package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/terravault/oracle/internal/storage"
)

// DeveloperHandlers handles developer-authenticated routes.
type DeveloperHandlers struct {
	cfg Config
}

// CreateProject registers a new project in the off-chain database.
func (h *DeveloperHandlers) CreateProject(w http.ResponseWriter, r *http.Request) {
	wallet, _ := r.Context().Value(ctxWallet).(string)

	var body struct {
		OnChainPubkey     string    `json:"on_chain_pubkey"`
		ProjectType       string    `json:"project_type"`
		MetadataURI       string    `json:"metadata_uri"`
		FundraiseTarget   int64     `json:"fundraise_target"`
		FundraiseHardCap  int64     `json:"fundraise_hard_cap"`
		FundraiseDeadline time.Time `json:"fundraise_deadline"`
		TokenPrice        int64     `json:"token_price"`
		TotalTokens       int64     `json:"total_tokens"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.OnChainPubkey == "" {
		writeError(w, http.StatusBadRequest, "on_chain_pubkey required")
		return
	}

	proj := &storage.Project{
		OnChainPubkey:     body.OnChainPubkey,
		DeveloperWallet:   wallet,
		State:             "Draft",
		ProjectType:       body.ProjectType,
		MetadataURI:       body.MetadataURI,
		FundraiseTarget:   body.FundraiseTarget,
		FundraiseHardCap:  body.FundraiseHardCap,
		FundraiseDeadline: body.FundraiseDeadline,
		TokenPrice:        body.TokenPrice,
		TotalTokens:       body.TotalTokens,
	}

	if err := h.cfg.DB.UpsertProject(proj); err != nil {
		h.cfg.Logger.Error("upserting project", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to create project")
		return
	}

	created, err := h.cfg.DB.GetProjectByPubkey(body.OnChainPubkey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "project created but failed to retrieve")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// ListDeveloperProjects returns all projects for the authenticated developer.
func (h *DeveloperHandlers) ListDeveloperProjects(w http.ResponseWriter, r *http.Request) {
	wallet, _ := r.Context().Value(ctxWallet).(string)
	// In production: filter by developer_wallet
	projects, err := h.cfg.DB.ListProjects(100, 0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list projects")
		return
	}
	var ownedProjects []storage.Project
	for _, p := range projects {
		if p.DeveloperWallet == wallet {
			ownedProjects = append(ownedProjects, p)
		}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"projects": ownedProjects})
}

// GetDeveloperProject returns detailed project info for the developer.
func (h *DeveloperHandlers) GetDeveloperProject(w http.ResponseWriter, r *http.Request) {
	pubkey := chi.URLParam(r, "pubkey")
	proj, err := h.cfg.DB.GetProjectByPubkey(pubkey)
	if err != nil {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	milestones, _ := h.cfg.DB.GetMilestonesByProject(proj.ID)
	distributions, _ := h.cfg.DB.GetDistributionsByProject(proj.ID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"project":       proj,
		"milestones":    milestones,
		"distributions": distributions,
	})
}

// AddMilestone records a new milestone off-chain.
func (h *DeveloperHandlers) AddMilestone(w http.ResponseWriter, r *http.Request) {
	pubkey := chi.URLParam(r, "pubkey")
	proj, err := h.cfg.DB.GetProjectByPubkey(pubkey)
	if err != nil {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}

	var body struct {
		MilestoneIndex int    `json:"milestone_index"`
		Description    string `json:"description"`
		ReleaseBPS     int    `json:"release_bps"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	m := &storage.Milestone{
		ProjectID:      proj.ID,
		MilestoneIndex: body.MilestoneIndex,
		Description:    body.Description,
		ReleaseBPS:     body.ReleaseBPS,
		Status:         "Pending",
	}
	if err := h.cfg.DB.UpsertMilestone(m); err != nil {
		h.cfg.Logger.Error("upserting milestone", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to add milestone")
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

// UploadDocument uploads a project document to IPFS and records it in the DB.
func (h *DeveloperHandlers) UploadDocument(w http.ResponseWriter, r *http.Request) {
	pubkey := chi.URLParam(r, "pubkey")
	proj, err := h.cfg.DB.GetProjectByPubkey(pubkey)
	if err != nil {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}

	r.ParseMultipartForm(10 << 20) // 10 MB max
	file, _, err := r.FormFile("document")
	if err != nil {
		writeError(w, http.StatusBadRequest, "document file required")
		return
	}
	defer file.Close()

	docType := r.FormValue("doc_type")
	if docType == "" {
		docType = "legal_doc"
	}

	cid, sha256hex, err := h.cfg.IPFS.UploadReader(file)
	if err != nil {
		h.cfg.Logger.Error("uploading to IPFS", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to upload document")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"project_id":  proj.ID,
		"ipfs_cid":    cid,
		"sha256_hash": sha256hex,
		"doc_type":    docType,
		"ipfs_url":    storage.BuildIPFSURL(cid),
	})
}

// DepositIncome records an income deposit event off-chain.
func (h *DeveloperHandlers) DepositIncome(w http.ResponseWriter, r *http.Request) {
	pubkey := chi.URLParam(r, "pubkey")
	proj, err := h.cfg.DB.GetProjectByPubkey(pubkey)
	if err != nil {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}

	var body struct {
		AmountUSDC   int64  `json:"amount_usdc"`
		Source       string `json:"source"`
		ClaimDeadline time.Time `json:"claim_deadline"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	dist := &storage.Distribution{
		ProjectID:     proj.ID,
		Round:         int32(proj.MilestoneCount), // simplified
		TotalUSDC:     body.AmountUSDC,
		Source:        body.Source,
		ClaimDeadline: body.ClaimDeadline,
	}
	if err := h.cfg.DB.InsertDistribution(dist); err != nil {
		h.cfg.Logger.Error("recording distribution", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to record distribution")
		return
	}
	writeJSON(w, http.StatusCreated, dist)
}


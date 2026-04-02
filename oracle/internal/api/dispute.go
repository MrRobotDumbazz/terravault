package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/terravault/oracle/internal/storage"
)

// DisputeHandlers handles dispute-related routes.
type DisputeHandlers struct {
	cfg Config
}

type raiseDisputeRequest struct {
	ProjectPubkey string `json:"project_pubkey"`
	ReasonHash    string `json:"reason_hash"` // hex-encoded [32]byte IPFS CID hash
}

type submitEvidenceRequest struct {
	ProjectPubkey string `json:"project_pubkey"`
	EvidenceHash  string `json:"evidence_hash"` // hex-encoded [32]byte IPFS CID hash
}

type resolveDisputeRequest struct {
	ProjectPubkey string `json:"project_pubkey"`
	Decision      string `json:"decision"` // PayInvestors | RefundAndExtend | ForceClose
}

// RaiseDispute records a dispute in the off-chain DB and returns the on-chain call params.
// POST /api/v1/disputes/raise
func (h *DisputeHandlers) RaiseDispute(w http.ResponseWriter, r *http.Request) {
	var req raiseDisputeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ProjectPubkey == "" || req.ReasonHash == "" {
		writeError(w, http.StatusBadRequest, "project_pubkey and reason_hash are required")
		return
	}

	callerWallet := walletFromContext(r.Context())
	deadline := time.Now().Add(72 * time.Hour)

	dispute := &storage.Dispute{
		ProjectPubkey: req.ProjectPubkey,
		RaisedBy:      callerWallet,
		ReasonHash:    req.ReasonHash,
		EvidenceHash:  req.ReasonHash,
		Status:        "open",
		Decision:      "",
		Deadline:      deadline,
	}

	if err := h.cfg.DB.UpsertDispute(dispute); err != nil {
		h.cfg.Logger.Error("upserting dispute", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to record dispute")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"project_pubkey": req.ProjectPubkey,
		"status":         "open",
		"deadline":       deadline.Unix(),
	})
}

// SubmitEvidence uploads evidence to IPFS (via configured IPFS client) and records the hash.
// POST /api/v1/disputes/evidence
func (h *DisputeHandlers) SubmitEvidence(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		writeError(w, http.StatusBadRequest, "failed to parse form")
		return
	}

	projectPubkey := r.FormValue("project_pubkey")
	if projectPubkey == "" {
		writeError(w, http.StatusBadRequest, "project_pubkey is required")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	cid, _, err := h.cfg.IPFS.UploadReader(file)
	if err != nil {
		h.cfg.Logger.Error("uploading evidence to IPFS", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to upload evidence")
		return
	}

	dispute, err := h.cfg.DB.GetDisputeByProject(projectPubkey)
	if err != nil {
		writeError(w, http.StatusNotFound, "no open dispute for this project")
		return
	}
	dispute.EvidenceHash = cid
	if err := h.cfg.DB.UpsertDispute(dispute); err != nil {
		h.cfg.Logger.Error("updating evidence hash", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to record evidence")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"project_pubkey": projectPubkey,
		"evidence_cid":   cid,
	})
}

// ResolveDispute records the admin resolution decision.
// POST /api/v1/disputes/resolve  (admin role required)
func (h *DisputeHandlers) ResolveDispute(w http.ResponseWriter, r *http.Request) {
	var req resolveDisputeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	switch req.Decision {
	case "PayInvestors", "RefundAndExtend", "ForceClose":
	default:
		writeError(w, http.StatusBadRequest, "decision must be PayInvestors | RefundAndExtend | ForceClose")
		return
	}

	dispute, err := h.cfg.DB.GetDisputeByProject(req.ProjectPubkey)
	if err != nil {
		writeError(w, http.StatusNotFound, "no open dispute for this project")
		return
	}

	now := time.Now()
	dispute.Status = "resolved"
	dispute.Decision = req.Decision
	dispute.ResolvedAt = &now

	if err := h.cfg.DB.UpsertDispute(dispute); err != nil {
		h.cfg.Logger.Error("resolving dispute", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to resolve dispute")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"project_pubkey": req.ProjectPubkey,
		"decision":       req.Decision,
		"status":         "resolved",
	})
}

// GetDispute returns the dispute record for a project.
// GET /api/v1/disputes/{project_id}
func (h *DisputeHandlers) GetDispute(w http.ResponseWriter, r *http.Request) {
	projectPubkey := chi.URLParam(r, "project_id")

	dispute, err := h.cfg.DB.GetDisputeByProject(projectPubkey)
	if err != nil {
		writeError(w, http.StatusNotFound, "dispute not found")
		return
	}

	writeJSON(w, http.StatusOK, dispute)
}

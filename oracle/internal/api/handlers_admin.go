package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// AdminHandlers handles admin-authenticated routes.
type AdminHandlers struct {
	cfg Config
}

// ListPendingKYC returns all KYC records with status 'pending'.
func (h *AdminHandlers) ListPendingKYC(w http.ResponseWriter, r *http.Request) {
	// In production: add a DB method to query by status
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"kyc_records": []interface{}{},
	})
}

// ApproveKYC manually approves KYC for a wallet address.
func (h *AdminHandlers) ApproveKYC(w http.ResponseWriter, r *http.Request) {
	wallet := chi.URLParam(r, "wallet")

	record, err := h.cfg.DB.GetKYCByWallet(wallet)
	if err != nil {
		writeError(w, http.StatusNotFound, "KYC record not found")
		return
	}

	now := time.Now()
	record.KYCStatus = "approved"
	record.VerifiedAt = &now
	record.VerificationLevel = 2

	if err := h.cfg.DB.UpsertKYC(record); err != nil {
		h.cfg.Logger.Error("approving KYC", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to approve KYC")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"wallet": wallet,
		"status": "approved",
	})
}

// RejectKYC manually rejects KYC for a wallet address.
func (h *AdminHandlers) RejectKYC(w http.ResponseWriter, r *http.Request) {
	wallet := chi.URLParam(r, "wallet")

	record, err := h.cfg.DB.GetKYCByWallet(wallet)
	if err != nil {
		writeError(w, http.StatusNotFound, "KYC record not found")
		return
	}

	record.KYCStatus = "rejected"
	record.VerifiedAt = nil

	if err := h.cfg.DB.UpsertKYC(record); err != nil {
		h.cfg.Logger.Error("rejecting KYC", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to reject KYC")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"wallet": wallet,
		"status": "rejected",
	})
}

// ListAllProjects returns all projects with admin-level detail.
func (h *AdminHandlers) ListAllProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.cfg.DB.ListProjects(1000, 0)
	if err != nil {
		h.cfg.Logger.Error("listing all projects for admin", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to list projects")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"projects": projects, "total": len(projects)})
}

// PauseProject marks a project as paused in the off-chain database.
func (h *AdminHandlers) PauseProject(w http.ResponseWriter, r *http.Request) {
	pubkey := chi.URLParam(r, "pubkey")
	proj, err := h.cfg.DB.GetProjectByPubkey(pubkey)
	if err != nil {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}

	proj.State = "Paused"
	if err := h.cfg.DB.UpsertProject(proj); err != nil {
		h.cfg.Logger.Error("pausing project", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to pause project")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"pubkey": pubkey,
		"state":  "Paused",
	})
}

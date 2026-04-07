package api

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/terravault/oracle/internal/storage"
)

// InternalHandlers handles internal oracle/listener API routes (API key auth).
type InternalHandlers struct {
	cfg Config
}

// SyncProject upserts a project record from on-chain data.
func (h *InternalHandlers) SyncProject(w http.ResponseWriter, r *http.Request) {
	var body storage.Project
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid project data")
		return
	}
	if err := h.cfg.DB.UpsertProject(&body); err != nil {
		h.cfg.Logger.Error("syncing project", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to sync project")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// SyncMilestone upserts a milestone record from on-chain data.
func (h *InternalHandlers) SyncMilestone(w http.ResponseWriter, r *http.Request) {
	var body storage.Milestone
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid milestone data")
		return
	}
	if err := h.cfg.DB.UpsertMilestone(&body); err != nil {
		h.cfg.Logger.Error("syncing milestone", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to sync milestone")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UpdateKYCStatus updates the KYC verification status for a wallet.
func (h *InternalHandlers) UpdateKYCStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		WalletAddress     string  `json:"wallet_address"`
		KYCStatus         string  `json:"kyc_status"`
		Country           string  `json:"country"`
		VerificationLevel int     `json:"verification_level"`
		ProviderSessionID string  `json:"provider_session_id"`
		VerifiedAt        *string `json:"verified_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid KYC data")
		return
	}

	record := &storage.KYCRecord{
		WalletAddress:     body.WalletAddress,
		KYCStatus:         body.KYCStatus,
		Country:           body.Country,
		VerificationLevel: body.VerificationLevel,
		ProviderSessionID: body.ProviderSessionID,
	}
	if body.VerifiedAt != nil {
		t, err := time.Parse(time.RFC3339, *body.VerifiedAt)
		if err == nil {
			record.VerifiedAt = &t
		}
	}

	if err := h.cfg.DB.UpsertKYC(record); err != nil {
		h.cfg.Logger.Error("updating KYC", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to update KYC")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RecordDistribution records a new distribution round.
func (h *InternalHandlers) RecordDistribution(w http.ResponseWriter, r *http.Request) {
	var body storage.Distribution
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid distribution data")
		return
	}
	if err := h.cfg.DB.InsertDistribution(&body); err != nil {
		h.cfg.Logger.Error("recording distribution", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to record distribution")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

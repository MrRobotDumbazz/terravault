package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// BlacklistHandlers handles developer blacklist routes.
type BlacklistHandlers struct {
	cfg Config
}

// GetBlacklist returns all flagged developer addresses.
// GET /api/v1/blacklist
func (h *BlacklistHandlers) GetBlacklist(w http.ResponseWriter, r *http.Request) {
	entries, err := h.cfg.DB.GetBlacklist()
	if err != nil {
		h.cfg.Logger.Error("fetching blacklist", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to fetch blacklist")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"entries": entries,
		"total":   len(entries),
	})
}

// CheckBlacklist checks if a specific developer pubkey is flagged.
// GET /api/v1/blacklist/{pubkey}
func (h *BlacklistHandlers) CheckBlacklist(w http.ResponseWriter, r *http.Request) {
	pubkey := chi.URLParam(r, "pubkey")

	entry, err := h.cfg.DB.GetBlacklistEntry(pubkey)
	if err != nil {
		// Not found means not blacklisted
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"pubkey":        pubkey,
			"is_blacklisted": false,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"pubkey":        pubkey,
		"is_blacklisted": true,
		"reason_hash":   entry.ReasonHash,
		"flagged_at":    entry.FlaggedAt.Unix(),
	})
}

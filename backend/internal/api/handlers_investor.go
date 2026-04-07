package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// InvestorHandlers handles investor-authenticated routes.
type InvestorHandlers struct {
	cfg Config
}

// GetPosition returns an investor's position in a specific project.
func (h *InvestorHandlers) GetPosition(w http.ResponseWriter, r *http.Request) {
	wallet, _ := r.Context().Value(ctxWallet).(string)
	projectPubkey := chi.URLParam(r, "pubkey")

	proj, err := h.cfg.DB.GetProjectByPubkey(projectPubkey)
	if err != nil {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}

	_ = proj
	// In a full implementation, query the on-chain InvestorPosition via RPC
	// and cross-reference with DB investor_positions table.
	// For now, return a structured response indicating the wallet and project.
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"wallet":          wallet,
		"project_pubkey":  projectPubkey,
		"tokens_held":     0,
		"usdc_invested":   0,
		"last_claimed_round": 0,
		"total_claimed_usdc": 0,
		"kyc_verified":    false,
	})
}

// GetPortfolio returns all projects an investor holds tokens in.
func (h *InvestorHandlers) GetPortfolio(w http.ResponseWriter, r *http.Request) {
	wallet, _ := r.Context().Value(ctxWallet).(string)
	_ = wallet
	// Query investor positions from DB (indexed by wallet)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"positions": []interface{}{},
	})
}

// GetClaimableDistributions returns distribution rounds the investor can claim.
func (h *InvestorHandlers) GetClaimableDistributions(w http.ResponseWriter, r *http.Request) {
	wallet, _ := r.Context().Value(ctxWallet).(string)
	_ = wallet
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"claimable": []interface{}{},
	})
}

// InitiateKYC starts a KYC verification session.
func (h *InvestorHandlers) InitiateKYC(w http.ResponseWriter, r *http.Request) {
	wallet, _ := r.Context().Value(ctxWallet).(string)

	// Check if already verified
	existing, _ := h.cfg.DB.GetKYCByWallet(wallet)
	if existing != nil && existing.KYCStatus == "approved" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status":     "approved",
			"message":    "KYC already verified",
			"session_id": existing.ProviderSessionID,
		})
		return
	}

	// For demo: return a mock session
	// In production: call PersonaProvider.InitiateVerification
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"session_id":   "persona_session_" + wallet[:8],
		"redirect_url": "https://withpersona.com/verify?session=demo",
		"status":       "pending",
	})
}

// GetKYCStatus returns the KYC status for the authenticated wallet.
func (h *InvestorHandlers) GetKYCStatus(w http.ResponseWriter, r *http.Request) {
	wallet, _ := r.Context().Value(ctxWallet).(string)

	record, err := h.cfg.DB.GetKYCByWallet(wallet)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"wallet": wallet,
			"status": "not_started",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"wallet":      record.WalletAddress,
		"status":      record.KYCStatus,
		"country":     record.Country,
		"verified_at": record.VerifiedAt,
	})
}

// writeJSON is a helper to write JSON responses.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// Best-effort: response already started
		return
	}
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// walletFromContext extracts the authenticated wallet address from request context.
func walletFromContext(ctx context.Context) string {
	wallet, _ := ctx.Value(ctxWallet).(string)
	return wallet
}


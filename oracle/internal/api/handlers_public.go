package api

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/terravault/oracle/internal/storage"
)

// PublicHandlers handles unauthenticated public routes.
type PublicHandlers struct {
	cfg Config
	// In-memory nonce store (production: use Redis with TTL)
	nonces   map[string]nonceEntry
	noncesMu sync.Mutex
}

type nonceEntry struct {
	nonce     string
	expiresAt time.Time
}

func (h *PublicHandlers) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": "1.0.0"})
}

func (h *PublicHandlers) ListProjects(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	projects, err := h.cfg.DB.ListProjects(limit, offset)
	if err != nil {
		h.cfg.Logger.Error("listing projects", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to list projects")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"projects": projects, "limit": limit, "offset": offset})
}

func (h *PublicHandlers) GetProject(w http.ResponseWriter, r *http.Request) {
	pubkey := chi.URLParam(r, "pubkey")
	proj, err := h.cfg.DB.GetProjectByPubkey(pubkey)
	if err != nil {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	writeJSON(w, http.StatusOK, proj)
}

func (h *PublicHandlers) GetMilestones(w http.ResponseWriter, r *http.Request) {
	pubkey := chi.URLParam(r, "pubkey")
	proj, err := h.cfg.DB.GetProjectByPubkey(pubkey)
	if err != nil {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	milestones, err := h.cfg.DB.GetMilestonesByProject(proj.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get milestones")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"milestones": milestones})
}

func (h *PublicHandlers) GetDistributions(w http.ResponseWriter, r *http.Request) {
	pubkey := chi.URLParam(r, "pubkey")
	proj, err := h.cfg.DB.GetProjectByPubkey(pubkey)
	if err != nil {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	distributions, err := h.cfg.DB.GetDistributionsByProject(proj.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get distributions")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"distributions": distributions})
}

// GetChallenge issues a random nonce that the wallet must sign.
func (h *PublicHandlers) GetChallenge(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Wallet string `json:"wallet"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Wallet == "" {
		writeError(w, http.StatusBadRequest, "wallet address required")
		return
	}

	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate nonce")
		return
	}
	nonceStr := base64.StdEncoding.EncodeToString(nonce)
	message := fmt.Sprintf("TerraVault authentication\nWallet: %s\nNonce: %s", body.Wallet, nonceStr)

	h.noncesMu.Lock()
	if h.nonces == nil {
		h.nonces = make(map[string]nonceEntry)
	}
	h.nonces[body.Wallet] = nonceEntry{nonce: nonceStr, expiresAt: time.Now().Add(5 * time.Minute)}
	h.noncesMu.Unlock()

	writeJSON(w, http.StatusOK, map[string]string{
		"message": message,
		"nonce":   nonceStr,
	})
}

// VerifySignature verifies a wallet signature over the challenge and issues a JWT.
func (h *PublicHandlers) VerifySignature(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Wallet    string `json:"wallet"`
		Signature string `json:"signature"` // base64 or hex
		Role      string `json:"role"`      // "investor" | "developer"
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	h.noncesMu.Lock()
	entry, ok := h.nonces[body.Wallet]
	if ok {
		delete(h.nonces, body.Wallet)
	}
	h.noncesMu.Unlock()

	if !ok || time.Now().After(entry.expiresAt) {
		writeError(w, http.StatusUnauthorized, "challenge expired or not found")
		return
	}

	// Reconstruct the signed message
	message := fmt.Sprintf("TerraVault authentication\nWallet: %s\nNonce: %s", body.Wallet, entry.nonce)

	// Decode signature
	var sigBytes []byte
	var err error
	sigBytes, err = base64.StdEncoding.DecodeString(body.Signature)
	if err != nil {
		sigBytes, err = hex.DecodeString(body.Signature)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid signature encoding")
			return
		}
	}

	// Decode wallet public key from base58
	pubkeyBytes := decodeBase58(body.Wallet)
	if len(pubkeyBytes) != ed25519.PublicKeySize {
		writeError(w, http.StatusBadRequest, "invalid wallet address")
		return
	}

	// Verify Ed25519 signature
	if !ed25519.Verify(ed25519.PublicKey(pubkeyBytes), []byte(message), sigBytes) {
		writeError(w, http.StatusUnauthorized, "signature verification failed")
		return
	}

	// Role is determined server-side — never trust client-supplied role.
	// Admin is granted only if the wallet is in the admin whitelist.
	role := "investor"
	if body.Role == "developer" {
		role = "developer"
	}
	if _, isAdmin := h.cfg.AdminWallets[body.Wallet]; isAdmin {
		role = "admin"
	}

	jti := uuid.New().String()
	token, err := IssueJWT(h.cfg.JWTSecret, body.Wallet, role, jti, 24*time.Hour)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to issue token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"token":  token,
		"wallet": body.Wallet,
		"role":   role,
	})
}

// PersonaWebhook handles Persona KYC webhooks.
func (h *PublicHandlers) PersonaWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	var event struct {
		Data struct {
			Attributes struct {
				Payload struct {
					Data struct {
						Attributes struct {
							Status      string `json:"status"`
							ReferenceID string `json:"reference-id"`
						} `json:"attributes"`
					} `json:"data"`
				} `json:"payload"`
			} `json:"attributes"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &event); err != nil {
		writeError(w, http.StatusBadRequest, "invalid webhook payload")
		return
	}

	attr := event.Data.Attributes.Payload.Data.Attributes
	if attr.ReferenceID != "" {
		record, err := h.cfg.DB.GetKYCByWallet(attr.ReferenceID)
		if err != nil || record == nil {
			record = &storage.KYCRecord{WalletAddress: attr.ReferenceID}
		}
		record.KYCStatus = attr.Status
		if attr.Status == "approved" {
			now := time.Now()
			record.VerifiedAt = &now
		}
		h.cfg.DB.UpsertKYC(record)
	}

	w.WriteHeader(http.StatusOK)
}

// decodeBase58 decodes a base58-encoded public key.
func decodeBase58(s string) []byte {
	const alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	result := make([]byte, 0, 32)
	for _, c := range s {
		val := strings.IndexRune(alphabet, c)
		if val < 0 {
			return nil
		}
		carry := val
		for j := len(result) - 1; j >= 0; j-- {
			carry += 58 * int(result[j])
			result[j] = byte(carry & 0xff)
			carry >>= 8
		}
		for carry > 0 {
			result = append([]byte{byte(carry & 0xff)}, result...)
			carry >>= 8
		}
	}
	// Add leading zeros for leading '1's
	for _, c := range s {
		if c != '1' {
			break
		}
		result = append([]byte{0}, result...)
	}
	return result
}


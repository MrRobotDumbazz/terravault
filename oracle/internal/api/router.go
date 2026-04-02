package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/terravault/oracle/internal/storage"
)

// Config holds all dependencies for the API router.
type Config struct {
	DB             *storage.DB
	IPFS           *storage.IPFSClient
	Logger         *zap.Logger
	JWTSecret      []byte
	InternalAPIKey string
	SolanaRPCURL   string
	// AdminWallets is the set of wallet addresses that receive the "admin" role on auth.
	AdminWallets   map[string]struct{}
}

// NewRouter constructs the full Chi router with all API routes.
func NewRouter(cfg Config) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(zapLogger(cfg.Logger))
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	ph := &PublicHandlers{cfg: cfg}
	ih := &InvestorHandlers{cfg: cfg}
	dh := &DeveloperHandlers{cfg: cfg}
	intH := &InternalHandlers{cfg: cfg}
	ah := &AdminHandlers{cfg: cfg}
	dispH := &DisputeHandlers{cfg: cfg}
	blH := &BlacklistHandlers{cfg: cfg}

	// ── Public endpoints (no auth) ─────────────────────────────────────────
	r.Group(func(r chi.Router) {
		r.Get("/health", ph.Health)
		r.Get("/api/v1/projects", ph.ListProjects)
		r.Get("/api/v1/projects/{pubkey}", ph.GetProject)
		r.Get("/api/v1/projects/{pubkey}/milestones", ph.GetMilestones)
		r.Get("/api/v1/projects/{pubkey}/distributions", ph.GetDistributions)
		r.Post("/api/v1/auth/challenge", ph.GetChallenge)
		r.Post("/api/v1/auth/verify", ph.VerifySignature)
	})

	// ── Investor endpoints (JWT auth) ──────────────────────────────────────
	r.Group(func(r chi.Router) {
		r.Use(jwtMiddleware(cfg.JWTSecret))
		r.Get("/api/v1/investor/position/{pubkey}", ih.GetPosition)
		r.Get("/api/v1/investor/portfolio", ih.GetPortfolio)
		r.Get("/api/v1/investor/distributions", ih.GetClaimableDistributions)
		r.Post("/api/v1/kyc/initiate", ih.InitiateKYC)
		r.Get("/api/v1/kyc/status", ih.GetKYCStatus)
	})

	// ── Developer endpoints (JWT auth + developer role) ────────────────────
	r.Group(func(r chi.Router) {
		r.Use(jwtMiddleware(cfg.JWTSecret))
		r.Use(requireRole("developer"))
		r.Post("/api/v1/developer/projects", dh.CreateProject)
		r.Get("/api/v1/developer/projects", dh.ListDeveloperProjects)
		r.Get("/api/v1/developer/projects/{pubkey}", dh.GetDeveloperProject)
		r.Post("/api/v1/developer/projects/{pubkey}/milestones", dh.AddMilestone)
		r.Post("/api/v1/developer/projects/{pubkey}/documents", dh.UploadDocument)
		r.Post("/api/v1/developer/projects/{pubkey}/income", dh.DepositIncome)
	})

	// ── Internal endpoints (API key auth — oracle/listener use these) ──────
	r.Group(func(r chi.Router) {
		r.Use(apiKeyMiddleware(cfg.InternalAPIKey))
		r.Post("/internal/v1/projects/sync", intH.SyncProject)
		r.Post("/internal/v1/milestones/sync", intH.SyncMilestone)
		r.Post("/internal/v1/kyc/update", intH.UpdateKYCStatus)
		r.Post("/internal/v1/distributions/record", intH.RecordDistribution)
	})

	// ── Admin endpoints (admin role JWT) ──────────────────────────────────
	r.Group(func(r chi.Router) {
		r.Use(jwtMiddleware(cfg.JWTSecret))
		r.Use(requireRole("admin"))
		r.Get("/api/v1/admin/kyc/pending", ah.ListPendingKYC)
		r.Post("/api/v1/admin/kyc/{wallet}/approve", ah.ApproveKYC)
		r.Post("/api/v1/admin/kyc/{wallet}/reject", ah.RejectKYC)
		r.Get("/api/v1/admin/projects", ah.ListAllProjects)
		r.Post("/api/v1/admin/projects/{pubkey}/pause", ah.PauseProject)
		r.Post("/api/v1/disputes/resolve", dispH.ResolveDispute)
	})

	// ── Dispute endpoints (JWT auth — any authenticated user) ──────────────
	r.Group(func(r chi.Router) {
		r.Use(jwtMiddleware(cfg.JWTSecret))
		r.Post("/api/v1/disputes/raise", dispH.RaiseDispute)
		r.Post("/api/v1/disputes/evidence", dispH.SubmitEvidence)
		r.Get("/api/v1/disputes/{project_id}", dispH.GetDispute)
	})

	// ── Blacklist endpoints (public read) ──────────────────────────────────
	r.Get("/api/v1/blacklist", blH.GetBlacklist)
	r.Get("/api/v1/blacklist/{pubkey}", blH.CheckBlacklist)

	// ── KYC webhook (provider-signed) ─────────────────────────────────────
	r.Post("/webhooks/kyc/persona", ph.PersonaWebhook)

	return r
}

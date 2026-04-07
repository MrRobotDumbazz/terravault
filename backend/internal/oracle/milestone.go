package oracle

import (
	"context"
	"crypto/sha256"
	"fmt"

	"go.uber.org/zap"

	"github.com/terravault/oracle/internal/storage"
)

// MilestoneConfig holds dependencies for the milestone handler.
type MilestoneConfig struct {
	DB           *storage.DB
	Signer       *Signer
	ProgramID    string
	SolanaRPCURL string
	Logger       *zap.Logger
}

// MilestoneHandler processes on-chain milestone events and signs proofs.
type MilestoneHandler struct {
	cfg MilestoneConfig
	tx  *TransactionSender
}

// NewMilestoneHandler creates a new MilestoneHandler.
func NewMilestoneHandler(cfg MilestoneConfig) *MilestoneHandler {
	tx := NewTransactionSender(cfg.SolanaRPCURL, cfg.Signer, cfg.ProgramID)
	return &MilestoneHandler{cfg: cfg, tx: tx}
}

// HandleMilestoneSubmitted is called when a MilestoneProofSubmitted event is observed.
// For the oracle this is already done (oracle signed and submitted on-chain).
// This handler updates the off-chain database.
func (h *MilestoneHandler) HandleMilestoneSubmitted(ctx context.Context, event MilestoneSubmittedEvent) error {
	h.cfg.Logger.Info("milestone proof submitted",
		zap.String("project", event.ProjectPubkey),
		zap.Uint8("index", event.MilestoneIndex),
	)

	proj, err := h.cfg.DB.GetProjectByPubkey(event.ProjectPubkey)
	if err != nil {
		return fmt.Errorf("getting project %s: %w", event.ProjectPubkey, err)
	}

	milestones, err := h.cfg.DB.GetMilestonesByProject(proj.ID)
	if err != nil {
		return fmt.Errorf("getting milestones for project %d: %w", proj.ID, err)
	}

	for i, m := range milestones {
		if m.MilestoneIndex == int(event.MilestoneIndex) {
			milestones[i].Status = "Submitted"
			milestones[i].ProofHash = event.ProofHash
			return h.cfg.DB.UpsertMilestone(&milestones[i])
		}
	}
	return fmt.Errorf("milestone %d not found in DB for project %s", event.MilestoneIndex, event.ProjectPubkey)
}

// HandleFundraisingStarted updates the project state in DB.
func (h *MilestoneHandler) HandleFundraisingStarted(ctx context.Context, event FundraisingStartedEvent) error {
	h.cfg.Logger.Info("fundraising started", zap.String("project", event.ProjectPubkey))
	proj, err := h.cfg.DB.GetProjectByPubkey(event.ProjectPubkey)
	if err != nil {
		return err
	}
	proj.State = "Fundraising"
	return h.cfg.DB.UpsertProject(proj)
}

// HandleProjectActivated updates the project state to InMilestones.
func (h *MilestoneHandler) HandleProjectActivated(ctx context.Context, event ProjectActivatedEvent) error {
	h.cfg.Logger.Info("project activated", zap.String("project", event.ProjectPubkey))
	proj, err := h.cfg.DB.GetProjectByPubkey(event.ProjectPubkey)
	if err != nil {
		return err
	}
	proj.State = "InMilestones"
	proj.TotalRaised = event.TotalRaisedUSDC
	return h.cfg.DB.UpsertProject(proj)
}

// SignProofForMilestone generates an oracle Ed25519 signature over a proof hash.
// The proofURI is first hashed with SHA-256 to produce a consistent 32-byte message.
func (h *MilestoneHandler) SignProofForMilestone(proofURI string, externalHash [32]byte) ([64]byte, [32]byte, error) {
	var proofHash [32]byte
	if externalHash != ([32]byte{}) {
		proofHash = externalHash
	} else {
		proofHash = sha256.Sum256([]byte(proofURI))
	}

	sig, err := h.cfg.Signer.SignProofHash(proofHash)
	if err != nil {
		return [64]byte{}, [32]byte{}, fmt.Errorf("signing proof: %w", err)
	}
	return sig, proofHash, nil
}

// ─── Event types emitted by the listener ────────────────────────────────────

type MilestoneSubmittedEvent struct {
	ProjectPubkey  string
	MilestoneIndex uint8
	ProofHash      string
	SubmittedAt    int64
	DisputeDeadline int64
}

type FundraisingStartedEvent struct {
	ProjectPubkey string
}

type ProjectActivatedEvent struct {
	ProjectPubkey   string
	TotalRaisedUSDC int64
}

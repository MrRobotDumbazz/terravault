package kyc

import "context"

// Status represents the KYC verification result.
type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
	StatusExpired  Status = "expired"
)

// VerificationResult is the outcome of a KYC check.
type VerificationResult struct {
	WalletAddress     string
	Status            Status
	Country           string
	VerificationLevel int
	ProviderSessionID string
}

// Provider is the interface every KYC provider must implement.
type Provider interface {
	// InitiateVerification starts a KYC session and returns a redirect URL.
	InitiateVerification(ctx context.Context, walletAddress string) (sessionID string, redirectURL string, err error)

	// GetVerificationStatus retrieves the current status of a KYC session.
	GetVerificationStatus(ctx context.Context, sessionID string) (*VerificationResult, error)

	// HandleWebhook processes an inbound webhook payload from the provider.
	// Returns the VerificationResult extracted from the payload.
	HandleWebhook(ctx context.Context, payload []byte, signature string) (*VerificationResult, error)
}

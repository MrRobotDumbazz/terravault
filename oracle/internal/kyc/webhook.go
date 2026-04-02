package kyc

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// WebhookVerifier validates inbound KYC provider webhook signatures.
type WebhookVerifier struct {
	secret    []byte
	tolerance time.Duration
}

// NewWebhookVerifier creates a verifier with the given HMAC secret and timestamp tolerance.
func NewWebhookVerifier(secret string, tolerance time.Duration) *WebhookVerifier {
	return &WebhookVerifier{
		secret:    []byte(secret),
		tolerance: tolerance,
	}
}

// VerifyPersonaSignature checks the Persona-Signature header.
// Persona signature format: t=<timestamp>,v1=<hmac_hex>
func (v *WebhookVerifier) VerifyPersonaSignature(r *http.Request, body []byte) error {
	sigHeader := r.Header.Get("Persona-Signature")
	if sigHeader == "" {
		return fmt.Errorf("missing Persona-Signature header")
	}

	parts := strings.Split(sigHeader, ",")
	if len(parts) < 2 {
		return fmt.Errorf("invalid signature format")
	}

	var timestamp string
	var sigHex string
	for _, p := range parts {
		if strings.HasPrefix(p, "t=") {
			timestamp = strings.TrimPrefix(p, "t=")
		} else if strings.HasPrefix(p, "v1=") {
			sigHex = strings.TrimPrefix(p, "v1=")
		}
	}

	if timestamp == "" || sigHex == "" {
		return fmt.Errorf("could not parse timestamp or signature")
	}

	// Compute expected HMAC
	payload := timestamp + "." + string(body)
	mac := hmac.New(sha256.New, v.secret)
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(sigHex)) {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}

// ProcessWebhookResult dispatches a verified KYC result to the database updater.
type WebhookProcessor struct {
	provider Provider
	onResult func(ctx context.Context, result *VerificationResult) error
}

// NewWebhookProcessor creates a new WebhookProcessor.
func NewWebhookProcessor(provider Provider, onResult func(ctx context.Context, result *VerificationResult) error) *WebhookProcessor {
	return &WebhookProcessor{provider: provider, onResult: onResult}
}

// Process handles a raw webhook payload.
func (wp *WebhookProcessor) Process(ctx context.Context, payload []byte, sig string) error {
	result, err := wp.provider.HandleWebhook(ctx, payload, sig)
	if err != nil {
		return fmt.Errorf("handling webhook: %w", err)
	}
	return wp.onResult(ctx, result)
}

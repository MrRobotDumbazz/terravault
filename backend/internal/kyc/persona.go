package kyc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const personaAPIBase = "https://withpersona.com/api/v1"

// PersonaProvider implements the KYC Provider interface using Persona.
type PersonaProvider struct {
	apiKey     string
	templateID string
	httpClient *http.Client
}

// NewPersonaProvider creates a new Persona KYC provider.
func NewPersonaProvider(apiKey, templateID string) *PersonaProvider {
	return &PersonaProvider{
		apiKey:     apiKey,
		templateID: templateID,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// InitiateVerification creates a Persona inquiry for the given wallet address.
func (p *PersonaProvider) InitiateVerification(ctx context.Context, walletAddress string) (string, string, error) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"attributes": map[string]interface{}{
				"inquiry-template-id": p.templateID,
				"reference-id":        walletAddress,
			},
		},
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, personaAPIBase+"/inquiries", bytes.NewReader(body))
	if err != nil {
		return "", "", fmt.Errorf("building persona request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Persona-Version", "2023-01-05")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("calling persona API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", "", fmt.Errorf("persona API error %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			ID         string `json:"id"`
			Attributes struct {
				Status      string `json:"status"`
				RedirectURL string `json:"redirect-url"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("decoding persona response: %w", err)
	}

	return result.Data.ID, result.Data.Attributes.RedirectURL, nil
}

// GetVerificationStatus retrieves the current status of a Persona inquiry.
func (p *PersonaProvider) GetVerificationStatus(ctx context.Context, sessionID string) (*VerificationResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		personaAPIBase+"/inquiries/"+sessionID, nil)
	if err != nil {
		return nil, fmt.Errorf("building persona status request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Persona-Version", "2023-01-05")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling persona status API: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Attributes struct {
				Status     string `json:"status"`
				ReferenceID string `json:"reference-id"`
				Fields     struct {
					CountryCode struct {
						Value string `json:"value"`
					} `json:"country-code"`
				} `json:"fields"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding persona status: %w", err)
	}

	attr := result.Data.Attributes
	kyc := &VerificationResult{
		WalletAddress:     attr.ReferenceID,
		ProviderSessionID: sessionID,
		Country:           attr.Fields.CountryCode.Value,
		VerificationLevel: 1,
	}
	switch attr.Status {
	case "approved":
		kyc.Status = StatusApproved
	case "declined", "failed":
		kyc.Status = StatusRejected
	case "expired":
		kyc.Status = StatusExpired
	default:
		kyc.Status = StatusPending
	}
	return kyc, nil
}

// HandleWebhook processes a Persona webhook notification.
func (p *PersonaProvider) HandleWebhook(ctx context.Context, payload []byte, signature string) (*VerificationResult, error) {
	// In production: verify HMAC signature using p.apiKey
	// For now: parse the payload
	var event struct {
		Data struct {
			Attributes struct {
				Status  string `json:"status"`
				Payload struct {
					Data struct {
						Attributes struct {
							Status     string `json:"status"`
							ReferenceID string `json:"reference-id"`
							Fields     struct {
								CountryCode struct {
									Value string `json:"value"`
								} `json:"country-code"`
							} `json:"fields"`
						} `json:"attributes"`
					} `json:"data"`
				} `json:"payload"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("parsing persona webhook: %w", err)
	}

	attr := event.Data.Attributes.Payload.Data.Attributes
	kyc := &VerificationResult{
		WalletAddress:     attr.ReferenceID,
		Country:           attr.Fields.CountryCode.Value,
		VerificationLevel: 1,
	}
	switch attr.Status {
	case "approved":
		kyc.Status = StatusApproved
	case "declined", "failed":
		kyc.Status = StatusRejected
	default:
		kyc.Status = StatusPending
	}
	return kyc, nil
}

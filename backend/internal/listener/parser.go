package listener

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"

	"github.com/terravault/oracle/internal/oracle"
)

// ParsedEvent is a discriminated union of all possible on-chain events.
// Concrete types: oracle.MilestoneSubmittedEvent, oracle.FundraisingStartedEvent, oracle.ProjectActivatedEvent.
type ParsedEvent = any

// Parser extracts TerraVault events from Solana program log lines.
type Parser struct {
	programID string
}

// NewParser creates a new log parser.
func NewParser(programID string) *Parser {
	return &Parser{programID: programID}
}

// ParseLogs scans log lines and returns all recognizable TerraVault events.
// Anchor emits events as base64-encoded borsh in lines like:
//   Program data: <base64>
func (p *Parser) ParseLogs(logs []string) []ParsedEvent {
	var events []ParsedEvent
	inProgram := false

	for _, line := range logs {
		if strings.Contains(line, "Program "+p.programID+" invoke") {
			inProgram = true
			continue
		}
		if strings.Contains(line, "Program "+p.programID+" success") ||
			strings.Contains(line, "Program "+p.programID+" failed") {
			inProgram = false
			continue
		}
		if !inProgram {
			continue
		}

		const prefix = "Program data: "
		if !strings.HasPrefix(line, prefix) {
			continue
		}

		b64 := strings.TrimPrefix(line, prefix)
		data, err := base64.StdEncoding.DecodeString(b64)
		if err != nil || len(data) < 8 {
			continue
		}

		event := p.decodeEvent(data)
		if event != nil {
			events = append(events, event)
		}
	}
	return events
}

// decodeEvent attempts to decode a Borsh-encoded Anchor event from raw bytes.
// The first 8 bytes are the event discriminator (sha256("event:<EventName>")[0:8]).
func (p *Parser) decodeEvent(data []byte) ParsedEvent {
	if len(data) < 8 {
		return nil
	}

	disc := [8]byte{}
	copy(disc[:], data[:8])
	payload := data[8:]

	switch {
	case disc == milestoneProofSubmittedDisc:
		return decodeMilestoneSubmittedEvent(payload)
	case disc == fundraisingStartedDisc:
		return decodeFundraisingStartedEvent(payload)
	case disc == projectActivatedDisc:
		return decodeProjectActivatedEvent(payload)
	}
	return nil
}

// Event discriminators — computed as sha256("event:<Name>")[0:8]
// These would typically be generated from the IDL; hard-coded here for clarity.
var (
	milestoneProofSubmittedDisc = eventDisc("MilestoneProofSubmitted")
	fundraisingStartedDisc      = eventDisc("FundraisingStarted")
	projectActivatedDisc        = eventDisc("ProjectActivated")
)

func eventDisc(name string) [8]byte {
	h := sha256.Sum256([]byte("event:" + name))
	var d [8]byte
	copy(d[:], h[:8])
	return d
}

func decodeMilestoneSubmittedEvent(payload []byte) ParsedEvent {
	// Borsh layout: project(32) + milestone_index(1) + proof_uri_hash(32) + submitted_at(8) + dispute_deadline(8)
	if len(payload) < 32+1+32+8+8 {
		return nil
	}
	projectPubkey := base58Encode(payload[:32])
	milestoneIndex := payload[32]
	proofHash := payload[33:65]
	// submitted_at = int64 LE at [65:73]
	// dispute_deadline = int64 LE at [73:81]
	return oracle.MilestoneSubmittedEvent{
		ProjectPubkey:  projectPubkey,
		MilestoneIndex: milestoneIndex,
		ProofHash:      base58Encode(proofHash),
	}
}

func decodeFundraisingStartedEvent(payload []byte) ParsedEvent {
	if len(payload) < 32 {
		return nil
	}
	return oracle.FundraisingStartedEvent{
		ProjectPubkey: base58Encode(payload[:32]),
	}
}

func decodeProjectActivatedEvent(payload []byte) ParsedEvent {
	if len(payload) < 32+8 {
		return nil
	}
	total := int64(uint64(payload[32]) |
		uint64(payload[33])<<8 |
		uint64(payload[34])<<16 |
		uint64(payload[35])<<24 |
		uint64(payload[36])<<32 |
		uint64(payload[37])<<40 |
		uint64(payload[38])<<48 |
		uint64(payload[39])<<56)
	return oracle.ProjectActivatedEvent{
		ProjectPubkey:   base58Encode(payload[:32]),
		TotalRaisedUSDC: total,
	}
}

// base58Encode converts raw bytes to a base58-encoded Solana public key string.
func base58Encode(b []byte) string {
	const alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	// Simple base58 encode
	result := []byte{}
	x := make([]byte, len(b))
	copy(x, b)

	for _, v := range x {
		carry := int(v)
		for j := len(result) - 1; j >= 0; j-- {
			carry += 256 * int(result[j])
			result[j] = byte(carry % 58)
			carry /= 58
		}
		for carry > 0 {
			result = append([]byte{byte(carry % 58)}, result...)
			carry /= 58
		}
	}

	// Add leading '1's for leading zero bytes
	for _, v := range x {
		if v != 0 {
			break
		}
		result = append([]byte{0}, result...)
	}

	encoded := make([]byte, len(result))
	for i, v := range result {
		encoded[i] = alphabet[v]
	}
	return string(encoded)
}

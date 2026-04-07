package oracle

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gagliardetto/solana-go"
)

// Signer holds the oracle Ed25519 keypair and provides signing utilities.
type Signer struct {
	account solana.PrivateKey
}

// NewSignerFromFile loads a Solana keypair from a JSON file (array of 64 bytes).
func NewSignerFromFile(path string) (*Signer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading keypair file %s: %w", path, err)
	}

	var rawBytes []byte
	// Try array format first (Solana CLI format)
	var byteSlice []int
	if err := json.Unmarshal(data, &byteSlice); err == nil {
		rawBytes = make([]byte, len(byteSlice))
		for i, b := range byteSlice {
			rawBytes[i] = byte(b)
		}
	} else {
		// Try base58 string format
		var keyStr string
		if err := json.Unmarshal(data, &keyStr); err != nil {
			return nil, fmt.Errorf("parsing keypair: %w", err)
		}
		pk, err := solana.PrivateKeyFromBase58(keyStr)
		if err != nil {
			return nil, fmt.Errorf("decoding base58 keypair: %w", err)
		}
		return &Signer{account: pk}, nil
	}

	if len(rawBytes) != 64 {
		return nil, fmt.Errorf("invalid keypair length %d, expected 64", len(rawBytes))
	}

	pk := solana.PrivateKey(rawBytes)
	return &Signer{account: pk}, nil
}

// PublicKey returns the oracle's public key.
func (s *Signer) PublicKey() solana.PublicKey {
	return s.account.PublicKey()
}

// Sign signs the given message bytes using Ed25519.
// Returns a 64-byte signature.
func (s *Signer) Sign(message []byte) ([]byte, error) {
	sig, err := s.account.Sign(message)
	if err != nil {
		return nil, fmt.Errorf("signing: %w", err)
	}
	return sig[:], nil
}

// SignProofHash signs a 32-byte proof hash and returns the 64-byte signature.
func (s *Signer) SignProofHash(hash [32]byte) ([64]byte, error) {
	sig, err := s.account.Sign(hash[:])
	if err != nil {
		return [64]byte{}, fmt.Errorf("signing proof hash: %w", err)
	}
	return sig, nil
}

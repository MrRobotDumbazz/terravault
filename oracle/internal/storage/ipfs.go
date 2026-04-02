package storage

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	shell "github.com/ipfs/go-ipfs-api"
)

// IPFSClient wraps the IPFS shell for TerraVault document storage.
type IPFSClient struct {
	sh *shell.Shell
}

// NewIPFSClient creates a new IPFS client connected to the given API URL.
func NewIPFSClient(apiURL string) *IPFSClient {
	return &IPFSClient{sh: shell.NewShell(apiURL)}
}

// UploadBytes uploads raw bytes to IPFS and returns (cid, sha256hex, error).
func (c *IPFSClient) UploadBytes(data []byte) (cid string, sha256hex string, err error) {
	hash := sha256.Sum256(data)
	sha256hex = hex.EncodeToString(hash[:])

	cid, err = c.sh.Add(bytes.NewReader(data))
	if err != nil {
		return "", "", fmt.Errorf("ipfs add: %w", err)
	}
	return cid, sha256hex, nil
}

// UploadReader uploads data from a reader to IPFS.
func (c *IPFSClient) UploadReader(r io.Reader) (cid string, sha256hex string, err error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", "", fmt.Errorf("reading upload data: %w", err)
	}
	return c.UploadBytes(data)
}

// GetBytes retrieves content from IPFS by CID.
func (c *IPFSClient) GetBytes(cid string) ([]byte, error) {
	r, err := c.sh.Cat(cid)
	if err != nil {
		return nil, fmt.Errorf("ipfs cat %s: %w", cid, err)
	}
	defer r.Close()
	return io.ReadAll(r)
}

// Pin explicitly pins a CID to prevent garbage collection.
func (c *IPFSClient) Pin(cid string) error {
	return c.sh.Pin(cid)
}

// BuildIPFSURL returns the IPFS gateway URL for a given CID.
func BuildIPFSURL(cid string) string {
	return fmt.Sprintf("ipfs://%s", cid)
}

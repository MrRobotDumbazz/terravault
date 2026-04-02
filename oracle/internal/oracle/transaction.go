package oracle

import (
	"context"
	"fmt"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

// TransactionSender handles building and sending Solana transactions.
type TransactionSender struct {
	rpcURL    string
	wsURL     string
	signer    *Signer
	programID solana.PublicKey
	client    *rpc.Client
}

// NewTransactionSender creates a new TransactionSender.
func NewTransactionSender(rpcURL string, signer *Signer, programID string) *TransactionSender {
	client := rpc.New(rpcURL)
	pid := solana.MustPublicKeyFromBase58(programID)
	return &TransactionSender{
		rpcURL:    rpcURL,
		signer:    signer,
		programID: pid,
		client:    client,
	}
}

// SendAndConfirm builds, signs and confirms a transaction.
// instructions: list of pre-built solana.Instruction
func (ts *TransactionSender) SendAndConfirm(
	ctx context.Context,
	instructions []solana.Instruction,
) (solana.Signature, error) {
	// Get recent blockhash
	recent, err := ts.client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("getting recent blockhash: %w", err)
	}

	tx, err := solana.NewTransaction(
		instructions,
		recent.Value.Blockhash,
		solana.TransactionPayer(ts.signer.PublicKey()),
	)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("building transaction: %w", err)
	}

	// Sign transaction
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(ts.signer.PublicKey()) {
			pk := ts.signer.account
			return &pk
		}
		return nil
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("signing transaction: %w", err)
	}

	// Use websocket for confirmation if wsURL is set
	wsURL := wsURLFromRPC(ts.rpcURL)
	wsClient, err := ws.Connect(ctx, wsURL)
	if err != nil {
		// Fall back to polling confirmation
		return ts.sendWithPolling(ctx, tx)
	}
	defer wsClient.Close()

	sig, err := confirm.SendAndConfirmTransaction(ctx, ts.client, wsClient, tx)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("send and confirm: %w", err)
	}
	return sig, nil
}

// sendWithPolling submits the transaction and polls for confirmation.
func (ts *TransactionSender) sendWithPolling(
	ctx context.Context,
	tx *solana.Transaction,
) (solana.Signature, error) {
	sig, err := ts.client.SendTransaction(ctx, tx)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("sending transaction: %w", err)
	}

	// Poll for confirmation up to 60 seconds
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return sig, ctx.Err()
		default:
		}

		resp, err := ts.client.GetSignatureStatuses(ctx, false, sig)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		if resp != nil && len(resp.Value) > 0 && resp.Value[0] != nil {
			status := resp.Value[0]
			if status.Err != nil {
				return sig, fmt.Errorf("transaction failed: %v", status.Err)
			}
			if status.ConfirmationStatus == rpc.ConfirmationStatusFinalized ||
				status.ConfirmationStatus == rpc.ConfirmationStatusConfirmed {
				return sig, nil
			}
		}
		time.Sleep(2 * time.Second)
	}
	return sig, fmt.Errorf("transaction confirmation timeout")
}

// wsURLFromRPC converts a http/https RPC URL to a ws/wss WS URL.
func wsURLFromRPC(rpcURL string) string {
	if len(rpcURL) >= 5 && rpcURL[:5] == "https" {
		return "wss" + rpcURL[5:]
	}
	if len(rpcURL) >= 4 && rpcURL[:4] == "http" {
		return "ws" + rpcURL[4:]
	}
	return rpcURL
}

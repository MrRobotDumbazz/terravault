package listener

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/terravault/oracle/internal/oracle"
	"github.com/terravault/oracle/internal/storage"
)

// WSConfig configures the WebSocket listener.
type WSConfig struct {
	WSURL     string
	ProgramID string
	DB        *storage.DB
	Logger    *zap.Logger
	Handlers  Handlers
}

// Handlers holds callbacks for program events.
type Handlers struct {
	OnMilestoneSubmitted func(ctx context.Context, e oracle.MilestoneSubmittedEvent) error
	OnFundraisingStarted func(ctx context.Context, e oracle.FundraisingStartedEvent) error
	OnProjectActivated   func(ctx context.Context, e oracle.ProjectActivatedEvent) error
}

// WebSocketListener subscribes to Solana logs for the TerraVault program.
type WebSocketListener struct {
	cfg WSConfig
}

// NewWebSocketListener creates a new WebSocketListener.
func NewWebSocketListener(cfg WSConfig) *WebSocketListener {
	return &WebSocketListener{cfg: cfg}
}

// Listen connects to the Solana WebSocket endpoint and streams log events.
func (l *WebSocketListener) Listen(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := l.connectAndSubscribe(ctx); err != nil {
			l.cfg.Logger.Error("websocket error, reconnecting in 5s", zap.Error(err))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
			}
		}
	}
}

func (l *WebSocketListener) connectAndSubscribe(ctx context.Context) error {
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, l.cfg.WSURL, nil)
	if err != nil {
		return fmt.Errorf("dialing websocket %s: %w", l.cfg.WSURL, err)
	}
	defer conn.Close()

	l.cfg.Logger.Info("websocket connected", zap.String("url", l.cfg.WSURL))

	// Subscribe to program logs
	subscribeMsg := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "logsSubscribe",
		"params": []interface{}{
			map[string]string{"mentions": l.cfg.ProgramID},
			map[string]string{"commitment": "confirmed"},
		},
	}
	if err := conn.WriteJSON(subscribeMsg); err != nil {
		return fmt.Errorf("writing subscribe message: %w", err)
	}

	parser := NewParser(l.cfg.ProgramID)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		_, msg, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("reading websocket message: %w", err)
		}

		var notification LogNotification
		if err := json.Unmarshal(msg, &notification); err != nil {
			continue
		}
		if notification.Method != "logsNotification" {
			continue
		}

		events := parser.ParseLogs(notification.Params.Result.Value.Logs)
		for _, event := range events {
			if err := l.dispatchEvent(ctx, event); err != nil {
				l.cfg.Logger.Error("dispatching event", zap.Error(err), zap.Any("event", event))
			}
		}
	}
}

func (l *WebSocketListener) dispatchEvent(ctx context.Context, event ParsedEvent) error {
	switch e := event.(type) {
	case oracle.MilestoneSubmittedEvent:
		if l.cfg.Handlers.OnMilestoneSubmitted != nil {
			return l.cfg.Handlers.OnMilestoneSubmitted(ctx, e)
		}
	case oracle.FundraisingStartedEvent:
		if l.cfg.Handlers.OnFundraisingStarted != nil {
			return l.cfg.Handlers.OnFundraisingStarted(ctx, e)
		}
	case oracle.ProjectActivatedEvent:
		if l.cfg.Handlers.OnProjectActivated != nil {
			return l.cfg.Handlers.OnProjectActivated(ctx, e)
		}
	}
	return nil
}

// LogNotification is the JSON-RPC notification from logsSubscribe.
type LogNotification struct {
	Method string `json:"method"`
	Params struct {
		Result struct {
			Context struct {
				Slot uint64 `json:"slot"`
			} `json:"context"`
			Value struct {
				Signature string   `json:"signature"`
				Err       interface{} `json:"err"`
				Logs      []string `json:"logs"`
			} `json:"value"`
		} `json:"result"`
	} `json:"params"`
}

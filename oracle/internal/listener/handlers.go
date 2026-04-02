package listener

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/terravault/oracle/internal/oracle"
	"github.com/terravault/oracle/internal/storage"
)

// EventHandlers provides concrete implementations for all program event callbacks.
type EventHandlers struct {
	db      *storage.DB
	oracle  *oracle.MilestoneHandler
	logger  *zap.Logger
}

// NewEventHandlers creates a new EventHandlers.
func NewEventHandlers(db *storage.DB, oh *oracle.MilestoneHandler, logger *zap.Logger) *EventHandlers {
	return &EventHandlers{db: db, oracle: oh, logger: logger}
}

// OnMilestoneSubmitted handles a MilestoneProofSubmitted event.
func (h *EventHandlers) OnMilestoneSubmitted(ctx context.Context, e oracle.MilestoneSubmittedEvent) error {
	h.logger.Info("handling milestone submitted event",
		zap.String("project", e.ProjectPubkey),
		zap.Uint8("index", e.MilestoneIndex),
	)
	if err := h.oracle.HandleMilestoneSubmitted(ctx, e); err != nil {
		return fmt.Errorf("oracle.HandleMilestoneSubmitted: %w", err)
	}
	return nil
}

// OnFundraisingStarted handles a FundraisingStarted event.
func (h *EventHandlers) OnFundraisingStarted(ctx context.Context, e oracle.FundraisingStartedEvent) error {
	h.logger.Info("handling fundraising started event", zap.String("project", e.ProjectPubkey))
	if err := h.oracle.HandleFundraisingStarted(ctx, e); err != nil {
		return fmt.Errorf("oracle.HandleFundraisingStarted: %w", err)
	}
	return nil
}

// OnProjectActivated handles a ProjectActivated event.
func (h *EventHandlers) OnProjectActivated(ctx context.Context, e oracle.ProjectActivatedEvent) error {
	h.logger.Info("handling project activated event", zap.String("project", e.ProjectPubkey))
	if err := h.oracle.HandleProjectActivated(ctx, e); err != nil {
		return fmt.Errorf("oracle.HandleProjectActivated: %w", err)
	}
	return nil
}

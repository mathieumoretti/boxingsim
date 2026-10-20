//go:build integration

package store

import (
	"context"
	"testing"

	"github.com/mormm/boxing/internal/model"
)

// GetPendingByBoxerIDAndTypeDebug adds debug logging for integration tests
func (s *ScheduledEventStore) GetPendingByBoxerIDAndTypeDebug(ctx context.Context, boxerID int, eventType model.EventType, t *testing.T) ([]*model.ScheduledEvent, error) {
	events, err := s.GetPendingByBoxerIDAndType(ctx, boxerID, eventType)
	if err != nil {
		return nil, err
	}

	for i, event := range events {
		t.Logf("DEBUG Event %d: ID=%d, DataLength=%d", i, event.ID, len(event.EventData))
		t.Logf("DEBUG Event %d: Raw bytes=[%v]", i, []byte(event.EventData))
		t.Logf("DEBUG Event %d: String repr=%q", i, string(event.EventData))
	}

	return events, err
}

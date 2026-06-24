package model

import "time"

// WorldTick represents a world game tick
type WorldTick struct {
	ID          int
	TickNumber  int
	StartTime   time.Time
	EndTime     *time.Time
	ProcessedAt *time.Time
}

func (w *WorldTick) Number() int {
	return w.TickNumber
}
package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/platform/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBoxerStore extends the real store with mocking capabilities for testing
type MockBoxerStore struct {
	mock.Mock
}

func (m *MockBoxerStore) GetByID(ctx context.Context, id int) (*model.Boxer, error) {
	args := m.Called(ctx, id)
	if args.Get(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Boxer), nil
}

func (m *MockBoxerStore) Update(ctx context.Context, boxer *model.Boxer) error {
	args := m.Called(ctx, boxer)
	return args.Error(0)
}

// TestFatigueService_CheckCanTrain validates exhaustion and forced rest blocking
func TestFatigueService_CheckCanTrain(t *testing.T) {
	mockStore := &MockBoxerStore{}
	lg := logger.New("TEST")
	fatigueService := NewFatigueService(mockStore, lg)

	t.Run("allows training when fatigue is below exhaustion threshold", func(t *testing.T) {
		boxer := &model.Boxer{
			ID:              1,
			Name:            "Test Boxer",
			FatigueScore:    79.0,
			ForcedRestUntil: nil,
		}

		canTrain, errMsg := fatigueService.CheckCanTrain(boxer)

		assert.True(t, canTrain)
		assert.Empty(t, errMsg)
	})

	t.Run("blocks training when fatigue equals exhaustion threshold", func(t *testing.T) {
		boxer := &model.Boxer{
			ID:              1,
			Name:            "Test Boxer",
			FatigueScore:    ExhaustionThreshold,
			ForcedRestUntil: nil,
		}

		canTrain, errMsg := fatigueService.CheckCanTrain(boxer)

		assert.False(t, canTrain)
		assert.Contains(t, errMsg, "Exhausted")
	})

	t.Run("blocks training when fatigue exceeds exhaustion threshold", func(t *testing.T) {
		boxer := &model.Boxer{
			ID:              1,
			Name:            "Test Boxer",
			FatigueScore:    95.0,
			ForcedRestUntil: nil,
		}

		canTrain, errMsg := fatigueService.CheckCanTrain(boxer)

		assert.False(t, canTrain)
		assert.Contains(t, errMsg, "Exhausted")
	})

	t.Run("blocks training when forced rest is active (future timestamp)", func(t *testing.T) {
		futureTime := time.Now().Add(48 * time.Hour)
		boxer := &model.Boxer{
			ID:              1,
			Name:            "Test Boxer",
			FatigueScore:    50.0,
			ForcedRestUntil: &futureTime,
		}

		canTrain, errMsg := fatigueService.CheckCanTrain(boxer)

		assert.False(t, canTrain)
		assert.Contains(t, errMsg, "Mandatory rest")
	})

	t.Run("allows training when forced rest has expired (past timestamp)", func(t *testing.T) {
		pastTime := time.Now().Add(-24 * time.Hour)
		boxer := &model.Boxer{
			ID:              1,
			Name:            "Test Boxer",
			FatigueScore:    50.0,
			ForcedRestUntil: &pastTime,
		}

		canTrain, errMsg := fatigueService.CheckCanTrain(boxer)

		assert.True(t, canTrain)
		assert.Empty(t, errMsg)
	})

	t.Run("allows training with zero fatigue and no forced rest", func(t *testing.T) {
		boxer := &model.Boxer{
			ID:              1,
			Name:            "Fresh Boxer",
			FatigueScore:    0.0,
			ForcedRestUntil: nil,
		}

		canTrain, errMsg := fatigueService.CheckCanTrain(boxer)

		assert.True(t, canTrain)
		assert.Empty(t, errMsg)
	})
}

// TestFatigueService_CalculateFatigueIncrease validates fatigue calculation from training duration
func TestFatigueService_CalculateFatigueIncrease(t *testing.T) {
	mockStore := &MockBoxerStore{}
	lg := logger.New("TEST")
	fatigueService := NewFatigueService(mockStore, lg)

	t.Run("calculates fatigue for 1 hour of training", func(t *testing.T) {
		fatigue := fatigueService.CalculateFatigueIncrease(1.0)
		assert.Equal(t, 15.0, fatigue, "1 hour should add 15 fatigue points")
	})

	t.Run("calculates fatigue for 2 hours of training", func(t *testing.T) {
		fatigue := fatigueService.CalculateFatigueIncrease(2.0)
		assert.Equal(t, 30.0, fatigue, "2 hours should add 30 fatigue points")
	})

	t.Run("calculates fatigue for 4 hours of training", func(t *testing.T) {
		fatigue := fatigueService.CalculateFatigueIncrease(4.0)
		assert.Equal(t, 60.0, fatigue, "4 hours should add 60 fatigue points")
	})

	t.Run("calculates fatigue for 0.5 hours of training", func(t *testing.T) {
		fatigue := fatigueService.CalculateFatigueIncrease(0.5)
		assert.Equal(t, 7.5, fatigue, "0.5 hours should add 7.5 fatigue points")
	})

	t.Run("calculates fatigue for 8 hours max training", func(t *testing.T) {
		fatigue := fatigueService.CalculateFatigueIncrease(8.0)
		assert.Equal(t, 120.0, fatigue, "8 hours should add 120 fatigue points (will cap at 100)")
	})

	t.Run("returns zero for zero duration", func(t *testing.T) {
		fatigue := fatigueService.CalculateFatigueIncrease(0.0)
		assert.Equal(t, 0.0, fatigue, "0 hours should add 0 fatigue points")
	})
}

// TestFatigueService_GetRecoveryBenefits validates recovery benefits calculation
func TestFatigueService_GetRecoveryBenefits(t *testing.T) {
	mockStore := &MockBoxerStore{}
	lg := logger.New("TEST")
	fatigueService := NewFatigueService(mockStore, lg)

	t.Run("returns 50% energy and 20 fatigue reduction for 1 day rest", func(t *testing.T) {
		benefits := fatigueService.GetRecoveryBenefits(1)
		assert.Equal(t, 50.0, benefits.EnergyRecoveryPercent)
		assert.Equal(t, 20.0, benefits.FatigueReduction)
		assert.Equal(t, 0.0, benefits.StatDecayRisk)
	})

	t.Run("returns 80% energy and 40 fatigue reduction for 2 days rest", func(t *testing.T) {
		benefits := fatigueService.GetRecoveryBenefits(2)
		assert.Equal(t, 80.0, benefits.EnergyRecoveryPercent)
		assert.Equal(t, 40.0, benefits.FatigueReduction)
		assert.Equal(t, 0.0, benefits.StatDecayRisk)
	})

	t.Run("returns 100% energy and 60 fatigue reduction for 3-6 days rest", func(t *testing.T) {
		benefits := fatigueService.GetRecoveryBenefits(5)
		assert.Equal(t, 100.0, benefits.EnergyRecoveryPercent)
		assert.Equal(t, 60.0, benefits.FatigueReduction)
		assert.Equal(t, 0.0, benefits.StatDecayRisk)
	})

	t.Run("returns full recovery with stat decay risk for 7+ days rest", func(t *testing.T) {
		benefits := fatigueService.GetRecoveryBenefits(10)
		assert.Equal(t, 100.0, benefits.EnergyRecoveryPercent)
		assert.Equal(t, 100.0, benefits.FatigueReduction)
		assert.Equal(t, 5.0, benefits.StatDecayRisk)
	})
}

// TestFatigueService_ApplyFatigueIncrease tests the database update for fatigue increase
func TestFatigueService_ApplyFatigueIncrease(t *testing.T) {
	mockStore := &MockBoxerStore{}
	lg := logger.New("TEST")
	fatigueService := NewFatigueService(mockStore, lg)
	ctx := context.Background()

	t.Run("successfully increases fatigue after training", func(t *testing.T) {
		existingBoxer := &model.Boxer{
			ID:           1,
			Name:         "Test Boxer",
			FatigueScore: 50.0,
		}

		mockStore.On("GetByID", ctx, 1).Return(existingBoxer, nil)
		// Update is called with mutated boxer (FatigueScore increased by 15)
		mockStore.On("Update", mock.Anything, mock.MatchedBy(func(b *model.Boxer) bool {
			return b.ID == 1 && b.FatigueScore == 65.0
		})).Return(nil)

		err := fatigueService.ApplyFatigueIncrease(ctx, 1, 15.0)

		assert.NoError(t, err)
		assert.Equal(t, 65.0, existingBoxer.FatigueScore, "Fatigue should increase from 50 to 65")
		mockStore.AssertExpectations(t)
	})

	t.Run("caps fatigue at MaxFatigueScore (100)", func(t *testing.T) {
		existingBoxer := &model.Boxer{
			ID:           2, // Different ID to avoid mock collision with first subtest
			Name:         "Exhausted Boxer",
			FatigueScore: 95.0,
		}

		mockStore.On("GetByID", ctx, 2).Return(existingBoxer, nil)
		mockStore.On("Update", mock.Anything, mock.MatchedBy(func(b *model.Boxer) bool {
			return b.ID == 2 && b.FatigueScore == 100.0
		})).Return(nil)

		err := fatigueService.ApplyFatigueIncrease(ctx, 2, 20.0)

		assert.NoError(t, err)
		assert.Equal(t, 100.0, existingBoxer.FatigueScore, "Fatigue should cap at 100")
		mockStore.AssertExpectations(t)
	})

	t.Run("returns error when boxer not found", func(t *testing.T) {
		mockStore.On("GetByID", ctx, 999).Return(nil, sql.ErrNoRows)

		err := fatigueService.ApplyFatigueIncrease(ctx, 999, 15.0)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "boxer not found")
		mockStore.AssertExpectations(t)
	})

	t.Run("returns error when database update fails", func(t *testing.T) {
		existingBoxer := &model.Boxer{
			ID:           3, // Different ID to avoid mock collision
			Name:         "Test Boxer",
			FatigueScore: 50.0,
		}

		dbError := errors.New("database connection lost")
		mockStore.On("GetByID", ctx, 3).Return(existingBoxer, nil)
		mockStore.On("Update", mock.Anything, mock.MatchedBy(func(b *model.Boxer) bool {
			return b.ID == 3 && b.FatigueScore == 65.0
		})).Return(dbError)

		err := fatigueService.ApplyFatigueIncrease(ctx, 3, 15.0)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update fatigue")
		mockStore.AssertExpectations(t)
	})
}

// TestFatigueService_ReduceFatigue tests the database update for fatigue reduction
func TestFatigueService_ReduceFatigue(t *testing.T) {
	mockStore := &MockBoxerStore{}
	lg := logger.New("TEST")
	fatigueService := NewFatigueService(mockStore, lg)
	ctx := context.Background()

	t.Run("successfully reduces fatigue after rest", func(t *testing.T) {
		existingBoxer := &model.Boxer{
			ID:           10,
			Name:         "Test Boxer",
			FatigueScore: 50.0,
		}

		mockStore.On("GetByID", ctx, 10).Return(existingBoxer, nil)
		mockStore.On("Update", mock.Anything, mock.MatchedBy(func(b *model.Boxer) bool {
			return b.ID == 10 && b.FatigueScore == 35.0
		})).Return(nil)

		err := fatigueService.ReduceFatigue(ctx, 10, 15.0)

		assert.NoError(t, err)
		assert.Equal(t, 35.0, existingBoxer.FatigueScore, "Fatigue should reduce from 50 to 35")
		mockStore.AssertExpectations(t)
	})

	t.Run("floors fatigue at zero (negative prevented)", func(t *testing.T) {
		existingBoxer := &model.Boxer{
			ID:           11,
			Name:         "Fresh Boxer",
			FatigueScore: 5.0,
		}

		mockStore.On("GetByID", ctx, 11).Return(existingBoxer, nil)
		mockStore.On("Update", mock.Anything, mock.MatchedBy(func(b *model.Boxer) bool {
			return b.ID == 11 && b.FatigueScore == 0.0
		})).Return(nil)

		err := fatigueService.ReduceFatigue(ctx, 11, 35.0)

		assert.NoError(t, err)
		assert.Equal(t, 0.0, existingBoxer.FatigueScore, "Fatigue should floor at 0")
		mockStore.AssertExpectations(t)
	})

	t.Run("returns error when boxer not found", func(t *testing.T) {
		mockStore.On("GetByID", ctx, 999).Return(nil, sql.ErrNoRows)

		err := fatigueService.ReduceFatigue(ctx, 999, 10.0)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "boxer not found")
		mockStore.AssertExpectations(t)
	})

	t.Run("returns error when database update fails", func(t *testing.T) {
		existingBoxer := &model.Boxer{
			ID:           12,
			Name:         "Test Boxer",
			FatigueScore: 50.0,
		}

		dbError := errors.New("database lock timeout")
		mockStore.On("GetByID", ctx, 12).Return(existingBoxer, nil)
		mockStore.On("Update", mock.Anything, mock.MatchedBy(func(b *model.Boxer) bool {
			return b.ID == 12 && b.FatigueScore == 35.0
		})).Return(dbError)

		err := fatigueService.ReduceFatigue(ctx, 12, 15.0)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update fatigue")
		mockStore.AssertExpectations(t)
	})
}

// TestFatigueService_ScheduleForcedRest tests forced rest scheduling functionality
func TestFatigueService_ScheduleForcedRest(t *testing.T) {
	mockStore := &MockBoxerStore{}
	lg := logger.New("TEST")
	fatigueService := NewFatigueService(mockStore, lg)
	ctx := context.Background()

	t.Run("successfully schedules forced rest for 3 days", func(t *testing.T) {
		existingBoxer := &model.Boxer{
			ID:           20,
			Name:         "Exhausted Boxer",
			FatigueScore: 85.0,
		}

		mockStore.On("GetByID", ctx, 20).Return(existingBoxer, nil)
		mockStore.On("Update", mock.Anything, mock.MatchedBy(func(b *model.Boxer) bool {
			return b.ID == 20 && b.ForcedRestUntil != nil
		})).Return(nil)

		err := fatigueService.ScheduleForcedRest(ctx, 20, 3)

		assert.NoError(t, err)
		assert.NotNil(t, existingBoxer.ForcedRestUntil)
		expectedTime := time.Now().Add(3 * 24 * time.Hour)
		timeDiff := expectedTime.Sub(*existingBoxer.ForcedRestUntil).Abs()
		assert.Less(t, timeDiff, 2*time.Second, "Forced rest should be scheduled for ~3 days from now")
		mockStore.AssertExpectations(t)
	})

	t.Run("uses default duration when boxer is exhausted", func(t *testing.T) {
		existingBoxer := &model.Boxer{
			ID:           21,
			Name:         "Exhausted Boxer",
			FatigueScore: 90.0,
		}

		mockStore.On("GetByID", ctx, 21).Return(existingBoxer, nil)
		mockStore.On("Update", mock.Anything, mock.MatchedBy(func(b *model.Boxer) bool {
			return b.ID == 21 && b.ForcedRestUntil != nil
		})).Return(nil)

		err := fatigueService.ScheduleForcedRest(ctx, 21, DefaultForcedRestDays)

		assert.NoError(t, err)
		assert.NotNil(t, existingBoxer.ForcedRestUntil)
		mockStore.AssertExpectations(t)
	})

	t.Run("returns error when boxer not found", func(t *testing.T) {
		mockStore.On("GetByID", ctx, 998).Return(nil, sql.ErrNoRows)

		err := fatigueService.ScheduleForcedRest(ctx, 998, 3)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "boxer not found")
		mockStore.AssertExpectations(t)
	})

	t.Run("returns error when database update fails", func(t *testing.T) {
		existingBoxer := &model.Boxer{
			ID:           22,
			Name:         "Test Boxer",
			FatigueScore: 85.0,
		}

		dbError := errors.New("database constraint violation")
		mockStore.On("GetByID", ctx, 22).Return(existingBoxer, nil)
		mockStore.On("Update", mock.Anything, mock.MatchedBy(func(b *model.Boxer) bool {
			return b.ID == 22 && b.ForcedRestUntil != nil
		})).Return(dbError)

		err := fatigueService.ScheduleForcedRest(ctx, 22, 3)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to schedule forced rest")
		mockStore.AssertExpectations(t)
	})
}

// TestFatigueService_ClearForcedRest tests clearing forced rest restriction
func TestFatigueService_ClearForcedRest(t *testing.T) {
	mockStore := &MockBoxerStore{}
	lg := logger.New("TEST")
	fatigueService := NewFatigueService(mockStore, lg)
	ctx := context.Background()

	t.Run("successfully clears forced rest", func(t *testing.T) {
		futureTime := time.Now().Add(48 * time.Hour)
		existingBoxer := &model.Boxer{
			ID:              30,
			Name:            "Test Boxer",
			FatigueScore:    50.0,
			ForcedRestUntil: &futureTime,
		}

		mockStore.On("GetByID", ctx, 30).Return(existingBoxer, nil)
		mockStore.On("Update", mock.Anything, mock.MatchedBy(func(b *model.Boxer) bool {
			return b.ID == 30 && b.ForcedRestUntil == nil
		})).Return(nil)

		err := fatigueService.ClearForcedRest(ctx, 30)

		assert.NoError(t, err)
		assert.Nil(t, existingBoxer.ForcedRestUntil, "Forced rest should be cleared")
		mockStore.AssertExpectations(t)
	})

	t.Run("returns nil when no forced rest is active", func(t *testing.T) {
		existingBoxer := &model.Boxer{
			ID:              31,
			Name:            "Test Boxer",
			FatigueScore:    50.0,
			ForcedRestUntil: nil,
		}

		mockStore.On("GetByID", ctx, 31).Return(existingBoxer, nil)

		err := fatigueService.ClearForcedRest(ctx, 31)

		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})
}

// TestFatigueService_CheckExhaustionThreshold tests exhaustion check
func TestFatigueService_CheckExhaustionThreshold(t *testing.T) {
	mockStore := &MockBoxerStore{}
	lg := logger.New("TEST")
	fatigueService := NewFatigueService(mockStore, lg)

	t.Run("returns true when fatigue equals exhaustion threshold", func(t *testing.T) {
		boxer := &model.Boxer{FatigueScore: ExhaustionThreshold}
		assert.True(t, fatigueService.CheckExhaustionThreshold(boxer))
	})

	t.Run("returns true when fatigue exceeds exhaustion threshold", func(t *testing.T) {
		boxer := &model.Boxer{FatigueScore: 90.0}
		assert.True(t, fatigueService.CheckExhaustionThreshold(boxer))
	})

	t.Run("returns false when fatigue is below exhaustion threshold", func(t *testing.T) {
		boxer := &model.Boxer{FatigueScore: 75.0}
		assert.False(t, fatigueService.CheckExhaustionThreshold(boxer))
	})
}

// TestFatigueService_IsOnForcedRest tests forced rest status check
func TestFatigueService_IsOnForcedRest(t *testing.T) {
	mockStore := &MockBoxerStore{}
	lg := logger.New("TEST")
	fatigueService := NewFatigueService(mockStore, lg)

	t.Run("returns false when no forced rest is set", func(t *testing.T) {
		boxer := &model.Boxer{ForcedRestUntil: nil}
		assert.False(t, fatigueService.IsOnForcedRest(boxer))
	})

	t.Run("returns true when forced rest is active (future)", func(t *testing.T) {
		futureTime := time.Now().Add(48 * time.Hour)
		boxer := &model.Boxer{ForcedRestUntil: &futureTime}
		assert.True(t, fatigueService.IsOnForcedRest(boxer))
	})

	t.Run("returns false when forced rest has expired (past)", func(t *testing.T) {
		pastTime := time.Now().Add(-24 * time.Hour)
		boxer := &model.Boxer{ForcedRestUntil: &pastTime}
		assert.False(t, fatigueService.IsOnForcedRest(boxer))
	})
}

// TestFatigueService_Integration tests end-to-end fatigue workflow
func TestFatigueService_Integration(t *testing.T) {
	mockStore := &MockBoxerStore{}
	lg := logger.New("TEST")
	fatigueService := NewFatigueService(mockStore, lg)
	ctx := context.Background()

	t.Run("boxer becomes exhausted after multiple training sessions", func(t *testing.T) {
		boxer := &model.Boxer{
			ID:           40,
			Name:         "Rising Star",
			FatigueScore: 50.0,
		}

		var fatigueHistory []float64

		for i := 0; i < 3; i++ {
			mockStore.On("GetByID", ctx, 40).Return(boxer, nil)
			mockStore.On("Update", mock.Anything, mock.MatchedBy(func(b *model.Boxer) bool {
				return b.ID == 40 && b.FatigueScore <= 100.0
			})).Run(func(args mock.Arguments) {
				fatigueHistory = append(fatigueHistory, boxer.FatigueScore)
			}).Return(nil)

			err := fatigueService.ApplyFatigueIncrease(ctx, 40, 30.0)
			assert.NoError(t, err)
		}

		mockStore.AssertExpectations(t)

		assert.Len(t, fatigueHistory, 3)
		assert.Equal(t, 80.0, fatigueHistory[0], "First session: 50 + 30 = 80")
		assert.Equal(t, 100.0, fatigueHistory[1], "Second session: 80 + 30 = 110, capped at 100")
		assert.Equal(t, 100.0, fatigueHistory[2], "Third session: stays at 100")

		canTrain, errMsg := fatigueService.CheckCanTrain(boxer)
		assert.False(t, canTrain)
		assert.Contains(t, errMsg, "Exhausted")
	})

	t.Run("fatigue reduction floors at zero", func(t *testing.T) {
		boxer := &model.Boxer{
			ID:           41,
			Name:         "Veteran",
			FatigueScore: 25.0,
		}

		mockStore.On("GetByID", ctx, 41).Return(boxer, nil)
		mockStore.On("Update", mock.Anything, mock.MatchedBy(func(b *model.Boxer) bool {
			return b.ID == 41 && b.FatigueScore == 0.0
		})).Return(nil)

		err := fatigueService.ReduceFatigue(ctx, 41, 50.0)
		assert.NoError(t, err)
		assert.Equal(t, 0.0, boxer.FatigueScore, "Fatigue should floor at 0")
		mockStore.AssertExpectations(t)
	})
}

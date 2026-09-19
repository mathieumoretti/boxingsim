package boxer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mormm/boxing/internal/model"
)

// MockBoxerRepository implements the BoxerRepository interface for testing
type MockBoxerRepository struct {
	mock.Mock
}

func (m *MockBoxerRepository) Create(ctx context.Context, boxer *model.Boxer) error {
	args := m.Called(ctx, boxer)
	return args.Error(0)
}

func (m *MockBoxerRepository) GetByID(ctx context.Context, id int) (*model.Boxer, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Boxer), args.Error(1)
}

func (m *MockBoxerRepository) GetByUserID(ctx context.Context, userID int) ([]*model.Boxer, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*model.Boxer), args.Error(1)
}

func (m *MockBoxerRepository) Update(ctx context.Context, boxer *model.Boxer) error {
	args := m.Called(ctx, boxer)
	return args.Error(0)
}

func (m *MockBoxerRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBoxerRepository) UpdateFightResult(ctx context.Context, boxer1ID, boxer2ID int, result FightResult) error {
	args := m.Called(ctx, boxer1ID, boxer2ID, result)
	return args.Error(0)
}

func TestNewBoxerService(t *testing.T) {
	t.Run("Creates service with repository", func(t *testing.T) {
		mockRepo := new(MockBoxerRepository)
		service := NewBoxerService(mockRepo)

		assert.NotNil(t, service)
		assert.Equal(t, mockRepo, service.repo)
	})
}

func TestBoxerServiceCreateBoxer(t *testing.T) {
	t.Run("Successfully creates boxer", func(t *testing.T) {
		mockRepo := new(MockBoxerRepository)
		service := NewBoxerService(mockRepo)

		createReq := &model.BoxerCreate{
			Name:      "Test Boxer",
			Nickname:  stringPtr("TB"),
			PositionX: 0,
			PositionY: 0,
			Strength:  10,
			Defense:   10,
			Agility:   10,
		}

		// Create expected boxer with time.Now() to capture the exact moment
		expectedBoxer := &model.Boxer{
			ID:         0, // ID is set by the repository after creation
			UserID:     1,
			Name:       "Test Boxer",
			Nickname:   stringPtr("TB"),
			PositionX:  0,
			PositionY:  0,
			Health:     100.0,
			Energy:     100.0,
			Strength:   10,
			Defense:    10,
			Agility:    10,
			Experience: 0.0,
			Level:      1,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		// Use mock.Anything for the boxer parameter to avoid strict time matching
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

		result, err := service.CreateBoxer(context.Background(), 1, createReq)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedBoxer.Name, result.Name)
		assert.Equal(t, expectedBoxer.UserID, result.UserID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Returns error when repository fails", func(t *testing.T) {
		mockRepo := new(MockBoxerRepository)
		service := NewBoxerService(mockRepo)

		createReq := &model.BoxerCreate{
			Name:      "Test Boxer",
			Nickname:  stringPtr("TB"),
			PositionX: 0,
			PositionY: 0,
			Strength:  10,
			Defense:   10,
			Agility:   10,
		}

		expectedError := errors.New("database error")
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(expectedError)

		result, err := service.CreateBoxer(context.Background(), 1, createReq)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)

		mockRepo.AssertExpectations(t)
	})
}

func TestBoxerServiceGetBoxer(t *testing.T) {
	t.Run("Successfully retrieves boxer", func(t *testing.T) {
		mockRepo := new(MockBoxerRepository)
		service := NewBoxerService(mockRepo)

		expectedBoxer := &model.Boxer{
			ID:         1,
			UserID:     1,
			Name:       "Test Boxer",
			Nickname:   stringPtr("TB"),
			PositionX:  0,
			PositionY:  0,
			Health:     100.0,
			Energy:     100.0,
			Strength:   10,
			Defense:    10,
			Agility:    10,
			Experience: 0.0,
			Level:      1,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		mockRepo.On("GetByID", mock.Anything, 1).Return(expectedBoxer, nil)

		result, err := service.GetBoxer(context.Background(), 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedBoxer.ID, result.ID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Returns error when repository fails", func(t *testing.T) {
		mockRepo := new(MockBoxerRepository)
		service := NewBoxerService(mockRepo)

		expectedError := errors.New("not found")
		mockRepo.On("GetByID", mock.Anything, 1).Return((*model.Boxer)(nil), expectedError)

		result, err := service.GetBoxer(context.Background(), 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)

		mockRepo.AssertExpectations(t)
	})
}

func TestBoxerServiceGetBoxersByUser(t *testing.T) {
	t.Run("Successfully retrieves boxers by user", func(t *testing.T) {
		mockRepo := new(MockBoxerRepository)
		service := NewBoxerService(mockRepo)

		expectedBoxers := []*model.Boxer{
			{
				ID:         1,
				UserID:     1,
				Name:       "Test Boxer 1",
				Nickname:   stringPtr("TB1"),
				PositionX:  0,
				PositionY:  0,
				Health:     100.0,
				Energy:     100.0,
				Strength:   10,
				Defense:    10,
				Agility:    10,
				Experience: 0.0,
				Level:      1,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			{
				ID:         2,
				UserID:     1,
				Name:       "Test Boxer 2",
				Nickname:   stringPtr("TB2"),
				PositionX:  0,
				PositionY:  0,
				Health:     100.0,
				Energy:     100.0,
				Strength:   10,
				Defense:    10,
				Agility:    10,
				Experience: 0.0,
				Level:      1,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		}

		mockRepo.On("GetByUserID", mock.Anything, 1).Return(expectedBoxers, nil)

		result, err := service.GetBoxersByUser(context.Background(), 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 2)
		assert.Equal(t, expectedBoxers[0].ID, result[0].ID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Returns error when repository fails", func(t *testing.T) {
		mockRepo := new(MockBoxerRepository)
		service := NewBoxerService(mockRepo)

		expectedError := errors.New("database error")
		mockRepo.On("GetByUserID", mock.Anything, 1).Return([]*model.Boxer(nil), expectedError)

		result, err := service.GetBoxersByUser(context.Background(), 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)

		mockRepo.AssertExpectations(t)
	})
}

func TestBoxerServiceUpdateBoxer(t *testing.T) {
	t.Run("Successfully updates boxer", func(t *testing.T) {
		mockRepo := new(MockBoxerRepository)
		service := NewBoxerService(mockRepo)

		// First get the existing boxer
		existingBoxer := &model.Boxer{
			ID:         1,
			UserID:     1,
			Name:       "Old Name",
			Nickname:   stringPtr("ON"),
			PositionX:  0,
			PositionY:  0,
			Health:     100.0,
			Energy:     100.0,
			Strength:   10,
			Defense:    10,
			Agility:    10,
			Experience: 0.0,
			Level:      1,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		updateReq := &model.BoxerUpdate{
			Name:      stringPtr("New Name"),
			Nickname:  stringPtr("NN"),
			PositionX: float64Ptr(5),
			PositionY: float64Ptr(5),
			Strength:  float64Ptr(15),
			Defense:   float64Ptr(15),
			Agility:   float64Ptr(15),
		}

		// Instead of using exact time expectations, let's just mock with anything
		mockRepo.On("GetByID", mock.Anything, 1).Return(existingBoxer, nil)
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

		result, err := service.UpdateBoxer(context.Background(), 1, updateReq)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "New Name", result.Name)
		assert.Equal(t, "NN", *result.Nickname)
		assert.Equal(t, 5.0, result.PositionX)
		assert.Equal(t, 5.0, result.PositionY)
		assert.Equal(t, 15.0, result.Strength)
		assert.Equal(t, 15.0, result.Defense)
		assert.Equal(t, 15.0, result.Agility)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Returns error when get fails", func(t *testing.T) {
		mockRepo := new(MockBoxerRepository)
		service := NewBoxerService(mockRepo)

		expectedError := errors.New("not found")
		mockRepo.On("GetByID", mock.Anything, 1).Return((*model.Boxer)(nil), expectedError)

		result, err := service.UpdateBoxer(context.Background(), 1, &model.BoxerUpdate{})

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Returns error when update fails", func(t *testing.T) {
		mockRepo := new(MockBoxerRepository)
		service := NewBoxerService(mockRepo)

		existingBoxer := &model.Boxer{
			ID:         1,
			UserID:     1,
			Name:       "Test Boxer",
			Nickname:   stringPtr("TB"),
			PositionX:  0,
			PositionY:  0,
			Health:     100.0,
			Energy:     100.0,
			Strength:   10,
			Defense:    10,
			Agility:    10,
			Experience: 0.0,
			Level:      1,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		updateReq := &model.BoxerUpdate{
			Name: stringPtr("New Name"),
		}

		mockRepo.On("GetByID", mock.Anything, 1).Return(existingBoxer, nil)
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update error"))

		result, err := service.UpdateBoxer(context.Background(), 1, updateReq)

		assert.Error(t, err)
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
	})
}

// Helper functions to create pointers for tests
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}

func TestBoxerServiceUpdateFightResult(t *testing.T) {
	t.Run("Successfully updates stats for boxer 1 win by KO", func(t *testing.T) {
		mockRepo := new(MockBoxerRepository)
		service := NewBoxerService(mockRepo)

		result := FightResult{
			Boxer1Wins:        true,
			IsDraw:            false,
			IsKnockout:        true,
			Boxer1Knockdowned: false,
			Boxer2Knockdowned: true,
		}

		mockRepo.On("UpdateFightResult", mock.Anything, 1, 2, result).Return(nil)

		err := service.UpdateFightResult(context.Background(), 1, 2, result)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Successfully updates stats for boxer 2 win by decision", func(t *testing.T) {
		mockRepo := new(MockBoxerRepository)
		service := NewBoxerService(mockRepo)

		result := FightResult{
			Boxer1Wins:        false,
			IsDraw:            false,
			IsKnockout:        false,
			Boxer1Knockdowned: false,
			Boxer2Knockdowned: false,
		}

		mockRepo.On("UpdateFightResult", mock.Anything, 1, 2, result).Return(nil)

		err := service.UpdateFightResult(context.Background(), 1, 2, result)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Successfully updates stats for draw", func(t *testing.T) {
		mockRepo := new(MockBoxerRepository)
		service := NewBoxerService(mockRepo)

		result := FightResult{
			Boxer1Wins:        false,
			IsDraw:            true,
			IsKnockout:        false,
			Boxer1Knockdowned: false,
			Boxer2Knockdowned: false,
		}

		mockRepo.On("UpdateFightResult", mock.Anything, 1, 2, result).Return(nil)

		err := service.UpdateFightResult(context.Background(), 1, 2, result)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Returns error when repository fails", func(t *testing.T) {
		mockRepo := new(MockBoxerRepository)
		service := NewBoxerService(mockRepo)

		result := FightResult{
			Boxer1Wins: true,
			IsDraw:     false,
			IsKnockout: false,
		}

		expectedError := errors.New("database error")
		mockRepo.On("UpdateFightResult", mock.Anything, 1, 2, result).Return(expectedError)

		err := service.UpdateFightResult(context.Background(), 1, 2, result)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
	})
}

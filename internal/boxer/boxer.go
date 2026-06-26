package boxer

import (
	"context"
	"time"

	"github.com/mormm/boxing/internal/model"
)

// BoxerRepository defines the interface for boxer operations
type BoxerRepository interface {
	Create(ctx context.Context, boxer *model.Boxer) error
	GetByID(ctx context.Context, id int) (*model.Boxer, error)
	GetByUserID(ctx context.Context, userID int) ([]*model.Boxer, error)
	Update(ctx context.Context, boxer *model.Boxer) error
	Delete(ctx context.Context, id int) error
}

// BoxerService handles boxer business logic
type BoxerService struct {
	repo BoxerRepository
}

func NewBoxerService(repo BoxerRepository) *BoxerService {
	return &BoxerService{
		repo: repo,
	}
}

// CreateBoxer creates a new boxer for a user
func (s *BoxerService) CreateBoxer(ctx context.Context, userID int, createReq *model.BoxerCreate) (*model.Boxer, error) {
	boxer := &model.Boxer{
		UserID:     userID,
		Name:       createReq.Name,
		Nickname:   createReq.Nickname,
		PositionX:  createReq.PositionX,
		PositionY:  createReq.PositionY,
		Health:     100.0,
		Energy:     100.0,
		Strength:   createReq.Strength,
		Defense:    createReq.Defense,
		Agility:    createReq.Agility,
		Experience: 0.0,
		Level:      1,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := s.repo.Create(ctx, boxer)
	if err != nil {
		return nil, err
	}

	return boxer, nil
}

// GetBoxer gets a boxer by ID
func (s *BoxerService) GetBoxer(ctx context.Context, id int) (*model.Boxer, error) {
	return s.repo.GetByID(ctx, id)
}

// GetBoxersByUser gets all boxers for a user
func (s *BoxerService) GetBoxersByUser(ctx context.Context, userID int) ([]*model.Boxer, error) {
	return s.repo.GetByUserID(ctx, userID)
}

// UpdateBoxer updates a boxer's information
func (s *BoxerService) UpdateBoxer(ctx context.Context, id int, updateReq *model.BoxerUpdate) (*model.Boxer, error) {
	boxer, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if updateReq.Name != nil {
		boxer.Name = *updateReq.Name
	}
	if updateReq.Nickname != nil {
		boxer.Nickname = updateReq.Nickname
	}
	if updateReq.PositionX != nil {
		boxer.PositionX = *updateReq.PositionX
	}
	if updateReq.PositionY != nil {
		boxer.PositionY = *updateReq.PositionY
	}
	if updateReq.Strength != nil {
		boxer.Strength = *updateReq.Strength
	}
	if updateReq.Defense != nil {
		boxer.Defense = *updateReq.Defense
	}
	if updateReq.Agility != nil {
		boxer.Agility = *updateReq.Agility
	}

	boxer.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, boxer)
	if err != nil {
		return nil, err
	}

	return boxer, nil
}
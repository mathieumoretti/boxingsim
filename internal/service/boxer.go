package service

import (
	"context"
	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/boxer"
)

// BoxerService handles boxer-related business logic
type BoxerService struct {
	boxerService *boxer.BoxerService
}

// NewBoxerService creates a new BoxerService
func NewBoxerService(boxerService *boxer.BoxerService) *BoxerService {
	return &BoxerService{boxerService: boxerService}
}

// CreateBoxer creates a new boxer for a user
func (s *BoxerService) CreateBoxer(ctx context.Context, userID int, createReq *model.BoxerCreate) (*model.Boxer, error) {
	return s.boxerService.CreateBoxer(ctx, userID, createReq)
}

// GetBoxer retrieves a boxer by ID
func (s *BoxerService) GetBoxer(ctx context.Context, id int) (*model.Boxer, error) {
	return s.boxerService.GetBoxer(ctx, id)
}

// UpdateBoxer updates a boxer's information
func (s *BoxerService) UpdateBoxer(ctx context.Context, id int, updateReq *model.BoxerUpdate) (*model.Boxer, error) {
	return s.boxerService.UpdateBoxer(ctx, id, updateReq)
}
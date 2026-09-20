package store

import (
	"context"
	"database/sql"

	"github.com/mormm/boxing/internal/model"
)

// RankingsStore handles ranking-related database queries
type RankingsStore struct {
	db *sql.DB
}

// NewRankingsStore creates a new RankingsStore
func NewRankingsStore(db *sql.DB) *RankingsStore {
	return &RankingsStore{db: db}
}

// getRankingQuery returns the SQL query for ranking by the specified criteria
func (s *RankingsStore) getRankingQuery(criteria model.RankingCriteria, limit int) (string, []any) {
	var orderBySQL string
	switch criteria {
	case model.RankingByWinRate:
		// win_rate = wins / (wins + losses), handle division by zero
		orderBySQL = `(CASE WHEN (wins + losses) > 0 THEN CAST(wins AS FLOAT) / (wins + losses) ELSE 0 END) DESC, wins DESC`
	case model.RankingByTotalFights:
		orderBySQL = `(wins + losses + draws) DESC, wins DESC`
	case model.RankingByLevel:
		orderBySQL = `level DESC, wins DESC`
	case model.RankingByStrength:
		orderBySQL = `strength DESC, wins DESC`
	case model.RankingByPowerScore:
		// power_score = (wins * 2) + knockouts + (level * 3)
		orderBySQL = `((wins * 2) + knockouts + (level * 3)) DESC, wins DESC`
	default:
		// Default to win rate
		orderBySQL = `(CASE WHEN (wins + losses) > 0 THEN CAST(wins AS FLOAT) / (wins + losses) ELSE 0 END) DESC, wins DESC`
	}

	query := `
		SELECT
			id, name, nickname, wins, losses, draws, knockouts, level, strength
		FROM boxers
		ORDER BY ` + orderBySQL + `
		LIMIT $1`

	return query, []any{limit}
}

// GetRankings retrieves boxers ranked by specified criteria with a limit
func (s *RankingsStore) GetRankings(ctx context.Context, criteria model.RankingCriteria, limit int) ([]*model.RankedBoxer, error) {
	if limit <= 0 {
		limit = 100 // Default limit
	}
	if limit > 1000 {
		limit = 1000 // Max limit for performance
	}

	query, args := s.getRankingQuery(criteria, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	boxers := make([]*model.RankedBoxer, 0)
	rank := 1

	for rows.Next() {
		boxer := &model.RankedBoxer{}
		err := rows.Scan(
			&boxer.ID, &boxer.Name, &boxer.Nickname,
			&boxer.Wins, &boxer.Losses, &boxer.Draws, &boxer.Knockouts,
			&boxer.Level, &boxer.Strength,
		)
		if err != nil {
			return nil, err
		}

		// Calculate ranking score based on criteria
		var score float64
		switch criteria {
		case model.RankingByWinRate:
			if boxer.Wins+boxer.Losses > 0 {
				score = float64(boxer.Wins) / float64(boxer.Wins+boxer.Losses)
			}
		case model.RankingByTotalFights:
			score = float64(boxer.Wins + boxer.Losses + boxer.Draws)
		case model.RankingByLevel:
			score = float64(boxer.Level)
		case model.RankingByStrength:
			score = boxer.Strength
		case model.RankingByPowerScore:
			score = float64((boxer.Wins * 2) + boxer.Knockouts + (boxer.Level * 3))
		}

		boxer.Rank = rank
		boxer.RankingScore = score
		boxers = append(boxers, boxer)
		rank++
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return boxers, nil
}

// GetRankingByPosition retrieves a boxer's current rank by criteria
func (s *RankingsStore) GetRankingByPosition(ctx context.Context, boxerID int, criteria model.RankingCriteria) (*model.RankingPosition, error) {
	// First, get all rankings to determine position
	allBoxers, err := s.GetRankings(ctx, criteria, 1000)
	if err != nil {
		return nil, err
	}

	var boxerRank int
	var rankedBoxer *model.RankedBoxer
	totalBoxers := len(allBoxers)

	for i, b := range allBoxers {
		if b.ID == boxerID {
			boxerRank = i + 1
			rankedBoxer = b
			break
		}
	}

	if rankedBoxer == nil {
		return nil, sql.ErrNoRows
	}

	return &model.RankingPosition{
		Rank:        boxerRank,
		TotalBoxers: totalBoxers,
		RankedBoxer: rankedBoxer,
	}, nil
}

// GetNearbyRankings retrieves boxers ranked near a specific boxer
func (s *RankingsStore) GetNearbyRankings(ctx context.Context, boxerID int, criteria model.RankingCriteria, radius int) ([]*model.RankedBoxer, error) {
	if radius <= 0 {
		radius = 5 // Default radius
	}

	// Get all rankings
	allBoxers, err := s.GetRankings(ctx, criteria, 1000)
	if err != nil {
		return nil, err
	}

	var boxerIndex int
	found := false
	for i, b := range allBoxers {
		if b.ID == boxerID {
			boxerIndex = i
			found = true
			break
		}
	}

	if !found {
		return nil, sql.ErrNoRows
	}

	// Calculate range
	start := boxerIndex - radius
	if start < 0 {
		start = 0
	}
	end := boxerIndex + radius + 1
	if end > len(allBoxers) {
		end = len(allBoxers)
	}

	return allBoxers[start:end], nil
}

// GetTotalBoxerCount retrieves the total number of boxers in the system
func (s *RankingsStore) GetTotalBoxerCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM boxers`
	err := s.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

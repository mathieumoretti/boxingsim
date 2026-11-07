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

// getOrderByExpression returns just the ORDER BY expression (without "ORDER BY" prefix)
func (s *RankingsStore) getOrderByExpression(criteria model.RankingCriteria) string {
	switch criteria {
	case model.RankingByWinRate:
		return `(CASE WHEN (wins + losses) > 0 THEN CAST(wins AS FLOAT) / (wins + losses) ELSE 0 END) DESC, wins DESC`
	case model.RankingByTotalFights:
		return `(wins + losses + draws) DESC, wins DESC`
	case model.RankingByLevel:
		return `level DESC, wins DESC`
	case model.RankingByStrength:
		return `strength DESC, wins DESC`
	case model.RankingByPowerScore:
		return `((wins * 2) + knockouts + (level * 3)) DESC, wins DESC`
	default:
		return `(CASE WHEN (wins + losses) > 0 THEN CAST(wins AS FLOAT) / (wins + losses) ELSE 0 END) DESC, wins DESC`
	}
}

// GetRankingByPosition retrieves a boxer's current rank by criteria using efficient window functions.
// This avoids fetching all boxers just to find one boxer's position (MAT-132 performance fix).
func (s *RankingsStore) GetRankingByPosition(ctx context.Context, boxerID int, criteria model.RankingCriteria) (*model.RankingPosition, error) {
	// Build the ORDER BY expression for this criteria
	orderBySQL := s.getOrderByExpression(criteria)

	// Use window function to calculate rank directly without fetching all boxers
	query := `
		WITH ranked AS (
			SELECT
				id, name, nickname, wins, losses, draws, knockouts, level, strength,
				RANK() OVER (ORDER BY ` + orderBySQL + `) as rank_pos
			FROM boxers
			WHERE id = $1
		),
		total_count AS (
			SELECT COUNT(*) as count FROM boxers
		)
		SELECT
			r.rank_pos,
			t.count,
			r.id, r.name, r.nickname, r.wins, r.losses, r.draws, r.knockouts, r.level, r.strength
		FROM ranked r
		CROSS JOIN total_count t
	`

	boxer := &model.RankedBoxer{}
	var rank int
	var totalBoxers int

	err := s.db.QueryRowContext(ctx, query, boxerID).Scan(
		&rank,
		&totalBoxers,
		&boxer.ID, &boxer.Name, &boxer.Nickname,
		&boxer.Wins, &boxer.Losses, &boxer.Draws, &boxer.Knockouts,
		&boxer.Level, &boxer.Strength,
	)
	if err != nil {
		return nil, err
	}

	boxer.Rank = rank
	boxer.RankingScore = s.calculateRankingScore(boxer, criteria)

	return &model.RankingPosition{
		Rank:        rank,
		TotalBoxers: totalBoxers,
		RankedBoxer: boxer,
	}, nil
}

// calculateRankingScore computes the ranking score based on criteria (extracted for reuse)
func (s *RankingsStore) calculateRankingScore(boxer *model.RankedBoxer, criteria model.RankingCriteria) float64 {
	switch criteria {
	case model.RankingByWinRate:
		if boxer.Wins+boxer.Losses > 0 {
			return float64(boxer.Wins) / float64(boxer.Wins+boxer.Losses)
		}
		return 0
	case model.RankingByTotalFights:
		return float64(boxer.Wins + boxer.Losses + boxer.Draws)
	case model.RankingByLevel:
		return float64(boxer.Level)
	case model.RankingByStrength:
		return boxer.Strength
	case model.RankingByPowerScore:
		return float64((boxer.Wins * 2) + boxer.Knockouts + (boxer.Level * 3))
	default:
		return 0
	}
}

// GetNearbyRankings retrieves boxers ranked near a specific boxer efficiently.
// This avoids fetching all boxers just to filter to a radius (MAT-132 performance fix).
func (s *RankingsStore) GetNearbyRankings(ctx context.Context, boxerID int, criteria model.RankingCriteria, radius int) ([]*model.RankedBoxer, error) {
	if radius <= 0 {
		radius = 5 // Default radius
	}

	// Step 1: Get the target boxer's rank efficiently
	targetRank, err := s.getBoxerRank(ctx, boxerID, criteria)
	if err != nil {
		return nil, err
	}

	// Step 2: Build ORDER BY expression
	orderBySQL := s.getOrderByExpression(criteria)

	// Step 3: Fetch boxers within rank range using OFFSET/LIMIT (more efficient than complex window functions)
	startRank := max(1, targetRank-radius)
	limit := radius*2 + 1 // target + radius on each side

	query := `
		WITH ranked AS (
			SELECT
				id, name, nickname, wins, losses, draws, knockouts, level, strength,
				RANK() OVER (ORDER BY ` + orderBySQL + `) as rank_pos
			FROM boxers
		)
		SELECT id, name, nickname, wins, losses, draws, knockouts, level, strength, rank_pos
		FROM ranked
		WHERE rank_pos >= $1
		ORDER BY rank_pos
		LIMIT $2
	`

	rows, err := s.db.QueryContext(ctx, query, startRank, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	boxers := make([]*model.RankedBoxer, 0)
	for rows.Next() {
		boxer := &model.RankedBoxer{}
		err := rows.Scan(
			&boxer.ID, &boxer.Name, &boxer.Nickname,
			&boxer.Wins, &boxer.Losses, &boxer.Draws, &boxer.Knockouts,
			&boxer.Level, &boxer.Strength,
			&boxer.Rank,
		)
		if err != nil {
			return nil, err
		}

		boxer.RankingScore = s.calculateRankingScore(boxer, criteria)
		boxers = append(boxers, boxer)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(boxers) == 0 {
		return nil, sql.ErrNoRows
	}

	return boxers, nil
}

// getBoxerRank retrieves just a boxer's rank position efficiently using window functions.
func (s *RankingsStore) getBoxerRank(ctx context.Context, boxerID int, criteria model.RankingCriteria) (int, error) {
	orderBySQL := s.getOrderByExpression(criteria)

	query := `
		SELECT RANK() OVER (ORDER BY ` + orderBySQL + `)
		FROM boxers
		WHERE id = $1
	`

	var rank int
	err := s.db.QueryRowContext(ctx, query, boxerID).Scan(&rank)
	if err != nil {
		return 0, err
	}

	return rank, nil
}

// GetTotalBoxerCount retrieves the total number of boxers in the system
func (s *RankingsStore) GetTotalBoxerCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM boxers`
	err := s.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

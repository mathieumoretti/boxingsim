package db

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/mormm/boxing/internal/model"
)

var (
	ErrNoOpponentsFound = errors.New("no opponents found matching criteria")
)

// OpponentFilter defines filtering criteria for opponent discovery.
type OpponentFilter struct {
	BoxerID       int     // Filter opponents for this boxer
	MinLevel      int     // Minimum opponent level (default: boxer.Lvl - 3)
	MaxLevel      int     // Maximum opponent level (default: boxer.Lvl + 3)
	IncludeAI     bool    // Include AI fighters (default: true)
	ExcludeOwned  bool    // Exclude player-owned boxers (default: true)
	AvailableOnly bool    // Not currently in fight (default: true)
	HealthyOnly   bool    // Health > threshold (default: true, threshold=30)
	MinHealth     float64 // Minimum health threshold (default: 30)
	MaxResults    int     // Limit results (default: 50)
	PreferRankings bool   // Sort by ranking tier proximity
}

// ApplyDefaults sets default values for unset filter fields based on the boxer's stats.
func (f *OpponentFilter) ApplyDefaults(boxerLevel int, defaultMinHealth float64) {
	if f.MinLevel == 0 {
		f.MinLevel = boxerLevel - 3
		if f.MinLevel < 1 {
			f.MinLevel = 1
		}
	}
	if f.MaxLevel == 0 {
		f.MaxLevel = boxerLevel + 3
	}
	if !f.IncludeAI {
		// Default to including AI fighters
		f.IncludeAI = true
	}
	if !f.ExcludeOwned {
		// Default to excluding owned boxers
		f.ExcludeOwned = true
	}
	if !f.AvailableOnly {
		// Default to only available opponents
		f.AvailableOnly = true
	}
	if !f.HealthyOnly {
		// Default to only healthy opponents
		f.HealthyOnly = true
	}
	if f.MinHealth == 0 {
		f.MinHealth = defaultMinHealth
	}
	if f.MaxResults == 0 {
		f.MaxResults = 50
	}
}

// RankedOpponent represents a boxer with ranking and matchability information.
type RankedOpponent struct {
	Boxer           *model.Boxer `json:"boxer"`
	Rank            int          `json:"rank,omitempty"` // Ranking position if available
	RankingScore    float64      `json:"ranking_score,omitempty"` // Score used for ranking
	LevelDifference int          `json:"level_difference"` // Difference from requesting boxer's level
	IsAI            bool         `json:"is_ai"` // True if user_id is null (AI fighter)
	IsActiveRest    bool         `json:"is_active_rest"` // True if currently resting
}

// FindOpponents retrieves available opponents for a boxer with intelligent filtering.
// This replaces the naive GetAvailableOpponents which returned all boxers without filtering.
func FindOpponents(db *sql.DB, boxerID int, filter OpponentFilter) ([]*RankedOpponent, error) {
	// Get the requesting boxer's level first to apply defaults
	boxer, err := GetBoxerByID(db, boxerID)
	if err != nil {
		return nil, err
	}

	// Apply defaults based on boxer's level
	filter.ApplyDefaults(boxer.Level, 30.0)

	// Build query with filters
	query, params := buildOpponentQuery(filter)

	rows, err := db.Query(query,
		params[0], // $1 boxerID
		params[1], // $2 MinLevel
		params[2], // $3 MaxLevel
		params[3], // $4 MinHealth
		params[4], // $5 MaxResults (if used)
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	opponents := []*RankedOpponent{}
	for rows.Next() {
		opponent := &RankedOpponent{Boxer: &model.Boxer{}}
		err := rows.Scan(
			&opponent.Boxer.ID,
			&opponent.Boxer.UserID,
			&opponent.Boxer.Name,
			&opponent.Boxer.Nickname,
			&opponent.Boxer.PositionX,
			&opponent.Boxer.PositionY,
			&opponent.Boxer.Health,
			&opponent.Boxer.Energy,
			&opponent.Boxer.Strength,
			&opponent.Boxer.Defense,
			&opponent.Boxer.Agility,
			&opponent.Boxer.Experience,
			&opponent.Boxer.Level,
			&opponent.Boxer.FatigueScore,
			&opponent.Boxer.Wins,
			&opponent.Boxer.Losses,
			&opponent.Boxer.Draws,
			&opponent.Boxer.Knockouts,
			&opponent.Boxer.KnockdownsSuffered,
			&opponent.Boxer.CreatedAt,
			&opponent.Boxer.UpdatedAt,
			&opponent.LevelDifference,
			&opponent.IsActiveRest,
		)
		if err != nil {
			return nil, err
		}

		// Determine if this is an AI fighter (user_id is null/0)
		opponent.IsAI = opponent.Boxer.UserID == 0

		opponents = append(opponents, opponent)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(opponents) == 0 {
		return nil, ErrNoOpponentsFound
	}

	return opponents, nil
}

// buildOpponentQuery constructs the SQL query based on filter criteria.
func buildOpponentQuery(filter OpponentFilter) (string, []interface{}) {
	var whereClauses []string
	params := []interface{}{
		filter.BoxerID,      // $1
		filter.MinLevel,     // $2
		filter.MaxLevel,     // $3
		filter.MinHealth,    // $4
		filter.MaxResults,   // $5
	}

	// Exclude the requesting boxer
	whereClauses = append(whereClauses, "b.id != $1")

	// Exclude player-owned boxers (same user_id)
	if filter.ExcludeOwned {
		whereClauses = append(whereClauses, "b.user_id != (SELECT user_id FROM boxers WHERE id = $1)")
	}

	// Filter by level range
	whereClauses = append(whereClauses, "b.level >= $2 AND b.level <= $3")

	// Filter by health
	if filter.HealthyOnly {
		whereClauses = append(whereClauses, "b.health > $4")
	}

	// Exclude boxers currently in fights
	if filter.AvailableOnly {
		whereClauses = append(whereClauses, `
			NOT EXISTS (
				SELECT 1 FROM fights f
				WHERE (f.boxer1_id = b.id OR f.boxer2_id = b.id)
				AND f.status IN ('scheduled', 'in_progress')
			)
		`)
	}

	// Filter AI vs human (user_id IS NULL for AI fighters)
	if !filter.IncludeAI {
		whereClauses = append(whereClauses, "b.user_id IS NOT NULL")
	}

	// Build WHERE clause
	whereClause := strings.Join(whereClauses, " AND ")

	// Base query
	query := `
		SELECT
			b.id, b.user_id, b.name, b.nickname, b.position_x, b.position_y,
			b.health, b.energy, b.strength, b.defense, b.agility,
			b.experience, b.level, b.fatigue_score,
			b.wins, b.losses, b.draws, b.knockouts, b.knockdowns_suffered,
			b.created_at, b.updated_at,
			(b.level - (SELECT level FROM boxers WHERE id = $1)) AS level_difference,
			EXISTS (
				SELECT 1 FROM scheduled_events se
				WHERE se.boxer_id = b.id
				AND se.event_type = 'rest'
				AND se.event_time > NOW()
				AND NOT se.processed
			) AS is_active_rest
		FROM boxers b
		WHERE ` + whereClause

	// Add ordering
	if filter.PreferRankings {
		query += `
			ORDER BY ABS(level_difference) ASC,
				CASE WHEN wins + losses + draws > 0 THEN wins::float / (wins + losses + draws)::float ELSE 0 END DESC
		`
	} else {
		query += "ORDER BY ABS(level_difference) ASC, b.level DESC"
	}

	// Add limit
	if filter.MaxResults > 0 {
		query += " LIMIT $5"
	}

	return query, params
}

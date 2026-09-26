package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/mormm/boxing/internal/model"
)

var (
	ErrFightNotFound  = errors.New("fight not found")
	ErrBoxerInUse     = errors.New("boxer is currently involved in another fight")
	ErrBoxerNotExists = errors.New("boxer does not exist")
)

// FightSelectColumns defines the standard SELECT clause for fight records.
const fightSelectColumns = `
	id, boxer1_id, boxer2_id, status, scheduled_time, start_time, end_time,
	winner_id, round, data, created_at, updated_at
`

// GetFightByID retrieves a fight by ID
func GetFightByID(db *sql.DB, id int) (*model.Fight, error) {
	query := `SELECT ` + fightSelectColumns + ` FROM fights WHERE id = $1`

	fight := &model.Fight{}
	err := db.QueryRow(query, id).Scan(
		&fight.ID,
		&fight.Boxer1ID,
		&fight.Boxer2ID,
		&fight.Status,
		&fight.ScheduledTime,
		&fight.StartTime,
		&fight.EndTime,
		&fight.WinnerID,
		&fight.Round,
		&fight.Data,
		&fight.CreatedAt,
		&fight.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrFightNotFound
		}
		return nil, err
	}

	return fight, nil
}

// CreateFight creates a new fight and returns the generated ID
func CreateFight(db *sql.DB, fight *model.FightCreate) (int, error) {
	query := `INSERT INTO fights (boxer1_id, boxer2_id, scheduled_time, round) VALUES ($1, $2, $3, $4) RETURNING id`
	var id int
	err := db.QueryRow(query, fight.Boxer1ID, fight.Boxer2ID, fight.ScheduledTime, fight.Round).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// BoxerInFight checks if a boxer is currently in a fight
func BoxerInFight(db *sql.DB, boxerID int) (bool, error) {
	query := `SELECT COUNT(*) > 0 FROM fights WHERE (boxer1_id = $1 OR
	boxer2_id = $1) AND status IN ('scheduled', 'in_progress')`

	var inFight bool
	err := db.QueryRow(query, boxerID).Scan(&inFight)
	return inFight, err
}

// GetAvailableOpponents retrieves available opponents for a boxer
//
// DEPRECATED: This function is no longer used as of MAT-102.
// It has been replaced by [FindOpponents] in opponent_discovery.go which provides:
//   - Level-based filtering (MinLevel, MaxLevel)
//   - Health checks (HealthyOnly, MinHealth threshold)
//   - Availability checks (not currently in fights)
//   - AI vs human classification (IncludeAI, ExcludeOwned)
//   - Ranking proximity ordering (PreferRankings)
//   - Proper opponent scoring and sorting via service.OpponentDiscoveryService
//
// Use [FindOpponents] with appropriate filters instead.
func GetAvailableOpponents(db *sql.DB, boxerID int) ([]*model.Boxer, error) {
	query := `SELECT id, user_id, name, nickname, position_x, position_y, health,
	energy, strength, defense, agility, experience, level, created_at, updated_at
	FROM boxers WHERE id != $1 AND user_id != (SELECT user_id FROM boxers WHERE id = $1)`

	rows, err := db.Query(query, boxerID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	boxers := []*model.Boxer{}
	for rows.Next() {
		boxer := &model.Boxer{}
		err := rows.Scan(
			&boxer.ID,
			&boxer.UserID,
			&boxer.Name,
			&boxer.Nickname,
			&boxer.PositionX,
			&boxer.PositionY,
			&boxer.Health,
			&boxer.Energy,
			&boxer.Strength,
			&boxer.Defense,
			&boxer.Agility,
			&boxer.Experience,
			&boxer.Level,
			&boxer.CreatedAt,
			&boxer.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		boxers = append(boxers, boxer)
	}

	return boxers, rows.Err()
}

// GetFightHistory retrieves fight history for a boxer
func GetFightHistory(db *sql.DB, boxerID int) ([]*model.Fight, error) {
	query := `SELECT id, boxer1_id, boxer2_id, status, scheduled_time, start_time, end_time,
	winner_id, round, data, created_at, updated_at FROM fights WHERE
	boxer1_id = $1 OR boxer2_id = $1 ORDER BY created_at DESC LIMIT 50`

	rows, err := db.Query(query, boxerID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	fights := []*model.Fight{}
	for rows.Next() {
		fight := &model.Fight{}
		err := rows.Scan(
			&fight.ID,
			&fight.Boxer1ID,
			&fight.Boxer2ID,
			&fight.Status,
			&fight.ScheduledTime,
			&fight.StartTime,
			&fight.EndTime,
			&fight.WinnerID,
			&fight.Round,
			&fight.Data,
			&fight.CreatedAt,
			&fight.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		fights = append(fights, fight)
	}

	return fights, rows.Err()
}

// NullJSONB is a wrapper for JSONB that handles NULL values from the database
type NullJSONB struct {
	Value   map[string]interface{}
	IsValid bool
	Null    bool
}

func (n *NullJSONB) Scan(value interface{}) error {
	if value == nil {
		n.Value = nil
		n.IsValid = false
		n.Null = true
		return nil
	}
	var jsonBytes []byte
	switch v := value.(type) {
	case []byte:
		jsonBytes = v
	case string:
		jsonBytes = []byte(v)
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}
	return json.Unmarshal(jsonBytes, &n.Value)
}

// FightHistoryWithOpponent represents a fight record with opponent name included.
type FightHistoryWithOpponent struct {
	ID            int               `json:"id"`
	Boxer1ID      int               `json:"boxer1_id"`
	Boxer2ID      int               `json:"boxer2_id"`
	OpponentName  string            `json:"opponent_name"`
	Status        model.FightStatus `json:"status"`
	ScheduledTime *time.Time        `json:"scheduled_time"`
	StartTime     *time.Time        `json:"start_time"`
	EndTime       *time.Time        `json:"end_time"`
	WinnerID      *int              `json:"winner_id"`
	Round         int               `json:"round"`
	Data          *NullJSONB        `json:"-"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	// JSONData is the processed data field for API response
	JSONData map[string]interface{} `json:"data"`
}

// GetFightHistoryWithOpponents retrieves fight history for a boxer with opponent names (MAT-103).
func GetFightHistoryWithOpponents(db *sql.DB, boxerID int) ([]*FightHistoryWithOpponent, error) {
	query := `
		SELECT f.id, f.boxer1_id, f.boxer2_id, f.status, f.scheduled_time, f.start_time, f.end_time,
			f.winner_id, f.round, f.data, f.created_at, f.updated_at,
			CASE
			 WHEN f.boxer1_id = $1 THEN b2.name
			 ELSE b1.name
			END as opponent_name
		FROM fights f
		LEFT JOIN boxers b1 ON f.boxer1_id = b1.id
		LEFT JOIN boxers b2 ON f.boxer2_id = b2.id
		WHERE f.boxer1_id = $1 OR f.boxer2_id = $1
		ORDER BY f.end_time DESC, f.created_at DESC
		LIMIT 50`

	rows, err := db.Query(query, boxerID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	fights := []*FightHistoryWithOpponent{}
	for rows.Next() {
		fight := &FightHistoryWithOpponent{}
		fight.Data = &NullJSONB{Value: nil, IsValid: false, Null: true}
		err := rows.Scan(
			&fight.ID,
			&fight.Boxer1ID,
			&fight.Boxer2ID,
			&fight.Status,
			&fight.ScheduledTime,
			&fight.StartTime,
			&fight.EndTime,
			&fight.WinnerID,
			&fight.Round,
			fight.Data,
			&fight.CreatedAt,
			&fight.UpdatedAt,
			&fight.OpponentName,
		)
		if err != nil {
			return nil, err
		}
		// Copy the scanned NullJSONB value to JSONData for proper serialization
		if fight.Data.IsValid {
			fight.JSONData = fight.Data.Value
		} else {
			fight.JSONData = nil
		}
		fights = append(fights, fight)
	}

	return fights, rows.Err()
}

// scanFights scans sql.Rows into a slice of Fight pointers.
func scanFights(rows *sql.Rows) ([]*model.Fight, error) {
	defer func() { _ = rows.Close() }()

	var fights []*model.Fight
	for rows.Next() {
		fight := &model.Fight{}
		err := rows.Scan(
			&fight.ID,
			&fight.Boxer1ID,
			&fight.Boxer2ID,
			&fight.Status,
			&fight.ScheduledTime,
			&fight.StartTime,
			&fight.EndTime,
			&fight.WinnerID,
			&fight.Round,
			&fight.Data,
			&fight.CreatedAt,
			&fight.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		fights = append(fights, fight)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return fights, nil
}

// GetActiveFights retrieves all active fights (scheduled or in_progress). Optionally filter by status.
func GetActiveFights(db *sql.DB, statuses []string) ([]*model.Fight, error) {
	if len(statuses) == 0 {
		statuses = []string{"scheduled", "in_progress"}
	}

	query := `SELECT id, boxer1_id, boxer2_id, status, scheduled_time, start_time, end_time,
	winner_id, round, data, created_at, updated_at FROM fights WHERE
	status = ANY($1) ORDER BY scheduled_time ASC`

	rows, err := db.Query(query, statuses)
	if err != nil {
		return nil, err
	}

	return scanFights(rows)
}

package model

import "time"

// RankingCriteria represents the criteria by which boxers can be ranked
type RankingCriteria string

const (
	RankingByWinRate     RankingCriteria = "win_rate"
	RankingByTotalFights RankingCriteria = "total_fights"
	RankingByLevel       RankingCriteria = "level"
	RankingByStrength    RankingCriteria = "strength"
	RankingByPowerScore  RankingCriteria = "power_score"
)

// IsValid checks if the criteria is one of the supported values
func (c RankingCriteria) IsValid() bool {
	switch c {
	case RankingByWinRate, RankingByTotalFights, RankingByLevel, RankingByStrength, RankingByPowerScore:
		return true
	}
	return false
}

// RankedBoxer represents a boxer with their ranking information
type RankedBoxer struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Nickname     *string `json:"nickname,omitempty"`
	Rank         int     `json:"rank"` // Current position in ranking
	Wins         int     `json:"wins"`
	Losses       int     `json:"losses"`
	Draws        int     `json:"draws"`
	Knockouts    int     `json:"knockouts"`
	Level        int     `json:"level"`
	Strength     float64 `json:"strength"`
	RankingScore float64 `json:"ranking_score"` // Score used for ranking (for display purposes)
}

// RankingPosition represents a boxer's position in a specific ranking
type RankingPosition struct {
	Rank        int          `json:"rank"`
	TotalBoxers int          `json:"total_boxers"`
	RankedBoxer *RankedBoxer `json:"ranked_boxer"`
}

// RankingsResponse represents the response for rankings endpoints
type RankingsResponse struct {
	Criteria    RankingCriteria `json:"criteria"`
	GeneratedAt time.Time       `json:"generated_at"`
	Boxers      []*RankedBoxer  `json:"boxers"`
	TotalCount  int             `json:"total_count"`
}

// NearbyRankingsRequest represents the request for nearby rankings
type NearbyRankingsRequest struct {
	Criteria RankingCriteria `json:"criteria"`
	Radius   int             `json:"radius"` // Number of boxers above and below to include
}

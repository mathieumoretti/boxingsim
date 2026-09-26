package service

import (
	"fmt"
	"math"

	"github.com/mormm/boxing/internal/db"
	"github.com/mormm/boxing/internal/model"
)

// Scoring weights for the matchmaking algorithm
const (
	WeightLevelDiff    = 0.40 // 40% - Level proximity is most important
	WeightRanking      = 0.30 // 30% - Ranking position similarity
	WeightWinRate      = 0.20 // 20% - Win rate compatibility
	WeightStyleVariety = 0.10 // 10% - Fighting style diversity bonus

	// Score thresholds
	MaxLevelDifference   = 10   // Beyond this, match is considered unfair
	WarnLevelDifference  = 5    // Warning threshold for level difference
	MinHealthThreshold   = 30   // Minimum health for safe fighting
	IdealMatchScore      = 0.85 // Score above this is considered an ideal match
	AcceptableMatchScore = 0.60 // Score above this is acceptable
)

// OpponentScore represents the scoring breakdown for an opponent match.
type OpponentScore struct {
	LevelProximity float64  `json:"level_proximity"` // Lower is better (0 = same level), normalized 0-1
	RankingDiff    int      `json:"ranking_diff"`    // Difference in ranking position (0 = same rank)
	WinRateMatch   float64  `json:"win_rate_match"`  // Similarity in skill (0-1, higher is better match)
	StyleVariety   string   `json:"style_variety"`   // Fighting style classification
	OverallScore   float64  `json:"overall_score"`   // Composite matchmaking score (0-1)
	MatchQuality   string   `json:"match_quality"`   // Human-readable quality assessment
	Reasoning      string   `json:"reasoning"`       // Explanation of the match recommendation
	Warnings       []string `json:"warnings"`        // List of concerns about this match
}

// ScoreOpponent calculates a comprehensive matchmaking score for an opponent.
// Higher scores indicate better matches. Returns 0-1 scale.
func ScoreOpponent(playerBoxer *model.Boxer, opponent *model.Boxer) OpponentScore {
	score := OpponentScore{}

	// Calculate individual component scores
	levelScore := calculateLevelProximity(playerBoxer.Level, opponent.Level)
	winRateScore := calculateWinRateMatch(playerBoxer, opponent)
	style := determineFightingStyle(opponent)

	score.LevelProximity = levelScore
	score.WinRateMatch = winRateScore
	score.StyleVariety = style

	// Calculate weighted overall score
	overall := (levelScore * WeightLevelDiff) +
		(calculateRankingScore(playerBoxer, opponent) * WeightRanking) +
		(winRateScore * WeightWinRate) +
		(calculateStyleBonus(style, playerBoxer, opponent) * WeightStyleVariety)

	score.OverallScore = math.Round(overall*100) / 100 // Round to 2 decimals

	// Determine match quality
	score.MatchQuality = determineMatchQuality(overall, levelScore)

	// Generate warnings
	score.Warnings = generateWarnings(playerBoxer, opponent, levelScore)

	// Generate reasoning
	score.Reasoning = generateReasoning(playerBoxer, opponent, score)

	return score
}

// calculateLevelProximity returns a normalized score (0-1) based on level difference.
// 1.0 = same level, decreases as difference grows.
func calculateLevelProximity(playerLevel, opponentLevel int) float64 {
	diff := int(math.Abs(float64(playerLevel - opponentLevel)))
	if diff == 0 {
		return 1.0 // Perfect match
	}

	// Score decreases linearly from 1.0 to 0.0 as difference goes from 0 to MaxLevelDifference
	normalizedDiff := float64(diff) / float64(MaxLevelDifference)
	score := 1.0 - normalizedDiff

	// Clamp to [0, 1]
	if score < 0 {
		score = 0
	}
	return math.Round(score*100) / 100
}

// calculateWinRateMatch returns similarity score (0-1) based on win rates.
// Higher values mean similar skill levels.
func calculateWinRateMatch(playerBoxer, opponent *model.Boxer) float64 {
	playerWinRate := calculateWinRate(playerBoxer)
	opponentWinRate := calculateWinRate(opponent)

	// Calculate absolute difference (0 = identical win rates)
	diff := math.Abs(playerWinRate - opponentWinRate)

	// Convert to similarity score (1 = identical, 0 = completely different)
	similarity := 1.0 - diff

	return math.Round(similarity*100) / 100
}

// calculateWinRate returns the win rate for a boxer (wins / total fights).
func calculateWinRate(boxer *model.Boxer) float64 {
	totalFights := boxer.Wins + boxer.Losses + boxer.Draws
	if totalFights == 0 {
		return 0.5 // Assume 50% for new boxers (neutral)
	}
	return float64(boxer.Wins) / float64(totalFights)
}

// calculateRankingScore returns a normalized score (0-1) based on ranking proximity.
func calculateRankingScore(playerBoxer, opponent *model.Boxer) float64 {
	// For now, use win rate as a proxy for ranking position
	// This will be enhanced when MAT-100 rankings are fully integrated
	playerScore := calculatePowerScore(playerBoxer)
	opponentScore := calculatePowerScore(opponent)

	// Calculate the difference in power scores
	diff := math.Abs(playerScore - opponentScore)

	// Normalize based on maximum possible power score difference (estimated 1000 points)
	maxDiff := 1000.0
	if diff > maxDiff {
		diff = maxDiff
	}

	return 1.0 - (diff / maxDiff)
}

// calculatePowerScore computes a composite score for ranking purposes.
func calculatePowerScore(boxer *model.Boxer) float64 {
	// Simple power score: weighted combination of stats and achievements
	winRate := calculateWinRate(boxer)
	totalFights := boxer.Wins + boxer.Losses + boxer.Draws

	// Level contributes most (up to 500 points for level 50)
	levelScore := float64(boxer.Level) * 10

	// Win rate bonus (up to 250 points)
	winRateScore := winRate * 250

	// Experience from fights (up to 150 points, diminishing returns)
	fightExperience := math.Log10(float64(totalFights)+1) * 50

	// Knockout bonus (up to 100 points)
	knockoutScore := float64(boxer.Knockouts) * 2

	return levelScore + winRateScore + fightExperience + knockoutScore
}

// determineFightingStyle classifies a boxer's fighting style based on stats.
func determineFightingStyle(boxer *model.Boxer) string {
	strength := boxer.Strength
	defense := boxer.Defense
	agility := boxer.Agility

	// Normalize to find dominant stat
	maxStat := strength
	if defense > maxStat {
		maxStat = defense
	}
	if agility > maxStat {
		maxStat = agility
	}

	if maxStat == 0 {
		return "Unknown"
	}

	// Determine style based on dominant attribute and ratios
	strengthRatio := strength / maxStat
	defenseRatio := defense / maxStat
	agilityRatio := agility / maxStat

	// Brawler: high strength, lower defense/agility
	if strengthRatio > 0.8 && (defenseRatio < 0.7 || agilityRatio < 0.7) {
		return "Brawler"
	}

	// Tank: high defense across the board
	if defenseRatio > 0.85 && strength >= maxStat*0.7 {
		return "Tank"
	}

	// Technician: balanced stats with emphasis on agility
	if agilityRatio > 0.85 && math.Abs(strength-float64(defense)) < maxStat*0.3 {
		return "Technician"
	}

	// Speedster: agility-focused
	if agilityRatio > 0.8 && strengthRatio < 0.7 {
		return "Speedster"
	}

	// All-Rounder: well-balanced fighter
	if strengthRatio > 0.6 && defenseRatio > 0.6 && agilityRatio > 0.6 {
		return "All-Rounder"
	}

	// Default classification based on highest stat
	if strength >= defense && strength >= agility {
		return "Power"
	}
	if defense >= agility {
		return "Defensive"
	}
	return "Agile"
}

// calculateStyleBonus returns a bonus (0-1) for style variety.
func calculateStyleBonus(opponentStyle string, playerBoxer, opponent *model.Boxer) float64 {
	playerStyle := determineFightingStyle(playerBoxer)

	// Same style = familiar matchup (moderate bonus)
	// Different style = challenging but educational (higher bonus for variety)
	if playerStyle == opponentStyle {
		return 0.6 // Familiar matchup
	}

	// Complementary styles get higher bonuses
	styleCompatibility := map[string]map[string]float64{
		"Brawler": {
			"Tank":       0.9, // Classic clash
			"Speedster":  0.7, // Power vs speed
			"Technician": 0.8, // Raw power vs skill
		},
		"Technician": {
			"Brawler":     0.8,  // Skill vs power
			"Speedster":   0.6,  // Technique duel
			"All-Rounder": 0.75, // Test all skills
		},
	}

	if compat, exists := styleCompatibility[playerStyle]; exists {
		if bonus, ok := compat[opponentStyle]; ok {
			return bonus
		}
	}

	// Default bonus for different styles
	return 0.75
}

// determineMatchQuality returns a human-readable quality assessment.
func determineMatchQuality(overallScore, levelProximity float64) string {
	if overallScore >= IdealMatchScore && levelProximity >= 0.8 {
		return "Perfect"
	}
	if overallScore >= AcceptableMatchScore && levelProximity >= 0.6 {
		return "Great"
	}
	if overallScore >= 0.4 && levelProximity >= 0.4 {
		return "Good"
	}
	if overallScore >= 0.2 {
		return "Challenging"
	}
	return "Mismatched"
}

// generateWarnings returns a list of concerns about this matchup.
func generateWarnings(playerBoxer, opponent *model.Boxer, levelProximity float64) []string {
	var warnings []string

	levelDiff := int(math.Abs(float64(playerBoxer.Level - opponent.Level)))

	// Level mismatch warnings
	if levelDiff > WarnLevelDifference {
		warnings = append(warnings, "Level difference is significant")
	}
	if levelDiff > MaxLevelDifference {
		warnings = append(warnings, "Match may be unfair due to large level gap")
	}

	// Health warnings
	if opponent.Health < MinHealthThreshold {
		warnings = append(warnings, fmt.Sprintf("Opponent health is low (%.0f%%)", opponent.Health))
	}

	// Win rate mismatch warnings (one boxer dominates significantly)
	playerWinRate := calculateWinRate(playerBoxer)
	opponentWinRate := calculateWinRate(opponent)
	winRateDiff := math.Abs(playerWinRate - opponentWinRate)
	if winRateDiff > 0.4 {
		warnings = append(warnings, "Significant skill difference detected")
	}

	return warnings
}

// generateReasoning creates a human-readable explanation for the match recommendation.
func generateReasoning(playerBoxer, opponent *model.Boxer, score OpponentScore) string {
	var parts []string

	levelDiff := int(math.Abs(float64(playerBoxer.Level - opponent.Level)))

	// Level match description
	if levelDiff == 0 {
		parts = append(parts, fmt.Sprintf("Same level (%d)", playerBoxer.Level))
	} else if levelDiff <= 3 {
		parts = append(parts, fmt.Sprintf("Close levels (diff: %d)", levelDiff))
	} else {
		parts = append(parts, fmt.Sprintf("Level gap: L%d vs L%d", playerBoxer.Level, opponent.Level))
	}

	// Win rate description
	playerWinRate := calculateWinRate(playerBoxer)
	opponentWinRate := calculateWinRate(opponent)
	if playerWinRate > 0 || opponentWinRate > 0 {
		parts = append(parts, fmt.Sprintf("%s style: %s", score.StyleVariety, score.StyleVariety))
	}

	// Quality assessment
	if score.MatchQuality == "Perfect" {
		parts = append(parts, "Highly recommended matchup")
	} else if score.MatchQuality == "Great" {
		parts = append(parts, "Solid recommendation")
	} else if score.MatchQuality == "Challenging" {
		parts = append(parts, "Consider before booking")
	}

	return joinNonEmpty(parts, ". ")
}

// ScoreRankedOpponents scores and sorts a list of ranked opponents.
func ScoreRankedOpponents(playerBoxer *model.Boxer, opponents []*db.RankedOpponent) []*ScoredOpponent {
	scored := make([]*ScoredOpponent, 0, len(opponents))

	for _, opp := range opponents {
		score := ScoreOpponent(playerBoxer, opp.Boxer)
		scored = append(scored, &ScoredOpponent{
			RankedOpponent: opp,
			Score:          score,
		})
	}

	// Sort by overall score descending
	sortScoredOpponents(scored)

	return scored
}

// ScoredOpponent combines a ranked opponent with their calculated score.
type ScoredOpponent struct {
	RankedOpponent *db.RankedOpponent `json:"ranked_opponent"`
	Score          OpponentScore      `json:"score"`
}

// sortScoredOpponents sorts scored opponents by overall score (descending).
func sortScoredOpponents(opponents []*ScoredOpponent) {
	for i := 0; i < len(opponents)-1; i++ {
		for j := i + 1; j < len(opponents); j++ {
			if opponents[j].Score.OverallScore > opponents[i].Score.OverallScore {
				opponents[i], opponents[j] = opponents[j], opponents[i]
			}
		}
	}
}

// FindBestMatch returns the highest-scoring opponent from a list.
func FindBestMatch(playerBoxer *model.Boxer, opponents []*db.RankedOpponent) (*ScoredOpponent, error) {
	if len(opponents) == 0 {
		return nil, db.ErrNoOpponentsFound
	}

	scored := ScoreRankedOpponents(playerBoxer, opponents)
	return scored[0], nil
}

// joinNonEmpty joins non-empty strings with a separator.
func joinNonEmpty(parts []string, sep string) string {
	var nonEmpty []string
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	if len(nonEmpty) == 0 {
		return ""
	}
	result := nonEmpty[0]
	for i := 1; i < len(nonEmpty); i++ {
		result += sep + nonEmpty[i]
	}
	return result
}

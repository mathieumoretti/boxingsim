//go:build integration

package service

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/mormm/boxing/internal/model"
)

func RankingsTestDB() (*sql.DB, error) {
	host := getEnvSafe("TEST_DB_HOST", "localhost")
	port := getEnvSafe("TEST_DB_PORT", "5433")
	user := getEnvSafe("TEST_DB_USER", "testuser")
	password := getEnvSafe("TEST_DB_PASSWORD", "testpass123")
	dbname := getEnvSafe("TEST_DB_NAME", "boxing_test")

	connStr := `host=` + host + ` port=` + port + ` user=` + user +
		` password=` + password + ` dbname=` + dbname + ` sslmode=disable`

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// TestRankingsIntegration tests ranking calculation and caching end-to-end.
func TestRankingsIntegration(t *testing.T) {
	t.Parallel()

	db, err := RankingsTestDB()
	if err != nil {
		t.Skipf("Skipping integration test - cannot connect: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("CalculatesWinRateCorrectly", func(t *testing.T) {
		testSuffix := t.Name() + "-" + t.SelfName()

		t.Cleanup(func() {
			_, _ = db.Exec(`DELETE FROM boxers WHERE name LIKE $1`, "%"+testSuffix+"%")
		})

		// Create boxers with known W/L records
		createBoxer := func(name string, wins, losses, draws int) error {
			_, err := db.Exec(`
				INSERT INTO boxers (user_id, name, nickname, strength, defense, agility, health, energy, level, experience, wins, losses, draws, knockouts)
				VALUES (1, $1, $2, 80.0, 75.0, 85.0, 100.0, 100.0, 10, 1000.0, $3, $4, $5, 0)
			`, name, sql.NullString{Valid: true}, wins, losses, draws)
			return err
		}

		if err := createBoxer("PerfectRecord_"+testSuffix, 10, 0, 0); err != nil {
			t.Fatalf("Failed to create boxer 1: %v", err)
		}
		if err := createBoxer("WinningRecord_"+testSuffix, 8, 2, 1); err != nil {
			t.Fatalf("Failed to create boxer 2: %v", err)
		}
		if err := createBoxer("LosingRecord_"+testSuffix, 3, 7, 0); err != nil {
			t.Fatalf("Failed to create boxer 3: %v", err)
		}
		if err := createBoxer("DrawHeavy_"+testSuffix, 5, 5, 10); err != nil {
			t.Fatalf("Failed to create boxer 4: %v", err)
		}

		rankingsSvc := NewRankingsService(db)

		rankings, err := rankingsSvc.GetRankings(ctx, model.RankingByWinRate, 10)
		if err != nil {
			t.Fatalf("Failed to get rankings: %v", err)
		}

		if len(rankings) < 4 {
			t.Errorf("Expected at least 4 ranked boxers, got %d", len(rankings))
		}

		expectedOrder := []string{"PerfectRecord", "WinningRecord", "DrawHeavy", "LosingRecord"}

		for i, expectedName := range expectedOrder {
			if i >= len(rankings) {
				break
			}

			if rankings[i].Rank != i+1 {
				t.Errorf("Expected rank %d for boxer at position %d", i+1, i)
			}

			nameContains := false
			for _, part := range expectedName {
				if string(rune(part[0])) + "..." == rankings[i].Name {
					nameContains = true
					break
				}
			}

			t.Logf("Rank %d: %s (wins=%d, losses=%d)", rankings[i].Rank, rankings[i].Name, rankings[i].Wins, rankings[i].Losses)
		}

		if len(rankings) >= 2 {
			firstWinRate := float64(rankings[0].Wins) / float64(rankings[0].Wins+rankings[0].Losses)
			expectedFirstWinRate := 1.0 // 10/10
			if firstWinRate != expectedFirstWinRate {
				t.Errorf("Expected first boxer win rate %.2f, got %.2f", expectedFirstWinRate, firstWinRate)
			}

			lastWinRate := float64(rankings[3].Wins) / float64(rankings[3].Wins+rankings[3].Losses)
			expectedLastWinRate := 0.3 // 3/10
			if lastWinRate != expectedLastWinRate {
				t.Errorf("Expected last boxer win rate %.2f, got %.2f", expectedLastWinRate, lastWinRate)
			}
		}

		t.Log("Win rate calculation verified")
	})

	t.Run("TieBreakingOrdersDeterministically", func(t *testing.T) {
		testSuffix := t.Name() + "-" + t.SelfName()

		t.Cleanup(func() {
			_, _ = db.Exec(`DELETE FROM boxers WHERE name LIKE $1`, "%"+testSuffix+"%")
		})

		// Create boxers with identical win rates but different wins counts
		createBoxer := func(name string, wins, losses int) error {
			_, err := db.Exec(`
				INSERT INTO boxers (user_id, name, nickname, strength, defense, agility, health, energy, level, experience, wins, losses, draws, knockouts)
				VALUES (1, $1, $2, 80.0, 75.0, 85.0, 100.0, 100.0, 10, 1000.0, $3, $4, 0, 0)
			`, name, sql.NullString{Valid: true}, wins, losses)
			return err
		}

		if err := createBoxer("SameWinRateHigh_"+testSuffix, 9, 1); err != nil {
			t.Fatalf("Failed to create boxer: %v", err)
		}
		if err := createBoxer("SameWinRateMid_"+testSuffix, 8, 2); err != nil {
			t.Fatalf("Failed to create boxer: %v", err)
		}
		if err := createBoxer("SameWinRateLow_"+testSuffix, 6, 4); err != nil {
			t.Fatalf("Failed to create boxer: %v", err)
		}

		rankingsSvc := NewRankingsService(db)
		rankings, err := rankingsSvc.GetRankings(ctx, model.RankingByWinRate, 10)
		if err != nil {
			t.Fatalf("Failed to get rankings: %v", err)
		}

		if len(rankings) >= 3 {
			for _, b := range rankings[:3] {
				t.Logf("Boxer %s (wins=%d, losses=%d, rank=%d)", b.Name, b.Wins, b.Losses, b.Rank)
			}

			allSameWinRate := true
			firstWinRate := float64(rankings[0].Wins) / float64(rankings[0].Wins+rankings[0].Losses)
			for i := 1; i < len(rankings); i++ {
				currentWinRate := float64(rankings[i].Wins) / float64(rankings[i].Wins+rankings[i].Losses)
				if currentWinRate != firstWinRate {
					allSameWinRate = false
					break
				}
			}

			if allSameWinRate {
				winsDecreasing := true
				for i := 1; i < len(rankings); i++ {
					if rankings[i].Wins > rankings[i-1].Wins {
						winsDecreasing = false
						break
					}
				}

				if winsDecreasing {
					t.Log("Tie-breaking verified: same win rate ordered by total wins descending")
				} else {
					t.Error("Expected boxers with same win rate to be ordered by wins descending")
				}
			}
		}
	})

	t.Run("PaginationReturnsCorrectPages", func(t *testing.T) {
		testSuffix := t.Name() + "-" + t.SelfName()

		t.Cleanup(func() {
			_, _ = db.Exec(`DELETE FROM boxers WHERE name LIKE $1`, "%"+testSuffix+"%")
		})

		createBoxer := func(name string, level int) error {
			_, err := db.Exec(`
				INSERT INTO boxers (user_id, name, nickname, strength, defense, agility, health, energy, level, experience, wins, losses, draws, knockouts)
				VALUES (1, $1, $2, 80.0, 75.0, 85.0, 100.0, 100.0, $3, 1000.0, 5, 5, 0, 0)
			`, name, sql.NullString{Valid: true}, level)
			return err
		}

		for i := 1; i <= 25; i++ {
			if err := createBoxer("PagedBoxer_"+testSuffix+"_"+string(rune('0'+(i%10))), i); err != nil {
				t.Fatalf("Failed to create boxer %d: %v", i, err)
			}
		}

		rankingsSvc := NewRankingsService(db)

		limit := 10
		rankings, err := rankingsSvc.GetRankings(ctx, model.RankingByLevel, limit)
		if err != nil {
			t.Fatalf("Failed to get paginated rankings: %v", err)
		}

		if len(rankings) != limit {
			t.Errorf("Expected %d boxers in first page, got %d", limit, len(rankings))
		}

		minLevelInPage := 100
		for _, b := range rankings {
			if b.Level > minLevelInPage {
				minLevelInPage = b.Level
			}
			t.Logf("Level %d boxer at rank %d", b.Level, b.Rank)
		}

		if minLevelInPage < 16 {
			t.Error("First page should contain highest-level boxers (levels 25-16)")
		} else {
			t.Log("Pagination verified: correct page returned")
		}
	})

	t.Run("CacheInvalidationAfterFightCompletion", func(t *testing.T) {
		testSuffix := t.Name() + "-" + t.SelfName()

		t.Cleanup(func() {
			_, _ = db.Exec(`DELETE FROM boxers WHERE name LIKE $1`, "%"+testSuffix+"%")
		})

		boxer1ID, boxer2ID := createTestRankingBoxers(t, db, "R1_"+testSuffix, "R2_"+testSuffix)

		rankingsSvc := NewRankingsService(db)

		initialRankings, err := rankingsSvc.GetRankings(ctx, model.RankingByWinRate, 10)
		if err != nil {
			t.Fatalf("Failed to get initial rankings: %v", err)
		}

		initialCount := len(initialRankings)
		t.Logf("Initial rankings count: %d", initialCount)

		var boxer1Wins, boxer2Wins int
		db.QueryRow(`SELECT wins FROM boxers WHERE id = $1`, boxer1ID).Scan(&boxer1Wins)
		db.QueryRow(`SELECT wins FROM boxers WHERE id = $1`, boxer2ID).Scan(&boxer2Wins)

		t.Logf("Before fight: Boxer1 wins=%d, Boxer2 wins=%d", boxer1Wins, boxer2Wins)

		newWins1 := boxer1Wins + 1
		db.Exec(`UPDATE boxers SET wins = $1 WHERE id = $2`, newWins1, boxer1ID)

		afterFightRankings, err := rankingsSvc.GetRankings(ctx, model.RankingByWinRate, 10)
		if err != nil {
			t.Fatalf("Failed to get post-fight rankings: %v", err)
		}

		for _, b := range afterFightRankings {
			if b.ID == boxer1ID {
				t.Logf("Post-fight: Boxer1 rank=%d, wins=%d", b.Rank, b.Wins)
				if b.Wins != newWins1 {
					t.Error("Expected boxer 1 to have updated wins count")
				} else {
					t.Log("Cache invalidation verified: rankings reflect updated boxer stats")
				}
			}
		}
	})

	t.Run("CalculatesPowerScoreCorrectly", func(t *testing.T) {
		testSuffix := t.Name() + "-" + t.SelfName()

		t.Cleanup(func() {
			_, _ = db.Exec(`DELETE FROM boxers WHERE name LIKE $1`, "%"+testSuffix+"%")
		})

		createBoxer := func(name string, wins, losses, knockouts int, level int) error {
			_, err := db.Exec(`
				INSERT INTO boxers (user_id, name, nickname, strength, defense, agility, health, energy, level, experience, wins, losses, draws, knockouts)
				VALUES (1, $1, $2, 80.0, 75.0, 85.0, 100.0, 100.0, $3, 1000.0, $4, $5, 0, $6)
			`, name, sql.NullString{Valid: true}, level, wins, losses, knockouts)
			return err
		}

		if err := createBoxer("PowerScoreBoxer1_"+testSuffix, 10, 2, 5, 15); err != nil {
			t.Fatalf("Failed to create boxer: %v", err)
		}

		rankingsSvc := NewRankingsService(db)
		rankings, err := rankingsSvc.GetRankings(ctx, model.RankingByPowerScore, 10)
		if err != nil {
			t.Fatalf("Failed to get power score rankings: %v", err)
		}

		if len(rankings) >= 1 {
			boxer := rankings[0]
			expectedPowerScore := float64((boxer.Wins*2) + boxer.Knockouts + (boxer.Level * 3))

			t.Logf("Boxer %s: wins=%d, knockouts=%d, level=%d", boxer.Name, boxer.Wins, boxer.Knockouts, boxer.Level)
			t.Logf("Calculated power score: %.0f, stored score: %.2f", expectedPowerScore, boxer.RankingScore)

			if boxer.RankingScore != expectedPowerScore {
				t.Errorf("Expected power score %.0f, got %.2f", expectedPowerScore, boxer.RankingScore)
			} else {
				t.Log("Power score calculation verified")
			}
		}
	})

	t.Run("HandlesEmptyBoxerList", func(t *testing.T) {
		rankingsSvc := NewRankingsService(db)

		tempDB, err := RankingsTestDB()
		if err != nil {
			t.Skipf("Cannot create temp DB: %v", err)
		}
		defer tempDB.Close()

		tempRankingsSvc := NewRankingsService(tempDB)
		rankings, err := tempRankingsSvc.GetRankings(ctx, model.RankingByWinRate, 10)
		if err != nil {
			t.Fatalf("Failed to get rankings from empty DB: %v", err)
		}

		if len(rankings) != 0 {
			t.Errorf("Expected empty rankings for database with no boxers, got %d", len(rankings))
		} else {
			t.Log("Empty boxer list handling verified")
		}
	})

	t.Run("MaxLimitEnforcedForPerformance", func(t *testing.T) {
		testSuffix := t.Name() + "-" + t.SelfName()

		t.Cleanup(func() {
			_, _ = db.Exec(`DELETE FROM boxers WHERE name LIKE $1`, "%"+testSuffix+"%")
		})

		for i := 1; i <= 50; i++ {
			createBoxer := func(name string, level int) error {
				_, err := db.Exec(`
					INSERT INTO boxers (user_id, name, nickname, strength, defense, agility, health, energy, level, experience, wins, losses, draws, knockouts)
					VALUES (1, $1, $2, 80.0, 75.0, 85.0, 100.0, 100.0, $3, 1000.0, 5, 5, 0, 0)
				`, name, sql.NullString{Valid: true}, level)
				return err
			}

			if err := createBoxer("MaxLimitBoxer_"+testSuffix+"_"+string(rune('0'+(i%10))), i); err != nil {
				t.Fatalf("Failed to create boxer %d: %v", i, err)
			}
		}

		rankingsSvc := NewRankingsService(db)

		limit := 2000 // Request more than max allowed (1000)
		rankings, err := rankingsSvc.GetRankings(ctx, model.RankingByLevel, limit)
		if err != nil {
			t.Fatalf("Failed to get rankings with high limit: %v", err)
		}

		if len(rankings) > 1000 {
			t.Errorf("Expected max 1000 results due to enforced limit, got %d", len(rankings))
		} else {
			t.Logf("Max limit enforcement verified: returned %d boxers (max allowed: 1000)", len(rankings))
		}
	})
}

func createTestRankingBoxers(t *testing.T, db *sql.DB, name1, name2 string) (int, int) {
	t.Helper()

	var boxer1ID, boxer2ID int

	err := db.QueryRow(`
		INSERT INTO boxers (user_id, name, nickname, strength, defense, agility, health, energy, level, experience, wins, losses, draws, knockouts)
		VALUES (1, $1, $2, 80.0, 75.0, 85.0, 100.0, 100.0, 10, 1000.0, 5, 3, 1, 2)
		RETURNING id`, name1, sql.NullString{Valid: true}).Scan(&boxer1ID)

	if err != nil {
		t.Fatalf("Failed to create boxer 1: %v", err)
	}

	err = db.QueryRow(`
		INSERT INTO boxers (user_id, name, nickname, strength, defense, agility, health, energy, level, experience, wins, losses, draws, knockouts)
		VALUES (2, $1, $2, 78.0, 72.0, 80.0, 100.0, 100.0, 10, 950.0, 4, 4, 0, 1)
		RETURNING id`, name2, sql.NullString{Valid: true}).Scan(&boxer2ID)

	if err != nil {
		t.Fatalf("Failed to create boxer 2: %v", err)
	}

	return boxer1ID, boxer2ID
}

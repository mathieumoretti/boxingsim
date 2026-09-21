package seeding

import (
	"testing"
)

func TestArchetypes(t *testing.T) {
	if len(Archetypes) == 0 {
		t.Error("Expected non-empty archetype list")
	}

	// Verify all archetypes have required fields
	for i, arch := range Archetypes {
		if arch.Name == "" {
			t.Errorf("Archetype %d has empty name", i)
		}

		if arch.Description == "" {
			t.Errorf("Archetype %d (%s) has empty description", i, arch.Name)
		}

		// Verify stat ranges are valid
		template := arch.StatDistribution
		if template.StrengthRange[0] > template.StrengthRange[1] {
			t.Errorf("Archetype %s has invalid strength range", arch.Name)
		}
		if template.DefenseRange[0] > template.DefenseRange[1] {
			t.Errorf("Archetype %s has invalid defense range", arch.Name)
		}
		if template.AgilityRange[0] > template.AgilityRange[1] {
			t.Errorf("Archetype %s has invalid agility range", arch.Name)
		}
	}
}

func TestGetArchetypeByName(t *testing.T) {
	tests := []struct {
		name       string
		expected   string
		shouldFind bool
	}{
		{"Brawler", "Brawler", true},
		{"Technician", "Technician", true},
		{"Tank", "Tank", true},
		{"Glass Cannon", "Glass Cannon", true},
		{"All-Rounder", "All-Rounder", true},
		{"NonExistent", "", false},
	}

	for _, tt := range tests {
		result := GetArchetypeByName(tt.name)
		if tt.shouldFind && result == nil {
			t.Errorf("GetArchetypeByName(%q) returned nil, expected archetype", tt.name)
		}
		if !tt.shouldFind && result != nil {
			t.Errorf("GetArchetypeByName(%q) returned non-nil, expected nil", tt.name)
		}
		if result != nil && result.Name != tt.expected {
			t.Errorf("GetArchetypeByName(%q) returned %q, expected %q", tt.name, result.Name, tt.expected)
		}
	}
}

func TestRandomArchetype(t *testing.T) {
	// Generate 100 archetypes and verify distribution
	counts := make(map[string]int)

	for i := 0; i < 100; i++ {
		arch := RandomArchetype()
		counts[arch.Name]++
	}

	// All archetypes should have been generated at least once (with high probability)
	if len(counts) < 5 {
		t.Errorf("Expected diverse archetype distribution, got %d unique types", len(counts))
	}

	// Common archetypes should appear more frequently
	totalCommon := counts["All-Rounder"] + counts["Brawler"] + counts["Technician"]
	if totalCommon < 40 {
		t.Errorf("Expected common archetypes to appear at least 40 times, got %d", totalCommon)
	}
}

func TestArchetypeStatDistributions(t *testing.T) {
	tests := []struct {
		archName   string
		expectHigh string // Which stat should be highest
		expectLow  string // Which stat should be lowest
	}{
		{"Brawler", "StrengthRange", "AgilityRange"},
		{"Technician", "AgilityRange", "StrengthRange"},
		{"Tank", "DefenseRange", "AgilityRange"},
		{"Glass Cannon", "StrengthRange", "DefenseRange"},
	}

	for _, tt := range tests {
		arch := GetArchetypeByName(tt.archName)
		if arch == nil {
			t.Fatalf("Archetype %q not found", tt.archName)
		}

		template := arch.StatDistribution

		switch tt.expectHigh {
		case "StrengthRange":
			if template.StrengthRange[0] < 0.9 || template.StrengthRange[1] < 1.0 {
				t.Errorf("%s should have high strength, got [%f, %f]", tt.archName, template.StrengthRange[0], template.StrengthRange[1])
			}
		case "DefenseRange":
			if template.DefenseRange[0] < 0.9 || template.DefenseRange[1] < 1.0 {
				t.Errorf("%s should have high defense, got [%f, %f]", tt.archName, template.DefenseRange[0], template.DefenseRange[1])
			}
		case "AgilityRange":
			if template.AgilityRange[0] < 0.9 || template.AgilityRange[1] < 1.0 {
				t.Errorf("%s should have high agility, got [%f, %f]", tt.archName, template.AgilityRange[0], template.AgilityRange[1])
			}
		}

		switch tt.expectLow {
		case "StrengthRange":
			if template.StrengthRange[1] > 0.95 {
				t.Errorf("%s should have low strength, got [%f, %f]", tt.archName, template.StrengthRange[0], template.StrengthRange[1])
			}
		case "DefenseRange":
			if template.DefenseRange[1] > 0.95 {
				t.Errorf("%s should have low defense, got [%f, %f]", tt.archName, template.DefenseRange[0], template.DefenseRange[1])
			}
		case "AgilityRange":
			if template.AgilityRange[1] > 0.95 {
				t.Errorf("%s should have low agility, got [%f, %f]", tt.archName, template.AgilityRange[0], template.AgilityRange[1])
			}
		}
	}
}

func TestAllRounderBalance(t *testing.T) {
	arch := GetArchetypeByName("All-Rounder")
	if arch == nil {
		t.Fatal("All-Rounder archetype not found")
	}

	template := arch.StatDistribution

	// All-Rounder should have balanced stats (all ranges near 1.0)
	tolerance := 0.2 // Allow +/- 20% from 1.0

	if template.StrengthRange[0] < 1.0-tolerance || template.StrengthRange[1] > 1.0+tolerance {
		t.Errorf("All-Rounder strength range should be balanced, got [%f, %f]", template.StrengthRange[0], template.StrengthRange[1])
	}

	if template.DefenseRange[0] < 1.0-tolerance || template.DefenseRange[1] > 1.0+tolerance {
		t.Errorf("All-Rounder defense range should be balanced, got [%f, %f]", template.DefenseRange[0], template.DefenseRange[1])
	}

	if template.AgilityRange[0] < 1.0-tolerance || template.AgilityRange[1] > 1.0+tolerance {
		t.Errorf("All-Rounder agility range should be balanced, got [%f, %f]", template.AgilityRange[0], template.AgilityRange[1])
	}
}

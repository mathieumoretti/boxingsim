package seeding

import (
	"strings"
	"testing"
)

func TestNewNameGenerator(t *testing.T) {
	g := NewNameGenerator()

	if g == nil {
		t.Fatal("Expected non-nil NameGenerator")
	}

	if len(g.firstNamesMale) == 0 {
		t.Error("Expected non-empty male first names list")
	}

	if len(g.firstNamesFemale) == 0 {
		t.Error("Expected non-empty female first names list")
	}

	if len(g.lastNames) == 0 {
		t.Error("Expected non-empty last names list")
	}

	if len(g.nicknames) == 0 {
		t.Error("Expected non-empty nicknames list")
	}
}

func TestGenerateBoxerNameMale(t *testing.T) {
	g := NewNameGenerator()
	name := g.GenerateBoxerName("male", false)

	if name.FirstName == "" {
		t.Error("Expected non-empty first name")
	}

	if name.LastName == "" {
		t.Error("Expected non-empty last name")
	}

	if name.FullName == "" {
		t.Error("Expected non-empty full name")
	}

	// Check that full name contains first and last name
	if !strings.Contains(name.FullName, name.FirstName) || !strings.Contains(name.FullName, name.LastName) {
		t.Error("Full name should contain first and last name")
	}
}

func TestGenerateBoxerNameFemale(t *testing.T) {
	g := NewNameGenerator()
	name := g.GenerateBoxerName("female", false)

	if name.FirstName == "" {
		t.Error("Expected non-empty first name")
	}

	// Verify first name is from female list (simple check - not exhaustive)
	found := false
	for _, fn := range g.firstNamesFemale {
		if fn == name.FirstName {
			found = true
			break
		}
	}
	if !found {
		t.Error("First name should be from female names list")
	}
}

func TestGenerateBoxerNameWithNickname(t *testing.T) {
	g := NewNameGenerator()

	// Generate multiple times to find one with a nickname
	var hasNickname bool
	for i := 0; i < 20; i++ {
		name := g.GenerateBoxerName("male", true)
		if name.Nickname != nil {
			hasNickname = true
			break
		}
	}

	if !hasNickname {
		t.Error("Expected at least one boxer with nickname after 20 attempts (70% probability)")
	}
}

func TestGenerateBoxerNameFullNameFormat(t *testing.T) {
	g := NewNameGenerator()

	// Find a name with nickname to test format
	var testName BoxerName
	for i := 0; i < 20; i++ {
		testName = g.GenerateBoxerName("male", true)
		if testName.Nickname != nil {
			break
		}
	}

	if testName.Nickname != nil {
		// Full name should be: FirstName 'Nickname' LastName
		expectedFormat := testName.FirstName + " '" + *testName.Nickname + "' " + testName.LastName
		if testName.FullName != expectedFormat {
			t.Errorf("Full name format incorrect. Expected '%s', got '%s'", expectedFormat, testName.FullName)
		}

		// Display name should be: Nickname LastName
		expectedDisplay := *testName.Nickname + " " + testName.LastName
		if testName.DisplayName != expectedDisplay {
			t.Errorf("Display name format incorrect. Expected '%s', got '%s'", expectedDisplay, testName.DisplayName)
		}
	}
}

func TestNameGenerationDiversity(t *testing.T) {
	g := NewNameGenerator()

	// Generate 50 names and check for reasonable diversity
	names := make(map[string]bool)
	for i := 0; i < 50; i++ {
		name := g.GenerateBoxerName("male", false)
		names[name.FullName] = true
	}

	// Should have at least 40 unique names out of 50 (allowing some duplicates)
	if len(names) < 40 {
		t.Errorf("Expected at least 40 unique names, got %d", len(names))
	}
}

func TestGlobalRandInit(t *testing.T) {
	// Test that global rand is initialized and works
	val1 := randFloat64()
	val2 := randFloat64()

	if val1 < 0 || val1 > 1 {
		t.Errorf("randFloat64 returned value out of range: %f", val1)
	}

	if val2 < 0 || val2 > 1 {
		t.Errorf("randFloat64 returned value out of range: %f", val2)
	}

	val3 := randIntn(100)
	if val3 < 0 || val3 >= 100 {
		t.Errorf("randIntn(100) returned value out of range: %d", val3)
	}
}

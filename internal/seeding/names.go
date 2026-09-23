package seeding

import (
	"math/rand"
	"time"
)

// NameGenerator provides realistic boxing name generation with ethnic diversity.
type NameGenerator struct {
	firstNamesMale   []string
	firstNamesFemale []string
	lastNames        []string
	nicknames        []string
}

// NewNameGenerator creates a new NameGenerator with populated name lists.
func NewNameGenerator() *NameGenerator {
	return &NameGenerator{
		firstNamesMale: []string{
			// Classic American names
			"Marcus", "Michael", "Robert", "James", "John", "Joseph", "David",
			"Anthony", "Daniel", "Matthew", "Christopher", "Joshua", "Andrew",
			"Brandon", "Kevin", "Timothy", "Nathan", "Tyler", "Jeremy", "Jordan",
			// African American names
			"Tyrone", "Lamont", "DeAndre", "Corey", "Malik", "Terrell", "Darnell",
			"Gregory", "Melvin", "Calvin", "Reggie", "Marquis", "Shawn", "Kenneth",
			"Walter", "Rashad", "Antoine", "Demetrius", "Victor", "Adrian",
			// Latino/Hispanic names
			"Roberto", "Carlos", "Ricardo", "Miguel", "Jose", "Juan", "Luis",
			"Oscar", "Antonio", "Rafael", "Fernando", "Daniel", "Emanuel", "Jorge",
			"Gerard", "Sergio", "Angel", "Gilberto", "Salvador", "Hector",
			// European names
			"Bert", "Frank", "Steve", "Barry", "George", "Larry", "Ray",
			"Willy", "Henry", "Tommy", "Eddie", "Danny", "Ronnie", "Billy",
			"Kevin", "Richard", "Paul", "Chris", "Mark", "Brian",
			// Irish names
			"Ryan", "Patrick", "Sean", "Kevin", "Barry", "Dermot", "Conor",
			"Liam", "Fionn", "Colm", "Ciaran", "Gerard", "Michael", "Thomas",
			// Italian names
			"Lorenzo", "Giuseppe", "Francesco", "Antonio", "Marco", "Luigi", "Paolo",
			"Vincent", "Frankie", "Salvatore", "Gennaro", "Rocco", "Tony", "Vito",
			// Filipino names (boxing heritage)
			"Manny", "Nonito", "Juan", "Jessie", "Mark", "Christian", "Paul",
			"Edgar", "Marco", "Jerwin", "Jhon", "Sergio", "Reyson", "Stephen",
			// Additional realistic names
			"Emanuel", "Erislandy", "Vasiliy", "Oleksandr", "Canelo", "Gennady",
			"Artur", "Clemente", "Jermall", "Errol", "Shawn", "Keith", "Sergio",
			"Alexander", "Ryota", "Naoya", "Terence", "Craig", "Joseph", "Anthony",
			"Oleksandr", "Dmitry", "Mairis", "Audley", "Joe", "Chris", "Jorge",
			"Nicolai", "Wladimir", "Lamontro", "Ismael", "Abner", "Marvin", "David",
			"Benavidez", "Hearn", "Gervonta", "Davis", "Anthony", "Joshua", "Andy",
		},
		firstNamesFemale: []string{
			// Classic American names
			"Maria", "Sarah", "Jessica", "Ashley", "Jennifer", "Amanda", "Michelle",
			"Stephanie", "Nicole", "Elizabeth", "Melissa", "Rebecca", "Laura",
			"Kimberly", "Angela", "Heather", "Amber", "Crystal", "Tiffany", "Courtney",
			// Latino/Hispanic names
			"Sylvia", "Claressa", "Maricruz", "Katie", "Terrence", "Eden", "Bianca",
			"Cecilia", "Denisha", "Shannon", "Jenny", "Dana", "Christy", "Kelly",
			"Mary", "Laila", "Xenia", "Anastasia", "Hanna", "Nicolle",
			// European names
			"Laura", "Emma", "Katie", "Nicola", "Lauren", "Chantal", "Erika",
			"Sara", "Dana", "Catherine", "Lucie", "Jade", "Haley", "Christina",
			// Additional realistic names
			"Claressa", "Savannah", "Christina", "Kellie", "Erin", "Valerie", "Meagan",
			"Frida", "Teresa", "Daniela", "Geraldine", "Mary", "Linda", "Patricia",
		},
		lastNames: []string{
			// Common surnames
			"Jackson", "Williams", "Johnson", "Brown", "Davis", "Miller", "Wilson",
			"Morrison", "Taylor", "Anderson", "Thomas", "White", "Harris", "Martin",
			"Garcia", "Rodriguez", "Martinez", "Lopez", "Hernandez", "Gonzalez",
			"Perez", "Sanchez", "Ramirez", "Torres", "Flores", "Rivera", "Castillo",
			// Boxing family names
			"Tyson", "Ali", "Mayweather", "Pacquiao", "Foreman", "Frazier", "Hagler",
			"Hearns", "Leonard", "Duran", "De La Hoya", "Holyfield", "Toney", "LaMotta",
			"Canzoneri", "Sacco", "Conn", "Walcott", "Marciano", "Louis", "Braddock",
			"Dempsey", "Robinson", "Calzaghe", "Lewis", "McLellan", "Haye", "Bryan",
			"Irving", "Smith", "Jones", "Patterson", "Noble", "Gresley", "Williams",
			"Malignaggi", "Armstrong", "Bradley", "Spence", "Crolla", "Kane", "Kelly",
			"Díaz", "Barrera", "Zepeda", "Luna", "Magdaleno", "Soto", "Cruz",
			"Castro", "Guevara", "Ortiz", "Vasquez", "Rios", "Guerrero", "Camacho",
			"Barrera", "Kovalev", "Beterbiev", "Usyk", "Lomachenko", "Dyachenko",
			"Nashibullin", "Gerald", "Takasu", "Inoue", "Magomedov", "Sharratt",
		},
		nicknames: []string{
			// Classic nicknames
			"Iron", "The Beast", "The Destroyer", "The Killer", "The Assassin",
			"Money", "The Greatest", "The Phantom", "The Cobra", "The Hurricane",
			"The Mauler", "The Bomber", "The Hitman", "Marvelous", "The Golden Boy",
			"Hands of Stone", "High on Action", "Smiling", "Big", "Little", "Real",
			// Descriptive nicknames
			"Thunder", "Lightning", "Storm", "Fire", "Blaze", "Ice", "Stone",
			"Steel", "Rocky", "Wild", "Dangerous", "Deadly", "Silent", "Crazy",
			"Quick", "Fast", "Swift", "Rapid", "Express", "Turbo", "Power",
			// Animal nicknames
			"The Tiger", "The Lion", "The Eagle", "The Wolf", "The Snake",
			"The Shark", "The Bear", "The Bulldog", "The Ram", "The Hawk",
			"The Falcon", "The Viper", "The Cobra", "The Dragon", "The Panther",
			// Style nicknames
			"Clean Cut", "Pretty", "Savage", "Ruthless", "Merciless", "Brutal",
			"Vicious", "Fierce", "Tough", "Hard", "Solid", "Strong", "Mighty",
			// Location-based nicknames
			"Bronx", "Brooklyn", "Queens", "Chicago", "Detroit", "Houston",
			"Dallas", "Miami", "LA", "Texas", "New York", "Philadelphia",
			"Pittsburgh", "Cleveland", "Buffalo", "Boston", "Seattle",
			// Personality nicknames
			"Sweet", "Smiling", "Happy", "Cool", "Bad", "Good", "Righteous",
			"Holy", "Sainted", "Blessed", "Divine", "Royal", "King", "Prince",
			"Papa", "Big Poppa", "Young", "Junior", "Senior", "Professor", "Dr.",
			// Combined style nicknames (two-word)
			"Iron Man", "Thunder Bolt", "Storm Bringer", "Fire Storm", "Ice Cold",
			"Stone Face", "Steel Chin", "Rock Hard", "Wild Child", "Danger Zone",
			"Deadly Sin", "Silent Death", "Crazy Horse", "Quick Draw", "Fast Eddie",
			"Power House", "Knockout King", "Punch Out", "Ring Master", "Fight Night",
		},
	}
}

// BoxerName represents a generated boxer name with multiple formats.
type BoxerName struct {
	FirstName   string  // First name (e.g., "Marcus")
	LastName    string  // Last name (e.g., "Jackson")
	Nickname    *string // Nickname (e.g., "Iron") - can be nil
	FullName    string  // Full name for database (e.g., "Marcus 'Iron' Jackson")
	DisplayName string  // Display name (e.g., "Iron Jackson")
}

// GenerateBoxerName generates a random boxer name with optional nickname.
func (g *NameGenerator) GenerateBoxerName(gender string, includeNickname bool) BoxerName {
	firstName := g.firstNamesMale[randIntn(len(g.firstNamesMale))]
	if gender == "female" {
		firstName = g.firstNamesFemale[randIntn(len(g.firstNamesFemale))]
	}

	lastName := g.lastNames[randIntn(len(g.lastNames))]

	var nickname *string
	if includeNickname && randFloat64() < 0.7 { // 70% chance of nickname
		nick := g.nicknames[randIntn(len(g.nicknames))]
		nickname = &nick
	}

	// Build full name format
	fullName := firstName + " " + lastName
	displayName := firstName + " " + lastName

	if nickname != nil {
		fullName = firstName + " '" + *nickname + "' " + lastName
		displayName = *nickname + " " + lastName
	}

	return BoxerName{
		FirstName:   firstName,
		LastName:    lastName,
		Nickname:    nickname,
		FullName:    fullName,
		DisplayName: displayName,
	}
}

// Helper functions for randomness (using package-level rand)
var globalRand *rand.Rand

func init() {
	globalRand = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func randIntn(n int) int {
	return globalRand.Intn(n)
}

func randFloat64() float64 {
	return globalRand.Float64()
}

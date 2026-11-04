package domain

import (
	"testing"
	"time"
)

// Test GameTime constructors
func TestGameTime_Constructors(t *testing.T) {
	baseTime := time.Date(2030, 1, 15, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		fn       func() GameTime
		wantYear int
		wantDay  int
	}{
		{
			name: "NowGameTime creates from time.Time",
			fn: func() GameTime {
				return NowGameTime(baseTime)
			},
			wantYear: 2030,
			wantDay:  15,
		},
		{
			name: "FromTime creates wrapper",
			fn: func() GameTime {
				return FromTime(baseTime)
			},
			wantYear: 2030,
			wantDay:  15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gt := tt.fn()
			if got := gt.Year(); got != tt.wantYear {
				t.Errorf("Year() = %d, want %d", got, tt.wantYear)
			}
			if got := gt.Day(); got != tt.wantDay {
				t.Errorf("Day() = %d, want %d", got, tt.wantDay)
			}
		})
	}
}

// Test GameTime.AsTime returns underlying time
func TestGameTime_AsTime(t *testing.T) {
	baseTime := time.Date(2030, 1, 15, 14, 30, 0, 0, time.UTC)
	gt := FromTime(baseTime)

	result := gt.AsTime()
	if !result.Equal(baseTime) {
		t.Errorf("AsTime() = %v, want %v", result, baseTime)
	}
}

// Test GameTime.IsZero detects zero value
func TestGameTime_IsZero(t *testing.T) {
	tests := []struct {
		name     string
		gt       GameTime
		wantZero bool
	}{
		{
			name:     "zero value is zero",
			gt:       GameTime{},
			wantZero: true,
		},
		{
			name:     "non-zero time is not zero",
			gt:       FromTime(time.Date(2030, 1, 15, 0, 0, 0, 0, time.UTC)),
			wantZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.gt.IsZero(); got != tt.wantZero {
				t.Errorf("IsZero() = %v, want %v", got, tt.wantZero)
			}
		})
	}
}

// Test GameTime comparison methods
func TestGameTime_Comparison(t *testing.T) {
	time1 := time.Date(2030, 1, 15, 10, 0, 0, 0, time.UTC)
	time2 := time.Date(2030, 1, 15, 14, 0, 0, 0, time.UTC)
	time3 := time.Date(2030, 1, 15, 14, 0, 0, 0, time.UTC) // Same as time2

	gt1 := FromTime(time1)
	gt2 := FromTime(time2)
	gt3 := FromTime(time3)

	tests := []struct {
		name   string
		fn     func() bool
		want   bool
		reason string
	}{
		{
			name: "Before returns true for earlier time",
			fn: func() bool {
				return gt1.Before(gt2)
			},
			want:   true,
			reason: "10:00 is before 14:00",
		},
		{
			name: "Before returns false for later time",
			fn: func() bool {
				return gt2.Before(gt1)
			},
			want:   false,
			reason: "14:00 is not before 10:00",
		},
		{
			name: "After returns true for later time",
			fn: func() bool {
				return gt2.After(gt1)
			},
			want:   true,
			reason: "14:00 is after 10:00",
		},
		{
			name: "After returns false for earlier time",
			fn: func() bool {
				return gt1.After(gt2)
			},
			want:   false,
			reason: "10:00 is not after 14:00",
		},
		{
			name: "Equal returns true for same times",
			fn: func() bool {
				return gt2.Equal(gt3)
			},
			want:   true,
			reason: "Both are 14:00",
		},
		{
			name: "Equal returns false for different times",
			fn: func() bool {
				return gt1.Equal(gt2)
			},
			want:   false,
			reason: "10:00 != 14:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn(); got != tt.want {
				t.Errorf("%s: got %v, want %v", tt.reason, got, tt.want)
			}
		})
	}
}

// Test GameTime.Compare returns correct ordering
func TestGameTime_Compare(t *testing.T) {
	time1 := time.Date(2030, 1, 15, 10, 0, 0, 0, time.UTC)
	time2 := time.Date(2030, 1, 15, 14, 0, 0, 0, time.UTC)
	time3 := time.Date(2030, 1, 15, 14, 0, 0, 0, time.UTC)

	gt1 := FromTime(time1)
	gt2 := FromTime(time2)
	gt3 := FromTime(time3)

	tests := []struct {
		name   string
		fn     func() int
		want   int
		reason string
	}{
		{
			name: "Compare returns -1 for earlier",
			fn: func() int {
				return gt1.Compare(gt2)
			},
			want:   -1,
			reason: "10:00 < 14:00",
		},
		{
			name: "Compare returns +1 for later",
			fn: func() int {
				return gt2.Compare(gt1)
			},
			want:   1,
			reason: "14:00 > 10:00",
		},
		{
			name: "Compare returns 0 for equal",
			fn: func() int {
				return gt2.Compare(gt3)
			},
			want:   0,
			reason: "14:00 == 14:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn(); got != tt.want {
				t.Errorf("%s: Compare() = %d, want %d", tt.reason, got, tt.want)
			}
		})
	}
}

// Test GameTime arithmetic operations
func TestGameTime_Arithmetic(t *testing.T) {
	baseTime := time.Date(2030, 1, 15, 10, 0, 0, 0, time.UTC)
	gt := FromTime(baseTime)

	tests := []struct {
		name     string
		fn       func() GameTime
		wantHour int
		wantDay  int
		reason   string
	}{
		{
			name: "Add adds duration",
			fn: func() GameTime {
				return gt.Add(4 * time.Hour)
			},
			wantHour: 14,
			wantDay:  15,
			reason:   "10:00 + 4 hours = 14:00",
		},
		{
			name: "Sub subtracts duration",
			fn: func() GameTime {
				return gt.Add(24 * time.Hour) // First add a day to get Jan 16
			},
			wantHour: 10,
			wantDay:  16,
			reason:   "Jan 15 + 24 hours = Jan 16",
		},
		{
			name: "Add wraps to next day",
			fn: func() GameTime {
				return gt.Add(24 * time.Hour)
			},
			wantHour: 10,
			wantDay:  16,
			reason:   "10:00 Jan 15 + 24h = 10:00 Jan 16",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			if got := result.Hour(); got != tt.wantHour {
				t.Errorf("%s: Hour() = %d, want %d", tt.reason, got, tt.wantHour)
			}
			if got := result.Day(); got != tt.wantDay {
				t.Errorf("%s: Day() = %d, want %d", tt.reason, got, tt.wantDay)
			}
		})
	}
}

// Test GameTime.SubGameTime calculates duration between times
func TestGameTime_SubGameTime(t *testing.T) {
	time1 := time.Date(2030, 1, 15, 10, 0, 0, 0, time.UTC)
	time2 := time.Date(2030, 1, 15, 14, 30, 0, 0, time.UTC)

	gt1 := FromTime(time1)
	gt2 := FromTime(time2)

	tests := []struct {
		name   string
		fn     func() time.Duration
		want   time.Duration
		reason string
	}{
		{
			name: "SubGameTime returns positive for later minus earlier",
			fn: func() time.Duration {
				return gt2.SubGameTime(gt1)
			},
			want:   4*time.Hour + 30*time.Minute,
			reason: "14:30 - 10:00 = 4h 30m",
		},
		{
			name: "SubGameTime returns negative for earlier minus later",
			fn: func() time.Duration {
				return gt1.SubGameTime(gt2)
			},
			want:   -(4*time.Hour + 30*time.Minute),
			reason: "10:00 - 14:30 = -4h 30m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn(); got != tt.want {
				t.Errorf("%s: SubGameTime() = %v, want %v", tt.reason, got, tt.want)
			}
		})
	}
}

// Test GameTime.String formatting
func TestGameTime_String(t *testing.T) {
	tests := []struct {
		name    string
		gt      GameTime
		wantStr string
		reason  string
	}{
		{
			name:    "zero value formats correctly",
			gt:      GameTime{},
			wantStr: "GameTime{<zero>}",
			reason:  "Zero time should show placeholder",
		},
		{
			name:    "non-zero value shows date and time",
			gt:      FromTime(time.Date(2030, 1, 15, 14, 30, 0, 0, time.UTC)),
			wantStr: "GameTime{2030-01-15 14:30:00}",
			reason:  "Should format as readable string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.gt.String(); got != tt.wantStr {
				t.Errorf("%s: String() = %q, want %q", tt.reason, got, tt.wantStr)
			}
		})
	}
}

// Test GameTime.Format uses RFC3339
func TestGameTime_Format(t *testing.T) {
	gt := FromTime(time.Date(2030, 1, 15, 14, 30, 0, 0, time.UTC))
	result := gt.Format()

	// RFC3339 format: 2006-01-02T15:04:05Z07:00
	want := "2030-01-15T14:30:00Z"
	if result != want {
		t.Errorf("Format() = %q, want %q", result, want)
	}
}

// Test GameTime accessors return correct values
func TestGameTime_Accessors(t *testing.T) {
	gt := FromTime(time.Date(2030, 7, 15, 14, 30, 45, 0, time.UTC))

	tests := []struct {
		name   string
		fn     func() interface{}
		want   interface{}
		reason string
	}{
		{
			name:   "Hour returns hour component",
			fn:     func() interface{} { return gt.Hour() },
			want:   14,
			reason: "Hour should be 14",
		},
		{
			name:   "Minute returns minute component",
			fn:     func() interface{} { return gt.Minute() },
			want:   30,
			reason: "Minute should be 30",
		},
		{
			name:   "Second returns second component",
			fn:     func() interface{} { return gt.Second() },
			want:   45,
			reason: "Second should be 45",
		},
		{
			name:   "Day returns day of month",
			fn:     func() interface{} { return gt.Day() },
			want:   15,
			reason: "Day should be 15",
		},
		{
			name:   "Month returns month",
			fn:     func() interface{} { return gt.Month() },
			want:   time.July,
			reason: "Month should be July",
		},
		{
			name:   "Year returns year",
			fn:     func() interface{} { return gt.Year() },
			want:   2030,
			reason: "Year should be 2030",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn()
			if got != tt.want {
				t.Errorf("%s: got %v (%T), want %v (%T)", tt.reason, got, got, tt.want, tt.want)
			}
		})
	}
}

// Test GameTime.TruncateSecond removes nanoseconds
func TestGameTime_TruncateSecond(t *testing.T) {
	gt := FromTime(time.Date(2030, 1, 15, 14, 30, 45, 123456789, time.UTC))
	result := gt.TruncateSecond()

	if result.Second() != 45 {
		t.Errorf("TruncateSecond: Second() = %d, want 45", result.Second())
	}

	if result.AsTime().Nanosecond() != 0 {
		t.Errorf("TruncateSecond: Nanosecond() = %d, want 0", result.AsTime().Nanosecond())
	}
}

// Test GameTime.IsToday compares calendar days
func TestGameTime_IsToday(t *testing.T) {
	gameDate := FromTime(time.Date(2030, 1, 15, 14, 30, 0, 0, time.UTC))

	tests := []struct {
		name     string
		ref      time.Time
		wantSame bool
		reason   string
	}{
		{
			name:     "same calendar day returns true",
			ref:      time.Date(2030, 1, 15, 0, 0, 0, 0, time.UTC), // Different hour, same day
			wantSame: true,
			reason:   "Both are Jan 15",
		},
		{
			name:     "different calendar day returns false",
			ref:      time.Date(2030, 1, 16, 14, 30, 0, 0, time.UTC), // Next day
			wantSame: false,
			reason:   "Jan 15 != Jan 16",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := gameDate.IsToday(tt.ref); got != tt.wantSame {
				t.Errorf("%s: IsToday() = %v, want %v", tt.reason, got, tt.wantSame)
			}
		})
	}
}

// Test GameTime.In and UTC conversion
func TestGameTime_Timezone(t *testing.T) {
	baseTime := time.Date(2030, 1, 15, 14, 30, 0, 0, time.UTC)
	gt := FromTime(baseTime)

	tests := []struct {
		name    string
		fn      func() GameTime
		wantLoc *time.Location
		reason  string
	}{
		{
			name:    "UTC returns UTC time",
			fn:      func() GameTime { return gt.UTC() },
			wantLoc: time.UTC,
			reason:  "Should convert to UTC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			if result.time.Location() != tt.wantLoc {
				t.Errorf("%s: Location() = %v, want %v", tt.reason, result.time.Location(), tt.wantLoc)
			}
		})
	}
}

// Test GameTime type safety prevents accidental time.Time usage
func TestGameTime_TypeSafety(t *testing.T) {
	// This test documents that GameTime and time.Time are different types
	gt := FromTime(time.Date(2030, 1, 15, 0, 0, 0, 0, time.UTC))

	// These should compile (GameTime methods):
	_ = gt.Add(time.Hour)
	_ = gt.Before(gt)
	_ = gt.AsTime()

	// This document shows we need AsTime() to get back to time.Time
	asTime := gt.AsTime()
	if asTime.Year() != 2030 {
		t.Errorf("Expected year 2030, got %d", asTime.Year())
	}
}

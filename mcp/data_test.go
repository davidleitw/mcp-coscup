package mcp

import (
	"testing"
	"mcp-coscup/mcp/testutil"
)

// Tests for functions in data.go

func TestTimeToMinutes(t *testing.T) {
	tests := []struct {
		name     string
		timeStr  string
		expected int
	}{
		{"Morning time", "09:30", 570},
		{"Noon", "12:00", 720},
		{"Afternoon", "14:45", 885},
		{"Evening", "18:30", 1110},
		{"Midnight", "00:00", 0},
		{"Late night", "23:59", 1439},
		{"Invalid format", "25:70", 0},
		{"Empty string", "", 0},
		{"Single digit", "9:5", 545},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := timeToMinutes(tt.timeStr)
			testutil.AssertEqual(t, tt.expected, result, "timeToMinutes result")
		})
	}
}

func TestHasTimeConflict(t *testing.T) {
	tests := []struct {
		name     string
		start1   string
		end1     string
		start2   string
		end2     string
		expected bool
	}{
		{"No conflict - session1 before session2", "09:00", "10:00", "11:00", "12:00", false},
		{"No conflict - session2 before session1", "11:00", "12:00", "09:00", "10:00", false},
		{"Complete overlap - session1 contains session2", "09:00", "12:00", "10:00", "11:00", true},
		{"Complete overlap - session2 contains session1", "10:00", "11:00", "09:00", "12:00", true},
		{"Partial overlap - sessions intersect", "09:00", "10:30", "10:00", "11:30", true},
		{"Adjacent sessions - no overlap", "09:00", "10:00", "10:00", "11:00", false},
		{"Same time periods", "09:00", "10:00", "09:00", "10:00", true},
		{"One minute overlap", "09:00", "10:01", "10:00", "11:00", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasTimeConflict(tt.start1, tt.end1, tt.start2, tt.end2)
			testutil.AssertEqual(t, tt.expected, result, "hasTimeConflict result")
		})
	}
}

func TestIsValidDay(t *testing.T) {
	tests := []struct {
		name     string
		day      string
		expected bool
	}{
		{"Valid Aug9", "Aug9", true},
		{"Valid Aug10", "Aug10", true},
		{"Invalid day", "Aug11", false},
		{"Invalid format", "August9", false},
		{"Empty string", "", false},
		{"Lowercase", "aug9", false},
		{"Wrong format", "9Aug", false},
		{"Random string", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidDay(tt.day)
			testutil.AssertEqual(t, tt.expected, result, "IsValidDay result")
		})
	}
}

func TestConvertDayFormat(t *testing.T) {
	tests := []struct {
		name     string
		userDay  string
		expected string
	}{
		{"Aug9 to Aug.9", "Aug9", "Aug.9"},
		{"Aug10 to Aug.10", "Aug10", "Aug.10"},
		{"Invalid input", "Aug11", "Aug11"},
		{"Empty string", "", ""},
		{"Already formatted", "Aug.9", "Aug.9"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertDayFormat(tt.userDay)
			testutil.AssertEqual(t, tt.expected, result, "convertDayFormat result")
		})
	}
}
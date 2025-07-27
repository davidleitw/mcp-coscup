package mcp

import (
	"strconv"
	"strings"
)

// Session represents a COSCUP session
type Session struct {
	Code       string   `json:"code"`
	Title      string   `json:"title"`
	Speakers   []string `json:"speakers"`
	Start      string   `json:"start"`
	End        string   `json:"end"`
	Track      string   `json:"track"`
	Abstract   string   `json:"abstract"`
	Language   string   `json:"language"`
	Difficulty string   `json:"difficulty"`
	Room       string   // derived from JSON structure
	Day        string   // "Aug.9" or "Aug.10"
	URL        string   `json:"url"`  // Official COSCUP session URL
	Tags       []string `json:"tags"` // Universal tags for categorization
}

// Global data storage - initialized at package load time
var (
	allSessions   []Session
	sessionsByDay = make(map[string][]Session)
)

// init initializes COSCUP session data from embedded data
// This happens automatically when the package is loaded
func init() {
	// Process embedded data from embedded_data.go
	for day, rooms := range COSCUPData {
		for _, sessions := range rooms {
			for _, session := range sessions {
				// Add official COSCUP URL
				session.URL = "https://coscup.org/2025/sessions/" + session.Code

				// Tags are already defined in embedded_data.go
				// No need to generate tags - they come from the embedded data

				allSessions = append(allSessions, session)
				sessionsByDay[day] = append(sessionsByDay[day], session)
			}
		}
	}
}

// FindSessionByCode finds a session by its code
// Returns a safe copy since allSessions is global data - preserves complete abstract for detailed view
func FindSessionByCode(code string) *Session {
	for _, session := range allSessions {
		if session.Code == code {
			// Return a copy to protect global data while preserving complete abstract
			result := session
			return &result
		}
	}
	return nil
}

// GetFirstSession returns the first session of the day (usually Welcome)
func GetFirstSession(day string) []Session {
	sessions := sessionsByDay[day]
	if len(sessions) == 0 {
		return nil
	}

	// Find the earliest start time
	earliest := "23:59"
	for _, session := range sessions {
		if session.Start < earliest {
			earliest = session.Start
		}
	}

	// Find all sessions that start at the earliest time
	var earliestSessions []Session
	for _, session := range sessions {
		if session.Start == earliest {
			earliestSessions = append(earliestSessions, session)
		}
	}

	return getSimplifiedSessions(earliestSessions)
}

// timeToMinutes converts "HH:MM" to minutes since midnight
func timeToMinutes(timeStr string) int {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0
	}

	hours, err1 := strconv.Atoi(parts[0])
	minutes, err2 := strconv.Atoi(parts[1])

	// Validate input range
	if err1 != nil || err2 != nil || hours < 0 || hours > 23 || minutes < 0 || minutes > 59 {
		return 0
	}

	return hours*60 + minutes
}

// IsValidDay checks if the given day is valid
func IsValidDay(day string) bool {
	return day == DayAug9 || day == DayAug10
}

// convertDayFormat converts user input format to internal format
func convertDayFormat(userDay string) string {
	switch userDay {
	case DayAug9:
		return DayFormatAug9
	case DayAug10:
		return DayFormatAug10
	default:
		return userDay
	}
}

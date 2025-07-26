package mcp

import (
	"strconv"
	"strings"
)

// Session represents a COSCUP session
type Session struct {
	Code     string   `json:"code"`
	Title    string   `json:"title"`
	Speakers []string `json:"speakers"`
	Start    string   `json:"start"`
	End      string   `json:"end"`
	Track    string   `json:"track"`
	Abstract string   `json:"abstract"`
	Language string   `json:"language"`
	Difficulty string `json:"difficulty"`
	Room     string   // derived from JSON structure
	Day      string   // "Aug.9" or "Aug.10"
	URL      string   `json:"url"` // Official COSCUP session URL
	Tags     []string `json:"tags"` // Universal tags for categorization
}


// Global data storage
var (
	allSessions     []Session
	sessionsByDay   = make(map[string][]Session)
	sessionsLoaded  = false
)

// LoadCOSCUPData loads embedded COSCUP data
func LoadCOSCUPData() error {
	if sessionsLoaded {
		return nil
	}

	// Use embedded data from embedded_data.go
	for day, rooms := range EmbeddedCOSCUPData {
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

	sessionsLoaded = true
	return nil
}


// FindSessionByCode finds a session by its code
func FindSessionByCode(code string) *Session {
	if !sessionsLoaded {
		if err := LoadCOSCUPData(); err != nil {
			return nil
		}
	}

	for _, session := range allSessions {
		if session.Code == code {
			return &session
		}
	}
	return nil
}

// GetFirstSession returns the first session of the day (usually Welcome)
func GetFirstSession(day string) []Session {
	if !sessionsLoaded {
		if err := LoadCOSCUPData(); err != nil {
			return nil
		}
	}

	sessions := sessionsByDay[day]
	if len(sessions) == 0 {
		return nil
	}

	// Find the earliest start time
	earliestTime := "23:59"
	for _, session := range sessions {
		if session.Start < earliestTime {
			earliestTime = session.Start
		}
	}

	// Return all sessions that start at the earliest time
	var result []Session
	for _, session := range sessions {
		if session.Start == earliestTime {
			result = append(result, session)
		}
	}

	return result
}

// timeToMinutes converts "HH:MM" to minutes since midnight
func timeToMinutes(timeStr string) int {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0
	}

	hours, _ := strconv.Atoi(parts[0])
	minutes, _ := strconv.Atoi(parts[1])
	return hours*60 + minutes
}

// IsValidDay checks if the given day is valid
func IsValidDay(day string) bool {
	return day == "Aug9" || day == "Aug10"
}

// convertDayFormat converts user input format to internal format
func convertDayFormat(userDay string) string {
	switch userDay {
	case "Aug9":
		return "Aug.9"
	case "Aug10":
		return "Aug.10"
	default:
		return userDay
	}
}



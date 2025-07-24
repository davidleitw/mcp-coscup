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

	// Use embedded data instead of reading from file
	for day, rooms := range EmbeddedCOSCUPData {
		for _, sessions := range rooms {
			for _, session := range sessions {
				allSessions = append(allSessions, session)
				sessionsByDay[day] = append(sessionsByDay[day], session)
			}
		}
	}

	sessionsLoaded = true
	return nil
}

// FindSessionsAfter returns sessions that start after the given time on the given day
func FindSessionsAfter(day, afterTime string) []Session {
	if !sessionsLoaded {
		if err := LoadCOSCUPData(); err != nil {
			return nil
		}
	}

	var result []Session
	afterMinutes := timeToMinutes(afterTime)

	for _, session := range sessionsByDay[day] {
		startMinutes := timeToMinutes(session.Start)
		if startMinutes >= afterMinutes {
			result = append(result, session)
		}
	}

	return result
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
	return day == "Aug.9" || day == "Aug.10"
}

// GetAllTracks returns all unique tracks
func GetAllTracks() []string {
	if !sessionsLoaded {
		if err := LoadCOSCUPData(); err != nil {
			return nil
		}
	}

	trackSet := make(map[string]bool)
	for _, session := range allSessions {
		trackSet[session.Track] = true
	}

	var tracks []string
	for track := range trackSet {
		tracks = append(tracks, track)
	}
	return tracks
}
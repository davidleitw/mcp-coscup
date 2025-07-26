package mcp

import (
	"fmt"
	"sync"
	"time"
)

// UserState represents the planning state for a user session
type UserState struct {
	SessionID    string    `json:"session_id"`
	Day          string    `json:"day"`           // "Aug.9" or "Aug.10"
	Schedule     []Session `json:"schedule"`      // selected sessions
	LastEndTime  string    `json:"last_end_time"` // end time of last selected session
	Profile      []string  `json:"profile"`       // interested tracks
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
}

// Response represents the standard MCP tool response
type Response struct {
	Success    bool   `json:"success"`
	Data       any    `json:"data"`
	CallReason string `json:"call_reason"`
	Message    string `json:"message"`
}

// Global in-memory storage with thread safety
var (
	userStates = make(map[string]*UserState)
	stateMutex = sync.RWMutex{}
)

// CreateUserState creates a new user planning session
func CreateUserState(sessionID, day string) *UserState {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	state := &UserState{
		SessionID:    sessionID,
		Day:          day,
		Schedule:     make([]Session, 0),
		LastEndTime:  "08:00", // start from early morning
		Profile:      make([]string, 0),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	userStates[sessionID] = state
	return state
}

// GetUserState retrieves user state by session ID
func GetUserState(sessionID string) *UserState {
	stateMutex.RLock()
	defer stateMutex.RUnlock()

	if state, exists := userStates[sessionID]; exists {
		// Update last activity
		state.LastActivity = time.Now()
		return state
	}
	return nil
}

// UpdateUserState updates the user state
func UpdateUserState(sessionID string, updater func(*UserState)) error {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	state, exists := userStates[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	updater(state)
	state.LastActivity = time.Now()
	return nil
}

// AddSessionToSchedule adds a selected session to user's schedule
func AddSessionToSchedule(sessionID, sessionCode string) error {
	session := FindSessionByCode(sessionCode)
	if session == nil {
		return fmt.Errorf("session %s not found", sessionCode)
	}

	return UpdateUserState(sessionID, func(state *UserState) {
		// Add to schedule
		state.Schedule = append(state.Schedule, *session)

		// Update last end time
		state.LastEndTime = session.End

		// Update profile based on the selected track
		addToProfile(state, session.Track)
	})
}

// addToProfile adds a track to user's profile if not already present
func addToProfile(state *UserState, track string) {
	for _, existing := range state.Profile {
		if existing == track {
			return // already in profile
		}
	}
	state.Profile = append(state.Profile, track)
}

// FindNextAvailableInEachRoom finds next available session in each room after given time
func FindNextAvailableInEachRoom(day, afterTime string, userSchedule []Session) []Session {
	if !sessionsLoaded {
		if err := LoadCOSCUPData(); err != nil {
			return nil
		}
	}

	// Group sessions by room
	roomSessions := make(map[string][]Session)
	for _, session := range sessionsByDay[day] {
		roomSessions[session.Room] = append(roomSessions[session.Room], session)
	}

	var nextSessions []Session
	afterMinutes := timeToMinutes(afterTime)

	// Find next available session in each room
	for _, sessions := range roomSessions {
		
		// Sort sessions in this room by start time
		roomSessionsSorted := make([]Session, len(sessions))
		copy(roomSessionsSorted, sessions)
		
		// Simple bubble sort by start time
		for i := 0; i < len(roomSessionsSorted); i++ {
			for j := i + 1; j < len(roomSessionsSorted); j++ {
				if timeToMinutes(roomSessionsSorted[i].Start) > timeToMinutes(roomSessionsSorted[j].Start) {
					roomSessionsSorted[i], roomSessionsSorted[j] = roomSessionsSorted[j], roomSessionsSorted[i]
				}
			}
		}

		// Find the first available session in this room
		for _, session := range roomSessionsSorted {
			startMinutes := timeToMinutes(session.Start)
			
			// Must start after afterTime
			if startMinutes >= afterMinutes {
				// Check if it conflicts with user schedule
				if !hasConflictWithSchedule(session, userSchedule) {
					nextSessions = append(nextSessions, session)
					break // Found the next available session for this room
				}
				// If it conflicts, continue to check the next session in this room
			}
		}
	}
	
	return nextSessions
}

// hasConflictWithSchedule checks if session conflicts with user's existing schedule
func hasConflictWithSchedule(session Session, userSchedule []Session) bool {
	for _, scheduled := range userSchedule {
		if hasTimeConflict(session.Start, session.End, scheduled.Start, scheduled.End) {
			return true
		}
	}
	return false
}

// hasTimeConflict checks if two time periods overlap
func hasTimeConflict(start1, end1, start2, end2 string) bool {
	start1Min := timeToMinutes(start1)
	end1Min := timeToMinutes(end1)
	start2Min := timeToMinutes(start2)
	end2Min := timeToMinutes(end2)
	
	// Two time periods overlap if:
	// session1 start < session2 end && session1 end > session2 start
	return start1Min < end2Min && end1Min > start2Min
}

// GetRecommendations returns recommended sessions for the user using new room-based logic
func GetRecommendations(sessionID string, limit int) ([]Session, error) {
	state := GetUserState(sessionID)
	if state == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	// Use new room-based logic to find next available sessions
	nextSessions := FindNextAvailableInEachRoom(state.Day, state.LastEndTime, state.Schedule)
	
	// Return all room sessions (no artificial limit)
	// Let LLM handle the sorting and presentation based on user preferences
	return nextSessions, nil
}

// ScoredSession represents a session with recommendation score
type ScoredSession struct {
	Session Session
	Score   float64
}

// calculateScore calculates recommendation score for a session
func calculateScore(session Session, profile []string) float64 {
	score := 1.0 // base score

	// Boost score if track matches user profile
	for _, track := range profile {
		if session.Track == track {
			score += 2.0
			break
		}
	}

	// Boost for Chinese language (assuming most users prefer Chinese)
	if session.Language == "漢語" {
		score += 0.5
	}

	// Small boost for beginner/intermediate level
	if session.Difficulty == "入門" || session.Difficulty == "中階" {
		score += 0.3
	}

	return score
}

// CleanupOldSessions removes sessions older than 24 hours
func CleanupOldSessions() {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour)
	for sessionID, state := range userStates {
		if state.LastActivity.Before(cutoff) {
			delete(userStates, sessionID)
		}
	}
}

// GetSessionStats returns basic statistics about active sessions
func GetSessionStats() map[string]any {
	stateMutex.RLock()
	defer stateMutex.RUnlock()

	return map[string]any{
		"active_sessions": len(userStates),
		"timestamp":       time.Now().Format(time.RFC3339),
	}
}

// IsScheduleComplete checks if the user has planned the full day
func IsScheduleComplete(sessionID string) bool {
	state := GetUserState(sessionID)
	if state == nil {
		return false
	}

	// Consider complete if last end time is after 17:00 (5 PM)
	lastEndMinutes := timeToMinutes(state.LastEndTime)
	return lastEndMinutes >= 17*60 // 17:00 = 17*60 minutes
}

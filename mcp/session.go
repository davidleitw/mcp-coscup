package mcp

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"log"
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

// Simple sharded storage with 16 shards for better concurrency
const NumShards = 16

type SessionShard struct {
	mu       sync.RWMutex
	sessions map[string]*UserState
}

// Global sharded storage
var sessionShards [NumShards]*SessionShard

func init() {
	// Initialize all shards
	for i := 0; i < NumShards; i++ {
		sessionShards[i] = &SessionShard{
			sessions: make(map[string]*UserState),
		}
	}
}

// getShardIndex returns which shard a sessionID should go to
func getShardIndex(sessionID string) int {
	hash := fnv.New32a()
	hash.Write([]byte(sessionID))
	return int(hash.Sum32() % NumShards)
}

// GenerateSecureSessionID creates a cryptographically secure session ID
func GenerateSecureSessionID(day string) string {
	// Generate 8 random bytes
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails
		return fmt.Sprintf("user_%s_%d_fallback", day, time.Now().UnixNano())
	}
	
	// Create a unique ID with timestamp and random component
	timestamp := time.Now().Unix()
	randomHex := hex.EncodeToString(randomBytes)
	
	return fmt.Sprintf("user_%s_%d_%s", day, timestamp, randomHex)
}

// GenerateSessionIDWithCollisionCheck generates a session ID and ensures it's unique
func GenerateSessionIDWithCollisionCheck(day string) string {
	maxAttempts := 10
	for attempt := 0; attempt < maxAttempts; attempt++ {
		sessionID := GenerateSecureSessionID(day)
		
		// Check if this ID already exists in the appropriate shard
		shardIndex := getShardIndex(sessionID)
		shard := sessionShards[shardIndex]
		
		shard.mu.RLock()
		_, exists := shard.sessions[sessionID]
		shard.mu.RUnlock()
		
		if !exists {
			return sessionID
		}
		
		// If collision occurs, wait a tiny bit and try again
		time.Sleep(1 * time.Millisecond)
	}
	
	// Ultimate fallback - use nano timestamp
	return fmt.Sprintf("user_%s_%d_ultimate", day, time.Now().UnixNano())
}

// CreateUserState creates a new user planning session
func CreateUserState(sessionID, day string) *UserState {
	shardIndex := getShardIndex(sessionID)
	shard := sessionShards[shardIndex]
	
	shard.mu.Lock()
	defer shard.mu.Unlock()

	state := &UserState{
		SessionID:    sessionID,
		Day:          day,
		Schedule:     make([]Session, 0),
		LastEndTime:  "08:00", // start from early morning
		Profile:      make([]string, 0),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	shard.sessions[sessionID] = state
	log.Printf("🆕 [%s] Created new user session for day %s (Shard: %d)", 
		sessionID, day, shardIndex)
	return state
}

// GetUserState retrieves user state by session ID
func GetUserState(sessionID string) *UserState {
	shardIndex := getShardIndex(sessionID)
	shard := sessionShards[shardIndex]
	
	shard.mu.RLock()
	defer shard.mu.RUnlock()

	if state, exists := shard.sessions[sessionID]; exists {
		// Update last activity
		state.LastActivity = time.Now()
		log.Printf("🔍 [%s] Session accessed, last activity updated", sessionID)
		return state
	}
	log.Printf("❌ [%s] Session not found", sessionID)
	return nil
}

// UpdateUserState updates the user state
func UpdateUserState(sessionID string, updater func(*UserState)) error {
	shardIndex := getShardIndex(sessionID)
	shard := sessionShards[shardIndex]
	
	shard.mu.Lock()
	defer shard.mu.Unlock()

	state, exists := shard.sessions[sessionID]
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
		log.Printf("❌ [%s] Failed to add session %s - session not found", sessionID, sessionCode)
		return fmt.Errorf("session %s not found", sessionCode)
	}

	log.Printf("➕ [%s] Adding session %s (%s) to schedule", sessionID, sessionCode, session.Title)

	return UpdateUserState(sessionID, func(state *UserState) {
		// Add to schedule
		state.Schedule = append(state.Schedule, *session)

		// Update last end time
		state.LastEndTime = session.End

		// Update profile based on the selected track
		addToProfile(state, session.Track)
		
		log.Printf("✅ [%s] Session added successfully. Schedule size: %d, End time: %s", 
			sessionID, len(state.Schedule), session.End)
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


// CleanupOldSessions removes sessions older than 24 hours (parallel cleanup)
func CleanupOldSessions() {
	cutoff := time.Now().Add(-24 * time.Hour)
	totalCleaned := 0
	
	// Clean each shard in parallel
	var wg sync.WaitGroup
	cleanedCounts := make([]int, NumShards)
	
	for i := 0; i < NumShards; i++ {
		wg.Add(1)
		go func(shardIndex int) {
			defer wg.Done()
			
			shard := sessionShards[shardIndex]
			shard.mu.Lock()
			defer shard.mu.Unlock()
			
			cleaned := 0
			for sessionID, state := range shard.sessions {
				if state.LastActivity.Before(cutoff) {
					log.Printf("🗑️  [%s] Cleaning up expired session (inactive since %v)", 
						sessionID, state.LastActivity.Format("2006-01-02 15:04:05"))
					delete(shard.sessions, sessionID)
					cleaned++
				}
			}
			cleanedCounts[shardIndex] = cleaned
		}(i)
	}
	
	wg.Wait()
	
	// Sum up cleaned sessions
	for _, count := range cleanedCounts {
		totalCleaned += count
	}
	
	if totalCleaned > 0 {
		activeCount := 0
		for i := 0; i < NumShards; i++ {
			shard := sessionShards[i]
			shard.mu.RLock()
			activeCount += len(shard.sessions)
			shard.mu.RUnlock()
		}
		log.Printf("🧹 Cleaned up %d expired sessions, %d sessions remain active", totalCleaned, activeCount)
	}
}

// GetSessionStats returns basic statistics about active sessions
func GetSessionStats() map[string]any {
	totalSessions := 0
	shardStats := make([]int, NumShards)
	
	for i := 0; i < NumShards; i++ {
		shard := sessionShards[i]
		shard.mu.RLock()
		count := len(shard.sessions)
		shard.mu.RUnlock()
		
		shardStats[i] = count
		totalSessions += count
	}

	return map[string]any{
		"active_sessions": totalSessions,
		"shard_stats":     shardStats,
		"num_shards":      NumShards,
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

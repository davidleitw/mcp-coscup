package mcp

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"log"
	"slices"
	"sort"
	"strings"
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
	IsCompleted  bool      `json:"is_completed"`  // user manually finished planning
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
}

// Response represents the standard MCP tool response
type Response struct {
	Success bool   `json:"success"`
	Data    any    `json:"data"`
	Message string `json:"message"`
}

// buildStandardResponse creates a standardized response with sessionID always included
func buildStandardResponse(sessionID string, data map[string]any, message string) Response {
	// Ensure sessionId is always in the response
	if data == nil {
		data = make(map[string]any)
	}
	data["sessionId"] = sessionID

	return Response{
		Success: true,
		Data:    data,
		Message: message,
	}
}

// Simple sharded storage for better concurrency
const NumShards = DefaultNumShards

type SessionShard struct {
	mu       sync.RWMutex
	sessions map[string]*UserState
}

// Global sharded storage
var sessionShards [NumShards]*SessionShard

func init() {
	// Initialize all shards
	for i := range NumShards {
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
	for range maxAttempts {
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
		IsCompleted:  false, // planning not finished yet
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
		log.Printf("[%s] Session accessed, last activity updated", sessionID)
		return state
	}
	log.Printf("[%s] Session not found", sessionID)
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
		log.Printf("[%s] Failed to add session %s - session not found", sessionID, sessionCode)
		return fmt.Errorf("session %s not found", sessionCode)
	}

	// Get current user state to check for conflicts
	state := GetUserState(sessionID)
	if state == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	// Check for time conflicts with existing schedule
	if hasConflictWithSchedule(*session, state.Schedule) {
		// Find the conflicting session(s)
		conflictingSessions := findConflictingSessions(*session, state.Schedule)
		conflictList := ""
		for i, conflict := range conflictingSessions {
			if i > 0 {
				conflictList += ", "
			}
			conflictList += fmt.Sprintf("%s-%s %s", conflict.Start, conflict.End, conflict.Title)
		}

		log.Printf("[%s] Time conflict detected for session %s (%s-%s)",
			sessionID, sessionCode, session.Start, session.End)
		return fmt.Errorf("時間衝突：您選擇的議程 %s-%s「%s」與已安排的議程重疊：%s。請選擇其他時段的議程",
			session.Start, session.End, session.Title, conflictList)
	}

	log.Printf("[%s] Adding session %s (%s) to schedule", sessionID, sessionCode, session.Title)

	return UpdateUserState(sessionID, func(state *UserState) {
		// Add to schedule
		state.Schedule = append(state.Schedule, *session)

		// Update last end time (only if this session ends later)
		if timeToMinutes(session.End) > timeToMinutes(state.LastEndTime) {
			state.LastEndTime = session.End
		}

		// Update profile based on the selected track
		addToProfile(state, session.Track)

		log.Printf("[%s] Session added successfully. Schedule size: %d, End time: %s",
			sessionID, len(state.Schedule), session.End)
	})
}

// addToProfile adds a track to user's profile if not already present
func addToProfile(state *UserState, track string) {
	if slices.Contains(state.Profile, track) {
		return // already in profile
	}
	state.Profile = append(state.Profile, track)
}

// sortSessionsByStartTime sorts sessions by start time using efficient sort.Slice
func sortSessionsByStartTime(sessions []Session) {
	sort.Slice(sessions, func(i, j int) bool {
		return timeToMinutes(sessions[i].Start) < timeToMinutes(sessions[j].Start)
	})
}

// getSimplifiedSessions creates safe copies of sessions and clears fields not needed for list display
func getSimplifiedSessions(sessions []Session) []Session {
	// Create safe copies since sessionsByDay is global data - avoid modifying original sessions
	result := make([]Session, len(sessions))
	for i, session := range sessions {
		result[i] = session
		result[i].Abstract = ""   // Clear abstract to reduce response size
		result[i].Difficulty = "" // Clear difficulty to reduce response size
	}
	return result
}

// FinishPlanning marks user's planning as completed
func FinishPlanning(sessionID string) error {
	return UpdateUserState(sessionID, func(state *UserState) {
		state.IsCompleted = true
		log.Printf("[%s] User manually finished planning with %d sessions",
			sessionID, len(state.Schedule))
	})
}

// FindNextAvailableInEachRoom finds next available session in each room after given time
func FindNextAvailableInEachRoom(day, afterTime string, userSchedule []Session) []Session {

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
		sortSessionsByStartTime(roomSessionsSorted)

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

	return getSimplifiedSessions(nextSessions)
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

// findConflictingSessions returns all sessions that conflict with the given session
func findConflictingSessions(session Session, userSchedule []Session) []Session {
	var conflicts []Session
	for _, scheduled := range userSchedule {
		if hasTimeConflict(session.Start, session.End, scheduled.Start, scheduled.End) {
			conflicts = append(conflicts, scheduled)
		}
	}
	return conflicts
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
func GetRecommendations(sessionID string) ([]Session, error) {
	state := GetUserState(sessionID)
	if state == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	// Use new room-based logic to find next available sessions
	nextSessions := FindNextAvailableInEachRoom(state.Day, state.LastEndTime, state.Schedule)

	// Filter out long-duration social activities (Hacking Corner, etc.)
	filteredSessions := filterOutSocialActivities(nextSessions)

	return filteredSessions, nil
}

// CleanupOldSessions removes sessions older than configured hours (parallel cleanup)
func CleanupOldSessions() {
	cutoff := time.Now().Add(-SessionCleanupHours * time.Hour)
	totalCleaned := 0

	// Clean each shard in parallel
	var wg sync.WaitGroup
	cleanedCounts := make([]int, NumShards)

	for i := range NumShards {
		wg.Add(1)
		go func(shardIndex int) {
			defer wg.Done()

			shard := sessionShards[shardIndex]
			shard.mu.Lock()
			defer shard.mu.Unlock()

			cleaned := 0
			for sessionID, state := range shard.sessions {
				if state.LastActivity.Before(cutoff) {
					log.Printf("[%s] Cleaning up expired session (inactive since %v)",
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
		for i := range NumShards {
			shard := sessionShards[i]
			shard.mu.RLock()
			activeCount += len(shard.sessions)
			shard.mu.RUnlock()
		}
		log.Printf("Cleaned up %d expired sessions, %d sessions remain active", totalCleaned, activeCount)
	}
}

// GetSessionStats returns basic statistics about active sessions
func GetSessionStats() map[string]any {
	totalSessions := 0
	shardStats := make([]int, NumShards)

	for i := range NumShards {
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

	// Check if there are still available sessions to choose from
	nextSessions := FindNextAvailableInEachRoom(state.Day, state.LastEndTime, state.Schedule)

	// Schedule is complete only if:
	// 1. No more available sessions, OR
	// 2. Last end time is after 17:00 (5 PM) AND user has selected at least 3 sessions
	lastEndMinutes := timeToMinutes(state.LastEndTime)
	hasLateEndTime := lastEndMinutes >= 17*60 // 17:00 = 17*60 minutes
	hasEnoughSessions := len(state.Schedule) >= 3

	return len(nextSessions) == 0 || (hasLateEndTime && hasEnoughSessions)
}

// generateTimelineView creates a formatted timeline view of user's schedule
func generateTimelineView(state *UserState) string {
	if len(state.Schedule) == 0 {
		return "尚未選擇任何議程"
	}

	// Sort schedule by start time
	sortedSchedule := make([]Session, len(state.Schedule))
	copy(sortedSchedule, state.Schedule)
	sortSessionsByStartTime(sortedSchedule)

	timeline := fmt.Sprintf("您的 %s 議程安排\n\n", state.Day)

	for i, session := range sortedSchedule {
		// Add time gap if needed
		if i > 0 {
			prevEndTime := sortedSchedule[i-1].End
			currentStartTime := session.Start

			prevEndMin := timeToMinutes(prevEndTime)
			currentStartMin := timeToMinutes(currentStartTime)

			if currentStartMin > prevEndMin {
				gapMinutes := currentStartMin - prevEndMin
				timeline += fmt.Sprintf("⏰ %s-%s | 🆓 空檔時間 (%d分鐘)\n\n",
					prevEndTime, currentStartTime, gapMinutes)
			}
		}

		// Format session info
		tags := ""
		if len(session.Tags) > 0 {
			tags = session.Tags[0] // Use first tag as primary
		}

		timeline += fmt.Sprintf("%s-%s | %s\n   %s %s\n   %s | %s | %s %s\n\n",
			session.Start, session.End, session.Room,
			tags, session.Title,
			formatSpeakers(session.Speakers), session.Track,
			session.Language, session.Difficulty)
	}

	// Add statistics
	totalSessions := len(sortedSchedule)
	if totalSessions > 0 {
		firstStart := sortedSchedule[0].Start
		lastEnd := sortedSchedule[totalSessions-1].End

		startMin := timeToMinutes(firstStart)
		endMin := timeToMinutes(lastEnd)
		totalHours := (endMin - startMin) / 60

		timeline += fmt.Sprintf("統計：共選擇 %d 個 session，總時間跨度 %d 小時",
			totalSessions, totalHours)
	}

	return timeline
}

// formatSpeakers formats speaker list for display
func formatSpeakers(speakers []string) string {
	if len(speakers) == 0 {
		return "未知講者"
	}
	if len(speakers) == 1 {
		return speakers[0]
	}
	if len(speakers) <= 3 {
		result := ""
		for i, speaker := range speakers {
			if i > 0 {
				result += ", "
			}
			result += speaker
		}
		return result
	}
	return fmt.Sprintf("%s 等 %d 位講者", speakers[0], len(speakers))
}

// GetNextSession returns next session information with current status
func GetNextSession(sessionID string) (map[string]any, error) {
	provider := &RealTimeProvider{}
	return GetNextSessionWithTime(sessionID, provider)
}

// GetNextSessionWithTime returns next session information with injectable time provider
func GetNextSessionWithTime(sessionID string, timeProvider TimeProvider) (map[string]any, error) {
	state := GetUserState(sessionID)
	if state == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	// Get current time from provider
	now := timeProvider.Now()

	// Check if within COSCUP period
	if !isInCOSCUPPeriod(now) {
		return buildOutsideCOSCUPPeriodResponse(), nil
	}

	// If no schedule planned yet
	if len(state.Schedule) == 0 {
		return map[string]any{
			"status":  "no_schedule",
			"message": "您還沒有規劃行程。請先使用 start_planning 工具選擇要規劃的日期，開始安排您的 COSCUP 議程。",
		}, nil
	}

	// Format time for session analysis
	currentTime := formatTimeForSession(now)
	currentStatus := analyzeCurrentStatus(state, currentTime)

	switch currentStatus.Status {
	case "ongoing":
		return buildOngoingResponse(currentStatus), nil
	case "break":
		return buildBreakResponse(currentStatus), nil
	case "just_ended":
		return buildJustEndedResponse(currentStatus), nil
	case "schedule_complete":
		// Check if user has manually finished planning
		if state.IsCompleted {
			return buildCompleteResponse(currentStatus), nil
		}

		// Before returning complete status, check if there are still sessions available to choose
		nextSessions := FindNextAvailableInEachRoom(state.Day, state.LastEndTime, state.Schedule)
		if len(nextSessions) > 0 {
			// There are still sessions available, suggest continuing planning
			return map[string]any{
				"status":             "planning_available",
				"message":            fmt.Sprintf("您目前已安排 %d 個議程，結束時間是 %s。系統發現還有 %d 個時段可以選擇更多議程。\n\n**重要提示給 LLM：請主動詢問用戶：**\n1. 是否滿意目前的規劃想要結束？請使用 finish_planning 工具\n2. 還是想要查看更多議程選項？請使用 get_options 工具\n\n請根據用戶回應採取相應行動，主動引導用戶做出選擇，不要讓用戶自己決定使用哪個工具。", len(state.Schedule), state.LastEndTime, len(nextSessions)),
				"available_sessions": len(nextSessions),
			}, nil
		}
		return buildCompleteResponse(currentStatus), nil
	default:
		return map[string]any{
			"status":  "unknown",
			"message": "無法判斷當前狀態，請稍後再試。",
		}, nil
	}
}

// TimeProvider interface for time dependency injection (used in tests)
type TimeProvider interface {
	Now() time.Time
}

// RealTimeProvider implements TimeProvider using actual system time
type RealTimeProvider struct{}

func (r *RealTimeProvider) Now() time.Time {
	// Return actual current time
	return time.Now()
}

// MockTimeProvider for testing with custom time
type MockTimeProvider struct {
	fixedTime time.Time
}

func (m *MockTimeProvider) Now() time.Time {
	return m.fixedTime
}

// Helper functions for time handling
func formatTimeForSession(t time.Time) string {
	return t.Format("15:04")
}

func getCOSCUPDay(t time.Time) string {
	if t.Year() == COSCUPYear && t.Month() == COSCUPMonth && t.Day() == COSCUPDay1 {
		return DayAug9
	} else if t.Year() == COSCUPYear && t.Month() == COSCUPMonth && t.Day() == COSCUPDay2 {
		return DayAug10
	}
	return StatusOutsideCOSCUP
}

func isInCOSCUPPeriod(t time.Time) bool {
	return t.Year() == COSCUPYear && t.Month() == COSCUPMonth && (t.Day() == COSCUPDay1 || t.Day() == COSCUPDay2)
}

// SessionStatus represents current session status
type SessionStatus struct {
	Status           string
	CurrentSession   *Session
	NextSession      *Session
	RemainingMinutes int
	BreakMinutes     int
	Route            *RouteInfo
}

// RouteInfo represents route between venues
type RouteInfo struct {
	FromRoom    string
	ToRoom      string
	WalkingTime int // minutes
	RouteDesc   string
	EnoughTime  bool
}

// analyzeCurrentStatus analyzes user's current status
func analyzeCurrentStatus(state *UserState, currentTime string) *SessionStatus {
	currentMinutes := timeToMinutes(currentTime)

	// Sort schedule by start time
	sortedSchedule := make([]Session, len(state.Schedule))
	copy(sortedSchedule, state.Schedule)
	sortSessionsByStartTime(sortedSchedule)

	// Find current and next sessions
	var currentSession, nextSession *Session

	for i, session := range sortedSchedule {
		startMin := timeToMinutes(session.Start)
		endMin := timeToMinutes(session.End)

		// Check if currently in this session
		if currentMinutes >= startMin && currentMinutes < endMin {
			currentSession = &session
			if i+1 < len(sortedSchedule) {
				nextSession = &sortedSchedule[i+1]
			}

			return &SessionStatus{
				Status:           "ongoing",
				CurrentSession:   currentSession,
				NextSession:      nextSession,
				RemainingMinutes: endMin - currentMinutes,
				Route:            calculateRoute(currentSession, nextSession),
			}
		}

		// Check if this is the next session
		if currentMinutes < startMin {
			nextSession = &session

			// Find if there was a previous session that just ended
			var prevSession *Session
			if i > 0 {
				prevSession = &sortedSchedule[i-1]
				prevEndMin := timeToMinutes(prevSession.End)

				// If just ended (within 10 minutes)
				if currentMinutes-prevEndMin <= 10 && currentMinutes >= prevEndMin {
					return &SessionStatus{
						Status:       "just_ended",
						NextSession:  nextSession,
						BreakMinutes: startMin - currentMinutes,
						Route:        calculateRoute(prevSession, nextSession),
					}
				}
			}

			// In break time
			return &SessionStatus{
				Status:       "break",
				NextSession:  nextSession,
				BreakMinutes: startMin - currentMinutes,
				Route:        calculateRoute(nil, nextSession),
			}
		}
	}

	// All sessions in user's personal schedule are completed
	return &SessionStatus{
		Status: "schedule_complete",
	}
}

// calculateRoute calculates route information between sessions
func calculateRoute(fromSession, toSession *Session) *RouteInfo {
	if toSession == nil {
		return nil
	}

	var fromRoom string
	if fromSession != nil {
		fromRoom = fromSession.Room
	}
	toRoom := toSession.Room

	// If same room or no previous room
	if fromRoom == "" || fromRoom == toRoom {
		return &RouteInfo{
			FromRoom:    fromRoom,
			ToRoom:      toRoom,
			WalkingTime: 0,
			RouteDesc:   fmt.Sprintf("議程在相同地點 %s", toRoom),
			EnoughTime:  true,
		}
	}

	// Calculate walking time between different venues
	walkingTime := calculateWalkingTime(fromRoom, toRoom)
	routeDesc := generateRouteDescription(fromRoom, toRoom)

	return &RouteInfo{
		FromRoom:    fromRoom,
		ToRoom:      toRoom,
		WalkingTime: walkingTime,
		RouteDesc:   routeDesc,
		EnoughTime:  true, // We'll calculate this based on break time in the calling function
	}
}

// getBuildingFromRoom returns building code from room name
func getBuildingFromRoom(room string) string {
	if room == BuildingAU || room == "AU101" {
		return BuildingAU
	}
	if room == "RB-101" || room == "RB-102" || room == "RB-105" {
		return BuildingRB
	}
	if len(room) >= 2 && room[:2] == BuildingTR {
		return BuildingTR
	}
	return "Unknown"
}

// calculateWalkingTime returns estimated walking time in minutes between rooms
// WARNING: These are rough estimates only. Actual travel time may be longer due to:
// - Crowded hallways during session breaks
// - Elevator waiting times
// - Getting lost or needing directions
// - Physical accessibility needs
func calculateWalkingTime(fromRoom, toRoom string) int {
	fromBuilding := getBuildingFromRoom(fromRoom)
	toBuilding := getBuildingFromRoom(toRoom)

	// Estimated walking times between buildings (minutes)
	// NOTE: These are conservative estimates and actual time may vary
	walkingTimes := map[string]map[string]int{
		BuildingAU: {BuildingAU: SameBuildingWalkTime, BuildingRB: AUToRBWalkTime, BuildingTR: AUToTRWalkTime},
		BuildingRB: {BuildingAU: RBToAUWalkTime, BuildingRB: RBToRBWalkTime, BuildingTR: RBToTRWalkTime},
		BuildingTR: {BuildingAU: TRToAUWalkTime, BuildingRB: TRToRBWalkTime, BuildingTR: TRInternalWalkTime},
	}

	if times, exists := walkingTimes[fromBuilding]; exists {
		if time, exists := times[toBuilding]; exists {
			return time
		}
	}

	return UnknownWalkTime // Default safe estimate
}

// generateRouteDescription generates human-readable route description
func generateRouteDescription(fromRoom, toRoom string) string {
	buildingNames := map[string]string{
		"AU": "視聽館",
		"RB": "綜合研究大樓",
		"TR": "研揚大樓",
	}

	fromBuilding := getBuildingFromRoom(fromRoom)
	toBuilding := getBuildingFromRoom(toRoom)

	fromName, fromExists := buildingNames[fromBuilding]
	toName, toExists := buildingNames[toBuilding]

	// Handle unknown buildings
	if !fromExists {
		fromName = "Unknown"
	}
	if !toExists {
		toName = "Unknown"
	}

	if fromBuilding == toBuilding && fromExists {
		return fmt.Sprintf("在 %s 內移動：%s → %s", fromName, fromRoom, toRoom)
	}

	return fmt.Sprintf("%s %s → %s %s", fromName, fromRoom, toName, toRoom)
}

// Response builders
func buildOngoingResponse(status *SessionStatus) map[string]any {
	data := map[string]any{
		"status":            "ongoing",
		"current_session":   status.CurrentSession,
		"remaining_minutes": status.RemainingMinutes,
	}

	var message string
	if status.NextSession != nil {
		data["next_session"] = status.NextSession
		data["route"] = status.Route

		message = fmt.Sprintf("🎯 您目前正在 %s 參加「%s」，還有 %d 分鐘結束。\n\n下一場：%s-%s 在 %s\n「%s」\n\n",
			status.CurrentSession.Room,
			status.CurrentSession.Title,
			status.RemainingMinutes,
			status.NextSession.Start,
			status.NextSession.End,
			status.NextSession.Room,
			status.NextSession.Title)

		if status.Route != nil && status.Route.WalkingTime > 0 {
			message += fmt.Sprintf("🚶 移動路線：%s（預估 %d 分鐘，實際可能更久）",
				status.Route.RouteDesc,
				status.Route.WalkingTime)
		}
	} else {
		message = fmt.Sprintf("🎯 您目前正在 %s 參加「%s」，還有 %d 分鐘結束。這是今天最後一場議程。",
			status.CurrentSession.Room,
			status.CurrentSession.Title,
			status.RemainingMinutes)
	}

	data["message"] = message
	return data
}

func buildBreakResponse(status *SessionStatus) map[string]any {
	data := map[string]any{
		"status":        "break",
		"next_session":  status.NextSession,
		"break_minutes": status.BreakMinutes,
		"route":         status.Route,
	}

	message := fmt.Sprintf("⏰ 您目前有 %d 分鐘空檔時間。\n\n下一場：%s-%s 在 %s\n「%s」\n\n",
		status.BreakMinutes,
		status.NextSession.Start,
		status.NextSession.End,
		status.NextSession.Room,
		status.NextSession.Title)

	if status.Route != nil && status.Route.WalkingTime > 0 {
		timeBuffer := status.BreakMinutes - status.Route.WalkingTime
		if timeBuffer > 5 {
			message += fmt.Sprintf("🚶 移動建議：%s（預估 %d 分鐘，實際可能更久）\n✅ 時間很充裕，您還有 %d 分鐘可以休息或逛攤位。",
				status.Route.RouteDesc,
				status.Route.WalkingTime,
				timeBuffer)
		} else if timeBuffer > 0 {
			message += fmt.Sprintf("🚶 移動建議：%s（預估 %d 分鐘，實際可能更久）\n⏱️ 建議現在就開始移動。",
				status.Route.RouteDesc,
				status.Route.WalkingTime)
		} else {
			message += fmt.Sprintf("🚶 移動建議：%s（預估 %d 分鐘，實際可能更久）\n🏃 時間較緊迫，建議立即前往！",
				status.Route.RouteDesc,
				status.Route.WalkingTime)
		}
	} else {
		message += "📍 下一場議程在相同地點，您可以繼續留在原地。"
	}

	data["message"] = message
	return data
}

func buildJustEndedResponse(status *SessionStatus) map[string]any {
	data := map[string]any{
		"status":        "just_ended",
		"next_session":  status.NextSession,
		"break_minutes": status.BreakMinutes,
		"route":         status.Route,
	}

	message := fmt.Sprintf("✅ 議程剛結束！距離下一場還有 %d 分鐘。\n\n下一場：%s-%s 在 %s\n「%s」\n\n",
		status.BreakMinutes,
		status.NextSession.Start,
		status.NextSession.End,
		status.NextSession.Room,
		status.NextSession.Title)

	if status.Route != nil && status.Route.WalkingTime > 0 {
		timeBuffer := status.BreakMinutes - status.Route.WalkingTime
		if timeBuffer > 5 {
			message += fmt.Sprintf("🚶 移動路線：%s（預估 %d 分鐘，實際可能更久）\n😌 時間充裕，可以先休息一下再出發。",
				status.Route.RouteDesc,
				status.Route.WalkingTime)
		} else {
			message += fmt.Sprintf("🚶 移動路線：%s（預估 %d 分鐘，實際可能更久）\n⏰ 建議現在就開始移動。",
				status.Route.RouteDesc,
				status.Route.WalkingTime)
		}
	} else {
		message += "📍 下一場議程在相同地點，您可以留在原地等待。"
	}

	data["message"] = message
	return data
}

func buildCompleteResponse(status *SessionStatus) map[string]any {
	return map[string]any{
		"status":  "schedule_complete",
		"message": "🎉 恭喜！您今天的所有議程都已完成。希望您在 COSCUP 2025 度過了充實的一天！\n\n您可以：\n- 逛逛攤位區域\n- 參加 BoF 活動\n- 與其他與會者交流",
	}
}

func buildOutsideCOSCUPPeriodResponse() map[string]any {
	return map[string]any{
		"status":  "outside_coscup_period",
		"message": "🗓️ 目前不在 COSCUP 2025 活動期間內（8月9-10日）。\n\n如果您想：\n- 📋 查看已規劃的議程：使用 get_schedule\n- 🔍 瀏覽議程資訊：使用 get_session_detail 加上議程代碼\n- 📍 查看會場資訊：使用 get_venue_map\n\n期待與您在 COSCUP 2025 相見！",
	}
}

// filterOutSocialActivities removes long-duration social activities from recommendations
// These are typically 4+ hour activities like Hacking Corner that aren't traditional talks
func filterOutSocialActivities(sessions []Session) []Session {
	if len(sessions) == 0 {
		return sessions
	}

	var filtered []Session
	for _, session := range sessions {
		// Skip if it's a long-duration social activity
		if isSocialActivity(session) {
			continue
		}
		filtered = append(filtered, session)
	}
	return filtered
}

// isSocialActivity checks if a session is a long-duration social activity
func isSocialActivity(session Session) bool {
	// Check for Hacking Corner activities
	if strings.Contains(session.Title, "Hacking Corner") {
		return true
	}

	// Check for activities in hallways (social spaces)
	if strings.Contains(session.Room, "Hallway") {
		return true
	}

	// Check for very long sessions
	startMin := timeToMinutes(session.Start)
	endMin := timeToMinutes(session.End)
	duration := endMin - startMin
	if duration >= LongSessionMinutes {
		return true
	}

	return false
}

// FindRoomSessions returns all sessions for a specific room on a given day
func FindRoomSessions(day, room string) []Session {

	var roomSessions []Session
	for _, session := range sessionsByDay[day] {
		if session.Room == room {
			roomSessions = append(roomSessions, session)
		}
	}

	result := getSimplifiedSessions(roomSessions)

	// Sort by start time using efficient sort.Slice
	sort.Slice(result, func(i, j int) bool {
		return timeToMinutes(result[i].Start) < timeToMinutes(result[j].Start)
	})

	return result
}

// GetCurrentRoomSession returns the session currently running in a room
func GetCurrentRoomSession(room, day, currentTime string) *Session {
	roomSessions := FindRoomSessions(day, room)
	currentMinutes := timeToMinutes(currentTime)

	for _, session := range roomSessions {
		startMin := timeToMinutes(session.Start)
		endMin := timeToMinutes(session.End)

		// Check if current time is within session period
		if currentMinutes >= startMin && currentMinutes < endMin {
			return &session
		}
	}

	return nil
}

// GetNextRoomSession returns the next session in a room after the current time
func GetNextRoomSession(room, day, currentTime string) *Session {
	roomSessions := FindRoomSessions(day, room)
	currentMinutes := timeToMinutes(currentTime)

	for _, session := range roomSessions {
		startMin := timeToMinutes(session.Start)

		// Find first session that starts after current time
		if startMin > currentMinutes {
			return &session
		}
	}

	return nil
}

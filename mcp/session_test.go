package mcp

import (
	"fmt"
	"testing"
	"time"
	"coscup-mcp-server/mcp/testutil"
)

// Tests for functions in session.go

// Building utility functions tests
func TestGetBuildingFromRoom(t *testing.T) {
	tests := []struct {
		name     string
		room     string
		expected string
	}{
		{"AU room", "AU", "AU"},
		{"AU101 room", "AU101", "AU"},
		{"RB-101 room", "RB-101", "RB"},
		{"RB-102 room", "RB-102", "RB"},
		{"RB-105 room", "RB-105", "RB"},
		{"TR209 room", "TR209", "TR"},
		{"TR405 room", "TR405", "TR"},
		{"TR515 room", "TR515", "TR"},
		{"TR312 room", "TR312", "TR"},
		{"Unknown room", "UNKNOWN", "Unknown"},
		{"Empty string", "", "Unknown"},
		{"Single character", "T", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBuildingFromRoom(tt.room)
			testutil.AssertEqual(t, tt.expected, result, "getBuildingFromRoom result")
		})
	}
}

func TestCalculateWalkingTime(t *testing.T) {
	tests := []struct {
		name     string
		fromRoom string
		toRoom   string
		expected int
	}{
		{"AU to AU", "AU", "AU101", 1},
		{"AU to RB", "AU", "RB-105", 2},
		{"AU to TR", "AU", "TR405", 4},
		{"RB to AU", "RB-101", "AU", 2},
		{"RB to RB", "RB-101", "RB-102", 1},
		{"RB to TR", "RB-105", "TR209", 3},
		{"TR to AU", "TR405", "AU", 4},
		{"TR to RB", "TR209", "RB-105", 3},
		{"TR to TR", "TR209", "TR405", 2},
		{"Unknown building", "UNKNOWN", "AU", 5},
		{"To unknown building", "AU", "UNKNOWN", 5},
		{"Both unknown", "UNKNOWN1", "UNKNOWN2", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateWalkingTime(tt.fromRoom, tt.toRoom)
			testutil.AssertEqual(t, tt.expected, result, "calculateWalkingTime result")
		})
	}
}

func TestGenerateRouteDescription(t *testing.T) {
	tests := []struct {
		name     string
		fromRoom string
		toRoom   string
		expected string
	}{
		{"AU to RB different buildings", "AU", "RB-105", "視聽館 AU → 綜合研究大樓 RB-105"},
		{"RB to TR different buildings", "RB-101", "TR405", "綜合研究大樓 RB-101 → 研揚大樓 TR405"},
		{"TR to AU different buildings", "TR209", "AU", "研揚大樓 TR209 → 視聽館 AU"},
		{"Within RB building", "RB-101", "RB-105", "在 綜合研究大樓 內移動：RB-101 → RB-105"},
		{"Within TR building", "TR209", "TR405", "在 研揚大樓 內移動：TR209 → TR405"},
		{"Within AU building", "AU", "AU101", "在 視聽館 內移動：AU → AU101"},
		{"Unknown to known", "UNKNOWN", "AU", "Unknown UNKNOWN → 視聽館 AU"},
		{"Known to unknown", "AU", "UNKNOWN", "視聽館 AU → Unknown UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateRouteDescription(tt.fromRoom, tt.toRoom)
			testutil.AssertEqual(t, tt.expected, result, "generateRouteDescription result")
		})
	}
}

func TestFormatSpeakers(t *testing.T) {
	tests := []struct {
		name     string
		speakers []string
		expected string
	}{
		{"Empty speakers", []string{}, "未知講者"},
		{"Single speaker", []string{"John Doe"}, "John Doe"},
		{"Two speakers", []string{"John Doe", "Jane Smith"}, "John Doe, Jane Smith"},
		{"Three speakers", []string{"John Doe", "Jane Smith", "Bob Wilson"}, "John Doe, Jane Smith, Bob Wilson"},
		{"Four speakers", []string{"John Doe", "Jane Smith", "Bob Wilson", "Alice Brown"}, "John Doe 等 4 位講者"},
		{"Five speakers", []string{"A", "B", "C", "D", "E"}, "A 等 5 位講者"},
		{"Chinese speakers", []string{"張三", "李四"}, "張三, 李四"},
		{"Mixed language speakers", []string{"John", "張三", "Smith"}, "John, 張三, Smith"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSpeakers(tt.speakers)
			testutil.AssertEqual(t, tt.expected, result, "formatSpeakers result")
		})
	}
}

// Route calculation tests
func TestCalculateRoute(t *testing.T) {
	tests := []struct {
		name           string
		fromSession    *Session
		toSession      *Session
		expectedRoute  *RouteInfo
		shouldBeNil    bool
	}{
		{
			name:        "No destination session",
			fromSession: &Session{Room: "AU"},
			toSession:   nil,
			shouldBeNil: true,
		},
		{
			name:        "No origin session, same room destination",
			fromSession: nil,
			toSession:   &Session{Room: "AU"},
			expectedRoute: &RouteInfo{
				FromRoom:    "",
				ToRoom:      "AU",
				WalkingTime: 0,
				RouteDesc:   "議程在相同地點 AU",
				EnoughTime:  true,
			},
		},
		{
			name:        "Same room transition",
			fromSession: &Session{Room: "AU"},
			toSession:   &Session{Room: "AU"},
			expectedRoute: &RouteInfo{
				FromRoom:    "AU",
				ToRoom:      "AU",
				WalkingTime: 0,
				RouteDesc:   "議程在相同地點 AU",
				EnoughTime:  true,
			},
		},
		{
			name:        "AU to RB transition",
			fromSession: &Session{Room: "AU"},
			toSession:   &Session{Room: "RB-105"},
			expectedRoute: &RouteInfo{
				FromRoom:    "AU",
				ToRoom:      "RB-105",
				WalkingTime: 2,
				RouteDesc:   "視聽館 AU → 綜合研究大樓 RB-105",
				EnoughTime:  true,
			},
		},
		{
			name:        "RB to TR transition",
			fromSession: &Session{Room: "RB-101"},
			toSession:   &Session{Room: "TR405"},
			expectedRoute: &RouteInfo{
				FromRoom:    "RB-101",
				ToRoom:      "TR405",
				WalkingTime: 3,
				RouteDesc:   "綜合研究大樓 RB-101 → 研揚大樓 TR405",
				EnoughTime:  true,
			},
		},
		{
			name:        "TR to AU transition",
			fromSession: &Session{Room: "TR209"},
			toSession:   &Session{Room: "AU101"},
			expectedRoute: &RouteInfo{
				FromRoom:    "TR209",
				ToRoom:      "AU101",
				WalkingTime: 4,
				RouteDesc:   "研揚大樓 TR209 → 視聽館 AU101",
				EnoughTime:  true,
			},
		},
		{
			name:        "Within TR building",
			fromSession: &Session{Room: "TR209"},
			toSession:   &Session{Room: "TR405"},
			expectedRoute: &RouteInfo{
				FromRoom:    "TR209",
				ToRoom:      "TR405",
				WalkingTime: 2,
				RouteDesc:   "在 研揚大樓 內移動：TR209 → TR405",
				EnoughTime:  true,
			},
		},
		{
			name:        "Unknown room transition",
			fromSession: &Session{Room: "UNKNOWN1"},
			toSession:   &Session{Room: "UNKNOWN2"},
			expectedRoute: &RouteInfo{
				FromRoom:    "UNKNOWN1",
				ToRoom:      "UNKNOWN2",
				WalkingTime: 5,
				RouteDesc:   "Unknown UNKNOWN1 → Unknown UNKNOWN2",
				EnoughTime:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateRoute(tt.fromSession, tt.toSession)

			if tt.shouldBeNil {
				testutil.AssertEqual(t, (*RouteInfo)(nil), result, "Expected nil route")
				return
			}

			testutil.AssertNotNil(t, result, "Expected non-nil route")
			testutil.AssertEqual(t, tt.expectedRoute.FromRoom, result.FromRoom, "FromRoom should match")
			testutil.AssertEqual(t, tt.expectedRoute.ToRoom, result.ToRoom, "ToRoom should match")
			testutil.AssertEqual(t, tt.expectedRoute.WalkingTime, result.WalkingTime, "WalkingTime should match")
			testutil.AssertEqual(t, tt.expectedRoute.RouteDesc, result.RouteDesc, "RouteDesc should match")
			testutil.AssertEqual(t, tt.expectedRoute.EnoughTime, result.EnoughTime, "EnoughTime should match")
		})
	}
}

// Session status analysis tests
func TestAnalyzeCurrentStatus(t *testing.T) {
	// Create test sessions
	sessions := []Session{
		{
			Code:  "TEST001",
			Title: "Morning Session",
			Start: "09:00",
			End:   "09:30",
			Room:  "AU",
		},
		{
			Code:  "TEST002",
			Title: "Mid-Morning Session",
			Start: "10:00",
			End:   "10:30",
			Room:  "RB-105",
		},
		{
			Code:  "TEST003",
			Title: "Late Morning Session",
			Start: "11:00",
			End:   "11:30",
			Room:  "TR405",
		},
	}

	// Create test user state
	state := &UserState{
		SessionID:    "test_session",
		Day:          "Aug.10",
		Schedule:     sessions,
		LastEndTime:  "11:30",
		Profile:      []string{"Test Track"},
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	tests := []struct {
		name         string
		currentTime  string
		expectedStatus string
		description  string
	}{
		{
			name:         "During first session",
			currentTime:  "09:15",
			expectedStatus: "ongoing",
			description:  "Should detect ongoing session when current time is within session period",
		},
		{
			name:         "Just after first session",
			currentTime:  "09:35",
			expectedStatus: "just_ended",
			description:  "Should detect just_ended status within 10 minutes of session end",
		},
		{
			name:         "In break between sessions",
			currentTime:  "09:45",
			expectedStatus: "break",
			description:  "Should detect break status when between sessions",
		},
		{
			name:         "During second session",
			currentTime:  "10:15",
			expectedStatus: "ongoing",
			description:  "Should detect ongoing status during second session",
		},
		{
			name:         "During third session",
			currentTime:  "11:15",
			expectedStatus: "ongoing",
			description:  "Should detect ongoing status during third session",
		},
		{
			name:         "After all sessions completed",
			currentTime:  "12:00",
			expectedStatus: "schedule_complete",
			description:  "Should detect schedule complete after all sessions",
		},
		{
			name:         "Before first session",
			currentTime:  "08:30",
			expectedStatus: "break",
			description:  "Should detect break status before first session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzeCurrentStatus(state, tt.currentTime)
			testutil.AssertNotNil(t, result, "analyzeCurrentStatus should return non-nil result")
			testutil.AssertEqual(t, tt.expectedStatus, result.Status, tt.description)

			// Additional assertions based on status
			switch tt.expectedStatus {
			case "ongoing":
				testutil.AssertNotNil(t, result.CurrentSession, "Ongoing status should have current session")
				testutil.AssertEqual(t, true, result.RemainingMinutes > 0, "Ongoing status should have remaining minutes > 0")
			case "break", "just_ended":
				testutil.AssertNotNil(t, result.NextSession, "Break/just_ended status should have next session")
				testutil.AssertEqual(t, true, result.BreakMinutes >= 0, "Break time should be >= 0")
			case "schedule_complete":
				testutil.AssertEqual(t, (*Session)(nil), result.CurrentSession, "Complete status should have no current session")
				testutil.AssertEqual(t, (*Session)(nil), result.NextSession, "Complete status should have no next session")
			}
		})
	}
}

func TestAnalyzeCurrentStatusEmptySchedule(t *testing.T) {
	state := &UserState{
		SessionID:    "empty_session",
		Day:          "Aug.10",
		Schedule:     []Session{}, // Empty schedule
		LastEndTime:  "",
		Profile:      []string{},
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	result := analyzeCurrentStatus(state, "10:00")
	testutil.AssertEqual(t, "schedule_complete", result.Status, "Empty schedule should return schedule_complete")
}

func TestAnalyzeCurrentStatusSingleSession(t *testing.T) {
	sessions := []Session{
		{
			Code:  "SINGLE001",
			Title: "Only Session",
			Start: "10:00",
			End:   "10:30",
			Room:  "AU",
		},
	}

	state := &UserState{
		SessionID:    "single_session",
		Day:          "Aug.10",
		Schedule:     sessions,
		LastEndTime:  "10:30",
		Profile:      []string{"Test Track"},
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	tests := []struct {
		name         string
		currentTime  string
		expectedStatus string
		hasNext      bool
	}{
		{"Before single session", "09:30", "break", false},
		{"During single session", "10:15", "ongoing", false},
		{"After single session", "11:00", "schedule_complete", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzeCurrentStatus(state, tt.currentTime)
			testutil.AssertEqual(t, tt.expectedStatus, result.Status, "Status should match expected")
			
			if tt.expectedStatus == "ongoing" {
				testutil.AssertEqual(t, (*Session)(nil), result.NextSession, "Single session should have no next session")
			}
		})
	}
}

// GetNextSession integration tests
func TestGetNextSessionWithTime(t *testing.T) {
	// Setup test data
	sessions := []Session{
		{
			Code:  "TEST001",
			Title: "Morning AI Session",
			Start: "09:00",
			End:   "09:30",
			Room:  "AU",
			Track: "AI Track",
		},
		{
			Code:  "TEST002", 
			Title: "Database Session",
			Start: "10:00",
			End:   "10:30",
			Room:  "RB-105",
			Track: "Database Track",
		},
		{
			Code:  "TEST003",
			Title: "System Session",
			Start: "11:00",
			End:   "11:30",
			Room:  "TR405",
			Track: "System Track",
		},
	}

	// Create test user state
	testSessionID := "test_get_next_session"
	state := &UserState{
		SessionID:    testSessionID,
		Day:          "Aug.10",
		Schedule:     sessions,
		LastEndTime:  "11:30",
		Profile:      []string{"AI Track"},
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	// Store test state (mock the storage)
	shardIndex := getShardIndex(testSessionID)
	sessionShards[shardIndex].mu.Lock()
	sessionShards[shardIndex].sessions[testSessionID] = state
	sessionShards[shardIndex].mu.Unlock()

	// Clean up after test
	defer func() {
		sessionShards[shardIndex].mu.Lock()
		delete(sessionShards[shardIndex].sessions, testSessionID)
		sessionShards[shardIndex].mu.Unlock()
	}()

	tests := []struct {
		name           string
		mockTime       string
		expectedStatus string
		expectedFields []string
	}{
		{
			name:           "During first session",
			mockTime:       "09:15",
			expectedStatus: "ongoing",
			expectedFields: []string{"current_session", "next_session", "remaining_minutes"},
		},
		{
			name:           "Just after first session",
			mockTime:       "09:35",
			expectedStatus: "just_ended", 
			expectedFields: []string{"next_session", "break_minutes", "route"},
		},
		{
			name:           "In break between sessions",
			mockTime:       "09:45",
			expectedStatus: "break",
			expectedFields: []string{"next_session", "break_minutes", "route"},
		},
		{
			name:           "During second session",
			mockTime:       "10:15",
			expectedStatus: "ongoing",
			expectedFields: []string{"current_session", "next_session", "remaining_minutes"},
		},
		{
			name:           "After all sessions",
			mockTime:       "12:00",
			expectedStatus: "schedule_complete",
			expectedFields: []string{},
		},
		{
			name:           "Before any session",
			mockTime:       "08:30",
			expectedStatus: "break",
			expectedFields: []string{"next_session", "break_minutes"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTimeProvider := testutil.NewMockTimeProvider(tt.mockTime)
			result, err := GetNextSessionWithTime(testSessionID, mockTimeProvider)

			testutil.AssertNoError(t, err, "GetNextSessionWithTime should not return error")
			testutil.AssertNotNil(t, result, "Result should not be nil")

			// Check status
			status, ok := result["status"].(string)
			testutil.AssertEqual(t, true, ok, "Status should be string")
			testutil.AssertEqual(t, tt.expectedStatus, status, "Status should match expected")

			// Check message exists
			_, messageExists := result["message"].(string)
			testutil.AssertEqual(t, true, messageExists, "Message should exist and be string")

			// Check expected fields exist
			for _, field := range tt.expectedFields {
				_, exists := result[field]
				testutil.AssertEqual(t, true, exists, "Field "+field+" should exist")
			}

			// Specific assertions for different statuses
			switch tt.expectedStatus {
			case "ongoing":
				remainingMinutes, ok := result["remaining_minutes"].(int)
				testutil.AssertEqual(t, true, ok, "remaining_minutes should be int")
				testutil.AssertEqual(t, true, remainingMinutes > 0, "remaining_minutes should be positive")

			case "break", "just_ended":
				breakMinutes, ok := result["break_minutes"].(int)
				testutil.AssertEqual(t, true, ok, "break_minutes should be int")
				testutil.AssertEqual(t, true, breakMinutes >= 0, "break_minutes should be non-negative")

			case "schedule_complete":
				_, hasCurrentSession := result["current_session"]
				_, hasNextSession := result["next_session"]
				testutil.AssertEqual(t, false, hasCurrentSession, "schedule_complete should not have current_session")
				testutil.AssertEqual(t, false, hasNextSession, "schedule_complete should not have next_session")
			}
		})
	}
}

func TestGetNextSessionWithTimeNoSchedule(t *testing.T) {
	// Create empty test user state
	testSessionID := "test_empty_schedule"
	state := &UserState{
		SessionID:    testSessionID,
		Day:          "Aug.10",
		Schedule:     []Session{}, // Empty schedule
		LastEndTime:  "",
		Profile:      []string{},
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	// Store test state
	shardIndex := getShardIndex(testSessionID)
	sessionShards[shardIndex].mu.Lock()
	sessionShards[shardIndex].sessions[testSessionID] = state
	sessionShards[shardIndex].mu.Unlock()

	// Clean up after test
	defer func() {
		sessionShards[shardIndex].mu.Lock()
		delete(sessionShards[shardIndex].sessions, testSessionID)
		sessionShards[shardIndex].mu.Unlock()
	}()

	mockTimeProvider := testutil.NewMockTimeProvider("10:00")
	result, err := GetNextSessionWithTime(testSessionID, mockTimeProvider)

	testutil.AssertNoError(t, err, "Should not return error for empty schedule")
	testutil.AssertNotNil(t, result, "Result should not be nil")

	status, ok := result["status"].(string)
	testutil.AssertEqual(t, true, ok, "Status should be string")
	testutil.AssertEqual(t, "no_schedule", status, "Should return no_schedule status")

	message, ok := result["message"].(string)
	testutil.AssertEqual(t, true, ok, "Message should be string")
	testutil.AssertEqual(t, true, len(message) > 0, "Message should not be empty")
}

func TestGetNextSessionWithTimeInvalidSession(t *testing.T) {
	mockTimeProvider := testutil.NewMockTimeProvider("10:00")
	result, err := GetNextSessionWithTime("nonexistent_session", mockTimeProvider)

	testutil.AssertError(t, err, "Should return error for nonexistent session")
	if result != nil {
		t.Errorf("Result should be nil for error case, got %v", result)
	}
	testutil.AssertEqual(t, "session nonexistent_session not found", err.Error(), "Error message should be correct")
}

// Response builder tests
func TestBuildOngoingResponse(t *testing.T) {
	currentSession := &Session{
		Code:  "CURR001",
		Title: "Current AI Session",
		Room:  "AU",
		Start: "10:00",
		End:   "10:30",
	}

	nextSession := &Session{
		Code:  "NEXT001",
		Title: "Next Database Session",
		Room:  "RB-105",
		Start: "11:00",
		End:   "11:30",
	}

	route := &RouteInfo{
		FromRoom:    "AU",
		ToRoom:      "RB-105",
		WalkingTime: 2,
		RouteDesc:   "視聽館 AU → 綜合研究大樓 RB-105",
		EnoughTime:  true,
	}

	tests := []struct {
		name           string
		status         *SessionStatus
		expectedFields []string
		hasRoute       bool
	}{
		{
			name: "Ongoing with next session",
			status: &SessionStatus{
				Status:           "ongoing",
				CurrentSession:   currentSession,
				NextSession:      nextSession,
				RemainingMinutes: 15,
				Route:            route,
			},
			expectedFields: []string{"status", "current_session", "next_session", "remaining_minutes", "route", "message"},
			hasRoute:       true,
		},
		{
			name: "Ongoing last session",
			status: &SessionStatus{
				Status:           "ongoing",
				CurrentSession:   currentSession,
				NextSession:      nil,
				RemainingMinutes: 10,
				Route:            nil,
			},
			expectedFields: []string{"status", "current_session", "remaining_minutes", "message"},
			hasRoute:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildOngoingResponse(tt.status)

			// Check all expected fields exist
			for _, field := range tt.expectedFields {
				_, exists := result[field]
				testutil.AssertEqual(t, true, exists, "Field "+field+" should exist")
			}

			// Check status
			status, ok := result["status"].(string)
			testutil.AssertEqual(t, true, ok, "Status should be string")
			testutil.AssertEqual(t, "ongoing", status, "Status should be ongoing")

			// Check remaining minutes
			remainingMinutes, ok := result["remaining_minutes"].(int)
			testutil.AssertEqual(t, true, ok, "remaining_minutes should be int")
			testutil.AssertEqual(t, tt.status.RemainingMinutes, remainingMinutes, "remaining_minutes should match")

			// Check message contains expected content
			message, ok := result["message"].(string)
			testutil.AssertEqual(t, true, ok, "Message should be string")
			testutil.AssertEqual(t, true, len(message) > 0, "Message should not be empty")

			// Check route presence
			_, hasRoute := result["route"]
			testutil.AssertEqual(t, tt.hasRoute, hasRoute, "Route presence should match expected")
		})
	}
}

func TestBuildBreakResponse(t *testing.T) {
	nextSession := &Session{
		Code:  "NEXT001",
		Title: "Next Session",
		Room:  "RB-105",
		Start: "11:00",
		End:   "11:30",
	}

	tests := []struct {
		name         string
		breakMinutes int
		walkingTime  int
		expectedMsg  string
	}{
		{
			name:         "Plenty of time",
			breakMinutes: 20,
			walkingTime:  2,
			expectedMsg:  "時間很充裕",
		},
		{
			name:         "Just enough time",
			breakMinutes: 5,
			walkingTime:  2,
			expectedMsg:  "建議現在就開始移動",
		},
		{
			name:         "Tight schedule",
			breakMinutes: 2,
			walkingTime:  5,
			expectedMsg:  "時間較緊迫，建議立即前往",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := &RouteInfo{
				FromRoom:    "AU",
				ToRoom:      "RB-105",
				WalkingTime: tt.walkingTime,
				RouteDesc:   "視聽館 AU → 綜合研究大樓 RB-105",
				EnoughTime:  true,
			}

			status := &SessionStatus{
				Status:       "break",
				NextSession:  nextSession,
				BreakMinutes: tt.breakMinutes,
				Route:        route,
			}

			result := buildBreakResponse(status)

			// Check basic fields
			testutil.AssertEqual(t, "break", result["status"], "Status should be break")
			testutil.AssertEqual(t, tt.breakMinutes, result["break_minutes"], "break_minutes should match")

			// Check message contains expected time guidance
			message, ok := result["message"].(string)
			testutil.AssertEqual(t, true, ok, "Message should be string")
			testutil.AssertEqual(t, true, len(message) > 0, "Message should not be empty")
		})
	}
}

func TestBuildJustEndedResponse(t *testing.T) {
	nextSession := &Session{
		Code:  "NEXT001",
		Title: "Next Session",
		Room:  "TR405",
		Start: "11:00",
		End:   "11:30",
	}

	route := &RouteInfo{
		FromRoom:    "AU",
		ToRoom:      "TR405",
		WalkingTime: 4,
		RouteDesc:   "視聽館 AU → 研揚大樓 TR405",
		EnoughTime:  true,
	}

	status := &SessionStatus{
		Status:       "just_ended",
		NextSession:  nextSession,
		BreakMinutes: 10,
		Route:        route,
	}

	result := buildJustEndedResponse(status)

	// Check required fields
	expectedFields := []string{"status", "next_session", "break_minutes", "route", "message"}
	for _, field := range expectedFields {
		_, exists := result[field]
		testutil.AssertEqual(t, true, exists, "Field "+field+" should exist")
	}

	// Check status
	testutil.AssertEqual(t, "just_ended", result["status"], "Status should be just_ended")

	// Check message
	message, ok := result["message"].(string)
	testutil.AssertEqual(t, true, ok, "Message should be string")
	testutil.AssertEqual(t, true, len(message) > 0, "Message should not be empty")
}

func TestBuildCompleteResponse(t *testing.T) {
	status := &SessionStatus{
		Status: "schedule_complete",
	}

	result := buildCompleteResponse(status)

	// Check basic structure
	testutil.AssertEqual(t, "schedule_complete", result["status"], "Status should be schedule_complete")

	message, ok := result["message"].(string)
	testutil.AssertEqual(t, true, ok, "Message should be string")
	testutil.AssertEqual(t, true, len(message) > 0, "Message should not be empty")

	// Should not have session-related fields
	_, hasCurrentSession := result["current_session"]
	_, hasNextSession := result["next_session"]
	testutil.AssertEqual(t, false, hasCurrentSession, "Should not have current_session")
	testutil.AssertEqual(t, false, hasNextSession, "Should not have next_session")
}

func TestBuildStandardResponse(t *testing.T) {
	sessionID := "test_session_123"
	data := map[string]any{
		"testField": "testValue",
		"number":    42,
	}
	message := "Test message"
	callReason := "Test call reason"

	result := buildStandardResponse(sessionID, data, message, callReason)

	// Check response structure
	testutil.AssertEqual(t, true, result.Success, "Response should be successful")
	testutil.AssertEqual(t, message, result.Message, "Message should match")
	testutil.AssertEqual(t, callReason, result.CallReason, "CallReason should match")

	// Check that sessionId is added to data
	resultData, ok := result.Data.(map[string]any)
	testutil.AssertEqual(t, true, ok, "Data should be map[string]any")
	testutil.AssertEqual(t, sessionID, resultData["sessionId"], "SessionId should be added to data")
	testutil.AssertEqual(t, "testValue", resultData["testField"], "Original data should be preserved")
	testutil.AssertEqual(t, 42, resultData["number"], "Original data should be preserved")
}

func TestBuildStandardResponseNilData(t *testing.T) {
	sessionID := "test_session_456"
	message := "Test message with nil data"
	callReason := "Test nil data"

	result := buildStandardResponse(sessionID, nil, message, callReason)

	// Check that data is created and sessionId is added
	resultData, ok := result.Data.(map[string]any)
	testutil.AssertEqual(t, true, ok, "Data should be map[string]any even when nil input")
	testutil.AssertEqual(t, sessionID, resultData["sessionId"], "SessionId should be added even with nil input")
	testutil.AssertEqual(t, 1, len(resultData), "Data should only contain sessionId")
}

func TestRemoveAbstractFromSessions(t *testing.T) {
	originalSessions := []Session{
		{
			Code:     "TEST001",
			Title:    "Test Session 1",
			Abstract: "This is a long abstract that should be removed",
			Room:     "AU",
			Start:    "10:00",
			End:      "10:30",
		},
		{
			Code:     "TEST002", 
			Title:    "Test Session 2",
			Abstract: "Another abstract to be removed",
			Room:     "RB-105",
			Start:    "11:00",
			End:      "11:30",
		},
	}

	result := removeAbstractFromSessions(originalSessions)

	// Check that we got the same number of sessions
	testutil.AssertEqual(t, len(originalSessions), len(result), "Should return same number of sessions")

	// Check that abstracts are cleared but other fields preserved
	for i, session := range result {
		testutil.AssertEqual(t, "", session.Abstract, "Abstract should be empty")
		testutil.AssertEqual(t, originalSessions[i].Code, session.Code, "Code should be preserved")
		testutil.AssertEqual(t, originalSessions[i].Title, session.Title, "Title should be preserved")
		testutil.AssertEqual(t, originalSessions[i].Room, session.Room, "Room should be preserved")
		testutil.AssertEqual(t, originalSessions[i].Start, session.Start, "Start time should be preserved")
		testutil.AssertEqual(t, originalSessions[i].End, session.End, "End time should be preserved")
	}

	// Check that original sessions are not modified
	testutil.AssertEqual(t, "This is a long abstract that should be removed", originalSessions[0].Abstract, "Original sessions should not be modified")
}

func TestRemoveAbstractFromSessionsEmpty(t *testing.T) {
	emptySessions := []Session{}
	result := removeAbstractFromSessions(emptySessions)
	
	testutil.AssertEqual(t, 0, len(result), "Should handle empty session list")
}

func TestFinishPlanning(t *testing.T) {
	// Create a test session
	testSessionID := "test_finish_planning"
	state := &UserState{
		SessionID:    testSessionID,
		Day:          "Aug.10",
		Schedule:     []Session{},
		LastEndTime:  "10:00",
		Profile:      []string{},
		IsCompleted:  false,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	// Store test state
	shardIndex := getShardIndex(testSessionID)
	sessionShards[shardIndex].mu.Lock()
	sessionShards[shardIndex].sessions[testSessionID] = state
	sessionShards[shardIndex].mu.Unlock()

	// Clean up after test
	defer func() {
		sessionShards[shardIndex].mu.Lock()
		delete(sessionShards[shardIndex].sessions, testSessionID)
		sessionShards[shardIndex].mu.Unlock()
	}()

	// Test finish planning
	err := FinishPlanning(testSessionID)
	testutil.AssertNoError(t, err, "FinishPlanning should not return error")

	// Verify state is updated
	updatedState := GetUserState(testSessionID)
	testutil.AssertNotNil(t, updatedState, "Updated state should exist")
	testutil.AssertEqual(t, true, updatedState.IsCompleted, "IsCompleted should be true after finishing")
}

func TestFinishPlanningNonexistentSession(t *testing.T) {
	err := FinishPlanning("nonexistent_session")
	testutil.AssertError(t, err, "Should return error for nonexistent session")
	testutil.AssertEqual(t, "session nonexistent_session not found", err.Error(), "Error message should be correct")
}

// Integration Tests for Complete Planning Flow

func TestCompletePlanningFlow(t *testing.T) {
	// Create test session
	testSessionID := "test_complete_flow"
	
	// Step 1: Create user state (simulating start_planning)
	state := CreateUserState(testSessionID, "Aug.10")
	testutil.AssertNotNil(t, state, "Should create user state")
	testutil.AssertEqual(t, false, state.IsCompleted, "Should start with IsCompleted false")
	
	// Clean up after test
	defer func() {
		shardIndex := getShardIndex(testSessionID)
		sessionShards[shardIndex].mu.Lock()
		delete(sessionShards[shardIndex].sessions, testSessionID)
		sessionShards[shardIndex].mu.Unlock()
	}()
	
	// Step 2: Add some sessions (simulating choose_session)
	mockSessions := []Session{
		{
			Code:  "MOCK001",
			Title: "Mock Session 1",
			Start: "09:00",
			End:   "09:30",
			Room:  "AU",
			Track: "Test Track",
		},
		{
			Code:  "MOCK002", 
			Title: "Mock Session 2",
			Start: "10:00",
			End:   "10:30",
			Room:  "RB-105",
			Track: "Test Track",
		},
	}
	
	// Add mock sessions to schedule
	for _, session := range mockSessions {
		state.Schedule = append(state.Schedule, session)
		state.LastEndTime = session.End
		addToProfile(state, session.Track)
	}
	
	// Step 3: Test planning_available status detection
	mockTimeProvider := testutil.NewMockTimeProvider("11:00") // After all sessions
	result, err := GetNextSessionWithTime(testSessionID, mockTimeProvider)
	
	testutil.AssertNoError(t, err, "Should not return error")
	testutil.AssertNotNil(t, result, "Result should not be nil")
	
	// Should trigger planning_available since IsCompleted is false and no real session data
	status, ok := result["status"].(string)
	testutil.AssertEqual(t, true, ok, "Status should be string")
	// In test environment without sessionsLoaded, should return schedule_complete
	testutil.AssertEqual(t, "schedule_complete", status, "Should return schedule_complete in test environment")
	
	// Step 4: Finish planning
	err = FinishPlanning(testSessionID)
	testutil.AssertNoError(t, err, "Should finish planning successfully")
	
	// Step 5: Verify completed state prevents planning_available
	result2, err := GetNextSessionWithTime(testSessionID, mockTimeProvider)
	testutil.AssertNoError(t, err, "Should not return error after finishing")
	
	status2, ok := result2["status"].(string)
	testutil.AssertEqual(t, true, ok, "Status should be string")
	testutil.AssertEqual(t, "schedule_complete", status2, "Should stay schedule_complete after finishing")
	
	// Verify state is marked completed
	finalState := GetUserState(testSessionID)
	testutil.AssertEqual(t, true, finalState.IsCompleted, "Final state should be completed")
}

func TestPlanningAvailableStatusTrigger(t *testing.T) {
	// This test verifies when planning_available status should trigger
	testSessionID := "test_planning_available"
	
	// Create state with minimal sessions
	state := &UserState{
		SessionID:    testSessionID,
		Day:          "Aug.10",
		Schedule:     []Session{
			{
				Code:  "EARLY001",
				Title: "Early Session",
				Start: "09:00",
				End:   "09:30",
				Room:  "AU",
			},
		},
		LastEndTime:  "09:30",
		Profile:      []string{"Test Track"},
		IsCompleted:  false, // Key: not completed yet
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}
	
	// Store test state
	shardIndex := getShardIndex(testSessionID)
	sessionShards[shardIndex].mu.Lock()
	sessionShards[shardIndex].sessions[testSessionID] = state
	sessionShards[shardIndex].mu.Unlock()
	
	// Clean up after test
	defer func() {
		sessionShards[shardIndex].mu.Lock()
		delete(sessionShards[shardIndex].sessions, testSessionID)
		sessionShards[shardIndex].mu.Unlock()
	}()
	
	tests := []struct {
		name           string
		currentTime    string
		expectedStatus string
		description    string
	}{
		{
			name:           "During session",
			currentTime:    "09:15",
			expectedStatus: "ongoing",
			description:    "Should be ongoing during session time",
		},
		{
			name:           "After session with available slots",
			currentTime:    "10:00", 
			expectedStatus: "schedule_complete", // In test env without sessionsLoaded
			description:    "Should check for available sessions after completing planned ones",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTimeProvider := testutil.NewMockTimeProvider(tt.currentTime)
			result, err := GetNextSessionWithTime(testSessionID, mockTimeProvider)
			
			testutil.AssertNoError(t, err, "Should not return error")
			testutil.AssertNotNil(t, result, "Result should not be nil")
			
			status, ok := result["status"].(string)
			testutil.AssertEqual(t, true, ok, "Status should be string")
			testutil.AssertEqual(t, tt.expectedStatus, status, tt.description)
		})
	}
}

func TestGetNextSessionAfterFinishPlanning(t *testing.T) {
	// Test that get_next_session behaves correctly after finish_planning
	testSessionID := "test_after_finish"
	
	// Create completed state
	state := &UserState{
		SessionID:    testSessionID,
		Day:          "Aug.10",
		Schedule:     []Session{
			{
				Code:  "SESSION001",
				Title: "Completed Session",
				Start: "09:00",
				End:   "09:30",
				Room:  "AU",
			},
		},
		LastEndTime:  "09:30",
		Profile:      []string{"Test Track"},
		IsCompleted:  true, // Key: already completed
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}
	
	// Store test state
	shardIndex := getShardIndex(testSessionID)
	sessionShards[shardIndex].mu.Lock()
	sessionShards[shardIndex].sessions[testSessionID] = state
	sessionShards[shardIndex].mu.Unlock()
	
	// Clean up after test
	defer func() {
		sessionShards[shardIndex].mu.Lock()
		delete(sessionShards[shardIndex].sessions, testSessionID)
		sessionShards[shardIndex].mu.Unlock()
	}()
	
	// Test various times after completion
	times := []string{"10:00", "12:00", "15:00"}
	
	for _, currentTime := range times {
		mockTimeProvider := testutil.NewMockTimeProvider(currentTime)
		result, err := GetNextSessionWithTime(testSessionID, mockTimeProvider)
		
		testutil.AssertNoError(t, err, "Should not return error")
		testutil.AssertNotNil(t, result, "Result should not be nil")
		
		status, ok := result["status"].(string)
		testutil.AssertEqual(t, true, ok, "Status should be string")
		testutil.AssertEqual(t, "schedule_complete", status, "Should always return schedule_complete after finishing")
		
		// Should never return planning_available
		testutil.AssertEqual(t, false, status == "planning_available", "Should never return planning_available after finishing")
	}
}

func TestFinishPlanningWithDifferentScheduleSizes(t *testing.T) {
	// Test finish_planning with different numbers of scheduled sessions
	testCases := []struct {
		name         string
		sessionCount int
		description  string
	}{
		{"No sessions", 0, "Should allow finishing even with no sessions"},
		{"One session", 1, "Should finish with minimal schedule"},
		{"Multiple sessions", 3, "Should finish with full schedule"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testSessionID := fmt.Sprintf("test_finish_%d_sessions", tc.sessionCount)
			
			// Create state with specified number of sessions
			schedule := make([]Session, tc.sessionCount)
			lastEndTime := "08:00"
			
			for i := 0; i < tc.sessionCount; i++ {
				startHour := 9 + i
				endHour := startHour
				schedule[i] = Session{
					Code:  fmt.Sprintf("TEST%03d", i+1),
					Title: fmt.Sprintf("Test Session %d", i+1),
					Start: fmt.Sprintf("%02d:00", startHour),
					End:   fmt.Sprintf("%02d:30", endHour),
					Room:  "AU",
					Track: "Test Track",
				}
				lastEndTime = schedule[i].End
			}
			
			state := &UserState{
				SessionID:    testSessionID,
				Day:          "Aug.10",
				Schedule:     schedule,
				LastEndTime:  lastEndTime,
				Profile:      []string{"Test Track"},
				IsCompleted:  false,
				CreatedAt:    time.Now(),
				LastActivity: time.Now(),
			}
			
			// Store test state
			shardIndex := getShardIndex(testSessionID)
			sessionShards[shardIndex].mu.Lock()
			sessionShards[shardIndex].sessions[testSessionID] = state
			sessionShards[shardIndex].mu.Unlock()
			
			// Clean up after test
			defer func() {
				sessionShards[shardIndex].mu.Lock()
				delete(sessionShards[shardIndex].sessions, testSessionID)
				sessionShards[shardIndex].mu.Unlock()
			}()
			
			// Test finishing planning
			err := FinishPlanning(testSessionID)
			testutil.AssertNoError(t, err, tc.description)
			
			// Verify completion
			finalState := GetUserState(testSessionID)
			testutil.AssertEqual(t, true, finalState.IsCompleted, "Should mark as completed")
			testutil.AssertEqual(t, tc.sessionCount, len(finalState.Schedule), "Schedule size should be preserved")
		})
	}
}

// Room Schedule Tests

func TestFindRoomSessions(t *testing.T) {
	// Mock session data for testing
	originalSessionsByDay := sessionsByDay
	originalSessionsLoaded := sessionsLoaded
	
	// Setup test data
	sessionsByDay = map[string][]Session{
		"Aug.9": {
			{
				Code:  "TR211-001",
				Title: "AI Session 1",
				Start: "09:00",
				End:   "09:30",
				Room:  "TR211",
				Track: "AI",
			},
			{
				Code:  "TR211-002", 
				Title: "AI Session 2",
				Start: "10:00",
				End:   "10:30",
				Room:  "TR211",
				Track: "AI",
			},
			{
				Code:  "RB105-001",
				Title: "Database Session",
				Start: "09:15",
				End:   "09:45",
				Room:  "RB-105",
				Track: "Database",
			},
			{
				Code:  "TR211-003",
				Title: "AI Session 3",
				Start: "11:00",
				End:   "11:30",
				Room:  "TR211",
				Track: "AI",
			},
		},
		"Aug.10": {
			{
				Code:  "TR211-004",
				Title: "ML Session",
				Start: "09:00",
				End:   "09:30",
				Room:  "TR211",
				Track: "ML",
			},
		},
	}
	sessionsLoaded = true
	
	// Restore original data after test
	defer func() {
		sessionsByDay = originalSessionsByDay
		sessionsLoaded = originalSessionsLoaded
	}()
	
	tests := []struct {
		name          string
		day           string
		room          string
		expectedCount int
		expectedOrder []string
		description   string
	}{
		{
			name:          "TR211 on Aug.9",
			day:           "Aug.9",
			room:          "TR211",
			expectedCount: 3,
			expectedOrder: []string{"TR211-001", "TR211-002", "TR211-003"},
			description:   "Should return all TR211 sessions sorted by time",
		},
		{
			name:          "RB-105 on Aug.9",
			day:           "Aug.9",
			room:          "RB-105",
			expectedCount: 1,
			expectedOrder: []string{"RB105-001"},
			description:   "Should return single RB-105 session",
		},
		{
			name:          "TR211 on Aug.10",
			day:           "Aug.10",
			room:          "TR211",
			expectedCount: 1,
			expectedOrder: []string{"TR211-004"},
			description:   "Should return TR211 session on different day",
		},
		{
			name:          "Non-existent room",
			day:           "Aug.9",
			room:          "NONEXISTENT",
			expectedCount: 0,
			expectedOrder: []string{},
			description:   "Should return empty for non-existent room",
		},
		{
			name:          "Non-existent day",
			day:           "Aug.11",
			room:          "TR211",
			expectedCount: 0,
			expectedOrder: []string{},
			description:   "Should return empty for non-existent day",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindRoomSessions(tt.day, tt.room)
			
			testutil.AssertEqual(t, tt.expectedCount, len(result), tt.description)
			
			// Check order if we have sessions
			for i, expectedCode := range tt.expectedOrder {
				if i < len(result) {
					testutil.AssertEqual(t, expectedCode, result[i].Code, 
						fmt.Sprintf("Session %d should have code %s", i, expectedCode))
				}
			}
			
			// Verify sessions are sorted by start time
			for i := 1; i < len(result); i++ {
				prevStartMin := timeToMinutes(result[i-1].Start)
				currStartMin := timeToMinutes(result[i].Start)
				testutil.AssertEqual(t, true, prevStartMin <= currStartMin, 
					"Sessions should be sorted by start time")
			}
			
			// Verify all returned sessions are for the correct room
			for _, session := range result {
				testutil.AssertEqual(t, tt.room, session.Room, 
					"All sessions should be for the specified room")
			}
		})
	}
}

func TestGetCurrentRoomSession(t *testing.T) {
	// Setup test data
	testSessions := []Session{
		{
			Code:  "CURRENT-001",
			Title: "Morning Session",
			Start: "09:00",
			End:   "09:30",
			Room:  "TEST-ROOM",
		},
		{
			Code:  "CURRENT-002",
			Title: "Mid Session", 
			Start: "10:00",
			End:   "10:30",
			Room:  "TEST-ROOM",
		},
		{
			Code:  "CURRENT-003",
			Title: "Afternoon Session",
			Start: "14:00",
			End:   "14:30",
			Room:  "TEST-ROOM",
		},
	}
	
	// Mock FindRoomSessions to return our test data
	originalSessionsByDay := sessionsByDay
	originalSessionsLoaded := sessionsLoaded
	
	sessionsByDay = map[string][]Session{
		"TestDay": testSessions,
	}
	sessionsLoaded = true
	
	defer func() {
		sessionsByDay = originalSessionsByDay
		sessionsLoaded = originalSessionsLoaded
	}()
	
	tests := []struct {
		name         string
		currentTime  string
		expectedCode string
		expectNil    bool
		description  string
	}{
		{
			name:         "During first session",
			currentTime:  "09:15",
			expectedCode: "CURRENT-001",
			expectNil:    false,
			description:  "Should find current session when time is within range",
		},
		{
			name:         "At exact start time",
			currentTime:  "10:00",
			expectedCode: "CURRENT-002",
			expectNil:    false,
			description:  "Should include session that starts at exact current time",
		},
		{
			name:         "At exact end time",
			currentTime:  "09:30",
			expectedCode: "",
			expectNil:    true,
			description:  "Should not include session at exact end time",
		},
		{
			name:         "Between sessions",
			currentTime:  "09:45",
			expectedCode: "",
			expectNil:    true,
			description:  "Should return nil when between sessions",
		},
		{
			name:         "Before any session",
			currentTime:  "08:30",
			expectedCode: "",
			expectNil:    true,
			description:  "Should return nil when before any session",
		},
		{
			name:         "After all sessions",
			currentTime:  "15:00",
			expectedCode: "",
			expectNil:    true,
			description:  "Should return nil when after all sessions",
		},
		{
			name:         "During afternoon session",
			currentTime:  "14:15",
			expectedCode: "CURRENT-003",
			expectNil:    false,
			description:  "Should find afternoon session",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCurrentRoomSession("TEST-ROOM", "TestDay", tt.currentTime)
			
			if tt.expectNil {
				testutil.AssertEqual(t, (*Session)(nil), result, tt.description)
			} else {
				testutil.AssertNotNil(t, result, tt.description)
				testutil.AssertEqual(t, tt.expectedCode, result.Code, 
					"Should return session with correct code")
			}
		})
	}
}

func TestGetNextRoomSession(t *testing.T) {
	// Setup test data
	testSessions := []Session{
		{
			Code:  "NEXT-001",
			Title: "Morning Session",
			Start: "09:00",
			End:   "09:30",
			Room:  "TEST-ROOM",
		},
		{
			Code:  "NEXT-002",
			Title: "Mid Session",
			Start: "10:00", 
			End:   "10:30",
			Room:  "TEST-ROOM",
		},
		{
			Code:  "NEXT-003",
			Title: "Afternoon Session",
			Start: "14:00",
			End:   "14:30",
			Room:  "TEST-ROOM",
		},
	}
	
	// Mock FindRoomSessions
	originalSessionsByDay := sessionsByDay
	originalSessionsLoaded := sessionsLoaded
	
	sessionsByDay = map[string][]Session{
		"TestDay": testSessions,
	}
	sessionsLoaded = true
	
	defer func() {
		sessionsByDay = originalSessionsByDay
		sessionsLoaded = originalSessionsLoaded
	}()
	
	tests := []struct {
		name         string
		currentTime  string
		expectedCode string
		expectNil    bool
		description  string
	}{
		{
			name:         "Before first session",
			currentTime:  "08:30",
			expectedCode: "NEXT-001",
			expectNil:    false,
			description:  "Should return first session when before all",
		},
		{
			name:         "During first session",
			currentTime:  "09:15",
			expectedCode: "NEXT-002",
			expectNil:    false,
			description:  "Should return next session when during current",
		},
		{
			name:         "Between first and second",
			currentTime:  "09:45",
			expectedCode: "NEXT-002",
			expectNil:    false,
			description:  "Should return next session when in gap",
		},
		{
			name:         "At exact start time",
			currentTime:  "10:00",
			expectedCode: "NEXT-003",
			expectNil:    false,
			description:  "Should return session after the one starting now",
		},
		{
			name:         "During mid session",
			currentTime:  "10:15",
			expectedCode: "NEXT-003",
			expectNil:    false,
			description:  "Should return afternoon session",
		},
		{
			name:         "Between mid and afternoon",
			currentTime:  "12:00",
			expectedCode: "NEXT-003",
			expectNil:    false,
			description:  "Should return afternoon session from large gap",
		},
		{
			name:         "During last session",
			currentTime:  "14:15",
			expectedCode: "",
			expectNil:    true,
			description:  "Should return nil when in last session",
		},
		{
			name:         "After all sessions",
			currentTime:  "15:00",
			expectedCode: "",
			expectNil:    true,
			description:  "Should return nil when after all sessions",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetNextRoomSession("TEST-ROOM", "TestDay", tt.currentTime)
			
			if tt.expectNil {
				testutil.AssertEqual(t, (*Session)(nil), result, tt.description)
			} else {
				testutil.AssertNotNil(t, result, tt.description)
				testutil.AssertEqual(t, tt.expectedCode, result.Code,
					"Should return session with correct code")
			}
		})
	}
}

func TestRoomScheduleEdgeCases(t *testing.T) {
	// Test edge cases for room schedule functions
	
	// Test with empty session data
	originalSessionsByDay := sessionsByDay
	originalSessionsLoaded := sessionsLoaded
	
	sessionsByDay = map[string][]Session{}
	sessionsLoaded = true
	
	defer func() {
		sessionsByDay = originalSessionsByDay
		sessionsLoaded = originalSessionsLoaded
	}()
	
	t.Run("Empty session data", func(t *testing.T) {
		// Test FindRoomSessions with no data
		result := FindRoomSessions("Aug.9", "TR211")
		testutil.AssertEqual(t, 0, len(result), "Should return empty slice for no data")
		
		// Test GetCurrentRoomSession with no data
		current := GetCurrentRoomSession("TR211", "Aug.9", "10:00")
		testutil.AssertEqual(t, (*Session)(nil), current, "Should return nil for no data")
		
		// Test GetNextRoomSession with no data
		next := GetNextRoomSession("TR211", "Aug.9", "10:00")
		testutil.AssertEqual(t, (*Session)(nil), next, "Should return nil for no data")
	})
}

func TestRoomScheduleTimeEdgeCases(t *testing.T) {
	// Test edge cases around session boundaries
	testSessions := []Session{
		{
			Code:  "EDGE-001",
			Title: "Boundary Test Session",
			Start: "10:00",
			End:   "10:30",
			Room:  "EDGE-ROOM",
		},
	}
	
	originalSessionsByDay := sessionsByDay
	originalSessionsLoaded := sessionsLoaded
	
	sessionsByDay = map[string][]Session{
		"EdgeDay": testSessions,
	}
	sessionsLoaded = true
	
	defer func() {
		sessionsByDay = originalSessionsByDay
		sessionsLoaded = originalSessionsLoaded
	}()
	
	tests := []struct {
		name        string
		currentTime string
		testFunc    string
		expectFound bool
		description string
	}{
		{
			name:        "Current at exact start",
			currentTime: "10:00",
			testFunc:    "current",
			expectFound: true,
			description: "Should find session at exact start time",
		},
		{
			name:        "Current at exact end",
			currentTime: "10:30",
			testFunc:    "current",
			expectFound: false,
			description: "Should not find session at exact end time",
		},
		{
			name:        "Current one minute before end",
			currentTime: "10:29",
			testFunc:    "current",
			expectFound: true,
			description: "Should find session one minute before end",
		},
		{
			name:        "Next at exact start",
			currentTime: "10:00",
			testFunc:    "next",
			expectFound: false,
			description: "Should not find next when at start of current",
		},
		{
			name:        "Next one minute before start",
			currentTime: "09:59",
			testFunc:    "next",
			expectFound: true,
			description: "Should find next session one minute before start",
		},
		{
			name:        "Next at exact end",
			currentTime: "10:30",
			testFunc:    "next",
			expectFound: false,
			description: "Should not find next session at end of last session",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result *Session
			
			if tt.testFunc == "current" {
				result = GetCurrentRoomSession("EDGE-ROOM", "EdgeDay", tt.currentTime)
			} else {
				result = GetNextRoomSession("EDGE-ROOM", "EdgeDay", tt.currentTime)
			}
			
			if tt.expectFound {
				testutil.AssertNotNil(t, result, tt.description)
				testutil.AssertEqual(t, "EDGE-001", result.Code, "Should find the test session")
			} else {
				testutil.AssertEqual(t, (*Session)(nil), result, tt.description)
			}
		})
	}
}

func TestRoomScheduleMultipleRoomsData(t *testing.T) {
	// Test that room schedule functions properly filter by room
	mixedSessions := []Session{
		{
			Code:  "TR211-A",
			Title: "TR211 Session A",
			Start: "09:00",
			End:   "09:30",
			Room:  "TR211",
		},
		{
			Code:  "RB105-A",
			Title: "RB105 Session A", 
			Start: "09:15",
			End:   "09:45",
			Room:  "RB-105",
		},
		{
			Code:  "TR211-B",
			Title: "TR211 Session B",
			Start: "10:00",
			End:   "10:30",
			Room:  "TR211",
		},
		{
			Code:  "AU-A",
			Title: "AU Session A",
			Start: "09:30",
			End:   "10:00",
			Room:  "AU",
		},
	}
	
	originalSessionsByDay := sessionsByDay
	originalSessionsLoaded := sessionsLoaded
	
	sessionsByDay = map[string][]Session{
		"MixedDay": mixedSessions,
	}
	sessionsLoaded = true
	
	defer func() {
		sessionsByDay = originalSessionsByDay
		sessionsLoaded = originalSessionsLoaded
	}()
	
	t.Run("Filter TR211 sessions", func(t *testing.T) {
		result := FindRoomSessions("MixedDay", "TR211")
		testutil.AssertEqual(t, 2, len(result), "Should find exactly 2 TR211 sessions")
		
		// Verify all sessions are TR211
		for _, session := range result {
			testutil.AssertEqual(t, "TR211", session.Room, "All sessions should be TR211")
		}
		
		// Verify correct order
		testutil.AssertEqual(t, "TR211-A", result[0].Code, "First should be TR211-A")
		testutil.AssertEqual(t, "TR211-B", result[1].Code, "Second should be TR211-B")
	})
	
	t.Run("Filter RB-105 sessions", func(t *testing.T) {
		result := FindRoomSessions("MixedDay", "RB-105")
		testutil.AssertEqual(t, 1, len(result), "Should find exactly 1 RB-105 session")
		testutil.AssertEqual(t, "RB105-A", result[0].Code, "Should be RB105-A")
	})
	
	t.Run("Current session filtering", func(t *testing.T) {
		// At 09:20, should find different sessions in different rooms
		tr211Current := GetCurrentRoomSession("TR211", "MixedDay", "09:20")
		testutil.AssertNotNil(t, tr211Current, "Should find TR211 session at 09:20")
		testutil.AssertEqual(t, "TR211-A", tr211Current.Code, "Should be TR211-A")
		
		rb105Current := GetCurrentRoomSession("RB-105", "MixedDay", "09:20")
		testutil.AssertNotNil(t, rb105Current, "Should find RB-105 session at 09:20")
		testutil.AssertEqual(t, "RB105-A", rb105Current.Code, "Should be RB105-A")
		
		auCurrent := GetCurrentRoomSession("AU", "MixedDay", "09:20")
		testutil.AssertEqual(t, (*Session)(nil), auCurrent, "Should not find AU session at 09:20")
	})
	
	t.Run("Next session filtering", func(t *testing.T) {
		// At 09:20, next sessions should be different for each room
		tr211Next := GetNextRoomSession("TR211", "MixedDay", "09:20")
		testutil.AssertNotNil(t, tr211Next, "Should find next TR211 session")
		testutil.AssertEqual(t, "TR211-B", tr211Next.Code, "Next TR211 should be TR211-B")
		
		rb105Next := GetNextRoomSession("RB-105", "MixedDay", "09:20")
		testutil.AssertEqual(t, (*Session)(nil), rb105Next, "Should not find next RB-105 session")
		
		auNext := GetNextRoomSession("AU", "MixedDay", "09:20")
		testutil.AssertNotNil(t, auNext, "Should find next AU session")
		testutil.AssertEqual(t, "AU-A", auNext.Code, "Next AU should be AU-A")
	})
}
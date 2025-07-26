package testutil

import (
	"time"
)

// Session represents a test session (duplicated from main package for testing)
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
	Room       string   `json:"room"`
	Day        string   `json:"day"`
	Tags       []string `json:"tags"`
}

// UserState represents test user state
type UserState struct {
	SessionID    string    `json:"session_id"`
	Day          string    `json:"day"`
	Schedule     []Session `json:"schedule"`
	LastEndTime  string    `json:"last_end_time"`
	Profile      []string  `json:"profile"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
}

// CreateTestSessions returns a set of test sessions for unit tests
func CreateTestSessions() []Session {
	return []Session{
		{
			Code:       "TEST001",
			Title:      "Test Session 1",
			Speakers:   []string{"Speaker A"},
			Start:      "09:00",
			End:        "09:30",
			Track:      "Test Track",
			Abstract:   "Test abstract 1",
			Language:   "英語",
			Difficulty: "入門",
			Room:       "AU",
			Day:        "Aug.10",
			Tags:       []string{"AI"},
		},
		{
			Code:       "TEST002",
			Title:      "Test Session 2",
			Speakers:   []string{"Speaker B"},
			Start:      "10:00",
			End:        "10:30",
			Track:      "Test Track",
			Abstract:   "Test abstract 2",
			Language:   "漢語",
			Difficulty: "中階",
			Room:       "RB-105",
			Day:        "Aug.10",
			Tags:       []string{"Languages"},
		},
		{
			Code:       "TEST003",
			Title:      "Test Session 3",
			Speakers:   []string{"Speaker C", "Speaker D"},
			Start:      "11:00",
			End:        "11:30",
			Track:      "Test Track",
			Abstract:   "Test abstract 3",
			Language:   "英語",
			Difficulty: "進階",
			Room:       "TR405",
			Day:        "Aug.10",
			Tags:       []string{"System"},
		},
	}
}

// CreateTestUserState creates a test user state with given sessions
func CreateTestUserState(sessionID, day string, sessions []Session, lastEndTime string) *UserState {
	return &UserState{
		SessionID:    sessionID,
		Day:          day,
		Schedule:     sessions,
		LastEndTime:  lastEndTime,
		Profile:      []string{"Test Track"},
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}
}

// CreateEmptyUserState creates a user state with no scheduled sessions
func CreateEmptyUserState(sessionID, day string) *UserState {
	return &UserState{
		SessionID:    sessionID,
		Day:          day,
		Schedule:     []Session{},
		LastEndTime:  "",
		Profile:      []string{},
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}
}
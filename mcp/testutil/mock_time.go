package testutil

import "time"

// MockTimeProvider implements TimeProvider for testing
type MockTimeProvider struct {
	fixedTime time.Time
}

// Now returns the mocked time
func (m *MockTimeProvider) Now() time.Time {
	return m.fixedTime
}

// NewMockTimeProvider creates a new mock time provider with Aug9 10:23
func NewMockTimeProvider(timeStr string) *MockTimeProvider {
	// Parse time string like "10:23" and create a Aug9 time
	parsedTime, err := time.Parse("15:04", timeStr)
	if err != nil {
		// Default to 10:23 if parsing fails
		parsedTime, _ = time.Parse("15:04", "10:23")
	}
	
	// Set to COSCUP 2025 Aug 9 with the specified time
	fixedTime := time.Date(2025, 8, 9, parsedTime.Hour(), parsedTime.Minute(), 0, 0, time.UTC)
	return &MockTimeProvider{fixedTime: fixedTime}
}

// NewMockTimeProviderWithDay creates a new mock time provider with custom day
func NewMockTimeProviderWithDay(timeStr, day string) *MockTimeProvider {
	parsedTime, err := time.Parse("15:04", timeStr)
	if err != nil {
		parsedTime, _ = time.Parse("15:04", "10:23")
	}
	
	var fixedTime time.Time
	switch day {
	case "Aug9":
		fixedTime = time.Date(2025, 8, 9, parsedTime.Hour(), parsedTime.Minute(), 0, 0, time.UTC)
	case "Aug10":
		fixedTime = time.Date(2025, 8, 10, parsedTime.Hour(), parsedTime.Minute(), 0, 0, time.UTC)
	default:
		// Outside COSCUP period - use 2025/8/8 as example
		fixedTime = time.Date(2025, 8, 8, parsedTime.Hour(), parsedTime.Minute(), 0, 0, time.UTC)
	}
	
	return &MockTimeProvider{fixedTime: fixedTime}
}

// NewMockTimeProviderWithTime creates a mock time provider with exact time.Time
func NewMockTimeProviderWithTime(t time.Time) *MockTimeProvider {
	return &MockTimeProvider{fixedTime: t}
}

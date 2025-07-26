package testutil

// MockTimeProvider implements TimeProvider for testing
type MockTimeProvider struct {
	MockTime string
	MockDay  string
}

// GetCurrentTime returns the mocked time
func (m *MockTimeProvider) GetCurrentTime() string {
	return m.MockTime
}

// GetCurrentDay returns the mocked day
func (m *MockTimeProvider) GetCurrentDay() string {
	if m.MockDay == "" {
		return "Aug9" // default to Aug9
	}
	return m.MockDay
}

// NewMockTimeProvider creates a new mock time provider
func NewMockTimeProvider(time string) *MockTimeProvider {
	return &MockTimeProvider{MockTime: time, MockDay: "Aug9"}
}

// NewMockTimeProviderWithDay creates a new mock time provider with custom day
func NewMockTimeProviderWithDay(time, day string) *MockTimeProvider {
	return &MockTimeProvider{MockTime: time, MockDay: day}
}
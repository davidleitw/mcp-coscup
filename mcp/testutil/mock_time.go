package testutil

// MockTimeProvider implements TimeProvider for testing
type MockTimeProvider struct {
	MockTime string
}

// GetCurrentTime returns the mocked time
func (m *MockTimeProvider) GetCurrentTime() string {
	return m.MockTime
}

// NewMockTimeProvider creates a new mock time provider
func NewMockTimeProvider(time string) *MockTimeProvider {
	return &MockTimeProvider{MockTime: time}
}
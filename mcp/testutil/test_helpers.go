package testutil

import (
	"testing"
)

// AssertEqual checks if two values are equal
func AssertEqual(t *testing.T, expected, actual interface{}, message string) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s: expected %v, got %v", message, expected, actual)
	}
}

// AssertNotNil checks if value is not nil
func AssertNotNil(t *testing.T, value interface{}, message string) {
	t.Helper()
	if value == nil {
		t.Errorf("%s: expected non-nil value", message)
	}
}

// AssertNil checks if value is nil
func AssertNil(t *testing.T, value interface{}, message string) {
	t.Helper()
	if value != nil {
		// Handle typed nil pointers
		switch v := value.(type) {
		case *interface{}:
			if v != nil {
				t.Errorf("%s: expected nil, got %v", message, value)
			}
		default:
			t.Errorf("%s: expected nil, got %v", message, value)
		}
	}
}

// AssertError checks if error is not nil
func AssertError(t *testing.T, err error, message string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected error but got nil", message)
	}
}

// AssertNoError checks if error is nil
func AssertNoError(t *testing.T, err error, message string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: expected no error but got: %v", message, err)
	}
}

// AssertContains checks if slice contains expected value
func AssertContains(t *testing.T, slice []string, expected string, message string) {
	t.Helper()
	for _, item := range slice {
		if item == expected {
			return
		}
	}
	t.Errorf("%s: slice %v does not contain %s", message, slice, expected)
}

// AssertSliceEqual checks if two slices are equal
func AssertSliceEqual(t *testing.T, expected, actual []string, message string) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Errorf("%s: slice lengths differ - expected %d, got %d", message, len(expected), len(actual))
		return
	}
	for i := range expected {
		if expected[i] != actual[i] {
			t.Errorf("%s: slices differ at index %d - expected %s, got %s", message, i, expected[i], actual[i])
			return
		}
	}
}

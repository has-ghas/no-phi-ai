package az

import "testing"

func TestNewEntityDetectionAI(t *testing.T) {
	// Test case 1: Valid service and key
	engine, err := NewEntityDetectionAI("https://example.com", "valid-key")
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
	if engine == nil {
		t.Error("Expected engine to be created, but got nil")
	}

	// Test case 2: Empty key
	engine, err = NewEntityDetectionAI("https://example.com", "")
	if err == nil {
		t.Error("Expected error, but got nil")
	}
	if engine != nil {
		t.Error("Expected engine to be nil, but got non-nil")
	}
	if err.Error() != "EntityDetectionAI requires a valid authentication key" {
		t.Errorf("Expected error message: 'EntityDetectionAI requires a valid authentication key', but got: %v", err.Error())
	}

	// Test case 3: Empty service
	engine, err = NewEntityDetectionAI("", "valid-key")
	if err == nil {
		t.Error("Expected error, but got nil")
	}
	if engine != nil {
		t.Error("Expected engine to be nil, but got non-nil")
	}
	if err.Error() != "EntityDetectionAI requires a valid service address" {
		t.Errorf("Expected error message: 'EntityDetectionAI requires a valid service address', but got: %v", err.Error())
	}
}

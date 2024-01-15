package server

import (
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	// set the path to the test config file
	os.Setenv("NOPHI_CONFIG_PATH", "../../config/test.yml.example")

	// Call the Run function
	err := Run()
	// error expected due to 0 value for github.integration_id
	if err == nil {
		t.Fatal("expected error, but got nil")
	}

	// TODO: Add more assertions to test the behavior of the Run function
}

package server

import (
	"net/http"
	//"net/http/httptest"
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	// Create a new HTTP request to simulate the server
	_, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// set the path to the test config file
	os.Setenv("NOPHI_CONFIG_PATH", "../../config/test.yml.example")

	// Call the Run function
	err = Run()
	// error expected due to 0 value for github.integration_id
	if err == nil {
		t.Fatal("expected error, but got nil")
	}

	/*
		// Create a new response recorder to record the response
		rr := httptest.NewRecorder()

		// Check the status code of the response
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status code %d, but got %d", http.StatusOK, rr.Code)
		}
	*/

	// TODO: Add more assertions to test the behavior of the Run function
}

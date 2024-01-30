package scanner

import (
	"time"

	"github.com/google/uuid"
)

// ScanObject struct is the base struct for all types of scanned objects
// and provides a common set of fields and methods for tracking the status
// of a scan for some uniquely identifiable object.
type ScanObject struct {
	// ID is the unique identifier for the object.
	ID string `json:"id"`
	// Name is the friendly name of the object, such as the name
	// of the file, organization, repository, etc.
	Name string `json:"name"`
	// Status tracks the different statuses of the object scan.
	Status *ScanStatus `json:"status"`
	// Type can be one of:
	//   - "commit"
	//   = "issue_comment"
	//   - "issue"
	//   - "file"
	//   = "organization"
	//   = "pull_request_comment"
	//   - "pull_request"
	//   - "repository"
	Type string `json:"type"`
	// The unique URL associated with the object.
	URL string `json:"url"`
}

// NewScanObject() function initializes a new ScanObject struct with
// the provided name, type, and URL. Sets default values for the
// Status of the ScanObject.
func NewScanObject(name, object_type, url string) *ScanObject {
	return &ScanObject{
		ID:     uuid.New().String(),
		Name:   name,
		Status: NewScanStatus(),
		Type:   object_type,
		URL:    url,
	}
}

// ScanStatus struct is used to track the status of a scan for the
// associated ScanObject, where ScanStatus is embedded in ScanObject.
type ScanStatus struct {
	Completed   bool  `json:"completed"`
	CompletedAt int64 `json:"completed_at"`
	Errored     bool  `json:"errored"`
	ErroredAt   int64 `json:"errored_at"`
	Started     bool  `json:"started"`
	StartedAt   int64 `json:"started_at"`
}

// NewScanStatus() function initializes a new ScanStatus struct with
// default values, meaning that booleans are set to false and timestamps
// are set to 0.
func NewScanStatus() *ScanStatus {
	return &ScanStatus{
		Completed:   false,
		CompletedAt: 0,
		Errored:     false,
		ErroredAt:   0,
		Started:     false,
		StartedAt:   0,
	}
}

// SetCompleted() method checks if the scan status is already set to
// completed and, if not, sets the Completed and CompletedAt fields
// and resets the Errored and ErroredAt fields.
func (s *ScanStatus) SetCompleted() {
	if !s.Completed {
		s.Completed = true
		s.CompletedAt = time.Now().Unix()
		s.Errored = false
		s.ErroredAt = 0
		// any scan that is completed is also started
		s.Started = true
	}
}

// SetErrored() method sets the Errored and ErroredAt fields to indicate
// an error occurred during the scan of the associated object and resets
// the Completed and CompletedAt fields.
func (s *ScanStatus) SetErrored() {
	s.Completed = false
	s.CompletedAt = 0
	s.Errored = true
	s.ErroredAt = time.Now().Unix()
}

// SetStarted() method checks if the scan status is already set to
// started and, if not, sets the Started and StartedAt fields.
func (s *ScanStatus) SetStarted() {
	if !s.Started {
		s.Started = true
		s.StartedAt = time.Now().Unix()
	}
}

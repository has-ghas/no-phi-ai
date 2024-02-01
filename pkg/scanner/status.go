package scanner

import "time"

// Status struct is used to track the status of a scan for the
// associated ScanObject, where Status is embedded in ScanObject.
type Status struct {
	Completed   bool   `json:"completed"`
	CompletedAt int64  `json:"completed_at"`
	ErrorMsg    string `json:"error"`
	Errored     bool   `json:"errored"`
	ErroredAt   int64  `json:"errored_at"`
	Started     bool   `json:"started"`
	StartedAt   int64  `json:"started_at"`
}

// NewStatus() function initializes a new Status struct with
// default values, meaning that booleans are set to false and timestamps
// are set to 0.
func NewStatus() *Status {
	return &Status{
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
func (s *Status) SetCompleted() {
	if !s.Completed {
		s.Completed = true
		s.CompletedAt = time.Now().Unix()
		s.Errored = false
		s.ErroredAt = 0
		s.ErrorMsg = ""
		// any scan that is completed is also started
		s.Started = true
	}
}

// SetErrored() method sets the Errored and ErroredAt fields to indicate
// an error occurred during the scan of the associated object and resets
// the Completed and CompletedAt fields.
func (s *Status) SetErrored(err_msg string) {
	s.Completed = false
	s.CompletedAt = 0
	s.ErrorMsg = err_msg
	s.Errored = true
	s.ErroredAt = time.Now().Unix()
}

// SetStarted() method checks if the scan status indicates that the scan
// has started and, if not, sets the Started and StartedAt fields.
func (s *Status) SetStarted() {
	if !s.Started {
		s.Started = true
		s.StartedAt = time.Now().Unix()
	}
}

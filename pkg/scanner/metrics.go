package scanner

import "sync"

// ScanMetrics struct is the top-level struct for capturing the metrics
// from some scan run, where the metrics are broken down by the type of
// object scanned (e.g. repositories, commits, files, documents, etc.).
type ScanMetrics struct {
	Commits      ScanMetricsObjectMetrics `json:"commits"`
	Documents    ScanMetricsObjectMetrics `json:"documents"`
	Files        ScanMetricsObjectMetrics `json:"files"`
	Repositories ScanMetricsObjectMetrics `json:"repositories"`

	mu *sync.RWMutex
}

// NewScanMetrics() function initializes a new ScanMetrics object for
// use in tracking metrics across all types of scanned objects.
func NewScanMetrics() *ScanMetrics {
	return &ScanMetrics{
		Commits:      NewScanMetricsObjectMetrics(ScanObjectTypeCommit),
		Documents:    NewScanMetricsObjectMetrics(ScanObjectTypeDocument),
		Files:        NewScanMetricsObjectMetrics(ScanObjectTypeFile),
		Repositories: NewScanMetricsObjectMetrics(ScanObjectTypeRepository),
		mu:           &sync.RWMutex{},
	}
}

// SetCommits() method sets metrics for scanned commits.
func (sm *ScanMetrics) Set(object_type string, metrics ScanMetricsObjectMetrics) (e error) {
	// lock the mutex to ensure that the metrics are not being updated
	// by a concurrent goroutine
	sm.mu.Lock()
	// unlock the mutex when the function returns
	defer sm.mu.Unlock()

	// update the metrics for the specified object type
	switch metrics.Type {
	case ScanObjectTypeCommit:
		sm.Commits = metrics
		return
	case ScanObjectTypeDocument:
		sm.Documents = metrics
		return
	case ScanObjectTypeFile:
		sm.Files = metrics
		return
	case ScanObjectTypeRepository:
		sm.Repositories = metrics
		return
	default:
		e = ErrScanMetricsInvalidObjectType
		return
	}
}

// ScanMetricsObjectMetrics struct wraps the metrics for a single type
// of scanned object.
type ScanMetricsObjectMetrics struct {
	Results ScanMetricsObjectResults `json:"results"`
	Status  ScanMetricsObjectStatus  `json:"status"`
	Type    string                   `json:"type"`
}

// NewScanMetricsObjectMetrics() function initializes a new
// ScanMetricsObjectMetrics object for the provided object type.
func NewScanMetricsObjectMetrics(object_type string) ScanMetricsObjectMetrics {
	return ScanMetricsObjectMetrics{
		Results: NewScanMetricsObjectResults(),
		Status:  NewScanMetricsObjectStatus(),
		Type:    object_type,
	}
}

// ScanMetricsObjectResults struct captures metrics related to the
// aggregated scan RESULTS a single object type. These metrics provide
// details about whether the scan found any PHI/PII data associated
// with the object, either in the text associated with the object or
// for some child(ren) of the object.
type ScanMetricsObjectResults struct {
	Clean   int `json:"clean,omitempty"`
	Dirty   int `json:"dirty,omitempty"`
	Error   int `json:"error,omitempty"`
	Unknown int `json:"unknown,omitempty"`
}

// NewScanMetricsObjectResults() function initializes a new
// ScanMetricsObjectResults object with all metrics set to zero.
func NewScanMetricsObjectResults() ScanMetricsObjectResults {
	return ScanMetricsObjectResults{
		Clean:   0,
		Dirty:   0,
		Error:   0,
		Unknown: 0,
	}
}

// ScanMetricsObjectStatus struct captures metrics related to the
// aggregated scan STATUS of a single object type. These metrics provide
// details about how well the scan is progressing and if there are any
// issues with the scan process.
type ScanMetricsObjectStatus struct {
	Completed   int `json:"completed,omitempty"`
	Errored     int `json:"errored,omitempty"`
	Ignored     int `json:"ignored,omitempty"`
	Initialized int `json:"initialized,omitempty"`
	Requested   int `json:"requested,omitempty"`
	Responded   int `json:"responded,omitempty"`
	Started     int `json:"started,omitempty"`
}

// NewScanMetricsObjectStatus() function initializes a new
// ScanMetricsObjectStatus object with all metrics set to zero.
func NewScanMetricsObjectStatus() ScanMetricsObjectStatus {
	return ScanMetricsObjectStatus{
		Completed:   0,
		Errored:     0,
		Ignored:     0,
		Initialized: 0,
		Requested:   0,
		Responded:   0,
		Started:     0,
	}
}

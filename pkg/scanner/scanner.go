package scanner

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/has-ghas/no-phi-ai/pkg/cfg"
	"github.com/has-ghas/no-phi-ai/pkg/client/az"
	nogit "github.com/has-ghas/no-phi-ai/pkg/client/no-git"
)

// scannerContext is global variable for storing the context used by the
// Scanner and the objects it creates, where the context must be set before
// iterating over commits and files in a repository.
var scannerContext context.Context

// Scanner struct provides a management wrapper for scanning a GitHub
// organization and/or a set of git repositories for PHI/PII.
//
// Scanner allows for scanning multiple repositories concurrently while
// tracking the status of each scan via a map of ScanTracker objects.
type Scanner struct {
	ID string `json:"id"`

	ai               *az.EntityDetectionAI
	channelDocuments chan az.AsyncDocumentWrapper
	channelQuit      chan error
	gitConfig        *cfg.GitConfig
	git              *nogit.GitManager
	scans            map[string]*ScanTracker
}

// NewScanner() function initializes a new Scanner instance.
func NewScanner(config *cfg.Config, ctx context.Context, logger *zerolog.Logger) (*Scanner, error) {
	// ensure that the context is not nil
	if ctx == nil {
		ctx = context.Background()
	}

	// set the global scannerContext variable
	scannerContext = ctx

	// create a common *az.EntityDetectionAI, which can be used for detecting
	// "entities" of interest within text documents submitted to the the API
	// for the Azure AI Language service
	ai, ai_err := az.NewEntityDetectionAI(config)
	if ai_err != nil {
		return nil, ai_err
	}

	return &Scanner{
		ID: uuid.NewString(),
		ai: ai,
		// create the channel used to allow ScanObject instances to
		// communicate by sending documents to be scanned
		channelDocuments: make(chan az.AsyncDocumentWrapper, config.Git.Scan.Limits.MaxRequestsOutstanding),
		// create the channel used to interrupt the scan
		channelQuit: make(chan error),
		git:         nogit.NewGitManager(&config.Git, scannerContext, logger),
		gitConfig:   &config.Git,
		scans:       make(map[string]*ScanTracker),
	}, nil
}

// ScanReposForPHI() method scans the configured GitHub organization and/or
// git repositories for PHI/PII data. Returns an error if any part of the
// scan fails.
func (s *Scanner) ScanReposForPHI() (e error) {
	var scan_id string
	// initialize a new ScanTracker for the scan
	scan_id, e = s.initScanTracker()
	if e != nil {
		return
	}

	// run the scan
	if e = s.runScanTracker(scan_id); e != nil {
		return
	}

	return
}

// initScanTracker() method initializes a new ScanTracker object as preparation
// for running a new scan, adds the ScanTracker to the map of scans, and returns
// the ID of the ScanTracker object. Returns a non-nil error if any failure is
// encountered in the initialization process.
func (s *Scanner) initScanTracker() (scan_id string, e error) {
	if s.git == nil {
		e = errors.New("faled to init scan tracker because git manager is nil")
		return
	}
	channel_documents := s.channelDocuments
	channel_quit := s.channelQuit

	// TODO : derive the actual list of repositories from a combination
	//        of config and discovery of organization repositories
	repo_list := s.gitConfig.Scan.Repositories

	// initialize a new ScanTracker object
	scan_tracker, err := NewScanTracker(
		s.git,
		s.gitConfig.Scan.Organization,
		repo_list,
		channel_documents,
		channel_quit,
	)
	if err != nil {
		e = err
		return
	}
	// set the returned scan_id to the ID of the scan tracker
	scan_id = scan_tracker.ID
	// add the scan tracker to the map of scans
	s.scans[scan_id] = scan_tracker

	return
}

// getScanTracker() method allows for looking up a ScanTracker object by its ID.
func (s *Scanner) getScanTracker(scan_id string) (st *ScanTracker, e error) {
	var exists bool
	st, exists = s.scans[scan_id]
	if !exists {
		e = errors.New("failed to get scan tracker : scan not found for ID = " + scan_id)
		return
	}
	return
}

// runScanTracker() method runs the scan for the ScanTracker object with the
// provided ID and returns an error if any part of the scan fails.
func (s *Scanner) runScanTracker(scan_id string) (e error) {
	var scan_tracker *ScanTracker
	// get the ScanTracker object for the provided ID
	scan_tracker, e = s.getScanTracker(scan_id)
	if e != nil {
		return
	}

	wg := &sync.WaitGroup{}
	// add a slot to the wait group to allow for s.ai.ScanDocuments() to run
	// in the background until the all documents have been processed
	wg.Add(1)
	// use a separate goroutine to begin listening on the s.channelDocuments
	// for documents to scan while allowing the scan to be interrupted via
	// quit channel or context cancellation
	go s.ai.ScanDocuments(wg, scannerContext, s.channelDocuments, s.channelQuit)

	// create a channel to receive errors from the scan tracker
	err_chan := make(chan error)
	// while listening for documents to be created from the scan, start running
	// the scan of the configured GitHub organization and/or git repositories;
	if e = scan_tracker.Scan(); e != nil {
		// return a non-nil error if any part of the scan fails to run
		e = errors.Wrap(e, "error running scan tracker")
		return
	}

	// add a slot to the wait group to allow for tracking the scan
	// until completed or interrupted
	wg.Add(1)
	// track the progress of the scan
	go scan_tracker.Track(wg, scannerContext, err_chan)

	// wait for the wait group to be marked as done before completing this
	// run of the scan tracker
	wg.Wait()
	return
}

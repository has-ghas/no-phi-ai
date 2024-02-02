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

// Scanner struct provides a management wrapper for scanning a GitHub
// organization and/or a set of git repositories for PHI/PII.
//
// Scanner allows for scanning multiple repositories concurrently while
// tracking the status of each scan via a map of ScanTracker objects.
type Scanner struct {
	ID string `json:"id"`

	channelDocuments chan az.AsyncDocumentWrapper
	channelQuit      chan error
	config           *cfg.GitConfig
	ctx              context.Context
	git              *nogit.GitManager
	logger           *zerolog.Logger
	scans            map[string]*ScanTracker
}

// NewScanner() function initializes a new Scanner instance.
func NewScanner(config *cfg.GitConfig, ctx context.Context, logger *zerolog.Logger) *Scanner {
	// ensure that the context is not nil
	if ctx == nil {
		ctx = context.Background()
	}

	return &Scanner{
		ID: uuid.NewString(),
		// create the channel used to allow ScanObject instances to
		// communicate by sending documents to be scanned
		channelDocuments: make(chan az.AsyncDocumentWrapper, config.Scan.Limits.MaxRequestsOutstanding),
		// create the channel used to interrupt the scan
		channelQuit: make(chan error),
		config:      config,
		ctx:         ctx,
		git:         nogit.NewGitManager(config, ctx, logger),
		logger:      logger,
		scans:       make(map[string]*ScanTracker),
	}
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
	if s.logger == nil {
		e = errors.New("failed to init scan tracker because logger is nil")
		return
	}
	channel_documents := s.channelDocuments
	channel_quit := s.channelQuit

	// TODO : derive the actual list of repositories from a combination
	//        of config and discovery of organization repositories
	repo_list := s.config.Scan.Repositories

	// initialize a new ScanTracker object
	scan_tracker, err := NewScanTracker(
		s.ctx,
		s.git,
		s.logger,
		s.config.Scan.Organization,
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

func (s *Scanner) test(wait_group *sync.WaitGroup) {
	// TODO : remove TRACE
	s.logger.Trace().Msg("Scanner.test() : marking wait group as done")
	wait_group.Done()
}

// processDocuments() method processes documents received from the s.channelDocuments
// channel. This method is intended to be run as a separate goroutine.
func (s *Scanner) processDocuments(wait_group *sync.WaitGroup) {
	defer s.test(wait_group)
	// TODO : remove doc_count
	doc_count := 0
	// TODO : remove test_queue
	test_queue := make(chan az.AsyncDocumentWrapper, s.config.Scan.Limits.MaxRequestsOutstanding)
	for {
		select {
		case <-s.ctx.Done():
			s.logger.Warn().Msg("Scanner.processDocuments() : context done")
			return
		case <-s.channelQuit:
			s.logger.Warn().Msg("Scanner.processDocuments() : received quit signal")
			return
		case doc := <-s.channelDocuments:
			// TODO : remoe doc_count
			// increment the count of documents received
			doc_count++
			// TODO : remove TRACE
			s.logger.Trace().Msgf("Scanner.processDocuments() : received document #%d : ID = %s", doc_count, doc.Document.ID)
			// TODO : remove test_queue
			// write the document to the test queue
			test_queue <- doc
			// TODO : remove TRACE
			s.logger.Trace().Msgf("Scanner.processDocuments() : sent document #%d  to queue : ID = %s", doc_count, doc.Document.ID)

			// send the document to the Azure AI Language service API for analysis
			// TODO

			// wait for the document response to be received from the API
			// TODO
		}
	}
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
	// setup a wait group to allow for s.processDocuments() to run in the
	// background until the all documents have been processed
	wg := &sync.WaitGroup{}
	wg.Add(1)
	// use a separate goroutine to begin listening on the s.channelDocuments
	// for documents to scan while allowing the scan to be interrupted via
	// quit channel or context cancellation
	go s.processDocuments(wg)
	// while listening for documents to be created from the scan, start running
	// the scan of the configured GitHub organization and/or git repositories;
	//
	// return a non-nil error if any part of the scan fails to run
	if e = scan_tracker.Scan(); e != nil {
		e = errors.Wrap(e, "error running scan tracker")
		return
	}
	// wait for the wait group as done to allow this function to return
	wg.Wait()
	return
}

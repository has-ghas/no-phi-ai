package scanner

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

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

// runScanTracker() method uses the ScanTracker associated with scan_id to run
// the scan of the configured GitHub organization and/or git repositories for
// PHI/PII data and tracks the progress of the scan. Returns a non-nil error if
// unable to find a ScanTracker for the provided ID.
func (s *Scanner) runScanTracker(scan_id string) (e error) {
	var scan_tracker *ScanTracker
	// get the ScanTracker object for the provided ID
	scan_tracker, e = s.getScanTracker(scan_id)
	if e != nil {
		return
	}

	// create a channel to receive errors from goroutines
	err_chan := make(chan error)
	// close the error channel before returning
	defer close(err_chan)
	// create the channels used to signal the completion of scan phases
	done_chan_1 := make(chan bool)
	done_chan_2 := make(chan bool)
	done_chan_3 := make(chan bool)
	done_chan_4 := make(chan bool)
	// use separate goroutines to:
	//   1. scan documents generated by the scan tracker;
	//   2. generate documents to be scanned and tracked;
	//   3. track the progress of objects in the scan tracker;
	//   4. listen for errors send from other goroutines.
	go scan_tracker.Scan(
		scannerContext, // cancel context to stop immediately
		err_chan,       // sends errors on err_chan
		done_chan_1,    // sends quit signal on done_chan_1
	)
	go s.ai.ScanReceiver(
		scannerContext,     // cancel context to stop immediately
		s.channelDocuments, // receives documents on s.channelDocuments
		done_chan_1,        // receives quit signal to stop gracefully
		err_chan,           // sends errors on err_chan
		done_chan_2,        // sends done signal on done_chan_2
	)
	go scan_tracker.Track(
		scannerContext, // cancel the context to stop immediately
		done_chan_2,    // receives quit signal to stop gracefully
		err_chan,       // sends errors on err_chan
		done_chan_3,    // sends done signal on done_chan_3
	)
	go listenForErrors(
		scannerContext, // context used for logging
		err_chan,       // receives errors to log and evaluate
		done_chan_3,    // receives quit signal on done_chan_3
		done_chan_4,    // sends done signal on done_chan_4
	)

	// wait for done_chan_4 to receive a signal / be closed
	<-done_chan_4

	return
}

// listenForErrors() function listens for errors returned via the
// error channel or the context done channel, and logs any error
// received. Returns when the context is done or the quit channel
// is closed.
func listenForErrors(ctx context.Context, err_chan <-chan error, quit_chan <-chan bool, done_chan chan<- bool) {
	log.Ctx(ctx).Debug().Msg("listening for errors")
	defer close(done_chan)

	select {
	case <-ctx.Done():
		return
	case <-quit_chan:
		log.Ctx(ctx).Debug().Msg("quit signal received : stopping error listener")
		return
	case e := <-err_chan:
		// log the error
		log.Ctx(ctx).Error().Err(e).Msg("error occurred while scanning")

		// evaluate the error and determine whether the scan
		// should be stopped by cancelling the context
		// TODO
	}
}

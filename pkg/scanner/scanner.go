package scanner

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/has-ghas/no-phi-ai/pkg/cfg"
	nogit "github.com/has-ghas/no-phi-ai/pkg/client/no-git"
)

// Scanner struct provides a management wrapper for scanning a GitHub
// organization and/or a set of git repositories for PHI/PII.
//
// Scanner allows for scanning multiple repositories concurrently while
// tracking the status of each scan via a map of ScanTracker objects.
type Scanner struct {
	ID string `json:"id"`

	ctx    context.Context
	config *cfg.GitConfig
	git    *nogit.GitManager
	logger *zerolog.Logger
	scans  map[string]*ScanTracker
}

// NewScanner() function initializes a new Scanner instance.
func NewScanner(config *cfg.GitConfig, ctx context.Context, logger *zerolog.Logger) *Scanner {
	// ensure that the context is not nil
	if ctx == nil {
		ctx = context.Background()
	}

	return &Scanner{
		ID:     uuid.NewString(),
		config: config,
		ctx:    ctx,
		git:    nogit.NewGitManager(config, ctx, logger),
		logger: logger,
		scans:  make(map[string]*ScanTracker),
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
	// TODO : get the actual list of repositories from a combination of
	//        config and discovery of organization repositories
	repo_list := s.config.Scan.Repositories
	// initialize a new ScanTracker object
	scan_tracker, err := NewScanTracker(
		s.ctx,
		s.git,
		s.logger,
		s.config.Scan.Organization,
		repo_list,
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
	// run the scan, let ScanTracker handle the details, and return any error
	e = scan_tracker.Scan()
	return
}

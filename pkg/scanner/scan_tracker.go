package scanner

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/has-ghas/no-phi-ai/pkg/client/az"
	nogit "github.com/has-ghas/no-phi-ai/pkg/client/no-git"
)

// ScanTracker struct is used to store the state of a scan for PHI/PII
// data in git repositories, optionally within a specific GitHub organization.
type ScanTracker struct {
	ID string `json:"id"`

	commits      []*ScanObject
	files        []*ScanObject
	git          *nogit.GitManager
	organization *ScanOrganization
	repositories []*ScanRepository
}

// NewScanTracker() function initializes a new ScanTracker object for
// use in runnging a scan of git repositories while tracking the progress
// of different aspects of the scan.
func NewScanTracker(
	gitManager *nogit.GitManager,
	org_URL string,
	repo_URLs []string,
	channel_documents chan<- az.AsyncDocumentWrapper,
	channel_quit <-chan error,
) (*ScanTracker, error) {
	if scannerContext == nil {
		return nil, ErrScanTrackerContextNil
	}
	if channel_documents == nil {
		return nil, ErrScanTrackerChannelDocumentsNil
	}
	if gitManager == nil {
		return nil, ErrScanTrackerGitManagerNil
	}

	// create a ScanObject for the organization, if provided
	var org_object *ScanOrganization
	if org_URL != "" {
		var org_err error
		org_object, org_err = NewScanOrganization(
			org_URL,
			channel_documents,
			channel_quit,
		)
		if org_err != nil {
			return nil, org_err
		}
	}
	// create a ScanObject for each repository
	repo_objects := make([]*ScanRepository, 0)
	for _, repo_URL := range repo_URLs {
		scan_repo, err := NewScanRepository(
			repo_URL,
			channel_documents,
			channel_quit,
		)
		if err != nil {
			return nil, err
		}
		repo_objects = append(repo_objects, scan_repo)
	}

	return &ScanTracker{
		ID:           uuid.NewString(),
		commits:      []*ScanObject{},
		files:        []*ScanObject{},
		git:          gitManager,
		organization: org_object,
		repositories: repo_objects,
	}, nil
}

// Scan() method runs the scan of the organization and/or repositories.
func (st *ScanTracker) Scan(wg *sync.WaitGroup, ctx context.Context, err_chan chan<- error) {
	defer wg.Done()
	log.Ctx(ctx).Debug().Msg("starting ScanTracker scan")
	// clone the repositories first in order to minimize the number of API
	// calls to GitHub and/or the time wasted via network latency.
	if e := st.scanRepositories(); e != nil {
		// send the error to the error channel
		err_chan <- errors.Wrap(e, "error scanning repositories")
	}
	log.Ctx(scannerContext).Debug().Msg("finished ScanTracker scan")
}

// Track() method tracks the progress of the scan until it is complete,
// or until the scan is interrupted via context cancellation / timeout.
func (st *ScanTracker) Track(wg *sync.WaitGroup, ctx context.Context, err_chan chan<- error) {
	defer wg.Done()
	log.Ctx(ctx).Debug().Msg("starting ScanTracker track")

	// TODO : implement the Track() method

	// wait for a signal from context cancellation
	<-ctx.Done()

	log.Ctx(ctx).Debug().Msg("finished ScanTracker track")
}

// scanRepositories() method clones each repository in ScanTracker.repositories
// into a temporary directory. Returns a non-nil error if any part of the cloning
// process fails.
func (st *ScanTracker) scanRepositories() (e error) {
	// scan each repository in series
	// TODO : allow for scanning repositories in parallel
	for _, scan_repo := range st.repositories {
		// scan the repository for PHI/PII data
		if e = scan_repo.ScanForPHI(st.git); e != nil {
			return
		}
	}
	return
}

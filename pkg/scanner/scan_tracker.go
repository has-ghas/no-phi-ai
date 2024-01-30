package scanner

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	nogit "github.com/has-ghas/no-phi-ai/pkg/client/no-git"
)

// ScanTracker struct is used to store the state of a scan for PHI/PII
// data in git repositories, optionally within a specific GitHub organization.
type ScanTracker struct {
	ID string `json:"id"`

	commits      []*ScanObject
	ctx          context.Context
	files        []*ScanObject
	git          *nogit.GitManager
	logger       *zerolog.Logger
	organization *ScanObject
	repositories []*ScanRepository
}

// NewScanTracker() function initializes a new ScanTracker object for
// use in runnging a scan of git repositories while tracking the progress
// of different aspects of the scan.
func NewScanTracker(
	ctx context.Context,
	gitManager *nogit.GitManager,
	logger *zerolog.Logger,
	org string,
	repo_URLs []string,
) (*ScanTracker, error) {
	if gitManager == nil {
		return nil, errors.New("failed to create scan tracker : gitManager cannot be nil")
	}

	// create a ScanObject for the organization, if provided
	var org_object *ScanObject
	if org != "" {
		org_object = NewScanObject(org, "organization", "")
	}
	// create a ScanObject for each repository
	repo_objects := make([]*ScanRepository, 0)
	for _, repo_URL := range repo_URLs {
		repo_name, err := nogit.ParseRepoNameFromURL(repo_URL)
		if err != nil {
			return nil, err
		}
		repo_objects = append(repo_objects, NewScanRepository(repo_name, repo_URL))
	}

	return &ScanTracker{
		ID:           uuid.NewString(),
		commits:      []*ScanObject{},
		ctx:          ctx,
		files:        []*ScanObject{},
		git:          gitManager,
		logger:       logger,
		organization: org_object,
		repositories: repo_objects,
	}, nil
}

// Scan() method runs the scan of the organization and/or repositories.
func (st *ScanTracker) Scan() (e error) {
	if e = st.preScan(); e != nil {
		return
	}
	// TODO : remove debug logging
	st.logger.Debug().Ctx(st.ctx).Msg("started ScanTracker.Scan()")
	// clone the repositories first in order to minimize the number of API
	// calls to GitHub and/or the time wasted via network latency.
	if e = st.respositoriesClone(); e != nil {
		return
	}
	// TODO : remove debug logging
	st.logger.Debug().Ctx(st.ctx).Msg("finished ScanTracker.Scan()")

	// TODO ; do more stuff after cloning the repositories

	return
}

// preScan() method performs some basic checks to ensure that the ScanTracker
// object is ready to run a scan (and will not panic due to nil pointers).
func (st *ScanTracker) preScan() (e error) {
	if st.ctx == nil {
		e = errors.New("scan tracker failed to run scan : context cannot be nil")
		return
	}
	if st.git == nil {
		e = errors.New("scan tracker failed to run scan : gitManager cannot be nil")
		return
	}
	if st.logger == nil {
		e = errors.New("scan tracker failed to run scan : logger cannot be nil")
		return
	}

	return
}

// respositoriesClone() method clones each repository in ScanTracker.repositories
// into a temporary directory. Returns a non-nil error if any part of the cloning
// process fails.
func (st *ScanTracker) respositoriesClone() (e error) {
	// clone each repository into a temporary directory
	for _, scan_repo := range st.repositories {
		// ensure the scan status shows that the scan has started
		scan_repo.Status.SetStarted()
		// clone the repository
		repo, err := st.git.CloneRepo(scan_repo.URL)
		if err != nil {
			e = err
		}
		// set the repository for the ScanRepository object
		scan_repo.SetRepository(repo)
	}
	// TODO : keep track of the cloned repositories and their scan status
	return
}

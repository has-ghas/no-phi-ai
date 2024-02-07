package scanner

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/has-ghas/no-phi-ai/pkg/client/az"
	nogit "github.com/has-ghas/no-phi-ai/pkg/client/no-git"
)

// ScanTracker struct is used to store the state of a scan for PHI/PII
// data in git repositories, optionally within a specific GitHub organization.
type ScanTracker struct {
	ID      string       `json:"id"`
	Metrics *ScanMetrics `json:"metrics"`

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

	// create a scan object for the organization to scan, if provided
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
	// create a scan object for each repository to be scanned
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
		Metrics:      NewScanMetrics(),
		git:          gitManager,
		organization: org_object,
		repositories: repo_objects,
	}, nil
}

// GetRepositories() method returns the current slice of ScanRepository objects
// associated with (i.e. known to) the scan tracker.
func (st *ScanTracker) GetRepositories() []*ScanRepository {
	return st.repositories
}

// Scan() method runs the scan of the organization and/or repositories.
func (st *ScanTracker) Scan(
	ctx context.Context,
	err_chan chan<- error,
	done_chan chan<- bool,
) {
	log.Ctx(ctx).Debug().Msg("starting ScanTracker scan")
	defer log.Ctx(scannerContext).Debug().Msg("finished ScanTracker scan")
	// close done_channel when done to signal that the scan is complete
	defer close(done_chan)

	// clone the repositories first in order to minimize the number of API
	// calls to GitHub and/or the time wasted via network latency.
	if e := st.scanRepositories(); e != nil {
		// send the error to the error channel
		err_chan <- errors.Wrap(e, "error scanning repositories")
	}
}

// Track() method tracks the progress of the scan until it is complete, or until
// the scan is interrupted via context cancellation / timeout.
func (st *ScanTracker) Track(
	ctx context.Context,
	quit_chan <-chan bool,
	err_chan chan<- error,
	done_chan chan<- bool,
) {
	defer close(done_chan)
	log.Ctx(ctx).Debug().Msg("starting ScanTracker track")
	defer log.Ctx(ctx).Debug().Msg("finished ScanTracker track")

	for {
		select {
		case <-ctx.Done():
			// stop tracking the scan when the context is cancelled
			return
		case <-time.After(time.Duration(ScanMetricsRefreshSeconds) * time.Second):
			// update the scan metrics every ScanMetricsRefreshSeconds
			if err := st.updateScanMetrics(ctx); err != nil {
				// send the error to the error channel
				err_chan <- errors.Wrap(err, "error updating scan metrics")
			}
		case <-quit_chan:
			// update the scan metrics before returning
			if err := st.updateScanMetrics(ctx); err != nil {
				// send the error to the error channel
				err_chan <- errors.Wrap(err, "error updating scan metrics")
			}
			// stop tracking the scan when the done channel is closed
			return
		}
	}
}

// logScanMetrics() method logs the current scan metrics as pretty-printed JSON.
func (st *ScanTracker) logScanMetrics(ctx context.Context) (e error) {
	if st.Metrics == nil {
		err := errors.New("scan metrics are nil")
		log.Ctx(ctx).Error().Err(err).Msg("error logging scan metrics")
		// just log the error and return (e =) nil
		return
	}
	metrics_bytes, err := json.MarshalIndent(st.Metrics, "", "  ")
	if err != nil {
		e = errors.Wrap(err, "error marshalling scan metrics to pretty-printed JSON")
		return
	}
	// log the updated scan metrics
	log.Ctx(ctx).Debug().Msgf("updated scan metrics...\n%s", string(metrics_bytes))
	return
}

// scanRepositories() method clones each repository known to the ScanTracker
// into a temporary directory. Returns a non-nil error if any part of the
// cloning process fails.
func (st *ScanTracker) scanRepositories() (e error) {
	// scan each repository in series
	// TODO : allow for scanning repositories in parallel
	for _, scan_repo := range st.GetRepositories() {
		// scan the repository for PHI/PII data
		if e = scan_repo.ScanForPHI(st.git); e != nil {
			return
		}
	}
	return
}

// setMetrics() method sets the Metrics field of the ScanTracker object to point
// to the provided ScanMetrics object.
func (st *ScanTracker) setMetrics(metrics *ScanMetrics) {
	st.Metrics = metrics
}

// updateScanMetrics() method updates the scan metrics for the scan tracker
func (st *ScanTracker) updateScanMetrics(ctx context.Context) (e error) {
	log.Ctx(ctx).Trace().Msg("updating metrics for scan tracker")

	// create a new scan metrics object for calculating the
	// updated scan metrics from scratch
	scan_metrics := NewScanMetrics()

	// iterate through the repositories in the scan tracker
	for _, repo := range st.GetRepositories() {
		scan_metrics.Repositories.Status.Initialized++
		if repo.Status.IsCompleted() {
			scan_metrics.Repositories.Status.Completed++
		}
		if repo.Status.IsErrored() {
			scan_metrics.Repositories.Status.Errored++
		}
		if repo.Status.IsStarted() {
			scan_metrics.Repositories.Status.Started++
		}

		// iterate through the commits in the repository
		for _, commit := range repo.GetCommits() {
			scan_metrics.Commits.Status.Initialized++
			if commit.Status.IsCompleted() {
				scan_metrics.Commits.Status.Completed++
			}
			if commit.Status.IsErrored() {
				scan_metrics.Commits.Status.Errored++
			}
			if commit.Status.IsStarted() {
				scan_metrics.Commits.Status.Started++
			}

			// iterate through the files in the commit
			for _, file := range commit.GetFiles() {
				scan_metrics.Files.Status.Initialized++
				if file.Status.IsCompleted() {
					scan_metrics.Files.Status.Completed++
				}
				if file.Status.IsIgnored() {
					scan_metrics.Files.Status.Ignored++
					continue
				}
				if file.Status.IsErrored() {
					scan_metrics.Files.Status.Errored++
				}
				if file.Status.IsStarted() {
					scan_metrics.Files.Status.Started++
				}

				var docs_map DocumentTrackerMap
				docs_map, docs_err := file.GetDocuments()
				if docs_err != nil {
					log.Ctx(ctx).Error().Err(docs_err).Msgf("error getting documents for file = %s", file.Name)
					continue
				}

				// iterate through the documents created from the file
				for _, document_tracker := range docs_map {
					scan_metrics.Documents.Status.Initialized++
					if document_tracker.Status.IsCompleted() {
						scan_metrics.Documents.Status.Completed++
					}
					if document_tracker.Status.IsRequested() {
						scan_metrics.Documents.Status.Requested++
					}
					if document_tracker.Status.IsResponded() {
						scan_metrics.Documents.Status.Responded++
					}
					if document_tracker.Status.IsErrored() {
						scan_metrics.Documents.Status.Errored++
					}
					if document_tracker.Status.IsStarted() {
						scan_metrics.Documents.Status.Started++
					}
					// process document results
					if document_tracker.Status.IsResultClean() {
						scan_metrics.Documents.Results.Clean++
					}
					if document_tracker.Status.IsResultDirty() {
						scan_metrics.Documents.Results.Dirty++
					}
					if document_tracker.Status.IsResultError() {
						scan_metrics.Documents.Results.Error++
					}
					if document_tracker.Status.IsResultUnknown() {
						scan_metrics.Documents.Results.Unknown++
					}
				}
			}
		}
	}

	// update the scan metrics for the scan tracker
	st.setMetrics(scan_metrics)

	e = st.logScanMetrics(ctx)

	return
}

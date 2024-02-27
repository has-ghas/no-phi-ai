package scannerv2

import (
	"context"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/has-ghas/no-phi-ai/pkg/cfg"
	nogit "github.com/has-ghas/no-phi-ai/pkg/client/no-git"
)

// ScanRepository struct embeds the ScanObject struct and adds fields
// and methods specific to scanning a git.Repository.
type ScanRepository struct {
	// ID should be the (escaped) URL of the repository.
	ID string `json:"id"`
	// Name is the friendly name of the object, such as the name
	// of the file, organization, repository, etc.
	Name string `json:"name"`
	// The unique URL associated with the object.
	URL string `json:"url"`

	TrackerCommits *KeyTracker
	TrackerFiles   *KeyTracker

	channel_requests chan<- Request
	config           *cfg.GitScanConfig
	ctx              context.Context
	logger           *zerolog.Logger
	repository       *git.Repository
}

// NewScanRepository() function initializes a new ScanRepository object.
func NewScanRepository(
	ctx context.Context,
	url string,
	config *cfg.GitScanConfig,
	channel_requests chan<- Request,
) (*ScanRepository, error) {
	if ctx == nil {
		return nil, errors.Wrap(ErrScanRepositoryContextNil, ErrMsgScanRepositoryCreate)
	}
	if config == nil {
		return nil, errors.Wrap(ErrScanRepositoryConfigNil, ErrMsgScanRepositoryCreate)
	}
	if channel_requests == nil {
		return nil, errors.Wrap(ErrScanRepositoryChannelDocumentsNil, ErrMsgScanRepositoryCreate)
	}

	name, err := nogit.ParseRepoNameFromURL(url)
	if err != nil {
		return nil, errors.Wrap(err, ErrMsgScanRepositoryCreate)
	}

	logger := zerolog.Ctx(ctx)

	// create a new KeyTracker for tracking scanned commits
	tracker_commits, t_err := NewKeyTracker(ScanObjectTypeCommit, logger)
	if t_err != nil {
		return nil, errors.Wrap(t_err, ErrMsgScanRepositoryCreate)
	}
	// create a new KeyTracker for tracking scanned files
	tracker_files, t_err := NewKeyTracker(ScanObjectTypeFile, logger)
	if t_err != nil {
		return nil, errors.Wrap(t_err, ErrMsgScanRepositoryCreate)
	}

	return &ScanRepository{
		ID:               url,
		Name:             name,
		URL:              url,
		channel_requests: channel_requests,
		ctx:              ctx,
		config:           config,
		logger:           logger,
		repository:       nil,
		TrackerCommits:   tracker_commits,
		TrackerFiles:     tracker_files,
	}, nil
}

// GetRepository() method returns a pointer to the git.Repository
// associated with the ScanRepository.
func (sr *ScanRepository) GetRepository() *git.Repository {
	return sr.repository
}

// Scan() method runs the scan of the repository and keeps track of the
// progress of the scan by updating private fields of the ScanRepository.
func (sr *ScanRepository) Scan(gm *nogit.GitManager) (e error) {
	sr.logger.Debug().Msgf("starting scan of repository %s", sr.URL)

	// ensure the repository has been cloned locally and its object is
	// referenced by the ScanRepository.repository field
	if e = sr.clone(gm); e != nil {
		return
	}

	// get an iterator for the commits in the repository
	var commit_iterator object.CommitIter
	commit_iterator, e = sr.repository.CommitObjects()
	if e != nil {
		if commit_iterator != nil {
			commit_iterator.Close()
		}
		return
	}
	defer commit_iterator.Close()

	// iterate through the commits in the repository history
	e = commit_iterator.ForEach(sr.scanCommit)
	if e != nil {
		e = errors.Wrapf(e, "failed to iterate through commits in repository %s", sr.URL)
		// return any error encountered while iterating through the commits
		return
	}

	// print the counts of the scanned commits and files
	sr.TrackerCommits.PrintCounts()
	sr.TrackerFiles.PrintCounts()

	return
}

// clone() method clones the repository from the ScanRepository.URL and
// sets the ScanRepository.repository field to the git.Repository object
// that references the cloned repository.
func (sr *ScanRepository) clone(gm *nogit.GitManager) (e error) {
	if gm == nil {
		e = ErrScanRepositoryCloneGitManagerNil
		return
	}

	var repo *git.Repository
	// clone the repository from the URL
	repo, e = gm.CloneRepo(sr.URL)
	if e != nil {
		e = errors.Wrapf(e, "failed to clone repository from %s", sr.URL)
		return
	}

	// set the ScanRepository.repository field to associate the git.Repository
	e = sr.setRepository(repo)

	return
}

// scanCommit() method scans the tree of the object.Commit for files
// containing any PHI/PII entities.
func (sr *ScanRepository) scanCommit(commit *object.Commit) error {
	update_code, init_err := sr.TrackerCommits.Update(
		commit.Hash.String(),
		KeyCodeInit,
		"",
		[]string{},
	)
	if init_err != nil {
		return errors.Wrapf(init_err, ErrMsgTrackerUpdateCommit, commit.Hash.String())
	}

	// skip commits that have already been scanned
	if update_code > KeyCodeInit {
		sr.logger.Trace().Msgf(
			"repository %s : skipping previously scanned commit %s",
			sr.URL,
			commit.Hash.String(),
		)
		return nil
	}

	sr.logger.Debug().Msgf(
		"repository %s : scanning commit %s",
		sr.URL,
		commit.Hash.String(),
	)

	// get the tree of objects associated with the commit
	tree, err := commit.Tree()
	if err != nil {
		_, err = sr.TrackerCommits.Update(
			commit.Hash.String(),
			KeyCodeError,
			err.Error(),
			[]string{},
		)
		if err != nil {
			return errors.Wrapf(err, ErrMsgTrackerUpdateCommit, commit.Hash.String())
		}
	}

	// iterate through the files in the commit tree
	err = tree.Files().ForEach(sr.scanFile(commit))
	if err != nil {
		err = errors.Wrapf(err, ErrMsgTrackerUpdateCommit, commit.Hash.String())
		sr.TrackerCommits.Update(
			commit.Hash.String(),
			KeyCodeError,
			err.Error(),
			[]string{},
		)
		return err
	}

	// attempt to update the commit code to "complete" status, but ignore any error
	// and accept that the commit may be left in "pending" status if the key has
	// children that are still in an incomplete (bool=false) state
	sr.TrackerCommits.Update(
		commit.Hash.String(),
		KeyCodeComplete,
		"",
		[]string{},
	)

	return nil
}

// scanFile() method returns an anonymous function that can be used to iterate through
// the files in the associated commit tree and scan each file for PHI/PII entities.
func (sr *ScanRepository) scanFile(commit *object.Commit) func(*object.File) error {
	return func(file *object.File) error {
		code, err := sr.TrackerFiles.Update(
			file.Hash.String(),
			KeyCodeInit,
			"",
			[]string{},
		)
		if err != nil {
			return errors.Wrapf(err, ErrMsgScanTrackerUpdateFile, file.Hash.String())
		}
		// skip files that have already been scanned
		if code > KeyCodeInit {
			sr.logger.Trace().Msgf(
				"commit %s : skipping previously scanned file %s : code=%d",
				commit.Hash.String(),
				file.Hash.String(),
				code,
			)
			return nil
		}

		// check if the file should be ignored instead of scanned
		should_ignore, ignore_reason := IgnoreFileObject(
			file,
			sr.config.Extensions,
			sr.config.IgnoreExtensions,
		)
		if should_ignore {
			sr.logger.Trace().Msgf(
				"commit %s : skipping scan of file %s : %s",
				commit.Hash.String(),
				file.Hash.String(),
				ignore_reason,
			)
			_, err = sr.TrackerFiles.Update(
				file.Hash.String(),
				KeyCodeIgnore,
				ignore_reason,
				[]string{},
			)
			return err
		}
		if ignore_reason != "" {
			sr.logger.Warn().Msgf(
				"commit %s : file %s : ignore reason => %s",
				commit.Hash.String(),
				file.Hash.String(),
				ignore_reason,
			)
		}
		// update tracker for the associated commit to indicate "pending" status
		_, err = sr.TrackerCommits.Update(
			commit.Hash.String(),
			KeyCodePending,
			"",
			[]string{file.Hash.String()},
		)
		if err != nil {
			return errors.Wrapf(err, ErrMsgTrackerUpdateCommit, commit.Hash.String())
		}

		sr.logger.Debug().Msgf(
			"commit %s : scanning file %s : %s",
			commit.Hash.String(),
			file.Hash.String(),
			file.Name,
		)
		// generate and send requests for the contents of the file
		requests, r_err := ChunkFileToRequests(ChunkFileInput{
			CommitID:     commit.Hash.String(),
			File:         file,
			MaxChunkSize: sr.config.Limits.MaxRequestChunkSize,
			RepoID:       sr.ID,
		})
		if r_err != nil {
			sr.logger.Error().Err(r_err).Msgf("commit %s : failed to generate requests for file %s", commit.Hash.String(), file.Hash.String())
			sr.TrackerFiles.Update(
				file.Hash.String(),
				KeyCodeError,
				r_err.Error(),
				[]string{},
			)
			return r_err
		}
		if len(requests) == 0 {
			if file.Size > 0 {
				sr.logger.Warn().Msgf(
					"commit %s : no requests generated for file ID=%s : Name=%s : size=%d",
					commit.Hash.String(),
					file.Hash.String(),
					file.Name,
					file.Size,
				)
			}
			// no requests were generated for the file, so nothing is pending
			// and there is no need to update the tracker
			return nil
		}
		var child_keys []string
		// send each request to the channel for processing
		for _, req := range requests {
			child_keys = append(child_keys, req.ID)
			sr.channel_requests <- req
		}
		// update tracker to mark the scan of this file as "pending"
		_, err = sr.TrackerFiles.Update(
			file.Hash.String(),
			KeyCodePending,
			"",
			child_keys,
		)
		if err != nil {
			return errors.Wrapf(err, ErrMsgScanTrackerUpdateFile, file.Hash.String())
		}

		return nil
	}
}

// setRepository() method stores a pointer to the git.Repository associated
// with the ScanRepository.
func (sr *ScanRepository) setRepository(repo *git.Repository) (e error) {
	if repo == nil {
		e = ErrScanRepositorySetRepositoryNil
		return
	}

	sr.repository = repo
	return
}

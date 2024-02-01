package scanner

import (
	"fmt"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"

	nogit "github.com/has-ghas/no-phi-ai/pkg/client/no-git"
)

// ScanRepository struct embeds the ScanObject struct and adds fields
// and methods specific to scanning a git.Repository.
type ScanRepository struct {
	// embed the ScanObject struct, along with its fields and methods
	ScanObject

	commits    []*ScanCommit
	repository *git.Repository
}

// NewScanRepository() function initializes a new ScanRepository object.
func NewScanRepository(url string) (*ScanRepository, error) {
	name, err := nogit.ParseRepoNameFromURL(url)
	if err != nil {
		return nil, err
	}
	return &ScanRepository{
		ScanObject: *NewScanObject(
			"",
			name,
			ScanObjectTypeRepository,
			url,
		),
		commits:    []*ScanCommit{},
		repository: nil,
	}, nil
}

// GetCommits() method returns a slice of ScanCommit object pointers.
func (sr *ScanRepository) GetCommits() []*ScanCommit {
	return sr.commits
}

// GetRepository() method returns a pointer to the git.Repository
// associated with the ScanRepository.
func (sr *ScanRepository) GetRepository() *git.Repository {
	return sr.repository
}

// Scan() method runs the scan of the repository and keeps track of the
// progress of the scan by updating private fields of the ScanRepository.
func (sr *ScanRepository) Scan(gm *nogit.GitManager, commit_scan_func func(*object.Commit) error) (e error) {
	// ensure the ScanRepository.Status reflects that the scan has started
	sr.Status.SetStarted()

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

	// wrap the provided commit_scan_func() function with a method that tracks
	// each scanned commit by adding a ScanCommit object to the list of
	// commits in the ScanRepository, which allows for scan tracking
	commit_scan_func_wrapper := sr.scanCommitWrapperFunc(commit_scan_func)

	// TODO : modify the object.CommitIter to allow for skipping commits
	//        that have already been scanned and/or limiting the scan to
	//        only the diffs between the current and previous commits

	// iterate through the commits in the repository, processing each commit
	// with the commit_scan_func_wrapper() function, which contains the provided
	// commit_scan_func() function wrapped with the scan tracking code
	e = commit_iterator.ForEach(commit_scan_func_wrapper)
	if e != nil {
		// return any error encountered while iterating through the commits
		return
	}

	return
}

// ScanForPHI() method runs the scan of the repository for any PHI/PII.
func (sr *ScanRepository) ScanForPHI(gm *nogit.GitManager) error {
	// TODO : remove TRACE
	fmt.Println("TRACE : ScanRepository.ScanForPHI()")
	return sr.Scan(gm, sr.scanCommitForPHI)
}

// clone() method clones the repository from the ScanRepository.URL and
// sets the ScanRepository.repository field to the git.Repository object
// that references the cloned repository.
func (sr *ScanRepository) clone(gm *nogit.GitManager) (e error) {
	var repo *git.Repository
	// clone the repository from the URL
	repo, e = gm.CloneRepo(sr.URL)
	if e != nil {
		e = errors.Wrapf(e, "failed to clone repository from %s", sr.URL)
		return
	}

	// set the ScanRepository.repository field to associate the git.Repository
	sr.setRepository(repo)

	return
}

// findScanCommit() method uses the provided object.Commit to find the
// associated ScanCommit object in the ScanRepository.commits slice.
func (sr *ScanRepository) findScanCommit(commit *object.Commit) (*ScanCommit, error) {
	// TODO : remove TRACE
	fmt.Println("TRACE : ScanRepository.findScanCommit()")
	for _, sc := range sr.commits {
		if sc.GetHash().String() == commit.ID().String() {
			// return the pointer to the associated ScanCommit object if it
			// is found, along with a nil error
			return sc, nil
		}
	}
	// return a non-nil error if the ScanCommit object was not found
	return nil, ErrScanRepositoryScanCommitNotFound
}

// postScanCommit() method updates Status of the ScanCommit object stored
// in the ScanRepository.commits list to reflect the successful completion
// of the scan for that commit.
func (sr *ScanRepository) postScanCommit(scan_commit *ScanCommit) {
	// update the scan_commit.Status in order to track the successful
	// completion of the scan for that commit
	scan_commit.Status.SetCompleted()
}

// preScanCommit() method creates a new ScanCommit object from the
// input object.Commit and adds it to the list of commits tracked
// in the ScanRepository.
func (sr *ScanRepository) preScanCommit(commit *object.Commit) *ScanCommit {
	// create a new ScanCommit object from the input object.Commit
	scan_commit := NewScanCommit(commit)
	// ensure the ScanCommit.Status reflects that the scan has completed
	scan_commit.Status.SetCompleted()
	// add the new ScanCommit object to the state of the ScanRepository
	sr.commits = append(sr.commits, scan_commit)

	return scan_commit
}

func (sr *ScanRepository) scanCommitForPHI(commit *object.Commit) error {
	// get the tree of objects associated with the commit
	tree, err := commit.Tree()
	if err != nil {
		return err
	}
	// use the object.Commmit to lookup the associated ScanCommit object
	// in the ScanRepository.commits slice
	sc, sc_err := sr.findScanCommit(commit)
	if sc_err != nil {
		return sc_err
	}

	// iterate through the files in the commit tree
	err = tree.Files().ForEach(sc.ScanFile)
	if err != nil {
		return err
	}

	return err
}

func (sr *ScanRepository) scanCommitWrapperFunc(scan_func func(*object.Commit) error) func(*object.Commit) error {
	return func(commit *object.Commit) error {
		// perform pre-scan processing of the object.Commit in order to track
		// a new ScanCommit object and associate it with this ScanRepository
		scan_commit := sr.preScanCommit(commit)

		// run the provided scan_func function and process any error
		if scan_func_err := scan_func(commit); scan_func_err != nil {
			// update the scan_commit.Status in order to track the error
			scan_commit.Status.SetErrored(scan_func_err.Error())
			// return the error
			return scan_func_err
		}

		// perform post-scan processing in order to update the status of the
		// ScanCommit within this ScanRepository
		sr.postScanCommit(scan_commit)

		return nil
	}
}

// setRepository() method stores a pointer to the git.Repository
// associated with the ScanRepository.
func (sr *ScanRepository) setRepository(repo *git.Repository) {
	sr.repository = repo
}
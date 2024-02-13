package scanner

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/has-ghas/no-phi-ai/pkg/client/az"
)

// ScanCommit struct embeds the ScanObject struct and adds fields
// and methods specific to scanning a git commit.
type ScanCommit struct {
	// embed the ScanObject struct, along with its fields and methods
	ScanObjectHashed

	commit *object.Commit
	files  []*ScanFile
}

// NewScanCommit() function initializes a new ScanCommit object using
// the provided object.Commit.
func NewScanCommit(
	commit *object.Commit,
	channel_documents chan<- az.AsyncDocumentWrapper,
	channel_quit <-chan error,
) (*ScanCommit, error) {
	if scannerContext == nil {
		return nil, ErrScanCommitContextNil
	}
	if commit == nil {
		return nil, ErrScanCommitInputCommitNil
	}
	if channel_documents == nil {
		return nil, ErrScanCommitChannelDocumentsNil
	}

	return &ScanCommit{
		ScanObjectHashed: *NewScanObjectHashed(commit.ID(), &ScanObjectInput{
			ChannelDocuments: channel_documents,
			ChannelQuit:      channel_quit,
			ID:               commit.ID().String(),
			Name:             commit.ID().String(),
			ObjectType:       ScanObjectTypeCommit,
			URL:              "", // TODO : fix empty URL for ScanObject
		}),
		commit: commit,
		files:  []*ScanFile{},
	}, nil
}

// GetHash() method returns the plumbing.Hash of this ScanCommit.
// Overrides the GetHash() method of the embedded ScanObjectHashed struct.
func (sc *ScanCommit) GetHash() plumbing.Hash {
	return sc.commit.Hash
}

// GetCommit() method returns a pointer to the object.Commit associated
// with the ScanCommit.
func (sc *ScanCommit) GetCommit() *object.Commit {
	return sc.commit
}

// GetFiles() method returns the slice of ScanFile object pointers currently
// associated with the ScanCommit.
func (sc *ScanCommit) GetFiles() []*ScanFile {
	return sc.files
}

// findScanFile() method uses the provided object.File to find the associated
// ScanFile object in the ScanCommit.files slice.
func (sc *ScanCommit) findScanFile(file *object.File) (*ScanFile, error) {
	if len(sc.files) == 0 || sc.files == nil {
		return nil, ErrScanCommitFilesNotSet
	}
	for _, scan_file := range sc.files {
		if scan_file.GetHash().String() == file.Hash.String() {
			return scan_file, nil
		}
	}

	return nil, ErrScanCommitScanFileNotFound
}

// ignoreScanFile() method returns boolean true if the object.File should
// not be added to the list of files to be scanned, and boolean false if the
// object.File should be added to (i.e. should not be ignored from) the list
// of files to be scanned.
func (sc *ScanCommit) ignoreScanFile(file *object.File) (ignore bool, reason string) {
	// explicitly set defaults for return values
	ignore = false
	reason = ""

	// ignore binary files as we are just scanning text for PHI/PII data
	if file_is_binary, _ := file.IsBinary(); file_is_binary {
		ignore = true
		reason = IgnoreReasonFileIsBinary
		return
	}

	// ignore files with zero size
	if file.Size == 0 {
		ignore = true
		reason = IgnoreReasonFileIsEmpty
		return
	}

	// ignore files with names that match an entry in the ignore map
	return IgnoreFilePath(file.Name)
}

// postScanFile() method updates the Status of the ScanFile object stored
// in the ScanCommit.files slice to reflect the results of the scan.
func (sc *ScanCommit) postScanFile(file *object.File) error {
	var e error
	var scan_file *ScanFile

	// use the object.File to lookup the associated ScanFile object in
	// the ScanCommit.files slice
	scan_file, e = sc.findScanFile(file)
	if e != nil {
		e = errors.Wrap(e, "scan commit failed to find scan file for post-processing")
		// return the error
		return e
	}

	// update the scan_file.Status in order to track the successful
	// completion of the scan for that file
	scan_file.Status.SetCompleted(ResultCleanCode, "")

	return e
}

// preScanFile() method creates a new ScanFile object from the input
// object.File and adds it to the list of files tracked in the ScanCommit.
func (sc *ScanCommit) preScanFile(file *object.File) (*ScanFile, error) {
	if file == nil {
		return nil, ErrScanCommitFileNil
	}

	// create a new ScanFile object from the input object.File
	scan_file, err := NewScanFile(file, sc.channelDocuments, sc.channelQuit)
	if err != nil {
		return nil, err
	}

	// check if the file should be marked as ignored by the scan
	if ignore, reason := sc.ignoreScanFile(file); ignore {
		// update the Status of the ScanFile to reflect that the file
		// has been ignored by the scan, which helps in tracking the
		// progress of the scan
		scan_file.Status.SetIgnored(reason)
	} else {
		// ensure the ScanFile.Status reflects that the scan has started
		scan_file.Status.SetStarted("")
	}

	// add the new ScanFile object to the state of the ScanCommit
	sc.files = append(sc.files, scan_file)

	return scan_file, nil
}

// scanFile() method implements the logic for scanning a single object.File
// by first converting the object.File to a ScanFile, then generating the
// (PHI/PII) entity detection documents from the file, and updating the
// status of the ScanFile object to reflect the results of the scan.
func (sc *ScanCommit) scanFile(file *object.File) error {
	// perform pre-scan processing of the object.File in order to track
	// a new ScanFile object and associate it with this ScanCommit
	scan_file, pre_scan_err := sc.preScanFile(file)
	if pre_scan_err != nil {
		// return the error
		return pre_scan_err
	}

	if scan_file.Status.IsIgnored() {
		log.Ctx(scannerContext).Trace().Msgf(
			"scan file ignored : path = %s : reason = %s",
			scan_file.Name,
			scan_file.Status.StateMessage,
		)
		// return early if the file is ignored
		return nil
	}

	// run the scan_file.scan() method and process any error
	if gen_docs_err := scan_file.Scan(file); gen_docs_err != nil {
		// update the scan_file.Status in order to track the error
		scan_file.Status.SetErrored(gen_docs_err.Error())
		// return the error
		return gen_docs_err
	}

	// perform post-scan processing in order to update the status of the
	// ScanFile within this ScanCommit
	if post_scan_err := sc.postScanFile(file); post_scan_err != nil {
		// update the scan_file.Status in order to track the error
		scan_file.Status.SetErrored(post_scan_err.Error())
		// return the error
		return post_scan_err
	}

	return nil
}

package scanner

import (
	"fmt" // TODO : remove

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
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
func NewScanCommit(commit *object.Commit) *ScanCommit {
	return &ScanCommit{
		ScanObjectHashed: *NewScanObjectHashed(
			commit.ID(),
			commit.ID().String(),
			ScanObjectTypeCommit,
			"",
		),
		commit: commit,
		files:  []*ScanFile{},
	}
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

// ScanFile() method scans the tree of objects (e.g. files) associated
// with the commit, keeping track of the progress of the scan by updating
// private fields of the ScanCommit.
func (sc *ScanCommit) ScanFile(file *object.File) error {
	var e error

	// TODO : remove TRACE
	fmt.Println("TRACE : ScanCommit.scanFileForPHI() start")

	// get an iterator for the files in the commit
	var files_iterator *object.FileIter
	files_iterator, e = sc.commit.Files()
	if e != nil {
		if files_iterator != nil {
			files_iterator.Close()
		}
		return e
	}
	defer files_iterator.Close()

	// iterate over the files in the commit, processing each file with the
	// file_scan_func_wrapper() function, which contains the scan_file.Scan()
	// function wrapped with code to track the progress of the file scan
	e = files_iterator.ForEach(sc.scanFile)
	if e != nil {
		// return any error encountered while iterating over the files
		return e
	}

	return e
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

// postScanFile() method updates the Status of the ScanFile object stored
// in the ScanCommit.files slice to reflect the results of the scan.
func (sc *ScanCommit) postScanFile(file *object.File) error {
	var e error
	var scan_file *ScanFile

	// use the object.File to lookup the associated ScanFile object in
	// the ScanCommit.files slice
	scan_file, e = sc.findScanFile(file)
	if e != nil {
		// return the error
		return e
	}

	// update the scan_file.Status in order to track the successful
	// completion of the scan for that file
	scan_file.Status.SetCompleted()

	return e
}

// shouldScanFile() method returns boolean false if the object.File should
// not be added to the list of files to be scanned, and boolean true if
// the object.File should be added to the list of files to be scanned.
func (sc *ScanCommit) shouldScanFile(file *object.File) (should_scan bool) {
	// assume that we probably should scan the file
	should_scan = true

	if file == nil {
		should_scan = false
		return
	}

	// ignore binary files as we are just scanning text for PHI/PII data
	if file_is_binary, _ := file.IsBinary(); file_is_binary {
		should_scan = false
	}

	return
}

// preScanFile() method creates a new ScanFile object from the input
// object.File and adds it to the list of files tracked in the ScanCommit.
func (sc *ScanCommit) preScanFile(file *object.File) (*ScanFile, error) {
	// create a new ScanFile object from the input object.File
	scan_file, err := NewScanFile(file)
	if err != nil {
		return nil, err
	}
	// ensure the ScanFile.Status reflects that the scan has started
	scan_file.Status.SetStarted()
	// add the new ScanFile object to the state of the ScanCommit
	sc.files = append(sc.files, scan_file)

	return scan_file, nil
}

// scanFile() method implements the logic for scanning a single object.File
// by first converting the object.File to a ScanFile, then generating the
// (PHI/PII) entity detection documents from the file, and updating the
// status of the ScanFile object to reflect the results of the scan.
func (sc *ScanCommit) scanFile(file *object.File) error {
	// TODO : remove TRACE
	fmt.Println("TRACE : ScanCommit.scanFile() : start")

	// check if the file should be scanned
	if !sc.shouldScanFile(file) {
		fmt.Printf("TRACE : skipping scan of file %s\n", file.Name)
		return nil
	}

	// perform pre-scan processing of the object.File in order to track
	// a new ScanFile object and associate it with this ScanCommit
	scan_file, pre_scan_err := sc.preScanFile(file)
	if pre_scan_err != nil {
		// return the error
		return pre_scan_err
	}

	// run the scan_file.generateDocuments() method and process any error
	if gen_docs_err := scan_file.generateDocuments(file); gen_docs_err != nil {
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
	// TODO : remove TRACE
	fmt.Println("TRACE : ScanCommit.scanFile() : finish")

	return nil
}

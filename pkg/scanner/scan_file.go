package scanner

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/object"
)

// ScanFile struct embeds the ScanObject struct and adds fields
// and methods specific to scanning a GitHub organization.
type ScanFile struct {
	// embed the ScanObject struct, along with its fields and methods
	ScanObjectHashed
}

// NewScanFile() function initializes a new ScanFile object using
// the provided URL for the GitHub organization.
func NewScanFile(file *object.File) (*ScanFile, error) {
	if file == nil {
		return nil, ErrScanFileInputNil
	}
	// initialize and return a new ScanFile object
	return &ScanFile{
		ScanObjectHashed: *NewScanObjectHashed(
			file.ID(),
			file.Name,
			ScanObjectTypeFile,
			"", // TODO
		),
	}, nil
}

// generateDocuments() function generates PHI/PII entity detection
// requests from the provided object.File, which is a file or blob in a
// git repository. Requests are limited to a maximum of 5 "documents",
// with a limit of 5,000 characters per "document".
func (sf *ScanFile) generateDocuments(file *object.File) (e error) {
	documents := NewDocumentTrackerMap()

	// TODO : remove TRACE
	fmt.Println("TRACE : ScanFile.generateDocuments() : start")

	// split the file / blob into chunks and prepare to scan each chunk
	// for PHI by creating a new DocumentTracker object for each chunk of
	// the file
	// TODO

	// set the documents field of the ScanFile object to the generated
	// map of DocumentTracker objects
	if documents.CountTotal() > 0 {
		if e = sf.SetDocuments(documents); e != nil {
			return
		}
	}

	// TODO : remove TRACE
	fmt.Println("TRACE : ScanFile.generateDocuments() : finish")

	return
}

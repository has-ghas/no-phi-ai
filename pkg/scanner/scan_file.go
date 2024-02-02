package scanner

import (
	"bufio"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"

	"github.com/has-ghas/no-phi-ai/pkg/client/az"
)

// ScanFile struct embeds the ScanObject struct and adds fields
// and methods specific to scanning a GitHub organization.
type ScanFile struct {
	// embed the ScanObject struct, along with its fields and methods
	ScanObjectHashed
}

// NewScanFile() function initializes a new ScanFile object using
// the provided URL for the GitHub organization.
func NewScanFile(
	file *object.File,
	channel_documents chan<- az.AsyncDocumentWrapper,
	channel_quit <-chan error,
) (*ScanFile, error) {
	if file == nil {
		return nil, ErrScanFileInputNil
	}
	// initialize and return a new ScanFile object
	return &ScanFile{
		ScanObjectHashed: *NewScanObjectHashed(file.ID(), &ScanObjectInput{
			ChannelDocuments: channel_documents,
			ChannelQuit:      channel_quit,
			ID:               file.ID().String(),
			Name:             file.Name,
			ObjectType:       ScanObjectTypeFile,
			URL:              "", // TODO
		}),
	}, nil
}

// Scan() method wraps the private methods that implement the scanning
// the text of the file (blob) for PHI/PII entities. The method is
// expected to return an error if the scan fails to run for that file.
func (sf *ScanFile) Scan(file *object.File) (e error) {

	// generate a new DocumentTracker for each chunk of the file, where
	// the chunk size is determined by the az.DocumentCharacterLimit
	if e = sf.generateDocuments(file); e != nil {
		return
	}

	// batch the documents into groups of az.RequestDocumentLimit
	// and create a new az.PiiEntityRecognitionRequest for each batch
	// TODO

	// submit the requests to the Azure AI Language service API
	// TODO

	// process the responses from the Azure AI Language service API
	// TODO

	return
}

// generateDocuments() function generates PHI/PII entity detection
// requests from the provided object.File, which is a file or blob in a
// git repository. Requests are limited to a maximum of 5 "documents",
// obeying the az.DocumentCharacterLimit per "document".
func (sf *ScanFile) generateDocuments(file *object.File) (e error) {

	file_reader, err := file.Reader()
	if err != nil {
		e = errors.Wrap(err, "failed to get file reader for generating documents")
		return
	}
	file_scanner := bufio.NewScanner(file_reader)
	file_scanner.Split(bufio.ScanRunes)
	var current_offset int
	var current_text string
	// scan the file / blob and split into chunks and prepare to scan each
	// chunk for PHI by creating a new DocumentTracker object that allows
	// for mapping any documents to their respective offsets in the file
	for file_scanner.Scan() {
		char := file_scanner.Text()
		current_text += char

		if len(current_text) == az.DocumentCharacterLimit {
			if e = sf.makeDocumentTracker(current_offset, current_text); e != nil {
				return
			}
			// reset the current_text variable to prepare for the next
			// chunk of the file
			current_text = ""
			// increment the current_offset variable to the next character
			current_offset += len(char)
		}
	}

	// handle the last chunk of characters from the file
	if len(current_text) > 0 {
		e = sf.makeDocumentTracker(current_offset, current_text)
		if e != nil {
			return
		}
	}

	// check the file_scanner for any error
	if err := file_scanner.Err(); err != nil {
		e = errors.Wrap(err, "file scanner returned error when generating documents")
		return
	}

	return
}

// makeDocumentTracker() method uses the context of the ScanFile object
// along with the provided offset and text to create a new DocumentTracker
// for the text.
func (sf *ScanFile) makeDocumentTracker(offset int, text string) error {
	// create a unique identifier for the document
	id := MakeDocumentID(sf.GetHash(), offset, []byte(text))

	// TODO : remove TRACE
	fmt.Printf(
		"TRACE : ScanFile.makeDocumentTracker() : hash=%s : path=%s : id=%s : offset=%d : text_length=%d\n",
		sf.hash.String(),
		sf.Name,
		id,
		offset,
		len(text),
	)

	// create the az.Document object which is a required input for a new DocumentTracker
	az_doc := az.NewDocument(id, text, "")

	document_tracker, err := NewDocumentTracker(&DocumentTrackerInput{
		ChannelDocuments: sf.channelDocuments,
		ChannelQuit:      sf.channelQuit,
		Document:         &az_doc,
		ID:               az_doc.ID,
		Offset:           offset,
		Path:             sf.Name,
	})
	if err != nil {
		return errors.Wrap(err, "failed to setup DocumentTracker for scan file")
	}

	// add the document_tracker object to the map of documents
	// stored in the ScanFile object
	if err := sf.documents.Set(document_tracker); err != nil {
		return err
	}

	// use the document tracker to wrap and send the document to be scanned
	go document_tracker.Scan()

	return nil
}

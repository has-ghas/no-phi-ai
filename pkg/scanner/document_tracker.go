package scanner

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/has-ghas/no-phi-ai/pkg/client/az"
)

// DocumentTrackerInput struct is used to provide the input parameters for
// creating a new DocumentTracker object. Wraps the az.Document object and
// adds fields for tracking the source of the document text back to the
// original source path (e.g. file, comment, issue body, etc.) and the
// starting offset of the source text within the document.
type DocumentTrackerInput struct {
	ChannelDocuments chan<- az.AsyncDocumentWrapper
	ChannelQuit      <-chan error
	Document         *az.Document
	ID               string
	Offset           int
	Path             string
}

// DocumentTracker struct provides a management wrapper for tracking the status
// of PHI/PII entity detection for a given source of text -- such as a file, a
// comment on a GitHub issue, the body of a pull request, etc.
type DocumentTracker struct {
	// ID is the unique identifier for the document being tracked, where the
	// input ID is the ID of the Document being scanned and must match the ID
	// of the DocumentResponse object for a successful mapping of the response.
	ID string `json:"id"`
	// Offset is the starting position of the source Text within the document,
	// set to the index of the first character (from the source Text) in relation
	// to the start of the full text source.
	Offset int `json:"offset"`
	// Path is the location of the document being tracked, such as a file path
	// or the URL of a comment on a GitHub issue.
	Path string `json:"path"`

	channelDocuments chan<- az.AsyncDocumentWrapper
	channelQuit      <-chan error
	channelResponse  chan az.DocumentResponse
	document         *az.Document
	response         *az.DocumentResponse
}

// NewDocumentTracker() function initializes a new DocumentTracker object
// for use in tracking the status of a document scan and mapping the
// response back to the source Path (e.g. of the file) and Offset
// (e.g. starting character of the target text within the source file.
func NewDocumentTracker(in *DocumentTrackerInput) (*DocumentTracker, error) {
	if scannerContext == nil {
		return nil, ErrDocumentTrackerContextNil
	}
	return &DocumentTracker{
		ID:               in.ID,
		Offset:           in.Offset,
		Path:             in.Path,
		channelDocuments: in.ChannelDocuments,
		channelResponse:  make(chan az.DocumentResponse),
		channelQuit:      in.ChannelQuit,
		document:         in.Document,
		response:         nil,
	}, nil
}

// GetResponse() method returns the response for the tracked document if the
// response has been received and set, otherwise returns nil.
func (dt *DocumentTracker) GetResponse() *az.DocumentResponse {
	return dt.response
}

// Scan() method creates a new az.AsyncDocumentWrapper object and sends it to
// the channelDocuments channel for processing and waits for a response to be
// received on the channelResponse channel created for -- and embedded in --
// the az.AsyncDocumentWrapper object.
func (dt *DocumentTracker) Scan() (e error) {
	// create a new az.AsyncDocumentWrapper object
	wrapper := az.NewAsyncDocumentWrapper(
		dt.document.ID,
		dt.document.Text,
		"", // use default value for document language
		dt.channelResponse,
	)
	// send the wrapper to the channelDocuments channel
	dt.channelDocuments <- *wrapper

	// wait for a response to be received on the dt.channelResponse channel,
	// or for the dt.channelQuit channel to be closed
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case response := <-dt.channelResponse:
			log.Ctx(scannerContext).Trace().Msgf(
				"document tracker channel received response ID=%s",
				response.ID,
			)
			dt.SetResponse(&response)
		case err := <-dt.channelQuit:
			if err != nil {
				e = errors.Wrap(e, "document tracker received error from quit channel")
				return
			}
			log.Ctx(scannerContext).Trace().Msg("document tracker received interrup signal")
			// return nil error if the quit channel is closed
			return
		}
	}()
	wg.Wait()
	if e != nil {
		e = errors.Wrap(e, "document tracker channel response error")
		return
	}
	log.Ctx(scannerContext).Trace().Msgf(
		"document tracker goroutine done : ID=%s : Offset=%d",
		wrapper.Document.ID,
		dt.Offset,
	)
	return
}

// SetResponse() method sets the response for the document being tracked.
func (dt *DocumentTracker) SetResponse(response *az.DocumentResponse) (e error) {
	if response == nil {
		return ErrDocumentResponseIsNil
	}
	if response.ID != dt.ID {
		return ErrDocumentResponseMismatch
	}
	dt.response = response
	return
}

// DocumentTrackerMap type is a map of DocumentTracker objects, where the key
// is the ID of the Document being tracked and the value is a pointer to the
// DocumentTracker object that wraps the Document and its response.
type DocumentTrackerMap map[string]*DocumentTracker

// NewDocumentTrackerMap() function initializes a new DocumentTrackerMap object.
func NewDocumentTrackerMap() DocumentTrackerMap {
	return make(DocumentTrackerMap)
}

// CountTotal() method returns the total number of DocumentTracker objects in
// the DocumentTrackerMap.
func (m DocumentTrackerMap) CountTotal() int {
	return len(m)
}

// Get() method returns the DocumentTracker object for the provided ID, or nil
// if no such object exists in the map.
func (m DocumentTrackerMap) Get(id string) *DocumentTracker {
	document_tracker, exists := m[id]
	if !exists {
		return nil
	}
	return document_tracker
}

// Set() method sets the DocumentTracker object for the provided ID in the map.
// If the ID already exists in the map, the existing DocumentTracker object is
// replaced with the provided object. Returns a non-nil error if the provided
// DocumentTracker object is nil.
func (m DocumentTrackerMap) Set(document_tracker *DocumentTracker) (e error) {
	if document_tracker == nil {
		return ErrDocumentTrackerMapInputIsNil
	}
	m[document_tracker.ID] = document_tracker
	return
}

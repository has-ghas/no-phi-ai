package az

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// AsyncDocumentResponseWrapper struct provides a wrapper for a single
// DocumentResponse object, where the RequestedAt and RespondedAt
// timestamps are used to track the time between the request and
// the response for the document.
type AsyncDocumentResponseWrapper struct {
	DocumentResponse *DocumentResponse
	RequestedAt      int64
	RespondedAt      int64
}

// NewAsyncDocumentResponseWrapper() function initializes a new
// AsyncDocumentResponseWrapper object with the provided DocumentResponse.
func NewAsyncDocumentResponseWrapper(document_response *DocumentResponse) *AsyncDocumentResponseWrapper {
	return &AsyncDocumentResponseWrapper{
		DocumentResponse: document_response,
		RequestedAt:      0,
		RespondedAt:      0,
	}
}

// SetRequested() method sets the RequestedAt timestamp of the document
// wrapper to the current time.
func (wrapper *AsyncDocumentResponseWrapper) SetRequested(timestamp int64) {
	if timestamp <= 0 {
		timestamp = time.Now().Unix()
	}
	wrapper.RequestedAt = timestamp
}

// SetResponded() method sets the RespondedAt timestamp of the document
// wrapper to the current time.
func (wrapper *AsyncDocumentResponseWrapper) SetResponded(timestamp int64) {
	if timestamp <= 0 {
		timestamp = time.Now().Unix()
	}
	wrapper.RespondedAt = timestamp
}

// AsyncDocumentWrapper struct provides a async (channel-based) wrapper for
// a single Document object, where the DocumentResponse is sent back to the
// provided channel once the response has been received and validated.
type AsyncDocumentWrapper struct {
	ChanQuit     chan error
	ChanResponse chan<- AsyncDocumentResponseWrapper
	Document     *Document
}

// sendRequest() method sends the Document object to the provided response
// channel if the DocumentResponse is valid in the context of the original
// (wrapped) Document object.
func (wrapper *AsyncDocumentWrapper) sendResponse(ctx context.Context, response_wrapper *AsyncDocumentResponseWrapper) (e error) {
	// validate the response channel
	if wrapper.ChanResponse == nil {
		e = errors.New("invalid response channel for AsyncDocumentWrapper")
		wrapper.ChanQuit <- e
		return
	}
	// validate the DocumentResponse
	if wrapper.Document.ID != response_wrapper.DocumentResponse.ID {
		e = errors.New("invalid DocumentResponse for AsyncDocumentWrapper : ID mismatch")
		wrapper.ChanQuit <- e
		return
	}

	// send the response to the response channel
	wrapper.ChanResponse <- *response_wrapper
	// close the response channel
	close(wrapper.ChanResponse)

	log.Ctx(ctx).Trace().Msgf("sent document response for ID = %s", wrapper.Document.ID)
	return
}

// NewAsyncDocumentWrapper() function initializes a new AsyncDocumentWrapper
// object for use in asynchronous processing of a Document object and returns
// a pointer to the wrapper.
func NewAsyncDocumentWrapper(id, text, language string, response chan<- AsyncDocumentResponseWrapper) *AsyncDocumentWrapper {
	document := NewDocument(id, text, language)
	return &AsyncDocumentWrapper{
		ChanQuit:     make(chan error),
		ChanResponse: response,
		Document:     &document,
	}
}

// isValid() method returns boolean true if the document wrapper is valid,
// otherwise returns boolean false.
func (wrapper *AsyncDocumentWrapper) isValid() bool {
	return wrapper.Document != nil && wrapper.Document.ID != "" && wrapper.Document.Text != ""
}

// AsyncDocumentWrapperMap provides convenience methods for managing a map of
// AsyncDocumentWrapper objects, where the key is the ID of the document
// wrapper and the value is a pointer to the wrapper.
type AsyncDocumentWrapperMap map[string]*AsyncDocumentWrapper

// NewAsyncDocumentWrapperMap() function makes a new, empty map of
// AsyncDocumentWrapper objects and returns the map.
func NewAsyncDocumentWrapperMap() AsyncDocumentWrapperMap {
	return make(AsyncDocumentWrapperMap)
}

// add() method adds the provided wrapper to the map if the wrapper is valid
// and if the map is not full. The method returns boolean values indicating
// whether the wrapper is valid and whether the wrapper was added to the map.
func (m AsyncDocumentWrapperMap) add(wrapper *AsyncDocumentWrapper) (valid, added bool) {
	// assume the wrapper is NOT valid until proven otherwise
	valid = false
	// assume the wrapper is NOT added to the map until proven otherwise
	added = false
	if wrapper == nil {
		return
	}
	if wrapper.isValid() {
		valid = true
	}
	if m.isFull() {
		// wrapper is valid, but the map is full
		return
	}
	// add the wrapper to the map
	m[wrapper.Document.ID] = wrapper
	// mark the wrapper as added to the map
	added = true
	return
}

// get() method returns the wrapper for the provided ID if it exists in the
// map and boolean true, otherwise it returns nil and boolean false.
func (m AsyncDocumentWrapperMap) get(id string) (wrapper *AsyncDocumentWrapper, exists bool) {
	wrapper, exists = m[id]
	return
}

// isEmpty() method returns true if the map is empty, otherwise false.
func (m AsyncDocumentWrapperMap) isEmpty() bool {
	return m.length() == 0
}

// isFull() method returns true if the map is full, otherwise false. The map
// is considered full if the number of documents in the map is equal or
// greater than the RequestDocumentLimit.
func (m AsyncDocumentWrapperMap) isFull() bool {
	return m.length() >= RequestDocumentLimit
}

// length() method returns the number of documents in the map.
func (m AsyncDocumentWrapperMap) length() int {
	return len(m)
}

// toPiiEntityRecognitionRequest() method converts the AsyncDocumentWrapperMap
// of documents to a PiiEntityRecognitionRequest object, which can be used in
// HTTP requests sent to the Azure AI Language service API. Since a request is
// being generated for the documents, this method also sets the RequestedAt
// timestamp for each document in the map. If the map contains more documents
// documents than the RequestDocumentLimit, or no documents at all, then a
// non-nil error is returned.
func (m AsyncDocumentWrapperMap) toPiiEntityRecognitionRequest() (*PiiEntityRecognitionRequest, error) {
	if m.length() < 1 {
		return nil, ErrTooFewDocumentsForEntityRequest
	}
	if m.length() > RequestDocumentLimit {
		return nil, ErrTooManyDocumentsForEntityRequest
	}

	// create an empty slice of Document objects
	documents := make([]Document, 0)
	for _, wrapper := range m {
		// append the document to the documents slice
		documents = append(documents, *wrapper.Document)
	}

	// create and return a new PiiEntityRecognitionRequest object
	// from the documents slice, along with a nil error
	return NewPiiEntityRecognitionRequest(documents), nil
}

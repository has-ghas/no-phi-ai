package scanner

import "github.com/has-ghas/no-phi-ai/pkg/client/az"

// DocumentTrackerInput struct is used to provide the input parameters for
// creating a new DocumentTracker object.
type DocumentTrackerInput struct {
	Document     *az.Document
	Path         string
	OffsetLine   int
	OffsetColumn int
}

// DocumentTracker struct provides a management wrapper for tracking the status
// of PHI/PII entity detection for a given source of text -- such as a file, a
// comment on a GitHub issue, the body of a pull request, etc.
type DocumentTracker struct {
	// ID is the unique identifier for the document being tracked, where the
	// input ID is the ID of the Document being scanned and must match the ID
	// of the DocumentResponse object for a successful mapping of the response.
	ID string `json:"id"`
	// Path is the location of the document being tracked, such as a file path
	// or the URL of a comment on a GitHub issue.
	Path string `json:"location"`
	// OffsetColumn is the starting character number within the OffsetLine
	OffsetColumn int `json:"offset_column"`
	// OffsetLine is the starting line number within the document, such as the
	// line number within a file or comment body.
	OffsetLine int `json:"offset_line"`

	document *az.Document
	response *az.DocumentResponse
}

// NewDocumentTracker() function initializes a new DocumentTracker object for
// use in tracking the status of a document scan and mapping the response back
// to the source Path (e.g. of the file), OffsetLine (e.g. line number within
// the file), and OffsetColumn (e.g. character number within the line).
func NewDocumentTracker(in DocumentTrackerInput) *DocumentTracker {
	return &DocumentTracker{
		ID:           in.Document.ID,
		OffsetColumn: in.OffsetColumn,
		OffsetLine:   in.OffsetLine,
		Path:         in.Path,
		document:     in.Document,
		response:     nil,
	}
}

func (dt *DocumentTracker) GetDocument() *az.Document {
	return dt.document
}

func (dt *DocumentTracker) GetResponse() *az.DocumentResponse {
	return dt.response
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

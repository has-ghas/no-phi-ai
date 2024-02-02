package scanner

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/google/uuid"
	"github.com/has-ghas/no-phi-ai/pkg/client/az"
)

// ScanObjectInput struct provides the input parameters for creating a new
// ScanObject instance. The struct is used to pass the input parameters to
// the NewScanObject() function.
type ScanObjectInput struct {
	ChannelDocuments chan<- az.AsyncDocumentWrapper
	ChannelQuit      <-chan error
	ID               string
	Name             string
	ObjectType       string
	URL              string
}

// ScanObject struct is the base struct for all types of scanned objects
// and provides a common set of fields and methods for tracking the status
// of a scan for some uniquely identifiable object.
type ScanObject struct {
	// ID is the unique identifier for the object.
	ID string `json:"id"`
	// Name is the friendly name of the object, such as the name
	// of the file, organization, repository, etc.
	Name string `json:"name"`
	// Status tracks the different statuses of the object scan.
	Status *Status `json:"status"`
	// Type can be one of:
	//   - "commit"
	//   = "issue_comment"
	//   - "issue"
	//   - "file"
	//   = "organization"
	//   = "pull_request_comment"
	//   - "pull_request"
	//   - "repository"
	Type string `json:"type"`
	// The unique URL associated with the object.
	URL string `json:"url"`

	documents        DocumentTrackerMap // TODO : remove this field and any associated methods
	channelDocuments chan<- az.AsyncDocumentWrapper
	channelQuit      <-chan error
}

// NewScanObject() function initializes a new ScanObject struct with
// the provided name, type, and URL. Sets default values for the
// Status of the ScanObject.
func NewScanObject(in *ScanObjectInput) *ScanObject {
	// create a random, unique ID if one is not provided
	if in.ID == "" {
		in.ID = uuid.New().String()
	}
	return &ScanObject{
		ID:               in.ID,
		Name:             in.Name,
		Status:           NewStatus(),
		Type:             in.ObjectType,
		URL:              in.URL,
		channelDocuments: in.ChannelDocuments,
		channelQuit:      in.ChannelQuit,
		documents:        NewDocumentTrackerMap(),
	}
}

// GetID() method returns the ID of the ScanObject.
func (so *ScanObject) GetID() string {
	return so.ID
}

// GetName() method returns the Name of the ScanObject.
func (so *ScanObject) GetName() string {
	return so.Name
}

// GetType() method returns the Type of the ScanObject.
func (so *ScanObject) GetType() string {
	return so.Type
}

// GetURL() method returns the URL of the ScanObject.
func (so *ScanObject) GetURL() string {
	return so.URL
}

// GetDocuments() method returns the DocumentTrackerMap of objects
// created for -- and associated with -- the ScanObject. Returns a non-nil
// error if the request has not been created/set for the ScanObject.
func (so *ScanObject) GetDocuments() (documents DocumentTrackerMap, e error) {
	if len(so.documents) == 0 {
		e = ErrScanObjectDocumentsNotSet
		return
	}
	documents = so.documents
	return
}

// SetDocuments() method sets the map of documents tracked for the ScanObject.
func (so *ScanObject) SetDocuments(documents DocumentTrackerMap) (e error) {
	if len(documents) == 0 || documents == nil {
		return ErrScanObjectDocumentsNotValid
	}
	so.documents = documents
	return
}

// ScanObjectHashed struct is a ScanObject that has been extended to include
// a private field for storing the plumbing.Hash of the object and a public
// method for getting that hash. Useful for scan objects that can always be
// associated with some plumbing.Hash, such as a commit or a file.
type ScanObjectHashed struct {
	ScanObject

	hash plumbing.Hash
}

// NewScanObjectHashed() function initializes a new ScanObjectHashed struct
// with the provided plumbing.Hash, name, type, and URL.
func NewScanObjectHashed(hash plumbing.Hash, in *ScanObjectInput) *ScanObjectHashed {
	// override the object ID with the string representation of the hash
	// to ensure consistency between the hash and the object ID
	in.ID = hash.String()
	return &ScanObjectHashed{
		ScanObject: *NewScanObject(in),
		hash:       hash,
	}
}

// GetHash() method returns the plumbing.Hash of the ScanObjectHashed object.
func (so *ScanObjectHashed) GetHash() plumbing.Hash {
	return so.hash
}

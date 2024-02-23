package scannerv2

import (
	"github.com/google/uuid"
)

// MetadataRequestResponset struct contains metadata specific to a request and
// its associated response, where this metadata is typically copied from the
// request to the response.
type MetadataRequestResponse struct {
	// ID is the unique identifier of the request.
	ID string `json:"id"`
	// Commit struct contains information about the associated commit.
	Commit MetadataRequestResponseCommit `json:"commit"`
	// Object struct contains information about the associated object (e.g. file).
	Object MetadataRequestResponseObject `json:"object"`
	// Repository struct contains information about the associated repository.
	Repository MetadataRequestResponseRepository `json:"repository"`
	// Time struct contains timestamps set during request processing.
	Time MetadataRequestResponseTime `json:"time"`
}

type MetadataRequestResponseCommit struct {
	ID string `json:"id"`
}

type MetadataRequestResponseObject struct {
	// ID is the string version of the file's SHA1 hash, which is unique
	// to the file's content and context (e.g. repository, commit, etc.)
	ID string `json:"id"`
	// Length is the number of characters in the source text.
	Length int `json:"length"`
	// Offset is the starting character position of the source text within its
	// original context (e.g. offset fromn start of file).
	Offset int `json:"offset"`
}

type MetadataRequestResponseRepository struct {
	// ID is the unique identifier created by this app for the purpose of
	// tracking the repository.
	ID string `json:"id"`
	// URL is the unique URL used to interact with the repository.
	URL string `json:"url"`
}

type MetadataRequestResponseTime struct {
	// Start is the time the request was received.
	Start int64 `json:"start"`
	// Stop is the time the request was completed.
	Stop int64 `json:"stop"`
}

// Request struct contains all the information needed to process a request to
// detect PHI/PII data in some source (e.g. file) object and to identify the
// offending data within the source.
type Request struct {
	// embed the MetadataRequestResponse struct
	MetadataRequestResponse

	// Text is the source text to be scanned for PHI/PII data and is only
	// included in the Request (not the Response) object in order to limit the
	// size of the response and the exposure of the source text.
	Text string `json:"text"`
}

// NewRequest() function initializes a new Request object.
func NewRequest(repo_id, commit_id, object_id string) (*Request, error) {
	if repo_id == "" {
		return nil, ErrNewRequestEmptyRepositoryID
	}
	if object_id == "" {
		return nil, ErrNewRequestEmptyObjectID
	}
	if commit_id == "" {
		return nil, ErrNewRequestEmptyCommitID
	}

	return &Request{
		MetadataRequestResponse: MetadataRequestResponse{
			ID: uuid.NewString(),
			Commit: MetadataRequestResponseCommit{
				ID: commit_id,
			},
			Object: MetadataRequestResponseObject{
				ID: object_id,
			},
			Repository: MetadataRequestResponseRepository{
				ID: repo_id,
			},
			Time: MetadataRequestResponseTime{
				Start: TimestampNow(),
				Stop:  0,
			},
		},
	}, nil
}

// Response struct embeds the Request struct and adds fields and methods
// specific to the response, such as the results from the detection services.
type Response struct {
	// embed the MetadataRequestResponse struct
	MetadataRequestResponse
	// Results is a slice of detection results from the detection services.
	Results []Result `json:"results"`
}

// NewResponse() function initializes a new Response object from the
// provided request.
func NewResponse(req *Request) *Response {
	return &Response{
		MetadataRequestResponse: req.MetadataRequestResponse,
		Results:                 make([]Result, 0),
	}
}

// Result struct contains the detection results from a single service in
// response to a request to process a portion of the source object/text.
type Result struct {
	// Category is the result type
	Category string `json:"category"`
	// ConfidenceScore is specific to the extracted result and
	// is a value between 0 and 1, where 0.99 represents extreme
	// confidence that PHI/PII data was detected.
	ConfidenceScore float64 `json:"confidenceScore"`
	// Length is the number of characters in the result text.
	Length int `json:"length"`
	// Offset is the start position of the result text within the source
	// text, which may have its own offset.
	Offset int `json:"offset"`
	// Subcategory is the (optional) result sub-type.
	Subcategory string `json:"subcategory"`
	// Text is the result text as it appears in the document
	Text string `json:"text"`
	// Service is the location (e.g. URL) of the service that processed
	// the request and returned the result.
	Service string `json:"service"`
}

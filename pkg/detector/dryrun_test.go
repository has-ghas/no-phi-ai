package detector

import (
	"context"
	"testing"

	"github.com/has-ghas/no-phi-ai/pkg/scannerv2"
	"github.com/stretchr/testify/assert"
)

// TestNewDryRunPhiDetector() unit test function tests the
// NewDryRunPhiDetector() function.
func TestNewDryRunPhiDetector(t *testing.T) {
	t.Parallel()

	d := NewDryRunPhiDetector()
	assert.NotNil(t, d)
}

// TestDryRunPhiDetector_Run() unit test function tests the
// Run() method of the DryRunPhiDetector struct.
func TestDryRunPhiDetector_Run(t *testing.T) {
	t.Parallel()

	d := &DryRunPhiDetector{}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	chan_requests_in := make(chan scannerv2.Request)
	chan_responses_out := make(chan scannerv2.Response)

	go d.Run(ctx, chan_requests_in, chan_responses_out)

	request_1 := scannerv2.Request{
		MetadataRequestResponse: scannerv2.MetadataRequestResponse{
			ID: "request-1",
			Commit: scannerv2.MetadataRequestResponseCommit{
				ID: "commit-1",
			},
			Object: scannerv2.MetadataRequestResponseObject{
				ID: "object-1",
			},
			Repository: scannerv2.MetadataRequestResponseRepository{
				ID: "repository-1",
			},
		},
	}
	request_2 := scannerv2.Request{
		MetadataRequestResponse: scannerv2.MetadataRequestResponse{
			ID: "request-2",
			Commit: scannerv2.MetadataRequestResponseCommit{
				ID: "commit-2",
			},
			Object: scannerv2.MetadataRequestResponseObject{
				ID: "object-2",
			},
			Repository: scannerv2.MetadataRequestResponseRepository{
				ID: "repository-2",
			},
		},
	}
	request_3 := scannerv2.Request{
		MetadataRequestResponse: scannerv2.MetadataRequestResponse{
			ID: "request-3",
			Commit: scannerv2.MetadataRequestResponseCommit{
				ID: "commit-3",
			},
			Object: scannerv2.MetadataRequestResponseObject{
				ID: "object-3",
			},
			Repository: scannerv2.MetadataRequestResponseRepository{
				ID: "repository-3",
			},
		},
	}
	// define the expected result common to both responses
	expectedResult := scannerv2.Result{
		Category:        DryRunCategory,
		ConfidenceScore: DryRunConfidenceScore,
		Length:          DryRunLength,
		Offset:          DryRunOffset,
		Service:         DryRunService,
		Subcategory:     DryRunSubcategory,
		Text:            DryRunText,
	}

	chan_requests_in <- request_1
	response_1 := <-chan_responses_out
	assert.Equal(t, []scannerv2.Result{expectedResult}, response_1.Results)
	assert.Equal(t, request_1.MetadataRequestResponse, response_1.MetadataRequestResponse)

	chan_requests_in <- request_2
	response_2 := <-chan_responses_out
	assert.Equal(t, []scannerv2.Result{expectedResult}, response_2.Results)
	assert.Equal(t, request_2.MetadataRequestResponse, response_2.MetadataRequestResponse)

	chan_requests_in <- request_3
	response_3 := <-chan_responses_out
	assert.Equal(t, []scannerv2.Result{expectedResult}, response_3.Results)
	assert.Equal(t, request_3.MetadataRequestResponse, response_3.MetadataRequestResponse)

	// cancel the context to stop the DryRunPhiDetector() function
	cancel()
	// the response channel should be closed
	_, ok := <-chan_responses_out
	assert.False(t, ok, "attempt to read from closed channel should return false")
}

package scannerv2

import (
	"context"

	"github.com/rs/zerolog"
)

// DryRunPhiDetector() function listens for requests, performs no operations
// other than translating the Request to a new Response, and sends responses
// using the provided channels.
//
// Useful for testing the performance of the Scanner in generating requests
// and processing responses for chunks of text contained within the files and
// commits of the scanned repositories.
func DryRunPhiDetector(
	ctx context.Context,
	chan_requests_in <-chan Request,
	chan_responses_out chan<- Response,
) {
	logger := zerolog.Ctx(ctx)
	logger.Info().Msg("started dry run detector")
	defer logger.Info().Msg("finished dry run detector")

	for {
		select {
		case <-ctx.Done():
			logger.Warn().Msg("finished dry run detector : context done")
			// exit the function when the context is done
			return
		case request := <-chan_requests_in:
			// create a Response from the Request metadata
			response := NewResponse(&request)
			// create a dummy Result for the Response
			result := Result{
				Category:        "dry-run_category",
				ConfidenceScore: 0.0,
				Length:          100,
				Offset:          0,
				Service:         "dry-run_service",
				Subcategory:     "dry-run_subcategory",
				Text:            "dry-run_text",
			}
			// add the dummy Result to the Response
			response.Results = append(response.Results, result)
			// send the Response to the output channel
			chan_responses_out <- response
			logger.Info().Msgf(
				"dry run detector processed request ID = %s : RepositoryID = %s : CommitID = %s : ObjectID = %s",
				request.ID,
				request.Repository.ID,
				request.Commit.ID,
				request.Object.ID,
			)
		}
	}
}

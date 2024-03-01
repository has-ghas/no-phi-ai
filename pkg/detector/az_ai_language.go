package detector

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/has-ghas/no-phi-ai/pkg/rrr"
)

// AzAiLanguagePhiDetector struct type is a wrapper for the Run() method.
type AzAiLanguagePhiDetector struct{}

// NewAzAiLanguagePhiDetector() function returns a new AzAiLanguagePhiDetector instance.
func NewAzAiLanguagePhiDetector() *AzAiLanguagePhiDetector {
	return &AzAiLanguagePhiDetector{}
}

// Run() method listens for requests, performs no operations other than
// translating the rrr.Request to a new rrr.Response, and sends responses
// using the provided channels.
//
// Useful for testing the performance of the Scanner in generating requests
// and processing responses for chunks of text contained within the files and
// commits of the scanned repositories.
func (detector *AzAiLanguagePhiDetector) Run(
	ctx context.Context,
	chan_requests_in <-chan rrr.Request,
	chan_responses_out chan<- rrr.Response,
) {
	defer close(chan_responses_out)

	logger := zerolog.Ctx(ctx)
	logger.Info().Msg("started dry run detector")
	defer logger.Info().Msg("finished dry run detector")

	for {
		select {
		case <-ctx.Done():
			logger.Warn().Msg("stopping dry run detector : context done")
			// exit the function when the context is done
			return
		case request := <-chan_requests_in:
			// send the request to the Azure AI Language service API
			logger.Debug().Msgf("sending request ID=%s to Azure AI Language service", request.ID)
			//
			// TODO
			//
		}
	}
}

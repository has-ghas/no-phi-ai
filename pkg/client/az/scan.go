package az

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// ScanDocuments() method listens on the provided documentsChan channel for new
// documents to be scanned, batching them into groups of up to RequestDocumentLimit
// size before sending them to the Azure AI Language service for detection of
// PHI/PII entities in the ducument text.
func (ai *EntityDetectionAI) ScanDocuments(wg *sync.WaitGroup, ctx context.Context, documentsChan <-chan AsyncDocumentWrapper, errChan chan<- error) {
	defer wg.Done()
	// count the number of documents received
	var doc_count int32
	// create a new map of documents for each request
	doc_map := NewAsyncDocumentWrapperMap()

	for {
		// select from the channels
		select {
		case <-ctx.Done():
			log.Ctx(ctx).Warn().Msg("aync detection : context done")
			return
		case wrapper := <-documentsChan:
			// increment the count of documents received
			doc_count++

			log.Ctx(ctx).Trace().Msgf("received document %d : ID = %s", doc_count, wrapper.Document.ID)
			// check if the document map is full
			if doc_map.isFull() {
				// create and a new PiiEntityRecognitionRequest using the existing
				// map of documents, then convert to an HTTP request and send the
				// request to the Azure AI Language service API
				if err := ai.asyncDetectPiiEntities(ctx, doc_map); err != nil {
					log.Ctx(ctx).Error().Err(err).Msg("failed to send PiiEntityRecognitionRequest")
					// send the error to the error channel
					errChan <- errors.Wrap(err, "failed to send PiiEntityRecognitionRequest")
				}
				// create a new map of documents for the next request
				doc_map = NewAsyncDocumentWrapperMap()
			}

			// attempt to add the document wrapper to the map
			valid, added := doc_map.add(&wrapper)
			if !valid {
				log.Ctx(ctx).Warn().Msg("invalid document received")
				continue
			}
			if !added {
				log.Ctx(ctx).Error().Msg("failed to add document to map")
				continue
			}
			// check if the document map is full
			if doc_map.isFull() {
				// create and a new PiiEntityRecognitionRequest using the existing
				// map of documents, then convert to an HTTP request and send the
				// request to the Azure AI Language service API
				if err := ai.asyncDetectPiiEntities(ctx, doc_map); err != nil {
					log.Ctx(ctx).Error().Err(err).Msg("failed to send PiiEntityRecognitionRequest")
					// send the error to the error channel
					errChan <- errors.Wrap(err, "failed to send PiiEntityRecognitionRequest")
				}
				// create a new map of documents for the next request
				doc_map = NewAsyncDocumentWrapperMap()
			}
			log.Ctx(ctx).Trace().Msgf("finished document %d : ID = %s", doc_count, wrapper.Document.ID)
		}
	}
}

// asyncDetectPiiEntities() method converts the input map of documents to a
// PiiEntityRecognitionRequest, sends the request to the Azure AI Language
// service API, processes the response JSON to a PiiEntityRecognitionResponse,
// and maps the DocumentResponse objects (contained in the
// PiiEntityRecognitionResponse) back to their respective AsyncDocumentWrapper
// objects, and sends the DocuemtnResponse to the response channel of the
// AsyncDocumentWrapper.
func (ai *EntityDetectionAI) asyncDetectPiiEntities(ctx context.Context, doc_map AsyncDocumentWrapperMap) (e error) {

	log.Ctx(ctx).Trace().Msgf(
		"running async detection of PHI/PII entities for %d documents",
		doc_map.length(),
	)

	// convert the document map to a PiiEntityRecognitionRequest
	var ai_request *PiiEntityRecognitionRequest
	ai_request, e = doc_map.toPiiEntityRecognitionRequest()
	if e != nil {
		e = errors.Wrap(e, "failed to create PiiEntityRecognitionRequest")
		return
	}

	timestamp_request := time.Now().Unix()
	// send the request to the Azure AI Language service API and wait for the response
	ai_response, err := ai.requestAiResponse(ctx, ai_request)
	if err != nil {
		e = errors.Wrap(err, "failed async detection of entitites")
	}
	timestamp_response := time.Now().Unix()

	// check for nil response, which is only expected in dry run mode
	if ai_response == nil {
		if ai.dryRun {
			log.Ctx(ctx).Warn().Msg("dry run mode enabled : skipping processing of async response")
			return
		} else {
			e = ErrNilResponseFromAPI
			return
		}
	}

	// map each DocumentResponse back to its AsyncDocumentWrapper
	for _, document_response := range ai_response.Results.Documents {
		document_wrapper, ok := doc_map.get(document_response.ID)
		if !ok {
			log.Ctx(ctx).Warn().Msgf(
				"failed to map DocumentResponse to AsyncDocumentWrapper : ID = %s",
				document_response.ID,
			)
			continue
		}
		// create a new AsyncDocumentResponseWrapper for the DocumentResponse
		document_response_wrapper := NewAsyncDocumentResponseWrapper(&document_response)
		// set the RequestedAt and ReceivedAt timestamps in the response wrapper
		document_response_wrapper.SetRequested(timestamp_request)
		document_response_wrapper.SetResponded(timestamp_response)
		// send the AsyncDocumentResponseWrapper to its response channel
		if doc_send_err := document_wrapper.sendResponse(ctx, document_response_wrapper); doc_send_err != nil {
			log.Ctx(ctx).Warn().Msgf(
				"failed to send DocumentResponse to channel : ID = %s : ERROR = %s",
				document_response.ID,
				doc_send_err.Error(),
			)
			continue
		}
	}

	return
}
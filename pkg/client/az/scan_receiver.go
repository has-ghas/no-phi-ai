package az

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// ScanReceiverMetrics struct is the top-level struct for capturing the metrics
// from requests and responses flowing through the scan receiver.
type ScanReceiverMetrics struct {
	Requests  ScanReceiverMetricsRequests  `json:"requests"`
	Responses ScanReceiverMetricsResponses `json:"responses"`
}

// NewScanReceiverMetrics() function initializes a new ScanReceiverMetrics
// object for use in tracking metrics for requests and responses processed by
// the scan receiver.
func NewScanReceiverMetrics() *ScanReceiverMetrics {
	return &ScanReceiverMetrics{
		Requests:  NewScanReceiverMetricsRequests(),
		Responses: NewScanReceiverMetricsResponses(),
	}
}

type ScanReceiverMetricsRequests struct {
	Received int `json:"received"`
	Sent     int `json:"sent"`

	mu *sync.RWMutex
}

func NewScanReceiverMetricsRequests() ScanReceiverMetricsRequests {
	return ScanReceiverMetricsRequests{
		Received: 0,
		Sent:     0,
		mu:       &sync.RWMutex{},
	}
}

func (m_req *ScanReceiverMetricsRequests) IncrementReceived() {
	m_req.mu.Lock()
	m_req.Received++
	m_req.mu.Unlock()
}

func (m_req *ScanReceiverMetricsRequests) IncrementSent() {
	m_req.mu.Lock()
	m_req.Sent++
	m_req.mu.Unlock()
}

type ScanReceiverMetricsResponses struct {
	Received int `json:"received"`
	Sent     int `json:"sent"`

	mu *sync.RWMutex
}

func NewScanReceiverMetricsResponses() ScanReceiverMetricsResponses {
	return ScanReceiverMetricsResponses{
		Received: 0,
		Sent:     0,
		mu:       &sync.RWMutex{},
	}
}

func (m_resp *ScanReceiverMetricsResponses) IncrementReceived() {
	m_resp.mu.Lock()
	m_resp.Received++
	m_resp.mu.Unlock()
}

func (m_resp *ScanReceiverMetricsResponses) IncrementSent() {
	m_resp.mu.Lock()
	m_resp.Sent++
	m_resp.mu.Unlock()
}

// ScanReceiver() method listens on the provided documents_chan channel for new
// documents to be scanned, batching them into groups of up to RequestDocumentLimit
// size before sending them to the Azure AI Language service for detection of
// PHI/PII entities in the ducument text.
func (ai *EntityDetectionAI) ScanReceiver(
	ctx context.Context,
	documents_chan <-chan AsyncDocumentWrapper,
	quit_chan <-chan bool,
	err_chan chan<- error,
	done_chan chan<- bool,
) {
	log.Ctx(ctx).Debug().Msg("started scanning for documents to process")
	defer log.Ctx(ctx).Debug().Msg("finished scanning for documents to process")
	defer close(done_chan)

	receiver_chan := make(chan AsyncDocumentWrapper, len(documents_chan))
	defer close(receiver_chan)
	go ai.receiveDocuments(ctx, receiver_chan, err_chan)

	for {
		// select from the channels
		select {
		case <-ctx.Done():
			return
		case <-quit_chan:
			log.Ctx(ctx).Debug().Msg("scan receiver quit channel closed")
			return
		case wrapper := <-documents_chan:
			// send the document wrapper to the receiver channel
			receiver_chan <- wrapper
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

	//log.Ctx(ctx).Debug().Msgf("ai_request:\n%+v\n\nai_response\n%+v", *ai_request, *ai_response)

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

		log.Ctx(ctx).Debug().Msgf("document_wrapper.Document.ID = %s  :  document_response.ID = %+v", document_wrapper.Document.ID, document_response.ID)

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

// receiveDocuments() method receives documents from the documents channel
// and sends them to the Azure AI Language service for entity detection;
// sends any errors to the error channel; returns when the context is done
// or the documents channel is closed.
func (ai *EntityDetectionAI) receiveDocuments(
	ctx context.Context,
	documents_chan <-chan AsyncDocumentWrapper,
	err_chan chan<- error,
) {
	log.Ctx(ctx).Debug().Msg("starting document receiver")
	defer log.Ctx(ctx).Debug().Msg("finished document receiver")
	// count the number of documents received
	var doc_count int32
	// create a new map of documents
	doc_map := NewAsyncDocumentWrapperMap()

	for {
		select {
		case <-ctx.Done():
			return
		case wrapper, ok := <-documents_chan:
			if !ok || wrapper.Document == nil {
				// create and send a new PiiEntityRecognitionRequest for
				// any remaining documents in the map
				if !doc_map.isEmpty() {
					log.Ctx(ctx).Debug().Msg("documents receiver sending final request for remaining documents")
					doc_map_copy := doc_map.deepCopy()
					if err := ai.asyncDetectPiiEntities(ctx, doc_map_copy); err != nil {
						log.Ctx(ctx).Error().Err(err).Msg("failed to send PiiEntityRecognitionRequest")
						err_chan <- errors.Wrap(err, "failed to send PiiEntityRecognitionRequest")
					}
				}
				log.Ctx(ctx).Debug().Msg("closing documents receiver")
				return
			}
			// increment the count of documents received
			doc_count++

			log.Ctx(ctx).Trace().Msgf("received document %d : ID = %s", doc_count, wrapper.Document.ID)
			// check if the document map is full
			if doc_map.isFull() {
				// create and a new PiiEntityRecognitionRequest using the existing
				// map of documents, then convert to an HTTP request and send the
				// request to the Azure AI Language service API
				doc_map_copy := doc_map.deepCopy()
				if err := ai.asyncDetectPiiEntities(ctx, doc_map_copy); err != nil {
					log.Ctx(ctx).Error().Err(err).Msg("failed to send PiiEntityRecognitionRequest")
					// send the error to the error channel
					err_chan <- errors.Wrap(err, "failed to send PiiEntityRecognitionRequest")
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
				doc_map_copy := doc_map.deepCopy()
				if err := ai.asyncDetectPiiEntities(ctx, doc_map_copy); err != nil {
					log.Ctx(ctx).Error().Err(err).Msg("failed to send PiiEntityRecognitionRequest")
					// send the error to the error channel
					err_chan <- errors.Wrap(err, "failed to send PiiEntityRecognitionRequest")
				}
				log.Ctx(ctx).Debug().Msg("replacing async document map wrapper")
				// create a new map of documents for the next request
				doc_map = NewAsyncDocumentWrapperMap()
			}
			log.Ctx(ctx).Trace().Msgf("finished document %d : ID = %s", doc_count, wrapper.Document.ID)
		}
	}
}

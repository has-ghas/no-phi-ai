package az

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/has-ghas/no-phi-ai/pkg/cfg"
	"github.com/rs/zerolog/log"
)

// EntityDetectionAI struct provides methods for (1) sending requests to
// detect entities of interest in natural language documents and (2) processing
// responses from the Azure AI Language service.
type EntityDetectionAI struct {
	client     *http.Client
	confidence float64
	endpoint   string
	key        string
}

// NewEntityDetectionAI() function requires the Azure service host and
// authentication key as inputs, and returns an initialized EntityDetectionAI.
func NewEntityDetectionAI(c *cfg.Config) (*EntityDetectionAI, error) {
	if len(c.AzureAI.AuthKey) == 0 {
		err := errors.New("EntityDetectionAI requires a valid authentication key")
		return nil, err
	}
	if len(c.AzureAI.Service) == 0 {
		err := errors.New("EntityDetectionAI requires a valid service address")
		return nil, err
	}

	// add query param(s) to the endpoint for the Azure AI Languge service
	//  e.g. to "show stats" on document analysis
	endpoint := fmt.Sprintf(
		"%s/%s",
		c.AzureAI.Service,
		GetDefaultDetectionApi(c.AzureAI.ShowStats),
	)

	return &EntityDetectionAI{
		client:     &http.Client{},
		confidence: c.AzureAI.ConfidenceThreshold,
		endpoint:   endpoint,
		key:        c.AzureAI.AuthKey,
	}, nil
}

// AsyncDetectPiiEntities() method wraps the DetectPiiEntities() method
// and sends the pass|fail (bool) result to a channel for asynchronous
// processing.
func (ai *EntityDetectionAI) AsyncDetectPiiEntities(ctx context.Context, requestData *PiiEntityRecognitionRequest, resultChan chan<- bool) {
	detected, e := ai.DetectPiiEntities(ctx, requestData)
	if e != nil {
		log.Ctx(ctx).Error().Msgf("error detecting entities: %s", e)
	}
	resultChan <- detected
}

// DetectPiiEntities() method accepts a PiiEntityRecognitionRequest as
// input and returns an error (or nil) status depending upon the
// PiiEntityRecognitionResponse from the Azure AI language service.
func (ai *EntityDetectionAI) DetectPiiEntities(ctx context.Context, requestData *PiiEntityRecognitionRequest) (detected bool, e error) {
	// explicitly set detected to false
	detected = false

	requestBytes, err := json.Marshal(*requestData)
	if err != nil {
		e = err
		log.Ctx(ctx).Error().Msgf("error marshalling data for entity detection request: %s", e)
		return
	}
	// TODO : remove debug logging of requestBytes
	log.Ctx(ctx).Debug().Msgf("entity detection request JSON:\n%s", string(requestBytes))

	req, err := http.NewRequest("POST", ai.endpoint, bytes.NewBuffer(requestBytes))
	if err != nil {
		e = err
		log.Ctx(ctx).Error().Msgf("error creating entity detection request: %s", e)
		return
	}

	// set the required headers for the HTTP request
	ai.setHttpRequestHeaders(req)

	resp, err := ai.client.Do(req)
	if err != nil {
		e = err
		log.Ctx(ctx).Error().Msgf("error executing entity detection request: %s", e)
		return
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		e = err
		log.Ctx(ctx).Error().Msgf("error reading entity detection response body: %s", e)
		return
	}
	log.Ctx(ctx).Debug().Msgf("entity detection AI confidence threshold: %f", ai.confidence)
	// TODO : remove debug logging of responseBody
	log.Ctx(ctx).Debug().Msgf("entity detection response body:\n%s", string(responseBody))

	var textResponse PiiEntityRecognitionResults
	e = json.Unmarshal(responseBody, &textResponse)
	if e != nil {
		log.Ctx(ctx).Error().Msgf("error unmarshalling entity detection response body: %s", e)
		return
	}

	for _, doc := range textResponse.Results.Documents {
		var entities []Entity
		for _, entity := range doc.Entities {
			if entity.ConfidenceScore >= ai.confidence {
				entities = append(entities, entity)
			}
		}
		if len(entities) > 0 {
			// set detected to true if any entities are found over the confidence threshold
			detected = true

			// TODO : remove/replace processing of entities
			result := struct {
				ID       string   `json:"id"`
				Redacted string   `json:"redacted"`
				Entities []Entity `json:"entities"`
			}{
				ID:       doc.ID,
				Redacted: doc.RedactedText,
				Entities: entities,
			}
			var resultBytes []byte
			resultBytes, e = json.MarshalIndent(result, "", "  ")
			if e != nil {
				log.Ctx(ctx).Error().Msgf("error marshalling result within entity detection response: %s", e)
				return
			}
			// TODO : remove debug logging of resultBytes
			log.Ctx(ctx).Debug().Msgf("entity detection result:\n%s", string(resultBytes))
		}
	}

	return
}

// GetServiceEndpoint() method return the full URL of the service API endpoint
func (ai *EntityDetectionAI) GetServiceEndpoint(service string) string {
	return ai.endpoint
}

// setHttpRequestHeaders() method sets the required headers for any HTTP
// request to the Azure AI Language service API.
func (ai *EntityDetectionAI) setHttpRequestHeaders(req *http.Request) {
	req.Header.Add("Ocp-Apim-Subscription-Key", ai.key)
	req.Header.Add("Content-Type", "application/json")
}

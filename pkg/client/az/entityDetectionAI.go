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

	return &EntityDetectionAI{
		client:     &http.Client{},
		confidence: c.AzureAI.ConfidenceThreshold,
		endpoint:   fmt.Sprintf("%s/%s", c.AzureAI.Service, DefaultDetectionApi),
		key:        c.AzureAI.AuthKey,
	}, nil
}

// GetServiceEndpoint() method return the full URL of the service API endpoint
func (ai *EntityDetectionAI) GetServiceEndpoint(service string) string {
	return ai.endpoint
}

// DetectPiiEntities() method accepts a PiiEntityRecognitionRequest as input and
// returns an error (or nil) status depending upon the PiiEntityRecognitionResponse
// from the Azure AI language service.
func (ai *EntityDetectionAI) DetectPiiEntities(ctx context.Context, requestData *PiiEntityRecognitionRequest) error {
	requestBytes, err := json.Marshal(*requestData)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("error marshalling data for entity recognition request: %s", err)
		return err
	}
	// TODO : remove debug logging of requestBytes
	log.Ctx(ctx).Debug().Msgf("entity detection request JSON:\n%s", string(requestBytes))

	req, err := http.NewRequest("POST", ai.endpoint, bytes.NewBuffer(requestBytes))
	if err != nil {
		log.Ctx(ctx).Error().Msgf("error creating entity recognition request: %s", err)
		return err
	}

	req.Header.Add("Ocp-Apim-Subscription-Key", ai.key)
	req.Header.Add("Content-Type", "application/json")

	resp, err := ai.client.Do(req)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("error executing entity recognition request: %s", err)
		return err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("error reading entity recognition response body: %s", err)
		return err
	}
	log.Ctx(ctx).Debug().Msgf("entity detection AI confidence threshold: %f", ai.confidence)
	// TODO : remove debug logging of responseBody
	log.Ctx(ctx).Debug().Msgf("entity detection response body:\n%s", string(responseBody))

	var textResponse PiiEntityRecognitionResults
	err = json.Unmarshal(responseBody, &textResponse)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("error unmarshalling entity recognition response body: %s", err)
		return err
	}

	for _, doc := range textResponse.Results.Documents {
		var entities []Entity
		for _, entity := range doc.Entities {
			if entity.ConfidenceScore >= ai.confidence {
				entities = append(entities, entity)
			}
		}
		if len(entities) > 0 {
			result := struct {
				ID       string   `json:"id"`
				Redacted string   `json:"redacted"`
				Entities []Entity `json:"entities"`
			}{
				ID:       doc.ID,
				Redacted: doc.RedactedText,
				Entities: entities,
			}
			resultBytes, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				log.Ctx(ctx).Error().Msgf("error marshalling result within entity recognition response: %s", err)
				return err
			}
			// TODO : remove debug logging of resultBytes
			log.Ctx(ctx).Debug().Msgf("entity detection result:\n%s", string(resultBytes))
		}
	}

	return nil
}

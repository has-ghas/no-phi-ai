package az

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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
func NewEntityDetectionAI(service, key string) (*EntityDetectionAI, error) {
	if len(key) == 0 {
		err := errors.New("EntityDetectionAI requires a valid authentication key")
		return nil, err
	}
	if len(service) == 0 {
		err := errors.New("EntityDetectionAI requires a valid service address")
		return nil, err
	}

	return &EntityDetectionAI{
		client:     &http.Client{},
		confidence: DefaultConfidenceMinimum,
		endpoint:   fmt.Sprintf("%s/%s", service, DefaultDetectionApi),
		key:        key,
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
		fmt.Println("error marshalling data for entity recognition request:", err)
		return err
	}

	req, err := http.NewRequest("POST", ai.endpoint, bytes.NewBuffer(requestBytes))
	if err != nil {
		fmt.Println("error creating entity recognition request:", err)
		return err
	}

	req.Header.Add("Ocp-Apim-Subscription-Key", ai.key)
	req.Header.Add("Content-Type", "application/json")

	resp, err := ai.client.Do(req)
	if err != nil {
		fmt.Println("error executing entity recognition request:", err)
		return err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error reading entity recognition response body:", err)
		return err
	}

	var textResponse PiiEntityRecognitionResults
	err = json.Unmarshal(responseBody, &textResponse)
	if err != nil {
		fmt.Println("error unmarshalling entity recognition response body:", err)
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
				fmt.Println("error marshalling result:", err)
				return err
			}
			fmt.Println(string(resultBytes))
		}
	}

	return nil
}

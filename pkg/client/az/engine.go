package az

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// EntityDetectionEngine struct provides methods for (1) sending requests to
// detect entities of interest in natural language documents and (2) processing
// responses from the Azure AI Language service.
type EntityDetectionEngine struct {
	client     *http.Client
	confidence float64
	endpoint   string
	key        string
}

// NewEntityDetectionEngine() function requires the Azure service host and
// authentication key as inputs, and returns an initialized EntityDetectionEngine.
func NewEntityDetectionEngine(service, key string) (*EntityDetectionEngine, error) {
	if len(key) == 0 {
		err := errors.New("EntityDetectionEngine requires a valid authentication key")
		return nil, err
	}
	if len(service) == 0 {
		err := errors.New("EntityDetectionEngine requires a valid service address")
		return nil, err
	}

	return &EntityDetectionEngine{
		client:     &http.Client{},
		confidence: DefaultConfidenceMinimum,
		endpoint:   fmt.Sprintf("%s/%s", service, DefaultDetectionApi),
		key:        key,
	}, nil
}

// GetServiceEndpoint() method return the full URL of the service API endpoint
func (ede *EntityDetectionEngine) GetServiceEndpoint(service string) string {
	return ede.endpoint
}

// DetectPiiEntities() method accepts a PiiEntityRecognitionRequest as input and
// returns an error (or nil) status depending upon the PiiEntityRecognitionResponse
// from the Azure AI language service.
func (ede *EntityDetectionEngine) DetectPiiEntities(requestData *PiiEntityRecognitionRequest) error {
	requestBytes, err := json.Marshal(*requestData)
	if err != nil {
		fmt.Println("error marshalling data for entity recognition request:", err)
		return err
	}

	req, err := http.NewRequest("POST", ede.endpoint, bytes.NewBuffer(requestBytes))
	if err != nil {
		fmt.Println("error creating entity recognition request:", err)
		return err
	}

	req.Header.Add("Ocp-Apim-Subscription-Key", ede.key)
	req.Header.Add("Content-Type", "application/json")

	resp, err := ede.client.Do(req)
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
			if entity.ConfidenceScore >= ede.confidence {
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

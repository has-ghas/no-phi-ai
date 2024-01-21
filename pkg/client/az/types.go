package az

type AnalysisInput struct {
	Documents []Document `json:"documents"`
}

type Document struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Language string `json:"language"`
}

func NewDocument(id, text, language string) Document {
	if len(language) == 0 {
		language = DefaultLanguage
	}

	return Document{
		Language: language,
		ID:       id,
		Text:     text,
	}
}

type DocumentResponse struct {
	RedactedText string    `json:"redactedText"`
	ID           string    `json:"id"`
	Entities     []Entity  `json:"entities"`
	Warnings     []Warning `json:"warnings"`
}

type Entity struct {
	Text            string  `json:"text"`
	Category        string  `json:"category"`
	Offset          int     `json:"offset"`
	Length          int     `json:"length"`
	ConfidenceScore float64 `json:"confidenceScore"`
}

type Error struct{}

type Parameters struct {
	Domain string `json:"domain"`
}

type PiiEntityRecognitionResults struct {
	Kind    string  `json:"kind"`
	Results Results `json:"results"`
}

type PiiEntityRecognitionRequest struct {
	Kind          string        `json:"kind"`
	AnalysisInput AnalysisInput `json:"analysisInput"`
	Parameters    Parameters    `json:"parameters"`
}

func NewPiiEntityRecognitionRequest(documents []Document) *PiiEntityRecognitionRequest {
	return &PiiEntityRecognitionRequest{
		Kind: "PiiEntityRecognition",
		AnalysisInput: AnalysisInput{
			Documents: documents,
		},
		Parameters: Parameters{
			Domain: DefaultDomain,
		},
	}
}

type Results struct {
	Documents    []DocumentResponse `json:"documents"`
	Errors       []Error            `json:"errors"`
	ModelVersion string             `json:"modelVersion"`
}

type Warning struct{}

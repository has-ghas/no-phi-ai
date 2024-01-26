package az

// ref: https://learn.microsoft.com/en-us/rest/api/language/text-analysis-runtime/analyze-text?view=rest-language-2023-04-01&tabs=HTTP#multilanguageanalysisinput
type AnalysisInput struct {
	Documents []Document `json:"documents"`
}

// ref: https://learn.microsoft.com/en-us/rest/api/language/text-analysis-runtime/analyze-text?view=rest-language-2023-04-01&tabs=HTTP#multilanguageinput
type Document struct {
	ID       string `json:"id"`
	Language string `json:"language"`
	Text     string `json:"text"`
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

// ref: https://learn.microsoft.com/en-us/rest/api/language/text-analysis-runtime/analyze-text?view=rest-language-2023-04-01&tabs=HTTP#error
type DocumentError struct {
	ID    string `json:"id"`
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Target  string `json:"target"`
	} `json:"error"`
}

// ref: https://learn.microsoft.com/en-us/rest/api/language/text-analysis-runtime/analyze-text?view=rest-language-2023-04-01&tabs=HTTP#documents
type DocumentResponse struct {
	Entities     []Entity           `json:"entities"`
	ID           string             `json:"id"`
	RedactedText string             `json:"redactedText"`
	Statistics   DocumentStatistics `json:"statistics"`
	Warnings     []Warning          `json:"warnings"`
}

// ref: https://learn.microsoft.com/en-us/rest/api/language/text-analysis-runtime/analyze-text?view=rest-language-2023-04-01&tabs=HTTP#documentstatistics
type DocumentStatistics struct {
	// number of text elements recognized in the document
	CharactersCount int `json:"charactersCount"`
	// number of transactions processed for the document
	TransactionsCount int `json:"transactionsCount"`
}

// ref: https://learn.microsoft.com/en-us/rest/api/language/text-analysis-runtime/analyze-text?view=rest-language-2023-04-01&tabs=HTTP#entity
type Entity struct {
	// Category is the entity type
	Category string `json:"category"`
	// ConfidenceScore is specific to the extracted entity and
	// is a value between 0 and 1, where 0.99 represents extreme
	// confidence that the entity was correctly recognized
	ConfidenceScore float64 `json:"confidenceScore"`
	// Length of the entity text
	Length int `json:"length"`
	// Offset is the start position of the entity text
	Offset int `json:"offset"`
	// Subcategory is the (optional) entity sub-type.
	Subcategory string `json:"subcategory"`
	// Text is the entity text as it appears in the document
	Text string `json:"text"`
}

// ref: https://learn.microsoft.com/en-us/rest/api/language/text-analysis-runtime/analyze-text?view=rest-language-2023-04-01&tabs=HTTP#piitaskparameters
type Parameters struct {
	// domain can be "none" or "phi", but should pretty much
	// always be "phi" for the purposes of this app
	//
	// (default = "phi")
	Domain string `json:"domain"`
	// (default = true)
	LoggingOptOut bool `json:"loggingOptOut"`
	// modelVersion is the version of the model to use
	// for the analysis (default = "latest")
	ModelVersion string `json:"modelVersion"`
	// piiCategories is the explicit list of PII categories
	// for which "entities" should be detected
	//
	// ref:
	// - https://learn.microsoft.com/en-us/rest/api/language/text-analysis-runtime/analyze-text?view=rest-language-2023-04-01&tabs=HTTP#piicategory
	// - https://learn.microsoft.com/en-us/azure/ai-services/language-service/personally-identifiable-information/how-to-call#select-which-entities-to-be-returned
	//
	// (default = ["Default"])
	PiiCategories []string `json:"piiCategories"`
}

// ref: https://learn.microsoft.com/en-us/rest/api/language/text-analysis-runtime/analyze-text?view=rest-language-2023-04-01&tabs=HTTP#piitaskresult
type PiiEntityRecognitionResults struct {
	Kind    string  `json:"kind"`
	Results Results `json:"results"`
}

// ref: https://learn.microsoft.com/en-us/rest/api/language/text-analysis-runtime/analyze-text?view=rest-language-2023-04-01&tabs=HTTP#analyzetextpiientitiesrecognitioninput
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
			Domain:        "phi",
			LoggingOptOut: true,
			ModelVersion:  "latest",
			PiiCategories: []string{"Default"},
		},
	}
}

// RequestStatistics will only be returned in the response from the API
// if showState=true was passed as a URI parameter in the request to the
// Azure AI Language service.
//
// ref: https://learn.microsoft.com/en-us/rest/api/language/text-analysis-runtime/analyze-text?view=rest-language-2023-04-01&tabs=HTTP#requeststatistics
type RequestStatistics struct {
	// number of documents submitted in the request
	DocumentsCount int `json:"documentsCount"`
	// number of invalid documents submitted in the request
	ErroneousDocumentsCount int `json:"erroneousDocumentsCount"`
	// number of transactions processed for the request
	TransactionsCount int `json:"transactionsCount"`
	// number of valid documents submitted in the request
	ValidDocumentsCount int `json:"validDocumentsCount"`
}

// Results is the contents of the response from the API
//
// ref: https://learn.microsoft.com/en-us/rest/api/language/text-analysis-runtime/analyze-text?view=rest-language-2023-04-01&tabs=HTTP#piiresult
type Results struct {
	Documents    []DocumentResponse `json:"documents"`
	Errors       []DocumentError    `json:"errors"`
	ModelVersion string             `json:"modelVersion"`
	Statistics   RequestStatistics  `json:"statistics"`
}

// ref: https://learn.microsoft.com/en-us/rest/api/language/text-analysis-runtime/analyze-text?view=rest-language-2023-04-01&tabs=HTTP#documentwarning
type Warning struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	TargetRef string `json:"targetRef"`
}

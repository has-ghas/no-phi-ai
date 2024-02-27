package scannerv2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestResult_Hash unit test function tests the Hash() method of the Result struct.
func TestResult_Hash(t *testing.T) {
	tests := []struct {
		commit_id       string
		expected_output string
		name            string
		object_id       string
		repo_id         string
		result          Result
	}{
		{
			commit_id:       "test_commit",
			expected_output: "da6544f2c9819e324b61b1e8de214dfe302fb969",
			name:            "TestResultHash1",
			object_id:       "test_object",
			repo_id:         "test_repo",
			result: Result{
				Category:        "test_category",
				ConfidenceScore: 0.99,
				Length:          1234,
				Offset:          500,
				Service:         "test_service",
				Subcategory:     "test_subcategory",
				Text:            "test_text",
			},
		},
		{
			commit_id:       "test_commit",
			expected_output: "b41daeb80879a4e20de48a4cff1be86da8818919",
			name:            "TestResultHash2",
			object_id:       "test_object",
			repo_id:         "test_repo",
			result: Result{
				Category:        "test_category",
				ConfidenceScore: 0.75,
				Length:          10,
				Offset:          20,
				Service:         "test_service",
				Subcategory:     "test_subcategory",
				Text:            "test_text",
			},
		},
		{
			commit_id:       "",
			expected_output: "d20f36ec299a4d84b03e7b3c7afcc9789ca12c31",
			name:            "TestResultHashEmptyIDs",
			object_id:       "",
			repo_id:         "",
			result: Result{
				Category:        "test_category",
				ConfidenceScore: 0.75,
				Length:          10,
				Offset:          20,
				Service:         "test_service",
				Subcategory:     "test_subcategory",
				Text:            "test_text",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := test.result.Hash(test.repo_id, test.commit_id, test.object_id)

			assert.Equal(t, test.expected_output, output)
		})
	}
}

// TestResult_String unit test function tests the String() method of the Result
// struct.
func TestResult_String(t *testing.T) {
	tests := []struct {
		commit_id       string
		expected_output string
		name            string
		object_id       string
		repo_id         string
		result          Result
	}{
		{
			commit_id:       "test_commit",
			expected_output: "test_repo__test_commit__test_object__test_category__0.75__10__20__test_service__test_subcategory__test_text",
			name:            "TestResultString1",
			object_id:       "test_object",
			repo_id:         "test_repo",
			result: Result{
				Category:        "test_category",
				ConfidenceScore: 0.75,
				Length:          10,
				Offset:          20,
				Service:         "test_service",
				Subcategory:     "test_subcategory",
				Text:            "test_text",
			},
		},
		{
			commit_id:       "",
			expected_output: "###@@@###__###@@@###__###@@@###__test_category__0.75__10__20__test_service__test_subcategory__test_text",
			name:            "TestResultStringEmptyIDs",
			object_id:       "",
			repo_id:         "",
			result: Result{
				Category:        "test_category",
				ConfidenceScore: 0.75,
				Length:          10,
				Offset:          20,
				Service:         "test_service",
				Subcategory:     "test_subcategory",
				Text:            "test_text",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := test.result.String(test.repo_id, test.commit_id, test.object_id)
			assert.Equal(t, test.expected_output, output)
		})
	}
}

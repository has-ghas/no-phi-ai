package gh

import (
	"os"

	git "github.com/go-git/go-git/v5"
)

func ConvertFileToChunks(filepath string) ([]string, error) {
	// Read the file
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	// Convert the file data to text
	text := string(data)

	// Split the text into chunks of no more than 500 characters
	var chunks []string
	for len(text) > 0 {
		if len(text) <= 500 {
			chunks = append(chunks, text)
			break
		}
		chunks = append(chunks, text[:500])
		text = text[500:]
	}

	return chunks, nil
}

func CloneRepo(work_dir, repo_url string) (e error) {
	_, e = git.PlainClone(work_dir, false, &git.CloneOptions{
		URL:      repo_url,
		Progress: os.Stdout,
	})
	return
}

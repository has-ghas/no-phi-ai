package nogit

import (
	"errors"
	"strings"
)

// parseRepoNameFromURL() method is used to parse the repository name
// from a GitHub repository URL.
func parseRepoNameFromURL(url string) (string, error) {
	i := strings.LastIndex(url, "/")
	if i == -1 {
		return "", errors.New("could not parse repo name from url = " + url)
	}
	repoName := url[i+1:]

	repoName = strings.TrimSuffix(repoName, ".git")

	return repoName, nil
}

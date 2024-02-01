package nogit

import (
	"errors"
	"strings"
)

// ParseRepoNameFromURL() method is used to parse the repository name
// from a GitHub repository URL.
func ParseRepoNameFromURL(url string) (string, error) {
	i := strings.LastIndex(url, "/")
	if i == -1 {
		return "", errors.New("failed to parse repo name from url = " + url)
	}
	repo_name := url[i+1:]

	repo_name = strings.TrimSuffix(repo_name, ".git")

	return repo_name, nil
}

func ParseOrgNameFromURL(url string) (string, error) {
	i := strings.LastIndex(url, "/")
	if i == -1 {
		return "", errors.New("failed to parse org name from url = " + url)
	}
	org_name := url[i+1:]

	if strings.Contains(org_name, ".git") {
		j := strings.LastIndex(url[:i], "/")
		if j == -1 {
			return "", errors.New("failed to parse org name from url = " + url)
		}
	}

	return org_name, nil
}

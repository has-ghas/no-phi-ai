package gh

import (
	"context"

	"github.com/google/go-github/v58/github"
	"github.com/pkg/errors"
)

const LabelCleanPHI string = "Clean scan by no-phi-ai"
const LabelDirtyPHI string = "PHI detected by AI"

func applyLabelsToIssue(ctx context.Context, client *github.Client, owner string, repo string, issueNum int, labelsToAdd, labelsToRemove []string) error {
	if len(labelsToAdd) == 0 && len(labelsToRemove) == 0 {
		return errors.New("cannot apply labels : input label lists are empty")
	}

	issue, _, err := client.Issues.Get(ctx, owner, repo, issueNum)
	if err != nil {
		return err
	}

	newLabels := []string{}
	// check if each label has already been applied to the issue
	for _, label := range labelsToAdd {
		if !hasLabel(issue.Labels, label) {
			newLabels = append(newLabels, label)
		}
	}
	// add newLabels to the issue
	if _, _, err := client.Issues.AddLabelsToIssue(ctx, owner, repo, issueNum, newLabels); err != nil {
		return err
	}

	// TODO : remove labelsToRemove from the issue

	return nil
}

func hasLabel(labels []*github.Label, label string) bool {
	for _, l := range labels {
		if l.GetName() == label {
			return true
		}
	}
	return false
}

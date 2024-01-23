package gh

import (
	"context"

	"github.com/google/go-github/v58/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
)

func (cms *ClientManager) ApplyLabelForIssueComment(ctx context.Context, event github.IssueCommentEvent, label string) error {

	if label == "" {
		return errors.New("cannot apply empty label to issue comment")
	}

	installationID := githubapp.GetInstallationIDFromEvent(&event)
	issueNum := event.GetIssue().GetNumber()
	repo := event.GetRepo()
	repoOwner := repo.GetOwner().GetLogin()
	repoName := repo.GetName()

	// use the NewInstallationClient() method from the githubapp.ClientCreator interface
	client, err := cms.NewInstallationClient(installationID)
	if err != nil {
		return err
	}

	labelsAdd := []string{label}
	labelsRm := []string{}

	ctx, logger := githubapp.PreparePRContext(ctx, installationID, repo, issueNum)

	if err := applyLabelsToIssue(ctx, client, repoOwner, repoName, issueNum, labelsAdd, labelsRm); err != nil {
		logger.Error().Err(err).Msg("failed to comment on pull request")
	}
	logger.Info().Msgf("updated issue_#=%d with label=%s", issueNum, label)

	return nil
}

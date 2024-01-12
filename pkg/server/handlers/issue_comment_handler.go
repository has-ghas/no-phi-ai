package handlers

import (
	"context"
	"encoding/json"
	//"fmt"
	//"strings"

	"github.com/google/go-github/v57/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type IssueCommentHandler struct {
	githubapp.ClientCreator

	Preamble string
}

func (h *IssueCommentHandler) Handles() []string {
	return []string{EventTypeIssueComment}
}

func (h *IssueCommentHandler) Handle(ctx context.Context, eventType, deliveryID string, payload []byte) error {
	var event github.IssueCommentEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return errors.Wrap(err, "failed to parse payload for event type="+EventTypeIssueComment)
	}
	zerolog.Ctx(ctx).Debug().Msgf("%s received webhook event type=%s", h.name(), eventType)
	// TODO : remove vulnerable use of payload as unfiltered input to logging function
	zerolog.Ctx(ctx).Debug().Msgf("%s received webhook event:\n%s", h.name(), string(payload))

	if !event.GetIssue().IsPullRequest() {
		zerolog.Ctx(ctx).Debug().Msg("issue comment event is not for a pull request")
		return nil
	}

	/*
		repo := event.GetRepo()
		prNum := event.GetIssue().GetNumber()
		installationID := githubapp.GetInstallationIDFromEvent(&event)

		ctx, logger := githubapp.PreparePRContext(ctx, installationID, repo, event.GetIssue().GetNumber())

		if event.GetAction() != "created" {
			return nil
		}

		client, err := h.NewInstallationClient(installationID)
		if err != nil {
			return err
		}

		repoOwner := repo.GetOwner().GetLogin()
		repoName := repo.GetName()
		author := event.GetComment().GetUser().GetLogin()
		body := event.GetComment().GetBody()

		if strings.HasSuffix(author, "[bot]") {
			logger.Debug().Msg("issue comment was created by a bot")
			return nil
		}

		logger.Debug().Msgf("echoing comment on %s/%s#%d by %s", repoOwner, repoName, prNum, author)
		msg := fmt.Sprintf("%s\n%s said\n```\n%s\n```\n", h.Preamble, author, body)
		prComment := github.IssueComment{
			Body: &msg,
		}

		if _, _, err := client.Issues.CreateComment(ctx, repoOwner, repoName, prNum, &prComment); err != nil {
			logger.Error().Err(err).Msg("failed to comment on pull request")
		}
	*/

	return nil
}

// IssueCommentHandler.name() method is NOT required by any interface.
func (h *IssueCommentHandler) name() string {
	return "IssueCommentHandler"
}

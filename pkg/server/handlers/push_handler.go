package handlers

import (
	"context"
	"encoding/json"

	"github.com/google/go-github/v57/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/has-ghas/no-phi-ai/pkg/client/az"
)

type PushHandler struct {
	githubapp.ClientCreator

	AI *az.EntityDetectionAI
}

func (h *PushHandler) Handles() []string {
	return []string{EventTypePush}
}

func (h *PushHandler) Handle(ctx context.Context, eventType, deliveryID string, payload []byte) error {
	var event github.PushEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return errors.Wrap(err, "failed to parse payload for event type="+EventTypePush)
	}
	zerolog.Ctx(ctx).Debug().Msgf("%s received webhook event type=%s", h.name(), eventType)
	// TODO : remove vulnerable use of payload as unfiltered input to logging function
	zerolog.Ctx(ctx).Debug().Msgf("%s received webhook event:\n%s", h.name(), string(payload))

	// TODO

	return nil
}

// PushHandler.name() method is NOT required by any interface.
func (h *PushHandler) name() string {
	return "PushHandler"
}

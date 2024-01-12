package handlers

import (
	"context"
	"encoding/json"

	"github.com/google/go-github/v57/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type InstallationHandler struct {
	githubapp.ClientCreator
}

func (h *InstallationHandler) Handles() []string {
	return []string{EventTypeInstallation}
}

func (h *InstallationHandler) Handle(ctx context.Context, eventType, deliveryID string, payload []byte) error {
	var event github.InstallationEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return errors.Wrap(err, "failed to parse payload for event type="+EventTypeInstallation)
	}
	zerolog.Ctx(ctx).Debug().Msgf("%s received webhook event type=%s", h.name(), eventType)
	// TODO : remove vulnerable use of payload as unfiltered input to logging function
	zerolog.Ctx(ctx).Debug().Msgf("%s received webhook event:\n%s", h.name(), string(payload))

	// TODO

	return nil
}

// InstallationHandler.name() method is NOT required by any interface.
func (h *InstallationHandler) name() string {
	return "InstallationHandler"
}

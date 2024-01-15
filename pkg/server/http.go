package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gregjones/httpcache"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/rcrowley/go-metrics"
	"github.com/rs/zerolog/log"

	"github.com/has-ghas/no-phi-ai/pkg/cfg"
	"github.com/has-ghas/no-phi-ai/pkg/client/az"
	"github.com/has-ghas/no-phi-ai/pkg/server/handlers"
)

// Run() function auto-loads configuration from file and env vars,
// then uses this configuration to setup clients for Azure AI Language
// service APIs and GitHub APIs before registering HTTP handlers for
// GitHub webhook events.
func Run() error {
	// parse config from file and env vars, where env vars take precedence
	//
	// use the config as the basis for setting up the HTTP server and
	// registering HTTP handlers for GitHub webhook events
	config, err := cfg.ParseConfig()
	if err != nil {
		return err
	}

	// TODO
	metricsRegistry := metrics.DefaultRegistry

	// create a common githubapp.ClientCreator, which can be used to get an
	// installation client for interacting with GitHub APIs
	cc, err := githubapp.NewDefaultCachingClientCreator(
		config.GitHub,
		githubapp.WithClientUserAgent(cfg.AppUserAgent),
		githubapp.WithClientTimeout(3*time.Second),
		githubapp.WithClientCaching(false, func() httpcache.Cache { return httpcache.NewMemoryCache() }),
		githubapp.WithClientMiddleware(
			githubapp.ClientMetrics(metricsRegistry),
		),
	)
	if err != nil {
		return err
	}

	// create a common *az.EntityDetectionAI, which can be used for
	// detecting "entities" of interest via the Azure AI Language service
	ai, ai_err := az.NewEntityDetectionAI(config.AzureAI.Service, config.AzureAI.AuthKey)
	if ai_err != nil {
		return ai_err
	}

	// define the event handlers
	installationHandler := &handlers.InstallationHandler{
		ClientCreator: cc,
	}
	issueCommentHandler := &handlers.IssueCommentHandler{
		AI:            ai,
		ClientCreator: cc,
		Preamble:      config.App.PullRequestPreamble,
	}
	pullRequestHandler := &handlers.PullRequestHandler{
		ClientCreator: cc,
	}
	pushHandler := &handlers.PushHandler{
		ClientCreator: cc,
	}

	// register the event handlers with a new/default event dispatcher
	eventDispatcher := githubapp.NewDefaultEventDispatcher(
		config.GitHub,
		installationHandler,
		issueCommentHandler,
		pullRequestHandler,
		pushHandler,
	)

	// add the HTTP route associated with the webhook handler
	http.Handle(githubapp.DefaultWebhookRoute, eventDispatcher)

	// run the HTTP server
	addr := fmt.Sprintf("%s:%d", config.Server.Address, config.Server.Port)
	log.Info().Msgf("starting HTTP server on %s", addr)
	if err = http.ListenAndServe(addr, nil); err != nil {
		return err
	}

	return nil
}

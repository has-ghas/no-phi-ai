package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gregjones/httpcache"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/rcrowley/go-metrics"
	"github.com/rs/zerolog/log"

	"github.com/has-ghas/no-phi-ai/pkg/cfg"
	"github.com/has-ghas/no-phi-ai/pkg/server/handlers"
)

func main() {
	// parse config from file and env vars, where env vars take precedence
	config, err := cfg.ParseConfig()
	if err != nil {
		panic(err)
	}

	metricsRegistry := metrics.DefaultRegistry
	// generate a client for interacting with GitHub APIs
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
		panic(err)
	}

	// define the event handlers
	installationHandler := &handlers.InstallationHandler{
		ClientCreator: cc,
	}
	issueCommentHandler := &handlers.IssueCommentHandler{
		ClientCreator: cc,
		Preamble:      config.App.PullRequestPreamble,
	}
	pullRequestHandler := &handlers.PullRequestHandler{
		ClientCreator: cc,
	}
	pushHandler := &handlers.PushHandler{
		ClientCreator: cc,
	}

	// register event handlers with a new/default event dispatcher
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
	log.Info().Msgf("starting server on %s", addr)
	if err = http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}

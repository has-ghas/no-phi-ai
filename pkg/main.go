package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gregjones/httpcache"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/rcrowley/go-metrics"
	"github.com/rs/zerolog"

	"github.com/has-ghas/no-phi-ai/pkg/cfg"
	"github.com/has-ghas/no-phi-ai/pkg/server/handlers"
)

func main() {
	// define flags
	debug := flag.Bool("debug", false, "enable logging at debug level")
	configPath := flag.String("config", "../config/test.yml", "local relative path to the config file")

	// parse flags
	flag.Parse()

	// setup logger and level
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	logger.Debug().Msgf("logger level=%s", zerolog.GlobalLevel())
	zerolog.DefaultContextLogger = &logger

	// load config
	config, err := cfg.ReadConfig(*configPath)
	if err != nil {
		panic(err)
	}
	logger.Debug().Msgf("loaded YAML config from path=%s", *configPath)

	metricsRegistry := metrics.DefaultRegistry
	// generate a client for interacting with GitHub APIs
	cc, err := githubapp.NewDefaultCachingClientCreator(
		config.Github,
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
		config.Github,
		installationHandler,
		issueCommentHandler,
		pullRequestHandler,
		pushHandler,
	)

	// add the HTTP route associated with the webhook handler
	http.Handle(githubapp.DefaultWebhookRoute, eventDispatcher)

	// run the HTTP server
	addr := fmt.Sprintf("%s:%d", config.Server.Address, config.Server.Port)
	logger.Info().Msgf("starting server on %s", addr)
	if err = http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}

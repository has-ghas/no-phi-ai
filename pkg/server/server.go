package server

import (
	"fmt"
	"net/http"
	"time"

	tollbooth "github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth/v6/limiter"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/rs/zerolog/log"

	"github.com/has-ghas/no-phi-ai/pkg/cfg"
	"github.com/has-ghas/no-phi-ai/pkg/client/az"
	"github.com/has-ghas/no-phi-ai/pkg/client/gh"
	"github.com/has-ghas/no-phi-ai/pkg/server/handlers"
)

// Manager struct holds the configuration and state for the HTTP server.
type Manager struct {
	Config *cfg.Config
	Server *http.Server
}

// NewManagerOrDie() function returns a new Manager instance for
// the HTTP server. Generates a fatal error if unable to:
//   - parse the configuration from file and env vars, or...
//   - setup the HTTP server, or...
//   - register HTTP handlers for GitHub webhook events.
func NewManagerOrDie() *Manager {
	var config *cfg.Config
	var err error
	var eventDispatcher http.Handler

	// parse config from file and env vars, where env vars take precedence.
	//
	// use the config as the basis for setting up the HTTP server and
	// registering HTTP handlers for GitHub webhook events.
	config, err = cfg.ParseConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse config for new Manager")
	}

	// setup the rate limiter for the HTTP server
	lmt := tollbooth.NewLimiter(config.Server.RateLimit, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour})
	lmt.SetHeader("Authorization", []string{})
	lmt.SetHeaderEntryExpirationTTL(time.Hour)
	lmt.SetMessage(`{"code":429,"response":"You have reached maximum request limit. Please try again in a few seconds."}`)
	lmt.SetMessageContentType("application/json")

	// set the "mode" for the (gin) HTTP server
	gin.SetMode(gin.ReleaseMode)
	// setup a new router for the HTTP server
	router := gin.New()
	router.Use(requestid.New(
		requestid.WithGenerator(func() string {
			return uuid.NewString()
		}),
	))
	router.Use(gin.Logger())

	// setup an http.Handler as the event dispatcher for GitHub webhook events
	eventDispatcher, err = setupEventDispatcher(config)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to setup event handler for new Manager")
	}

	// setup routes for the HTTP server
	v1 := router.Group(cfg.RouteGroupGHv1)
	v1.POST(cfg.RouteWebhook, handlers.LimitHandler(lmt), gin.WrapH(eventDispatcher))

	// populate the Manager struct
	return &Manager{
		Config: config,
		Server: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", config.Server.Address, config.Server.Port),
			Handler: router,
		},
	}
}

// Serve() method auto-loads configuration from file and env vars,
// then uses this configuration to setup clients for Azure AI Language
// service APIs and GitHub APIs before registering HTTP handlers for
// GitHub webhook events.
func (m *Manager) Serve() {
	m.logRoutes()
	// run the HTTP server
	if err := m.Server.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Msg("runtime error in HTTP server")
	}
}

// logRoutes() function logs the HTTP routes registered with the gin Router.
func (m *Manager) logRoutes() {
	for _, route := range m.Server.Handler.(*gin.Engine).Routes() {
		log.Info().Msgf("serving endpoint -> %s %s%s", route.Method, m.Server.Addr, route.Path)
	}
}

func setupEventDispatcher(config *cfg.Config) (http.Handler, error) {

	ghcm, err := gh.NewClientManager(config)
	if err != nil {
		return nil, err
	}

	// create a common *az.EntityDetectionAI, which can be used for
	// detecting "entities" of interest via the Azure AI Language service
	ai, ai_err := az.NewEntityDetectionAI(config)
	if ai_err != nil {
		return nil, ai_err
	}

	// define the event handlers
	installationHandler := &handlers.InstallationHandler{
		GHCM: ghcm,
	}
	issueCommentHandler := &handlers.IssueCommentHandler{
		AI:       ai,
		GHCM:     ghcm,
		Preamble: config.App.PullRequestPreamble,
	}
	pullRequestHandler := &handlers.PullRequestHandler{
		AI:   ai,
		GHCM: ghcm,
	}
	pushHandler := &handlers.PushHandler{
		AI:   ai,
		GHCM: ghcm,
	}

	// register the event handlers with a new/default event dispatcher
	eventDispatcher := githubapp.NewDefaultEventDispatcher(
		config.GitHub,
		installationHandler,
		issueCommentHandler,
		pullRequestHandler,
		pushHandler,
	)

	return eventDispatcher, nil
}

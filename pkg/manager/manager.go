package manager

import (
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/has-ghas/no-phi-ai/pkg/cfg"
)

// Manager struct holds the configuration and state for app.
type Manager struct {
	Config *cfg.Config
	Logger *zerolog.Logger
	Server *http.Server
}

// New() function returns a new Manager instance for the app.
// Generates a fatal error if unable to:
//   - parse the configuration from file and env vars, or...
//   - setup the HTTP server, or...
//   - register HTTP handlers for GitHub webhook events.
func New() *Manager {
	var config *cfg.Config
	var err error
	var logger *zerolog.Logger

	// parse config from file and env vars, where env vars take precedence.
	//
	// use the config as the basis for setting up the HTTP server and
	// registering HTTP handlers for GitHub webhook events.
	config, logger, err = cfg.ParseConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse config for new Manager")
	}

	// populate the Manager struct
	return &Manager{
		Config: config,
		Logger: logger,
	}
}

func (m *Manager) GetAppMode() string {
	return m.Config.App.Mode
}

func (m *Manager) Init() {
	switch m.GetAppMode() {
	case "cli":
		m.initCLI()
		return
	case "gh_app":
	default:
		m.initServer()
	}
}

// Run() method runs the Manager in the configured mode.
func (m *Manager) Run() {
	switch m.GetAppMode() {
	case "cli":
		m.runCLI()
		return
	case "gh_app":
	default:
		m.runServer()
	}
}

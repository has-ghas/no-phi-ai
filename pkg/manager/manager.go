package manager

import (
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/has-ghas/no-phi-ai/pkg/cfg"
	nogit "github.com/has-ghas/no-phi-ai/pkg/client/no-git"
)

// Manager struct holds the configuration and state for app.
type Manager struct {
	Config *cfg.Config
	Git    *nogit.GitManager
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
		msg := "failed to parse config for new Manager"
		if logger == nil {
			log.Fatal().Err(err).Msg(msg)
		} else {
			logger.Fatal().Err(err).Msg(msg)
		}
	}

	// populate the Manager struct
	return &Manager{
		Config: config,
		Logger: logger,
	}
}

// GetAppMode() method returns the configured app mode for the Manager.
func (m *Manager) GetAppMode() string {
	return m.Config.App.Mode
}

// Init() method runs initialization steps that are specific to the configured mode.
func (m *Manager) Init() {
	m.Logger.Trace().Msg("initializing Manager")
	switch m.GetAppMode() {
	case cfg.AppModeCLI:
		m.initCLI()
		return
	case cfg.AppModeServer:
		m.initServer()
		return
	default:
		m.Logger.Fatal().Msgf("Manager refusing to Init() invalid app mode: %s", m.GetAppMode())
	}
}

// Run() method runs the Manager in the configured mode.
func (m *Manager) Run() {
	m.Logger.Trace().Msg("running Manager")
	switch m.GetAppMode() {
	case cfg.AppModeCLI:
		if err := m.runCLI(); err != nil {
			m.Logger.Fatal().Err(err).Msgf("error running in '%s' mode", m.GetAppMode())
		}
		return
	case cfg.AppModeServer:
		if err := m.runServer(); err != nil {
			m.Logger.Fatal().Err(err).Msgf("error running in '%s' mode", m.GetAppMode())
		}
		return
	default:
		m.Logger.Fatal().Msgf("Manager refusing to Run() invalid app mode: %s", m.GetAppMode())
	}
}

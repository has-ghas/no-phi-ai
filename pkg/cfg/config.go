package cfg

import (
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

// AppConfig struct contains the configuration items used by the running app.
type AppConfig struct {
	LogLevel            string `yaml:"log_level"`
	Name                string `yaml:"name"`
	PullRequestPreamble string `yaml:"pull_request_preamble"`
}

// AzureAIConfig struct contains the configuration items used by the Azure AI API.
type AzureAIConfig struct {
	AuthKey             string  `yaml:"auth_key"`
	ConfidenceThreshold float64 `yaml:"confidence_threshold"`
	Service             string  `yaml:"service"`
}

// ServerConfig struct contains the configuration used to start the HTTP server.
type ServerConfig struct {
	Address   string  `yaml:"address"`
	Port      int     `yaml:"port"`
	RateLimit float64 `yaml:"rate_limit"`
}

// Config struct is the top-level configuration object for the app.
type Config struct {
	App     AppConfig        `yaml:"app"`
	AzureAI AzureAIConfig    `yaml:"azure_ai"`
	GitHub  githubapp.Config `yaml:"github"`
	Server  ServerConfig     `yaml:"server"`
}

// envOverride() method overrides *Config values with values from environment variables.
func (c *Config) envOverride() error {
	if appName := os.Getenv(NOPHI_APP_NAME); appName != "" {
		c.App.Name = appName
	}
	if logLevel := os.Getenv(NOPHI_APP_LOG_LEVEL); logLevel != "" {
		c.App.LogLevel = logLevel
	}
	if integrationID := os.Getenv(NOPHI_GH_INTEGRATION_ID); integrationID != "" {
		integrationIDInt, err := strconv.ParseInt(integrationID, 10, 64)
		if err != nil {
			return errors.Wrap(err, "failed parsing NOPHI_GH_INTEGRATION_ID env var")
		}
		c.GitHub.App.IntegrationID = integrationIDInt
	}
	if privateKey := os.Getenv(NOPHI_GH_PRIVATE_KEY); privateKey != "" {
		c.GitHub.App.PrivateKey = privateKey
	}
	if webhookSecret := os.Getenv(NOPHI_GH_WEBHOOK_SECRET); webhookSecret != "" {
		c.GitHub.App.WebhookSecret = webhookSecret
	}
	if V3APIURL := os.Getenv(NOPHI_GH_V3APIURL); V3APIURL != "" {
		c.GitHub.V3APIURL = V3APIURL
	}
	if V4APIURL := os.Getenv(NOPHI_GH_V4APIURL); V4APIURL != "" {
		c.GitHub.V4APIURL = V4APIURL
	}
	if serverAddress := os.Getenv(NOPHI_SERVER_ADDRESS); serverAddress != "" {
		c.Server.Address = serverAddress
	}
	if serverPort := os.Getenv(NOPHI_SERVER_PORT); serverPort != "" {
		serverPortInt, err := strconv.Atoi(serverPort)
		if err != nil {
			return errors.Wrap(err, "failed parsing SERVER_PORT env var")
		}
		c.Server.Port = serverPortInt
	}

	return nil
}

// setupLogger() method uses the Config settings to setup logging
// for the app.
func (c *Config) setupLogger() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// set the global log level
	switch strings.ToLower(c.App.LogLevel) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "err", "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "warn", "warning":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	zerolog.DefaultContextLogger = &logger
	logger.Info().Msgf("%s app using log_level=%s", c.App.Name, zerolog.GlobalLevel())

	return
}

// default() method sets default values for optional config fields.
func (c *Config) defaultConfig() {
	// set defaults for optional c.App config values
	if c.App.Name == "" {
		c.App.Name = DefaultAppName
	}
	if c.App.LogLevel == "" {
		c.App.LogLevel = DefaultAppLogLevel
	}
	// set defaults for optional c.GitHub config values
	if c.GitHub.V3APIURL == "" {
		c.GitHub.V3APIURL = DefaultGitHubV3APIURL
	}
	// set defaults for optional c.Server config values
	if c.Server.Address == "" {
		c.Server.Address = DefaultServerAddress
	}
	if c.Server.Port == 0 {
		c.Server.Port = DefaultServerPort
	}
	if c.Server.RateLimit == 0 {
		c.Server.RateLimit = DefaultRateLimit
	}
}

// verifyConfig() method returns an error if a required value has not
// been set in the Config, sets defaults for optional values, and/or
// returns a nil error if all required values are set.
func (c *Config) verifyConfig() (e error) {
	// check the c.AzureAI config values
	if c.AzureAI.Service == "" {
		e = errors.New("missing required config value: azure_ai.service")
		return
	}
	if c.AzureAI.AuthKey == "" {
		e = errors.New("missing required config value: azure_ai.auth_key")
		return
	}
	if c.AzureAI.ConfidenceThreshold == 0 {
		c.AzureAI.ConfidenceThreshold = DefaultConfidenceThreshold
	}

	// check the c.GitHub config values
	if c.GitHub.App.IntegrationID == 0 {
		e = errors.New("missing required config value: github.app.integration_id")
		return
	}
	if c.GitHub.App.PrivateKey == "" {
		e = errors.New("missing required config value: github.app.private_key")
		return
	}
	if c.GitHub.App.WebhookSecret == "" {
		e = errors.New("missing required config value: github.app.webhook_secret")
		return
	}

	return
}

// readConfigFile() function reads config data from the input file path,
// or returns an empty Config if path is empty in order to allow for the
// entire config to be provided via environment variables.
func readConfigFile(path string) (*Config, error) {
	var c Config

	if path == "" {
		return &c, nil
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed reading server config file: %s", path)
	}

	if err := yaml.UnmarshalStrict(bytes, &c); err != nil {
		return nil, errors.Wrap(err, "failed parsing server config file")
	}

	log.Debug().Msgf("loaded YAML config from path=%s", path)

	return &c, nil
}

// ParseConfig() function parses the config file and environment variables.
func ParseConfig() (*Config, error) {
	// define flags
	configPath := flag.String("config", "", "local relative path to the config file")

	// parse flags
	flag.Parse()

	// get config path from environment variable if not set via flag
	if *configPath == "" {
		if envConfigPath, found := os.LookupEnv(NOPHI_CONFIG_PATH); found {
			*configPath = envConfigPath
		}
	}

	// get config from file
	c, err := readConfigFile(*configPath)
	if err != nil {
		return c, err
	}

	// override config values with environment variables
	if err := c.envOverride(); err != nil {
		return c, err
	}

	// fill in default values for optional config fields
	c.defaultConfig()

	// verify required config values are set (i.e. not empty)
	if err := c.verifyConfig(); err != nil {
		return c, err
	}

	// setup logger and level
	c.setupLogger()

	return c, nil
}

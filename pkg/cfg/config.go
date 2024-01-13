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

// Config struct is the top-level configuration object for the app.
type Config struct {
	App     AppConfig        `yaml:"app"`
	AzureAI AzureAIConfig    `yaml:"azure_ai"`
	GitHub  githubapp.Config `yaml:"github"`
	Server  ServerConfig     `yaml:"server"`
}

// ServerConfig struct contains the configuration used to start the HTTP server.
type ServerConfig struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

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

// ParseConfig() function parses the config file and environment variables.
func ParseConfig() (*Config, error) {
	// define flags
	configPath := flag.String("config", "", "local relative path to the config file")

	// parse flags
	flag.Parse()

	c, err := configFileRead(*configPath)
	if err != nil {
		return nil, err
	}

	if err := configEnvOverride(c); err != nil {
		return nil, err
	}

	// setup logger and level
	setupLogger(c)

	return c, nil
}

// setupLogger() function sets up logging for the app.
func setupLogger(c *Config) {
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

// configEnvOverride() function overrides config values with values from environment variables.
func configEnvOverride(c *Config) error {
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

// configFileRead() function reads config data from the input file path, or returns an
// empty config if path is empty.
func configFileRead(path string) (*Config, error) {
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

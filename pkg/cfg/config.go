package cfg

import (
	"flag"
	"os"

	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
)

// AppConfig struct contains the configuration items used by the running app.
type AppConfig struct {
	Log  AppLogConfig `yaml:"log"`
	Name string       `yaml:"name"`
}

// AzureAIConfig struct contains the configuration items used by the Azure AI API.
type AzureAIConfig struct {
	AuthKey             string  `yaml:"auth_key"`
	ConfidenceThreshold float64 `yaml:"confidence_threshold"`
	Service             string  `yaml:"service"`
}

// AppLogConfig struct contains the configuration used to initialize the logger.
type AppLogConfig struct {
	// ConsoleEnable controls whether or not to log to standard output
	ConsoleEnable bool `yaml:"console_enable"`
	// ConsolePretty controls whether console output is pretty printed (true)
	// or printed as structured JSON (false)
	ConsolePretty bool `yaml:"console_pretty"`
	// "trace", "debug", "info", "warn", or "error"
	Level string `yaml:"level"`
	// log to standard output -> "" or "stdout"
	// log to a file -> "../relative/file/path" or "/absolute/file/path"
	File string `yaml:"file"`
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

// default() method sets default values for optional config fields.
func (c *Config) defaultConfig() {
	// set defaults for optional c.App config values
	if c.App.Name == "" {
		c.App.Name = DefaultAppName
	}
	if c.App.Log.Level == "" {
		c.App.Log.Level = DefaultAppLogLevel
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

	return &c, nil
}

// ParseConfig() function parses the config file and environment variables.
func ParseConfig() (*Config, *zerolog.Logger, error) {
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
		return c, nil, err
	}

	// fill in default values for optional config fields
	c.defaultConfig()

	// override config values with environment variables
	if err := c.envOverride(); err != nil {
		return c, nil, err
	}

	// verify required config values are set (i.e. not empty)
	if err := c.verifyConfig(); err != nil {
		return c, nil, err
	}

	return c, c.setupLogger(), nil
}

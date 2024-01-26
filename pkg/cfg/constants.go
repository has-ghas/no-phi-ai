package cfg

import "time"

const AppModeCLI string = "cli"
const AppModeServer string = "server"
const AppVersion string = "1.0.0"

const CommandRunHelp string = "help"
const CommandRunListOrgRepos string = "list-org-repos"
const CommandRunScanOrg string = "scan-org"
const CommandRunScanRepos string = "scan-repos"
const CommandRunVersion string = "version"

const DefaultAppLogLevel string = "info"
const DefaultAppMode string = AppModeServer
const DefaultAppName string = "no-phi-ai"
const DefaultAppUserAgent string = DefaultAppName + "/" + AppVersion
const DefaultAzureAIShowStats bool = true
const DefaultClientTimeout time.Duration = 3 * time.Second
const DefaultCommandRun string = CommandRunHelp
const DefaultCommandWorkDir string = "/tmp/" + DefaultAppName
const DefaultConfidenceThreshold float64 = 0.6
const DefaultGitHubV3APIURL string = "https://api.github.com"
const DefaultRateLimit float64 = 1000.0
const DefaultServerAddress string = "127.0.0.1"
const DefaultServerPort int = 8080

const RouteGroupGHv1 string = "/api/v1/github"
const RouteWebhook string = "/hook"

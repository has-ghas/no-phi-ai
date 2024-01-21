package cfg

import "time"

const AppName string = "no-phi-ai"
const AppUserAgent string = AppName + "/" + AppVersion
const AppVersion string = "1.0.0"

const DefaultAppLogLevel string = "info"
const DefaultAppName string = "no-phi-ai"
const DefaultClientTimeout time.Duration = 3 * time.Second
const DefaultConfidenceThreshold float64 = 0.6
const DefaultGitHubV3APIURL string = "https://api.github.com"
const DefaultRateLimit float64 = 1000.0
const DefaultServerAddress string = "127.0.0.1"
const DefaultServerPort int = 8080

const NOPHI_APP_LOG_LEVEL string = "NOPHI_APP_LOG_LEVEL"
const NOPHI_APP_NAME string = "NOPHI_APP_NAME"
const NOPHI_CONFIG_PATH string = "NOPHI_CONFIG_PATH"
const NOPHI_GH_INTEGRATION_ID string = "NOPHI_GH_INTEGRATION_ID"
const NOPHI_GH_PRIVATE_KEY string = "NOPHI_GH_PRIVATE_KEY"
const NOPHI_GH_V3APIURL string = "NOPHI_GH_V3APIURL"
const NOPHI_GH_V4APIURL string = "NOPHI_GH_V4APIURL"
const NOPHI_GH_WEBHOOK_SECRET = "NOPHI_GH_WEBHOOK_SECRET"
const NOPHI_SERVER_ADDRESS string = "NOPHI_SERVER_ADDRESS"
const NOPHI_SERVER_PORT string = "NOPHI_SERVER_PORT"

const RouteGroupGHv1 string = "/api/v1/github"
const RouteWebhook string = RouteGroupGHv1 + "/hook"

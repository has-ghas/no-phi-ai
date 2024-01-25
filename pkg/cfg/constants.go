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

const RouteGroupGHv1 string = "/api/v1/github"
const RouteWebhook string = "/hook"

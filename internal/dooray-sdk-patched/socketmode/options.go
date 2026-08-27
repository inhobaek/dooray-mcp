package socketmode

import (
	"log"
	"net/http"
	"time"
)

const (
	ServiceMessenger = "messenger"
	ServiceTask      = "task"
	ServiceWiki      = "wiki"

	defaultBaseURL      = "https://api.dooray.com"
	defaultPingInterval = 30 * time.Second
	defaultPingTimeout  = 10 * time.Second
	defaultReconnectMin = 1 * time.Second
	defaultReconnectMax = 30 * time.Second
)

// Option configures the Agent.
type Option func(*Agent)

// WithBaseURL sets the Dooray API base URL.
func WithBaseURL(url string) Option {
	return func(a *Agent) {
		a.baseURL = url
	}
}

// WithDomain sets the Dooray domain (e.g. "company").
func WithDomain(domain string) Option {
	return func(a *Agent) {
		a.domain = domain
	}
}

// WithHTTPClient sets a custom HTTP client for REST API calls.
func WithHTTPClient(client *http.Client) Option {
	return func(a *Agent) {
		a.httpClient = client
	}
}

// WithLogger sets a custom logger.
func WithLogger(logger *log.Logger) Option {
	return func(a *Agent) {
		a.logger = logger
	}
}

// WithPingInterval sets the WebSocket ping interval.
func WithPingInterval(d time.Duration) Option {
	return func(a *Agent) {
		a.pingInterval = d
	}
}

// WithReconnectBackoff sets the min and max reconnect backoff durations.
func WithReconnectBackoff(min, max time.Duration) Option {
	return func(a *Agent) {
		a.reconnectMin = min
		a.reconnectMax = max
	}
}

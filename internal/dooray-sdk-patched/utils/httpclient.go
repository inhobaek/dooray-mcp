package utils

import (
	"net"
	"net/http"
	"time"
)

const (
	DefaultConnectTimeout = 3 * time.Second
	DefaultReadTimeout    = 10 * time.Second
)

// NewDefaultHTTPClient creates an *http.Client with sensible default timeouts.
//   - Connection timeout: 3s
//   - Total request timeout (read): 10s
func NewDefaultHTTPClient() *http.Client {
	return NewHTTPClient(DefaultConnectTimeout, DefaultReadTimeout)
}

// NewHTTPClient creates an *http.Client with the given connect and read timeouts.
func NewHTTPClient(connectTimeout, readTimeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: readTimeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: connectTimeout,
			}).DialContext,
			TLSHandshakeTimeout: connectTimeout,
		},
	}
}

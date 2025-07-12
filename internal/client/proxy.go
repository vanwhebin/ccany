package client

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"

	"golang.org/x/net/proxy"
)

// ProxyConfig holds proxy configuration
type ProxyConfig struct {
	Enabled             bool
	Type                string // "http" or "socks5"
	HTTPProxy           string
	SOCKS5Proxy         string
	SOCKS5ProxyUser     string
	SOCKS5ProxyPassword string
	IgnoreSSL           bool
}

// BuildHTTPTransport creates an HTTP transport with proxy support
func BuildHTTPTransport(config *ProxyConfig) *http.Transport {
	var transport *http.Transport

	if !config.Enabled {
		// Return default transport if proxy is disabled
		return &http.Transport{}
	}

	switch config.Type {
	case "http":
		if config.HTTPProxy != "" {
			httpProxy, err := url.Parse(config.HTTPProxy)
			if err == nil && httpProxy != nil {
				transport = &http.Transport{
					Proxy: http.ProxyURL(httpProxy),
				}
			}
		}
	case "socks5":
		if config.SOCKS5Proxy != "" {
			var auth *proxy.Auth
			if config.SOCKS5ProxyUser != "" || config.SOCKS5ProxyPassword != "" {
				auth = &proxy.Auth{
					User:     config.SOCKS5ProxyUser,
					Password: config.SOCKS5ProxyPassword,
				}
			}

			dialer, err := proxy.SOCKS5("tcp", config.SOCKS5Proxy, auth, proxy.Direct)
			if err == nil {
				dialContext := func(ctx context.Context, network, address string) (net.Conn, error) {
					return dialer.Dial(network, address)
				}
				transport = &http.Transport{
					DialContext: dialContext,
				}
			}
		}
	}

	// If no transport was created, create a default one
	if transport == nil {
		transport = &http.Transport{}
	}

	// Apply SSL verification settings
	if config.IgnoreSSL {
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		transport.TLSClientConfig.InsecureSkipVerify = true
	}

	return transport
}

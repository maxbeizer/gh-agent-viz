package capi

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cli/go-gh/v2/pkg/auth"
)

const (
	baseCAPIURL    = "https://api.githubcopilot.com"
	capiHost       = "api.githubcopilot.com"
	integrationID  = "copilot-4-cli"
	apiVersion     = "2026-01-09"
	defaultPageSize = 50
)

// Client communicates with the Copilot API (api.githubcopilot.com).
type Client struct {
	httpClient *http.Client
}

// capiTransport injects Copilot auth headers into every request.
type capiTransport struct {
	base  http.RoundTripper
	token string
}

func (ct *capiTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+ct.token)
	if req.URL.Host == capiHost {
		req.Header.Set("Copilot-Integration-Id", integrationID)
		req.Header.Set("X-GitHub-Api-Version", apiVersion)
	}
	return ct.base.RoundTrip(req)
}

// NewClient creates a CAPI client using the user's gh OAuth token.
func NewClient() (*Client, error) {
	token, err := resolveToken()
	if err != nil {
		return nil, err
	}
	transport := &capiTransport{
		base:  http.DefaultTransport,
		token: token,
	}
	return &Client{
		httpClient: &http.Client{Transport: transport},
	}, nil
}

// resolveToken retrieves the user's gh OAuth token for github.com.
func resolveToken() (string, error) {
	host, _ := auth.DefaultHost()
	if host == "" {
		host = "github.com"
	}
	token, _ := auth.TokenForHost(host)
	if token == "" {
		return "", fmt.Errorf("no auth token found; run 'gh auth login'")
	}
	if !strings.HasPrefix(token, "gho_") {
		return "", fmt.Errorf("copilot API requires an OAuth token (gho_ prefix); re-authenticate with: gh auth login")
	}
	return token, nil
}

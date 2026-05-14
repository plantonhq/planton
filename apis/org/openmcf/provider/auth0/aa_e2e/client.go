package aa_e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// ManagementClient is a lightweight HTTP client for the Auth0 Management API.
// It handles token acquisition via client_credentials and resource lookups
// for E2E verification (exists / not-exists checks).
type ManagementClient struct {
	domain      string
	accessToken string
	httpClient  *http.Client
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// NewManagementClient authenticates via client_credentials and returns a ready client.
func NewManagementClient(domain, clientID, clientSecret string) (*ManagementClient, error) {
	httpClient := &http.Client{Timeout: 30 * time.Second}

	tokenURL := fmt.Sprintf("https://%s/oauth/token", domain)
	payload, _ := json.Marshal(map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"audience":      fmt.Sprintf("https://%s/api/v2/", domain),
		"grant_type":    "client_credentials",
	})

	resp, err := httpClient.Post(tokenURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, errors.Wrap(err, "failed to request access token")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.Errorf("token request returned %d: %s", resp.StatusCode, body)
	}

	var tok tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return nil, errors.Wrap(err, "failed to decode token response")
	}

	return &ManagementClient{
		domain:      domain,
		accessToken: tok.AccessToken,
		httpClient:  httpClient,
	}, nil
}

// ResourceExists checks whether a resource at the given Management API path exists.
// path is relative to /api/v2/ (e.g., "clients/abc123").
// Returns true if 200, false if 404, error for anything else.
func (c *ManagementClient) ResourceExists(path string) (bool, error) {
	url := fmt.Sprintf("https://%s/api/v2/%s", c.domain, path)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, errors.Wrap(err, "failed to build request")
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, errors.Wrapf(err, "GET %s failed", path)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		body, _ := io.ReadAll(resp.Body)
		return false, errors.Errorf("GET %s returned %d: %s", path, resp.StatusCode, body)
	}
}

// VerifyConnectivity does a lightweight check that the token and API are working.
func (c *ManagementClient) VerifyConnectivity() error {
	url := fmt.Sprintf("https://%s/api/v2/clients?per_page=1&fields=client_id", c.domain)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return errors.Wrap(err, "failed to build connectivity check request")
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "connectivity check failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return errors.Errorf("connectivity check returned %d: %s", resp.StatusCode, body)
	}
	return nil
}

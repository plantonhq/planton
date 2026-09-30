package aa_e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// stripeAPIBase is Stripe's REST API host. A path names its API version first: "v1/prices/...",
// "v2/core/event_destinations/...".
const stripeAPIBase = "https://api.stripe.com"

// stripeAPIVersion is the API version the pinned provider (stripe/stripe 0.3.0) sends on every
// call, so the harness reads objects in the shape the module wrote them. It moves with the pin.
const stripeAPIVersion = "2026-05-27.dahlia"

// Client reads Stripe objects over the REST API with the lane's key. It never writes: the
// modules under test create, update and destroy; the harness only observes.
type Client struct {
	apiKey        string
	stripeAccount string
	baseURL       string
	http          *http.Client
}

// NewClient returns a client for the key's account, or for the account stripeAccount names when
// it is set (sent as the Stripe-Context header, as the provider does).
func NewClient(apiKey, stripeAccount string) *Client {
	return &Client{
		apiKey:        apiKey,
		stripeAccount: stripeAccount,
		baseURL:       stripeAPIBase,
		http:          &http.Client{Timeout: 30 * time.Second},
	}
}

// ReadResource GETs one object by its API path (e.g. "v1/webhook_endpoints/we_123") and returns its
// JSON body and whether it exists. A 404 is an honest "does not exist"; any other failure is an
// error carrying Stripe's own message.
func (c *Client) ReadResource(path string) (map[string]interface{}, bool, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/"+path, nil)
	if err != nil {
		return nil, false, errors.Wrapf(err, "building the request for %s", path)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Stripe-Version", stripeAPIVersion)
	if c.stripeAccount != "" {
		req.Header.Set("Stripe-Context", c.stripeAccount)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, false, errors.Wrapf(err, "GET %s", path)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, errors.Wrapf(err, "reading the response for %s", path)
	}

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return nil, false, nil
	case resp.StatusCode != http.StatusOK:
		return nil, false, errors.Errorf("GET %s answered %d: %s", path, resp.StatusCode, stripeErrorMessage(body))
	}
	var object map[string]interface{}
	if err := json.Unmarshal(body, &object); err != nil {
		return nil, false, errors.Wrapf(err, "decoding the response for %s", path)
	}
	return object, true, nil
}

// ResourceExists reports whether the object at path exists.
func (c *Client) ResourceExists(path string) (bool, error) {
	_, exists, err := c.ReadResource(path)
	return exists, err
}

// VerifyConnectivity proves the key authenticates by listing one webhook endpoint -- a read
// every Stripe lane's key can make, whatever else it carries.
func (c *Client) VerifyConnectivity() error {
	if _, _, err := c.ReadResource("v1/webhook_endpoints?limit=1"); err != nil {
		return errors.Wrap(err, "the Stripe key could not read the account")
	}
	return nil
}

// stripeErrorMessage extracts error.message from a Stripe error body, which names what failed
// (a missing permission, an unknown id) in Stripe's words.
func stripeErrorMessage(body []byte) string {
	var envelope struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &envelope) == nil && envelope.Error.Message != "" {
		return envelope.Error.Message
	}
	return fmt.Sprintf("%.200s", body)
}

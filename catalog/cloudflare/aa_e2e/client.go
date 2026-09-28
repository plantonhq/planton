package aa_e2e

// Cloudflare v4 REST client for E2E verification: authenticated GETs with a
// typed exists/absent answer and 429-aware Retry-After backoff.

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/cloudflare/aa_e2e/verify"
)

const (
	apiBaseURL = "https://api.cloudflare.com/client/v4"

	// maxAttempts bounds the 429 retry loop. Cloudflare rate-limits at 1,200
	// requests per 5 minutes per token; verification traffic is far below
	// that, so more than a handful of retries means something else is wrong.
	maxAttempts = 4

	// defaultRetryDelay applies when a 429 response carries no Retry-After
	// header. Doubles per attempt.
	defaultRetryDelay = 2 * time.Second
)

// Client is a lightweight HTTP client for the Cloudflare v4 REST API, scoped
// to what E2E verification needs: authenticated GETs with a typed
// exists/absent answer. It retries 429 responses honoring Retry-After --
// Cloudflare enforces API rate limits, so backoff is part of the client
// contract here, not an optimization.
type Client struct {
	apiToken   string
	accountID  string
	httpClient *http.Client
}

// NewClient returns a client for the given API token and account.
func NewClient(apiToken, accountID string) *Client {
	return &Client{
		apiToken:   apiToken,
		accountID:  accountID,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// AccountID returns the Cloudflare account the harness is scoped to. Account-
// scoped verifiers prepend it to their API paths.
func (c *Client) AccountID() string {
	return c.accountID
}

// cloudflareEnvelope is the standard v4 response wrapper. Only the pieces
// verification reads are modeled.
type cloudflareEnvelope struct {
	Success bool `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

// unknownObjectError reports whether a 400 body carries one of Cloudflare's
// deleted-or-unknown-object error answers -- what several endpoints give
// for a GET on an identifier that no longer exists, instead of a clean 404:
//   - 7003 "could not route to that endpoint" (many collection endpoints)
//   - 1001 "Invalid zone identifier" (zones/{zone_id} after the zone is
//     deleted -- measured live on the zone destroy verification)
//   - 1103 "Location ID is invalid." (gateway/locations/{id} after the
//     location is deleted -- measured live on the DNS location destroy
//     verification)
//   - 1472 "Certificate not found." (accounts/{a}/mtls_certificates/{id}
//     after the certificate is deleted -- measured live 2026-08-28 on the
//     mTLS certificate destroy verification)
//   - code-less "requested snippet not found" (zones/{id}/snippets/{name}
//     after the snippet is deleted -- the snippets endpoint's errors carry
//     a message but NO numeric code, measured live 2026-08-27 on the
//     snippet destroy verification; matched on the exact message because
//     there is no code to match)
func unknownObjectError(body []byte) bool {
	var envelope cloudflareEnvelope
	if json.Unmarshal(body, &envelope) != nil {
		return false
	}
	for _, e := range envelope.Errors {
		if e.Code == 7003 || e.Code == 1001 || e.Code == 1103 || e.Code == 1472 {
			return true
		}
		if e.Code == 0 && e.Message == "requested snippet not found" {
			return true
		}
	}
	return false
}

// ResourceExists reports whether a GET on the given path (relative to
// /client/v4/, e.g. "zones/abc123") finds a resource. 200 means present --
// and so does 202: the origin_tls_client_auth family answers EVERY verb
// with 202 Accepted, including a plain GET on its settings surface
// (measured live 2026-08-28). 404 means absent; a 400 carrying one of the
// unknown-object error codes (see unknownObjectError) also means absent.
// Anything else is a real error.
func (c *Client) ResourceExists(ctx context.Context, path string) (bool, error) {
	resp, body, err := c.get(ctx, path)
	if err != nil {
		return false, err
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusAccepted:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	case http.StatusBadRequest:
		if unknownObjectError(body) {
			return false, nil
		}
		return false, errors.Errorf("GET %s returned 400: %s", path, body)
	default:
		return false, errors.Errorf("GET %s returned %d: %s", path, resp.StatusCode, body)
	}
}

// ResourcePresent reports whether a GET finds a resource that is still
// present after applying opts. 404 and Cloudflare's 400 unknown-object
// answers (see unknownObjectError) are always absent. A 200 (or a 202 --
// the origin_tls_client_auth family answers every verb with 202 Accepted,
// measured live 2026-08-28) is present unless SoftDeleted sees a non-null
// result.deleted_at, IsDeletedFlag sees a non-zero result.is_deleted,
// RevokedAt sees a non-empty result.revoked_at, or result.status matches
// AbsentStatuses. Parses the v4 envelope, so it must never replace
// ResourceExists on raw-body endpoints (e.g. the KV value endpoint).
func (c *Client) ResourcePresent(ctx context.Context, path string, opts verify.EnvelopePresence) (bool, error) {
	resp, body, err := c.get(ctx, path)
	if err != nil {
		return false, err
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusAccepted:
		// Fast path: no absence shape requested, any 2xx is present. EVERY
		// EnvelopePresence field must appear in this guard -- a new probe
		// shape missing here is silently bypassed and its family reports a
		// false "still exists" on every destroy (shipped once with
		// RevokedAt: the lane polled for minutes while the body it never
		// parsed carried revoked_at the whole time).
		if !opts.SoftDeleted && len(opts.AbsentStatuses) == 0 && !opts.EmptyResultArray && !opts.IsDeletedFlag && !opts.RevokedAt {
			return true, nil
		}
		var envelope struct {
			Result json.RawMessage `json:"result"`
		}
		if err := json.Unmarshal(body, &envelope); err != nil {
			return false, errors.Errorf("GET %s returned 200 with an unparseable body: %s", path, body)
		}
		if opts.EmptyResultArray {
			var items []json.RawMessage
			if err := json.Unmarshal(envelope.Result, &items); err == nil && len(items) == 0 {
				return false, nil
			}
		}
		if opts.SoftDeleted || len(opts.AbsentStatuses) > 0 || opts.IsDeletedFlag || opts.RevokedAt {
			var object struct {
				DeletedAt *string  `json:"deleted_at"`
				Status    string   `json:"status"`
				IsDeleted *float64 `json:"is_deleted"`
				RevokedAt *string  `json:"revoked_at"`
			}
			if err := json.Unmarshal(envelope.Result, &object); err != nil {
				return false, errors.Errorf("GET %s returned 200 with an unparseable result object: %s", path, body)
			}
			if opts.SoftDeleted && object.DeletedAt != nil && *object.DeletedAt != "" {
				return false, nil
			}
			// Origin CA's revoke-is-not-delete: the revoked certificate keeps
			// answering GET 200 with revoked_at set (measured live 2026-08-28).
			if opts.RevokedAt && object.RevokedAt != nil && *object.RevokedAt != "" {
				return false, nil
			}
			// The Workflows API's soft-delete variant: is_deleted is a
			// required NUMERIC field in the GET response (not a timestamp),
			// non-zero once the workflow is deleted.
			if opts.IsDeletedFlag && object.IsDeleted != nil && *object.IsDeleted != 0 {
				return false, nil
			}
			for _, status := range opts.AbsentStatuses {
				if object.Status == status {
					return false, nil
				}
			}
		}
		return true, nil
	case http.StatusNotFound:
		return false, nil
	case http.StatusBadRequest:
		if unknownObjectError(body) {
			return false, nil
		}
		return false, errors.Errorf("GET %s returned 400: %s", path, body)
	default:
		return false, errors.Errorf("GET %s returned %d: %s", path, resp.StatusCode, body)
	}
}

// ResourceActive is the deleted_at-aware probe: a 200 whose envelope carries
// a non-null result.deleted_at counts as ABSENT. Kept as a one-line wrapper
// so existing tunnel/route registrations and the API interface do not move.
func (c *Client) ResourceActive(ctx context.Context, path string) (bool, error) {
	return c.ResourcePresent(ctx, path, verify.EnvelopePresence{SoftDeleted: true})
}

// VerifyConnectivity proves the token is valid and active using Cloudflare's
// purpose-built, side-effect-free verify endpoints, then proves the token can
// see the configured account. Failing fast here keeps credential problems from
// surfacing later as confusing mid-lane verification errors.
//
// Cloudflare issues two token forms and each has its OWN verify endpoint that
// rejects the other form: user-owned tokens verify at user/tokens/verify,
// account-owned tokens (the `cfat_`-prefixed kind minted under an account's
// "API Tokens" page) verify ONLY at accounts/{account_id}/tokens/verify --
// the user endpoint answers 401 "Invalid API Token" for them (measured
// live 2026-08-25). Both engines accept either form as a plain bearer
// credential, so the harness must too: try the user endpoint first, fall
// back to the account endpoint, and fail only when both reject.
//
// A verified account-owned token is still NOT universal: a small class of
// endpoints demands a USER-ACTOR credential and refuses account-owned tokens
// on every verb with 401 code 1039 "malformed actor email claim" (first
// measured live 2026-08-27 on zones/{zone}/devices/policy/certificates --
// GET and PATCH alike, on active and pending zones). An upstream acceptance
// test that blanks CLOUDFLARE_API_TOKEN is the tell for this class; a lane
// arm that hits 1039 is credential-blocked, not broken, and its honest
// outcome is an offline plan arm whose unblock is a user-actor credential.
func (c *Client) VerifyConnectivity(ctx context.Context) error {
	resp, userBody, err := c.get(ctx, "user/tokens/verify")
	if err != nil {
		return errors.Wrap(err, "token verification request failed")
	}
	if resp.StatusCode != http.StatusOK {
		resp, acctBody, acctErr := c.get(ctx, "accounts/"+c.accountID+"/tokens/verify")
		if acctErr != nil {
			return errors.Wrap(acctErr, "account-token verification request failed")
		}
		if resp.StatusCode != http.StatusOK {
			return errors.Errorf(
				"token verification failed on both endpoints: user/tokens/verify returned %s; accounts/%s/tokens/verify returned %d: %s",
				userBody, c.accountID, resp.StatusCode, acctBody)
		}
	}

	resp, body, err := c.get(ctx, "accounts/"+c.accountID)
	if err != nil {
		return errors.Wrap(err, "account lookup request failed")
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	// A token scoped to zones (DNS edit on one zone, the least a lane that
	// only writes records needs) cannot read the account itself, but lists
	// the zones it reaches. Seeing one of this account's zones is what a
	// zone-scoped lane requires; an account-level kind still fails at its own
	// first call, with the provider's own refusal.
	zresp, zbody, zerr := c.get(ctx, "zones?per_page=1&account.id="+c.accountID)
	if zerr != nil {
		return errors.Wrap(zerr, "zone lookup request failed")
	}
	if zresp.StatusCode == http.StatusOK && zoneListHasResult(zbody) {
		return nil
	}
	return errors.Errorf("account %s lookup returned %d and the token reaches none of its zones (is the token scoped to this account?): %s",
		c.accountID, resp.StatusCode, body)
}

// zoneListHasResult reports whether a zones-list response names at least one zone.
func zoneListHasResult(body []byte) bool {
	var parsed struct {
		Result []json.RawMessage `json:"result"`
	}
	return json.Unmarshal(body, &parsed) == nil && len(parsed.Result) > 0
}

// get performs one authenticated GET with 429-aware retries and returns the
// final response alongside its fully-read body (the response body is already
// closed).
func (c *Client) get(ctx context.Context, path string) (*http.Response, []byte, error) {
	url := fmt.Sprintf("%s/%s", apiBaseURL, path)

	for attempt := 1; ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to build request for %s", path)
		}
		req.Header.Set("Authorization", "Bearer "+c.apiToken)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "GET %s failed", path)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusTooManyRequests {
			return resp, body, nil
		}
		if attempt >= maxAttempts {
			return nil, nil, errors.Errorf("GET %s rate-limited after %d attempts", path, attempt)
		}

		delay := defaultRetryDelay * time.Duration(1<<(attempt-1))
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if seconds, err := strconv.Atoi(ra); err == nil && seconds > 0 {
				delay = time.Duration(seconds) * time.Second
			}
		}
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		}
	}
}

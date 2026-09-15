package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenBAOInitResult holds the secrets produced by a successful initialization:
// the root token and, depending on the seal, either the unseal keys (built-in
// seal) or the recovery keys (a cloud seal). Exactly one of the two slices is
// populated; the caller must never store one under the other's name.
type OpenBAOInitResult struct {
	UnsealKeys   []string
	RecoveryKeys []string
	RootToken    string
}

// OpenBAOInitOptions is the shape of a /sys/init request, and the shape is
// decided by the seal: a built-in-seal server takes secret_shares and
// secret_threshold and returns unseal keys; a server under a cloud seal
// REFUSES those two parameters ("not applicable to seal type"), forces its
// own barrier to one share, and takes recovery_shares and recovery_threshold
// instead, returning recovery keys. The operator asks for one shape or the
// other; a request carrying both is rejected by every server.
type OpenBAOInitOptions struct {
	// SecretShares/SecretThreshold: the built-in seal's key shares.
	SecretShares    int
	SecretThreshold int
	// RecoveryShares/RecoveryThreshold: a cloud seal's recovery quorum.
	RecoveryShares    int
	RecoveryThreshold int
}

// ShamirInit is the built-in seal's request shape.
func ShamirInit(shares, threshold int) OpenBAOInitOptions {
	return OpenBAOInitOptions{SecretShares: shares, SecretThreshold: threshold}
}

// RecoveryInit is a cloud seal's request shape.
func RecoveryInit(shares, threshold int) OpenBAOInitOptions {
	return OpenBAOInitOptions{RecoveryShares: shares, RecoveryThreshold: threshold}
}

func (o OpenBAOInitOptions) body() map[string]int {
	body := map[string]int{}
	if o.SecretShares != 0 {
		body["secret_shares"] = o.SecretShares
		body["secret_threshold"] = o.SecretThreshold
	}
	if o.RecoveryShares != 0 {
		body["recovery_shares"] = o.RecoveryShares
		body["recovery_threshold"] = o.RecoveryThreshold
	}
	return body
}

// OpenBAOHealthStatus represents the state of an OpenBAO instance.
type OpenBAOHealthStatus struct {
	Initialized bool
	Sealed      bool
}

// CheckOpenBAOHealth queries /v1/sys/health to determine initialization and
// seal status. The OpenBAO health endpoint returns different HTTP status codes:
//   - 200: initialized, unsealed, active
//   - 429: unsealed, standby
//   - 472: data recovery mode
//   - 501: not initialized
//   - 503: sealed
//
// All status codes return a JSON body with "initialized" and "sealed" fields.
func CheckOpenBAOHealth(ctx context.Context, client *http.Client, apiAddr string) (*OpenBAOHealthStatus, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiAddr+"/v1/sys/health", nil)
	if err != nil {
		return nil, fmt.Errorf("building health request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("checking health: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Initialized bool `json:"initialized"`
		Sealed      bool `json:"sealed"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding health response: %w", err)
	}

	return &OpenBAOHealthStatus{
		Initialized: result.Initialized,
		Sealed:      result.Sealed,
	}, nil
}

// InitializeOpenBAO calls /v1/sys/init to initialize a fresh OpenBAO instance
// in the shape the seal requires (OpenBAOInitOptions). Returns the root token
// and the keys the server produced -- unseal keys or recovery keys. This is a
// one-time operation; calling it on an already initialized instance returns
// an error. Under a cloud seal the server re-seals itself after init and its
// own start-up loop opens it within seconds (WaitUntilUnsealed); the caller
// never calls /sys/unseal there.
func InitializeOpenBAO(ctx context.Context, client *http.Client, apiAddr string, opts OpenBAOInitOptions) (*OpenBAOInitResult, error) {
	body, err := json.Marshal(opts.body())
	if err != nil {
		return nil, fmt.Errorf("marshaling init request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, apiAddr+"/v1/sys/init", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("building init request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("initializing OpenBAO: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("init returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Keys         []string `json:"keys"`
		RecoveryKeys []string `json:"recovery_keys"`
		RootToken    string   `json:"root_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding init response: %w", err)
	}

	return &OpenBAOInitResult{
		UnsealKeys:   result.Keys,
		RecoveryKeys: result.RecoveryKeys,
		RootToken:    result.RootToken,
	}, nil
}

// WaitUntilUnsealed polls /v1/sys/health until the server reports itself
// initialized and unsealed, or the timeout passes. Under a cloud seal the
// server opens itself: after init, and on every start, its own loop fetches
// the stored key through the seal and retries every five seconds until it
// succeeds. The caller sizes the timeout to that interval; a vault still
// sealed afterwards is reported, not retried here.
func WaitUntilUnsealed(ctx context.Context, client *http.Client, apiAddr string, timeout time.Duration) (bool, error) {
	deadline := time.Now().Add(timeout)
	for {
		health, err := CheckOpenBAOHealth(ctx, client, apiAddr)
		if err != nil {
			return false, err
		}
		if health.Initialized && !health.Sealed {
			return true, nil
		}
		if time.Now().After(deadline) {
			return false, nil
		}
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
}

// EnsureOpenBAOMounts idempotently enables the two secrets engines the
// platform expects: KV v2 at "secret/" and Transit at "transit/" (the mounts
// PlatformVaultClient hardcodes). A fresh server-mode OpenBAO starts with NO
// secrets engines -- dev mode auto-mounts a KV engine, the Helm chart runs
// server mode, and Transit is never auto-mounted in any mode -- so without
// this every platform vault call fails with "no handler for route".
// Enabling engines is infrastructure provisioning and therefore the
// operator's job, never the control plane's.
func EnsureOpenBAOMounts(ctx context.Context, client *http.Client, apiAddr, rootToken string) error {
	mounts, err := listMounts(ctx, client, apiAddr, rootToken)
	if err != nil {
		return fmt.Errorf("listing OpenBAO mounts: %w", err)
	}

	if _, ok := mounts["secret/"]; !ok {
		if err := enableMount(ctx, client, apiAddr, rootToken, "secret", map[string]any{
			"type":    "kv",
			"options": map[string]string{"version": "2"},
		}); err != nil {
			return fmt.Errorf("enabling KV v2 engine at secret/: %w", err)
		}
	}

	if _, ok := mounts["transit/"]; !ok {
		if err := enableMount(ctx, client, apiAddr, rootToken, "transit", map[string]any{
			"type": "transit",
		}); err != nil {
			return fmt.Errorf("enabling Transit engine at transit/: %w", err)
		}
	}

	return nil
}

func listMounts(ctx context.Context, client *http.Client, apiAddr, rootToken string) (map[string]json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiAddr+"/v1/sys/mounts", nil)
	if err != nil {
		return nil, fmt.Errorf("building mounts request: %w", err)
	}
	req.Header.Set("X-Vault-Token", rootToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("listing mounts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sys/mounts returned %d: %s", resp.StatusCode, string(respBody))
	}

	// Mount names arrive as top-level keys ("secret/", "transit/", "sys/", ...);
	// newer API versions nest them under "data" as well -- reading the top level
	// works for both because the legacy shape is preserved.
	var result map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding mounts response: %w", err)
	}
	return result, nil
}

func enableMount(ctx context.Context, client *http.Client, apiAddr, rootToken, path string, config map[string]any) error {
	body, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshaling mount config: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiAddr+"/v1/sys/mounts/"+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("building mount request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Vault-Token", rootToken)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("enabling mount %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("sys/mounts/%s returned %d: %s", path, resp.StatusCode, string(respBody))
	}

	return nil
}

// UnsealOpenBAO calls /v1/sys/unseal with each key until the threshold is met
// and the instance is unsealed. Returns nil on success.
func UnsealOpenBAO(ctx context.Context, client *http.Client, apiAddr string, keys []string) error {
	for i, key := range keys {
		body, err := json.Marshal(map[string]string{"key": key})
		if err != nil {
			return fmt.Errorf("marshaling unseal request %d: %w", i, err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, apiAddr+"/v1/sys/unseal", bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("building unseal request %d: %w", i, err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("unsealing (key %d): %w", i, err)
		}

		var result struct {
			Sealed bool `json:"sealed"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return fmt.Errorf("decoding unseal response %d: %w", i, err)
		}
		resp.Body.Close()

		if !result.Sealed {
			return nil
		}
	}

	return fmt.Errorf("still sealed after applying all %d keys", len(keys))
}

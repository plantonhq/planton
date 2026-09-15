package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// The vault's access surface, as the operator drives it over the raw API:
// the Kubernetes auth method the operator signs in through, the policies
// that bound what each caller may do, the token role the control plane's
// token is minted through, and the token operations that keep that token
// alive. Everything here is a plain request and a plain answer; what the
// operator DOES with these -- which policy says what, when a role is
// rewritten, when a token is re-minted -- is the component's, in
// component/openbao_login.go and component/openbao_token.go.
//
// Two facts from the server shape this file. Enabling an auth method
// (sys/auth/<path>) needs the sudo capability, so it is done with the root
// token at initialization and never by the operator's own session. And a
// batch token cannot create tokens, so the operator's session -- the thing
// that mints the control plane's token -- is a service token, created at
// sign-in and revoked when the pass ends.

// APIError is a non-2xx answer from the vault with the server's own words.
type APIError struct {
	Status int
	Errors []string
}

func (e *APIError) Error() string {
	if len(e.Errors) == 0 {
		return fmt.Sprintf("the vault answered %d", e.Status)
	}
	return fmt.Sprintf("the vault answered %d: %s", e.Status, strings.Join(e.Errors, "; "))
}

// IsPermissionDenied reports a 403: the token presented lacks the capability
// the path needs (or, on a login, the identity is not bound to the role).
func (e *APIError) IsPermissionDenied() bool { return e.Status == http.StatusForbidden }

// IsNotMounted reports the answer a path under an auth method or engine
// that is not enabled gets: a 404 whose error names the missing handler.
func (e *APIError) IsNotMounted() bool {
	if e.Status != http.StatusNotFound {
		return false
	}
	for _, msg := range e.Errors {
		if strings.Contains(msg, "no handler for route") || strings.Contains(msg, "unsupported path") {
			return true
		}
	}
	return false
}

// AsAPIError unwraps the vault's answer from an error, if that is what it is.
func AsAPIError(err error) (*APIError, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}

// vaultCall is one request: method, path under /v1, an optional token, an
// optional JSON body. It returns the decoded envelope (nil on 204) or an
// APIError carrying the server's errors array.
func vaultCall(ctx context.Context, client *http.Client, apiAddr, method, path, token string, body any) (map[string]any, error) {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling %s %s: %w", method, path, err)
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, apiAddr+"/v1/"+path, reader)
	if err != nil {
		return nil, fmt.Errorf("building %s %s: %w", method, path, err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("X-Vault-Token", token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{Status: resp.StatusCode}
		var envelope struct {
			Errors []string `json:"errors"`
		}
		if json.Unmarshal(raw, &envelope) == nil {
			apiErr.Errors = envelope.Errors
		}
		return nil, fmt.Errorf("%s %s: %w", method, path, apiErr)
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, nil
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decoding %s %s: %w", method, path, err)
	}
	return out, nil
}

// dataOf returns the envelope's "data" object, or nil.
func dataOf(envelope map[string]any) map[string]any {
	if envelope == nil {
		return nil
	}
	data, _ := envelope["data"].(map[string]any)
	return data
}

// notFound reports a 404 that is NOT a missing mount: the path exists and
// holds nothing (a config never written, a role or policy that does not
// exist). A missing mount is surfaced as the error it is.
func notFound(err error) (bool, error) {
	apiErr, ok := AsAPIError(err)
	if !ok || apiErr.Status != http.StatusNotFound || apiErr.IsNotMounted() {
		return false, err
	}
	return true, nil
}

// ── the Kubernetes auth method ─────────────────────────────────────────────

// EnableKubernetesAuth mounts the kubernetes auth method at the given path.
// Needs sudo (the root token at initialization). Enabling a path that is
// already in use is not an error: the operator asks for the state, not the
// act.
func EnableKubernetesAuth(ctx context.Context, client *http.Client, apiAddr, token, path string) error {
	_, err := vaultCall(ctx, client, apiAddr, http.MethodPost, "sys/auth/"+path, token, map[string]any{"type": "kubernetes"})
	if err != nil {
		if apiErr, ok := AsAPIError(err); ok && apiErr.Status == http.StatusBadRequest {
			for _, msg := range apiErr.Errors {
				if strings.Contains(msg, "already in use") {
					return nil
				}
			}
		}
		return fmt.Errorf("enabling the kubernetes auth method at auth/%s: %w", path, err)
	}
	return nil
}

// KubernetesAuthConfig is the auth method's configuration. In production
// only Host is set: the method reads the pod's own CA certificate and
// ServiceAccount token for its TokenReview calls, which is what lets a
// vault restored onto a different cluster keep a working auth method --
// nothing here ever names the old cluster. The other three exist for the
// real-server proof, where the vault runs outside a cluster and reviews
// tokens against a stand-in API: the method reads the pod's CA file the
// moment the config is written unless DisableLocalCAJWT is set, and then
// insists on a CACert of its own.
type KubernetesAuthConfig struct {
	Host              string
	CACert            string
	TokenReviewerJWT  string
	DisableLocalCAJWT bool
}

// ReadKubernetesAuthConfig reads the method's config at the given path.
// found is false when the method is enabled but never configured; a method
// that is not mounted is an error (IsNotMounted).
func ReadKubernetesAuthConfig(ctx context.Context, client *http.Client, apiAddr, token, path string) (*KubernetesAuthConfig, bool, error) {
	envelope, err := vaultCall(ctx, client, apiAddr, http.MethodGet, "auth/"+path+"/config", token, nil)
	if err != nil {
		if missing, nerr := notFound(err); missing {
			return nil, false, nil
		} else if nerr != nil {
			return nil, false, fmt.Errorf("reading the kubernetes auth config: %w", nerr)
		}
	}
	data := dataOf(envelope)
	if data == nil {
		return nil, false, nil
	}
	cfg := &KubernetesAuthConfig{
		Host:   stringOf(data["kubernetes_host"]),
		CACert: stringOf(data["kubernetes_ca_cert"]),
	}
	return cfg, true, nil
}

// WriteKubernetesAuthConfig writes the method's config. Keys left empty are
// not sent, so the server keeps its defaults (the local CA and token).
func WriteKubernetesAuthConfig(ctx context.Context, client *http.Client, apiAddr, token, path string, cfg KubernetesAuthConfig) error {
	body := map[string]any{"kubernetes_host": cfg.Host}
	if cfg.CACert != "" {
		body["kubernetes_ca_cert"] = cfg.CACert
	}
	if cfg.TokenReviewerJWT != "" {
		body["token_reviewer_jwt"] = cfg.TokenReviewerJWT
	}
	if cfg.DisableLocalCAJWT {
		body["disable_local_ca_jwt"] = true
	}
	if _, err := vaultCall(ctx, client, apiAddr, http.MethodPost, "auth/"+path+"/config", token, body); err != nil {
		return fmt.Errorf("writing the kubernetes auth config: %w", err)
	}
	return nil
}

// KubernetesAuthRole binds ServiceAccounts to policies: a login presenting a
// token for one of the named accounts in one of the named namespaces gets a
// token carrying Policies, living TTL. No audience is pinned: the mounted
// ServiceAccount token's audience is the cluster's own, and pinning it would
// bind the role to one cluster -- the one thing a restore must not depend on.
type KubernetesAuthRole struct {
	ServiceAccountNames      []string
	ServiceAccountNamespaces []string
	Policies                 []string
	TTL                      time.Duration
	MaxTTL                   time.Duration
}

// ReadKubernetesAuthRole reads a role; found is false when it does not exist.
func ReadKubernetesAuthRole(ctx context.Context, client *http.Client, apiAddr, token, path, name string) (*KubernetesAuthRole, bool, error) {
	envelope, err := vaultCall(ctx, client, apiAddr, http.MethodGet, "auth/"+path+"/role/"+name, token, nil)
	if err != nil {
		if missing, nerr := notFound(err); missing {
			return nil, false, nil
		} else if nerr != nil {
			return nil, false, fmt.Errorf("reading the kubernetes auth role %s: %w", name, nerr)
		}
	}
	data := dataOf(envelope)
	if data == nil {
		return nil, false, nil
	}
	return &KubernetesAuthRole{
		ServiceAccountNames:      stringsOf(data["bound_service_account_names"]),
		ServiceAccountNamespaces: stringsOf(data["bound_service_account_namespaces"]),
		Policies:                 stringsOf(data["token_policies"]),
		TTL:                      secondsOf(data["token_ttl"]),
		MaxTTL:                   secondsOf(data["token_max_ttl"]),
	}, true, nil
}

// WriteKubernetesAuthRole creates or replaces a role.
func WriteKubernetesAuthRole(ctx context.Context, client *http.Client, apiAddr, token, path, name string, role KubernetesAuthRole) error {
	body := map[string]any{
		"bound_service_account_names":      role.ServiceAccountNames,
		"bound_service_account_namespaces": role.ServiceAccountNamespaces,
		"token_policies":                   role.Policies,
		"token_ttl":                        durationString(role.TTL),
		"token_max_ttl":                    durationString(role.MaxTTL),
	}
	if _, err := vaultCall(ctx, client, apiAddr, http.MethodPost, "auth/"+path+"/role/"+name, token, body); err != nil {
		return fmt.Errorf("writing the kubernetes auth role %s: %w", name, err)
	}
	return nil
}

// Token is what a login or a token creation hands back: the token itself
// and its accessor, the public handle the token is looked up, renewed, and
// revoked by without ever presenting it.
type Token struct {
	ClientToken string
	Accessor    string
}

// KubernetesLogin signs in through the method with a ServiceAccount token
// and a role. A refused login is an APIError with the server's reason (the
// account is not bound to the role; the TokenReview failed).
func KubernetesLogin(ctx context.Context, client *http.Client, apiAddr, path, role, jwt string) (*Token, error) {
	envelope, err := vaultCall(ctx, client, apiAddr, http.MethodPost, "auth/"+path+"/login", "", map[string]any{"role": role, "jwt": jwt})
	if err != nil {
		return nil, fmt.Errorf("signing in through auth/%s as role %s: %w", path, role, err)
	}
	return tokenOf(envelope, "the kubernetes login")
}

// ── policies ───────────────────────────────────────────────────────────────

// ReadPolicy returns an ACL policy's text; found is false when it does not
// exist.
func ReadPolicy(ctx context.Context, client *http.Client, apiAddr, token, name string) (string, bool, error) {
	envelope, err := vaultCall(ctx, client, apiAddr, http.MethodGet, "sys/policies/acl/"+name, token, nil)
	if err != nil {
		if missing, nerr := notFound(err); missing {
			return "", false, nil
		} else if nerr != nil {
			return "", false, fmt.Errorf("reading policy %s: %w", name, nerr)
		}
	}
	data := dataOf(envelope)
	if data == nil {
		return "", false, nil
	}
	return stringOf(data["policy"]), true, nil
}

// WritePolicy creates or replaces an ACL policy.
func WritePolicy(ctx context.Context, client *http.Client, apiAddr, token, name, policy string) error {
	if _, err := vaultCall(ctx, client, apiAddr, http.MethodPut, "sys/policies/acl/"+name, token, map[string]any{"policy": policy}); err != nil {
		return fmt.Errorf("writing policy %s: %w", name, err)
	}
	return nil
}

// ── token roles and tokens ─────────────────────────────────────────────────

// TokenRole is the shape tokens minted through auth/token/create/<role> take.
// Orphan and Period come from the role, which is what lets a token WITHOUT
// sudo mint an orphan, periodic token: asked for directly, both need sudo;
// set on the role, they do not.
type TokenRole struct {
	AllowedPolicies []string
	Orphan          bool
	Renewable       bool
	Period          time.Duration
	TokenType       string
}

// ReadTokenRole reads a token role; found is false when it does not exist.
func ReadTokenRole(ctx context.Context, client *http.Client, apiAddr, token, name string) (*TokenRole, bool, error) {
	envelope, err := vaultCall(ctx, client, apiAddr, http.MethodGet, "auth/token/roles/"+name, token, nil)
	if err != nil {
		if missing, nerr := notFound(err); missing {
			return nil, false, nil
		} else if nerr != nil {
			return nil, false, fmt.Errorf("reading token role %s: %w", name, nerr)
		}
	}
	data := dataOf(envelope)
	if data == nil {
		return nil, false, nil
	}
	orphan, _ := data["orphan"].(bool)
	renewable, _ := data["renewable"].(bool)
	return &TokenRole{
		AllowedPolicies: stringsOf(data["allowed_policies"]),
		Orphan:          orphan,
		Renewable:       renewable,
		Period:          secondsOf(data["token_period"]),
		TokenType:       stringOf(data["token_type"]),
	}, true, nil
}

// WriteTokenRole creates or replaces a token role.
func WriteTokenRole(ctx context.Context, client *http.Client, apiAddr, token, name string, role TokenRole) error {
	body := map[string]any{
		"allowed_policies": role.AllowedPolicies,
		"orphan":           role.Orphan,
		"renewable":        role.Renewable,
		"token_period":     durationString(role.Period),
		"token_type":       role.TokenType,
	}
	if _, err := vaultCall(ctx, client, apiAddr, http.MethodPost, "auth/token/roles/"+name, token, body); err != nil {
		return fmt.Errorf("writing token role %s: %w", name, err)
	}
	return nil
}

// CreateTokenOptions names the token being minted so its accessor reads in
// the vault's own records as what it is.
type CreateTokenOptions struct {
	Policies    []string
	DisplayName string
	Meta        map[string]string
}

// CreateTokenWithRole mints a token through a token role; the role's orphan,
// period, and allowed policies apply.
func CreateTokenWithRole(ctx context.Context, client *http.Client, apiAddr, token, role string, opts CreateTokenOptions) (*Token, error) {
	body := map[string]any{"policies": opts.Policies}
	if opts.DisplayName != "" {
		body["display_name"] = opts.DisplayName
	}
	if len(opts.Meta) > 0 {
		body["meta"] = opts.Meta
	}
	envelope, err := vaultCall(ctx, client, apiAddr, http.MethodPost, "auth/token/create/"+role, token, body)
	if err != nil {
		return nil, fmt.Errorf("minting a token through role %s: %w", role, err)
	}
	return tokenOf(envelope, "the token role "+role)
}

// TokenInfo is what a lookup by accessor tells about a token that exists.
type TokenInfo struct {
	TTL       time.Duration
	Renewable bool
	Policies  []string
}

// LookupAccessor reads a token by its accessor. found is false when the
// accessor names no live token -- revoked, expired, or never this vault's --
// which is the one answer the caller acts on by minting anew.
func LookupAccessor(ctx context.Context, client *http.Client, apiAddr, token, accessor string) (*TokenInfo, bool, error) {
	envelope, err := vaultCall(ctx, client, apiAddr, http.MethodPost, "auth/token/lookup-accessor", token, map[string]any{"accessor": accessor})
	if err != nil {
		if apiErr, ok := AsAPIError(err); ok && (apiErr.Status == http.StatusBadRequest || apiErr.Status == http.StatusForbidden) && !apiErr.IsNotMounted() {
			for _, msg := range apiErr.Errors {
				if strings.Contains(msg, "invalid accessor") {
					return nil, false, nil
				}
			}
		}
		return nil, false, fmt.Errorf("looking up a token by accessor: %w", err)
	}
	data := dataOf(envelope)
	if data == nil {
		return nil, false, nil
	}
	renewable, _ := data["renewable"].(bool)
	return &TokenInfo{
		TTL:       secondsOf(data["ttl"]),
		Renewable: renewable,
		Policies:  stringsOf(data["policies"]),
	}, true, nil
}

// RenewAccessor renews a token by its accessor. A periodic token renewed
// within its period never expires and keeps the same string, so a consumer
// reading it from a Secret needs nothing done.
func RenewAccessor(ctx context.Context, client *http.Client, apiAddr, token, accessor string) error {
	if _, err := vaultCall(ctx, client, apiAddr, http.MethodPost, "auth/token/renew-accessor", token, map[string]any{"accessor": accessor}); err != nil {
		return fmt.Errorf("renewing a token by accessor: %w", err)
	}
	return nil
}

// RevokeAccessor revokes a token by its accessor. Revoking one that is
// already gone is not an error.
func RevokeAccessor(ctx context.Context, client *http.Client, apiAddr, token, accessor string) error {
	if _, err := vaultCall(ctx, client, apiAddr, http.MethodPost, "auth/token/revoke-accessor", token, map[string]any{"accessor": accessor}); err != nil {
		if apiErr, ok := AsAPIError(err); ok && apiErr.Status == http.StatusBadRequest {
			for _, msg := range apiErr.Errors {
				if strings.Contains(msg, "invalid accessor") {
					return nil
				}
			}
		}
		return fmt.Errorf("revoking a token by accessor: %w", err)
	}
	return nil
}

// RevokeSelf ends the session the token is.
func RevokeSelf(ctx context.Context, client *http.Client, apiAddr, token string) error {
	if _, err := vaultCall(ctx, client, apiAddr, http.MethodPost, "auth/token/revoke-self", token, nil); err != nil {
		return fmt.Errorf("revoking the session token: %w", err)
	}
	return nil
}

// ── decoding ───────────────────────────────────────────────────────────────

func tokenOf(envelope map[string]any, what string) (*Token, error) {
	auth, _ := envelope["auth"].(map[string]any)
	tok := &Token{ClientToken: stringOf(auth["client_token"]), Accessor: stringOf(auth["accessor"])}
	if tok.ClientToken == "" || tok.Accessor == "" {
		return nil, fmt.Errorf("%s returned no token", what)
	}
	return tok, nil
}

func stringOf(v any) string {
	s, _ := v.(string)
	return s
}

func stringsOf(v any) []string {
	items, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// secondsOf reads a duration the server reports in seconds (JSON numbers
// decode as float64).
func secondsOf(v any) time.Duration {
	switch n := v.(type) {
	case float64:
		return time.Duration(n) * time.Second
	case json.Number:
		if i, err := n.Int64(); err == nil {
			return time.Duration(i) * time.Second
		}
	}
	return 0
}

// durationString is the form the server accepts on write: whole seconds
// with the unit, so "600s" and "168h" both round-trip to the seconds the
// server reports back.
func durationString(d time.Duration) string {
	return fmt.Sprintf("%ds", int64(d/time.Second))
}

// Package keycloak converges the operator-owned subset of a live Keycloak
// realm through the Admin API, and refuses by construction to touch anything
// else.
//
// The realm import (internal/resources) remains the first-boot bootstrap;
// this package is the steady state that makes the import's frozen-in-time
// nature irrelevant: missing realm furniture is created, drifted owned fields
// are corrected, and admin-created state is never diffed, never written,
// never reported as drift. Both derive from the same shared constants in
// internal/resources, so what the import bakes and what this package
// converges cannot disagree.
//
// This is a deliberately thin, hand-rolled client -- no gocloak, no SDK. A
// full client library's convenience is full-representation round-trips
// through TYPED structs, and Go's encoding/json silently DROPS any field a
// struct does not declare -- so a typed GET-mutate-PUT would erase every
// admin-set field it didn't model. That is exactly the clobber hazard the
// never-clobber contract forbids. Representations here are therefore raw
// JSON maps end to end: a GET returns everything the server sent, owned
// fields are mutated in place, and a PUT carries every unowned field back
// byte-equivalent.
package keycloak

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Representation is a Keycloak Admin API object as the server sent it: a raw
// JSON map. See the package comment for why this is a map and not a struct.
type Representation map[string]any

// AdminClient talks to exactly the admin surfaces the realm reconciler
// consumes -- the token endpoint, realm settings, clients and their protocol
// mappers, and service-account role mappings. Every write's field set is
// visible in this package's own code and reviewable; there is no generic
// "update object" convenience to misuse.
type AdminClient struct {
	// httpClient is injected by the caller (the operator's thin-client
	// idiom: see internal/bootstrap).
	httpClient *http.Client

	// serverRoot is the identity server's base URL including the serving
	// path prefix (e.g. http://{crName}-identity.{ns}.svc.cluster.local/idp)
	// -- the split-horizon in-cluster address, never the public front door.
	serverRoot string

	token string
}

// NewAdminClient returns an unauthenticated client; call Authenticate before
// any admin call.
func NewAdminClient(httpClient *http.Client, serverRoot string) *AdminClient {
	return &AdminClient{httpClient: httpClient, serverRoot: strings.TrimSuffix(serverRoot, "/")}
}

// Authenticate obtains an admin token via the password grant on the MASTER
// realm as the bootstrap admin -- the credential the operator itself
// provisions at first boot (KC_BOOTSTRAP_ADMIN_*). The token is fetched once
// per reconcile pass and never cached across passes: a pass lasts seconds,
// and a cache would add an expiry surface for zero gain.
func (c *AdminClient) Authenticate(ctx context.Context, username, password string) error {
	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("client_id", "admin-cli")
	form.Set("username", username)
	form.Set("password", password)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.serverRoot+"/realms/master/protocol/openid-connect/token",
		strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("building token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("requesting admin token: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &TokenRefusedError{Status: resp.StatusCode, Body: string(body)}
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("decoding admin token response: %w", err)
	}
	if tokenResp.AccessToken == "" {
		return fmt.Errorf("admin token response carried no access_token")
	}
	c.token = tokenResp.AccessToken
	return nil
}

// TokenRefusedError is the token endpoint answering anything but a token:
// typed so a caller can tell a refused credential (401, the restored-realm
// case the identity component recovers from) apart from a server that is
// not answering at all.
type TokenRefusedError struct {
	Status int
	Body   string
}

func (e *TokenRefusedError) Error() string {
	return fmt.Sprintf("admin token request returned %d: %s", e.Status, e.Body)
}

// IsCredentialRefused reports whether err is the token endpoint refusing the
// credential itself (HTTP 401) -- as opposed to the server being down,
// unreachable, or broken, which the reconcile cadence retries.
func IsCredentialRefused(err error) bool {
	var refused *TokenRefusedError
	return errors.As(err, &refused) && refused.Status == http.StatusUnauthorized
}

// GetRealm returns the realm's full representation.
func (c *AdminClient) GetRealm(ctx context.Context, realm string) (Representation, error) {
	var rep Representation
	if err := c.do(ctx, http.MethodGet, c.adminPath(realm, ""), nil, http.StatusOK, &rep); err != nil {
		return nil, fmt.Errorf("getting realm %s: %w", realm, err)
	}
	return rep, nil
}

// UpdateRealm PUTs the realm representation back. Callers pass the SAME map
// GetRealm returned, mutated only on owned fields, so unowned fields ride
// through untouched.
func (c *AdminClient) UpdateRealm(ctx context.Context, realm string, rep Representation) error {
	if err := c.do(ctx, http.MethodPut, c.adminPath(realm, ""), rep, http.StatusNoContent, nil); err != nil {
		return fmt.Errorf("updating realm %s: %w", realm, err)
	}
	return nil
}

// FindClientByClientID looks a client up by its clientId (the human name,
// not the server UUID). Returns found=false when absent.
func (c *AdminClient) FindClientByClientID(ctx context.Context, realm, clientID string) (Representation, bool, error) {
	var reps []Representation
	path := c.adminPath(realm, "/clients?clientId="+url.QueryEscape(clientID))
	if err := c.do(ctx, http.MethodGet, path, nil, http.StatusOK, &reps); err != nil {
		return nil, false, fmt.Errorf("finding client %s: %w", clientID, err)
	}
	// The query is an exact-match filter server-side, but verify: some
	// Keycloak versions have matched by substring here.
	for _, rep := range reps {
		if id, _ := rep["clientId"].(string); id == clientID {
			return rep, true, nil
		}
	}
	return nil, false, nil
}

// CreateClient POSTs a new client. On create (unlike update) Keycloak honors
// the embedded protocolMappers and client-scope lists, so the desired
// representation goes in whole.
func (c *AdminClient) CreateClient(ctx context.Context, realm string, rep Representation) error {
	clientID, _ := rep["clientId"].(string)
	if err := c.do(ctx, http.MethodPost, c.adminPath(realm, "/clients"), rep, http.StatusCreated, nil); err != nil {
		return fmt.Errorf("creating client %s: %w", clientID, err)
	}
	return nil
}

// UpdateClient PUTs a client representation back by its server UUID. Same
// round-trip contract as UpdateRealm. NOTE: Keycloak IGNORES protocolMappers
// and client-scope lists on update -- mappers converge through the dedicated
// endpoints below.
func (c *AdminClient) UpdateClient(ctx context.Context, realm, id string, rep Representation) error {
	clientID, _ := rep["clientId"].(string)
	if err := c.do(ctx, http.MethodPut, c.adminPath(realm, "/clients/"+id), rep, http.StatusNoContent, nil); err != nil {
		return fmt.Errorf("updating client %s: %w", clientID, err)
	}
	return nil
}

// GetClientSecret returns a confidential client's current secret.
func (c *AdminClient) GetClientSecret(ctx context.Context, realm, id string) (string, error) {
	var rep struct {
		Value string `json:"value"`
	}
	if err := c.do(ctx, http.MethodGet, c.adminPath(realm, "/clients/"+id+"/client-secret"), nil, http.StatusOK, &rep); err != nil {
		return "", fmt.Errorf("getting client secret: %w", err)
	}
	return rep.Value, nil
}

// ListProtocolMappers returns a client's protocol mappers.
func (c *AdminClient) ListProtocolMappers(ctx context.Context, realm, clientUUID string) ([]Representation, error) {
	var reps []Representation
	path := c.adminPath(realm, "/clients/"+clientUUID+"/protocol-mappers/models")
	if err := c.do(ctx, http.MethodGet, path, nil, http.StatusOK, &reps); err != nil {
		return nil, fmt.Errorf("listing protocol mappers: %w", err)
	}
	return reps, nil
}

// CreateProtocolMapper adds a protocol mapper to a client.
func (c *AdminClient) CreateProtocolMapper(ctx context.Context, realm, clientUUID string, rep Representation) error {
	name, _ := rep["name"].(string)
	path := c.adminPath(realm, "/clients/"+clientUUID+"/protocol-mappers/models")
	if err := c.do(ctx, http.MethodPost, path, rep, http.StatusCreated, nil); err != nil {
		return fmt.Errorf("creating protocol mapper %s: %w", name, err)
	}
	return nil
}

// UpdateProtocolMapper PUTs a mapper representation back by its id. Same
// round-trip contract as UpdateRealm.
func (c *AdminClient) UpdateProtocolMapper(ctx context.Context, realm, clientUUID, mapperID string, rep Representation) error {
	name, _ := rep["name"].(string)
	path := c.adminPath(realm, "/clients/"+clientUUID+"/protocol-mappers/models/"+mapperID)
	if err := c.do(ctx, http.MethodPut, path, rep, http.StatusNoContent, nil); err != nil {
		return fmt.Errorf("updating protocol mapper %s: %w", name, err)
	}
	return nil
}

// DeleteProtocolMapper removes a protocol mapper by id. Used only on mappers
// the operator owns by constant name -- the arm-specific groups mapper when
// federation is unbound or the arm changes (a mapper's TYPE cannot be
// changed in place).
func (c *AdminClient) DeleteProtocolMapper(ctx context.Context, realm, clientUUID, mapperID string) error {
	path := c.adminPath(realm, "/clients/"+clientUUID+"/protocol-mappers/models/"+mapperID)
	if err := c.do(ctx, http.MethodDelete, path, nil, http.StatusNoContent, nil); err != nil {
		return fmt.Errorf("deleting protocol mapper: %w", err)
	}
	return nil
}

// ListComponents returns realm components filtered by parent and provider
// type -- user-federation providers (parent = the realm's internal id) and
// their mappers (parent = the federation component's id).
func (c *AdminClient) ListComponents(ctx context.Context, realm, parentID, providerType string) ([]Representation, error) {
	var reps []Representation
	path := c.adminPath(realm, "/components?parent="+url.QueryEscape(parentID)+"&type="+url.QueryEscape(providerType))
	if err := c.do(ctx, http.MethodGet, path, nil, http.StatusOK, &reps); err != nil {
		return nil, fmt.Errorf("listing components: %w", err)
	}
	return reps, nil
}

// CreateComponent POSTs a new realm component (a user-federation provider or
// one of its mappers).
func (c *AdminClient) CreateComponent(ctx context.Context, realm string, rep Representation) error {
	name, _ := rep["name"].(string)
	if err := c.do(ctx, http.MethodPost, c.adminPath(realm, "/components"), rep, http.StatusCreated, nil); err != nil {
		return fmt.Errorf("creating component %s: %w", name, err)
	}
	return nil
}

// UpdateComponent PUTs a component representation back by id. Same round-trip
// contract as UpdateRealm: callers pass the map the server sent, mutated only
// on owned config keys.
func (c *AdminClient) UpdateComponent(ctx context.Context, realm, id string, rep Representation) error {
	name, _ := rep["name"].(string)
	if err := c.do(ctx, http.MethodPut, c.adminPath(realm, "/components/"+id), rep, http.StatusNoContent, nil); err != nil {
		return fmt.Errorf("updating component %s: %w", name, err)
	}
	return nil
}

// DeleteComponent removes a component by id. Keycloak cascades a federation
// provider's child mappers with it. Used only on marked federation state.
func (c *AdminClient) DeleteComponent(ctx context.Context, realm, id string) error {
	if err := c.do(ctx, http.MethodDelete, c.adminPath(realm, "/components/"+id), nil, http.StatusNoContent, nil); err != nil {
		return fmt.Errorf("deleting component: %w", err)
	}
	return nil
}

// SyncResult is Keycloak's user-storage synchronization outcome.
type SyncResult struct {
	Added   int    `json:"added"`
	Updated int    `json:"updated"`
	Removed int    `json:"removed"`
	Failed  int    `json:"failed"`
	Status  string `json:"status"`
}

// TriggerUserStorageSync runs a federation user sync (action is
// triggerFullSync or triggerChangedUsersSync). Deliberately called only by
// the verification pass, never the steady-state reconcile -- a sync probes
// the company's directory.
func (c *AdminClient) TriggerUserStorageSync(ctx context.Context, realm, componentID, action string) (*SyncResult, error) {
	var result SyncResult
	path := c.adminPath(realm, "/user-storage/"+componentID+"/sync?action="+url.QueryEscape(action))
	if err := c.do(ctx, http.MethodPost, path, nil, http.StatusOK, &result); err != nil {
		return nil, fmt.Errorf("triggering user sync: %w", err)
	}
	return &result, nil
}

// TriggerLDAPMapperSync runs one LDAP mapper's sync (the group mapper's
// fedToKeycloak direction pulls directory groups into realm groups). Same
// verification-only cadence as TriggerUserStorageSync.
func (c *AdminClient) TriggerLDAPMapperSync(ctx context.Context, realm, componentID, mapperID, direction string) (*SyncResult, error) {
	var result SyncResult
	path := c.adminPath(realm, "/user-storage/"+componentID+"/mappers/"+mapperID+"/sync?direction="+url.QueryEscape(direction))
	if err := c.do(ctx, http.MethodPost, path, nil, http.StatusOK, &result); err != nil {
		return nil, fmt.Errorf("triggering group mapper sync: %w", err)
	}
	return &result, nil
}

// TestLDAPConnection asks the identity server to probe the directory --
// action "testConnection" reaches the server, "testAuthentication" binds as
// the service account. The identity server does the LDAP speaking (the
// operator never opens an LDAP connection itself); a non-2xx response's body
// carries Keycloak's reason, returned as the error for verdict composition.
func (c *AdminClient) TestLDAPConnection(ctx context.Context, realm string, rep Representation) error {
	path := c.adminPath(realm, "/testLDAPConnection")
	if err := c.do(ctx, http.MethodPost, path, rep, http.StatusNoContent, nil); err != nil {
		return err
	}
	return nil
}

// ListIdentityProviders returns the realm's identity-broker instances.
func (c *AdminClient) ListIdentityProviders(ctx context.Context, realm string) ([]Representation, error) {
	var reps []Representation
	if err := c.do(ctx, http.MethodGet, c.adminPath(realm, "/identity-provider/instances"), nil, http.StatusOK, &reps); err != nil {
		return nil, fmt.Errorf("listing identity providers: %w", err)
	}
	return reps, nil
}

// CreateIdentityProvider POSTs a new identity-broker instance.
func (c *AdminClient) CreateIdentityProvider(ctx context.Context, realm string, rep Representation) error {
	alias, _ := rep["alias"].(string)
	if err := c.do(ctx, http.MethodPost, c.adminPath(realm, "/identity-provider/instances"), rep, http.StatusCreated, nil); err != nil {
		return fmt.Errorf("creating identity provider %s: %w", alias, err)
	}
	return nil
}

// UpdateIdentityProvider PUTs a broker instance back by alias. Same
// round-trip contract as UpdateRealm.
func (c *AdminClient) UpdateIdentityProvider(ctx context.Context, realm, alias string, rep Representation) error {
	path := c.adminPath(realm, "/identity-provider/instances/"+url.PathEscape(alias))
	if err := c.do(ctx, http.MethodPut, path, rep, http.StatusNoContent, nil); err != nil {
		return fmt.Errorf("updating identity provider %s: %w", alias, err)
	}
	return nil
}

// DeleteIdentityProvider removes a broker instance (and its mappers) by
// alias. Used only on marked federation state.
func (c *AdminClient) DeleteIdentityProvider(ctx context.Context, realm, alias string) error {
	path := c.adminPath(realm, "/identity-provider/instances/"+url.PathEscape(alias))
	if err := c.do(ctx, http.MethodDelete, path, nil, http.StatusNoContent, nil); err != nil {
		return fmt.Errorf("deleting identity provider %s: %w", alias, err)
	}
	return nil
}

// ListFlowExecutions returns the executions of an authentication flow by its
// alias, in flow order. Each carries providerId (the authenticator), its
// requirement, and -- when one is attached -- authenticationConfig (the
// config's id). Read-only: the flow itself is never written by the operator.
func (c *AdminClient) ListFlowExecutions(ctx context.Context, realm, flowAlias string) ([]Representation, error) {
	var reps []Representation
	path := c.adminPath(realm, "/authentication/flows/"+url.PathEscape(flowAlias)+"/executions")
	if err := c.do(ctx, http.MethodGet, path, nil, http.StatusOK, &reps); err != nil {
		return nil, fmt.Errorf("listing executions of flow %s: %w", flowAlias, err)
	}
	return reps, nil
}

// CreateExecutionConfig attaches an authenticator config to one execution.
func (c *AdminClient) CreateExecutionConfig(ctx context.Context, realm, executionID string, rep Representation) error {
	alias, _ := rep["alias"].(string)
	path := c.adminPath(realm, "/authentication/executions/"+url.PathEscape(executionID)+"/config")
	if err := c.do(ctx, http.MethodPost, path, rep, http.StatusCreated, nil); err != nil {
		return fmt.Errorf("creating authenticator config %s: %w", alias, err)
	}
	return nil
}

// GetAuthenticatorConfig reads an authenticator config by id (alias + config map).
func (c *AdminClient) GetAuthenticatorConfig(ctx context.Context, realm, configID string) (Representation, error) {
	var rep Representation
	path := c.adminPath(realm, "/authentication/config/"+url.PathEscape(configID))
	if err := c.do(ctx, http.MethodGet, path, nil, http.StatusOK, &rep); err != nil {
		return nil, fmt.Errorf("reading authenticator config %s: %w", configID, err)
	}
	return rep, nil
}

// UpdateAuthenticatorConfig PUTs an authenticator config back by id.
func (c *AdminClient) UpdateAuthenticatorConfig(ctx context.Context, realm, configID string, rep Representation) error {
	path := c.adminPath(realm, "/authentication/config/"+url.PathEscape(configID))
	if err := c.do(ctx, http.MethodPut, path, rep, http.StatusNoContent, nil); err != nil {
		return fmt.Errorf("updating authenticator config %s: %w", configID, err)
	}
	return nil
}

// DeleteAuthenticatorConfig removes an authenticator config by id. Used only
// on the operator's own config (matched by alias), never on an admin's.
func (c *AdminClient) DeleteAuthenticatorConfig(ctx context.Context, realm, configID string) error {
	path := c.adminPath(realm, "/authentication/config/"+url.PathEscape(configID))
	if err := c.do(ctx, http.MethodDelete, path, nil, http.StatusNoContent, nil); err != nil {
		return fmt.Errorf("deleting authenticator config %s: %w", configID, err)
	}
	return nil
}

// ListIdentityProviderMappers returns a broker instance's mappers.
func (c *AdminClient) ListIdentityProviderMappers(ctx context.Context, realm, alias string) ([]Representation, error) {
	var reps []Representation
	path := c.adminPath(realm, "/identity-provider/instances/"+url.PathEscape(alias)+"/mappers")
	if err := c.do(ctx, http.MethodGet, path, nil, http.StatusOK, &reps); err != nil {
		return nil, fmt.Errorf("listing identity provider mappers: %w", err)
	}
	return reps, nil
}

// CreateIdentityProviderMapper adds a mapper to a broker instance.
func (c *AdminClient) CreateIdentityProviderMapper(ctx context.Context, realm, alias string, rep Representation) error {
	name, _ := rep["name"].(string)
	path := c.adminPath(realm, "/identity-provider/instances/"+url.PathEscape(alias)+"/mappers")
	if err := c.do(ctx, http.MethodPost, path, rep, http.StatusCreated, nil); err != nil {
		return fmt.Errorf("creating identity provider mapper %s: %w", name, err)
	}
	return nil
}

// UpdateIdentityProviderMapper PUTs a broker mapper back by id.
func (c *AdminClient) UpdateIdentityProviderMapper(ctx context.Context, realm, alias, mapperID string, rep Representation) error {
	name, _ := rep["name"].(string)
	path := c.adminPath(realm, "/identity-provider/instances/"+url.PathEscape(alias)+"/mappers/"+mapperID)
	if err := c.do(ctx, http.MethodPut, path, rep, http.StatusNoContent, nil); err != nil {
		return fmt.Errorf("updating identity provider mapper %s: %w", name, err)
	}
	return nil
}

// FindUsersByEmail returns the realm users holding exactly this email --
// the seeded-admin collision check's read.
func (c *AdminClient) FindUsersByEmail(ctx context.Context, realm, email string) ([]Representation, error) {
	var reps []Representation
	path := c.adminPath(realm, "/users?exact=true&email="+url.QueryEscape(email))
	if err := c.do(ctx, http.MethodGet, path, nil, http.StatusOK, &reps); err != nil {
		return nil, fmt.Errorf("finding users by email: %w", err)
	}
	return reps, nil
}

// FindUserByUsername returns the realm user with exactly this username;
// found=false when there is none.
func (c *AdminClient) FindUserByUsername(ctx context.Context, realm, username string) (Representation, bool, error) {
	var reps []Representation
	path := c.adminPath(realm, "/users?exact=true&username="+url.QueryEscape(username))
	if err := c.do(ctx, http.MethodGet, path, nil, http.StatusOK, &reps); err != nil {
		return nil, false, fmt.Errorf("finding user by username: %w", err)
	}
	for _, rep := range reps {
		if name, _ := rep["username"].(string); strings.EqualFold(name, username) {
			return rep, true, nil
		}
	}
	return nil, false, nil
}

// ResetUserPassword sets a user's password to a permanent value (no
// change-at-next-sign-in). Used only on the master admin the operator itself
// provisions, when a restored realm carries its source's password.
func (c *AdminClient) ResetUserPassword(ctx context.Context, realm, userID, password string) error {
	body := Representation{"type": "password", "value": password, "temporary": false}
	if err := c.do(ctx, http.MethodPut, c.adminPath(realm, "/users/"+userID+"/reset-password"), body, http.StatusNoContent, nil); err != nil {
		return fmt.Errorf("resetting user password: %w", err)
	}
	return nil
}

// DeleteUser removes a realm user. Used only on the temporary recovery admin
// the operator itself created.
func (c *AdminClient) DeleteUser(ctx context.Context, realm, userID string) error {
	if err := c.do(ctx, http.MethodDelete, c.adminPath(realm, "/users/"+userID), nil, http.StatusNoContent, nil); err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}
	return nil
}

// GetGroupByPath returns the realm group at the given path; found=false on
// 404 (the one admin call where absence is an expected answer, not an error).
func (c *AdminClient) GetGroupByPath(ctx context.Context, realm, path string) (Representation, bool, error) {
	var rep Representation
	err := c.do(ctx, http.MethodGet, c.adminPath(realm, "/group-by-path/"+url.PathEscape(strings.TrimPrefix(path, "/"))), nil, http.StatusOK, &rep)
	if err != nil {
		var statusErr *unexpectedStatusError
		if errors.As(err, &statusErr) && statusErr.status == http.StatusNotFound {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("getting group by path %s: %w", path, err)
	}
	return rep, true, nil
}

// CreateGroup POSTs a new top-level realm group.
func (c *AdminClient) CreateGroup(ctx context.Context, realm string, rep Representation) error {
	name, _ := rep["name"].(string)
	if err := c.do(ctx, http.MethodPost, c.adminPath(realm, "/groups"), rep, http.StatusCreated, nil); err != nil {
		return fmt.Errorf("creating group %s: %w", name, err)
	}
	return nil
}

// GetServiceAccountUser returns the user behind a service-accounts-enabled
// client.
func (c *AdminClient) GetServiceAccountUser(ctx context.Context, realm, clientUUID string) (Representation, error) {
	var rep Representation
	path := c.adminPath(realm, "/clients/"+clientUUID+"/service-account-user")
	if err := c.do(ctx, http.MethodGet, path, nil, http.StatusOK, &rep); err != nil {
		return nil, fmt.Errorf("getting service-account user: %w", err)
	}
	return rep, nil
}

// ListClientRoles returns the roles a client DEFINES (e.g. realm-management's
// manage-users) -- the source pool for role-mapping grants.
func (c *AdminClient) ListClientRoles(ctx context.Context, realm, clientUUID string) ([]Representation, error) {
	var reps []Representation
	if err := c.do(ctx, http.MethodGet, c.adminPath(realm, "/clients/"+clientUUID+"/roles"), nil, http.StatusOK, &reps); err != nil {
		return nil, fmt.Errorf("listing client roles: %w", err)
	}
	return reps, nil
}

// ListUserClientRoleMappings returns the client-level roles a user HOLDS.
func (c *AdminClient) ListUserClientRoleMappings(ctx context.Context, realm, userID, clientUUID string) ([]Representation, error) {
	var reps []Representation
	path := c.adminPath(realm, "/users/"+userID+"/role-mappings/clients/"+clientUUID)
	if err := c.do(ctx, http.MethodGet, path, nil, http.StatusOK, &reps); err != nil {
		return nil, fmt.Errorf("listing user role mappings: %w", err)
	}
	return reps, nil
}

// AddUserClientRoleMappings grants client-level roles to a user.
func (c *AdminClient) AddUserClientRoleMappings(ctx context.Context, realm, userID, clientUUID string, roles []Representation) error {
	path := c.adminPath(realm, "/users/"+userID+"/role-mappings/clients/"+clientUUID)
	if err := c.do(ctx, http.MethodPost, path, roles, http.StatusNoContent, nil); err != nil {
		return fmt.Errorf("adding user role mappings: %w", err)
	}
	return nil
}

// RemoveUserClientRoleMappings revokes client-level roles from a user. Used
// only on OWNED service accounts, where an extra admin role is a
// least-privilege drift, not an admin choice.
func (c *AdminClient) RemoveUserClientRoleMappings(ctx context.Context, realm, userID, clientUUID string, roles []Representation) error {
	path := c.adminPath(realm, "/users/"+userID+"/role-mappings/clients/"+clientUUID)
	if err := c.doWithBody(ctx, http.MethodDelete, path, roles, http.StatusNoContent, nil); err != nil {
		return fmt.Errorf("removing user role mappings: %w", err)
	}
	return nil
}

func (c *AdminClient) adminPath(realm, suffix string) string {
	return c.serverRoot + "/admin/realms/" + realm + suffix
}

// unexpectedStatusError carries the response the server actually sent, typed
// so callers can branch on status (absence-is-an-answer reads) without
// parsing the message.
type unexpectedStatusError struct {
	method, endpoint string
	status, want     int
	body             string
}

func (e *unexpectedStatusError) Error() string {
	return fmt.Sprintf("%s %s returned %d (want %d): %s",
		e.method, e.endpoint, e.status, e.want, e.body)
}

// ResponseBody exposes the server's own words for verdict composition -- a
// failed directory probe's reason lives in Keycloak's error body.
func (e *unexpectedStatusError) ResponseBody() string { return e.body }

// do issues one admin request: JSON body in (when non-nil), one expected
// status out, JSON decoded into out (when non-nil). Any other status becomes
// an error carrying the response body -- the reconcile loop retries next
// pass; there is deliberately no retry here (the operator's thin-client
// idiom: retry IS the reconcile cadence).
func (c *AdminClient) do(ctx context.Context, method, endpoint string, body any, expectStatus int, out any) error {
	return c.doWithBody(ctx, method, endpoint, body, expectStatus, out)
}

func (c *AdminClient) doWithBody(ctx context.Context, method, endpoint string, body any, expectStatus int, out any) error {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
		reader = strings.NewReader(string(payload))
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != expectStatus {
		respBody, _ := io.ReadAll(resp.Body)
		return &unexpectedStatusError{
			method: method, endpoint: endpoint,
			status: resp.StatusCode, want: expectStatus, body: string(respBody),
		}
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}
	return nil
}

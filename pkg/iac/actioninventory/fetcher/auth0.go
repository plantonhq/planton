package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/plantonhq/planton/pkg/iac/actioninventory"
	"github.com/plantonhq/planton/pkg/iac/permissions"
)

// The Auth0 arm authenticates exactly as the catalog's Auth0 E2E harness
// does (catalog/auth0/aa_e2e): a machine-to-machine application's client
// id and secret for the tenant AUTH0_DOMAIN names (a bare domain, e.g.
// "example.eu.auth0.com"). The application needs only read:resource_servers
// on the tenant's Management API -- it reads one definition and nothing
// else.
const (
	auth0DomainEnvVar       = "AUTH0_DOMAIN"
	auth0ClientIDEnvVar     = "AUTH0_CLIENT_ID"
	auth0ClientSecretEnvVar = "AUTH0_CLIENT_SECRET"
)

// auth0ScopeSourceURL is the snapshot's recorded source, in route form.
// The inventory is the tenant's own definition of its Management API --
// the system resource server whose identifier is the Management API
// audience -- read by that audience. The published API reference is NOT
// the source: it omits scopes the tenant enforces (it lists no
// "connections_options" scope, yet the tenant refuses a connection-options
// write without update:connections_options), and a snapshot that misses an
// enforced scope would refuse a correct manifest. The tenant's definition
// is the set Auth0 itself checks a grant against, and a tenant cannot edit
// it (the resource server is system-owned). The route form, never a
// concrete tenant, keeps the committed snapshot independent of whichever
// tenant fetched it.
const auth0ScopeSourceURL = "https://{tenant}/api/v2/resource-servers/{management-api-audience}"

// refreshAuth0 rewrites the committed Auth0 snapshot
// (pkg/iac/actioninventory/auth0.yaml) from the tenant's Management API
// definition, scoped to exactly the scope resources the committed
// permissions manifests reference. An Auth0 scope is "verb:resource", so
// the snapshot's Service prefix is the RESOURCE (the part after the
// colon) and its actions are the verbs -- the reverse of DigitalOcean's
// spelling, the same shape. Deterministic-or-dead: a referenced resource
// the definition does not list is a hard error, and an answer that is not
// the system Management API, or lists no scopes, refuses.
func refreshAuth0(repoRoot string) error {
	resources, err := referencedAuth0Resources(repoRoot)
	if err != nil {
		return err
	}
	if len(resources) == 0 {
		fmt.Println("no Auth0 scopes in any permissions manifest -- auth0 snapshot not written")
		return nil
	}

	domain := strings.TrimSpace(os.Getenv(auth0DomainEnvVar))
	clientID := strings.TrimSpace(os.Getenv(auth0ClientIDEnvVar))
	clientSecret := strings.TrimSpace(os.Getenv(auth0ClientSecretEnvVar))
	if domain == "" || clientID == "" || clientSecret == "" {
		return fmt.Errorf("the Auth0 scope inventory needs a credential: set %s (the tenant's bare domain), %s, and %s to a machine-to-machine application granted read:resource_servers on that tenant's Management API -- the variables the catalog's Auth0 E2E harness reads", auth0DomainEnvVar, auth0ClientIDEnvVar, auth0ClientSecretEnvVar)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	audience := "https://" + domain + "/api/v2/"
	token, err := auth0ManagementToken(client, domain, clientID, clientSecret, audience)
	if err != nil {
		return err
	}
	published, err := fetchAuth0Scopes(client, domain, token, audience)
	if err != nil {
		return err
	}

	today := time.Now().UTC().Format("2006-01-02")
	inv := &actioninventory.Inventory{Provider: "auth0"}
	for _, resource := range resources {
		verbs := published[resource]
		if len(verbs) == 0 {
			return fmt.Errorf("scope resource %q (used by a permissions manifest) does not exist in the tenant's Management API definition -- the resource itself is wrong", resource)
		}
		sort.Strings(verbs)
		inv.Services = append(inv.Services, actioninventory.Service{
			Prefix:      resource,
			SourceURL:   auth0ScopeSourceURL,
			RetrievedOn: today,
			Actions:     verbs,
		})
	}

	out := filepath.Join(repoRoot, "pkg", "iac", "actioninventory", actioninventory.Auth0FileName)
	if err := os.WriteFile(out, []byte(actioninventory.Render(inv)), 0o644); err != nil {
		return err
	}
	fmt.Printf("wrote %d scope resource list(s) to %s\n", len(inv.Services), out)
	return nil
}

// referencedAuth0Resources collects the distinct Auth0 scope resources
// (the segment after the colon) named by every committed permissions
// manifest, sorted.
func referencedAuth0Resources(repoRoot string) ([]string, error) {
	discovered, err := permissions.Discover(repoRoot)
	if err != nil {
		return nil, err
	}
	set := map[string]bool{}
	for provider, components := range discovered {
		for _, component := range components {
			manifest, err := permissions.Load(repoRoot, provider, component)
			if err != nil {
				return nil, err
			}
			for _, group := range manifest.GetSpec().GetAuth0().GetGroups() {
				for _, scope := range group.GetScopes() {
					_, resource, found := strings.Cut(scope, ":")
					if !found {
						return nil, fmt.Errorf("%s/%s: auth0 scope %q has no resource segment", provider, component, scope)
					}
					set[resource] = true
				}
			}
		}
	}
	resources := make([]string, 0, len(set))
	for resource := range set {
		resources = append(resources, resource)
	}
	sort.Strings(resources)
	return resources, nil
}

// auth0ManagementToken performs the client-credentials grant every
// Management API call begins with. A refusal carries Auth0's own error
// words (unauthorized_client, access_denied) -- they name the cause and
// never the secret.
func auth0ManagementToken(client *http.Client, domain, clientID, clientSecret, audience string) (string, error) {
	body, err := json.Marshal(map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     clientID,
		"client_secret": clientSecret,
		"audience":      audience,
	})
	if err != nil {
		return "", err
	}
	tokenURL := "https://" + domain + "/oauth/token"
	response, err := client.Post(tokenURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("POST %s: %w", tokenURL, err)
	}
	defer response.Body.Close()
	var grant struct {
		AccessToken      string `json:"access_token"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.NewDecoder(response.Body).Decode(&grant); err != nil {
		return "", fmt.Errorf("POST %s: %s, and the body is not a token response: %w", tokenURL, response.Status, err)
	}
	if response.StatusCode != http.StatusOK || grant.AccessToken == "" {
		return "", fmt.Errorf("POST %s: %s (%s: %s) -- the tenant refused the client credentials; check %s and %s name a machine-to-machine application authorized for %s", tokenURL, response.Status, grant.Error, grant.ErrorDescription, auth0ClientIDEnvVar, auth0ClientSecretEnvVar, audience)
	}
	return grant.AccessToken, nil
}

// fetchAuth0Scopes reads the tenant's Management API definition by its
// audience and returns every scope grouped by resource, verbs
// de-duplicated. It refuses anything that is not the system resource
// server: a tenant-defined API could share no scope with the Management
// API, and committing its scopes as Auth0's inventory would be a
// fabrication.
func fetchAuth0Scopes(client *http.Client, domain, token, audience string) (map[string][]string, error) {
	definitionURL := "https://" + domain + "/api/v2/resource-servers/" + url.PathEscape(audience)
	request, err := http.NewRequest(http.MethodGet, definitionURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", definitionURL, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s -- the application needs read:resource_servers on the Management API", definitionURL, response.Status)
	}
	var definition struct {
		Identifier string `json:"identifier"`
		IsSystem   bool   `json:"is_system"`
		Scopes     []struct {
			Value string `json:"value"`
		} `json:"scopes"`
	}
	if err := json.NewDecoder(response.Body).Decode(&definition); err != nil {
		return nil, fmt.Errorf("GET %s: decoding: %w", definitionURL, err)
	}
	if definition.Identifier != audience || !definition.IsSystem {
		return nil, fmt.Errorf("GET %s: answered %q (system: %t), not the tenant's system Management API -- refusing to snapshot a tenant-defined API as Auth0's inventory", definitionURL, definition.Identifier, definition.IsSystem)
	}

	published := map[string][]string{}
	seen := map[string]bool{}
	for _, scope := range definition.Scopes {
		if seen[scope.Value] {
			continue
		}
		seen[scope.Value] = true
		verb, resource, found := strings.Cut(scope.Value, ":")
		if !found {
			return nil, fmt.Errorf("GET %s: scope %q is not verb:resource -- the definition's format changed and the fetcher must be updated", definitionURL, scope.Value)
		}
		published[resource] = append(published[resource], verb)
	}
	if len(published) == 0 {
		return nil, fmt.Errorf("GET %s: the definition lists no scopes -- refusing to commit an empty inventory", definitionURL)
	}
	return published, nil
}

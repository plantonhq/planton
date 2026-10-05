package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/plantonhq/planton/pkg/iac/actioninventory"
	"github.com/plantonhq/planton/pkg/iac/permissions"
)

// digitalOceanScopesURL is DigitalOcean's published token-scope
// reference, served as machine-readable markdown (the docs site's own
// convention: any page's markdown form is index.html.md). DigitalOcean
// exposes NO scope-inventory API -- this page IS the provider's published
// inventory, and the snapshot's provenance says so: the URL plus the
// retrieval date, never a claim of an API the provider does not offer.
const digitalOceanScopesURL = "https://docs.digitalocean.com/reference/api/scopes/index.html.md"

// digitalOceanScopeRowPattern matches one scope table row's scope cell,
// e.g. "| [`droplet:create`](./droplet/create) | Create Droplets |".
// Underscores are legal in both segments (block_storage_action,
// view_credentials).
var digitalOceanScopeRowPattern = regexp.MustCompile("\\[`([a-z0-9_]+:[a-z0-9_]+)`\\]")

// refreshDigitalOcean rewrites the committed DigitalOcean snapshot
// (pkg/iac/actioninventory/digitalocean.yaml) from the provider's
// published scope reference, scoped to exactly the scope prefixes (the
// resource segment before the colon) the committed permissions manifests
// reference. Deterministic-or-dead: a referenced prefix the reference
// does not list is a hard error (a genuinely wrong prefix), and a page
// that yields no scopes at all refuses -- the reference's format changed
// and the parser must be updated, never silently committed around.
func refreshDigitalOcean(repoRoot string) error {
	prefixes, err := referencedDigitalOceanPrefixes(repoRoot)
	if err != nil {
		return err
	}
	if len(prefixes) == 0 {
		fmt.Println("no DigitalOcean scopes in any permissions manifest -- digitalocean snapshot not written")
		return nil
	}

	client := &http.Client{Timeout: 60 * time.Second}
	published, err := fetchDigitalOceanScopes(client)
	if err != nil {
		return err
	}

	today := time.Now().UTC().Format("2006-01-02")
	inv := &actioninventory.Inventory{Provider: "digitalocean"}
	for _, prefix := range prefixes {
		actions := published[prefix]
		if len(actions) == 0 {
			return fmt.Errorf("scope prefix %q (used by a permissions manifest) does not exist in DigitalOcean's scope reference -- the prefix itself is wrong", prefix)
		}
		sort.Strings(actions)
		inv.Services = append(inv.Services, actioninventory.Service{
			Prefix:      prefix,
			SourceURL:   digitalOceanScopesURL,
			RetrievedOn: today,
			Actions:     actions,
		})
	}

	out := filepath.Join(repoRoot, "pkg", "iac", "actioninventory", actioninventory.DigitalOceanFileName)
	if err := os.WriteFile(out, []byte(actioninventory.Render(inv)), 0o644); err != nil {
		return err
	}
	fmt.Printf("wrote %d scope prefix list(s) to %s\n", len(inv.Services), out)
	return nil
}

// referencedDigitalOceanPrefixes collects the distinct DigitalOcean scope
// prefixes named by every committed permissions manifest, sorted.
func referencedDigitalOceanPrefixes(repoRoot string) ([]string, error) {
	discovered, err := permissions.Discover(repoRoot)
	if err != nil {
		return nil, err
	}
	set := map[string]bool{}
	for provider, kinds := range discovered {
		for _, kind := range kinds {
			manifest, err := permissions.Load(repoRoot, provider, kind)
			if err != nil {
				return nil, err
			}
			for _, group := range manifest.GetSpec().GetDigitalOcean().GetGroups() {
				for _, scope := range group.GetScopes() {
					prefix, _, found := strings.Cut(scope, ":")
					if !found {
						return nil, fmt.Errorf("%s/%s: digitalocean scope %q has no resource prefix", provider, kind, scope)
					}
					set[prefix] = true
				}
			}
		}
	}
	prefixes := make([]string, 0, len(set))
	for prefix := range set {
		prefixes = append(prefixes, prefix)
	}
	sort.Strings(prefixes)
	return prefixes, nil
}

// fetchDigitalOceanScopes reads the published scope reference and returns
// every listed scope grouped by prefix, de-duplicated.
func fetchDigitalOceanScopes(client *http.Client) (map[string][]string, error) {
	response, err := client.Get(digitalOceanScopesURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", digitalOceanScopesURL, response.Status)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("GET %s: reading: %w", digitalOceanScopesURL, err)
	}

	published := map[string][]string{}
	seen := map[string]bool{}
	for _, match := range digitalOceanScopeRowPattern.FindAllStringSubmatch(string(body), -1) {
		scope := match[1]
		if seen[scope] {
			continue
		}
		seen[scope] = true
		prefix, action, _ := strings.Cut(scope, ":")
		published[prefix] = append(published[prefix], action)
	}
	if len(published) == 0 {
		return nil, fmt.Errorf("GET %s: the scope reference yielded no scopes -- its format changed and the fetcher's parser must be updated", digitalOceanScopesURL)
	}
	return published, nil
}

// Command fetcher refreshes the committed action-inventory snapshots from
// each provider's own published inventory: AWS
// (pkg/iac/actioninventory/aws.yaml) from the machine-readable service
// reference, Azure (azure.yaml) from ARM's provider-operations metadata
// (see azure.go for its arm and credential contract), GCP (gcp.yaml) from
// IAM's testable-permissions inventory (see gcp.go), and Cloudflare,
// DigitalOcean, and Auth0 from their own arms' files. It reads the
// committed runner permissions manifests to learn which service prefixes
// / namespaces / services the catalog actually uses, fetches exactly
// those inventories, and rewrites each snapshot in its canonical form.
// Deterministic-or-dead: a manifest prefix the provider does not define
// is a hard error (a genuinely wrong prefix, not a refresh problem), and a
// service with an empty action list refuses rather than committing an
// inventory that would fail every action.
//
// With no arguments every arm refreshes, in the order below. Naming arms
// (`go run ./pkg/iac/actioninventory/fetcher auth0`) refreshes only those:
// several arms need their own provider credential, and refreshing one
// provider must neither require the others' credentials nor rewrite their
// snapshots with a new retrieval date and nothing else changed.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/plantonhq/planton/pkg/iac/actioninventory"
	"github.com/plantonhq/planton/pkg/iac/permissions"
)

// indexURL lists every service the reference covers with its per-service
// document URL and modification stamp.
const indexURL = "https://servicereference.us-east-1.amazonaws.com/"

// arms are the inventory arms by the catalog provider directory name each
// snapshot serves, in the order a full refresh runs them.
var arms = []struct {
	provider string
	refresh  func(repoRoot string) error
}{
	{"aws", refreshAws},
	{"azure", refreshAzure},
	{"gcp", refreshGcp},
	{"cloudflare", refreshCloudflare},
	{"digitalocean", refreshDigitalOcean},
	{"auth0", refreshAuth0},
}

func main() {
	repoRoot, err := os.Getwd()
	if err != nil {
		fatal(err)
	}

	selected := map[string]bool{}
	known := make([]string, 0, len(arms))
	for _, arm := range arms {
		known = append(known, arm.provider)
	}
	for _, name := range os.Args[1:] {
		found := false
		for _, arm := range arms {
			found = found || arm.provider == name
		}
		if !found {
			fatal(fmt.Errorf("no inventory arm named %q -- the arms are %s", name, strings.Join(known, ", ")))
		}
		selected[name] = true
	}

	for _, arm := range arms {
		if len(selected) > 0 && !selected[arm.provider] {
			continue
		}
		if err := arm.refresh(repoRoot); err != nil {
			fatal(fmt.Errorf("%s: %w", arm.provider, err))
		}
	}
}

// refreshAws rewrites the committed AWS snapshot from the service
// reference, scoped to the service prefixes the manifests reference.
func refreshAws(repoRoot string) error {
	prefixes, err := referencedAwsPrefixes(repoRoot)
	if err != nil {
		return err
	}
	if len(prefixes) == 0 {
		return fmt.Errorf("no AWS actions found in any permissions manifest -- nothing to inventory")
	}

	client := &http.Client{Timeout: 60 * time.Second}
	index, err := fetchIndex(client)
	if err != nil {
		return err
	}

	today := time.Now().UTC().Format("2006-01-02")
	inv := &actioninventory.Inventory{Provider: "aws"}
	for _, prefix := range prefixes {
		entry, ok := index[prefix]
		if !ok {
			return fmt.Errorf("service prefix %q (used by a permissions manifest) does not exist in the AWS service reference -- the prefix itself is wrong", prefix)
		}
		actions, nonScopable, err := fetchServiceActions(client, entry.URL)
		if err != nil {
			return fmt.Errorf("service %q: %w", prefix, err)
		}
		if len(actions) == 0 {
			return fmt.Errorf("service %q: the reference lists no actions -- refusing to commit an empty inventory", prefix)
		}
		inv.Services = append(inv.Services, actioninventory.Service{
			Prefix:             prefix,
			SourceURL:          entry.URL,
			SourceModified:     time.Unix(entry.Modified, 0).UTC().Format("2006-01-02"),
			RetrievedOn:        today,
			Actions:            actions,
			NonScopableActions: nonScopable,
		})
	}

	out := filepath.Join(repoRoot, "pkg", "iac", "actioninventory", actioninventory.AwsFileName)
	if err := os.WriteFile(out, []byte(actioninventory.Render(inv)), 0o644); err != nil {
		return err
	}
	fmt.Printf("wrote %d service action list(s) to %s\n", len(inv.Services), out)
	return nil
}

// referencedAwsPrefixes collects the distinct AWS service prefixes named by
// every committed permissions manifest, sorted.
func referencedAwsPrefixes(repoRoot string) ([]string, error) {
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
			for _, statement := range manifest.GetSpec().GetAws().GetStatements() {
				for _, action := range statement.GetActions() {
					prefix, _, found := strings.Cut(action, ":")
					if !found {
						return nil, fmt.Errorf("%s/%s: action %q has no service prefix", provider, component, action)
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

// indexEntry is one service's row in the reference index.
type indexEntry struct {
	URL      string
	Modified int64
}

func fetchIndex(client *http.Client) (map[string]indexEntry, error) {
	var rows []struct {
		Service  string `json:"service"`
		URL      string `json:"url"`
		Modified int64  `json:"modified"`
	}
	if err := getJSON(client, indexURL, &rows); err != nil {
		return nil, fmt.Errorf("fetching the service reference index: %w", err)
	}
	index := make(map[string]indexEntry, len(rows))
	for _, row := range rows {
		index[row.Service] = indexEntry{URL: row.URL, Modified: row.Modified}
	}
	if len(index) == 0 {
		return nil, fmt.Errorf("the service reference index is empty")
	}
	return index, nil
}

// fetchServiceActions reads one service's reference document and returns
// its action names plus the subset that declares no resource types, both
// sorted and de-duplicated. The reference marks a scopable action with a
// Resources list naming the resource types its ARNs can name; an action
// with NO Resources entry is evaluated by IAM against Resource "*" only,
// and the snapshot must carry that fact so the scopability gate can hold
// statements to it.
func fetchServiceActions(client *http.Client, url string) ([]string, []string, error) {
	var doc struct {
		Actions []struct {
			Name      string `json:"Name"`
			Resources []struct {
				Name string `json:"Name"`
			} `json:"Resources"`
		} `json:"Actions"`
	}
	if err := getJSON(client, url, &doc); err != nil {
		return nil, nil, err
	}
	seen := map[string]bool{}
	var actions, nonScopable []string
	for _, action := range doc.Actions {
		if action.Name == "" || seen[action.Name] {
			continue
		}
		seen[action.Name] = true
		actions = append(actions, action.Name)
		if len(action.Resources) == 0 {
			nonScopable = append(nonScopable, action.Name)
		}
	}
	sort.Strings(actions)
	sort.Strings(nonScopable)
	return actions, nonScopable, nil
}

func getJSON(client *http.Client, url string, out any) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("GET %s: decoding: %w", url, err)
	}
	return nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "action-inventory fetcher:", err)
	os.Exit(1)
}

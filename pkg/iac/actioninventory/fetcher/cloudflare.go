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

// cloudflareTokenEnvVar carries the Cloudflare API token the fetcher
// authenticates with -- the same variable the repo's Cloudflare E2E lane
// uses. An ACCOUNT-owned token works (and is what the operator's account
// standardizes on): the permission-group catalog is global, served
// identically to every account, so the fetcher discovers the token's own
// account and reads the catalog through it. The token needs only the
// account-level token-read permission; it can touch no zones or
// resources.
const cloudflareTokenEnvVar = "CLOUDFLARE_API_TOKEN"

// cloudflareAccountsURL lists the accounts the token can see -- the
// fetcher's way of discovering an account id to route the
// permission-group read through.
const cloudflareAccountsURL = "https://api.cloudflare.com/client/v4/accounts"

// cloudflarePermissionGroupsURL is the permission-group inventory
// endpoint, in its documented route form. The snapshot records THIS form
// as source_url -- never a concrete account id -- because the catalog is
// global and the committed snapshot must not vary by which operator
// account fetched it.
const cloudflarePermissionGroupsURL = "https://api.cloudflare.com/client/v4/accounts/{account_id}/tokens/permission_groups"

// refreshCloudflare rewrites the committed Cloudflare snapshot
// (pkg/iac/actioninventory/cloudflare.yaml) from Cloudflare's own
// permission-group inventory, scoped to exactly the group names the
// committed permissions manifests reference. Deterministic-or-dead: a
// referenced group name the provider does not define is a hard error (a
// genuinely wrong name, not a refresh problem).
func refreshCloudflare(repoRoot string) error {
	referenced, err := referencedCloudflareGroups(repoRoot)
	if err != nil {
		return err
	}
	if len(referenced) == 0 {
		fmt.Println("no Cloudflare permission groups in any permissions manifest -- cloudflare snapshot not written")
		return nil
	}

	token := strings.TrimSpace(os.Getenv(cloudflareTokenEnvVar))
	if token == "" {
		return fmt.Errorf("the Cloudflare permission-group inventory needs a credential: set %s to an API token that can read the account's token permission groups (an account-owned token with the token-read permission suffices; it reads only Cloudflare's global permission-group catalog)", cloudflareTokenEnvVar)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	accountID, err := cloudflareAccountID(client, token)
	if err != nil {
		return err
	}

	groups, err := fetchCloudflarePermissionGroups(client, token, accountID)
	if err != nil {
		return err
	}

	byName := map[string][]actioninventory.PermissionGroup{}
	for _, group := range groups {
		byName[group.Name] = append(byName[group.Name], group)
	}

	inv := &actioninventory.GroupInventory{
		Provider:    "cloudflare",
		SourceURL:   cloudflarePermissionGroupsURL,
		RetrievedOn: time.Now().UTC().Format("2006-01-02"),
	}
	names := make([]string, 0, len(referenced))
	for name := range referenced {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		rows := byName[name]
		if len(rows) == 0 {
			return fmt.Errorf("permission group %q (used by a permissions manifest) does not exist in Cloudflare's inventory -- the name is invented, misspelled, or renamed by the provider", name)
		}
		inv.Groups = append(inv.Groups, rows...)
	}
	sort.Slice(inv.Groups, func(i, j int) bool {
		if inv.Groups[i].Name != inv.Groups[j].Name {
			return inv.Groups[i].Name < inv.Groups[j].Name
		}
		return inv.Groups[i].ID < inv.Groups[j].ID
	})

	out := filepath.Join(repoRoot, "pkg", "iac", "actioninventory", actioninventory.CloudflareFileName)
	if err := os.WriteFile(out, []byte(actioninventory.RenderGroups(inv)), 0o644); err != nil {
		return err
	}
	fmt.Printf("wrote %d permission group(s) to %s\n", len(inv.Groups), out)
	return nil
}

// referencedCloudflareGroups collects the distinct Cloudflare
// permission-group names named by every committed permissions manifest.
func referencedCloudflareGroups(repoRoot string) (map[string]bool, error) {
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
			for _, group := range manifest.GetSpec().GetCloudflare().GetGroups() {
				if group.GetName() == "" {
					return nil, fmt.Errorf("%s/%s: cloudflare group with empty name", provider, kind)
				}
				set[group.GetName()] = true
			}
		}
	}
	return set, nil
}

// cloudflareEnvelope is Cloudflare's standard response wrapper.
type cloudflareEnvelope struct {
	Success bool `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
	ResultInfo *struct {
		Count      int `json:"count"`
		TotalCount int `json:"total_count"`
	} `json:"result_info"`
}

// cloudflareAccountID discovers an account the token can see. WHICH
// account does not matter -- the permission-group catalog is global -- so
// the fetcher picks the lexicographically first id for determinism of
// behavior (the snapshot's content is account-independent either way).
func cloudflareAccountID(client *http.Client, token string) (string, error) {
	var response struct {
		cloudflareEnvelope
		Result []struct {
			ID string `json:"id"`
		} `json:"result"`
	}
	if err := cloudflareGetJSON(client, token, cloudflareAccountsURL, &response); err != nil {
		return "", err
	}
	if !response.Success || len(response.Result) == 0 {
		return "", fmt.Errorf("GET %s: the token can see no accounts (errors: %v) -- the inventory read routes through an account, so the token must carry account membership", cloudflareAccountsURL, response.Errors)
	}
	ids := make([]string, 0, len(response.Result))
	for _, account := range response.Result {
		if account.ID != "" {
			ids = append(ids, account.ID)
		}
	}
	if len(ids) == 0 {
		return "", fmt.Errorf("GET %s: accounts listed without ids", cloudflareAccountsURL)
	}
	sort.Strings(ids)
	return ids[0], nil
}

// fetchCloudflarePermissionGroups reads the full permission-group catalog
// through one account. The endpoint serves the whole catalog in one
// response today; if Cloudflare ever paginates it, the count check dies
// loudly rather than committing a silently truncated inventory.
func fetchCloudflarePermissionGroups(client *http.Client, token, accountID string) ([]actioninventory.PermissionGroup, error) {
	url := strings.Replace(cloudflarePermissionGroupsURL, "{account_id}", accountID, 1)
	var response struct {
		cloudflareEnvelope
		Result []struct {
			ID     string   `json:"id"`
			Name   string   `json:"name"`
			Scopes []string `json:"scopes"`
		} `json:"result"`
	}
	if err := cloudflareGetJSON(client, token, url, &response); err != nil {
		return nil, err
	}
	if !response.Success {
		return nil, fmt.Errorf("GET %s: request failed (errors: %v)", cloudflarePermissionGroupsURL, response.Errors)
	}
	if len(response.Result) == 0 {
		return nil, fmt.Errorf("GET %s: the inventory came back empty -- refusing to commit an empty snapshot", cloudflarePermissionGroupsURL)
	}
	if response.ResultInfo != nil && response.ResultInfo.TotalCount > len(response.Result) {
		return nil, fmt.Errorf("GET %s: the endpoint returned %d of %d groups -- it has become paginated and the fetcher must learn to page before a truncated inventory can ship", cloudflarePermissionGroupsURL, len(response.Result), response.ResultInfo.TotalCount)
	}
	groups := make([]actioninventory.PermissionGroup, 0, len(response.Result))
	for _, row := range response.Result {
		if row.ID == "" || row.Name == "" {
			continue
		}
		scopes := append([]string(nil), row.Scopes...)
		sort.Strings(scopes)
		groups = append(groups, actioninventory.PermissionGroup{ID: row.ID, Name: row.Name, Scopes: scopes})
	}
	return groups, nil
}

// cloudflareGetJSON issues one authenticated GET against Cloudflare's API
// and decodes the enveloped response. Auth failures surface the HTTP
// status; envelope-level failures are the callers' to interpret.
func cloudflareGetJSON(client *http.Client, token, url string, out any) error {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: %s", url, response.Status)
	}
	if err := json.NewDecoder(response.Body).Decode(out); err != nil {
		return fmt.Errorf("GET %s: decoding: %w", url, err)
	}
	return nil
}

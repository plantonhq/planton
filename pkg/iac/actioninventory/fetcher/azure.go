package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/plantonhq/planton/pkg/iac/actioninventory"
	"github.com/plantonhq/planton/pkg/iac/permissions"
)

// providerOperationsURL is ARM's provider-operations metadata endpoint for
// one namespace. The $expand carries the per-resource-type operations,
// which is where most of a namespace's inventory lives.
const providerOperationsURL = "https://management.azure.com/providers/Microsoft.Authorization/providerOperations/%s?api-version=2022-04-01&$expand=resourceTypes"

// refreshAzure rewrites the committed Azure snapshot
// (pkg/iac/actioninventory/azure.yaml) from ARM's provider-operations
// metadata -- the same inventory `az provider operation list` reads --
// scoped to exactly the ARM namespaces the committed permissions
// manifests reference on either plane. The credential contract mirrors
// the GCP price fetcher's: the operator's own authenticated Azure CLI
// supplies the ARM bearer token (network in make, never in CI). Azure
// splits management-plane operations from data-plane operations, and the
// snapshot preserves that split so the conformance gate can hold a
// manifest's `actions` and `data_actions` each to its own plane.
func refreshAzure(repoRoot string) error {
	namespaces, err := referencedAzureNamespaces(repoRoot)
	if err != nil {
		return err
	}
	if len(namespaces) == 0 {
		fmt.Println("no Azure actions in any permissions manifest -- azure snapshot not written")
		return nil
	}

	token, err := azureAccessToken()
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 60 * time.Second}
	today := time.Now().UTC().Format("2006-01-02")
	inv := &actioninventory.Inventory{Provider: "azure"}
	for _, namespace := range namespaces {
		sourceURL := fmt.Sprintf(providerOperationsURL, namespace)
		operations, dataOperations, err := fetchNamespaceOperations(client, token, sourceURL, namespace)
		if err != nil {
			return fmt.Errorf("namespace %q: %w", namespace, err)
		}
		if len(operations) == 0 && len(dataOperations) == 0 {
			return fmt.Errorf("namespace %q: ARM lists no operations -- refusing to commit an empty inventory", namespace)
		}
		inv.Services = append(inv.Services, actioninventory.Service{
			Prefix:      namespace,
			SourceURL:   sourceURL,
			RetrievedOn: today,
			Actions:     operations,
			DataActions: dataOperations,
		})
	}

	out := filepath.Join(repoRoot, "pkg", "iac", "actioninventory", actioninventory.AzureFileName)
	if err := os.WriteFile(out, []byte(actioninventory.Render(inv)), 0o644); err != nil {
		return err
	}
	fmt.Printf("wrote %d namespace operation list(s) to %s\n", len(inv.Services), out)
	return nil
}

// referencedAzureNamespaces collects the distinct ARM namespaces named by
// every committed permissions manifest on either plane, sorted.
func referencedAzureNamespaces(repoRoot string) ([]string, error) {
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
			for _, group := range manifest.GetSpec().GetAzure().GetGroups() {
				for _, action := range append(append([]string(nil), group.GetActions()...), group.GetDataActions()...) {
					namespace, _, found := strings.Cut(action, "/")
					if !found {
						return nil, fmt.Errorf("%s/%s: azure action %q has no namespace segment", provider, kind, action)
					}
					set[namespace] = true
				}
			}
		}
	}
	namespaces := make([]string, 0, len(set))
	for namespace := range set {
		namespaces = append(namespaces, namespace)
	}
	sort.Strings(namespaces)
	return namespaces, nil
}

// azureAccessToken obtains an ARM bearer token from the operator's own
// authenticated Azure CLI -- the fetcher's credential contract. A missing
// or signed-out CLI is a hard error naming the way in.
func azureAccessToken() (string, error) {
	cmd := exec.Command("az", "account", "get-access-token", "--resource", "https://management.azure.com/", "--query", "accessToken", "-o", "tsv")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("obtaining an ARM token via `az account get-access-token` (is the Azure CLI installed and signed in?): %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	token := strings.TrimSpace(stdout.String())
	if token == "" {
		return "", fmt.Errorf("`az account get-access-token` returned an empty token")
	}
	return token, nil
}

// fetchNamespaceOperations reads one namespace's provider-operations
// document and returns its operation names partitioned by plane, sorted,
// de-duplicated, and stripped of the namespace prefix.
func fetchNamespaceOperations(client *http.Client, token, url, namespace string) (operations, dataOperations []string, err error) {
	type armOperation struct {
		Name         string `json:"name"`
		IsDataAction bool   `json:"isDataAction"`
	}
	var doc struct {
		Operations    []armOperation `json:"operations"`
		ResourceTypes []struct {
			Operations []armOperation `json:"operations"`
		} `json:"resourceTypes"`
	}
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := client.Do(request)
	if err != nil {
		return nil, nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("GET %s: %s", url, response.Status)
	}
	if err := json.NewDecoder(response.Body).Decode(&doc); err != nil {
		return nil, nil, fmt.Errorf("GET %s: decoding: %w", url, err)
	}

	all := append([]armOperation(nil), doc.Operations...)
	for _, resourceType := range doc.ResourceTypes {
		all = append(all, resourceType.Operations...)
	}
	prefix := namespace + "/"
	seen := map[string]bool{}
	for _, operation := range all {
		name, found := strings.CutPrefix(operation.Name, prefix)
		if !found || name == "" {
			// ARM occasionally lists operations under aliased namespaces
			// (case variants of the queried one); normalize by matching
			// case-insensitively and keeping the queried namespace's form.
			lowered := strings.ToLower(operation.Name)
			if !strings.HasPrefix(lowered, strings.ToLower(prefix)) {
				continue
			}
			name = operation.Name[len(prefix):]
			if name == "" {
				continue
			}
		}
		key := fmt.Sprintf("%t/%s", operation.IsDataAction, name)
		if seen[key] {
			continue
		}
		seen[key] = true
		if operation.IsDataAction {
			dataOperations = append(dataOperations, name)
		} else {
			operations = append(operations, name)
		}
	}
	sort.Strings(operations)
	sort.Strings(dataOperations)
	return operations, dataOperations, nil
}

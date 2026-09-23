package actioninventory

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/plantonhq/planton/pkg/iac/permissions"
)

// repoRoot resolves the repository root from this file's location so the
// gate always reads the live tree it ships in.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve caller path")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(thisFile))))
}

func packageDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve caller path")
	}
	return filepath.Dir(thisFile)
}

// TestAwsActionsExist is the gate: every AWS action every committed runner
// permissions manifest names must exist in the provider's own inventory
// snapshot. Exact names must match a published spelling (case-insensitive,
// IAM's evaluation semantics); wildcard patterns must match at least one
// published action -- a pattern matching nothing is a fabricated
// permission wearing a wildcard. Providers without an inventory arm
// (kubernetes) are exempt HERE deliberately: their structural validation
// lives in pkg/iac/permissions, and their existence arms join this gate
// when their machine-readable inventories are proven. Azure's arm is
// TestAzureActionsExist and GCP's is TestGcpPermissionsExist below.
func TestAwsActionsExist(t *testing.T) {
	root := repoRoot(t)
	inv, err := LoadAws(packageDir(t))
	if err != nil {
		t.Fatalf("loading inventory: %v", err)
	}

	discovered, err := permissions.Discover(root)
	if err != nil {
		t.Fatalf("discovering permissions manifests: %v", err)
	}

	referenced := map[string]bool{}
	for provider, components := range discovered {
		for _, component := range components {
			manifest, err := permissions.Load(root, provider, component)
			if err != nil {
				t.Fatalf("loading %s/%s: %v", provider, component, err)
			}
			for _, statement := range manifest.GetSpec().GetAws().GetStatements() {
				// The scopability census for this statement: one
				// non-scopable action anywhere in it decides the whole
				// statement's resource shape (IAM evaluates every action
				// against the same Resource list).
				nonScopableAction := ""
				scopableAction := ""
				allResolved := true
				for _, action := range statement.GetActions() {
					prefix, name, found := strings.Cut(action, ":")
					if !found {
						t.Errorf("%s/%s: action %q has no service prefix", provider, component, action)
						allResolved = false
						continue
					}
					referenced[prefix] = true
					svc := inv.Lookup(prefix)
					if svc == nil {
						t.Errorf("%s/%s: action %q names service %q which the inventory snapshot does not cover -- run `make generate-action-inventory`", provider, component, action, prefix)
						allResolved = false
						continue
					}
					total := MatchAction(svc.Actions, name)
					if total == 0 {
						t.Errorf("%s/%s: action %q does not exist in AWS's service reference for %q -- the name is invented or misspelled", provider, component, action, prefix)
						allResolved = false
						continue
					}
					nonScopable := MatchAction(svc.NonScopableActions, name)
					if nonScopable > 0 && nonScopableAction == "" {
						nonScopableAction = action
					}
					if total > nonScopable && scopableAction == "" {
						scopableAction = action
					}
				}
				// The scopability gate. AWS's reference declares, per
				// action, the resource types an ARN can name; an action
				// declaring NONE is evaluated against Resource "*" only,
				// so a statement carrying one must grant exactly "*" --
				// an ARN-scoped grant reads tighter than required and
				// silently DENIES at runtime, the quiet failure mode a
				// least-privilege catalog must make impossible. The
				// converse holds too: a "*" statement whose actions all
				// scope is wider than required -- split the statement and
				// scope what scopes (the manifests' own sibling-split
				// idiom).
				resources := statement.GetResources()
				starOnly := len(resources) == 1 && resources[0] == "*"
				if nonScopableAction != "" && !starOnly {
					t.Errorf("%s/%s: statement %q grants %s, which AWS's reference lists NO resource types for -- IAM evaluates it against Resource \"*\" only, so this statement's resources %v never match and the grant denies at runtime; make the resources exactly [\"*\"] (moving scopable siblings to their own scoped statement)", provider, component, statement.GetSid(), nonScopableAction, resources)
				}
				if starOnly && allResolved && nonScopableAction == "" && scopableAction != "" {
					t.Errorf("%s/%s: statement %q grants Resource \"*\" but every action in it (e.g. %s) supports resource-level scoping -- scope the statement to the resources the module manages, or defend the wildcard where truly unavoidable", provider, component, statement.GetSid(), scopableAction)
				}
			}
		}
	}

	// Dead weight is rejected like the price books' dead prices: a
	// snapshot service no manifest references bloats every refresh for
	// nothing -- the fetcher scopes to referenced services, so a stale
	// extra means the manifests moved and the snapshot did not.
	for _, svc := range inv.Services {
		if !referenced[svc.Prefix] {
			t.Errorf("inventory covers service %q which no permissions manifest references -- run `make generate-action-inventory`", svc.Prefix)
		}
	}
}

// TestAzureActionsExist is the Azure arm of the gate: every ARM operation
// every committed permissions manifest names must exist in ARM's own
// provider-operations inventory -- ON ITS OWN PLANE. Azure role
// definitions separate management-plane `actions` from data-plane
// `dataActions`, the permissions schema mirrors the split, and the
// snapshot records each plane; an operation that exists only on the other
// plane is a modeling error the gate names distinctly, because a role
// definition carrying it would silently grant nothing.
func TestAzureActionsExist(t *testing.T) {
	root := repoRoot(t)
	inv, err := LoadAzure(packageDir(t))
	if err != nil {
		t.Fatalf("loading inventory: %v", err)
	}

	discovered, err := permissions.Discover(root)
	if err != nil {
		t.Fatalf("discovering permissions manifests: %v", err)
	}

	referenced := map[string]bool{}
	for provider, components := range discovered {
		for _, component := range components {
			manifest, err := permissions.Load(root, provider, component)
			if err != nil {
				t.Fatalf("loading %s/%s: %v", provider, component, err)
			}
			for _, group := range manifest.GetSpec().GetAzure().GetGroups() {
				checkAzurePlane(t, inv, referenced, provider, component, "actions", group.GetActions(), false)
				checkAzurePlane(t, inv, referenced, provider, component, "data_actions", group.GetDataActions(), true)
			}
		}
	}

	// The dead-weight rule, as on the AWS arm: a snapshot namespace no
	// manifest references means the manifests moved and the snapshot did
	// not.
	for _, svc := range inv.Services {
		if !referenced[svc.Prefix] {
			t.Errorf("inventory covers namespace %q which no permissions manifest references -- run `make generate-action-inventory`", svc.Prefix)
		}
	}
}

// TestGcpPermissionsExist is the GCP arm of the gate: every dotted IAM
// permission every committed permissions manifest names must exist in
// GCP's own testable-permissions inventory. The snapshot keys permissions
// by their service segment (the part before the first dot); a permission
// whose service the snapshot lacks names the refresh command, and a name
// the service never defined is called out as invented -- the structural
// regex in pkg/iac/permissions would happily accept
// "container.clusters.frobnicate"; this gate is what makes that class of
// wrong impossible.
func TestGcpPermissionsExist(t *testing.T) {
	root := repoRoot(t)
	inv, err := LoadGcp(packageDir(t))
	if err != nil {
		t.Fatalf("loading inventory: %v", err)
	}

	discovered, err := permissions.Discover(root)
	if err != nil {
		t.Fatalf("discovering permissions manifests: %v", err)
	}

	referenced := map[string]bool{}
	for provider, components := range discovered {
		for _, component := range components {
			manifest, err := permissions.Load(root, provider, component)
			if err != nil {
				t.Fatalf("loading %s/%s: %v", provider, component, err)
			}
			for _, group := range manifest.GetSpec().GetGcp().GetGroups() {
				for _, permission := range group.GetPermissions() {
					service, name, found := strings.Cut(permission, ".")
					if !found {
						t.Errorf("%s/%s: gcp permission %q has no service segment", provider, component, permission)
						continue
					}
					referenced[service] = true
					published := inv.ServiceActions(service)
					if published == nil {
						t.Errorf("%s/%s: permission %q names service %q which the inventory snapshot does not cover -- run `make generate-action-inventory`", provider, component, permission, service)
						continue
					}
					if MatchAction(published, name) == 0 {
						t.Errorf("%s/%s: permission %q does not exist in GCP's IAM inventory for %q -- the name is invented or misspelled", provider, component, permission, service)
					}
				}
			}
		}
	}

	// The dead-weight rule, as on the AWS and Azure arms.
	for _, svc := range inv.Services {
		if !referenced[svc.Prefix] {
			t.Errorf("inventory covers service %q which no permissions manifest references -- run `make generate-action-inventory`", svc.Prefix)
		}
	}
}

// TestDigitalOceanScopesExist is the DigitalOcean arm of the gate: every
// token scope every committed permissions manifest names must exist in
// the provider's own published scope reference. Matching is EXACT string
// membership -- deliberately not MatchAction's IAM glob semantics,
// because DigitalOcean does not evaluate wildcards and borrowing IAM's
// matcher would claim semantics the provider does not have. The
// structural regex in pkg/iac/permissions would happily accept
// "droplet:frobnicate"; this gate is what makes that class of wrong
// impossible.
func TestDigitalOceanScopesExist(t *testing.T) {
	root := repoRoot(t)
	inv, err := LoadDigitalOcean(packageDir(t))
	if err != nil {
		t.Fatalf("loading inventory: %v", err)
	}

	discovered, err := permissions.Discover(root)
	if err != nil {
		t.Fatalf("discovering permissions manifests: %v", err)
	}

	referenced := map[string]bool{}
	for provider, components := range discovered {
		for _, component := range components {
			manifest, err := permissions.Load(root, provider, component)
			if err != nil {
				t.Fatalf("loading %s/%s: %v", provider, component, err)
			}
			for _, group := range manifest.GetSpec().GetDigitalOcean().GetGroups() {
				for _, scope := range group.GetScopes() {
					prefix, action, found := strings.Cut(scope, ":")
					if !found {
						t.Errorf("%s/%s: scope %q has no resource prefix", provider, component, scope)
						continue
					}
					referenced[prefix] = true
					published := inv.ServiceActions(prefix)
					if published == nil {
						t.Errorf("%s/%s: scope %q names resource %q which the inventory snapshot does not cover -- run `make generate-action-inventory`", provider, component, scope, prefix)
						continue
					}
					if !containsExact(published, action) {
						t.Errorf("%s/%s: scope %q does not exist in DigitalOcean's scope reference for %q -- the name is invented or misspelled", provider, component, scope, prefix)
					}
				}
			}
		}
	}

	// The dead-weight rule, as on every arm.
	for _, svc := range inv.Services {
		if !referenced[svc.Prefix] {
			t.Errorf("inventory covers resource %q which no permissions manifest references -- run `make generate-action-inventory`", svc.Prefix)
		}
	}
}

// auth0RefreshHint is the one command that refreshes the Auth0 snapshot
// alone; it needs the tenant credential the fetcher's Auth0 arm names.
const auth0RefreshHint = "run `make generate-action-inventory ARMS=auth0` with AUTH0_DOMAIN, AUTH0_CLIENT_ID, and AUTH0_CLIENT_SECRET set"

// TestAuth0ScopesExist is the Auth0 arm of the gate: every Management API
// scope every committed permissions manifest names must exist in the
// tenant's own Management API definition. An Auth0 scope is
// "verb:resource", so the snapshot is keyed by the resource and the verb
// is matched exactly -- Auth0 grants no wildcards. The structural regex in
// pkg/iac/permissions would accept "update:connection_options" (one letter
// short of the real resource); this gate refuses it, which is the
// difference between a manifest that installs and a grant Auth0 refuses on
// the first write.
func TestAuth0ScopesExist(t *testing.T) {
	root := repoRoot(t)
	inv, err := LoadAuth0(packageDir(t))
	if err != nil {
		t.Fatalf("loading inventory: %v", err)
	}

	discovered, err := permissions.Discover(root)
	if err != nil {
		t.Fatalf("discovering permissions manifests: %v", err)
	}

	referenced := map[string]bool{}
	for provider, components := range discovered {
		for _, component := range components {
			manifest, err := permissions.Load(root, provider, component)
			if err != nil {
				t.Fatalf("loading %s/%s: %v", provider, component, err)
			}
			for _, group := range manifest.GetSpec().GetAuth0().GetGroups() {
				for _, scope := range group.GetScopes() {
					verb, resource, found := strings.Cut(scope, ":")
					if !found {
						t.Errorf("%s/%s: scope %q has no resource segment", provider, component, scope)
						continue
					}
					referenced[resource] = true
					published := inv.ServiceActions(resource)
					if published == nil {
						t.Errorf("%s/%s: scope %q names resource %q, which the tenant's Management API definition does not list -- the resource is invented or misspelled (if Auth0 added it since the snapshot, %s)", provider, component, scope, resource, auth0RefreshHint)
						continue
					}
					if !containsExact(published, verb) {
						t.Errorf("%s/%s: scope %q does not exist in the tenant's Management API definition (the verbs for %q are %s) -- the name is invented or misspelled", provider, component, scope, resource, strings.Join(published, ", "))
					}
				}
			}
		}
	}

	// The dead-weight rule, as on every arm.
	for _, svc := range inv.Services {
		if !referenced[svc.Prefix] {
			t.Errorf("inventory covers resource %q which no permissions manifest references -- %s", svc.Prefix, auth0RefreshHint)
		}
	}
}

// TestCloudflareGroupsExist is the Cloudflare arm of the gate: every
// (permission-group name, scope) pair every committed permissions
// manifest names must exist in Cloudflare's own permission-group
// inventory. Names are matched exactly (they are not IAM-evaluated
// patterns), and the two failure classes are named distinctly: a name the
// provider never defined is invented, while a real name at the wrong
// scope is a modeling error -- a token policy carrying it would grant
// nothing at the intended level.
func TestCloudflareGroupsExist(t *testing.T) {
	root := repoRoot(t)
	inv, err := LoadCloudflare(packageDir(t))
	if err != nil {
		t.Fatalf("loading inventory: %v", err)
	}

	discovered, err := permissions.Discover(root)
	if err != nil {
		t.Fatalf("discovering permissions manifests: %v", err)
	}

	referenced := map[string]bool{}
	for provider, components := range discovered {
		for _, component := range components {
			manifest, err := permissions.Load(root, provider, component)
			if err != nil {
				t.Fatalf("loading %s/%s: %v", provider, component, err)
			}
			for _, group := range manifest.GetSpec().GetCloudflare().GetGroups() {
				name, scope := group.GetName(), group.GetScope()
				referenced[name] = true
				if inv.HasGroup(name, scope) {
					continue
				}
				if published := inv.GroupScopes(name); len(published) > 0 {
					t.Errorf("%s/%s: permission group %q exists but not at scope %q (Cloudflare defines it at %v) -- a token policy carrying it here would grant nothing at the intended level", provider, component, name, scope, published)
					continue
				}
				t.Errorf("%s/%s: permission group %q does not exist in Cloudflare's inventory -- the name is invented, misspelled, or renamed by the provider (run `make generate-action-inventory`)", provider, component, name)
			}
		}
	}

	// The dead-weight rule, keyed by group name: the fetcher scopes the
	// snapshot to referenced names, so an unreferenced name means the
	// manifests moved and the snapshot did not.
	seen := map[string]bool{}
	for _, group := range inv.Groups {
		if seen[group.Name] {
			continue
		}
		seen[group.Name] = true
		if !referenced[group.Name] {
			t.Errorf("inventory covers permission group %q which no permissions manifest references -- run `make generate-action-inventory`", group.Name)
		}
	}
}

// containsExact reports exact string membership -- the matching semantics
// of providers that do not evaluate wildcards.
func containsExact(published []string, name string) bool {
	for _, entry := range published {
		if entry == name {
			return true
		}
	}
	return false
}

// checkAzurePlane holds one manifest field's operations to its plane of
// the namespace inventory, naming a wrong-plane operation distinctly from
// a nonexistent one.
func checkAzurePlane(t *testing.T, inv *Inventory, referenced map[string]bool, provider, component, plane string, operations []string, dataPlane bool) {
	t.Helper()
	for _, operation := range operations {
		namespace, name, found := strings.Cut(operation, "/")
		if !found {
			t.Errorf("%s/%s: azure %s entry %q has no namespace segment", provider, component, plane, operation)
			continue
		}
		referenced[namespace] = true
		svc := inv.Lookup(namespace)
		if svc == nil {
			t.Errorf("%s/%s: %s entry %q names namespace %q which the inventory snapshot does not cover -- run `make generate-action-inventory`", provider, component, plane, operation, namespace)
			continue
		}
		own, other := svc.Actions, svc.DataActions
		if dataPlane {
			own, other = svc.DataActions, svc.Actions
		}
		if MatchAction(own, name) > 0 {
			continue
		}
		if MatchAction(other, name) > 0 {
			t.Errorf("%s/%s: %s entry %q exists on the OTHER plane -- a role definition carrying it here would grant nothing; move it to the right field", provider, component, plane, operation)
			continue
		}
		t.Errorf("%s/%s: %s entry %q does not exist in ARM's provider operations for %q -- the name is invented or misspelled", provider, component, plane, operation, namespace)
	}
}

// TestMatchAction pins the matching semantics the gate stands on.
func TestMatchAction(t *testing.T) {
	published := []string{"CreateFunction", "DeleteFunction", "GetFunction", "GetFunctionConfiguration", "ListFunctions", "TagResource"}
	cases := []struct {
		name    string
		matches int
	}{
		{"CreateFunction", 1},
		{"createfunction", 1}, // IAM evaluates action names case-insensitively
		{"Get*", 2},           // wildcard expands against published names
		{"GetFunction?onfiguration", 1},
		{"*", 6},
		{"List*", 1},
		{"DeleteFunctionScalingConfig", 0}, // the invented-name class the gate exists for
		{"Create*Url", 0},                  // a wildcard matching nothing is still a fabrication
	}
	for _, c := range cases {
		if got := MatchAction(published, c.name); got != c.matches {
			t.Errorf("MatchAction(%q) = %d, want %d", c.name, got, c.matches)
		}
	}
}

// TestRenderLoadRoundTrip proves the renderer and the strict loader agree:
// what Render writes, LoadAws accepts, byte-stably.
func TestRenderLoadRoundTrip(t *testing.T) {
	inv := &Inventory{
		Provider: "aws",
		Services: []Service{
			{
				Prefix:             "lambda",
				SourceURL:          "https://servicereference.us-east-1.amazonaws.com/v1/lambda/lambda.json",
				SourceModified:     "2026-08-01",
				RetrievedOn:        "2026-08-15",
				Actions:            []string{"CreateFunction", "DeleteFunction", "ListFunctions"},
				NonScopableActions: []string{"ListFunctions"},
			},
			{
				Prefix:         "s3",
				SourceURL:      "https://servicereference.us-east-1.amazonaws.com/v1/s3/s3.json",
				SourceModified: "2026-08-02",
				RetrievedOn:    "2026-08-15",
				Actions:        []string{"GetObject", "PutObject"},
			},
		},
	}
	dir := t.TempDir()
	rendered := Render(inv)
	if err := writeFile(t, filepath.Join(dir, AwsFileName), rendered); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadAws(dir)
	if err != nil {
		t.Fatalf("round-trip load: %v", err)
	}
	if Render(loaded) != rendered {
		t.Error("render -> load -> render is not byte-stable")
	}
}

// TestLoadAwsRefusals pins the loader's structural invariants: unsorted or
// duplicated content refuses loudly rather than letting gate results
// depend on snapshot ordering accidents.
func TestLoadAwsRefusals(t *testing.T) {
	base := func() *Inventory {
		return &Inventory{
			Provider: "aws",
			Services: []Service{{
				Prefix:         "lambda",
				SourceURL:      "https://example.invalid/lambda.json",
				SourceModified: "2026-08-01",
				RetrievedOn:    "2026-08-15",
				Actions:        []string{"CreateFunction", "DeleteFunction"},
			}},
		}
	}
	cases := []struct {
		name    string
		mutate  func(*Inventory)
		wantErr string
	}{
		{"wrong provider", func(i *Inventory) { i.Provider = "gcp" }, "provider"},
		{"unsorted actions", func(i *Inventory) { i.Services[0].Actions = []string{"DeleteFunction", "CreateFunction"} }, "not sorted"},
		{"duplicate action", func(i *Inventory) { i.Services[0].Actions = []string{"CreateFunction", "CreateFunction"} }, "duplicates"},
		{"empty actions", func(i *Inventory) { i.Services[0].Actions = nil }, "no actions"},
		{"missing provenance", func(i *Inventory) { i.Services[0].RetrievedOn = "" }, "required"},
		{"non-scopable entry not published", func(i *Inventory) { i.Services[0].NonScopableActions = []string{"FrobnicateFunction"} }, "not in actions"},
		{"unsorted non-scopable actions", func(i *Inventory) {
			i.Services[0].NonScopableActions = []string{"DeleteFunction", "CreateFunction"}
		}, "not sorted"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			inv := base()
			c.mutate(inv)
			dir := t.TempDir()
			// Render sorts, so unsorted/duplicate cases write raw YAML to
			// exercise the loader against genuinely malformed bytes.
			raw := renderRaw(inv)
			if err := writeFile(t, filepath.Join(dir, AwsFileName), raw); err != nil {
				t.Fatal(err)
			}
			_, err := LoadAws(dir)
			if err == nil || !strings.Contains(err.Error(), c.wantErr) {
				t.Errorf("LoadAws error = %v, want it to contain %q", err, c.wantErr)
			}
		})
	}
}

// TestRenderLoadRoundTripAzure proves the renderer and the strict loader
// agree on the Azure shape too: split planes, no source_modified (ARM
// publishes none), byte-stably.
func TestRenderLoadRoundTripAzure(t *testing.T) {
	inv := &Inventory{
		Provider: "azure",
		Services: []Service{
			{
				Prefix:      "Microsoft.Network",
				SourceURL:   "https://management.azure.com/providers/Microsoft.Authorization/providerOperations/Microsoft.Network?api-version=2022-04-01&$expand=resourceTypes",
				RetrievedOn: "2026-08-16",
				Actions:     []string{"dnsZones/read", "dnsZones/write"},
			},
			{
				Prefix:      "Microsoft.Storage",
				SourceURL:   "https://management.azure.com/providers/Microsoft.Authorization/providerOperations/Microsoft.Storage?api-version=2022-04-01&$expand=resourceTypes",
				RetrievedOn: "2026-08-16",
				Actions:     []string{"storageAccounts/read"},
				DataActions: []string{"storageAccounts/blobServices/containers/blobs/read"},
			},
		},
	}
	dir := t.TempDir()
	rendered := Render(inv)
	if err := writeFile(t, filepath.Join(dir, AzureFileName), rendered); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadAzure(dir)
	if err != nil {
		t.Fatalf("round-trip load: %v", err)
	}
	if Render(loaded) != rendered {
		t.Error("render -> load -> render is not byte-stable")
	}
}

// TestLoadAzureRefusals pins the Azure loader's structural invariants,
// including the plane lists' own ordering rules and the
// at-least-one-plane requirement.
func TestLoadAzureRefusals(t *testing.T) {
	base := func() *Inventory {
		return &Inventory{
			Provider: "azure",
			Services: []Service{{
				Prefix:      "Microsoft.Network",
				SourceURL:   "https://example.invalid/providerOperations",
				RetrievedOn: "2026-08-16",
				Actions:     []string{"dnsZones/read", "dnsZones/write"},
				DataActions: []string{"someService/data/read"},
			}},
		}
	}
	cases := []struct {
		name    string
		mutate  func(*Inventory)
		wantErr string
	}{
		{"wrong provider", func(i *Inventory) { i.Provider = "aws" }, "provider"},
		{"unsorted data actions", func(i *Inventory) { i.Services[0].DataActions = []string{"b/read", "a/read"} }, "not sorted"},
		{"duplicate data action", func(i *Inventory) { i.Services[0].DataActions = []string{"a/read", "a/read"} }, "duplicates"},
		{"both planes empty", func(i *Inventory) { i.Services[0].Actions = nil; i.Services[0].DataActions = nil }, "no actions"},
		{"missing provenance", func(i *Inventory) { i.Services[0].RetrievedOn = "" }, "required"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			inv := base()
			c.mutate(inv)
			dir := t.TempDir()
			raw := renderRaw(inv)
			if err := writeFile(t, filepath.Join(dir, AzureFileName), raw); err != nil {
				t.Fatal(err)
			}
			_, err := LoadAzure(dir)
			if err == nil || !strings.Contains(err.Error(), c.wantErr) {
				t.Errorf("LoadAzure error = %v, want it to contain %q", err, c.wantErr)
			}
		})
	}
}

// TestRenderLoadRoundTripGroups proves the group-shape renderer and its
// strict loader agree: what RenderGroups writes, LoadCloudflare accepts,
// byte-stably -- including names whose ": " would change meaning emitted
// unquoted (the yamlemit contract).
func TestRenderLoadRoundTripGroups(t *testing.T) {
	inv := &GroupInventory{
		Provider:    "cloudflare",
		SourceURL:   "https://api.cloudflare.com/client/v4/accounts/{account_id}/tokens/permission_groups",
		RetrievedOn: "2026-08-22",
		Groups: []PermissionGroup{
			{ID: "1ba6ab4cacdb454b913bbb93e1b8cb8c", Name: "Access: Apps and Policies Read", Scopes: []string{"com.cloudflare.api.account"}},
			{ID: "82e64a83756745bbbb1c9c2701bf816b", Name: "DNS Read", Scopes: []string{"com.cloudflare.api.account.zone"}},
			{ID: "4755a26eedb94da69e1066d98aa820be", Name: "DNS Write", Scopes: []string{"com.cloudflare.api.account.zone"}},
		},
	}
	dir := t.TempDir()
	rendered := RenderGroups(inv)
	if err := writeFile(t, filepath.Join(dir, CloudflareFileName), rendered); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadCloudflare(dir)
	if err != nil {
		t.Fatalf("round-trip load: %v", err)
	}
	if RenderGroups(loaded) != rendered {
		t.Error("render -> load -> render is not byte-stable")
	}
	if !loaded.HasGroup("Access: Apps and Policies Read", "com.cloudflare.api.account") {
		t.Error("the quoted colon-carrying name did not survive the round trip")
	}
}

// TestLoadCloudflareRefusals pins the group loader's structural
// invariants: unsorted or duplicated content refuses loudly rather than
// letting gate results depend on snapshot ordering accidents.
func TestLoadCloudflareRefusals(t *testing.T) {
	base := func() *GroupInventory {
		return &GroupInventory{
			Provider:    "cloudflare",
			SourceURL:   "https://example.invalid/permission_groups",
			RetrievedOn: "2026-08-22",
			Groups: []PermissionGroup{
				{ID: "aaa", Name: "DNS Read", Scopes: []string{"com.cloudflare.api.account.zone"}},
				{ID: "bbb", Name: "DNS Write", Scopes: []string{"com.cloudflare.api.account.zone"}},
			},
		}
	}
	cases := []struct {
		name    string
		mutate  func(*GroupInventory)
		wantErr string
	}{
		{"wrong provider", func(i *GroupInventory) { i.Provider = "aws" }, "provider"},
		{"missing provenance", func(i *GroupInventory) { i.RetrievedOn = "" }, "required"},
		{"no groups", func(i *GroupInventory) { i.Groups = nil }, "no groups"},
		{"duplicate id", func(i *GroupInventory) { i.Groups[1].ID = "aaa" }, "duplicate group id"},
		{"unsorted groups", func(i *GroupInventory) { i.Groups[0], i.Groups[1] = i.Groups[1], i.Groups[0] }, "not sorted"},
		{"empty scopes", func(i *GroupInventory) { i.Groups[0].Scopes = nil }, "no scopes"},
		{"unsorted scopes", func(i *GroupInventory) {
			i.Groups[0].Scopes = []string{"com.cloudflare.api.account.zone", "com.cloudflare.api.account"}
		}, "not sorted"},
		{"duplicate scope", func(i *GroupInventory) {
			i.Groups[0].Scopes = []string{"com.cloudflare.api.account", "com.cloudflare.api.account"}
		}, "duplicates scope"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			inv := base()
			c.mutate(inv)
			dir := t.TempDir()
			// Render sorts, so unsorted/duplicate cases write raw YAML to
			// exercise the loader against genuinely malformed bytes.
			raw := renderGroupsRaw(inv)
			if err := writeFile(t, filepath.Join(dir, CloudflareFileName), raw); err != nil {
				t.Fatal(err)
			}
			_, err := LoadCloudflare(dir)
			if err == nil || !strings.Contains(err.Error(), c.wantErr) {
				t.Errorf("LoadCloudflare error = %v, want it to contain %q", err, c.wantErr)
			}
		})
	}
}

// TestLoadDigitalOceanRefusals pins the DigitalOcean loader's identity
// invariant on the shared Service shape (the structural invariants
// themselves are pinned by the AWS and Azure refusal suites -- one loader
// serves all three).
func TestLoadDigitalOceanRefusals(t *testing.T) {
	inv := &Inventory{
		Provider: "aws",
		Services: []Service{{
			Prefix:      "database",
			SourceURL:   "https://example.invalid/scopes",
			RetrievedOn: "2026-08-22",
			Actions:     []string{"create", "read"},
		}},
	}
	dir := t.TempDir()
	if err := writeFile(t, filepath.Join(dir, DigitalOceanFileName), renderRaw(inv)); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadDigitalOcean(dir); err == nil || !strings.Contains(err.Error(), "provider") {
		t.Errorf("LoadDigitalOcean error = %v, want a provider identity refusal", err)
	}
}

// TestLoadAuth0Refusals pins the Auth0 loader's identity invariant on the
// shared Service shape: a DigitalOcean snapshot saved under Auth0's name
// has the same prefix-and-verbs look and must still be refused.
func TestLoadAuth0Refusals(t *testing.T) {
	inv := &Inventory{
		Provider: "digitalocean",
		Services: []Service{{
			Prefix:      "connections",
			SourceURL:   "https://example.invalid/definition",
			RetrievedOn: "2026-09-23",
			Actions:     []string{"create", "read"},
		}},
	}
	dir := t.TempDir()
	if err := writeFile(t, filepath.Join(dir, Auth0FileName), renderRaw(inv)); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadAuth0(dir); err == nil || !strings.Contains(err.Error(), "provider") {
		t.Errorf("LoadAuth0 error = %v, want a provider identity refusal", err)
	}
}

// renderGroupsRaw writes a group inventory verbatim, WITHOUT
// RenderGroups's sorting, so refusal tests can present genuinely
// malformed snapshots. Quoting rides the same yamlemit rules as the real
// renderer.
func renderGroupsRaw(inv *GroupInventory) string {
	var b strings.Builder
	b.WriteString("provider: " + inv.Provider + "\n")
	b.WriteString("source_url: " + inv.SourceURL + "\n")
	if inv.RetrievedOn != "" {
		b.WriteString("retrieved_on: \"" + inv.RetrievedOn + "\"\n")
	}
	if len(inv.Groups) == 0 {
		b.WriteString("groups: []\n")
		return b.String()
	}
	b.WriteString("groups:\n")
	for _, group := range inv.Groups {
		b.WriteString("  - id: " + group.ID + "\n")
		b.WriteString("    name: \"" + group.Name + "\"\n")
		if len(group.Scopes) == 0 {
			b.WriteString("    scopes: []\n")
			continue
		}
		b.WriteString("    scopes:\n")
		for _, scope := range group.Scopes {
			b.WriteString("      - " + scope + "\n")
		}
	}
	return b.String()
}

// renderRaw writes an inventory verbatim, WITHOUT Render's sorting, so
// refusal tests can present genuinely malformed snapshots.
func renderRaw(inv *Inventory) string {
	var b strings.Builder
	b.WriteString("provider: " + inv.Provider + "\n")
	b.WriteString("services:\n")
	for _, svc := range inv.Services {
		b.WriteString("  - prefix: " + svc.Prefix + "\n")
		b.WriteString("    source_url: " + svc.SourceURL + "\n")
		if svc.SourceModified != "" {
			b.WriteString("    source_modified: \"" + svc.SourceModified + "\"\n")
		}
		b.WriteString("    retrieved_on: \"" + svc.RetrievedOn + "\"\n")
		if len(svc.Actions) == 0 && len(svc.DataActions) == 0 {
			b.WriteString("    actions: []\n")
			continue
		}
		if len(svc.Actions) > 0 {
			b.WriteString("    actions:\n")
			for _, action := range svc.Actions {
				b.WriteString("      - " + action + "\n")
			}
		}
		if len(svc.NonScopableActions) > 0 {
			b.WriteString("    non_scopable_actions:\n")
			for _, action := range svc.NonScopableActions {
				b.WriteString("      - " + action + "\n")
			}
		}
		if len(svc.DataActions) > 0 {
			b.WriteString("    data_actions:\n")
			for _, action := range svc.DataActions {
				b.WriteString("      - " + action + "\n")
			}
		}
	}
	return b.String()
}

func writeFile(t *testing.T, path, content string) error {
	t.Helper()
	return os.WriteFile(path, []byte(content), 0o644)
}

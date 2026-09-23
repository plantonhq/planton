package permissions

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"

	permissionsv1 "github.com/plantonhq/planton/iac/componentpermissions/v1"
)

var (
	// awsActionPattern is IAM's "service:Action" spelling. Wildcards are
	// syntactically legal; the gate separately demands a defense in notes.
	awsActionPattern = regexp.MustCompile(`^[a-z0-9-]+:[A-Za-z0-9*]+$`)
	// gcpPermissionPattern is GCP IAM's dotted permission form,
	// e.g. "container.clusters.create" (3+ segments).
	gcpPermissionPattern = regexp.MustCompile(`^[a-z][a-zA-Z0-9]*(\.[a-zA-Z0-9]+){2,}$`)
	// azureActionPattern is Azure RBAC's "Provider.Namespace/type/action"
	// form, e.g. "Microsoft.ContainerService/managedClusters/write".
	azureActionPattern = regexp.MustCompile(`^[A-Za-z]+\.[A-Za-z]+(/[A-Za-z0-9*]+)+$`)
	// kubernetesVerbs is the closed RBAC verb vocabulary. "*" is absent
	// deliberately -- a wildcard verb is never least privilege. escalate
	// (on roles and clusterroles) and bind (on their bindings) are the
	// verbs Kubernetes defines for a principal that grants permissions it
	// does not itself hold -- what installing an operator IS -- so a kind
	// that installs one can state that exactly instead of being run as
	// cluster-admin in practice and described as least privilege on paper.
	kubernetesVerbs = map[string]bool{
		"get": true, "list": true, "watch": true, "create": true,
		"update": true, "patch": true, "delete": true, "deletecollection": true,
		"escalate": true, "bind": true,
	}
	// cloudflareScopePattern is Cloudflare's scope identifier spelling,
	// e.g. "com.cloudflare.api.account.zone". The spelling is checked
	// here; EXISTENCE of the (name, scope) pair is the inventory gate's
	// job (pkg/iac/actioninventory) -- Cloudflare grows the scope
	// vocabulary, so the snapshot, never a regex, is the closed set.
	cloudflareScopePattern = regexp.MustCompile(`^com\.cloudflare\.[a-z0-9]+(\.[a-z0-9]+)*$`)
	// digitalOceanScopePattern is DigitalOcean's "resource:action" token
	// scope spelling. Underscores are legal in BOTH segments, and action
	// segments go beyond CRUD (view_credentials, access_cluster, admin) --
	// a closed verb vocabulary here would wrongly reject real published
	// scopes, so the spelling is checked here and existence against the
	// provider's own inventory is the real check (pkg/iac/actioninventory).
	digitalOceanScopePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*:[a-z][a-z0-9_]*$`)
	// digitalOceanAliasScopes are the global alias scopes. Each expands to
	// every current AND future endpoint, so neither can ever appear in a
	// least-privilege manifest -- the gate refuses them outright.
	digitalOceanAliasScopes = map[string]bool{"api:read": true, "api:write": true}
	// digitalOceanSpacesPermissions is the CLOSED grant-level vocabulary of
	// DigitalOcean's Spaces keys, verified against the provider's own
	// Spaces-keys API reference (a grant is {bucket, permission} with
	// exactly these levels). Three values from the provider's API contract
	// -- no machine inventory exists to snapshot, so this closed set IS
	// the existence check, stated here rather than silently absent from
	// the inventory gates.
	digitalOceanSpacesPermissions = map[string]bool{"read": true, "readwrite": true, "fullaccess": true}
	// auth0ScopePattern is Auth0's Management API "verb:resource" scope
	// spelling -- the verb FIRST, the reverse of DigitalOcean's. Verbs go
	// beyond CRUD ("blacklist:tokens"), so the verb is not a closed set
	// here; existence against the tenant's own Management API definition
	// is the real check (pkg/iac/actioninventory).
	auth0ScopePattern = regexp.MustCompile(`^[a-z]+:[a-z_]+$`)
	// tokenScopedProviders are catalog providers whose modules authenticate
	// with a bearer credential that offers NO finer-grained vocabulary --
	// so, per the schema's own absence semantics, their manifests may
	// legally carry no provider section. openfga is the one tenant: a
	// pre-shared key grants the server's entire API, so there is nothing
	// a manifest could truthfully declare. The exemption is the honest
	// model, not a schema gap.
	//
	// Cloudflare, DigitalOcean, and Auth0 left this map when their arms
	// landed: their manifests declare token permission groups / scopes as
	// first-class sections, held to the providers' own inventories.
	tokenScopedProviders = map[string]bool{
		"openfga": true,
	}
)

// TestPermissionsConformance holds every authored permissions manifest to
// its contract, offline:
//
//  1. The manifest parses strictly against its proto schema and names its
//     component (metadata.name equals the component directory).
//  2. It declares at least one provider section -- a permissions file that
//     grants nothing describes no module.
//  3. Every entry is structurally sound for its provider (action spelling,
//     RBAC verb vocabulary), carries provenance, and defends its trust
//     posture: derived entries cite the module resources they cover, and
//     any wildcard resource scope carries a defense in notes.
//
// What this gate deliberately CANNOT prove: that the actions are SUFFICIENT
// or MINIMAL against a live cloud. That proof is the capture harness's job
// (e2e running under a role built from this very manifest); until an entry
// graduates to proven, its provenance says so honestly.
func TestPermissionsConformance(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "catalog")); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test and the lint.catalog-data lane")
	}

	discovered, err := Discover(root)
	if err != nil {
		t.Fatalf("discovering permissions manifests: %v", err)
	}
	if len(discovered) == 0 {
		t.Skip("no permissions manifests authored yet")
	}

	for provider, components := range discovered {
		for _, component := range components {
			component := component
			t.Run(provider+"/"+component, func(t *testing.T) {
				manifest, err := Load(root, provider, component)
				if err != nil {
					t.Fatalf("permissions manifest: %v", err)
				}
				if manifest.GetKind() != "ComponentPermissions" {
					t.Fatalf("kind is %q, want ComponentPermissions", manifest.GetKind())
				}
				if manifest.GetMetadata().GetName() != component {
					t.Errorf("metadata.name is %q, want %q", manifest.GetMetadata().GetName(), component)
				}

				spec := manifest.GetSpec()
				// Counted from the schema's own populated fields, never a
				// hand-written list of sections: a provider section added
				// to the schema counts here the day it lands.
				declared := 0
				spec.ProtoReflect().Range(func(protoreflect.FieldDescriptor, protoreflect.Value) bool {
					declared++
					return true
				})
				if declared == 0 && !tokenScopedProviders[provider] {
					t.Fatal("manifest declares no provider section -- a permissions file that grants nothing describes no module")
				}

				checkAws(t, spec.GetAws())
				checkGcp(t, spec.GetGcp())
				checkAzure(t, spec.GetAzure())
				checkKubernetes(t, spec.GetKubernetes())
				checkCloudflare(t, spec.GetCloudflare())
				checkDigitalOcean(t, spec.GetDigitalOcean())
				checkAuth0(t, spec.GetAuth0())
			})
		}
	}
}

func checkProvenance(t *testing.T, where string, provenance permissionsv1.Provenance, notes string) {
	t.Helper()
	switch provenance {
	case permissionsv1.Provenance_provenance_unspecified:
		t.Errorf("%s: provenance is unspecified -- derived and proven are never blurred", where)
	case permissionsv1.Provenance_derived:
		if strings.TrimSpace(notes) == "" {
			t.Errorf("%s: derived entry cites no module resources in notes -- a derivation without its source is unverifiable", where)
		}
	}
}

func checkAws(t *testing.T, aws *permissionsv1.AwsPermissions) {
	t.Helper()
	if aws == nil {
		return
	}
	if len(aws.GetStatements()) == 0 {
		t.Error("aws section is present but declares no statements")
	}
	sids := map[string]bool{}
	for _, statement := range aws.GetStatements() {
		sid := statement.GetSid()
		if sid == "" {
			t.Error("aws statement with empty sid")
			continue
		}
		if sids[sid] {
			t.Errorf("aws: duplicate sid %q", sid)
		}
		sids[sid] = true
		if len(statement.GetActions()) == 0 {
			t.Errorf("aws %s: no actions", sid)
		}
		for _, action := range statement.GetActions() {
			if !awsActionPattern.MatchString(action) {
				t.Errorf("aws %s: action %q is not service:Action spelling", sid, action)
			}
		}
		if len(statement.GetResources()) == 0 {
			t.Errorf("aws %s: no resources -- scope the statement or defend '*'", sid)
		}
		for _, resource := range statement.GetResources() {
			if resource == "*" && strings.TrimSpace(statement.GetNotes()) == "" {
				t.Errorf("aws %s: resource '*' without a defense in notes -- least privilege demands the reason", sid)
			}
		}
		checkProvenance(t, "aws "+sid, statement.GetProvenance(), statement.GetNotes())
	}
}

func checkGcp(t *testing.T, gcp *permissionsv1.GcpPermissions) {
	t.Helper()
	if gcp == nil {
		return
	}
	if len(gcp.GetGroups()) == 0 {
		t.Error("gcp section is present but declares no groups")
	}
	for _, group := range gcp.GetGroups() {
		if strings.TrimSpace(group.GetPurpose()) == "" {
			t.Error("gcp group with empty purpose")
		}
		if len(group.GetPermissions()) == 0 {
			t.Errorf("gcp %s: no permissions", group.GetPurpose())
		}
		for _, permission := range group.GetPermissions() {
			if !gcpPermissionPattern.MatchString(permission) {
				t.Errorf("gcp %s: permission %q is not dotted IAM form", group.GetPurpose(), permission)
			}
		}
		checkProvenance(t, "gcp "+group.GetPurpose(), group.GetProvenance(), group.GetNotes())
	}
}

func checkAzure(t *testing.T, azure *permissionsv1.AzurePermissions) {
	t.Helper()
	if azure == nil {
		return
	}
	if len(azure.GetGroups()) == 0 {
		t.Error("azure section is present but declares no groups")
	}
	for _, group := range azure.GetGroups() {
		if strings.TrimSpace(group.GetPurpose()) == "" {
			t.Error("azure group with empty purpose")
		}
		if len(group.GetActions()) == 0 && len(group.GetDataActions()) == 0 {
			t.Errorf("azure %s: no actions or data_actions", group.GetPurpose())
		}
		for _, action := range append(append([]string{}, group.GetActions()...), group.GetDataActions()...) {
			if !azureActionPattern.MatchString(action) {
				t.Errorf("azure %s: action %q is not Provider.Namespace/type/action form", group.GetPurpose(), action)
			}
		}
		checkProvenance(t, "azure "+group.GetPurpose(), group.GetProvenance(), group.GetNotes())
	}
}

func checkKubernetes(t *testing.T, kubernetes *permissionsv1.KubernetesPermissions) {
	t.Helper()
	if kubernetes == nil {
		return
	}
	if len(kubernetes.GetRules()) == 0 {
		t.Error("kubernetes section is present but declares no rules")
	}
	for i, rule := range kubernetes.GetRules() {
		if len(rule.GetResources()) == 0 {
			t.Errorf("kubernetes rule %d: no resources", i)
		}
		if len(rule.GetVerbs()) == 0 {
			t.Errorf("kubernetes rule %d: no verbs", i)
		}
		for _, verb := range rule.GetVerbs() {
			if !kubernetesVerbs[verb] {
				t.Errorf("kubernetes rule %d: verb %q is not in the RBAC verb vocabulary (wildcards are never least privilege)", i, verb)
			}
		}
		checkProvenance(t, "kubernetes rule "+ruleLabel(rule), rule.GetProvenance(), rule.GetNotes())
	}
}

func ruleLabel(rule *permissionsv1.KubernetesRule) string {
	return strings.Join(rule.GetResources(), ",")
}

func checkCloudflare(t *testing.T, cloudflare *permissionsv1.CloudflarePermissions) {
	t.Helper()
	if cloudflare == nil {
		return
	}
	if len(cloudflare.GetGroups()) == 0 {
		t.Error("cloudflare section is present but declares no groups")
	}
	for _, group := range cloudflare.GetGroups() {
		if strings.TrimSpace(group.GetPurpose()) == "" {
			t.Error("cloudflare group with empty purpose")
		}
		if group.GetName() == "" {
			t.Errorf("cloudflare %s: no permission-group name", group.GetPurpose())
		}
		if !cloudflareScopePattern.MatchString(group.GetScope()) {
			t.Errorf("cloudflare %s: scope %q is not Cloudflare's scope identifier spelling (e.g. com.cloudflare.api.account.zone)", group.GetPurpose(), group.GetScope())
		}
		checkProvenance(t, "cloudflare "+group.GetPurpose(), group.GetProvenance(), group.GetNotes())
	}
}

func checkDigitalOcean(t *testing.T, digitalOcean *permissionsv1.DigitalOceanPermissions) {
	t.Helper()
	if digitalOcean == nil {
		return
	}
	if len(digitalOcean.GetGroups()) == 0 && len(digitalOcean.GetSpacesGrants()) == 0 {
		t.Error("digitalocean section is present but declares no groups and no spaces grants")
	}
	for _, grant := range digitalOcean.GetSpacesGrants() {
		if strings.TrimSpace(grant.GetPurpose()) == "" {
			t.Error("digitalocean spaces grant with empty purpose")
		}
		if !digitalOceanSpacesPermissions[grant.GetPermission()] {
			t.Errorf("digitalocean spaces %s: permission %q is not a Spaces key grant level (read, readwrite, or fullaccess)", grant.GetPurpose(), grant.GetPermission())
		}
		checkProvenance(t, "digitalocean spaces "+grant.GetPurpose(), grant.GetProvenance(), grant.GetNotes())
	}
	for _, group := range digitalOcean.GetGroups() {
		if strings.TrimSpace(group.GetPurpose()) == "" {
			t.Error("digitalocean group with empty purpose")
		}
		if len(group.GetScopes()) == 0 {
			t.Errorf("digitalocean %s: no scopes", group.GetPurpose())
		}
		for _, scope := range group.GetScopes() {
			if digitalOceanAliasScopes[scope] {
				t.Errorf("digitalocean %s: scope %q is a global alias that expands to every current and future endpoint -- never least privilege; declare the resource scopes the modules actually need", group.GetPurpose(), scope)
				continue
			}
			if !digitalOceanScopePattern.MatchString(scope) {
				t.Errorf("digitalocean %s: scope %q is not resource:action spelling", group.GetPurpose(), scope)
			}
		}
		checkProvenance(t, "digitalocean "+group.GetPurpose(), group.GetProvenance(), group.GetNotes())
	}
}

func checkAuth0(t *testing.T, auth0 *permissionsv1.Auth0Permissions) {
	t.Helper()
	if auth0 == nil {
		return
	}
	if len(auth0.GetGroups()) == 0 {
		t.Error("auth0 section is present but declares no groups")
	}
	purposes := map[string]bool{}
	for _, group := range auth0.GetGroups() {
		purpose := group.GetPurpose()
		if strings.TrimSpace(purpose) == "" {
			t.Error("auth0 group with empty purpose")
		}
		if purposes[purpose] {
			t.Errorf("auth0: duplicate purpose %q -- a reader tells groups apart by purpose, so one purpose is one group", purpose)
		}
		purposes[purpose] = true
		if len(group.GetScopes()) == 0 {
			t.Errorf("auth0 %s: no scopes", purpose)
		}
		for _, scope := range group.GetScopes() {
			if !auth0ScopePattern.MatchString(scope) {
				t.Errorf("auth0 %s: scope %q is not verb:resource spelling (Auth0 puts the verb first, e.g. \"update:connections\")", purpose, scope)
			}
		}
		checkProvenance(t, "auth0 "+purpose, group.GetProvenance(), group.GetNotes())
	}
}

// repoRoot walks up from this test file to the directory containing go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(thisFile)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found walking up from test file")
		}
		dir = parent
	}
}

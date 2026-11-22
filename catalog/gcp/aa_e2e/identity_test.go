package aa_e2e

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cloud.google.com/go/auth"
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/gcp/aa_e2e/verify"
	"github.com/plantonhq/planton/e2e/framework/provider"
	"github.com/plantonhq/planton/e2e/framework/runner"
	cloudbilling "google.golang.org/api/cloudbilling/v1"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/googleapi"
)

// repoRoot is the repository root these tests read the catalog from. Under
// Bazel the catalog tree is not part of a test's inputs, so they skip there,
// as the permissions conformance gate does, and run under go test.
func repoRoot(t *testing.T) string {
	t.Helper()
	if os.Getenv("TEST_WORKSPACE") != "" {
		t.Skip("repo-reading test; skipped under Bazel")
	}
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

// A scenario that sets the field a conditional group names is granted that
// group, and the role it gets is named for the group.
func TestDeclaredGrantTurnsOnTheConditionalGroupTheScenarioSets(t *testing.T) {
	root := repoRoot(t)
	scenario := filepath.Join(root, "catalog", "gcp", "gcpdatastreamstream", "e2e", "scenarios", "not-started-postgresql-to-bigquery.yaml")
	grant, err := declaredGrant(root, "gcpdatastreamstream", []string{scenario})
	if err != nil {
		t.Fatal(err)
	}
	if grant.kind != "GcpDatastreamStream" {
		t.Fatalf("kind = %q", grant.kind)
	}
	if len(grant.conditionalPurposes) != 1 {
		t.Fatalf("conditional groups = %v, want the desired_state group alone", grant.conditionalPurposes)
	}
	if !strings.HasPrefix(declaredRoleID(grant.kind, grant.conditionalPurposes), "plantonE2eDeclaredGcpDatastreamStreamC") {
		t.Fatalf("role id = %q", declaredRoleID(grant.kind, grant.conditionalPurposes))
	}
}

// A scenario that leaves the field unset is granted only the unconditional
// groups, under the kind's plain role name.
func TestDeclaredGrantLeavesOutConditionalGroupsTheScenarioDoesNotSet(t *testing.T) {
	root := repoRoot(t)
	source := filepath.Join(root, "catalog", "gcp", "gcpdatastreamstream", "e2e", "scenarios", "not-started-postgresql-to-bigquery.yaml")
	raw, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	unset := strings.Replace(string(raw), "desiredState: NOT_STARTED", "", 1)
	scenario := filepath.Join(t.TempDir(), "scenario.yaml")
	if err := os.WriteFile(scenario, []byte(unset), 0o600); err != nil {
		t.Fatal(err)
	}
	grant, err := declaredGrant(root, "gcpdatastreamstream", []string{scenario})
	if err != nil {
		t.Fatal(err)
	}
	if len(grant.conditionalPurposes) != 0 {
		t.Fatalf("conditional groups = %v, want none", grant.conditionalPurposes)
	}
	if got := declaredRoleID(grant.kind, nil); got != "plantonE2eDeclaredGcpDatastreamStream" {
		t.Fatalf("role id = %q", got)
	}
	for _, permission := range grant.target.permissions {
		if strings.HasSuffix(permission, ".start") || strings.HasSuffix(permission, ".pause") {
			t.Fatalf("the state-change group leaked in: %s", permission)
		}
	}
}

// One identity deploys both acts, so a conditional group only the second act
// turns on is granted from the start: the upgrade would otherwise be refused
// a permission the kind's manifest does list.
func TestDeclaredGrantHoldsWhatTheSecondActTurnsOn(t *testing.T) {
	root := repoRoot(t)
	second := filepath.Join(root, "catalog", "gcp", "gcpdatastreamstream", "e2e", "scenarios", "not-started-postgresql-to-bigquery.yaml")
	raw, err := os.ReadFile(second)
	if err != nil {
		t.Fatal(err)
	}
	first := filepath.Join(t.TempDir(), "first.yaml")
	if err := os.WriteFile(first, []byte(strings.Replace(string(raw), "desiredState: NOT_STARTED", "", 1)), 0o600); err != nil {
		t.Fatal(err)
	}
	alone, err := declaredGrant(root, "gcpdatastreamstream", []string{first})
	if err != nil {
		t.Fatal(err)
	}
	both, err := declaredGrant(root, "gcpdatastreamstream", []string{first, second})
	if err != nil {
		t.Fatal(err)
	}
	if len(alone.conditionalPurposes) != 0 || len(both.conditionalPurposes) != 1 {
		t.Fatalf("conditional groups: first act alone %v, both acts %v; want none, then the desired_state group", alone.conditionalPurposes, both.conditionalPurposes)
	}
}

// A lane's identity is granted in the harness's project unless the scenario
// names another test project, and a name that is no test project is refused
// before anything is minted.
func TestIdentityProjectIsATestProject(t *testing.T) {
	t.Setenv(secondProjectEnv, "second-project")
	h := NewHarness()
	h.services = &verify.Services{Project: "harness-project"}
	manifest := func(annotation string) *provider.KindTestContext {
		t.Helper()
		body := "apiVersion: gcp.planton.dev/v1alpha1\nkind: GcpBinaryAuthorizationPolicy\nmetadata:\n  name: x\n"
		if annotation != "" {
			body += "  annotations:\n    " + provider.IdentityProjectAnnotation + ": " + annotation + "\n"
		}
		path := filepath.Join(t.TempDir(), "scenario.yaml")
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		return &provider.KindTestContext{ManifestPath: path}
	}
	cases := []struct {
		annotation, want string
		refused          bool
	}{
		{"", "harness-project", false},
		{"second-project", "second-project", false},
		{`"harness-project, second-project"`, "harness-project,second-project", false},
		{"someone-elses-project", "", true},
		{`"harness-project, someone-elses-project"`, "", true},
	}
	for _, c := range cases {
		got, err := h.identityProjects(manifest(c.annotation))
		if c.refused != (err != nil) || strings.Join(got, ",") != c.want {
			t.Errorf("annotation %q: got (%q, %v), want %q refused=%v", c.annotation, got, err, c.want, c.refused)
		}
	}
	if projects := h.testProjects(); len(projects) != 2 || projects[1] != "second-project" {
		t.Errorf("test projects = %v, want the harness's then the second", projects)
	}
}

// syntheticKind writes a kind whose permissions manifest is the given
// groups into a scratch repository root beside the real inventory, so a test
// of how the harness splits a grant reads no catalog kind's live manifest.
func syntheticKind(t *testing.T, gcpBody string) (root, scenario string) {
	t.Helper()
	real := repoRoot(t)
	root = t.TempDir()
	inventory := filepath.Join(root, "pkg", "iac")
	if err := os.MkdirAll(inventory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(real, "pkg", "iac", "actioninventory"), filepath.Join(inventory, "actioninventory")); err != nil {
		t.Fatal(err)
	}
	kindDir := filepath.Join(root, "catalog", "gcp", "gcpsynthetic", "iac")
	if err := os.MkdirAll(kindDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := "apiVersion: iac.planton.dev/v1\nkind: CatalogKindPermissions\nmetadata:\n  name: gcpsynthetic\nspec:\n  gcp:\n" + gcpBody
	if err := os.WriteFile(filepath.Join(kindDir, "permissions.yaml"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	scenario = filepath.Join(root, "scenario.yaml")
	if err := os.WriteFile(scenario, []byte("apiVersion: gcp.planton.dev/v1alpha1\nkind: GcpSynthetic\nmetadata:\n  name: x\nspec:\n  region: us-central1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return root, scenario
}

// Each group is granted where it reaches: the project's custom role and
// roles, the organization's custom role and roles, the billing account's
// role. A predefined role's listed permissions are what readiness waits on
// for it, and the manifest's caller-project services ride along.
func TestDeclaredGrantPutsEachGroupWhereItReaches(t *testing.T) {
	root, scenario := syntheticKind(t, `    caller_project_services:
      - billingbudgets.googleapis.com
    groups:
      - purpose: ManageInProject
        permissions: [compute.networks.get, compute.networks.create]
        provenance: derived
        reach: target
        privilege: ordinary
        notes: n
      - purpose: ProjectRole
        permissions: [gkehub.features.create]
        provenance: derived
        reach: target
        privilege: ordinary
        predefined_role: roles/gkehub.admin
        notes: n
      - purpose: CreateUnderOrganization
        permissions: [resourcemanager.projects.create]
        provenance: derived
        reach: organization
        privilege: ordinary
        notes: n
      - purpose: OrganizationRole
        permissions: [orgpolicy.policies.create]
        provenance: derived
        reach: organization
        privilege: ordinary
        predefined_role: roles/orgpolicy.policyAdmin
        notes: n
      - purpose: Budgets
        permissions: [billing.budgets.create, billing.budgets.get]
        provenance: derived
        reach: billing_account
        privilege: ordinary
        predefined_role: roles/billing.costsManager
        notes: n
`)
	grant, err := declaredGrant(root, "gcpsynthetic", []string{scenario})
	if err != nil {
		t.Fatal(err)
	}
	want := declaredGrantSet{
		kind:                  "GcpSynthetic",
		target:                scopeGrant{permissions: []string{"compute.networks.create", "compute.networks.get"}, predefinedRoles: []string{"roles/gkehub.admin"}, rolePermissions: []string{"gkehub.features.create"}},
		organization:          scopeGrant{permissions: []string{"resourcemanager.projects.create"}, predefinedRoles: []string{"roles/orgpolicy.policyAdmin"}, rolePermissions: []string{"orgpolicy.policies.create"}},
		billingAccount:        scopeGrant{predefinedRoles: []string{"roles/billing.costsManager"}, rolePermissions: []string{"billing.budgets.create", "billing.budgets.get"}},
		callerProjectServices: []string{"billingbudgets.googleapis.com"},
	}
	if fmt.Sprintf("%+v", grant) != fmt.Sprintf("%+v", want) {
		t.Fatalf("grant\n got %+v\nwant %+v", grant, want)
	}
	if got := grant.organization.awaited(); strings.Join(got, ",") != "orgpolicy.policies.create,resourcemanager.projects.create" {
		t.Fatalf("organization awaits %v", got)
	}
}

// A billing account group must name its predefined role: the harness cannot
// bind a custom role there, and says so before anything is minted.
func TestDeclaredGrantRefusesABillingAccountGroupWithoutARole(t *testing.T) {
	root, scenario := syntheticKind(t, `    groups:
      - purpose: Budgets
        permissions: [billing.budgets.create]
        provenance: derived
        reach: billing_account
        privilege: ordinary
        notes: n
`)
	_, err := declaredGrant(root, "gcpsynthetic", []string{scenario})
	if err == nil || !strings.Contains(err.Error(), `"Budgets"`) || !strings.Contains(err.Error(), "predefined role") {
		t.Fatalf("err = %v, want a refusal naming the group and the missing predefined role", err)
	}
}

// A grant above the project needs its scope's variable; without it the lane
// fails naming the variable, and with it the scope is the variable's
// resource, with or without its collection prefix.
func TestGrantScopesNameTheMissingVariable(t *testing.T) {
	h := NewHarness()
	h.services = &verify.Services{Project: "harness-project"}
	grant := declaredGrantSet{
		kind:           "GcpSynthetic",
		target:         scopeGrant{permissions: []string{"compute.networks.get"}},
		organization:   scopeGrant{predefinedRoles: []string{"roles/orgpolicy.policyAdmin"}},
		billingAccount: scopeGrant{predefinedRoles: []string{"roles/billing.costsManager"}},
	}
	t.Setenv(organizationEnv, "")
	t.Setenv(billingAccountEnv, "012345-6789AB-CDEF01")
	if _, err := h.grantScopes([]string{"harness-project"}, grant); err == nil || !strings.Contains(err.Error(), organizationEnv) {
		t.Fatalf("err = %v, want the unset %s named", err, organizationEnv)
	}
	t.Setenv(organizationEnv, "organizations/123456789012")
	t.Setenv(billingAccountEnv, "billingAccounts/012345-6789AB-CDEF01")
	scopes, err := h.grantScopes([]string{"harness-project"}, grant)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, scope := range scopes {
		names = append(names, scope.policy.name)
	}
	if got := strings.Join(names, ","); got != "projects/harness-project,organizations/123456789012,billingAccounts/012345-6789AB-CDEF01" {
		t.Fatalf("scopes = %s", got)
	}
	t.Setenv(billingAccountEnv, "")
	if _, err := h.grantScopes([]string{"harness-project"}, grant); err == nil || !strings.Contains(err.Error(), billingAccountEnv) {
		t.Fatalf("err = %v, want the unset %s named", err, billingAccountEnv)
	}
}

// A billing account's IAM policy edits through Resource Manager's type and
// back unchanged.
func TestBillingAccountPolicyConvertsBothWays(t *testing.T) {
	billing := &cloudbilling.Policy{Etag: "BwX", Version: 1, Bindings: []*cloudbilling.Binding{
		{Role: "roles/billing.costsManager", Members: []string{"serviceAccount:pe2e-0123456789@p.iam.gserviceaccount.com"}},
	}}
	var edited cloudresourcemanager.Policy
	if err := convertPolicy(billing, &edited); err != nil {
		t.Fatal(err)
	}
	if !removeMember(&edited, "serviceAccount:pe2e-0123456789@p.iam.gserviceaccount.com") {
		t.Fatal("the converted policy lost its member")
	}
	var back cloudbilling.Policy
	if err := convertPolicy(&edited, &back); err != nil {
		t.Fatal(err)
	}
	if back.Etag != "BwX" || len(back.Bindings) != 0 {
		t.Fatalf("converted back: %+v", back)
	}
}

// Every role id fits Google's 64-character limit, and every lane account id
// its 30-character limit, for the longest kind in the catalog.
func TestIdentityNamesFitGoogleLimits(t *testing.T) {
	longest := "GcpVertexAiFeatureOnlineStore"
	if id := declaredRoleID(longest, []string{"A", "B"}); len(id) > 64 {
		t.Fatalf("role id %q is %d characters", id, len(id))
	}
	tc := &provider.KindTestContext{Kind: "gcpvertexaifeatureonlinestore", Engine: "terraform", ManifestPath: "/x/storage-optimized-with-a-long-name.yaml", RunID: "0a1b2c3d"}
	id := laneAccountID(laneLabel(tc))
	if len(id) < 6 || len(id) > 30 || !strings.HasPrefix(id, laneAccountPrefix) {
		t.Fatalf("lane account id %q", id)
	}
	other := *tc
	other.Engine = "pulumi"
	if laneAccountID(laneLabel(&other)) == id {
		t.Fatal("two engines' lanes share an account id")
	}
}

// Adding is idempotent, and removing drops the member and any binding it
// leaves empty, without touching others.
func TestPolicyMemberEdits(t *testing.T) {
	policy := &cloudresourcemanager.Policy{Bindings: []*cloudresourcemanager.Binding{
		{Role: "roles/viewer", Members: []string{"user:someone@example.com"}},
	}}
	member := "serviceAccount:pe2e-0123456789@p.iam.gserviceaccount.com"
	if !addMember(policy, member, []string{"projects/p/roles/r", "roles/viewer"}) {
		t.Fatal("adding reported no change")
	}
	if addMember(policy, member, []string{"projects/p/roles/r", "roles/viewer"}) {
		t.Fatal("adding twice reported a change")
	}
	if !removeMember(policy, member) {
		t.Fatal("removing reported no change")
	}
	if len(policy.Bindings) != 1 || len(policy.Bindings[0].Members) != 1 || policy.Bindings[0].Members[0] != "user:someone@example.com" {
		t.Fatalf("policy after removal = %+v", policy.Bindings)
	}
	if !isLaneOrDeletedMember("deleted:serviceAccount:x@p.iam.gserviceaccount.com?uid=1") || isLaneOrDeletedMember("user:someone@example.com") {
		t.Fatal("leftover member classification is wrong")
	}
	// On a shared billing account only lane accounts go, deleted or not.
	if !isLaneMember(member) || !isLaneMember("deleted:"+member+"?uid=1") ||
		isLaneMember("deleted:serviceAccount:x@p.iam.gserviceaccount.com?uid=1") || isLaneMember("user:someone@example.com") {
		t.Fatal("billing account leftover classification is wrong")
	}
}

// Every GCP scenario that declares the identity has a grant the harness can
// mint: no permission Google keeps out of custom roles, no billing account
// group without its predefined role. A lane would refuse either at its IDENTITY phase; this finds it
// before any lane runs.
func TestEveryDeclaredScenarioHasAMintableGrant(t *testing.T) {
	root := repoRoot(t)
	t.Setenv("PLANTON_E2E_GCP_PROJECT_ID", "planton-e2e-unit")
	scenarios, err := filepath.Glob(filepath.Join(root, "catalog", "gcp", "*", "e2e", "scenarios", "*.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	declared := 0
	for _, scenario := range scenarios {
		raw, err := os.ReadFile(scenario)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(raw), "planton.dev/e2e-identity: declared") {
			continue
		}
		declared++
		kindDir := filepath.Base(filepath.Dir(filepath.Dir(filepath.Dir(scenario))))
		acts := []string{scenario}
		if secondAct, _ := runner.ManifestAnnotation(scenario, runner.UpgradeManifestAnnotation); secondAct != "" {
			acts = append(acts, filepath.Join(filepath.Dir(scenario), secondAct))
		}
		if _, err := declaredGrant(root, kindDir, acts); err != nil {
			rel, _ := filepath.Rel(root, scenario)
			t.Errorf("%s: %v", rel, err)
		}
	}
	if declared == 0 {
		t.Fatal("no GCP scenario declares the identity; the net checks nothing")
	}
}

// A key minted seconds ago is refused by Google's token endpoint with
// invalid_grant until it propagates; the readiness wait must read that, as it
// reads a 401 or 403, as "not yet", and nothing else.
func TestReadinessReadsAFreshKeysRefusalsAsNotYet(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"token endpoint invalid_grant, as the client wraps it", fmt.Errorf("Post %q: %w", "https://cloudresourcemanager.googleapis.com/v1/projects/p:testIamPermissions", &auth.Error{Response: &http.Response{StatusCode: 400}, Body: []byte(`{"error":"invalid_grant","error_description":"Invalid JWT Signature."}`)}), true},
		{"token endpoint invalid_client", &auth.Error{Response: &http.Response{StatusCode: 401}, Body: []byte(`{"error":"invalid_client"}`)}, false},
		{"API 401", &googleapi.Error{Code: 401}, true},
		{"API 403", &googleapi.Error{Code: 403}, true},
		{"API 404", &googleapi.Error{Code: 404}, false},
	}
	for _, c := range cases {
		if got := keyNotYetAccepted(c.err); got != c.want {
			t.Errorf("%s: keyNotYetAccepted = %v, want %v", c.name, got, c.want)
		}
	}
}

// A refusal of a permission the lane's role holds is a grant still
// propagating; a refusal of one it lacks, or an error naming none, is not.
func TestHeldPermissionRefusedReadsGooglesTwoWordings(t *testing.T) {
	h := NewHarness()
	tc := &provider.KindTestContext{}
	h.laneGrants[tc] = map[string]bool{"discoveryengine.engines.create": true, "dialogflow.tools.get": true}
	cases := []struct {
		err        string
		permission string
		held       bool
	}{
		{"googleapi: Error 403: Permission 'discoveryengine.engines.create' denied on resource '//discoveryengine.googleapis.com/projects/p/locations/global/collections/default_collection' (or it may not exist).", "discoveryengine.engines.create", true},
		{"googleapi: Error 403: IAM permission 'dialogflow.tools.get' on 'projects/p/locations/global/agents/a/tools/t' denied.", "dialogflow.tools.get", true},
		{"googleapi: Error 403: IAM permission 'dialogflow.agents.create' on 'projects/p/locations/global' denied.", "dialogflow.agents.create", false},
		{"googleapi: Error 400: control.useCase can only have 1 value but 2 were found.", "", false},
	}
	for _, c := range cases {
		permission, held := h.HeldPermissionRefused(tc, errors.New(c.err))
		if permission != c.permission || held != c.held {
			t.Errorf("%q: got (%q, %v), want (%q, %v)", c.err, permission, held, c.permission, c.held)
		}
	}
	if _, held := h.HeldPermissionRefused(&provider.KindTestContext{}, errors.New(cases[0].err)); held {
		t.Error("a lane with no recorded grant must not read a refusal as propagation")
	}
}

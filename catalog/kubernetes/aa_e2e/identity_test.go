package aa_e2e

import (
	"reflect"
	"strings"
	"testing"

	"github.com/plantonhq/planton/e2e/framework/provider"
	permissionsv1 "github.com/plantonhq/planton/iac/catalogkindpermissions/v1"
)

func TestParseIdentitySpec(t *testing.T) {
	if withheld, err := parseIdentitySpec("declared"); err != nil || withheld != nil {
		t.Fatalf("declared withholds nothing: %v %v", withheld, err)
	}
	withheld, err := parseIdentitySpec("declared-minus:apiextensions.k8s.io/customresourcedefinitions:create, patch,delete")
	if err != nil {
		t.Fatal(err)
	}
	if withheld.apiGroup != "apiextensions.k8s.io" || withheld.resource != "customresourcedefinitions" || len(withheld.verbs) != 3 || !withheld.verbs["patch"] {
		t.Fatalf("parsed wrong: %+v", withheld)
	}
	core, err := parseIdentitySpec("declared-minus:/namespaces:delete")
	if err != nil || core.apiGroup != "" || core.resource != "namespaces" {
		t.Fatalf("the core group is the empty string before the slash: %+v %v", core, err)
	}
	for _, bad := range []string{"admin", "declared-minus:", "declared-minus:apps/deployments", "declared-minus:deployments:create"} {
		if _, err := parseIdentitySpec(bad); err == nil {
			t.Errorf("%q must be refused", bad)
		}
	}
}

func TestClusterRoleRulesWithholdOnlyTheNamedVerbs(t *testing.T) {
	declared := []*permissionsv1.KubernetesRule{
		{ApiGroups: []string{""}, Resources: []string{"namespaces"}, Verbs: []string{"get", "create", "delete"}},
		{ApiGroups: []string{"apiextensions.k8s.io"}, Resources: []string{"customresourcedefinitions"}, Verbs: []string{"get", "list", "create", "update", "patch", "delete"}},
		{ApiGroups: []string{"apps"}, Resources: []string{"deployments", "statefulsets"}, Verbs: []string{"get", "create"}},
	}
	withheld, _ := parseIdentitySpec("declared-minus:apiextensions.k8s.io/customresourcedefinitions:create,update,patch,delete")
	rules, err := clusterRoleRules(declared, withheld)
	if err != nil {
		t.Fatal(err)
	}
	want := []rbacRule{
		{APIGroups: []string{""}, Resources: []string{"namespaces"}, Verbs: []string{"get", "create", "delete"}},
		{APIGroups: []string{"apiextensions.k8s.io"}, Resources: []string{"customresourcedefinitions"}, Verbs: []string{"get", "list"}},
		{APIGroups: []string{"apps"}, Resources: []string{"deployments", "statefulsets"}, Verbs: []string{"get", "create"}},
	}
	if !reflect.DeepEqual(rules, want) {
		t.Fatalf("rules:\n got %+v\nwant %+v", rules, want)
	}

	// A resource named beside others loses the verbs alone; its siblings keep them.
	split, _ := parseIdentitySpec("declared-minus:apps/deployments:create")
	rules, err = clusterRoleRules(declared[2:], split)
	if err != nil {
		t.Fatal(err)
	}
	want = []rbacRule{
		{APIGroups: []string{"apps"}, Resources: []string{"statefulsets"}, Verbs: []string{"get", "create"}},
		{APIGroups: []string{"apps"}, Resources: []string{"deployments"}, Verbs: []string{"get"}},
	}
	if !reflect.DeepEqual(rules, want) {
		t.Fatalf("split rules:\n got %+v\nwant %+v", rules, want)
	}
	if got, err := clusterRoleRules(declared, nil); err != nil || len(got) != 3 {
		t.Fatalf("declared alone keeps every rule, got %d %v", len(got), err)
	}
}

// The generic Helm kind declares a wildcard for the arbitrary chart it
// installs; withholding a CRD verb from it would leave the identity holding
// the verb anyway, so the harness refuses the spec rather than run a lane
// whose refusal can never come.
func TestClusterRoleRulesRefuseAWithholdAWildcardWouldVoid(t *testing.T) {
	declared := []*permissionsv1.KubernetesRule{
		{ApiGroups: []string{"apiextensions.k8s.io"}, Resources: []string{"customresourcedefinitions"}, Verbs: []string{"get", "list", "create", "patch"}},
		{ApiGroups: []string{"*"}, Resources: []string{"*"}, Verbs: []string{"*"}},
	}
	withheld, _ := parseIdentitySpec("declared-minus:apiextensions.k8s.io/customresourcedefinitions:patch")
	if _, err := clusterRoleRules(declared, withheld); err == nil || !strings.Contains(err.Error(), "wildcard") {
		t.Fatalf("expected the wildcard refusal, got %v", err)
	}
}

func TestIdentityNameIsAValidObjectName(t *testing.T) {
	name := identityName(&provider.KindTestContext{Kind: "kuberneteshelmrelease", Engine: "terraform", RunID: "20260903T1200_Z.abc"})
	if len(name) > 63 || strings.ContainsAny(name, "_.TZ") || !strings.HasPrefix(name, "lane-kuberneteshelmrelease-terraform-") {
		t.Fatalf("not a valid lowercase DNS label: %q", name)
	}
}

func TestIdentityManifestCarriesTheFourObjects(t *testing.T) {
	manifest, err := identityManifest("lane-x", []rbacRule{{APIGroups: []string{""}, Resources: []string{"namespaces"}, Verbs: []string{"get"}}})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"kind: Namespace", "kind: ServiceAccount", "kind: ClusterRole\n", "kind: ClusterRoleBinding", "name: " + identityNamespace} {
		if !strings.Contains(string(manifest), want) {
			t.Errorf("manifest lacks %q:\n%s", want, manifest)
		}
	}
}

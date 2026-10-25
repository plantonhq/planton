package verify

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/plantonhq/planton/pkg/failure"
)

// The Agent Substrate lanes are only as honest as the verifier's reading of
// their scenarios: who owns the namespace, which database and dataplane run,
// which grants the credential provider must hold, who the bootstrap owner
// is. Each committed scenario is read here offline, so a scenario edit that
// changes what a lane asserts shows up as a failing test, not as a lane that
// quietly proves something else. The refusal phrases the lanes look for are
// held to the texts the modules actually write, and the release to both
// modules' own.

const agentSubstrateScenarios = "../../kubernetesagentsubstrate/e2e/scenarios"

func TestAgentSubstrateVerifierReadsTheScenarios(t *testing.T) {
	research := []agentSubstrateGrant{{Atespace: "research", AllowedNamespaces: []string{"research-credentials"}}}
	changed := []agentSubstrateGrant{
		{Atespace: "research", AllowedNamespaces: []string{"research-credentials", "shared-credentials"}},
		{Atespace: "support", AllowedNamespaces: []string{"support-credentials"}},
	}
	type lane struct {
		create, bundled, agentgateway, keep, refused, policy, gcs bool
		provider, installed                                       string
		grants                                                    []agentSubstrateGrant
		crds                                                      []string
	}
	kind := lane{bundled: true, policy: true, provider: "disabled", crds: agentSubstrateCrds}
	with := func(edit func(*lane)) lane { l := kind; edit(&l); return l }
	cases := map[string]lane{
		"minimal.yaml":                     with(func(l *lane) { l.keep = true }),
		"agentgateway.yaml":                with(func(l *lane) { l.agentgateway = true }),
		"full-surface.yaml":                kind,
		"credential-policy-change.yaml":    with(func(l *lane) { l.provider, l.grants = "kubernetes", research }),
		"credential-policy-change.to.yaml": with(func(l *lane) { l.provider, l.grants = "kubernetes", changed }),
		"external-database.yaml":           with(func(l *lane) { l.bundled = false }),
		"gke-keyless.yaml":                 with(func(l *lane) { l.create, l.policy, l.gcs = true, false, true }),
		"namespace-missing.yaml":           with(func(l *lane) { l.keep, l.refused = true, true }),
		"namespace-owned-elsewhere.yaml":   with(func(l *lane) { l.create, l.keep, l.refused = true, true, true }),
		"workload-owned-elsewhere.yaml":    with(func(l *lane) { l.keep, l.refused = true, true }),
		"crds-owned-elsewhere.yaml": with(func(l *lane) {
			l.create, l.refused, l.crds = true, true, []string{"workerpools.ate.dev"}
		}),
		"upgrade-refused.yaml":        with(func(l *lane) { l.keep, l.refused, l.installed = true, true, "v0.5.0" }),
		"volume-upgrade-refused.yaml": with(func(l *lane) { l.keep, l.refused, l.installed = true, true, "v0.5" }),
	}
	files, err := filepath.Glob(filepath.Join(agentSubstrateScenarios, "*.yaml"))
	if err != nil || len(files) != len(cases) {
		t.Fatalf("every scenario needs a case here: found %d files, %d cases (%v)", len(files), len(cases), err)
	}
	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			v := newAgentSubstrateVerifier(filepath.Join(agentSubstrateScenarios, name))
			got := lane{
				create: v.CreateNamespace, bundled: v.BundledDatabase, agentgateway: v.Agentgateway,
				keep: v.KeepOnUninstall, refused: v.DeployRefused, policy: v.EnforcesNetworkPolicy, gcs: v.GcsBucket != "",
				provider: v.CredentialProvider, installed: v.ExpectedVersion, grants: v.Grants, crds: v.CRDs,
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("read %+v, want %+v", got, want)
			}
			if v.ChartVersion != agentSubstrateRelease || v.SourceLabel != agentSubstrateBundle {
				t.Errorf("stamps version=%q source=%q, want %s %s", v.ChartVersion, v.SourceLabel, agentSubstrateRelease, agentSubstrateBundle)
			}
		})
	}
}

// Authorization is always on and the Envoy router serves only the clients it
// names, so a lane that installs proves its promises only if the probe is
// one of its bootstrap owners and, behind the Envoy router, one of its
// allowed router clients: the API and the router would refuse it otherwise.
func TestEveryInstallingLaneAdmitsTheProbe(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(agentSubstrateScenarios, "*.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	client := "spiffe://cluster.local/ns/" + agentSubstrateProbeNamespace + "/sa/" + agentSubstrateProbeClient
	for _, path := range append(files, "../../kubernetesagentsubstrate/e2e/prerequisite.yaml") {
		v := newAgentSubstrateVerifier(path)
		if v.DeployRefused {
			continue
		}
		if !slices.Contains(v.BootstrapOwners, v.probeOwner()) {
			t.Errorf("%s names bootstrap owners %v; the probe calls as %s", filepath.Base(path), v.BootstrapOwners, v.probeOwner())
		}
		router, _ := specField(manifestSpecMap(path), "router").(map[string]interface{})
		allowed, _ := specField(router, "allowedClientSpiffeIds").([]interface{})
		if !v.Agentgateway && !containsValue(allowed, client) {
			t.Errorf("%s allows router clients %v; the probe's router caller is %s", filepath.Base(path), allowed, client)
		}
	}
}

func containsValue(list []interface{}, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}

// Each refusal phrase the lanes look for must appear in the text the
// modules write: the shared failure constructors render it, and the upgrade
// rule and remedies are read from both engines' sources.
func TestAgentSubstrateRefusalTextsMatchTheModules(t *testing.T) {
	read := func(path string) string {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return string(src)
	}
	goSource := read("../../kubernetesagentsubstrate/iac/pulumi/module/upgrade.go")
	tfSource := read("../../kubernetesagentsubstrate/iac/tf/preflight.tf")
	rule := regexp.MustCompile(`ruleBeforeOne = "([^"]+)"`).FindStringSubmatch(goSource)
	if rule == nil {
		t.Fatal("upgrade.go no longer names ruleBeforeOne")
	}
	for _, phrase := range []string{"destroy this resource", "delete the volume (`kubectl -n %s delete pvc %s`)"} {
		if !strings.Contains(goSource, phrase) {
			t.Errorf("upgrade.go no longer says %q", phrase)
		}
	}
	if !strings.Contains(tfSource, "delete the volume (`kubectl -n ${local.namespace} delete pvc ${local.postgres_volume_claim}`)") || !strings.Contains(tfSource, "destroy this resource") {
		t.Error("preflight.tf no longer names the remedies the lanes look for")
	}

	rendered := map[string]string{
		failureWorkloadOwnedElsewhere:  failure.KubernetesWorkloadOwnedElsewhere(agentSubstrateKind, "Deployment", agentSubstrateNamespace, agentSubstrateAPIServer, "no owner label").Error(),
		failureNamespaceMissing:        failure.KubernetesNamespaceMissing(agentSubstrateKind, agentSubstrateNamespace).Error(),
		failureNamespaceOwnedElsewhere: failure.KubernetesNamespaceOwnedElsewhere(agentSubstrateKind, agentSubstrateNamespace, "no owner label").Error(),
	}
	for expectation, text := range rendered {
		wants, err := agentSubstrateRefusalTexts(expectation, "", agentSubstrateRelease)
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range wants {
			if !strings.Contains(text, want) {
				t.Errorf("%s: the module writes %q, which lacks %q", expectation, text, want)
			}
		}
	}

	for installed, record := range map[string]string{
		"v0.5.0": "Deployment " + agentSubstrateNamespace + "/" + agentSubstrateAPIServer,
		"v0.5":   "PersistentVolumeClaim " + agentSubstrateNamespace + "/" + agentSubstrateVolumeClaim + ", the bundled database's volume left by an earlier installation,",
	} {
		remedy := "back up what you need from the database, destroy this resource"
		if installed == "v0.5" {
			remedy = "delete the volume (`kubectl -n " + agentSubstrateNamespace + " delete pvc " + agentSubstrateVolumeClaim + "`)"
		}
		text := failure.KubernetesUpgradeRefused(agentSubstrateKind, "Agent Substrate", record, installed, agentSubstrateRelease, rule[1], remedy).Error()
		wants, err := agentSubstrateRefusalTexts(failureUpgradeRefused, installed, agentSubstrateRelease)
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range wants {
			if !strings.Contains(text, want) {
				t.Errorf("upgrade from %s: the module writes %q, which lacks %q", installed, text, want)
			}
		}
	}
}

func TestAgentSubstrateVerifierPinsTheModulesRelease(t *testing.T) {
	for path, pattern := range map[string]*regexp.Regexp{
		"../../kubernetesagentsubstrate/iac/pulumi/module/vars.go": regexp.MustCompile(`Release:\s+"([^"]+)"`),
		"../../kubernetesagentsubstrate/iac/tf/locals.tf":          regexp.MustCompile(`release\s*=\s*"([^"]+)"`),
		"internal/ateapipb/upstream_test.go":                       regexp.MustCompile(`upstreamRelease = "([^"]+)"`),
	} {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		match := pattern.FindSubmatch(src)
		if match == nil || string(match[1]) != agentSubstrateRelease {
			t.Errorf("%s pins %q; the verifier expects %s", path, match, agentSubstrateRelease)
		}
	}
}

// The API's authorizer names a denied principal in its store's escaped form;
// the full-surface lane asserts the stranger is denied by exactly that name.
func TestAteAuthorizationUserEscapesLikeTheAPI(t *testing.T) {
	got := ateAuthorizationUser("system:serviceaccount:e2e-ate-probe:e2e-ate-stranger")
	if want := "user:system%3Aserviceaccount%3Ae2e-ate-probe%3Ae2e-ate-stranger"; got != want {
		t.Errorf("ateAuthorizationUser = %q, want %q (the principal the API printed in a live denial)", got, want)
	}
}

package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"
)

// writeKindManifest writes a real, loadable manifest (unlike writeManifest's
// Placeholder, these go through manifest.LoadManifest in the checker).
func writeKindManifest(t *testing.T, repoRoot, relPath, content string) string {
	t.Helper()
	full := filepath.Join(repoRoot, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", relPath, err)
	}
	if err := os.WriteFile(full, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", relPath, err)
	}
	return full
}

const fakeVpcPrereq = `apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVpcNetwork
metadata:
  name: fake-vpc-prereq
spec:
  networkName: fake-vpc
  autoCreateSubnetworks: false
`

func fakeSubnetworkScenario(vpcRefName string) string {
	return `apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSubnetwork
metadata:
  name: fake-subnet-scenario
spec:
  subnetworkName: fake-subnet
  region: us-central1
  ipCidrRange: 10.10.0.0/24
  vpcSelfLink:
    valueFrom:
      kind: GcpVpcNetwork
      name: ` + vpcRefName + `
      fieldPath: status.outputs.network_self_link
`
}

func TestFixtureIntegrity_ExactNameResolves(t *testing.T) {
	repoRoot := t.TempDir()
	writeKindManifest(t, repoRoot, "catalog/gcp/gcpvpcnetwork/e2e/prerequisite.yaml", fakeVpcPrereq)
	scenario := writeKindManifest(t, repoRoot, "catalog/gcp/gcpsubnetwork/e2e/scenarios/minimal.yaml",
		fakeSubnetworkScenario("fake-vpc-prereq"))

	findings, err := CheckScenarioFixtureIntegrity(repoRoot, "gcp", "gcpsubnetwork", scenario)
	if err != nil {
		t.Fatalf("CheckScenarioFixtureIntegrity: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %v", findings)
	}
}

// A name that matches no deployed instance still resolves when exactly one
// instance of the kind exists -- the runner's documented sole-instance
// fallback (scenario references name real-world topology, not install-profile
// files). This is design, not a defect; the gate must never flag it.
func TestFixtureIntegrity_SoleInstanceFallbackIsLegal(t *testing.T) {
	repoRoot := t.TempDir()
	writeKindManifest(t, repoRoot, "catalog/gcp/gcpvpcnetwork/e2e/prerequisite.yaml", fakeVpcPrereq)
	scenario := writeKindManifest(t, repoRoot, "catalog/gcp/gcpsubnetwork/e2e/scenarios/minimal.yaml",
		fakeSubnetworkScenario("some-topology-name"))

	findings, err := CheckScenarioFixtureIntegrity(repoRoot, "gcp", "gcpsubnetwork", scenario)
	if err != nil {
		t.Fatalf("CheckScenarioFixtureIntegrity: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected the sole-instance fallback to resolve, got findings: %v", findings)
	}
}

// With SEVERAL deployed instances of the referenced kind, a name matching none
// of them is the ambiguity lookupRefValue hard-errors on -- at deploy time,
// after the fixtures are already up. The gate catches it offline.
func TestFixtureIntegrity_AmbiguousNameAmongMultipleInstances(t *testing.T) {
	repoRoot := t.TempDir()
	twoVpcs := fakeVpcPrereq + "---\n" + strings.Replace(fakeVpcPrereq, "fake-vpc-prereq", "fake-vpc-second", 1)
	writeKindManifest(t, repoRoot, "catalog/gcp/gcpvpcnetwork/e2e/prerequisite.yaml", twoVpcs)
	scenario := writeKindManifest(t, repoRoot, "catalog/gcp/gcpsubnetwork/e2e/scenarios/minimal.yaml",
		fakeSubnetworkScenario("neither-of-them"))

	findings, err := CheckScenarioFixtureIntegrity(repoRoot, "gcp", "gcpsubnetwork", scenario)
	if err != nil {
		t.Fatalf("CheckScenarioFixtureIntegrity: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 ambiguity finding, got %v", findings)
	}
	if !strings.Contains(findings[0].Reason, "matches none of the 2 deployed instances") {
		t.Errorf("unexpected reason: %s", findings[0].Reason)
	}
}

// A reference whose kind the chain never deploys is left unresolved by the
// runner (by design) and the module deploys without the value -- the failure
// surfaces at DEPLOY as a provider error that reads like a module defect.
func TestFixtureIntegrity_KindNotInChain(t *testing.T) {
	repoRoot := t.TempDir()
	writeKindManifest(t, repoRoot, "catalog/gcp/gcpvpcnetwork/e2e/prerequisite.yaml", fakeVpcPrereq)
	scenario := writeKindManifest(t, repoRoot, "catalog/gcp/gcpsubnetwork/e2e/scenarios/minimal.yaml",
		strings.Replace(fakeSubnetworkScenario("fake-vpc-prereq"), "kind: GcpVpcNetwork", "kind: GcpAddress", 1))

	findings, err := CheckScenarioFixtureIntegrity(repoRoot, "gcp", "gcpsubnetwork", scenario)
	if err != nil {
		t.Fatalf("CheckScenarioFixtureIntegrity: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 kind-not-in-chain finding, got %v", findings)
	}
	if !strings.Contains(findings[0].Reason, "deploys no instance of this kind") {
		t.Errorf("unexpected reason: %s", findings[0].Reason)
	}
	if findings[0].RefKind != "GcpAddress" {
		t.Errorf("RefKind = %q, want GcpAddress", findings[0].RefKind)
	}
}

// A bare polymorphic reference (a field with no default_kind) that omits the
// explicit `kind:` can never resolve -- the documented trap no other offline
// gate catches (e2e/README.md, "Bare polymorphic references"). A DNS record's
// values are the catalog's plainest such field: an entry can be any
// resource's output, so the field names no kind.
func TestFixtureIntegrity_BareRefWithoutKind(t *testing.T) {
	repoRoot := t.TempDir()
	writeKindManifest(t, repoRoot, "catalog/gcp/gcpvpcnetwork/e2e/prerequisite.yaml", fakeVpcPrereq)
	writeKindManifest(t, repoRoot, "catalog/gcp/gcpdnszone/e2e/prerequisite.yaml",
		`apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDnsZone
metadata:
  name: fake-zone-prereq
spec: {}
`)
	scenario := writeKindManifest(t, repoRoot, "catalog/gcp/gcpdnsrecord/e2e/scenarios/minimal.yaml",
		`apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDnsRecord
metadata:
  name: fake-record-scenario
spec:
  type: TXT
  values:
    - valueFrom:
        name: fake-dns-authorization
        fieldPath: status.outputs.dns_record_data
`)

	findings, err := CheckScenarioFixtureIntegrity(repoRoot, "gcp", "gcpdnsrecord", scenario)
	if err != nil {
		t.Fatalf("CheckScenarioFixtureIntegrity: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 bare-ref finding, got %v", findings)
	}
	if !strings.Contains(findings[0].Reason, "no kind") {
		t.Errorf("unexpected reason: %s", findings[0].Reason)
	}
}

// A component whose profile records `status: deferred` is skipped by the
// CATALOG walk: its deferral is the record that its lanes cannot run, so an
// unresolvable chain is not a finding. Any other status keeps the check
// armed -- the same broken chain must fail the moment the deferral lifts.
func TestFixtureIntegrity_DeferredProfileSkipsCatalogCheck(t *testing.T) {
	profileYAML := func(status string) string {
		return `apiVersion: qa.planton.dev/v1
kind: ComponentE2EProfile
metadata:
  name: gcpsubnetwork
spec:
  tier: 1
  status: ` + status + `
`
	}
	// The scenario's vpc prerequisite has no install manifest anywhere --
	// exactly the chain-resolution failure the org-wall class produces.
	brokenScenario := fakeSubnetworkScenario("fake-vpc-prereq")

	for _, tc := range []struct {
		status       string
		wantFindings int
	}{
		{"deferred", 0},
		{"pending_proof", 1},
	} {
		repoRoot := t.TempDir()
		writeKindManifest(t, repoRoot, "catalog/gcp/gcpsubnetwork/e2e/scenarios/minimal.yaml", brokenScenario)
		writeKindManifest(t, repoRoot, "catalog/gcp/gcpsubnetwork/e2e/profile.yaml", profileYAML(tc.status))

		findings, err := CheckCatalogFixtureIntegrity(repoRoot)
		if err != nil {
			t.Fatalf("CheckCatalogFixtureIntegrity (status %s): %v", tc.status, err)
		}
		if len(findings) != tc.wantFindings {
			t.Errorf("status %s: got %d findings (%v), want %d", tc.status, len(findings), findings, tc.wantFindings)
		}
	}
}

// A chain that cannot resolve at all (a prerequisite with no install
// manifest) is reported as a finding, not swallowed.
func TestFixtureIntegrity_ChainResolutionFailureIsAFinding(t *testing.T) {
	repoRoot := t.TempDir()
	scenario := writeKindManifest(t, repoRoot, "catalog/gcp/gcpsubnetwork/e2e/scenarios/minimal.yaml",
		fakeSubnetworkScenario("fake-vpc-prereq"))

	findings, err := CheckScenarioFixtureIntegrity(repoRoot, "gcp", "gcpsubnetwork", scenario)
	if err != nil {
		t.Fatalf("CheckScenarioFixtureIntegrity: %v", err)
	}
	if len(findings) != 1 || !strings.Contains(findings[0].Reason, "chain resolution failed") {
		t.Fatalf("expected a chain-resolution finding, got %v", findings)
	}
}

// TestCatalogFixtureIntegrity is the repo-wide gate: every committed scenario's
// reference graph must resolve against the prerequisite chain the runner will
// deploy for it. This is the offline stand-in for the DEPENDENCIES-UP phase --
// a failure here would otherwise cost a live run minutes of fixture deploys
// before surfacing (or fail silently at DEPLOY for bare references).
//
// Pre-existing defects found when the gate was introduced live in the
// committed baseline file (fixture_integrity_baseline.yaml) with their owners,
// exactly like the provider-parity checker's baseline: a finding NOT in the
// baseline fails the gate (fix the manifest, never extend the baseline), and a
// baseline entry whose finding no longer occurs fails as stale (delete the
// entry -- the baseline only ever burns down).
func TestCatalogFixtureIntegrity(t *testing.T) {
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("locating repo root: %v", err)
	}

	findings, err := CheckCatalogFixtureIntegrity(repoRoot)
	if err != nil {
		t.Fatalf("CheckCatalogFixtureIntegrity: %v", err)
	}

	baseline, err := loadFixtureIntegrityBaseline()
	if err != nil {
		t.Fatalf("loading baseline: %v", err)
	}

	matched := make(map[string]bool, len(baseline))
	for _, f := range findings {
		key := findingKey(repoRoot, f)
		if _, known := baseline[key]; known {
			matched[key] = true
			continue
		}
		t.Errorf("unbaselined finding (fix the manifest; do not extend the baseline)\n  key: %s\n  %s", key, f)
	}
	for key, entry := range baseline {
		if !matched[key] {
			t.Errorf("stale baseline entry (the finding no longer occurs -- delete it from %s)\n  key: %s\n  owner: %s", fixtureIntegrityBaselineFile, key, entry.Owner)
		}
	}
}

const fixtureIntegrityBaselineFile = "fixture_integrity_baseline.yaml"

type fixtureIntegrityBaselineEntry struct {
	Key   string `json:"key"`
	Owner string `json:"owner"`
	Note  string `json:"note"`
}

// findingKey renders a finding as a stable, repo-relative identity (reasons
// are excluded -- they carry environment-specific detail like error text).
func findingKey(repoRoot string, f FixtureIntegrityFinding) string {
	rel := func(p string) string {
		if r, err := filepath.Rel(repoRoot, p); err == nil {
			return r
		}
		return p
	}
	return strings.Join([]string{rel(f.ScenarioPath), rel(f.ManifestPath), f.Field, f.RefKind + "/" + f.RefName}, "|")
}

func loadFixtureIntegrityBaseline() (map[string]fixtureIntegrityBaselineEntry, error) {
	raw, err := os.ReadFile(fixtureIntegrityBaselineFile)
	if os.IsNotExist(err) {
		return nil, nil // no baseline -- every finding fails the gate
	}
	if err != nil {
		return nil, err
	}
	var entries []fixtureIntegrityBaselineEntry
	if err := yaml.Unmarshal(raw, &entries); err != nil {
		return nil, err
	}
	baseline := make(map[string]fixtureIntegrityBaselineEntry, len(entries))
	for _, e := range entries {
		baseline[e.Key] = e
	}
	return baseline, nil
}

// findRepoRoot walks up from the package directory to the repository root
// (the directory holding both go.mod and catalog/).
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if pathExists(filepath.Join(dir, "go.mod")) && pathExists(filepath.Join(dir, "catalog")) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

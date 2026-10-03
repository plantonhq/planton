package setdeploy

import (
	"strings"
	"testing"

	"github.com/plantonhq/planton/pkg/iac/provisioner"
	"github.com/plantonhq/planton/pkg/iac/tofu/backendconfig"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// fakeProbes verifies everything by default; tests flip individual outcomes.
type fakeProbes struct {
	hclBinary       ProbeResult
	pulumiBinary    ProbeResult
	module          ProbeResult
	tofuBackend     ProbeResult
	pulumiBackend   ProbeResult
	credentials     map[catalogkind.CatalogProvider]ProbeResult
	kubeContext     ProbeResult
	tofuBackendSeen []*backendconfig.TofuBackendConfig
}

func newFakeProbes() *fakeProbes {
	return &fakeProbes{
		hclBinary:     verified("tofu binary on PATH"),
		pulumiBinary:  verified("pulumi binary on PATH"),
		module:        verified("module published"),
		tofuBackend:   verified("state backend reachable"),
		pulumiBackend: verified("pulumi backend reachable"),
		credentials:   map[catalogkind.CatalogProvider]ProbeResult{},
		kubeContext:   verified("kube context exists"),
	}
}

func (f *fakeProbes) HclBinary(string) ProbeResult { return f.hclBinary }
func (f *fakeProbes) PulumiBinary() ProbeResult    { return f.pulumiBinary }
func (f *fakeProbes) ModulePublished(string, provisioner.ProvisionerType, string) ProbeResult {
	return f.module
}
func (f *fakeProbes) TofuBackend(cfg *backendconfig.TofuBackendConfig) ProbeResult {
	f.tofuBackendSeen = append(f.tofuBackendSeen, cfg)
	return f.tofuBackend
}
func (f *fakeProbes) PulumiBackend(string) ProbeResult { return f.pulumiBackend }
func (f *fakeProbes) ProviderCredentials(p catalogkind.CatalogProvider) ProbeResult {
	if r, ok := f.credentials[p]; ok {
		return r
	}
	return verified("%s credentials authenticate", p)
}
func (f *fakeProbes) KubeContext(string) ProbeResult { return f.kubeContext }

// docsOf builds Docs from inline YAML manifests.
func docsOf(t *testing.T, sources map[string]string) []Doc {
	t.Helper()
	var docs []Doc
	// Deterministic order: sorted by label, matching authored-file sorting.
	labels := make([]string, 0, len(sources))
	for label := range sources {
		labels = append(labels, label)
	}
	for i := 0; i < len(labels); i++ {
		for j := i + 1; j < len(labels); j++ {
			if labels[j] < labels[i] {
				labels[i], labels[j] = labels[j], labels[i]
			}
		}
	}
	for _, label := range labels {
		docs = append(docs, Doc{Bytes: []byte(sources[label]), Source: label})
	}
	return docs
}

const producerYaml = `apiVersion: _test.planton.dev/v1alpha2
kind: TestCatalogKindGeneric
metadata:
  name: producer
  env: dev
spec:
  requiredRef:
    value: literal
`

const consumerYaml = `apiVersion: _test.planton.dev/v1alpha2
kind: TestCatalogKindGeneric
metadata:
  name: consumer
  env: dev
spec:
  requiredRef:
    value: literal
  annotatedRef:
    valueFrom:
      name: producer
`

func checkByName(t *testing.T, report *Report, name string) *Check {
	t.Helper()
	for i := range report.Checks {
		if report.Checks[i].Name == name {
			return &report.Checks[i]
		}
	}
	t.Fatalf("report has no check %q", name)
	return nil
}

func requireRefused(t *testing.T, report *Report, checkName, needle string) {
	t.Helper()
	check := checkByName(t, report, checkName)
	for _, e := range check.Entries {
		if e.Severity == SeverityRefusal && strings.Contains(e.Message, needle) {
			return
		}
	}
	t.Fatalf("check %q has no refusal containing %q; entries: %+v", checkName, needle, check.Entries)
}

func TestPreflight_TwoNodeSetPasses(t *testing.T) {
	docs := docsOf(t, map[string]string{"01-producer.yaml": producerYaml, "02-consumer.yaml": consumerYaml})
	plan := Preflight(docs, Flags{}, newFakeProbes())

	if plan.Report.Refused() {
		t.Fatalf("expected a passing wall; report: %+v", plan.Report)
	}
	if len(plan.Order) != 2 {
		t.Fatalf("expected 2 nodes in order, got %d", len(plan.Order))
	}
	first := plan.Set.Nodes[plan.Order[0]].Identity
	second := plan.Set.Nodes[plan.Order[1]].Identity
	if first.Slug != "producer" || second.Slug != "consumer" {
		t.Fatalf("expected producer before consumer, got %s -> %s", first, second)
	}
	// The annotation-riding reference must have created the edge (defaults
	// materialize before ordering — the one-rule-set law).
	if len(plan.Graph.DependsOn[plan.Order[1]]) != 1 {
		t.Fatalf("expected the consumer to depend on the producer")
	}
}

func TestPreflight_ProvisionerDefaultsToTofu(t *testing.T) {
	docs := docsOf(t, map[string]string{"01-producer.yaml": producerYaml})
	plan := Preflight(docs, Flags{}, newFakeProbes())
	if plan.Nodes[0].Provisioner != provisioner.ProvisionerTypeTofu || !plan.Nodes[0].ProvisionerDefault {
		t.Fatalf("expected tofu default; got %+v", plan.Nodes[0])
	}
	engine := checkByName(t, plan.Report, "engine-and-modules")
	found := false
	for _, fact := range engine.Verified {
		if strings.Contains(fact, "defaulted to tofu") {
			found = true
		}
	}
	if !found {
		t.Fatalf("the default must be STATED in the report; verified: %v", engine.Verified)
	}
}

func TestPreflight_SchemaViolationRefusesWithFieldLine(t *testing.T) {
	// requiredRef is required by the kind's schema; omitting it must land as
	// a per-field refusal under load-and-schema, and the healthy remainder
	// still flows through the graph checks.
	broken := `apiVersion: _test.planton.dev/v1alpha2
kind: TestCatalogKindGeneric
metadata:
  name: broken
  env: dev
spec: {}
`
	docs := docsOf(t, map[string]string{"01-broken.yaml": broken, "02-producer.yaml": producerYaml})
	plan := Preflight(docs, Flags{}, newFakeProbes())
	requireRefused(t, plan.Report, "load-and-schema", "required_ref")
	if len(plan.Set.Nodes) != 1 {
		t.Fatalf("the healthy remainder must still form the set; got %d nodes", len(plan.Set.Nodes))
	}
}

func TestPreflight_DuplicateIdentityRefuses(t *testing.T) {
	docs := docsOf(t, map[string]string{"01-a.yaml": producerYaml, "02-b.yaml": producerYaml})
	plan := Preflight(docs, Flags{}, newFakeProbes())
	requireRefused(t, plan.Report, "identity", "duplicate resource identity")
}

func TestPreflight_ExternalValueFromRefuses(t *testing.T) {
	lonelyConsumer := docsOf(t, map[string]string{"01-consumer.yaml": consumerYaml})
	plan := Preflight(lonelyConsumer, Flags{}, newFakeProbes())
	requireRefused(t, plan.Report, "references", "no backend exists here to discover it")
}

func TestPreflight_BackendResolvedValueRefuses(t *testing.T) {
	withSecret := `apiVersion: _test.planton.dev/v1alpha2
kind: TestCatalogKindGeneric
metadata:
  name: secretful
  env: dev
spec:
  requiredRef:
    value: literal
  sensitiveString: $secret/db-password
`
	docs := docsOf(t, map[string]string{"01-secretful.yaml": withSecret})
	plan := Preflight(docs, Flags{}, newFakeProbes())
	requireRefused(t, plan.Report, "backend-resolved-values", "$secret")
	requireRefused(t, plan.Report, "backend-resolved-values", "provider-native secret references")
}

func TestPreflight_CycleRefusesNamingChain(t *testing.T) {
	a := `apiVersion: _test.planton.dev/v1alpha2
kind: TestCatalogKindGeneric
metadata:
  name: alpha
  env: dev
spec:
  requiredRef:
    value: literal
  annotatedRef:
    valueFrom:
      name: beta
`
	b := `apiVersion: _test.planton.dev/v1alpha2
kind: TestCatalogKindGeneric
metadata:
  name: beta
  env: dev
spec:
  requiredRef:
    value: literal
  annotatedRef:
    valueFrom:
      name: alpha
`
	docs := docsOf(t, map[string]string{"01-a.yaml": a, "02-b.yaml": b})
	plan := Preflight(docs, Flags{}, newFakeProbes())
	requireRefused(t, plan.Report, "cycles", "dependency cycle")
	if plan.Order != nil {
		t.Fatalf("a cyclic set must have no order")
	}
}

func TestPreflight_RemoteBackendMissingKeyRefusesNamingAnnotation(t *testing.T) {
	docs := docsOf(t, map[string]string{"01-producer.yaml": producerYaml})
	plan := Preflight(docs, Flags{BackendType: "s3", BackendBucket: "state", BackendRegion: "us-east-1"}, newFakeProbes())
	requireRefused(t, plan.Report, "state-backend", "tofu.planton.dev/backend.key")
}

func TestPreflight_StateKeyCollisionRefuses(t *testing.T) {
	sharedKey := func(name string) string {
		return `apiVersion: _test.planton.dev/v1alpha2
kind: TestCatalogKindGeneric
metadata:
  name: ` + name + `
  env: dev
  annotations:
    tofu.planton.dev/backend.key: shared/one.tfstate
spec:
  requiredRef:
    value: literal
`
	}
	docs := docsOf(t, map[string]string{"01-a.yaml": sharedKey("alpha"), "02-b.yaml": sharedKey("beta")})
	plan := Preflight(docs, Flags{BackendType: "s3", BackendBucket: "state", BackendRegion: "us-east-1"}, newFakeProbes())
	requireRefused(t, plan.Report, "state-backend", "overwrite each other")
}

func TestPreflight_LocalStateIsStatedWithCiNotice(t *testing.T) {
	docs := docsOf(t, map[string]string{"01-producer.yaml": producerYaml, "02-consumer.yaml": consumerYaml})
	plan := Preflight(docs, Flags{}, newFakeProbes())
	backend := checkByName(t, plan.Report, "state-backend")
	statedNodes := 0
	for _, fact := range backend.Verified {
		if strings.Contains(fact, "local state in this node's own workspace") {
			statedNodes++
		}
	}
	if statedNodes != 2 {
		t.Fatalf("every local node's state location must be stated; got %d of 2", statedNodes)
	}
	noticed := false
	for _, e := range backend.Entries {
		if e.Severity == SeverityWarning && strings.Contains(e.Message, "remote backend") {
			noticed = true
		}
	}
	if !noticed {
		t.Fatalf("local state must carry the CI notice; entries: %+v", backend.Entries)
	}
}

func TestPreflight_PulumiNodeNeedsStackIdentity(t *testing.T) {
	pulumiNode := `apiVersion: _test.planton.dev/v1alpha2
kind: TestCatalogKindGeneric
metadata:
  name: pnode
  env: dev
  annotations:
    planton.dev/provisioner: pulumi
spec:
  requiredRef:
    value: literal
`
	docs := docsOf(t, map[string]string{"01-p.yaml": pulumiNode})
	plan := Preflight(docs, Flags{}, newFakeProbes())
	requireRefused(t, plan.Report, "state-backend", "pulumi.planton.dev/stack.fqdn")
}

func TestPreflight_PulumiStackCollisionRefuses(t *testing.T) {
	pulumiNode := func(name string) string {
		return `apiVersion: _test.planton.dev/v1alpha2
kind: TestCatalogKindGeneric
metadata:
  name: ` + name + `
  env: dev
  annotations:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/stack.fqdn: org/proj/shared
spec:
  requiredRef:
    value: literal
`
	}
	docs := docsOf(t, map[string]string{"01-a.yaml": pulumiNode("alpha"), "02-b.yaml": pulumiNode("beta")})
	plan := Preflight(docs, Flags{}, newFakeProbes())
	requireRefused(t, plan.Report, "state-backend", "sharing one stack overwrite each other")
}

func TestPreflight_ProbeRefusalsSurface(t *testing.T) {
	probes := newFakeProbes()
	probes.hclBinary = refused("tofu is not on PATH — install it")
	docs := docsOf(t, map[string]string{"01-producer.yaml": producerYaml})
	plan := Preflight(docs, Flags{}, probes)
	requireRefused(t, plan.Report, "engine-and-modules", "not on PATH")
}

func TestPreflight_ModuleFallbackAssumptionRendersAsWarning(t *testing.T) {
	probes := newFakeProbes()
	probes.module = assumed("module artifact for X at v1 is not published (HTTP 404) — the deploy falls back to a source checkout, which is slower")
	docs := docsOf(t, map[string]string{"01-producer.yaml": producerYaml})
	plan := Preflight(docs, Flags{}, probes)
	engine := checkByName(t, plan.Report, "engine-and-modules")
	for _, e := range engine.Entries {
		if strings.Contains(e.Message, "falls back to a source checkout") {
			if e.Severity != SeverityWarning {
				t.Fatalf("module fallback must be a WARNING (loud beats slow-and-silent), got %s", e.Severity)
			}
			return
		}
	}
	t.Fatalf("module fallback entry missing")
}

func TestPreflight_TestProviderNeedsNoCredentials(t *testing.T) {
	docs := docsOf(t, map[string]string{"01-producer.yaml": producerYaml})
	probes := newFakeProbes()
	probes.credentials[catalogkind.CatalogProvider__test] = refused("must never be called")
	plan := Preflight(docs, Flags{}, probes)
	creds := checkByName(t, plan.Report, "provider-credentials")
	if creds.refusals() != 0 {
		t.Fatalf("the _test provider must not be credential-probed; entries: %+v", creds.Entries)
	}
}

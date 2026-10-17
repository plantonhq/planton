package helmcrds

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"
)

// Every test here runs offline: charts load from testdata directories (Helm's
// LocateChart returns a local path as-is) and the bundle is served by an
// in-process HTTP server. The network paths are exercised by the e2e lanes.

func fixtureChart(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(path, "Chart.yaml")); err != nil {
		t.Fatalf("fixture chart %s missing: %v", name, err)
	}
	return path
}

func decode(t *testing.T, doc string) map[string]interface{} {
	t.Helper()
	var object map[string]interface{}
	if err := yaml.Unmarshal([]byte(doc), &object); err != nil {
		t.Fatalf("stamped document is not YAML: %v\n%s", err, doc)
	}
	return object
}

func metadataMap(t *testing.T, object map[string]interface{}, key string) map[string]interface{} {
	t.Helper()
	metadata, _ := object["metadata"].(map[string]interface{})
	m, _ := metadata[key].(map[string]interface{})
	if m == nil {
		t.Fatalf("metadata.%s missing", key)
	}
	return m
}

func TestDeriveTemplatedCRDsUsesReleaseValuesAndIdentity(t *testing.T) {
	src := Source{
		Chart:   fixtureChart(t, "templated-crds"),
		Version: "1.2.3",
		// The release's own values: the module pins fullnameOverride and
		// keeps the chart's CRD switch OFF for the install.
		Values:      []string{"fullnameOverride: my-op\ncrds:\n  create: false\n"},
		CRDOverride: "crds:\n  create: true\n",
		APIVersions: []string{"cert-manager.io/v1"},
	}
	derived, err := Derive(context.Background(), src, Policy{ExpectCRDs: true}, "my-op", "ops")
	crds := derived.Owned
	if err != nil {
		t.Fatal(err)
	}
	if len(crds) != 2 {
		t.Fatalf("expected 2 CRDs, got %d", len(crds))
	}
	// Sorted by name, so gadgets before widgets.
	if crds[0].Name != "gadgets.fixture.planton.ai" || crds[1].Name != "widgets.fixture.planton.ai" {
		t.Fatalf("unexpected names: %s, %s", crds[0].Name, crds[1].Name)
	}

	widgets := decode(t, crds[1].YAML)
	annotations := metadataMap(t, widgets, "annotations")
	if got := annotations["cert-manager.io/inject-ca-from"]; got != "ops/my-op-serving-cert" {
		t.Fatalf("release-derived annotation not rendered from the real identity: %v", got)
	}
	if got := annotations[AnnotationSourceVersion]; got != "1.2.3" {
		t.Fatalf("source version stamp: %v", got)
	}
	if got := annotations[AnnotationSourceChart]; got != "/"+src.Chart {
		t.Fatalf("source chart stamp: %v", got)
	}
	labels := metadataMap(t, widgets, "labels")
	if got := labels[LabelSource]; got != src.Chart {
		t.Fatalf("source label: %v", got)
	}

	// The mid-line "---" inside a description survived the split; the
	// conversion webhook points at the release's service.
	if !strings.Contains(crds[1].YAML, "rwx---r--") {
		t.Fatal("description containing --- was corrupted by the document split")
	}
	spec := widgets["spec"].(map[string]interface{})
	service := spec["conversion"].(map[string]interface{})["webhook"].(map[string]interface{})["clientConfig"].(map[string]interface{})["service"].(map[string]interface{})
	if service["name"] != "my-op-webhook" || service["namespace"] != "ops" {
		t.Fatalf("conversion webhook not derived from release identity: %v", service)
	}
	// Types survive the stamping round trip: the port stays a number.
	if _, isNumber := service["port"].(float64); !isNumber {
		t.Fatalf("port type changed through stamping: %T", service["port"])
	}
}

func TestDeriveTemplatedCRDsDropsCapabilityGatedContentWithoutAPIVersion(t *testing.T) {
	src := Source{
		Chart:       fixtureChart(t, "templated-crds"),
		Version:     "1.2.3",
		CRDOverride: "crds:\n  create: true\n",
		// No cert-manager.io/v1 declared: the chart omits the annotation.
	}
	derived, err := Derive(context.Background(), src, Policy{ExpectCRDs: true}, "op", "ns")
	if err != nil {
		t.Fatal(err)
	}
	widgets := decode(t, derived.Owned[1].YAML)
	annotations := metadataMap(t, widgets, "annotations")
	if _, present := annotations["cert-manager.io/inject-ca-from"]; present {
		t.Fatal("capability-gated annotation rendered without the API version declared")
	}
}

func TestDeriveCRDsDirectoryChart(t *testing.T) {
	src := Source{Chart: fixtureChart(t, "crds-dir"), Version: "0.7.1"}
	derived, err := Derive(context.Background(), src, Policy{ExpectCRDs: true}, "op", "ns")
	if err != nil {
		t.Fatal(err)
	}
	crds := derived.Owned
	if len(crds) != 1 || crds[0].Name != "platforms.fixture.planton.ai" {
		t.Fatalf("expected the one crds/-directory CRD, got %+v", crds)
	}
}

func TestDeriveNoCRDsExplainsItself(t *testing.T) {
	src := Source{Chart: fixtureChart(t, "no-crds"), Version: "0.1.0", CRDOverride: "crds:\n  create: true\n"}
	_, err := Derive(context.Background(), src, Policy{ExpectCRDs: true}, "op", "ns")
	assertFailure(t, err, "produced no CustomResourceDefinition documents", "CRD switch", "helm show values")
}

func TestDeriveCRDSwitchOffYieldsNothing(t *testing.T) {
	// Without the override the templated chart renders no CRDs: the
	// override is load-bearing, and forgetting it is reported, not silent.
	src := Source{Chart: fixtureChart(t, "templated-crds"), Version: "1.2.3"}
	_, err := Derive(context.Background(), src, Policy{ExpectCRDs: true}, "op", "ns")
	if err == nil {
		t.Fatal("expected the zero-CRD failure")
	}
}

// The generic Helm kind renders whatever chart the user named, with no CRD
// override, so ownership follows Helm's two surfaces: the crds/ directory is
// the module's; templated CRDs are Helm's, refused unless the chart keeps them
// itself or the user accepted Helm-managed CRDs.
func TestDeriveWithoutOverrideSplitsTheTwoSurfaces(t *testing.T) {
	chart := fixtureChart(t, "mixed-surfaces")

	t.Run("templated CRD without keep is refused with its name", func(t *testing.T) {
		_, err := Derive(context.Background(), Source{Chart: chart, Version: "2.0.0"}, Policy{}, "rel", "ns")
		assertFailure(t, err, "gizmos.fixture.planton.ai", "deletes them when the release is uninstalled", "allow_helm_managed")
		if !strings.Contains(err.Error(), "1 CustomResourceDefinition(s)") || strings.Contains(err.Error(), "things.fixture") {
			t.Fatalf("the refusal must name only the templated CRD: %s", err)
		}
	})

	t.Run("accepted Helm-managed CRDs are reported, the directory CRD is owned", func(t *testing.T) {
		derived, err := Derive(context.Background(), Source{Chart: chart, Version: "2.0.0"}, Policy{AllowHelmManaged: true}, "rel", "ns")
		if err != nil {
			t.Fatal(err)
		}
		if len(derived.Owned) != 1 || derived.Owned[0].Name != "things.fixture.planton.ai" {
			t.Fatalf("expected only the crds/-directory CRD to be owned, got %+v", derived.Owned)
		}
		if len(derived.HelmManaged) != 1 || derived.HelmManaged[0] != "gizmos.fixture.planton.ai" {
			t.Fatalf("expected the templated CRD reported as Helm-managed, got %v", derived.HelmManaged)
		}
	})

	t.Run("a templated CRD the chart keeps itself is neither refused nor owned", func(t *testing.T) {
		src := Source{Chart: chart, Version: "2.0.0", Values: []string{"crds:\n  keep: true\n"}}
		derived, err := Derive(context.Background(), src, Policy{}, "rel", "ns")
		if err != nil {
			t.Fatalf("a chart that owns its CRD lifecycle must not be refused: %v", err)
		}
		if len(derived.Owned) != 1 || len(derived.HelmManaged) != 0 {
			t.Fatalf("expected one owned, none Helm-managed, got %+v", derived)
		}
	})

	t.Run("with an override the module owns both surfaces", func(t *testing.T) {
		src := Source{Chart: chart, Version: "2.0.0", CRDOverride: "crds:\n  keep: false\n"}
		derived, err := Derive(context.Background(), src, Policy{ExpectCRDs: true}, "rel", "ns")
		if err != nil {
			t.Fatal(err)
		}
		if len(derived.Owned) != 2 || len(derived.HelmManaged) != 0 {
			t.Fatalf("expected both CRDs owned, got %+v", derived)
		}
	})

	t.Run("a chart without CRDs is ordinary when none are expected", func(t *testing.T) {
		derived, err := Derive(context.Background(), Source{Chart: fixtureChart(t, "no-crds"), Version: "0.1.0"}, Policy{}, "rel", "ns")
		if err != nil {
			t.Fatal(err)
		}
		if len(derived.Owned) != 0 {
			t.Fatalf("expected nothing owned, got %+v", derived.Owned)
		}
	})
}

// A private repository must be read with the same credentials the install
// uses; the render carries them through Helm's own chart locator.
func TestDeriveFromPrivateRepositoryUsesCredentials(t *testing.T) {
	repoDir := t.TempDir()
	chrt, err := loader.Load(fixtureChart(t, "crds-dir"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := chartutil.Save(chrt, repoDir); err != nil {
		t.Fatal(err)
	}
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user, pass, ok := r.BasicAuth(); !ok || user != "reader" || pass != "s3cret" {
			w.Header().Set("WWW-Authenticate", `Basic realm="charts"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.URL.Path == "/index.yaml" {
			index, err := repo.IndexDirectory(repoDir, server.URL)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			index.SortEntries()
			body, _ := yaml.Marshal(index)
			_, _ = w.Write(body)
			return
		}
		http.ServeFile(w, r, filepath.Join(repoDir, filepath.Base(r.URL.Path)))
	}))
	defer server.Close()

	src := Source{Repository: server.URL, Chart: "crds-dir", Version: "0.7.1", Username: "reader", Password: "s3cret"}
	derived, err := Derive(context.Background(), src, Policy{ExpectCRDs: true}, "op", "ns")
	if err != nil {
		t.Fatalf("credentials must reach the repository: %v", err)
	}
	if len(derived.Owned) != 1 {
		t.Fatalf("expected the chart's one CRD, got %+v", derived.Owned)
	}
	if _, err := Derive(context.Background(), Source{Repository: server.URL, Chart: "crds-dir", Version: "0.7.1"}, Policy{ExpectCRDs: true}, "op", "ns"); err == nil {
		t.Fatal("without credentials the private repository must refuse the render")
	}
}

func TestDeriveBundle(t *testing.T) {
	bundle, err := os.ReadFile(filepath.Join("testdata", "bundle.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/solr-operator/v0.9.1/crds/all.yaml":
			_, _ = w.Write(bundle)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	src := Source{
		Version:      "0.9.1",
		VersionField: "spec.chart_version",
		BundleURL:    server.URL + "/solr-operator/v{{version}}/crds/all.yaml",
		BundleName:   "solr-operator",
	}
	derived, err := Derive(context.Background(), src, Policy{ExpectCRDs: true}, "solr", "ns")
	if err != nil {
		t.Fatal(err)
	}
	crds := derived.Owned
	if len(crds) != 2 {
		t.Fatalf("expected 2 CRDs (the Namespace document filtered out), got %d", len(crds))
	}
	zk := decode(t, crds[1].YAML)
	labels := metadataMap(t, zk, "labels")
	if labels["upstream"] != "kept" {
		t.Fatal("stamping dropped an upstream label")
	}
	if labels[LabelSource] != "solr-operator" {
		t.Fatalf("bundle source label should be the bundle's own name, got %v", labels[LabelSource])
	}

	// 404 at a version a user chose: the setting is the remedy.
	missing := src
	missing.Version = "0.9.0"
	_, err = Derive(context.Background(), missing, Policy{ExpectCRDs: true}, "solr", "ns")
	assertFailure(t, err, "answered HTTP 404", "no bundle is published at this version", "set spec.chart_version to a published version")

	// 404 at a version the module pins: no setting can fix it, so the text
	// names the pinned release and never sends the reader to a field.
	pinned := missing
	pinned.VersionField = ""
	_, err = Derive(context.Background(), pinned, Policy{ExpectCRDs: true}, "solr", "ns")
	assertFailure(t, err, "answered HTTP 404", "solr-operator release 0.9.0, the release this module pins", "correct the module's bundle URL template or its pin")
	if strings.Contains(err.Error(), "spec.") {
		t.Fatalf("a module-pinned source must not name a spec field: %s", err)
	}
}

// Two upstreams that publish from one host are two owners. A label derived
// from the download host would read them as one, and the foreign-owner
// refusal would never fire between their kinds.
func TestBundlesFromOneHostAreDistinctOwners(t *testing.T) {
	sandbox := Source{Version: "v1.0.7", BundleURL: "https://github.com/kubernetes-sigs/agent-sandbox/releases/download/{{version}}/sandbox.yaml", BundleName: "agent-sandbox"}
	other := Source{Version: "v2.0.0", BundleURL: "https://github.com/example/other/releases/download/{{version}}/crds.yaml", BundleName: "other"}
	stampedByOther := ExistingFromObject("sandboxes.agents.x-k8s.io",
		map[string]string{LabelSource: labelSourceValue(other)},
		map[string]string{AnnotationSourceVersion: "v2.0.0"}, nil)
	assertFailure(t, CheckOwnership([]ExistingCRD{stampedByOther}, sandbox), "derived from other by another Planton module", "", "")
}

// A module author's mistakes in assembling a source are refused before any
// fetch or render, so no half-stamped CRD can ever reach a cluster.
func TestDeriveRefusesAMalformedSource(t *testing.T) {
	cases := map[string]Source{
		"bundle without a name":        {Version: "v1.0.0", BundleURL: "https://example.invalid/{{version}}/crds.yaml"},
		"bundle name not a label":      {Version: "v1.0.0", BundleURL: "https://example.invalid/{{version}}/crds.yaml", BundleName: "kubernetes-sigs/agent-sandbox"},
		"both a chart and a bundle":    {Version: "v1.0.0", BundleURL: "https://example.invalid/{{version}}/crds.yaml", BundleName: "x", Chart: "x"},
		"neither a chart nor a bundle": {Version: "1.0.0"},
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Derive(context.Background(), src, Policy{ExpectCRDs: true}, "r", "ns"); err == nil {
				t.Fatal("a malformed source must be refused")
			}
		})
	}
}

func TestClassifyLocateErrors(t *testing.T) {
	http := Source{Repository: "https://charts.example.invalid", Chart: "op", Version: "9.9.9"}
	oci := Source{Repository: "oci://ghcr.io/example/charts", Chart: "op", Version: "9.9.9"}

	cases := []struct {
		name     string
		src      Source
		raw      string
		observed string
		next     string
	}{
		{"version not published", http,
			`chart "op" version "9.9.9" not found in https://charts.example.invalid repository`,
			"is not in the index of", "index.yaml"},
		{"repository unreachable", http,
			`looks like "https://charts.example.invalid" is not a valid chart repository or cannot be reached: Get "https://charts.example.invalid/index.yaml": dial tcp: lookup charts.example.invalid: no such host`,
			"could not be reached", "curl -I"},
		{"oci version not published", oci,
			`failed to perform "FetchReference" on source: ghcr.io/example/charts/op:9.9.9: not found`,
			"OCI registry has no chart", "helm show chart"},
		{"stale local repository cache", http,
			`no cached repo found. (try 'helm repo update'): open /home/x/.cache/helm/repository/neo4j-index.yaml: no such file or directory`,
			"repository index on this machine", "helm repo update"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := classifyLocateError(tc.src, errors.New(tc.raw))
			assertFailure(t, err, tc.observed, "", tc.next)
			if !strings.Contains(err.Error(), tc.raw) {
				t.Fatal("the raw Helm text must stay visible inside the observation")
			}
		})
	}
}

func TestCheckNoDowngrade(t *testing.T) {
	existing := []ExistingCRD{
		{Name: "a.example", Version: "0.120.0"},
		{Name: "b.example", Version: ""}, // unstamped: never a downgrade
	}
	chart := func(version string) Source {
		return Source{Repository: "https://charts.example.com", Chart: "demo", Version: version, VersionField: "spec.chart_version"}
	}
	if err := CheckNoDowngrade(existing, chart("0.120.0")); err != nil {
		t.Fatalf("same version must pass: %v", err)
	}
	if err := CheckNoDowngrade(existing, chart("0.120.3")); err != nil {
		t.Fatalf("upgrade must pass: %v", err)
	}
	err := CheckNoDowngrade(existing, chart("0.119.0"))
	assertFailure(t, err, "derived from chart version 0.120.0; the manifest asks for chart version 0.119.0", "strip fields", "set spec.chart_version to 0.120.0 or higher")
	if !strings.Contains(err.Error(), "kubectl delete crd a.example") {
		t.Fatalf("remedy must name the exact CRD: %s", err)
	}
	if err := CheckNoDowngrade(nil, chart("0.1.0")); err != nil {
		t.Fatalf("first install has nothing to compare: %v", err)
	}
	assertFailure(t, CheckNoDowngrade(existing, chart("latest")), "the requested chart version \"latest\" is not a semantic version", "", "0.120.0")

	// A generic kind whose version field is named differently is sent to
	// that field, never to a field it does not have.
	generic := chart("0.119.0")
	generic.VersionField = "spec.version"
	assertFailure(t, CheckNoDowngrade(existing, generic), "", "", "set spec.version to 0.120.0 or higher")

	// A module-pinned upstream release, written as its tag: the leading "v"
	// orders like any version, and the remedy is a Planton release.
	release := []ExistingCRD{{Name: "sandboxes.agents.x-k8s.io", Version: "v1.0.8"}}
	pinned := func(version string) Source {
		return Source{Version: version, BundleURL: "https://example.invalid/{{version}}/sandbox.yaml", BundleName: "agent-sandbox"}
	}
	if err := CheckNoDowngrade(release, pinned("v1.0.8")); err != nil {
		t.Fatalf("same release must pass: %v", err)
	}
	err = CheckNoDowngrade(release, pinned("v1.0.7"))
	assertFailure(t, err,
		"derived from agent-sandbox release v1.0.8; this module pins agent-sandbox release v1.0.7",
		"strip fields",
		"deploy with a Planton release whose module pins agent-sandbox release v1.0.8 or higher (this module pins v1.0.7)")
	if strings.Contains(err.Error(), "spec.") || strings.Contains(err.Error(), "chart") {
		t.Fatalf("a module-pinned release must not be spoken of as a chart or a setting: %s", err)
	}
	assertFailure(t, CheckNoDowngrade(release, pinned("main")), "the requested agent-sandbox release \"main\" is not a semantic version", "", "correct the module's pin")
}

func TestCheckOwnership(t *testing.T) {
	src := Source{Repository: "https://charts.example.com", Chart: "demo", Version: "1.2.0"}
	ours := ExistingFromObject("a.example",
		map[string]string{LabelSource: "demo"},
		map[string]string{AnnotationSourceVersion: "1.1.0", AnnotationSourceChart: "https://charts.example.com/demo"},
		[]string{"kubectl"})
	if err := CheckOwnership([]ExistingCRD{ours}, src); err != nil {
		t.Fatalf("a CRD this source stamped is ours to re-apply: %v", err)
	}
	if err := CheckOwnership(nil, src); err != nil {
		t.Fatalf("a first install finds nothing: %v", err)
	}

	// Helm owns it as a release resource: the owner is the release, and the
	// hand-over remedy first frees it from Helm.
	helmOwned := ExistingFromObject("a.example", nil,
		map[string]string{"meta.helm.sh/release-name": "demo-by-hand", "meta.helm.sh/release-namespace": "ops"},
		[]string{"helm"})
	err := CheckOwnership([]ExistingCRD{helmOwned}, src)
	assertFailure(t, err, "owned by Helm release demo-by-hand in namespace ops", "two owners would write one schema", "spec.crds.install to false")
	for _, want := range []string{"uninstall Helm release demo-by-hand", "kubectl label crd a.example planton.ai/crd-source=demo", "planton.ai/crd-source-version=1.2.0"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("remedy lacks %q: %s", want, err)
		}
	}

	// Another Planton module derived the same CRD name from a different chart.
	otherSource := ExistingFromObject("a.example", map[string]string{LabelSource: "other-chart"},
		map[string]string{AnnotationSourceVersion: "9.0.0"}, nil)
	assertFailure(t, CheckOwnership([]ExistingCRD{otherSource}, src), "derived from other-chart by another Planton module", "", "")

	// Applied by hand: the field managers are the only mark.
	byHand := ExistingFromObject("a.example", nil, nil, []string{"kubectl-client-side-apply", "kube-apiserver"})
	err = CheckOwnership([]ExistingCRD{byHand}, src)
	assertFailure(t, err, "last written by field manager(s) kubectl-client-side-apply, kube-apiserver", "", "")
	if strings.Contains(err.Error(), "uninstall Helm release") {
		t.Fatal("a CRD Helm does not own must not be sent to helm uninstall")
	}

	// No marks at all: say so rather than invent an owner.
	assertFailure(t, CheckOwnership([]ExistingCRD{{Name: "a.example"}}, src), "it carries no ownership marks", "", "")

	// A bundle kind installs the upstream's own workloads, so leaving the
	// definitions with their owner is right only when no other installation
	// of that upstream runs; the remedy says so and speaks of the release.
	bundle := Source{Version: "v1.0.7", BundleURL: "https://example.invalid/{{version}}/sandbox.yaml", BundleName: "agent-sandbox"}
	err = CheckOwnership([]ExistingCRD{byHand}, bundle)
	assertFailure(t, err, "was not applied by this module", "newer or older than agent-sandbox release v1.0.7",
		"set spec.crds.install to false only when another tool manages these definitions and no other installation of agent-sandbox runs on this cluster")
	for _, want := range []string{"hand them to this module once you know they match agent-sandbox release v1.0.7", "planton.ai/crd-source=agent-sandbox"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("remedy lacks %q: %s", want, err)
		}
	}
	if strings.Contains(err.Error(), "chart version") {
		t.Fatalf("a bundle refusal must not speak of a chart version: %s", err)
	}
}

func TestCRDApplyDeniedFailure(t *testing.T) {
	err := CRDApplyDeniedFailure("system:serviceaccount:ci:deployer", "patch", "SelfSubjectAccessReview answered denied")
	assertFailure(t, err, "system:serviceaccount:ci:deployer) may not patch customresourcedefinitions.apiextensions.k8s.io at the cluster scope", "as kept resources apart from the release or manifest it installs", "iac/permissions.yaml")
}

func TestFailureRendersThreeStableLines(t *testing.T) {
	f := &Failure{Observed: "o", Meaning: "m", NextStep: "n"}
	if f.Error() != "observed: o\nmeaning: m\nnext step: n" {
		t.Fatalf("unexpected rendering: %q", f.Error())
	}
}

func TestSplitDocumentsIgnoresCommentOnlyFragments(t *testing.T) {
	docs := splitDocuments("# license\n# header\n---\n---\nkind: A\n---\n# Source: x\n---\nkind: B\ndescription: a---b\n")
	if len(docs) != 2 {
		t.Fatalf("expected 2 documents, got %d: %q", len(docs), docs)
	}
	if !strings.Contains(docs[1], "a---b") {
		t.Fatal("mid-line --- must not split")
	}
}

func assertFailure(t *testing.T, err error, observed, meaning, next string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected a failure")
	}
	var failure *Failure
	if !errors.As(err, &failure) {
		t.Fatalf("expected a *Failure, got %T: %v", err, err)
	}
	if observed != "" && !strings.Contains(failure.Observed, observed) {
		t.Fatalf("observed %q lacks %q", failure.Observed, observed)
	}
	if meaning != "" && !strings.Contains(failure.Meaning, meaning) {
		t.Fatalf("meaning %q lacks %q", failure.Meaning, meaning)
	}
	if next != "" && !strings.Contains(failure.NextStep, next) {
		t.Fatalf("next step %q lacks %q", failure.NextStep, next)
	}
	for _, part := range []string{failure.Observed, failure.Meaning, failure.NextStep} {
		if strings.TrimSpace(part) == "" {
			t.Fatalf("every failure carries all three parts: %+v", failure)
		}
	}
}

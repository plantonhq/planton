package setdeploy

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// The behavior corpus (pkg/manifestgraph/testdata/corpus) pins the SHARED
// semantics: edges, order, classification. This test pins the OFFLINE
// POLICY over those semantics — which corpus scenarios a backendless deploy
// refuses, and which it deploys with stated assumptions. A change here is a
// policy change and must be made deliberately, in the open.
func TestCorpusOfflinePolicy(t *testing.T) {
	// Scenario -> whether the offline wall refuses it. Probes are faked to
	// all-verified so only the input decides. Two fixtures refuse at the
	// SCHEMA check rather than a graph check: the corpus pins graph semantics
	// and keeps some real-kind fixtures deliberately schema-minimal — the
	// wall validates schema too, and a document that fails to load also
	// leaves the healthy remainder (making its dependents' references
	// external). That interplay is itself pinned behavior.
	policy := map[string]bool{
		"annotation-riding-ref":             false,
		"connection-placement":              true, // schema-minimal cluster fixture: refuses at load-and-schema
		"connection-placement-default-name": true, // schema-minimal cluster fixture: refuses at load-and-schema
		"cycle":                             true,
		"derived-edge-cycle":                false, // the inference yields to the author's order — a stated fact, not a refusal
		"derived-namespace":                 false, // a derived target is a stated fact, not a refusal
		"duplicate-identity":                true,
		"env-external":                      true, // needs a value from another environment — no backend to find it
		"explicit-kind-ref":                 false,
		"external-relationship":             false, // ordering fact only — deploys with a stated assumption
		"external-valuefrom":                true,  // needs a value the set cannot produce
		"literal-sibling":                   false,
		"literal-sibling-no-match":          false,
		"map-ref":                           false,
		"namespace-edge":                    true, // schema-minimal fixture: refuses at load-and-schema
		"operator-prerequisite":             false,
		"operator-prerequisite-ambiguous":   false,
		"phantom-node-slug":                 false,
		"ref-rule-violation":                true,
		"relationships-edge":                false,
		"two-node-real-kinds":               true, // schema-minimal producer refuses at load; its dependent's ref becomes external
	}

	corpusDir := filepath.Join("..", "manifestgraph", "testdata", "corpus")
	entries, err := os.ReadDir(corpusDir)
	if err != nil {
		t.Fatalf("reading corpus: %v", err)
	}

	seen := map[string]bool{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		scenario := entry.Name()
		seen[scenario] = true
		expectRefusal, known := policy[scenario]
		if !known {
			t.Errorf("corpus scenario %q has no offline policy — decide and pin whether the backendless lane refuses it", scenario)
			continue
		}
		t.Run(scenario, func(t *testing.T) {
			docs := corpusScenarioDocs(t, filepath.Join(corpusDir, scenario))
			plan := Preflight(docs, Flags{}, newFakeProbes())
			if plan.Report.Refused() != expectRefusal {
				t.Fatalf("offline policy for %q: expected refused=%v; report:\n%s",
					scenario, expectRefusal, reportDump(plan.Report))
			}
		})
	}
	for scenario := range policy {
		if !seen[scenario] {
			t.Errorf("policy names corpus scenario %q which no longer exists", scenario)
		}
	}
}

func corpusScenarioDocs(t *testing.T, dir string) []Doc {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading scenario %s: %v", dir, err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || e.Name() == "golden.yaml" || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	var docs []Doc
	for _, name := range names {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("reading %s: %v", name, err)
		}
		fileDocs, err := CollectDocsFromBytes(b, name)
		if err != nil {
			t.Fatalf("splitting %s: %v", name, err)
		}
		docs = append(docs, fileDocs...)
	}
	return docs
}

func reportDump(report *Report) string {
	var b strings.Builder
	for _, check := range report.Checks {
		for _, e := range check.Entries {
			b.WriteString(string(e.Severity) + " [" + check.Name + "] " + e.Message + "\n")
		}
	}
	return b.String()
}

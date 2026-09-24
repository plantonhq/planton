//go:build !codegen
// +build !codegen

package providerparity

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRenderPublicReport_Hermetic proves the renderer over the shared
// fixture: deterministic bytes, the honest claim language, the embedded
// generation parameters, both accounting directions surfaced, and proof
// status rendered only from the joined map.
func TestRenderPublicReport_Hermetic(t *testing.T) {
	spec, modules, schemas, manifests, ledger := accountingFixture()
	// Drop the fixture's deliberately-stale ledger rows: the render fixture
	// wants a clean breadth story.
	ledger = ledger[:1]
	acc := buildAccounting("gcp", spec, modules, schemas, "google", manifests, ledger, nil)
	rep := buildReport("gcp", spec, modules, schemas)
	proofs := map[string]E2EProof{
		"TestWidget": {Green: true, Engines: []string{"pulumi", "terraform"}},
		"TestPlain":  {Green: true, Engines: []string{"pulumi"}},
	}

	page := RenderPublicReport(rep, acc, proofs)
	if page != RenderPublicReport(rep, acc, proofs) {
		t.Fatal("render is not deterministic")
	}

	providerName, gaSchema, err := ParseReportParams(page)
	if err != nil || providerName != "gcp" || gaSchema != "google" {
		t.Errorf("embedded parameters = %q/%q (%v), want gcp/google", providerName, gaSchema, err)
	}

	for _, want := range []string{
		"GENERATED FILE -- DO NOT EDIT",
		"built for 100% Terraform parity",
		"`google@6.50.0`",
		// Green with both engines is proven; green with one is partial.
		"| TestWidget | 12 | 5 | 3 | 4 | 1 | ❌ | ✅ pulumi, terraform |",
		"| TestPlain | 1 | 0 | 0 | 0 | 2 | ❌ | partial: pulumi |",
		"### Composed (1)",
		"| `google_composed_thing` | covered by TestWidget's composed_thing field |",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("rendered page misses %q", want)
		}
	}
	if strings.Contains(page, "UNDISPOSITIONED") == (acc.DispositionTotals[""] == 0) {
		t.Errorf("undispositioned row rendered inconsistently with totals %v", acc.DispositionTotals)
	}
}

// TestRenderPublicReport_BaselineListsOnlyOwnSchemas proves the measurement
// baseline table is scoped to the page's own provider: the GA baseline and
// any schema this catalog's modules pin appear; schema artifacts committed
// for OTHER providers' catalogs never leak onto the page.
func TestRenderPublicReport_BaselineListsOnlyOwnSchemas(t *testing.T) {
	spec, modules, schemas, manifests, ledger := accountingFixture()
	ledger = ledger[:1]
	// A sibling catalog's artifact (loaded because schemas are loaded as a
	// set) and an unpinned secondary channel for our own provider: neither
	// is a yardstick for this page.
	schemas["aws"] = &Schema{Provider: "aws", Source: "hashicorp/aws", Version: "6.58.0"}
	schemas["google-beta"] = &Schema{Provider: "google-beta", Source: "hashicorp/google-beta", Version: "6.50.0"}
	acc := buildAccounting("gcp", spec, modules, schemas, "google", manifests, ledger, nil)
	rep := buildReport("gcp", spec, modules, schemas)

	page := RenderPublicReport(rep, acc, nil)
	if !strings.Contains(page, "`google@6.50.0`") {
		t.Error("GA baseline schema missing from the measurement baseline table")
	}
	for _, leaked := range []string{"aws@6.58.0", "google-beta@6.50.0"} {
		if strings.Contains(page, leaked) {
			t.Errorf("unpinned schema %q leaked onto the page", leaked)
		}
	}

	// A pin admits the schema back: the beta channel returns to the table
	// the moment a module pins it.
	if modules[0].Pins == nil {
		modules[0].Pins = map[string]string{}
	}
	modules[0].Pins["google-beta"] = "~> 6.0"
	rep = buildReport("gcp", spec, modules, schemas)
	if page := RenderPublicReport(rep, acc, nil); !strings.Contains(page, "`google-beta@6.50.0`") {
		t.Error("pinned secondary-channel schema missing from the measurement baseline table")
	}
}

// TestPublicReportDrift freezes the committed parity report pages: each
// catalog/<provider>/terraform-parity.md must be byte-identical to a fresh
// render from the tree, so the page can never be hand-edited or go stale.
// Each page embeds its own generation parameters, so enrollment is file
// presence -- no per-provider configuration here. Regenerate with
// PLANTON_REGEN_PROVIDERPARITY_REPORT=1 (the Makefile's
// generate-provider-parity-report target).
func TestPublicReportDrift(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, catalogRoot)); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test")
	}
	pages, err := filepath.Glob(filepath.Join(root, catalogRoot, "*", PublicReportFileName))
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) == 0 {
		t.Skip("no committed parity report pages yet")
	}
	schemas, err := LoadSchemas("schemas")
	if err != nil {
		t.Fatalf("committed schemas: %v", err)
	}
	for _, page := range pages {
		committed, err := os.ReadFile(page)
		if err != nil {
			t.Fatal(err)
		}
		providerName, gaSchema, err := ParseReportParams(string(committed))
		if err != nil {
			t.Errorf("%s: %v", page, err)
			continue
		}
		if dir := filepath.Base(filepath.Dir(page)); dir != providerName {
			t.Errorf("%s: embedded provider %q does not match its directory %q", page, providerName, dir)
			continue
		}
		provider, err := providerFromName(providerName)
		if err != nil {
			t.Errorf("%s: %v", page, err)
			continue
		}
		fresh, err := GeneratePublicReport(root, provider, schemas, gaSchema, "", "")
		if err != nil {
			t.Errorf("%s: regenerating: %v", page, err)
			continue
		}
		if os.Getenv("PLANTON_REGEN_PROVIDERPARITY_REPORT") == "1" {
			if err := os.WriteFile(page, []byte(fresh), 0o644); err != nil {
				t.Fatalf("regenerating %s: %v", page, err)
			}
			t.Logf("%s regenerated -- review the diff before committing", page)
			continue
		}
		if string(committed) != fresh {
			t.Errorf("%s is stale or hand-edited -- regenerate with the Makefile's generate-provider-parity-report target (never edit by hand)", page)
		}
	}
}

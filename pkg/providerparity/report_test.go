//go:build !codegen
// +build !codegen

package providerparity

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/plantonhq/planton/shared/cloudresourcekind"
)

func TestBuildReport(t *testing.T) {
	schemas := map[string]*Schema{"google": fixtureSchema()}
	spec := []KindCensus{
		{Kind: "GcpGcsBucket", SpecFieldPaths: []string{"spec.name", "spec.location", "spec.labels"}},
	}
	modules := []ModuleCensus{
		{
			Kind:      "GcpGcsBucket",
			ModuleDir: "catalog/gcp/gcpgcsbucket/iac/tf",
			Resources: []string{"google_storage_bucket", "google_made_up_thing", "random_password"},
			Pins:      map[string]string{"google": "~> 6.0"},
		},
	}

	// random_password is judged internal by the kind's manifest: plumbing
	// no cloud schema serves, never an unknown.
	r := buildReport("gcp", spec, modules, schemas, map[string]map[string]bool{"GcpGcsBucket": {"random_password": true}})

	if r.Kinds != 1 || r.TotalSpecFields != 3 || r.DistinctResources != 3 {
		t.Errorf("aggregates = kinds %d, specFields %d, distinct %d; want 1, 3, 3",
			r.Kinds, r.TotalSpecFields, r.DistinctResources)
	}
	// Only the known resource's non-deprecated configurable surface counts.
	if r.TotalConfigurableArgs != 5 {
		t.Errorf("TotalConfigurableArgs = %d, want 5", r.TotalConfigurableArgs)
	}
	// A resource the pinned schema does not serve is surfaced, never dropped.
	if len(r.UnknownResources) != 1 || r.UnknownResources[0] != "google_made_up_thing" {
		t.Errorf("UnknownResources = %v, want [google_made_up_thing]", r.UnknownResources)
	}
	if r.SchemaVersions["google"] != "6.50.0" {
		t.Errorf("SchemaVersions = %v", r.SchemaVersions)
	}
	if r.PinDistribution["google"]["~> 6.0"] != 1 {
		t.Errorf("PinDistribution = %v", r.PinDistribution)
	}
	kr := r.KindReports[0]
	if kr.Resources[0].Schema != "google" || kr.Resources[0].ConfigurableArgs != 5 {
		t.Errorf("resource resolution = %+v", kr.Resources[0])
	}
	if kr.Resources[1].Schema != "" {
		t.Errorf("unknown resource resolved to %q, want empty", kr.Resources[1].Schema)
	}
}

// TestGcpMeasuredReport is the live measurement over the whole GCP catalog
// against the committed schema artifacts. It guards the joined pipeline
// (three censuses + committed schemas) and logs the catalog-level numbers --
// the same aggregation the CLI and the public report generator render.
func TestGcpMeasuredReport(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, catalogRoot)); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test")
	}
	schemas, err := LoadSchemas("schemas")
	if err != nil {
		t.Fatalf("committed schemas: %v", err)
	}

	r, err := BuildReport(root, cloudresourcekind.CloudResourceProvider_gcp, schemas)
	if err != nil {
		t.Fatalf("report: %v", err)
	}

	// A consumed resource no loaded schema serves means a module drifted
	// from its pin (or the artifacts are stale) -- either way the parity
	// numbers would be lies. Fail, and name the resources.
	if len(r.UnknownResources) > 0 {
		t.Errorf("resources consumed by modules but absent from every pinned schema: %v", r.UnknownResources)
	}

	t.Logf("gcp catalog: %d kinds, %d spec fields, %d distinct resources consumed, %d configurable args across them (schemas: %v)",
		r.Kinds, r.TotalSpecFields, r.DistinctResources, r.TotalConfigurableArgs, r.SchemaVersions)
	for name, dist := range r.PinDistribution {
		t.Logf("pin distribution %s: %v", name, dist)
	}

	if raw, err := json.MarshalIndent(r, "", "  "); err == nil && os.Getenv("PLANTON_PROVIDERPARITY_DUMP") != "" {
		path := filepath.Join(os.TempDir(), "providerparity-gcp-report.json")
		if err := os.WriteFile(path, raw, 0o644); err == nil {
			t.Logf("full report written to %s", path)
		}
	}

	// Depth ordering sanity: the kinds the audit found deepest in debt exist
	// and are measured (their numbers move with the depth waves).
	measured := map[string]bool{}
	for _, kr := range r.KindReports {
		measured[kr.Kind] = true
	}
	for _, giant := range []string{"GcpGkeCluster", "GcpCloudSql", "GcpUrlMap"} {
		if !measured[giant] {
			t.Errorf("expected kind %s missing from the measurement", giant)
		}
	}

	sorted := sort.SliceIsSorted(r.KindReports, func(i, j int) bool {
		return r.KindReports[i].Kind < r.KindReports[j].Kind
	})
	if !sorted {
		t.Error("KindReports must be deterministically sorted by kind")
	}
}

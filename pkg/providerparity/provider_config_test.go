//go:build !codegen
// +build !codegen

package providerparity

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"gopkg.in/yaml.v3"
)

// --- HCL provider-block extraction (module census) ---

// TestScanModule_ProviderBlockArgs proves the HCL walk against the shapes the
// live catalog actually carries: an empty canonical block (no entry), a
// populated nested block (azurerm's features), a flat attribute, and an
// empty-but-present nested block (the bare features {} every azurerm module
// declares) -- the shape a regex reader can never see into.
func TestScanModule_ProviderBlockArgs(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"provider.tf": `provider "aws" {
  # canonical empty block -- must yield NO entry
}

provider "azurerm" {
  features {
    machine_learning {
      purge_soft_deleted_workspace_on_destroy = true
    }
  }
}

provider "google" {
  user_project_override = true
}
`,
		"empty_features.tf": `provider "azurerm2" {
  features {}
}
`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	census, err := ScanModule(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if _, ok := census.ProviderBlockArgs["aws"]; ok {
		t.Error("empty provider block must yield no entry")
	}
	want := map[string][]string{
		"azurerm":  {"features.machine_learning.purge_soft_deleted_workspace_on_destroy"},
		"azurerm2": {"features"},
		"google":   {"user_project_override"},
	}
	for name, args := range want {
		if !reflect.DeepEqual(census.ProviderBlockArgs[name], args) {
			t.Errorf("%s args = %v, want %v", name, census.ProviderBlockArgs[name], args)
		}
	}
}

func TestScanModule_InvalidHclFailsLoudly(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "broken.tf"), []byte(`provider "aws" {`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ScanModule(dir); err == nil {
		t.Error("unparseable HCL must fail the census, not silently under-report")
	}
}

// --- provider-config census ---

func TestProviderConfigCensus_Aws(t *testing.T) {
	paths, err := ProviderConfigCensus(cloudresourcekind.CloudResourceProvider_aws)
	if err != nil {
		t.Fatalf("census: %v", err)
	}
	for _, want := range []string{
		"config.account_id",
		"config.assume_role_chain.role_arn",
		"config.assume_role_chain.transitive_tag_keys",
		"config.default_tags.tags",
		"config.endpoints",
		"config.max_retries",
		"config.retry_mode",
		"config.web_identity.web_identity_token",
	} {
		if !slices.Contains(paths, want) {
			t.Errorf("census is missing %s", want)
		}
	}
	if !slices.IsSorted(paths) {
		t.Error("census paths must be sorted")
	}
}

// --- manifest validation ---

func TestProviderConfigManifest_Validation(t *testing.T) {
	cases := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{"mapping must be config-rooted",
			"mappings:\n  - config: spec.foo\n    arg: foo\n", "config-rooted"},
		{"exclusion needs a reason",
			"exclusions:\n  - arg: foo\n", "no reason"},
		{"moduleOwned needs a reason",
			"moduleOwned:\n  - arg: foo\n", "no reason"},
		{"pattern must compile",
			"exclusionPatterns:\n  - pattern: '['\n    reason: r\n", "not a valid regular expression"},
		{"pattern needs a reason",
			"exclusionPatterns:\n  - pattern: _custom$\n", "no reason"},
		{"configExclusions must be config-rooted",
			"configExclusions:\n  - field: spec.foo\n    reason: r\n", "config-rooted"},
		{"an arg judged twice is refused",
			"exclusions:\n  - arg: foo\n    reason: a\nmoduleOwned:\n  - arg: foo\n    reason: b\n", "judged twice"},
		{"unknown fields are refused",
			"resources: {}\n", "not found in type"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var m ProviderConfigManifest
			dec := yaml.NewDecoder(strings.NewReader(tc.yaml))
			dec.KnownFields(true)
			err := dec.Decode(&m)
			if err == nil {
				err = m.validate()
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("want error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

// --- the accounting join ---

// fixtureProviderBlock is a small provider block exercising every judgment
// class: an exact-named attribute, a renamed one, a collapsible per-service
// block, a pattern class, a module-owned flag, and an excluded lever.
func fixtureProviderBlock() *Block {
	return &Block{
		Attributes: map[string]*Attribute{
			"region":                {Optional: true}, // exact match
			"token":                 {Optional: true}, // mapped
			"legacy_flag":           {Optional: true, Deprecated: true},
			"svc_a_custom_endpoint": {Optional: true}, // pattern class
			"svc_b_custom_endpoint": {Optional: true}, // pattern class
			"behavior_flag":         {Optional: true}, // module-owned
			"insecure":              {Optional: true}, // excluded
		},
		Blocks: map[string]*NestedBlock{
			"endpoints": {NestingMode: "single", Block: &Block{
				Attributes: map[string]*Attribute{
					"svc_a": {Optional: true},
					"svc_b": {Optional: true},
				},
			}},
		},
	}
}

func fixtureConfigPaths() []string {
	return []string{
		"config.api_token", // <- token
		"config.endpoints", // <- endpoints.* collapse
		"config.platform_identity",
		"config.region", // <- region exact
	}
}

func fixtureManifest() *ProviderConfigManifest {
	return &ProviderConfigManifest{
		Mappings: []ConfigMapping{
			{Config: "config.api_token", Arg: "token"},
			{Config: "config.endpoints", Arg: "endpoints", Collapse: true},
		},
		Exclusions: []ArgExclusion{
			{Arg: "insecure", Reason: "never offered"},
		},
		ExclusionPatterns: []PatternExclusion{
			{Pattern: "_custom_endpoint$", Reason: "per-service class"},
		},
		ModuleOwned: []ArgExclusion{
			{Arg: "behavior_flag", Reason: "modules own it"},
		},
		ConfigExclusions: []ConfigExclusion{
			{Field: "config.platform_identity", Reason: "platform concept"},
		},
	}
}

func TestProviderConfigAccounting_TotalAccounting(t *testing.T) {
	pc, findings := buildProviderConfigAccounting("test", fixtureConfigPaths(), fixtureProviderBlock(),
		fixtureManifest(), map[string][]string{"behavior_flag": {"KindA"}, "token": {"KindB"}})

	if len(findings) != 0 {
		t.Fatalf("want zero findings, got %v", findings)
	}
	if !pc.Accounted() {
		t.Fatalf("want accounted, got %+v", pc)
	}
	// 6 non-deprecated args + 2 endpoint leaves; legacy_flag is deprecated.
	if pc.TotalArgs != 8 {
		t.Errorf("TotalArgs = %d, want 8", pc.TotalArgs)
	}
	if pc.MatchedArgs != 1 { // region
		t.Errorf("MatchedArgs = %d, want 1", pc.MatchedArgs)
	}
	if pc.MappedArgs != 3 { // token + endpoints.svc_a + endpoints.svc_b
		t.Errorf("MappedArgs = %d, want 3", pc.MappedArgs)
	}
	if pc.ModuleOwnedArgs != 1 { // behavior_flag
		t.Errorf("ModuleOwnedArgs = %d, want 1", pc.ModuleOwnedArgs)
	}
	if pc.ExcludedArgs != 3 { // insecure + the two pattern-class endpoints
		t.Errorf("ExcludedArgs = %d, want 3", pc.ExcludedArgs)
	}
}

func TestProviderConfigAccounting_Gaps(t *testing.T) {
	manifest := fixtureManifest()
	// Drop the token mapping: the arg becomes unaccounted AND the config
	// field becomes uncovered AND KindB's module-set token loses judgment.
	manifest.Mappings = manifest.Mappings[1:]

	pc, findings := buildProviderConfigAccounting("test", fixtureConfigPaths(), fixtureProviderBlock(),
		manifest, map[string][]string{"behavior_flag": {"KindA"}, "token": {"KindB"}})

	if pc.Accounted() {
		t.Fatal("want unaccounted")
	}
	if len(pc.UnaccountedArgs) != 1 || pc.UnaccountedArgs[0] != "token" {
		t.Errorf("UnaccountedArgs = %v, want [token]", pc.UnaccountedArgs)
	}
	if len(pc.UncoveredConfigFields) != 1 || pc.UncoveredConfigFields[0] != "config.api_token" {
		t.Errorf("UncoveredConfigFields = %v, want [config.api_token]", pc.UncoveredConfigFields)
	}
	if len(pc.UnjudgedModuleArgs) != 1 || !strings.Contains(pc.UnjudgedModuleArgs[0], "KindB") {
		t.Errorf("UnjudgedModuleArgs = %v, want the KindB token leak", pc.UnjudgedModuleArgs)
	}
	for _, f := range findings {
		if f.BaselineKey != "provider:test" {
			t.Errorf("finding key = %q, want provider:test", f.BaselineKey)
		}
	}
	if len(findings) != 3 {
		t.Errorf("want 3 findings, got %d: %v", len(findings), findings)
	}
}

func TestProviderConfigAccounting_Staleness(t *testing.T) {
	manifest := fixtureManifest()
	manifest.Mappings = append(manifest.Mappings, ConfigMapping{Config: "config.gone", Arg: "gone_arg"})
	manifest.Exclusions = append(manifest.Exclusions, ArgExclusion{Arg: "gone_lever", Reason: "r"})
	manifest.ModuleOwned = append(manifest.ModuleOwned, ArgExclusion{Arg: "region", Reason: "no module sets it"})
	manifest.ExclusionPatterns = append(manifest.ExclusionPatterns, PatternExclusion{Pattern: "^never_matches$", Reason: "r"})
	manifest.ConfigExclusions = append(manifest.ConfigExclusions, ConfigExclusion{Field: "config.gone_field", Reason: "r"})

	pc, _ := buildProviderConfigAccounting("test", fixtureConfigPaths(), fixtureProviderBlock(),
		manifest, map[string][]string{"behavior_flag": {"KindA"}, "token": {"KindB"}})

	wantStale := []string{
		"mapping arg gone_arg",
		"mapping config config.gone",
		"exclusion gone_lever",
		"moduleOwned region is set by no module",
		"exclusionPatterns ^never_matches$",
		"configExclusions: config.gone_field",
	}
	for _, want := range wantStale {
		found := false
		for _, s := range pc.ManifestStale {
			if strings.Contains(s, want) {
				found = true
			}
		}
		if !found {
			t.Errorf("ManifestStale is missing %q: %v", want, pc.ManifestStale)
		}
	}
}

// TestEnrolledProvidersAtTotalProviderConfigAccounting is the live gate's
// twin, scoped: every provider that ships a provider-config manifest is at
// total provider-block accounting right now (the shared baseline carries no
// provider: entries).
func TestEnrolledProvidersAtTotalProviderConfigAccounting(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, catalogRoot)); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test")
	}
	schemas, err := LoadSchemas("schemas")
	if err != nil {
		t.Fatalf("schemas: %v", err)
	}
	enrollments, err := DiscoverEnrollments(root)
	if err != nil {
		t.Fatalf("enrollments: %v", err)
	}
	withManifest := 0
	for _, e := range enrollments {
		manifest, err := LoadProviderConfigManifest(root, e.Provider)
		if err != nil {
			t.Fatalf("%s: %v", e.Provider, err)
		}
		if manifest == nil {
			continue
		}
		withManifest++
		acc, err := BuildAccounting(root, e.Provider, schemas, e.GASchema, "", "")
		if err != nil {
			t.Fatalf("%s: %v", e.Provider, err)
		}
		if acc.ProviderConfig == nil {
			t.Errorf("%s: manifest present but no provider-config accounting ran", e.Provider)
			continue
		}
		if !acc.ProviderConfig.Accounted() {
			t.Errorf("%s: provider block not at total accounting: %+v", e.Provider, acc.ProviderConfig)
		}
	}
	if withManifest == 0 {
		t.Error("no provider ships a provider-config manifest -- enrollment collapsed")
	}
}

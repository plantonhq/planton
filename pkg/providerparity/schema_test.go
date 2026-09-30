//go:build !codegen
// +build !codegen

package providerparity

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// fixtureSchema exercises every distillation shape: required/optional/
// computed-only/sensitive/deprecated attributes, nested blocks (including a
// deprecated one whose args must inherit deprecation), and a deprecated
// resource.
func fixtureSchema() *Schema {
	return &Schema{
		Provider: "google",
		Source:   "hashicorp/google",
		Version:  "6.50.0",
		Resources: map[string]*Block{
			"google_storage_bucket": {
				Attributes: map[string]*Attribute{
					"name":      {Type: json.RawMessage(`"string"`), Required: true},
					"location":  {Type: json.RawMessage(`"string"`), Optional: true},
					"self_link": {Type: json.RawMessage(`"string"`), Computed: true},
					"secret":    {Type: json.RawMessage(`"string"`), Optional: true, Sensitive: true},
					"old_knob":  {Type: json.RawMessage(`"string"`), Optional: true, Deprecated: true},
				},
				Blocks: map[string]*NestedBlock{
					"lifecycle_rule": {
						NestingMode: "list",
						Block: &Block{
							Attributes: map[string]*Attribute{"age": {Optional: true}},
							Blocks: map[string]*NestedBlock{
								"condition": {
									NestingMode: "single",
									Block: &Block{
										Attributes: map[string]*Attribute{"prefix": {Optional: true}},
									},
								},
							},
						},
					},
					"legacy_block": {
						NestingMode: "list",
						Block: &Block{
							Deprecated: true,
							Attributes: map[string]*Attribute{"value": {Optional: true}},
						},
					},
				},
			},
			"google_legacy_thing": {
				Deprecated: true,
				Attributes: map[string]*Attribute{"name": {Required: true}},
			},
		},
	}
}

func TestConfigurableArgs(t *testing.T) {
	bucket := fixtureSchema().Resources["google_storage_bucket"]

	args := bucket.ConfigurableArgs("")
	byPath := map[string]Arg{}
	for _, a := range args {
		byPath[a.Path] = a
	}

	want := map[string]Arg{
		"name":                            {Path: "name", Required: true},
		"location":                        {Path: "location"},
		"secret":                          {Path: "secret", Sensitive: true},
		"old_knob":                        {Path: "old_knob", Deprecated: true},
		"lifecycle_rule.age":              {Path: "lifecycle_rule.age"},
		"lifecycle_rule.condition.prefix": {Path: "lifecycle_rule.condition.prefix"},
		// legacy_block is a deprecated block: its args inherit deprecation.
		"legacy_block.value": {Path: "legacy_block.value", Deprecated: true},
	}
	if len(byPath) != len(want) {
		t.Fatalf("args = %v, want exactly %d paths (computed-only self_link must be excluded)", byPath, len(want))
	}
	for path, w := range want {
		if byPath[path] != w {
			t.Errorf("arg %s = %+v, want %+v", path, byPath[path], w)
		}
	}

	// Non-deprecated configurable count: name, location, secret, the two
	// lifecycle args -- old_knob and legacy_block.value are deprecated.
	if got := bucket.ConfigurableArgCount(); got != 5 {
		t.Errorf("ConfigurableArgCount = %d, want 5", got)
	}
}

func TestSchemaFileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := fixtureSchema()

	path1, err := WriteSchemaFile(dir, s)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	loaded, err := LoadSchema(dir, "google")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.Version != s.Version || loaded.Source != s.Source || len(loaded.Resources) != len(s.Resources) {
		t.Fatalf("round trip lost identity: %+v", loaded)
	}
	if loaded.Resources["google_storage_bucket"].ConfigurableArgCount() != 5 {
		t.Error("round trip lost block structure")
	}

	// Determinism: rewriting the identical schema is byte-identical.
	before, _ := os.ReadFile(path1)
	if _, err := WriteSchemaFile(dir, s); err != nil {
		t.Fatalf("rewrite: %v", err)
	}
	after, _ := os.ReadFile(path1)
	if string(before) != string(after) {
		t.Error("regenerating an unchanged schema is not byte-identical")
	}

	// A pin bump replaces the artifact instead of accumulating versions.
	bumped := fixtureSchema()
	bumped.Version = "7.0.0"
	path2, err := WriteSchemaFile(dir, bumped)
	if err != nil {
		t.Fatalf("write bumped: %v", err)
	}
	if _, err := os.Stat(path1); !os.IsNotExist(err) {
		t.Errorf("stale artifact %s survived the pin bump", path1)
	}
	loaded, err = LoadSchema(dir, "google")
	if err != nil {
		t.Fatalf("load after bump: %v", err)
	}
	if loaded.Version != "7.0.0" {
		t.Errorf("loaded version = %s, want 7.0.0 (from %s)", loaded.Version, path2)
	}
}

func TestLoadSchemasRejectsAmbiguousPin(t *testing.T) {
	dir := t.TempDir()
	s := fixtureSchema()
	if _, err := WriteSchemaFile(dir, s); err != nil {
		t.Fatal(err)
	}
	// Simulate a stale artifact surviving a pin bump (e.g. hand-copied).
	bumped := fixtureSchema()
	bumped.Version = "7.0.0"
	tmp := t.TempDir()
	stale, err := WriteSchemaFile(tmp, bumped)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(stale)
	if err := os.WriteFile(filepath.Join(dir, filepath.Base(stale)), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSchemas(dir); err == nil {
		t.Error("two artifacts for one provider must fail loudly, got nil error")
	}
}

// TestCommittedArtifacts guards the real committed schemas: loadable, one per
// provider, carrying identity and a non-trivial resource surface. Skipped in
// sandboxes without the source tree (bazel), like the anatomy gate.
func TestCommittedArtifacts(t *testing.T) {
	if _, err := os.Stat("schemas"); err != nil {
		t.Skip("schemas dir not present (bazel sandbox); runs under go test")
	}
	schemas, err := LoadSchemas("schemas")
	if err != nil {
		t.Fatalf("committed artifacts unreadable: %v", err)
	}
	if len(schemas) == 0 {
		t.Fatal("no committed schema artifacts")
	}
	for name, s := range schemas {
		if s.Version == "" || s.Source == "" {
			t.Errorf("%s: missing identity: %+v", name, s)
		}
		// The floor catches a truncated or garbage artifact (a distiller
		// failure writes a near-empty file), so it sits well below the
		// smallest real provider surface: hyperscalers register hundreds to
		// thousands of resources, while the smaller IaaS providers
		// (DigitalOcean, Hetzner Cloud) legitimately register only tens.
		if len(s.Resources) < 10 {
			t.Errorf("%s: only %d resources -- even the smallest real provider registers tens; the artifact is truncated", name, len(s.Resources))
		}
		// The provider block is part of the parity surface: every provider
		// takes at least its credential there, so an artifact whose block is
		// absent or has no configurable argument was distilled by a build that
		// silently dropped it -- the regression this guard exists to catch.
		// The floor is one, not "a handful": a provider can be that small
		// (stripe/stripe takes exactly api_key and stripe_account), and the
		// provider-config accounting judges every argument of an enrolled
		// provider in both directions, so a count adds nothing past non-empty.
		if s.ProviderConfig == nil {
			t.Errorf("%s: artifact carries no provider block -- regenerate with the current distiller", name)
			continue
		}
		if len(s.ProviderConfig.ConfigurableArgs("")) == 0 {
			t.Errorf("%s: provider block carries no configurable argument -- the block was dropped; regenerate with the current distiller", name)
		}
	}
}

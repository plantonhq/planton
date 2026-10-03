package controlcatalog

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	controlcatalogv1 "github.com/plantonhq/planton/compliance/controlcatalog/v1"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// controlIDPattern is the stable, OSCAL-compatible id shape: lowercase
// dashed tokens ("enc-at-rest"). Ids are referenced by every kind
// profile and crosswalk, so the shape is enforced at the source.
var controlIDPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// TestControlCatalogConformance holds the central compliance vocabulary to
// its contract:
//
//  1. The catalog parses strictly against its proto schema; ids are unique,
//     well-shaped, and every control carries a name, a testable statement,
//     and a category.
//  2. Every crosswalk parses, names its framework and version, and every
//     control id it references exists in the catalog -- a dangling
//     reference would render a framework answer from nothing.
//
// Enrollment is file presence (the catalog path and frameworks/*.yaml);
// there is no allowlist to keep in sync.
func TestControlCatalogConformance(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "catalog")); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test and the lint.catalog-data lane")
	}

	catalog, err := Load(root)
	if err != nil {
		t.Fatalf("control catalog: %v", err)
	}
	if catalog.GetKind() != "ControlCatalog" {
		t.Fatalf("kind is %q, want ControlCatalog", catalog.GetKind())
	}
	if len(catalog.GetSpec().GetControls()) == 0 {
		t.Fatal("control catalog declares no controls")
	}

	ids := map[string]bool{}
	for _, c := range catalog.GetSpec().GetControls() {
		if !controlIDPattern.MatchString(c.GetId()) {
			t.Errorf("control id %q is not lowercase-dashed", c.GetId())
		}
		if ids[c.GetId()] {
			t.Errorf("duplicate control id %q", c.GetId())
		}
		ids[c.GetId()] = true
		if strings.TrimSpace(c.GetName()) == "" {
			t.Errorf("control %s has no name", c.GetId())
		}
		if strings.TrimSpace(c.GetStatement()) == "" {
			t.Errorf("control %s has no statement -- a control without a testable sentence asserts nothing", c.GetId())
		}
		if c.GetCategory() == controlcatalogv1.Category_category_unspecified {
			t.Errorf("control %s has no category", c.GetId())
		}
	}

	frameworks, err := DiscoverCrosswalks(root)
	if err != nil {
		t.Fatalf("discovering crosswalks: %v", err)
	}
	for _, framework := range frameworks {
		framework := framework
		t.Run(framework, func(t *testing.T) {
			crosswalk, err := LoadCrosswalk(root, framework)
			if err != nil {
				t.Fatalf("crosswalk: %v", err)
			}
			if crosswalk.GetKind() != "FrameworkCrosswalk" {
				t.Fatalf("kind is %q, want FrameworkCrosswalk", crosswalk.GetKind())
			}
			if crosswalk.GetMetadata().GetName() != framework {
				t.Errorf("metadata.name is %q, want %q (the filename is the framework's identity)",
					crosswalk.GetMetadata().GetName(), framework)
			}
			if strings.TrimSpace(crosswalk.GetSpec().GetFrameworkName()) == "" {
				t.Error("framework_name is empty")
			}
			if strings.TrimSpace(crosswalk.GetSpec().GetFrameworkVersion()) == "" {
				t.Error("framework_version is empty -- a crosswalk must state what revision it was authored against")
			}
			if len(crosswalk.GetSpec().GetMappings()) == 0 {
				t.Error("crosswalk declares no mappings")
			}
			// Provider scope: the schema's enum already refuses invented
			// providers at parse time; the gate holds what the schema
			// cannot express -- no meaningless values, no duplicates.
			// (An EMPTY list is the provider-neutral statement and is
			// always legal.)
			seenProviders := map[catalogkind.CatalogProvider]bool{}
			for _, provider := range crosswalk.GetSpec().GetProviders() {
				switch provider {
				case catalogkind.CatalogProvider_catalog_provider_unspecified:
					t.Error("providers lists the unspecified value -- provider-neutral is stated by an EMPTY list, never by unspecified")
				case catalogkind.CatalogProvider__test:
					t.Error("providers lists the synthetic _test provider, which must never reach a user-facing surface")
				}
				if seenProviders[provider] {
					t.Errorf("providers lists %s more than once", provider)
				}
				seenProviders[provider] = true
			}
			seen := map[string]bool{}
			for _, m := range crosswalk.GetSpec().GetMappings() {
				if strings.TrimSpace(m.GetRequirementId()) == "" {
					t.Error("mapping with empty requirement_id")
					continue
				}
				if seen[m.GetRequirementId()] {
					t.Errorf("duplicate requirement_id %q", m.GetRequirementId())
				}
				seen[m.GetRequirementId()] = true
				if strings.TrimSpace(m.GetRequirementName()) == "" {
					t.Errorf("requirement %s has no name", m.GetRequirementId())
				}
				if len(m.GetControlIds()) == 0 {
					t.Errorf("requirement %s maps to no controls -- an empty mapping asserts nothing", m.GetRequirementId())
				}
				for _, id := range m.GetControlIds() {
					if !ids[id] {
						t.Errorf("requirement %s references control %q, which does not exist in the control catalog",
							m.GetRequirementId(), id)
					}
				}
			}
		})
	}
}

// repoRoot walks up from this test file to the directory containing go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(thisFile)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found walking up from test file")
		}
		dir = parent
	}
}

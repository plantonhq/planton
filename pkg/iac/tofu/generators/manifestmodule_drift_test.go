package generators

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/kubernetes/manifestprojection"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// manifestModuleGaps names the projection kinds whose committed module the
// generator cannot yet express, each with what it does that the generator
// does not. Such a module stays hand-owned until the generator learns the
// shape or the kind stops being a pure projection; it is never hand-patched
// to match. The list only shrinks: a listed kind whose module matches the
// generator fails the guard until its entry is removed.
var manifestModuleGaps = map[string]string{
	"KubernetesHttpRoute":         "exports first_host, a value derived from spec.hostnames that no generator rule maps",
	"KubernetesListenerSet":       "its gateway-name output is the resolved spec.parentRef.name, which the generator's *_name rule would map to metadata.name",
	"KubernetesKarpenterNodePool": "its locals default nodeClassRef.group and kind and reshape template metadata, so its spec is not passed through verbatim",
}

// TestManifestModuleDrift asserts that every Kubernetes manifest projection
// kind's committed iac/tf module is byte-identical to what GenerateManifestModule
// produces. A projection kind's Terraform module is generator output by rule
// (the forge's fast path), so a hand edit, or a generator change nobody
// regenerated, would otherwise ship unnoticed to every adopter of the kind.
//
// Enrollment is read from the kind registry: every kind whose metadata carries
// kubernetes_manifest_projection is guarded the day it is registered, minus the
// named manifestModuleGaps. Files the generator does not produce (README.md,
// the lock file) are not compared, and a generated file missing from the
// module is a failure. Run with PLANTON_REGEN_MANIFEST_MODULES=1 to (re)write
// the enrolled modules from the generator instead of comparing.
func TestManifestModuleDrift(t *testing.T) {
	root := repoRoot(t)
	regenerate := os.Getenv("PLANTON_REGEN_MANIFEST_MODULES") == "1"

	enrolled := 0
	for _, kind := range catalogkindreflect.KindsList() {
		if manifestprojection.ProjectionOf(kind) == nil {
			if _, listed := manifestModuleGaps[kind.String()]; listed {
				t.Errorf("manifestModuleGaps names %s, which is not a projection kind; remove the stale entry", kind)
			}
			continue
		}
		kind := kind
		if _, gap := manifestModuleGaps[kind.String()]; gap {
			t.Run(kind.String()+"/gap", func(t *testing.T) {
				if moduleMatchesGenerator(t, root, kind) {
					t.Errorf("%s now matches the generator; remove it from manifestModuleGaps", kind)
				}
			})
			continue
		}
		enrolled++
		t.Run(kind.String(), func(t *testing.T) {
			msg, err := catalogkindreflect.NewInstance(kind)
			if err != nil {
				t.Fatalf("NewInstance(%s): %v", kind, err)
			}
			files, err := GenerateManifestModule(kind, msg)
			if err != nil {
				t.Fatalf("GenerateManifestModule(%s): %v", kind, err)
			}
			moduleDir := filepath.Dir(moduleVariablesPath(root, msg))

			names := make([]string, 0, len(files))
			for name := range files {
				// The runtime writes backend.tf before every init
				// (tfbackend.WriteBackendFile), so a committed copy is never
				// read and most modules commit none.
				if name == "backend.tf" {
					continue
				}
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				want := files[name]
				path := filepath.Join(moduleDir, name)
				if regenerate {
					if err := os.WriteFile(path, []byte(want), 0o644); err != nil {
						t.Fatalf("write %s: %v", path, err)
					}
					continue
				}
				got, err := os.ReadFile(path)
				if err != nil {
					t.Errorf("read %s: %v", path, err)
					continue
				}
				if strings.TrimRight(string(got), "\n") != strings.TrimRight(want, "\n") {
					t.Errorf("%s is out of sync with the generator.\n"+
						"Run: PLANTON_REGEN_MANIFEST_MODULES=1 go test ./pkg/iac/tofu/generators/ -run TestManifestModuleDrift\n"+
						"path: %s", name, path)
				}
			}
		})
	}
	if enrolled == 0 {
		t.Fatal("no Kubernetes manifest projection kind is registered; the registry walk is broken")
	}
}

// moduleMatchesGenerator reports whether a kind's committed module equals the
// generator's output (backend.tf aside). A kind the generator refuses never
// matches.
func moduleMatchesGenerator(t *testing.T, root string, kind catalogkind.CatalogKind) bool {
	t.Helper()
	msg, err := catalogkindreflect.NewInstance(kind)
	if err != nil {
		t.Fatalf("NewInstance(%s): %v", kind, err)
	}
	files, err := GenerateManifestModule(kind, msg)
	if err != nil {
		return false
	}
	moduleDir := filepath.Dir(moduleVariablesPath(root, msg))
	for name, want := range files {
		if name == "backend.tf" {
			continue
		}
		got, err := os.ReadFile(filepath.Join(moduleDir, name))
		if err != nil || strings.TrimRight(string(got), "\n") != strings.TrimRight(want, "\n") {
			return false
		}
	}
	return true
}

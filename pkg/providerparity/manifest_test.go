//go:build !codegen
// +build !codegen

package providerparity

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plantonhq/planton/shared/catalogkind"
)

func writeManifest(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ManifestFileName)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadManifest_Valid(t *testing.T) {
	path := writeManifest(t, `
resources:
  google_widget:
    mappings:
      - spec: spec.widget_name
        arg: name
      - spec: spec.rule_list
        arg: rules
    exclusions:
      - arg: rules.condition.send_age_if_zero
        reason: proto3 optional presence covers the zero-vs-unset distinction
  google_widget_iam_member:
    specRoot: spec.iam_members
    exclusions:
      - arg: widget
        reason: wired by the module to the enclosing widget
  google_project_service:
    internal: API enablement plumbing; arguments are module decisions
  azapi_resource:
    external: raw-ARM surface pinned at type@api-version by recorded admission
specExclusions:
  - field: spec.platform_only
    reason: platform-level concept with no provider counterpart
`)
	m, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(m.Resources) != 4 || len(m.SpecExclusions) != 1 {
		t.Fatalf("parsed shape = %+v", m)
	}
	if m.Resources["google_project_service"].Internal == "" {
		t.Error("internal disposition lost in parsing")
	}
	if m.Resources["azapi_resource"].External == "" {
		t.Error("external disposition lost in parsing")
	}
	if m.Resources["google_widget_iam_member"].SpecRoot != "spec.iam_members" {
		t.Error("specRoot lost in parsing")
	}
}

// TestLoadManifest_Rejections proves every validation class fails loudly: a
// manifest records judgment, and a half-read record is worse than none.
func TestLoadManifest_Rejections(t *testing.T) {
	cases := map[string]struct {
		yaml    string
		wantErr string
	}{
		"unknown field (strict parsing)": {
			yaml:    "resources:\n  google_widget:\n    mapings: []\n",
			wantErr: "not valid (strict) YAML",
		},
		"internal plus other judgment": {
			yaml:    "resources:\n  google_widget:\n    internal: plumbing\n    specRoot: spec.x\n",
			wantErr: "internal is the whole judgment",
		},
		"external plus other judgment": {
			yaml:    "resources:\n  azapi_resource:\n    external: raw-ARM admission\n    specRoot: spec.x\n",
			wantErr: "external is the whole judgment",
		},
		"internal plus external": {
			yaml:    "resources:\n  azapi_resource:\n    internal: plumbing\n    external: raw-ARM admission\n",
			wantErr: "mutually exclusive",
		},
		"exclusion without reason": {
			yaml:    "resources:\n  google_widget:\n    exclusions:\n      - arg: name\n",
			wantErr: "no reason",
		},
		"mapping not spec-rooted": {
			yaml:    "resources:\n  google_widget:\n    mappings:\n      - spec: widget_name\n        arg: name\n",
			wantErr: "must be spec-rooted",
		},
		"arg judged twice": {
			yaml:    "resources:\n  google_widget:\n    mappings:\n      - spec: spec.a\n        arg: name\n    exclusions:\n      - arg: name\n        reason: r\n",
			wantErr: "judged twice",
		},
		"identical mapping recorded twice": {
			yaml:    "resources:\n  google_widget:\n    mappings:\n      - spec: spec.a\n        arg: name\n      - spec: spec.a\n        arg: name\n",
			wantErr: "recorded twice",
		},
		"empty resource entry": {
			yaml:    "resources:\n  google_widget: {}\n",
			wantErr: "records nothing",
		},
		"spec exclusion without reason": {
			yaml:    "specExclusions:\n  - field: spec.x\n",
			wantErr: "no reason",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := LoadManifest(writeManifest(t, tc.yaml))
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("err = %v, want it to contain %q", err, tc.wantErr)
			}
		})
	}
}

// TestLoadManifest_FanInMappings proves the dual of the name/value idiom
// parses: one map-typed argument may be recorded by several mappings, each
// covering one honest spec field.
func TestLoadManifest_FanInMappings(t *testing.T) {
	m, err := LoadManifest(writeManifest(t,
		"resources:\n  google_widget:\n    mappings:\n"+
			"      - spec: spec.resources.cpu\n        arg: resources.limits\n"+
			"      - spec: spec.resources.memory\n        arg: resources.limits\n"))
	if err != nil {
		t.Fatalf("fan-in mappings must parse, got err %v", err)
	}
	if got := len(m.Resources["google_widget"].Mappings); got != 2 {
		t.Fatalf("mappings = %d, want 2", got)
	}
}

func TestLoadKindManifest_AbsenceIsNotAnError(t *testing.T) {
	// Enrollment is file presence; a kind without a manifest is an accepted
	// gap in the baseline, never a load failure.
	m, err := LoadKindManifest(t.TempDir(),
		catalogkind.CatalogProvider_gcp,
		catalogkind.CatalogKind_GcpGcsBucket)
	if err != nil {
		t.Fatalf("absence must be (nil, nil), got err %v", err)
	}
	if m != nil {
		t.Fatalf("absence must be (nil, nil), got %+v", m)
	}
}

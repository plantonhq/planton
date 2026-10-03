package setdeploy

import (
	"testing"

	"github.com/plantonhq/planton/internal/manifest"
)

// The detector is prefix detection over every populated string container —
// singular fields, repeated messages, map values, and StringValueOrRef
// literal arms — never a parse of the reference grammar (the grammar's one
// home is server-side).
func TestCollectBackendRefs(t *testing.T) {
	manifestYaml := `apiVersion: _test.planton.dev/v1alpha2
kind: TestCatalogKindGeneric
metadata:
  name: refs
  env: dev
spec:
  requiredRef:
    value: $secret/db-password
  sensitiveString: $var/api-key
  labels:
    ok: plain-value
    hot: $secret/from-a-map
  steps:
    - command: $var/from-a-list
    - command: echo plain
  displayName: "not-a-ref $secret/mid-string stays"
`
	msg, err := manifest.LoadManifestBytes([]byte(manifestYaml), "refs.yaml")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	uses := CollectBackendRefs(msg)

	found := map[string]string{}
	for _, use := range uses {
		found[use.FieldPath] = use.Prefix
	}
	expect := map[string]string{
		"spec.required_ref.value": secretRefPrefix,
		"spec.sensitive_string":   varRefPrefix,
		"spec.labels.hot":         secretRefPrefix,
		"spec.steps[0].command":   varRefPrefix,
	}
	for path, prefix := range expect {
		if found[path] != prefix {
			t.Fatalf("expected %s at %s; found: %v", prefix, path, found)
		}
	}
	if len(found) != len(expect) {
		t.Fatalf("prefix detection must not fire mid-string or on plain values; found: %v", found)
	}
}

package verify

import (
	"encoding/json"
	"reflect"
	"testing"
)

// The read-back comparison is what lets a projection lane say every field it
// set landed. These cases pin its semantics: server-written defaults (extra
// keys, including inside list items) are allowed; a pruned key, a changed
// value, a number written as text, or a list of a different length is named
// by its path.

func decode(t *testing.T, s string) interface{} {
	t.Helper()
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatal(err)
	}
	return v
}

func TestMissingFromLive_AllowsServerDefaults(t *testing.T) {
	want := decode(t, `{"podTemplate":{"spec":{"containers":[{"name":"app","ports":[{"containerPort":8888}]}]}}}`)
	live := decode(t, `{"operatingMode":"Running","podTemplate":{"spec":{"containers":[{"name":"app","ports":[{"containerPort":8888,"protocol":"TCP"}]}]}}}`)
	if got := missingFromLive(want, live, "spec"); len(got) != 0 {
		t.Errorf("findings = %q, want none", got)
	}
}

func TestMissingFromLive_NamesEveryDivergence(t *testing.T) {
	want := decode(t, `{"replicas":2,"sandboxTemplateRef":{"name":"t"},"podTemplate":{"spec":{"hostIPC":true,"containers":[{"name":"a"},{"name":"b"}]}},"runAsUser":1000}`)
	live := decode(t, `{"replicas":3,"sandboxTemplateRef":{"name":"t"},"podTemplate":{"spec":{"containers":[{"name":"a"}]}},"runAsUser":"1000"}`)
	got := missingFromLive(want, live, "spec")
	wantFindings := []string{
		`spec.podTemplate.spec.containers: want 2 element(s), live has [{"name":"a"}]`,
		`spec.podTemplate.spec.hostIPC: set by the manifest, absent from the live object (pruned)`,
		`spec.replicas: want 2, live has 3`,
		`spec.runAsUser: want 1000, live has "1000"`,
	}
	if !reflect.DeepEqual(got, wantFindings) {
		t.Errorf("findings =\n%q\nwant\n%q", got, wantFindings)
	}
}

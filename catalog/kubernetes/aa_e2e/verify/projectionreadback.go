package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/kubernetes/manifestprojection"
)

// proveProjectionReadback proves every field a projection kind's scenario sets
// reads back from the live object: the spec the manifest renders (the same
// projection both engines apply) must be contained in the live spec, value for
// value, and the declared labels and annotations must be on the object. An API
// server silently prunes a key its definition does not serve (a miscased key,
// a field the installed definition predates), and only reading the object back
// catches that.
//
// Containment rather than equality: the API server writes the defaults its
// definition declares into the stored object (a large custom resource such as
// one embedding a pod carries many, inside list items too), and an exact
// comparison would need a hand-kept copy of upstream's defaults. Every key and
// value the manifest sets must be present as set; keys only the server added
// are allowed.
func proveProjectionReadback(ctx context.Context, kubeconfig, resource, name, namespace, manifestPath string) error {
	want, err := expectedProjection(manifestPath)
	if err != nil {
		return err
	}
	liveSpec, err := kubectlGetJSONPath(ctx, kubeconfig, resource, name, namespace, "{.spec}")
	if err != nil {
		return errors.Wrap(err, "reading the live spec")
	}
	var live interface{}
	if err := json.Unmarshal([]byte(liveSpec), &live); err != nil {
		return errors.Wrapf(err, "decoding the live spec of %s %s", resource, name)
	}
	encoded, err := json.Marshal(want.Spec)
	if err != nil {
		return err
	}
	var expected interface{}
	if err := json.Unmarshal(encoded, &expected); err != nil {
		return err
	}
	if missing := missingFromLive(expected, live, "spec"); len(missing) > 0 {
		return errors.Errorf("READ-BACK: %d field(s) the manifest sets did not land on the live %s %s:\n  %s",
			len(missing), resource, name, joinLines(missing))
	}
	if err := proveObjectMetadata(ctx, kubeconfig, resource, name, namespace, want); err != nil {
		return err
	}
	fmt.Printf("  [verify] READ-BACK: every field the manifest sets is on the live object, labels and annotations included\n")
	return nil
}

// missingFromLive lists, by path, every value in want that is absent from or
// different in got. Objects compare key by key (extra keys in got are the
// server's defaults); lists compare element by element and must be the same
// length, since an atomic list is stored as written; scalars compare as JSON
// values.
func missingFromLive(want, got interface{}, path string) []string {
	switch w := want.(type) {
	case map[string]interface{}:
		g, ok := got.(map[string]interface{})
		if !ok {
			return []string{fmt.Sprintf("%s: want an object, live has %s", path, describeJSON(got))}
		}
		keys := make([]string, 0, len(w))
		for k := range w {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var out []string
		for _, k := range keys {
			gv, present := g[k]
			if !present {
				out = append(out, fmt.Sprintf("%s.%s: set by the manifest, absent from the live object (pruned)", path, k))
				continue
			}
			out = append(out, missingFromLive(w[k], gv, path+"."+k)...)
		}
		return out
	case []interface{}:
		g, ok := got.([]interface{})
		if !ok || len(g) != len(w) {
			return []string{fmt.Sprintf("%s: want %d element(s), live has %s", path, len(w), describeJSON(got))}
		}
		var out []string
		for i := range w {
			out = append(out, missingFromLive(w[i], g[i], path+"["+strconv.Itoa(i)+"]")...)
		}
		return out
	default:
		wj, _ := json.Marshal(want)
		gj, _ := json.Marshal(got)
		if string(wj) != string(gj) {
			return []string{fmt.Sprintf("%s: want %s, live has %s", path, wj, gj)}
		}
		return nil
	}
}

// proveObjectMetadata checks the declared labels and annotations are on the
// live object, under Planton's identity labels.
func proveObjectMetadata(ctx context.Context, kubeconfig, resource, name, namespace string, want *manifestprojection.Object) error {
	liveLabels, err := liveStringMap(ctx, kubeconfig, resource, name, namespace, "{.metadata.labels}")
	if err != nil {
		return err
	}
	for k, val := range want.Labels {
		// The id, organization and environment labels carry values the
		// harness assigns at deploy time; the rest are fixed by the manifest.
		if k == manifestprojection.LabelResourceID || k == manifestprojection.LabelOrganization || k == manifestprojection.LabelEnvironment {
			continue
		}
		if liveLabels[k] != val {
			return errors.Errorf("PROJECTION: label %s = %q on the live object, want %q", k, liveLabels[k], val)
		}
	}
	if len(want.Annotations) > 0 {
		liveAnnotations, err := liveStringMap(ctx, kubeconfig, resource, name, namespace, "{.metadata.annotations}")
		if err != nil {
			return err
		}
		for k, val := range want.Annotations {
			if liveAnnotations[k] != val {
				return errors.Errorf("PROJECTION: annotation %s = %q on the live object, want %q", k, liveAnnotations[k], val)
			}
		}
	}
	return nil
}

func describeJSON(v interface{}) string {
	if v == nil {
		return "nothing"
	}
	encoded, _ := json.Marshal(v)
	return string(encoded)
}

func joinLines(lines []string) string {
	out := ""
	for i, l := range lines {
		if i > 0 {
			out += "\n  "
		}
		out += l
	}
	return out
}

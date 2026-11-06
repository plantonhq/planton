package module

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

// Upstream's installer turns some settings into kustomize components and
// overlays: the agentgateway dataplane, the cordoned control plane, the
// size10 pod-certificate rate. The module applies those same patches, read
// from the published tree, rather than retyping what upstream decided: a new
// release that changes a component changes what is applied, and a patch that
// no longer finds its target refuses the deploy instead of being skipped.
//
// Only the two patch forms upstream's components use are understood: JSON
// patch operations (add, replace, remove, test) and strategic-merge patches
// over the lists Kubernetes merges by name (by mountPath for volume mounts),
// with `$patch: delete` and null removals. The Terraform twin applies the
// same edits, one by one, in locals.tf.

// componentPatch is one entry of a kustomization's patches list: the patch
// inline, or the file beside the kustomization that holds it (kustomize's
// `path:`, which resolvePaths reads in before the patch is applied).
type componentPatch struct {
	Target struct {
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"target"`
	Patch string `json:"patch"`
	Path  string `json:"path"`
}

// kustomization is the part of a kustomization or component the module reads.
type kustomization struct {
	Patches []componentPatch `json:"patches"`
}

// parseKustomization reads a published kustomization file.
func parseKustomization(text string) (*kustomization, error) {
	k := &kustomization{}
	if err := yaml.Unmarshal([]byte(text), k); err != nil {
		return nil, errors.Wrap(err, "parsing a release kustomization")
	}
	return k, nil
}

// resolvePaths reads every `path:` patch from the release file beside the
// kustomization at kustomizationPath, so applyComponent sees each patch's
// text. A path the module does not fetch is a pin that needs a catalog fix.
func (k *kustomization) resolvePaths(kustomizationPath string, f releaseFileSet) error {
	dir := kustomizationPath[:strings.LastIndex(kustomizationPath, "/")+1]
	for i := range k.Patches {
		p := &k.Patches[i]
		if p.Path == "" {
			continue
		}
		text, ok := f.files[dir+p.Path]
		if !ok {
			return errors.Errorf("Agent Substrate %s's %s patches from %s, which the module does not fetch; the module's pin needs a catalog fix", vars.Release, kustomizationPath, dir+p.Path)
		}
		p.Patch = text
	}
	return nil
}

// matches reports whether a patch targets a document: the kind must match,
// the name as an anchored regular expression (upstream's cordon component
// names several workloads in one alternation), and the namespace when given.
func (p componentPatch) matches(d document) bool {
	if p.Target.Kind != "" && p.Target.Kind != str(d, "kind") {
		return false
	}
	if p.Target.Namespace != "" && p.Target.Namespace != str(d, "metadata", "namespace") {
		return false
	}
	if p.Target.Name != "" {
		matched, err := regexp.MatchString("^(?:"+p.Target.Name+")$", str(d, "metadata", "name"))
		if err != nil || !matched {
			return false
		}
	}
	return true
}

// applyComponent applies every patch of a component to the documents it
// targets, in order. A patch that targets nothing in the set is fine
// (upstream's components span several install streams); a patch that
// deletes its target removes the document. Every named target the caller
// lists in required must be found.
func applyComponent(k *kustomization, documents []document, required ...string) ([]document, error) {
	found := map[string]bool{}
	for _, p := range k.Patches {
		var kept []document
		for _, d := range documents {
			if !p.matches(d) {
				kept = append(kept, d)
				continue
			}
			found[str(d, "kind")+"/"+str(d, "metadata", "name")] = true
			deleted, err := applyPatchText(d, p.Patch)
			if err != nil {
				return nil, errors.Wrapf(err, "applying upstream's patch to %s", documentID(d))
			}
			if !deleted {
				kept = append(kept, d)
			}
		}
		documents = kept
	}
	for _, target := range required {
		if !found[target] {
			return nil, errors.Errorf("Agent Substrate %s's component patches no longer find %s; the module's pin needs a catalog fix", vars.Release, target)
		}
	}
	return documents, nil
}

// applyPatchText applies one patch: a YAML list is JSON patch operations, a
// map is a strategic-merge patch. It reports whether the patch deleted the
// whole document.
func applyPatchText(d document, text string) (bool, error) {
	var parsed interface{}
	if err := yaml.Unmarshal([]byte(text), &parsed); err != nil {
		return false, errors.Wrap(err, "parsing the patch")
	}
	switch patch := parsed.(type) {
	case []interface{}:
		return false, applyJSONPatch(d, patch)
	case map[string]interface{}:
		if patch["$patch"] == "delete" {
			return true, nil
		}
		// The patch's own identity is kustomize's to match on, never to merge:
		// upstream's cordon component names its target
		// "overridden-by-target" and selects by the target instead.
		delete(patch, "apiVersion")
		delete(patch, "kind")
		if metadata, ok := patch["metadata"].(map[string]interface{}); ok {
			delete(metadata, "name")
			delete(metadata, "namespace")
		}
		mergePatch(d, patch)
		return false, nil
	}
	return false, errors.New("a patch is neither a list of operations nor an object")
}

// applyJSONPatch applies JSON patch operations (RFC 6902) with JSON pointer
// paths. "-" as the last token appends to a list.
func applyJSONPatch(d document, ops []interface{}) error {
	for _, raw := range ops {
		op, _ := raw.(map[string]interface{})
		path, _ := op["path"].(string)
		tokens := strings.Split(strings.TrimPrefix(path, "/"), "/")
		parent, err := walk(d, tokens[:len(tokens)-1])
		if err != nil {
			return errors.Wrapf(err, "operation %v %s", op["op"], path)
		}
		last := tokens[len(tokens)-1]
		switch op["op"] {
		case "test":
			if got, _ := get(parent, last); !reflect.DeepEqual(got, op["value"]) {
				return errors.Errorf("upstream's patch expects %s to be %v, found %v", path, op["value"], got)
			}
		case "add", "replace", "remove":
			if err := set(d, tokens[:len(tokens)-1], parent, last, op["op"].(string), op["value"]); err != nil {
				return errors.Wrapf(err, "operation %v %s", op["op"], path)
			}
		default:
			return errors.Errorf("unsupported patch operation %v", op["op"])
		}
	}
	return nil
}

// walk follows tokens to the container (map or list) they name.
func walk(d document, tokens []string) (interface{}, error) {
	var current interface{} = d
	for _, token := range tokens {
		next, ok := get(current, token)
		if !ok {
			return nil, errors.Errorf("no %q along the path", token)
		}
		current = next
	}
	return current, nil
}

func get(container interface{}, token string) (interface{}, bool) {
	switch c := container.(type) {
	case map[string]interface{}:
		v, ok := c[token]
		return v, ok
	case []interface{}:
		i, err := strconv.Atoi(token)
		if err != nil || i < 0 || i >= len(c) {
			return nil, false
		}
		return c[i], true
	}
	return nil, false
}

// set applies add, replace or remove at last within parent. Lists are
// rebuilt, so the list's owner (reached again from the root through
// parentTokens) gets the new slice.
func set(root document, parentTokens []string, parent interface{}, last, op string, value interface{}) error {
	switch c := parent.(type) {
	case map[string]interface{}:
		if op == "remove" {
			delete(c, last)
		} else {
			c[last] = value
		}
		return nil
	case []interface{}:
		var updated []interface{}
		switch {
		case op == "add" && last == "-":
			updated = append(append([]interface{}{}, c...), value)
		default:
			i, err := strconv.Atoi(last)
			if err != nil || i < 0 || i > len(c) || (op != "add" && i == len(c)) {
				return errors.Errorf("index %q is out of range", last)
			}
			switch op {
			case "add":
				updated = append(append(append([]interface{}{}, c[:i]...), value), c[i:]...)
			case "replace":
				updated = append([]interface{}{}, c...)
				updated[i] = value
			case "remove":
				updated = append(append([]interface{}{}, c[:i]...), c[i+1:]...)
			}
		}
		grand, err := walk(root, parentTokens[:len(parentTokens)-1])
		if err != nil {
			return err
		}
		return set(root, parentTokens[:len(parentTokens)-1], grand, parentTokens[len(parentTokens)-1], "replace", updated)
	}
	return errors.New("the path does not end in an object or a list")
}

// mergePatch applies a strategic-merge patch: maps merge, a null removes a
// key, lists of named objects merge item by item (a `$patch: delete` item
// removes its match, and items already sharing a key become one), and any
// other list is replaced.
func mergePatch(target, patch map[string]interface{}) {
	for key, value := range patch {
		if key == "$patch" {
			continue
		}
		if value == nil {
			delete(target, key)
			continue
		}
		switch v := value.(type) {
		case map[string]interface{}:
			existing, ok := target[key].(map[string]interface{})
			if !ok {
				existing = map[string]interface{}{}
				target[key] = existing
			}
			mergePatch(existing, v)
		case []interface{}:
			target[key] = mergeList(key, target[key], v)
		default:
			target[key] = v
		}
	}
}

// mergeKey is the field Kubernetes merges a list's items by.
func mergeKey(listField string) string {
	if listField == "volumeMounts" {
		return "mountPath"
	}
	return "name"
}

func mergeList(field string, existing interface{}, patch []interface{}) []interface{} {
	key := mergeKey(field)
	base, _ := existing.([]interface{})
	keyed := len(patch) > 0
	for _, item := range patch {
		if m, ok := item.(map[string]interface{}); !ok || m[key] == nil {
			keyed = false
		}
	}
	if !keyed {
		return patch
	}
	// Kustomize merges a keyed list as a map, so items an earlier JSON patch
	// appended under a key already present collapse into the first one.
	var out []interface{}
	for _, b := range base {
		bm, ok := b.(map[string]interface{})
		duplicate := -1
		for i, o := range out {
			if om, isMap := o.(map[string]interface{}); ok && isMap && bm[key] != nil && om[key] == bm[key] {
				duplicate = i
			}
		}
		if duplicate >= 0 {
			mergePatch(out[duplicate].(map[string]interface{}), bm)
			continue
		}
		out = append(out, b)
	}
	for _, item := range patch {
		m := item.(map[string]interface{})
		index := -1
		for i, b := range out {
			if bm, ok := b.(map[string]interface{}); ok && bm[key] == m[key] {
				index = i
			}
		}
		switch {
		case m["$patch"] == "delete":
			if index >= 0 {
				out = append(out[:index], out[index+1:]...)
			}
		case index >= 0:
			mergePatch(out[index].(map[string]interface{}), m)
		default:
			clean := map[string]interface{}{}
			mergePatch(clean, m)
			out = append(out, clean)
		}
	}
	return out
}

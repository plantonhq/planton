// Package refannotations is the one Go reader of a reference field's
// foreign-key annotations: default_kind, default_kind_field_path, the
// candidate list and containment_exempt (shared/foreignkey/v1). Every Go lane
// that decides what a valueFrom may point at -- the manifest graph's rules and
// edges, the registry-wide reference gate, the catalog schema projection,
// explain, the YAML suggestions, the mapping evaluator, the e2e resolver --
// reads through it, so the composition-key rules below have one Go
// implementation. The platform's Java reader (ForeignKeyFieldAnnotations)
// applies the same rules; the case table in pkg/manifestgraph's rules test is
// the shared contract both copy.
//
// Composition keys. A reference field composes from a set of keys, each a
// kind plus the output path the field accepts from it: the candidate entries,
// plus (default_kind, default_kind_field_path) when both are present. A kind
// may carry more than one key (a subnetwork's primary range and its secondary
// ranges). The rules every reader applies:
//
//   - a valueFrom that names no kind takes default_kind;
//   - a valueFrom on a kind with exactly one key defaults its path to it;
//   - an explicit path on a keyed kind must equal one of that kind's keys or
//     extend it past a '.' (a list index or map key the annotation cannot
//     name);
//   - a kind with no key is accepted with an explicit path, unjudged.
package refannotations

import (
	"strings"

	"github.com/plantonhq/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Key is one composition key: a kind the reference can point at and the
// output path on that kind the field composes from.
type Key struct {
	Kind      cloudresourcekind.CloudResourceKind
	FieldPath string
}

// Field is the foreign-key annotations authored on one field. The zero value
// is a field with none.
type Field struct {
	DefaultKind          cloudresourcekind.CloudResourceKind
	DefaultKindFieldPath string
	Candidates           []Key
	ContainmentExempt    bool
}

// Of reads the annotations off a field descriptor (nil-safe).
func Of(fd protoreflect.FieldDescriptor) Field {
	if fd == nil || fd.Options() == nil {
		return Field{}
	}
	opts := fd.Options()
	var f Field
	f.DefaultKind, _ = proto.GetExtension(opts, foreignkeyv1.E_DefaultKind).(cloudresourcekind.CloudResourceKind)
	f.DefaultKindFieldPath, _ = proto.GetExtension(opts, foreignkeyv1.E_DefaultKindFieldPath).(string)
	f.ContainmentExempt, _ = proto.GetExtension(opts, foreignkeyv1.E_ContainmentExempt).(bool)
	candidates, _ := proto.GetExtension(opts, foreignkeyv1.E_Candidate).([]*foreignkeyv1.ReferenceCandidate)
	for _, c := range candidates {
		f.Candidates = append(f.Candidates, Key{Kind: c.GetKind(), FieldPath: c.GetFieldPath()})
	}
	return f
}

// IsReference reports whether the field carries any annotation that says
// what it may point at (a default kind or a candidate).
func (f Field) IsReference() bool {
	return f.DefaultKind != cloudresourcekind.CloudResourceKind_unspecified || len(f.Candidates) > 0
}

// Keys returns the field's composition keys: the default key first (when the
// field declares both a default kind and its path), then each candidate in
// authored order, without duplicates.
func (f Field) Keys() []Key {
	var keys []Key
	seen := map[Key]bool{}
	add := func(k Key) {
		if k.Kind == cloudresourcekind.CloudResourceKind_unspecified || k.FieldPath == "" || seen[k] {
			return
		}
		seen[k] = true
		keys = append(keys, k)
	}
	add(Key{Kind: f.DefaultKind, FieldPath: f.DefaultKindFieldPath})
	for _, c := range f.Candidates {
		add(c)
	}
	return keys
}

// Kinds returns the kinds the field names, the default kind first, each once.
// A console's reference picker offers exactly these.
func (f Field) Kinds() []cloudresourcekind.CloudResourceKind {
	var kinds []cloudresourcekind.CloudResourceKind
	seen := map[cloudresourcekind.CloudResourceKind]bool{}
	add := func(k cloudresourcekind.CloudResourceKind) {
		if k == cloudresourcekind.CloudResourceKind_unspecified || seen[k] {
			return
		}
		seen[k] = true
		kinds = append(kinds, k)
	}
	add(f.DefaultKind)
	for _, c := range f.Candidates {
		add(c.Kind)
	}
	return kinds
}

// KeysFor returns the output paths the field composes from on one kind, in
// Keys order; empty when the field names no key for it.
func (f Field) KeysFor(kind cloudresourcekind.CloudResourceKind) []string {
	var paths []string
	for _, k := range f.Keys() {
		if k.Kind == kind {
			paths = append(paths, k.FieldPath)
		}
	}
	return paths
}

// EffectiveKind is the kind a reference points at: its explicit kind, else
// the field's default kind, else unspecified.
func (f Field) EffectiveKind(explicit cloudresourcekind.CloudResourceKind) cloudresourcekind.CloudResourceKind {
	if explicit != cloudresourcekind.CloudResourceKind_unspecified {
		return explicit
	}
	return f.DefaultKind
}

// DefaultPath returns the path a reference to kind takes when it names none:
// the kind's key when it has exactly one. ok is false when the kind has no key
// or more than one (the reference must then name its path).
func (f Field) DefaultPath(kind cloudresourcekind.CloudResourceKind) (path string, ok bool) {
	keys := f.KeysFor(kind)
	if len(keys) != 1 {
		return "", false
	}
	return keys[0], true
}

// AcceptsPath reports whether an explicit path on kind is one the field
// composes from: it equals one of the kind's keys or extends one past a '.'.
// A kind with no key accepts any path (it is judged only by whether the path
// resolves on the kind).
func (f Field) AcceptsPath(kind cloudresourcekind.CloudResourceKind, path string) bool {
	keys := f.KeysFor(kind)
	if len(keys) == 0 {
		return true
	}
	for _, key := range keys {
		if path == key || strings.HasPrefix(path, key+".") {
			return true
		}
	}
	return false
}

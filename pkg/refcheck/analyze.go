// Package refcheck validates foreign-key reference integrity across the catalog kind
// registry: every composition key a field declares -- its (default_kind,
// default_kind_field_path) pair and each (foreignkey.v1.candidate) entry -- must point
// at a real field on the referenced kind's resolved target -- its status.outputs message
// for "status.outputs.*" paths, or its spec for "spec.*" paths. A dangling path is a
// composition that silently fails to resolve at deploy time (the orchestrator reads the
// referenced output and finds nothing), so this is a hard invariant, not a coverage metric.
//
// The descriptor walk mirrors secretcoverage.walk so a reader who knows one knows both:
// proto field-name dot paths, StringValueOrRef treated as a leaf, recurse into submessages,
// map/repeated handled. The annotation is read off whichever field carries it (typically a
// StringValueOrRef, singular or repeated).

//go:build !codegen
// +build !codegen

package refcheck

import (
	"sort"
	"strconv"
	"strings"

	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/refannotations"
	"github.com/plantonhq/planton/shared/catalogkind"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// StringValueOrRef is a reference leaf, not a message to recurse into: the FK annotation
// lives on the outer field, never inside the oneof.
const (
	stringValueOrRefFullName = "dev.planton.shared.foreignkey.v1.StringValueOrRef"
	valueFromRefFullName     = "dev.planton.shared.foreignkey.v1.ValueFromRef"
)

// Finding is one foreign-key annotation whose default_kind_field_path does not resolve
// against the referenced kind. Each is a hard gate failure.
type Finding struct {
	Kind       string // the kind that declares the reference, e.g. "AwsEksNodeGroup"
	Provider   string
	FieldPath  string // proto field-name dot path to the annotated field, e.g. "spec.subnet_ids"
	TargetKind string // the referenced kind (default_kind), e.g. "AwsSubnet"
	RefPath    string // the default_kind_field_path value, e.g. "status.outputs.subnet_id"
	Reason     string // why it does not resolve
}

// Analyze walks every production catalog kind and returns the foreign-key
// references that do not resolve, sorted deterministically. Hermetic `_test` kinds and
// unimplemented kinds are skipped -- matching the kind-map codegen -- so the report
// reflects the real surface.
func Analyze() []Finding {
	var findings []Finding
	for _, kind := range catalogkindreflect.KindsList() {
		provider := catalogkindreflect.GetProvider(kind)
		if provider == catalogkind.CatalogProvider_catalog_provider_unspecified {
			continue
		}
		// The `_test` provider holds hermetic fixtures; they are exercised directly by
		// unit tests, never part of the production invariant.
		if provider.String()[0] == '_' {
			continue
		}
		msg, err := catalogkindreflect.NewInstance(kind)
		if err != nil {
			// Enum value exists but the API package is not implemented yet.
			continue
		}
		specFd := msg.ProtoReflect().Descriptor().Fields().ByName("spec")
		if specFd == nil || specFd.Kind() != protoreflect.MessageKind {
			continue
		}
		walk(specFd.Message(), "spec", kind, provider.String(), map[protoreflect.FullName]bool{}, &findings)
	}
	sortFindings(findings)
	return findings
}

func walk(md protoreflect.MessageDescriptor, prefix string, declaringKind catalogkind.CatalogKind, provider string, visited map[protoreflect.FullName]bool, out *[]Finding) {
	if visited[md.FullName()] {
		return
	}
	visited[md.FullName()] = true
	defer delete(visited, md.FullName())

	fields := md.Fields()
	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		path := string(fd.Name())
		if prefix != "" {
			path = prefix + "." + path
		}

		*out = append(*out, checkField(fd, path, declaringKind, provider)...)

		// Recurse to find nested FK annotations, but never into the StringValueOrRef leaf.
		switch {
		case fd.IsMap():
			if v := fd.MapValue(); v.Kind() == protoreflect.MessageKind && string(v.Message().FullName()) != stringValueOrRefFullName {
				walk(v.Message(), path, declaringKind, provider, visited, out)
			}
		case fd.Kind() == protoreflect.MessageKind:
			if string(fd.Message().FullName()) != stringValueOrRefFullName {
				walk(fd.Message(), path, declaringKind, provider, visited, out)
			}
		}
	}
}

// checkField validates the FK annotations on a single field. It returns one Finding per
// declared composition key that does not resolve, plus the authoring defects the
// annotations can carry on their own:
//   - a default_kind_field_path with no default_kind;
//   - candidates declared beside a default kind that is not among them (default_kind is
//     the kind a bare literal is read as, so it must be one of the field's candidates, on
//     the same path);
//   - annotations on a field that is not a reference (never read there).
//
// A field with no annotations, or an intentional kind-less reference (a route target
// that can point at any kind), has nothing to validate.
func checkField(fd protoreflect.FieldDescriptor, fieldPath string, declaringKind catalogkind.CatalogKind, provider string) []Finding {
	annotations := refannotations.Of(fd)
	if !annotations.IsReference() && annotations.DefaultKindFieldPath == "" {
		return nil
	}

	var findings []Finding
	mk := func(targetKind catalogkind.CatalogKind, refPath, reason string) {
		findings = append(findings, Finding{
			Kind:       declaringKind.String(),
			Provider:   provider,
			FieldPath:  fieldPath,
			TargetKind: targetKind.String(),
			RefPath:    refPath,
			Reason:     reason,
		})
	}

	if !isReferenceField(fd) {
		mk(annotations.DefaultKind, annotations.DefaultKindFieldPath,
			"reference annotations sit on a field that is not a StringValueOrRef or ValueFromRef -- no reader sees them there")
		return findings
	}
	if annotations.DefaultKind == catalogkind.CatalogKind_unspecified && annotations.DefaultKindFieldPath != "" {
		mk(annotations.DefaultKind, annotations.DefaultKindFieldPath, "default_kind_field_path is set but default_kind is unspecified")
		return findings
	}
	if len(annotations.Candidates) > 0 && annotations.DefaultKind != catalogkind.CatalogKind_unspecified {
		defaultKey := refannotations.Key{Kind: annotations.DefaultKind, FieldPath: annotations.DefaultKindFieldPath}
		listed := false
		for _, c := range annotations.Candidates {
			listed = listed || c == defaultKey
		}
		if !listed {
			mk(annotations.DefaultKind, annotations.DefaultKindFieldPath,
				"the field declares candidates, and its default kind (on its default path) is not one of them -- list it as a candidate")
		}
	}

	for _, key := range annotations.Keys() {
		rootMd, rest, reason := targetRoot(key.Kind, key.FieldPath)
		if reason == "" {
			reason = resolvePath(rootMd, rest)
		}
		if reason == "" {
			reason = ownNameReason(key.Kind, key.FieldPath)
		}
		if reason != "" {
			mk(key.Kind, key.FieldPath, reason)
		}
	}
	return findings
}

// ownNameReason refuses a "metadata.name" path into a kind whose spec declares its own
// `name`. metadata.name is the Planton resource's name; such a kind names the object it
// provisions with spec.name, which a chart may set differently, so the reference would
// hand the consumer a name nothing carries. The path always resolves, which is why this
// needs its own rule. It returns an empty string when the path is sound.
func ownNameReason(kind catalogkind.CatalogKind, refPath string) string {
	if refPath != "metadata.name" {
		return ""
	}
	inst, err := catalogkindreflect.NewInstance(kind)
	if err != nil {
		return ""
	}
	specFd := inst.ProtoReflect().Descriptor().Fields().ByName("spec")
	if specFd == nil || specFd.Kind() != protoreflect.MessageKind || specFd.Message().Fields().ByName("name") == nil {
		return ""
	}
	return "'metadata.name' is the Planton resource's name, but " + kind.String() +
		" names the object it provisions with spec.name, which can differ -- point at the output that publishes the object's name (status.outputs.*_name)"
}

// isReferenceField reports whether the field carries a reference: a StringValueOrRef or a
// ValueFromRef, singular, repeated or as a map value.
func isReferenceField(fd protoreflect.FieldDescriptor) bool {
	md := fd.Message()
	if fd.IsMap() {
		md = fd.MapValue().Message()
	}
	if md == nil {
		return false
	}
	name := string(md.FullName())
	return name == stringValueOrRefFullName || name == valueFromRefFullName
}

// targetRoot resolves the message descriptor the path is rooted at, dispatching on the
// path prefix against the referenced kind's top-level API message:
//   - "status.outputs." -> the kind's outputs message (deploy-time results)
//   - "spec."           -> the kind's spec message (declared inputs)
//   - "metadata."       -> the kind's metadata message (e.g. referencing a parent by name)
//
// Any other root is itself a defect.
func targetRoot(kind catalogkind.CatalogKind, refPath string) (protoreflect.MessageDescriptor, string, string) {
	inst, err := catalogkindreflect.NewInstance(kind)
	if err != nil {
		return nil, "", "referenced kind " + kind.String() + " is not a registered/implemented kind"
	}
	top := inst.ProtoReflect().Descriptor()

	switch {
	case strings.HasPrefix(refPath, "status.outputs."):
		statusFd := top.Fields().ByName("status")
		if statusFd == nil || statusFd.Kind() != protoreflect.MessageKind {
			return nil, "", "target kind " + kind.String() + " has no status message"
		}
		outputsFd := statusFd.Message().Fields().ByName("outputs")
		if outputsFd == nil || outputsFd.Kind() != protoreflect.MessageKind {
			return nil, "", "target kind " + kind.String() + " has no status.outputs message"
		}
		return outputsFd.Message(), strings.TrimPrefix(refPath, "status.outputs."), ""
	case strings.HasPrefix(refPath, "spec."):
		return childMessage(top, "spec", kind, refPath)
	case strings.HasPrefix(refPath, "metadata."):
		return childMessage(top, "metadata", kind, refPath)
	default:
		return nil, "", "field path root must be 'status.outputs.', 'spec.', or 'metadata.' (got '" + refPath + "')"
	}
}

// childMessage returns the descriptor of a direct message field on the top-level API
// message (e.g. "spec" or "metadata") plus the path remainder beneath that root.
func childMessage(top protoreflect.MessageDescriptor, root string, kind catalogkind.CatalogKind, refPath string) (protoreflect.MessageDescriptor, string, string) {
	fd := top.Fields().ByName(protoreflect.Name(root))
	if fd == nil || fd.Kind() != protoreflect.MessageKind {
		return nil, "", "target kind " + kind.String() + " has no " + root + " message"
	}
	return fd.Message(), strings.TrimPrefix(refPath, root+"."), ""
}

// resolvePath descends md by the field-name segments of path, skipping index segments
// (`[*]`, `[0]`, or a bare `0`) that denote a repeated element. It returns an empty
// string when the path resolves, or a human-readable reason otherwise.
func resolvePath(md protoreflect.MessageDescriptor, path string) string {
	var segs []string
	for _, s := range strings.Split(path, ".") {
		if s == "" || isIndexSegment(s) {
			continue
		}
		segs = append(segs, s)
	}
	if len(segs) == 0 {
		return "field path resolves to no field"
	}

	current := md
	for i, seg := range segs {
		if current == nil {
			return "path '" + path + "' descends past a scalar field"
		}
		fd := current.Fields().ByName(protoreflect.Name(seg))
		if fd == nil {
			return "no field named '" + seg + "' on " + string(current.FullName())
		}
		if i == len(segs)-1 {
			return "" // terminal field exists
		}
		if fd.Kind() != protoreflect.MessageKind {
			return "cannot descend into scalar field '" + seg + "'"
		}
		current = fd.Message()
	}
	return ""
}

// isIndexSegment reports whether a path segment denotes a repeated-element index rather
// than a field name: a bracketed `[*]`/`[0]` or a bare integer.
func isIndexSegment(s string) bool {
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		return true
	}
	_, err := strconv.Atoi(s)
	return err == nil
}

func sortFindings(f []Finding) {
	sort.Slice(f, func(i, j int) bool {
		if f[i].Kind != f[j].Kind {
			return f[i].Kind < f[j].Kind
		}
		return f[i].FieldPath < f[j].FieldPath
	})
}

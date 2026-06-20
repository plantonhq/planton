// Package refcheck validates foreign-key reference integrity across the cloud-resource
// registry: every field annotated with (foreignkey.v1.default_kind_field_path) must point
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

	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"github.com/plantonhq/openmcf/pkg/crkreflect"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// StringValueOrRef is a reference leaf, not a message to recurse into: the FK annotation
// lives on the outer field, never inside the oneof.
const stringValueOrRefFullName = "org.openmcf.shared.foreignkey.v1.StringValueOrRef"

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

// Analyze walks every production cloud-resource kind and returns the foreign-key
// references that do not resolve, sorted deterministically. Hermetic `_test` kinds and
// unimplemented kinds are skipped -- matching the kind-map codegen -- so the report
// reflects the real surface.
func Analyze() []Finding {
	var findings []Finding
	for _, kind := range crkreflect.KindsList() {
		provider := crkreflect.GetProvider(kind)
		if provider == cloudresourcekind.CloudResourceProvider_cloud_resource_provider_unspecified {
			continue
		}
		// The `_test` provider holds hermetic fixtures; they are exercised directly by
		// unit tests, never part of the production invariant.
		if provider.String()[0] == '_' {
			continue
		}
		msg, err := crkreflect.NewInstance(kind)
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

func walk(md protoreflect.MessageDescriptor, prefix string, declaringKind cloudresourcekind.CloudResourceKind, provider string, visited map[protoreflect.FullName]bool, out *[]Finding) {
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

		if f, ok := checkField(fd, path, declaringKind, provider); ok {
			*out = append(*out, f)
		}

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

// checkField validates the FK annotation on a single field, if present. It returns a
// Finding only when the annotation is present and does not resolve.
func checkField(fd protoreflect.FieldDescriptor, fieldPath string, declaringKind cloudresourcekind.CloudResourceKind, provider string) (Finding, bool) {
	opts := fd.Options()
	if opts == nil {
		return Finding{}, false
	}
	refPath, _ := proto.GetExtension(opts, foreignkeyv1.E_DefaultKindFieldPath).(string)
	if refPath == "" {
		// Either an ordinary field, or an intentional FK with no default path (e.g. a
		// route target that can point at many kinds). Nothing to validate.
		return Finding{}, false
	}
	targetKind, _ := proto.GetExtension(opts, foreignkeyv1.E_DefaultKind).(cloudresourcekind.CloudResourceKind)

	mk := func(reason string) (Finding, bool) {
		return Finding{
			Kind:       declaringKind.String(),
			Provider:   provider,
			FieldPath:  fieldPath,
			TargetKind: targetKind.String(),
			RefPath:    refPath,
			Reason:     reason,
		}, true
	}

	if targetKind == cloudresourcekind.CloudResourceKind_unspecified {
		return mk("default_kind_field_path is set but default_kind is unspecified")
	}

	rootMd, rest, reason := targetRoot(targetKind, refPath)
	if reason != "" {
		return mk(reason)
	}
	if reason := resolvePath(rootMd, rest); reason != "" {
		return mk(reason)
	}
	return Finding{}, false
}

// targetRoot resolves the message descriptor the path is rooted at, dispatching on the
// path prefix against the referenced kind's top-level API message:
//   - "status.outputs." -> the kind's stack-outputs message (deploy-time results)
//   - "spec."           -> the kind's spec message (declared inputs)
//   - "metadata."       -> the kind's metadata message (e.g. referencing a parent by name)
//
// Any other root is itself a defect.
func targetRoot(kind cloudresourcekind.CloudResourceKind, refPath string) (protoreflect.MessageDescriptor, string, string) {
	inst, err := crkreflect.NewInstance(kind)
	if err != nil {
		return nil, "", "default_kind " + kind.String() + " is not a registered/implemented kind"
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
func childMessage(top protoreflect.MessageDescriptor, root string, kind cloudresourcekind.CloudResourceKind, refPath string) (protoreflect.MessageDescriptor, string, string) {
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

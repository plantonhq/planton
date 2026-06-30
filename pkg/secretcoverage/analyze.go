// Secret-coverage analyzer: walks every production cloud-resource kind and reports,
// per string-bearing field, whether the secret-by-default `sensitive` annotation is
// present, intentionally exempted, or missing on a field that looks like a secret.
//
// Per-field logic is a descriptor-level twin of the Java security authority
// secrets-commons.SensitiveFieldWalker: proto field-NAME dot paths, StringValueOrRef
// as a gated leaf, descend non-sensitive submessages, map/repeated handled. A reader
// who knows one walk knows both.
//
// Scope is the `spec` subtree only. The `sensitive` option is actionable solely on
// the INPUT surface -- it forces a value to be a managed-secret reference and resolves
// it JIT at deploy. `status.outputs.*` are provider-computed results (a generated
// password, a connection string) that the user never supplies and cannot make a
// reference, so they are not part of the annotation sweep and would only be noise here.

//go:build !codegen
// +build !codegen

package secretcoverage

import (
	"sort"

	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/apis/dev/planton/shared/options"
	"github.com/plantonhq/planton/pkg/crkreflect"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// StringValueOrRef is a secret-bearing leaf, not a message to recurse into: the
// `sensitive` annotation lives on the outer field and only its literal `value` can
// hold a secret (a `value_from` is a cross-resource reference owned by FK resolution).
const stringValueOrRefFullName = "dev.planton.shared.foreignkey.v1.StringValueOrRef"

// Classification is the secret-coverage verdict for a single string-bearing field.
type Classification string

const (
	Covered      Classification = "covered"       // annotated `sensitive` = true
	Exempt       Classification = "exempt"        // annotated `sensitive_exempt_reason`
	Gap          Classification = "gap"           // looks sensitive by name, unannotated
	NotSensitive Classification = "not_sensitive" // neither annotated nor heuristic-positive
)

// Finding is one secret-relevant field. NotSensitive fields are not emitted -- only
// fields that are covered, exempted, or gaps (the universe the coverage % is over).
type Finding struct {
	Kind         string
	Provider     string
	Path         string // proto field-name dot path from the cloud-object root, e.g. "spec.registry_password"
	FieldName    string // leaf field name, the heuristic input
	Class        Classification
	ExemptReason string   // populated when Class == Exempt
	Violations   []string // gate failures intrinsic to the annotation (contradiction / pointless exemption)
}

// classify is the pure decision for one field. Separated from the descriptor walk so
// the full truth table -- including the two annotation-level violations -- is unit
// tested without needing fixture protos in every contradictory shape.
//
// Note: an empty exemption reason is indistinguishable from "unset" (proto3 singular
// string), so "empty reason" is not a reachable state and is intentionally not a rule.
func classify(fieldName string, isSensitive bool, exemptReason string) (Classification, []string) {
	looks := LooksSensitiveByName(fieldName)
	switch {
	case isSensitive && exemptReason != "":
		// `sensitive` wins for safety, but the contradiction is flagged so it is fixed.
		return Covered, []string{"field sets both `sensitive` and `sensitive_exempt_reason` (contradiction)"}
	case isSensitive:
		return Covered, nil
	case exemptReason != "":
		if !looks {
			return Exempt, []string{"`sensitive_exempt_reason` is set on a field whose name does not look sensitive (pointless exemption)"}
		}
		return Exempt, nil
	case looks:
		return Gap, nil
	default:
		return NotSensitive, nil
	}
}

// Analyze walks every production cloud-resource kind and returns the secret-relevant
// findings, sorted deterministically. Hermetic `_test` kinds and unimplemented kinds
// are skipped -- matching the kind-map codegen -- so the report reflects real surface.
func Analyze() []Finding {
	var findings []Finding
	for _, kind := range crkreflect.KindsList() {
		provider := crkreflect.GetProvider(kind)
		if provider == cloudresourcekind.CloudResourceProvider_cloud_resource_provider_unspecified {
			continue
		}
		// The `_test` provider holds hermetic fixtures (testcloudresourcegeneric); they
		// are exercised directly by unit tests, never counted in production coverage.
		if provider.String()[0] == '_' {
			continue
		}
		msg, err := crkreflect.NewInstance(kind)
		if err != nil {
			// Enum value exists but the API package is not implemented yet.
			continue
		}
		specField := msg.ProtoReflect().Descriptor().Fields().ByName("spec")
		if specField == nil || specField.Kind() != protoreflect.MessageKind {
			continue
		}
		findings = append(findings, CollectFindings(specField.Message(), kind.String(), provider.String())...)
	}
	sortFindings(findings)
	return findings
}

// CollectFindings walks a single SPEC message descriptor, rooting paths at "spec" to
// match the enforcement walker's path format. Exposed so tests can drive it against
// the hermetic testcloudresourcegeneric spec in isolation.
func CollectFindings(specMd protoreflect.MessageDescriptor, kindName, provider string) []Finding {
	var out []Finding
	walk(specMd, "spec", kindName, provider, map[protoreflect.FullName]bool{}, &out)
	return out
}

func walk(md protoreflect.MessageDescriptor, prefix, kindName, provider string, visited map[protoreflect.FullName]bool, out *[]Finding) {
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

		switch {
		case fd.IsMap():
			// The annotation lives on the map field; the value is the secret slot.
			if v := fd.MapValue(); isStringLeaf(v) {
				addLeaf(fd, path, kindName, provider, out)
			} else if v.Kind() == protoreflect.MessageKind {
				walk(v.Message(), path, kindName, provider, visited, out)
			}
		case fd.IsList():
			if isStringLeaf(fd) {
				addLeaf(fd, path, kindName, provider, out)
			} else if fd.Kind() == protoreflect.MessageKind {
				walk(fd.Message(), path, kindName, provider, visited, out)
			}
		case fd.Kind() == protoreflect.MessageKind:
			if string(fd.Message().FullName()) == stringValueOrRefFullName {
				addLeaf(fd, path, kindName, provider, out)
			} else {
				walk(fd.Message(), path, kindName, provider, visited, out)
			}
		case fd.Kind() == protoreflect.StringKind:
			addLeaf(fd, path, kindName, provider, out)
		}
	}
}

// isStringLeaf reports whether a field (or a map value descriptor) carries a string
// secret slot: a string, or a StringValueOrRef whose literal value is a string.
func isStringLeaf(fd protoreflect.FieldDescriptor) bool {
	if fd.Kind() == protoreflect.StringKind {
		return true
	}
	return fd.Kind() == protoreflect.MessageKind && string(fd.Message().FullName()) == stringValueOrRefFullName
}

func addLeaf(fd protoreflect.FieldDescriptor, path, kindName, provider string, out *[]Finding) {
	sensitive, exemptReason := leafOptions(fd)
	class, violations := classify(string(fd.Name()), sensitive, exemptReason)
	if class == NotSensitive {
		return
	}
	*out = append(*out, Finding{
		Kind:         kindName,
		Provider:     provider,
		Path:         path,
		FieldName:    string(fd.Name()),
		Class:        class,
		ExemptReason: exemptReason,
		Violations:   violations,
	})
}

func leafOptions(fd protoreflect.FieldDescriptor) (sensitive bool, exemptReason string) {
	opts := fd.Options()
	if opts == nil {
		return false, ""
	}
	if v, ok := proto.GetExtension(opts, options.E_Sensitive).(bool); ok {
		sensitive = v
	}
	if v, ok := proto.GetExtension(opts, options.E_SensitiveExemptReason).(string); ok {
		exemptReason = v
	}
	return sensitive, exemptReason
}

func sortFindings(f []Finding) {
	sort.Slice(f, func(i, j int) bool {
		if f[i].Kind != f[j].Kind {
			return f[i].Kind < f[j].Kind
		}
		return f[i].Path < f[j].Path
	})
}

// Summary aggregates findings for the human/agent report.
type Summary struct {
	Covered    int
	Exempt     int
	Gap        int
	Violations int
}

// CoveragePercent is covered+exempt over the secret-relevant universe (covered+exempt+gap).
// Returns 100 when there is nothing to cover.
func (s Summary) CoveragePercent() float64 {
	denom := s.Covered + s.Exempt + s.Gap
	if denom == 0 {
		return 100
	}
	return 100 * float64(s.Covered+s.Exempt) / float64(denom)
}

func Summarize(findings []Finding) Summary {
	var s Summary
	for _, f := range findings {
		switch f.Class {
		case Covered:
			s.Covered++
		case Exempt:
			s.Exempt++
		case Gap:
			s.Gap++
		}
		s.Violations += len(f.Violations)
	}
	return s
}

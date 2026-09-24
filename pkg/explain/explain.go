// Package explain renders API schemas as reference documentation, offline,
// from the proto descriptors compiled into the binary -- the engine behind
// `planton explain` (kubectl-explain for every Planton API).
//
// The engine is deliberately host-neutral: it walks any resource message
// following the KRM envelope shape (apiVersion/kind/metadata/spec/status)
// and resolves dotted field paths against it. Everything host-specific
// enters through three seams:
//
//   - OptionInterpreter: reads custom proto options into report fields.
//     This repo's interpreter covers dev.planton.shared options (sensitive,
//     recommended defaults, foreign keys); a host that defines its own
//     option family adds another interpreter, never a fork of the walker.
//   - Dispatcher: continues path resolution across a kind-valued boundary
//     (a field whose VALUE selects another schema, e.g. an envelope's kind
//     enum next to an untyped payload). Without it such paths dead-end.
//   - DocLookup: resolves proto-source documentation by full name (the Go
//     protobuf runtime strips comments; pkg/protodocs restores them).
//
// The report this package produces is a published contract consumed by AI
// agents composing manifests and charts: extend it additively, never rename
// or repurpose fields casually.
package explain

import (
	"github.com/pkg/errors"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// DocLookup resolves proto-source documentation for a descriptor full name.
// Empty string means undocumented.
type DocLookup func(protoreflect.FullName) string

// Resource is one explainable root: a name the user can type, the KRM
// envelope message behind it, and its apiVersion.
type Resource struct {
	// Name is the canonical kind name rendered in reports (e.g. AwsVpc).
	Name string
	// ApiVersion is the KRM group/version (e.g. aws.planton.dev/v1alpha1).
	ApiVersion string
	// Message is the resource's envelope descriptor
	// (apiVersion/kind/metadata/spec/status).
	Message protoreflect.MessageDescriptor
}

// Engine explains resources. Construct once per binary with the host's
// interpreters, dispatchers, and doc sources; Explain is then pure.
type Engine struct {
	Interpreters []OptionInterpreter
	Dispatchers  []Dispatcher
	Docs         DocLookup
}

// Report is the machine-readable answer to one explain invocation. Exactly
// one of the two views is populated:
//
//   - the ROOT view (Path empty): Spec/SpecRules/Outputs -- the fields a
//     manifest author writes and the outputs other resources can reference.
//     The envelope plumbing (apiVersion, kind, metadata) is uniform across
//     the platform and deliberately not repeated in every root report,
//     though paths into it (e.g. `metadata`) still resolve honestly.
//   - the FIELD view (Path set): the resolved field with its children.
type Report struct {
	Kind       string `json:"kind"`
	ApiVersion string `json:"apiVersion,omitempty"`
	// Path is the resolved dotted field path below the resource name; empty
	// for the root view.
	Path string `json:"path,omitempty"`
	// Doc is the resource's own documentation (root view only).
	Doc string `json:"doc,omitempty"`

	Field *Field `json:"field,omitempty"`

	Spec []Field `json:"spec,omitempty"`
	// SpecRules are the spec's cross-field constraints (CEL), in the
	// empathetic wording the proto authors wrote for humans and agents.
	SpecRules []Rule `json:"specRules,omitempty"`
	// Outputs are the deployment outputs of this resource. Another resource
	// references them as valueFrom {kind, name, fieldPath:
	// "status.outputs.<name>"} -- names here are the exact fieldPath leaves.
	Outputs []Field `json:"outputs,omitempty"`
}

// Rule is one cross-field constraint on a message.
type Rule struct {
	Id      string `json:"id,omitempty"`
	Message string `json:"message"`
}

// EnumValue is one allowed value of an enum field, with its documentation.
type EnumValue struct {
	Name string `json:"name"`
	Doc  string `json:"doc,omitempty"`
}

// Field is one manifest field. Name is the protojson name -- the exact key
// written in YAML manifests and chart templates. ProtoName is the proto
// field name (snake_case) -- the canonical spelling of a valueFrom
// fieldPath, because the control plane stores cloud objects with proto
// field names and canonicalizes reference paths to them (camelCase is
// tolerated on input and rewritten).
type Field struct {
	Name      string `json:"name"`
	ProtoName string `json:"protoName"`
	Type      string `json:"type"`
	Doc       string `json:"doc,omitempty"`

	Required bool `json:"required,omitempty"`
	// Optional mirrors the proto3 `optional` keyword (explicit presence):
	// leaving the field unset is semantically different from the zero value.
	Optional  bool `json:"optional,omitempty"`
	Sensitive bool `json:"sensitive,omitempty"`
	// SecretHome is set on a field whose value is written where anyone who
	// can view the resource reads it: it names (in the authored spelling)
	// the sibling field where the kind keeps a secret instead. A platform
	// that resolves secret references refuses one anywhere within this field.
	SecretHome string `json:"secretHome,omitempty"`
	// RecommendedDefault is the value the platform recommends when the
	// field is left unset.
	RecommendedDefault string `json:"recommendedDefault,omitempty"`
	// RefKind/RefFieldPath describe the default target of a foreign-key
	// field (a StringValueOrRef): which kind it usually points at and which
	// output field the reference reads.
	RefKind      string `json:"refKind,omitempty"`
	RefFieldPath string `json:"refFieldPath,omitempty"`
	// Provenance labels who writes the field when it is not the manifest
	// author: "assembled" (tooling fills it from companion files) or
	// "computed" (the platform derives it). Empty means hand-authored.
	Provenance string `json:"provenance,omitempty"`

	Enum []EnumValue `json:"enum,omitempty"`
	// Constraints carries the field's validation rules: the human-worded
	// messages of CEL rules plus a compact JSON dump of the remaining
	// structured buf.validate rules so no constraint class is dropped.
	Constraints []string `json:"constraints,omitempty"`

	Fields []Field `json:"fields,omitempty"`
}

// Explain resolves path (dotted protojson segments below the resource name,
// already split) against the resource and renders the report. An empty path
// yields the root view.
func (e *Engine) Explain(res Resource, path []string) (*Report, error) {
	if res.Message == nil {
		return nil, errors.Errorf("resource %s has no message descriptor", res.Name)
	}
	if len(path) == 0 {
		return e.rootReport(res)
	}
	return e.resolvePath(res, path)
}

// rootReport renders the manifest-author view of a resource: spec fields,
// spec rules, and deployment outputs (status.outputs on the KRM envelope).
func (e *Engine) rootReport(res Resource) (*Report, error) {
	report := &Report{
		Kind:       res.Name,
		ApiVersion: res.ApiVersion,
	}

	specField := res.Message.Fields().ByName("spec")
	if specField == nil || specField.Message() == nil {
		return nil, errors.Errorf("resource %s has no spec message", res.Name)
	}
	specMsg := specField.Message()

	// The authored resource description lives on the SPEC message in this
	// catalog's house style; the envelope message usually carries only a
	// stub comment, so it is the fallback rather than the source.
	report.Doc = e.doc(specMsg.FullName())
	if report.Doc == "" {
		report.Doc = e.doc(res.Message.FullName())
	}
	report.Spec = e.walkFields(specMsg, map[protoreflect.FullName]bool{specMsg.FullName(): true})
	report.SpecRules = messageCelRules(specMsg)

	if statusField := res.Message.Fields().ByName("status"); statusField != nil && statusField.Message() != nil {
		if outputsField := statusField.Message().Fields().ByName("outputs"); outputsField != nil && outputsField.Message() != nil {
			outMsg := outputsField.Message()
			report.Outputs = e.walkFields(outMsg, map[protoreflect.FullName]bool{outMsg.FullName(): true})
		}
	}
	return report, nil
}

// doc looks up documentation, tolerating an engine constructed without a
// doc source (reports simply carry no prose).
func (e *Engine) doc(name protoreflect.FullName) string {
	if e.Docs == nil {
		return ""
	}
	return e.Docs(name)
}

// Package generators provides proto-aware Terraform artifact generation for
// Planton cloud components.
//
// It replaces the earlier pkg/iac/tofu/tfvars and pkg/iac/tofu/variablestf
// packages with a unified implementation that shares a single TypeRule registry
// across both generators. This ensures that Planton's domain types (such as
// StringValueOrRef and ValueFromRef) are handled consistently
// whether generating terraform.tfvars or variables.tf.
//
// # Architecture
//
// The package is organized around three concerns:
//
//  1. Type Rules (typerules.go) -- a registry mapping proto message full names
//     to Terraform translation behaviors: flatten to a primitive, skip entirely,
//     or recurse normally. Adding a new wrapper type is one registry entry.
//
//  2. tfvars generation (flatten.go, hclwrite.go, tfvars.go) -- converts a
//     proto message to HCL-formatted terraform.tfvars. The pipeline is:
//     protojson -> JSON map -> Flatten (applies type rules using proto
//     descriptors) -> WriteMapToHCL -> string. Two emission modes exist,
//     selected per kind by RenderTFVars:
//     - snake_case (ProtoToTFVars): keys renamed to proto snake_case to match
//     the generated snake_case variables.tf that provider-abstraction
//     modules consume.
//     - camelCase (ProtoToManifestTFVars): keys kept as the CRD's camelCase
//     JSON, for kinds whose CloudResourceKindMeta carries a
//     kubernetes_manifest_projection -- their `spec` is fed verbatim to a
//     kubernetes_manifest passthrough module (see manifestmodule.go).
//
//  3. variables.tf generation (tftype.go, variablestf.go) -- walks a proto
//     message descriptor to produce Terraform variable blocks. Consults the
//     same type rules to flatten wrapper types to primitives and skip
//     orchestrator-only fields.
//
//  4. thin manifest-module generation (manifestmodule.go) -- for projection
//     kinds, emits the entire iac/tf/ module (any-typed spec passthrough), so
//     no hand-written snake->camel/prune/oneOf locals.tf is needed.
//
// Note: despite the generic-sounding name, this package is planton-domain-aware
// (it hardcodes planton type rules and reads kind metadata via crkreflect); it
// is not a standalone proto->HCL library.
//
// # The optional() contract (renderer and module must agree)
//
// The tfvars renderer (ProtoToTFVars) marshals with protojson
// EmitUnpopulated=false, so any proto field left at its zero value is ABSENT
// from the emitted terraform.tfvars (it is "null-pruned"). A Terraform object
// type rejects a value that omits a non-optional attribute. Therefore every
// attribute that the renderer may prune MUST be declared optional() in
// variables.tf, with a default equal to the proto zero value so the pruned field
// reconstructs to the same zero. variables.tf generation enforces this by
// construction:
//
//   - An attribute is REQUIRED (left bare) only when the proto field carries a
//     presence guarantee: (buf.validate.field).required = true, or a
//     presence-implying constraint such as string min_len >= 1 or repeated
//     min_items >= 1. See isRequiredField.
//   - Every other attribute is optional(<type>, <zero>): string -> "", number ->
//     0, bool -> false, map -> {}, list -> []. Nested objects and `any` default
//     to null (consumers null-guard with try()/!= null).
//   - The shared resource envelope (CloudResourceMetadata) is emitted from one
//     canonical block (name required; id/org/env/labels/annotations/tags
//     optional), independent of the constraint-free envelope proto.
//
// Output is deterministic and offline -- it depends only on the compiled proto
// descriptor and the committed proto documentation index, never on a network
// call -- so the committed variables.tf can be regenerated and diffed.
// TestVariablesTFDrift guards this: for every enrolled kind, the committed
// variables.tf must equal the generator output, making the generator the
// single source of truth and preventing any regression to a hand-edited or
// legacy (all-required) schema. Enrollment has two shapes: a whole provider
// (every registered kind, so a new kind is guarded the day it is registered)
// or an individual kind, with named exceptions for a module the generator
// cannot yet express -- see the drift test.
//
// # The generated file is formatted and documented
//
// The text is passed through hclwrite.Format before it is returned, so it is
// byte-identical to what `tofu fmt` writes: the repository's formatting gate
// runs on every changed module, and a generator whose output that gate
// rewrites could not own the committed file.
//
// Each nested attribute carries its proto field's documentation as `#`
// comment lines directly above it. Nested object attributes have no
// `description` in Terraform's type language, so the comment is the only place
// the module can say what an input means, and it is what whoever wires main.tf
// reads first. The protobuf runtime strips comments from descriptors, so the
// text is read from the embedded pkg/protodocs index (distilled from the same
// .proto sources by `make generate-proto-docs`); a field with no entry gets no
// comment. The two inputs move together: regenerate the index before
// regenerating variables, or the comments lag the protos until the next
// regeneration. A field a type rule collapses from a wrapper message to a
// primitive additionally carries that rule's FlattenNote -- the proto
// documentation describes the wrapper, the attribute is the primitive, and the
// note bridges the two.
//
// # Free-form JSON maps must be `any`, not map(any)
//
// A field typed map<string, google.protobuf.Struct> (e.g. AwsIamRole.inline_policies,
// a map of policy-name -> arbitrary policy document) holds entries with independent,
// heterogeneous shapes. Terraform's map(any) requires every element to converge to a
// SINGLE common type and fails input validation with
//
//	attribute "X": all map elements must have the same type
//
// the moment two differently-shaped entries are passed. The generator therefore emits
// such fields as the bare `any` keyword (TFFreeFormMap), letting Terraform infer a
// heterogeneous object; the optional default stays an empty map ({}). A single
// google.protobuf.Struct field (e.g. AwsIamRole.trust_policy) is already `any` for the
// same reason -- only the map case needed fixing.
//
// Module contract for these fields: because the value is `any` (an object, not a map),
// a consuming module must NOT pass it straight to for_each. The canonical idiom is to
// encode at the boundary into a homogeneous map(string) first:
//
//	inline_policies_json = { for k, v in try(var.spec.inline_policies, {}) : k => jsonencode(v) }
//	# then: for_each = local.inline_policies_json ; policy = each.value
//
// # Extensibility
//
// To handle a new Planton wrapper type, add one entry to DefaultRules() in
// typerules.go. Both generators will immediately respect the new rule. No
// other code changes are required.
package generators

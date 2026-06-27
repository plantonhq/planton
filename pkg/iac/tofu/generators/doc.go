// Package generators provides proto-aware Terraform artifact generation for
// OpenMCF cloud components.
//
// It replaces the earlier pkg/iac/tofu/tfvars and pkg/iac/tofu/variablestf
// packages with a unified implementation that shares a single TypeRule registry
// across both generators. This ensures that OpenMCF's domain types (such as
// StringValueOrRef and KubernetesClusterSelector) are handled consistently
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
// Note: despite the generic-sounding name, this package is openmcf-domain-aware
// (it hardcodes openmcf type rules and reads kind metadata via crkreflect); it
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
// descriptor, never on a network call -- so the committed variables.tf can be
// regenerated and diffed. TestVariablesTFDrift guards this: for every migrated
// kind, the committed variables.tf must equal the generator output, making the
// generator the single source of truth and preventing any regression to a
// hand-edited or legacy (all-required) schema.
//
// # Extensibility
//
// To handle a new OpenMCF wrapper type, add one entry to DefaultRules() in
// typerules.go. Both generators will immediately respect the new rule. No
// other code changes are required.
package generators

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
//     descriptors) -> WriteMapToHCL -> string.
//
//  3. variables.tf generation (tftype.go, variablestf.go) -- walks a proto
//     message descriptor to produce Terraform variable blocks. Consults the
//     same type rules to flatten wrapper types to primitives and skip
//     orchestrator-only fields.
//
// # Extensibility
//
// To handle a new OpenMCF wrapper type, add one entry to DefaultRules() in
// typerules.go. Both generators will immediately respect the new rule. No
// other code changes are required.
package generators

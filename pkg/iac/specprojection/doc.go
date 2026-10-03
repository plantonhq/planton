// Package specprojection turns a Planton manifest message into the plain value
// map an IaC engine consumes, the same way for every engine.
//
// The pipeline is protojson (unpopulated fields omitted, so the map carries no
// nulls) -> JSON map -> Flatten, which walks the map beside the message
// descriptor and applies the type rules (typerules.go): Planton's wrapper types
// collapse to the primitive an engine expects (a StringValueOrRef becomes its
// resolved string), orchestrator-only messages are dropped, free-form JSON
// well-known types pass through verbatim, manifest-only fields never leave
// the manifest, and the two Kubernetes shape markers write an IntOrString as a
// number or a name and a list-valued map's wrapper values as bare lists.
//
// Two key styles exist (KeyStyle). Provider-abstraction Terraform modules read
// snake_case variables. Kubernetes-manifest-projection kinds read the custom
// resource's own keys, which are protojson's names; their Terraform module and
// their Pulumi module both build the object from this one projection, so the
// two engines cannot disagree about what a spec means.
//
// The package is engine-neutral on purpose: the Terraform generators
// (pkg/iac/tofu/generators) and the Kubernetes manifest projection
// (pkg/kubernetes/manifestprojection) both import it, and neither engine's
// libraries are imported here.
package specprojection

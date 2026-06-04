# IaC Module Parity (Tofu <-> Pulumi)

Every cloud-resource kind ships two IaC implementations under `apis/.../<kind>/v1/iac/`:
a Pulumi module (`pulumi/module/*.go`) and an OpenTofu module (`tf/*.tf`). For a given
`stack-input` they MUST produce the same cloud objects, names, labels, selectors,
environment, and stack outputs. A divergence here is not cosmetic: it silently changes
what gets deployed depending on which provisioner a resource happens to use.

This note is the standing "keep an eye out for drift" practice. Read it whenever you
touch a module on either side (or add a new kind).

## What is enforced automatically (don't re-litigate by hand)

- **Stack-outputs conformance** -- `pkg/outputs/conformance_test.go`
  (`TestStackOutputsConformance`). Both engines feed the same generic transformer
  (`pkg/outputs.TransformRaw` -> `Flatten` -> `populateMessage`), so a single bar per
  kind -- "this representative output set fully populates the `StackOutputs` proto with
  nothing left unmapped" -- enforces cross-engine output parity. Add a case for each
  kind whose outputs you care about. You can also dry-run a module ad hoc:
  `openmcf validate-outputs --kind <Kind> --module-dir <dir> --sample-outputs <json>`.
- **Output transform convention** -- emit outputs that flatten to the proto field
  paths. Scalars are plain outputs; nested proto messages (e.g. `KubernetesSecretKey`)
  are emitted as nested objects (`output "password_secret" { value = { name = ..., key = ... } }`),
  which `Flatten` turns into `password_secret.name` / `password_secret.key`. Do NOT emit
  flat names like `password_secret_name` -- they never reach the nested proto field. Of
  the 364 tofu modules, zero use a `transform-outputs` executable or
  `output_transform.yaml`; the generic path is the convention. Reach for an override
  only when an upstream provider's output genuinely cannot be shaped to the proto.

## Manual parity checklist (the hand-written logic no tool can diff)

When changing a provider-resource module (`locals.tf`/`database.tf`/`main.tf` and the
matching `pulumi/module/*.go`), confirm both sides agree on:

- **Namespace source.** Use `spec.namespace` (NOT a resource id or a derived name).
- **Resource naming basis.** Both engines name the created objects (operator CRs, pod
  annotations, secret names) off the SAME field -- `metadata.name` is the established
  basis. Don't introduce a parallel `metadata.id`-based name on one side.
- **Labels.** Same keys and values. The resource-identity labels are the
  `kuberneteslabelkeys` set (`planton.ai/resource`, `planton.ai/name`, `planton.ai/kind`,
  `planton.ai/id`, `planton.ai/organization`, `planton.ai/environment`); the kind value
  is the `CloudResourceKind` enum string (e.g. `KubernetesPostgres`), and the id label is
  present only when `metadata.id` is set.
- **Pod / service selectors.** Selectors must match the labels the operator/helm chart
  actually puts on the workload pods (e.g. Zalando/Spilo pods are `application: spilo`),
  NOT our resource-identity labels. A wrong selector matches zero pods and silently
  breaks connectivity while still "succeeding".
- **Spec feature coverage.** Every behavior on one side exists on the other: backup,
  restore/standby, ingress, env injection, resource sizing, etc. The proto `spec` is the
  contract -- if it has a field, both modules must honor it.
- **Outputs shape.** Both engines export the same `StackOutputs` field set (see the
  automated conformance guard above).

## variables.tf (a generated *scaffold*, curated in practice)

`openmcf tofu generate-variables <Kind>` (`pkg/iac/tofu/generators`) renders a starting
`variables.tf` from the spec proto, but the committed convention is the curated
`optional()` form (used by the large majority of modules, e.g. `kubernetesnamespace`,
`kubernetescronjob`). The generator's raw output makes every field required, which is not
runtime-compatible with the generated `terraform.tfvars` (it omits unset fields). So:

- Treat the generator as a reference for *coverage*, not a file to commit verbatim.
- When a spec field is added, add the matching `variable` (in the curated `optional()`
  style) so partial tfvars still apply. Diffing against `generate-variables` output is a
  quick way to spot a missing field.

## Worked example

The Postgres tofu module was brought to parity with its Pulumi counterpart (correct
namespace source, `metadata.name` naming basis, `application: spilo` LB selector,
`planton.ai/*` labels, backup + disaster-recovery standby/env, and nested secret
outputs). See the conformance guard's `KubernetesPostgres` case and its negative
counterpart `TestStackOutputsConformance_DetectsFlatSecretDrift`.

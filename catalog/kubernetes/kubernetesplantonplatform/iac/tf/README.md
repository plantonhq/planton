# KubernetesPlantonPlatform Terraform Module

Declares one `PlantonPlatform` custom resource that the Planton operator
reconciles into a running self-hosted platform, plus the optional owning
namespace. The CR is rendered from null-pruned locals
(`locals.platform_spec`) through `kubectl_manifest` + `yamlencode` — the
exact semantic twin of the Pulumi module's
`apiextensions.NewCustomResource` + `platformSpecBody`.

## Module Behavior

- **Keys render only when declared** — every block in `locals.tf` is a
  null-pruned object, so the operator's own defaulting stays
  authoritative for everything unset (the umbrella chart's
  verbatim-pass-through posture). Three-state optionals (the
  default-true toggles, the defaulted scalars) render exactly when
  PRESENT in the manifest.
- **The CR is namespaced and named from `metadata.name`** — the prefix
  of every object the operator creates for the platform, which is why
  the outputs (gateway Service, setup-code Secret, the two exact
  commands) derive deterministically at declaration time.
- **Destroy is garbage collection, not operator work** — every
  operator-created object is owner-referenced to the CR (the operator
  has no finalizers), so deletion completes even when the operator is
  already gone. The 15-minute delete timeout is headroom, never an
  expected wait.
- **`version` is always explicit** — required by the spec (mirroring the
  CRD), never defaulted by the module: a module default would turn a
  catalog update into a silent whole-platform upgrade on the next apply.
- **Server-side apply with `force_conflicts`** — re-applies converge
  against the operator's own field ownership without flapping.

## Provider Choice

`alekc/kubectl` because `kubectl_manifest` needs no cluster connection at
plan time — the CRD installed by the prerequisite
KubernetesPlantonOperator may not exist yet when a composed Infra Chart
plans this resource.

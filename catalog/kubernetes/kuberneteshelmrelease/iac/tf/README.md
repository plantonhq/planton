# Kubernetes Helm Release - Terraform Module

## Overview

This Terraform module installs an upstream Helm chart as a **real Helm release** via the `helm_release` resource: hooks run, the release secret is written, and `helm list` shows the release exactly as if the Helm CLI had installed it. It is the semantic twin of the kind's Pulumi module — every lifecycle knob maps 1:1 between the two, and both merge the values layers with identical precedence.

This kind is the catalog's sole intentional passthrough — for charts no first-class kind covers. Where a typed kind exists for the workload, it always wins.

## Architecture

```
iac/tf/
├── provider.tf       # Terraform, kubernetes, and helm provider requirements
├── variables.tf      # Input variables mirroring spec.proto
├── locals.tf         # Derived values: labels, release name, namespace, Helm defaults
├── main.tf           # Optional labeled kubernetes_namespace_v1 resource
├── helm_release.tf   # The helm_release resource: values layers + lifecycle knobs
├── outputs.tf        # Exports the release's observable handles
└── README.md         # This file
```

## How It Works

1. **Variable Input**: The `spec` variable mirrors the protobuf schema; the namespace value-or-ref arrives flattened to a plain string
2. **Namespace**: When `create_namespace` is true, `main.tf` creates the namespace as an explicit, module-owned `kubernetes_namespace_v1` stamped with the standard Planton governance labels — never via `helm_release`'s own `create_namespace` flag, which would create it unlabeled. The release depends on it explicitly
3. **Identity resolution**: `locals.tf` resolves the release name (`spec.release_name`, else `metadata.name`) and Helm's own defaults for the optional knobs (`timeout_seconds` 300, `max_history` 10) so both engines send identical values whether or not the spec set the fields
4. **Chart resolution**: `repository` + `chart` + `version` pass straight through; for `oci://` repositories the provider joins repo and chart into the full OCI reference internally (the Pulumi module performs the identical join client-side). Private-repo credentials ride `repository_username` / `repository_password`
5. **Values merge**: `values_yaml` rides the `values` list; `set` entries become type-`auto` set attributes (Helm `--set` coercion), `set_string` entries become type-`string` (literal strings), and `set_sensitive` entries ride the `set_sensitive` attribute. The provider merges them in exactly the documented precedence — values_yaml, then set, then set_string, then set_sensitive — and runs Helm's own `strvals` parser on the entries, so dotted paths and coercion behave identically to the Pulumi module
6. **Lifecycle knobs**: `atomic`, `cleanup_on_fail`, `wait` (the inversion of the spec's `skip_await`), `wait_for_jobs`, `timeout`, `dependency_update`, `max_history`, `replace`, `force_update`, `reuse_values`, `reset_values`, `disable_webhooks`, `disable_openapi_validation`, `take_ownership`, and `description` — each mapped 1:1 with the Pulumi module's arguments
7. **CRD lifecycle**: the generated `helm_crds.tf` (the catalog's shared block for charts that carry CRDs) renders the pinned chart at plan time with exactly the values, set layers, and credentials the release uses, keeps the CustomResourceDefinition documents from the chart's `crds/` directory, stamps them with `planton.ai/crd-source-chart` and `planton.ai/crd-source-version`, and applies each one as a `kubectl_manifest` keyed by the CRD's own name ahead of the release, which installs with `skip_crds = true`. `apply_only` keeps them on destroy unless `crds.keep_on_uninstall` is false; server-side apply re-adopts kept CRDs on reinstall; `data "kubernetes_resources"` reads the stamps so a `version` below what the cluster carries fails the plan with what was observed, what it means, and the next step. CRDs the chart templates as release resources stay Helm's: with `helm.sh/resource-policy: keep` on them the chart installs as is; without it the plan refuses with the CRD names and the remedies unless `crds.allow_helm_managed` is true. A chart without CRDs (most charts) renders nothing here
8. **Output Export**: The release's observable handles are read from the `helm_release` resource's recorded metadata — what Helm actually installed, not what the spec asked for

## Semantics Preserved by the Module

- **A real release** — hooks, history, the release secret; `helm status`, `helm history`, and `helm rollback` all work on it
- **Identical values merge to the Pulumi module** — the provider's list-order merge plus lexical map iteration matches the Pulumi module's sorted-key `strvals` application, keeping even same-path collisions deterministic
- **`set_sensitive` entries are masked individually** in plans and state (the Pulumi module masks more coarsely — the whole merged values map — when sensitive entries are present; same installed release either way)
- **`take_ownership` works here** — this module is currently the only engine that honors it (Helm `--take-ownership`, the migration knob for adopting resources an earlier tool created; requires helm provider >= 3.1). The Pulumi module rejects the flag loudly at its pinned SDK rather than silently dropping it

## Usage

```hcl
module "helm_release" {
  source = "./iac/tf"

  metadata = {
    name = "podinfo"
  }

  spec = {
    namespace        = "podinfo"
    create_namespace = true
    repo             = "https://stefanprodan.github.io/podinfo"
    chart            = "podinfo"
    version          = "6.9.2"

    values_yaml = <<-EOT
      replicaCount: 2
    EOT

    set_string = {
      "image.tag" = "6.9.2"
    }

    atomic          = true
    cleanup_on_fail = true
  }
}
```

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| `metadata` | Resource metadata (name, org, env) | object | yes |
| `spec` | KubernetesHelmRelease specification | object | yes |

## Outputs

| Name | Description |
|------|-------------|
| `namespace` | The namespace the release is installed in |
| `release_name` | The Helm release name (`helm list` NAME column) |
| `version` | The installed chart version |
| `app_version` | The chart's appVersion (the packaged application's upstream version) |
| `status` | The release status as Helm records it (e.g. `deployed`) |
| `revision` | The release revision number (1 on install, incremented by upgrades/rollbacks) |

> **Note**: Chart values pass through to the chart unvalidated — the chart's values surface is the contract, and a typo'd value surfaces when the chart renders or its resources misbehave, not at plan. `version` is required by the spec: unpinned installs are not reproducible.

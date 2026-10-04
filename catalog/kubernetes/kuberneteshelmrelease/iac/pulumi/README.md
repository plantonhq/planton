# Kubernetes Helm Release - Pulumi Module

## Overview

This Pulumi (Go) module installs an upstream Helm chart as a **real Helm release** via `helm.v3.Release`: hooks run, the release secret is written, and `helm list` shows the release exactly as if the Helm CLI had installed it. The render-only `helm.v3.Chart` resource is deliberately NOT used — it template-renders client-side without creating a release, which silently skips hooks and leaves nothing for Helm tooling to manage.

The module is the semantic twin of the kind's Terraform module: every lifecycle knob maps 1:1 onto a `helm_release` argument, and both engines merge the values layers with identical precedence.

This kind is the catalog's sole intentional passthrough — for charts no first-class kind covers. Where a typed kind exists for the workload, it always wins.

## Architecture

```
iac/pulumi/
├── main.go             # Entrypoint: loads IaC input, calls the module
├── Pulumi.yaml         # Pulumi project configuration
└── module/
    ├── main.go         # Orchestration: provider, namespace, helm.v3.Release, outputs
    ├── values.go       # Module-side values merge (Helm's own strvals parser)
    ├── locals.go       # Release name, namespace, labels, Helm defaults
    ├── namespace.go    # Optional labeled namespace resource
    └── outputs.go      # Exports the release's observable handles
```

## How It Works

1. **Provider**: A Kubernetes provider is built from the credential in the IaC input
2. **Namespace**: When `create_namespace` is true, the namespace is created as an explicit, module-owned resource stamped with the standard Planton governance labels — never via Helm's own create-namespace flag, which would create it unlabeled. The release depends on it explicitly
3. **Identity resolution**: `locals.go` resolves the release name (`spec.release_name`, else `metadata.name`) and Helm's own defaults for the optional knobs (`timeout_seconds` 300, `max_history` 10) so both engines send identical values whether or not the spec set the fields
4. **Values merge** (`values.go`): the layers merge module-side with the documented precedence — `values_yaml` is parsed as YAML, then `set` entries apply with Helm's own `strvals` `--set` parser (coercion: `"true"` → bool, digits → number, `"null"` deletes), then `set_string` and `set_sensitive` apply with the `--set-string` parser (literal strings). Entries apply in sorted-key order, matching Terraform's lexical map iteration, so even same-path collisions resolve identically on both engines. The Release receives one final merged map
5. **Chart resolution**: for `oci://` repositories the chart reference is joined client-side as `<repo>/<chart>`; for HTTP(S) repositories the bare chart name plus the repo URL ride the repository options. Private-repo credentials ride the repository options in both forms, with the password marked secret
6. **Release**: `helm.v3.Release` installs the chart with every lifecycle knob mapped 1:1 with the Terraform module's `helm_release` arguments (`SkipAwait` maps directly; Terraform inverts it as `wait`)
7. **Output export**: the release's observable handles are read from the Release resource's status — what Helm actually recorded, not what the spec asked for

## Semantics Preserved by the Module

- **A real release** — hooks, history, the release secret; `helm status`, `helm history`, and `helm rollback` all work on it
- **Identical values merge to the Terraform module** — both paths run Helm's own `strvals` parser on override entries in the same order
- **Secret handling** — when any `set_sensitive` entry is present, the whole merged values map is marked secret in Pulumi state. This is coarser than Terraform's per-entry masking but errs on the safe side; a documented behavioral difference, not a parity exception — the release Helm installs is identical

> **PARITY-EXCEPTION**: `take_ownership` (Helm `--take-ownership`, the migration knob for adopting resources an earlier tool created) is not expressible at the module's pinned pulumi-kubernetes SDK — `helm.v3.ReleaseArgs` gained the flag only in a later release. A set field must never be silently dropped, so the module **fails the deploy loudly** when `take_ownership: true`, with an error that routes you to the Terraform provisioner (which honors it) or to dropping the flag. The exception dissolves at the next pulumi-kubernetes SDK upgrade.

## Usage

Deploy with the CLI:

```shell
planton pulumi up --manifest <manifest.yaml>
```

Example manifest:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesHelmRelease
metadata:
  name: podinfo
spec:
  namespace:
    value: podinfo
  create_namespace: true
  repo: https://stefanprodan.github.io/podinfo
  chart: podinfo
  version: 6.9.2
  values_yaml: |
    replicaCount: 2
  set_string:
    image.tag: "6.9.2"
  atomic: true
  cleanup_on_fail: true
```

## Outputs

Exported to `status.outputs`, read from what Helm actually recorded:

| Name | Description |
|------|-------------|
| `namespace` | The namespace the release is installed in |
| `release_name` | The Helm release name (`helm list` NAME column) |
| `version` | The installed chart version |
| `app_version` | The chart's appVersion (the packaged application's upstream version) |
| `status` | The release status as Helm records it (e.g. `deployed`) |
| `revision` | The release revision number (1 on install, incremented by upgrades/rollbacks) |

> **Note**: Chart values pass through to the chart unvalidated — the chart's values surface is the contract, and a typo'd value surfaces when the chart renders or its resources misbehave, not at preview. `version` is required by the spec: unpinned installs are not reproducible.

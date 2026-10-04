# Kubernetes Manifest - Terraform Module

## Overview

This Terraform module applies raw multi-document Kubernetes YAML — the KubernetesManifest kind's escape-hatch contract — creating one `kubectl_manifest` resource per document, with an optional anchor namespace created first. The manifest content is applied exactly as written: no injected labels, no rewritten fields; the only defaulting is the anchor namespace, and only for documents that declare none.

## Architecture

```
iac/tf/
├── provider.tf     # Terraform provider requirements: alekc/kubectl + hashicorp/kubernetes
├── variables.tf    # Input variables mirroring spec.proto
├── locals.tf       # Labels, document splitting, identity keys, applied-resource inventory
├── main.tf         # Optional namespace + one kubectl_manifest per document
├── outputs.tf      # Exports namespace, applied_resources
└── README.md       # This file
```

## How It Works

1. **Variable Input**: the `spec` variable mirrors the protobuf schema — `namespace` (the anchor, flattened to a plain string by the generator), `create_namespace`, `manifest_yaml`, `skip_await`
2. **Namespace**: when `create_namespace` is `true`, a `kubernetes_namespace_v1` resource creates the anchor namespace with the standard Planton identity labels — labels go on the namespace object only, never into the manifest's documents. Every document `depends_on` the namespace
3. **Document splitting**: `locals.tf` splits `manifest_yaml` on lines starting with `---` (a newline is prepended so a manifest that starts with `---` does not lose its first document), drops blank and comment-only chunks, and fails the plan loudly on an invalid document — never silently skipping it. The Pulumi module's inventory parser uses the identical rule
4. **Identity-keyed `for_each`**: each document becomes a `kubectl_manifest` keyed by its full identity (`apiVersion/Kind/namespace/name`), not its position — reordering documents in the manifest never re-keys Terraform state addresses. Two documents with the same identity collide at plan time, because a duplicate document is a manifest bug, not something to apply twice
5. **Namespace anchoring**: `override_namespace` is set only on documents that declare no `metadata.namespace`; documents with an explicit namespace pass through untouched. Cluster-scoped documents without a namespace also receive the override, but the API server ignores `metadata.namespace` on cluster-scoped objects — the outcome matches the Pulumi provider's scope-aware defaulting (which skips them client-side instead)
6. **Apply**: `server_side_apply = true` (the same apply mechanism as the Pulumi provider) with `force_conflicts`. `kubectl_manifest` needs no cluster connection at plan time, so offline plans work and a CRD plus its custom resources apply in one pass
7. **Await**: `wait` and `wait_for_rollout` are both the inverse of `spec.skip_await` — by default the module blocks until workload rollouts complete on apply and until each document is actually gone (foreground propagation) on destroy, matching the Pulumi module's `SkipAwait`
8. **Outputs**: the anchor namespace and the applied-resource inventory (`apiVersion/Kind/name` per document, manifest order), derived from the input YAML so both engines export an identical list

## Semantics Preserved by the Module

- **The manifest is never mutated** — no label injection, no field rewriting; only the bounded namespace default described above
- **Namespace anchoring is scope-safe** — explicit namespaces always win; cluster-scoped documents are never distorted
- **The inventory comes from the input, not from state** — `applied_resources` is parsed from `manifest_yaml` with the same split rule as the Pulumi module, so the two engines export identical values by construction

> **Behavioral note (not a parity exception)**: await BREADTH differs benignly between engines — when awaiting, the Pulumi engine also readiness-checks non-workload kinds like Services, while kubectl awaits workload rollouts only. The applied objects are identical either way.

## Usage

```hcl
module "manifest" {
  source = "./iac/tf"

  metadata = {
    name = "app-config"
  }

  spec = {
    namespace        = "my-app"
    create_namespace = true

    manifest_yaml = <<-EOT
      apiVersion: v1
      kind: ConfigMap
      metadata:
        name: app-settings
      data:
        LOG_LEVEL: info
      ---
      apiVersion: v1
      kind: Secret
      metadata:
        name: app-credentials
      type: Opaque
      stringData:
        api-key: replace-me
    EOT
  }
}
```

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| `metadata` | Resource metadata (name, org, env) | object | yes |
| `spec` | KubernetesManifest specification | object | yes |

## Outputs

| Name | Description |
|------|-------------|
| `namespace` | The anchor namespace: where namespaced documents without an explicit `metadata.namespace` were applied |
| `applied_resources` | One `apiVersion/Kind/name` entry per document in manifest order, derived from the input YAML |

> **Note**: this module uses the community `alekc/kubectl` provider for the documents (no plan-time cluster connection, CRD+CR in one apply) and the `hashicorp/kubernetes` provider only for the optional anchor-namespace creation. Both are configured by the calling workspace's kubeconfig environment contract.

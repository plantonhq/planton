# Kubernetes Manifest - Pulumi Module

## Overview

This Pulumi module applies raw Kubernetes YAML — the KubernetesManifest kind's escape-hatch contract — through the Kubernetes provider's `yaml/v2` ConfigGroup, with an optional anchor namespace created first. The manifest content is applied exactly as written: no injected labels, no rewritten fields; the only defaulting is the anchor namespace, and only for namespaced documents that declare none.

## Architecture

```
iac/pulumi/
├── main.go              # Entrypoint: loads IaC input, calls module
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Make targets for preview/up/down/refresh
└── module/
    ├── main.go          # Orchestrator: provider init, namespace, ConfigGroup, output export
    ├── locals.go        # Derived values: labels, anchor namespace, applied-resource inventory
    ├── inventory.go     # Parses manifest_yaml into the applied-resource inventory
    ├── namespace.go     # Conditionally creates the anchor namespace
    └── outputs.go       # Exports namespace, applied_resources
```

## How It Works

1. **IaC Input Loading**: the entrypoint loads `KubernetesManifestIacInput` from Pulumi config
2. **Locals Initialization**: `locals.go` resolves the anchor namespace from the spec's value-or-ref, computes the standard Planton identity labels (stamped on the created namespace only — never injected into the manifest's documents), and parses the applied-resource inventory from `manifest_yaml`
3. **Provider Creation**: the Kubernetes provider is initialized from `provider_config` **with `spec.namespace` as its default namespace** — this is the anchoring mechanism. The provider resolves each kind's scope before defaulting, so namespaced documents without an explicit `metadata.namespace` land in the anchor, documents with one keep it, and cluster-scoped documents are never touched
4. **Namespace**: when `create_namespace` is `true`, the anchor namespace is created with the identity labels, and the ConfigGroup depends on it
5. **Manifest Application**: the whole manifest goes to a `yaml/v2` ConfigGroup, which splits multi-document YAML and **orders CRDs before the custom resources that use them** — a CRD and its custom resources apply in one pass. `SkipAwait` is wired to `spec.skip_await`; by default the deploy blocks until readiness (workload rollouts complete, other kinds pass the provider's readiness checks)
6. **Output Export**: the anchor namespace and the applied-resource inventory (`apiVersion/Kind/name` per document, manifest order) are exported

## Semantics Preserved by the Module

- **The manifest is never mutated** — no label injection, no field rewriting; only the provider-level namespace default described above
- **Namespace anchoring is scope-aware** — explicit namespaces always win; cluster-scoped documents (CRDs, ClusterRoles, ...) are applied as-is
- **The inventory comes from the input, not from engine state** — `applied_resources` is parsed from `manifest_yaml` with the same document-split rule as the Terraform module (split on lines starting with `---`, newline prepended so a leading `---` loses nothing, blank and comment-only chunks dropped, invalid documents rejected loudly), so both engines export identical values by construction

The Terraform module reaches the same anchoring outcome with a per-document `override_namespace` and the same await default with `wait`/`wait_for_rollout`. One benign breadth difference: when awaiting, this engine also readiness-checks non-workload kinds like Services, while the Terraform engine's kubectl await covers workload rollouts only — the applied objects are identical either way.

## Field Mapping

| Spec Field | Module Behavior |
|------------|-----------------|
| `namespace` | Provider default namespace (the anchor); exported as the `namespace` output |
| `create_namespace` | Conditionally creates the anchor namespace with Planton identity labels; documents depend on it |
| `manifest_yaml` | Applied verbatim through `yaml/v2` ConfigGroup; also parsed into `applied_resources` |
| `skip_await` | `SkipAwait` on the ConfigGroup; default is to await readiness |

## Usage

Wrap your raw YAML in a KubernetesManifest resource (see `../../e2e/manifest.yaml` for a full-surface example) and deploy through the CLI:

```bash
planton apply -f manifest.yaml
```

## Debug

```bash
# Download deps, vet, and format
make build

# Build the module
go build ./module/...

# Build the entrypoint
go build .
```

> **Note**: reach for KubernetesManifest only when no first-class catalog kind covers what you need to apply — typed kinds validate configuration before deploy and export composable outputs. This module deliberately validates nothing about the manifest's content beyond YAML well-formedness; the API server is what judges the kinds, exactly as with `kubectl apply`.

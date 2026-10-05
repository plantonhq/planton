# Kubernetes Priority Class - Pulumi Module

## Overview

This Pulumi module creates and manages a Kubernetes `scheduling.k8s.io/v1` PriorityClass. It supports the complete PriorityClass surface: the priority value, the global-default flag, the human description, and the preemption policy.

## Architecture

```
iac/pulumi/
├── main.go              # Entrypoint: loads IaC input, calls module
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Make targets for preview/up/down/refresh
└── module/
    ├── main.go          # Orchestrator: provider init, resource creation, output export
    ├── locals.go        # Derived values: labels, annotations, resolved preemption policy
    ├── priorityclass.go # Creates kubernetes.scheduling.v1.PriorityClass resource
    └── outputs.go       # Exports priority_class_name, value
```

## How It Works

1. **IaC Input Loading**: The entrypoint loads `KubernetesPriorityClassIacInput` from Pulumi config
2. **Locals Initialization**: `locals.go` computes:
   - Standard Planton labels merged with user labels (identity keys cannot be overridden)
   - User annotations
   - The resolved preemption policy: the API string for the explicit value, or the server default `PreemptLowerPriority` when the spec omits it
3. **Provider Creation**: Kubernetes provider is initialized from `provider_config`
4. **PriorityClass Creation**: A single `scheduling.k8s.io/v1` PriorityClass is created with the value, global-default flag, description, and preemption policy — with `DeleteBeforeReplace` set
5. **Output Export**: Class name and value are exported as outputs

## Semantics Preserved by the Module

- **`preemptionPolicy` is always sent explicitly** — with the API server's default (`PreemptLowerPriority`) applied module-side, so the Pulumi and Terraform modules submit byte-identical objects for the same manifest, and the deployed object never depends on which engine applied it
- **`globalDefault` is always sent explicitly** — the Kubernetes default is `false`; sending it keeps both engines' submitted objects identical
- **Value changes replace, delete-first** — the priority value is immutable upstream, so a change forces replacement; `DeleteBeforeReplace` avoids the name collision that a create-first replacement would cause (PriorityClass names are cluster-unique)
- **PriorityClasses are cluster-scoped** — no namespace is involved anywhere in the module

## Field Mapping

| Spec Field | PriorityClass Field | Notes |
|------------|--------------------|-------|
| `name` | `metadata.name` | The value pods reference in `priority_class_name` |
| `value` | `value` | Immutable; change forces delete-before-replace |
| `global_default` | `globalDefault` | Always sent explicitly (default `false`) |
| `description` | `description` | Surfaced by `kubectl describe priorityclass` |
| `preemption_policy` | `preemptionPolicy` | `preempt_lower_priority` → `PreemptLowerPriority`, `never` → `Never`; default applied when omitted |

## Usage

```bash
# Preview changes
make preview manifest=../../e2e/manifest.yaml

# Deploy
make up manifest=../../e2e/manifest.yaml

# Destroy
make down manifest=../../e2e/manifest.yaml
```

## Debug

```bash
# Build the module
go build ./module/...

# Build the entrypoint
go build .
```

> **Note**: `global_default: true` affects every pod in the cluster that names no priority class. Only one class should carry the flag — when several do, Kubernetes uses the smallest such value.

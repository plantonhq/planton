# Kubernetes Resource Quota - Pulumi Module

## Overview

This Pulumi module creates and manages a namespace-governance pair: a `core/v1` **ResourceQuota** carrying aggregate caps, scopes, and a scope selector, plus an optional companion `core/v1` **LimitRange** created when the spec sets `limit_defaults`. It supports the complete ResourceQuotaSpec and LimitRangeSpec surfaces: compute, storage, and object-count caps; all seven quota scopes; scope selectors with `In`/`NotIn`/`Exists`/`DoesNotExist`; and container, pod, and persistent-volume-claim limit items with defaults, bounds, and burst ratios.

## Architecture

```
iac/pulumi/
├── main.go              # Entrypoint: loads IaC input, calls module
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Make targets for preview/up/down/refresh
└── module/
    ├── main.go          # Orchestrator: provider init, quota + conditional LimitRange creation, output export
    ├── locals.go        # Derived values: labels, annotations, namespace default, LimitRange name, enum→API-string maps
    ├── resourcequota.go # Creates core/v1 ResourceQuota and the companion core/v1 LimitRange
    └── outputs.go       # Exports resource_quota_name, namespace, limit_range_name
```

## How It Works

1. **IaC Input Loading**: The entrypoint loads `KubernetesResourceQuotaIacInput` from Pulumi config
2. **Locals Initialization**: `locals.go` computes:
   - Standard Planton labels merged with user labels (identity keys cannot be overridden)
   - User annotations
   - The target namespace (foreign-key references are pre-resolved; falls back to `default` when omitted)
   - The companion LimitRange's name — the quota's own name when `limit_defaults` is set, empty otherwise
3. **Provider Creation**: Kubernetes provider is initialized from `provider_config`
4. **ResourceQuota Creation**: A `core/v1` ResourceQuota is created with the `hard` caps, scopes (mapped to API strings), and scope selector
5. **LimitRange Creation (conditional)**: When `limit_defaults` is non-empty, a `core/v1` LimitRange is created with one item per entry, mapping type and all quantity maps one-to-one
6. **Output Export**: Quota name, namespace, and the LimitRange name (empty when none exists) are exported as outputs

## The Two-Object Creation

This module manages **two Kubernetes resources as one governance pair**:

- The **ResourceQuota** is always created
- The **LimitRange** is created exactly when the spec carries `limit_defaults`, and it **shares the quota's name and namespace** — one governance pair, one identity
- The pairing is what keeps a compute quota livable: once a quota caps `requests.*`/`limits.*`, the API rejects pods that omit them; the LimitRange's container defaults let naive workloads inherit values instead of being rejected
- Both objects receive the same merged labels and annotations

> For simple T-shirt-size governance on namespaces Planton creates, the `KubernetesNamespace` kind's resource profiles (which manage a quota and limit range internally) are the simpler alternative. This module is the full-fidelity instrument.

## Field Mapping

| Spec Field | Kubernetes Field | Notes |
|------------|------------------|-------|
| `hard` | ResourceQuota `spec.hard` | Passed through as resource name → quantity |
| `scopes` | ResourceQuota `spec.scopes` | Enum values mapped to API strings (`best_effort` → `BestEffort`, ...) |
| `scope_selector[]` | ResourceQuota `spec.scopeSelector.matchExpressions[]` | Scope name mapped to API string; values sent only when non-empty |
| `limit_defaults[].type` | LimitRange `spec.limits[].type` | `container`/`pod`/`persistent_volume_claim` → `Container`/`Pod`/`PersistentVolumeClaim` |
| `limit_defaults[].default_limit` | LimitRange `spec.limits[].default` | The upstream field is named `default`; the spec names it `default_limit` for clarity |
| `limit_defaults[].default_request` | LimitRange `spec.limits[].defaultRequest` | |
| `limit_defaults[].max/min/max_limit_request_ratio` | LimitRange `spec.limits[].max/min/maxLimitRequestRatio` | Sent only when non-empty |

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

> **Note**: Quota is enforced at admission time — objects created before the quota existed count against usage but are never evicted, and a compute-capping quota without `limit_defaults` rejects pods that omit requests/limits.

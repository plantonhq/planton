# Kubernetes Pod Disruption Budget - Pulumi Module

## Overview

This Pulumi module creates and manages a Kubernetes `policy/v1` PodDisruptionBudget. It supports the complete PodDisruptionBudget surface: exact-match and set-based label selectors, absolute and percentage availability bounds (`min_available` / `max_unavailable`), and the unhealthy-pod eviction policy.

## Architecture

```
iac/pulumi/
├── main.go                  # Entrypoint: loads IaC input, calls module
├── Pulumi.yaml              # Pulumi project configuration
├── Makefile                 # Make targets for preview/up/down/refresh
└── module/
    ├── main.go              # Orchestrator: provider init, resource creation, output export
    ├── locals.go            # Derived values: labels, annotations, namespace default, resolved unhealthy-pod policy
    ├── poddisruptionbudget.go # Creates kubernetes.policy.v1.PodDisruptionBudget resource
    └── outputs.go           # Exports pod_disruption_budget_name, namespace
```

## How It Works

1. **IaC Input Loading**: The entrypoint loads `KubernetesPodDisruptionBudgetIacInput` from Pulumi config
2. **Locals Initialization**: `locals.go` computes:
   - Standard Planton labels merged with user labels (identity keys cannot be overridden)
   - User annotations
   - The target namespace (foreign-key references are pre-resolved; falls back to `default` when omitted)
   - The resolved unhealthy-pod eviction policy: the API string for the spec's enum, with the server default (`IfHealthyBudget`) applied when the spec omits it
3. **Provider Creation**: Kubernetes provider is initialized from `provider_config`
4. **PodDisruptionBudget Creation**: A single `policy/v1` PodDisruptionBudget is created with the selector, the chosen availability bound, and the eviction policy
5. **Output Export**: Budget name and namespace are exported as outputs

## Semantics Preserved by the Module

- **The selector is always sent** (it is required in the spec) — an empty selector block is the "all pods in the namespace" wire form, deliberately explicit, because a `policy/v1` budget with a NULL selector matches no pods at all
- **`unhealthyPodEvictionPolicy` is always sent explicitly** — with the server default applied module-side, so the deployed object never depends on server-side defaulting
- **Bounds pass through as IntOrString** — a numeric string (`"2"`) becomes an absolute count, a `%`-suffixed one (`"25%"`) a percentage; the spec enforces exactly one bound is set

## PARITY-EXCEPTION: unhealthy_pod_eviction_policy

The Terraform kubernetes provider's PDB resource cannot express `spec.unhealthyPodEvictionPolicy` at all; the Terraform module fails the plan with a precondition when the spec asks for `always_allow`. **This module is therefore the only engine that deploys `always_allow`.** It sends the field explicitly for every spec (server default `IfHealthyBudget` when omitted), so the deployed object is identical across engines for every spec the Terraform module accepts.

## Field Mapping

| Spec Field | PodDisruptionBudget Field | Notes |
|------------|---------------------------|-------|
| `selector` | `spec.selector` | Always rendered; empty → empty selector (all pods) |
| `min_available` | `spec.minAvailable` | IntOrString; numeric string → count, `%` → percentage |
| `max_unavailable` | `spec.maxUnavailable` | IntOrString; alternative to `min_available` |
| `unhealthy_pod_eviction_policy` | `spec.unhealthyPodEvictionPolicy` | `if_healthy_budget` → `IfHealthyBudget`, `always_allow` → `AlwaysAllow`; default applied when omitted |

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

> **Note**: The budget governs only voluntary disruptions (the eviction API). A budget can only refuse evictions — the selected workload must run enough replicas above the declared floor for node drains to proceed.

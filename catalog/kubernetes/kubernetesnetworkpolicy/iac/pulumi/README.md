# Kubernetes Network Policy - Pulumi Module

## Overview

This Pulumi module creates and manages a Kubernetes `networking/v1` NetworkPolicy. It supports the complete NetworkPolicy surface: pod selector, policy types, ingress and egress rules with all three peer forms (pod selector, namespace selector, IP block with exceptions), TCP/UDP/SCTP ports, named ports, and port ranges.

## Architecture

```
iac/pulumi/
├── main.go              # Entrypoint: loads IaC input, calls module
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Make targets for preview/up/down/refresh
└── module/
    ├── main.go          # Orchestrator: provider init, resource creation, output export
    ├── locals.go        # Derived values: labels, annotations, namespace default, resolved policy types
    ├── networkpolicy.go # Creates kubernetes.networking.v1.NetworkPolicy resource
    └── outputs.go       # Exports network_policy_name, namespace, policy_types
```

## How It Works

1. **IaC Input Loading**: The entrypoint loads `KubernetesNetworkPolicyIacInput` from Pulumi config
2. **Locals Initialization**: `locals.go` computes:
   - Standard Planton labels merged with user labels (identity keys cannot be overridden)
   - User annotations
   - The target namespace (foreign-key references are pre-resolved; falls back to `default` when omitted)
   - The resolved policy types: the explicit `policy_types` when set, otherwise the Kubernetes inference rule (ingress always; egress only when egress rules exist)
3. **Provider Creation**: Kubernetes provider is initialized from `provider_config`
4. **NetworkPolicy Creation**: A single `networking/v1` NetworkPolicy is created with the pod selector, policy types, and ingress/egress rules
5. **Output Export**: Policy name, namespace, and the governed directions are exported as outputs

## Semantics Preserved by the Module

- **Policy types are always sent explicitly** — using the resolved set from locals rather than API-server inference, so the Pulumi and Terraform modules submit byte-identical direction sets for the same manifest, and the deployed object never depends on which engine applied it
- **Each spec peer maps to exactly one API peer** — a peer with both `pod_selector` and `namespace_selector` stays one AND'd peer (pods matching the selector in the selected namespaces) and is never split into two OR'd peers
- **A nil or empty pod selector renders as the EMPTY selector** — "all pods in the namespace", the wire form default-deny policies depend on
- **Ports pass through as IntOrString** — a numeric string (`"5432"`) becomes a port number, anything else a named container port; `end_port` expresses a contiguous range anchored at a numeric port

## Field Mapping

| Spec Field | NetworkPolicy Field | Notes |
|------------|--------------------|-------|
| `pod_selector` | `spec.podSelector` | Empty/absent → empty selector (all pods) |
| `policy_types` | `spec.policyTypes` | Enum values mapped to `"Ingress"`/`"Egress"`; inferred when omitted |
| `ingress_rules[].from[]` / `egress_rules[].to[]` | `spec.ingress[].from[]` / `spec.egress[].to[]` | One-to-one peer mapping |
| `ports[].protocol` | `ports[].protocol` | Unspecified → `TCP` |
| `ports[].port` / `ports[].end_port` | `ports[].port` / `ports[].endPort` | Empty port → all ports; `end_port` 0 → single port |

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

> **Note**: The created object is only enforced by a NetworkPolicy-implementing CNI (Calico, Cilium, cloud CNIs with enforcement enabled). On clusters whose CNI ignores NetworkPolicies, the deployment succeeds but traffic is unaffected.

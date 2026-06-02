# KubernetesPeerAuthentication Pulumi Module Architecture

## Data Flow

```
StackInput (spec + cluster credentials; foreign keys already resolved)
   ↓
Locals (peer_authentication_name = metadata.name, namespace, labels)
   ↓
Resources (namespaced PeerAuthentication CR via typed crd2pulumi SDK)
   ↓
Outputs (peer_authentication_name, namespace)
```

## File Organization

| File | Responsibility |
|------|---------------|
| `main.go` | Orchestrates resource creation: assembles `istiosecurityv1.PeerAuthenticationSpecArgs` (mtls, selector, port_level_mtls) and exports outputs |
| `locals.go` | Extracts values from spec/metadata: resolves the `namespace` foreign key via `GetValue()`, computes standard labels |
| `outputs.go` | Defines output constant names matching `stack_outputs.proto` |

## Key Design Decisions

- **Namespaced resource** -- the module sets `metadata.namespace` from the resolved `spec.namespace` foreign key.
- **Foreign keys are pre-resolved** -- `spec.namespace` is a `StringValueOrRef`. The platform resolves `valueFrom` references to literals before the module runs, so the module simply reads `GetValue()`.
- **Plain selector** -- `selector.matchLabels` is matched at runtime by istiod against pod labels; it is not an OpenMCF foreign key and creates no DAG edge. InfraChart DAG ordering (via `metadata.relationships`) sequences the protected workload before the policy.
- **Typed crd2pulumi resource** -- uses `istiosecurityv1.NewPeerAuthentication` rather than an untyped `CustomResource`, so field-name and structure errors are caught at compile time. This matches every other OpenMCF Istio component.
- **Optional fields are conditionally set** -- `mtls`, `selector`, and `port_level_mtls` are only populated when present, so omitting them lets the policy inherit from its parent (namespace, then mesh) exactly as upstream intends.
- **Port-level mTLS key conversion** -- the proto keys `port_level_mtls` by `uint32`, but JSON/CRD map keys are strings, so the SDK models the field as `pulumi.StringMapMap` (`{"8080": {"mode": "STRICT"}}`). The module converts each `uint32` port to its decimal string form.

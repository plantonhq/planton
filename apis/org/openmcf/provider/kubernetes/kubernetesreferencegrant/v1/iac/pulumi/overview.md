# KubernetesReferenceGrant Pulumi Module Architecture

## Data Flow

```
StackInput (spec + cluster credentials; foreign keys already resolved)
   ↓
Locals (reference_grant_name = metadata.name, namespace, labels)
   ↓
Resources (namespaced ReferenceGrant CR via typed crd2pulumi SDK)
   ↓
Outputs (reference_grant_name, namespace)
```

## File Organization

| File | Responsibility |
|------|---------------|
| `main.go` | Orchestrates resource creation: assembles `gatewayv1.ReferenceGrantSpecArgs` (from / to) and exports outputs |
| `locals.go` | Extracts values from spec/metadata: resolves the `namespace` foreign key via `GetValue()`, computes standard labels |
| `outputs.go` | Defines output constant names matching `stack_outputs.proto` |
| `references.go` | Maps the `from` (trusted sources) and `to` (referenceable targets) lists |

## Key Design Decisions

- **Namespaced resource** -- the module sets `metadata.namespace` from the resolved `spec.namespace` foreign key. This is the "to" namespace: the grant lives alongside the resources it authorizes inbound references to.
- **Foreign keys are pre-resolved** -- `spec.namespace` is a `StringValueOrRef`. The platform resolves `valueFrom` references to literals before the module runs, so the module simply reads `GetValue()`.
- **No status subresource** -- upstream ReferenceGrant has no status, so the module exports only the identifying coordinates; there is nothing controller-managed to surface.
- **Typed crd2pulumi resource** -- uses `gatewayv1.NewReferenceGrant` rather than an untyped `CustomResource`, so field-name and structure errors are caught at compile time. The constructor sets the `apiVersion`/`kind` and registers a `v1beta1` alias, so the module never hard-codes the served version. This matches every other OpenMCF ingress component.
- **from / to are trust assertions, not foreign keys (DD-009)** -- the entries describe KINDS of resources permitted to reference (and be referenced), not specific OpenMCF resource instances. The one genuine cross-resource reference is `from[].namespace` (a source namespace); when it is OpenMCF-managed, infra-chart authors express that edge via `metadata.relationships` (`type: uses`). The grant itself is a low-dependency leaf the consuming Gateway/Route orders itself after.
- **Optional fields are conditionally set** -- `group` is only written when non-empty (empty = core API group), and `to[].name` only when present (absence = all resources of the group/kind); upstream/controller defaults are never baked into the module.
- **No matches, filters, parent or backend refs** -- ReferenceGrant has none of the route machinery; the module is intentionally minimal (just the two from/to lists).

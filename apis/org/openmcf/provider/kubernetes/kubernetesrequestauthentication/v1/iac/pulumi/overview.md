# KubernetesRequestAuthentication Pulumi Module Architecture

## Data Flow

```
StackInput (spec + cluster credentials; foreign keys already resolved)
   ↓
Locals (request_authentication_name = metadata.name, namespace, labels)
   ↓
Resources (namespaced RequestAuthentication CR via typed crd2pulumi SDK)
   ↓
Outputs (request_authentication_name, namespace)
```

## File Organization

| File | Responsibility |
|------|---------------|
| `main.go` | Orchestrates resource creation: assembles `istiosecurityv1.RequestAuthenticationSpecArgs` (jwt_rules, selector, target_refs) and exports outputs |
| `locals.go` | Extracts values from spec/metadata: resolves the `namespace` foreign key via `GetValue()`, computes standard labels |
| `outputs.go` | Defines output constant names matching `stack_outputs.proto` |

## Key Design Decisions

- **Namespaced resource** -- the module sets `metadata.namespace` from the resolved `spec.namespace` foreign key.
- **Foreign keys are pre-resolved** -- `spec.namespace` is a `StringValueOrRef`. The platform resolves `valueFrom` references to literals before the module runs, so the module simply reads `GetValue()`.
- **Plain selector / target_refs** -- `selector.matchLabels` and `target_refs` are resolved at runtime by istiod; they are not OpenMCF foreign keys and create no DAG edge. InfraChart DAG ordering (via `metadata.relationships`) sequences the targeted resources before the policy.
- **Typed crd2pulumi resource** -- uses `istiosecurityv1.NewRequestAuthentication` rather than an untyped `CustomResource`, so field-name and structure errors are caught at compile time. This matches every other OpenMCF Istio component.
- **Spec is assigned by value** -- `RequestAuthenticationArgs.Spec` is a `PtrInput` satisfied by the `RequestAuthenticationSpecArgs` value itself; the `*Ptr()` wrapper marshals to the wrong element type and panics at apply. Nested `Selector` / `TargetRefs` / `JwtRules` follow the same value-assignment rule.
- **Optional fields are conditionally set** -- `jwt_rules`, `selector`, `target_refs`, and each JWT rule's optional scalars (read via the proto3 `optional` pointer) are only populated when present, so omitting them lets upstream defaults apply.

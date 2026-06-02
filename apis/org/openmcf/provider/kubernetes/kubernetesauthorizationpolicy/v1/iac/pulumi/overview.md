# KubernetesAuthorizationPolicy Pulumi Module Architecture

## Data Flow

```
StackInput (spec + cluster credentials; foreign keys already resolved)
   ↓
Locals (authorization_policy_name = metadata.name, namespace, labels)
   ↓
Resources (namespaced AuthorizationPolicy CR via typed crd2pulumi SDK)
   ↓
Outputs (authorization_policy_name, namespace)
```

## File Organization

| File | Responsibility |
|------|---------------|
| `main.go` | Orchestrates resource creation: assembles `istiosecurityv1.AuthorizationPolicySpecArgs` (selector, target_refs, rules, action, provider) and exports outputs |
| `locals.go` | Extracts values from spec/metadata: resolves the `namespace` foreign key via `GetValue()`, computes standard labels |
| `outputs.go` | Defines output constant names matching `stack_outputs.proto` |

## Key Design Decisions

- **Namespaced resource** -- the module sets `metadata.namespace` from the resolved `spec.namespace` foreign key.
- **Foreign keys are pre-resolved** -- `spec.namespace` is a `StringValueOrRef`. The platform resolves `valueFrom` references to literals before the module runs, so the module simply reads `GetValue()`.
- **Plain selector / target_refs** -- `selector.matchLabels` and `target_refs` are resolved at runtime by istiod; they are not OpenMCF foreign keys and create no DAG edge. InfraChart DAG ordering (via `metadata.relationships`) sequences the targeted resources before the policy.
- **Typed crd2pulumi resource** -- uses `istiosecurityv1.NewAuthorizationPolicy` rather than an untyped `CustomResource`, so field-name and structure errors are caught at compile time. This matches every other OpenMCF Istio component.
- **Spec is assigned by value** -- `AuthorizationPolicyArgs.Spec` is a `PtrInput` satisfied by the `AuthorizationPolicySpecArgs` value itself; the `*Ptr()` wrapper marshals to the wrong element type and panics at apply (a bug the PeerAuthentication forge caught live). Nested `Selector`, `TargetRefs`, `Rules` (with `From.Source` / `To.Operation` / `When`), and `Provider` follow the same value-assignment rule.
- **The from/to wrappers are preserved** -- the CR shape is `from: [{ source: {...} }]` and `to: [{ operation: {...} }]`; the module builds `RulesFromArgs{Source: ...}` / `RulesToArgs{Operation: ...}` exactly, matching the typed SDK and the upstream CRD JSON.
- **Optional fields are conditionally set** -- `selector`, `target_refs`, `rules`, `action` (read via the proto3 `optional` pointer), `provider`, and every source/operation match list are only populated when present, so omitting them lets upstream defaults apply (e.g. an absent `action` becomes ALLOW).

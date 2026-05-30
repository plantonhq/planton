# KubernetesHttpRoute Pulumi Module Architecture

## Data Flow

```
StackInput (spec + cluster credentials; foreign keys already resolved)
   ↓
Locals (route_name = metadata.name, namespace, labels)
   ↓
Resources (namespaced HTTPRoute CR via typed crd2pulumi SDK)
   ↓
Outputs (route_name, namespace)
```

## File Organization

| File | Responsibility |
|------|---------------|
| `main.go` | Orchestrates resource creation: assembles `gatewayv1.HTTPRouteSpecArgs` from the builder functions and exports outputs |
| `locals.go` | Extracts values from spec/metadata: resolves the `namespace` foreign key via `GetValue()`, computes standard labels |
| `outputs.go` | Defines output constant names matching `stack_outputs.proto` |
| `parent_refs.go` | Maps the parent references (Gateways the route attaches to) |
| `rules.go` | Maps route rules and per-rule timeouts; delegates matches, filters, and backend refs |
| `matches.go` | Maps path, header, query-param, and method matchers |
| `filters.go` | Maps rule-level filters (header modify, redirect, rewrite, mirror, CORS, extension ref) |
| `backend_refs.go` | Maps backend refs and the parallel backend-ref-level filter tree |

## Key Design Decisions

- **Namespaced resource** -- the module sets `metadata.namespace` from the resolved `spec.namespace` foreign key.
- **Foreign keys are pre-resolved** -- `spec.namespace` is a `StringValueOrRef`. The platform resolves `valueFrom` references to literals before the module runs, so the module simply reads `GetValue()`.
- **Plain parent and backend references** -- `parentRefs` and `backendRefs` are upstream multi-field objects, not OpenMCF foreign keys; InfraChart DAG ordering sequences the Gateway and backends before the route.
- **Typed crd2pulumi resource** -- uses `gatewayv1.NewHTTPRoute` rather than an untyped `CustomResource`, so field-name and structure errors are caught at compile time. This matches every other OpenMCF ingress component.
- **Optional fields are conditionally set** -- hostnames, matches, filters, backend refs, and timeouts are only populated when present; upstream/controller defaults are never baked into the module.
- **Two filter type trees** -- crd2pulumi generates structurally identical but distinct Go types for rule-level filters (`HTTPRouteSpecRulesFilters*`) and backend-ref-level filters (`HTTPRouteSpecRulesBackendRefsFilters*`). A single proto `KubernetesHttpRouteFilter` is therefore mapped by two parallel builder sets (`filters.go` and `backend_refs.go`), the same pattern the Gateway component uses for its two namespace-selector types.
- **Split mapping files** -- the upstream `HTTPRouteSpec` is the largest in the family; mapping is split by concern so each file stays focused and reviewable.

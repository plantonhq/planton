# KubernetesGrpcRoute Pulumi Module Architecture

## Data Flow

```
StackInput (spec + cluster credentials; foreign keys already resolved)
   ↓
Locals (route_name = metadata.name, namespace, labels)
   ↓
Resources (namespaced GRPCRoute CR via typed crd2pulumi SDK)
   ↓
Outputs (route_name, namespace)
```

## File Organization

| File | Responsibility |
|------|---------------|
| `main.go` | Orchestrates resource creation: assembles `gatewayv1.GRPCRouteSpecArgs` from the builder functions and exports outputs |
| `locals.go` | Extracts values from spec/metadata: resolves the `namespace` foreign key via `GetValue()`, computes standard labels |
| `outputs.go` | Defines output constant names matching `stack_outputs.proto` |
| `parent_refs.go` | Maps the parent references (Gateways the route attaches to) |
| `rules.go` | Maps route rules; delegates matches, filters, and backend refs (GRPCRouteRule has no timeouts) |
| `matches.go` | Maps the method (service/method) matcher and header matchers |
| `filters.go` | Maps rule-level filters (request/response header modify, request mirror, extension ref) |
| `backend_refs.go` | Maps backend refs and the parallel backend-ref-level filter tree |

## Key Design Decisions

- **Namespaced resource** -- the module sets `metadata.namespace` from the resolved `spec.namespace` foreign key.
- **Foreign keys are pre-resolved** -- `spec.namespace` is a `StringValueOrRef`. The platform resolves `valueFrom` references to literals before the module runs, so the module simply reads `GetValue()`.
- **Plain parent and backend references (DD-009)** -- `parentRefs` and `backendRefs` are upstream multi-field objects, not OpenMCF foreign keys. InfraChart DAG ordering sequences the Gateway and backends before the route, and the cross-resource edges are expressed by infra-chart authors via `metadata.relationships` (`depends_on` Gateway, `uses` Service).
- **Typed crd2pulumi resource** -- uses `gatewayv1.NewGRPCRoute` rather than an untyped `CustomResource`, so field-name and structure errors are caught at compile time. This matches every other OpenMCF ingress component.
- **Optional fields are conditionally set** -- hostnames, matches, filters, and backend refs are only populated when present; upstream/controller defaults are never baked into the module.
- **Two filter type trees** -- crd2pulumi generates structurally identical but distinct Go types for rule-level filters (`GRPCRouteSpecRulesFilters*`) and backend-ref-level filters (`GRPCRouteSpecRulesBackendRefsFilters*`). A single proto `KubernetesGrpcRouteFilter` is therefore mapped by two parallel builder sets (`filters.go` and `backend_refs.go`), the same pattern HTTPRoute and the Gateway component use.
- **Smaller filter set than HTTPRoute** -- GRPCRoute supports only header modifiers, request mirror, and extension ref (no redirect, URL rewrite, or CORS), reflecting the upstream GRPCRouteFilter union.

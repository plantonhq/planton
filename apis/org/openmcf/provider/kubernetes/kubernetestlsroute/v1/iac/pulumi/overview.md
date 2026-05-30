# KubernetesTlsRoute Pulumi Module Architecture

## Data Flow

```
StackInput (spec + cluster credentials; foreign keys already resolved)
   ↓
Locals (route_name = metadata.name, namespace, labels)
   ↓
Resources (namespaced TLSRoute CR via typed crd2pulumi SDK)
   ↓
Outputs (route_name, namespace)
```

## File Organization

| File | Responsibility |
|------|---------------|
| `main.go` | Orchestrates resource creation: assembles `gatewayv1.TLSRouteSpecArgs` (hostnames, parent refs, rules) and exports outputs |
| `locals.go` | Extracts values from spec/metadata: resolves the `namespace` foreign key via `GetValue()`, computes standard labels |
| `outputs.go` | Defines output constant names matching `stack_outputs.proto` |
| `parent_refs.go` | Maps the parent references (Gateways the route attaches to) |
| `rules.go` | Maps route rules and their backend refs (a TLSRoute rule has no matches or filters) |

## Key Design Decisions

- **Namespaced resource** -- the module sets `metadata.namespace` from the resolved `spec.namespace` foreign key.
- **Foreign keys are pre-resolved** -- `spec.namespace` is a `StringValueOrRef`. The platform resolves `valueFrom` references to literals before the module runs, so the module simply reads `GetValue()`.
- **Plain parent and backend references (DD-009)** -- `parentRefs` and `backendRefs` are upstream multi-field objects, not OpenMCF foreign keys. InfraChart DAG ordering sequences the Gateway and backends before the route, and the cross-resource edges are expressed by infra-chart authors via `metadata.relationships` (`depends_on` Gateway, `uses` Service).
- **Typed crd2pulumi resource** -- uses `gatewayv1.NewTLSRoute` rather than an untyped `CustomResource`, so field-name and structure errors are caught at compile time. This matches every other OpenMCF ingress component.
- **Shared backend ref** -- because a TLS route has no per-backend filters, the rule mapping consumes the shared `KubernetesGatewayApiBackendRef` directly instead of a per-route backend type.
- **Optional fields are conditionally set** -- parent refs and the per-backend optional fields are only populated when present; upstream/controller defaults are never baked into the module. `hostnames` and `rules` are always set (both required by the spec).
- **No matches or filters** -- a TLS passthrough route routes only by SNI hostname (carried at the spec level), so there is no `matches.go`/`filters.go`; the module is intentionally the smallest of the route family.

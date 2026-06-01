# KubernetesTcpRoute Pulumi Module Architecture

## Data Flow

```
StackInput (spec + cluster credentials; foreign keys already resolved)
   ↓
Locals (route_name = metadata.name, namespace, labels)
   ↓
Resources (namespaced TCPRoute CR via typed crd2pulumi SDK, v1alpha2)
   ↓
Outputs (route_name, namespace)
```

## File Organization

| File | Responsibility |
|------|---------------|
| `main.go` | Orchestrates resource creation: assembles `gatewayv1alpha2.TCPRouteSpecArgs` (parent refs, use_default_gateways, rules) and exports outputs |
| `locals.go` | Extracts values from spec/metadata: resolves the `namespace` foreign key via `GetValue()`, computes standard labels |
| `outputs.go` | Defines output constant names matching `stack_outputs.proto` |
| `parent_refs.go` | Maps the parent references (Gateways the route attaches to) |
| `rules.go` | Maps route rules and their backend refs (a TCPRoute rule has no matches or filters) |

## Key Design Decisions

- **Experimental channel resource** -- TCPRoute is served as `gateway.networking.k8s.io/v1alpha2`, so the module imports the `gateway/v1alpha2` typed package and calls `gatewayv1alpha2.NewTCPRoute`. The experimental CRDs must be installed on the cluster.
- **Namespaced resource** -- the module sets `metadata.namespace` from the resolved `spec.namespace` foreign key.
- **Foreign keys are pre-resolved** -- `spec.namespace` is a `StringValueOrRef`. The platform resolves `valueFrom` references to literals before the module runs, so the module simply reads `GetValue()`.
- **`use_default_gateways` mapped when set** -- the experimental default-Gateway attachment field is passed through only when the user sets it.
- **Plain parent and backend references (DD-009)** -- `parentRefs` and `backendRefs` are upstream multi-field objects, not OpenMCF foreign keys. InfraChart DAG ordering sequences the Gateway and backends before the route, and the cross-resource edges are expressed by infra-chart authors via `metadata.relationships` (`depends_on` Gateway, `uses` Service).
- **Shared backend ref** -- because a TCP route has no per-backend filters, the rule mapping consumes the shared `KubernetesGatewayApiBackendRef` directly instead of a per-route backend type.
- **Optional fields are conditionally set** -- parent refs, use_default_gateways, and the per-backend optional fields are only populated when present; upstream/controller defaults are never baked into the module.
- **No matches or filters** -- a TCP route forwards by listener port with no application-layer visibility, so there is no `matches.go`/`filters.go`.

# KubernetesGateway Pulumi Module Architecture

## Data Flow

```
StackInput (spec + cluster credentials; foreign keys already resolved)
   ↓
Locals (gateway_name = metadata.name, namespace, gateway_class_name, labels)
   ↓
Resources (namespaced Gateway CR via typed crd2pulumi SDK)
   ↓
Outputs (gateway_name, namespace, gateway_class_name)
```

## File Organization

| File | Responsibility |
|------|---------------|
| `main.go` | Orchestrates resource creation: assembles `gatewayv1.GatewaySpecArgs` from the builder functions and exports outputs |
| `locals.go` | Extracts values from spec/metadata: resolves the `namespace` and `gateway_class_name` foreign keys via `GetValue()`, computes standard labels |
| `outputs.go` | Defines output constant names matching `stack_outputs.proto` |
| `listeners.go` | Maps listeners, per-listener TLS, certificate refs, and allowedRoutes |
| `tls.go` | Maps gateway-wide frontend (client-cert validation, per-port overrides) and backend TLS |
| `infrastructure.go` | Maps infrastructure labels/annotations/parametersRef and allowedListeners |
| `addresses.go` | Maps requested Gateway addresses |
| `selectors.go` | Maps the namespace label selector into the two distinct crd2pulumi selector types |

## Key Design Decisions

- **Namespaced resource** -- unlike GatewayClass, a Gateway is namespaced. The module sets `metadata.namespace` from the resolved `spec.namespace` foreign key.
- **Foreign keys are pre-resolved** -- `spec.namespace` and `spec.gateway_class_name` are `StringValueOrRef`. The platform resolves `valueFrom` references to literals before the module runs, so the module simply reads `GetValue()`.
- **Typed crd2pulumi resource** -- uses `gatewayv1.NewGateway` rather than an untyped `CustomResource`, so field-name and structure errors are caught at compile time. This matches every other OpenMCF ingress component.
- **Optional fields are conditionally set** -- listeners' optional fields, addresses, infrastructure, allowedListeners, and gateway-level TLS are only populated when present; upstream/controller defaults are never baked into the module.
- **Two selector builders** -- the Gateway CRD generates structurally identical but distinct Go types for the AllowedRoutes and AllowedListeners namespace selectors. A single proto `KubernetesGatewayLabelSelector` is therefore mapped by two dedicated builders in `selectors.go`.
- **Split mapping files** -- the upstream `GatewaySpec` is large; mapping is split by concern (listeners, tls, infrastructure, addresses, selectors) so each file stays focused and reviewable.

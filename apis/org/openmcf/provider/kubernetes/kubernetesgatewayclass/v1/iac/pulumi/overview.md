# KubernetesGatewayClass Pulumi Module Architecture

## Data Flow

```
StackInput (spec + cluster credentials)
   ↓
Locals (gateway_class_name = metadata.name, controller_name, labels)
   ↓
Resources (cluster-scoped GatewayClass CR via typed crd2pulumi SDK)
   ↓
Outputs (gateway_class_name, controller_name)
```

## File Organization

| File | Responsibility |
|------|---------------|
| `main.go` | Orchestrates resource creation: maps the spec to a typed `gatewayv1.GatewayClassArgs` and exports outputs |
| `locals.go` | Extracts and computes values from spec/metadata: cluster-scoped name, controller name, standard labels |
| `outputs.go` | Defines output constant names matching `stack_outputs.proto` |

## Key Design Decisions

- **Cluster-scoped, no namespace** -- GatewayClass is a cluster-scoped Kubernetes resource, so the module never sets or creates a namespace. Its resource name is the OpenMCF `metadata.name`.
- **Typed crd2pulumi resource** -- Uses `gatewayv1.NewGatewayClass` rather than an untyped `CustomResource`, so field-name and structure errors are caught at compile time. This matches how every other OpenMCF ingress component consumes the Gateway API typed SDK.
- **Optional fields are conditionally set** -- `description` and `parameters_ref` are only populated when present in the spec; upstream/controller defaults are never baked into the module.
- **`parameters_ref.namespace` is set only when provided** -- preserving the upstream rule that namespace must be unset for cluster-scoped parameter resources.

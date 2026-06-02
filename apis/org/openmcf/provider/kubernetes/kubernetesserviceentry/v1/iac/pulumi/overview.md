# KubernetesServiceEntry Pulumi Module Architecture

## Data Flow

```
StackInput (spec + cluster credentials; foreign keys already resolved)
   ↓
Locals (service_entry_name = metadata.name, namespace, labels)
   ↓
Resources (namespaced ServiceEntry CR via typed crd2pulumi SDK)
   ↓
Outputs (service_entry_name, namespace)
```

## File Organization

| File | Responsibility |
|------|---------------|
| `main.go` | Orchestrates resource creation: assembles `istionetworkingv1.ServiceEntrySpecArgs` (hosts, ports, endpoints, workload_selector, ...) and exports outputs |
| `locals.go` | Extracts values from spec/metadata: resolves the `namespace` foreign key via `GetValue()`, computes standard labels |
| `outputs.go` | Defines output constant names matching `stack_outputs.proto` |

## Key Design Decisions

- **Namespaced resource** -- the module sets `metadata.namespace` from the resolved `spec.namespace` foreign key.
- **Foreign keys are pre-resolved** -- `spec.namespace` is a `StringValueOrRef`. The platform resolves `valueFrom` references to literals before the module runs, so the module simply reads `GetValue()`.
- **Plain hosts / selector** -- `hosts`, `addresses`, and `workload_selector.labels` are registry values resolved by istiod at runtime; they are not OpenMCF foreign keys and create no DAG edge. InfraChart DAG ordering (via `metadata.relationships`) sequences any workloads a MESH_INTERNAL entry fronts.
- **Typed crd2pulumi resource** -- uses `istionetworkingv1.NewServiceEntry` rather than an untyped `CustomResource`, so field-name and structure errors are caught at compile time. This matches every other OpenMCF Istio component.
- **Spec is assigned by value** -- `ServiceEntryArgs.Spec` is a `PtrInput` satisfied by the `ServiceEntrySpecArgs` value itself; the `*Ptr()` wrapper marshals to the wrong element type and panics at apply (a bug the PeerAuthentication forge caught live).
- **uint32 -> int / IntMap conversions** -- proto `uint32` port numbers, target ports, and weights are cast to the SDK's `int` inputs, and the endpoint `map<string,uint32>` ports become a `pulumi.IntMap`.
- **Only `hosts` is unconditional** -- it is required upstream; `ports`, `endpoints`, `workload_selector`, the string lists, and the optional `location`/`resolution` scalars (read via the proto3 `optional` pointer) are only populated when present, so omitting them lets upstream defaults apply.

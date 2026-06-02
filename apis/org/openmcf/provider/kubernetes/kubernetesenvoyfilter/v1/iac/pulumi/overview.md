# KubernetesEnvoyFilter Pulumi Module Architecture

## Data Flow

```
StackInput (spec + cluster credentials; foreign keys already resolved)
   ↓
Locals (envoy_filter_name = metadata.name, namespace, labels)
   ↓
Resources (namespaced EnvoyFilter CR via typed crd2pulumi SDK)
   ↓
Outputs (envoy_filter_name, namespace)
```

## File Organization

| File | Responsibility |
|------|---------------|
| `main.go` | Orchestrates resource creation: assembles `istionetworkingv1alpha3.EnvoyFilterSpecArgs` (config_patches, workload_selector, target_refs, priority) via per-nested-message builders, converts the free-form `patch.value` Struct to a Pulumi map, and exports outputs |
| `locals.go` | Extracts values from spec/metadata: resolves the `namespace` foreign key via `GetValue()`, computes standard labels |
| `outputs.go` | Defines output constant names matching `stack_outputs.proto` |

## Key Design Decisions

- **Namespaced resource** -- the module sets `metadata.namespace` from the resolved `spec.namespace` foreign key.
- **Foreign keys are pre-resolved** -- `spec.namespace` is a `StringValueOrRef`. The platform resolves `valueFrom` references to literals before the module runs, so the module simply reads `GetValue()`.
- **Plain selector / target_refs** -- `workload_selector.labels` and `target_refs` are resolved by istiod at runtime; they are not OpenMCF foreign keys and create no DAG edge. InfraChart DAG ordering (via `metadata.relationships`) sequences any gateway/workload the filter patches.
- **Typed crd2pulumi resource** -- uses `istionetworkingv1alpha3.NewEnvoyFilter` (served at `networking/v1alpha3` -- EnvoyFilter has not graduated to v1) rather than an untyped `CustomResource`, so field-name and structure errors are caught at compile time. This matches every other OpenMCF Istio component.
- **Spec is assigned by value** -- `EnvoyFilterArgs.Spec` is a `PtrInput` satisfied by the `EnvoyFilterSpecArgs` value itself; the `*Ptr()` wrapper marshals to the wrong element type and panics at apply (a bug the PeerAuthentication forge caught live).
- **Free-form Struct value** -- `patch.value` is typed `pulumi.MapInput` by the SDK. The module converts the proto `google.protobuf.Struct` to a `pulumi.Map` with a small recursive helper (`structToPulumiMap`: map -> `pulumi.Map`, slice -> `pulumi.Array`, scalars -> typed inputs), preserving arbitrary nesting. The repo has no generic `pulumi.ToMap`, so the recursive converter is the robust path.
- **Everything optional is conditional** -- config_patches, the match branches (at most one of listener/route_configuration/cluster), the patch fields, priority, workload_selector, and target_refs are populated only when present (proto3 `optional`/empty checks), so omitting them lets upstream defaults apply.
- **uint32 -> int conversions** -- proto `uint32` port numbers and destination ports are cast to the SDK's `int` inputs.

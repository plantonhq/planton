# KubernetesDestinationRule Pulumi Module Architecture

## Data Flow

```
StackInput (spec + cluster credentials; foreign keys already resolved)
   ↓
Locals (destination_rule_name = metadata.name, namespace, labels)
   ↓
Resources (namespaced DestinationRule CR via typed crd2pulumi SDK)
   ↓
Outputs (destination_rule_name, namespace)
```

## File Organization

| File | Responsibility |
|------|---------------|
| `main.go` | Orchestrates resource creation: assembles `istionetworkingv1.DestinationRuleSpecArgs` (host, trafficPolicy, subsets, workloadSelector, exportTo) and exports outputs |
| `traffic_policy.go` | Per-path typed builders for the traffic-policy subtree, plus shared scalar helpers |
| `locals.go` | Extracts values from spec/metadata: resolves the `namespace` foreign key via `GetValue()`, computes standard labels |
| `outputs.go` | Defines output constant names matching `stack_outputs.proto` |

## Key Design Decisions

- **Typed crd2pulumi resource** -- uses `istionetworkingv1.NewDestinationRule` rather than an untyped `CustomResource`, so field-name and structure errors are caught at compile time. This matches every other OpenMCF Istio component.
- **Spec is assigned by value** -- `DestinationRuleArgs.Spec` is a `PtrInput` satisfied by the `DestinationRuleSpecArgs` value itself; the `*Ptr()` wrapper marshals to the wrong element type and panics at apply.
- **Unions are mapped by which proto field is set** -- there is no discriminator (see `docs/README.md`). `lb_policy`, `hash_key`, and `hash_algorithm` are read by checking which optional sibling is non-nil.
- **Path-typed builder duplication is intentional** -- crd2pulumi emits a distinct Go type for the `TrafficPolicy` subtree at each of its four reachable paths (spec / subset / their portLevelSettings), with no shared interface, so the builders cannot be merged. `traffic_policy.go` documents this; the leaf scalar mapping is centralized in `opt*` helpers so each per-path builder is a one-line struct literal.
- **Only `host` is unconditional** -- it is required upstream; `traffic_policy`, `subsets`, `workload_selector`, and `export_to` are attached only when present, so omitting them lets istiod defaults apply.
- **uint32/uint64 -> int conversions** -- proto port numbers, ring sizes, table sizes, and outlier counts are cast to the SDK's `int` inputs; the locality `to` map (`map<string,uint32>`) becomes a `pulumi.IntMap`.

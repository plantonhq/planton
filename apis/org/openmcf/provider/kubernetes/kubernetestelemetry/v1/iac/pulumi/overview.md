# KubernetesTelemetry Pulumi Module Architecture

## Data Flow

```
StackInput (spec + cluster credentials; foreign keys already resolved)
   ↓
Locals (telemetry_name = metadata.name, namespace, labels)
   ↓
Resources (namespaced Telemetry CR via apiextensions.CustomResource)
   ↓
Outputs (telemetry_name, namespace)
```

## File Organization

| File | Responsibility |
|------|---------------|
| `main.go` | Orchestrates resource creation: builds the `spec` map from the typed proto getters (selector, targetRefs, tracing, metrics, accessLogging) and exports outputs |
| `locals.go` | Extracts values from spec/metadata: resolves the `namespace` foreign key via `GetValue()`, computes standard labels |
| `outputs.go` | Defines output constant names matching `stack_outputs.proto` |

## Key Design Decisions

- **Untyped `apiextensions.CustomResource` (the one Istio exception)** -- crd2pulumi
  degrades the CRD's `tracing[].customTags` (a map of nested object-valued `oneOf`) to
  `map[string]map[string]string`, which cannot carry `{literal: {value: ...}}`. The typed
  SDK therefore cannot express a real custom tag, so the resource is built generically with
  the `spec` assembled from the typed proto getters. See `../../docs/README.md` section 5.
- **Input stays type-safe** -- every `spec` value is read through a typed proto getter
  (`GetTracing()`, `GetCustomTags()`, ...); only the Pulumi resource wrapper is generic.
- **Unions are mapped by which proto field is set** -- there is no discriminator. `CustomTag`
  emits only the set source (literal/environment/header); `MetricSelector` emits `metric` or
  `customMetric`.
- **Every block is attached only when present** -- unset fields are omitted from the `spec`
  map, so istiod defaults flow through.
- **Snake_case proto fields map to the CRD's camelCase JSON keys** (`randomSamplingPercentage`,
  `disableSpanReporting`, `customTags`, `tagOverrides`, `reportingInterval`,
  `useRequestIdForTraceSampling`, ...).

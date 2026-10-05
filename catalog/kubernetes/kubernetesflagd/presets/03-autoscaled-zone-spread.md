# Autoscaled Zone Spread

The production shape for heavy evaluation traffic: three to ten replicas autoscaled on CPU, spread across zones, with a disruption budget that keeps two pods serving. flagd is stateless, so every replica serves the same flags and scaling out adds evaluation throughput directly.

## When to Use

- Many services evaluate flags remotely against flagd (gRPC or OFREP)
- The cluster spans several availability zones
- Flag evaluation is on a request path where an outage is user-visible

## Key Configuration Choices

- `hpa` owns the replica count between 3 and 10 at 70% CPU; `replicas` is ignored while it is enabled
- `pdb.minAvailable: "2"` keeps two pods through voluntary disruptions
- `scheduling.topologySpreadConstraints` with no `matchLabels` spreads flagd's own pods (the module fills in its selector)
- `resources` raises the defaults for the higher load

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<namespace>` | Namespace for flagd and its flag file | Your cluster's namespace plan |
| `my-flags` | Name of the KubernetesFlagdFlagFile to serve | The flag file's `metadata.name` |

## Related Presets

- **Flag File Source** -- the two-replica standard shape

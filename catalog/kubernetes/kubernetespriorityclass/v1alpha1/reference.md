# KubernetesPriorityClass

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesPriorityClassSpec** defines one step of the cluster's workload
importance ladder. Pods reference the class by name (the shared workload
pod spec's `priority_class_name`); the scheduler places higher-priority
pods first when capacity is scarce and — unless preemption is disabled —
EVICTS lower-priority pods to make room for a higher-priority pod that
cannot otherwise schedule.

A typical ladder is small and deliberate: e.g. "critical" (1000000,
preempting) for revenue-path services, "standard" (1000, the global
default) for everything unmarked, and "batch" (-100, non-preempting) for
work that should never displace services. Built-in classes
`system-cluster-critical` and `system-node-critical` sit above the
user-definable range and belong to Kubernetes itself.

PriorityClasses are cluster-scoped. The spec covers the complete
scheduling.k8s.io/v1 PriorityClass surface.

## Example

```yaml
# Full-surface test manifest for the offline plan proofs. Exercises the
# priority value, global default, description, and non-preempting policy.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPriorityClass
metadata:
  name: test-priority-class
spec:
  name: test-priority-class
  labels:
    team: platform-engineering
  value: 1000000
  global_default: false
  description: High priority for revenue-path services; preempts lower tiers under pressure.
  preemption_policy: preempt_lower_priority
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.name` | `string` | yes |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.value` | `int32` |  |  |  |
| `spec.globalDefault` | `bool` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.preemptionPolicy` | `enum` |  | `preempt_lower_priority` |  |

## Field Details

### spec.name

`string` · required

The name of the PriorityClass (its `metadata.name` in the cluster) — the
value pods reference in `priority_class_name`. Must be a valid DNS
subdomain and must not use the reserved `system-` prefix.

- rule: Name must be a valid DNS subdomain (lowercase alphanumeric, hyphens, and dots, no leading/trailing dots or hyphens)
- rule: Names beginning with "system-" are reserved for Kubernetes' built-in priority classes
- rule: {"string":{"minLen":"1","maxLen":"253"}}

### spec.labels

`map<string, string>`

Additional labels to apply to the PriorityClass object.
Merged with the standard Planton governance labels.

### spec.annotations

`map<string, string>`

Annotations to apply to the PriorityClass object.

### spec.value

`int32`

The priority integer pods of this class receive — higher schedules (and
preempts) ahead of lower. User-defined classes must stay at or below
1,000,000,000 (one billion); the range above is reserved for Kubernetes
system classes. Negative values are valid and useful for
"always-preemptable" batch tiers. IMMUTABLE after creation: changing the
value replaces the class (both engines force replacement).

- rule: {"int32":{"lte":1000000000}}

### spec.globalDefault

`bool`

Makes this class the cluster-wide default for pods that name NO priority
class (pods that would otherwise get priority 0). Only one class should
be the global default — when several claim it, Kubernetes uses the
SMALLEST such value, which is rarely what anyone intended. Changing the
default never re-prioritizes existing pods; it applies to pods created
afterwards.

### spec.description

`string`

Human guidance on when to use this class (surfaced by `kubectl describe
priorityclass`). Write it for the next engineer choosing a class for
their workload.

### spec.preemptionPolicy

`enum` · optional (explicit presence)

Whether pending pods of this class preempt running pods.
Default: preempt_lower_priority

- default: `preempt_lower_priority`
- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `kubernetes_priority_class_preemption_policy_unspecified` -- Unspecified. Defaults to preempt_lower_priority.
- `preempt_lower_priority` -- Pods of this class evict lower-priority pods to make room — the default, and what "critical" tiers want.
- `never` -- Pods of this class wait in the queue ahead of lower-priority pods but never evict anything already running. The right policy for high-priority BATCH work: jump the line without disrupting services.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesPriorityClass, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.priority_class_name` | `string` | The name of the PriorityClass object as created in the cluster — the value pods put in `priority_class_name`. |
| `status.outputs.value` | `int32` | The priority integer pods of this class receive. |

## See Also

- [Overview](../README.md)

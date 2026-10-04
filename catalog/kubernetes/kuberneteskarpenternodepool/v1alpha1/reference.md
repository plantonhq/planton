# KubernetesKarpenterNodePool

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

KubernetesKarpenterNodePoolSpec declares a Karpenter NodePool — the fleet
shape Karpenter provisions machines from. A NodePool answers "what nodes
may exist": the instance-type/zone/capacity-type constraints
(requirements), the taints and labels new nodes carry, how long nodes
live, how aggressively under-used nodes are consolidated away, and the
resource ceiling the pool may reach. Clusters routinely run several —
a default on-demand pool, a spot pool, a GPU pool — and Karpenter picks
among them by weight.

The cloud-level machine template (AMIs, subnets, security groups, IAM)
lives in the NodeClass the pool references through node_class_ref —
KubernetesKarpenterEc2NodeClass on AWS.

The rendered NodePool is CLUSTER-SCOPED (no namespace) and named after
metadata.name. 100% fidelity with the upstream karpenter.sh/v1 NodePool
CRD at the pinned release (aws/karpenter-provider-aws
pkg/apis/crds/karpenter.sh_nodepools.yaml); the CRD's own CEL rules are
mirrored below so mistakes surface at validate time, not at apply.

## Example

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKarpenterNodePool
metadata:
  name: gp-spot-pool
spec:
  template:
    labels:
      pool-tier: general-purpose
      karpenter.sh/capacity-type: spot
    annotations:
      example.com/owner: platform-team
    nodeClassRef:
      group: karpenter.k8s.aws
      kind: EC2NodeClass
      name:
        value: gp-nodeclass
    requirements:
      - key: karpenter.sh/capacity-type
        operator: In
        values:
          - spot
      - key: kubernetes.io/arch
        operator: In
        values:
          - amd64
          - arm64
      - key: karpenter.k8s.aws/instance-category
        operator: NotIn
        values:
          - t
      - key: karpenter.k8s.aws/instance-generation
        operator: Gt
        values:
          - "2"
      - key: node.kubernetes.io/instance-type
        operator: In
        values:
          - m5.large
          - m5.xlarge
          - m6i.large
          - m6i.xlarge
        minValues: 2
      - key: topology.kubernetes.io/zone
        operator: Exists
    taints:
      - key: dedicated
        value: batch
        effect: NoSchedule
    startupTaints:
      - key: example.com/cni-not-ready
        effect: NoExecute
    expireAfter: 720h
    terminationGracePeriod: 48h
  disruption:
    consolidationPolicy: WhenEmptyOrUnderutilized
    consolidateAfter: 5m
    budgets:
      - nodes: 10%
      - nodes: "0"
        schedule: 0 9 * * 1-5
        duration: 8h
        reasons:
          - Underutilized
          - Empty
  limits:
    cpu: "1000"
    memory: 1000Gi
  weight: 50
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.template` | `KubernetesKarpenterNodePoolTemplate` | yes |  |  |
| `spec.template.labels` | `map<string, string>` |  |  |  |
| `spec.template.annotations` | `map<string, string>` |  |  |  |
| `spec.template.nodeClassRef` | `KubernetesKarpenterNodePoolNodeClassRef` | yes |  |  |
| `spec.template.nodeClassRef.group` | `string` |  | `karpenter.k8s.aws` |  |
| `spec.template.nodeClassRef.kind` | `string` |  | `EC2NodeClass` |  |
| `spec.template.nodeClassRef.name` | `string \| valueFrom` | yes |  | KubernetesKarpenterEc2NodeClass (`status.outputs.node_class_name`) |
| `spec.template.requirements` | `[]KubernetesKarpenterNodePoolRequirement` | yes |  |  |
| `spec.template.requirements[].key` | `string` | yes |  |  |
| `spec.template.requirements[].operator` | `string` | yes |  |  |
| `spec.template.requirements[].values` | `[]string` |  |  |  |
| `spec.template.requirements[].minValues` | `int32` |  |  |  |
| `spec.template.taints` | `[]KubernetesKarpenterNodePoolTaint` |  |  |  |
| `spec.template.taints[].key` | `string` | yes |  |  |
| `spec.template.taints[].value` | `string` |  |  |  |
| `spec.template.taints[].effect` | `string` | yes |  |  |
| `spec.template.startupTaints` | `[]KubernetesKarpenterNodePoolTaint` |  |  |  |
| `spec.template.startupTaints[].key` | `string` | yes |  |  |
| `spec.template.startupTaints[].value` | `string` |  |  |  |
| `spec.template.startupTaints[].effect` | `string` | yes |  |  |
| `spec.template.expireAfter` | `string` |  | `720h` |  |
| `spec.template.terminationGracePeriod` | `string` |  |  |  |
| `spec.disruption` | `KubernetesKarpenterNodePoolDisruption` |  |  |  |
| `spec.disruption.consolidationPolicy` | `string` |  | `WhenEmptyOrUnderutilized` |  |
| `spec.disruption.consolidateAfter` | `string` |  | `0s` |  |
| `spec.disruption.budgets` | `[]KubernetesKarpenterNodePoolDisruptionBudget` |  |  |  |
| `spec.disruption.budgets[].nodes` | `string` | yes |  |  |
| `spec.disruption.budgets[].schedule` | `string` |  |  |  |
| `spec.disruption.budgets[].duration` | `string` |  |  |  |
| `spec.disruption.budgets[].reasons` | `[]string` |  |  |  |
| `spec.limits` | `map<string, string>` |  |  |  |
| `spec.weight` | `int32` |  |  |  |
| `spec.replicas` | `int64` |  |  |  |

## Field Details

### spec.template

`KubernetesKarpenterNodePoolTemplate` · required

Template for the nodes this pool launches: the node's class reference,
scheduling requirements, taints, labels, and lifetime.

Deliberately FLATTER than the CRD: upstream nests these under
template.metadata (labels/annotations) and template.spec (everything
else); this spec folds both levels into one message because the split
carries no meaning for a manifest author. Both IaC modules rebuild the
CRD's exact nesting when rendering — when diffing this spec against the
upstream CRD, compare against the UNION of template.metadata and
template.spec fields.

- rule: {"required":true}

### spec.template.labels

`map<string, string>`

Labels applied to every node the pool launches (max 100). The
karpenter.sh and karpenter.k8s.aws label domains are restricted to
Karpenter's own well-known keys, and kubernetes.io/hostname can never
be set — mirrored from the CRD.

- rule: labels in the karpenter.sh domain are restricted to karpenter.sh/capacity-type, kubernetes.io/hostname can never be set, and karpenter.sh/nodepool is controller-owned
- rule: {"map":{"maxPairs":"100","values":{"string":{"maxLen":"63","pattern":"^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$"}}}}

### spec.template.annotations

`map<string, string>`

Annotations applied to every node the pool launches.

### spec.template.nodeClassRef

`KubernetesKarpenterNodePoolNodeClassRef` · required

Reference to the NodeClass carrying the cloud-level machine template.
Defaults address the AWS provider's EC2NodeClass.

- rule: {"required":true}

### spec.template.nodeClassRef.group

`string` · optional (explicit presence)

API group of the NodeClass. AWS provider: "karpenter.k8s.aws".

- default: `karpenter.k8s.aws`
- rule: node_class_ref group may not be empty
- rule: {"string":{"maxLen":"253","pattern":"^[^/]*$"}}

### spec.template.nodeClassRef.kind

`string` · optional (explicit presence)

Kind of the NodeClass. AWS provider: "EC2NodeClass".

- default: `EC2NodeClass`
- rule: node_class_ref kind may not be empty

### spec.template.nodeClassRef.name

`string | valueFrom` · required

Name of the NodeClass. Defaults to a KubernetesKarpenterEc2NodeClass
foreign key — wire it with valueFrom in an infra chart and the pool
deploys after its machine template. Pass the literal name with
`value:` when the NodeClass is not Planton-managed.

- references: KubernetesKarpenterEc2NodeClass (`status.outputs.node_class_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesKarpenterEc2NodeClass, name: <that resource's name>, fieldPath: status.outputs.node_class_name}} -- a bare string does not parse

### spec.template.requirements

`[]KubernetesKarpenterNodePoolRequirement` · required

Scheduling requirements constraining what nodes the pool may launch
(instance families, zones, capacity type, architecture, ...). Layered
with labels and applied to every node. At least one requirement is the
practical minimum — an unconstrained pool launches anything.

- rule: requirements with operator 'In' must have at least one value
- rule: requirements with operator 'Gt', 'Lt', 'Gte' or 'Lte' must have exactly one non-negative integer value
- rule: requirements with min_values must list at least that many values
- rule: {"repeated":{"minItems":"1","maxItems":"100"}}

### spec.template.requirements[].key

`string` · required

Label key the requirement applies to (a well-known scheduling key such
as karpenter.sh/capacity-type, node.kubernetes.io/instance-type,
topology.kubernetes.io/zone, kubernetes.io/arch, or any
karpenter.k8s.aws/instance-* discovery key). The karpenter.sh and
karpenter.k8s.aws domains are restricted to their well-known keys;
kubernetes.io/hostname and karpenter.sh/nodepool can never be
constrained — mirrored from the CRD.

- rule: kubernetes.io/hostname and karpenter.sh/nodepool cannot be constrained; the karpenter.sh domain only allows karpenter.sh/capacity-type
- rule: {"required":true,"string":{"maxLen":"316","pattern":"^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*(\\/))?([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$"}}

### spec.template.requirements[].operator

`string` · required

Relationship between the key and the values: In, NotIn, Exists,
DoesNotExist, Gt, Lt, Gte or Lte (the numeric operators take exactly
one integer value).

- rule: operator must be one of 'In', 'NotIn', 'Exists', 'DoesNotExist', 'Gt', 'Lt', 'Gte' or 'Lte'
- rule: {"required":true}

### spec.template.requirements[].values

`[]string`

Values for the operator (non-empty for In/NotIn; empty for
Exists/DoesNotExist; a single integer for Gt/Lt/Gte/Lte).

- rule: {"repeated":{"items":{"string":{"maxLen":"63","pattern":"^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$"}}}}

### spec.template.requirements[].minValues

`int32` · optional (explicit presence)

ALPHA: minimum number of distinct values the requirement must span in
the launched fleet — the instance-type-diversity knob for spot pools
(1-50).

- rule: {"int32":{"lte":50,"gte":1}}

### spec.template.taints

`[]KubernetesKarpenterNodePoolTaint`

Taints applied to every node — pods must tolerate them to schedule
onto the pool (the standard dedicated-pool pattern).

### spec.template.taints[].key

`string` · required

Taint key.

- rule: {"required":true,"string":{"minLen":"1","pattern":"^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*(\\/))?([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$"}}

### spec.template.taints[].value

`string`

Taint value (optional).

- rule: {"string":{"pattern":"^(([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*(\\/))?([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$"}}

### spec.template.taints[].effect

`string` · required

Effect on pods that do not tolerate the taint: NoSchedule,
PreferNoSchedule or NoExecute.

- rule: effect must be one of 'NoSchedule', 'PreferNoSchedule' or 'NoExecute'
- rule: {"required":true}

### spec.template.startupTaints

`[]KubernetesKarpenterNodePoolTaint`

Taints applied at startup and expected to be removed shortly by a
tolerating DaemonSet (initialization ordering). Unlike taints, pods do
NOT need to tolerate these for the pool to provision for them.

### spec.template.startupTaints[].key

`string` · required

Taint key.

- rule: {"required":true,"string":{"minLen":"1","pattern":"^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*(\\/))?([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$"}}

### spec.template.startupTaints[].value

`string`

Taint value (optional).

- rule: {"string":{"pattern":"^(([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*(\\/))?([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$"}}

### spec.template.startupTaints[].effect

`string` · required

Effect on pods that do not tolerate the taint: NoSchedule,
PreferNoSchedule or NoExecute.

- rule: effect must be one of 'NoSchedule', 'PreferNoSchedule' or 'NoExecute'
- rule: {"required":true}

### spec.template.expireAfter

`string` · optional (explicit presence)

Maximum node lifetime measured from creation (Go duration limited to
s/m/h, or "Never"). CRD default: "720h" (30 days) — expiry is
Karpenter's eventually-consistent node-recycling mechanism.

- default: `720h`
- rule: expire_after must be a duration using s/m/h units (e.g. "720h") or "Never"

### spec.template.terminationGracePeriod

`string` · optional (explicit presence)

Ceiling on how long node draining may take before pods are forcefully
terminated (duration, s/m/h). Overrides pod terminationGracePeriod
math and bypasses blocking PDBs once reached — the cluster admin's
guarantee that nodes CAN be cycled. Empty = wait indefinitely.

- rule: termination_grace_period must be a duration using s/m/h units (e.g. "48h")

### spec.disruption

`KubernetesKarpenterNodePoolDisruption`

Disruption policy: when Karpenter may consolidate or replace nodes it
provisioned from this pool, and how many may be disrupted at once.

### spec.disruption.consolidationPolicy

`string` · optional (explicit presence)

Which nodes consolidation may disrupt: "WhenEmptyOrUnderutilized" (CRD
default — replace or remove nodes whenever cheaper capacity fits the
pods), "WhenEmpty" (only nodes with no workload pods) or "Balanced".
Ignored in static mode (replicas set).

- default: `WhenEmptyOrUnderutilized`
- rule: consolidation_policy must be one of 'WhenEmpty', 'WhenEmptyOrUnderutilized' or 'Balanced'

### spec.disruption.consolidateAfter

`string` · optional (explicit presence)

How long a node must be underutilized/empty before consolidation acts
(duration s/m/h, or "Never" to disable consolidation). CRD default:
"0s" (immediate).

- default: `0s`
- rule: consolidate_after must be a duration using s/m/h units (e.g. "5m") or "Never"

### spec.disruption.budgets

`[]KubernetesKarpenterNodePoolDisruptionBudget`

Budgets limiting how many of the pool's nodes may be disrupted at
once, optionally on a schedule (e.g. business hours). CRD default when
omitted: one always-active budget of "10%". Multiple active budgets
apply the most restrictive value.

- rule: a budget's schedule and duration must be set together — the schedule starts the window, the duration ends it
- rule: {"repeated":{"maxItems":"50"}}

### spec.disruption.budgets[].nodes

`string` · required

Maximum nodes of this pool that may be disrupting at once — an
absolute count ("2") or a percentage ("10%"). "0" blocks disruption
entirely while the budget is active.

- rule: {"required":true,"string":{"pattern":"^((100|[0-9]{1,2})%|[0-9]+)$"}}

### spec.disruption.budgets[].schedule

`string` · optional (explicit presence)

Cron schedule (upstream cronjob syntax, or @hourly/@daily/...) at
which the budget becomes active. Timezones are not supported. Omitted
= always active. Must be set together with duration.

- rule: {"string":{"pattern":"^(@(annually|yearly|monthly|weekly|daily|midnight|hourly))|((.+)\\s(.+)\\s(.+)\\s(.+)\\s(.+))$"}}

### spec.disruption.budgets[].duration

`string` · optional (explicit presence)

How long the budget stays active after each schedule hit (minutes and
hours only — cron cannot express seconds). Must be set together with
schedule.

- rule: {"string":{"pattern":"^((([0-9]+(h|m))|([0-9]+h[0-9]+m))(0s)?)$"}}

### spec.disruption.budgets[].reasons

`[]string`

Disruption reasons the budget applies to (Underutilized, Empty,
Drifted). Empty list = all reasons.

- rule: {"repeated":{"maxItems":"50","items":{"cel":[{"id":"spec.disruption.budget.reason_enum","message":"each budget reason must be one of 'Underutilized', 'Empty' or 'Drifted'","expression":"this in ['Underutilized', 'Empty', 'Drifted']"}]}}}

### spec.limits

`map<string, string>`

Resource ceiling for the whole pool, as Kubernetes quantities keyed by
resource name (e.g. cpu: "1000", memory: "1000Gi", nodes: "10").
Provisioning stops when a limit is reached. Upstream: only
limits.nodes is supported when replicas (static mode) is set.

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.weight

`int32` · optional (explicit presence)

Priority of this pool relative to other pools during scheduling —
higher weights are considered first (1-100; a pool with no weight is
weight 0). Not supported when replicas (static mode) is set.

- rule: {"int32":{"lte":100,"gte":1}}

### spec.replicas

`int64` · optional (explicit presence)

ALPHA (static capacity): maintain a FIXED number of nodes instead of
scaling on pod demand. When set, disruption consolidation settings and
weight are ignored/forbidden, and only limits.nodes is allowed.
Requires the staticCapacity feature gate on the Karpenter controller.

- rule: {"int64":{"gte":"0"}}

## Validation Rules

- `spec.static_limits_nodes_only`: only the 'nodes' limit is supported on static NodePools (replicas set) — remove other limits or unset replicas
- `spec.static_weight_forbidden`: weight is not supported on static NodePools (replicas set) — static pools are not ranked, remove one of the two

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesKarpenterNodePool, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.node_pool_name` | `string` | Name of the NodePool object (cluster-scoped; equals metadata.name). The value of the karpenter.sh/nodepool label on every node the pool launches — the join key for pool-scoped scheduling and monitoring. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.template.nodeClassRef.name` | KubernetesKarpenterEc2NodeClass | `status.outputs.node_class_name` |

## See Also

- [Overview](../README.md)

# Kubernetes Karpenter Node Pool

## When NOT to Use This

**This declares WHAT nodes may exist — it is not the engine.** A
NodePool does nothing without a Karpenter installation
(KubernetesKarpenter) on the cluster; its CRD does not even exist before
the engine's CRD release is applied.

Also not the right kind when:

- **You need the cloud-level machine template** — AMIs, subnets,
  security groups, and IAM live in the NodeClass the pool references
  (KubernetesKarpenterEc2NodeClass on AWS), not here. One NodeClass is
  typically shared by several pools: the pools differ in constraints and
  taints; the class is the common "how a node is built".
- **Your capacity is pre-defined node groups** — growing and shrinking
  EC2 Auto Scaling groups or VM scale sets is
  KubernetesClusterAutoscaler territory; a NodePool describes machines
  Karpenter launches on demand.

## Overview

**KubernetesKarpenterNodePool** declares a Karpenter NodePool — the
fleet shape Karpenter provisions machines from. A NodePool answers "what
nodes may exist": the instance-type/zone/capacity-type constraints
(requirements), the taints and labels new nodes carry, how long nodes
live, how aggressively under-used nodes are consolidated away, and the
resource ceiling the pool may reach. Clusters routinely run several — a
default on-demand pool, a spot pool, a GPU pool — and Karpenter picks
among them by weight.

The rendered NodePool is CLUSTER-SCOPED (no namespace) and named after
`metadata.name`. The spec holds 100% fidelity with the upstream
`karpenter.sh/v1` NodePool CRD at the pinned release, and the CRD's own
CEL rules are mirrored into the spec so mistakes surface at validate
time, not at apply.

**Key design points:**

- **Requirements are the practical minimum** — at least one is required;
  an unconstrained pool launches anything. The `karpenter.sh` and
  `karpenter.k8s.aws` label domains are restricted to Karpenter's
  well-known keys, and `kubernetes.io/hostname` /
  `karpenter.sh/nodepool` can never be constrained — mirrored from the
  CRD.
- **Consolidation defaults on** — `WhenEmptyOrUnderutilized` with
  `consolidate_after: 0s` and node expiry at 720h (30 days) are the CRD
  defaults; disruption budgets bound how many nodes may be disrupted at
  once (default one always-active budget of 10%).
- **Static capacity is an alpha mode** — setting `replicas` maintains a
  FIXED node count instead of scaling on demand; it requires the
  staticCapacity feature gate on the Karpenter controller, forbids
  `weight`, and allows only the `nodes` limit — the CRD's own
  static-mode rules, mirrored as CEL so a manifest the API server would
  reject never leaves validation.

## Essential Configuration Fields

### Required

- **`spec.template.node_class_ref`**: the NodeClass carrying the machine
  template — `group`/`kind` default to the AWS provider's
  `karpenter.k8s.aws` / `EC2NodeClass`; `name` is a literal or a
  KubernetesKarpenterEc2NodeClass reference
- **`spec.template.requirements`**: 1–100 scheduling requirements
  (well-known keys such as `karpenter.sh/capacity-type`,
  `node.kubernetes.io/instance-type`, `topology.kubernetes.io/zone`,
  `kubernetes.io/arch`, or any `karpenter.k8s.aws/instance-*` discovery
  key) with Karpenter's extended operators — `In`, `NotIn`, `Exists`,
  `DoesNotExist`, and the single-integer `Gt`/`Lt`/`Gte`/`Lte`

### Common

- **`spec.template.labels` / `annotations`**: applied to every node the
  pool launches (labels max 100; the restricted domains apply)
- **`spec.template.taints` / `startup_taints`**: `taints` require pods
  to tolerate them (the dedicated-pool pattern); `startup_taints` are
  expected to be removed shortly by a tolerating DaemonSet, and pods do
  NOT need to tolerate them for provisioning
- **`spec.template.expire_after`**: maximum node lifetime (CRD default
  `720h`, or `Never`) — Karpenter's eventually-consistent node-recycling
  mechanism
- **`spec.template.termination_grace_period`**: ceiling on node draining
  before pods are forcefully terminated — overrides pod
  terminationGracePeriod math and bypasses blocking PDBs once reached;
  empty waits indefinitely
- **`spec.disruption`**: `consolidation_policy`
  (`WhenEmptyOrUnderutilized` CRD default, `WhenEmpty`, or `Balanced`),
  `consolidate_after` (CRD default `0s`, or `Never`), and `budgets`
  (absolute counts or percentages, optionally on a cron schedule with a
  duration — `"0"` blocks disruption entirely while active)
- **`spec.limits`**: resource ceiling for the whole pool as Kubernetes
  quantities (`cpu: "1000"`, `memory: 1000Gi`, `nodes: "10"`) —
  provisioning stops when a limit is reached
- **`spec.weight`**: priority among pools (1–100, higher first; absent
  means 0)
- **`spec.replicas`**: ALPHA static capacity — a fixed node count
  instead of demand-driven provisioning
- **`min_values`** on an `In` requirement: ALPHA — minimum distinct
  values the launched fleet must span (1–50), the
  instance-type-diversity knob for spot pools

## Outputs

| Output | Purpose |
|---|---|
| `node_pool_name` | Name of the cluster-scoped NodePool (equals `metadata.name`) — the value of the `karpenter.sh/nodepool` label on every node the pool launches, the join key for pool-scoped scheduling and monitoring |

## Composing in Infra Charts

- **`spec.template.node_class_ref.name`** is a foreign key (default kind
  KubernetesKarpenterEc2NodeClass, field path
  `status.outputs.node_class_name`) — wire it with `valueFrom` and the
  pool deploys after its machine template. Pass the literal name with
  `value:` when the NodeClass is not platform-managed.
- **The chain deploys in dependency order**: KubernetesKarpenter (the
  engine and CRDs) → KubernetesKarpenterEc2NodeClass →
  KubernetesKarpenterNodePool. The pool needs no reference to the engine
  itself — it is a cluster-visible custom resource the controller
  watches — but cannot be applied before the CRDs exist.
- **`node_class_ref.group` and `kind` are immutable** on the live API
  after creation.

## Examples

### General-purpose spot pool with instance diversity

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKarpenterNodePool
metadata:
  name: general-spot
spec:
  template:
    nodeClassRef:
      name:
        valueFrom:
          kind: KubernetesKarpenterEc2NodeClass
          name: default-al2023
          fieldPath: status.outputs.node_class_name
    requirements:
      - key: karpenter.sh/capacity-type
        operator: In
        values: [spot]
      - key: kubernetes.io/arch
        operator: In
        values: [amd64]
      - key: karpenter.k8s.aws/instance-category
        operator: NotIn
        values: [t]
  limits:
    cpu: "1000"
  weight: 50
```

### Dedicated GPU pool (taints + on-demand)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKarpenterNodePool
metadata:
  name: gpu-pool
spec:
  template:
    nodeClassRef:
      name:
        value: gpu-nodeclass
    labels:
      pool-tier: gpu
    taints:
      - key: nvidia.com/gpu
        value: "true"
        effect: NoSchedule
    requirements:
      - key: karpenter.sh/capacity-type
        operator: In
        values: [on-demand]
      - key: node.kubernetes.io/instance-type
        operator: In
        values: [g5.xlarge, g5.2xlarge]
  disruption:
    consolidationPolicy: WhenEmpty
    consolidateAfter: 5m
```

### Business-hours disruption freeze

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKarpenterNodePool
metadata:
  name: steady-pool
spec:
  template:
    nodeClassRef:
      name:
        value: default-al2023
    requirements:
      - key: karpenter.sh/capacity-type
        operator: In
        values: [on-demand]
  disruption:
    budgets:
      - nodes: 10%
      - nodes: "0" # block all disruption during business hours
        schedule: 0 9 * * 1-5
        duration: 8h
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

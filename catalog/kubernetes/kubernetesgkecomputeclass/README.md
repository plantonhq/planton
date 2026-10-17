# KubernetesGkeComputeClass

Declares a GKE `ComputeClass`: a cluster-scoped, ordered list of the machines GKE may create nodes from. When pods that select the class are pending, GKE tries the first priority rule, and when it cannot get that capacity (a stockout, a full reservation, a quota) it tries the next. The spec mirrors the `cloud.google.com/v1` ComputeClass field for field as GKE serves it at 1.36.4-gke.2046002, after a small Planton envelope: the object's name, labels and annotations.

## When NOT to Use This

- **The cluster is not on GKE.** ComputeClass is GKE's own resource. On EKS the equivalents are `KubernetesKarpenterNodePool` and `KubernetesKarpenterEc2NodeClass`.
- **You need one fixed pool of one machine type.** A `GcpGkeNodePool` is simpler. Reach for a class when the machine may change: fallback across families, Spot first, reservations first, accelerators.

## What Gets Created

- **One ComputeClass**, cluster-scoped, named by the spec's `name`.
- **Labels on the object**: the spec's `labels`, with Planton's identity labels stamped on top.

GKE reads the class whenever pods that select it are pending, and, with `nodePoolAutoCreation.enabled`, creates and deletes node pools for its rules. Those node pools are GKE's, created by the cluster autoscaler, not by this module.

## How Pods Use a Class

- **By name.** A pod sets the node selector `cloud.google.com/compute-class: <name>`; GKE adds the toleration for the class's taint.
- **Per namespace.** The namespace label `cloud.google.com/default-compute-class: <name>` applies the class to every pod in it that names none.
- **Cluster-wide.** A class named `default` applies to every pod that selects none, once the `GcpGkeCluster` turns on `clusterAutoscaling.defaultComputeClassEnabled`.

## Quick Start

The stockout-resilient cluster default:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGkeComputeClass
metadata:
  name: prod-default-compute-class
  annotations:
    planton.dev/connection: prod-gke # the cluster's Kubernetes connection
spec:
  name: default
  priorities:
    - machineFamily: n2d
      minCores: 8
    - machineFamily: c3d
      minCores: 8
    - machineFamily: n2
      minCores: 8
    - machineFamily: t2d
      minCores: 8
  whenUnsatisfiable: ScaleUpAnyway
  nodePoolAutoCreation:
    enabled: true
```

```shell
planton apply -f compute-class.yaml
```

## Essential Configuration Fields

### Envelope

| Field | Description |
|-------|-------------|
| `name` | The ComputeClass's name in the cluster: the value pods select, and `default` for the cluster default. Names beginning with `autopilot` are GKE's built-in classes. Required. |
| `labels` | The object's own labels, for policy engines and inventory. They do not reach nodes. |
| `annotations` | The object's own annotations. |

### The Class

| Field | Description |
|-------|-------------|
| `priorities` | The rules, tried in order (or by `priorityScore`). Each names a machine family or type, existing node pools, a GPU or TPU, or a pod family, and how to pay: Spot, flex-start or reservations. |
| `whenUnsatisfiable` | `DoNotScaleUp` (default; pods wait) or `ScaleUpAnyway` (GKE's default machine series runs them). |
| `nodePoolAutoCreation.enabled` | Lets GKE create node pools for the rules. Off, the class only uses pools that already exist. |
| `nodePoolConfig` | Image, service account, metadata server, labels, taints, sandbox and confidential computing for every node. |
| `priorityDefaults` | Default location and kubelet and kernel settings for every rule. |
| `activeMigration.optimizeRulePriority` | Moves pods back to higher-priority capacity once it returns. |
| `autoscalingPolicy` | How soon idle nodes are removed. |
| `minimumCapacity` | Nodes kept with no pods pending (commits spend). |
| `autopilot.enabled` | Autopilot-managed nodes for the class; fixed once set. |

The generated [reference](v1alpha1/reference.md) documents every field, including the per-rule storage, location, reservation, kubelet and sysctl settings.

## Outputs

| Output | Description |
|--------|-------------|
| `compute_class_name` | The class's name in the cluster (the spec's `name`): the value a workload's node selector carries. |

## Composing in Infra Charts

Deploy the class through the cluster's Kubernetes connection and turn the cluster's default-class switch on beside it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGkeComputeClass
metadata:
  name: "{{ values.env }}-gke-default-class"
  annotations:
    planton.dev/connection: "{{ values.cluster_connection_name }}"
  relationships:
    - kind: GcpGkeCluster
      name: "{{ values.env }}-gke"
      type: runs_on
spec:
  name: default
  priorities:
    - nodepools:
        - valueFrom:
            kind: GcpGkeNodePool
            name: "{{ values.env }}-gke-system"
            fieldPath: status.outputs.node_pool_name
    - machineFamily: n2d
      minCores: 4
  nodePoolAutoCreation:
    enabled: true
```

References compose the class with the rest of the graph: `nodepools` with `GcpGkeNodePool`, `nodePoolConfig.serviceAccount` with `GcpServiceAccount`, `storage.bootDiskKMSKey` with `GcpKmsKey`, and a reservation's or disk image's `project` with `GcpProject`.

## Examples

The presets carry four starting points: a stockout-resilient cluster default, Spot first with on-demand fallback, GPU inference on L4s, and reserved capacity first. The [GKE capacity fallback pattern](../../_patterns/gke-capacity-fallback.md) composes the class with the cluster and its node pools.

## Related Kinds

- **GcpGkeCluster**: the cluster; `clusterAutoscaling.defaultComputeClassEnabled` makes a `default` class cluster-wide.
- **GcpGkeNodePool**: fixed pools a `nodepools` rule can use before GKE creates new ones.
- **GcpServiceAccount**: the node identity for `nodePoolConfig.serviceAccount`.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

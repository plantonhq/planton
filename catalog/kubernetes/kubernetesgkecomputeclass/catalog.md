# GKE Compute Class

Declares a GKE ComputeClass: an ordered list of the machines GKE may create nodes from, tried in order, so a cluster keeps getting nodes when Google runs out of one machine type. Pods select a class by name, a namespace can default to one, and a class named `default` covers the whole cluster. The spec is faithful to the `cloud.google.com/v1` ComputeClass GKE serves (pinned to 1.36.4-gke.2046002 and checked field for field against it), after a Planton envelope that names the object and sets its own labels. This is how you make a GKE cluster's capacity resilient, cheap-first or reservation-first without hand-writing node pools.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **A ComputeClass** -- one cluster-scoped object GKE's cluster autoscaler reads whenever pods that select it are pending; with node pool auto-creation on, GKE creates and removes node pools for its rules.
- **Kubernetes Labels** -- your own labels on the object, with resource metadata labels (resource name, kind, organization, environment) applied automatically on top.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- the GKE cluster's connection. A **GKE Cluster** deployed through Planton publishes one; map it as the default for your environment, or name it with the `planton.dev/connection` annotation.

### Kubernetes Cluster

- **A GKE cluster** -- Standard or Autopilot. GKE serves the ComputeClass definition itself; there is nothing to install.
- **The default compute class switch** -- for a class named `default`, turn on `clusterAutoscaling.defaultComputeClassEnabled` on the **GKE Cluster**.
- **GKE 1.33.3-gke.1136000 or later** -- for node pool auto-creation without cluster-level node auto-provisioning. Individual fields name the version that introduced them.

## Deploy

### Console

Open the deployment store, find **GKE Compute Class**, and click **Deploy**. The creation wizard walks you through the class name (with a Cluster Default switch), the ordered priority rules, the node pool configuration and the scaling behavior.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGkeComputeClass
metadata:
  name: prod-default-compute-class
  org: acme-corp
  env: prod
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
  activeMigration:
    optimizeRulePriority: true
```

```shell
planton apply -f compute-class.yaml
```

With the cluster's default class switch on, every pod that selects no class now gets n2d nodes, and c3d, n2 or t2d nodes when n2d runs out. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, start the class with the chart's own system node pool so the cluster uses capacity it already pays for first:

```yaml
spec:
  name: default
  priorities:
    - nodepools:
        - valueFrom:
            kind: GcpGkeNodePool
            name: prod-gke-system
            fieldPath: status.outputs.node_pool_name
    - machineFamily: n2d
      minCores: 4
  nodePoolAutoCreation:
    enabled: true
```

The InfraPipeline creates the cluster and its node pool first, then the class through the cluster's Kubernetes connection.

## Key Configuration

These are the most important decisions when configuring a class. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Order is the design.** GKE tries the rules top to bottom and creates nodes from the first it can satisfy. List the shape you want first, then families that can run the same pods. Keep `minCores` equal across fallback rules, so a fallback node is not smaller than the pods expect.

**Nothing is created without auto-creation.** `nodePoolAutoCreation.enabled` defaults to false upstream; without it the class only places pods on existing pools that match a rule. A family fallback needs it on.

**What happens when every rule fails.** `DoNotScaleUp` (the default) leaves pods pending until a rule can be met; `ScaleUpAnyway` runs them on GKE's default series. A cluster-wide default usually wants `ScaleUpAnyway`; a GPU or confidential class wants `DoNotScaleUp`.

**Disks follow the family.** C4, C4D, N4 and other fourth-generation families boot only from Hyperdisk. Give their rules `storage.bootDiskType: hyperdisk-balanced`, or GKE skips them.

**Coming back from a fallback.** `activeMigration.optimizeRulePriority` moves pods back to the first rule once its capacity returns. Pods are evicted to move, so pair it with PodDisruptionBudgets.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpGkeNodePool** | `priorities[].nodepools[]` | `status.outputs.node_pool_name` |
| **GcpServiceAccount** | `nodePoolConfig.serviceAccount` | `status.outputs.email` |
| **GcpKmsKey** | `priorities[].storage.bootDiskKMSKey` | `status.outputs.key_id` |
| **GcpProject** | `priorities[].reservations.specific[].project`, `priorities[].storage.secondaryBootDisks[].project` | `status.outputs.project_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `compute_class_name` | The class's name in the cluster | A workload's `cloud.google.com/compute-class` node selector, or a namespace's default-class label |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Survive a machine-type stockout** -- a cluster default that falls back across four families of the same size. Start from the **Stockout-Resilient Default** preset.

**Run cheap first** -- Spot across several families, on-demand only when Spot runs out, back to Spot when it returns. Start from the **Spot First, On-Demand Fallback** preset.

**Serve models on GPUs** -- Spot L4s first, on-demand L4s next, never a CPU-only node. Start from the **GPU L4 With Fallback** preset.

**Spend reservations before buying** -- a named reservation, then any matching one, then on-demand. Start from the **Reservation First** preset.

## Works With

- [**GKE Cluster**](/infra-catalog/gcp-gke-cluster) -- the cluster the class lives in; its default compute class switch makes a `default` class cluster-wide.
- [**GKE Node Pool**](/infra-catalog/gcp-gke-node-pool) -- fixed pools a `nodepools` rule uses before GKE creates new ones.
- [**Service Account**](/infra-catalog/gcp-service-account) -- the least-privilege identity the class's nodes run as.

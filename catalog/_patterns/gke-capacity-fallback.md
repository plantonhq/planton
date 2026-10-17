---
kinds:
  - GcpGkeCluster
  - GcpGkeNodePool
  - KubernetesGkeComputeClass
---

# GKE Capacity Fallback: Keep Getting Nodes Through a Stockout

A GKE cluster whose nodes all come from one machine type stops growing the moment Google runs out of that type in the region. Node creation fails with `ZONE_RESOURCE_POOL_EXHAUSTED` (or `GCE_STOCKOUT`) in every zone, quota untouched, and everything that waits for nodes times out: an install that schedules pods, a rollout, a scale-up. A busy region does this for hours at a time. This pattern is the composition that turns it into a non-event: the cluster's default compute class switch, a `default` class that lists several machine families, and the fixed node pool the class uses first.

## The composition

Three kinds, one decision each:

1. **GcpGkeCluster** turns on `clusterAutoscaling.defaultComputeClassEnabled`. That makes a class named `default` apply to every pod that selects no class. It needs neither node auto-provisioning nor its resource limits on GKE 1.33.3-gke.1136000 or later.
2. **GcpGkeNodePool** is the small fixed pool the cluster's controllers already run on.
3. **KubernetesGkeComputeClass** named `default` lists that pool first, then several machine families of the same size, with node pool auto-creation on.

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGkeCluster
metadata:
  name: prod-gke
  annotations:
    planton.dev/connection-name: prod-gke
spec:
  projectId:
    value: my-project
  location: us-central1
  network:
    valueFrom:
      kind: GcpVpcNetwork
      name: prod-vpc
      fieldPath: status.outputs.network_self_link
  subnetwork:
    valueFrom:
      kind: GcpSubnetwork
      name: prod-subnet
      fieldPath: status.outputs.subnetwork_self_link
  clusterAutoscaling:
    defaultComputeClassEnabled: true
---
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGkeNodePool
metadata:
  name: prod-gke-system
spec:
  projectId:
    value: my-project
  clusterName:
    valueFrom:
      kind: GcpGkeCluster
      name: prod-gke
      fieldPath: status.outputs.name
  location:
    valueFrom:
      kind: GcpGkeCluster
      name: prod-gke
      fieldPath: status.outputs.location
  autoscaling:
    minNodes: 1
    maxNodes: 2
  nodeConfig:
    machineType: e2-standard-4
---
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGkeComputeClass
metadata:
  name: prod-gke-default-class
  annotations:
    planton.dev/connection: prod-gke
spec:
  name: default
  priorities:
    - nodepools:
        - valueFrom:
            kind: GcpGkeNodePool
            name: prod-gke-system
            fieldPath: status.outputs.node_pool_name
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

## Why the fixed pool comes first

Once a `default` class applies, a pod that selects no class schedules only onto nodes of that class: nodes GKE created for its rules, or pools a `nodepools` rule names. Leave the fixed pool out and it sits idle while GKE creates new nodes for the very controllers it was sized for. Listing it first uses the capacity the cluster already pays for, and the families below it are where GKE goes when the pool is full or unavailable.

## Choosing the families

- **Same architecture, same disks.** n2d, c3d, n2 and t2d are x86 and boot from pd-balanced. A fourth-generation family (C4, C4D, N4) boots only from Hyperdisk, so its rule needs `storage.bootDiskType: hyperdisk-balanced`. An ARM family needs arm64 images and lands behind GKE's ARM taint.
- **Same size.** The same `minCores` on every rule, so a fallback node fits exactly the pods the first choice would have run.
- **Different hardware lines.** AMD (n2d, c3d, t2d) and Intel (n2) draw on different capacity, which is what makes the list resilient. Four families of one line run out together more often.

## The two decisions that make or break it

- **`nodePoolAutoCreation.enabled: true`.** It defaults to false. Without it the class never creates the fallback family's pool, and only the fixed pool is ever used.
- **`whenUnsatisfiable`.** `ScaleUpAnyway` runs pods on GKE's default series when every rule fails, which keeps installs moving through a total stockout of the listed families. `DoNotScaleUp` leaves them pending; choose it for classes whose machine is a requirement, not for the cluster default.

## On the diagram

The class draws an edge to the node pool its `nodepools` rule references and runs on the cluster through its Kubernetes connection. The cluster's switch and the class's name `default` are what join them in behavior, and neither draws an edge, so a reviewer checks the pair deliberately: a `default` class on a cluster whose switch is off applies to nothing.

## See also

- [KubernetesGkeComputeClass](../kubernetes/kubernetesgkecomputeclass/GUIDE.md): order, Spot first, reservations first, and what GKE refuses at apply.
- [GcpGkeCluster](../gcp/gcpgkecluster/GUIDE.md): node auto-provisioning and the stockout entry.
- [GcpGkeNodePool](../gcp/gcpgkenodepool/GUIDE.md): fixed pools and their rotation.

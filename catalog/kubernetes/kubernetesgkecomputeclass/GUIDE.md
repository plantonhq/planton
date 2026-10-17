# KubernetesGkeComputeClass Guide

The judgment this guide carries: a compute class is only as good as its order and its fallbacks. GKE runs out of single machine types in busy regions for hours at a time, and a class that lists one family, or several families that cannot run the same pods, fails exactly when it was meant to help. Get the order, the sizes and the unsatisfiable posture right, and turn node pool auto-creation on.

## A class or a node pool

Use a [GcpGkeNodePool](../../gcp/gcpgkenodepool/GUIDE.md) when the machine is fixed and you want to see every node before it exists: a system pool for cluster controllers, a pool with a hand-tuned kubelet. Use a class when the machine may change: falling back across families during a stockout, Spot first, a reservation first, accelerators where the zone decides. The two compose. A `nodepools` rule at the top of a class sends pods to an existing pool first, and the rules below it are where GKE goes when that pool is full.

## Surviving a stockout

`ZONE_RESOURCE_POOL_EXHAUSTED` (or `GCE_STOCKOUT`) on node creation means Google has no capacity for that machine type in that zone, and in a busy region it is often every zone at once. Quota is untouched, so raising quota does nothing. A class answers it with breadth:

- **List families that run the same pods.** n2d, c3d, n2 and t2d are all x86, all boot from pd-balanced, and all run the same images. An ARM family (t2a, c4a) in the same list needs arm64 images and lands behind GKE's ARM taint.
- **Keep the size equal.** Set the same `minCores` on every fallback rule. Without it GKE may pick a smaller machine from the fallback family, and pods sized for 8 cores stay pending on 2-core nodes.
- **Mind the disks.** C4, C4D, N4 and the other fourth-generation families boot only from Hyperdisk. A rule for them without `storage.bootDiskType: hyperdisk-balanced` is skipped, and you lose the fallback you thought you had.
- **Turn auto-creation on.** `nodePoolAutoCreation.enabled` defaults to false. Without it the class never creates the fallback family's pool, and the class is a list of wishes.
- **Decide the last resort.** `whenUnsatisfiable: ScaleUpAnyway` runs pods on GKE's default series when every rule fails, which is right for a cluster default that installs and controllers depend on. `DoNotScaleUp`, the default, is right when an unlisted machine is worse than waiting: GPUs, confidential nodes, licensed software tied to a family.
- **Come back.** `activeMigration.optimizeRulePriority` moves pods to the first rule's family once it has capacity again. It evicts pods to move them, so give Deployments PodDisruptionBudgets.

The [GKE capacity fallback pattern](../../_patterns/gke-capacity-fallback.md) composes the cluster switch, the class and the node pools for a whole cluster.

## The default class, and who it applies to

A class named `default` applies to every pod that selects no class, but only after the [GcpGkeCluster](../../gcp/gcpgkecluster/GUIDE.md) turns on `clusterAutoscaling.defaultComputeClassEnabled`. Until then it is an ordinary class nobody selects.

Once it applies, pods without a class schedule only onto nodes of the class: nodes GKE created for its rules, or pools a `nodepools` rule names. A cluster whose controllers run on a fixed system pool must list that pool in a `nodepools` rule at the top of the `default` class, or the system pool sits idle while GKE creates new nodes for the controllers. A namespace can choose a different class for its pods with the label `cloud.google.com/default-compute-class: <name>`.

## Cheap first

Spot rules first, on-demand rules after them, is the cost pattern. Spot capacity runs out per family and per zone too, so list several Spot families; one Spot rule followed by on demand spends on-demand money every time that one family is short. A Specific reservation first does the same for committed capacity: `affinity: Specific` never falls back by itself, so the rules after it are the fallback.

## What GKE refuses that the spec cannot see

The spec mirrors the definition's own rules, so most mistakes fail before the apply. A few depend on the cluster, and GKE's admission refuses them at apply:

- zones outside the cluster's region, in `location.zones` or a reservation's `zones`;
- local SSDs (`storage.localSSDCount`) on a rule with no `machineType` or GPU;
- `priorityDefaults.location.zoneTypes` when any rule consumes a Specific reservation;
- a field the cluster's GKE version does not serve yet (each field names the version that introduced it).

Others GKE accepts and then marks unusable. After applying, `kubectl describe computeclass <name>` shows the class's conditions: `Health=True` when GKE can use it, or `CrdMisconfigured=True` with a reason and `Health=False`. Seen in practice:

- `ReservationNotFound`: a Specific reservation that does not exist or the cluster cannot reach (another project's reservation needs sharing to this one);
- `NoSuitableMachineExists`: a rule no machine can satisfy, such as a `minCores` larger than any machine of the family.

A misconfigured class schedules nothing, so check the conditions after every change. A rule that is valid but cannot get capacity right now (a stockout, a quota) is not misconfigured: GKE skips it at scale-up and the class stays healthy, which is exactly the fallback a class exists for.

## On the diagram

The class draws edges to what it references: the node pools of a `nodepools` rule, the node service account, a CMEK key, a reservation's project. It runs on the cluster through its Kubernetes connection, and charts add a `runs_on` relationship to the cluster. The pods that select it are joined by a node selector, not a reference, so the diagram does not draw them.

## Design rationale

The spec mirrors the served ComputeClass field for field, under the upstream keys, so a class written for GKE reads the same here. The kind adds three things:

- **The envelope.** `name` is the object's name because the name is configuration here: pods select a class by it, and `default` is special. Keeping it in the spec lets every cluster in an environment carry its own `default` class. `labels` and `annotations` are the object's own metadata.
- **References.** Node pools, the node service account, the boot disk key and a reservation's project are references, so a class composes with the resources that provide them.
- **The definition's rules, moved forward.** The ComputeClass definition's own validation rules are mirrored as spec rules. The two that compare a class with its previous version (Autopilot is fixed once set) cannot be checked before the apply; GKE refuses the change.

Keys proto spells differently keep their upstream spelling: `bootDiskKMSKey`, `localSSDCount`, `maxSharedClientsPerGPU`, `singleProcessOOMKill`, the `hugepage_size*` keys, and the sysctls under their dotted names (`net.core.somaxconn`). Write those keys exactly as upstream does.

## Parity accounting

- **Pinned:** the ComputeClass definition GKE serves at 1.36.4-gke.2046002, committed at `pkg/kubernetes/manifestprojection/testdata/schemas/cloud.google.com_computeclasses.yaml`.
- **Measured:** the mirror test compares the spec with the pinned definition both ways and type for type. Every property of the served spec is a field, and no field is unknown to the definition.
- **Composed:** nothing; the kind applies one object.
- **Excluded:** nothing. Fields GKE documents for later versions (for example `enableNestedVirtualization`, 1.37) enter when the pin moves to a version that serves them.

## Pairs well with

- [GcpGkeCluster](../../gcp/gcpgkecluster/GUIDE.md): the default compute class switch and node auto-provisioning.
- [GcpGkeNodePool](../../gcp/gcpgkenodepool/GUIDE.md): fixed pools a `nodepools` rule uses first.
- [The GKE capacity fallback pattern](../../_patterns/gke-capacity-fallback.md): the whole composition for a cluster.

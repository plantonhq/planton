# Kubernetes Persistent Volume Claim

## Overview

**KubernetesPersistentVolumeClaim** is a catalog kind that creates and manages Kubernetes PersistentVolumeClaims — the durable-disk primitive — as first-class, declaratively managed resources. A claim names how much storage it needs, how it will be accessed, and (optionally) which StorageClass provisions it; the cluster binds it to a PersistentVolume that satisfies the request.

The kind covers the complete `core/v1` PersistentVolumeClaimSpec surface: access modes, storage requests and limits, StorageClass selection (including the empty-vs-absent distinction), volume mode, static binding to a named volume, volume selectors, and data sources (clone a claim or restore a snapshot). The feature-gated `volumeAttributesClassName` and cross-namespace data sources are deliberately unmodeled until they graduate.

## Purpose

A standalone claim is the right shape for storage whose lifecycle is independent of any one workload: a volume shared by several pods (ReadWriteMany), a data volume a Deployment mounts (via the workload's volume mounts referencing the claim by name), or a pre-provisioned volume being adopted. Per-replica storage for a StatefulSet should use the workload's own `volume_claim_templates` instead — those claims are stamped and tracked by the StatefulSet controller.

**Key value over raw manifests:**

- **Schema-level validation**: Kubernetes-quantity format checks on sizes, access-mode vocabulary enforcement, selector operator contracts, and the class-name conflict rule — all caught before anything reaches the cluster
- **The empty-vs-absent class distinction, made typed**: Kubernetes distinguishes an EMPTY `storageClassName` (bind only pre-provisioned volumes) from an ABSENT one (cluster default applies) — a distinction a single string field cannot carry. The spec carries `disable_dynamic_provisioning` as its own field, and validation rejects combining it with a named class
- **Namespace and class by value or reference**: `spec.namespace` and `spec.storage_class_name` accept literal names or references to `KubernetesNamespace` / `KubernetesStorageClass` resources, so an infra chart can create the class, the namespace, and the claim in one run
- **Deploys never hang on binding**: Neither engine waits for the claim to reach Bound — deliberate, because a claim under a `wait_for_first_consumer` class is correctly Pending until a pod consumes it
- **Dual IaC support**: Both Pulumi and Terraform implementations with feature parity (one documented exception: data sources, below)
- **Lifecycle management**: Integrated with Planton's deployment lifecycle for status tracking and outputs

## Binding Timing: Pending Is Not an Error

With a `wait_for_first_consumer` StorageClass (the norm for zonal cloud disks and kind's local-path), a claim stays **Pending until a pod uses it** — the volume then provisions in the zone that pod scheduled into. This is correct behavior, not a failure. Neither IaC engine blocks on the claim reaching Bound (the Pulumi module sets the skip-await annotation; the Terraform module sets `wait_until_bound = false`), so deploys never hang waiting for a consumer that arrives later.

## Choosing the StorageClass

Three mutually understood shapes:

- **Omit `storage_class_name`**: the cluster's DEFAULT StorageClass applies — fine for ad-hoc claims
- **Set `storage_class_name`**: pin the claim to a named class (literal value or reference to a `KubernetesStorageClass` resource) — the production shape
- **Set `disable_dynamic_provisioning: true`**: pin `storageClassName` to the empty string; the claim binds only to a matching pre-provisioned PersistentVolume, never triggering dynamic provisioning. Mutually exclusive with `storage_class_name` — the schema rejects setting both

## Essential Configuration Fields

### Required

- **`spec.name`**: The claim name (DNS subdomain: lowercase alphanumeric, hyphens, dots, 1–253 chars) — the value workloads reference in their PVC volume mounts
- **`spec.storage_request`**: Requested storage as a Kubernetes quantity (e.g. `"10Gi"`, `"500Mi"`). Growing later requires the claim's class to allow volume expansion (and Kubernetes never shrinks a volume)

### Common

- **`spec.namespace`**: Literal namespace name or reference to a KubernetesNamespace resource. When omitted, the claim lands in the cluster's `default` namespace
- **`spec.access_modes`**: How the volume may be mounted. Defaults to `["ReadWriteOnce"]` — one node at a time, the mode every block-storage driver supports. `ReadWriteMany` needs a shared-filesystem driver (EFS, Filestore, Azure Files); `ReadOnlyMany` is many readers; `ReadWriteOncePod` is exactly one pod, the strictest isolation
- **`spec.storage_class_name`** / **`spec.disable_dynamic_provisioning`**: Class selection (see above)
- **`spec.volume_mode`**: `filesystem` (the default, what nearly every workload wants) or `block` (a raw block device, for workloads that manage their own on-disk format; requires driver support)
- **`spec.storage_limit`**: Upper bound on the volume's size. Rarely needed — only meaningful to drivers that honor limits
- **`spec.volume_name`**: Binds to one specific pre-provisioned PersistentVolume by name — the adopt-existing-data path
- **`spec.selector`**: Label selector narrowing which pre-provisioned volumes may bind. Only meaningful for static binding
- **`spec.data_source`**: Populates the new volume from an existing source instead of provisioning it empty — clone a `persistent_volume_claim` or restore a `volume_snapshot`, same-namespace only. Requires a CSI driver that implements the operation. **Deploys via the Pulumi engine only** — the Terraform Kubernetes provider cannot express PVC data sources, and the Terraform module rejects the field at plan time
- **`spec.labels`** / **`spec.annotations`**: Merged with standard Planton labels for tracking and governance

## Outputs

After deployment, the following outputs are available in `status.outputs`:

- **`pvc_name`**: The name of the PersistentVolumeClaim object as created in the cluster — the value workload volume mounts reference as their claim name
- **`namespace`**: The namespace the claim was created in
- **`storage_request`**: The requested storage size, as a Kubernetes quantity

The outputs deliberately avoid bind-time status (bound volume name, phase) because a claim under a `wait_for_first_consumer` class is correctly Pending until a pod consumes it.

## Quick Start

Create a file `pvc.yaml` — the simplest useful claim, using the cluster's default StorageClass:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPersistentVolumeClaim
metadata:
  name: app-data
spec:
  namespace:
    value: backend
  name: app-data
  storage_request: 10Gi
```

Deploy:

```shell
planton apply -f pvc.yaml
```

A workload then mounts the claim by name in its volume mounts. If the default class binds on first consumer, the claim shows Pending until that workload's pod schedules — expected.

## How It Works

This kind includes both **Pulumi** (Go) and **Terraform** (HCL) modules that:

1. Resolve the target namespace (literal value or resolved reference; `default` when omitted)
2. Merge user labels and annotations with standard Planton tracking labels
3. Apply the Kubernetes default access mode (`ReadWriteOnce`) when the spec omits `access_modes`, and resolve `volume_mode` to its API string (`Filesystem` default) — both sent explicitly so the engines submit identical claims
4. Compute the `storageClassName` wire value: the resolved class name, the empty string when dynamic provisioning is disabled, or absent (cluster default) when the spec names nothing
5. Create the `core/v1` PersistentVolumeClaim with resources, static-binding fields, selector, and (Pulumi only) data source — **without awaiting Bound**
6. Export the claim name, namespace, and storage request for downstream composition

Both IaC implementations follow identical logic, with **one parity exception**: the Terraform Kubernetes provider's PVC resource cannot express `spec.dataSource` (clone a PVC / restore a VolumeSnapshot), so the Terraform module fails the plan with a precondition when `data_source` is set — failing loudly beats silently provisioning an empty volume where restored data was asked for. Claims with a data source deploy via the Pulumi provisioner.

## When to Use

Use **KubernetesPersistentVolumeClaim** when you need:

- A data volume for a Deployment or shared by several pods, with a lifecycle independent of any one workload
- A ReadWriteMany volume on a shared-filesystem driver
- Adopting a pre-provisioned volume (static binding via `volume_name` or a selector)
- Cloning an existing claim or restoring a snapshot into a new volume
- Pinning a claim to a platform-defined StorageClass by reference

**Do NOT use** when:

- You need per-replica storage for a StatefulSet — use the workload's `volume_claim_templates`; those claims are stamped and tracked by the StatefulSet controller
- You need scratch space that can vanish with the pod — use an emptyDir volume in the workload spec
- You need to share configuration or secrets — that is what ConfigMaps and Secrets are for

## Prerequisites

- **Kubernetes Cluster**: Access to a Kubernetes cluster
- **Credentials**: Kubernetes cluster credentials (kubeconfig)
- **Namespace**: The target namespace must exist before creating the claim (unless deploying to `default`, or creating the namespace in the same chart via a reference)
- **StorageClass**: A default class (for the omitted-class shape), the named class, or — for static binding — a matching pre-provisioned PersistentVolume
- **CSI capability**: `ReadWriteMany` needs a shared-filesystem driver; data sources need a driver implementing clone/restore

## Best Practices

1. **Name the class in production claims**: the cluster default is a convenience that can change under you; `storage_class_name` (ideally as a reference) pins the tier explicitly
2. **Treat Pending as the steady state for unconsumed claims**: under `wait_for_first_consumer`, the event `waiting for first consumer to be created before binding` is the design working, not a fault to debug
3. **Size for growth, on an expandable class**: expansion is grow-only and driver-dependent; a claim on a non-expandable class outgrowing its disk means a migration
4. **Reserve `ReadWriteMany` for drivers that mean it**: block-storage drivers (EBS, PD, Azure Disk) do not support it — the claim will never provision; use EFS/Filestore/Azure Files class equivalents
5. **Keep StatefulSet storage in the StatefulSet**: standalone claims mounted by StatefulSets forfeit the controller's per-replica stamping and lifecycle tracking

## References

- [Kubernetes Persistent Volumes Documentation](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)
- [PersistentVolumeClaim API Reference](https://kubernetes.io/docs/reference/kubernetes-api/config-and-storage-resources/persistent-volume-claim-v1/)
- [Configure a Pod to Use a PersistentVolume](https://kubernetes.io/docs/tasks/configure-pod-container/configure-persistent-volume-storage/)
- [CSI Volume Cloning](https://kubernetes.io/docs/concepts/storage/volume-pvc-datasource/)
- [Volume Snapshots](https://kubernetes.io/docs/concepts/storage/volume-snapshots/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

# Kubernetes PersistentVolumeClaim

Deploys a Kubernetes PersistentVolumeClaim — the durable-disk primitive. A claim names how much storage it needs, how it will be accessed, and which StorageClass provisions it; the cluster binds it to a PersistentVolume that satisfies the request. The spec covers the complete core/v1 claim surface, including static binding to pre-provisioned volumes and populating a new volume from a clone or snapshot data source.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes PersistentVolumeClaim** -- a core/v1 claim in the specified namespace carrying capacity, access modes, storage-class selection, optional static binding, and an optional clone/snapshot data source
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- The target namespace must already exist (the module does not create it).
- The chosen StorageClass (or the cluster default) must exist; clone/snapshot data sources need a CSI driver that implements the operation.

## Deploy

### Console

Open the deployment store, find **Kubernetes PersistentVolumeClaim**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Dynamic Default Class** preset for the common case or **Shared RWM** for multi-pod filesystems in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPersistentVolumeClaim
metadata:
  name: shared-uploads
  org: acme-corp
  env: prod
spec:
  name: shared-uploads
  namespace:
    value: backend-services
  storageRequest: 10Gi
```

```shell
planton apply -f pvc.yaml
```

This requests a 10Gi ReadWriteOnce volume through the cluster's default StorageClass, ready for workloads to mount by name. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to place the claim in a managed namespace and pin its class to a composed StorageClass:

```yaml
spec:
  name: shared-uploads
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: backend-services
      fieldPath: spec.name
  storageRequest: 10Gi
  storageClassName:
    valueFrom:
      kind: KubernetesStorageClass
      name: fast-ssd
      fieldPath: status.outputs.storage_class_name
```

The InfraPipeline creates the namespace and StorageClass first, then the claim against them — workloads in the same chart mount it by name.

## Key Configuration

These are the most important decisions when configuring a Kubernetes PersistentVolumeClaim. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Size for growth** -- The request is a Kubernetes quantity ("10Gi", "500Mi"). Growing later requires the claim's StorageClass to allow expansion, and Kubernetes NEVER shrinks a volume — the only way down is a data migration.

**Access modes are driver promises** -- ReadWriteOnce (one node, the default) works everywhere; ReadWriteMany needs a shared-filesystem driver (EFS, Filestore, Azure Files); ReadWriteOncePod is the strictest isolation. A mode the backing storage cannot deliver leaves the claim Pending.

**Absent and empty class names differ** -- Omitting the class uses the cluster DEFAULT; `disableDynamicProvisioning` pins it to the EMPTY string, meaning no volume is ever provisioned — the claim binds only to a matching pre-provisioned PersistentVolume. The two are mutually exclusive by design.

**Pending can be correct** -- With a WaitForFirstConsumer class (the norm for zonal cloud disks), a claim stays Pending until a pod uses it. Deploys never block on the claim reaching Bound — deliberate, so they never hang waiting for a consumer that arrives later.

**Data sources have driver rules** -- Cloning runs as snapshot-then-restore on AWS EBS (minutes even for small volumes), and EC2 refuses to clone an UNENCRYPTED source; snapshots restore faster with no encryption constraint. Sources are same-namespace only.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** (optional) | `namespace` | `spec.name` |
| **KubernetesStorageClass** (optional) | `storageClassName` | `status.outputs.storage_class_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `pvc_name` | The name of the created claim | Workloads' PVC volume mounts |
| `namespace` | The namespace where the claim was created | Verifying consumer co-location (mounting is namespace-local) |
| `storage_request` | The requested capacity | Capacity audits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Default-class app storage** -- One claim, cluster default class, ReadWriteOnce. Start from the **Dynamic Default Class** preset.

**Pinned performance tier** -- A named class (SSD, encrypted) chosen deliberately. Start from the **Pinned Class** preset.

**Shared filesystem** -- ReadWriteMany through a shared-filesystem class for multi-pod access. Start from the **Shared RWM** preset.

**Environment stamping** -- A new claim born from an existing claim's data. Start from the **Clone From Claim** preset.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- reference the namespace so infra charts create it and this claim in dependency order
- [**Kubernetes StorageClass**](/infra-catalog/kubernetes-storage-class) -- reference a class created on this platform to pin the performance tier declaratively
- [**Kubernetes Deployment**](/infra-catalog/kubernetes-deployment) -- mounts the claim by name via its volume mounts, from the same namespace only; per-replica StatefulSet storage uses the workload's own volume claim templates instead

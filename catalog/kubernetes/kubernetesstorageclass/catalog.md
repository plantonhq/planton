# Kubernetes StorageClass

Deploys a cluster-scoped Kubernetes StorageClass — an entry on the cluster's storage menu. A class names the CSI provisioner that creates volumes, the provisioner-specific parameters (disk type, IOPS, encryption), and the lifecycle policies for the volumes it provisions.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes StorageClass** -- a storage.k8s.io/v1 StorageClass carrying the provisioner, parameters, reclaim and binding policies, expandability, topology restrictions, and the optional default-class annotation
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **The CSI driver must be installed** -- a class naming an absent provisioner leaves every claim of that class Pending forever.
- When claiming the cluster default, demote the managed cluster's built-in default first (EKS ships gp2/gp3, GKE standard-rwo, AKS default).

## Deploy

### Console

Open the deployment store, find **Kubernetes StorageClass**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **AWS EBS gp3** preset (or its GCP/Azure siblings) in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesStorageClass
metadata:
  name: fast-ssd
  org: acme-corp
  env: prod
spec:
  name: fast-ssd
  provisioner: ebs.csi.aws.com
  parameters:
    type: gp3
    encrypted: "true"
  volumeBindingMode: wait_for_first_consumer
  allowVolumeExpansion: true
```

```shell
planton apply -f storageclass.yaml
```

This creates an encrypted gp3 class with zonal-correct binding and expansion enabled — claims reference it via `storageClassName: fast-ssd`. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a Kubernetes StorageClass. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Parameters speak the driver's language** -- The map passes through verbatim: EBS CSI takes `type: gp3, iops: "6000", encrypted: "true"`; GCE PD CSI takes `type: pd-ssd`; Azure Disk CSI takes `skuName: Premium_LRS`. Invalid keys surface as provisioning failures on the first claim. Values naming secrets are Secret references, never material.

**Most of the class is frozen** -- Provisioner, parameters, reclaim policy, and binding mode are immutable at the API; editing them replaces the class. Volumes already provisioned are untouched by a replacement — only future claims see new settings.

**Binding mode decides zone correctness** -- Immediate (the default) provisions on claim creation; on multi-zone clusters the disk may land where the pod cannot schedule. WaitForFirstConsumer provisions where the pod actually lands — the right choice for zonal storage, and its claims staying Pending until consumed is correct behavior.

**Reclaim fate** -- Delete (the default) removes the backing disk with its claim; Retain parks it (data and bill intact) for manual reclamation. The policy stamps at PROVISION time.

**Enable expansion** -- It costs nothing on drivers that support it and turns "the database outgrew its disk" into a one-field claim edit instead of a migration.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign-key dependencies -- it is cluster-scoped and references nothing.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `storage_class_name` | The name of the created class | PersistentVolumeClaims' `storageClassName` (the Kubernetes PersistentVolumeClaim kind references it directly) |
| `provisioner` | The CSI driver behind the class | Auditing the storage menu |
| `is_default_class` | Whether this class is the cluster default | Auditing default posture |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Tuned cloud SSD tiers** -- Encrypted, expandable, zonal-correct classes per cloud. Start from the **AWS EBS gp3**, **GCP PD SSD**, or **Azure Premium** presets.

## Works With

- [**Kubernetes PersistentVolumeClaim**](/infra-catalog/kubernetes-persistent-volume-claim) -- claims order from the menu by this class's name (and can reference it declaratively on this platform).
- [**Kubernetes StatefulSet**](/infra-catalog/kubernetes-stateful-set) -- volume claim templates name the class for per-replica storage.

# Kubernetes Storage Class

## Overview

**KubernetesStorageClass** is a catalog kind that creates and manages Kubernetes StorageClasses — the cluster's storage menu — as first-class, declaratively managed resources. A StorageClass names a provisioner (the CSI driver that creates volumes), the provisioner-specific parameters (disk type, IOPS, encryption, filesystem), and the lifecycle policies for the volumes it provisions (reclaim, binding timing, expandability). PersistentVolumeClaims then request the class by name.

The kind covers the complete `storage.k8s.io/v1` StorageClass surface: provisioner, parameters, reclaim policy, volume binding mode, volume expansion, mount options, topology restrictions, and the default-class marker. There is nothing an upstream StorageClass can express that this spec cannot.

## Purpose

Managed clusters ship default classes (EKS: gp2/gp3 via `ebs.csi.aws.com`, GKE: standard-rwo/premium-rwo via `pd.csi.storage.gke.io`, AKS: default/managed-csi via `disk.csi.azure.com`), and a custom class is how a platform pins performance characteristics — SSD-backed, expandable, topology-constrained — instead of inheriting whatever the cluster default happens to be.

**Key value over raw manifests:**

- **Schema-level validation**: DNS-subdomain name checks, enum-constrained reclaim and binding policies, and topology-term shape contracts — all caught before anything reaches the cluster
- **Default-class as a first-class field**: `is_default_class` renders the `storageclass.kubernetes.io/is-default-class` annotation, so the upstream mechanism is a typed boolean instead of a magic string
- **Deterministic policies**: Both IaC modules always submit `reclaimPolicy` and `volumeBindingMode` explicitly (applying the Kubernetes defaults when the spec omits them), so the deployed object never depends on which engine applied it
- **Safe replacement on immutable fields**: `provisioner` and `parameters` are immutable upstream; both engines replace the class (delete-before-create, since StorageClass names are cluster-unique) instead of failing an in-place update
- **Dual IaC support**: Both Pulumi and Terraform implementations with feature parity (one documented exception, below)
- **Lifecycle management**: Integrated with Planton's deployment lifecycle for status tracking and outputs

## Cluster-Scoped, Provisioner-Owned

Two structural facts shape everything about StorageClasses:

- **StorageClasses are cluster-scoped.** One class serves claims from every namespace; there is no `spec.namespace`.
- **The parameters vocabulary belongs to the provisioner, not to Kubernetes.** `parameters` is a plain string map passed through to the CSI driver (e.g. EBS CSI: `{type: gp3, iops: "6000", encrypted: "true"}`; GCE PD CSI: `{type: pd-ssd}`; Azure Disk CSI: `{skuName: Premium_LRS}`). Kubernetes does not validate them at class creation — they are consumed at provision time, so a typo surfaces as a provisioning failure on the first claim, not at apply.

## Binding Timing: `wait_for_first_consumer`

`volume_binding_mode` decides when a claim of this class is bound and its volume provisioned:

- **`immediate`** (the Kubernetes default): bind and provision as soon as the claim is created. Simple, but on multi-zone clusters the volume may land in a zone where the consuming pod cannot schedule.
- **`wait_for_first_consumer`**: wait until a pod actually uses the claim, then provision in the topology that pod scheduled into. This is the right choice for zonal storage (EBS, GCE PD, Azure Disk) — and it means **a claim of this class stays Pending until a consumer arrives, which is correct behavior, not an error.**

`allowed_topologies` (restricting which zones volumes may provision in) is only honored with `wait_for_first_consumer` binding — immediate binding cannot see pod topology, so the combination is meaningless.

## The Default Class

`is_default_class: true` marks this class as the one claims receive when they name none, rendered as the `storageclass.kubernetes.io/is-default-class: "true"` annotation. A cluster should have exactly **one** default: setting this alongside an existing default (e.g. the managed cluster's built-in class) makes newest-wins behavior undefined across Kubernetes versions — demote the old default first.

## Essential Configuration Fields

### Required

- **`spec.name`**: The StorageClass name (DNS subdomain: lowercase alphanumeric, hyphens, dots, 1–253 chars) — the value claims put in `storage_class_name`
- **`spec.provisioner`**: The CSI driver (or in-tree plugin) that provisions volumes — e.g. `ebs.csi.aws.com`, `pd.csi.storage.gke.io`, `disk.csi.azure.com`, `rancher.io/local-path`. IMMUTABLE after creation; changing it replaces the class

### Common

- **`spec.parameters`**: Provisioner-specific parameters, passed through verbatim. IMMUTABLE after creation. Values that name secrets (e.g. `csi.storage.k8s.io/provisioner-secret-name`) are references to Secret objects, never secret material itself
- **`spec.reclaim_policy`**: What happens to a dynamically provisioned volume when its claim is deleted — `delete` (the default, right for disposable data) or `retain` (the volume survives and must be reclaimed manually)
- **`spec.volume_binding_mode`**: `immediate` (default) or `wait_for_first_consumer` (the right choice for zonal storage)
- **`spec.allow_volume_expansion`**: Lets claims of this class be resized after creation (grow only — Kubernetes never shrinks). Requires driver support; enabling it costs nothing on drivers that support it
- **`spec.mount_options`**: Mount options for volumes of this class (e.g. `["noatime"]`). NOT validated by Kubernetes — an invalid option fails at pod start
- **`spec.allowed_topologies`**: Zone/topology restrictions; each term's requirements AND together, multiple terms OR
- **`spec.is_default_class`**: Marks the cluster default (see above)
- **`spec.labels`** / **`spec.annotations`**: Merged with standard Planton labels for tracking and governance. The default-class annotation is managed by `is_default_class` — set that field instead of adding the annotation here

## Outputs

After deployment, the following outputs are available in `status.outputs`:

- **`storage_class_name`**: The name of the StorageClass object as created in the cluster — the value claims put in their `storage_class_name`
- **`provisioner`**: The provisioner (CSI driver) backing this class
- **`is_default_class`**: Whether this class is annotated as the cluster's default StorageClass

## Quick Start

Create a file `storage-class.yaml` — an expandable gp3 class for EKS:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesStorageClass
metadata:
  name: fast-ssd
spec:
  name: fast-ssd
  provisioner: ebs.csi.aws.com
  parameters:
    type: gp3
    encrypted: "true"
  volume_binding_mode: wait_for_first_consumer
  allow_volume_expansion: true
```

Deploy:

```shell
planton apply -f storage-class.yaml
```

Claims then request the class by name (`storage_class_name: {value: fast-ssd}`), or by referencing this resource so the name resolves through the deployment graph.

## How It Works

This kind includes both **Pulumi** (Go) and **Terraform** (HCL) modules that:

1. Merge user labels and annotations with standard Planton tracking labels
2. Render the default-class marker: `is_default_class: true` becomes the `storageclass.kubernetes.io/is-default-class: "true"` annotation
3. Resolve `reclaim_policy` and `volume_binding_mode` to their Kubernetes API strings, applying the API server's own defaults (`Delete`, `Immediate`) when the spec omits them — and always submit them explicitly
4. Create the `storage.k8s.io/v1` StorageClass with the provisioner, parameters, expansion flag, mount options, and topology terms
5. Export the class name, provisioner, and default-class flag for downstream composition

Both IaC implementations follow identical logic, with **one parity exception**: the Terraform Kubernetes provider models `allowed_topologies` as a single selector term, so the Terraform module rejects a spec with multiple terms at plan time (deploy such a class with the Pulumi provisioner, or combine the zone values into one term — values within one requirement already OR together). The Pulumi module sends every term.

## When to Use

Use **KubernetesStorageClass** when you need:

- A storage tier with pinned performance characteristics (SSD type, IOPS, throughput, encryption) instead of the cluster default
- `wait_for_first_consumer` binding for zonal cloud disks, so volumes provision where pods schedule
- Volume expansion enabled before a database outgrows its disk
- A `retain` reclaim policy for data that must survive claim deletion
- Topology-constrained provisioning (pin volumes to specific zones)
- Declaring or replacing the cluster's default StorageClass

**Do NOT use** when:

- The cluster's built-in classes already match your requirements — inheriting a managed default is fine until it isn't
- You need per-workload storage requests — that is the PersistentVolumeClaim's job; the class defines the tier, the claim requests from it
- You need to change `provisioner` or `parameters` on volumes that already exist — those fields are immutable, and replacing the class does not migrate existing volumes

## Prerequisites

- **Kubernetes Cluster**: Access to a Kubernetes cluster
- **Credentials**: Kubernetes cluster credentials (kubeconfig)
- **CSI Driver**: The named provisioner must be installed in the cluster for claims of this class to actually provision — the class object itself is creatable regardless (parameters are consumed at provision time)

## Best Practices

1. **Use `wait_for_first_consumer` for zonal storage**: EBS, GCE PD, and Azure Disk volumes are zone-bound; immediate binding can strand a volume in a zone where the pod cannot schedule
2. **Enable expansion up front**: `allow_volume_expansion: true` costs nothing on drivers that support it and saves a migration later — the field is mutable, but the habit is cheaper than the incident
3. **Treat `provisioner` and `parameters` as permanent**: changing either replaces the class; existing volumes keep their original settings. Name classes for what they provide (`fast-ssd`, `archive-hdd`), not for today's driver
4. **Keep exactly one default class**: demote the old default before promoting a new one — two defaults make claim behavior undefined
5. **Prefer `retain` only where you mean it**: retained volumes accrue cost until manually reclaimed; `delete` is the right default for anything recreatable

## References

- [Kubernetes Storage Classes Documentation](https://kubernetes.io/docs/concepts/storage/storage-classes/)
- [StorageClass API Reference](https://kubernetes.io/docs/reference/kubernetes-api/config-and-storage-resources/storage-class-v1/)
- [Dynamic Volume Provisioning](https://kubernetes.io/docs/concepts/storage/dynamic-provisioning/)
- [Volume Expansion](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#expanding-persistent-volumes-claims)
- [AWS EBS CSI Driver Parameters](https://github.com/kubernetes-sigs/aws-ebs-csi-driver/blob/master/docs/parameters.md)
- [GCE PD CSI Driver](https://github.com/kubernetes-sigs/gcp-compute-persistent-disk-csi-driver)
- [Azure Disk CSI Driver Parameters](https://github.com/kubernetes-sigs/azuredisk-csi-driver/blob/master/docs/driver-parameters.md)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

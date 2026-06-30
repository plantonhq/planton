---
title: "Volume"
description: "Volume deployment documentation"
icon: "package"
order: 100
componentName: "civovolume"
---

# Civo Volume

Deploys a Civo block storage volume that can be attached to compute instances for persistent, expandable storage. The component provisions a volume in a specified region and size, and exports its identifier for use by other components such as CivoComputeInstance.

## What Gets Created

When you deploy a CivoVolume resource, Planton provisions:

- **Block Storage Volume** — a `civo_volume` resource created in the target region with the requested capacity in GiB

## Prerequisites

- **Civo credentials** configured via environment variables or Planton provider config
- **A target region** — the volume must be created in the same region as any instance that will attach to it

## Quick Start

Create a file `civo-volume.yaml`:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoVolume
metadata:
  name: my-volume
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.CivoVolume.my-volume
spec:
  volumeName: my-volume
  region: nyc1
  sizeGib: 50
```

Deploy:

```shell
planton apply -f civo-volume.yaml
```

This creates a 50 GiB unformatted block storage volume in the New York region.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `volumeName` | `string` | Name of the volume. Must start with a lowercase letter, end with a letter or number, and contain only lowercase letters, numbers, and hyphens. | Required, 1-64 characters, pattern: `^[a-z]([a-z0-9-]*[a-z0-9])?$` |
| `region` | `enum` | Civo region where the volume is created. Valid values: `lon1`, `lon2`, `fra1`, `nyc1`, `phx1`, `mum1`. | Required |
| `sizeGib` | `uint32` | Size of the volume in GiB. | Required, 1-16000 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `filesystemType` | `enum` | `unformatted` | Initial filesystem to format the volume with. Valid values: `unformatted`, `ext4`, `xfs`. **Note:** the upstream Civo provider does not currently expose filesystem formatting. The volume is created unformatted regardless of this setting. Use cloud-init or configuration management to format the volume after attachment. |
| `snapshotId` | `string` | `""` | Snapshot ID to create the volume from. **Note:** snapshot-based creation is not currently supported on public Civo cloud. This field is reserved for CivoStack (private cloud) deployments or future provider support. The volume is created empty when this field is set. |
| `tags` | `string[]` | `[]` | Tags for organizational purposes. Each tag must be unique, at most 64 characters, and match `^[A-Za-z0-9:_-]+$`. **Note:** the upstream Civo Volume provider does not currently apply tags to the cloud resource. Tags are recorded in Planton metadata only. |

## Examples

### Basic Volume

A minimal 10 GiB volume for development use:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoVolume
metadata:
  name: dev-data
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.CivoVolume.dev-data
spec:
  volumeName: dev-data
  region: fra1
  sizeGib: 10
```

### Application Data Volume

A larger volume in the London region with tags and a requested filesystem type:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoVolume
metadata:
  name: app-storage
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.CivoVolume.app-storage
spec:
  volumeName: app-storage
  region: lon1
  sizeGib: 200
  filesystemType: ext4
  tags:
    - environment:staging
    - team:platform
```

### Large Production Volume

A high-capacity volume for production workloads:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoVolume
metadata:
  name: prod-db-vol
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.CivoVolume.prod-db-vol
spec:
  volumeName: prod-db-vol
  region: nyc1
  sizeGib: 2000
  filesystemType: xfs
  tags:
    - environment:production
    - service:database
    - managed-by:planton
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `volumeId` | `string` | Unique identifier of the created Civo volume |
| `attachedInstanceId` | `string` | ID of the Civo instance the volume is attached to. Empty until the volume is attached to an instance via a separate attachment resource or Kubernetes CSI driver. |
| `devicePath` | `string` | Device path of the volume on the attached instance. Empty until the volume is attached. |

## Related Components

- [CivoComputeInstance](/docs/catalog/civo/compute-instance) — attach the volume to a compute instance for persistent storage
- [CivoKubernetesCluster](/docs/catalog/civo/kubernetes-cluster) — use the volume as persistent storage for Kubernetes workloads via the Civo CSI driver
- [CivoVpc](/docs/catalog/civo/vpc) — provides the network in which the attached instance and volume reside
- [CivoFirewall](/docs/catalog/civo/firewall) — controls network access to instances using the volume

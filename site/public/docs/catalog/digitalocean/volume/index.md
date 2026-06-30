---
title: "Volume"
description: "Volume deployment documentation"
icon: "package"
order: 100
componentName: "digitaloceanvolume"
---

# DigitalOcean Volume

Deploys a DigitalOcean block storage volume that provides persistent, network-attached storage attachable to Droplets. The component handles volume creation, optional filesystem pre-formatting, snapshot-based provisioning, and tag management, exposing the volume UUID as a stack output.

## What Gets Created

When you deploy a DigitalOceanVolume resource, Planton provisions:

- **Block Storage Volume** — a `digitalocean_volume` resource in the specified region with the given size, optional filesystem formatting, and tags

## Prerequisites

- **DigitalOcean credentials** configured via environment variables or Planton provider config
- **A target region** that matches the region of any Droplet you intend to attach the volume to (volumes and Droplets must be co-located)
- **A volume snapshot ID** if creating the volume from an existing snapshot (optional)

## Quick Start

Create a file `volume.yaml`:

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanVolume
metadata:
  name: my-volume
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.DigitalOceanVolume.my-volume
spec:
  volumeName: my-volume
  region: nyc3
  sizeGib: 50
  filesystemType: ext4
```

Deploy:

```shell
planton apply -f volume.yaml
```

This creates a 50 GiB block storage volume in the NYC3 region, pre-formatted with ext4.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `volumeName` | `string` | Name of the volume in DigitalOcean. Lowercase letters, numbers, and hyphens only; must start with a letter and end with a letter or number. | Required, 1–64 characters, pattern: `^[a-z]([a-z0-9-]*[a-z0-9])?$` |
| `region` | `enum` | DigitalOcean region where the volume will be created. Valid values: `nyc3`, `sfo3`, `fra1`, `sgp1`, `lon1`, `tor1`, `blr1`, `ams3`. | Required |
| `sizeGib` | `uint32` | Size of the volume in GiB. | Required, 1–16000 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | `""` | Free-form description of the volume. Maximum 100 characters. |
| `filesystemType` | `enum` | `unformatted` | Initial filesystem to format the volume with. Valid values: `unformatted` (no formatting), `ext4`, `xfs`. Setting `ext4` or `xfs` eliminates the need for manual formatting after attaching to a Droplet. |
| `snapshotId` | `string` | `""` | UUID of an existing volume snapshot to create this volume from. When provided, the volume inherits the snapshot's data; `sizeGib` must be at least the snapshot's size. |
| `tags` | `repeated string` | `[]` | Tags to apply to the volume for organization and cost allocation. Each tag must be unique and consist of letters, numbers, colons, dashes, or underscores (max 64 characters per tag). |

## Examples

### Development Test Volume

A small, unformatted volume for experimentation:

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanVolume
metadata:
  name: dev-scratch
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.DigitalOceanVolume.dev-scratch
spec:
  volumeName: dev-scratch
  region: sfo3
  sizeGib: 10
  tags:
    - "env:dev"
    - "owner:testing"
```

### Database Storage with XFS

A pre-formatted XFS volume sized for a staging database. XFS handles large files and concurrent I/O well, making it a common choice for PostgreSQL or MySQL data directories:

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanVolume
metadata:
  name: staging-db-data
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.DigitalOceanVolume.staging-db-data
spec:
  volumeName: staging-db-data
  description: "Staging PostgreSQL data directory"
  region: fra1
  sizeGib: 200
  filesystemType: xfs
  tags:
    - "env:staging"
    - "service:postgres"
```

### Production Volume from Snapshot

A production volume restored from an existing snapshot, with ext4 formatting and compliance tags:

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanVolume
metadata:
  name: prod-app-data
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.DigitalOceanVolume.prod-app-data
spec:
  volumeName: prod-app-data
  description: "Production app data restored from nightly snapshot"
  region: nyc3
  sizeGib: 500
  filesystemType: ext4
  snapshotId: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  tags:
    - "env:prod"
    - "service:app"
    - "compliance:soc2"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `volumeId` | `string` | UUID of the created DigitalOcean block storage volume |

## Related Components

- [DigitalOceanDroplet](/docs/catalog/digitalocean/droplet) — the compute instance a volume attaches to; volume and Droplet must share the same region
- [DigitalOceanVpc](/docs/catalog/digitalocean/vpc) — provides the private network for Droplets that use attached volumes
- [DigitalOceanFirewall](/docs/catalog/digitalocean/firewall) — controls network access to Droplets with attached volumes

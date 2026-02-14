---
title: "Droplet"
description: "Droplet deployment documentation"
icon: "package"
order: 100
componentName: "digitaloceandroplet"
---

# DigitalOcean Droplet

Deploys a DigitalOcean Droplet (Linux virtual machine) with configurable size, base image, VPC placement, block storage attachments, and cloud-init user data. The component gives you full root-level control over the VM while managing provisioning through a declarative manifest.

## What Gets Created

When you deploy a DigitalOceanDroplet resource, OpenMCF provisions:

- **Droplet** — a `digitalocean_droplet` resource with the specified region, size slug, base image, VPC assignment, and optional features (IPv6, backups, monitoring agent)
- **Volume Attachments** — existing block storage volumes are attached to the Droplet when `volumeIds` is specified
- **Tags** — DigitalOcean tags are applied to the Droplet for organization and Cloud Firewall integration

## Prerequisites

- **DigitalOcean credentials** configured via environment variables or OpenMCF provider config
- **A DigitalOcean VPC** in the target region (can reference a DigitalOceanVpc resource via `valueFrom`)
- **A valid size slug** accepted by the DigitalOcean `/v2/sizes` API (e.g., `s-1vcpu-1gb`, `s-2vcpu-4gb`)
- **A valid image slug or snapshot ID** for the base OS (e.g., `ubuntu-22-04-x64`, `debian-12-x64`)

## Quick Start

Create a file `droplet.yaml`:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDroplet
metadata:
  name: my-droplet
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.DigitalOceanDroplet.my-droplet
spec:
  dropletName: my-droplet
  region: nyc3
  size: s-1vcpu-1gb
  image: ubuntu-22-04-x64
  vpc:
    value: "vpc-uuid-here"
```

Deploy:

```shell
openmcf apply -f droplet.yaml
```

This creates a single-vCPU Droplet running Ubuntu 22.04 in NYC3 with monitoring enabled by default.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `dropletName` | `string` | Hostname for the Droplet in DigitalOcean. Must be DNS-compatible. | Required, lowercase alphanumeric and hyphens, max 63 characters, pattern `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` |
| `region` | `enum` | DigitalOcean datacenter region. Valid values: `nyc3`, `sfo3`, `fra1`, `sgp1`, `lon1`, `tor1`, `blr1`, `ams3`. | Required |
| `size` | `string` | Droplet size slug determining CPU and memory allocation (e.g., `s-2vcpu-4gb`, `g-8vcpu-32gb`). Must match a slug from the DigitalOcean `/v2/sizes` API. | Required, pattern `^[a-z0-9]+(-[a-z0-9]+)+$` |
| `image` | `string` | Base image slug (e.g., `ubuntu-22-04-x64`) or custom snapshot ID. | Required, pattern `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` |
| `vpc` | `StringValueOrRef` | VPC UUID where the Droplet resides. Can reference a DigitalOceanVpc resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enableIpv6` | `bool` | `false` | Enables IPv6 networking on the Droplet. |
| `enableBackups` | `bool` | `false` | Enables automated daily backups. Recommended for production workloads. |
| `disableMonitoring` | `bool` | `false` | When `true`, disables the DigitalOcean monitoring agent. Monitoring is enabled by default because it is free and provides CPU, memory, disk, and network metrics. |
| `volumeIds` | `StringValueOrRef[]` | `[]` | Block storage volume IDs to attach to the Droplet. Volumes must reside in the same region. Can reference DigitalOceanVolume resources via `valueFrom`. |
| `tags` | `string[]` | `[]` | Tags applied to the Droplet for organization and Cloud Firewall integration. Must be unique. |
| `userData` | `string` | `""` | Cloud-init script executed on first boot. Maximum size is 32 KiB. Accepts both shell scripts and cloud-config YAML. |
| `timezone` | `enum` | `utc` | Timezone for the Droplet's system clock. Valid values: `utc`, `local`. |

## Examples

### Development Server

A minimal Droplet for development and testing:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDroplet
metadata:
  name: dev-server
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.DigitalOceanDroplet.dev-server
spec:
  dropletName: dev-server
  region: sfo3
  size: s-1vcpu-2gb
  image: ubuntu-24-04-x64
  vpc:
    value: "vpc-dev-uuid"
  tags:
    - dev
```

### Web Server with Cloud-Init and Backups

A staging web server that installs nginx on first boot, enables backups, and uses tags for Cloud Firewall integration:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDroplet
metadata:
  name: staging-web
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.DigitalOceanDroplet.staging-web
spec:
  dropletName: staging-web
  region: fra1
  size: s-2vcpu-4gb
  image: ubuntu-22-04-x64
  vpc:
    value: "vpc-staging-uuid"
  enableBackups: true
  enableIpv6: true
  tags:
    - staging
    - web
    - http-firewall
  userData: |
    #cloud-config
    package_update: true
    packages:
      - nginx
      - fail2ban
    runcmd:
      - systemctl enable nginx
      - systemctl start nginx
```

### Production Database with Attached Volume and VPC Reference

A production Droplet with an attached block storage volume for persistent data, referencing a DigitalOceanVpc and DigitalOceanVolume by name:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDroplet
metadata:
  name: prod-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.DigitalOceanDroplet.prod-db
spec:
  dropletName: prod-db
  region: nyc3
  size: g-4vcpu-16gb
  image: ubuntu-22-04-x64
  vpc:
    valueFrom:
      kind: DigitalOceanVpc
      name: prod-vpc
      fieldPath: status.outputs.vpc_id
  enableBackups: true
  volumeIds:
    - valueFrom:
        kind: DigitalOceanVolume
        name: prod-db-data
        fieldPath: status.outputs.volume_id
  tags:
    - production
    - database
    - db-firewall
  userData: |
    #!/bin/bash
    apt-get update
    apt-get install -y postgresql-16
    systemctl enable postgresql
    systemctl start postgresql
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `droplet_id` | `string` | Unique identifier of the created Droplet in DigitalOcean |
| `ipv4_address` | `string` | Primary IPv4 address (public if available, otherwise private) |
| `ipv6_address` | `string` | IPv6 address of the Droplet (empty if IPv6 was not enabled) |
| `image_id` | `int64` | Image ID of the Droplet's base image |
| `vpc_uuid` | `string` | UUID of the VPC network the Droplet resides in |

## Related Components

- [DigitalOceanVpc](/docs/catalog/digitalocean/digitaloceanvpc) — provides the VPC for Droplet network placement
- [DigitalOceanVolume](/docs/catalog/digitalocean/digitaloceanvolume) — provisions block storage volumes for persistent data
- [DigitalOceanFirewall](/docs/catalog/digitalocean/digitaloceanfirewall) — controls network access to the Droplet via tag-based rules
- [DigitalOceanLoadBalancer](/docs/catalog/digitalocean/digitaloceanloadbalancer) — distributes traffic across Droplets matched by tag
- [DigitalOceanDnsRecord](/docs/catalog/digitalocean/digitaloceandnsrecord) — maps DNS names to Droplet IP addresses

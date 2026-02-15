---
title: "Compute Instance"
description: "Compute Instance deployment documentation"
icon: "package"
order: 100
componentName: "gcpcomputeinstance"
---

# GCP Compute Instance

Deploys a Google Compute Engine VM instance with configurable machine type, boot disk, network interfaces, optional attached disks, service accounts, scheduling policies, and startup scripts. The component provisions a single managed instance in a specified zone and project, applying GCP-managed labels automatically.

## What Gets Created

When you deploy a GcpComputeInstance resource, OpenMCF provisions:

- **Compute Engine Instance** — a `google_compute_instance` with the specified machine type, boot disk image, and zone, placed in the target GCP project
- **Boot Disk** — created from the specified source image with configurable size, type, and auto-delete behavior
- **Network Interfaces** — one or more interfaces attached to a VPC network or subnetwork, with optional external IP access configs and alias IP ranges
- **Attached Disks** — additional persistent disks attached to the instance, created only when `attachedDisks` entries are provided
- **Service Account Binding** — applied only when `serviceAccount` is specified, granting the instance the configured OAuth scopes
- **Scheduling Configuration** — configured only when `scheduling` is specified or when `spot`/`preemptible` is set, controlling preemption behavior, automatic restart, and host maintenance policy
- **GCP Labels** — system labels (resource kind, name, organization, environment) merged with any user-provided labels

## Prerequisites

- **GCP credentials** configured via environment variables or OpenMCF provider config
- **A GCP project** where the Compute Engine instance will be created
- **Compute Engine API** enabled in the target project
- **A VPC network or subnetwork** for network interface configuration

## Quick Start

Create a file `compute-instance.yaml`:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpComputeInstance
metadata:
  name: my-vm
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpComputeInstance.my-vm
spec:
  projectId: my-gcp-project
  zone: us-central1-a
  machineType: e2-medium
  bootDisk:
    image: debian-cloud/debian-12
  networkInterfaces:
    - network: default
      accessConfigs:
        - {}
```

Deploy:

```shell
openmcf apply -f compute-instance.yaml
```

This creates an `e2-medium` instance running Debian 12 in `us-central1-a` with a public IP on the default VPC network.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `string` | GCP project ID where the instance is created. Can reference a GcpProject resource via `valueFrom`. | Required |
| `zone` | `string` | Zone for the instance (e.g., `us-central1-a`). | Pattern: `^[a-z]+-[a-z]+[0-9]-[a-z]$` |
| `machineType` | `string` | Machine type (e.g., `e2-medium`, `n1-standard-1`, `n2-standard-2`). | Minimum length: 1 |
| `bootDisk` | `object` | Boot disk configuration. See **Boot Disk** fields below. | Required |
| `bootDisk.image` | `string` | Source image (e.g., `debian-cloud/debian-12`, `ubuntu-os-cloud/ubuntu-2204-lts`). | Minimum length: 1 |
| `networkInterfaces` | `object[]` | Network interface configurations. At least one is required. | Minimum items: 1 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `bootDisk.sizeGb` | `int32` | `10` | Boot disk size in GB. Range: 10–65536. |
| `bootDisk.type` | `string` | `pd-standard` | Boot disk type: `pd-standard`, `pd-ssd`, or `pd-balanced`. |
| `bootDisk.autoDelete` | `bool` | `true` | Auto-delete the boot disk when the instance is deleted. |
| `networkInterfaces[].network` | `string` | — | VPC network for the interface. Can reference a GcpVpc resource via `valueFrom`. Either `network` or `subnetwork` is required. |
| `networkInterfaces[].subnetwork` | `string` | — | Subnetwork for the interface. Can reference a GcpSubnetwork resource via `valueFrom`. Either `network` or `subnetwork` is required. |
| `networkInterfaces[].accessConfigs` | `object[]` | `[]` | External IP access configurations. If empty, the interface has no external IP. |
| `networkInterfaces[].accessConfigs[].natIp` | `string` | — | Static external IP address. If omitted, an ephemeral IP is assigned. |
| `networkInterfaces[].accessConfigs[].networkTier` | `string` | — | Network tier: `PREMIUM` or `STANDARD`. |
| `networkInterfaces[].aliasIpRanges` | `object[]` | `[]` | Alias IP ranges for the interface. |
| `networkInterfaces[].aliasIpRanges[].ipCidrRange` | `string` | — | IP CIDR range for alias IPs. Required when specifying an alias IP range. |
| `networkInterfaces[].aliasIpRanges[].subnetworkRangeName` | `string` | — | Subnetwork secondary range name. |
| `attachedDisks` | `object[]` | `[]` | Additional persistent disks to attach. |
| `attachedDisks[].source` | `string` | — | Source disk self-link or name. Required per entry. |
| `attachedDisks[].deviceName` | `string` | `attached-disk-{index}` | Device name for the disk. |
| `attachedDisks[].mode` | `string` | `READ_WRITE` | Disk mode: `READ_WRITE` or `READ_ONLY`. |
| `serviceAccount` | `object` | — | Service account configuration. |
| `serviceAccount.email` | `string` | — | Service account email. Can reference a GcpServiceAccount resource via `valueFrom`. If omitted, the default Compute Engine service account is used. |
| `serviceAccount.scopes` | `string[]` | `["https://www.googleapis.com/auth/cloud-platform"]` | OAuth scopes for the service account. |
| `preemptible` | `bool` | `false` | Whether the instance is preemptible. Deprecated; use `spot` or `scheduling` instead. |
| `spot` | `bool` | `false` | Whether the instance is a Spot VM. |
| `deletionProtection` | `bool` | `false` | Prevents accidental deletion of the instance when enabled. |
| `metadata` | `map<string, string>` | `{}` | Custom metadata key-value pairs for the instance. |
| `labels` | `map<string, string>` | `{}` | Labels to apply to the instance, merged with system labels. |
| `tags` | `string[]` | `[]` | Network tags for firewall rule targeting. |
| `sshKeys` | `string[]` | `[]` | SSH keys in `username:ssh-key` format. |
| `startupScript` | `string` | — | Startup script to run when the instance boots. |
| `allowStoppingForUpdate` | `bool` | `true` | Allow the instance to be stopped for update operations. |
| `scheduling` | `object` | — | Scheduling configuration. Overrides `preemptible` and `spot` when set. |
| `scheduling.preemptible` | `bool` | `false` | Whether the instance is preemptible. |
| `scheduling.automaticRestart` | `bool` | `true` | Automatic restart on failure. |
| `scheduling.onHostMaintenance` | `string` | — | Host maintenance behavior: `MIGRATE` or `TERMINATE`. |
| `scheduling.provisioningModel` | `string` | — | Provisioning model: `STANDARD` or `SPOT`. |
| `scheduling.instanceTerminationAction` | `string` | — | Termination action for Spot VMs: `STOP` or `DELETE`. |

## Examples

### Basic Instance with Public IP

An `e2-small` instance running Ubuntu on the default network with an ephemeral external IP:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpComputeInstance
metadata:
  name: web-server
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpComputeInstance.web-server
spec:
  projectId: my-gcp-project
  zone: us-central1-a
  machineType: e2-small
  bootDisk:
    image: ubuntu-os-cloud/ubuntu-2204-lts
    sizeGb: 20
    type: pd-balanced
  networkInterfaces:
    - network: default
      accessConfigs:
        - {}
  tags:
    - http-server
    - https-server
```

### Private Instance with Startup Script and Service Account

An instance inside a custom VPC subnetwork with no external IP, a startup script that installs Nginx, and a dedicated service account:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpComputeInstance
metadata:
  name: app-backend
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.GcpComputeInstance.app-backend
spec:
  projectId: my-gcp-project
  zone: us-east1-b
  machineType: n2-standard-2
  bootDisk:
    image: debian-cloud/debian-12
    sizeGb: 30
    type: pd-ssd
  networkInterfaces:
    - subnetwork: projects/my-gcp-project/regions/us-east1/subnetworks/private-subnet
  serviceAccount:
    email: app-backend@my-gcp-project.iam.gserviceaccount.com
    scopes:
      - https://www.googleapis.com/auth/cloud-platform
  startupScript: |
    #!/bin/bash
    apt-get update && apt-get install -y nginx
    systemctl enable nginx && systemctl start nginx
  metadata:
    enable-oslogin: "TRUE"
  labels:
    team: backend
    tier: application
```

### Spot VM with Attached Disk and Foreign Key References

A cost-effective Spot VM with an additional data disk, referencing other OpenMCF-managed resources for the project, network, and service account:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpComputeInstance
metadata:
  name: batch-worker
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpComputeInstance.batch-worker
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      field: status.outputs.project_id
  zone: us-central1-a
  machineType: n1-standard-4
  bootDisk:
    image: debian-cloud/debian-12
    sizeGb: 20
    type: pd-ssd
  networkInterfaces:
    - subnetwork:
        valueFrom:
          kind: GcpSubnetwork
          name: worker-subnet
          field: status.outputs.subnetwork_self_link
      accessConfigs:
        - networkTier: STANDARD
  attachedDisks:
    - source: projects/my-gcp-project/zones/us-central1-a/disks/data-disk-1
      deviceName: data
      mode: READ_WRITE
  serviceAccount:
    email:
      valueFrom:
        kind: GcpServiceAccount
        name: batch-worker-sa
        field: status.outputs.email
    scopes:
      - https://www.googleapis.com/auth/cloud-platform
  scheduling:
    provisioningModel: SPOT
    instanceTerminationAction: STOP
    onHostMaintenance: TERMINATE
    automaticRestart: false
  deletionProtection: false
  allowStoppingForUpdate: true
  tags:
    - batch
    - internal
  sshKeys:
    - "deploy:ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIG... deploy@ci"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `instance_name` | `string` | Name of the Compute Engine instance |
| `instance_id` | `string` | Instance ID (unique numeric identifier) |
| `self_link` | `string` | GCP resource self link URL for the instance |
| `internal_ip` | `string` | Internal (private) IP address of the instance |
| `external_ip` | `string` | External (public) IP address of the instance (only set when an access config is present) |
| `status` | `string` | Current status of the instance (`RUNNING`, `STOPPED`, etc.) |
| `zone` | `string` | Zone where the instance is located |
| `machine_type` | `string` | Machine type of the instance |
| `cpu_platform` | `string` | CPU platform of the instance |

## Related Components

- [GcpProject](/docs/catalog/gcp/project) — provides the GCP project where the instance is created
- [GcpVpc](/docs/catalog/gcp/vpc) — provides the VPC network for network interface configuration
- [GcpSubnetwork](/docs/catalog/gcp/subnetwork) — provides the subnetwork for network interface configuration

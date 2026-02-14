# Civo Compute Instance

Deploys a virtual machine on Civo Cloud with a specified OS image, instance size, and network placement. The instance supports optional SSH key injection, firewall attachment, volume attachment, reserved IP assignment, and cloud-init scripting for first-boot configuration.

## What Gets Created

When you deploy a CivoComputeInstance resource, OpenMCF provisions:

- **Compute Instance** — a `civo_instance` resource in the specified region, running the chosen OS image at the requested size, attached to the given network

## Prerequisites

- **Civo credentials** configured via environment variables or OpenMCF provider config
- **An existing Civo network** in the target region (can be created with CivoVpc)
- **A valid instance size slug** for the target region (e.g., `g3.small`) — check Civo's available sizes
- **A valid OS image slug** for the target region (e.g., `ubuntu-focal`) — check Civo's available disk images

## Quick Start

Create a file `civo-instance.yaml`:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoComputeInstance
metadata:
  name: my-instance
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CivoComputeInstance.my-instance
spec:
  instanceName: my-instance
  region: nyc1
  size: g3.small
  image: ubuntu-focal
  network:
    value: network-uuid-here
```

Deploy:

```shell
openmcf apply -f civo-instance.yaml
```

This creates a `g3.small` instance in New York running Ubuntu Focal, attached to the specified network.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `instanceName` | `string` | Hostname for the instance. Lowercase letters, numbers, dashes, and dots. Cannot end with a dash or dot. | Required, max 63 chars, pattern `^[a-z0-9]([a-z0-9.\-]*[a-z0-9])?$` |
| `region` | `enum` | Civo region where the instance is created. Valid values: `lon1`, `lon2`, `fra1`, `nyc1`, `phx1`, `mum1`. | Required |
| `size` | `string` | Instance size (flavor) slug defining CPU and memory allocation (e.g., `g3.small`). | Required, pattern `^[a-z0-9]([a-z0-9.\-]*[a-z0-9])?$` |
| `image` | `string` | Base OS disk image slug for the instance (e.g., `ubuntu-focal`). | Required, pattern `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` |
| `network` | `StringValueOrRef` | Network ID where the instance is placed. Must be an existing network in the same region. Can reference a CivoVpc resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `sshKeyIds` | `string[]` | `[]` | SSH public key ID(s) to authorize for passwordless login. If empty, the instance is provisioned with a generated password. Only the first key is used (Civo API limitation). |
| `firewallIds` | `StringValueOrRef[]` | `[]` | Firewall ID(s) to attach to the instance. The firewall must belong to the same network. Only the first firewall is applied. Can reference CivoFirewall resources via `valueFrom`. |
| `volumeIds` | `StringValueOrRef[]` | `[]` | Existing storage volume ID(s) to attach. Volumes must reside in the same region. Can reference CivoVolume resources via `valueFrom`. |
| `reservedIpId` | `StringValueOrRef` | none | Reserved IP ID to assign a static public IPv4 address to this instance. Can reference a CivoIpAddress resource via `valueFrom`. |
| `tags` | `string[]` | `[]` | Tags for organizational purposes. Each tag must be unique within the list. |
| `userData` | `string` | `""` | Cloud-init user data script executed on first boot. Maximum size is 32 KiB. |

## Examples

### Basic Instance

A minimal instance with no SSH keys or firewalls:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoComputeInstance
metadata:
  name: basic-vm
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CivoComputeInstance.basic-vm
spec:
  instanceName: basic-vm
  region: fra1
  size: g3.small
  image: ubuntu-focal
  network:
    value: network-uuid-here
```

### Instance with SSH Access and Firewall

An instance configured with an SSH key for login and a firewall for network access control:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoComputeInstance
metadata:
  name: ssh-vm
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.CivoComputeInstance.ssh-vm
spec:
  instanceName: ssh-vm
  region: nyc1
  size: g3.medium
  image: ubuntu-jammy
  network:
    value: network-uuid-here
  sshKeyIds:
    - ssh-key-uuid-here
  firewallIds:
    - value: firewall-uuid-here
  tags:
    - environment:staging
    - role:webserver
```

### Full-Featured Instance with Foreign Key References and Cloud-Init

An instance referencing other OpenMCF-managed resources, with a reserved IP, attached volume, and a cloud-init script:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoComputeInstance
metadata:
  name: prod-app
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CivoComputeInstance.prod-app
spec:
  instanceName: prod-app
  region: lon1
  size: g3.large
  image: ubuntu-jammy
  network:
    valueFrom:
      kind: CivoVpc
      name: my-network
      field: status.outputs.network_id
  sshKeyIds:
    - ssh-key-uuid-here
  firewallIds:
    - valueFrom:
        kind: CivoFirewall
        name: my-firewall
        field: status.outputs.firewall_id
  volumeIds:
    - valueFrom:
        kind: CivoVolume
        name: my-data-volume
        field: status.outputs.volume_id
  reservedIpId:
    valueFrom:
      kind: CivoIpAddress
      name: my-static-ip
      field: status.outputs.reserved_ip_id
  tags:
    - environment:production
    - managed-by:openmcf
  userData: |
    #!/bin/bash
    apt-get update && apt-get install -y nginx
    systemctl enable nginx
    systemctl start nginx
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `instanceId` | `string` | Unique identifier (UUID) of the created compute instance |
| `publicIpv4` | `string` | Public IPv4 address assigned to the instance, if any |
| `privateIpv4` | `string` | Private IPv4 address of the instance within its network |
| `status` | `string` | Current instance status (e.g., `ACTIVE`, `BUILDING`) |
| `createdAtRfc3339` | `string` | Timestamp when the instance was created, in RFC 3339 format |

## Related Components

- [CivoVpc](/docs/catalog/civo/civovpc) — provides the network for instance placement
- [CivoFirewall](/docs/catalog/civo/civofirewall) — controls inbound and outbound network access
- [CivoVolume](/docs/catalog/civo/civovolume) — provides persistent block storage to attach to the instance
- [CivoIpAddress](/docs/catalog/civo/civoipaddress) — reserves a static public IP for the instance

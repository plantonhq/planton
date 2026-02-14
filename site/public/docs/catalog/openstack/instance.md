---
title: "Instance"
description: "Instance deployment documentation"
icon: "package"
order: 100
componentName: "openstackinstance"
---

# OpenStack Instance

Deploys an OpenStack Compute instance with configurable flavor, image or boot-from-volume source, network attachments via network UUID or pre-provisioned port, and optional placement controls via server groups and availability zones.

## What Gets Created

When you deploy an OpenStackInstance resource, OpenMCF provisions:

- **Compute Instance** — an `openstack_compute_instance_v2` resource with the specified flavor and image (or block device boot source), placed in the configured networks with attached security groups. When `blockDevice` entries are provided, the instance boots from persistent Cinder volumes instead of ephemeral image-based storage. When `serverGroupId` is set, scheduler hints control placement within the server group.

## Prerequisites

- **OpenStack credentials** configured via environment variables or OpenMCF provider config
- **At least one network** (a network UUID or a pre-provisioned port UUID) for the instance to attach to
- **A flavor** (by name or UUID) available in the target OpenStack project
- **A Glance image** (by name or UUID) if not booting from a block device volume
- **An SSH keypair** registered in OpenStack if setting `keyPair`
- **A server group** if using placement constraints via `serverGroupId`

## Quick Start

Create a file `instance.yaml`:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackInstance
metadata:
  name: my-instance
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OpenStackInstance.my-instance
spec:
  flavorName: m1.medium
  imageName: ubuntu-22.04
  networks:
    - uuid: 4a0e3c5b-2f1d-4e6a-8b9c-0d1e2f3a4b5c
```

Deploy:

```shell
openmcf apply -f instance.yaml
```

This creates a compute instance with the `m1.medium` flavor, booted from the `ubuntu-22.04` Glance image, attached to the specified network.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `flavorName` | `string` | Human-readable name of the instance flavor (e.g., `m1.medium`). Mutually exclusive with `flavorId`. | Exactly one of `flavorName` or `flavorId` must be set |
| `flavorId` | `string` | UUID of the instance flavor. Mutually exclusive with `flavorName`. | Exactly one of `flavorName` or `flavorId` must be set |
| `networks` | `InstanceNetwork[]` | Network attachments for the instance. Each entry connects the instance to a network via UUID or a pre-provisioned port. | Minimum 1 item required |
| `networks[].uuid` | `StringValueOrRef` | Network UUID to attach to. OpenStack auto-creates a port on this network. Can reference an OpenStackNetwork resource via `valueFrom`. Mutually exclusive with `port`. | Exactly one of `uuid` or `port` per network entry |
| `networks[].port` | `StringValueOrRef` | UUID of a pre-provisioned port to attach. Use for stable MAC/IP addresses or port-level security groups. Can reference an OpenStackNetworkPort resource via `valueFrom`. Mutually exclusive with `uuid`. | Exactly one of `uuid` or `port` per network entry |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `imageName` | `string` | — | Name of the Glance image to boot from (e.g., `ubuntu-22.04`). Optional when using `blockDevice` with a boot volume. |
| `imageId` | `string` | — | UUID of the Glance image to boot from. Alternative to `imageName`. Optional when using `blockDevice` with a boot volume. |
| `keyPair` | `StringValueOrRef` | — | SSH keypair name to inject into the instance. Can reference an OpenStackKeypair resource via `valueFrom`. |
| `securityGroups` | `StringValueOrRef[]` | default SG | Security group names to apply. Uses names, not UUIDs. Can reference OpenStackSecurityGroup resources via `valueFrom`. If omitted, OpenStack applies the default security group. |
| `blockDevice` | `BlockDevice[]` | `[]` | Block device mappings for boot-from-volume or additional volumes at launch. |
| `blockDevice[].sourceType` | `string` | — | Source type: `blank`, `image`, `snapshot`, or `volume`. Required per block device entry. |
| `blockDevice[].uuid` | `string` | — | UUID of the source image, volume, or snapshot. Required unless `sourceType` is `blank`. |
| `blockDevice[].destinationType` | `string` | `local` | Where the device is created: `local` (ephemeral) or `volume` (persistent Cinder volume). |
| `blockDevice[].bootIndex` | `int` | `0` | Boot order: `0` = boot device, `-1` = not bootable, higher values = lower priority. |
| `blockDevice[].volumeSize` | `int` | — | Size in GB. Required for image-to-volume and blank mappings. |
| `blockDevice[].deleteOnTermination` | `bool` | `false` | When `true`, the volume is deleted when the instance terminates. |
| `blockDevice[].volumeType` | `string` | — | Cinder volume type (e.g., `SSD`, `HDD`). Only applies when `destinationType` is `volume`. |
| `networks[].fixedIpV4` | `string` | — | Specific IPv4 address to request on the network. Only applies when `uuid` is used, not `port`. |
| `networks[].accessNetwork` | `bool` | `false` | Marks this network as the access network, determining which IP appears in the `access_ip_v4` output. |
| `userData` | `string` | — | Cloud-init configuration or script for first boot. Base64-encoded before passing to the instance. ForceNew: changing this recreates the instance. |
| `metadata` | `map<string, string>` | `{}` | Key-value pairs attached to the instance, visible in the OpenStack API. Can be updated without recreating the instance. |
| `configDrive` | `bool` | — | Enables a config drive containing instance metadata and user data on a local read-only disk. ForceNew: changing this recreates the instance. |
| `serverGroupId` | `StringValueOrRef` | — | Server group UUID for placement control. Maps to scheduler hints. Can reference an OpenStackServerGroup resource via `valueFrom`. ForceNew. |
| `availabilityZone` | `string` | — | AZ where the instance launches (e.g., `nova`, `az1`). If omitted, Nova selects one. ForceNew. |
| `tags` | `string[]` | `[]` | Tags for filtering and organization in the OpenStack API. Must be unique. |
| `region` | `string` | provider default | Overrides the region from the provider config for this instance. |

## Examples

### Basic Instance with Image Boot

A minimal instance booted from a Glance image on a single network with an SSH keypair:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackInstance
metadata:
  name: web-server
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OpenStackInstance.web-server
spec:
  flavorName: m1.small
  imageName: ubuntu-22.04
  keyPair: my-keypair
  networks:
    - uuid: 4a0e3c5b-2f1d-4e6a-8b9c-0d1e2f3a4b5c
  securityGroups:
    - web-sg
```

### Boot from Volume

An instance with a persistent 50 GB root disk created from a Glance image, recommended for production workloads where the root disk must survive instance rebuilds:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackInstance
metadata:
  name: db-server
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OpenStackInstance.db-server
spec:
  flavorName: m1.large
  keyPair: ops-keypair
  networks:
    - uuid: 4a0e3c5b-2f1d-4e6a-8b9c-0d1e2f3a4b5c
  securityGroups:
    - db-sg
  blockDevice:
    - sourceType: image
      uuid: 12345678-abcd-efgh-ijkl-123456789abc
      destinationType: volume
      bootIndex: 0
      volumeSize: 50
      deleteOnTermination: false
      volumeType: SSD
```

### Full-Featured Instance with Placement Controls

Production instance with server group placement, cloud-init configuration, metadata, multiple networks, and a specific availability zone:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackInstance
metadata:
  name: app-server-01
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OpenStackInstance.app-server-01
spec:
  flavorName: m1.xlarge
  imageName: ubuntu-22.04
  keyPair: prod-keypair
  availabilityZone: az1
  configDrive: true
  serverGroupId: 98765432-dcba-4321-fedc-ba9876543210
  networks:
    - uuid: 4a0e3c5b-2f1d-4e6a-8b9c-0d1e2f3a4b5c
      accessNetwork: true
    - uuid: 7d8e9f0a-1b2c-3d4e-5f6a-7b8c9d0e1f2a
  securityGroups:
    - app-sg
    - monitoring-sg
  userData: |
    #cloud-config
    packages:
      - nginx
      - prometheus-node-exporter
    runcmd:
      - systemctl enable nginx
      - systemctl start nginx
  metadata:
    environment: production
    team: platform
  tags:
    - production
    - app-tier
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding UUIDs:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackInstance
metadata:
  name: ref-instance
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OpenStackInstance.ref-instance
spec:
  flavorName: m1.medium
  imageName: ubuntu-22.04
  keyPair:
    valueFrom:
      kind: OpenStackKeypair
      name: my-keypair
      field: status.outputs.name
  networks:
    - uuid:
        valueFrom:
          kind: OpenStackNetwork
          name: my-network
          field: status.outputs.network_id
  securityGroups:
    - valueFrom:
        kind: OpenStackSecurityGroup
        name: app-sg
        field: status.outputs.name
  serverGroupId:
    valueFrom:
      kind: OpenStackServerGroup
      name: anti-affinity-group
      field: status.outputs.server_group_id
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `instance_id` | `string` | UUID of the created compute instance |
| `name` | `string` | Name of the instance, derived from `metadata.name` |
| `access_ip_v4` | `string` | Best IPv4 address for accessing the instance, prioritizing the access network if one is marked |
| `access_ip_v6` | `string` | Best IPv6 address for accessing the instance. Empty if the instance has no IPv6 connectivity. |
| `region` | `string` | OpenStack region where the instance was created |

## Related Components

- [OpenStackNetwork](/docs/catalog/openstack/network) — provides the network for instance attachment
- [OpenStackNetworkPort](/docs/catalog/openstack/network-port) — provides pre-provisioned ports for stable network identity
- [OpenStackKeypair](/docs/catalog/openstack/keypair) — manages the SSH keypair injected into the instance
- [OpenStackSecurityGroup](/docs/catalog/openstack/security-group) — controls inbound and outbound traffic rules
- [OpenStackServerGroup](/docs/catalog/openstack/server-group) — defines placement policies (affinity/anti-affinity) for instance groups

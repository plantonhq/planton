# OpenStackInstance

An OpenStack Compute instance (virtual machine) -- the fundamental compute resource.

## Overview

An instance is a virtual machine running on an OpenStack cloud. It binds together networking (networks/ports), storage (images/volumes), access control (keypairs/security groups), and placement (server groups/AZs) into a running VM. This is the most important resource after networking -- every developer workload, application server, and database runs as an instance.

## Spec Fields

| Field | Type | Required | FK | Description |
|-------|------|----------|-----|-------------|
| `flavor_name` | string | XOR with flavor_id | - | Flavor name (e.g., "m1.medium") |
| `flavor_id` | string | XOR with flavor_name | - | Flavor UUID |
| `image_name` | string | No | - | Image name (e.g., "ubuntu-22.04") |
| `image_id` | string | No | - | Image UUID |
| `key_pair` | StringValueOrRef | No | OpenStackKeypair | SSH keypair name |
| `networks` | repeated InstanceNetwork | Yes (min 1) | Network/Port | Network attachments |
| `security_groups` | repeated StringValueOrRef | No | OpenStackSecurityGroup | Security group names |
| `block_device` | repeated BlockDevice | No | - | Block device mappings |
| `user_data` | string | No | - | Cloud-init script (ForceNew) |
| `metadata` | map\<string,string\> | No | - | Instance metadata |
| `config_drive` | optional bool | No | - | Enable config drive (ForceNew) |
| `server_group_id` | StringValueOrRef | No | OpenStackServerGroup | Placement group |
| `availability_zone` | string | No | - | AZ placement (ForceNew) |
| `tags` | repeated string | No | - | Instance tags (unique) |
| `region` | string | No | - | Region override |

### InstanceNetwork (nested)

| Field | Type | FK | Description |
|-------|------|-----|-------------|
| `uuid` | StringValueOrRef | OpenStackNetwork | Network UUID (XOR with port) |
| `port` | StringValueOrRef | OpenStackNetworkPort | Port UUID (XOR with uuid) |
| `fixed_ip_v4` | string | - | Specific IPv4 address |
| `access_network` | bool | - | Mark as access network |

### BlockDevice (nested)

| Field | Type | Description |
|-------|------|-------------|
| `source_type` | string | Required: "blank", "image", "snapshot", "volume" |
| `uuid` | string | Source UUID (required unless "blank") |
| `destination_type` | string | "local" or "volume" |
| `boot_index` | int32 | 0 = boot device, -1 = not bootable |
| `volume_size` | int32 | Size in GB |
| `delete_on_termination` | bool | Delete volume on instance termination |
| `volume_type` | string | Volume type (e.g., "SSD") |

## Outputs

| Output | Description |
|--------|-------------|
| `instance_id` | UUID (used by VolumeAttach FK) |
| `name` | Instance name |
| `access_ip_v4` | Best IPv4 access address |
| `access_ip_v6` | Best IPv6 access address |
| `region` | Region where created |

## Usage

### Minimal (image boot)

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackInstance
metadata:
  name: dev-vm
spec:
  flavor_name: m1.medium
  image_name: ubuntu-22.04
  networks:
    - uuid:
        value: "network-uuid-here"
```

### InfraChart with FK References

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackInstance
metadata:
  name: app-server
spec:
  flavor_name: m1.large
  image_name: ubuntu-22.04
  key_pair:
    value_from:
      name: prod-keypair
  networks:
    - port:
        value_from:
          name: app-port
      access_network: true
  security_groups:
    - value_from:
        name: web-sg
    - value_from:
        name: ssh-sg
  server_group_id:
    value_from:
      name: ha-group
  user_data: |
    #!/bin/bash
    apt-get update && apt-get install -y nginx
  metadata:
    role: webserver
    environment: production
```

### Boot from Volume

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackInstance
metadata:
  name: persistent-vm
spec:
  flavor_name: m1.large
  networks:
    - uuid:
        value_from:
          name: app-network
  block_device:
    - source_type: image
      uuid: "image-uuid-here"
      destination_type: volume
      boot_index: 0
      volume_size: 50
      delete_on_termination: true
      volume_type: SSD
```

## Foreign Key Relationships

This component has the most FK connections of any OpenStack component:
- `key_pair` -> OpenStackKeypair.status.outputs.name
- `networks[].uuid` -> OpenStackNetwork.status.outputs.network_id
- `networks[].port` -> OpenStackNetworkPort.status.outputs.port_id
- `security_groups[]` -> OpenStackSecurityGroup.status.outputs.name
- `server_group_id` -> OpenStackServerGroup.status.outputs.server_group_id

## Important Notes

- **security_groups uses NAMES**: The Compute API (Nova) uses security group names, unlike Neutron which uses UUIDs. The FK resolves to `status.outputs.name`.
- **ForceNew fields**: `key_pair`, `networks`, `user_data`, `config_drive`, `server_group_id`, `availability_zone`, `block_device` all recreate the instance on change.
- **Flavor changes**: `flavor_name`/`flavor_id` changes trigger a resize (not recreate).
- **Image optional**: Not needed when using `block_device` with `boot_index=0`.

## Related Components

- **OpenStackKeypair** (2500) -- SSH access
- **OpenStackNetwork** (2501) -- Network attachment
- **OpenStackNetworkPort** (2507) -- Pre-provisioned port attachment
- **OpenStackSecurityGroup** (2505) -- Network access control
- **OpenStackServerGroup** (2509) -- Placement control
- **OpenStackVolumeAttach** (2511) -- Post-launch volume attachment (Phase 3)

# Scaleway Instance

Deploys a Scaleway compute Instance as a composite resource that bundles the server, an optional dedicated Flexible IP, optional additional local volumes, and an optional Private Network attachment into a single declarative manifest. The instance is provisioned in a specific Scaleway zone and can reference ScalewayInstanceSecurityGroup and ScalewayPrivateNetwork resources for network configuration.

## What Gets Created

When you deploy a ScalewayInstance resource, OpenMCF provisions:

- **Instance Server** — an `instance.Server` resource in the specified zone with the chosen commercial type, base image, root volume configuration, optional cloud-init script, instance state control, and deletion protection
- **Flexible IP** (optional) — a dedicated `instance.Ip` resource providing a public IPv4 address that has an independent lifecycle, surviving instance replacement to preserve DNS records and firewall rules. Created only when `publicIp` is set.
- **Additional Volumes** (optional) — one or more `instance.Volume` resources (local SSD or scratch) created and attached to the instance via `additionalVolumeIds`. These volumes share the instance's lifecycle.
- **Private Network Attachment** (optional) — an inline private NIC on the server connecting it to a ScalewayPrivateNetwork, enabling communication with other resources over private IPs. Created only when `privateNetworkId` is set.

## Prerequisites

- **Scaleway credentials** configured via environment variables or OpenMCF provider config
- **A target zone** where the instance will be created (e.g., `fr-par-1`, `nl-ams-1`, `pl-waw-1`). The zone must match the zone of any referenced Private Network or security group.
- **A base image** — either a UUID or a human-friendly label (e.g., `ubuntu_jammy`, `debian_bullseye`) available in the target zone
- **(Optional) A ScalewayPrivateNetwork** if you want the instance on an internal network
- **(Optional) A ScalewayInstanceSecurityGroup** if you need custom firewall rules

## Quick Start

Create a file `instance.yaml`:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayInstance
metadata:
  name: web-01
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.ScalewayInstance.web-01
spec:
  zone: fr-par-1
  type: DEV1-S
  image: ubuntu_jammy
  publicIp: {}
  state: started
```

Deploy:

```shell
openmcf apply -f instance.yaml
```

This creates a `DEV1-S` instance running Ubuntu Jammy in `fr-par-1` with a dedicated public IP address.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `zone` | `string` | Scaleway zone where the instance will be created (e.g., `"fr-par-1"`, `"nl-ams-1"`, `"pl-waw-1"`). Determines the physical data center. Cannot be changed after creation. | Required |
| `type` | `string` | Instance commercial type determining CPU, RAM, and storage allocation. Examples: `"DEV1-S"` (2 vCPU, 2 GB), `"DEV1-M"` (3 vCPU, 4 GB), `"GP1-S"` (8 vCPU, 32 GB), `"PRO2-S"` (2 vCPU, 8 GB). Can be changed after creation (causes stop/migrate/restart). | Required |
| `image` | `string` | Base image UUID or label. Labels: `"ubuntu_jammy"`, `"ubuntu_focal"`, `"debian_bullseye"`, `"centos_stream_9"`. Use UUIDs for reproducible deployments. Available images depend on the zone. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `publicIp` | `object` | not set | When present, a dedicated Flexible IPv4 address is created and attached. Omit for instances behind a Load Balancer or Public Gateway. |
| `publicIp.reverseDns` | `string` | Scaleway default | Reverse DNS hostname for the public IP (e.g., `"web-01.example.com"`). A matching DNS A record must exist before setting this. |
| `securityGroupId` | `string` or `valueFrom` | Scaleway default SG | Security group UUID or reference to a ScalewayInstanceSecurityGroup output. If omitted, Scaleway assigns its default security group (allows all traffic). |
| `privateNetworkId` | `string` or `valueFrom` | not set | Private Network UUID or reference to a ScalewayPrivateNetwork output. When set, the instance receives a private NIC for internal communication. |
| `rootVolume` | `object` | image defaults | Root volume configuration. If omitted, the image's default root volume settings are used. |
| `rootVolume.sizeInGb` | `int` | image default | Root volume size in GB. Minimum depends on the image (typically 10 GB). Can only be increased after creation. |
| `rootVolume.volumeType` | `string` | `"l_ssd"` | Root disk type: `"l_ssd"` (local SSD, high performance) or `"sbs_volume"` (network-attached, resizable, snapshottable). Changing after creation recreates the instance. |
| `rootVolume.deleteOnTermination` | `bool` | `true` | Delete the root volume when the instance is terminated. Set to `false` to preserve data. Only meaningful for SBS volumes. |
| `rootVolume.sbsIops` | `int` | Scaleway default | Guaranteed IOPS for the root volume. Only relevant when `volumeType` is `"sbs_volume"`. |
| `additionalVolumes` | `list` | `[]` | Additional local volumes to create and attach. These share the instance's lifecycle. |
| `additionalVolumes[].name` | `string` | auto-generated | Descriptive name for the volume. |
| `additionalVolumes[].volumeType` | `string` | `"l_ssd"` | Volume type: `"l_ssd"` (local SSD) or `"scratch"` (ephemeral, data lost on stop/restart). | Required per item |
| `additionalVolumes[].sizeInGb` | `int` | — | Volume size in GB. Total of all local volumes cannot exceed the instance type's maximum. | Required per item |
| `cloudInit` | `string` | not set | Cloud-init script (e.g., `#!/bin/bash` or `#cloud-config`) executed on first boot. Maximum ~127 KB. |
| `state` | `string` | `"started"` | Desired instance state: `"started"` (running), `"stopped"` (shut down, not billed for compute), or `"standby"` (suspended to RAM, faster resume). |
| `protected` | `bool` | `false` | When `true`, the instance cannot be deleted via the API without first disabling protection. Enable for production workloads. |

## Examples

### Minimal Development Instance

A small development instance with a public IP and the default root volume:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayInstance
metadata:
  name: dev-box
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.ScalewayInstance.dev-box
spec:
  zone: fr-par-1
  type: DEV1-S
  image: ubuntu_jammy
  publicIp: {}
  state: started
```

### Production Instance with Private Network and Security Group

A production instance on a Private Network, behind a custom security group, with a larger SBS root volume, cloud-init bootstrapping, and deletion protection enabled. The security group and Private Network are referenced from other OpenMCF resources:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayInstance
metadata:
  name: app-server
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.ScalewayInstance.app-server
spec:
  zone: fr-par-1
  type: PRO2-S
  image: ubuntu_jammy
  securityGroupId:
    valueFrom:
      kind: ScalewayInstanceSecurityGroup
      name: web-sg
      fieldPath: status.outputs.security_group_id
  privateNetworkId:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: app-network
      fieldPath: status.outputs.private_network_id
  rootVolume:
    sizeInGb: 50
    volumeType: sbs_volume
    deleteOnTermination: true
    sbsIops: 5000
  cloudInit: |
    #!/bin/bash
    apt-get update && apt-get install -y docker.io
    systemctl enable docker
    systemctl start docker
  state: started
  protected: true
```

### Instance with Additional Volumes and Custom Root Volume

A general-purpose instance with a larger local SSD root volume and two additional volumes — one for application data and one for temporary processing:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayInstance
metadata:
  name: data-processor
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.ScalewayInstance.data-processor
spec:
  zone: nl-ams-1
  type: GP1-S
  image: debian_bullseye
  publicIp:
    reverseDns: data-processor.example.com
  rootVolume:
    sizeInGb: 40
    volumeType: l_ssd
  additionalVolumes:
    - name: app-data
      volumeType: l_ssd
      sizeInGb: 100
    - name: scratch-space
      volumeType: scratch
      sizeInGb: 50
  cloudInit: |
    #!/bin/bash
    mkfs.ext4 /dev/vdb
    mkdir -p /mnt/app-data
    mount /dev/vdb /mnt/app-data
    echo '/dev/vdb /mnt/app-data ext4 defaults 0 2' >> /etc/fstab
  state: started
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `server_id` | `string` | Zoned ID of the created instance server (format: `{zone}/{uuid}`, e.g., `fr-par-1/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`). Used for management, monitoring, and lifecycle operations. |
| `public_ip_address` | `string` | Public IPv4 address of the instance's Flexible IP. Empty string if `publicIp` was not set. Use for DNS A records, SSH access, and external service whitelisting. |
| `public_ip_id` | `string` | Zoned ID of the Flexible IP resource. Empty string if `publicIp` was not set. The Flexible IP survives instance replacement, preserving DNS records and firewall rules. |
| `private_ip_address` | `string` | Private IP address on the attached Private Network. Empty string if `privateNetworkId` was not set. Use for Load Balancer backends, internal service discovery, and database connection allowlists. |

## Related Components

- [ScalewayInstanceSecurityGroup](/docs/catalog/scaleway/scalewayinstancesecuritygroup) — zonal firewall that controls inbound and outbound traffic rules; referenced via the `securityGroupId` field
- [ScalewayPrivateNetwork](/docs/catalog/scaleway/scalewayprivatenetwork) — internal network for private communication between instances, databases, and load balancers; referenced via the `privateNetworkId` field
- [ScalewayVpc](/docs/catalog/scaleway/scalewayvpc) — regional VPC that contains Private Networks
- [ScalewayPublicGateway](/docs/catalog/scaleway/scalewaypublicgateway) — provides NAT and DHCP for instances on a Private Network that do not have a public IP
- [ScalewayKapsuleCluster](/docs/catalog/scaleway/scalewaykapsulecluster) — managed Kubernetes cluster; instances can serve as standalone workloads alongside Kapsule clusters on the same Private Network
- [ScalewayRdbInstance](/docs/catalog/scaleway/scalewayrdbinstance) — managed database that instances connect to over the Private Network using `private_ip_address`

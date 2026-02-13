# ScalewayInstance

The **ScalewayInstance** resource provides a declarative way to provision and manage Scaleway compute instances through OpenMCF. It is a **composite resource** that bundles a virtual machine with its public IP, local volumes, and private network attachment into a single manifest.

## What It Represents

A [Scaleway Instance](https://www.scaleway.com/en/virtual-instances/) is a virtual machine running on Scaleway's cloud platform. Instances are available in a range of types from lightweight development VMs (DEV1) to production-grade compute-optimized machines (PRO2, GP1).

## Bundled Terraform Resources

Applying a single `ScalewayInstance` manifest creates up to 3 Scaleway resource types:

| Terraform Resource | Created When | Purpose |
|---|---|---|
| `scaleway_instance_server` | Always | The virtual machine (compute + root volume) |
| `scaleway_instance_ip` | Only if `spec.public_ip` is set | Dedicated Flexible IPv4 with independent lifecycle |
| `scaleway_instance_volume` | For each `spec.additional_volumes` entry | Local volumes (l_ssd, scratch) attached to the server |

The Private Network attachment uses an **inline block** on the server resource (not a separate `scaleway_instance_private_nic`), consistent with the ScalewayLoadBalancer pattern.

## Key Features

### Optional Public IP

Unlike network appliances (Load Balancer, Public Gateway) that always need a public IP, instances can run with **no public IP** -- reachable only via Private Network through a bastion host, VPN, or Load Balancer. This is the recommended production topology.

Set `spec.public_ip` to create a Flexible IP. Omit it for private-only instances.

### Security Group Integration

Attach a `ScalewayInstanceSecurityGroup` via the `security_group_id` field (using `StringValueOrRef`). This controls inbound and outbound firewall rules. If omitted, Scaleway assigns its default security group (which allows all traffic).

### Private Network Attachment

Attach to a `ScalewayPrivateNetwork` via the `private_network_id` field (using `StringValueOrRef`). The instance receives a private NIC and a private IP, enabling communication with other resources on the same network using private addresses.

### Root Volume Configuration

The root volume inherits defaults from the selected image and instance type. Override with `spec.root_volume` to:
- Increase disk size beyond the image default
- Switch between Local SSD (`l_ssd`) and SBS (`sbs_volume`) storage
- Configure IOPS for SBS volumes
- Control delete-on-termination behavior

### Additional Local Volumes

Create local volumes (`l_ssd`, `scratch`) that are automatically attached to the instance. These volumes share the instance's lifecycle. For persistent block storage with independent lifecycle, use `ScalewayBlockVolume` (a separate resource kind).

### Cloud-Init Bootstrapping

Pass a cloud-init script via `spec.cloud_init` to automatically configure the instance on first boot. Supports shell scripts and cloud-config YAML.

## Upstream Dependencies (What This Resource Needs)

| Dependency | Field | Required | Purpose |
|---|---|---|---|
| `ScalewayPrivateNetwork` | `spec.private_network_id` | No | Attach instance to a private network |
| `ScalewayInstanceSecurityGroup` | `spec.security_group_id` | No | Firewall rules for the instance |

## Downstream Dependents (What References This Resource)

| Dependent | Output Used | Purpose |
|---|---|---|
| `ScalewayDnsRecord` | `status.outputs.public_ip_address` | Create DNS A records pointing to the instance |
| `ScalewayLoadBalancer` | `status.outputs.private_ip_address` | Register as a backend server |

## Stack Outputs

| Output | Description |
|---|---|
| `server_id` | The zoned ID of the instance server |
| `public_ip_address` | Public IPv4 address (empty if no public IP) |
| `public_ip_id` | Flexible IP resource ID (empty if no public IP) |
| `private_ip_address` | Private IP on the attached Private Network (empty if none) |

## Instance Types

| Category | Types | Use Case |
|---|---|---|
| Development | `DEV1-S`, `DEV1-M`, `DEV1-L`, `DEV1-XL` | Development, testing, CI runners |
| General Purpose | `GP1-S`, `GP1-M`, `GP1-L`, `GP1-XL` | General workloads, web servers |
| Production | `PRO2-XXS`, `PRO2-XS`, `PRO2-S`, `PRO2-M`, `PRO2-L` | Production, databases, high-performance |

See [Scaleway Instance Types](https://www.scaleway.com/en/pricing/?tags=compute) for the full list with specifications and pricing.

## What's Not Included (Deferred)

The following Scaleway instance features are intentionally deferred to future versions:

- **Placement groups** -- Server affinity/anti-affinity scheduling
- **Boot type** -- Custom boot configurations (default "local" covers 99% of cases)
- **Filesystems** -- Scaleway filesystem attachments
- **Dynamic IP** -- Automatic ephemeral public IP (use explicit `public_ip` instead)
- **SBS volume attachment** -- Attaching externally-created block volumes (use `ScalewayBlockVolume` when R13 is implemented)
- **Multiple private networks** -- Attaching to more than one Private Network (max 8 supported by Scaleway)

All deferred features can be added as new optional fields without breaking changes.

## References

- [Scaleway Instances Documentation](https://www.scaleway.com/en/docs/compute/instances/)
- [Scaleway Instance Types and Pricing](https://www.scaleway.com/en/pricing/?tags=compute)
- [Terraform scaleway_instance_server Resource](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/instance_server)
- [Scaleway Cloud-Init Guide](https://www.scaleway.com/en/docs/compute/instances/how-to/use-cloud-init/)

# HetznerCloudNetwork

The **HetznerCloudNetwork** resource creates a private network in a Hetzner Cloud account with subnets and optional static routes. The network provides isolated IPv4 connectivity between cloud resources — servers and load balancers attach to subnets within the network and communicate using private RFC 1918 addresses, separate from the public internet.

## What It Represents

A [Hetzner Cloud Network](https://docs.hetzner.cloud/#networks) is a user-defined private address space divided into subnets assigned to network zones. It is the equivalent of a VPC in AWS or a VNet in Azure, but simpler — no NAT gateways, no internet gateways, no peering. Servers connect to the network through subnets and receive private IPs within the subnet's CIDR range. Optional static routes enable custom traffic paths for VPN gateways or NAT instances.

## Bundled Resources

| Terraform Resource | Count | Created When | Purpose |
|---|---|---|---|
| `hcloud_network` | 1 | Always | Creates the network with top-level CIDR, labels, and protection settings |
| `hcloud_network_subnet` | N | Always (min 1) | Creates subnets within the network, each in a specific zone |
| `hcloud_network_route` | M | When `routes` is non-empty | Creates static routes for custom traffic paths |

This is a multi-resource component. Subnets are required (minimum 1) because a Hetzner Cloud network is unusable without them — servers and load balancers attach to subnets, not directly to the network. Routes are optional because default routing handles most use cases.

## Key Features

### Multi-Zone Private Networking

Subnets can span multiple Hetzner Cloud network zones (`eu-central`, `us-east`, `us-west`, `ap-southeast`), enabling cross-region private connectivity within a single network. Traffic between zones traverses Hetzner's backbone with no egress charges.

### Three Subnet Types

- **cloud** — Standard subnet for Hetzner Cloud servers. The most common type.
- **server** — Subnet for connecting Hetzner Robot (dedicated) servers.
- **vswitch** — Subnet linked to a Hetzner Robot vSwitch for hybrid cloud/dedicated connectivity. Requires a `vswitchId`.

### Minimum Subnet Enforcement

The spec requires at least one subnet (`min_items: 1`). This prevents deploying a network that cannot host any workloads — a guardrail that Terraform and Pulumi do not enforce.

### CEL Validation for vSwitch Subnets

A CEL rule enforces that `vswitchId` is provided when subnet type is `vswitch`. This catches the misconfiguration at manifest validation time rather than at the Hetzner Cloud API level.

### Automatic Labeling

Standard labels (`resource`, `name`, `kind`, `org`, `env`, `id`) are applied to the Hetzner Cloud network from metadata. User-specified `metadata.labels` are merged in, with standard labels taking precedence. Subnets and routes do not support labels in the Hetzner Cloud API.

## Upstream Dependencies (What This Resource Needs)

None. `HetznerCloudNetwork` is a foundation resource with no foreign key dependencies.

## Downstream Dependents (What References This Resource)

| Dependent | Field | Purpose |
|---|---|---|
| `HetznerCloudServer` | `spec` (network reference) | Attach server to the network for private connectivity |
| `HetznerCloudLoadBalancer` | `spec` (network reference) | Attach load balancer to reach backend servers over private IPs |

## Stack Outputs

| Output | Description |
|---|---|
| `network_id` | Hetzner Cloud numeric ID of the created network (as string). Referenced by HetznerCloudServer and HetznerCloudLoadBalancer via StringValueOrRef. |

## References

- [Hetzner Cloud Networks Documentation](https://docs.hetzner.cloud/#networks)
- [Terraform hcloud_network Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/network)
- [Terraform hcloud_network_subnet Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/network_subnet)
- [Terraform hcloud_network_route Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/network_route)
- [Pulumi hcloud.Network Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/network/)
- [Pulumi hcloud.NetworkSubnet Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/networksubnet/)

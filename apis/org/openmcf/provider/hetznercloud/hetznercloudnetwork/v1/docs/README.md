# Hetzner Cloud Network — Research Documentation

## Introduction

A Hetzner Cloud network provides private IPv4 connectivity between cloud resources — servers, load balancers, and (optionally) dedicated Robot servers. It is the VPC equivalent in Hetzner's ecosystem: a user-defined address space carved into subnets, with optional static routes for custom traffic paths. Unlike the public network interface that every server receives by default, a Hetzner Cloud network creates an isolated Layer 3 overlay where resources communicate using private RFC 1918 addresses.

The `HetznerCloudNetwork` component provisions a network with its subnets and routes as a single declarative unit. It is a **foundation resource**: it has no dependencies, but it is referenced by `HetznerCloudServer` and `HetznerCloudLoadBalancer` via its `network_id` output through `StringValueOrRef`. Every non-trivial Hetzner Cloud deployment — anything beyond a single public-facing server — begins with a network.

OpenMCF bundles the network, its subnets, and its routes into a single component because a Hetzner Cloud network is unusable without at least one subnet. Servers and load balancers attach to subnets, not to the network directly. Routes are tightly coupled to the network's address space and routing table. Splitting these into separate components would force every user to deploy at least two manifests (network + subnet) for the simplest private networking setup — unnecessary friction for a relationship that is always 1:N and never cross-referenced.

## Historical Context

Private networking in cloud infrastructure has evolved through several distinct phases, each driven by scale, security requirements, and the operational complexity of managing network boundaries.

**Flat networking era:** Early cloud platforms gave every instance a public IP and an optional private IP on a shared flat network. All instances in the same data center could communicate over private addresses, but there was no isolation between customers or projects. Security relied entirely on host-level firewalls (iptables, Windows Firewall). This model was simple but fundamentally insecure — a misconfigured firewall meant any instance in the data center could reach any other.

**VPC era:** AWS introduced Virtual Private Cloud (VPC) in 2009, creating isolated network segments per customer with user-defined CIDR blocks, subnets mapped to availability zones, and route tables for custom traffic paths. Google Cloud followed with VPC networks (auto-mode and custom-mode), Azure with Virtual Networks (VNets). The key insight was that network isolation should be a platform primitive, not a host-level concern. VPCs became the foundation of every serious cloud deployment.

**Hetzner Cloud's approach:** Hetzner Cloud introduced networks as a lighter-weight alternative to full VPC implementations. The model is simpler than AWS VPC: there are no NAT gateways, no internet gateways, no VPC peering, no transit gateways, no endpoint services. A Hetzner network is a private CIDR block divided into subnets assigned to network zones (regions). Servers attach to subnets and communicate over private IPs. Routes provide custom static paths for VPN gateways or inter-network traffic. This simplicity reflects Hetzner's product philosophy — fewer abstractions, less operational surface area, lower cost.

**The hybrid twist:** Where Hetzner Cloud networks get interesting is the `vswitch` subnet type. Hetzner operates both a cloud platform (Hetzner Cloud) and a dedicated server platform (Hetzner Robot). A vSwitch subnet bridges the two, allowing cloud servers and dedicated servers to communicate over the same private network. This is a capability that AWS, GCP, and Azure achieve only through VPN tunnels or dedicated interconnects — Hetzner provides it natively because both platforms run in the same data centers.

**IaC era:** Terraform and Pulumi brought version control and drift detection to network management. But in both tools, a Hetzner network requires three separate resource declarations: the network, at least one subnet, and optionally routes. Each resource has its own lifecycle, and the subnet/route resources reference the network by ID. This is correct but verbose — a simple private network with two subnets requires 3 Terraform resources and careful dependency management.

**OpenMCF approach:** A single manifest that declares the network CIDR, subnets as a repeated list, and routes as an optional repeated list. The IaC modules handle resource creation order, ID passing between resources, and resource naming. The `network_id` output feeds into downstream components via `StringValueOrRef`, enabling declarative composition without hardcoded IDs.

## Deployment Methods Landscape

### Level 0: Manual (Hetzner Cloud Console)

1. Log in to [console.hetzner.cloud](https://console.hetzner.cloud)
2. Navigate to **Networks** in the left sidebar
3. Click **Create Network**
4. Enter a name and IP range (e.g., `10.0.0.0/16`)
5. Add at least one subnet:
   - Select network zone (eu-central, us-east, us-west, ap-southeast)
   - Enter subnet CIDR (e.g., `10.0.1.0/24`)
   - Select type (Cloud, Server, vSwitch)
6. Optionally add routes:
   - Enter destination CIDR
   - Enter gateway IP
7. Click **Create Network**

**Pros:**
- Zero tooling required
- Visual subnet and route builder
- Immediate feedback on CIDR overlap errors
- Network zone selection via dropdown

**Cons:**
- No audit trail beyond Hetzner's internal logs
- No version control — CIDR changes cannot be reviewed before applying
- Cannot reproduce across environments or projects
- Subnet modifications require navigating into the network detail page
- No way to enforce organizational standards for addressing schemes

**Verdict:** Acceptable for experiments and learning the networking model. Not suitable for any environment where network topology must be reproducible or auditable.

### Level 1: CLI (`hcloud`)

The Hetzner Cloud CLI provides network management commands. Unlike the single-step console workflow, the CLI requires multiple sequential commands:

```bash
# Create the network (top-level CIDR only)
hcloud network create --name my-network --ip-range 10.0.0.0/16

# Add a cloud subnet (eu-central zone)
hcloud network add-subnet my-network \
  --type cloud \
  --network-zone eu-central \
  --ip-range 10.0.1.0/24

# Add a second subnet (us-east zone)
hcloud network add-subnet my-network \
  --type cloud \
  --network-zone us-east \
  --ip-range 10.0.2.0/24

# Add a static route
hcloud network add-route my-network \
  --destination 172.16.0.0/12 \
  --gateway 10.0.1.1

# Inspect the network
hcloud network describe my-network

# List all networks
hcloud network list

# Enable delete protection
hcloud network enable-protection my-network delete

# Remove a subnet
hcloud network remove-subnet my-network --ip-range 10.0.2.0/24

# Remove a route
hcloud network remove-route my-network \
  --destination 172.16.0.0/12 \
  --gateway 10.0.1.1

# Delete the network
hcloud network delete my-network
```

**Pros:**
- Scriptable
- Incremental subnet/route management (add and remove individually)
- Good for debugging and inspecting network state
- Human-readable output from `describe`

**Cons:**
- No state tracking — cannot detect drift
- Network, subnets, and routes are managed as separate operations
- No atomic creation — the network exists without subnets between the first and second commands
- Shell scripts are fragile across environments
- No structured output for downstream resource referencing

**Verdict:** Good for ad-hoc operations, debugging, and verifying network state. Not a management solution for production network topologies.

### Level 2: IaC — Terraform

The `hcloud` Terraform provider (`hetznercloud/hcloud ~> 1.60`) provides three separate resources for network management:

```hcl
terraform {
  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "~> 1.60"
    }
  }
}

resource "hcloud_network" "main" {
  name                     = "my-network"
  ip_range                 = "10.0.0.0/16"
  delete_protection        = true
  expose_routes_to_vswitch = false

  labels = {
    environment = "production"
    team        = "platform"
  }
}

resource "hcloud_network_subnet" "cloud_eu" {
  network_id   = hcloud_network.main.id
  type         = "cloud"
  network_zone = "eu-central"
  ip_range     = "10.0.1.0/24"
}

resource "hcloud_network_subnet" "cloud_us" {
  network_id   = hcloud_network.main.id
  type         = "cloud"
  network_zone = "us-east"
  ip_range     = "10.0.2.0/24"
}

resource "hcloud_network_route" "vpn" {
  network_id  = hcloud_network.main.id
  destination = "172.16.0.0/12"
  gateway     = "10.0.1.1"
}

output "network_id" {
  value = hcloud_network.main.id
}
```

**Key provider behaviors:**

- `hcloud_network`: `ip_range` is `ForceNew` — changing it destroys and recreates the network (and all attached subnets, routes, and server attachments).
- `hcloud_network_subnet`: All fields (`network_id`, `type`, `network_zone`, `ip_range`, `vswitch_id`) are `ForceNew`. Any change to a subnet requires replacement. The `gateway` attribute is computed (read-only) — Hetzner assigns the first usable IP in the subnet range.
- `hcloud_network_route`: All fields (`network_id`, `destination`, `gateway`) are `ForceNew`. Any change to a route requires replacement.
- Subnet and route IDs use compound formats: `<network_id>-<ip_range>` for subnets, `<network_id>-<destination>` for routes.

**Pros:**
- State tracking and drift detection across all three resource types
- Plan/apply workflow shows exact infrastructure changes
- Explicit dependency graph (subnets/routes depend on network)
- Version-controlled network topology
- `for_each` enables DRY subnet/route declarations

**Cons:**
- Three separate resources for one logical network — verbose for simple setups
- All subnet/route fields are immutable (ForceNew) — any change is a replacement, not an update
- Subnet replacement disconnects attached servers during the operation
- Must manage inter-resource dependencies manually (or rely on implicit dependency from `network_id`)
- No built-in validation for CIDR overlap between subnets

**Verdict:** Production-grade for Terraform teams. The standard choice before OpenMCF. The three-resource model is correct but verbose.

### Level 3: IaC — Pulumi

The `pulumi-hcloud` SDK (bridged from the Terraform provider) exposes `Network`, `NetworkSubnet`, and `NetworkRoute`:

```go
package main

import (
    "github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        net, err := hcloud.NewNetwork(ctx, "main", &hcloud.NetworkArgs{
            Name:              pulumi.String("my-network"),
            IpRange:           pulumi.String("10.0.0.0/16"),
            DeleteProtection:  pulumi.Bool(true),
            Labels: pulumi.StringMap{
                "environment": pulumi.String("production"),
            },
        })
        if err != nil {
            return err
        }

        _, err = hcloud.NewNetworkSubnet(ctx, "cloud-eu", &hcloud.NetworkSubnetArgs{
            NetworkId:   net.ID().ApplyT(func(id pulumi.ID) (int, error) {
                return strconv.Atoi(string(id))
            }).(pulumi.IntOutput),
            Type:        pulumi.String("cloud"),
            NetworkZone: pulumi.String("eu-central"),
            IpRange:     pulumi.String("10.0.1.0/24"),
        })
        if err != nil {
            return err
        }

        _, err = hcloud.NewNetworkRoute(ctx, "vpn", &hcloud.NetworkRouteArgs{
            NetworkId:   net.ID().ApplyT(func(id pulumi.ID) (int, error) {
                return strconv.Atoi(string(id))
            }).(pulumi.IntOutput),
            Destination: pulumi.String("172.16.0.0/12"),
            Gateway:     pulumi.String("10.0.1.1"),
        })
        if err != nil {
            return err
        }

        ctx.Export("networkId", net.ID())
        return nil
    })
}
```

**A notable friction point:** The `Network` resource's `ID()` returns `IDOutput` (a string), but `NetworkSubnetArgs.NetworkId` and `NetworkRouteArgs.NetworkId` expect `IntInput`. This requires an `ApplyT` conversion at every reference site. The OpenMCF Pulumi module handles this once in a shared variable.

**Pros:**
- Full programming language (Go, TypeScript, Python)
- Type safety catches field name errors at compile time
- Programmatic subnet generation (e.g., loop over a list of zones)
- Built-in secret management for tokens
- Explicit dependency tracking via output references

**Cons:**
- The ID type mismatch (`string` ID vs `int` NetworkId) adds boilerplate
- Three separate resource types for one logical network
- More verbose than HCL for static subnet lists
- Smaller community for Hetzner Cloud specifically

**Verdict:** Excellent for Go/TypeScript teams and dynamic network generation. OpenMCF uses Pulumi (Go) internally for its IaC modules.

## Comparative Analysis

| Method | State Tracking | Drift Detection | Atomic Apply | Audit Trail | CIDR Validation |
|--------|---------------|-----------------|-------------|-------------|-----------------|
| Console | No | No | Yes (single step) | Minimal | UI-level |
| CLI | No | No | No (multi-step) | No | API-level |
| Terraform | Yes | Yes | Yes (plan/apply) | Via VCS | API-level |
| Pulumi | Yes | Yes | Yes (preview/up) | Via VCS | API-level |
| **OpenMCF** | **Yes** | **Yes** | **Yes** | **Via VCS** | **Proto + CEL** |

The key differentiators for OpenMCF:

1. **Single manifest**: One YAML file declares the network, all subnets, and all routes. Terraform and Pulumi require three separate resource declarations with explicit ID wiring.
2. **Proto-level validation**: The CEL rule `vswitch_id_required_for_vswitch_type` catches the most common subnet misconfiguration (missing vSwitch ID for vswitch-type subnets) at validation time, before any API call.
3. **Minimum subnet enforcement**: `buf.validate` requires `min_items: 1` for subnets, preventing the creation of an unusable network (one without subnets). Terraform and Pulumi allow creating a bare network with zero subnets.

## The OpenMCF Approach

### Manifest Format

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudNetwork
metadata:
  name: my-network
spec:
  ipRange: "10.0.0.0/16"
  subnets:
    - type: cloud
      networkZone: eu-central
      ipRange: "10.0.1.0/24"
    - type: cloud
      networkZone: us-east
      ipRange: "10.0.2.0/24"
  routes:
    - destination: "172.16.0.0/12"
      gateway: "10.0.1.1"
  deleteProtection: true
```

### What OpenMCF Automates

1. **Naming:** The network name in Hetzner Cloud is derived from `metadata.name` — no separate `name` field in the spec.
2. **Labeling:** Standard labels (`resource`, `name`, `kind`, `org`, `env`, `id`) are computed from metadata and merged with user-specified labels.
3. **Resource wiring:** The IaC modules create the network first, extract its numeric ID, and pass it to every subnet and route resource automatically. Users never handle IDs.
4. **Provider configuration:** Hetzner Cloud API token is resolved from provider config or environment variables, not hardcoded in the manifest.
5. **Dual IaC:** The same manifest drives both Pulumi and Terraform backends.
6. **Output referencing:** The `network_id` output feeds into `HetznerCloudServer` and `HetznerCloudLoadBalancer` via `StringValueOrRef`, enabling declarative composition without hardcoded IDs.

### The 80/20 Principle

The Hetzner Cloud network API has several attributes. OpenMCF's `HetznerCloudNetworkSpec` exposes the attributes that matter for 80% of use cases.

**Included:**
- `ip_range` — The top-level CIDR block. This is the network's identity and must be one of the RFC 1918 private ranges.
- `subnets` (repeated, min 1) — Subnets with type, zone, CIDR, and optional vSwitch ID. At least one is required because a network without subnets cannot attach any resources.
- `routes` (repeated, optional) — Static routes for VPN gateways, NAT instances, or inter-network routing. Optional because default routing handles most use cases.
- `delete_protection` — Prevents accidental deletion via the API. Important for production networks that anchor entire environments.
- `expose_routes_to_vswitch` — Controls whether network routes are visible to vSwitch connections. Only relevant when a vSwitch subnet is present.

**Handled by the platform:**
- `name` — Derived from `metadata.name`. Consistent naming across all components.
- `labels` — Computed from metadata (org, env, kind, id) with user labels merged in. Consistent labeling across all Hetzner Cloud resources.

**Deliberately excluded:**
- No equivalent to AWS VPC features (NAT gateways, internet gateways, VPC peering, transit gateways, endpoint services). Hetzner Cloud's network model does not have these concepts. The network is a simple private overlay — internet access goes through the server's public interface, not through the network.

### API Design Decisions

**Subnets as a repeated message, not a separate component:** The `HetznerCloudNetworkSpec` contains `repeated Subnet subnets` rather than having a separate `HetznerCloudNetworkSubnet` component. This is deliberate:

1. A Hetzner Cloud network is unusable without subnets — servers and load balancers attach to subnets, not to the network directly.
2. Subnet CIDR ranges must not overlap and must fall within the network's `ip_range`. Validating this constraint is natural when subnets are declared alongside the network.
3. Subnets reference the network by ID. Making them a separate component would require every subnet manifest to include a `valueFrom` reference to the network — boilerplate for a relationship that is always parent-child.

The same reasoning applies to routes.

**Minimum one subnet:** The `min_items: 1` validation on `subnets` prevents creating a network that cannot be used. While Terraform and Pulumi allow bare networks, deploying one serves no purpose — it consumes a network resource but cannot host any workloads until a subnet is added.

**CEL validation for vSwitch subnets:** The CEL rule `vswitch_id_required_for_vswitch_type` enforces that a `vswitch_id` is provided when the subnet type is `vswitch`. The Hetzner Cloud API rejects vswitch-type subnets without a vSwitch ID, but the error message is opaque. The CEL rule provides a clear, specific message at manifest validation time.

**Subnet type as enum:** The `SubnetType` enum (`cloud`, `server`, `vswitch`) prevents invalid type strings. Terraform accepts any string for the `type` field and rejects invalid values at API call time. OpenMCF rejects them at schema validation.

**Single output (network_id):** The component exports only the network's numeric ID. Individual subnet IDs or gateway IPs are not exported because downstream components (servers, load balancers) reference the network, not specific subnets. Server attachment to a subnet is handled by the Hetzner Cloud API based on the server's requested IP within the network's address space.

## Implementation Landscape

### Resources Created

| IaC Engine | Resource | Count | Description |
|------------|----------|-------|-------------|
| Pulumi | `hcloud.Network` | 1 | Network with top-level CIDR, labels, and protection settings |
| Pulumi | `hcloud.NetworkSubnet` | N (1 per subnet) | Subnets within the network, each in a specific zone |
| Pulumi | `hcloud.NetworkRoute` | M (1 per route) | Static routes within the network |
| Terraform | `hcloud_network` | 1 | Same as Pulumi |
| Terraform | `hcloud_network_subnet` | N (1 per subnet) | Keyed by `ip_range` via `for_each` |
| Terraform | `hcloud_network_route` | M (1 per route) | Keyed by `destination` via `for_each` |

### Resource Naming and Keying

**Pulumi module:** Subnets are named `subnet-{sanitized_cidr}` and routes are named `route-{sanitized_cidr}`. The `sanitizeCidr` helper converts CIDR notation to Pulumi-safe resource names by replacing dots, slashes, and colons with hyphens. For example, subnet `10.0.1.0/24` becomes resource name `subnet-10-0-1-0-24`.

**Terraform module:** Subnets use `for_each` keyed by `ip_range` and routes use `for_each` keyed by `destination`. This means:
- Adding a new subnet adds a resource without affecting existing subnets.
- Removing a subnet removes only that specific resource.
- Changing a subnet's `ip_range` is a key change — Terraform destroys the old subnet and creates a new one.

### ID Type Conversion

The Pulumi hcloud SDK has a type mismatch: `Network.ID()` returns `IDOutput` (a string representation of the numeric ID), but `NetworkSubnetArgs.NetworkId` expects `IntInput`. The module solves this once with an `ApplyT` conversion:

```go
networkIdInt := createdNetwork.ID().ApplyT(func(id pulumi.ID) (int, error) {
    return strconv.Atoi(string(id))
}).(pulumi.IntOutput)
```

This `networkIdInt` is then shared across all subnet and route resource creation calls, avoiding repeated conversion code.

### Dependency Role

`HetznerCloudNetwork` is a **foundation resource** — it has no foreign key dependencies. It is referenced by:

- `HetznerCloudServer` — Servers attach to the network for private connectivity
- `HetznerCloudLoadBalancer` — Load balancers attach to the network to reach backend servers over private IPs

In infra charts, the pattern is:

```
HetznerCloudNetwork (foundation)
  └── network_id output
        ├── HetznerCloudServer.spec (via StringValueOrRef)
        └── HetznerCloudLoadBalancer.spec (via StringValueOrRef)
```

### Label Management

Both IaC modules apply a standard label set to the `hcloud_network` resource (subnets and routes do not support labels in the Hetzner Cloud API):

| Label Key | Source | Example |
|-----------|--------|---------|
| `resource` | Constant | `"true"` |
| `name` | `metadata.name` | `"my-network"` |
| `kind` | Constant | `"HetznerCloudNetwork"` |
| `org` | `metadata.org` | `"my-org"` |
| `env` | `metadata.env` | `"production"` |
| `id` | `metadata.id` | `"hcnet-abc123"` |

User-specified `metadata.labels` are merged in, with standard labels taking precedence in case of key conflicts.

## Production Best Practices

### Choosing the Right CIDR Block

The network's `ip_range` must be one of the RFC 1918 private ranges:

| Range | Size | Typical Use |
|-------|------|-------------|
| `10.0.0.0/8` | 16,777,216 addresses | Large organizations with many subnets. Most common choice for cloud networks. |
| `172.16.0.0/12` | 1,048,576 addresses | Medium deployments. Less commonly used, which can be advantageous for avoiding conflicts with other private networks. |
| `192.168.0.0/16` | 65,536 addresses | Small deployments. Commonly used in home and office networks, which can cause VPN routing conflicts. |

**Recommendation:** Use `10.0.0.0/16` (or narrower within `10.0.0.0/8`) for most deployments. It provides 65,536 addresses — enough for extensive subnet subdivision without risking conflict with corporate VPN ranges that commonly use `192.168.0.0/16`.

### Subnet Sizing Strategy

Plan subnet sizes based on expected resource count per zone:

- `/24` (254 addresses) — Sufficient for most application subnets. Handles up to ~250 servers.
- `/20` (4,094 addresses) — Large subnets for environments with many servers or services.
- `/28` (14 addresses) — Minimal subnets for isolated purposes (management, bastion hosts).

Leave room for growth: it is easier to add new subnets to unused CIDR space than to resize existing subnets (which requires replacement).

### Network Zone Strategy

Hetzner Cloud supports four network zones: `eu-central`, `us-east`, `us-west`, `ap-southeast`. A single network can span multiple zones through subnets in different zones.

**Single-zone deployment:** If all resources are in one Hetzner data center, a single subnet in the corresponding zone is sufficient. Most Hetzner Cloud deployments are in `eu-central` (Falkenstein, Nuremberg, Helsinki).

**Multi-zone deployment:** For geographic redundancy, create subnets in multiple zones. Traffic between zones traverses Hetzner's backbone — latency is higher than intra-zone but no egress charges apply.

### Immutability and Replacement

All subnet and route fields in the Hetzner Cloud API are immutable. In both Terraform and Pulumi, changing any field on a subnet or route triggers resource replacement (destroy + create), not an in-place update. This has operational implications:

- **Subnet replacement disconnects servers:** Replacing a subnet detaches all servers connected to it during the operation. Plan maintenance windows for subnet changes.
- **Network CIDR is immutable:** Changing the network's `ip_range` replaces the entire network, including all subnets, routes, and server attachments. This is a destructive operation.
- **Add, don't modify:** When the addressing scheme needs to change, prefer adding new subnets alongside existing ones, migrating resources, and then removing old subnets.

### Delete Protection

Enable `deleteProtection: true` for production networks. A network anchors servers, load balancers, and potentially vSwitch connections to dedicated servers. Accidental deletion cascades through the entire environment. Delete protection must be explicitly disabled before the network can be removed.

### vSwitch Integration

The `vswitch` subnet type bridges Hetzner Cloud and Hetzner Robot (dedicated servers). Before creating a vswitch subnet:

1. Ensure you have a Robot vSwitch set up with the correct VLAN ID.
2. The `vswitchId` in the subnet must match the Robot vSwitch ID.
3. The subnet's `ipRange` must not conflict with IPs already assigned to the vSwitch.
4. The `exposeRoutesToVswitch` network flag controls whether custom routes are visible to dedicated servers — enable it if dedicated servers need to reach route destinations.

vSwitch integration is an advanced use case. Most deployments use only `cloud`-type subnets.

### Avoiding CIDR Conflicts

Within a network:
- Subnet `ipRange` values must not overlap with each other.
- Route `destination` values must not overlap with subnet `ipRange` values.
- The first IP of the network's `ipRange` (e.g., `10.0.0.0` for `10.0.0.0/16`) is reserved.
- The gateway `172.31.1.1` is reserved for the public network interface.

Across networks:
- If servers are attached to multiple networks, ensure the CIDR blocks of those networks do not overlap. Overlapping ranges cause ambiguous routing on the server.

### Static Routes

Custom routes are optional and unnecessary for most deployments. Default routing delivers traffic between subnets within the same network automatically. Use static routes when:

- **VPN gateway:** Route traffic for a remote network (e.g., `172.16.0.0/12`) through a VPN server within the network.
- **NAT instance:** Route internet-bound traffic from private-only servers through a NAT instance.
- **Inter-network routing:** Route traffic between two Hetzner Cloud networks via a dual-homed server.

The route's `gateway` must be an IP address within one of the network's subnets. The gateway server must have IP forwarding enabled at the OS level.

## References

- [Hetzner Cloud Networks Documentation](https://docs.hetzner.cloud/#networks)
- [Hetzner Cloud API — Networks](https://docs.hetzner.cloud/#networks-get-all-networks)
- [Hetzner Cloud API — Network Subnets](https://docs.hetzner.cloud/#network-actions-add-a-subnet-to-a-network)
- [Hetzner Cloud API — Network Routes](https://docs.hetzner.cloud/#network-actions-add-a-route-to-a-network)
- [Terraform hcloud_network Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/network)
- [Terraform hcloud_network_subnet Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/network_subnet)
- [Terraform hcloud_network_route Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/network_route)
- [Pulumi hcloud.Network Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/network/)
- [Pulumi hcloud.NetworkSubnet Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/networksubnet/)
- [Pulumi hcloud.NetworkRoute Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/networkroute/)

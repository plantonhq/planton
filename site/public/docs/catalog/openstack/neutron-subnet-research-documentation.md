---
title: "Neutron Subnet -- Research Documentation"
description: "Neutron Subnet -- Research Documentation deployment documentation"
icon: "package"
order: 100
componentName: "openstacksubnet"
---

# OpenStack Neutron Subnet -- Research Documentation

## Introduction

OpenStack Neutron subnets provide the IP address management layer within a network. While a network is an isolated Layer 2 broadcast domain, a subnet defines the Layer 3 addressing: CIDR block, gateway, DHCP, DNS, and IP allocation ranges. Together, networks and subnets form the foundation of every OpenStack deployment's connectivity model.

A single network can host multiple subnets (e.g., an IPv4 and an IPv6 subnet on the same network, or multiple IPv4 subnets for different address ranges). Each port on the network receives an IP address from one or more of the network's subnets.

## Historical Context

**Early Neutron subnets (2012-2014):** Neutron's subnet model was established early and has remained remarkably stable. The core API (`/v2.0/subnets`) has seen additive changes but no breaking revisions. This stability reflects a well-designed abstraction that maps naturally to how networks work.

**IPAM evolution (2015-present):** Neutron's built-in IPAM handles IP allocation from subnets. The `allocation_pools` feature allows carving the CIDR into allocatable ranges, reserving portions for static assignment or network appliances. External IPAM drivers (Infoblox, etc.) can replace the built-in allocator, but the subnet API remains the same.

**IPv6 support (2014-present):** IPv6 subnets added `ipv6_address_mode` and `ipv6_ra_mode` for controlling Stateless Address Auto-Configuration (SLAAC) and DHCPv6. These are excluded from our 80/20 spec but can be added when ARM requires IPv6 deployments.

**Subnet pools (2015-present):** `SubnetPool` resources manage CIDR allocation across subnets, preventing overlap and enabling delegated address management. This is an advanced feature excluded from our spec -- users specify CIDRs directly.

## Deployment Methods Landscape

### Level 0: Manual (Horizon Dashboard)

Subnets are typically created as part of the network creation wizard in Horizon:

1. Navigate to **Project > Network > Networks**
2. Click **Create Network** or select an existing network
3. In the **Subnet** tab, enter CIDR, gateway, DHCP settings
4. In the **Subnet Details** tab, configure DNS and allocation pools
5. Click **Create**

**Pros:**
- Visual, subnet wizard is intuitive
- Immediate feedback on IP allocation

**Cons:**
- Not reproducible or auditable
- Easy to create overlapping CIDRs
- Cannot enforce naming conventions

**Verdict:** Good for learning. Not suitable for production.

### Level 1: CLI (openstack client)

```bash
# Create a basic subnet
openstack subnet create dev-subnet \
  --network dev-network \
  --subnet-range 192.168.1.0/24

# Create with full options
openstack subnet create prod-subnet \
  --network prod-network \
  --subnet-range 10.0.0.0/16 \
  --gateway 10.0.0.1 \
  --dns-nameserver 8.8.8.8 \
  --dns-nameserver 8.8.4.4 \
  --allocation-pool start=10.0.1.0,end=10.0.254.255 \
  --description "Production application subnet"

# Create without gateway (isolated)
openstack subnet create storage-subnet \
  --network storage-network \
  --subnet-range 172.16.0.0/24 \
  --no-gateway

# Create IPv6 subnet
openstack subnet create ipv6-subnet \
  --network dual-stack-network \
  --subnet-range 2001:db8::/64 \
  --ip-version 6
```

**Pros:**
- Scriptable, full control over all parameters
- Good for quick operations and debugging

**Cons:**
- No state tracking, no drift detection
- Manual dependency management (network must exist first)

**Verdict:** Good for ad-hoc operations. Not recommended for managing infrastructure at scale.

### Level 2: IaC -- Terraform

```hcl
resource "openstack_networking_subnet_v2" "app_subnet" {
  name       = "app-subnet"
  network_id = openstack_networking_network_v2.app_network.id
  cidr       = "10.0.0.0/16"
  ip_version = 4

  gateway_ip  = "10.0.0.1"
  enable_dhcp = true

  dns_nameservers = ["8.8.8.8", "8.8.4.4"]

  allocation_pool {
    start = "10.0.1.0"
    end   = "10.0.254.255"
  }

  tags = ["env:production", "managed-by:terraform"]
}
```

**Pros:**
- Declarative, reproducible
- Automatic dependency ordering (subnet waits for network)
- State tracking and drift detection

**Cons:**
- HCL syntax required
- State management overhead
- No built-in FK validation across resources

**Verdict:** Industry standard. OpenMCF wraps this with a unified API and FK-based dependency management.

### Level 3: IaC -- Pulumi

```go
subnet, err := networking.NewSubnet(ctx, "app-subnet", &networking.SubnetArgs{
    Name:      pulumi.String("app-subnet"),
    NetworkId: network.ID(),
    Cidr:      pulumi.String("10.0.0.0/16"),
    IpVersion: pulumi.Int(4),
    GatewayIp: pulumi.StringPtr("10.0.0.1"),
    EnableDhcp: pulumi.BoolPtr(true),
    DnsNameservers: pulumi.StringArray{
        pulumi.String("8.8.8.8"),
        pulumi.String("8.8.4.4"),
    },
    AllocationPools: networking.SubnetAllocationPoolArray{
        &networking.SubnetAllocationPoolArgs{
            Start: pulumi.String("10.0.1.0"),
            End:   pulumi.String("10.0.254.255"),
        },
    },
})
```

**Pros:**
- Type-safe, compile-time error detection
- Native dependency tracking via resource references
- Real programming language capabilities

**Cons:**
- More verbose for simple resources
- Requires Go/Python/TypeScript knowledge

**Verdict:** Excellent for engineering teams. OpenMCF uses Pulumi internally for Go-native IaC modules.

## The OpenMCF Approach

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSubnet
metadata:
  name: app-subnet
spec:
  network_id:
    value_from:
      name: app-network
  cidr: "10.0.0.0/16"
  dns_nameservers:
    - "8.8.8.8"
    - "8.8.4.4"
```

### What OpenMCF Automates

1. **Foreign key resolution**: `network_id.value_from` automatically resolves to the network's UUID from its stack outputs
2. **Dependency ordering**: In InfraCharts, the subnet waits for the network to complete before deployment
3. **Provider configuration**: Credentials resolved from the platform's credential store
4. **Dual IaC engines**: Same manifest works with both Pulumi and Terraform backends
5. **Validation**: CIDR format, IP version, gateway mutual exclusion validated at the API level before deployment

### The 80/20 Principle

The Terraform `openstack_networking_subnet_v2` resource exposes 22 schema fields. OpenMCF's `OpenStackSubnetSpec` exposes 11 fields -- the ones that cover 95%+ of real-world use cases:

**Included:**
- `network_id` -- Parent network (FK)
- `cidr` -- IP address range
- `ip_version` -- IPv4 or IPv6 (default: 4)
- `gateway_ip` -- Custom gateway IP
- `no_gateway` -- Disable gateway entirely
- `enable_dhcp` -- DHCP toggle (default: true)
- `dns_nameservers` -- DNS servers pushed via DHCP
- `allocation_pools` -- IP allocation sub-ranges
- `description` -- Human-readable description
- `tags` -- Resource tagging
- `region` -- Region override

**Excluded:**
- `prefix_length` -- Alternative to CIDR for subnetpool-based allocation; niche
- `subnetpool_id` -- Subnetpool integration; niche
- `tenant_id` -- Admin-only
- `ipv6_address_mode`, `ipv6_ra_mode` -- IPv6-specific; can add later when ARM needs IPv6
- `dns_publish_fixed_ip` -- Niche DNS feature
- `segment_id` -- Multi-segment networks; niche
- `value_specs` -- Vendor escape hatch
- `service_types` -- Niche

### API Design Decisions

**`network_id` as StringValueOrRef FK:** This is the first OpenStack component with a foreign key. Using `StringValueOrRef` enables both literal UUIDs (`value: "..."`) and references to managed resources (`value_from: {name: "..."}`). The FK annotations (`default_kind: OpenStackNetwork`, `default_kind_field_path: "status.outputs.network_id"`) enable automatic resolution in InfraChart pipelines.

**`cidr` as required string:** Since we exclude `subnetpool_id` and `prefix_length`, CIDR is the only way to define the subnet's IP range. Making it required prevents users from accidentally creating a subnet with no address space.

**`optional int32 ip_version` with default 4:** Proto3 int32 defaults to 0, which is not a valid IP version. Using `optional` with default `"4"` ensures IPv4 is used unless explicitly overridden to IPv6.

**`gateway_ip` + `no_gateway` mutual exclusion:** Following the Terraform provider pattern exactly. A message-level CEL expression enforces this at validation time. Users get one of three behaviors: (1) omit both for auto-assigned gateway, (2) set `gateway_ip` for specific gateway, (3) set `no_gateway: true` for gatewayless subnet.

**`enable_dhcp` with default true:** Proto3 bool defaults to false, but DHCP enabled is the correct default for virtually all subnets. The `optional` + default pattern prevents the foot-gun of accidentally creating DHCP-disabled subnets.

## Implementation Landscape

### Resources Created

| Resource | Count | Description |
|----------|-------|-------------|
| `openstack_networking_subnet_v2` (TF) / `networking.Subnet` (Pulumi) | 1 | The Neutron subnet |

Single-resource component. Atomic and composable in InfraCharts.

### Dependency Role

OpenStackSubnet has one inbound FK and is referenced by multiple downstream components:

**References:**
- `network_id` -> `OpenStackNetwork.status.outputs.network_id`

**Referenced by:**
- `OpenStackRouterInterface.spec.subnet_id` -> `subnet_id`
- `OpenStackLoadBalancer.spec.vip_subnet_id` -> `subnet_id`
- `OpenStackLoadBalancerMember.spec.subnet_id` -> `subnet_id`
- `OpenStackContainerClusterTemplate.spec.fixed_subnet` -> `subnet_id`

This makes `subnet_id` the second most important output in the OpenStack component family (after `network_id`).

## Production Best Practices

### CIDR Planning

- Plan your CIDR ranges before creating subnets -- overlapping ranges within the same router cause routing conflicts
- Use `/24` for small subnets (254 hosts), `/16` for large ones (65,534 hosts)
- Reserve the first and last few IPs in each subnet for network infrastructure (gateways, DHCP agents)
- Document your IP allocation plan; OpenStack will not warn you about poor planning

### DHCP

- Keep DHCP enabled unless you have a specific reason to disable it (e.g., bare-metal with static IPs)
- OpenStack's DHCP agent runs on the Neutron network node; HA deployments run agents on multiple nodes
- DHCP leases are tied to ports, not instances -- a port retains its IP across instance rebuilds

### DNS

- Always specify `dns_nameservers` for subnets that host user workloads
- Internal DNS servers should be listed before external ones for faster resolution of internal names
- For OpenStack deployments with Designate, DNS integration can auto-register instance DNS names

### Allocation Pools

- Use allocation pools to carve out reserved ranges (e.g., `10.0.0.1-10.0.0.50` for network appliances)
- Multiple pools can be specified; they must not overlap with each other or the gateway IP
- If no pools are specified, the entire CIDR minus gateway and broadcast addresses is allocatable

### No-Gateway Subnets

- Use `no_gateway` for storage networks, inter-VM-only communication, or any subnet that does not need external routing
- Instances on no-gateway subnets cannot reach the internet or other subnets via routing (but can still communicate within the same subnet)

## References

- [Terraform openstack_networking_subnet_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/networking_subnet_v2)
- [Pulumi openstack.networking.Subnet](https://www.pulumi.com/registry/packages/openstack/api-docs/networking/subnet/)
- [OpenStack Neutron API -- Subnets](https://docs.openstack.org/api-ref/network/v2/#subnets)
- [OpenStack Networking Guide -- Subnets](https://docs.openstack.org/neutron/latest/admin/)
- [RFC 4632 -- CIDR](https://tools.ietf.org/html/rfc4632)

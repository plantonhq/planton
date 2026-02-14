---
title: "Octavia Load Balancer -- Research Documentation"
description: "Octavia Load Balancer -- Research Documentation deployment documentation"
icon: "package"
order: 100
componentName: "openstackloadbalancer"
---

# OpenStack Octavia Load Balancer -- Research Documentation

## Introduction

OpenStack Octavia is the load balancing service that replaced the legacy Neutron LBaaS v2 extension. It provides production-grade, scalable load balancing as a first-class OpenStack service. An Octavia load balancer creates a Virtual IP (VIP) on a specified subnet, serving as the entry point for all traffic distribution to backend pools.

The load balancer is the root object in Octavia's hierarchy: Load Balancer -> Listener -> Pool -> Member -> Health Monitor. Each level in this hierarchy is a separate resource, enabling flexible composition of traffic distribution topologies.

## Historical Context

**Neutron LBaaS v1 (2013-2015):** The original load balancing implementation was a Neutron extension with a simplified API. It supported basic Layer 4 load balancing but lacked advanced features like TLS termination, L7 policies, and scalable amphora management. It was deprecated in favor of LBaaS v2.

**Neutron LBaaS v2 (2015-2018):** LBaaS v2 introduced the hierarchical model (loadbalancer -> listener -> pool -> member) that Octavia inherits. However, it still ran as a Neutron extension and shared Neutron's agent-based architecture, limiting scalability. The API surface was identical to early Octavia.

**Octavia (2017-present):** Octavia became a standalone project, running its own control plane with dedicated amphora VMs (or OVN-based provider drivers). Key improvements:
- **Scalability**: Each load balancer runs on one or more dedicated VMs (amphorae), eliminating the bottleneck of shared agents
- **HA**: Active-standby amphora pairs with VRRP failover
- **Provider drivers**: Pluggable backends (amphora, OVN, F5, A10, etc.)
- **Flavors**: Resource profiles controlling amphora size, topology, and provider
- **Tags**: Resource tagging for organization and filtering

The `openstack_lb_loadbalancer_v2` Terraform resource and `loadbalancer.LoadBalancer` Pulumi resource both target the Octavia API (which uses the same v2 API path as legacy LBaaS v2 for backward compatibility).

## Architecture

### Octavia Object Model

```
OpenStackLoadBalancer (VIP on subnet)
  └── OpenStackLoadBalancerListener (protocol + port)
        └── OpenStackLoadBalancerPool (algorithm + members)
              ├── OpenStackLoadBalancerMember (backend server)
              └── OpenStackLoadBalancerMonitor (health check)
```

### VIP Networking

The load balancer's VIP is allocated on the specified subnet. This means:
1. The VIP is a Neutron port on the subnet's network
2. Instances on the same network can reach the VIP directly
3. For external access, a floating IP must be associated with the VIP port
4. Security groups can be applied to the VIP port for access control

### Amphora Architecture

In the default amphora provider:
- Each load balancer is backed by one or more amphora VMs
- Active-standby topology provides HA (VRRP between amphorae)
- The amphora runs HAProxy to handle traffic distribution
- Configuration is pushed to the amphora via the Octavia controller
- Amphorae are placed on the load balancer's management network

### OVN Provider

For lightweight load balancing:
- Traffic is handled by OVN's built-in load balancer (no amphora VMs)
- Lower latency and resource consumption
- Limited feature set (no L7 policies, no TLS termination)
- Ideal for simple TCP/UDP load balancing

## Deployment Methods Landscape

### Level 0: Manual (Horizon Dashboard)

Load balancers are created via the Horizon Load Balancer panel:

1. Navigate to **Project > Network > Load Balancers**
2. Click **Create Load Balancer**
3. Select the subnet for VIP allocation
4. Optionally specify a VIP address
5. Add listeners, pools, and members in the wizard
6. Click **Create**

**Pros:**
- Visual wizard guides through the full hierarchy
- Immediate feedback on VIP allocation

**Cons:**
- Not reproducible or auditable
- Wizard creates all levels at once, making troubleshooting harder
- Cannot enforce naming conventions or tagging standards

**Verdict:** Good for learning and one-off testing. Not suitable for production.

### Level 1: CLI (openstack client)

```bash
# Create a basic load balancer
openstack loadbalancer create \
  --name dev-lb \
  --vip-subnet-id e0a1f622-9aab-4a48-8c8c-3b0c7e2a9b1d

# Create with specific VIP address
openstack loadbalancer create \
  --name prod-lb \
  --vip-subnet-id e0a1f622-9aab-4a48-8c8c-3b0c7e2a9b1d \
  --vip-address 10.0.0.100 \
  --description "Production load balancer"

# Create with flavor
openstack loadbalancer create \
  --name premium-lb \
  --vip-subnet-id e0a1f622-9aab-4a48-8c8c-3b0c7e2a9b1d \
  --flavor a1b2c3d4-e5f6-7890-abcd-ef1234567890

# Check provisioning status
openstack loadbalancer show prod-lb -f value -c provisioning_status

# Wait for ACTIVE status before adding listeners
openstack loadbalancer wait prod-lb
```

**Pros:**
- Scriptable, full control over all parameters
- Good for quick operations and debugging
- Can wait for provisioning to complete

**Cons:**
- No state tracking, no drift detection
- Must manually wait for ACTIVE status before creating child resources
- Manual dependency management across the LB hierarchy

**Verdict:** Good for ad-hoc operations. Not recommended for managing infrastructure at scale.

### Level 2: IaC -- Terraform

```hcl
resource "openstack_lb_loadbalancer_v2" "app_lb" {
  name          = "app-lb"
  vip_subnet_id = openstack_networking_subnet_v2.app_subnet.id
  description   = "Application load balancer"

  tags = ["env:production", "managed-by:terraform"]
}

# Must wait for LB to become ACTIVE before creating listener
resource "openstack_lb_listener_v2" "http" {
  name            = "http-listener"
  loadbalancer_id = openstack_lb_loadbalancer_v2.app_lb.id
  protocol        = "HTTP"
  protocol_port   = 80
}
```

**Pros:**
- Declarative, reproducible
- Automatic dependency ordering
- State tracking and drift detection
- Handles the ACTIVE status wait internally

**Cons:**
- HCL syntax required
- State management overhead
- No built-in FK validation across resources

**Verdict:** Industry standard. OpenMCF wraps this with a unified API and FK-based dependency management.

### Level 3: IaC -- Pulumi

```go
lb, err := loadbalancer.NewLoadBalancer(ctx, "app-lb", &loadbalancer.LoadBalancerArgs{
    Name:        pulumi.String("app-lb"),
    VipSubnetId: subnet.ID(),
    Description: pulumi.StringPtr("Application load balancer"),
    Tags: pulumi.StringArray{
        pulumi.String("env:production"),
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
kind: OpenStackLoadBalancer
metadata:
  name: app-lb
spec:
  vip_subnet_id:
    value_from:
      name: app-subnet
  description: "Application load balancer"
```

### What OpenMCF Automates

1. **Foreign key resolution**: `vip_subnet_id.value_from` automatically resolves to the subnet's UUID from its stack outputs
2. **Dependency ordering**: In InfraCharts, the load balancer waits for the subnet to complete before deployment
3. **Provider configuration**: Credentials resolved from the platform's credential store
4. **Dual IaC engines**: Same manifest works with both Pulumi and Terraform backends
5. **Validation**: Required fields and tag uniqueness validated at the API level before deployment

### The 80/20 Principle

The Terraform `openstack_lb_loadbalancer_v2` resource exposes 14 schema fields. OpenMCF's `OpenStackLoadBalancerSpec` exposes 7 fields -- the ones that cover 95%+ of real-world use cases:

**Included:**
- `vip_subnet_id` -- VIP subnet (FK)
- `vip_address` -- Specific VIP address
- `description` -- Human-readable description
- `admin_state_up` -- Administrative state (default: true)
- `flavor_id` -- Octavia flavor for resource limits
- `tags` -- Resource tagging
- `region` -- Region override

**Excluded:**
- `vip_network_id` -- Alternative to `vip_subnet_id`; using subnet is more precise
- `vip_port_id` -- Pre-created port for the VIP; niche
- `tenant_id` -- Admin-only
- `security_group_ids` -- Can be managed via separate security group attachment
- `loadbalancer_provider` -- Use flavors instead for provider selection
- `availability_zone` -- Niche; most deployments use the default AZ
- `value_specs` -- Vendor escape hatch

### API Design Decisions

**`vip_subnet_id` as the sole required FK:** We use `vip_subnet_id` rather than `vip_network_id` because specifying the subnet is more precise and deterministic. When a network has multiple subnets, `vip_subnet_id` ensures the VIP lands on the intended subnet. The Terraform resource supports either, but we choose the more explicit option.

**`optional bool admin_state_up` with default true:** Proto3 bool defaults to false, but an active load balancer is the correct default. The `optional` + default pattern prevents accidentally creating disabled load balancers.

**`flavor_id` as a plain string:** Flavors are typically pre-created by cloud operators. Since they're not managed by OpenMCF, a plain string (not StringValueOrRef) is appropriate.

**Single-resource component:** Unlike some frameworks that create the entire LB hierarchy at once, OpenMCF keeps each level as a separate component. This enables:
- Independent lifecycle management (update a listener without recreating the LB)
- Flexible composition in InfraCharts
- Clear FK chains for dependency resolution

## Implementation Landscape

### Resources Created

| Resource | Count | Description |
|----------|-------|-------------|
| `openstack_lb_loadbalancer_v2` (TF) / `loadbalancer.LoadBalancer` (Pulumi) | 1 | The Octavia load balancer |

Single-resource component. Atomic and composable in InfraCharts.

### Dependency Role

OpenStackLoadBalancer has one inbound FK and is referenced by the listener:

**References:**
- `vip_subnet_id` -> `OpenStackSubnet.status.outputs.subnet_id`

**Referenced by:**
- `OpenStackLoadBalancerListener.spec.loadbalancer_id` -> `loadbalancer_id`

This makes `loadbalancer_id` the primary output, used as a foreign key by all child resources in the Octavia hierarchy.

## Production Best Practices

### VIP Planning

- Plan your VIP addresses when using shared subnets -- use `vip_address` to ensure a predictable IP
- For external access, always associate a floating IP with the VIP port
- Document your VIP allocations to avoid conflicts with DHCP-allocated addresses

### Flavors

- Use flavors to control resource allocation (amphora size, topology)
- The default flavor is typically "small" -- for production traffic, request a larger flavor from your cloud operator
- Flavors are immutable on existing load balancers (ForceNew) -- changing requires recreation

### High Availability

- Octavia creates active-standby amphora pairs by default for HA
- The VRRP failover typically completes in 5-10 seconds
- For zero-downtime failovers, configure health monitors on your pools
- Monitor the `operating_status` and `provisioning_status` in OpenStack

### Security

- Apply security groups to the VIP port (via `vip_port_id` output) to restrict access
- Use allowed_cidrs on listeners for additional access control
- For TLS termination, use TERMINATED_HTTPS on the listener with a Barbican secret

### Monitoring

- Monitor VIP reachability from outside the subnet
- Check Octavia's health manager for amphora health
- Set up alerts for `operating_status` changes (ONLINE -> ERROR)
- Track connection counts and bandwidth via Octavia's statistics API

## References

- [Terraform openstack_lb_loadbalancer_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/lb_loadbalancer_v2)
- [Pulumi openstack.loadbalancer.LoadBalancer](https://www.pulumi.com/registry/packages/openstack/api-docs/loadbalancer/loadbalancer/)
- [OpenStack Octavia API -- Load Balancers](https://docs.openstack.org/api-ref/load-balancer/v2/)
- [OpenStack Octavia User Guide](https://docs.openstack.org/octavia/latest/user/guides/)
- [Octavia Cookbook](https://docs.openstack.org/octavia/latest/user/guides/basic-cookbook.html)

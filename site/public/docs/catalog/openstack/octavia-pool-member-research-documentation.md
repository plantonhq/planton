---
title: "Octavia Pool Member -- Research Documentation"
description: "Octavia Pool Member -- Research Documentation deployment documentation"
icon: "package"
order: 100
componentName: "openstackloadbalancermember"
---

# OpenStack Octavia Pool Member -- Research Documentation

## Introduction

An Octavia pool member represents a backend server in the load balancer pool. Members are the actual endpoints that receive traffic distributed by the pool's load-balancing algorithm. Each member defines an IP address and port, along with optional weight and subnet configuration.

Members are the leaf nodes in Octavia's hierarchy: Load Balancer -> Listener -> Pool -> Member. The health monitor (also attached to the pool) determines whether individual members are healthy and should receive traffic.

## Historical Context

**Neutron LBaaS v1 (2013-2015):** Members were called "pool members" and were tightly coupled to the pool. No weight support, no subnet awareness, no tags.

**Neutron LBaaS v2 / Octavia (2015-present):** Members gained weight support for weighted load distribution, subnet_id for cross-subnet routing, and monitor_address/monitor_port for health check customization. Tags were added in later releases.

Key improvements:
- **Weight**: 0-256 range for fine-grained traffic distribution
- **Subnet awareness**: Cross-subnet routing via Octavia's L3 capabilities
- **Backup members**: backup=true for failover-only members (not exposed in this component)
- **Tags**: Resource tagging for organization and filtering

## Architecture

### Member Position in the Octavia Hierarchy

```
OpenStackLoadBalancer (VIP on subnet)
  +-- OpenStackLoadBalancerListener (protocol + port)
        +-- OpenStackLoadBalancerPool (algorithm + protocol)
              +-- OpenStackLoadBalancerMember (backend server)   <-- this component
              +-- OpenStackLoadBalancerMonitor (health check)
```

### Traffic Flow

1. Client sends request to the load balancer's VIP
2. Listener receives the request on its protocol/port
3. Pool selects a member using the configured algorithm
4. Traffic is forwarded to the member's address:protocol_port
5. If a health monitor is configured, unhealthy members are excluded

### Weight-Based Distribution

With ROUND_ROBIN and weights:
- Member A (weight 10): receives ~50% of traffic
- Member B (weight 5): receives ~25% of traffic
- Member C (weight 5): receives ~25% of traffic

Weight 0 means the member receives no traffic (drain mode). This is useful for graceful removal of backends during maintenance.

### Cross-Subnet Routing

When subnet_id is specified, Octavia routes traffic to the member via that subnet's gateway. This enables backends on different subnets than the VIP. Without subnet_id, Octavia assumes the member is on the same subnet as the VIP.

## Deployment Methods Landscape

### Level 0: Manual (Horizon Dashboard)

Members are added through the load balancer wizard or the member management panel. Good for learning. Not suitable for production.

### Level 1: CLI (openstack client)

```bash
# Add a member to a pool
openstack loadbalancer member create \
  --name web-backend-1 \
  --address 10.0.0.10 \
  --protocol-port 8080 \
  --weight 10 \
  web-pool

# Add a member with cross-subnet routing
openstack loadbalancer member create \
  --name cross-subnet-member \
  --address 10.1.0.10 \
  --protocol-port 8080 \
  --subnet-id b2c3d4e5-f6a7-8901-bcde-f12345678901 \
  web-pool

# Drain a member (set weight to 0)
openstack loadbalancer member set --weight 0 web-pool web-backend-1
```

Good for ad-hoc operations. Not recommended for managing infrastructure at scale.

### Level 2: IaC -- Terraform

```hcl
resource "openstack_lb_member_v2" "web_1" {
  name          = "web-backend-1"
  pool_id       = openstack_lb_pool_v2.web.id
  address       = "10.0.0.10"
  protocol_port = 8080
  weight        = 10
  subnet_id     = openstack_networking_subnet_v2.backend.id
}
```

Industry standard. OpenMCF wraps this with FK-based references and validation.

### Level 3: IaC -- Pulumi

```go
member, err := loadbalancer.NewMember(ctx, "web-1", &loadbalancer.MemberArgs{
    Name:         pulumi.String("web-backend-1"),
    PoolId:       pool.ID(),
    Address:      pulumi.String("10.0.0.10"),
    ProtocolPort: pulumi.Int(8080),
    Weight:       pulumi.IntPtr(10),
})
```

Excellent for engineering teams.

## The OpenMCF Approach

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMember
metadata:
  name: web-backend-1
spec:
  pool_id:
    value_from:
      name: web-pool
  address: "10.0.0.10"
  protocol_port: 8080
  weight: 10
```

### What OpenMCF Automates

1. **Foreign key resolution**: pool_id.value_from resolves to the pool UUID
2. **Dependency ordering**: Member waits for the pool to be ready
3. **Validation**: Protocol port range (1-65535), weight range (0-256) validated at the API level
4. **Dual IaC engines**: Same manifest works with Pulumi and Terraform

### The 80/20 Principle

The Terraform openstack_lb_member_v2 resource exposes 13 schema fields. OpenMCF exposes 8 fields:

**Included:** pool_id, address, protocol_port, subnet_id, weight, admin_state_up, tags, region

**Excluded:** tenant_id (admin-only), monitor_address/monitor_port (niche), backup (failover-only), value_specs (vendor escape hatch)

### API Design Decisions

**`optional int32 weight`**: Proto3 int32 defaults to 0, which means "drain" in Octavia. Using optional lets us distinguish "not set" (Octavia picks default 1) from "explicitly 0" (drain mode).

**`address` as plain string**: Member backends can be any IP (VMs, containers, bare metal, external services), so a plain string is more appropriate than StringValueOrRef.

**`pool_id` as separate attribute**: In Terraform, pool_id is a top-level attribute (not a nested object). The TF resource import format is `pool_id/member_id`, highlighting the tight relationship.

## Implementation Landscape

### Resources Created

| Resource | Count | Description |
|----------|-------|-------------|
| openstack_lb_member_v2 (TF) / loadbalancer.Member (Pulumi) | 1 | The Octavia pool member |

### Dependency Role

**References:**
- pool_id -> OpenStackLoadBalancerPool.status.outputs.pool_id
- subnet_id -> OpenStackSubnet.status.outputs.subnet_id (optional)

**Referenced by:** None (leaf node in the Octavia hierarchy)

## Production Best Practices

### Weight Management

- Use weight 0 for graceful drain before removing a member
- Higher weights for more powerful backends
- Equal weights for homogeneous backends (or omit weight entirely)

### Cross-Subnet Routing

- Always specify subnet_id when members are on a different subnet than the VIP
- Ensures Octavia routes traffic correctly via the subnet gateway
- Required for multi-tier architectures (frontend VIP on public subnet, backends on private subnet)

### Health Monitoring

- Always pair pools with health monitors
- Without monitors, failed members continue to receive traffic
- Members marked DOWN by monitors are automatically removed from rotation

### Scaling

- Add/remove members to scale the backend pool
- Weight changes take effect immediately (no drain delay)
- Use admin_state_up=false to temporarily remove a member without deleting it

## References

- [Terraform openstack_lb_member_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/lb_member_v2)
- [Pulumi openstack.loadbalancer.Member](https://www.pulumi.com/registry/packages/openstack/api-docs/loadbalancer/member/)
- [OpenStack Octavia API -- Members](https://docs.openstack.org/api-ref/load-balancer/v2/#members)
- [Octavia Cookbook](https://docs.openstack.org/octavia/latest/user/guides/basic-cookbook.html)

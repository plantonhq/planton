# OpenStack Octavia Pool -- Research Documentation

## Introduction

An Octavia pool is the second-to-last level in Octavia's resource hierarchy (Load Balancer -> Listener -> Pool -> Member). A pool groups backend members (servers) and defines how traffic from the parent listener is distributed across those members. The pool determines the backend protocol and load-balancing algorithm.

Each pool has exactly one parent listener (or can be shared via loadbalancer_id for L7 policy routing, which this component does not expose). Members and health monitors attach directly to the pool.

## Historical Context

**Neutron LBaaS v1 (2013-2015):** Pools were tightly coupled to the VIP. The concept of listeners did not exist -- each pool was directly bound to a protocol and port. This limited load balancers to one protocol per VIP.

**Neutron LBaaS v2 / Octavia (2015-present):** The listener abstraction was introduced, decoupling protocols from pools. A single load balancer can now have multiple listeners, each with its own pool. Pools define the backend protocol independently of the listener frontend protocol, enabling protocol translation (e.g., HTTPS termination on the listener with HTTP to backends).

Key improvements in the pool resource over time:
- **Session persistence**: SOURCE_IP, HTTP_COOKIE, and APP_COOKIE types
- **Additional algorithms**: SOURCE_IP_PORT added for fine-grained stickiness
- **Tags**: Resource tagging support (added in later Octavia releases)
- **PROXY protocol**: Support for PROXY protocol v1/v2 to preserve client IP information

## Architecture

### Pool Position in the Octavia Hierarchy

```
OpenStackLoadBalancer (VIP on subnet)
  +-- OpenStackLoadBalancerListener (protocol + port)
        +-- OpenStackLoadBalancerPool (algorithm + protocol)   <-- this component
              +-- OpenStackLoadBalancerMember (backend server)
              +-- OpenStackLoadBalancerMonitor (health check)
```

### Protocol Translation

The listener protocol defines what clients send; the pool protocol defines what backends receive:

| Listener Protocol | Pool Protocol | Use Case |
|---|---|---|
| HTTP | HTTP | Standard web traffic |
| HTTPS (pass-through) | HTTPS | End-to-end TLS |
| TERMINATED_HTTPS | HTTP | TLS offload at the LB |
| TCP | TCP | Generic TCP services |
| UDP | UDP | DNS, gaming, IoT |
| TCP | PROXY | TCP with client IP preservation |

### Load-Balancing Algorithms

| Algorithm | Description | Best For |
|---|---|---|
| ROUND_ROBIN | Equal distribution across all members | Stateless services, even load |
| LEAST_CONNECTIONS | Route to member with fewest active connections | Long-lived connections |
| SOURCE_IP | Hash client IP for consistent routing | Stateful apps without cookie support |
| SOURCE_IP_PORT | Hash client IP + port | Multi-stream clients |

### Session Persistence

| Type | Mechanism | Use Case |
|---|---|---|
| SOURCE_IP | Hash of client IP address | Layer 4, no cookie support |
| HTTP_COOKIE | Octavia inserts and tracks a cookie | Standard web apps |
| APP_COOKIE | Application manages the cookie | Apps with existing session cookies (e.g., JSESSIONID) |

**APP_COOKIE** requires the `cookie_name` field to specify which application cookie to use for affinity.

## Deployment Methods Landscape

### Level 0: Manual (Horizon Dashboard)

Pools are created through the load balancer wizard in Horizon. Good for learning. Not suitable for production.

### Level 1: CLI (openstack client)

```bash
openstack loadbalancer pool create \
  --name web-pool \
  --listener http-listener \
  --protocol HTTP \
  --lb-algorithm ROUND_ROBIN

openstack loadbalancer pool create \
  --name cookie-pool \
  --listener https-listener \
  --protocol HTTPS \
  --lb-algorithm LEAST_CONNECTIONS \
  --session-persistence type=APP_COOKIE,cookie_name=JSESSIONID
```

Good for debugging. Not recommended for managing infrastructure at scale.

### Level 2: IaC -- Terraform

```hcl
resource "openstack_lb_pool_v2" "web" {
  name        = "web-pool"
  listener_id = openstack_lb_listener_v2.http.id
  protocol    = "HTTP"
  lb_method   = "ROUND_ROBIN"

  persistence {
    type        = "APP_COOKIE"
    cookie_name = "JSESSIONID"
  }
}
```

Industry standard. OpenMCF wraps this with FK-based references.

### Level 3: IaC -- Pulumi

```go
pool, err := loadbalancer.NewPool(ctx, "web-pool", &loadbalancer.PoolArgs{
    Name:       pulumi.String("web-pool"),
    ListenerId: listener.ID(),
    Protocol:   pulumi.String("HTTP"),
    LbMethod:   pulumi.String("ROUND_ROBIN"),
})
```

Excellent for engineering teams.

## The OpenMCF Approach

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerPool
metadata:
  name: web-pool
spec:
  listener_id:
    value_from:
      name: http-listener
  protocol: "HTTP"
  lb_method: "LEAST_CONNECTIONS"
```

### What OpenMCF Automates

1. **Foreign key resolution**: listener_id.value_from resolves to the listener UUID
2. **Dependency ordering**: Pool waits for the listener to be ACTIVE
3. **Validation**: Protocol, lb_method, and persistence type validated at the API level
4. **CEL constraints**: cookie_name validated against persistence type
5. **Dual IaC engines**: Same manifest works with Pulumi and Terraform

### The 80/20 Principle

The Terraform openstack_lb_pool_v2 resource exposes 11 schema fields. OpenMCF exposes 8 fields:

**Included:** listener_id, protocol, lb_method, persistence, description, admin_state_up, tags, region

**Excluded:** loadbalancer_id (L7 policy routing), tenant_id (admin-only), value_specs (vendor escape hatch)

## Implementation Landscape

### Resources Created

| Resource | Count | Description |
|----------|-------|-------------|
| openstack_lb_pool_v2 (TF) / loadbalancer.Pool (Pulumi) | 1 | The Octavia backend pool |

### Dependency Role

**References:** listener_id -> OpenStackLoadBalancerListener.status.outputs.listener_id

**Referenced by:**
- OpenStackLoadBalancerMember.spec.pool_id -> pool_id
- OpenStackLoadBalancerMonitor.spec.pool_id -> pool_id

## Production Best Practices

### Algorithm Selection

- Use ROUND_ROBIN for stateless services with similar-capacity backends
- Use LEAST_CONNECTIONS when request processing times vary significantly
- Use SOURCE_IP when you need session affinity but cannot use cookies
- Use SOURCE_IP_PORT for clients opening multiple connections needing per-connection affinity

### Session Persistence

- Prefer HTTP_COOKIE for web applications when possible
- Use APP_COOKIE when your application already has a session cookie
- Use SOURCE_IP for non-HTTP protocols or when cookies are not an option
- Remember that SOURCE_IP affinity breaks when clients are behind NAT

### Health Monitoring

- Always pair pools with health monitors in production
- Without a monitor, unhealthy members remain in rotation
- Use HTTP monitors for HTTP pools, TCP monitors for TCP pools

## References

- [Terraform openstack_lb_pool_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/lb_pool_v2)
- [Pulumi openstack.loadbalancer.Pool](https://www.pulumi.com/registry/packages/openstack/api-docs/loadbalancer/pool/)
- [OpenStack Octavia API -- Pools](https://docs.openstack.org/api-ref/load-balancer/v2/#pools)
- [Octavia Cookbook](https://docs.openstack.org/octavia/latest/user/guides/basic-cookbook.html)

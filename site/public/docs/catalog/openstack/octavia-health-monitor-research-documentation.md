---
title: "Octavia Health Monitor -- Research Documentation"
description: "Octavia Health Monitor -- Research Documentation deployment documentation"
icon: "package"
order: 100
componentName: "openstackloadbalancermonitor"
---

# OpenStack Octavia Health Monitor -- Research Documentation

## Introduction

An Octavia health monitor periodically probes pool members to determine their health status. Unhealthy members are automatically removed from the pool's traffic rotation until they recover. Health monitors are attached to pools, and each pool can have at most one monitor.

Monitors are the operational safety net in the Octavia hierarchy: without a monitor, failed members continue to receive traffic indefinitely. In production, every pool should have a health monitor.

## Historical Context

**Neutron LBaaS v1 (2013-2015):** Health monitors existed but were limited to HTTP, HTTPS, and TCP types. No support for PING, TLS-HELLO, or UDP-CONNECT.

**Neutron LBaaS v2 / Octavia (2015-present):** Health monitors gained additional check types (PING, TLS-HELLO, UDP-CONNECT), max_retries_down for asymmetric failover/recovery thresholds, and improved timeout handling. Tags are NOT supported on health monitors (provider limitation).

Key improvements:
- **Additional types**: PING, TLS-HELLO, UDP-CONNECT
- **max_retries_down**: Separate threshold for marking members unhealthy
- **Admin state**: Pause monitoring without deleting the monitor
- **Name field**: Human-readable names for monitors

## Architecture

### Monitor Position in the Octavia Hierarchy

```
OpenStackLoadBalancer (VIP on subnet)
  +-- OpenStackLoadBalancerListener (protocol + port)
        +-- OpenStackLoadBalancerPool (algorithm + protocol)
              +-- OpenStackLoadBalancerMember (backend server)
              +-- OpenStackLoadBalancerMonitor (health check)   <-- this component
```

### Health Check Types

| Type | Mechanism | Use Case |
|---|---|---|
| HTTP | Send HTTP request, check response code | Web applications |
| HTTPS | Send HTTPS request, check response code | Secure web applications |
| PING | ICMP echo request | Basic reachability check |
| TCP | Attempt TCP connection | TCP services |
| TLS-HELLO | Perform TLS handshake | TLS services without HTTP |
| UDP-CONNECT | Send UDP datagram, check for ICMP errors | UDP services |

### Health Check Flow

1. Monitor sends a probe to the member at the configured interval (delay)
2. If the member responds within the timeout, the check succeeds
3. After max_retries consecutive successes, the member is marked ONLINE
4. After max_retries_down consecutive failures, the member is marked ERROR
5. Members in ERROR state are removed from the pool's traffic rotation
6. When a member recovers, max_retries consecutive successes restore it to ONLINE

### Timing Configuration

- **delay**: Time between consecutive health checks (seconds)
- **timeout**: Maximum time to wait for a single check response (seconds)
- **Best practice**: timeout < delay to avoid overlapping checks
- **Example**: delay=10, timeout=5 means check every 10s, wait up to 5s for response

## Deployment Methods Landscape

### Level 0: Manual (Horizon Dashboard)

Health monitors are created through the load balancer wizard. Good for learning. Not suitable for production.

### Level 1: CLI (openstack client)

```bash
# Create an HTTP health monitor
openstack loadbalancer healthmonitor create \
  --name http-health \
  --type HTTP \
  --delay 5 \
  --timeout 10 \
  --max-retries 3 \
  --url-path /healthz \
  --http-method GET \
  --expected-codes 200 \
  web-pool

# Create a TCP health monitor
openstack loadbalancer healthmonitor create \
  --name tcp-check \
  --type TCP \
  --delay 5 \
  --timeout 3 \
  --max-retries 3 \
  web-pool
```

Good for ad-hoc operations. Not recommended for managing infrastructure at scale.

### Level 2: IaC -- Terraform

```hcl
resource "openstack_lb_monitor_v2" "http" {
  name            = "http-health"
  pool_id         = openstack_lb_pool_v2.web.id
  type            = "HTTP"
  delay           = 5
  timeout         = 10
  max_retries     = 3
  url_path        = "/healthz"
  http_method     = "GET"
  expected_codes  = "200"
}
```

Industry standard. OpenMCF wraps this with FK-based references and validation.

### Level 3: IaC -- Pulumi

```go
monitor, err := loadbalancer.NewMonitor(ctx, "http-health", &loadbalancer.MonitorArgs{
    Name:          pulumi.String("http-health"),
    PoolId:        pool.ID(),
    Type:          pulumi.String("HTTP"),
    Delay:         pulumi.Int(5),
    Timeout:       pulumi.Int(10),
    MaxRetries:    pulumi.Int(3),
    UrlPath:       pulumi.StringPtr("/healthz"),
    HttpMethod:    pulumi.StringPtr("GET"),
    ExpectedCodes: pulumi.StringPtr("200"),
})
```

Excellent for engineering teams.

## The OpenMCF Approach

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMonitor
metadata:
  name: http-health
spec:
  pool_id:
    value_from:
      name: web-pool
  type: "HTTP"
  delay: 5
  timeout: 10
  max_retries: 3
  url_path: "/healthz"
  http_method: "GET"
  expected_codes: "200"
```

### What OpenMCF Automates

1. **Foreign key resolution**: pool_id.value_from resolves to the pool UUID
2. **Dependency ordering**: Monitor waits for the pool to be ready
3. **Validation**: Type, max_retries range, HTTP field restrictions validated at the API level
4. **CEL constraints**: url_path/http_method/expected_codes rejected on non-HTTP types
5. **Dual IaC engines**: Same manifest works with Pulumi and Terraform

### The 80/20 Principle

The Terraform openstack_lb_monitor_v2 resource exposes 14 schema fields. OpenMCF exposes 11 fields:

**Included:** pool_id, type, delay, timeout, max_retries, max_retries_down, url_path, http_method, expected_codes, admin_state_up, region

**Excluded:** tenant_id (admin-only), value_specs (vendor escape hatch), tags (not supported by provider)

### API Design Decisions

**CEL validation for HTTP fields**: The message-level CEL constraint ensures url_path, http_method, and expected_codes are only set for HTTP/HTTPS monitors. This catches a common misconfiguration at the API level.

**`optional int32 max_retries_down`**: Lets Octavia default to the same value as max_retries when not set. Important for symmetric behavior when only max_retries is configured.

**No tags field**: The Terraform OpenStack provider does not support tags on health monitors. This is documented in the proto and reflected in all implementations.

## Implementation Landscape

### Resources Created

| Resource | Count | Description |
|----------|-------|-------------|
| openstack_lb_monitor_v2 (TF) / loadbalancer.Monitor (Pulumi) | 1 | The Octavia health monitor |

### Dependency Role

**References:** pool_id -> OpenStackLoadBalancerPool.status.outputs.pool_id

**Referenced by:** None (leaf node in the Octavia hierarchy alongside members)

## Production Best Practices

### Monitor Type Selection

- Use HTTP monitors for HTTP pools (most accurate health signal)
- Use TCP monitors for non-HTTP TCP services
- Use PING as a last resort (only checks network reachability, not application health)
- Use TLS-HELLO for TLS services without HTTP endpoints

### Timing Configuration

- Set timeout < delay to avoid overlapping checks
- Use shorter delays (3-5s) for latency-sensitive applications
- Use longer delays (30-60s) for stable services to reduce monitoring overhead
- Balance max_retries: too low = false positives, too high = slow failure detection

### HTTP Monitor Best Practices

- Use a dedicated health endpoint (e.g., /healthz) that checks application dependencies
- Return 200 for healthy, 503 for unhealthy
- Keep the health endpoint lightweight (no database queries on every check)
- Use expected_codes "200" for strict checking, "200-299" for lenient checking

### max_retries vs max_retries_down

- max_retries: How many successes to recover (return to rotation)
- max_retries_down: How many failures to fail (remove from rotation)
- Set max_retries_down lower for fast failure detection
- Set max_retries higher for cautious recovery

## References

- [Terraform openstack_lb_monitor_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/lb_monitor_v2)
- [Pulumi openstack.loadbalancer.Monitor](https://www.pulumi.com/registry/packages/openstack/api-docs/loadbalancer/monitor/)
- [OpenStack Octavia API -- Health Monitors](https://docs.openstack.org/api-ref/load-balancer/v2/#health-monitors)
- [Octavia Cookbook](https://docs.openstack.org/octavia/latest/user/guides/basic-cookbook.html)

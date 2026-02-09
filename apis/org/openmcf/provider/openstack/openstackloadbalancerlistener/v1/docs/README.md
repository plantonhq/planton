# OpenStack Octavia Listener -- Research Documentation

## Introduction

An Octavia listener is the second level in the Octavia object hierarchy: Load Balancer -> **Listener** -> Pool -> Member -> Health Monitor. The listener binds a protocol and port to a load balancer, accepting incoming traffic on that combination and forwarding it to a backend pool.

Each listener belongs to exactly one load balancer, and a single load balancer can have multiple listeners (e.g., one for HTTP on port 80, another for HTTPS on port 443). This enables a single VIP to serve multiple protocols and ports.

## Historical Context

**Neutron LBaaS v1 (2013-2015):** The original Neutron LBaaS had no separate listener concept. The "VIP" object combined what Octavia separates into load balancer + listener. This made multi-port configurations awkward -- each port required a separate VIP and load balancer.

**Neutron LBaaS v2 (2015-2018):** LBaaS v2 introduced the hierarchical model with separate listener resources. This was a major improvement, enabling multiple listeners per load balancer. The API surface is identical to modern Octavia listeners.

**Octavia (2017-present):** Octavia inherited the LBaaS v2 listener model and added:
- **TLS termination**: TERMINATED_HTTPS protocol with Barbican secret integration
- **Insert headers**: X-Forwarded-For, X-Forwarded-Proto, X-Forwarded-Port
- **Allowed CIDRs**: Per-listener access control lists
- **Tags**: Resource tagging for organization and filtering
- **Connection limits**: Per-listener connection throttling
- **L7 policies**: Complex routing rules (not yet in OpenMCF scope)

The `openstack_lb_listener_v2` Terraform resource and `loadbalancer.Listener` Pulumi resource both target the Octavia API.

## Architecture

### Protocol Types

| Protocol | Layer | TLS | Use Case |
|----------|-------|-----|----------|
| HTTP | 7 | No | Unencrypted web traffic. Supports insert_headers and L7 policies |
| HTTPS | 4 | Pass-through | Encrypted traffic passed directly to backends. No TLS termination |
| TCP | 4 | No | Raw TCP (databases, custom protocols) |
| UDP | 4 | No | DNS, gaming, streaming |
| TERMINATED_HTTPS | 7 | Terminated | TLS terminated at LB. Requires Barbican cert. Supports insert_headers |

### TLS Termination

When using `TERMINATED_HTTPS`:
1. The client connects to the listener with TLS
2. The load balancer terminates TLS using the certificate from Barbican
3. The decrypted request is forwarded to the backend pool (typically over HTTP)
4. The `insert_headers` feature adds `X-Forwarded-Proto: https` so backends know the original protocol

The `default_tls_container_ref` must point to a Barbican secret container containing:
- The server certificate
- The private key
- Optional intermediate certificates

### Insert Headers

Only available for HTTP and TERMINATED_HTTPS protocols:

| Header | Value | Purpose |
|--------|-------|---------|
| `X-Forwarded-For` | `"true"` | Client's original IP address |
| `X-Forwarded-Proto` | `"true"` | Original protocol (http/https) |
| `X-Forwarded-Port` | `"true"` | Original port number |

These headers are critical for backends to know the client's real IP and protocol when behind a load balancer.

### Allowed CIDRs

Allowed CIDRs provide per-listener access control:
- When set, only traffic from listed CIDRs reaches the listener
- All other traffic is silently dropped
- When empty (default), all traffic is allowed
- CIDRs are specified in standard notation (e.g., `10.0.0.0/8`)

This is useful for:
- Restricting admin panels to internal networks
- Limiting API access to known clients
- Creating internal-only listeners on public-facing load balancers

## Deployment Methods Landscape

### Level 0: Manual (Horizon Dashboard)

Listeners are created via the Horizon Load Balancer wizard or the listener tab:

1. Navigate to **Project > Network > Load Balancers**
2. Select an existing load balancer
3. Click the **Listeners** tab
4. Click **Create Listener**
5. Select protocol, port, and optional parameters
6. Click **Create**

**Pros:**
- Visual interface for protocol selection
- Immediate feedback on conflicts (e.g., duplicate port/protocol)

**Cons:**
- Not reproducible or auditable
- No way to set insert_headers or allowed_cidrs via Horizon in many deployments
- Cannot enforce naming conventions

**Verdict:** Good for learning. Not suitable for production.

### Level 1: CLI (openstack client)

```bash
# Create a basic HTTP listener
openstack loadbalancer listener create \
  --name http-listener \
  --protocol HTTP \
  --protocol-port 80 \
  dev-lb

# Create TERMINATED_HTTPS listener with TLS
openstack loadbalancer listener create \
  --name https-listener \
  --protocol TERMINATED_HTTPS \
  --protocol-port 443 \
  --default-tls-container-ref https://barbican.example.com/v1/secrets/cert-123 \
  --insert-headers X-Forwarded-For=true \
  --insert-headers X-Forwarded-Proto=true \
  prod-lb

# Create listener with allowed CIDRs
openstack loadbalancer listener create \
  --name api-listener \
  --protocol HTTP \
  --protocol-port 8080 \
  --allowed-cidr 10.0.0.0/8 \
  --allowed-cidr 172.16.0.0/12 \
  api-lb

# Set connection limit
openstack loadbalancer listener create \
  --name limited-listener \
  --protocol HTTP \
  --protocol-port 80 \
  --connection-limit 10000 \
  prod-lb
```

**Pros:**
- Full control over all parameters
- Scriptable for automation

**Cons:**
- No state tracking, no drift detection
- Must manually ensure the load balancer is in ACTIVE state first
- Manual dependency management

**Verdict:** Good for ad-hoc operations. Not recommended for managing infrastructure at scale.

### Level 2: IaC -- Terraform

```hcl
resource "openstack_lb_listener_v2" "http" {
  name            = "http-listener"
  loadbalancer_id = openstack_lb_loadbalancer_v2.app_lb.id
  protocol        = "HTTP"
  protocol_port   = 80

  insert_headers = {
    "X-Forwarded-For"   = "true"
    "X-Forwarded-Proto" = "true"
  }

  tags = ["env:production", "managed-by:terraform"]
}

resource "openstack_lb_listener_v2" "https" {
  name            = "https-listener"
  loadbalancer_id = openstack_lb_loadbalancer_v2.app_lb.id
  protocol        = "TERMINATED_HTTPS"
  protocol_port   = 443

  default_tls_container_ref = barbican_secret_v1.cert.secret_ref

  insert_headers = {
    "X-Forwarded-For"   = "true"
    "X-Forwarded-Proto" = "true"
  }
}
```

**Pros:**
- Declarative, reproducible
- Automatic dependency ordering (waits for LB to be ACTIVE)
- State tracking and drift detection

**Cons:**
- HCL syntax required
- State management overhead
- No built-in FK validation across resources

**Verdict:** Industry standard. OpenMCF wraps this with a unified API and FK-based dependency management.

### Level 3: IaC -- Pulumi

```go
listener, err := loadbalancer.NewListener(ctx, "http-listener", &loadbalancer.ListenerArgs{
    Name:           pulumi.String("http-listener"),
    LoadbalancerId: lb.ID(),
    Protocol:       pulumi.String("HTTP"),
    ProtocolPort:   pulumi.Int(80),
    InsertHeaders: pulumi.StringMap{
        "X-Forwarded-For":   pulumi.String("true"),
        "X-Forwarded-Proto": pulumi.String("true"),
    },
})
```

**Pros:**
- Type-safe, compile-time error detection
- Native dependency tracking via resource references
- Real programming language capabilities

**Cons:**
- More verbose for simple resources

**Verdict:** Excellent for engineering teams. OpenMCF uses Pulumi internally for Go-native IaC modules.

## The OpenMCF Approach

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerListener
metadata:
  name: http-listener
spec:
  loadbalancer_id:
    value_from:
      name: app-lb
  protocol: "HTTP"
  protocol_port: 80
  insert_headers:
    X-Forwarded-For: "true"
    X-Forwarded-Proto: "true"
```

### What OpenMCF Automates

1. **Foreign key resolution**: `loadbalancer_id.value_from` automatically resolves to the LB's UUID from its stack outputs
2. **Dependency ordering**: In InfraCharts, the listener waits for the load balancer to complete before deployment
3. **Provider configuration**: Credentials resolved from the platform's credential store
4. **Dual IaC engines**: Same manifest works with both Pulumi and Terraform backends
5. **Validation**: Protocol enum, port range, TLS ref requirement, and tag uniqueness validated at the API level before deployment

### The 80/20 Principle

The Terraform `openstack_lb_listener_v2` resource exposes 16+ schema fields. OpenMCF's `OpenStackLoadBalancerListenerSpec` exposes 11 fields -- the ones that cover 95%+ of real-world use cases:

**Included:**
- `loadbalancer_id` -- Load balancer (FK)
- `protocol` -- Protocol type (validated enum)
- `protocol_port` -- Port number (validated range)
- `description` -- Human-readable description
- `connection_limit` -- Max connections
- `default_tls_container_ref` -- Barbican TLS secret
- `insert_headers` -- HTTP headers to inject
- `allowed_cidrs` -- Access control
- `admin_state_up` -- Administrative state (default: true)
- `tags` -- Resource tagging
- `region` -- Region override

**Excluded:**
- `tenant_id` -- Admin-only
- `timeout_client_data` -- Advanced timeout tuning
- `timeout_member_data` -- Advanced timeout tuning
- `timeout_member_connect` -- Advanced timeout tuning
- `timeout_tcp_inspect` -- Advanced timeout tuning
- `sni_container_refs` -- SNI multi-cert (niche)

### API Design Decisions

**`loadbalancer_id` as a StringValueOrRef FK:** The listener's only inbound FK. Using `value_from` enables dependency chains in InfraCharts where the listener automatically waits for its load balancer.

**Protocol validation via CEL enum:** The `string.in` constraint ensures only valid Octavia protocols are accepted. This catches typos (e.g., "Http" instead of "HTTP") at validation time rather than at deployment.

**`protocol_port` range validation:** CEL expression validates 1-65535 at the API level, preventing deployment failures from invalid port numbers.

**Cross-field validation for TERMINATED_HTTPS:** The CEL message-level validation ensures `default_tls_container_ref` is provided when `protocol` is `TERMINATED_HTTPS`. This prevents a common misconfiguration that would cause an Octavia API error.

**`insert_headers` as a map:** Maps naturally to both Pulumi's `StringMap` and Terraform's `map(string)`. The map type allows flexible header configuration without needing a predefined set of boolean flags.

**Single-resource component:** Each listener is a separate component, enabling:
- Independent lifecycle (add/remove listeners without recreating the LB)
- Flexible composition in InfraCharts (one chart per listener)
- Clear FK chains: LB -> Listener -> Pool -> Member

## Implementation Landscape

### Resources Created

| Resource | Count | Description |
|----------|-------|-------------|
| `openstack_lb_listener_v2` (TF) / `loadbalancer.Listener` (Pulumi) | 1 | The Octavia listener |

Single-resource component. Atomic and composable in InfraCharts.

### Dependency Role

OpenStackLoadBalancerListener has one inbound FK and is referenced by the pool:

**References:**
- `loadbalancer_id` -> `OpenStackLoadBalancer.status.outputs.loadbalancer_id`

**Referenced by:**
- `OpenStackLoadBalancerPool.spec.listener_id` -> `listener_id`

This makes `listener_id` the primary output, used as a foreign key by pools in the Octavia hierarchy.

## Production Best Practices

### Protocol Selection

- Use **HTTP** for internal services where TLS is not needed
- Use **TERMINATED_HTTPS** for external-facing services -- terminates TLS at the LB and forwards plaintext to backends
- Use **HTTPS** pass-through only when backends must handle their own TLS (e.g., mutual TLS)
- Use **TCP** for databases, message queues, and custom protocols
- Use **UDP** for DNS, gaming, and streaming

### TLS Management

- Store certificates in Barbican with proper access controls
- Rotate certificates by updating the `default_tls_container_ref` (triggers listener update)
- For multiple domains on one listener, consider SNI (not yet in OpenMCF scope)
- Always enable `X-Forwarded-Proto` header to let backends know traffic was originally HTTPS

### Access Control

- Use `allowed_cidrs` as a first line of defense for sensitive listeners
- Combine with security groups on the VIP port for defense in depth
- Document allowed CIDRs in your InfraChart for visibility

### Connection Limits

- Set connection limits to prevent a single listener from overwhelming backend resources
- Use -1 (unlimited) for internal services with known traffic patterns
- Monitor connection counts to tune limits based on actual traffic

### High Availability

- Listeners inherit the HA characteristics of their parent load balancer
- During failover, in-flight connections may be dropped -- design clients for retry
- For zero-downtime certificate rotation, use TERMINATED_HTTPS with the new cert

## References

- [Terraform openstack_lb_listener_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/lb_listener_v2)
- [Pulumi openstack.loadbalancer.Listener](https://www.pulumi.com/registry/packages/openstack/api-docs/loadbalancer/listener/)
- [OpenStack Octavia API -- Listeners](https://docs.openstack.org/api-ref/load-balancer/v2/#listeners)
- [OpenStack Octavia User Guide](https://docs.openstack.org/octavia/latest/user/guides/)
- [Octavia Cookbook](https://docs.openstack.org/octavia/latest/user/guides/basic-cookbook.html)

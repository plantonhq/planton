# HetznerCloudLoadBalancer

The **HetznerCloudLoadBalancer** resource provisions a complete load balancing stack in Hetzner Cloud — a load balancer with services (listeners), backend targets, and an optional private network attachment. This is the most complex component in the Planton Hetzner Cloud catalog: it bundles four provider resources into a single manifest that defines how incoming traffic reaches your servers.

## What It Represents

A [Hetzner Cloud Load Balancer](https://docs.hetzner.cloud/#load-balancers) distributes incoming traffic across one or more backend targets using configurable protocols, ports, and health checks. Each load balancer exposes one or more **services** (listeners) that accept traffic on a protocol/port combination and forward it to targets. Targets can be specific servers, dynamically selected via label selectors, or external IP addresses.

The load balancer supports Layer 7 (HTTP/HTTPS) and Layer 4 (TCP) protocols. HTTP/HTTPS services enable features like sticky sessions, TLS termination with Hetzner-managed or uploaded certificates, and automatic HTTP-to-HTTPS redirection. Health checks monitor target availability and remove unhealthy backends from rotation.

Optionally, the load balancer can be attached to a private Hetzner Cloud network so that traffic to targets flows over the private network instead of the public internet.

## Bundled Resources

| Terraform Resource | Count | Created When | Purpose |
|---|---|---|---|
| `hcloud_load_balancer` | 1 | Always | Provisions the load balancer with the specified type, location, algorithm, labels, and delete protection |
| `hcloud_load_balancer_service` | 1 per service | Always (at least 1 service required) | Configures a listener with protocol, ports, HTTP settings, and health check |
| `hcloud_load_balancer_target` | 1 per target | When any of `serverTargets`, `labelSelectorTargets`, or `ipTargets` has entries | Adds a backend target (server, label selector, or IP) to the load balancer |
| `hcloud_load_balancer_network` | 0 or 1 | When `network` is set | Attaches the load balancer to a private network with an optional fixed IP |

All four resource types are bundled because none has an independent lifecycle. A service only exists in the context of a load balancer. A target only makes sense when there are services to route traffic. The network attachment configures how the load balancer communicates with its targets. Removing any of these from the bundle would leave orphaned resources with no purpose.

## Key Features

### Load Balancer Types

The `loadBalancerType` field selects the hardware profile, which determines connection capacity and the maximum number of targets:

| Type | Max Targets | Connections/s | Bandwidth |
|------|-------------|---------------|-----------|
| `lb11` | 25 | 10,000 | 10 Gbps |
| `lb21` | 75 | 20,000 | 20 Gbps |
| `lb31` | 150 | 40,000 | 40 Gbps |

The type can be changed after creation (in-place resize). No downtime is involved.

### Algorithm Selection

The `algorithm` field controls how the load balancer distributes connections across healthy targets:

- **`round_robin`** (default) — Connections are distributed evenly across targets in order. Simple and predictable.
- **`least_connections`** — Each new connection goes to the target with the fewest active connections. Better when backends have varying response times or processing capacities.

### Services (Listeners)

Each entry in the `services` list creates a listener on the load balancer. A service binds to a `listenPort` and forwards traffic to targets on a `destinationPort` using the specified `protocol`. At least one service is required.

Three protocols are supported:

| Protocol | Layer | Default Listen Port | Features |
|----------|-------|---------------------|----------|
| `http` | 7 | 80 | Sticky sessions, HTTP health checks |
| `https` | 7 | 443 | TLS termination, certificates, HTTP redirect, sticky sessions |
| `tcp` | 4 | (required) | Pass-through, PROXY protocol |

For `http` and `https` services, `destinationPort` defaults to the `listenPort`. For `tcp` services, both `listenPort` and `destinationPort` are required.

Each service's `listenPort` must be unique across all services on the same load balancer. Changing the `protocol` or `listenPort` forces replacement of the service resource.

### HTTP Features

HTTP and HTTPS services support additional configuration via the `http` block:

- **Sticky sessions** — Cookie-based session affinity. When enabled, the load balancer sets a cookie on the first response and routes subsequent requests with that cookie to the same target. Configurable cookie name (default: `HCLBSTICKY`) and lifetime (default: 300 seconds).
- **TLS termination** — HTTPS services terminate TLS at the load balancer using certificates referenced via `certificateIds`. Each entry accepts a literal certificate ID or a `valueFrom` reference to a `HetznerCloudCertificate` resource.
- **HTTP-to-HTTPS redirect** — When `redirectHttp` is `true` on an HTTPS service using port 443, the load balancer automatically redirects HTTP traffic on port 80 to HTTPS.

### Health Checks

Each service can have a custom health check. When omitted, the provider creates a default health check matching the service protocol and destination port.

Health check fields:

| Field | Default | Description |
|-------|---------|-------------|
| `protocol` | Matches service (HTTPS defaults to `http`) | `http`, `https`, or `tcp` |
| `port` | Matches `destinationPort` | Port to check on the target |
| `interval` | 15 | Seconds between checks |
| `timeout` | 10 | Seconds to wait for a response (must be < interval) |
| `retries` | 3 | Consecutive failures before marking unhealthy |

HTTP/HTTPS health checks support additional fields: `path` (default: `/`), `domain` (HTTP Host header), `response` (expected body substring), `tls` (verify target TLS cert), and `statusCodes` (default: `["2??", "3??"]` — wildcard notation).

### Target Types

Targets are the backends that receive traffic. Three target types are supported, each in its own list:

- **Server targets** (`serverTargets`) — Add a specific Hetzner Cloud server by ID. Each entry's `serverId` accepts a literal ID or a `valueFrom` reference to a `HetznerCloudServer` resource. Set `usePrivateIp: true` to route traffic over the private network.

- **Label selector targets** (`labelSelectorTargets`) — Dynamically add all servers matching a Hetzner Cloud label selector expression (e.g., `env=production,role=web`). The target set updates automatically as server labels change. Set `usePrivateIp: true` for private routing.

- **IP targets** (`ipTargets`) — Add an external IP address as a backend. Use this for targets outside of Hetzner Cloud. IP targets cannot use private IP routing.

All target types can coexist on the same load balancer.

### Network Attachment

The `network` field attaches the load balancer to a private Hetzner Cloud network. A load balancer can be attached to at most one network.

When attached:
- The load balancer receives a private IP within the network's subnet range (auto-assigned or fixed via `ip`)
- Server and label selector targets with `usePrivateIp: true` receive traffic over the private network
- The public interface can be disabled via `enablePublicInterface: false` to make the load balancer private-only

The `networkId` field accepts a literal network ID or a `valueFrom` reference to a `HetznerCloudNetwork` resource.

### Delete Protection

When `deleteProtection` is `true`, the load balancer cannot be deleted via the Hetzner Cloud API until protection is explicitly removed.

### Automatic Labeling

Standard labels (`resource`, `name`, `kind`, `org`, `env`, `id`) are applied to the load balancer from metadata. User-specified `metadata.labels` are merged in, with standard labels taking precedence on key conflicts. Labels are applied only to the load balancer resource — service, target, and network resources do not support labels in the Hetzner Cloud API.

## Upstream Dependencies (What This Resource Needs)

| Dependency | Field | Required | Cardinality | Purpose |
|---|---|---|---|---|
| `HetznerCloudServer` | `spec.serverTargets[].serverId` | No | 0..N | Servers added as load balancer backends |
| `HetznerCloudNetwork` | `spec.network.networkId` | No | 0..1 | Private network for internal traffic routing |
| `HetznerCloudCertificate` | `spec.services[].http.certificateIds[]` | No | 0..N | TLS certificates for HTTPS termination |

All dependencies are optional. A load balancer can be created with literal IDs instead of `valueFrom` references.

## Downstream Dependents (What References This Resource)

No components currently reference the load balancer's outputs. The load balancer is a leaf node in the Hetzner Cloud dependency graph — it consumes references from servers, networks, and certificates but nothing depends on it.

## Stack Outputs

| Output | Description |
|---|---|
| `load_balancer_id` | Hetzner Cloud numeric ID of the created load balancer (as string). Can be referenced by other components via `StringValueOrRef`. |
| `ipv4_address` | The public IPv4 address assigned to the load balancer. Empty if the public interface is disabled via `network.enablePublicInterface = false`. |
| `ipv6_address` | The public IPv6 address assigned to the load balancer. Empty if the public interface is disabled. |

## References

- [Hetzner Cloud Load Balancers Documentation](https://docs.hetzner.cloud/#load-balancers)
- [Hetzner Cloud Load Balancer Types](https://docs.hetzner.cloud/#load-balancer-types)
- [Terraform hcloud_load_balancer Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/load_balancer)
- [Terraform hcloud_load_balancer_service Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/load_balancer_service)
- [Terraform hcloud_load_balancer_target Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/load_balancer_target)
- [Terraform hcloud_load_balancer_network Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/load_balancer_network)
- [Pulumi hcloud.LoadBalancer Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/loadbalancer/)
- [Pulumi hcloud.LoadBalancerService Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/loadbalancerservice/)
- [Pulumi hcloud.LoadBalancerTarget Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/loadbalancertarget/)
- [Pulumi hcloud.LoadBalancerNetwork Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/loadbalancernetwork/)

# Hetzner Cloud Load Balancer — Research Documentation

## Introduction

A Hetzner Cloud Load Balancer distributes incoming network traffic across a pool of backend servers, transforming a set of independent machines into a scalable, fault-tolerant service. It is the resource that turns "I have servers" into "I have a service" — without it, clients must know individual server addresses, handle failover manually, and accept that any single server failure means downtime.

The `HetznerCloudLoadBalancer` component is the most complex resource in the OpenMCF Hetzner Cloud catalog. It bundles four provider resources — the load balancer itself, its services (listeners), its targets (backends), and an optional network attachment — into a single manifest. This bundling reflects the reality that these resources have no independent lifecycle: a service only exists in the context of a load balancer, and a target only matters when there are services to route traffic.

What makes this component architecturally significant is its position as the **terminal consumer** in the dependency graph. It references servers (as targets), networks (for private routing), and certificates (for TLS termination), but nothing references it. Every other component in the catalog either produces something the load balancer consumes or is entirely unrelated to it. This means the load balancer manifest is where the infrastructure story comes together — it is the last piece deployed in a typical web application stack.

OpenMCF's abstraction consolidates what would otherwise be a multi-resource Terraform/Pulumi deployment into a single declarative manifest with proto-validated fields, foreign key references to other OpenMCF components, and sensible defaults that cover the common case while exposing the full breadth of configuration when needed.

## Historical Context

### The Evolution of Load Balancing

Load balancing predates cloud computing by decades. Hardware load balancers from F5, Citrix, and A10 Networks dominated enterprise networking in the 2000s — expensive, proprietary appliances that sat in front of server racks and distributed TCP connections. Configuration was done through vendor-specific CLIs or web interfaces, and provisioning a new load balancer meant a purchase order, a shipping delay, and a rack-and-stack operation.

Cloud computing replaced the hardware appliance with an API call. AWS launched Elastic Load Balancing in 2009, and every major cloud provider followed. The core concept remained the same — distribute traffic, check health, remove failed backends — but the provisioning model changed from "buy hardware" to "call an API." This shift moved load balancing from a capital expense to an operational one and made it accessible to small teams.

Modern cloud load balancers have converged on a common architecture:
- **Listeners** (services) bind to protocol/port combinations and accept incoming traffic
- **Target groups** (backends) define the pool of servers that receive forwarded traffic
- **Health checks** verify that targets are healthy before routing traffic to them
- **Rules** (HTTP path/host-based routing, in more complex providers) direct traffic to specific target groups

The complexity varies dramatically across providers. AWS ALB has dozens of configuration options, path-based routing rules, authentication integrations, and WAF attachments. GCP's load balancing suite spans seven distinct products. Azure has five load balancer types with overlapping feature sets.

### Hetzner Cloud's Position

Hetzner Cloud launched its load balancer product as a straightforward Layer 4/7 load balancer with a deliberately constrained feature set. Where AWS ALB has path-based routing, WAF integration, and authentication actions, Hetzner Cloud's load balancer focuses on the fundamentals:

- **Three protocols**: HTTP, HTTPS, TCP
- **Two algorithms**: round-robin and least-connections
- **Three target types**: server, label selector, IP
- **One network attachment**: at most one private network per load balancer
- **Three sizes**: lb11, lb21, lb31 with increasing capacity

This simplicity is intentional and aligns with Hetzner Cloud's philosophy of providing a focused, price-competitive product without enterprise complexity. There are no path-based routing rules, no request/response transformations, no authentication integrations, and no WAF. If you need those features, you deploy a reverse proxy (Nginx, Caddy, Traefik, HAProxy) on a server behind the load balancer.

What Hetzner Cloud's load balancer does well is the 80% use case: distribute HTTP/HTTPS/TCP traffic across servers with health checks, sticky sessions, and TLS termination. For most web applications, this is sufficient.

### The Multi-Resource Problem

The simplicity of Hetzner Cloud's load balancer concept contrasts with the operational complexity of provisioning one. In the raw Hetzner Cloud API (and by extension, in raw Terraform or Pulumi), a load balancer with HTTPS, health checks, and private networking requires four separate resources:

1. `hcloud_load_balancer` — the load balancer itself
2. `hcloud_load_balancer_service` — one per listener (e.g., HTTPS on 443, TCP on 5432)
3. `hcloud_load_balancer_target` — one per backend (e.g., server-1, server-2, label-selector)
4. `hcloud_load_balancer_network` — the network attachment

Each resource has its own ID type requirements (string vs. integer), its own lifecycle constraints (ForceNew fields), and its own implicit ordering dependencies (the network attachment must exist before targets can use private IPs). Managing these four resource types across a Terraform state file or Pulumi stack requires careful attention to dependency ordering and ID type conversions.

This is the problem OpenMCF solves: collapse four resources into one manifest with one `openmcf apply` command.

## Deployment Methods Landscape

### Level 0: Manual (Hetzner Cloud Console)

1. Log in to [console.hetzner.cloud](https://console.hetzner.cloud)
2. Navigate to **Load Balancers** in the left sidebar
3. Click **Create Load Balancer**
4. Select a location (Falkenstein, Nuremberg, Helsinki, Ashburn, Hillsboro, Singapore)
5. Select a type (LB11, LB21, LB31)
6. Choose an algorithm (Round Robin or Least Connections)
7. Add a network: select an existing private network and optionally assign a fixed IP
8. Add targets: pick servers from a list, enter label selectors, or type external IPs
9. Add services:
   - For each service, select protocol (HTTP/HTTPS/TCP), listen port, destination port
   - For HTTPS: select certificates, enable HTTP redirect
   - Configure health check: protocol, port, interval, timeout, retries, HTTP path
   - Configure sticky sessions: cookie name, lifetime
10. Enable/disable delete protection
11. Name the load balancer and add labels
12. Click **Create & Buy now**

**Pros:**
- Visual configuration with immediate feedback
- Server selection from a dropdown list
- Certificate selection from the project's certificate list
- Good for initial exploration and prototyping

**Cons:**
- Not repeatable — no record of configuration choices
- No version control
- Error-prone for complex configurations (multiple services, many targets)
- Cannot reference other resources by name, only by ID
- Changes require navigating multiple tabs (Services, Targets, Networking)

**Verdict:** Useful for learning and one-off experiments. Not suitable for production infrastructure.

### Level 1: CLI (`hcloud` CLI)

```bash
# Create the load balancer
hcloud load-balancer create \
  --name web-lb \
  --type lb11 \
  --location fsn1 \
  --algorithm-type round_robin \
  --label env=production

# Attach to a network
hcloud load-balancer attach-to-network \
  --network my-network \
  --ip 10.0.1.100 \
  web-lb

# Add an HTTPS service
hcloud load-balancer add-service \
  --protocol https \
  --listen-port 443 \
  --destination-port 8080 \
  --http-redirect-http \
  --http-certificates 12345 \
  web-lb

# Add a health check (configured per-service, set during add-service or update-service)
hcloud load-balancer update-service \
  --listen-port 443 \
  --health-check-protocol http \
  --health-check-port 8080 \
  --health-check-interval 10s \
  --health-check-timeout 5s \
  --health-check-retries 3 \
  --health-check-http-path /health \
  web-lb

# Add server targets
hcloud load-balancer add-target \
  --server web-01 \
  --use-private-ip \
  web-lb

hcloud load-balancer add-target \
  --server web-02 \
  --use-private-ip \
  web-lb

# Add a label selector target
hcloud load-balancer add-target \
  --label-selector "env=production,role=web" \
  --use-private-ip \
  web-lb
```

**Pros:**
- Scriptable and reproducible (put commands in a shell script)
- Direct access to all configuration options
- Fast feedback loop for prototyping

**Cons:**
- Imperative — commands must be run in the correct order
- No state tracking — adding a target twice may error or create duplicates
- No idempotency — running the script again fails on "already exists" errors
- Service updates require knowing the current listen port
- No dependency management between load balancer and referenced resources

**Verdict:** Better than the console for scripted workflows. Still lacks the state management and idempotency needed for production infrastructure.

### Level 2: Terraform

```hcl
resource "hcloud_load_balancer" "web" {
  name               = "web-lb"
  load_balancer_type = "lb11"
  location           = "fsn1"
  delete_protection  = true

  algorithm {
    type = "round_robin"
  }

  labels = {
    env  = "production"
    role = "web"
  }
}

resource "hcloud_load_balancer_network" "web" {
  load_balancer_id        = hcloud_load_balancer.web.id
  network_id              = hcloud_network.main.id
  ip                      = "10.0.1.100"
  enable_public_interface = true
}

resource "hcloud_load_balancer_service" "https" {
  load_balancer_id = hcloud_load_balancer.web.id
  protocol         = "https"
  listen_port      = 443
  destination_port = 8080

  http {
    certificates  = [hcloud_managed_certificate.web.id]
    redirect_http = true
  }

  health_check {
    protocol = "http"
    port     = 8080
    interval = 10
    timeout  = 5
    retries  = 3

    http {
      path = "/health"
    }
  }
}

resource "hcloud_load_balancer_target" "web_01" {
  load_balancer_id = hcloud_load_balancer.web.id
  type             = "server"
  server_id        = hcloud_server.web_01.id
  use_private_ip   = true

  depends_on = [hcloud_load_balancer_network.web]
}

resource "hcloud_load_balancer_target" "web_02" {
  load_balancer_id = hcloud_load_balancer.web.id
  type             = "server"
  server_id        = hcloud_server.web_02.id
  use_private_ip   = true

  depends_on = [hcloud_load_balancer_network.web]
}
```

**Pros:**
- Declarative — describes desired state, not steps
- State management — tracks all four resource types
- Idempotent — safe to run repeatedly
- Dependency graph — Terraform handles ordering (network before targets)
- Type-safe references — `hcloud_server.web_01.id` instead of hardcoded IDs

**Cons:**
- Four resource types to manage per load balancer
- Adding a server target requires a new `hcloud_load_balancer_target` resource block
- `depends_on` must be manually added for private IP targets
- ID type confusion — `load_balancer_id` is a string for services but an int for targets/network
- No schema validation until `terraform plan`
- HCL syntax requires understanding dynamic blocks for variable-length services

**Verdict:** The standard approach for teams managing infrastructure as code. Correct but verbose — a single HTTPS load balancer with two servers requires 5 resource blocks and careful dependency wiring.

### Level 3: Pulumi (Go)

```go
lb, err := hcloud.NewLoadBalancer(ctx, "web-lb", &hcloud.LoadBalancerArgs{
    Name:             pulumi.String("web-lb"),
    LoadBalancerType: pulumi.String("lb11"),
    Location:         pulumi.StringPtr("fsn1"),
    DeleteProtection: pulumi.Bool(true),
    Labels:           pulumi.ToStringMap(map[string]string{"env": "production"}),
    Algorithm: &hcloud.LoadBalancerAlgorithmArgs{
        Type: pulumi.StringPtr("round_robin"),
    },
})

lbNetwork, err := hcloud.NewLoadBalancerNetwork(ctx, "web-network", &hcloud.LoadBalancerNetworkArgs{
    LoadBalancerId:        lb.ID().ApplyT(func(id pulumi.ID) (int, error) { return strconv.Atoi(string(id)) }).(pulumi.IntOutput),
    NetworkId:             pulumi.IntPtr(networkId),
    Ip:                    pulumi.StringPtr("10.0.1.100"),
    EnablePublicInterface: pulumi.BoolPtr(true),
})

_, err = hcloud.NewLoadBalancerService(ctx, "https", &hcloud.LoadBalancerServiceArgs{
    LoadBalancerId:  lb.ID().ToStringOutput(),
    Protocol:        pulumi.String("https"),
    ListenPort:      pulumi.IntPtr(443),
    DestinationPort: pulumi.IntPtr(8080),
    Http: &hcloud.LoadBalancerServiceHttpArgs{
        Certificates: pulumi.IntArray{pulumi.Int(certId)},
        RedirectHttp: pulumi.BoolPtr(true),
    },
    HealthCheck: &hcloud.LoadBalancerServiceHealthCheckArgs{
        Protocol: pulumi.String("http"),
        Port:     pulumi.Int(8080),
        Interval: pulumi.Int(10),
        Timeout:  pulumi.Int(5),
        Retries:  pulumi.Int(3),
        Http: &hcloud.LoadBalancerServiceHealthCheckHttpArgs{
            Path: pulumi.StringPtr("/health"),
        },
    },
})

_, err = hcloud.NewLoadBalancerTarget(ctx, "web-01", &hcloud.LoadBalancerTargetArgs{
    LoadBalancerId: lb.ID().ApplyT(func(id pulumi.ID) (int, error) { return strconv.Atoi(string(id)) }).(pulumi.IntOutput),
    Type:           pulumi.String("server"),
    ServerId:       pulumi.IntPtr(server01Id),
    UsePrivateIp:   pulumi.BoolPtr(true),
}, pulumi.DependsOn([]pulumi.Resource{lbNetwork}))
```

**Pros:**
- Full programming language — loops, conditionals, type safety
- Compile-time checks catch some errors before deployment
- Same dependency management as Terraform but expressed in Go
- IDE support with autocompletion and type checking

**Cons:**
- Same four-resource pattern as Terraform — no less verbose
- ID type conversion is more painful: `ApplyT` with `strconv.Atoi` for every integer ID field
- The load balancer ID is `pulumi.StringOutput` for services but needs `pulumi.IntOutput` for targets/network
- Dependency ordering (`pulumi.DependsOn`) must still be manually specified
- Significant boilerplate for each target

**Verdict:** Better than Terraform for teams already using Go. The type conversion overhead for Hetzner Cloud's mixed-type IDs adds friction.

### Other Methods

**Ansible** — The `hetzner.hcloud` collection provides modules for load balancers, but Ansible's imperative task model makes multi-resource orchestration awkward. Each resource type is a separate task with manual ordering.

**Direct API** — The Hetzner Cloud REST API exposes load balancer operations at `https://api.hetzner.cloud/v1/load_balancers`. Useful for building custom tooling but requires implementing state management, idempotency, and error handling from scratch.

## Comparative Analysis

| Aspect | Console | CLI | Terraform | Pulumi | OpenMCF |
|--------|---------|-----|-----------|--------|---------|
| Declarative | No | No | Yes | Yes | Yes |
| State management | No | No | Yes | Yes | Yes |
| Idempotent | No | No | Yes | Yes | Yes |
| Single manifest | No | No | No (4 blocks) | No (4 calls) | Yes |
| Schema validation | At submit | At run | At plan | At compile | At parse (proto) |
| ID type handling | Hidden | Hidden | Manual | Manual + ApplyT | Automatic |
| Dependency ordering | Manual | Manual | Mostly auto | Mostly auto | Automatic |
| Foreign key refs | Hardcoded IDs | Hardcoded IDs | Resource refs | Resource refs | `valueFrom` |
| Repeatable | No | Partial | Yes | Yes | Yes |

## The OpenMCF Approach

### Single-Manifest Unification

OpenMCF collapses the four-resource deployment into a single YAML manifest:

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudLoadBalancer
metadata:
  name: web-lb
spec:
  loadBalancerType: lb11
  location: fsn1
  services:
    - protocol: https
      destinationPort: 8080
      http:
        certificateIds:
          - valueFrom:
              kind: HetznerCloudCertificate
              name: web-cert
              fieldPath: status.outputs.certificate_id
        redirectHttp: true
      healthCheck:
        protocol: http
        port: 8080
        http:
          path: /health
  serverTargets:
    - serverId:
        valueFrom:
          kind: HetznerCloudServer
          name: web-01
          fieldPath: status.outputs.server_id
      usePrivateIp: true
  network:
    networkId:
      valueFrom:
        kind: HetznerCloudNetwork
        name: main-vpc
        fieldPath: status.outputs.network_id
```

One file. One `openmcf apply`. The IaC modules handle resource creation ordering, ID type conversions, and dependency wiring.

### 80/20 Scoping Decisions

The spec exposes the configuration options that cover the vast majority of load balancer use cases while excluding rarely-used provider features:

**Included (the 80%):**
- Load balancer type, location, algorithm — core identity
- Services with all three protocols (HTTP, HTTPS, TCP) — the complete listener model
- HTTP features: sticky sessions, TLS certificates, HTTP redirect — essential for web applications
- Health checks with full HTTP customization — required for production reliability
- Three target types (server, label selector, IP) — all backend patterns Hetzner supports
- Network attachment with public interface control — private networking is critical for security
- Delete protection — production safety net

**Excluded (the 20%):**
- `network_zone` on the load balancer — `location` is more explicit and covers the same functionality. Network zone is inferred from the location.
- `subnet_id` on the network attachment — `network_id` is simpler and sufficient for 95% of cases. The provider auto-discovers the appropriate subnet.
- Deprecated `target` inline list on the load balancer resource — replaced by separate target resources, which is what OpenMCF uses.
- Service-level `proxyprotocol` fine-tuning beyond the boolean toggle — the on/off switch covers the practical use case.

### Three Separate Target Lists

The spec defines three separate repeated fields for targets (`serverTargets`, `labelSelectorTargets`, `ipTargets`) instead of a single polymorphic list with a `type` discriminator. This is a deliberate design choice:

1. **Type safety** — Each target type has different fields (`serverId` vs. `selector` vs. `ip`). Separate messages eliminate the need for `oneof` semantics and make validation straightforward.
2. **Proto validation** — `buf.validate` rules are specific to each target type (e.g., `serverId` is required on `ServerTarget`, `selector` is `min_len: 1` on `LabelSelectorTarget`).
3. **YAML clarity** — Users see exactly which target type they are configuring without inspecting a `type` field.
4. **IaC simplicity** — The Pulumi and Terraform modules iterate over each list independently, setting the `type` field on the provider resource. No runtime type discrimination needed.

### Why All Four Resources Are Bundled

The four provider resources (`hcloud_load_balancer`, `hcloud_load_balancer_service`, `hcloud_load_balancer_target`, `hcloud_load_balancer_network`) form a single logical unit:

- A **service** cannot exist without a load balancer. It has no independent use.
- A **target** only receives traffic when services route to it. Standalone targets are meaningless.
- A **network attachment** configures how the load balancer communicates with targets. It is a property of the load balancer, not a standalone resource.

Unlike the `HetznerCloudServer` component (where volumes and snapshots have independent lifecycles and are separate components), nothing in the load balancer bundle should survive independently. Deleting the load balancer should delete all its services, targets, and network attachment. Bundling ensures this.

## Implementation Landscape

### Pulumi Module Architecture

The Pulumi module (`iac/pulumi/module/`) consists of four files:

- **`main.go`** — Orchestrates resource creation: initializes locals, creates the Hetzner Cloud provider, calls `loadBalancer()`
- **`locals.go`** — Extracts provider config and target resource from stack input, builds the label map with standard labels + user labels (standard wins on conflict)
- **`load_balancer.go`** — Core resource file. Creates all four resource types with helper functions for services, targets, network attachment, HTTP config, and health checks
- **`outputs.go`** — Constants for output names (`load_balancer_id`, `ipv4_address`, `ipv6_address`)

The module handles two forms of the load balancer ID: `lbIdStr` (string, for services) and `lbIdInt` (integer, for targets and network). This split exists because the Pulumi hcloud SDK uses `StringInput` for `LoadBalancerService.LoadBalancerId` but `IntInput` for `LoadBalancerTarget.LoadBalancerId` and `LoadBalancerNetwork.LoadBalancerId`.

Resource creation order is deliberate: load balancer first, then network attachment (if specified), then services and targets. Targets that use `usePrivateIp` have an explicit `DependsOn` on the network attachment to ensure the private network route exists before the target is created.

### Terraform Module Architecture

The Terraform module (`iac/tf/`) uses `for_each` for services and targets, and `count` for the conditional network attachment:

- Services are keyed by effective listen port via a `locals` block that computes default ports for HTTP/HTTPS
- Server targets are keyed by `server_id`, label selector targets by `selector`, IP targets by `ip`
- The network attachment uses `count = var.spec.network != null ? 1 : 0`
- All targets include `depends_on = [hcloud_load_balancer_network.this]` to handle the private IP dependency

Both modules apply the same default logic: algorithm defaults to `round_robin`, health check protocol defaults to match the service protocol (with HTTPS defaulting to HTTP health checks), and listen/destination ports use protocol-specific defaults.

## Production Best Practices

### Health Check Configuration

The default health check (TCP on the destination port, 15s interval, 10s timeout, 3 retries) is a reasonable starting point but should be customized for production:

- **Use HTTP health checks for HTTP/HTTPS services** — A TCP check only verifies the port is open. An HTTP check verifies the application is responding correctly.
- **Check a dedicated health endpoint** — Use a path like `/health` or `/ready` that returns 200 when the application is ready to serve traffic. Avoid checking `/` which may be expensive or require authentication.
- **Tune the interval for your traffic pattern** — Lower intervals (5-10s) detect failures faster but generate more health check traffic. Higher intervals (15-30s) are lighter but mean longer detection times.
- **Keep timeout well below interval** — A timeout of 10s with a 15s interval works. A timeout of 14s with a 15s interval means a single slow response delays the next check.
- **Set retries based on tolerance** — 3 retries with a 15s interval means a failed server stays in rotation for up to 45 seconds. For latency-sensitive services, reduce retries to 2.

### Algorithm Selection

- **`round_robin`** — Use when all targets have similar capacity and response times. Simpler, more predictable. Good default.
- **`least_connections`** — Use when targets have different capacities (e.g., mixed server types) or when request processing times vary significantly. Also better when using sticky sessions, as it compensates for uneven session distribution.

### Private Networking

For production workloads, attach the load balancer to a private network and set `usePrivateIp: true` on all server and label selector targets. This keeps backend traffic off the public internet, reducing latency and improving security. The load balancer's public interface remains enabled to accept client traffic.

For internal-only load balancers (e.g., a database connection pool), set `enablePublicInterface: false` on the network attachment. The load balancer is then only reachable via its private network IP.

### TLS Configuration

- Always use HTTPS services for public-facing web traffic
- Enable `redirectHttp: true` to ensure all HTTP traffic is upgraded to HTTPS
- Use `HetznerCloudCertificate` resources with `valueFrom` references instead of hardcoded certificate IDs — this ensures certificate rotation is captured in the infrastructure graph
- For applications that terminate TLS themselves, use TCP pass-through with PROXY protocol to preserve client IP information

### Target Management

- **Prefer label selector targets for auto-scaling** — When servers are created and destroyed dynamically, label selectors automatically adjust the target pool. The manifest does not need updating when the server count changes.
- **Use server targets for fixed pools** — When you have a known, stable set of servers, explicit server targets with `valueFrom` references provide full traceability.
- **Use IP targets sparingly** — These are for external backends that are not managed by Hetzner Cloud. They cannot use private IP routing.

### Sizing

- Start with `lb11` for development and staging
- Use `lb21` for production workloads with up to 75 targets
- Use `lb31` for high-traffic services or large server pools (up to 150 targets)
- The type can be changed without downtime (in-place resize), so start small and scale up based on observed connection counts

## Conclusion

The Hetzner Cloud Load Balancer is the infrastructure component that turns a set of servers into a production service. It is conceptually simple — distribute traffic, check health, handle TLS — but operationally complex due to the four distinct provider resources, mixed ID types, and dependency ordering requirements.

OpenMCF's `HetznerCloudLoadBalancer` component absorbs this operational complexity into a single validated manifest. The proto schema enforces correct field types and required fields at parse time. The IaC modules handle resource ordering, ID type conversion, and default application. Foreign key references via `valueFrom` connect the load balancer to servers, networks, and certificates without hardcoded IDs.

For teams using Hetzner Cloud for web applications, the load balancer component is typically the final piece deployed after servers, networks, firewalls, and certificates are in place. Its manifest reads as a description of the desired traffic flow, not a sequence of API calls.

### References

- [Hetzner Cloud Load Balancers Documentation](https://docs.hetzner.cloud/#load-balancers)
- [Hetzner Cloud Load Balancer Pricing](https://www.hetzner.com/cloud/load-balancer)
- [Terraform hcloud Provider — Load Balancer Resources](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs)
- [Pulumi hcloud Package — Load Balancer Resources](https://www.pulumi.com/registry/packages/hcloud/api-docs/)
- [Hetzner Cloud API Reference — Load Balancers](https://docs.hetzner.cloud/#load-balancers)

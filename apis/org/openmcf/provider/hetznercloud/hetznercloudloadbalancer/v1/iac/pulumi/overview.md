# HetznerCloudLoadBalancer Pulumi Module — Architecture Overview

## Data Flow

```
manifest.yaml
  └─> HetznerCloudLoadBalancerStackInput (proto)
        ├── target: HetznerCloudLoadBalancer
        │     ├── metadata.name → load balancer name in Hetzner Cloud
        │     ├── metadata.org, env, id, labels → label computation
        │     └── spec
        │           ├── load_balancer_type (string, required) → LB size (lb11/lb21/lb31)
        │           ├── location (string, required) → datacenter
        │           ├── algorithm (enum, optional) → round_robin or least_connections
        │           ├── services (Service[], required, min 1) → listener configuration
        │           │     ├── protocol (enum, required) → http, https, or tcp
        │           │     ├── listen_port (optional int) → default 80 (http), 443 (https)
        │           │     ├── destination_port (optional int) → default = listen_port
        │           │     ├── proxyprotocol (bool) → PROXY protocol v1
        │           │     ├── http (HttpConfig, optional) → sticky sessions, certs, redirect
        │           │     └── health_check (HealthCheck, optional) → custom health check
        │           ├── server_targets (ServerTarget[], optional) → static server backends
        │           ├── label_selector_targets (LabelSelectorTarget[], optional) → dynamic backends
        │           ├── ip_targets (IpTarget[], optional) → external IP backends
        │           ├── network (NetworkAttachment, optional) → private network config
        │           │     ├── network_id (StringValueOrRef, required)
        │           │     ├── ip (string, optional) → fixed private IP
        │           │     └── enable_public_interface (optional bool, default true)
        │           └── delete_protection (bool) → API deletion guard
        └── provider_config: HetznerCloudProviderConfig
              └── hcloud_token (or HCLOUD_TOKEN env var)
```

## Module Structure

1. **main.go (entrypoint)**: Loads `HetznerCloudLoadBalancerStackInput` from the `STACK_INPUT` environment variable (base64-encoded YAML) via `stackinput.LoadStackInput`, then calls `module.Resources`.

2. **module/main.go**: Orchestrates resource creation:
   - Initializes locals from stack input
   - Creates a Hetzner Cloud Pulumi provider via `pulumihcloudprovider.Get`
   - Calls `loadBalancer()` to create all resources and export outputs

3. **module/locals.go**: Extracts provider config and target resource, then builds the label map:
   - Standard labels are set from metadata (`resource`, `name`, `kind`, `org`, `env`, `id`)
   - User-specified `metadata.labels` are merged in; standard labels take precedence on key conflicts

4. **module/load_balancer.go**: The core resource file. Creates up to four resource types with extensive helper functions:

   **Load balancer creation:** Creates `hcloud.NewLoadBalancer` with name, type, location, labels, delete protection, and algorithm. Produces two forms of the load balancer ID: `lbIdStr` (string, for services) and `lbIdInt` (integer, for targets and network).

   **Network attachment (optional):** Guarded by `if spec.Network != nil`. Creates `hcloud.NewLoadBalancerNetwork` with network ID (int-converted from string), optional fixed IP, and public interface toggle. Created before targets so private IP routing is available.

   **Services:** Iterates over `spec.Services`, creating one `hcloud.NewLoadBalancerService` per entry. Keyed by effective listen port (CG02 pattern). Helper functions:
   - `effectiveListenPort()` — defaults to 80 (HTTP), 443 (HTTPS), or 0 (TCP)
   - `effectiveDestinationPort()` — defaults to listen port
   - `buildHttpConfig()` — converts proto `HttpConfig` to Pulumi args, handles certificate ID int-conversion
   - `buildHealthCheck()` — converts proto `HealthCheck` to Pulumi args, applies defaults for protocol/port/interval/timeout/retries

   **Targets:** Iterates over all three target lists, creating one `hcloud.NewLoadBalancerTarget` per entry:
   - Server targets: keyed by server ID, type `"server"`, server ID int-converted, optional `DependsOn` for private IP
   - Label selector targets: keyed by sanitized selector, type `"label_selector"`, optional `DependsOn` for private IP
   - IP targets: keyed by sanitized IP, type `"ip"`, no private IP support

   **Output export:** Exports `load_balancer_id` (from `.ID()`), `ipv4_address` (from `.Ipv4`), `ipv6_address` (from `.Ipv6`).

5. **module/outputs.go**: Constants for output names (`load_balancer_id`, `ipv4_address`, `ipv6_address`), matching the `stack_outputs.proto` field names.

## Resource Graph

```
hcloud.LoadBalancer ("load-balancer")
  │
  ├── [if spec.Network != nil] hcloud.LoadBalancerNetwork ("network")
  │     ├── LoadBalancerId ← lb.ID() (int-converted via ApplyT)
  │     ├── NetworkId      ← spec.Network.NetworkId (int-converted from string)
  │     ├── [if ip set] Ip ← spec.Network.Ip
  │     └── EnablePublicInterface ← spec.Network.EnablePublicInterface (default true)
  │
  ├── [for each service] hcloud.LoadBalancerService ("service-{listenPort}")
  │     ├── LoadBalancerId  ← lb.ID() (string — no conversion needed)
  │     ├── Protocol        ← service.Protocol.String()
  │     ├── ListenPort      ← effectiveListenPort(service)
  │     ├── DestinationPort ← effectiveDestinationPort(service, listenPort)
  │     ├── Proxyprotocol   ← service.Proxyprotocol
  │     ├── [if http != nil && protocol != tcp] Http
  │     │     ├── StickySessions ← http.StickySessions
  │     │     ├── CookieName     ← http.CookieName (if non-empty)
  │     │     ├── CookieLifetime ← http.CookieLifetime (if > 0)
  │     │     ├── Certificates   ← http.CertificateIds[] (each int-converted)
  │     │     └── RedirectHttp   ← http.RedirectHttp
  │     └── [if healthCheck != nil] HealthCheck
  │           ├── Protocol ← defaultHealthCheckProtocol(hc, svc)
  │           ├── Port     ← hc.Port or destPort
  │           ├── Interval ← hc.Interval or 15
  │           ├── Timeout  ← hc.Timeout or 10
  │           ├── Retries  ← hc.Retries or 3
  │           └── [if hc.Http != nil] Http
  │                 ├── Domain, Path, Response, Tls, StatusCodes
  │
  ├── [for each serverTarget] hcloud.LoadBalancerTarget ("target-server-{id}")
  │     ├── LoadBalancerId ← lb.ID() (int-converted)
  │     ├── Type           ← "server"
  │     ├── ServerId       ← target.ServerId (int-converted)
  │     ├── UsePrivateIp   ← target.UsePrivateIp
  │     └── [if usePrivateIp && network exists] DependsOn: [network]
  │
  ├── [for each labelSelectorTarget] hcloud.LoadBalancerTarget ("target-label-{sanitized}")
  │     ├── LoadBalancerId ← lb.ID() (int-converted)
  │     ├── Type           ← "label_selector"
  │     ├── LabelSelector  ← target.Selector
  │     ├── UsePrivateIp   ← target.UsePrivateIp
  │     └── [if usePrivateIp && network exists] DependsOn: [network]
  │
  ├── [for each ipTarget] hcloud.LoadBalancerTarget ("target-ip-{sanitized}")
  │     ├── LoadBalancerId ← lb.ID() (int-converted)
  │     ├── Type           ← "ip"
  │     └── Ip             ← target.Ip
  │
  ├── Export: "load_balancer_id" ← lb.ID()
  ├── Export: "ipv4_address"     ← lb.Ipv4
  └── Export: "ipv6_address"     ← lb.Ipv6
```

## Key Design Points

- **Two forms of load balancer ID**: The Pulumi hcloud SDK uses `StringInput` for `LoadBalancerService.LoadBalancerId` but `IntInput` for `LoadBalancerTarget.LoadBalancerId` and `LoadBalancerNetwork.LoadBalancerId`. The module prepares both: `lbIdStr` from `createdLb.ID().ToStringOutput()` and `lbIdInt` via `createdLb.ID().ApplyT(strconv.Atoi)`. This is the only component in the catalog that needs two ID representations of the same resource.

- **Network-before-targets ordering**: When a network attachment is specified, `createNetworkAttachment()` runs before `createTargets()`. Targets with `usePrivateIp: true` add an explicit `pulumi.DependsOn([]pulumi.Resource{createdNetwork})` to ensure the private network route is established before the target is created. Without this, the target creation may fail with "not attached to network" errors.

- **Service keying by listen port (CG02)**: Each service resource is named `"service-{listenPort}"`. Listen port is unique per load balancer (enforced by the provider), making it a natural key for Pulumi resource names and for `for_each` in Terraform.

- **Health check protocol defaulting**: The `defaultHealthCheckProtocol()` function applies a non-obvious default: HTTPS services default to HTTP health checks (not HTTPS), because the load balancer terminates TLS and backends typically serve plain HTTP on the destination port. TCP services default to TCP health checks. This matches the Hetzner Cloud provider's behavior.

- **Port defaulting helpers**: `effectiveListenPort()` applies protocol-specific defaults (80 for HTTP, 443 for HTTPS). `effectiveDestinationPort()` defaults to the listen port when not explicitly set. These helpers are used both for building the service args and for computing the health check port default.

- **Certificate ID conversion**: Certificate IDs in the proto spec are `StringValueOrRef` (strings), but the Pulumi SDK expects `pulumi.IntArray`. `buildHttpConfig()` converts each certificate ID from string to int via `strconv.Atoi`, collecting them into a `[]pulumi.IntInput` slice.

- **Sanitize helpers for resource naming**: `sanitizeSelector()` replaces `=`, `,`, and spaces with hyphens (e.g., `env=production,role=web` becomes `env-production-role-web`). `sanitizeIp()` replaces `.` and `:` with hyphens (e.g., `203.0.113.50` becomes `203-0-113-50`). Both produce Pulumi-safe resource name components.

- **Label merge strategy**: Standard labels always win over user labels, preventing users from overriding management metadata. Labels are applied only to the load balancer resource — service, target, and network resources do not support labels in the Hetzner Cloud API.

- **Single resource file**: Despite the component's complexity (four resource types, eight helper functions), all resource creation lives in one file (`load_balancer.go`). This is appropriate because all resources are tightly coupled and created in a single function call chain. Helper functions (`createServices`, `createTargets`, `createNetworkAttachment`, `buildHttpConfig`, `buildHealthCheck`, etc.) provide modularity within the file.

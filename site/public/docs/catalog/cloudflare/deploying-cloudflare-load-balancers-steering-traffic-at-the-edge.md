---
title: "Deploying Cloudflare Load Balancers: Steering Traffic at the Edge"
description: "Deploying Cloudflare Load Balancers: Steering Traffic at the Edge deployment documentation"
icon: "package"
order: 100
componentName: "cloudflareloadbalancer"
---

# Deploying Cloudflare Load Balancers: Steering Traffic at the Edge

## Introduction

The conventional wisdom for decades was clear: if you want to load balance traffic across data centers, you need hardware. F5 appliances in your NOC. NGINX clusters at the edge of each region. DNS round-robin as a poor man's failover. The complexity was assumed to be unavoidable—global traffic management meant global infrastructure, and global infrastructure meant big budgets and bigger ops teams.

Cloudflare Load Balancer challenges that assumption entirely. It's not a virtual appliance you deploy. It's not software you install. It's a control plane that lives in Cloudflare's global network—over 330 data centers spanning every continent—that intelligently steers DNS responses and HTTP traffic based on real-time health checks, geographic proximity, and custom steering policies. You don't manage servers. You define policies. Cloudflare does the rest.

**Cloudflare Load Balancer** is a Global Server Load Balancing (GSLB) solution operating at the DNS level, tightly integrated with Cloudflare's reverse proxy and CDN infrastructure. Unlike static DNS round-robin, which blindly distributes requests regardless of server health, Cloudflare Load Balancer continuously probes your origin servers via health checks and dynamically routes traffic *away* from failures and *toward* healthy endpoints. This makes it a critical component for high-availability architectures, multi-cloud strategies, and geo-distributed applications where downtime isn't an option.

The service operates in two fundamental modes:

1. **Proxied (Orange Cloud)**: Traffic flows through Cloudflare's Layer 7 reverse proxy, enabling session affinity, WAF protection, CDN caching, and advanced routing rules. This is the default and most common mode.

2. **DNS-Only (Gray Cloud)**: Cloudflare's authoritative DNS returns the IP of a healthy origin directly to clients, bypassing the proxy. This mode is used for non-HTTP traffic or when proxying is undesirable.

Think of Cloudflare Load Balancer not as a load balancer in the traditional sense, but as a **steering control plane** with pluggable data planes. You can steer Layer 7 HTTP/S traffic through the proxy, Layer 4 TCP/UDP traffic through Spectrum, or even just DNS responses for protocols Cloudflare doesn't proxy.

### When to Use Cloudflare Load Balancer

Use Cloudflare Load Balancer when you need:

- **Global traffic management (GTM)**: Distribute traffic across geographically dispersed data centers to reduce latency and improve fault tolerance.
- **Active-passive failover**: Automatically route traffic from a primary origin to a backup/DR site when health checks fail, with failover times measured in seconds (not minutes).
- **Geo-steering**: Reduce latency by routing users to the origin pool closest to them, or enforce data sovereignty by keeping EU users' traffic within EU borders.
- **Multi-cloud or hybrid-cloud abstraction**: Balance traffic across origins in AWS, GCP, Azure, and on-premises data centers from a single, vendor-neutral control plane.
- **Blue-green deployments and A/B testing**: Use weighted traffic distribution to send a controlled percentage of production traffic to a new application version before full rollout.

### What Makes It Different

**vs. DNS Round-Robin**: Standard DNS allows multiple A records for one hostname, providing basic load distribution. But it has no health awareness, can't fail over intelligently, and offers no control over traffic steering. Cloudflare Load Balancer is the correct choice when high availability and intelligent routing are required.

**vs. Argo Smart Routing**: These two products are complementary, not competitive. The Load Balancer *selects* the optimal origin pool (e.g., "US-East" vs. "EU-West"). Then, *after* that decision, Argo finds the fastest network path from Cloudflare's edge to the chosen origin. A complete performance architecture often uses both.

**vs. Cloudflare Spectrum**: The standard Load Balancer is a Layer 7 (HTTP/S) product. Spectrum is a Layer 4 (TCP/UDP) reverse proxy for non-web applications like game servers, databases, and IoT protocols. You can combine them: the Load Balancer provides GSLB logic (health checks, steering) for applications proxied by Spectrum.

This document walks through the evolution of deployment methods for Cloudflare Load Balancer—from manual anti-patterns to production-grade Infrastructure-as-Code—and explains how OpenMCF abstracts the complexity into a clean, developer-friendly API.

---

## The Deployment Maturity Spectrum

Not all approaches to deploying Cloudflare Load Balancers are equally suitable for production. Here's the progression from anti-patterns to production-ready solutions.

### Level 0: The Manual Dashboard (Anti-Pattern for Production)

**What it is**: Using the Cloudflare web dashboard to manually create Monitors, Pools, and Load Balancers through a point-and-click workflow.

**The workflow**:
1. Create a Monitor (define health check path, expected HTTP codes, interval)
2. Create a Pool (define origin servers, attach the Monitor)
3. Create a Load Balancer (assign the hostname, select Pools, configure steering policy)

**What it solves**: Nothing that can't be solved better another way. The dashboard is valuable for learning the Cloudflare model, visualizing traffic flow in real-time, and performing emergency "break-glass" changes during incidents. It's an excellent tool for understanding *what* a load balancer does.

**What it doesn't solve**: Repeatability, version control, auditability, or disaster recovery. The biggest operational risk is the **timing anti-pattern**: creating a load balancer and immediately enabling it before its associated pools have been marked healthy by the monitoring system. Cloudflare's documentation explicitly warns against this, but it's trivially easy to do in the UI. The result? Production traffic routed to origins that haven't passed health checks.

**Verdict**: Use it to explore, visualize, and troubleshoot. Never use it as the source of truth for production infrastructure. If it's not in code, it's not reproducible.

---

### Level 1: CLI Scripting with `cloudflared` (Niche Use Case)

**What it is**: A common misconception is that Cloudflare's CLI tools can manage Load Balancers. They can't—at least not in the way most engineers assume.

- **`wrangler` CLI**: Not applicable. Its scope is strictly the Developer Platform: Workers, Pages, KV, Durable Objects. It cannot create or modify Load Balancers.
  
- **`cloudflared` CLI**: Highly relevant for a *specific integration pattern*, but cannot create the Load Balancer resource itself. Its purpose is to manage **Cloudflare Tunnels**—secure connections from private networks to Cloudflare's edge. The command `cloudflared tunnel route lb` registers a tunnel as an origin *within* an existing, pre-configured Load Balancer pool.

**What it solves**: Enables hybrid-cloud architectures where private origins (on-premises servers, VPCs without public IPs) participate in load balancing. Once a tunnel is established, you can add its UUID-based hostname (`<tunnel-uuid>.cfargotunnel.com`) as an origin in a pool.

**What it doesn't solve**: Creating the Load Balancer, Pools, or Monitors themselves. Those must still be provisioned via the dashboard, API, or IaC.

**Verdict**: Essential for hybrid-cloud scenarios with Cloudflare Tunnel integration. Not a standalone deployment method.

---

### Level 2: Direct API Integration (Flexible but High-Maintenance)

**What it is**: Calling the Cloudflare REST API directly using HTTP clients (`curl`, custom scripts, or SDKs in Go, Python, or TypeScript).

**The API design reveals a critical complexity**: Monitors and Pools are **account-level resources**, while Load Balancers are **zone-level resources**.

- Monitors: `POST /accounts/{account_id}/load_balancers/monitors`
- Pools: `POST /accounts/{account_id}/load_balancers/pools`
- Load Balancers: `POST /zones/{zone_id}/load_balancers`

This means a Load Balancer resource in a Zone depends on Pools and Monitors that live *outside* that zone, at the parent account level. This design implies that Pools and Monitors are intended to be shared and reused across multiple zones and load balancers. It's an architectural decision that optimizes for reusability at the cost of operational complexity.

**What it solves**: Maximum flexibility. You can integrate Cloudflare Load Balancer management into any language or tool that speaks HTTP. Cloudflare provides official, modern SDKs for **Go, Python, and TypeScript** that wrap the API in typed, idiomatic interfaces.

**What it doesn't solve**: State management, dependency orchestration, or idempotency. When you create a Load Balancer via the API, you must:
1. Create the Monitor and capture its ID
2. Create the Pool(s), referencing the Monitor ID, and capture the Pool ID(s)
3. Create the Load Balancer, referencing the Pool IDs
4. Handle failures at any step and implement rollback or cleanup logic

You're essentially building your own Infrastructure-as-Code layer. For most teams, that's wasted effort.

**Verdict**: Useful if you're building a custom control plane, integrating Cloudflare into a broader orchestration system, or implementing a Kubernetes operator. For standard infrastructure provisioning, higher-level IaC tools handle the API calls and state management for you.

---

### Level 3: Infrastructure-as-Code (Production-Ready)

**What it is**: Using declarative IaC tools—Terraform, Pulumi, or OpenTofu—with Cloudflare's official providers to define Load Balancers, Pools, and Monitors as code.

This is the dominant and most mature deployment method. Cloudflare itself uses Terraform to manage its own infrastructure, and the company has publicly endorsed the `cloudflare/cloudflare` Terraform provider as production-ready.

**Example: Active-Passive Failover with Terraform**

This configuration routes all traffic to a primary pool. If health checks fail, traffic automatically moves to the secondary pool:

```hcl
# 1. Define the Health Check (Account-level)
resource "cloudflare_load_balancer_monitor" "http_check" {
  account_id       = var.cloudflare_account_id
  type             = "https"
  path             = "/health"
  expected_codes   = "200"
  interval         = 60  # 60s minimum for Pro plan
  timeout          = 5
  retries          = 2
  follow_redirects = true
}

# 2. Define the Primary Pool (Account-level)
resource "cloudflare_load_balancer_pool" "primary" {
  account_id       = var.cloudflare_account_id
  name             = "primary-pool-us-east"
  monitor          = cloudflare_load_balancer_monitor.http_check.id
  
  origins {
    name    = "primary-origin-1"
    address = "198.51.100.1"
    enabled = true
  }
}

# 3. Define the Secondary/Backup Pool (Account-level)
resource "cloudflare_load_balancer_pool" "secondary" {
  account_id       = var.cloudflare_account_id
  name             = "secondary-pool-us-west"
  monitor          = cloudflare_load_balancer_monitor.http_check.id
  
  origins {
    name    = "backup-origin-1"
    address = "198.51.100.2"
    enabled = true
  }
}

# 4. Define the Load Balancer (Zone-level)
resource "cloudflare_load_balancer" "failover_lb" {
  zone_id = var.cloudflare_zone_id
  name    = "app.example.com"
  proxied = true  # Orange cloud, enables WAF, caching, session affinity
  
  # steering_policy "off" means "use default_pools as a priority list"
  steering_policy = "off"
  
  # Failover priority: primary -> secondary
  default_pool_ids = [
    cloudflare_load_balancer_pool.primary.id,
    cloudflare_load_balancer_pool.secondary.id
  ]
  
  # Pool of last resort. If both primary and secondary fail, send here.
  fallback_pool_id = cloudflare_load_balancer_pool.secondary.id
  
  # Enable sticky sessions
  session_affinity = "cookie"
}
```

**What it solves**:
- **Declarative configuration**: State *what* you want, not *how* to achieve it
- **State management**: Terraform tracks what exists and what changed
- **Idempotency**: Running `terraform apply` twice produces the same result
- **Version control**: Configuration lives in Git, with history and code review
- **Dependency orchestration**: Terraform automatically creates resources in the correct order (Monitor → Pool → Load Balancer) and handles references by ID
- **Drift detection**: `terraform plan` shows configuration changes made outside of IaC

**What it doesn't solve out-of-the-box**: The critical operational challenge is **secret management**. By default, Terraform stores its state file (`terraform.tfstate`) in **plain text**. If a Cloudflare API token is defined as a variable, it will be visible in the state file. This is a significant security risk.

The established best practice is to:
1. Never hardcode secrets in `.tf` files
2. Inject the API token at runtime via environment variables (`CLOUDFLARE_API_TOKEN`)
3. Store the state file in a remote backend (S3, Terraform Cloud) with encryption at rest and strict access controls
4. Use external secret management (HashiCorp Vault, AWS Secrets Manager) to provide tokens to CI/CD pipelines

**Verdict**: This is the recommended method for production. Terraform has the largest ecosystem, the most battle-tested provider, and is internally used by Cloudflare. OpenTofu (the open-source fork of Terraform) is fully compatible with the `cloudflare/cloudflare` provider.

---

### Level 4: Kubernetes-Native IaC with Crossplane (Advanced)

**What it is**: Using the Crossplane Cloudflare provider to manage Load Balancers as Kubernetes Custom Resources (CRs) directly from within a cluster.

**Example**:

```yaml
apiVersion: cloudflare.crossplane.io/v1alpha1
kind: LoadBalancer
metadata:
  name: app-lb
spec:
  forProvider:
    name: app.example.com
    zoneId: abc123
    proxied: true
    defaultPoolIds:
      - pool-primary-id
    fallbackPoolId: pool-secondary-id
```

**What it solves**: For teams operating in a Kubernetes-first workflow, Crossplane allows Load Balancers to be managed with the same GitOps practices used for application deployments. You can define Cloudflare infrastructure in the same repository and CI/CD pipeline as your Kubernetes manifests.

**What it doesn't solve**: The same state management and secret handling challenges as Terraform. Crossplane is essentially Terraform-like functionality wrapped in Kubernetes CRDs. It's not inherently more secure or simpler—it's just Kubernetes-native.

**Verdict**: Excellent for teams already invested in Crossplane for multi-cloud orchestration. Overkill if you're not already using Crossplane for other infrastructure.

---

## IaC Provider Deep Dive: Terraform vs. Pulumi

For teams building an IaC framework, the choice between Terraform and Pulumi for Cloudflare Load Balancer management comes down to trade-offs in secret management, language flexibility, and operational maturity.

### Production Readiness

**Terraform**: The `cloudflare/cloudflare` provider is exceptionally mature. Cloudflare has publicly endorsed it and uses it internally. Load Balancing resources have been supported since v1.0 of the provider. It's the most widely adopted and battle-tested option.

**Pulumi**: The `pulumi-cloudflare` provider is also production-ready. Pulumi's architecture can bridge Terraform providers, allowing it to maintain resource parity with the Terraform provider and adopt upstream fixes rapidly. While its user base is smaller, it's fully capable of managing complex load balancer configurations.

**Verdict**: Both are production-ready. Maturity is not a deciding factor.

### Resource Coverage

Both providers offer complete, 1:1 coverage for all Cloudflare Load Balancer components:

**Terraform**:
- `cloudflare_load_balancer`
- `cloudflare_load_balancer_pool`
- `cloudflare_load_balancer_monitor`

**Pulumi**:
- `cloudflare.LoadBalancer`
- `cloudflare.LoadBalancerPool`
- `cloudflare.LoadBalancerMonitor`

**Verdict**: Parity. Both fully cover the resource model.

### Secret Management (The Key Differentiator)

This is where the two tools diverge most significantly.

**Terraform**: By default, the state file is stored in **plain text**. If an API token is defined as a variable, it will be visible in the state file. The recommended mitigation is:
- Use environment variables (`CLOUDFLARE_API_TOKEN`) instead of hardcoding tokens
- Store the state file in a remote backend with encryption at rest
- Use external secret management (HashiCorp Vault, AWS Secrets Manager)

This is not a Terraform flaw—it's an architectural choice that prioritizes simplicity and transparency. But it requires external tooling and discipline to secure.

**Pulumi**: Designed with this problem in mind. Pulumi provides **built-in, default-on secret management**. Any configuration value can be marked as a secret:

```python
import pulumi

config = pulumi.Config()
api_token = config.require_secret("cloudflare-api-token")
```

Pulumi automatically encrypts the secret before storing it in the state backend—whether using the managed Pulumi Service or a self-hosted backend (S3, Azure Blob, GCS). This significantly simplifies the security posture.

**Verdict**: Pulumi's secret management is simpler and more secure out-of-the-box. For teams managing sensitive credentials, this is a compelling advantage.

### Multi-Environment Patterns (dev/staging/prod)

The single most important best practice for managing multiple environments with IaC is to use **separate Cloudflare accounts** (e.g., `company-dev`, `company-prod`) with **separate domains** (e.g., `example-staging.com`, `example.com`).

This is not optional—it's a structural necessity dictated by the Cloudflare API design. As noted earlier, Monitors and Pools are **account-level resources**. If "dev" and "prod" environments were to exist within the *same account* (e.g., as separate load balancers on `dev.example.com` and `prod.example.com`), their IaC configurations would reference the same global pool of Monitors and Pools.

**The risk**: A `terraform apply` in the "dev" workspace could inadvertently modify or delete a production-critical Pool or Monitor, causing a catastrophic production outage.

**The solution**: Account-level segregation is the only pattern that guarantees complete environment isolation. The recommended directory structure reflects this:

```
terraform/
├── accounts/
│   ├── prod/
│   │   ├── zones/
│   │   │   ├── example.com/
│   │   │   │   ├── load_balancer/
│   │   │   │   │   ├── main.tf
│   │   ├── global/
│   │   │   ├── pools/
│   │   │   ├── monitors/
│   ├── dev/
│   │   ├── zones/
│   │   │   ├── example-staging.com/
```

**Verdict**: The core challenge (account-level resources) is identical for both Terraform and Pulumi. Neither tool solves this problem—it must be addressed through architectural discipline.

### Comparison Table

| Feature | Terraform (with OpenTofu) | Pulumi |
|---------|---------------------------|--------|
| **Maturity** | Excellent. Used internally by Cloudflare. | Excellent. Often bridges the Terraform provider. |
| **Resource Coverage** | Complete (`cloudflare_load_balancer`, `cloudflare_load_balancer_pool`, `cloudflare_load_balancer_monitor`) | Complete (`cloudflare.LoadBalancer`, `cloudflare.LoadBalancerPool`, `cloudflare.LoadBalancerMonitor`) |
| **Secret Management** | Poor (default). State is plain text. Requires external tooling (Vault) or env vars. | Excellent (default). Secrets are encrypted in state by default. |
| **Language** | HCL (Declarative DSL) | General-purpose (Go, Python, TypeScript, etc.) |
| **State Management** | Manual setup for remote state (e.g., S3) | Managed service (Pulumi Service) is the default |
| **Multi-Env Pattern** | Requires disciplined directory structure or workspaces | Same. Requires disciplined code/stack structure |

**OpenMCF's Choice**: The superior secret management and modern SDK experience make Pulumi a strong fit for teams prioritizing security and developer experience. However, Terraform's ecosystem maturity and Cloudflare's internal endorsement make it the safe, conservative choice.

---

## Production Essentials: Health Checks, Failover, and Traffic Steering

### Health Monitoring: The Brain of the System

Health monitors are the *brain* of the load balancer. A misconfigured health check renders the entire service useless and can lead to outages.

**Configuration Essentials**:
- **Type**: `http`, `https`, or `tcp`
- **Path**: The endpoint to check (e.g., `/health`)
- **Expected Codes**: HTTP status codes that signal a healthy response (e.g., `200` or `2xx`)
- **Interval**: How frequently to probe (minimum interval is plan-gated—see Cost Considerations below)

**Quorum System**: Health is determined by a quorum. Cloudflare probes from **three separate data centers** within each configured region. An origin is marked healthy only if the *majority* of probes in that region succeed. A *pool* is considered healthy if the number of healthy origins it contains remains at or above its configured `health_threshold`.

**The RTO Trade-Off**: The minimum health check interval is the single largest factor in your Recovery Time Objective (RTO):

- **Pro Plan**: 60 seconds (minimum)
- **Business Plan**: 15 seconds (minimum)
- **Enterprise Plan**: 10 seconds (minimum)

A user on a Pro plan *cannot* achieve a failover time of less than 60 seconds. The "cost" of high availability is a mandatory plan upgrade.

### Failover Patterns

**Active-Passive Failover** is the most common resilience pattern. The configuration is counter-intuitive:

1. `steering_policy` must be set to **`"off"`** (not `"failover"`)
2. The `default_pool_ids` list is treated as an **ordered priority list**
3. Traffic is sent *only* to the first healthy pool in the list

**Fallback Pool**: A `fallback_pool_id` is required. This is the pool of last resort. If all pools in the `default_pool_ids` list are marked unhealthy, Cloudflare will send all traffic to the `fallback_pool_id` *regardless of its own health status*. This ensures traffic continues flowing, even if it's to a degraded origin.

**Failback**: When a higher-priority pool (e.g., the primary pool) recovers and is marked healthy again, the load balancer will automatically "fail back," shifting traffic away from the passive pool and back to the primary.

### Traffic Steering Policies

The `steering_policy` parameter dictates how the load balancer selects a pool:

- **`"off"`** (Default): Disables dynamic steering and enables Active-Passive Failover. Uses `default_pool_ids` as a static failover priority.
- **`"random"`**: Distributes traffic randomly across pools based on configured weights. Used for active-active load balancing or A/B testing.
- **`"geo"`**: Steers traffic based on the user's geographic location, matching it against `region_pools` or `country_pools` maps.
- **`"dynamic_latency"`** (Enterprise only): Steers traffic to the pool with the lowest measured Round Trip Time (RTT) from the Cloudflare edge PoP.

### Session Affinity (Sticky Sessions)

Session affinity ensures that subsequent requests from the same end-user are "stuck" to the same origin server. This is critical for stateful applications like e-commerce shopping carts.

**Configuration**: This feature is **only supported on proxied (orange-cloud) load balancers**. When `session_affinity = "cookie"` is set, Cloudflare generates a `__cflb` cookie on the client's first request. This cookie is used to route all subsequent requests to the same origin for the duration of the cookie's TTL.

### Common Anti-Patterns and Pitfalls

**Missing Health Checks**: Configuring a load balancer without an attached monitor provides no health awareness and will continue to route traffic to failed origins.

**Firewall Misconfiguration**: The origin server's firewall (e.g., iptables, AWS Security Group) blocks requests from Cloudflare's health check IP addresses. The probes fail, Cloudflare marks the origin as unhealthy, and traffic is failed over unnecessarily.

**Health Check Mismatch**:
- **Response Code Mismatch**: The monitor expects `expected_codes = "200"`, but the origin's health endpoint issues an HTTP 301 or 302 redirect. Solution: Set `follow_redirects = true` on the monitor.
- **TLS Mismatch**: An HTTPS monitor is used against an origin with a self-signed or invalid certificate. The TLS handshake fails. Solution: Use a free Cloudflare Origin CA certificate on the origin or, if security is not paramount, set `allow_insecure = true` on the monitor.

**No Fallback Pool**: Failing to define a `fallback_pool_id`. If all `default_pool_ids` become unhealthy, there is no "pool of last resort," and users will receive an error (e.g., HTTP 530).

**Single Origin**: Configuring a load balancer with a single pool containing only one origin. This provides no resilience, no failover, and serves no purpose beyond that of a standard DNS record.

---

## Cost Considerations: Pricing Model and Plan Requirements

### Plan Requirements

Cloudflare Load Balancing is a **paid add-on service**. It is **not available on the Free plan** (except as a paid add-on).

The primary limitations are gated by the underlying plan (Pro, Business, Enterprise):

- **Health Check Interval**: This is the most significant functional and financial trade-off.
  - **Pro Plan ($20/mo)**: 60 seconds (minimum interval)
  - **Business Plan ($200/mo)**: 15 seconds (minimum interval)
  - **Enterprise Plan (Custom)**: 10 seconds (minimum interval)

- **Resource Limits (Non-Enterprise)**:
  - Load Balancers: 20
  - Pools: 20
  - Origins (Endpoints): 20

### Detailed Pricing Model

For non-Enterprise plans, the pricing model is multi-vector but predictable:

1. **Base Fee**: **$5.00 / month**. This fee includes the first **two** origin servers.
2. **Additional Origin Fee**: **$5.00 / month *per additional origin***.
   - *Example: 4 origins = $5 (base for 2) + $10 (for 3rd & 4th) = $15/mo.*
3. **Geo-Routing Add-on**: **$10.00 / month** (flat fee). Required to enable `steering_policy = "geo"`.
4. **DNS Query Usage Fee**:
   - First 500,000 queries/month: **Free**.
   - Additional queries: **$0.50 per 500,000 queries**.

**Key Insight**: The usage-based DNS query fee is primarily relevant for **DNS-only (gray-cloud)** load balancers. The 80% use case of a **proxied (orange-cloud)** load balancer does not incur significant usage-based fees; its cost is a predictable flat monthly fee based on the number of origins and features. This contrasts favorably with public cloud load balancers, which often charge per-GB-processed.

### Health Check Frequency vs. Cost

There is **no direct, per-check cost** for health checks. The cost is bundled into the per-origin monthly fee.

The *true* trade-off is not cost-vs-frequency, but **Plan vs. Failover Speed**. To achieve a 15-second failover, an organization *must* subscribe to the Business plan, which carries a $200/month platform fee (vs. $20/month for Pro). The RTO is therefore a direct function of the monthly plan cost.

---

## OpenMCF's Abstraction: Simplifying the Complexity

### The Core Complexity: Account-Level vs. Zone-Level Resources

The single greatest source of complexity in the Cloudflare Load Balancer model is the separation of account-level resources (Pools, Monitors) from zone-level resources (Load Balancers). This forces users to define, link, and manage a dependency graph of 3-4 distinct resources just to create one load balancer.

When using Terraform directly, the workflow looks like this:

1. Create a `cloudflare_load_balancer_monitor` (account-level)
2. Capture its ID
3. Create a `cloudflare_load_balancer_pool` (account-level), referencing the monitor ID
4. Capture the pool ID(s)
5. Create a `cloudflare_load_balancer` (zone-level), referencing the pool IDs

This is verbose, error-prone, and requires deep knowledge of Cloudflare's resource model.

### OpenMCF's Solution: A Denormalized, Inline API

OpenMCF simplifies this by providing a **denormalized, "flattened" API resource**. The `CloudflareLoadBalancerSpec` allows users to define origins inline, without worrying about the underlying Pool and Monitor resources:

```protobuf
message CloudflareLoadBalancerSpec {
  // The DNS hostname for the load balancer (e.g., "app.example.com")
  string hostname = 1;
  
  // Foreign key reference to a Cloudflare DNS zone
  StringValueOrRef zone_id = 2;
  
  // List of origin servers (defined inline)
  repeated CloudflareLoadBalancerOrigin origins = 3;
  
  // Whether the LB is proxied (orange cloud). Default: true.
  bool proxied = 4;
  
  // HTTP path to use for health monitoring. Default: "/"
  string health_probe_path = 5;
  
  // Session affinity. "NONE" (default) or "COOKIE"
  CloudflareLoadBalancerSessionAffinity session_affinity = 6;
  
  // Traffic steering policy. "OFF" (failover), "GEO", or "RANDOM"
  CloudflareLoadBalancerSteeringPolicy steering_policy = 7;
}

message CloudflareLoadBalancerOrigin {
  string name = 1;
  string address = 2;
  int32 weight = 3;  // Default: 1
}
```

**What this abstraction provides**:

1. **No Resource Fan-Out**: The user defines origins inline. The OpenMCF controller handles the "fan-out" logic of creating the underlying Cloudflare Monitor, Pool, and Load Balancer resources and linking them by ID.

2. **Intuitive Defaults**: `proxied` defaults to `true` (the 80% case). `health_probe_path` defaults to `/` (the most common health endpoint).

3. **Simplified Steering**: The `steering_policy` enum uses intuitive values (`OFF`, `GEO`, `RANDOM`) that map to Cloudflare's less-obvious API values (`"off"`, `"geo"`, `"random"`).

4. **80/20 Configuration**: The API exposes only the fields that 80% of users need 80% of the time. Advanced use cases (custom health check headers, per-origin TLS settings, dynamic latency steering) can be added later without breaking the core API.

### Example: Active-Passive Failover with OpenMCF

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareLoadBalancer
metadata:
  name: app-lb
spec:
  hostname: app.example.com
  zone_id:
    ref: example-zone  # Foreign key to CloudflareDnsZone resource
  proxied: true
  health_probe_path: /health
  session_affinity: SESSION_AFFINITY_COOKIE
  steering_policy: STEERING_OFF  # Active-passive failover
  origins:
    - name: primary
      address: 198.51.100.1
      weight: 1
    - name: secondary
      address: 198.51.100.2
      weight: 1
```

Behind the scenes, the OpenMCF controller:
1. Creates a Cloudflare Monitor (type: `https`, path: `/health`, expected_codes: `200`)
2. Creates two Cloudflare Pools (one for each origin), both referencing the Monitor
3. Creates a Cloudflare Load Balancer with:
   - `steering_policy = "off"`
   - `default_pool_ids` set to the primary pool ID (index 0) and secondary pool ID (index 1)
   - `fallback_pool_id` set to the secondary pool ID
   - `session_affinity = "cookie"`

The user sees a simple, intuitive API. OpenMCF handles the complexity.

### Best Practices Codified in the API

The API design makes correct configuration easy and incorrect configuration difficult:

- **Default `proxied` to `true`**: The vast majority of use cases benefit from Cloudflare's Layer 7 proxy (WAF, caching, session affinity).
- **Require at least one origin**: The API validation enforces `repeated.min_items = 1` to prevent the "no origins" anti-pattern.
- **Sensible health check defaults**: The default `health_probe_path` is `/`, and the controller automatically uses `expected_codes = "200"`.

Future enhancements can add:
- Validation to require at least two origins for production environments
- Auto-configuration of `fallback_pool_id` to the last origin in the list
- Built-in geo-routing templates (e.g., "US-East + US-West + EU")

---

## Conclusion: From Global Complexity to Local Simplicity

Cloudflare Load Balancer represents a fundamental shift in how we think about traffic management. It's not a box you deploy. It's not software you maintain. It's a global control plane that makes intelligent, real-time decisions about where to send every request, based on health, geography, and custom policies you define.

The challenge has always been the operational complexity. Cloudflare's API model—with account-level Pools and Monitors referencing zone-level Load Balancers—optimizes for reusability at the cost of user-facing simplicity. Deploying a load balancer with raw Terraform requires orchestrating 3-4 dependent resources, managing IDs, and understanding Cloudflare's architectural quirks.

**OpenMCF abstracts this complexity**. By providing a denormalized, protobuf-defined API that allows inline origin definitions, it reduces the cognitive load on developers and operators. You define *what* you want—a load balancer for `app.example.com` with two origins and health checks—and OpenMCF handles the *how*.

This is the 80/20 principle in action: focus the API on the 20% of configuration that 80% of users need, and make it trivial to get right. Active-passive failover, geo-routing, and session affinity become simple YAML declarations, not multi-step orchestration puzzles.

For teams building multi-cloud, globally distributed systems, Cloudflare Load Balancer is an essential tool. And for teams using OpenMCF, deploying and managing it is as simple as defining a resource and letting the platform do the rest.


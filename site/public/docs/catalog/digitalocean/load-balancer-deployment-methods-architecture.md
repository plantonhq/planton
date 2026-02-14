---
title: "Load Balancer: Deployment Methods & Architecture"
description: "Load Balancer: Deployment Methods & Architecture deployment documentation"
icon: "package"
order: 100
componentName: "digitaloceanloadbalancer"
---

# DigitalOcean Load Balancer: Deployment Methods & Architecture

## Introduction

For years, load balancers were the infrastructure everyone needed but few wanted to manage. The conventional wisdom was simple: either pay for an expensive managed solution from your cloud provider, or roll your own HAProxy instance and become responsible for its uptime, security patches, and scaling—ironically creating a single point of failure while trying to eliminate one.

DigitalOcean Load Balancers represent a pragmatic middle ground. Built on battle-tested HAProxy under the hood, they offer the reliability of self-managed infrastructure with the operational simplicity of a managed service—at $12/month for regional load balancers. This isn't just about cost savings; it's about focusing engineering time on application logic rather than HAProxy configuration files.

The challenge isn't whether to use DigitalOcean Load Balancers, but **how** to deploy them in a way that's repeatable, auditable, and production-ready. From UI clicks to declarative IaC, this document maps the landscape of deployment methods and explains why OpenMCF chose to model them as protobuf-defined, Pulumi-deployed infrastructure resources.

## Understanding DigitalOcean's Load Balancer Architecture

### Two Distinct Products, Two Use Cases

DigitalOcean offers two fundamentally different load balancing products that must be understood as separate resources:

**Regional Load Balancers** are the workhorse. Operating at Layer 4 (TCP/UDP) and Layer 7 (HTTP/HTTPS/HTTP/2), they distribute traffic across Droplets within a single datacenter region. They support up to 40 forwarding rules, SSL termination with Let's Encrypt integration, tag-based backend discovery, and sophisticated health checks. This is what most teams need 95% of the time.

**Global Load Balancers** operate at a higher level, routing traffic across multiple regions using geo-proximity and failover logic. Their "backends" aren't Droplets—they're Regional Load Balancers themselves. With a usage-based pricing model (base $15/month + per-request charges), they serve a specific need: multi-region applications requiring intelligent geographic distribution.

OpenMCF models these as separate resource types because their APIs, protocols, backend targets, and pricing models are fundamentally incompatible. This documentation focuses on **Regional Load Balancers**, as they represent the 80% use case for infrastructure teams.

### Protocol Support: Layer 4 vs. Layer 7

Regional Load Balancers support both network-layer (L4) and application-layer (L7) protocols:

- **Layer 7 (Application)**: HTTP, HTTPS, HTTP/2, HTTP/3, and WebSockets. This enables SSL termination, modern protocol support (including gRPC over HTTP/2), and path-based routing.
- **Layer 4 (Network)**: TCP and UDP passthrough. Critical for non-HTTP services like databases (PostgreSQL on TCP 5432), message queues, or game servers (UDP). UDP load balancing has a gotcha: health checks must use TCP or HTTP on a different port, since UDP is connectionless.

The pricing reflects this capability matrix: Layer 7 (HTTP/HTTPS) load balancers are $12/month per node, while Layer 4 (Network) load balancers are $15/month per node.

### The Managed vs. Self-Managed Trade-off

Under the hood, DigitalOcean Load Balancers run HAProxy, the industry-standard proxy solution. This architectural choice frames the fundamental decision: **managed convenience vs. total control**.

**Managed (DigitalOcean LB)**:
- Fixed, predictable pricing ($12-15/month)
- Automatic high availability (multi-node by default)
- Free Let's Encrypt integration
- No OS patches, security updates, or failover management
- Limited to DigitalOcean's API configuration surface

**Self-Managed (HAProxy on Droplet)**:
- Complete control over HAProxy version, configuration, and logging
- Access to advanced features (complex ACLs, custom backends)
- Lower raw latency in some benchmarks (community reports show $5 Droplets outperforming managed LBs)
- Full operational burden: OS patching, monitoring, and—critically—ensuring the HAProxy Droplet itself is highly available

For teams practicing infrastructure-as-code, the managed service's API-driven configuration and built-in HA make it the pragmatic choice. You're trading the last 5% of configurability for elimination of operational toil.

### Integration with the DigitalOcean Ecosystem

Regional Load Balancers are designed to integrate deeply with DigitalOcean's platform:

- **VPC (Virtual Private Cloud)**: Load balancers should be created within a VPC to communicate with backend Droplets over the private, unmetered network. This is both a security and cost optimization best practice.
- **Droplet Tagging**: Dynamic backend discovery via tags (e.g., `web-prod`) enables autoscaling and blue-green deployments. The load balancer polls the API and automatically adjusts its pool as Droplets are created or destroyed with matching tags.
- **DOKS (DigitalOcean Kubernetes)**: Integration is transparent—creating a Kubernetes Service with `type: LoadBalancer` triggers the Cloud Controller Manager to provision a DigitalOcean LB automatically, configured via Service annotations.
- **DNS and SSL**: Load balancers receive a public IP that DNS A records point to. The Let's Encrypt feature requires DigitalOcean-managed DNS for automated certificate validation and renewal.

## The Maturity Spectrum: Deployment Methods from Manual to Declarative

### Level 0: The Anti-Pattern (ClickOps)

**What it is**: Using the DigitalOcean Control Panel's web UI to manually create load balancers through a wizard—selecting region, clicking Droplets, typing forwarding rules into form fields.

**Why teams do it**: Immediate gratification. The visual interface makes it easy to explore features and get a load balancer online in under 5 minutes during proof-of-concept work.

**Why it fails in production**:
- **Zero repeatability**: Creating a second load balancer in a different region means manually replicating 20+ clicks and hoping you remembered every setting.
- **No audit trail**: When the staging load balancer differs from production, you have no idea when, why, or who made the change.
- **Configuration drift**: Teams inevitably make manual "quick fixes" during incidents that never get documented or replicated across environments.

**Verdict**: Acceptable for learning the service or emergency debugging. Unacceptable for any infrastructure that matters.

### Level 1: Imperative Scripting (doctl CLI)

**What it is**: Using DigitalOcean's official CLI tool, `doctl`, to script load balancer creation and management with commands like:

```bash
doctl compute load-balancer create \
  --name load-balancer-1 \
  --region nyc1 \
  --forwarding-rules entry_protocol:http,entry_port:80,target_protocol:http,target_port:80
```

**What it solves**: Scripts are repeatable and version-controllable. You can document your infrastructure in shell scripts and run them consistently.

**What it doesn't solve**: Idempotency and complexity. The `doctl` tool exposes complex nested objects (forwarding rules, sticky sessions, health checks) as comma-separated key-value strings. This format is error-prone and fragile. Running the script twice creates two load balancers, not one configured load balancer. State management is manual—you must track resource IDs yourself.

**When to use it**: Ad-hoc administrative tasks, debugging production issues, or simple automation for teams already living in bash. Not suitable for managing production infrastructure at scale.

### Level 2: Configuration Management (Ansible)

**What it is**: Using the `community.digitalocean` Ansible collection to define load balancers as declarative YAML tasks:

```yaml
- name: Create production load balancer
  community.digitalocean.digital_ocean_load_balancer:
    name: prod-web-lb
    region: sfo3
    tag: web-prod
    forwarding_rules:
      - entry_protocol: https
        entry_port: 443
        target_protocol: http
        target_port: 80
    healthcheck:
      protocol: http
      port: 80
      path: /healthz
```

**What it solves**: Idempotency (running the playbook multiple times converges to the desired state), structured data (no comma-separated strings), and integration with server configuration management.

**What it doesn't solve**: State management at scale. Ansible's state tracking is file-based and limited. Dependency graphing (e.g., "create VPC before load balancer") must be manually orchestrated. For teams already using Ansible for server configuration, adding load balancer management is natural. For infrastructure-only teams, dedicated IaC tools offer more robust state management.

**When to use it**: When your infrastructure is tightly coupled to server configuration (e.g., Ansible installs Nginx and configures the load balancer in a single playbook). Not ideal for multi-team environments where infrastructure state must be shared and locked.

### Level 3: Production-Grade IaC (Terraform/Pulumi)

**What it is**: Using stateful Infrastructure-as-Code tools to define load balancers as declarative resources with managed state, dependency graphing, and plan/preview workflows.

**Terraform** (HCL-based, industry standard):

```hcl
resource "digitalocean_loadbalancer" "prod_lb" {
  name     = "prod-web-lb"
  region   = "sfo3"
  vpc_uuid = digitalocean_vpc.main.id
  tag      = "web-prod"

  forwarding_rule {
    entry_protocol   = "https"
    entry_port       = 443
    target_protocol  = "http"
    target_port      = 80
    certificate_name = "my-le-cert-name"
  }

  healthcheck {
    protocol = "http"
    port     = 80
    path     = "/healthz"
  }
}
```

**Pulumi** (Code-native, general-purpose languages):

```typescript
const lb = new digitalocean.LoadBalancer("prod-lb", {
  name: "prod-web-lb",
  region: "sfo3",
  vpcUuid: vpc.id,
  tag: "web-prod",
  forwardingRules: [{
    entryProtocol: "https",
    entryPort: 443,
    targetProtocol: "http",
    targetPort: 80,
    certificateName: "my-le-cert-name"
  }],
  healthcheck: {
    protocol: "http",
    port: 80,
    path: "/healthz"
  }
});
```

**What they solve**:
- **Stateful management**: Both track infrastructure state in remote backends (Terraform state files, Pulumi Service), enabling multi-user collaboration with locking.
- **Dependency graphing**: Automatically understand that VPC must exist before load balancer, Droplets before attachment.
- **Plan/preview workflows**: See exactly what will change before applying.
- **Multi-environment support**: Terraform Workspaces and Pulumi Stacks provide identical functionality for managing dev/staging/prod from a single codebase.

**Critical architectural detail**: Pulumi's DigitalOcean provider is a bridge to Terraform's provider. This means **feature parity is guaranteed**—the resource schema, maturity, and capabilities are identical. The only difference is the language used to define resources (HCL vs. TypeScript/Python/Go).

**When to use it**: Production infrastructure. Always. These tools are the gold standard for managing cloud resources at scale.

### Level 4: Kubernetes-Native Orchestration

**Method 1: DOKS Cloud Controller Manager**

For teams running DigitalOcean Kubernetes (DOKS), load balancer management is abstracted entirely. Creating a Kubernetes Service with `type: LoadBalancer` triggers automatic provisioning:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx
  annotations:
    service.beta.kubernetes.io/do-loadbalancer-protocol: "http"
    service.beta.kubernetes.io/do-loadbalancer-healthcheck-path: "/healthz"
    service.beta.kubernetes.io/do-loadbalancer-certificate-name: "my-cert"
spec:
  type: LoadBalancer
  selector:
    app: nginx
  ports:
    - port: 80
      targetPort: 8080
```

The DigitalOcean Cloud Controller Manager reconciles this into a fully configured load balancer. Configuration happens via annotations, not by editing the load balancer directly. This list of annotations (`do-loadbalancer-protocol`, `do-loadbalancer-certificate-name`, `do-loadbalancer-healthcheck-path`) represents DigitalOcean's own 80/20 abstraction—battle-tested for the most common Kubernetes use cases.

**Method 2: Crossplane**

Crossplane takes a different philosophical approach: it extends the Kubernetes API with Custom Resource Definitions (CRDs) that mirror cloud provider resources. Instead of a Service, you create a DigitalOcean-specific `LB` resource:

```yaml
apiVersion: loadbalancer.do.crossplane.io/v1alpha1
kind: LB
metadata:
  name: prod-lb
spec:
  name: prod-web-lb
  region: sfo3
  tag: web-prod
  forwardingRules:
    - entryProtocol: https
      entryPort: 443
```

This is infrastructure *extension*, not integration. For teams managing all cloud infrastructure from Kubernetes (databases, VPCs, DNS), Crossplane provides a unified control plane. For teams using Kubernetes only for application workloads, the DOKS CCM approach is simpler.

## Comparative Analysis: Choosing Your IaC Tool

### Provider Maturity and Feature Coverage

For DigitalOcean Load Balancers, all three major IaC tools—Terraform, Pulumi, and OpenTofu—offer production-ready providers:

- **Terraform**: The `digitalocean/digitalocean` provider is official, mature, and widely adopted. The `digitalocean_loadbalancer` resource covers the full API surface.
- **Pulumi**: Reuses Terraform's provider via bridging, guaranteeing 100% feature parity and maturity. The `digitalocean.LoadBalancer` resource is architecturally identical to Terraform's.
- **OpenTofu**: As a Terraform fork, it maintains complete provider compatibility and can use the same `digitalocean/digitalocean` provider.

**Decision criteria**: You're not choosing based on capabilities—they're equivalent. You're choosing based on philosophy:

| Aspect | Terraform/OpenTofu | Pulumi |
|--------|-------------------|--------|
| **Language** | HCL (Domain-Specific Language) | TypeScript, Python, Go (General-purpose) |
| **Ecosystem** | Largest provider registry, most community modules | Code-native testing (unittest, pytest), standard tooling |
| **State Management** | Manual remote backend setup (DO Spaces, S3) | Managed Pulumi Service by default, or self-hosted |
| **Learning Curve** | Learning HCL syntax | Using familiar programming languages |
| **Licensing** | OpenTofu: MPL 2.0, Terraform: BSL | Apache 2.0 |

### Resource Schema: The 80/20 Design Pattern

Both Terraform and Pulumi model load balancers using a nested block pattern that reflects production best practices:

**Top-level attributes**: `name`, `region`, `vpc_uuid`, `tag`

**Nested blocks** (Terraform) / **Input objects** (Pulumi):
- `forwarding_rule`: List of port/protocol mappings, each with optional `certificate_name` for SSL termination
- `healthcheck`: Single configuration object defining protocol, port, path, and timing parameters
- `sticky_sessions`: Optional session affinity configuration

**Critical API design insight**: Use `certificate_name`, not `certificate_id`. Let's Encrypt certificate IDs change on renewal, breaking IaC state. Certificate names remain stable. DigitalOcean's DOKS documentation explicitly recommends this pattern.

### Multi-Environment Strategies

Managing dev/staging/prod environments is a solved problem in both tools:

**Terraform Workspaces**: A single codebase deploys to multiple state files. Environment-specific configuration uses `.tfvars` files:

```bash
terraform workspace select prod
terraform apply -var-file=prod.tfvars
```

**Pulumi Stacks**: Identical concept, different name. Environment-specific config uses YAML files:

```bash
pulumi stack select prod
pulumi up
```

Both support remote state (Terraform State in DO Spaces, Pulumi Service) with locking for multi-user collaboration. The choice is semantic, not functional.

## Production Best Practices

### Traffic Routing: Forwarding Rules That Scale

Forwarding rules are the heart of load balancer configuration. Key patterns:

**HTTP-to-HTTPS redirect**: Use the `redirect_http_to_https` flag (Terraform/Pulumi) or annotation (DOKS). This automatically redirects port 80 traffic to 443, simplifying backend application logic—your Droplets can listen on HTTP only.

**SSL Termination**: The most common pattern. Load balancer decrypts HTTPS (port 443) and forwards HTTP (port 80) to backends over the private VPC network. Centralizes certificate management and offloads CPU-intensive decryption from application servers.

**TCP Passthrough**: For non-HTTP services (databases, message queues), use Layer 4 TCP forwarding:
- Entry: TCP 3306 → Target: TCP 3306 (MySQL)
- Entry: TCP 5432 → Target: TCP 5432 (PostgreSQL)

**Rule limits**: A single Regional Load Balancer supports up to 40 forwarding rules. For cost optimization, use one load balancer with multiple rules rather than deploying multiple load balancers.

### Health Checks: The HA Foundation

Health checks are the mechanism that enables high availability. A production-ready health check configuration is non-negotiable:

**HTTP/HTTPS checks**: The load balancer sends a GET request to a specific path (e.g., `/healthz` or `/health`). The backend must return a 2xx status code. This verifies the application is responding, not just the server.

**TCP checks**: Opens a TCP connection on the specified port. Confirms the service is listening but doesn't verify application health. Use for databases or when HTTP endpoints aren't available.

**The UDP gotcha**: Load balancers cannot health-check UDP ports (connectionless protocol). You must configure a TCP or HTTP check on a different port (e.g., check SSH port 22 to verify the Droplet is online).

**Production tuning**:
```hcl
healthcheck {
  protocol                 = "http"
  port                     = 80
  path                     = "/healthz"
  check_interval_seconds   = 10   # How often to check
  response_timeout_seconds = 5    # How long to wait for response
  unhealthy_threshold      = 3    # Failures before marking DOWN
  healthy_threshold        = 5    # Successes before marking UP
}
```

**Common failure mode**: 503 Service Unavailable errors from the load balancer mean zero healthy backends. This is almost always caused by:
1. Misconfigured health checks (wrong port, wrong path)
2. Backend application not responding to health check requests
3. All Droplets actually being down

### Scalability: Tag-Based vs. ID-Based Targeting

Load balancers can target backends using either static Droplet IDs or dynamic tags:

**ID-based targeting**: Manually specify Droplet IDs. Brittle and suitable only for static, manually-managed setups. Every new Droplet requires updating the load balancer configuration.

**Tag-based targeting**: Specify a tag (e.g., `web-prod`). The load balancer automatically discovers all Droplets with that tag and adjusts its backend pool as they're created or destroyed.

**Tag-based is the production pattern** because it enables two critical workflows:

1. **Autoscaling**: Autoscaling services create/destroy Droplets with the correct tag. Load balancer automatically handles pool management.
2. **Blue-Green Deployments**:
   - Production LB targets tag `blue` (current version)
   - Deploy new version to Droplets tagged `green`
   - Test `green` deployment independently
   - Cutover: Single API call switches LB to target `green`
   - Instant traffic shift, instant rollback capability

### Security: VPC Isolation and SSL Management

**VPC best practice**:
1. Create a VPC
2. Place load balancer and all backend Droplets in the VPC
3. Use DigitalOcean Cloud Firewalls to **block all public inbound traffic to Droplets**
4. Result: The only path to your application is Internet → LB → Private VPC → Droplet

This architecture prevents attackers from discovering and attacking Droplets directly, bypassing load balancer protections and any Web Application Firewall.

**SSL/TLS management**:

**Let's Encrypt** (recommended): Free, automated certificate provisioning and renewal. Requires DigitalOcean-managed DNS. Set `certificate_name` in forwarding rule, load balancer handles the rest.

**Custom certificates**: If using third-party DNS (Cloudflare, Route53) or internal CAs, upload certificates via API. You're responsible for renewal monitoring and updates.

**PROXY Protocol pitfall**: Enabling PROXY protocol on the load balancer (to preserve client IP) without configuring backends to expect it breaks SSL handshakes. PROXY protocol headers appear before TLS ClientHello, confusing SSL parsers. **Only enable PROXY protocol if backends (e.g., Nginx) are configured for it**. For most use cases, SSL termination at the load balancer is simpler.

### Monitoring and Observability

**Built-in metrics**: DigitalOcean provides frontend/backend connection counts, throughput, and HTTP request rates in the Control Panel. These are load balancer-level metrics.

**Droplet metrics**: Install `digitalocean-agent` on Droplets to get CPU, memory, disk, and load metrics.

**Alerting gap**: DigitalOcean Monitoring supports alerts based on Droplet metrics (CPU > 80%), but not load balancer metrics (5xx error rate > 2%). For advanced alerting, integrate external monitoring (Datadog, Prometheus with `digitalocean_exporter`) to scrape API metrics and configure sophisticated alert rules.

### Common Anti-Patterns to Avoid

❌ **Single-Droplet backend**: Using a load balancer for one Droplet adds cost and latency for zero HA benefit.

❌ **Missing health checks**: Deploying without health checks means the load balancer can't detect backend failures, defeating its purpose.

❌ **Public Droplet exposure**: Leaving Droplets accessible from the internet bypasses load balancer security and monitoring.

❌ **PROXY Protocol mismatch**: Enabling PROXY protocol on load balancer but not on backend servers breaks SSL and causes connection failures.

❌ **Certificate ID in IaC**: Using `certificate_id` instead of `certificate_name`. Let's Encrypt renewals change IDs, breaking Terraform/Pulumi state.

## The 80/20 Configuration Philosophy

### What Most Teams Actually Need

Analysis of Terraform provider schemas, DOKS annotations, and production configurations reveals a clear 80/20 split:

**Essential (Required 80% of the time)**:
- Name, region, VPC reference
- At least one forwarding rule (entry/target ports and protocols)
- Backend targeting (either Droplet IDs or tag)
- Health check configuration (protocol, port, path)

**Common (Used 15% of the time)**:
- SSL certificate name (for HTTPS termination)
- HTTP-to-HTTPS redirect flag
- Sticky sessions
- Health check timing parameters

**Rare (Advanced 5%)**:
- PROXY protocol settings
- Algorithm selection (round-robin vs. least-connections)
- Size tuning (node count)

### Real-World Configuration Examples

**Example 1: Simple HTTP Load Balancer (Dev/Test)**

```hcl
resource "digitalocean_loadbalancer" "dev_lb" {
  name   = "dev-lb"
  region = "nyc3"
  tag    = "web-dev"

  forwarding_rule {
    entry_protocol  = "http"
    entry_port      = 80
    target_protocol = "http"
    target_port     = 80
  }

  healthcheck {
    protocol = "tcp"
    port     = 80
  }
}
```

**Example 2: Production HTTPS Web Service**

```hcl
resource "digitalocean_loadbalancer" "prod_lb" {
  name                   = "prod-web-lb"
  region                 = "sfo3"
  vpc_uuid               = var.vpc_id
  tag                    = "web-prod"
  redirect_http_to_https = true

  forwarding_rule {
    entry_protocol   = "https"
    entry_port       = 443
    target_protocol  = "http"
    target_port      = 80
    certificate_name = "my-le-cert-name"
  }

  healthcheck {
    protocol                 = "http"
    port                     = 80
    path                     = "/healthz"
    check_interval_seconds   = 10
    response_timeout_seconds = 5
    unhealthy_threshold      = 3
    healthy_threshold        = 5
  }
}
```

**Example 3: TCP Database Load Balancer**

```hcl
resource "digitalocean_loadbalancer" "db_lb" {
  name     = "prod-galera-lb"
  region   = "fra1"
  vpc_uuid = var.vpc_id
  tag      = "db-prod"

  forwarding_rule {
    entry_protocol  = "tcp"
    entry_port      = 3306
    target_protocol = "tcp"
    target_port     = 3306
  }

  healthcheck {
    protocol = "tcp"
    port     = 3306
  }
}
```

## The OpenMCF Choice: Protobuf + Pulumi

### Why Protobuf-Defined APIs

OpenMCF models DigitalOcean Load Balancers as protobuf messages (`DigitalOceanLoadBalancerSpec`) for several strategic reasons:

1. **Language-agnostic schema**: Protobuf generates client libraries for every language (Go, TypeScript, Python, Java), enabling multi-language tooling ecosystems.

2. **Built-in validation**: Buf validate constraints (`min_len`, `required`, `pattern`) enforce correctness at the API layer, preventing invalid configurations from reaching deployment.

3. **Versioned evolution**: Protobuf's backward compatibility guarantees mean API v1 clients continue working as the schema evolves.

4. **80/20 forcing function**: Protobuf's explicit field definitions force API designers to make conscious decisions about what to include. OpenMCF's schema includes essential and common fields while omitting rarely-used parameters (algorithm, PROXY protocol, advanced timeouts).

### Why Pulumi for Deployment

While Terraform is the "800-pound gorilla" of IaC, OpenMCF uses Pulumi as the deployment engine:

**Programmatic generation**: OpenMCF generates Pulumi programs from protobuf specs. Pulumi's code-native approach (TypeScript, Go, Python) makes programmatic generation significantly simpler than HCL generation. The Pulumi program is the intermediate representation, not the user-facing configuration.

**State management parity**: Pulumi's state management is architecturally identical to Terraform (remote backends, locking, plan/preview workflows). There's no loss of production-readiness.

**Provider bridging advantage**: Pulumi reuses Terraform providers, inheriting their maturity and feature coverage. This means using Pulumi provides access to the entire Terraform ecosystem while maintaining code-native flexibility.

**Testing and validation**: Pulumi programs are executable code in general-purpose languages, enabling standard testing frameworks (unittest, pytest, Go testing) to validate infrastructure logic before deployment.

### The Foreign Key Pattern

OpenMCF's protobuf schema uses `StringValueOrRef` for the `vpc` field:

```protobuf
org.openmcf.shared.foreignkey.v1.StringValueOrRef vpc = 3 [
  (buf.validate.field).required = true,
  (org.openmcf.shared.foreignkey.v1.default_kind) = DigitalOceanVpc,
  (org.openmcf.shared.foreignkey.v1.default_kind_field_path) = "status.outputs.vpc_id"
];
```

This enables **resource references**, not just literal values. Users can specify:
- A literal VPC ID: `vpc: { value: "abc-123" }`
- A reference to another resource: `vpc: { ref: "my-vpc" }`

The deployment engine resolves references, extracting `status.outputs.vpc_id` from the referenced `DigitalOceanVpc` resource. This creates a dependency graph and ensures resources are created in the correct order (VPC before load balancer).

The same pattern applies to `droplet_ids`, enabling references to `DigitalOceanDroplet` resources.

### What We Exclude and Why

OpenMCF's API intentionally excludes several DigitalOcean Load Balancer features:

**`redirect_http_to_https`**: While useful, this is syntactic sugar. Users can achieve the same result with two forwarding rules (one HTTP:80→HTTP:80, one HTTPS:443→HTTP:80). The 80/20 principle favors explicit over convenient.

**`enable_proxy_protocol`**: Advanced feature with a high failure rate (misconfiguration breaks SSL). Teams needing this can use Pulumi/Terraform directly.

**`algorithm` (round-robin vs. least-connections)**: Terraform deprecated this as a top-level field. The default round-robin is sufficient for 99% of use cases.

**Health check advanced tuning**: `check_interval_sec` is exposed with a recommended default (10 seconds). Other timing parameters (`response_timeout_seconds`, `unhealthy_threshold`, `healthy_threshold`) are omitted to prevent over-configuration.

**Sticky sessions detail**: Exposed as a boolean (`enable_sticky_sessions`) rather than a configuration object. For advanced sticky session configuration (cookie names, TTL), teams use Pulumi/Terraform.

This isn't about limiting capabilities—it's about presenting the 20% of configuration that 80% of users need, making the default path the correct path.

## Cost Considerations

### Pricing Model

**Regional Load Balancers**:
- Layer 7 (HTTP/HTTPS): **$12/month per node**
- Layer 4 (TCP/UDP): **$15/month per node**
- Includes free Let's Encrypt certificates

**Sizing**: Each node provides:
- 10,000 requests/second
- 10,000 simultaneous connections
- 250 new SSL connections/second

Default is 1 node. Scale horizontally (add nodes) only when monitoring shows capacity limits are being reached.

### Optimization Strategies

**Use VPC for unmetered traffic**: Load balancer → Droplet communication over VPC is unmetered and doesn't count against bandwidth quotas. This is both a security and cost optimization.

**Consolidate with forwarding rules**: A single load balancer supports up to 40 forwarding rules. Use one LB with multiple rules for different services (e.g., `api.example.com` on port 8080, `www.example.com` on port 80) rather than deploying separate LBs.

**When to use multiple load balancers**:
- Different regions (regional isolation)
- Different VPCs (network segmentation)
- Different backend pools with zero overlap
- Separating L4 and L7 workloads (different pricing tiers)

**Let's Encrypt integration**: Free automated certificates eliminate recurring CA costs. Requires DigitalOcean-managed DNS.

## Conclusion

The evolution of load balancer deployment mirrors the maturation of infrastructure practices: from manual configuration to imperative scripts to declarative, stateful IaC. DigitalOcean Load Balancers, built on HAProxy but delivered as a managed service, represent the pragmatic middle ground between operational simplicity and production capability.

For teams practicing infrastructure-as-code, the choice isn't whether to use IaC—it's which tool aligns with your philosophy. Terraform's HCL dominance, Pulumi's code-native approach, and OpenTofu's open-source fork all offer production-ready DigitalOcean providers with feature parity. The decision is about language preference and ecosystem, not capability.

OpenMCF's approach—protobuf-defined APIs generating Pulumi deployments—optimizes for a different dimension: **API clarity over configuration completeness**. By exposing the essential 20% of load balancer configuration through a typed, validated schema, we make the correct path the easy path. Teams needing the advanced 5% can drop to Pulumi/Terraform directly, but most teams ship faster by starting with the 80/20 defaults.

The paradigm shift isn't about load balancer features—it's about treating infrastructure configuration as a versioned, typed, validated API rather than an ad-hoc collection of key-value pairs. That's the foundation for building production infrastructure at scale.


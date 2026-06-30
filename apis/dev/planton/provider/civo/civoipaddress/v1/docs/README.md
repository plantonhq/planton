# Civo Reserved IP Addresses: Deployment Methods & Best Practices

## Introduction

In cloud infrastructure, IP addresses come in two flavors: **dynamic** (ephemeral, assigned at resource creation) and **static** (persistent, survive resource lifecycle changes). The conventional wisdom used to be simple: dynamic IPs are free and convenient, static IPs cost extra and require manual management. For many years, this meant most services lived behind load balancers or used DNS tricks to handle IP changes during deployments.

Civo Cloud challenges this model. On Civo, **reserved IPs are free when attached** to resources—no monthly charges like AWS (\$3.60/mo), GCP (\$8.76/mo), or Azure (\$13.14/mo). This pricing shift changes the strategic calculus: there's little reason *not* to use a reserved IP for any service that needs network stability.

A **Civo Reserved IP** is a persistent public IPv4 address you can reserve in a specific region and attach to instances or load balancers. Unlike the default dynamic IPs that disappear when you delete a resource, reserved IPs persist in your account until you explicitly delete them. This makes them ideal for:

- **Persistent public endpoints**: Web services where DNS records need stable targets
- **Static load balancer IPs**: Kubernetes Services of type LoadBalancer with predictable addresses  
- **High-availability failover**: Moving an IP between primary and standby instances during outages
- **Compliance requirements**: Environments where firewall rules or audit logs require fixed IP addresses

Reserved IPs are **region-scoped** (a London IP can't attach to a New York instance) and subject to account quotas. This document explores how to provision and manage Civo Reserved IPs across the deployment maturity spectrum—from manual console clicks to production-grade Infrastructure as Code.

---

## The Deployment Maturity Spectrum

### Level 0: Manual Console Management

The Civo Dashboard provides a **Networking → Reserved IPs** section where you can click "Create a reserved IP," provide a descriptive label, and obtain a static IP in seconds. To attach it to an instance, you navigate to the instance details page and use the "Assign Reserved IP" action.

**What it solves**: Immediate access for testing, learning Civo's IP model, or one-off scenarios.

**What it doesn't solve**:  
- No audit trail or version control of infrastructure changes  
- Easy to forget which IPs are attached where (especially in multi-team accounts)  
- Manual cleanup required—orphaned IPs count against quota and may incur charges  
- Cross-region deployments require switching regions in the UI repeatedly

**Verdict**: Acceptable for sandbox environments or initial exploration. For anything beyond a single development instance, the lack of reproducibility becomes a liability.

---

### Level 1: CLI Automation

The **Civo CLI** (`civo` command) exposes IP management under the `civo ip` subcommand:

```bash
# Reserve a new IP with a descriptive label
civo ip reserve -n prod-web-ip

# List all reserved IPs in the current region
civo ip ls

# Attach an IP to a running instance
civo ip assign 74.220.24.88 --instance my-vm

# Detach (unassign) an IP
civo ip unassign 74.220.24.88

# Delete a reserved IP (must be unattached first)
civo ip delete prod-web-ip
```

The CLI authenticates via API keys (`civo apikey save`) and respects a default region (`civo region use LON1`), making it scriptable for basic automation or CI pipelines.

**What it solves**:  
- Repeatable commands that can be version-controlled in shell scripts  
- Faster than clicking through the console for bulk operations  
- Works well for teams comfortable with imperative workflows

**What it doesn't solve**:  
- Still imperative, not declarative—no desired-state management  
- State drift (did the script run? did someone manually detach the IP?)  
- No dependency tracking (you must manually ensure instances exist before attaching IPs)  
- Limited multi-environment support (dev/staging/prod requires separate scripts or careful variable management)

**Verdict**: A step up from manual for teams with simple needs or short-lived environments. For production systems, the lack of state management and dependency orchestration makes it fragile.

---

### Level 2: API & Configuration Management

Civo's **REST API** (`https://api.civo.com/v2/`) supports full CRUD operations on reserved IPs. You authenticate with an API key via header (`Authorization: bearer <API_KEY>`) and can integrate IP management into any language or tool:

```bash
# Reserve an IP via API
curl -H "Authorization: bearer <API_KEY>" -X POST \
  https://api.civo.com/v2/ips \
  -d region=LON1 -d name=prod-web-ip
```

Civo provides official **Go** (`civogo`) and **Ruby** SDKs, plus community libraries for Python and other languages. This enables custom automation or integration with existing infrastructure management systems.

For **Ansible** users (despite no official Civo modules), you can use the `uri` module to call the API directly or the `command` module to invoke the Civo CLI. Example pattern:

```yaml
- name: Reserve a Civo IP via API
  uri:
    url: https://api.civo.com/v2/ips
    method: POST
    headers:
      Authorization: "bearer {{ civo_api_key }}"
    body: "region={{ region }}&name={{ ip_name }}"
    body_format: form-urlencoded
  register: ip_result
```

**What it solves**:  
- Flexibility to integrate with any tool or language  
- Programmatic control for complex workflows (e.g., multi-cloud orchestration)  
- Works when Terraform/Pulumi aren't options (restrictive environments)

**What it doesn't solve**:  
- Requires custom state management (you must track which IPs exist and where they're attached)  
- No built-in idempotency—running the same API call twice creates duplicate IPs  
- Error handling, retry logic, and dependency graphs are your responsibility  
- Ansible playbooks become complex quickly without proper abstractions

**Verdict**: Powerful for specialized use cases or integration with non-IaC tooling. But for general infrastructure management, declarative IaC tools provide better guarantees.

---

### Level 3: Infrastructure as Code (Production Standard)

This is where reserved IPs become **first-class infrastructure** managed alongside instances, clusters, networks, and DNS. Two mature options exist: **Terraform** and **Pulumi**, both backed by Civo's official provider.

#### Terraform

Civo maintains an official Terraform provider (`civo/civo`, version 1.x) with full support for reserved IPs via the `civo_reserved_ip` resource:

```hcl
resource "civo_reserved_ip" "www" {
  name   = "nginx-www"
  region = "LON1"
}

resource "civo_instance" "web" {
  hostname      = "web-server"
  size          = "g4s.small"
  reserved_ipv4 = civo_reserved_ip.www.ip  # Attach at creation
}
```

The provider handles dependency ordering automatically—Terraform knows to create the IP before the instance. If you destroy the instance but keep the IP resource in your config, the IP persists (as desired for static addresses).

**Authentication**: Set the `CIVO_TOKEN` environment variable (the provider auto-detects it) or specify `token` in the provider block. For multi-region deployments, you can either:
- Use separate Terraform workspaces per region  
- Instantiate the provider with aliases for each region  
- Set `region` explicitly on each resource

**Attachment patterns**:
- **Inline**: Specify `reserved_ipv4` on the `civo_instance` resource (most common)  
- **Separate resource**: Use `civo_instance_reserved_ip_assignment` for independent lifecycle management (e.g., swapping IPs between instances)

For **Kubernetes LoadBalancer Services**, reserve the IP in Terraform, then annotate the Service manifest:

```yaml
metadata:
  annotations:
    kubernetes.civo.com/ipv4-address: "74.220.24.88"
```

Civo's cloud controller will claim that IP when creating the load balancer.

#### Pulumi

Pulumi's Civo provider (`@pulumi/civo` for Node.js, `pulumi-civo` for Python, etc.) is a **bridge to the Terraform provider**, meaning resource coverage and behavior are identical. Example in TypeScript:

```typescript
const ip = new civo.ReservedIp("www", { name: "web-ip" });

const vm = new civo.Instance("webserver", {
  size: "g4s.small",
  region: "LON1",
  reservedIpv4: ip.ip,  // Attach the reserved IP
});
```

**Authentication**: Same as Terraform—`CIVO_TOKEN` environment variable or Pulumi config (`pulumi config set civo:token <key> --secret`). Pulumi also auto-detects credentials from the Civo CLI's `~/.civo.json` for local development convenience.

**State management**: Pulumi Service or self-managed backends (S3, Azure Blob, etc.) handle state. Like Terraform, destroying an instance doesn't delete a reserved IP unless you remove the IP resource from code.

#### Crossplane (Limited Support)

The community-driven **Crossplane provider for Civo** (v0.1) supports `CivoKubernetes` and `CivoInstance` CRDs but **does not yet include** a `CivoReservedIP` custom resource. For now, Crossplane users must manage IPs separately via Terraform/Pulumi or the API.

**What Level 3 solves**:  
- **Declarative state**: Describe desired infrastructure; IaC tools converge to it  
- **Dependency orchestration**: Automatic ordering (create IP → create instance → attach)  
- **Change preview**: `terraform plan` or `pulumi preview` shows exactly what will happen  
- **GitOps-friendly**: Infrastructure changes go through code review, CI/CD, and audit logs  
- **Multi-environment**: Reuse modules/stacks with different configs for dev/staging/prod  
- **Idempotency**: Re-running applies is safe—no duplicate resources

**What it doesn't solve**:  
- Learning curve for teams new to IaC  
- State file management complexity (requires remote backends for teams)  
- Initial setup overhead (provider configuration, backend setup)

**Verdict**: This is the **production standard** for any team managing more than a handful of resources or operating in regulated environments. The benefits of auditability, repeatability, and disaster recovery vastly outweigh the initial investment.

---

## IaC Tool Comparison: Terraform vs Pulumi

Both tools manage Civo reserved IPs identically under the hood (Pulumi uses the Terraform provider code), so the decision comes down to **workflow preference**, not technical capability.

| Aspect | Terraform | Pulumi |
|--------|-----------|--------|
| **Language** | HCL (declarative DSL) | TypeScript, Python, Go, C#, Java (real programming languages) |
| **State Management** | Terraform state file (local or remote backend) | Pulumi Service or self-managed backend |
| **Reserved IP Support** | `civo_reserved_ip` resource (v1.0+) | `civo.ReservedIp` (bridged from Terraform provider) |
| **Attachment** | `reserved_ipv4` field on instance | `reservedIpv4` property on instance |
| **Community/Docs** | Larger Terraform community, more examples | Growing Pulumi community, fewer Civo-specific examples |
| **Best For** | Teams preferring declarative config files, GitOps workflows | Teams wanting programmatic control (loops, conditionals, abstraction) |
| **Production Readiness** | Mature, widely adopted | Stable, officially supported by Civo |

**Key considerations**:

- **State isolation**: Destroying an instance via IaC does **not** destroy the reserved IP (it's a separate resource). To fully clean up, remove both from config.  
- **IP reassignment**: If you swap which IP is attached to an instance in code (change `reserved_ipv4` to point to a different IP), the IaC tool may trigger an instance replacement (destroy + recreate) since IP changes often can't be hot-swapped on running VMs. For zero-downtime IP swaps, use the CLI or API to move the IP, then update IaC to match.  
- **Multi-region**: Both tools support multiple regions via provider aliases or separate stacks. For reserved IPs (which are region-scoped), it's often cleaner to have one stack per region rather than mixing regions in a single configuration.

**Recommendation**: If your team already uses Terraform or Pulumi for other clouds, extend it to Civo. If starting fresh, Terraform has a shallower learning curve and more community resources, but Pulumi offers more power for teams with strong programming backgrounds.

---

## Production Best Practices

### 1. Treat IPs as Managed Resources

Don't create reserved IPs ad-hoc in the console and forget about them. Define them in IaC alongside the resources they serve. Use descriptive labels (`prod-web-ip`, `frontend-vip`) to make their purpose obvious in the Civo dashboard and audit logs.

**Anti-pattern**: Allocating 10 IPs "just in case" and leaving them unattached. Civo doesn't charge for attached IPs, but unattached IPs may incur charges and definitely count against your quota. Regularly audit with `civo ip ls` and delete unused IPs.

### 2. Plan Attachment Lifecycle

Reserve the IP **before** creating the resource that will use it. This allows you to configure DNS records pointing to the IP even before the service is live (useful for blue/green deployments).

For **high-availability failover**, leverage Civo's instant IP reassignment: bring up a standby instance, then use `civo ip assign <ip> --instance backup` to move the IP from the failed primary. Clients experience a brief reconnection blip, but the IP remains stable.

**Note**: If an IP is already attached to instance A and you assign it to instance B via CLI or IaC, it will **immediately detach from A and attach to B**. This "hot swap" is powerful but dangerous—ensure you're not accidentally stealing an IP from a running production service.

### 3. Integrate with DNS

**Never hardcode IP addresses** in application configs or client code. Always use DNS names that point to reserved IPs. This gives you flexibility to change the underlying IP (in rare disaster scenarios) without redeploying applications.

For Civo DNS, use Terraform's `civo_dns_domain_record` to create A records pointing to `civo_reserved_ip.www.ip`. For external DNS (Cloudflare, Route53), use their respective providers or ExternalDNS for Kubernetes.

### 4. Automation Workflow

Recommended pipeline for new services:

1. **Reserve IP** (via IaC or pre-allocated in a pool)  
2. **Create instance/cluster** with IP attached  
3. **Configure DNS** to point to the reserved IP  
4. **Deploy application** and verify connectivity via DNS name

During **teardown**:

1. **Remove DNS record** (prevent clients from attempting connections)  
2. **Detach IP** from resource (if keeping the IP for reuse)  
3. **Delete resource**  
4. **Delete IP** (if no longer needed)

This ordering prevents dangling DNS and ensures IPs are explicitly released.

### 5. Cost Optimization

Civo's free-when-attached model means the cost concern is **quota exhaustion**, not billing. Delete IPs you're not using. For ephemeral environments (CI/CD test stacks), skip reserved IPs entirely—use dynamic IPs and clean up the whole stack after tests.

For **production services**, the small risk of an unattached IP fee (if Civo's policy changes) is negligible compared to the operational cost of IP changes causing downtime.

---

## The Planton Choice

Planton's `CivoIpAddress` resource follows the **80/20 principle**: focus on the 20% of configuration that 80% of users need.

### Minimal Spec

The protobuf spec for `CivoIpAddressSpec` includes just two fields:

- **`region`** (required): The Civo region where the IP will be reserved (e.g., `LON1`, `NYC1`, `FRA1`). Must match the region of resources you'll attach it to.  
- **`description`** (optional): A human-readable label for the IP (e.g., "production web IP", "corp VPN IP"). Appears in the Civo dashboard and helps with auditing.

No other configuration is needed—Civo allocates the actual IP address automatically, and attachment to instances or load balancers is handled by the instance/LB resource referencing the IP.

### Why This Simplicity Works

Civo's reserved IP model is inherently simple: an IP is just a regional, persistent IPv4 address. It has no tags, no network associations, no advanced routing options. By keeping the API minimal, we:

- **Reduce user error**: Fewer fields = fewer ways to misconfigure  
- **Improve clarity**: The spec clearly shows the two decisions users must make (where and what to call it)  
- **Maintain flexibility**: Attachment patterns vary (inline vs. separate), so we don't enforce one in the IP resource itself

### Default Implementation: Pulumi

Planton uses **Pulumi (Go)** as the default IaC engine for Civo resources. The `CivoIpAddress` module:

1. Calls `civo.NewReservedIp()` with the region and description from the spec  
2. Outputs the allocated IP address in `stack_outputs.proto` as `ipAddress` (for use in DNS or other resources)  
3. Tracks the resource ID for lifecycle management (updates, deletions)

This choice aligns with Planton's strategy: use production-proven, open-source IaC tools (Pulumi bridges to Civo's official Terraform provider) while providing a consistent, multi-cloud API layer via protobuf specs.

### When to Use This Resource

Use `CivoIpAddress` when:

- **Stability matters**: Web services, VPNs, or APIs where IP changes break client configurations  
- **Compliance requires it**: Firewall allowlists or audit logs tied to specific IPs  
- **HA/failover**: You need to move an IP between instances during outages

Skip it when:

- **Short-lived environments**: CI/CD test stacks that live for minutes  
- **Behind a load balancer with dynamic IP**: If your service is already behind a managed LB and DNS handles changes, dynamic IPs are fine

---

## Deep Dive References

For teams needing more detail on specific integration patterns, consult these guides (to be published separately):

- **[Kubernetes LoadBalancer Integration](./kubernetes-loadbalancer-guide.md)**: How to use `kubernetes.civo.com/ipv4-address` annotations with Civo Cloud Controller Manager  
- **[High-Availability Failover Patterns](./ha-failover-guide.md)**: Scripting IP reassignment for multi-instance HA setups  
- **[DNS Automation with Civo Reserved IPs](./dns-automation-guide.md)**: Integrating Civo DNS or external providers (Cloudflare, Route53) in IaC workflows

---

## Conclusion

The shift from "dynamic by default, static costs extra" to "static for free" changes how we should think about IP management in Civo Cloud. Reserved IPs aren't a premium feature to be used sparingly—they're a **default best practice** for any service that needs network stability.

The maturity spectrum from manual console to production IaC is short on Civo: you can move from experimentation to fully-automated GitOps in days, not months. The tooling is mature (official Terraform/Pulumi providers), the API is simple (just region and description), and the cost model aligns with best practices (no penalty for using static IPs).

By managing Civo Reserved IPs as code—whether via Terraform, Pulumi, or Planton's protobuf APIs—you gain **auditability** (every IP change is a Git commit), **repeatability** (rebuilding environments is trivial), and **reliability** (no manual steps to forget during 3 AM incidents).

Start with IaC from day one. Your future self, debugging a production outage at midnight, will thank you for having a clear, version-controlled record of which IPs are where and why.


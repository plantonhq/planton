# Securing DigitalOcean Infrastructure: A Production Guide to Cloud Firewalls

## Introduction

In cloud security, there's a deceptively simple principle: **default deny**. Block everything, then carefully poke holes for the services that need them. Yet many production outages and security incidents trace back to a single, common pitfall: the "double firewall" problem.

Picture this: You've just deployed a web application to a DigitalOcean Droplet. You carefully configure a DigitalOcean Cloud Firewall to allow HTTPS traffic on port 443. You test it. Connection times out. You check the firewall rules in the control panel—they look perfect. You scratch your head, check again, and then realize: there's *another* firewall running *inside* the Droplet itself—`ufw` or `iptables`—still blocking that port because no one ever told it to allow traffic.

This is the reality of DigitalOcean's firewall architecture: **two independent layers** that don't communicate with each other. The **Cloud Firewall** operates at the network edge, outside the Droplet, dropping unwanted packets before they even touch your VM. The **host-based firewall** (`ufw`, `iptables`) runs inside the operating system, enforcing its own rules. Traffic must pass *both* to reach your application.

What makes DigitalOcean Cloud Firewalls compelling is that they solve the right problem: **network-edge enforcement with centralized, scalable management**. They're stateful (you only define rules for initiating traffic, not return packets), free (included with the platform), and designed to scale to thousands of Droplets via tag-based targeting. They protect your infrastructure from resource-exhaustion attacks by dropping malicious traffic before it consumes CPU or memory on your VMs. And they integrate natively with Droplets, Kubernetes clusters, and Load Balancers, enabling production-grade, multi-tier security architectures.

But simplicity at the API level doesn't mean simple operations. You need to understand the deployment spectrum—from manual "click-ops" (an anti-pattern for production) to declarative Infrastructure-as-Code with Terraform, Pulumi, or Crossplane. You need to know when to use static Droplet IDs (almost never) versus tag-based targeting (the production standard). And you need a coherent strategy for organizing rules, managing multi-environment configs, and troubleshooting connectivity issues when both firewalls are in play.

This guide walks you through the landscape of DigitalOcean Cloud Firewall deployment methods, explains the production patterns that separate brittle configurations from scalable ones, and shows how Planton abstracts these choices into a clean, protobuf-defined API.

---

## The Deployment Spectrum: From Manual to Production

Not all methods of managing firewalls are created equal. Here's how they stack up, from what to avoid to what works at scale:

### Level 0: The Control Panel (Anti-Pattern for Production)

**What it is:** Using the DigitalOcean web dashboard to click through firewall creation, rule configuration, and Droplet assignment.

**What it solves:** Discovery and learning. The control panel is fine for understanding how firewalls work, testing a rule quickly, or emergency troubleshooting when you need to see the current state visually.

**What it doesn't solve:** Repeatability, auditability, version control, or scalability. Manual configurations drift over time. Rules added "just for testing" get forgotten and stay in production. Applying a firewall to 50 Droplets means clicking through 50 checkboxes. If you can't codify it, you can't reliably reproduce it across environments or hand it off to another engineer without a knowledge-transfer meeting.

**The common trap:** Creating overly permissive rules during development (SSH open to `0.0.0.0/0`, for example) and forgetting to lock them down before going live.

**Verdict:** Use it to explore the interface and understand the workflow. Never for production or even staging environments that matter.

---

### Level 1: CLI Scripting (Better, But Still Brittle)

**What it is:** Using the `doctl` CLI to create and manage firewalls in shell scripts:

```bash
doctl compute firewall create \
  --name web-firewall \
  --inbound-rules "protocol:tcp,ports:443,address:0.0.0.0/0" \
  --tag-names web-tier
```

**What it solves:** Automation. You can script provisioning, integrate it into CI/CD, and version-control your scripts. The `doctl` CLI is synchronous, supports JSON output for parsing, and covers all firewall operations: create, update rules, assign tags, delete.

**What it doesn't solve:** State management. Scripts don't track what was created or what changed. If a script runs twice, you might create duplicate firewalls or fail because resources already exist. Cleanup on failure is manual. Rule syntax is passed as dense, quoted strings (e.g., `"protocol:tcp,ports:22,droplet_id:386734086"`), which are error-prone and hard to review.

**A useful pattern:** Shell scripts that dynamically update firewall rules with your current public IP for SSH access—common for developers working from dynamic IP addresses. But even this is better handled by IaC with variables.

**Verdict:** Acceptable for throwaway dev environments or quick one-off operations. Not suitable for production, where you need state tracking, idempotency, and rollback capabilities.

---

### Level 2: Direct API Integration (Flexible but High-Maintenance)

**What it is:** Calling the DigitalOcean v2 REST API directly from custom tooling or configuration management systems (like Ansible's `uri` module or Python scripts using the `pydo` library).

**Example (Python with `pydo`):**

```python
from pydo import Client

client = Client(token=api_token)
firewall = client.firewalls.create(
    name="db-firewall",
    inbound_rules=[{
        "protocol": "tcp",
        "ports": "5432",
        "sources": {"tags": ["web-tier"]}
    }],
    tags=["db-tier"]
)
```

**What it solves:** Maximum flexibility. You can integrate firewall management into any HTTP-capable tool or custom orchestration system. The API is well-documented, supports all advanced features (tag-based sources/destinations, Load Balancer UIDs, Kubernetes cluster IDs), and returns structured JSON.

**What it doesn't solve:** Abstraction. You're managing HTTP calls, handling authentication (API keys in headers), sequencing operations (create firewall before assigning Droplets), and implementing idempotency yourself. There's no built-in state file to tell you what exists or what changed. You're essentially building your own IaC layer.

**Verdict:** Useful if you're building a custom provisioning system or integrating DigitalOcean into a broader orchestration framework. But for most teams, higher-level tools (Terraform, Pulumi) handle the API calls and state management for you.

---

### Level 3: Infrastructure-as-Code (Production-Ready)

**What it is:** Using Terraform, Pulumi, or Ansible with declarative configurations to define firewalls and their lifecycle.

**Terraform example:**

```hcl
provider "digitalocean" {
  token = var.do_token
}

resource "digitalocean_firewall" "web" {
  name = "web-firewall"

  tags = ["web-tier"]

  inbound_rule {
    protocol         = "tcp"
    port_range       = "443"
    source_load_balancer_uids = [digitalocean_loadbalancer.main.id]
  }

  inbound_rule {
    protocol         = "tcp"
    port_range       = "22"
    source_addresses = ["203.0.113.0/24"]  # Office IP
  }

  outbound_rule {
    protocol              = "tcp"
    port_range            = "5432"
    destination_tags      = ["db-tier"]
  }

  outbound_rule {
    protocol              = "tcp"
    port_range            = "443"
    destination_addresses = ["0.0.0.0/0", "::/0"]  # For OS updates, external APIs
  }

  outbound_rule {
    protocol              = "udp"
    port_range            = "53"
    destination_addresses = ["0.0.0.0/0", "::/0"]  # DNS
  }
}
```

**Pulumi example (TypeScript):**

```typescript
import * as digitalocean from "@pulumi/digitalocean";

const webFirewall = new digitalocean.Firewall("web-firewall", {
    name: "web-firewall",
    tags: ["web-tier"],
    inboundRules: [
        {
            protocol: "tcp",
            portRange: "443",
            sourceLoadBalancerUids: [loadBalancer.id],
        },
        {
            protocol: "tcp",
            portRange: "22",
            sourceAddresses: ["203.0.113.0/24"],
        },
    ],
    outboundRules: [
        {
            protocol: "tcp",
            portRange: "5432",
            destinationTags: ["db-tier"],
        },
        {
            protocol: "tcp",
            portRange: "443",
            destinationAddresses: ["0.0.0.0/0", "::/0"],
        },
    ],
});
```

**What it solves:** Everything. You get:
- **Declarative configuration**: State what you want, not how to get there
- **State management**: Terraform/Pulumi track what exists and what changed
- **Idempotency**: Running the same config twice produces the same result
- **Plan/preview**: See changes before applying them (`terraform plan`, `pulumi preview`)
- **Version control**: Treat infrastructure as code, with diffs, reviews, and rollbacks
- **Multi-environment support**: Reuse configs for dev/staging/prod with different parameters
- **Modular organization**: Extract firewall rules into reusable modules

**What it doesn't solve:** The underlying architectural constraints (10-Droplet limit for static IDs, maximum 50 rules per firewall, no native rule priorities). But it makes managing those constraints predictable, reproducible, and auditable.

**Verdict:** This is the production standard. Terraform has the broadest adoption and ecosystem maturity. Pulumi offers programming language flexibility (TypeScript, Python, Go) over HCL. Ansible bridges configuration management and provisioning. All three are solid choices.

---

### Level 4: Kubernetes-Native Control Planes (Emerging)

**What it is:** Using Crossplane with the `provider-upjet-digitalocean` to manage DigitalOcean firewalls as Kubernetes Custom Resource Definitions (CRDs).

**Example (Kubernetes YAML):**

```yaml
apiVersion: digitalocean.upjet.crossplane.io/v1alpha1
kind: Firewall
metadata:
  name: web-firewall
spec:
  forProvider:
    name: web-firewall
    tags:
      - web-tier
    inboundRules:
      - protocol: tcp
        portRange: "443"
        sourceLoadBalancerUids:
          - lb-abc123
    outboundRules:
      - protocol: tcp
        portRange: "5432"
        destinationTags:
          - db-tier
```

You apply this to a Kubernetes cluster running Crossplane (`kubectl apply -f firewall.yaml`). A controller observes the CRD and reconciles it with the DigitalOcean API, creating or updating the actual firewall.

**What it solves:** Unified control plane for all infrastructure. If your organization already manages application workloads in Kubernetes, Crossplane lets you manage cloud resources (firewalls, Droplets, databases) using the same `kubectl` workflows, RBAC, and GitOps pipelines.

**What it doesn't solve:** Complexity. Crossplane adds operational overhead (running controllers, managing provider packages). The DigitalOcean provider is emerging (the modern `provider-upjet-digitalocean` replaced the older `provider-digitalocean` in 2023) and less battle-tested than Terraform or Pulumi.

**Verdict:** Promising for organizations with heavy Kubernetes investment and GitOps pipelines (Flux, ArgoCD). Not yet the default recommendation for teams just starting with IaC.

---

## IaC Tool Comparison: Terraform, Pulumi, Ansible, Crossplane

All four tools support DigitalOcean Cloud Firewalls in production. Here's how they compare:

| Tool | Resource Name | Maturity | State Management | Multi-Env Pattern | Key Differentiator |
|------|---------------|----------|------------------|-------------------|-------------------|
| **Terraform** | `digitalocean_firewall` | Very High | Local/Remote State File | Workspaces / Directory-based Modules | HCL syntax; industry standard; mature ecosystem |
| **Pulumi** | `digitalocean.Firewall` | High (inherited from Terraform) | Local/SaaS State Backend | Native Stacks | Real programming languages (TypeScript, Python, Go) |
| **Ansible** | `community.digitalocean.digital_ocean_firewall` | High | None by default (asserts state) | `group_vars` / Inventory | Integrates configuration management and provisioning |
| **Crossplane** | `Firewall.digitalocean.upjet.io` | Emerging | State stored in Kubernetes etcd | Kubernetes Namespaces / Labels | Kubernetes-native control plane; GitOps-friendly |

### Terraform: The Battle-Tested Standard

**Maturity:** Terraform has been the IaC standard for years. The DigitalOcean provider is officially maintained by DigitalOcean, stable, and comprehensively documented.

**Configuration Model:** Declarative HCL. You define resources and their relationships, and Terraform figures out the dependency graph and execution order.

**State Management:** Local or remote backends (S3, Terraform Cloud, etc.). State tracks resource IDs, making updates and deletions predictable.

**Strengths:**
- Broad ecosystem and community support (thousands of modules, examples)
- Familiar syntax across ops teams
- Robust `plan` workflow shows exactly what will change before applying
- Excellent support for dynamic blocks (e.g., iterating over a list of ports to generate rules programmatically)

**Limitations:**
- HCL is less expressive than a full programming language (limited loops, conditionals)
- Complex provisioning logic (e.g., "if prod, create stricter rules") can be verbose

**Verdict:** The default choice for teams already using Terraform or prioritizing stability and ecosystem maturity. Perfect for standard firewall provisioning.

---

### Pulumi: The Programmer's IaC

**Maturity:** Newer than Terraform, but production-ready. The Pulumi DigitalOcean provider is auto-generated from the Terraform provider (via a bridge), so it has 1:1 resource parity.

**Configuration Model:** Real programming languages (TypeScript, Python, Go). You write infrastructure logic as code, with loops, conditionals, functions, and unit tests.

**State Management:** Pulumi Cloud (SaaS) or self-managed backends (S3, Azure Blob). Similar state tracking to Terraform.

**Strengths:**
- Full programming language expressiveness (easier to build dynamic configs)
- Better for complex provisioning logic or integration with application code
- Native testing frameworks (write unit tests for your infrastructure)
- Native support for "stacks" (one codebase, multiple environments)

**Limitations:**
- Smaller community than Terraform (fewer third-party modules and examples)
- Requires a runtime (Node.js, Python, etc.)
- Bridged provider means any quirks in Terraform's provider carry over

**Verdict:** Great if your team prefers coding infrastructure in familiar languages or needs complex orchestration logic (dynamic rule generation, multi-cloud abstractions). Slightly more overhead than Terraform for simple use cases.

---

### Ansible: Bridging Configuration and Provisioning

**Maturity:** The `community.digitalocean` collection is well-maintained and widely used.

**Configuration Model:** YAML playbooks with the `digital_ocean_firewall` module.

**State Management:** Ansible is traditionally agentless and doesn't maintain a state file like Terraform/Pulumi. It asserts the infrastructure matches the playbook definition on each run.

**Strengths:**
- Integrates provisioning (create firewall) with configuration management (install `ufw` rules inside Droplets)
- Familiar for teams already using Ansible for server config
- Simple YAML syntax

**Limitations:**
- Lacks robust state-tracking, drift-detection, and planning capabilities of dedicated provisioning tools
- Less suitable for large-scale, multi-tier infrastructure orchestration

**Verdict:** Good for teams heavily invested in Ansible for configuration management. But for pure infrastructure provisioning, Terraform or Pulumi offer more mature state management and preview workflows.

---

### Crossplane: Kubernetes-Native Infrastructure

**Maturity:** Emerging. The modern `provider-upjet-digitalocean` (released 2023) is the official, Upjet-generated provider, replacing the older archived provider.

**Configuration Model:** Kubernetes CRDs (Custom Resource Definitions). Firewalls are defined as YAML manifests and applied via `kubectl`.

**State Management:** State is stored in Kubernetes etcd. Controllers reconcile desired state (CRD) with actual state (DigitalOcean API).

**Strengths:**
- Unified control plane for infrastructure and applications
- GitOps-friendly (Flux, ArgoCD)
- Kubernetes RBAC for infrastructure access control

**Limitations:**
- Operational complexity (requires running Crossplane controllers)
- Less battle-tested than Terraform/Pulumi for DigitalOcean
- Steep learning curve if team isn't already Kubernetes-native

**Verdict:** Promising for organizations with deep Kubernetes investment and GitOps pipelines. Not yet the default for teams starting with IaC.

---

### Which Should You Choose?

- **Default to Terraform** if you want the most mature, widely-adopted solution with straightforward declarative configs.
- **Choose Pulumi** if you prefer writing infrastructure in TypeScript/Python/Go and need advanced logic or testing.
- **Use Ansible** if you're already managing server configs with Ansible and want to unify provisioning and CM.
- **Consider Crossplane** if you're Kubernetes-native, use GitOps, and want a single control plane for everything.

All four work equally well for standard firewall provisioning. The choice is more about team preference and existing tooling than capability.

---

## Production Patterns: Security, Scalability, and Sanity

### The Tag-Based Targeting Strategy (Production Essential)

The single most important production pattern for DigitalOcean Cloud Firewalls is **tag-based targeting**.

DigitalOcean lets you apply firewalls to Droplets in two ways:
1. **Static Droplet IDs**: Explicitly list the IDs of the Droplets the firewall should protect (e.g., `droplet_ids = [386734086, 386734087]`)
2. **Tags**: Apply the firewall to all Droplets (existing and future) that have a specific tag (e.g., `tags = ["web-tier"]`)

**Why tags are the production standard:**

- **Scalability**: A firewall can be applied to a maximum of 10 static Droplet IDs, but there's no limit to the number of Droplets that can share a tag. Tags let you exceed the 10-Droplet limit.
- **Automation**: When auto-scaling provisions a new Droplet, it automatically receives the tags defined in the launch template. The firewall applies instantly—no manual intervention required.
- **Composability**: A single Droplet can have multiple tags and therefore multiple firewalls applied. Rules are additive (the Droplet receives the "union" of all rules).
- **Readability**: Tags like `web-tier`, `db-tier`, `prod`, `staging` make configs self-documenting.

**When to use static Droplet IDs:**
- Development or testing environments where Droplets are manually created and destroyed
- One-off, temporary exceptions (e.g., allowing SSH to a specific Droplet for debugging)

**Production anti-pattern:** Hardcoding Droplet IDs in production configs. When auto-scaling provisions a new instance, it won't be protected until you manually update the firewall config with the new ID—creating a security gap.

---

### Rule Organization: Composable Firewalls by Role

Instead of creating one large, monolithic firewall with dozens of rules, adopt a **"split by role"** strategy. Create multiple, focused firewalls and apply them based on function.

**Example: Multi-Tier Web Application**

1. **`management-fw`** (applied to all Droplets via `tag: all-instances`):
   - Inbound: SSH (port 22) from office bastion IP (`203.0.113.10/32`)
   - Outbound: DNS (UDP port 53), HTTPS (TCP port 443) for OS updates

2. **`web-tier-fw`** (applied via `tag: web-tier`):
   - Inbound: HTTPS (port 443) from Load Balancer UID
   - Outbound: PostgreSQL (TCP port 5432) to `tag: db-tier`

3. **`db-tier-fw`** (applied via `tag: db-tier`):
   - Inbound: PostgreSQL (port 5432) from `tag: web-tier` only
   - Outbound: None (or only OS updates to specific repository IPs)

**Why this works:**
- **Separation of concerns**: Management, public traffic, and internal service communication are isolated
- **Reusability**: The `management-fw` can be reused across all environments
- **Least privilege**: The database is completely isolated from the public internet. Only the web tier can reach it.

---

### Security Best Practices

1. **Default Deny**: DigitalOcean Cloud Firewalls block all traffic by default. Only add explicit "allow" rules. This is the foundation of least-privilege security.

2. **Never expose management ports to the internet**: SSH (port 22), RDP (port 3389), database ports (5432, 3306) should *never* have `source_addresses: ["0.0.0.0/0"]` in production. Lock them to office IPs, bastion hosts, or VPN exit points.

3. **Use Load Balancer UIDs for public services**: Instead of allowing HTTPS from `0.0.0.0/0` directly to web servers, allow it only from the Load Balancer UID. This forces all public traffic through the LB, enabling centralized SSL termination, rate limiting, and logging.

4. **Implement outbound rules**: While many configurations default to allowing all outbound traffic (`0.0.0.0/0` on TCP/UDP), high-security environments (like database servers) should implement "default deny" outbound policies, explicitly allowing only necessary traffic (DNS, OS repos, internal services).

5. **Beware the "double firewall" trap**: Traffic must pass both the Cloud Firewall (network edge) and the host-based firewall (`ufw`, `iptables`). The two are independent. Check both when troubleshooting.

---

### Troubleshooting Connectivity Issues

When a connection times out, it's almost always a firewall misconfiguration. Here's the debugging flow:

1. **Check Cloud Firewall**: In the DigitalOcean control panel, navigate to the Droplet's "Networking" tab. This shows a comprehensive table of all rules from all firewalls applied to that Droplet. This is the network-edge truth.

2. **Check Host Firewall**: Use the Web Console to access the Droplet (bypassing SSH). Run:
   ```bash
   sudo ufw status verbose  # If using UFW
   sudo iptables -L -n -v   # If using iptables directly
   ```

3. **Resolve Conflicts**: If a conflicting host rule is found, the simplest fix is to disable the host firewall and rely solely on the Cloud Firewall:
   ```bash
   sudo ufw disable
   ```

4. **Verify Service**: If both firewalls appear correct, check the service is running and bound:
   ```bash
   systemctl status sshd
   ss -ltn  # Show listening TCP ports
   ```

**Pro tip:** Use an external "probe" service to test connectivity. If the probe fails and the host logs show no incoming traffic, the Cloud Firewall is blocking it. If the host logs show connection attempts with drops, the host firewall is the culprit.

---

## Planton's Approach: Abstraction with Production Pragmatism

Planton abstracts DigitalOcean Cloud Firewall provisioning behind a clean, protobuf-defined API (`DigitalOceanFirewall`). This provides a consistent interface across clouds while respecting DigitalOcean's unique characteristics.

### What We Abstract

The `DigitalOceanFirewallSpec` follows the **80/20 principle**: 80% of users need only these fields:

- **`name`**: Human-readable firewall identifier (unique per account)
- **`tags`**: Droplet tags to apply the firewall to (the production-preferred method)
- **`droplet_ids`**: Static Droplet IDs (for dev/testing scenarios)
- **`inbound_rules`**: List of rules allowing traffic *to* Droplets
- **`outbound_rules`**: List of rules allowing traffic *from* Droplets

Each rule supports the full spectrum of DigitalOcean's capabilities:
- **Protocol**: `tcp`, `udp`, or `icmp`
- **Port Range**: Single port (`"22"`), range (`"8000-9000"`), or all (`"1-65535"`)
- **IP-Based Sources/Destinations**: CIDR blocks (`"0.0.0.0/0"`, `"192.168.1.0/24"`)
- **Resource-Based Sources/Destinations**: Droplet IDs, tags, Load Balancer UIDs, Kubernetes cluster IDs

This unifies the simple (allow HTTPS from anywhere) and the advanced (allow PostgreSQL from web-tier tags only) in a single, consistent API.

### Default Choices

- **No default rules**: We enforce explicit configuration. No magic "allow all SSH" rules that could become security holes.
- **Tag-first philosophy**: Examples and documentation emphasize tag-based targeting for production, with Droplet IDs shown as dev/testing alternatives.
- **Outbound flexibility**: We don't assume "allow all outbound." High-security environments can define restrictive outbound policies.

### Under the Hood: Pulumi

Planton currently uses **Pulumi (Go)** for DigitalOcean Firewall provisioning. Why?

- **Language Flexibility**: Pulumi's Go SDK fits naturally into our broader multi-cloud orchestration.
- **Equivalent Coverage**: Pulumi's DigitalOcean provider (bridged from Terraform) supports all firewall features.
- **Future-Proofing**: Pulumi's programming model simplifies conditional logic, multi-environment strategies, and custom integrations.

That said, Terraform would work equally well. The choice is implementation detail—the protobuf API remains the same.

---

## Configuration Examples: Dev, Staging, Production

### Development: Permissive Web Server

**Use Case:** Simple web application for developer testing. Easy access, no strict security.

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanFirewall
metadata:
  name: dev-web-fw
spec:
  name: dev-web-firewall
  tags:
    - dev-web
  inbound_rules:
    - protocol: tcp
      port_range: "22"
      source_addresses:
        - "0.0.0.0/0"
    - protocol: tcp
      port_range: "80"
      source_addresses:
        - "0.0.0.0/0"
    - protocol: tcp
      port_range: "443"
      source_addresses:
        - "0.0.0.0/0"
    - protocol: icmp
      source_addresses:
        - "0.0.0.0/0"
  outbound_rules:
    - protocol: tcp
      port_range: "1-65535"
      destination_addresses:
        - "0.0.0.0/0"
    - protocol: udp
      port_range: "1-65535"
      destination_addresses:
        - "0.0.0.0/0"
```

**Rationale:**
- Open SSH, HTTP, HTTPS for easy testing
- Allow all outbound for flexibility
- Fine for dev; **never use in production**

---

### Production Web Tier: Secure and Scalable

**Use Case:** Production web servers behind a Load Balancer. Serve HTTPS traffic, allow SSH from office only.

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanFirewall
metadata:
  name: prod-web-fw
spec:
  name: prod-web-firewall
  tags:
    - prod-web-tier
  inbound_rules:
    - protocol: tcp
      port_range: "443"
      source_load_balancer_uids:
        - "lb-abc123"  # Load Balancer UID
    - protocol: tcp
      port_range: "80"
      source_load_balancer_uids:
        - "lb-abc123"  # For HTTP-to-HTTPS redirect
    - protocol: tcp
      port_range: "22"
      source_addresses:
        - "203.0.113.10/32"  # Office bastion host
  outbound_rules:
    - protocol: tcp
      port_range: "5432"
      destination_tags:
        - prod-db-tier  # Allow connections to database tier
    - protocol: tcp
      port_range: "443"
      destination_addresses:
        - "0.0.0.0/0"  # For external APIs, OS updates
    - protocol: udp
      port_range: "53"
      destination_addresses:
        - "0.0.0.0/0"  # DNS
```

**Rationale:**
- HTTPS/HTTP only from Load Balancer (not directly from internet)
- SSH locked to office IP
- Outbound to database tier via tags (scalable, dynamic)
- Explicit DNS and HTTPS outbound for operations

---

### Production Database Tier: Internal-Only

**Use Case:** PostgreSQL database reachable only by web tier and administrators. No public access.

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanFirewall
metadata:
  name: prod-db-fw
spec:
  name: prod-db-firewall
  tags:
    - prod-db-tier
  inbound_rules:
    - protocol: tcp
      port_range: "5432"
      source_tags:
        - prod-web-tier  # Only web tier can connect
    - protocol: tcp
      port_range: "22"
      source_addresses:
        - "203.0.113.10/32"  # Office bastion for admin access
  outbound_rules:
    - protocol: tcp
      port_range: "443"
      destination_addresses:
        - "91.189.88.0/21"  # Ubuntu repos (example, use actual IPs)
    - protocol: udp
      port_range: "53"
      destination_addresses:
        - "1.1.1.1/32"  # Specific DNS resolver (Cloudflare)
```

**Rationale:**
- Database port (5432) only accessible from web tier
- No public internet access
- Outbound locked to specific OS repos and DNS (high security)
- SSH for administrative access only

---

## Key Takeaways

1. **DigitalOcean Cloud Firewalls are stateful, network-edge firewalls** that protect Droplets, Kubernetes nodes, and Load Balancers. They're free, scalable, and enforce a "default deny" security model.

2. **The "double firewall" trap is real.** Cloud Firewalls operate at the network edge. Host-based firewalls (`ufw`, `iptables`) run inside the OS. Both are independent. Traffic must pass both. Troubleshoot both.

3. **Manual management is an anti-pattern.** Use IaC (Terraform, Pulumi, Ansible, or Crossplane) for production. All are mature and support the full firewall API.

4. **Tag-based targeting is the production standard.** Static Droplet IDs hit a 10-resource limit and don't auto-scale. Tags scale infinitely and enable dynamic, automated security.

5. **Compose firewalls by role.** Instead of one monolithic firewall, create focused firewalls for management, web tier, database tier, etc. Apply multiple firewalls to a Droplet via tags. Rules are additive.

6. **The 80/20 config is name, tags, inbound rules, and outbound rules.** Advanced features (Load Balancer UIDs, Kubernetes cluster IDs, tag-based sources/destinations) are essential for multi-tier production architectures.

7. **Planton abstracts the API** into a clean protobuf spec, making multi-cloud deployments consistent while respecting DigitalOcean's unique features (tag-based targeting, resource-aware rules).

---

## Further Reading

- **DigitalOcean Cloud Firewalls Documentation:** [DigitalOcean Docs - Cloud Firewalls](https://docs.digitalocean.com/products/networking/firewalls/)
- **Terraform DigitalOcean Provider:** [digitalocean_firewall Resource](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/firewall)
- **Pulumi DigitalOcean Provider:** [digitalocean.Firewall](https://www.pulumi.com/registry/packages/digitalocean/api-docs/firewall/)
- **Ansible DigitalOcean Collection:** [community.digitalocean.digital_ocean_firewall](https://docs.ansible.com/ansible/latest/collections/community/digitalocean/digital_ocean_firewall_module.html)
- **Crossplane DigitalOcean Provider:** [provider-upjet-digitalocean](https://github.com/crossplane-contrib/provider-upjet-digitalocean)
- **DigitalOcean API Reference:** [Firewalls API](https://docs.digitalocean.com/reference/api/api-reference/#tag/Firewalls)

---

**Bottom Line:** DigitalOcean Cloud Firewalls give you stateful, network-edge security with centralized management and tag-based scalability. Manage them with Terraform or Pulumi, avoid the "double firewall" trap, and adopt tag-first targeting for production. Planton makes this simple with a protobuf API that hides complexity while exposing the essential configuration you actually need.


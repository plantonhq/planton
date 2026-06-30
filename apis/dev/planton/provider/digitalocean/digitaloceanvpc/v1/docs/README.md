# DigitalOcean VPC Deployment: Simplicity as a Strategic Choice

## Introduction

In cloud networking, there's a persistent tension between flexibility and simplicity. AWS gives you subnets, route tables, internet gateways, NAT gateways, and an almost infinite configuration surface. GCP offers VPC peering, shared VPCs, custom route advertisements, and deep integration with Cloud Interconnect. These platforms expose the full complexity of enterprise networking because their customers demand it.

DigitalOcean made a different bet: **what if the VPC was intentionally simple?**

The result is a networking primitive that feels more like a traditional Layer 2 network than a cloud abstraction. There are no managed subnets. No route tables to configure. No complex routing policies. The DigitalOcean VPC is a single, contiguous IP address range that provides network isolation for your Droplets, Kubernetes clusters, and managed databases—and that's it.

This simplicity is not a limitation—it's a feature aligned with DigitalOcean's philosophy. For the vast majority of use cases (development environments, small-to-medium production workloads, startups optimizing for velocity), the added complexity of subnet-level segmentation provides diminishing returns. You don't need five public subnets and three private subnets across availability zones when you're running a web app with a database. You need **isolation from the public internet** and **private communication between your resources**. The DigitalOcean VPC delivers exactly that.

This document explains the deployment methods available for DigitalOcean VPCs, the critical constraints that shape network planning, and how Planton exposes this platform through a minimal, production-ready API.

## Why DigitalOcean VPCs Are Different

### The L2 Domain Model

The DigitalOcean VPC is designed to "look and feel like a Layer 2 domain." What does that mean in practice?

When you inspect the routing table on a Droplet within a VPC, you'll notice something unusual: the route for the VPC's CIDR block (e.g., `10.110.16.0/20`) has **no gateway IP assigned**. Traffic between Droplets in the same VPC is **L2-forwarded using MAC addresses**, not L3-routed through a virtual router.

This explains why there are no subnet or route table resources—all resources within a VPC are in the same broadcast domain. It's conceptually simpler, but it also means that any fine-grained network segmentation (like creating "public" and "private" subnets) is not a platform-level construct. If you need that pattern, you must implement it manually at the OS level by configuring a Droplet as an internet gateway and manipulating OS-level routes.

### What You Get (The Essentials)

**Core VPC Features:**
- **Private network isolation**: Resources in a VPC communicate via private IP addresses, invisible to the public internet
- **Regional scope**: One VPC per region (e.g., `nyc3`, `sfo3`)
- **Free internal traffic**: All traffic within a VPC, or between peered VPCs in the same datacenter, is free and doesn't count against bandwidth quotas
- **Automatic IP assignment**: Resources provisioned into a VPC automatically receive a private IP from the VPC's CIDR range

**Managed Adjacent Services:**
- **VPC Peering**: Connect two VPCs (same region or cross-region) for secure, private communication without traversing the public internet
- **VPC-local DNS**: Built-in internal DNS resolver on the second-to-last IP of the network range (e.g., `10.10.255.254`)
- **NAT Gateway** (Public Preview): Centralized outbound internet access for private resources without exposing them via public IPs
- **Partner Network Connect**: Private interconnections with third-party providers (niche use case)

### The Critical Constraints

Understanding these hard limits is essential for network planning:

**Immutability (The "No Resize" Rule):**  
A VPC's IP address range **cannot be changed or resized** after creation. If you run out of IP addresses, your only option is to provision a new, larger VPC and execute a painful, resource-by-resource migration. This makes initial CIDR block planning the most critical strategic decision.

**Regional Confinement:**  
A VPC is strictly confined to a single datacenter region. Multi-region architectures require multiple VPCs connected via VPC Peering.

**Resource Limits:**
- Maximum 10,000 resources per VPC (Droplets, database nodes, etc.)
- Resources cannot be in multiple VPCs (no multi-homed networking)
- **Critical**: DigitalOcean Kubernetes (DOKS) clusters and Load Balancers **cannot be migrated** into a VPC after creation—they must be created inside the VPC from day one

**CIDR Block Restrictions:**
1. Must use RFC1918 private address space (`10.0.0.0/8`, `172.16.0.0/12`, or `192.168.0.0/16`)
2. Cannot overlap with other VPCs in your account
3. Cannot overlap with VPCs you intend to peer with (even in different regions)
4. Cannot use DigitalOcean's reserved ranges: `10.244.0.0/16`, `10.245.0.0/16`, `10.246.0.0/24`, `10.229.0.0/16`, and region-specific reservations like `10.10.0.0/16` in `nyc1`
5. Minimum size: `/28` (16 IPs) technically supported, but `/24` (256 IPs) is the practical minimum due to reserved platform IPs

## The Deployment Maturity Spectrum

### Level 0: The Cloud Console (The "Easy Button" and Its Traps)

**Method**: DigitalOcean Cloud Control Panel (web UI)  
**Verdict**: ✅ Perfect for learning and experimentation | ❌ Anti-pattern for production

The web console provides the simplest possible workflow:
1. Navigate to Networking → VPC
2. Click "Create VPC Network"
3. Enter a **Name** and select a **Region**
4. Choose IP addressing: "Generate an IP range for me" or "Choose a custom range"

**The "Generate" Option—The 80% Use Case:**  
When you select "Generate an IP range for me," DigitalOcean automatically allocates a non-conflicting `/20` CIDR block (4,096 IP addresses). This represents the 80% use case: users who want isolation without managing IP address planning.

**Common Pitfalls:**
1. **CIDR Overlap**: Manually choosing `192.168.1.0/24` because it's familiar, only to discover later that it conflicts with your home network, making VPN connections impossible
2. **Undersizing**: Choosing a `/24` because it "seems big enough," then being trapped by the immutability constraint when you need more IPs
3. **The "VPC-Later" Mistake**: Creating Droplets and databases first, then creating a VPC later, only to discover that DOKS clusters and Load Balancers cannot be migrated

**What it teaches you**: The basic parameters (name, region, IP range) and the trade-off between automated IP allocation and explicit control.

**What it doesn't solve**: Repeatability, version control, multi-environment consistency, or disaster recovery. Every manually-created VPC is a snowflake.

### Level 1: CLI Scripting with doctl (Scriptable, but Imperative)

**Method**: DigitalOcean CLI (`doctl`)  
**Verdict**: ⚠️ Suitable for dev/test automation and one-time migrations | ❌ Not ideal for production infrastructure

The official CLI provides scriptable VPC provisioning:

```bash
# Minimal (80% use case): Auto-generated /20 CIDR
doctl vpcs create \
  --name "dev-vpc" \
  --region "nyc3"

# Explicit (20% use case): User-defined /16 CIDR
doctl vpcs create \
  --name "prod-vpc" \
  --region "nyc3" \
  --ip-range "10.100.0.0/16" \
  --description "Main production VPC"
```

**Key Behavior:**  
The `--ip-range` flag is **optional**. When omitted, DigitalOcean automatically generates a non-conflicting `/20` CIDR block. This matches the web console's "easy button" behavior and is a critical API capability that Planton must support.

**What it solves**: Automation and CI/CD integration. Scripts can be version-controlled and reused across environments.

**What it doesn't solve**: State management. If you run the script twice, you create two VPCs. If creation fails halfway, cleanup is manual. There's no drift detection or rollback.

**Use cases**: Quick experimentation, ephemeral dev environments, one-time migrations from other platforms.

### Level 2: Configuration Management with Ansible (Declarative-ish)

**Method**: Ansible (`community.digitalocean.digital_ocean_vpc` or `digitalocean.cloud.vpc`)  
**Verdict**: ⚠️ Useful when paired with Terraform | ⚠️ Not ideal as a standalone provisioning tool

Ansible provides an idempotent VPC module:

```yaml
- name: Create production VPC
  digitalocean.cloud.vpc:
    state: present
    oauth_token: "{{ lookup('env', 'DO_API_TOKEN') }}"
    name: "prod-vpc"
    region: "nyc3"
    ip_range: "10.100.0.0/16"
    description: "Production VPC for all services"
```

**The `ip_range` Default:**  
Like `doctl`, the Ansible module confirms: "If no IP range is specified, a `/20` network range is generated that won't conflict with other VPC networks in your account."

**The `default` Parameter Trap:**  
The Ansible module uniquely provides a boolean `default` input parameter for setting the regional default VPC. This appears to contradict the DigitalOcean API, which handles the default status automatically (the first VPC in a region becomes the default).

Closer inspection reveals the truth: when you set `default: true`, the Ansible module performs a **non-atomic, two-step operation**:
1. Sends a POST request to create the VPC
2. Sends a separate PUT/POST request to a "Set Default VPC" endpoint

If the second step fails, you're left in an inconsistent state. This is a high-level abstraction that masks the platform's true behavior—exactly the kind of abstraction Planton avoids.

**Best Practice Pattern**: **Terraform (provision) → Ansible (configure)**. Use Terraform for stateful infrastructure provisioning (VPCs, Droplets, firewalls), then use Ansible for what it's best at: installing software and configuring services on those resources.

### Level 3: Infrastructure as Code (The Production Standard)

**Methods**: Terraform, Pulumi, OpenTofu  
**Verdict**: ✅ **Recommended for all production deployments**

This is the gold standard for managing DigitalOcean VPCs. IaC provides state management, drift detection, plan previews, version control, and a declarative model.

#### Terraform

DigitalOcean maintains the official `digitalocean` provider in the HashiCorp Registry.

**The Essential Resource:**

```hcl
resource "digitalocean_vpc" "prod_vpc" {
  name     = "prod-vpc"
  region   = "nyc3"
  ip_range = "10.100.0.0/16"
  description = "Main production VPC"
}

# Reference in dependent resources
resource "digitalocean_droplet" "web_server" {
  name     = "web-01"
  region   = "nyc3"
  size     = "s-2vcpu-4gb"
  image    = "ubuntu-22-04-x64"
  vpc_uuid = digitalocean_vpc.prod_vpc.id  # Implicit dependency
}
```

**Critical API Behaviors:**

1. **`ip_range` is Optional**: The Terraform documentation confirms this field is optional. When omitted, the value appears as `(known after apply)` in the plan, proving the provider correctly handles DigitalOcean's auto-generation behavior.

2. **`default` is Output-Only**: The Terraform docs list `default` *only* under "Attributes Reference" (outputs), **not** under "Argument Reference" (inputs). This is the correct model: `default` reflects the VPC's current state as reported by the API, not a user's intent at creation time. It's automatically set to `true` only if the VPC is the first created in that region.

**Managing Dependencies:**  
Resource dependencies are handled implicitly through attribute references. When you reference `digitalocean_vpc.prod_vpc.id` in another resource's `vpc_uuid` field, Terraform understands it must create the VPC first, wait for its `id` to be available, then create the dependent resource.

#### Pulumi

Pulumi provides the same resource coverage via its `digitalocean` package, available in TypeScript, Python, Go, and C#.

**Example (TypeScript):**

```typescript
import * as digitalocean from "@pulumi/digitalocean";

const prodVpc = new digitalocean.Vpc("prod-vpc", {
    name: "prod-vpc",
    region: "nyc3",
    ipRange: "10.100.0.0/16",
    description: "Production VPC",
});

const webServer = new digitalocean.Droplet("web-server", {
    name: "web-01",
    region: "nyc3",
    size: "s-2vcpu-4gb",
    image: "ubuntu-22-04-x64",
    vpcUuid: prodVpc.id,  // Implicit dependency
});
```

**Advantages over Terraform:**
- **Programming model**: Use loops, conditionals, functions, and type checking
- **Built-in secret management**: Encrypted configuration storage via `pulumi config set --secret`
- **Managed state by default**: Pulumi Cloud handles state storage and concurrency locking automatically

**Equivalence**: For standard VPC provisioning, Terraform and Pulumi are equally capable. The choice is team preference: HCL's declarative simplicity vs. a general-purpose programming language's flexibility.

#### OpenTofu

As a compatible fork of Terraform, OpenTofu works seamlessly with the `digitalocean` provider (which is itself open-source). No code changes required—simply use `tofu` instead of `terraform` commands.

### Level 4: Multi-Environment Strategy (Directory-Based Isolation)

**The Anti-Pattern: Terraform Workspaces**

A common misconception is using Terraform workspaces for environment separation (e.g., `terraform workspace new dev`). This is dangerous.

Workspaces only create separate **state files**. They **do not** isolate code, variables, or backend configuration. This forces engineers to write risky conditional logic:

```hcl
# ANTI-PATTERN: Don't do this
resource "digitalocean_vpc" "vpc" {
  name     = terraform.workspace == "prod" ? "prod-vpc" : "dev-vpc"
  ip_range = terraform.workspace == "prod" ? "10.100.0.0/16" : "10.10.0.0/20"
  region   = "nyc3"
}
```

If you run a command in the wrong workspace, you could damage or destroy the production environment. Workspaces are only suitable for creating multiple identical instances of the same stack (e.g., per-feature-branch test environments), not for separating dev/staging/prod.

**Best Practice: Directory-Based Isolation**

The industry-standard pattern is a separate directory for each environment:

```
/infra-live/
  /modules/
    /vpc/
      main.tf       # Defines digitalocean_vpc resource
      variables.tf  # Declares var.ip_range, var.name
      outputs.tf    # Outputs vpc.id
  /environments/
    /dev/
      main.tf           # Calls module.vpc
      terraform.tfvars  # Sets ip_range = "10.10.0.0/20"
      backend.tf        # Sets state key = "env/dev.tfstate"
    /staging/
      main.tf           # Calls module.vpc
      terraform.tfvars  # Sets ip_range = "10.20.0.0/20"
      backend.tf        # Sets state key = "env/staging.tfstate"
    /prod/
      main.tf           # Calls module.vpc
      terraform.tfvars  # Sets ip_range = "10.30.0.0/16"
      backend.tf        # Sets state key = "env/prod.tfstate"
```

**Why This Works:**
- Each environment has its **own state file**, often in separate remote backends
- Each environment has its **own variable definitions**
- Shared, reusable logic lives in `/modules`
- A `terraform apply` in `/dev` has **no ability to modify `/prod` state**
- Non-overlapping CIDR blocks are configured per-environment via `.tfvars` files

This provides maximum isolation and is the correct mechanism for managing VPCs across multiple environments.

### Level 5: Kubernetes-Native Provisioning with Crossplane

**Method**: Crossplane (`provider-upjet-digitalocean`)  
**Verdict**: ⚠️ Use only if you've standardized on a Kubernetes-native control plane

For organizations that manage all infrastructure as Kubernetes Custom Resources, Crossplane is an option. The modern, maintained provider is `crossplane-contrib/provider-upjet-digitalocean` (the older `provider-digitalocean` is archived).

**Example VPC CRD:**

```yaml
apiVersion: digitalocean.upjet.crossplane.io/v1beta1
kind: Vpc
metadata:
  name: prod-vpc-from-k8s
spec:
  forProvider:
    name: "prod-vpc"
    region: "nyc3"
    ipRange: "10.200.0.0/16"
    description: "Production VPC managed by Crossplane"
```

**The "Upjet" Designation:**  
This provider is auto-generated from the Terraform provider, meaning it shares the exact same API schema and behavior. The YAML structure maps directly to Terraform's HCL arguments.

**When to Use It:**  
Only if you're already running Crossplane and want to manage all infrastructure through Kubernetes. This pattern adds significant complexity—for most teams, Terraform or Pulumi is simpler and more portable.

## Production Best Practices

### CIDR Block Planning: The "T-Shirt Sizing" Strategy

Because VPCs cannot be resized, initial CIDR block planning is critical. Here's the recommended approach:

**RFC1918 Selection:**  
All VPC ranges must come from private address spaces. The `10.0.0.0/8` block is the most common choice for cloud-native infrastructure, as it provides a massive, contiguous space to carve up.

**Strategic IPAM (IP Address Management):**  
Don't choose CIDR blocks randomly. Best practice: perform organizational-level IPAM **before** creating any VPCs.

For example, assign a large `/16` block to each DigitalOcean region:
- `10.100.0.0/16` for `nyc3`
- `10.101.0.0/16` for `sfo3`
- `10.102.0.0/16` for `fra1`

Then, carve out smaller blocks for individual VPCs within each region:
- `10.100.0.0/20` for `nyc3-dev`
- `10.100.16.0/20` for `nyc3-staging`
- `10.100.32.0/16` for `nyc3-prod` (larger for production scaling)

This centralized planning prevents all future routing and peering conflicts.

**Recommended Sizes ("T-Shirt Sizing"):**

| **Size** | **CIDR** | **Usable IPs** | **Use Case** |
|----------|----------|----------------|--------------|
| **Small** | `/24` | 256 | Minimum practical size. Suitable for small, single-purpose apps, sandboxes, isolated utility VPCs (e.g., bastion hosts). |
| **Medium** | `/20` | 4,096 | **DigitalOcean's auto-generated default**. Recommended for dev, staging, and medium-sized production applications. |
| **Large** | `/16` | 65,536 | **Maximum allowed size**. Use for main production VPCs expected to host large, scaling workloads, especially VPC-native DOKS clusters (where pods consume IPs directly from the VPC range). |

**Why Over-Provision:**  
Because resizing is impossible, always over-provision IP address space. The cost is zero—you're not charged per IP, only per resource. A `/16` costs the same as a `/24`. When in doubt, go bigger.

### The "VPC-First" Imperative

**The Anti-Pattern**: Creating Droplets, databases, and Kubernetes clusters first, then creating a VPC later and trying to migrate resources into it.

**The Problem**: 
- **DOKS clusters cannot be migrated**—they must be destroyed and recreated from scratch inside the new VPC
- **Load Balancers cannot be migrated**—they must be destroyed and recreated, causing DNS-level cutover and potential downtime
- **Droplets require a painful migration**: power off → snapshot → recreate from snapshot → update all DNS/firewall/load balancer references

**The Production Standard:**  
The VPC must be the **first resource** created in any new environment. All subsequent resources (databases, Kubernetes clusters, Droplets) are provisioned **into** the VPC from day one.

**Terraform Pattern:**

```hcl
# 1. Create VPC FIRST
resource "digitalocean_vpc" "prod" {
  name     = "prod-vpc"
  region   = "nyc3"
  ip_range = "10.100.0.0/16"
}

# 2. Create resources INTO the VPC
resource "digitalocean_kubernetes_cluster" "prod_cluster" {
  name     = "prod-cluster"
  region   = "nyc3"
  version  = "1.28.2-do.0"
  vpc_uuid = digitalocean_vpc.prod.id  # Cannot be changed later
  
  node_pool {
    name       = "worker-pool"
    size       = "s-2vcpu-4gb"
    node_count = 3
  }
}

resource "digitalocean_database_cluster" "postgres" {
  name       = "prod-postgres"
  engine     = "pg"
  version    = "15"
  size       = "db-s-2vcpu-4gb"
  region     = "nyc3"
  node_count = 1
  
  private_network_uuid = digitalocean_vpc.prod.id  # Private access only
}
```

### Critical Anti-Patterns to Avoid

❌ **The Overlapping CIDR Block**  
If you create `dev-vpc` with `10.10.0.0/16` and `prod-vpc` with `10.10.10.0/24`, these ranges overlap. While the DigitalOcean control panel will warn you, IaC tools often won't. The consequence: it becomes **permanently impossible** to establish VPC Peering between these networks, severing all private connectivity.

❌ **The "Too-Small" VPC**  
Choosing a `/24` for a production application expected to grow. As you add Droplets, database replicas, and a DOKS cluster, you exhaust the 256 available IPs. The only solution is a full, disruptive migration to a new, larger VPC.

❌ **Ignoring Reserved Ranges**  
Using `10.244.0.0/16` or `10.10.0.0/16` in `nyc1` because you didn't check DigitalOcean's reserved ranges. The VPC creation will fail with a cryptic error message.

❌ **Home Network Overlap**  
Choosing `192.168.1.0/24` because it's familiar, then discovering you can't establish a VPN or SSH tunnel from your home network (which also uses `192.168.1.0/24`). Always use the `10.0.0.0/8` space for cloud infrastructure.

### Migration Strategies (When Unavoidable)

If you're stuck with the "VPC-Later" scenario, here's how to migrate:

**Migrating Managed Databases (Graceful, Non-Disruptive):**  
This is the only resource with in-place migration support:
1. Navigate to the database cluster in the Control Panel
2. Click "Settings" → "Cluster datacenter" → "Edit"
3. Select the target VPC from the dropdown
4. Click "Save"

DigitalOcean migrates the database in the background with zero downtime, eventually assigning it a new private IP within the target VPC's range.

**Migrating Droplets (Disruptive, Requires Downtime):**
1. Power off the source Droplet (`sudo shutdown -h now`)
2. Take a Snapshot in the Control Panel
3. Create a new Droplet from the snapshot
4. Select the target VPC during creation
5. Manually update all DNS records, load balancers, and firewall rules to point to the new private IP

**Migrating DOKS Clusters and Load Balancers:**  
These **cannot be migrated**. They must be destroyed and recreated inside the new VPC. For DOKS, this means a full cluster rebuild and workload redeployment—essentially a complete environment cutover.

## The Planton Choice

Planton's `DigitalOceanVpc` API is designed following the **80/20 principle**: expose the 20% of configuration that 80% of users need, enforce security and best practices by default, and avoid abstractions that don't exist in the native platform.

### The Minimal, Production-Ready API

Based on the research findings, the Planton API specification should include:

**Input (Spec) Fields:**

| **Field** | **Type** | **Required?** | **Rationale** |
|-----------|----------|---------------|---------------|
| `name` | string | ✅ Yes | Human-readable VPC name. Must be unique per region. |
| `region` | enum | ✅ Yes | DigitalOcean region slug (e.g., `nyc3`, `sfo3`). Determines geographical location. |
| `ip_range` | string | ❌ **No** | CIDR block (e.g., `10.100.0.0/20`). **Optional to support the 80% use case**: when omitted, DigitalOcean auto-generates a non-conflicting `/20` block. |
| `description` | string | ❌ No | Free-form text (max 255 chars) describing the VPC's purpose. Useful for documentation and auditing. |

**Output (Status) Fields:**

| **Field** | **Type** | **Description** |
|-----------|----------|-----------------|
| `id` | string | Unique identifier (UUID) for the VPC. Used in `vpc_uuid` fields of dependent resources. |
| `urn` | string | Uniform resource name for the VPC. |
| `is_default` | bool | Indicates whether this VPC is the default for its region. **Output-only**: automatically set to `true` by DigitalOcean if this is the first VPC created in the region. Cannot be set as an input. |
| `ip_range_computed` | string | The actual IP range assigned to the VPC. Essential for discovering what range was auto-generated when `ip_range` was omitted. |
| `created_at` | timestamp | When the VPC was created. |

### Critical Design Decisions

**Why `ip_range` Must Be Optional:**  
This is the 80% use case. Most users want the platform to handle IP allocation automatically. Making this field required forces users to manage CIDR planning even when they don't need to. When `ip_range` is omitted, the Planton provider must send a Create request **without** the `ip_range` key, triggering DigitalOcean's auto-generation behavior.

**Why `is_default` Must Be Output-Only:**  
The native DigitalOcean API does not support setting `is_default` during VPC creation. The first VPC created in a region automatically becomes the default—this is a side effect of creation order, not a settable parameter.

The Ansible module's `default` input parameter is a high-level abstraction that performs a non-atomic, two-step operation (create VPC, then call a separate "Set Default" endpoint). This masks the platform's true behavior and creates inconsistency risks.

Planton avoids this abstraction. If users need to change the default VPC, that must be implemented as a separate Update operation or dedicated action, accurately reflecting the non-atomic nature of the operation.

**Why Description is Optional:**  
While useful for production documentation, it's not required for a functioning VPC. The 80% use case (dev/test environments, quick provisioning) often skips this field.

### Reference Configurations

**Dev VPC (Minimal, Auto-Generated CIDR):**

```hcl
# This represents the 80% use case
resource "digitalocean_vpc" "dev_vpc" {
  name   = "dev-vpc"
  region = "nyc3"
}

# Result: DigitalOcean auto-generates a /20 block (e.g., 10.116.0.0/20)
# If this is the first VPC in nyc3, is_default = true automatically
```

**Staging VPC (Explicit /20 for Non-Overlapping CIDR):**

```hcl
# This represents the 20% use case: explicit IP planning
resource "digitalocean_vpc" "staging_vpc" {
  name     = "staging-vpc"
  region   = "nyc3"
  ip_range = "10.100.16.0/20"  # Non-overlapping with dev and prod
}
```

**Production VPC (Maximum Size for Scaling):**

```hcl
# Production VPC: max-size /16 for DOKS VPC-native networking
resource "digitalocean_vpc" "prod_vpc" {
  name        = "prod-vpc"
  region      = "nyc3"
  ip_range    = "10.101.0.0/16"  # 65,536 IPs for growth
  description = "Main production VPC for all services"
}
```

## Conclusion

The DigitalOcean VPC is a deliberate exercise in **simplicity as a strategic choice**. By removing the complexity of subnets, route tables, and multi-tier networking, DigitalOcean delivers a networking primitive that's fast to deploy, easy to understand, and sufficient for the vast majority of production workloads.

But this simplicity comes with rigid constraints:
- **Immutability**: No resizing means over-provision or face painful migrations
- **VPC-First**: DOKS clusters and Load Balancers cannot be migrated, making the VPC the foundation of all infrastructure
- **CIDR Planning**: Overlapping ranges permanently prevent peering and VPN connectivity

The production standard for managing VPCs is **Infrastructure as Code** (Terraform or Pulumi), with **directory-based environment isolation** for dev/staging/prod. The 80% use case is auto-generated CIDR blocks (the platform's `/20` default). The 20% use case is explicit CIDR planning using organizational IPAM strategies.

**Planton's API design enforces these patterns**: `ip_range` is optional (supporting the 80% use case), `is_default` is output-only (accurately modeling the platform), and the configuration surface is minimal—name, region, optional IP range, optional description. This gives users the flexibility to "just create a VPC" for dev environments while supporting explicit, production-grade IP planning for critical workloads.

For organizations that value **velocity, simplicity, and cost efficiency** over enterprise networking complexity, the DigitalOcean VPC delivers exactly what's needed—and nothing more.

---

**For comprehensive guides on related topics, see:**

- [VPC Peering Strategies for Multi-Region Architectures](./vpc-peering-guide.md) *(placeholder)*
- [Integrating VPC-Native DOKS Clusters with Managed Databases](./vpc-native-doks.md) *(placeholder)*
- [IP Address Management (IPAM) Best Practices for Multi-Cloud](./ipam-guide.md) *(placeholder)*
- [Migrating Legacy Infrastructure into DigitalOcean VPCs](./migration-guide.md) *(placeholder)*


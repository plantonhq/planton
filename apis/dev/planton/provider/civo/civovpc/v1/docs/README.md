# Deploying Civo VPCs: Embracing Simplicity in Network Isolation

## Introduction

In the hyperscale cloud era, VPC architecture has become synonymous with complexity. AWS offers VPC Peering, Transit Gateways, PrivateLink, and multi-AZ subnet orchestration. GCP provides global VPCs with regional subnets and Private Service Connect. Azure adds VNet Peering and Service Endpoints to the mix. For teams managing enterprise-scale deployments across multiple regions and hundreds of microservices, these features are essential. But for the majority of workloads—the 80% of businesses that use fewer than ten cloud services—this complexity is overkill.

**Civo VPCs** (called "private networks" in Civo's documentation) represent a different philosophy: **simplicity as a feature, not a limitation**. A Civo VPC is a strict Layer 3 network isolation boundary, purpose-built for the most common use cases: environment separation (dev/staging/prod), multi-tenant workload isolation, and secure bastion-based access patterns. There are no subnets to manage, no availability zones to orchestrate, and no complex routing tables to configure. A Civo VPC is a single, flat network—limited to a /24 CIDR block (256 IP addresses)—that lives in one region and provides true isolation for Kubernetes clusters, compute instances, and load balancers.

What makes this architecture compelling is not what it lacks, but what it enables: **zero data egress fees**, **transparent pricing**, and **operational simplicity**. Unlike hyperscalers that charge for inter-AZ traffic or VPC endpoint hours, Civo's model encourages best practices like creating separate networks for each environment without financial penalty. The platform's OpenStack-based networking ensures true tenant isolation—not the pseudo-private networks some clouds offer where resources from different accounts might share infrastructure.

This guide explains the landscape of deployment methods for Civo VPCs, from manual dashboard clicks to production-grade Infrastructure as Code, and shows how Planton abstracts these choices into a clean, protobuf-defined API that works consistently across clouds.

---

## The Deployment Spectrum: From Manual to Production

Not all approaches to managing network isolation are created equal. Here's how the methods stack up, from what to avoid to what works at scale:

### Level 0: The Default Network (Anti-Pattern for Production)

**What it is:** Every Civo region automatically provisions a "Default" network when you create your account. Any resource (Kubernetes cluster, compute instance) created without specifying a network ID lands here.

**What it solves:** Initial convenience. You can spin up a Kubernetes cluster in seconds without thinking about networking.

**What it doesn't solve:** Isolation, security, or predictability. The Default network is a shared space within your account—mixing dev experiments, staging workloads, and potentially production resources. It's the equivalent of AWS's default VPC: functional for learning, unacceptable for production.

**Verdict:** Never deploy production workloads here. Treat the Default network as a landing zone for throwaway tests, nothing more. Architecturally, using the Default network is like running all your applications in the same Kubernetes namespace—technically possible, operationally dangerous.

---

### Level 1: Manual Dashboard Provisioning (Learning Phase Only)

**What it is:** Using Civo's web console to create a network by navigating to "Networking," clicking "Create a network," and entering a human-readable label. The console auto-allocates a CIDR block if you don't specify one.

**What it solves:** Exploration and learning. The workflow is simple enough to understand Civo's networking model in minutes. You see that networks are regional, that CIDR blocks are limited to /24, and that assignment to resources is permanent.

**What it doesn't solve:** Repeatability, version control, or auditability. Every click is a manual operation with no record beyond your browser history. Scaling to multiple environments means duplicating manual steps—a recipe for misconfigurations and drift.

**Verdict:** Use the dashboard to understand Civo's networking concepts, then immediately transition to Infrastructure as Code for any environment that matters (including dev).

---

### Level 2: CLI Scripting (Scriptable, But Stateless)

**What it is:** Using the `civo` CLI to provision networks in shell scripts:

```bash
civo network create prod-main-network --cidr 10.10.1.0/24 --region LON1
```

**What it solves:** Automation and repeatability. You can script network creation, integrate it into CI/CD pipelines, and version-control your scripts. The CLI is synchronous and returns structured output (JSON mode available with `--output json`).

**What it doesn't solve:** State management. Scripts are imperative—they execute commands in order but don't track what already exists. Run the script twice, and you'll get an error (network already exists) or create duplicates with slightly different names. There's no declarative model, no plan/preview, and no automatic cleanup on failure.

**Verdict:** Acceptable for integration tests or throwaway dev environments where you create and destroy everything in one pass. Not suitable for production, where you need idempotency, state tracking, and rollback capabilities.

---

### Level 3: Direct API Integration (Maximum Flexibility, High Maintenance)

**What it is:** Calling Civo's REST API directly from custom tooling, orchestration systems, or configuration management tools (e.g., Ansible using the `uri` module).

**What it solves:** Complete control. The API exposes a `POST /v2/networks` endpoint with clear parameters: `label` (required), `region` (required), and `cidr_v4` (optional, auto-allocated if omitted). You can integrate network provisioning into any HTTP-capable system.

**What it doesn't solve:** Abstraction or state management. You're handling HTTP authentication (Bearer tokens), implementing idempotency yourself (checking if the network exists before creating it), and sequencing operations manually (create network → create firewall → attach to cluster). Essentially, you're building your own IaC layer.

**Critical Gap: No Ansible Support.** Unlike Terraform or Pulumi, there is **no official Ansible collection for Civo**. The `community.network` collection is deprecated and doesn't include Civo modules. Teams standardized on Ansible must use workarounds—calling the API via `ansible.builtin.uri` or shelling out to the `civo` CLI with `ansible.builtin.command`. This is a significant friction point for Ansible-first organizations.

**Verdict:** Useful if you're building a custom control plane (like Planton) or integrating Civo into a broader orchestration framework. For most teams, higher-level IaC tools (Terraform, Pulumi) handle API calls and state management for you.

---

### Level 4: Infrastructure as Code (Production-Ready)

**What it is:** Using Terraform or Pulumi with Civo's official provider to declaratively define networks and their lifecycle.

**Terraform example:**

```hcl
provider "civo" {
  token  = var.civo_api_key
  region = "LON1"
}

resource "civo_network" "prod_net" {
  label   = "prod-main-network"
  region  = "LON1"
  cidr_v4 = "10.10.1.0/24"
}

resource "civo_firewall" "prod_fw" {
  name       = "prod-firewall"
  region     = "LON1"
  network_id = civo_network.prod_net.id
}

resource "civo_kubernetes_cluster" "prod_cluster" {
  name        = "prod-cluster"
  region      = "LON1"
  network_id  = civo_network.prod_net.id
  firewall_id = civo_firewall.prod_fw.id
}
```

**Pulumi example (TypeScript):**

```typescript
import * as civo from "@pulumi/civo";

const prodNet = new civo.Network("prod-net", {
    label: "prod-main-network",
    region: "LON1",
    cidrv4: "10.10.1.0/24",
});

const prodFw = new civo.Firewall("prod-fw", {
    name: "prod-firewall",
    region: "LON1",
    networkId: prodNet.id,
});

const prodCluster = new civo.KubernetesCluster("prod-cluster", {
    name: "prod-cluster",
    region: "LON1",
    networkId: prodNet.id,
    firewallId: prodFw.id,
});
```

**What it solves:** Everything. You get:
- **Declarative configuration**: State what you want, not how to get there
- **State management**: Terraform/Pulumi track what exists and detect drift
- **Idempotency**: Running the same config twice produces the same result
- **Plan/preview**: See changes before applying them (critical for the "permanent assignment" constraint—see below)
- **Version control**: Treat infrastructure as code, with diffs, reviews, and rollbacks
- **Multi-environment support**: Reuse configs for dev/staging/prod with different parameters
- **Dependency graphs**: Automatically sequence operations (create network before firewall, firewall before cluster)

**What it doesn't solve:** The underlying architectural constraints of Civo VPCs (no peering, no subnets, max /24, permanent network assignment). But it makes managing those constraints predictable and reproducible.

**Verdict:** This is the production standard. Both Terraform and Pulumi are fully supported by Civo and production-ready. The choice between them is team preference, not capability.

---

### Level 5: Kubernetes Control Plane (Crossplane)

**What it is:** Using Crossplane's GitOps model to define Civo networks as Kubernetes custom resources.

**The Ecosystem Problem:** Civo's Crossplane story is fragmented and immature compared to hyperscalers. There are **two competing providers**:

1. **`crossplane-contrib/provider-civo`** (official contrib): Only supports Kubernetes clusters and compute instances. **Does not support networks.** This is the provider you'd expect to use, but it's incomplete.

2. **`upsidr/provider-civo-upjet`** (community, Upjet-based): Auto-generated from the Terraform provider. **Does support networks.** This is the one you must use if you need network management via Crossplane.

**What it solves (when using the Upjet provider):** GitOps-based network provisioning. You define networks as Kubernetes `Network` resources, and Crossplane reconciles them. This fits naturally into teams already using ArgoCD or Flux for GitOps.

**What it doesn't solve:** The confusion. Having two providers with overlapping but incomplete functionality is a friction point. Teams must research which provider to use, and community providers lack the long-term maintenance guarantees of official providers.

**Verdict:** Only use Crossplane for Civo if you're already committed to Crossplane for other clouds and need unified GitOps workflows. For Civo-only deployments, Terraform or Pulumi are simpler and more mature.

---

## IaC Tool Comparison: Terraform vs. Pulumi

Both Terraform and Pulumi are production-ready for Civo VPCs. Here's how they compare:

### Terraform: The Battle-Tested Standard

**Maturity:** Terraform has been the IaC gold standard for years. Civo's official provider (`civo/civo`) is stable, well-documented, and covers all network operations: create, update (limited), delete. The provider is open-source on GitHub and actively maintained by Civo.

**Configuration Model:** Declarative HCL. You define resources and their relationships, and Terraform builds a dependency graph and execution plan.

**State Management:** Local or remote backends (Civo Object Store, S3, Terraform Cloud). Terraform tracks resource IDs (like `network_id`) in state, making updates and deletions predictable.

**Strengths:**
- **Broad adoption**: Most ops teams know Terraform. Onboarding is easier.
- **Mature ecosystem**: Thousands of providers, modules, and community resources.
- **Clear workflow**: `terraform plan` shows exactly what will change before you apply it. This is critical for Civo because network assignment is **permanent**—changing a cluster's `network_id` triggers a destructive replacement (the cluster is destroyed and recreated).
- **Proven at scale**: Battle-tested in production for complex multi-cloud environments.

**Limitations:**
- **HCL expressiveness**: HCL is less flexible than a full programming language. Complex logic (conditional resource creation, loops over dynamic data) can be verbose or require workarounds.
- **No compile-time validation**: You only discover errors when you run `terraform plan`, not when you write code.

**Verdict:** Default choice for teams already using Terraform or prioritizing ecosystem maturity and team familiarity.

---

### Pulumi: The Programmer's IaC

**Maturity:** Newer than Terraform, but production-ready. Pulumi's Civo provider is **bridged from the Terraform provider**, meaning it's auto-generated from the same source code. This guarantees 1:1 feature parity for all resources, including `civo.Network`.

**Configuration Model:** Real programming languages (TypeScript, Python, Go, C#). Infrastructure is expressed as code with loops, conditionals, functions, and unit tests.

**State Management:** Pulumi Cloud (managed) or self-hosted backends (S3, Azure Blob, Civo Object Store). Similar state tracking to Terraform.

**Strengths:**
- **Language familiarity**: If your team writes TypeScript or Python, Pulumi feels natural. No new DSL to learn.
- **Expressiveness**: Complex provisioning logic (e.g., "create N networks based on a config file") is cleaner in code than in HCL.
- **Testability**: Write unit tests for infrastructure logic using standard test frameworks (Jest, pytest).
- **Abstraction**: You can build reusable infrastructure components as libraries and publish them to package registries (npm, PyPI).

**Limitations:**
- **Smaller ecosystem**: Fewer community modules and resources than Terraform.
- **Runtime dependency**: You need Node.js, Python, or Go runtime. Terraform is a single binary.
- **Bridged provider**: Any quirks or bugs in the Terraform provider carry over to Pulumi. You're dependent on Civo's Terraform provider updates.

**Verdict:** Great choice if your team prefers writing infrastructure in familiar languages or needs complex orchestration logic (dynamic resource generation, conditionals, multi-stage deployments). Slightly more overhead than Terraform for simple use cases.

---

### Which Should You Choose?

**Decision Matrix:**

| Factor | Choose Terraform | Choose Pulumi |
|--------|-----------------|---------------|
| **Team skillset** | Terraform experience, ops-focused | TypeScript/Python/Go expertise, dev-focused |
| **Complexity** | Simple, declarative network configs | Complex logic, dynamic resource generation |
| **Ecosystem** | Need mature modules and community | Need language-native libraries |
| **Testing** | `terraform validate` and manual | Unit tests with standard frameworks |
| **Onboarding** | Most engineers know HCL basics | Engineers already know programming languages |

**Both work equally well** for standard Civo VPC provisioning. The choice is team preference, not capability. Planton supports both by abstracting the underlying IaC tool behind a protobuf API.

---

## Civo VPC Architecture: Understanding the Constraints

Before using any deployment method, it's critical to understand Civo's architectural model and how it differs from hyperscalers.

### The /24 Limit: Many Small Networks, Not One Big One

**Key Constraint:** Civo networks are limited to a maximum prefix size of **/24** (256 IP addresses). This is fundamentally different from AWS (up to /16, 65,536 IPs) or GCP (custom subnet sizes).

**Why It Works:** Civo Kubernetes clusters use **overlay networking** (Flannel or Cilium). Pods get IPs from a separate virtual network (e.g., `10.42.0.0/16`), decoupled from the host network. The Civo VPC's /24 address space is only consumed by:
- Kubernetes cluster nodes (usually 3-10)
- Standalone compute instances
- Civo Load Balancers

**Design Pattern:** Embrace **many small, isolated networks** instead of one large, monolithic network. For example:
- `prod-lon1-network`: `10.10.1.0/24` (production in London)
- `stage-lon1-network`: `10.10.2.0/24` (staging in London)
- `dev-nyc1-network`: `10.20.1.0/24` (dev in New York)

**CIDR Planning Best Practice:** Use a hierarchical schema to prevent future conflicts:
- `10.A.B.0/24` where:
  - `A` = Region ID (e.g., 10 for LON1, 20 for NYC1, 30 for FRA1)
  - `B` = Environment ID (e.g., 1 for prod, 2 for staging, 3 for dev)

This ensures that if you ever need to connect networks via VPN (self-managed), you won't have overlapping CIDRs.

---

### Permanent Network Assignment: The Migration Trap

**Critical Constraint:** Once a resource (Kubernetes cluster or compute instance) is assigned to a network, it **cannot be moved** to a different network. The only migration path is **destroy and recreate**.

**IaC Implications:** If you change a `civo_kubernetes_cluster` resource's `network_id` in Terraform or Pulumi, the tool will plan a **replacement operation**:
1. Destroy the existing cluster (data loss if not backed up)
2. Create a new cluster in the new network

This is not an update—it's a destructive, data-loss-incurring operation. Always run `terraform plan` or `pulumi preview` before applying changes to network assignments.

**Best Practice:** Treat network assignment as a foundational, "create-once" decision. Design your CIDR scheme carefully upfront. If you need to change networks in production, plan for:
- Full data backup
- DNS cutover
- Application migration windows

---

### Regional Constraint: Networks Are Not Global

**Key Constraint:** Civo networks are **region-specific**. A network in `LON1` (London) is not visible in `NYC1` (New York). This is similar to AWS and Azure, not GCP (which has global VPCs with regional subnets).

**Multi-Region Pattern:** If you deploy across multiple regions, create separate networks per region:
- `prod-lon1-network` in `LON1`
- `prod-nyc1-network` in `NYC1`

There is no native inter-region peering. If workloads in different regions need to communicate, they must use public IPs (over the internet, secured with TLS) or self-managed VPNs (WireGuard, Tailscale).

---

### Two-Tier Security Model: Firewalls + Network Policies

Civo provides **two independent security layers**:

1. **Layer 1: Platform Firewall (`civo_firewall`)**
   - Controls **North-South traffic** (traffic entering/leaving the network)
   - Stateful firewall attached to a `civo_network`
   - Default-deny: All traffic blocked unless you add explicit allow rules
   - Use case: Open port 6443 for Kubernetes API access, 80/443 for web traffic to load balancers, 22 for SSH to bastion hosts

2. **Layer 2: Kubernetes CNI Network Policy**
   - Controls **East-West traffic** (pod-to-pod communication inside the cluster)
   - Not a Civo resource—standard Kubernetes `NetworkPolicy` objects enforced by the CNI (Cilium or Flannel)
   - Use case: Zero-trust security. Deny all cross-namespace traffic by default, then explicitly allow frontend pods to talk to backend pods on specific ports.

**Best Practice:** Use **both layers**. The Civo firewall protects the network perimeter from the public internet. CNI policies enforce fine-grained, zero-trust security within the cluster.

---

## Production Essentials: Best Practices and Anti-Patterns

### Best Practices

1. **Never Use the Default Network**
   - Create dedicated networks for each environment (dev, staging, prod) in each region
   - Tag networks for cost allocation and organization

2. **Plan CIDR Blocks Upfront**
   - Use a hierarchical schema (`10.A.B.0/24`) to prevent future conflicts
   - Document your CIDR allocation in version control

3. **Use IaC from Day One**
   - Even for dev environments. Manual networks lead to drift and misconfiguration.
   - Store Terraform/Pulumi state in Civo Object Store (S3-compatible) for a fully self-contained Civo stack

4. **Implement Two-Tier Security**
   - Define Civo firewalls for perimeter security (North-South)
   - Use Kubernetes Network Policies for intra-cluster security (East-West)

5. **Test Network Changes in Preview Mode**
   - Always run `terraform plan` or `pulumi preview` before applying
   - Pay special attention to any resource marked for "replacement"—this is destructive for clusters and instances

---

### Anti-Patterns to Avoid

1. **Deploying Production to the Default Network**
   - No isolation, no predictability, no security boundary

2. **Copy-Pasting CIDR Blocks**
   - Forgetting to change `cidr_v4` when reusing configs for different environments leads to overlapping address spaces
   - This breaks routing if you ever connect networks via VPN

3. **Assuming Network Changes Are Non-Destructive**
   - Changing `network_id` on a cluster or instance triggers a replacement (destroy + recreate)
   - Always preview changes and plan migration windows

4. **Mixing Security Layers**
   - Don't assume the Civo firewall controls pod-to-pod traffic (it doesn't)
   - Don't forget to configure CNI policies (the Civo firewall only protects the perimeter)

5. **Ignoring the /24 Limit**
   - Don't try to build a monolithic network with thousands of instances
   - Embrace the "many small networks" pattern instead

---

## The Planton Choice: Abstraction with Pragmatism

Planton abstracts Civo VPC provisioning behind a clean, protobuf-defined API (`CivoVpc`). This provides a consistent interface across clouds while respecting Civo's unique characteristics.

### What We Abstract

**The `CivoVpcSpec` includes:**

- **`civo_credential_id`** (required): Reference to the Civo API credential for authentication
- **`network_name`** (required): DNS-friendly label for the network (lowercase alphanumeric + hyphens)
- **`region`** (required): Civo region code (e.g., `LON1`, `NYC1`, `FRA1`)
- **`ip_range_cidr`** (optional): IPv4 CIDR block (max /24). If omitted, Civo auto-allocates from available address pools.
- **`is_default_for_region`** (optional): Whether this network should be the default for the region. Only one default per region is allowed. Default: `false`.
- **`description`** (optional): Human-readable description (max 100 characters)

This follows the **80/20 principle**: 80% of users need only these fields. The remaining 20% (advanced use cases like VLAN integration for private cloud) are omitted to avoid API clutter.

### Default Choices

- **CIDR Auto-Allocation:** If `ip_range_cidr` is empty, we let Civo auto-allocate. This simplifies dev/test workflows. For production, always specify an explicit CIDR.
- **Not Default Network:** We default `is_default_for_region` to `false` to prevent accidental misuse of the default network anti-pattern.
- **No VLAN Fields:** Civo's API includes VLAN-related fields (`vlan_id`, `vlan_physical_interface`, etc.) for private cloud deployments (CivoStack Enterprise). We omit these from the public cloud API to avoid confusion.

### Under the Hood: Pulumi (Go)

Planton uses **Pulumi (Go)** for Civo VPC provisioning. Why?

- **Language Consistency:** Our broader multi-cloud orchestration is Go-based. Pulumi's Go SDK integrates naturally.
- **Equivalent Coverage:** Pulumi's Civo provider (bridged from Terraform) supports all network operations we need.
- **Programming Model:** Pulumi's code-based approach makes it easier to add conditional logic, multi-network strategies, or custom integrations in the future.

That said, Terraform would work equally well. The choice is an implementation detail—the protobuf API remains the same.

### Stack Outputs: What You Get Back

After creating a Civo VPC, Planton captures the following outputs in `CivoVpcStackOutputs`:

- **`network_id`**: The unique ID of the created network. This is the critical linker value—used as `network_id` in Kubernetes clusters, compute instances, and firewalls.
- **`cidr_block`**: The actual IPv4 CIDR block (auto-allocated if you didn't specify one).
- **`created_at_rfc3339`**: Timestamp when the network was created (useful for auditing and tracking).

These outputs are stored in the resource's `status` field and can be referenced by dependent resources.

---

## Configuration Examples: Dev, Staging, Production

### Development: Auto-Allocated CIDR

**Use Case:** Quick dev network for testing. Let Civo handle CIDR allocation.

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoVpc
metadata:
  name: dev-test-network
spec:
  civo_credential_id: civo-cred-123
  network_name: dev-test-network
  region: LON1
  description: "Development test network"
```

**Rationale:**
- No `ip_range_cidr` specified—Civo auto-allocates from available address pools
- Simplifies dev workflows (no CIDR planning required)
- Not default network (`is_default_for_region` defaults to `false`)

---

### Staging: Explicit CIDR for Predictability

**Use Case:** Staging environment with planned CIDR for future VPN integration.

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoVpc
metadata:
  name: staging-main-network
spec:
  civo_credential_id: civo-cred-123
  network_name: staging-main-network
  region: NYC1
  ip_range_cidr: "10.20.2.0/24"
  description: "Staging environment network (NYC1)"
```

**Rationale:**
- Explicit CIDR (`10.20.2.0/24`) following hierarchical schema:
  - `10` = Civo namespace
  - `20` = NYC1 region ID
  - `2` = Staging environment ID
- Prevents conflicts if we later add production (`10.20.1.0/24`) or dev (`10.20.3.0/24`) in the same region
- Makes it safe to connect networks via self-managed VPN without overlapping address spaces

---

### Production: Multi-Region with Explicit CIDRs

**Use Case:** Production networks in two regions (LON1 and FRA1), ready for global traffic distribution.

```yaml
# Production - London
---
apiVersion: civo.planton.dev/v1
kind: CivoVpc
metadata:
  name: prod-lon1-network
spec:
  civo_credential_id: civo-cred-123
  network_name: prod-lon1-network
  region: LON1
  ip_range_cidr: "10.10.1.0/24"
  description: "Production network (London)"

# Production - Frankfurt
---
apiVersion: civo.planton.dev/v1
kind: CivoVpc
metadata:
  name: prod-fra1-network
spec:
  civo_credential_id: civo-cred-123
  network_name: prod-fra1-network
  region: FRA1
  ip_range_cidr: "10.30.1.0/24"
  description: "Production network (Frankfurt)"
```

**Rationale:**
- Separate networks per region (Civo's constraint)
- Non-overlapping CIDRs using hierarchical schema:
  - LON1: `10.10.1.0/24` (region ID 10)
  - FRA1: `10.30.1.0/24` (region ID 30)
- Both use environment ID `1` for production
- Enables future multi-region connectivity via self-managed VPN or mesh networking (Tailscale, WireGuard)

---

## Key Takeaways

1. **Civo VPCs are simple, regional, and purpose-built for isolation.** No subnets, no peering, no transit gateways. Just a flat /24 network that provides true Layer 3 isolation for Kubernetes clusters, compute instances, and load balancers.

2. **Manual management is an anti-pattern.** Use IaC (Terraform or Pulumi) from day one, even for dev environments. Both are officially supported by Civo and production-ready.

3. **The /24 limit is a feature, not a bug.** Embrace "many small networks" instead of one monolithic network. Plan your CIDR scheme upfront using a hierarchical schema to prevent future conflicts.

4. **Network assignment is permanent.** Changing a resource's `network_id` in IaC triggers a destructive replacement (destroy + recreate). Always preview changes and plan migration windows.

5. **There are no native snapshots, peering, or multi-region networking.** Civo's philosophy is simplicity. Advanced connectivity (VPN, mesh networking) is self-managed, but the lack of data egress fees makes this economically viable.

6. **Two-tier security is essential.** Use Civo firewalls for perimeter security (North-South traffic) and Kubernetes Network Policies for intra-cluster security (East-West traffic).

7. **Planton abstracts the API** into a clean protobuf spec, making multi-cloud deployments consistent while respecting Civo's unique characteristics. The 80/20 config is credential, name, region, and optional CIDR.

---

## Conclusion

Civo's networking philosophy is a deliberate rejection of hyperscaler complexity. Instead of offering every advanced feature under the sun—VPC peering, transit gateways, PrivateLink, global routing—Civo focuses on what the majority of workloads actually need: **strict isolation, transparent pricing, and zero egress fees**.

The /24 CIDR limit forces a design pattern of many small, single-purpose networks—exactly the isolation strategy security teams recommend (separate dev from staging, separate customers from each other). The permanent network assignment constraint encourages thoughtful upfront planning instead of ad-hoc, drift-prone reconfigurations. The lack of subnets and availability zones eliminates an entire class of operational complexity.

What you lose in flexibility, you gain in operational simplicity and cost predictability. For the 80% of businesses that don't need multi-region VPC peering or sub-millisecond inter-AZ latency guarantees, Civo's model is refreshingly straightforward.

Planton makes this even simpler. Define your network once in a protobuf spec, let our control plane handle the Pulumi/Terraform orchestration, and move on to building your application. The complexity is abstracted, but the simplicity—and the cost savings—remain.

---

## Further Reading

- **Civo VPC Documentation:** [Civo Docs - Private Networks](https://www.civo.com/docs/networking/private-networks)
- **Why Create Multiple Networks?** [Civo Learn - Multi-Network Strategies](https://www.civo.com/learn/why-create-multiple-networks)
- **Terraform Civo Provider:** [GitHub - civo/terraform-provider-civo](https://github.com/civo/terraform-provider-civo)
- **Pulumi Civo Provider:** [Pulumi Registry - Civo](https://www.pulumi.com/registry/packages/civo/)
- **Civo Firewalls:** [Civo Docs - Firewalls](https://www.civo.com/docs/networking/firewalls)
- **Kubernetes Network Policies with Cilium:** [Civo Learn - Network Policies](https://www.civo.com/learn/network-policies-with-cilium)
- **Civo API Reference:** [Civo API - Networks](https://www.civo.com/api/networks)

---

**Bottom Line:** Civo VPCs are simple, regional, /24-limited networks that provide true isolation for cloud-native workloads. Manage them with Terraform or Pulumi (never manually), plan your CIDR scheme upfront, and embrace the "many small networks" pattern. Planton abstracts this into a protobuf API that works consistently across clouds while respecting Civo's deliberate simplicity.


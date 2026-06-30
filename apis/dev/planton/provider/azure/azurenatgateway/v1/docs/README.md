# Azure NAT Gateway Deployment: From SNAT Chaos to Predictable Egress

## Introduction: The Hidden Bottleneck in Cloud Networking

For years, developers deploying private workloads on Azure—AKS clusters, VM scale sets, App Services—would experience a puzzling, intermittent failure: outbound connections would suddenly fail. Database queries would hang. API calls would timeout. Container image pulls would stall. The root cause was often invisible until production broke under load: **SNAT port exhaustion**.

Azure's default outbound connectivity model relies on pre-allocating a fixed number of Source Network Address Translation (SNAT) ports to each VM instance. This pre-allocation is fundamentally inefficient. In a dynamic, high-scale workload, some instances exhaust their allocated ports while others have thousands sitting unused. The result is connection failures that are difficult to diagnose and even harder to solve after the fact.

Azure NAT Gateway represents a paradigm shift away from this broken model. It's a fully managed, software-defined networking service that provides **dynamic SNAT**, creating a shared, on-demand pool of ports for all resources in a subnet. A single Standard SKU public IP provides 64,512 SNAT ports. A NAT Gateway can use up to 16 public IP addresses, yielding over 1 million ports—a scale that transforms SNAT exhaustion from an inevitable production failure into a solved problem.

This document explores the landscape of outbound connectivity methods for Azure, explains why NAT Gateway has become the production-standard solution, and details how Planton provides a declarative API to deploy and manage NAT Gateways across your infrastructure.

## The Outbound Connectivity Spectrum

Not all outbound connectivity methods are created equal. Understanding the evolution from basic approaches to production-ready solutions helps clarify when and why NAT Gateway is the right choice.

### Level 0: Default Implicit SNAT (The Anti-Pattern)

**What it is:** Azure provides implicit outbound connectivity for private VMs. No explicit configuration required.

**Why it's tempting:** Zero setup. Just deploy a VM or AKS cluster in a private subnet, and it "just works."

**Why it fails in production:** This method uses Azure Load Balancer's implicit SNAT with extremely limited port allocation. For VM scale sets and AKS node pools, this defaults to as few as 64-256 ports per instance. High-churn workloads (think microservices making thousands of API calls) exhaust this allocation in seconds.

**The verdict:** Acceptable only for development or proof-of-concept environments. **Never use this for production workloads.**

### Level 1: Load Balancer Outbound Rules (The Band-Aid)

**What it is:** Explicitly configure a Standard Load Balancer with outbound rules to provide SNAT for backend pool members.

**What it solves:** Gives you control over the number of frontend IPs and ports allocated per instance. Better than implicit SNAT.

**What it doesn't solve:** Still uses pre-allocation. You might configure 1,024 ports per instance, but if one instance is idle and another is under load, the busy instance still can't access the idle one's unused ports. You're also coupling your inbound load balancing configuration with your outbound egress strategy.

**The verdict:** A partial improvement, but fundamentally still fighting the pre-allocation problem. Outbound rules are complex to configure correctly and don't scale as cleanly as NAT Gateway.

### Level 2: Instance-Level Public IPs (The Edge Case)

**What it is:** Assign a public IP directly to a VM's network interface.

**What it solves:** That VM gets the full 64,000 ephemeral port range. No SNAT exhaustion for that single instance.

**What it doesn't solve:** Doesn't scale. You can't assign instance-level public IPs to AKS node pools managed by virtual machine scale sets. Your egress IP changes every time the VM is redeployed. You lose centralized egress control and create a firewall rule management nightmare.

**The verdict:** Useful only for very specific single-VM scenarios (e.g., a bastion host). Not a solution for cluster or fleet workloads.

### Level 3: Azure NAT Gateway (The Production Solution)

**What it is:** A fully managed, software-defined NAT service that operates at the subnet level. Once associated with a subnet, **all resources in that subnet automatically use the NAT Gateway for outbound internet traffic.**

**What it solves:**

- **Dynamic SNAT:** No more pre-allocation. Ports are available on-demand to any instance that needs them from a shared pool.
- **Massive scale:** Each public IP provides 64,512 ports. Associate up to 16 IPs for over 1 million ports per subnet.
- **Predictable egress:** Your outbound traffic always uses the same static public IP or IP prefix, simplifying firewall allow-listing.
- **Automatic precedence:** NAT Gateway takes precedence over all other outbound methods (Load Balancer rules, instance-level IPs, even Azure Firewall unless explicitly overridden). This makes the egress path explicit and eliminates ambiguity.

**What it costs:** NAT Gateway has a two-component pricing model: a fixed hourly charge (~$0.045/hour, ~$33/month) plus a variable charge for data processed (~$0.045/GB). The key cost optimization strategy is using **Private Link or Service Endpoints** to route traffic to Azure PaaS services (Storage, SQL Database, Key Vault) over the Azure backbone, bypassing the NAT Gateway entirely and reserving its capacity for true public internet egress.

**The verdict:** This is the **production-standard solution** for private workloads needing internet access. Microsoft recommends it as the default for AKS clusters, VM scale sets, and any scenario where SNAT port exhaustion is a risk.

## Public IP vs. Public IP Prefix: Scaling Your Egress

NAT Gateway requires at least one Standard SKU public IP address or public IP prefix to function. The choice between individual IPs and prefixes is strategic.

### When to Use Individual Public IPs

**Best for:** Development, testing, or small-scale production workloads with predictable, low-volume egress.

**Characteristics:**
- One IP = 64,512 SNAT ports (sufficient for many workloads)
- Simple to set up
- You can add more individual IPs later, up to 16 total

**Example use case:** A staging AKS cluster with a dozen nodes running batch jobs.

### When to Use Public IP Prefixes (Production Pattern)

**Best for:** Any production workload requiring scale, IP allow-listing, or proactive capacity planning.

**Characteristics:**
- **Scalability:** A /28 prefix provides 16 IPs (1,032,192 SNAT ports) from day one. No need to add IPs incrementally as you scale.
- **IP Whitelisting:** A prefix provides a contiguous, static, predictable range of IPs. This is often a hard requirement when third-party partners or compliance controls need to add your egress IPs to firewall allow-lists.
- **Operational simplicity:** One prefix resource vs. managing 16 individual IP resources.

**Recommendation:** Use public IP prefixes for all production deployments. The operational benefits far outweigh the minimal additional cost.

## Availability Zones and High Availability

Azure NAT Gateway is inherently resilient—it's a software-defined service with multiple fault domains. But for mission-critical workloads, you can use **Availability Zones** to add an explicit layer of isolation.

### Option 1: "No-Zone" Deployment (Default)

When you don't specify an availability zone, Azure automatically places the NAT Gateway in a single zone (not visible to you). You can pair this with **zone-redundant public IPs**, which is a common and simple HA pattern for most workloads.

**Best for:** Standard production workloads where automatic zone placement is acceptable.

### Option 2: "Zonal Stacks" (Maximum Isolation)

For the highest level of control and fault isolation, deploy **multiple NAT Gateways, one per availability zone**, each associated with a zone-specific subnet.

**Example architecture:**
- `nat-gateway-zone-1` (pinned to Zone 1) → `subnet-zone-1`
- `nat-gateway-zone-2` (pinned to Zone 2) → `subnet-zone-2`
- `nat-gateway-zone-3` (pinned to Zone 3) → `subnet-zone-3`

**Benefit:** A catastrophic datapath failure in Zone 1 only affects resources in `subnet-zone-1`. Zones 2 and 3 continue operating independently.

**Best for:** Mission-critical, zone-aware AKS clusters or VM scale sets where you need guaranteed isolation.

## The Idle Timeout Problem (And How to Solve It)

One of the most common NAT Gateway misconfigurations is leaving the default **TCP idle timeout** at 4 minutes.

### What Happens

NAT Gateway maintains a SNAT port mapping for each active TCP connection. If no packets are sent for longer than the idle timeout, the gateway silently drops the mapping. When the client attempts to send the next packet, it goes into a void—the connection appears to hang until a higher-level TCP timeout occurs (often 60+ seconds).

This creates mysterious, intermittent failures for:
- Long-running SSH sessions
- Database connection pools with idle connections
- HTTP/2 persistent connections
- Any long-lived but low-traffic flows

### The Solutions

**Option 1: Increase the Idle Timeout**

Set `idle_timeout_in_minutes` to a higher value (10, 30, or 60 minutes). This is a simple, effective fix for applications known to have long-lived idle connections.

**Planton Default:** The Planton API defaults to **10 minutes** (not the Azure default of 4), providing a safer, more production-ready baseline.

**Option 2: Use TCP Keepalives (Best Practice)**

Configure your application or host OS to send TCP keepalive packets at an interval shorter than the timeout (e.g., every 3 minutes). This keeps the NAT Gateway's state table entry "active" and prevents the timer from expiring.

**For UDP:** The idle timeout for UDP is fixed at 4 minutes and cannot be changed. Long-running UDP flows must use application-level keepalives.

## Integration with Azure Kubernetes Service (AKS)

NAT Gateway is the **Microsoft-recommended egress solution** for AKS clusters, especially for private clusters or those needing predictable outbound IPs.

### Two Deployment Patterns

**Pattern 1: `managedNatGateway` (Simple, Less Control)**

AKS provisions and manages the NAT Gateway in the cluster's node resource group.

```bash
az aks create ... --outbound-type managedNatGateway
```

**Best for:** Quick setup, development, or simple production scenarios where you don't need fine-grained control over the NAT Gateway configuration.

**Pattern 2: `userAssignedNatGateway` (Enterprise Standard)**

You (or your platform team) pre-deploy the VNet, subnet, and NAT Gateway using IaC. The AKS cluster is then deployed into that pre-configured subnet.

```bash
# 1. Deploy VNet, subnet, and NAT Gateway (using Terraform, Pulumi, or Planton)
# 2. Associate NAT Gateway with AKS node subnet
# 3. Deploy AKS cluster
az aks create ... --outbound-type userAssignedNatGateway --vnet-subnet-id <subnet-id>
```

**Best for:** Enterprise deployments where the network team manages VNet infrastructure separately from the cluster team. This model correctly separates lifecycle concerns: the network infrastructure is long-lived, while clusters are ephemeral.

### Private Clusters and API Server Traffic

For private AKS clusters, the API server traffic also respects the `outboundType`. If using `userAssignedNatGateway`, control plane traffic will egress through the NAT Gateway to the public API server endpoint. To keep this traffic private, you must use **API Server VNet Integration** or a private endpoint.

## Monitoring NAT Gateway: The Essential Metrics

Effective monitoring is critical. Unlike traditional infrastructure, there is **no metric for "SNAT port usage percentage."** You must infer health and utilization from several key metrics.

### Critical Metrics (Azure Monitor)

| Metric | Type | What to Monitor | Alert Threshold |
|--------|------|----------------|----------------|
| **DatapathAvailability** | Gauge (Average) | Health of the NAT Gateway datapath | Alert if < 99% |
| **SNATConnectionCount** | Sum | Total active SNAT connections | Monitor against max capacity (Total IPs × 64,512) |
| **PacketDropCount** | Sum | Dropped packets | Alert if > 0 (indicates SNAT exhaustion or datapath failure) |
| **ByteCount** | Sum | Data processed (for cost analysis) | Track for billing correlation |
| **TotalConnectionCount** | Sum | New connections per second | Monitor for traffic churn patterns |

### What You Can't Measure Directly

- **SNAT port usage percentage:** Not available. Use `SNATConnectionCount` as a proxy.
- **Per-VM connection distribution:** NAT Gateway operates at the subnet level. You see aggregate metrics, not per-instance breakdowns.

## Common Anti-Patterns to Avoid

**❌ Anti-Pattern 1: No NAT Gateway for Private Clusters**

Deploying private AKS clusters or VM scale sets with default implicit SNAT and hoping for the best. This leads to inevitable SNAT port exhaustion in production.

**✅ Solution:** Always deploy NAT Gateway for production workloads needing public internet egress.

---

**❌ Anti-Pattern 2: Single Public IP for High-Scale Workloads**

Provisioning a NAT Gateway with only one public IP for a massive AKS cluster (e.g., 100+ nodes, high-churn microservices).

**✅ Solution:** Use a Public IP Prefix (e.g., /28 for 16 IPs, 1M+ ports) to proactively scale.

---

**❌ Anti-Pattern 3: Ignoring Idle Timeout**

Using the 4-minute default for applications with long-lived database connections, leading to mysterious connection hangs.

**✅ Solution:** Set `idle_timeout_minutes` to 10+ or configure TCP keepalives.

---

**❌ Anti-Pattern 4: Processing Azure PaaS Traffic Through NAT Gateway**

Routing all traffic (including traffic to Azure Storage, SQL Database, Key Vault) through the NAT Gateway, incurring unnecessary data processing charges.

**✅ Solution:** Use **Private Link** or **Service Endpoints** to route Azure PaaS traffic over the Azure backbone, bypassing the NAT Gateway and eliminating data processing fees for internal-Azure traffic.

## The Planton Approach

Planton provides a declarative, protobuf-based API for deploying Azure NAT Gateways. The design philosophy prioritizes production-ready defaults and simplicity for the 80% use case while exposing advanced configuration for the 20% edge cases.

### Production-Ready Defaults

**Idle Timeout:** The Planton API defaults `idle_timeout_minutes` to **10 minutes** (not the Azure default of 4). This provides a safer baseline for applications with persistent connections, reducing the likelihood of mysterious timeout issues.

**SKU:** The API hardcodes the NAT Gateway SKU to `Standard` (the only supported SKU). There's no reason to expose this as a configuration option—it only adds confusion.

### The Subnet Association Model

Azure NAT Gateway operates at the **subnet level**. Once associated with a subnet, all resources in that subnet automatically use the gateway for outbound traffic.

In the Azure API, subnet association is a **property of the subnet resource**, not the NAT Gateway. The Planton API reflects this reality:

- The `AzureNatGateway` resource creates and manages the NAT Gateway itself
- Subnet association is handled via the `subnet_id` field, which references an existing Azure subnet

This model mirrors the native Azure API structure and avoids architectural conflicts where multiple controllers might fight for control of subnet configuration.

### Public IP Prefix for Scale

For production deployments, use the `public_ip_prefix_length` field to provision a Public IP Prefix instead of individual IPs:

- `/28` = 16 IPs = 1,032,192 SNAT ports (recommended for large-scale production)
- `/29` = 8 IPs = 516,096 SNAT ports
- `/30` = 4 IPs = 258,048 SNAT ports
- `/31` = 2 IPs = 129,024 SNAT ports

### Example: Development/Staging

A simple NAT Gateway for a dev AKS cluster:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureNatGateway
metadata:
  name: dev-aks-nat-gateway
spec:
  subnetId: ${ref:dev-aks-vpc.status.outputs.nodes_subnet_id}
  idle_timeout_minutes: 10
  tags:
    environment: dev
    team: platform
```

This configuration:
- Associates with an existing AKS node subnet
- Uses the default 10-minute idle timeout
- Provisions a single public IP (suitable for dev/staging)

### Example: Production with HA and Scale

A production-grade NAT Gateway with IP prefix for scale and whitelisting:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureNatGateway
metadata:
  name: prod-aks-nat-gateway-z1
spec:
  subnetId: ${ref:prod-aks-vpc-z1.status.outputs.nodes_subnet_id}
  idle_timeout_minutes: 30
  public_ip_prefix_length: 28  # 16 IPs, 1M+ ports
  tags:
    environment: production
    availability_zone: "1"
    cost_center: platform-engineering
```

This configuration:
- Uses a /28 public IP prefix (16 IPs, 1M+ SNAT ports)
- Sets a 30-minute idle timeout for long-running connections
- Designed as part of a "zonal stack" HA pattern (one gateway per zone)

## Conclusion: From Bottleneck to Foundation

Azure NAT Gateway represents a shift from viewing outbound connectivity as a default, implicit behavior to treating it as a first-class infrastructure component that requires deliberate design.

The days of diagnosing SNAT port exhaustion at 3 AM in production are over—if you architect correctly. NAT Gateway's dynamic SNAT model, massive port capacity, and predictable egress behavior make it the production standard for any private workload needing internet access.

Planton's declarative API abstracts the complexity of resource associations and provides production-ready defaults (like a sensible idle timeout) while still exposing the full power of Azure NAT Gateway for advanced scenarios.

When you deploy your next AKS cluster, VM scale set, or private application, don't rely on default SNAT. Make your egress path explicit. Make it scalable. Make it predictable.

Deploy a NAT Gateway.


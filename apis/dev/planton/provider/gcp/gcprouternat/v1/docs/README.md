# GCP Cloud Router NAT Deployment: From Manual Console Clicks to Declarative Infrastructure

## Introduction

In the early days of cloud networking, the default assumption was simple: if a VM needs internet access, give it a public IP. This straightforward approach worked for small deployments, but as cloud architectures matured and security requirements tightened, the shortcomings became apparent. Hundreds or thousands of VMs, each with their own public IP, created a sprawling attack surface and made IP allowlisting a nightmare for partners who needed to trust your egress traffic.

**Google Cloud Router NAT** (commonly known as **Cloud NAT**) represents a fundamental shift in how we think about egress connectivity. Rather than treating public IPs as a per-instance commodity, Cloud NAT consolidates outbound internet access through a managed, regional NAT gateway. Private instances—GKE nodes, Compute Engine VMs, Cloud SQL databases—can reach the internet without ever exposing themselves to unsolicited inbound traffic. The service operates at Google's software-defined networking layer (Andromeda), making it distributed, scalable, and remarkably transparent to your workloads.

What makes Cloud NAT particularly elegant is its **proxy-less architecture**. Unlike traditional NAT implementations that funnel traffic through dedicated appliances (creating bottlenecks and single points of failure), Cloud NAT distributes translation across Google's infrastructure. Each VM's traffic is translated at the host level, maintaining full bandwidth and minimal latency. There's no NAT VM to maintain, no throughput ceiling to hit, and no availability concerns—the service is inherently highly available within a region.

This document explores the landscape of Cloud Router NAT deployment methods, examines how approaches have evolved from manual provisioning to infrastructure-as-code, and explains why Planton has chosen specific defaults for its protobuf-based API design.

## The Deployment Maturity Spectrum

Cloud Router NAT deployment methods fall along a spectrum from manual, error-prone approaches to fully declarative, version-controlled infrastructure.

### Level 0: The Manual Console Workflow

**The approach:** Navigate to the GCP Console, click through the Cloud NAT creation wizard, fill in form fields for router name, NAT gateway name, IP allocation mode, and subnet coverage. Repeat for each environment and region.

**What it solves:** Gets you a working NAT gateway quickly for exploration or proof-of-concept work.

**What it doesn't solve:** Manual console work is inherently non-reproducible. When you need to set up NAT in a new region or replicate your production configuration to staging, you're clicking through the same forms again, hoping you remember all the settings. The console provides no audit trail of who changed what, and it's easy to overlook critical configuration like logging or subnet coverage. One common pitfall: forgetting that Cloud NAT is **regional**, so setting it up in `us-central1` does nothing for your resources in `europe-west1`.

**Verdict:** Use the console for initial exploration and learning, but don't build production infrastructure this way. The moment you need to answer "What NAT settings are in production?" or "Who changed the NAT IP allocation last week?", you'll wish you had infrastructure-as-code.

### Level 1: Shell Scripts and gcloud CLI

**The approach:** Automate NAT provisioning using `gcloud compute routers create` and `gcloud compute routers nats create` commands wrapped in bash scripts.

```bash
gcloud compute routers create my-router \
  --network=my-vpc --region=us-central1

gcloud compute routers nats create my-nat \
  --router=my-router --region=us-central1 \
  --nat-ip-allocation-option=AUTO_ONLY \
  --source-subnet-ip-ranges-to-nat=ALL_SUBNETWORKS_ALL_IP_RANGES \
  --enable-logging --log-filter=ERRORS_ONLY
```

**What it solves:** Shell scripts provide repeatability. The same script can provision NAT in multiple regions or environments, and you can version control the scripts. This is a significant improvement over manual console work—at least your configuration is documented in code.

**What it doesn't solve:** Shell scripts are imperative, not declarative. They tell Google *how* to create resources but don't describe the *desired state*. If someone manually changes a NAT setting via the console, your script has no way to detect drift. Re-running the script might fail if resources already exist, requiring custom logic to handle updates versus creates. Error handling becomes complex, and managing dependencies (ensure the VPC exists before creating the router, ensure the router exists before creating the NAT) requires careful scripting.

**Verdict:** Shell scripts are useful for simple automation or one-off migrations, but they lack the state management and drift detection that production infrastructure demands. You'll spend more time maintaining the scripts than the infrastructure.

### Level 2: Configuration Management (Ansible, Chef)

**The approach:** Use configuration management tools to orchestrate Cloud NAT provisioning, either by calling `gcloud` commands or using GCP modules.

**What it solves:** Configuration management tools bring better orchestration, templating, and inventory management to infrastructure provisioning. You can define variables for different environments and use roles to organize NAT configuration alongside other infrastructure tasks.

**What it doesn't solve:** As of recent versions, Ansible's `google.cloud` collection lacks a dedicated first-class module for Cloud Router NAT. Teams typically work around this by using the `command` or `shell` module to execute `gcloud` CLI commands, which brings us back to the imperative shell script problem. The tool can orchestrate, but it's not truly managing the resource declaratively.

**Verdict:** Configuration management tools can work for NAT provisioning if you're already standardized on them for other operations, but they're not the natural fit for cloud infrastructure. Their imperative nature and gaps in GCP resource coverage make them suboptimal compared to infrastructure-as-code tools designed specifically for cloud APIs.

### Level 3: Infrastructure-as-Code (Terraform, Pulumi, OpenTofu)

**The approach:** Define Cloud Router and Cloud NAT as declarative resources using tools like Terraform, Pulumi, or OpenTofu. Describe the desired state, let the tool handle create/update/delete operations, and store state to detect configuration drift.

**Terraform example:**

```hcl
resource "google_compute_router" "nat_router" {
  name    = "nat-router-uscentral1"
  network = google_compute_network.vpc.id
  region  = "us-central1"
}

resource "google_compute_router_nat" "nat" {
  name                               = "nat-uscentral1"
  router                             = google_compute_router.nat_router.name
  region                             = google_compute_router.nat_router.region
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
  
  log_config {
    enable = true
    filter = "ERRORS_ONLY"
  }
}
```

**What it solves:** Infrastructure-as-code tools provide the holy trinity of production-ready infrastructure management:

1. **Declarative desired state**: You specify *what* you want, not *how* to create it.
2. **State management**: The tool knows what currently exists and can detect drift or changes.
3. **Plan/apply workflow**: Preview changes before applying them, reducing the risk of surprises.

Terraform and Pulumi both provide comprehensive support for Cloud Router NAT with full coverage of all configuration options—IP allocation modes, subnet coverage, logging, port allocation, and more. The resource schema maps one-to-one with the GCP API, and both tools handle dependencies gracefully (ensuring the VPC and router exist before creating the NAT).

**What it doesn't solve:** Infrastructure-as-code requires learning the tool's patterns and investing in proper state management (where to store the state file, how to handle concurrency, etc.). For Terraform specifically, the HCL language has a learning curve, though it's well-documented. Pulumi's code-centric approach (TypeScript, Python, Go) may be easier for some teams but introduces its own complexity around SDK usage.

**Verdict:** This is the production-ready approach. Terraform has the most mature ecosystem and is the de facto standard for GCP infrastructure provisioning, with Pulumi as a strong alternative for teams who prefer general-purpose programming languages. OpenTofu (the Terraform fork) is functionally identical and appeals to teams prioritizing open-source licensing.

### Level 4: Kubernetes-Native (Crossplane, Config Connector)

**The approach:** Manage Cloud Router NAT as a Kubernetes Custom Resource using Crossplane or Google's Config Connector. Define your infrastructure in YAML, apply it to a Kubernetes cluster, and let the operator reconcile the desired state with GCP.

**Config Connector example:**

```yaml
apiVersion: compute.cnrm.cloud.google.com/v1beta1
kind: ComputeRouterNAT
metadata:
  name: nat-uscentral1
spec:
  routerRef:
    name: nat-router
  region: us-central1
  natIpAllocateOption: AUTO_ONLY
  sourceSubnetworkIpRangesToNat: ALL_SUBNETWORKS_ALL_IP_RANGES
  logConfig:
    enable: true
    filter: ERRORS_ONLY
```

**What it solves:** For organizations standardized on Kubernetes and GitOps workflows, managing infrastructure through Kubernetes CRDs provides a unified control plane. All your infrastructure—network, compute, data—lives in the same declarative format, applied through the same workflow. Config Connector is Google-supported and GA, while Crossplane provides multi-cloud abstractions if that's valuable.

**What it doesn't solve:** This approach requires a Kubernetes cluster to manage your infrastructure, which adds operational complexity. If your team isn't already deeply invested in Kubernetes-based workflows, requiring a cluster just to provision NAT gateways is overkill. There's also latency—changes applied to the cluster must be reconciled by the operator, which can take longer than direct API calls via Terraform.

**Verdict:** Excellent for platform engineering teams building Kubernetes-centric infrastructure platforms, particularly if you're already using Crossplane or Config Connector for other GCP resources. For most teams, Terraform/Pulumi's direct API approach is simpler and faster.

## The 80/20 Configuration Principle

Cloud Router NAT exposes dozens of configuration parameters, but the vast majority of production deployments use a small subset. Here's what actually matters:

### Essential Configuration (80% of Use Cases)

**1. Region and VPC Context**

Every Cloud NAT gateway is regional and attached to a Cloud Router in a specific VPC. This isn't optional—it's the foundation. Most teams provision one Cloud Router and one NAT gateway per region where they have private resources needing internet access.

**2. NAT IP Allocation: Auto vs. Manual**

The most consequential decision:

- **Auto-allocation** (`AUTO_ONLY`): Google automatically assigns and scales external IP addresses as needed. This is the recommended default—simple, hands-off, and cost-efficient. Use this unless you have a specific reason not to.

- **Manual allocation** (`MANUAL_ONLY`): You reserve specific static external IP addresses and assign them to the NAT. Use this **only** when you need deterministic egress IPs for external allowlisting. If a partner requires "whitelist these IPs to allow our traffic through your firewall," manual allocation is your answer.

**Real-world pattern**: Most deployments start with auto-allocation. When a business requirement emerges (vendor wants to whitelist IPs), you reserve 1-2 static IPs and switch to manual mode. The transition is straightforward but does briefly disrupt existing connections, so plan it during a maintenance window.

**3. Subnet Coverage: All vs. Specific**

Cloud NAT can cover:
- **All subnets in the region** (all IP ranges): The simplest and most common choice. Any subnet you create in that region automatically gets NAT—no config updates needed.
- **Specific subnets only**: Used when you want fine-grained control, perhaps to exclude certain subnets from internet access or to use separate NAT gateways for different workload tiers.

**Best practice**: Use "all subnets" unless you have a specific security requirement. It's future-proof—new subnets automatically inherit NAT coverage without manual intervention.

**4. Logging: Errors or All**

Cloud NAT can log:
- **Nothing** (default): No logging, which is fine for cost-conscious non-production environments.
- **Errors only**: Logs connection failures due to port exhaustion or other issues. This is the production minimum—you want to know if connections are being dropped.
- **All**: Logs every connection translation. Useful for security auditing or detailed troubleshooting, but generates significant log volume and costs.

**Recommendation**: Enable `ERRORS_ONLY` logging in production. The overhead is minimal, and it's invaluable when debugging connectivity issues.

### Advanced Configuration (20% of Use Cases)

**Port Allocation**

By default, each VM gets 64 source ports allocated. This is sufficient for most workloads—64 concurrent outbound connections per VM is plenty for typical microservices. If you have VMs that open thousands of concurrent connections (proxies, web scrapers, connection-intensive applications), you can:
- Increase `minPortsPerVm` (up to 65,536 per VM, though that's extreme)
- Enable `dynamicPortAllocation` to let VMs scale up to `maxPortsPerVm` as needed

**Endpoint-Independent Mapping**

Enabled by default for public NAT. It allows reusing a source port for connections to different destinations, conserving ports. Rarely disabled except in specialized security contexts.

**Custom Timeouts**

TCP established connections default to 1,200 seconds idle timeout, UDP to 30 seconds. These defaults work for almost everything. Adjust only if you have long-lived UDP flows or need aggressive timeout for security reasons.

**NAT Rules**

Advanced feature for routing specific destination IP ranges through specific NAT IPs. Very rarely used—most deployments have uniform egress requirements.

## Production Patterns and Integration

### Private GKE Clusters

Cloud NAT is essential for private GKE clusters. Without it, cluster nodes (which have no external IPs) can't pull container images from Docker Hub, reach package repositories, or call external APIs. The integration is seamless:

1. Create VPC network and subnets (including secondary ranges for pods)
2. Configure Cloud Router NAT covering "all subnets, all IP ranges"
3. Create the GKE private cluster in those subnets

The NAT automatically handles both node-level traffic (system updates, node initialization) and pod egress (application traffic). You don't need any in-cluster configuration—it's entirely at the network layer.

**Key detail**: Enable **Private Google Access** on your subnets so that traffic to Google APIs (GCR, Cloud Storage, Pub/Sub) stays internal and doesn't hairpin through NAT to the internet and back. Cloud NAT and Private Google Access work together: Private Google Access for Google services (efficient, no egress costs), Cloud NAT for general internet (external APIs, package repos, third-party services).

### Cloud SQL Private IP

When using Cloud SQL with private IP only (no public interface), the database instance needs Cloud NAT if it must reach the internet—for example, to import data from external URLs or replicate from an external MySQL server. Simply ensure the Cloud SQL's subnet is included in your NAT configuration, which happens automatically if using "all subnets" coverage.

### Multi-Region High Availability

Cloud NAT is inherently highly available **within a region**—it's distributed across Google's infrastructure with no single point of failure. However, it's also inherently **regional**, so multi-region HA means deploying one NAT gateway per region where you have workloads.

Pattern: If you have active-active or active-passive deployments across regions, configure Cloud NAT in each region. Each NAT will use its own external IPs, so if partners need to allowlist your egress, they'll need IPs from all regions you operate in.

### Cost Considerations

Cloud NAT pricing has three components:

1. **Gateway hourly cost**: ~$0.0014 per VM per hour, capped at $0.044/hour for 32+ VMs (~$32/month)
2. **Data processing**: $0.045 per GB of traffic through NAT (on top of normal egress costs)
3. **IP address cost**: $0.005 per hour per external IP (~$3.60/month per IP)

For a typical deployment with 20 VMs and 100 GB monthly egress, total NAT costs are roughly $30/month—insignificant compared to compute costs. The data processing fee becomes noticeable at high bandwidth (terabytes), at which point teams might consider alternatives like direct external IPs for bandwidth-heavy VMs.

**Cost optimization**: Use auto-allocation to avoid paying for unused static IPs, enable Private Google Access to keep Google API traffic internal, and monitor usage patterns to right-size port allocation.

## Planton's Approach

Planton's `GcpRouterNat` API follows the 80/20 principle, exposing the configuration that matters while using sensible defaults for the rest:

```protobuf
message GcpRouterNatSpec {
  // VPC network reference (required)
  StringValueOrRef vpc_self_link = 1;
  
  // Region for the Cloud Router and NAT (required)
  string region = 2;
  
  // Optional: Specific subnets to NAT (empty = all subnets)
  repeated StringValueOrRef subnetwork_self_links = 3;
  
  // Optional: Static external IPs (empty = auto-allocate)
  repeated StringValueOrRef nat_ip_names = 4;
}
```

This minimal schema captures the decisions that users actually need to make:

- **Where**: VPC and region (required context)
- **What**: Which subnets to cover (defaults to all)
- **How**: Auto or manual IP allocation (defaults to auto)

Behind the scenes, Planton's implementation applies production-ready defaults:
- Logging enabled at `ERRORS_ONLY` level
- Coverage of all IP ranges (primary and secondary) for listed subnets
- Default port allocation (64 per VM, sufficient for most workloads)
- Endpoint-independent mapping enabled

This design reflects a core philosophy: **the API should make the common case trivial and the complex case possible**. Users who need advanced features like dynamic port allocation or custom NAT rules can access them through the underlying IaC modules, but they're not cluttering the core API that 80% of users interact with.

## Comparative Landscape

When choosing a deployment method for Cloud Router NAT, the ecosystem breaks down as:

**Terraform**: The industry standard, with the most mature modules and largest community. Use this if you're provisioning infrastructure across clouds or want the richest ecosystem of pre-built modules.

**Pulumi**: Excellent alternative if your team prefers general-purpose programming languages (TypeScript, Python, Go) over HCL. Functionally equivalent to Terraform for GCP resources.

**OpenTofu**: The open-source fork of Terraform, identical in functionality but with an Apache 2.0 license. Use this if you prioritize open-source governance.

**Crossplane/Config Connector**: Best for teams deeply invested in Kubernetes-centric workflows and GitOps. Requires running operators in a Kubernetes cluster, adding operational overhead.

**gcloud CLI / shell scripts**: Useful for one-off tasks or migrations but lack the state management and drift detection that production infrastructure demands.

The decision largely comes down to organizational context: Are you already standardized on Terraform? Are you building on Kubernetes? What does your team find most maintainable? For most teams, Terraform or Pulumi provides the best balance of power, maturity, and operational simplicity.

## Conclusion

Cloud Router NAT represents a maturity milestone in cloud networking: the shift from treating public IPs as a per-instance necessity to consolidating egress through managed, scalable infrastructure. The service's proxy-less, distributed architecture means you get security benefits (reduced attack surface, controlled egress IPs) without sacrificing performance or introducing management complexity.

The deployment method you choose matters less than having **any** infrastructure-as-code approach. Whether you use Terraform, Pulumi, or Kubernetes CRDs, the key is making your NAT configuration declarative, version-controlled, and reproducible. The days of clicking through consoles to provision network infrastructure should be behind us.

Planton's approach distills Cloud Router NAT configuration to its essence: specify where (VPC and region), what (which subnets), and how (IP allocation), then get out of your way with sensible defaults. The result is infrastructure that's both simple to provision and production-ready—exactly what multi-cloud deployment demands.


# GCP Subnetwork Deployment: From Ad-Hoc to Production-Ready

## Introduction

In the early days of Google Cloud Platform, the default VPC network seemed like a convenience: pre-created subnets in every region, ready to use. But this "convenience" became a liability as organizations scaled. Auto-mode VPCs allocate identical IP ranges across all projects, making VPC peering impossible and causing headaches when connecting to on-premises networks. Google's own documentation now states plainly: **"Production networks should be planned using custom mode VPC networks."**

This shift from auto-mode to custom mode isn't just about best practices—it's about treating network infrastructure as a first-class architectural decision. Subnetworks are the building blocks of your cloud network topology, defining where workloads live, how they communicate, and whether they can scale. Get subnet sizing wrong, and you'll face IP exhaustion in a growing GKE cluster. Forget to enable Private Google Access, and your locked-down VMs can't reach Cloud Storage. Overlap IP ranges, and you'll block future VPC peering or hybrid connectivity.

This document explores the deployment methods for GCP subnetworks, from manual console clicks to declarative infrastructure-as-code, and explains how Planton abstracts the complexity while maintaining the flexibility you need for production environments.

## The Deployment Maturity Spectrum

### Level 0: Auto-Mode Networks (The Anti-Pattern)

Auto-mode VPCs automatically create one subnet per region using a predefined /20 range from the 10.128.0.0/9 block. This sounds convenient, but it's a trap:

- **Identical IP ranges across projects**: Every auto-mode VPC uses the same IP space, preventing VPC peering between projects
- **No CIDR control**: You can't customize ranges to fit your IP addressing scheme or avoid conflicts with on-premises networks
- **Inflexible for GKE**: No secondary ranges for pods and services, forcing you into routes-based mode (deprecated by Google)

**Verdict**: Avoid auto-mode VPCs entirely in production. They're only suitable for quick demos or learning exercises.

### Level 1: Manual Console Creation (Learning Mode)

Creating subnets via the GCP Console teaches the fundamentals: navigate to VPC networks, click "Add subnet," specify region and CIDR range, toggle Private Google Access. This manual approach works for understanding the concepts and for one-off test environments.

**Common pitfalls**:
- Forgetting to plan for growth—a /24 (256 IPs) might seem fine today but won't scale
- Overlooking secondary ranges needed for GKE, requiring subnet recreation later
- Not documenting IP allocations, leading to accidental CIDR overlaps in multi-team environments

**Verdict**: Fine for learning and small experiments, but manual creation doesn't scale and leaves no audit trail. Production environments need repeatability and version control.

### Level 2: CLI and Scripting (Semi-Automated)

The `gcloud` CLI enables scripting subnet creation:

```bash
gcloud compute networks subnets create my-subnet \
  --network=my-vpc \
  --region=us-central1 \
  --range=10.0.0.0/20 \
  --enable-private-ip-google-access \
  --enable-flow-logs \
  --secondary-range=pods=10.0.16.0/20,services=10.0.32.0/24
```

This approach is scriptable and can be integrated into CI/CD pipelines. You can also use the Python SDK or other language clients for programmatic control.

**Limitations**:
- No built-in state management—you must track what's deployed separately
- Imperative rather than declarative—scripts describe actions, not desired state
- Error handling and idempotency require custom logic

**Verdict**: Useful for custom automation and one-time migrations, but for collaborative infrastructure management, declarative IaC tools provide better safety and clarity.

### Level 3: Declarative Infrastructure-as-Code (Production Standard)

This is where production deployments live. Declarative IaC tools like **Terraform**, **Pulumi**, and **OpenTofu** treat subnet configuration as code: version-controlled, peer-reviewed, and consistently applied.

#### Terraform Example

```hcl
resource "google_compute_network" "vpc" {
  name                    = "production-vpc"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "app_subnet" {
  name          = "app-subnet-us-central1"
  region        = "us-central1"
  network       = google_compute_network.vpc.id
  ip_cidr_range = "10.100.0.0/20"

  secondary_ip_range {
    range_name    = "pods"
    ip_cidr_range = "10.100.16.0/20"
  }

  secondary_ip_range {
    range_name    = "services"
    ip_cidr_range = "10.100.32.0/24"
  }

  private_ip_google_access = true
}
```

#### Pulumi Example (Python)

```python
import pulumi_gcp as gcp

vpc = gcp.compute.Network(
    "vpc",
    name="production-vpc",
    auto_create_subnetworks=False
)

subnet = gcp.compute.Subnetwork(
    "app-subnet",
    name="app-subnet-us-central1",
    region="us-central1",
    network=vpc.id,
    ip_cidr_range="10.100.0.0/20",
    secondary_ip_ranges=[
        {"range_name": "pods", "ip_cidr_range": "10.100.16.0/20"},
        {"range_name": "services", "ip_cidr_range": "10.100.32.0/24"}
    ],
    private_ip_google_access=True
)
```

Both tools provide:
- **State management**: Track what's deployed and detect drift
- **Plan/preview**: See changes before applying them
- **Dependency management**: Ensure the VPC exists before creating subnets
- **Module reusability**: Encapsulate patterns for consistent multi-region deployments

**Key differences**:
- **Terraform/OpenTofu**: HCL syntax, mature module ecosystem, remote state backends (GCS, S3, etc.)
- **Pulumi**: General-purpose languages (Python, TypeScript, Go), integrates easily with existing app code, Pulumi Service or self-managed state

**Verdict**: This is the production standard. Both Terraform and Pulumi are excellent choices—pick based on team expertise and ecosystem fit. OpenTofu (the open-source Terraform fork) offers the same functionality without HashiCorp's recent license restrictions.

### Level 4: Kubernetes-Native Management (Specialized Use Case)

**Crossplane** extends Kubernetes to manage cloud infrastructure via CRDs. You define a subnet as a Kubernetes resource:

```yaml
apiVersion: compute.gcp.crossplane.io/v1beta1
kind: Subnetwork
metadata:
  name: app-subnet
spec:
  forProvider:
    region: us-west1
    ipCidrRange: 10.150.0.0/20
    networkRef:
      name: production-network
    secondaryIpRanges:
      - rangeName: pods
        ipCidrRange: 10.150.16.0/20
      - rangeName: services
        ipCidrRange: 10.150.32.0/24
    privateIpGoogleAccess: true
  providerConfigRef:
    name: gcp-provider
```

Crossplane's continuous reconciliation means changes are automatically corrected, and you can compose higher-level abstractions (e.g., a "Platform VPC" that bundles network, subnets, and firewall rules).

**Trade-offs**:
- **Pros**: GitOps-friendly, no state file locking, compositional abstractions, unified management if you already run Kubernetes
- **Cons**: Requires a Kubernetes cluster, adds operational complexity, smaller community than Terraform/Pulumi

**Verdict**: Excellent if you're already deeply invested in Kubernetes and want unified control plane management. For most teams, Terraform or Pulumi offers a simpler path.

## Production Essentials for GCP Subnetworks

### Custom Mode Is Non-Negotiable

Always create custom mode VPCs for production. This gives you control over:
- **CIDR allocation**: Plan IP space to avoid overlaps with on-prem, other clouds, or peered VPCs
- **Regional placement**: Only create subnets in regions you actually use
- **Secondary ranges**: Define alias IP ranges for GKE pods and services

### Subnet Sizing and CIDR Planning

Choosing the right subnet size is critical because **you cannot change a subnet's CIDR or region after creation**. Common sizes:

- **/24 (256 IPs)**: Small services or dedicated subnets for VPC connectors (e.g., Cloud Run)
- **/22 (1,024 IPs)**: Moderate workloads or smaller GKE clusters
- **/20 (4,096 IPs)**: Standard for production GKE clusters—enough for hundreds of nodes
- **/16 (65,536 IPs)**: Reserved for very large environments or when you need significant growth headroom

**Planning principles**:
1. **Plan for growth**: Allocate 2-3x your initial capacity
2. **Avoid overlap**: Coordinate with on-prem networks and other VPCs (VPC peering requires non-overlapping ranges)
3. **Document allocations**: Maintain an IP address management (IPAM) spreadsheet or tool
4. **Reserve ranges**: Don't use every IP block—leave room for future subnets and environments

Remember that GCP reserves **4 IPs per subnet** (network address, broadcast, gateway, DNS), so a /24 actually gives you 252 usable addresses.

### Secondary IP Ranges for GKE

If you're running GKE clusters, you **must** define secondary ranges at subnet creation time:

- **Pods range**: Size based on `max nodes × max pods per node` (default 110 pods/node)
  - Example: 100 nodes × 110 pods = 11,000 IPs → use at least a /18 (16,384 IPs)
- **Services range**: Size based on the number of Kubernetes Services (ClusterIPs)
  - Example: 200 services → use a /24 (256 IPs)

Common mistake: Under-sizing the pod range and hitting IP exhaustion as the cluster grows. Err on the larger side—a /18 or /17 for pods is typical in production.

### Private Google Access

Enable **Private Google Access** when:
- VMs or GKE nodes **don't have external IPs** (common security practice)
- Workloads need to access Google APIs (GCS, GCR, BigQuery, etc.)
- You're running a private GKE cluster (control plane has no public endpoint)

This setting allows internal-only instances to reach Google services via Google's private network, avoiding public internet transit.

**Critical**: Private Google Access is **per-subnet**. Enable it on every subnet where internal-only workloads need API access.

### VPC Flow Logs

VPC Flow Logs capture network flow metadata (5-tuple: source IP, dest IP, ports, protocol, byte count) for security analysis, troubleshooting, and capacity planning.

**When to enable**:
- **Production subnets**: For audit trails and incident response
- **Regulated environments**: PCI, HIPAA, or SOC 2 compliance often requires network flow logs

**When to skip**:
- **High-traffic dev/test**: Logs generate cost and volume—use sampling (e.g., 0.5 sample rate)
- **Low-value subnets**: If you don't need forensics, skip it to save costs

You can configure sampling rates and metadata inclusion (VM tags, instance details) to control volume.

### Anti-Patterns to Avoid

1. **Using auto-mode VPCs**: Inflexible, prevents peering, causes conflicts
2. **Overlapping CIDR ranges**: Blocks VPC peering and hybrid connectivity
3. **Forgetting secondary ranges for GKE**: Forces subnet recreation or routes-based mode (deprecated)
4. **Under-sizing subnets**: You can't expand a subnet—you'd have to recreate it
5. **Not documenting IP allocations**: Leads to accidental overlaps in multi-team environments

## IaC Tool Comparison: Terraform vs. Pulumi vs. Crossplane

| Feature                    | Terraform/OpenTofu        | Pulumi                    | Crossplane               |
|----------------------------|---------------------------|---------------------------|--------------------------|
| **Language**               | HCL                       | Python, TypeScript, Go    | YAML (CRDs)              |
| **State Management**       | Remote state (GCS, S3)    | Pulumi Service or self-hosted | Kubernetes etcd (no state file) |
| **Plan/Preview**           | ✅ `terraform plan`        | ✅ `pulumi preview`        | ✅ Kubernetes dry-run     |
| **Module Ecosystem**       | Large (Terraform Registry)| Growing (Pulumi Registry) | Smaller (Crossplane packages) |
| **Learning Curve**         | Moderate (HCL syntax)     | Low (familiar languages)  | Moderate (K8s knowledge required) |
| **GitOps Integration**     | Via tools (Atlantis, etc.)| Via Pulumi Operator       | Native (Argo CD, Flux)   |
| **Production Readiness**   | ✅ Mature                  | ✅ Mature                  | ✅ Growing adoption       |

**Recommendation**:
- **Terraform/OpenTofu**: Best for teams with existing Terraform expertise or who want the largest module ecosystem
- **Pulumi**: Best for teams who prefer general-purpose languages and want tighter integration with application code
- **Crossplane**: Best for Kubernetes-centric teams who want unified infrastructure and app management

All three are production-ready. The choice comes down to team skills and architectural preferences.

## The Planton Approach

Planton uses **Pulumi** under the hood to deploy GCP subnetworks, providing a simple, declarative API that abstracts away infrastructure details while maintaining full flexibility.

### Why Pulumi?

- **Multi-cloud consistency**: Pulumi's programming model works across AWS, Azure, GCP, and Kubernetes
- **Real programming languages**: Easier to express complex logic (e.g., generating subnets for multiple regions)
- **Open source**: Pulumi's core engine is Apache 2.0 licensed

### The GcpSubnetwork API

Planton's `GcpSubnetwork` API distills subnet configuration to the essential 20% that covers 80% of use cases:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpSubnetwork
metadata:
  name: prod-app-subnet
spec:
  project_id: my-gcp-project
  vpc_self_link: projects/my-gcp-project/global/networks/my-vpc
  region: us-central1
  ip_cidr_range: 10.100.0.0/20
  secondary_ip_ranges:
    - range_name: pods
      ip_cidr_range: 10.100.16.0/20
    - range_name: services
      ip_cidr_range: 10.100.32.0/24
  private_ip_google_access: true
```

**What's included**:
- **project_id**: The GCP project (can reference a `GcpProject` resource)
- **vpc_self_link**: Parent VPC network (can reference a `GcpVpc` resource)
- **region**: Where the subnet lives (immutable after creation)
- **ip_cidr_range**: Primary IPv4 CIDR block
- **secondary_ip_ranges**: Alias IP ranges for GKE or other uses (optional)
- **private_ip_google_access**: Enable internal-only access to Google APIs (boolean)

**What's omitted** (defaults or rarely used):
- **Flow logs**: Can be added in future versions if needed
- **Purpose and role**: Default to `PRIVATE` (standard subnets)
- **IPv6**: Most organizations are IPv4-only today; can add later

This minimal API reduces complexity while covering the vast majority of production subnet configurations.

### Multi-Region Patterns

For high availability, you'll typically create one subnet per region:

```yaml
# US East
apiVersion: gcp.planton.dev/v1
kind: GcpSubnetwork
metadata:
  name: app-subnet-us-east
spec:
  project_id: prod-project
  vpc_self_link: projects/prod-project/global/networks/prod-vpc
  region: us-east1
  ip_cidr_range: 10.0.0.0/20
  private_ip_google_access: true
---
# Europe West
apiVersion: gcp.planton.dev/v1
kind: GcpSubnetwork
metadata:
  name: app-subnet-eu-west
spec:
  project_id: prod-project
  vpc_self_link: projects/prod-project/global/networks/prod-vpc
  region: europe-west1
  ip_cidr_range: 10.0.16.0/20
  private_ip_google_access: true
```

Each region gets a non-overlapping CIDR block from your overall IP plan (e.g., 10.0.0.0/18 divided into /20 chunks).

### Integration with GKE

For GKE clusters, define secondary ranges in the subnet, then reference them in your `GkeCluster` resource:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpSubnetwork
metadata:
  name: gke-cluster-subnet
spec:
  project_id: prod-project
  vpc_self_link: projects/prod-project/global/networks/prod-vpc
  region: us-west1
  ip_cidr_range: 10.50.0.0/20          # Nodes
  secondary_ip_ranges:
    - range_name: pods
      ip_cidr_range: 10.50.16.0/18     # 16,384 IPs for pods
    - range_name: services
      ip_cidr_range: 10.50.80.0/24     # 256 IPs for services
  private_ip_google_access: true
```

The GKE cluster resource would then specify:
- `subnetwork: gke-cluster-subnet`
- `cluster_secondary_range_name: pods`
- `services_secondary_range_name: services`

This separation of concerns—subnets define IP space, GKE consumes it—keeps networking configuration reusable and composable.

## CIDR Planning Deep Dive

**Choosing Subnet Sizes**

Start with your workload requirements and work backward:

1. **Compute Engine VMs**: Count expected instances, add 50% growth buffer
   - Example: 50 VMs → use /26 (64 IPs) or /25 (128 IPs)

2. **GKE Clusters**:
   - **Node range** (primary): `max nodes × 1` (one IP per node)
   - **Pod range** (secondary): `max nodes × max pods per node` (default 110)
   - **Service range** (secondary): `max Kubernetes Services` (typically 100-500)
   - Example: 100 nodes, 110 pods/node → use /20 primary, /18 pods, /24 services

3. **Serverless VPC Connectors** (Cloud Run, Cloud Functions):
   - Requires a **dedicated /28 subnet** (16 IPs) per connector
   - Cannot share with other resources

**IP Address Allocation Strategy**

Divide your private IP space (e.g., 10.0.0.0/8) into blocks:

- **10.0.0.0/16**: Production environment
  - 10.0.0.0/20: us-central1 app subnet
  - 10.0.16.0/20: us-east1 app subnet
  - 10.0.32.0/18: us-central1 GKE cluster (primary + secondaries)
- **10.1.0.0/16**: Staging environment
- **10.2.0.0/16**: Development environment

**Tools and Calculators**

- **CIDR Calculator**: [cidr.xyz](https://cidr.xyz), [subnet-calculator.com](https://www.subnet-calculator.com)
- **Terraform functions**: `cidrsubnet()` for programmatic CIDR allocation
- **IPAM tools**: NetBox, phpIPAM, or internal spreadsheets

**Reserved and Restricted Ranges**

- **GCP reserves 4 IPs per subnet**: network, broadcast, gateway, DNS (RFC 3927)
- **Avoid Google-owned ranges**: 8.8.8.0/24, 35.199.192.0/19 (Private Google Access)
- **Avoid Docker default bridge**: 172.17.0.0/16 (can conflict with VM Docker installs)
- **Stick to RFC 1918**: 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16

## Conclusion

GCP subnetworks are deceptively simple—just specify a region and CIDR range—but the decisions you make ripple through your architecture. Undersize a subnet, and you'll face IP exhaustion. Overlap CIDR ranges, and you'll block VPC peering. Forget secondary ranges, and you'll have to recreate subnets to support GKE.

The maturity progression is clear: move from manual console clicks to declarative infrastructure-as-code, whether that's Terraform, Pulumi, or Crossplane. All three are production-ready; the choice depends on your team's skills and existing toolchain.

Planton abstracts this complexity with a minimal API surface—just six fields cover 80% of subnet configurations—while Pulumi handles the underlying deployment. This approach balances simplicity for common cases with the ability to scale to multi-region, multi-environment architectures.

The key insight? **Treat network design as a first-class architectural decision.** Spend time upfront planning your IP space, allocating CIDR blocks per environment and region, and documenting your choices. A few hours of CIDR planning now will save days of painful network redesign later.

Your subnets are the foundation. Build them right, and everything else becomes easier.

---

## Further Reading

- [GCP VPC Best Practices](https://cloud.google.com/architecture/best-practices-vpc-design) - Official Google Cloud architecture guide
- [GKE Alias IPs and VPC-Native Clusters](https://cloud.google.com/kubernetes-engine/docs/concepts/alias-ips) - Secondary range sizing for Kubernetes
- [VPC Flow Logs Documentation](https://cloud.google.com/vpc/docs/flow-logs) - Network monitoring and security analysis
- [Terraform Google Network Module](https://registry.terraform.io/modules/terraform-google-modules/network/google/latest) - Community-maintained Terraform patterns
- [Pulumi GCP Subnetwork Resource](https://www.pulumi.com/registry/packages/gcp/api-docs/compute/subnetwork/) - Pulumi API reference


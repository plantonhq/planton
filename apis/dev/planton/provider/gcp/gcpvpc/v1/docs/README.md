# GCP VPC Deployment: From Console Chaos to Infrastructure as Code

## The Network Foundation Challenge

Here's a truth that catches many teams off guard: the way you create your VPC network on day one often becomes the blueprint—or the bottleneck—for everything that follows. Choose auto-mode because it's the default? You've just planted the seeds for IP conflicts when you scale to multiple environments or connect to on-premises. Click through the console to "quickly test something"? That one-off network becomes production, and nobody remembers what firewall rules were added at 2 AM during that incident.

Google Cloud's VPC is deceptively simple on the surface—just a network with some subnets, right? But beneath that simplicity lies a sophisticated software-defined network that spans regions globally, routes traffic through an implicit mesh, and integrates with everything from Kubernetes to serverless functions. The question isn't whether you *can* create a VPC—anyone can click through the console in 5 minutes. The real question is: **how do you deploy VPCs consistently, with the right architecture choices, in a way that doesn't haunt you six months later?**

This document examines the evolution of GCP VPC deployment methods, from manual provisioning through modern infrastructure-as-code approaches, and explains why Planton chose a streamlined, custom-mode-first design that prevents the most common networking pitfalls before they happen.

## The Maturity Spectrum: How VPC Provisioning Evolved

### Level 0: Manual Console Provisioning (The Anti-Pattern)

In the beginning, there was the web console. You log in, navigate to VPC networks, click "Create VPC network," accept the defaults (auto-mode, because it sounds convenient), maybe add a subnet or two, and you're done. For a weekend project or learning exercise, this is fine.

For anything you intend to keep running, **this is actively harmful**.

The problems compound quickly:
- **Configuration amnesia**: Six months later, when you need to replicate the setup for staging, nobody remembers whether that firewall rule was priority 1000 or 1001, or whether Private Google Access was enabled.
- **Auto-mode traps**: The console defaults to auto-mode VPCs, which automatically create subnets in *every* GCP region using predefined IP ranges (`10.128.0.0/9`). This seems convenient until you try to peer two VPCs together—GCP outright refuses because the IP ranges overlap. Or until you connect to on-premises and discover your corporate network uses the same `10.128.x.x` space.
- **Drift and chaos**: Manual changes accumulate. Someone adds a firewall rule via console. Another engineer modifies it via `gcloud` CLI. A third person doesn't know either change happened. The result is a network nobody fully understands.

As one infrastructure engineer put it: *"Manually provisioning resources like VPCs... is a BAD IDEA!"* Human memory fades. Documentation goes stale. Manual processes don't scale, don't audit, and don't prevent the same mistakes from happening twice.

**Verdict**: Avoid manual provisioning for anything beyond learning. If you must use the console initially, document every step and transition to automation immediately.

### Level 1: Scripted CLI (Imperative Automation)

The natural evolution from clicking is scripting: `gcloud compute networks create my-vpc --subnet-mode=custom --bgp-routing-mode=regional`. You write a bash script or Python program using the Cloud Client Libraries, and now your network creation is repeatable.

This is better than manual. It's auditable (the script is in version control), and you can run it consistently. Many teams operate at this level successfully.

But there are limitations:
- **Imperative mindset**: Scripts issue commands to *do things*: "create this network," "add that subnet." If the script runs twice, you need error handling to check if resources already exist. If reality diverges from what the script expects (someone deleted a subnet manually), the script might fail or create duplicates.
- **No built-in state tracking**: The script doesn't remember what it created last time. You have to query GCP APIs to determine current state, then decide what actions to take.
- **Drift detection requires extra work**: Knowing whether your actual VPC configuration matches what the script would create requires you to query and compare. There's no automatic "plan" showing differences.

**Verdict**: Scripted CLI is a step up from manual and works for simple scenarios or CI/CD integrations. But for complex, stateful infrastructure, declarative tools provide clearer semantics and better change management.

### Level 2: Google Deployment Manager (Native IaC, Quietly Fading)

Google Cloud Deployment Manager was GCP's original answer to infrastructure-as-code. You write YAML (or Jinja/Python templates) describing resources like `compute.v1.network`, then run `gcloud deployment-manager deployments create`. DM handles resource creation order, maintains a manifest of what was deployed, and tracks state within GCP itself (no separate state file to manage).

For GCP-only environments, DM worked:
- **Native integration**: No extra authentication needed; it uses your project credentials directly.
- **Managed state**: Google stores the deployment state, so you don't need to set up remote state storage.

But DM never quite took off in the broader community:
- **GCP-only**: No multi-cloud support. If your organization uses AWS or Azure, you need a different tool.
- **Limited ecosystem**: While it supported most GCP services, the module ecosystem and community tooling remained small compared to Terraform's.
- **Google's own shift**: In recent years, Google has quietly de-emphasized DM. New features sometimes lag in DM support, and Google's own blueprints increasingly reference Terraform.

**Verdict**: Deployment Manager is stable and usable, but it's not the future. Google's direction points elsewhere, and the industry momentum is clearly with multi-cloud IaC tools.

### Level 3: Terraform and OpenTofu (Industry Standard IaC)

Terraform changed the infrastructure game with declarative, stateful configuration. You describe the desired end state—a VPC with specific subnets, routing mode, firewall rules—and Terraform figures out how to make reality match:

```hcl
resource "google_compute_network" "vpc" {
  name                    = "prod-network"
  auto_create_subnetworks = false
  routing_mode            = "GLOBAL"
}

resource "google_compute_subnetwork" "subnet_west" {
  name          = "prod-uswest1"
  ip_cidr_range = "10.10.0.0/16"
  region        = "us-west1"
  network       = google_compute_network.vpc.id

  secondary_ip_range {
    range_name    = "gke-pods"
    ip_cidr_range = "10.10.192.0/18"
  }
}
```

Terraform's power lies in its **plan/apply cycle**: before making changes, you see a preview of what will be created, modified, or destroyed. This turns infrastructure changes into code review opportunities. Teams can collaborate via remote state storage (Cloud Storage, Terraform Cloud), and drift is detected on the next `terraform plan`.

**OpenTofu** emerged in late 2023 as a community fork when HashiCorp changed Terraform's licensing. It's functionally identical to Terraform for GCP use cases—same HCL syntax, same providers—but with open-source governance. For practical purposes, Terraform and OpenTofu are interchangeable in how they handle GCP VPCs.

**Strengths**:
- **Mature and battle-tested**: Terraform has been the de facto IaC standard for nearly a decade. The GCP provider is comprehensive and well-maintained.
- **Rich ecosystem**: Terraform Registry hosts hundreds of community modules for GCP networking patterns (Shared VPC, hub-and-spoke, etc.).
- **Multi-cloud**: One tool for GCP, AWS, Azure, and 3,000+ other providers.

**Considerations**:
- **State management complexity**: The state file is critical and must be stored remotely for team collaboration. Locking mechanisms (via Cloud Storage or Terraform Cloud) prevent concurrent modifications but add operational overhead.
- **On-demand drift detection**: Terraform doesn't continuously monitor infrastructure. Drift is detected only when you run `terraform plan`. If someone changes a firewall rule manually, you won't know until the next plan.

**Verdict**: Terraform/OpenTofu is the production-ready choice for most teams. It's proven, flexible, and has massive community support. If you're deploying GCP VPCs at any meaningful scale, this is the tier you want to be operating at.

### Level 4: Pulumi (General-Purpose Languages for Infrastructure)

Pulumi offers a different take on IaC: instead of learning a domain-specific language (HCL), you write infrastructure code in TypeScript, Python, Go, or other general-purpose languages. A VPC in Pulumi might look like:

```typescript
const network = new gcp.compute.Network("vpc", {
    autoCreateSubnetworks: false,
    routingMode: "GLOBAL"
});

const subnet = new gcp.compute.Subnetwork("subnet-west", {
    network: network.id,
    ipCidrRange: "10.10.0.0/16",
    region: "us-west1",
    secondaryIpRanges: [{
        rangeName: "gke-pods",
        ipCidrRange: "10.10.192.0/18"
    }]
});
```

This approach appeals to teams who prefer standard programming constructs—loops, functions, type checking—over declarative DSLs. You get package management, IDE support, and the ability to encapsulate complex logic in familiar ways.

**Strengths**:
- **Developer-friendly**: No new language to learn if you already know Python or TypeScript.
- **Full programming power**: Create abstractions, share logic via packages, use conditionals and loops naturally.
- **Modern CLI and state management**: Pulumi's CLI is polished, and state can be stored in Pulumi Service (SaaS) or self-hosted backends.

**Considerations**:
- **Smaller ecosystem for GCP**: While Pulumi's GCP support is solid (it can even use Terraform providers under the hood), the community module ecosystem is smaller than Terraform's.
- **Flexibility can be complexity**: Giving teams full programming power means they *will* use it. Complex abstractions can make infrastructure code harder to understand than straightforward Terraform HCL.

**Verdict**: Pulumi is excellent if your team values general-purpose languages and wants tight integration with application code. It's production-ready and well-supported. The choice between Pulumi and Terraform often comes down to team preference and existing workflows rather than capability differences.

### Level 5: Crossplane (Continuous Reconciliation via Kubernetes)

Crossplane represents a fundamentally different model: infrastructure-as-Kubernetes-resources. You define a GCP VPC as a custom resource in a Kubernetes cluster:

```yaml
apiVersion: compute.gcp.crossplane.io/v1beta1
kind: Network
metadata:
  name: prod-vpc
spec:
  forProvider:
    autoCreateSubnetworks: false
    routingConfig:
      routingMode: GLOBAL
```

A Crossplane controller continuously watches this resource and ensures the corresponding GCP VPC exists and matches the specification. If someone modifies the VPC outside of Crossplane, the controller detects the drift and reverts it—**continuous reconciliation** rather than on-demand changes.

This model is powerful for platform teams building internal developer platforms:
- **GitOps native**: Infrastructure changes flow through Git commits, pull requests, and Kubernetes admission controllers.
- **Self-healing**: Drift is automatically corrected. Manual changes are undone.
- **Higher-level abstractions**: Platform teams can define composite resources like "StandardVPC" that bundle a network, subnets, firewall rules, and Cloud Router into a single developer-friendly API.

**Considerations**:
- **Requires Kubernetes**: You need a control-plane cluster running 24/7. This adds operational complexity.
- **Paradigm shift**: Managing infrastructure via Kubernetes CRDs requires teams to embrace the Kubernetes operational model (reconciliation loops, eventual consistency, etc.).
- **Best for platforms**: Crossplane shines when you're building a platform *on top of* cloud primitives, not when you're just provisioning individual VPCs.

**Verdict**: Crossplane is the most sophisticated approach and represents the future for platform engineering teams. But it's overkill for straightforward VPC provisioning. Choose this if you're building an internal platform where developers should never touch raw GCP APIs—otherwise, Terraform or Pulumi will be simpler.

## Comparative Analysis: Choosing Your IaC Tool

| Criterion | Terraform/OpenTofu | Pulumi | Crossplane | Deployment Manager |
|-----------|-------------------|--------|------------|-------------------|
| **Learning Curve** | Moderate (HCL syntax) | Low (if familiar with the language) | High (K8s + IaC concepts) | Moderate (YAML + GCP APIs) |
| **Multi-Cloud** | ✅ Excellent (3000+ providers) | ✅ Excellent | ✅ Good (via providers) | ❌ GCP only |
| **Community & Modules** | ✅ Huge ecosystem | 🟡 Growing ecosystem | 🟡 Smaller, but active | ⚠️ Limited, declining |
| **State Management** | Remote state (GCS, TF Cloud) | Remote state (Pulumi Service, S3, etc.) | Kubernetes etcd | GCP-managed |
| **Drift Detection** | On-demand (`plan`) | On-demand (`preview`) | ✅ Continuous | On-demand |
| **Best For** | Teams wanting mature, proven IaC with HCL | Teams preferring code in Python/TS/Go | Platform teams building on K8s | Legacy GCP-only projects |
| **Production Readiness** | ✅ Battle-tested | ✅ Production-ready | ✅ Production-ready (with K8s expertise) | 🟡 Stable, but not Google's focus |

**Decision Framework**:
- **Choose Terraform/OpenTofu** if you want the industry-standard tool with the largest ecosystem and proven track record. This is the safe, pragmatic choice.
- **Choose Pulumi** if your team prefers general-purpose languages and wants tighter integration with application code.
- **Choose Crossplane** if you're building an internal platform and want continuous reconciliation with GitOps workflows.
- **Avoid Deployment Manager** for new projects unless you have a specific requirement for GCP-native tooling and no multi-cloud needs.

## Planton's Approach: Streamlined Custom-Mode VPCs

Planton's GCP VPC module is designed around a simple principle: **make the right thing easy and the wrong thing hard**.

### What Planton Provides

The `GcpVpc` API in Planton is intentionally minimal, focusing on the 20% of configuration that covers 80% of real-world use cases:

```protobuf
message GcpVpcSpec {
  // Project ID where the VPC will be created
  StringValueOrRef project_id = 1;

  // Auto-create subnets? (Default: false - custom mode recommended)
  bool auto_create_subnetworks = 2;

  // Routing mode: REGIONAL (default) or GLOBAL
  optional GcpVpcRoutingMode routing_mode = 3;
}
```

**What's included**:
- **Custom mode by default**: `auto_create_subnetworks` defaults to `false`, steering users away from the auto-mode pitfalls documented earlier (IP overlaps, lack of control).
- **Explicit routing mode**: Expose the `routing_mode` choice (REGIONAL vs GLOBAL) because it's fundamental for hybrid connectivity, but default to REGIONAL (the GCP default) since most use cases don't need global route advertisement.
- **Project ID with foreign key support**: Reference a `GcpProject` resource directly, enabling cross-resource dependencies.

**What's excluded** (and why):
- **MTU settings**: 99% of users never change the default 1460 bytes. The 1% who need jumbo frames for HPC can override via advanced configuration.
- **IPv6 and ULA**: Still low adoption. Can be added when demand grows.
- **Direct subnet management**: Subnets, firewall rules, and routes are typically managed via separate resources (`GcpSubnetwork`, etc.) for better modularity and reuse.

### The Philosophy: Guard Rails, Not Handcuffs

Planton doesn't try to abstract away GCP's VPC entirely. Instead, it provides a thin, opinionated layer that:

1. **Defaults to best practices**: Custom mode, regional routing, explicit project references.
2. **Prevents common mistakes**: You can't accidentally create an auto-mode VPC and get into IP conflicts later without explicitly choosing that path.
3. **Integrates with the ecosystem**: Under the hood, Planton generates Terraform or Pulumi code (or calls GCP APIs directly), leveraging battle-tested tooling rather than reinventing the wheel.
4. **Keeps configuration auditable**: The protobuf API is version-controlled, typed, and can be validated at CI time. You get all the benefits of infrastructure-as-code with a cleaner, more focused schema.

This isn't about hiding complexity for complexity's sake. It's about encoding the lessons from the research—the knowledge that auto-mode causes problems, that routing mode matters for hybrid connectivity, that most advanced knobs can safely default—into an API that guides users toward success.

### When to Use Planton's GcpVpc

**Use it when**:
- You want a streamlined, best-practice VPC without boilerplate.
- You're managing infrastructure via protobuf-defined APIs (e.g., in a GitOps or platform engineering context).
- You need cross-cloud consistency (Planton provides similar abstractions for AWS, Azure, etc.).

**Consider raw Terraform/Pulumi instead when**:
- You need extremely fine-grained control over every GCP VPC knob (though you can often extend Planton or use it for the base network and Terraform for advanced tweaks).
- You're deeply embedded in the Terraform ecosystem and prefer direct HCL.

## Architecture Patterns Enabled by Planton

Planton's VPC design supports the full range of GCP networking patterns:

### Pattern 1: Single-Region Custom VPC (Dev/Test)

A simple, custom-mode VPC in one region with a subnet for internal workloads:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpVpc
metadata:
  name: dev-network
spec:
  project_id: my-dev-project
  auto_create_subnetworks: false
  routing_mode: REGIONAL
```

Subnets, firewall rules, and routes are defined separately, keeping the base network clean.

### Pattern 2: Multi-Region Shared VPC (Production)

A global VPC with regional subnets, used as a Shared VPC host for multiple service projects:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpVpc
metadata:
  name: prod-network
spec:
  project_id: network-host-project
  auto_create_subnetworks: false
  routing_mode: GLOBAL  # For multi-region hybrid connectivity
```

The `GLOBAL` routing mode ensures Cloud Routers in any region can advertise routes to all subnets, critical for hybrid VPN/Interconnect setups. Shared VPC attachment (enabling the host project and adding service projects) is typically handled via separate configuration or Terraform resources that reference this VPC.

### Pattern 3: GKE-Ready VPC with Secondary Ranges

While the `GcpVpc` resource defines the base network, subnets with secondary IP ranges (for GKE pods and services) are configured via `GcpSubnetwork` resources, which reference the VPC. This separation allows flexible IP planning per cluster without bloating the VPC configuration.

## Conclusion: Infrastructure as Code Isn't Optional—It's the Baseline

The journey from console clicks to mature IaC is a progression every team makes, often painfully. Starting with custom-mode VPCs, using declarative IaC tools (Terraform, Pulumi, or Crossplane depending on your context), and avoiding the auto-mode trap aren't advanced techniques—they're the baseline for production-grade networking.

Planton's GCP VPC module distills the lessons from thousands of deployments into a simple API: custom mode by default, explicit routing choices, and a focus on what actually matters. It's not about hiding GCP's capabilities. It's about providing a saner starting point so you spend less time debugging IP conflicts and more time building what matters.

Whether you're provisioning your first VPC or refactoring your tenth, the principles remain the same: **plan your IP space, use custom mode, codify everything, and choose tools that make drift visible**. Follow those rules, and your network becomes a foundation you can build on—not a liability you're stuck with.

For further reading:
- [GCP VPC Best Practices (Google Cloud Architecture Center)](https://docs.cloud.google.com/architecture/best-practices-vpc-design)
- [IP Address Planning for Large-Scale GKE Deployments](https://medium.com/google-cloud/ip-address-planning-for-large-scale-gke-deployments-48fdee0f7722)
- [Terraform vs. Crossplane: A Practical Comparison](https://medium.com/@paolo.salvatori/terraform-vs-crossplane-a-practical-comparison-992dc9745e08)


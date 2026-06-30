# Azure Virtual Network Deployment: Beyond the ClickOps

## Introduction

Virtual networks are the foundation of cloud infrastructure, yet the conventional wisdom about how to create them has evolved dramatically. For years, the Azure Portal's point-and-click interface was the entry point for most users learning Azure networking. Click through a wizard, fill in some address ranges, and you have a VNet. Simple, visual, immediate—and completely unsustainable in production.

The problem isn't that the Azure Portal exists. It's that what works for learning and one-off experiments becomes a liability the moment you need to reproduce an environment, audit changes, or scale to multiple teams. Without a codified definition, there's no single source of truth. Manual changes accumulate like sediment, creating configuration drift that's nearly impossible to reverse without careful archaeology through Azure's activity logs.

Modern Azure networking has matured beyond these manual workflows. Infrastructure-as-Code (IaC) tools have become production-ready, Azure itself has introduced native declarative languages, and Kubernetes-native approaches have emerged for teams operating at scale. The challenge isn't finding a way to deploy a VNet—it's choosing the right approach for your production requirements.

This document explains the maturity spectrum of Azure VNet deployment methods, compares the leading IaC tools, and clarifies how Planton leverages these foundations to provide a simplified, production-ready abstraction for teams building on Azure.

## The Maturity Spectrum: From Manual to Production-Ready

### Level 0: The Anti-Pattern (Manual Portal Workflows)

Creating Azure VNets through the Azure Portal is intuitive for newcomers but fails in production for several fundamental reasons:

**No Version Control or Audit Trail**: When you click through the Azure Portal to create a VNet, there's no record of exactly what settings were chosen except in Azure's own activity log. If you need to recreate that network in another region or subscription, you're relying on memory and screenshots. There's no diff, no commit history, no way to see who changed what and when.

**Human Error at Scale**: Setting up one VNet manually might work. Setting up ten across dev, staging, and production environments means ten opportunities for typos, forgotten settings, or inconsistent configurations. Research shows that as manual setup increases, the probability of human error increases significantly—especially in subnet sizing, NSG rules, and DNS configuration.

**Configuration Drift**: The real killer is drift. Someone manually adds a subnet for a quick test. Another engineer tweaks an NSG rule without documenting it. Six months later, your staging environment doesn't match production, and nobody knows why. Azure can't tell you what *should* be there, only what *is* there.

**The Verdict**: Manual workflows are acceptable for learning Azure or running temporary experiments. For anything that needs to be maintained, reproduced, or audited, they're a non-starter.

### Level 1: Scriptable but Imperative (Azure CLI & PowerShell)

The Azure CLI (`az network vnet create`) and PowerShell (`New-AzVirtualNetwork`) represent a step forward: you can script VNet creation, check those scripts into git, and run them in CI/CD pipelines.

**What This Solves**: Repeatability. You can run the same script to create the same VNet multiple times. Commands are idempotent—running them again updates or reports existing resources rather than failing.

**What It Doesn't Solve**: These are still imperative scripts. They tell Azure what to *do*, not what state to *maintain*. If someone manually changes a subnet outside your script, the script won't know unless you add explicit checks. You're also responsible for managing dependencies (create VNet first, then subnets, then NSGs, then associate them) and handling partial failures.

**The Verdict**: CLI and PowerShell are excellent for quick automation, one-off migrations, or integrating Azure tasks into larger scripts. For managing the desired state of infrastructure over time, declarative IaC is better.

### Level 2: Programmatic Control (Azure SDKs)

Azure offers SDKs in Python, Go, Java, C#, and other languages, letting you programmatically define and deploy VNets by calling Azure management APIs directly.

**What This Solves**: Full control and the ability to integrate complex logic—conditionals, loops, error handling—into your provisioning code. Useful for building custom provisioning services, test frameworks, or platforms that need to create infrastructure dynamically.

**What It Doesn't Solve**: You're essentially writing imperative code that interacts with Azure APIs. You manage API versions, authentication, retries, and state tracking yourself. For straightforward VNet creation, this is overkill. The complexity outweighs the benefits unless you're building a platform *on top of* Azure provisioning.

**The Verdict**: Powerful for specialized use cases (custom platforms, testing harnesses), but too low-level for most production infrastructure management.

### Level 3: Declarative Configuration Management (Ansible)

Tools like Ansible bring declarative YAML playbooks to Azure infrastructure. The `azure.azcollection.azure_rm_virtualnetwork` module lets you describe a VNet's desired state, and Ansible will create or update it to match.

**What This Solves**: Declarative intent and integration with broader configuration management workflows. If you're already using Ansible to configure VMs, databases, and applications, managing VNets in the same playbook ecosystem is convenient.

**What It Doesn't Solve**: Ansible is agentless and stateless. It doesn't maintain a separate record of what it deployed last time—it checks Azure's current state on each run and converges toward your playbook's definition. This works, but drift detection is implicit, not explicit. Ansible also doesn't have the same deep infrastructure focus as dedicated IaC tools; its sweet spot is configuration management post-provisioning.

**The Verdict**: Good for teams already invested in Ansible and wanting a single tool for infra + config. Less ideal as the primary IaC tool for complex, multi-resource Azure environments.

### Level 4: Production-Ready Infrastructure-as-Code

This is where the industry has settled for production Azure deployments. Four tools dominate: **Terraform** (and its open-source fork **OpenTofu**), **Pulumi**, **Bicep/ARM**, and **Crossplane**. Each is mature, battle-tested, and capable of managing Azure VNets at scale—but they differ significantly in philosophy, workflow, and fit.

## Comparing Production IaC Tools

### Terraform (and OpenTofu): The Pragmatic Standard

**Community & Adoption**: Terraform is the most widely adopted multi-cloud IaC tool, with a massive community, thousands of modules, and extensive enterprise use. OpenTofu, its open-source fork (created after HashiCorp changed Terraform's license in 2023), maintains compatibility with Terraform's syntax and providers while remaining fully open-source.

**How It Works**: Terraform uses HashiCorp Configuration Language (HCL) to declaratively define resources like `azurerm_virtual_network`. You run `terraform plan` to preview changes, `terraform apply` to execute them. Terraform maintains a **state file** that maps your configuration to actual Azure resource IDs, enabling it to detect drift and calculate minimal change sets.

**Strengths**:
- **Mature and proven**: Enterprises manage enormous Azure environments with Terraform.
- **Multi-cloud portability**: The same tooling works for AWS, GCP, Kubernetes, and more.
- **Predictable workflow**: Plan before apply, clear diff of changes, remote state with locking for team collaboration.

**Trade-offs**:
- **State management overhead**: You must store and protect the state file (often in Azure Storage with locking).
- **Drift detection is manual**: Terraform won't automatically fix drift—you have to run `terraform plan` to see it and `apply` to correct it.
- **Learning curve**: HCL is straightforward but adds a language to learn.

**Best Fit**: Teams wanting a battle-tested, multi-cloud IaC solution with broad community support and flexibility.

### Pulumi: Code-First Infrastructure

**Community & Adoption**: Pulumi is younger than Terraform but growing rapidly, especially among developer-heavy teams. It's production-ready and backed by Pulumi Corp, which offers commercial support.

**How It Works**: Instead of HCL, Pulumi lets you define infrastructure in general-purpose languages—TypeScript, Python, Go, C#, Java. Under the hood, Pulumi's `azure-native` provider maps to Azure Resource Manager APIs, much like Terraform's provider. Pulumi also manages state (in Pulumi Cloud or self-hosted backends).

**Strengths**:
- **Familiar languages**: Developers can use loops, functions, conditionals, and their favorite libraries to define infrastructure.
- **Strong Azure coverage**: The `azure-native` provider auto-generates resources from Azure's API definitions, ensuring parity with Azure features.
- **Integration with application code**: Easier to embed infrastructure logic into applications or share code between infra and app layers.

**Trade-offs**:
- **Overkill for simple resources**: Defining a basic VNet in code can be more verbose than equivalent HCL.
- **Smaller ecosystem**: Fewer community modules and examples compared to Terraform.
- **Complexity for simple cases**: If all you need is a VNet with subnets, Pulumi's power may feel like using a sledgehammer to crack a nut.

**Best Fit**: Developer-centric teams comfortable with programming languages, or cases where infrastructure needs to interact with complex logic or application code.

### Bicep/ARM: The Azure-Native Path

**Community & Adoption**: Bicep is Microsoft's recommended way to write Azure Resource Manager (ARM) templates. It reached v1.0 and is now the standard for Azure-native IaC. The community is Azure-focused—not used outside Azure.

**How It Works**: Bicep is a domain-specific language that compiles to ARM JSON templates. ARM templates are deployed directly to Azure Resource Manager, which handles incremental deployment. There's **no client-side state file**—Azure itself knows what was last deployed and applies diffs.

**Strengths**:
- **100% Azure coverage on day one**: Every new Azure feature and resource type is available immediately in Bicep (Microsoft maintains the schemas).
- **No external state management**: Azure Resource Manager is the source of truth; you deploy a template, and Azure figures out what changed.
- **Excellent tooling**: Visual Studio Code Bicep extension provides autocomplete, validation, and export-from-portal features.

**Trade-offs**:
- **Azure-only**: Zero portability to other clouds. If your strategy is multi-cloud, Bicep won't help outside Azure.
- **No automatic drift correction**: Like Terraform, Bicep only reconciles on deployment. If someone changes a resource manually, Azure won't revert it unless you re-deploy.

**Best Fit**: Organizations standardized on Azure and wanting the tightest integration with Azure's own tooling. Great for Azure-centric teams with no multi-cloud requirements.

### Crossplane: Kubernetes-Native Infrastructure

**Community & Adoption**: Crossplane is a CNCF Incubating project with a smaller but enthusiastic community, especially among Kubernetes-heavy organizations. It's newer than Terraform but rapidly maturing, with production use at companies like IBM and Alibaba.

**How It Works**: Crossplane runs inside a Kubernetes cluster and treats infrastructure as Kubernetes Custom Resources. You define a `VirtualNetwork` CRD with your desired spec, submit it to the cluster, and Crossplane's controllers continuously reconcile that state with Azure. If someone changes the VNet manually in Azure, Crossplane detects the drift and reverts it.

**Strengths**:
- **Continuous reconciliation**: Unlike Terraform/Pulumi (which apply changes when you run them), Crossplane is always running, always converging to desired state.
- **GitOps integration**: Infrastructure definitions are YAML manifests in git, managed via tools like Flux or Argo CD.
- **Composition and abstraction**: You can define higher-level "composite resources" (e.g., a "Network" that automatically creates a VNet, subnets, NSGs) and expose them as self-service APIs to developers.

**Trade-offs**:
- **Requires a Kubernetes control plane**: You must operate a Kubernetes cluster to run Crossplane, which adds operational overhead.
- **Smaller ecosystem**: Fewer examples and modules than Terraform; some Azure features may lag provider updates.
- **Learning curve**: Requires understanding Kubernetes concepts (CRDs, controllers, reconciliation loops).

**Best Fit**: Platform engineering teams already running Kubernetes who want to offer infrastructure as Kubernetes APIs. Excellent for self-service platforms and GitOps workflows. Overkill if you just need to deploy a few static networks.

## Production Best Practices: What Actually Matters

The research into Azure VNet deployments reveals patterns that separate hobby projects from production infrastructure:

### Address Space Planning

**Plan for growth, avoid overlap**: Choose a CIDR block large enough to accommodate future expansion (a /16 like `10.0.0.0/16` supports 65,536 addresses) but ensure it doesn't overlap with on-premises networks, other VNets, or other cloud environments. Overlapping IPs prevent VNet peering and hybrid connectivity. Common practice: allocate each environment or region a portion of your organization's IP space (e.g., dev gets `10.10.0.0/16`, staging `10.20.0.0/16`, production `10.0.0.0/16`).

### Subnet Segmentation

**Segregate by role, not excessively**: Create subnets for meaningful boundaries—web tier, app tier, database tier, Kubernetes nodes. This allows applying Network Security Groups (NSGs) at the subnet level to control traffic. However, avoid creating dozens of tiny subnets; Azure recommends against overly small subnets because they add management overhead without significant security gain. Each subnet consumes 5 IPs for Azure's internal use, so extremely small subnets (/29, /28) leave few usable addresses. A practical approach: use /24 or /22 subnets and rely on NSGs for fine-grained control.

### Network Security Groups (NSGs)

**Adopt least privilege**: By default, Azure allows all traffic within a VNet. In production, use NSGs to restrict traffic to only required paths (e.g., web subnet can reach app subnet on port 443, app can reach database on port 5432, and nothing else). Avoid broad "allow all" rules—they defeat the security perimeter. Use **Application Security Groups (ASGs)** to simplify NSG management by grouping resources logically (e.g., all app servers in one ASG) rather than maintaining lists of IPs.

### Outbound Connectivity: NAT Gateway vs. Azure Firewall

For production workloads that need outbound Internet access (pulling updates, calling external APIs), you have two primary options:

- **Azure NAT Gateway**: Provides highly scalable outbound NAT with static public IPs and over 64,000 SNAT ports per IP. It's **outbound-only** (no inbound connections allowed, enhancing security) and cost-effective (~$32/month plus data charges). Use NAT Gateway when you need predictable egress IPs and high connection volumes without the overhead of a full firewall.

- **Azure Firewall**: A fully managed Layer 3-7 firewall that can handle outbound *and* inbound traffic, perform URL filtering, threat intelligence, and intrusion detection. It's significantly more expensive (~$910/month for Standard SKU plus data charges). Use Azure Firewall when security requirements demand deep packet inspection or centralized traffic control across many VNets.

Many production environments use **both**: Azure Firewall for security policies in a hub VNet, with NAT Gateway attached to the firewall's subnet to offload SNAT scaling.

### Hub-and-Spoke Architecture

For complex environments, the **hub-and-spoke** model is the recommended pattern. A central hub VNet contains shared services (Azure Firewall, NAT Gateway, VPN or ExpressRoute gateways, DNS relays), and multiple spoke VNets host isolated workloads. VNet peering connects spokes to the hub (peering is non-transitive, so spokes can communicate via the hub). This provides isolation (teams can manage their own spokes) while centralizing connectivity and governance in the hub.

### AKS-Specific Considerations

If your VNet will host an Azure Kubernetes Service (AKS) cluster, subnet sizing becomes critical:

- **Azure CNI (default)**: Every pod gets an IP from the VNet subnet. With the default 30 pods per node, a 50-node cluster needs approximately 1,600 IPs (50 nodes × 30 pods + 50 node IPs + buffer for upgrades). Microsoft recommends a **/21 subnet (2,048 addresses)** for a 50-node cluster. Undersizing the subnet will cause IP exhaustion and prevent pod scheduling.

- **Azure CNI Overlay** (newer option): Pods get IPs from an overlay network, not the VNet, so the subnet only needs to accommodate node IPs. This saves VNet address space but is a newer feature.

- **Kubenet** (older option): Similar to overlay—pods use non-VNet IPs. Simpler on IP usage but has limitations with Azure features like Private Link.

**Rule of thumb**: Provision generously for AKS. A /24 subnet (256 IPs) might seem large for 10 nodes, but with CNI and upgrades, it's reasonable. Always leave 20-30% free space in the VNet for adding subnets later.

## The 80/20 Configuration: What You Actually Need

Azure VNets expose dozens of configuration options, but most production deployments use a small subset:

**Essential (always required)**:
- **Address space** (e.g., `10.0.0.0/16`)
- **Subnets** with name and CIDR (e.g., `nodes-subnet: 10.0.0.0/18`)
- **Region** (e.g., `eastus`)
- **Tags** (environment, owner, cost-center)

**Common (frequently used)**:
- **NAT Gateway** (enable for outbound Internet access)
- **NSGs** (attached to subnets for security)
- **DNS servers** (custom DNS for hybrid scenarios; defaults to Azure DNS otherwise)

**Rare (advanced scenarios)**:
- **DDoS Protection Plan** (only for mission-critical public endpoints)
- **BGP community** (auto-set by Azure for ExpressRoute)
- **Encryption** (rarely enabled due to NIC support requirements)

Planton's API focuses on the essential and common fields, allowing teams to deploy production-ready VNets without navigating Azure's full complexity.

## How Planton Approaches Azure VNets

Planton uses **Pulumi** as the underlying IaC engine to deploy Azure VNets. Why Pulumi?

1. **Multi-cloud consistency**: Planton's protobuf-defined APIs work across AWS, GCP, Azure, and Kubernetes. Pulumi's support for multiple languages and clouds aligns with this philosophy.

2. **Developer-friendly**: Pulumi's use of general-purpose languages (TypeScript, Go, Python) allows Planton to embed complex logic (CIDR calculation, validation, defaults) within the provisioning code without extending a DSL.

3. **Production-ready**: Pulumi's `azure-native` provider offers full parity with Azure Resource Manager APIs, ensuring that all Azure features are available.

Planton abstracts Pulumi's complexity by exposing a minimal, opinionated API via protobuf:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureVpc
metadata:
  name: prod-vnet
spec:
  address_space_cidr: "10.0.0.0/16"
  nodes_subnet_cidr: "10.0.0.0/18"
  is_nat_gateway_enabled: true
  tags:
    environment: production
    team: platform
```

This configuration generates a production-ready VNet with:
- The specified address space and subnet
- An optional NAT Gateway for outbound connectivity
- Azure-recommended defaults for DNS, security, and networking

For teams deploying AKS clusters, Planton's `AzureVpc` resource is typically created as a dependency, ensuring that the cluster has a properly sized, secure network foundation without requiring deep Azure networking expertise.

## Conclusion

The journey from clicking through the Azure Portal to managing VNets as declarative, version-controlled code represents a fundamental shift in how we think about cloud infrastructure. Manual workflows are learning tools, not production strategies. Declarative IaC tools—whether Terraform for multi-cloud flexibility, Bicep for Azure-native simplicity, Pulumi for code-first infrastructure, or Crossplane for Kubernetes-native control planes—are now the standard.

Planton builds on these foundations by providing a simplified, opinionated abstraction layer. Teams get the production-readiness of Pulumi and Azure best practices without needing to master subnet CIDR math, NAT Gateway configuration, or NSG rule syntax. The result is infrastructure that's easier to deploy, safer to operate, and faster to understand—whether you're a platform engineer managing dozens of clusters or a developer who just needs a network to deploy into.

For teams building on Azure, the choice isn't whether to use IaC—it's which tool best fits your workflow. And for teams using Planton, that choice has already been made, letting you focus on what matters: shipping applications on production-ready infrastructure.


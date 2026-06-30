# Azure Container Registry: Deployment Approaches and Best Practices

## Introduction

Azure Container Registry (ACR) is Microsoft's fully managed Docker/OCI registry service, designed to store and manage private container images within the Azure ecosystem. Built on the open-source Docker Registry 2.0, ACR offers seamless integration with Azure services like Azure Kubernetes Service (AKS), Azure DevOps, and other container-based platforms.

The question isn't whether to use a container registry—that's a given for any containerized workload—but rather *how* to provision and manage it. The landscape has evolved from manual portal clicks to sophisticated infrastructure-as-code approaches, each with distinct trade-offs. For Azure-centric deployments, ACR stands out due to its tight integration with Azure AD for authentication, Azure Monitor for observability, and Azure's networking primitives for security.

This document explores the spectrum of deployment methods for Azure Container Registry, from basic manual provisioning to production-ready automation. We'll examine the maturity progression, compare infrastructure-as-code tools, and explain Planton's approach to abstracting ACR configuration while maintaining flexibility for production requirements.

## Understanding Azure Container Registry

### When to Use ACR vs. Alternatives

If you're already deployed on Azure, ACR offers compelling advantages:

- **Azure AD Integration**: Native support for role-based access control (RBAC) using Azure AD identities, service principals, and managed identities
- **Network Proximity**: Images stored in the same region as your compute resources eliminate cross-region bandwidth costs and latency
- **Seamless AKS Integration**: AKS can automatically authenticate to ACR using managed identities, eliminating manual credential management
- **Azure Ecosystem**: First-class integration with Azure DevOps, GitHub Actions (via OIDC), Azure Monitor, and Defender for Cloud

Docker Hub remains the largest public registry and works well for publicly distributed images, but it has rate limits and lacks cloud-specific integration. Amazon ECR and Google Artifact Registry serve analogous roles in their respective clouds. Harbor provides a powerful open-source alternative for self-hosted, cloud-agnostic deployments, though it shifts operational burden to your team.

The choice is straightforward: **use ACR for Azure-centric workloads** where integration, performance, and unified security matter. Consider alternatives only when multi-cloud portability trumps Azure-specific benefits.

### SKU Tiers: Choosing the Right Fit

ACR offers three SKUs that differ primarily in storage, performance, and advanced features:

| SKU | Storage | Throughput | Geo-Replication | Private Link | Typical Cost |
|-----|---------|------------|-----------------|--------------|--------------|
| **Basic** | 10 GB | ~1,000 pulls/min | No | No | ~$5/month |
| **Standard** | 100 GB | ~3,000 pulls/min | No | No | ~$20/month |
| **Premium** | 500 GB | ~10,000 pulls/min | Yes | Yes | ~$20/month + ~$50/month per replica |

**Basic** suits development environments, learning, and low-volume scenarios. It shares the same core API capabilities (Azure AD integration, webhooks, image deletion) but lacks advanced networking and geo-replication.

**Standard** serves most production workloads with substantially higher throughput and storage. It's the default choice when you don't need Premium-exclusive features.

**Premium** unlocks enterprise capabilities:
- **Geo-replication** across multiple Azure regions for global deployments
- **Private Link** and firewall rules for network isolation
- **Content Trust** (Docker Content Trust/Notary) for image signing
- **Customer-managed encryption keys** for compliance scenarios
- **Repository-scoped tokens** for fine-grained access control

Notably, Premium's base price roughly matches Standard when used in a single region, making it cost-effective if you need its performance or features even without geo-replication.

## The Deployment Maturity Spectrum

The journey from manual provisioning to production-grade automation follows a predictable evolution. Understanding this progression helps you choose the right approach for your team's maturity and requirements.

### Level 0: Manual Azure Portal Provisioning

**What it is:** Using the Azure Portal UI to create and configure ACR through a web browser.

**When to use it:** One-off experiments, learning ACR features, or quick prototypes.

**The pitfalls:**
- **Global Naming Conflicts**: Registry names must be globally unique across all of Azure (5-50 alphanumeric characters, no dashes). The Portal now offers a Domain Name Label (DNL) option that appends a unique hash to prevent DNS reuse, but this changes your pull URL permanently.
- **SKU Selection Mistakes**: Choosing Basic or Standard only to realize later you need geo-replication or private endpoints (Premium-only) requires migration and potential downtime.
- **Configuration Drift**: Manual changes aren't tracked in version control, making it impossible to reproduce environments or audit changes.
- **Admin User Temptation**: The Portal makes enabling the admin user (a single username/password credential) trivial, but this anti-pattern bypasses Azure AD auditing and creates a security liability.

**Verdict:** Acceptable for learning and throwaway environments, but avoid for anything that needs to be recreated, audited, or secured. Manual provisioning doesn't scale beyond initial exploration.

### Level 1: Scripting with Azure CLI

**What it is:** Automating ACR creation using the `az acr` command-line interface in scripts.

**Example:**

```bash
az acr create \
  --resource-group rg-prod \
  --name mycompanyacr \
  --sku Premium \
  --location eastus \
  --admin-enabled false
```

**What it solves:**
- **Repeatability**: Scripts can be version-controlled and rerun to provision identical registries
- **CI Integration**: Azure CLI works in CI/CD pipelines (Azure DevOps, GitHub Actions, GitLab CI)
- **Feature Access**: The CLI exposes all ACR capabilities, including ACR Tasks, geo-replication management, and image operations

**What it doesn't solve:**
- **State Management**: Scripts are imperative—they don't track what's deployed. Running the same script twice might error out or create duplicates.
- **Drift Detection**: If someone modifies the registry manually, your script won't detect or correct it.
- **Complexity at Scale**: Managing dependencies, ordering, and error handling across multiple resources becomes unwieldy.

Azure PowerShell (`New-AzContainerRegistry`) serves the same role for PowerShell-centric teams. Azure SDKs (Python, .NET, Java, Go) allow programmatic control but introduce similar challenges around state and orchestration.

**Verdict:** A step up from manual clicking, suitable for simple automation or as a complement to IaC for one-off operations. Not recommended as the primary provisioning method for production infrastructure.

### Level 2: Configuration Management (Ansible)

**What it is:** Using Ansible's `azure.azcollection.azure_rm_containerregistry` module to declare ACR configuration in YAML playbooks.

**Example:**

```yaml
- name: Create Azure Container Registry
  azure.azcollection.azure_rm_containerregistry:
    name: mycompanyacr
    resource_group: rg-prod
    location: eastus
    sku: Premium
    admin_user_enabled: false
    state: present
```

**What it solves:**
- **Idempotency**: Ansible checks current state and only makes changes if needed, making playbooks safely rerunnable
- **Declarative Style**: You specify desired state, not the steps to achieve it
- **Unified Tooling**: Teams already using Ansible for VM configuration can manage Azure resources in the same workflow

**What it doesn't solve:**
- **Limited Ecosystem**: Ansible's Azure modules may lag behind new Azure features compared to Azure-native tools
- **State Complexity**: While idempotent, Ansible doesn't provide robust dependency graphing or rollback like dedicated IaC tools
- **Multi-Cloud Focus**: Ansible's strength is breadth, not depth—Azure-specific features may require workarounds

**Verdict:** Viable for shops heavily invested in Ansible, especially when combining Azure provisioning with traditional configuration management. Not the first choice for greenfield Azure infrastructure.

### Level 3: Infrastructure as Code—The Production Approaches

This is where the landscape matures into truly production-ready options. Three tools dominate: **Terraform**, **Pulumi**, and **Azure Bicep/ARM**. All three provide declarative configuration, state tracking, and dependency management, but they differ in philosophy, ecosystem, and Azure integration depth.

## Infrastructure as Code: Comparative Analysis

### Terraform (with AzureRM Provider)

**Approach:** Declarative HCL (HashiCorp Configuration Language) with external state management.

**Strengths:**
- **Multi-Cloud Maturity**: If you manage infrastructure across Azure, AWS, and GCP, Terraform provides a consistent workflow and language
- **Ecosystem**: Extensive module registry, community support, and third-party integrations
- **Plan/Apply Workflow**: `terraform plan` shows a diff of changes before applying, providing confidence and audit trails
- **State Management**: Explicit state files (local or remote backends like Azure Storage) enable drift detection and collaborative workflows

**Considerations:**
- **Azure Feature Lag**: New Azure capabilities may take days or weeks to appear in the AzureRM provider
- **State Overhead**: Managing remote state backends, locking, and potential corruption requires operational discipline
- **Azure Policy Integration**: Policy violations are detected at apply time, not during plan, potentially causing late failures

**ACR Configuration Example:**

```hcl
resource "azurerm_container_registry" "main" {
  name                = "mycompanyacr"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  sku                 = "Premium"
  admin_enabled       = false

  georeplications {
    location = "westeurope"
    tags     = {}
  }

  georeplications {
    location = "southeastasia"
    tags     = {}
  }

  network_rule_set {
    default_action = "Deny"
  }
}
```

**Best for:** Multi-cloud enterprises, teams with existing Terraform expertise, or organizations requiring a cloud-agnostic IaC strategy.

### Pulumi

**Approach:** Infrastructure as code using general-purpose programming languages (TypeScript, Python, Go, C#).

**Strengths:**
- **Familiar Languages**: Developers can use loops, conditionals, and functions from their primary language without learning a DSL
- **Multi-Cloud Support**: Like Terraform, Pulumi spans clouds while offering Azure Native providers for full API coverage
- **Abstraction Flexibility**: Easier to build custom abstractions and reusable components in real code
- **State Management**: Pulumi Service (SaaS) or self-managed backends handle state similarly to Terraform

**Considerations:**
- **Determinism Burden**: General-purpose code can introduce non-deterministic behavior if not written carefully
- **Review Complexity**: Infrastructure changes require reviewing imperative code, which can be harder to audit than declarative configs
- **Smaller Ecosystem**: Fewer third-party modules and community resources compared to Terraform

**ACR Configuration Example (TypeScript):**

```typescript
import * as azure from "@pulumi/azure-native";

const acr = new azure.containerregistry.Registry("mycompanyacr", {
    resourceGroupName: resourceGroup.name,
    location: resourceGroup.location,
    sku: { name: "Premium" },
    adminUserEnabled: false,
});
```

**Best for:** Software-heavy teams, startups preferring to avoid DSLs, or scenarios requiring complex dynamic infrastructure logic.

### Azure Bicep / ARM Templates

**Approach:** Azure-native declarative templates (Bicep is a cleaner syntax that transpiles to ARM JSON).

**Strengths:**
- **Azure-First Integration**: Same-day support for new Azure features since Bicep compiles directly to ARM templates
- **No External State**: Azure Resource Manager is the source of truth; no separate state file to manage
- **Portal Export**: You can export existing resources as Bicep/ARM templates for reverse-engineering or migration
- **Azure Policy Preflight**: ARM deployments can detect policy violations before attempting changes, failing fast
- **Incremental Deployment**: ARM's incremental mode adds or updates resources without needing to track prior state explicitly

**Considerations:**
- **Azure-Only**: Bicep is Azure-specific; you can't use it for AWS or GCP (though you can call it from multi-cloud orchestrators)
- **Drift Handling**: No explicit drift detection like Terraform—redeployment overwrites differences without warning
- **Import Not Needed**: Unlike Terraform, you don't need to import manually created resources; ARM just updates them

**ACR Configuration Example (Bicep):**

```bicep
resource acr 'Microsoft.ContainerRegistry/registries@2023-01-01-preview' = {
  name: 'mycompanyacr'
  location: resourceGroup().location
  sku: {
    name: 'Premium'
  }
  properties: {
    adminUserEnabled: false
    publicNetworkAccess: 'Disabled'
  }
}

resource replication 'Microsoft.ContainerRegistry/registries/replications@2023-01-01-preview' = {
  parent: acr
  name: 'westeurope'
  location: 'westeurope'
}
```

**Best for:** Azure-only organizations, teams wanting first-party support and minimal external dependencies, or enterprises requiring tight Azure Policy integration.

## Production Best Practices

Regardless of which IaC tool you choose, production ACR deployments share common requirements:

### Authentication and Access Control

**Default to Azure AD**: Never rely on the admin user in production. Use Azure RBAC roles (`AcrPull`, `AcrPush`, `AcrDelete`) assigned to:
- **Managed Identities** for Azure services (AKS, Azure Container Instances, Azure Functions)
- **Service Principals** for CI/CD pipelines (Azure DevOps, GitHub Actions, GitLab CI)
- **User Identities** for developer access during testing

**Repository-Scoped Tokens**: For external systems or third-party integrations that can't use Azure AD, use ACR's repository-scoped tokens instead of the admin user. These tokens can be scoped to specific repositories and actions (pull-only, push/pull) for least-privilege access.

### Network Security

**Private Link (Premium)**: For production workloads in regulated industries, disable public network access and attach ACR to your VNet via private endpoints. This ensures all image pulls and pushes traverse Azure's private backbone, not the internet.

**IP Allowlisting (Premium)**: If Private Link is too restrictive, use IP firewall rules to whitelist known build agents or on-premises networks.

**Service Endpoints**: Azure VNet service endpoints (in preview for ACR) offer a middle ground, routing traffic over Azure's backbone without requiring private IPs.

### Geo-Replication for Global Deployments

If you operate AKS clusters or container workloads across multiple Azure regions:

1. **Enable Geo-Replication (Premium)** to replicate images to each deployment region
2. **Reduce Latency**: AKS nodes pull from the nearest replica, improving pod startup times
3. **Eliminate Bandwidth Costs**: Cross-region image pulls incur egress charges; replicas avoid this
4. **Increase Resilience**: If one region's storage is unavailable, other replicas continue serving images

Azure Traffic Manager automatically routes image pulls to the nearest healthy replica. You can monitor replication status via webhooks or Azure Event Grid.

### Security and Compliance

**Vulnerability Scanning**: Enable Microsoft Defender for Cloud integration to automatically scan images for CVEs upon push. Results surface in Azure Security Center with remediation guidance.

**Content Trust (Premium)**: For high-security environments (financial, healthcare), enable Docker Content Trust to require signed images using Notary.

**Customer-Managed Keys (Premium)**: If compliance mandates controlling encryption keys, configure ACR to use Azure Key Vault keys for at-rest encryption instead of Microsoft-managed keys.

### Lifecycle Management

**Retention Policies**: Enable automatic deletion of untagged manifests after 7-30 days to prevent storage bloat from CI/CD builds.

**Image Purge**: Use ACR Tasks with the `acr purge` command to delete old tagged images based on age or count (e.g., keep only the last 10 versions of each image).

**Immutable Tags**: Consider using semantic versioning or commit SHAs as image tags instead of mutable tags like `latest` to ensure reproducibility and traceability.

## The 80/20 Configuration Principle

When designing an API for ACR, the goal is to expose the 20% of configuration that covers 80% of use cases while allowing escape hatches for advanced scenarios.

### Essential Configuration (The 20%)

These fields satisfy the vast majority of deployments:

1. **`registry_name`**: Globally unique, 5-50 alphanumeric characters
2. **`sku`**: `BASIC`, `STANDARD`, or `PREMIUM` (default: `STANDARD`)
3. **`location`**: Azure region (e.g., `eastus`, `westeurope`)
4. **`admin_user_enabled`**: Boolean (default: `false`; enable only for dev/test)
5. **`geo_replication_regions`**: List of regions for Premium SKU (empty for Basic/Standard)

### Advanced Configuration (The 80%)

These settings are needed only for specific scenarios and can default to sensible values:

- **Network isolation**: Private endpoints, IP allowlists, or public access (default: public)
- **Retention policies**: Untagged manifest cleanup (default: disabled or 30 days)
- **Content trust**: Image signing enforcement (default: disabled)
- **Anonymous pull**: Public access to specific repositories (default: disabled)
- **Customer-managed keys**: BYO encryption keys via Key Vault (default: Microsoft-managed)
- **Repository permissions**: Azure AD ABAC for fine-grained repo-level access (default: registry-level RBAC)

### Example Configurations

**Dev/Test (Basic SKU):**

```yaml
registry_name: "myappacrdev"
sku: BASIC
admin_user_enabled: true  # Convenience for local docker login
```

**Production (Premium SKU with Geo-Replication):**

```yaml
registry_name: "mycompanyacr"
sku: PREMIUM
admin_user_enabled: false
geo_replication_regions:
  - "westeurope"
  - "southeastasia"
# Network isolation handled via separate private endpoint config
```

By focusing on these essential fields in the API, Planton enables users to provision ACR quickly while the underlying implementation applies production defaults (disabled admin user, retention policies, tags) automatically.

## Planton's Approach

Planton's `AzureContainerRegistrySpec` abstracts ACR configuration to its essential elements while leveraging Pulumi for deployment automation. This approach balances simplicity with production-readiness:

**Why Pulumi?**
- **Multi-Cloud Consistency**: Pulumi enables a unified deployment experience across Azure, AWS, GCP, and Kubernetes-native resources
- **Azure Native Coverage**: Pulumi's `@pulumi/azure-native` provider offers complete Azure API coverage with same-day feature support
- **Programmability**: Complex logic (conditional geo-replication, dynamic network rules) is easier to express in Go than HCL or Bicep
- **State Management**: Pulumi's state backend works seamlessly with Planton's workflow without requiring separate Terraform Cloud or Azure storage setup

**Abstraction Philosophy:**

The protobuf API exposes only the fields most users need to decide:
- **`registry_name`**: Required, validated against Azure naming rules
- **`sku`**: Defaults to `STANDARD` for production-ready performance
- **`admin_user_enabled`**: Defaults to `false` for security
- **`geo_replication_regions`**: Optional, validated to ensure Premium SKU when specified

Advanced features (network isolation, CMK encryption, retention policies) are configured via additional optional fields or sensible defaults. This keeps common deployments concise while allowing power users to override as needed.

**Infrastructure Philosophy:**

Rather than forcing users to choose between Terraform, Bicep, or Pulumi, Planton makes the choice for them based on our multi-cloud requirements, then exposes a cloud-agnostic API. Users specify *what* they want (a Premium ACR in East US with replicas in Europe and Asia), not *how* to provision it (which Pulumi resources, dependencies, and state to manage).

This abstraction allows the underlying implementation to evolve—switching from Pulumi to Bicep or adding Crossplane support—without breaking existing configurations.

## Cost Optimization

Azure Container Registry costs scale with SKU, storage, geo-replication, and bandwidth:

**SKU Selection**: Don't over-provision. Standard satisfies most production needs (~$20/month). Premium is justified only when you need geo-replication, private networking, or extreme throughput.

**Geo-Replication**: Each replica region adds ~$50/month. Only replicate to regions with actual deployments.

**Storage Cleanup**: Enable retention policies to delete untagged manifests after 7-30 days. Use `acr purge` tasks to remove old tagged images.

**Bandwidth**: Keep ACR and compute resources in the same region to avoid cross-region egress charges (which can be significant for large images pulled frequently).

**Defender for Cloud**: Security scanning costs ~$15/registry/month plus per-image fees. Enable only on production registries.

For a typical production setup (Premium ACR in one region with two replicas, 100 GB storage, minimal egress), expect ~$120-150/month—a reasonable cost for centralized, secure image management.

## Integration Patterns

### AKS (Azure Kubernetes Service)

The recommended approach is to attach ACR to AKS during cluster creation or via:

```bash
az aks update --attach-acr mycompanyacr
```

This grants the AKS cluster's managed identity the `AcrPull` role, eliminating the need for `imagePullSecrets` in Kubernetes manifests. AKS nodes authenticate automatically when pulling images.

For multi-region AKS deployments, use ACR geo-replication to place images near each cluster for faster pulls and lower bandwidth costs.

### CI/CD Pipelines

**Azure DevOps**: Use the built-in Azure service connection with a service principal granted `AcrPush`. The `Docker@2` task handles authentication and push automatically.

**GitHub Actions**: Use OpenID Connect (OIDC) federated credentials to authenticate without storing secrets:

```yaml
- uses: azure/login@v1
  with:
    client-id: ${{ secrets.AZURE_CLIENT_ID }}
    tenant-id: ${{ secrets.AZURE_TENANT_ID }}
    subscription-id: ${{ secrets.AZURE_SUBSCRIPTION_ID }}
- uses: azure/docker-login@v1
  with:
    login-server: mycompanyacr.azurecr.io
```

**GitLab CI**: Store service principal credentials in GitLab CI variables and use `az acr login` or repository-scoped tokens.

### ACR Tasks

ACR Tasks enable cloud-native image builds without maintaining dedicated build infrastructure:

- **On-Commit Builds**: Connect ACR to GitHub/Azure Repos to trigger builds on `git push`
- **Base Image Updates**: Automatically rebuild when upstream images (e.g., `mcr.microsoft.com/dotnet/runtime:6.0`) are updated
- **Multi-Architecture Images**: Build ARM64 and AMD64 variants in a single task for IoT or edge scenarios
- **Scheduled Builds**: Run nightly builds or periodic image purges via cron-like schedules

This offloads build complexity to Azure while keeping images close to storage for faster pushes.

## Common Anti-Patterns to Avoid

**Using the Admin User in Production**: This bypasses Azure AD auditing, creates a shared credential risk, and doesn't integrate with Azure RBAC. Always use managed identities or service principals.

**Ignoring Geo-Replication for Global Deployments**: Serving images from a single region to worldwide AKS clusters increases latency and bandwidth costs. Use Premium geo-replication to place images near workloads.

**No Image Lifecycle Management**: Letting CI/CD push images indefinitely without cleanup leads to storage bloat and unnecessary costs. Enable retention policies and periodic purges.

**Pulling Base Images from Docker Hub at Runtime**: Relying on external registries (especially Docker Hub with its rate limits) in production is fragile. Import base images into ACR and reference them locally.

**Choosing Basic for Production**: Basic's lower throughput (~1,000 pulls/min) can throttle busy deployments. Standard or Premium ensures headroom for scaling.

**Not Scanning Images**: Pushing unscanned images to production risks deploying known vulnerabilities. Enable Defender for Cloud or integrate a scanner (Trivy, Aqua) into CI.

## Conclusion

Azure Container Registry deployment has matured from manual portal provisioning to sophisticated infrastructure-as-code orchestration. While Azure CLI scripts and Ansible playbooks serve specific niches, production workloads demand the state management, drift detection, and collaboration features of true IaC tools.

For Azure-only teams, **Bicep** offers first-party support, zero state overhead, and same-day feature availability. For multi-cloud enterprises, **Terraform** provides a battle-tested, cloud-agnostic workflow. For development-heavy teams, **Pulumi** bridges infrastructure and application code in familiar languages.

Planton abstracts this choice, using **Pulumi** internally to deliver a consistent multi-cloud API while applying production defaults (disabled admin user, appropriate SKU selection, geo-replication configuration) automatically. This lets users focus on *what* they need—a secure, performant container registry—rather than *how* to wire together Azure Resource Manager primitives.

By understanding the deployment spectrum and applying the 80/20 principle to configuration, you can provision Azure Container Registry that is simple enough for rapid iteration yet robust enough for global production deployments. The registry becomes invisible infrastructure: always available, seamlessly integrated, and secure by default.

---

**Further Reading:**

- [Azure Container Registry SKU Tiers](https://learn.microsoft.com/en-us/azure/container-registry/container-registry-skus) - Official Microsoft documentation on feature and pricing differences
- [AKS + ACR Integration](https://learn.microsoft.com/en-us/azure/aks/cluster-container-registry-integration) - Best practices for connecting AKS clusters to ACR
- [ACR Best Practices](https://learn.microsoft.com/en-us/azure/container-registry/container-registry-best-practices) - Microsoft's production recommendations
- [Terraform vs. Bicep Comparison](https://learn.microsoft.com/en-us/azure/developer/terraform/comparing-terraform-and-bicep) - Official guidance on choosing between Azure-native and multi-cloud IaC


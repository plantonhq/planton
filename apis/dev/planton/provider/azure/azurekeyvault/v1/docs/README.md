# Azure Key Vault Deployment Methods

## Introduction

"Just put the secrets in a config file for now," said every developer ever — until the inevitable security audit comes knocking. Azure Key Vault has quietly become one of the most critical infrastructure pieces in modern Azure deployments, yet the path to deploying it properly is littered with pitfalls that seem obvious only in hindsight.

The promise is simple: centralize your secrets, keys, and certificates in a managed service with military-grade security and seamless Azure integration. The reality is more nuanced. How do you deploy the vault itself? What access model should you use? How locked-down should the network be? What about multi-tenant scenarios or compliance requirements?

This document maps out the full spectrum of Azure Key Vault deployment approaches — from the quick-and-dirty portal clicks that work for demos to the production-hardened Infrastructure-as-Code patterns that pass SOC 2 audits. We'll explore what works, what doesn't, and why Planton defaults to certain choices for its Azure Key Vault implementation.

## The Maturity Spectrum: From Manual Clicks to Production IaC

### Level 0: The Portal Deployment (Quick Start, Not Production)

The Azure Portal offers the fastest path to a Key Vault: fill out a form, click "Create," and you're done in two minutes. The portal will even helpfully add an access policy giving your user account full permissions.

This approach is perfectly fine for learning, prototyping, or isolated development environments. But it reveals its limitations quickly:

- **Configuration Drift**: Manual deployments are impossible to replicate consistently across environments. Your dev vault might have different settings than prod, and you won't notice until something breaks.

- **Missing Security Defaults**: The portal makes it easy to skip critical production settings. Purge protection? Easy to forget. Network restrictions? That's an extra click you might skip. Before you know it, your vault is wide open to the public internet with no audit logging.

- **Global Name Collisions**: Key Vault names must be globally unique across all of Azure (they become DNS names like `your-vault.vault.azure.net`). Manual provisioning means trial-and-error until you find an unused name — not ideal in automated pipelines.

- **No Change Tracking**: When something goes wrong, there's no Git history to review. Who changed the network rules? When was purge protection disabled? The portal doesn't remember.

**Verdict**: Portal deployment is the training wheels version. Use it to understand Key Vault's capabilities, then graduate to automation.

### Level 1: CLI and PowerShell Scripts (Repeatable, But Imperative)

Azure CLI and PowerShell bring repeatability through scripting. You can capture your vault configuration as code:

```bash
az keyvault create \
  --name myapp-prod-kv \
  --resource-group myapp-rg \
  --location eastus \
  --sku standard \
  --enable-soft-delete true \
  --enable-purge-protection true \
  --enable-rbac-authorization true
```

This is a substantial improvement over manual clicking. Scripts can be version-controlled, reviewed, and run repeatedly. DevOps pipelines can execute them as deployment steps.

The challenge is that CLI/PowerShell scripts are **imperative** — they describe steps, not desired state. This creates subtle issues:

- **Idempotency concerns**: Running the same script twice might fail or create unexpected duplicates unless you carefully handle `create-or-update` semantics.

- **Dependency management**: You have to manually orchestrate ordering. Create the resource group, then the vault, then the access policies, then the private endpoint. Miss a dependency and the script fails cryptically.

- **State tracking**: The script doesn't inherently know what already exists. You end up writing conditional logic: "If the vault doesn't exist, create it; otherwise, update it."

- **Limited validation**: Syntax errors might not surface until runtime, potentially failing mid-deployment.

For teams comfortable with shell scripting and CI/CD orchestration, this approach can work. But it requires discipline and careful error handling to make scripts truly production-ready.

**Verdict**: Scripts are a step up from manual, suitable for small-scale automation. But they lack the declarative elegance and state management of true Infrastructure-as-Code.

### Level 2: Azure Resource Manager Templates and Bicep (Azure-Native IaC)

Azure Resource Manager (ARM) templates are JSON files that declaratively specify infrastructure. Bicep is Microsoft's newer domain-specific language that compiles to ARM but reads like actual code instead of JSON soup.

A Bicep template for Key Vault might look like:

```bicep
resource keyVault 'Microsoft.KeyVault/vaults@2023-02-01' = {
  name: 'myapp-${uniqueString(resourceGroup().id)}-kv'
  location: resourceGroup().location
  properties: {
    tenantId: subscription().tenantId
    sku: {
      family: 'A'
      name: 'standard'
    }
    enablePurgeProtection: true
    enableRbacAuthorization: true
    networkAcls: {
      defaultAction: 'Deny'
      bypass: 'AzureServices'
    }
  }
}
```

ARM/Bicep deployments are **idempotent** and **declarative**: you describe what you want, Azure figures out how to get there. Re-run the same template, and Azure makes only the necessary changes.

Key advantages:

- **Built-in validation**: ARM validates templates before deployment, catching many errors early.

- **Atomic deployments**: If any resource in the template fails, the whole deployment rolls back (in complete mode).

- **Parameter files**: Separate environment-specific values from infrastructure definitions, enabling easy dev/staging/prod variations.

- **Deep Azure integration**: ARM templates can reference other Azure resources, fetch secrets at deployment time (if `enabledForTemplateDeployment` is set), and leverage Azure's deployment engine features.

The limitation is ecosystem lock-in: ARM/Bicep is Azure-only. If your organization has multi-cloud infrastructure or wants to use the same IaC tool across providers, you'll look elsewhere.

**Verdict**: ARM/Bicep is the production-ready choice for Azure-centric teams. It's mature, well-supported, and aligns with Azure's design philosophy. Use Bicep over raw ARM for better readability.

### Level 3: Terraform and OpenTofu (Multi-Cloud Declarative IaC)

HashiCorp Terraform has become the de facto standard for multi-cloud infrastructure management. Using the AzureRM provider, you can define Key Vault alongside AWS, GCP, and Kubernetes resources in a single configuration.

A Terraform Key Vault resource looks like:

```hcl
resource "azurerm_key_vault" "main" {
  name                        = "myapp-prod-kv"
  location                    = azurerm_resource_group.main.location
  resource_group_name         = azurerm_resource_group.main.name
  tenant_id                   = data.azurerm_client_config.current.tenant_id
  sku_name                    = "standard"
  purge_protection_enabled    = true
  soft_delete_retention_days  = 90
  
  enable_rbac_authorization   = true
  
  network_acls {
    default_action = "Deny"
    bypass         = "AzureServices"
  }
}
```

Terraform brings several production benefits:

- **State management**: Terraform tracks deployed infrastructure in a state file, enabling accurate drift detection and updates.

- **Dependency graphs**: Terraform automatically determines resource dependencies and parallelizes deployments where possible.

- **Plan before apply**: The `terraform plan` command shows exactly what will change before you commit, reducing surprise failures.

- **Module ecosystem**: Reusable Terraform modules encapsulate best practices, so you can deploy a compliant Key Vault with a single module reference.

- **Multi-cloud portability**: The same tool manages Azure, AWS, GCP, and dozens of other providers.

**OpenTofu** is the open-source fork of Terraform (created after HashiCorp's license change in 2023). It's functionally identical for most use cases, using the same AzureRM provider and HCL syntax. Organizations concerned about licensing or seeking community governance are adopting OpenTofu.

Critical considerations:

- **Secret handling**: Never put actual secret values in Terraform code. Use variables marked `sensitive` and inject values at runtime from secure sources (Azure DevOps variable groups, environment variables, etc.). Terraform state can contain sensitive data, so encrypt and restrict access to state files.

- **Access policy vs RBAC**: Terraform supports both Key Vault access models. Modern deployments should use RBAC (`enable_rbac_authorization = true`) and manage permissions via `azurerm_role_assignment` resources for better integration with Azure's identity management.

**Verdict**: Terraform/OpenTofu is the gold standard for teams with multi-cloud infrastructure or strong preferences for vendor-neutral tooling. It offers production-grade declarative IaC with a massive ecosystem.

### Level 4: Pulumi (IaC in Real Programming Languages)

Pulumi takes a different approach: write infrastructure code in TypeScript, Python, Go, C#, or Java instead of learning a DSL. For developers who think in for-loops and conditionals rather than declarative YAML, this can feel natural.

A TypeScript Pulumi program for Key Vault:

```typescript
import * as azure_native from "@pulumi/azure-native";

const vault = new azure_native.keyvault.Vault("keyVault", {
  resourceGroupName: resourceGroup.name,
  properties: {
    tenantId: config.require("azureTenantId"),
    sku: { name: "standard", family: "A" },
    enablePurgeProtection: true,
    enableRbacAuthorization: true,
    networkAcls: {
      defaultAction: "Deny",
      bypass: "AzureServices",
    },
  },
});
```

Pulumi excels when:

- **Complex logic is needed**: Generating multiple vaults programmatically, dynamically constructing access policies based on external data, or integrating with APIs during deployment.

- **Developer preference**: Teams already fluent in TypeScript/Python can hit the ground running without learning HCL or Bicep.

- **Secrets management**: Pulumi has built-in encrypted secrets in its config system, avoiding plaintext credential exposure.

Trade-offs:

- **Smaller ecosystem**: Pulumi's community is growing but still smaller than Terraform's. Fewer examples and modules available.

- **State management**: Pulumi requires managing state (either via Pulumi Cloud SaaS or self-hosted backends).

- **Learning curve for ops teams**: If your operations team is comfortable with HCL but not Python, Pulumi might be a harder sell.

**Verdict**: Pulumi is a strong choice for developer-centric teams that value using real programming languages for infrastructure. It's production-ready but less ubiquitous than Terraform or ARM.

### Level 5: Configuration Management Tools (Ansible, Chef, Puppet)

Tools like Ansible can manage Azure Key Vault via Azure collection modules:

```yaml
- name: Create Key Vault
  azure.azcollection.azure_rm_keyvault:
    resource_group: "myapp-rg"
    name: "myapp-prod-kv"
    location: "eastus"
    sku: standard
    tenant_id: "{{ azure_tenant_id }}"
    purge_protection_enabled: yes
    enable_rbac_authorization: yes
```

These tools originated in configuration management (installing software on servers) but have added cloud resource provisioning capabilities. They work, but feel like square pegs in round holes:

- **Imperative roots**: While Ansible strives for idempotency, its modules can be less robust than purpose-built IaC tools.

- **State handling**: Ansible doesn't maintain persistent state like Terraform, relying on querying Azure's current state each run.

- **Limited adoption for cloud**: Most teams use Ansible for VM configuration and Terraform/ARM for cloud resource provisioning.

**Verdict**: Use configuration management tools if you're already heavily invested in them, but dedicated IaC tools (Terraform, ARM/Bicep) are better suited for Azure infrastructure.

### Level 6: Advanced Abstractions (Crossplane, Azure CDK)

**Crossplane** allows managing Azure resources via Kubernetes Custom Resources. You declare a Key Vault as a Kubernetes YAML manifest, and Crossplane reconciles it into Azure.

This is powerful for teams operating in a "Kubernetes as control plane" model, but introduces significant complexity. Most organizations won't need this level of abstraction unless pursuing advanced multi-cloud or GitOps patterns.

**Azure CDK** (Cloud Development Kit) is experimental, aiming to provide an AWS CDK-like experience for Azure. It's not yet mature enough for production recommendation.

**Verdict**: Bleeding edge, suitable only for teams with specific architectural needs (like Kubernetes-centric control planes).

## The Production Decision: What Actually Matters

After surveying the landscape, three deployment methods emerge as production-ready:

1. **ARM/Bicep**: Best for Azure-native teams, tight integration with Azure ecosystem, official Microsoft support.

2. **Terraform/OpenTofu**: Best for multi-cloud environments, mature tooling, vendor-neutral, huge community.

3. **Pulumi**: Best for developer-centric teams comfortable with real programming languages.

The choice depends less on technical superiority and more on **organizational fit**:

- **Do you already have Terraform/OpenTofu expertise?** Use that. Don't introduce Bicep just for Key Vault.

- **Are you Azure-only with Microsoft-centric operations?** ARM/Bicep aligns better with Azure DevOps pipelines and Microsoft support channels.

- **Do developers own infrastructure in your org?** Pulumi might feel more natural than learning HCL.

## What Planton Supports

Planton's Azure Key Vault implementation uses **Pulumi** as the deployment engine. This decision reflects several priorities:

**Multi-Cloud Consistency**: Planton provides a unified interface across AWS, Azure, GCP, and Kubernetes. Pulumi's multi-language, multi-cloud approach aligns with this philosophy better than Azure-specific tooling.

**Developer Accessibility**: Pulumi's use of familiar programming languages (TypeScript, Python, Go) lowers the barrier for developers who need to understand or extend infrastructure code. The same team writing application code can read and modify infrastructure definitions.

**Abstraction Layer**: Planton's protobuf-based API abstracts the underlying IaC tool. Users define Key Vault configurations in a provider-neutral spec, and Planton's engine translates to Pulumi code. This allows swapping implementations without changing user-facing APIs.

**Production Proven**: Pulumi is production-ready, well-funded, and actively developed. Its Azure Native provider closely tracks Azure's API surface.

While Terraform might have a larger market share and ARM/Bicep might be "more Azure-native," Pulumi offers the best balance of **expressiveness, multi-cloud portability, and developer ergonomics** for Planton's architecture.

Users of Planton don't write Pulumi directly — they define a `AzureKeyVault` resource in protobuf with high-level configuration (secrets to create, RBAC settings, network rules), and Planton generates and executes the Pulumi code behind the scenes.

## The 80/20 Configuration Philosophy

Research into real-world Azure Key Vault deployments reveals that **80% of users need only 20% of available configuration options**. Planton's API design reflects this insight.

**Essential Configuration** (always needed):
- **Vault Name**: Globally unique identifier
- **Location/Region**: Azure region for deployment
- **SKU**: Standard (software-protected) vs Premium (HSM-backed)
- **Tenant ID**: Azure AD tenant (usually implicit from context)
- **Access Control Model**: RBAC vs Access Policies (strongly prefer RBAC)
- **Purge Protection**: Should always be enabled in production
- **Soft Delete Retention**: Defaults to 90 days

**Network Security** (critical for production):
- **Public Network Access**: Enabled or Disabled
- **Default Action**: Allow or Deny for network ACLs
- **IP Allowlist**: Specific IPs/CIDRs allowed to access
- **VNet/Subnet Rules**: Virtual networks allowed access
- **Bypass Azure Services**: Allow trusted Azure services to access vault

**Rare/Advanced Configuration** (excluded from minimal spec):
- `enabledForDeployment`, `enabledForDiskEncryption`, `enabledForTemplateDeployment`: Specific to VM encryption scenarios
- Certificate Authority integrations: Enterprise PKI scenarios
- HSM-specific settings: Only needed for Premium SKU compliance use cases
- Custom rotation policies: Advanced automation

Planton's protobuf spec focuses on the essential and network security fields, providing sensible defaults for everything else. This keeps the API surface small and learnable while still enabling production-hardened deployments.

## Deployment Patterns: Development vs Production

### Development Vault Configuration

For non-production environments, prioritize **speed and accessibility** over maximum security:

- **SKU**: Standard (no need for HSM costs in dev)
- **Access Control**: Access Policies can be simpler for small dev teams (though RBAC is still recommended for consistency)
- **Network**: Public access enabled with IP restrictions to office/VPN ranges
- **Purge Protection**: Optional (disabled allows cleanup/testing of deletion)
- **Logging**: Nice to have, but not critical

**Example minimal dev config**:
```yaml
sku: standard
enable_rbac_authorization: false
network_acls:
  default_action: Deny
  ip_rules: ["203.0.113.0/24"]  # Office IP range
purge_protection_enabled: false
```

### Production Vault Configuration

Production environments demand **defense in depth**:

- **SKU**: Standard for most applications, Premium if compliance requires HSM (PCI-DSS, FIPS 140-3)
- **Access Control**: Azure RBAC exclusively (better integration, centralized management)
- **Network**: Private endpoint with no public access, or VNet rules with firewall
- **Purge Protection**: Always enabled (prevents accidental permanent deletion)
- **Soft Delete**: Enabled with 90-day retention (default)
- **Logging**: Mandatory, shipped to Log Analytics or SIEM
- **Monitoring**: Alerts on failed access attempts, configuration changes, secret near-expiry

**Example production config**:
```yaml
sku: standard
enable_rbac_authorization: true
network_acls:
  default_action: Deny
  public_network_access: Disabled
  # Private endpoint created separately
purge_protection_enabled: true
soft_delete_retention_days: 90
tags:
  Environment: Production
  Compliance: SOC2
```

### Enterprise Vault Configuration

For highly regulated industries (finance, healthcare, government):

- **SKU**: Premium (HSM-backed keys for maximum security)
- **Access Control**: RBAC with just-in-time access (Azure AD Privileged Identity Management)
- **Network**: Private endpoint only, no public internet access ever
- **Multi-tenant separation**: One vault per customer or sensitive domain
- **Encryption**: Customer-managed keys (BYOK) where required
- **Compliance**: Aligned with SOC 2, HIPAA, PCI-DSS, FedRAMP controls
- **Backup**: Offline backups of critical HSM keys stored securely

**Example enterprise config**:
```yaml
sku: premium
enable_rbac_authorization: true
network_acls:
  default_action: Deny
  public_network_access: Disabled
purge_protection_enabled: true
tags:
  Environment: Production
  DataClassification: Highly-Sensitive
  Compliance: HIPAA-PCI
```

## Security and Compliance Essentials

### Access Control: RBAC vs Access Policies

Azure Key Vault offers two authentication models:

**Access Policies** (legacy): Define granular permissions (Get, List, Set, Delete) for specific Azure AD principals directly on the vault. Simple for small teams but doesn't scale well.

**Azure RBAC** (recommended): Use standard Azure role assignments (Key Vault Secrets User, Key Vault Administrator, etc.). Integrates with Azure AD groups, Privileged Identity Management, and conditional access policies.

**Recommendation**: Use RBAC for all new deployments. It provides better integration with enterprise identity management and conditional access policies.

### Network Security

**Never leave a production Key Vault open to the public internet without restrictions.**

Options in order of security:

1. **Private Endpoint** (most secure): Vault accessible only via private IP in your VNet. No public internet route exists.

2. **VNet Service Endpoints + Firewall**: Restrict access to specific Azure VNets and/or IP ranges.

3. **IP Allowlist**: Permit only known IP addresses (office, VPN gateways, CI/CD runners).

Always set `bypass: AzureServices` to allow trusted Azure services (Azure Monitor, Backup, etc.) to access the vault even when firewalled.

### Secret Management Anti-Patterns to Avoid

❌ **Hardcoding secrets in code or IaC templates**: Never put actual secret values in Git. Use secure variable injection at deployment time.

❌ **Storing vault URLs in application config**: Vault URLs are not secrets, but storing them separately from the code that uses them creates unnecessary coupling. Use managed identities and Key Vault references instead.

❌ **Disabling soft delete and purge protection**: These features prevent accidental data loss and should always be enabled in production.

❌ **Granting overly broad permissions**: Don't give every service "Key Vault Administrator" when it only needs to read a single secret. Use least privilege.

❌ **Sharing vaults across environments**: Never put dev, staging, and production secrets in the same vault. Separate vaults provide isolation and reduce blast radius.

### Compliance Certifications

Azure Key Vault is certified for major compliance frameworks:

- **FIPS 140-2/140-3**: Standard tier (Level 1), Premium tier (Level 3)
- **SOC 1/2/3, ISO/IEC 27001**
- **PCI-DSS**: HSM-backed keys in Premium tier support compliance
- **HIPAA/HITRUST**: Suitable for healthcare PHI encryption keys
- **FedRAMP**: Available in Azure Government regions

Enabling audit logging and ensuring proper access controls are critical for compliance audits.

## Secret Lifecycle Management

### Versioning and Rotation

Azure Key Vault automatically versions every secret. When you update a secret, the old value remains accessible via its version ID. Applications fetching secrets without a version always get the latest.

**Best practice rotation flow**:
1. Generate new secret value (via automation or manual process)
2. Add as new version to Key Vault (old version still active)
3. Update applications to use new version (or rely on automatic refresh)
4. Monitor for any services still using old version
5. Disable old version after transition period
6. Optionally delete old version after retention period

### Integration with Applications

**Azure-native services** (App Service, Functions, AKS):
- Use **Key Vault References** in application settings: `@Microsoft.KeyVault(SecretUri=https://vault.vault.azure.net/secrets/DbPassword)`
- Use **CSI Secret Store Driver** in AKS to mount secrets as files or sync to Kubernetes Secrets
- Leverage **managed identities** for authentication (no credentials in code)

**Custom applications**:
- Use Azure SDK for Key Vault (available in .NET, Python, Java, Node.js, Go)
- Authenticate via `DefaultAzureCredential` (discovers managed identity, service principal, or CLI credentials)
- Cache secrets in memory to reduce vault calls and latency
- Subscribe to Azure Event Grid events for secret rotation notifications

## Conclusion

The journey from "just store it in a config file" to a production-hardened secret management system is longer than it appears. Azure Key Vault provides the security foundation, but deploying it correctly requires understanding the tradeoffs between access models, network configurations, and compliance requirements.

The proliferation of deployment tools — Portal, CLI, ARM/Bicep, Terraform, Pulumi, Ansible, Crossplane — reflects the diverse needs of organizations. There's no single "best" choice, only the best choice **for your team's skills and constraints**.

Planton's use of Pulumi strikes a balance: production-proven, developer-friendly, and multi-cloud consistent. By abstracting Key Vault configuration into a simple protobuf API, it allows teams to deploy compliant vaults without becoming Azure or IaC experts.

The paradigm shift in modern cloud security is treating secrets as first-class infrastructure, managed with the same rigor as compute and networking. Azure Key Vault, deployed thoughtfully with Infrastructure-as-Code, makes that paradigm real. Your secrets deserve better than a `.env` file in Git — give them the vault they deserve.


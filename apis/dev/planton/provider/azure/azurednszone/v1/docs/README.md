# Azure DNS Zone Deployment Guide

## Introduction

When Azure first launched its DNS service, the conventional wisdom was that it was just another managed DNS offering in a crowded market. Why move away from established providers like Route 53, Cloudflare, or your domain registrar's DNS?

The reality today is quite different. Azure DNS has matured into a strategic component for cloud-native architectures, not merely a commodity service. It offers seamless Azure integration, global anycast infrastructure, and unique features like **alias records** (eliminating hardcoded IPs), **private DNS zones** (for internal VNet resolution), and **DNSSEC support** (for cryptographic validation). For organizations building on Azure, hosting DNS outside the platform means managing yet another integration point, maintaining separate access controls, and missing out on unified infrastructure-as-code.

This document examines how to deploy Azure DNS zones effectively – from anti-patterns to production-ready solutions – and explains Planton's approach to DNS management in a multi-cloud world.

## The Azure DNS Landscape

### What is Azure DNS?

Azure DNS is Microsoft's managed authoritative DNS hosting service. It allows you to host DNS zones and records using Azure's global network, with the same credentials, APIs, and billing as other Azure services. Key characteristics:

- **High availability**: Anycast-based DNS with answers from multiple global locations
- **Public zones**: For internet-facing domains (e.g., `example.com`)
- **Private zones**: For internal VNet name resolution (e.g., `internal.corp.local`)
- **Azure integrations**: Alias records for Azure resources, RBAC for access control, Private Endpoints support
- **DNSSEC**: Cryptographic signing support (GA as of 2025)

**Critical distinction**: Azure DNS is **not** a domain registrar. You must purchase domains elsewhere (GoDaddy, Namecheap, etc.) and then delegate nameserver authority to Azure DNS.

### Azure DNS vs. Alternatives

**AWS Route 53**  
Route 53 combines DNS hosting with advanced traffic routing (latency-based, weighted, health-check failover). In Azure, those routing features live in a separate service (Traffic Manager) while Azure DNS handles domain hosting. Both support DNSSEC. Route 53 can also act as a domain registrar; Azure DNS cannot.

**Google Cloud DNS**  
Similar feature set to Azure DNS (global anycast, DNSSEC, zone management). Both are PaaS offerings with comparable performance. GCP integrates with Google Domains for registration; Azure requires external registrars. Choose based on your primary cloud platform.

**Cloudflare DNS**  
Cloudflare operates one of the world's fastest DNS networks (300+ PoPs) and offers built-in DDoS protection, CNAME flattening at the apex, and CDN/proxy capabilities. Unlike Azure DNS which is purely authoritative DNS, Cloudflare can also front your traffic. Cloudflare excels for globally distributed, multi-cloud apps. Azure DNS excels for Azure-centric workloads with tight integration.

### Public vs. Private DNS Zones

**Public DNS Zones** are accessible from the Internet. After creating a public zone in Azure, you delegate your domain to Azure's nameservers (e.g., `ns1-01.azure-dns.com`) at your registrar. Azure provides four nameserver addresses per zone for redundancy. Note: Azure does **not** support vanity nameservers – you must use the Azure-provided NS names.

**Private DNS Zones** are only resolvable within your Azure Virtual Networks. They're used for internal naming:
- Internal application domains (e.g., `app.internal.corp.com`)
- Azure Private Endpoint domains (e.g., `privatelink.database.windows.net` for private Azure SQL connections)

Private zones must be linked to one or more VNets. You can enable **auto-registration** so VMs in a VNet automatically get DNS records when they boot – eliminating manual DNS management for dynamic VM deployments.

## The Maturity Spectrum: Deployment Methods

Deploying Azure DNS zones spans a spectrum from manual portal clicks to fully automated infrastructure-as-code. Here's how approaches evolve from anti-patterns to production-ready solutions.

### Level 0: The Anti-Pattern (Manual Portal Management)

**Approach**: Create DNS zones and records through the Azure Portal UI. Search for "DNS zones", click through the blade, manually enter each record set.

**What it solves**: Quick exploration, proof-of-concept, one-off domains.

**What it doesn't**: 
- **No version control**: Changes aren't tracked. Who changed the A record for `api.example.com` last Tuesday? Unknown.
- **Error-prone**: Forgetting to click "Save" after entering record data is a common pitfall. Records silently disappear.
- **Team coordination nightmare**: Multiple admins clicking in the portal leads to configuration drift and conflicts.
- **No automation**: Scaling to dozens of records or multiple zones becomes tedious and risky.

**Verdict**: Acceptable only for personal projects or early prototyping. Never for production DNS that impacts revenue or security.

### Level 1: Scripted Imperative Changes (CLI/PowerShell)

**Approach**: Use Azure CLI (`az network dns`) or PowerShell (`New-AzDnsZone`) to script DNS operations.

```bash
az network dns zone create -g prod-network-rg -n example.com
az network dns record-set a add-record -g prod-network-rg -z example.com \
  -n www --ipv4-address 203.0.113.10
```

**What it solves**: 
- Repeatable commands (can document or version-control the scripts)
- Bulk operations (script a loop to add 50 records)
- CI/CD integration (run scripts in pipelines)

**What it doesn't**:
- **State management**: Scripts don't know current state. If a record already exists, the script may fail or duplicate entries unless you add conditional logic.
- **Drift detection**: If someone manually deletes a record, the script won't notice or recreate it on the next run.
- **Complexity**: Managing complex DNS with scripts (checking existence, handling updates, deleting stale records) quickly becomes unwieldy.

**Verdict**: Step up from manual, useful for one-time migrations (like bulk importing records from another provider) or embedding DNS updates in deployment scripts. Not ideal as the primary DNS management strategy.

### Level 2: Zone File Import/Export (Traditional DNS Workflow)

**Approach**: Maintain DNS in a BIND-style zone file and use Azure CLI to import/export.

```bash
# Export current state
az network dns zone export -g prod-network-rg -n example.com -f example.com.zone

# Edit the zone file, then import
az network dns zone import -g prod-network-rg -n example.com -f example.com.zone
```

**What it solves**:
- **Familiar format**: DNS admins know BIND zone files. Portable across providers.
- **Bulk changes**: Add 100 records in a text editor, import once.
- **Version control**: Store zone files in Git. Diff changes before applying.

**What it doesn't**:
- **Merge behavior**: Azure import **merges** with existing records, which can be confusing. If you remove a record from the zone file and re-import, the record remains in Azure unless explicitly deleted first.
- **Azure-specific limitations**: SOA and NS records are special. Azure uses its own nameservers; zone file NS data is ignored. Multi-string TXT records (common for SPF/DKIM) can have quirks on import.
- **No automated reconciliation**: Still requires manual process to apply. If someone changes DNS outside the zone file process, next import could conflict.

**Verdict**: Valid for organizations with existing DNS workflows or multi-cloud scenarios (same zone file format across providers). Works well when combined with CI/CD to automate imports. Not as seamless as declarative IaC but better than raw CLI.

### Level 3: Declarative Infrastructure as Code (Production Standard)

**Approach**: Define DNS zones and records as code using Terraform, Pulumi, OpenTofu, or Azure-native tools (ARM/Bicep). Treat DNS like any other infrastructure resource.

#### Terraform / OpenTofu

```hcl
resource "azurerm_dns_zone" "example" {
  name                = "example.com"
  resource_group_name = azurerm_resource_group.network.name
}

resource "azurerm_dns_a_record" "www" {
  name                = "www"
  zone_name           = azurerm_dns_zone.example.name
  resource_group_name = azurerm_resource_group.network.name
  ttl                 = 3600
  records             = ["203.0.113.10"]
}
```

- **State management**: Terraform tracks current state. Run `terraform plan` to see exactly what will change before applying.
- **Drift detection**: If someone manually adds/removes records, Terraform detects drift and can reconcile.
- **Modularity**: Define reusable modules for DNS patterns (e.g., "standard web domain" module with apex A, www CNAME, MX, SPF).

#### Pulumi (Azure Native)

```typescript
import * as azure_native from "@pulumi/azure-native";

const dnsZone = new azure_native.network.Zone("exampleZone", {
  zoneName: "example.com",
  resourceGroupName: "prod-network-rg",
  zoneType: "Public",
});

const wwwRecord = new azure_native.network.RecordSet("wwwRecord", {
  zoneName: dnsZone.name,
  resourceGroupName: "prod-network-rg",
  relativeRecordSetName: "www",
  recordType: "A",
  ttl: 3600,
  aRecords: [{ ipv4Address: "203.0.113.10" }],
});
```

- **Real programming languages**: Write DNS logic in TypeScript, Python, Go. Use loops, conditionals, data structures.
- **Azure Native provider**: Auto-generated from Azure API specs. Always up-to-date with latest features (like DNSSEC).
- **Stack outputs**: Export nameservers or other metadata for use in other stacks.

#### ARM Templates / Bicep

```bicep
resource dnsZone 'Microsoft.Network/dnsZones@2018-05-01' = {
  name: 'example.com'
  location: 'global'
}

resource wwwA 'Microsoft.Network/dnsZones/A@2018-05-01' = {
  name: 'www@${dnsZone.name}'
  properties: {
    TTL: 3600
    ARecords: [ { ipv4Address: '203.0.113.10' } ]
  }
}
```

- **Azure-native**: First-class support in Azure DevOps, Blueprints, Policy integration.
- **Bicep simplifies ARM**: Less verbose than raw JSON templates, compiles to ARM.

**What these solve**:
- **Version control**: DNS config lives in Git. Pull requests for DNS changes with peer review.
- **Automated pipelines**: CI/CD deploys DNS changes automatically (with approval gates for production).
- **Consistency**: Same tool manages compute, networking, and DNS. One deployment graph.
- **Audit trail**: Git history + pipeline logs = complete change tracking.

**Pitfalls to avoid**:
- **Mixing tools**: Never manage the same zone with Terraform in one pipeline and manual changes in another. Pick one source of truth.
- **TTL mismanagement**: If migrating a domain, lower TTLs days before the change so caches expire faster. After migration, raise TTLs back.
- **Record set confusion**: Azure groups records with the same name+type into a "record set". If you define two separate resources for `www` A records in your IaC, they'll conflict. Combine values into one resource.

**Verdict**: **This is the production standard.** Declarative IaC eliminates drift, enables GitOps workflows, and scales to hundreds of zones. The choice between Terraform/Pulumi/Bicep depends on team preference and existing tooling, but the declarative approach is non-negotiable for serious deployments.

### Level 4: High-Level Abstractions (Advanced Scenarios)

**Approach**: Use higher-level control planes like Crossplane (Kubernetes-native) or custom operators.

**Crossplane Example**: Define a Kubernetes Custom Resource for an Azure DNS zone. Crossplane reconciles that to Azure's actual state.

```yaml
apiVersion: dns.azure.crossplane.io/v1alpha1
kind: Zone
metadata:
  name: example-zone
spec:
  forProvider:
    resourceGroupName: prod-network-rg
    name: example.com
  providerConfigRef:
    name: azure-provider
```

**What it solves**:
- **GitOps for DNS**: Manage DNS via Kubernetes manifests in Git, with Flux or ArgoCD reconciling.
- **Multi-cloud abstraction**: Crossplane can manage DNS across Azure, AWS, GCP using similar Kubernetes interfaces.
- **Dynamic provisioning**: A Kubernetes operator could auto-create DNS zones per tenant or namespace.

**What it doesn't**:
- **Additional complexity**: Requires running Crossplane/operators in a Kubernetes cluster. More moving parts.
- **Feature lag**: Community providers may lag behind official Azure APIs for newest features (e.g., DNSSEC support).

**Verdict**: Excellent for platform teams building self-service infrastructure on Kubernetes, or multi-cloud environments where Kubernetes is the control plane. Overkill for simple single-cloud DNS management. Use when the benefits of Kubernetes-native workflows outweigh the operational overhead.

## Production DNS Essentials

### Domain Delegation

After creating a public DNS zone in Azure, you must delegate your domain to Azure's nameservers at your registrar:

1. **Retrieve nameservers**: Azure assigns four NS records (e.g., `ns1-01.azure-dns.com`, `ns2-01.azure-dns.net`, `ns3-01.azure-dns.org`, `ns4-01.azure-dns.info`).
2. **Update registrar**: Log into your domain registrar (GoDaddy, Namecheap, etc.) and replace the current nameservers with Azure's four addresses.
3. **Propagation**: DNS changes can take up to 48 hours to propagate globally (usually much faster).
4. **Validation**: Use `dig NS example.com` to confirm delegation is complete.

**Pro tip**: Before switching, lower TTLs at your old DNS provider a few days ahead. This minimizes stale cache issues during migration.

### DNS Record Management Patterns

**Declarative (Recommended)**  
Define DNS state in IaC configuration. Use pull requests for changes. CI/CD applies updates. Version control provides audit trail and rollback capability.

**Zone File Workflow**  
Maintain a master BIND zone file in Git. Use CI pipeline to run `az network dns zone import` on changes. Hybrid approach between traditional DNS and IaC.

**Dynamic Updates**  
For Kubernetes workloads, use **ExternalDNS** to automatically create/delete DNS records based on Ingress or Service annotations. No manual DNS changes needed when services scale or move.

### DNSSEC Considerations

Azure DNS supports DNSSEC signing (GA 2025). Enabling DNSSEC:

1. **Enable signing**: `az network dns zone update --name example.com --signing-enabled true`
2. **Azure manages keys**: Azure uses Azure Key Vault to store KSK/ZSK pairs. Automatic key rotation handled by Azure.
3. **Update registrar**: Retrieve the DS (Delegation Signer) record from Azure and add it at your domain registrar to complete the chain of trust.

**When to use DNSSEC**:
- Compliance requirements (some government/financial sectors mandate it)
- High-security domains (prevents DNS spoofing)
- Customer trust signals

**When to skip**:
- Internal private zones (DNSSEC not applicable)
- Development environments (adds complexity without benefit)
- If registrar doesn't support DS records (rare but possible)

### Monitoring and Backups

**Monitoring**:
- **Azure Monitor metrics**: Track query count, alert on unexpected drops (could indicate delegation issues).
- **Activity logs**: Capture who changed DNS records and when via Azure Resource Manager logs.
- **Private DNS analytics**: For private zones, enable DNS query logging to Log Analytics (useful for troubleshooting internal resolution).

**Backups**:
- **Zone file export**: Regularly export zones to BIND format: `az network dns zone export -g rg -n example.com -f backup.zone`
- **IaC as backup**: If DNS is defined in Terraform/Pulumi, your Git repository **is** your backup. Accidental deletion? Redeploy from code.
- **No native versioning**: Azure DNS doesn't version record changes. Use Git + IaC for change history.

### Common Pitfalls

**Mixing management tools**: Managing the same zone with Terraform and manual portal changes creates drift. Terraform may overwrite manual entries on next run. Solution: Pick one source of truth (IaC), enforce it via access controls.

**Forgetting TTLs during migration**: If you change an A record's IP but the old TTL was 24 hours, clients cache the old IP for a day. Solution: Lower TTL before critical changes, raise after.

**SOA misconfiguration on import**: Azure auto-manages SOA records. If you repeatedly import zone files with stale serial numbers, external secondary DNS (if any) might not update. Solution: Use Azure's default SOA or ensure serial increments on each import.

**Private zone not linked**: Creating a private DNS zone without linking it to a VNet means no resources can query it. Solution: Always link to at least one VNet, verify resolution from a VM.

## 80/20 Configuration: What Most Users Actually Need

DNS has dozens of record types, but production deployments overwhelmingly use a small subset. Here's what matters:

### Essential DNS Record Types

| Record Type | Usage | Frequency |
|-------------|-------|-----------|
| **A** | Map hostname to IPv4 address | ~80% of all records |
| **AAAA** | Map hostname to IPv6 address | Growing (recommend alongside A) |
| **CNAME** | Alias one name to another | Very common for subdomains |
| **TXT** | SPF, DKIM, domain verification | Nearly every domain has 1-3 |
| **MX** | Mail server routing | If domain handles email |
| **NS** | Subdomain delegation | Rare (only for delegating subzones) |
| **CAA** | Certificate authority authorization | Security best practice (5-10%) |
| **SRV** | Service locator (LDAP, SIP, etc.) | Niche (Active Directory, VoIP) |

**Rarely used**: PTR (reverse DNS, managed separately in Azure), SOA (Azure auto-manages), exotic types like APL or RP (not supported).

### Common Configuration Patterns

#### Basic Public Zone (Web Hosting)

```yaml
zone_name: "example.com"
resource_group: "prod-network-rg"
records:
  - name: "@"              # Apex record
    type: A
    ttl: 3600
    values: ["203.0.113.10"]
  
  - name: "www"
    type: CNAME
    ttl: 3600
    values: ["example.com."]
```

**Explanation**: Root domain points to web server IP. `www` is a CNAME to the root. Both HTTP and HTTPS traffic work. TTL of 1 hour balances caching and flexibility.

#### Production Zone (Web + Email)

```yaml
zone_name: "contoso.com"
resource_group: "prod-network-rg"
records:
  # Web hosting
  - name: "@"
    type: A
    ttl: 300
    values: ["198.51.100.45"]
  
  - name: "www"
    type: CNAME
    ttl: 300
    values: ["contoso.com."]
  
  # Email routing
  - name: "@"
    type: MX
    ttl: 3600
    values:
      - "10 mail.contoso.com."
      - "20 backup-mail.contoso.com."
  
  - name: "mail"
    type: A
    ttl: 300
    values: ["198.51.100.100"]
  
  - name: "backup-mail"
    type: A
    ttl: 300
    values: ["198.51.100.101"]
  
  # Email authentication
  - name: "@"
    type: TXT
    ttl: 300
    values: ["v=spf1 include:mail.contoso.com -all"]
  
  - name: "_dmarc"
    type: TXT
    ttl: 300
    values: ["v=DMARC1; p=reject; rua=mailto:dmarc@contoso.com"]
  
  # Certificate authority authorization
  - name: "@"
    type: CAA
    ttl: 86400
    values: ["0 issue \"letsencrypt.org\""]
```

**Explanation**: Comprehensive setup for a domain hosting a website and handling email. MX records point to mail servers (with failover). SPF and DMARC TXT records improve email deliverability and security. CAA record restricts certificate issuance to Let's Encrypt only (prevents rogue CAs).

#### Private Zone (Internal Services)

```yaml
zone_name: "internal.corp.local"
zone_type: "Private"
resource_group: "prod-network-rg"
vnet_links:
  - vnet_id: "/subscriptions/.../virtualNetworks/vnet-prod-west"
    auto_registration: true
  - vnet_id: "/subscriptions/.../virtualNetworks/vnet-prod-east"
    auto_registration: false

records:
  - name: "app1"
    type: A
    ttl: 10
    values: ["10.10.1.5", "10.20.1.5"]  # Multi-region instances
  
  - name: "db"
    type: CNAME
    ttl: 60
    values: ["sqlserver001.database.windows.net."]
```

**Explanation**: Private zone for internal name resolution. Linked to two VNets (one with auto-registration so VMs get DNS entries automatically). `app1` has two IPs for simple round-robin load distribution. `db` is a CNAME to an Azure SQL endpoint (could be private endpoint via `privatelink` zone in practice). Very low TTLs for fast failover.

### Minimal API Design

Based on the 80/20 analysis, Planton's `AzureDnsZoneSpec` focuses on essentials:

- `zone_name`: The DNS domain (e.g., `example.com`)
- `resource_group`: Azure resource group for the zone
- `records`: List of DNS records (type, name, values, TTL)

**What's omitted** (intentionally):
- SOA customization (Azure auto-manages)
- DNSSEC toggle (can be enabled post-creation via Azure directly)
- Private zone VNet links (separate concern, handled by Azure networking)
- Azure-specific alias records (specialty feature, not core DNS)

This minimal schema covers 95% of use cases while keeping the API simple and cloud-agnostic. Advanced features can be layered via Azure-specific extensions or manual configuration.

## Integration Scenarios

### Migrating from Other DNS Providers

**From AWS Route 53 / GCP Cloud DNS**:
1. Export zone from source provider (use AWS CLI `list-resource-record-sets` or GCP `gcloud dns record-sets export`)
2. Convert to BIND zone file format
3. Import to Azure: `az network dns zone import -g rg -n example.com -f zone.txt`
4. Validate records in Azure Portal or CLI
5. Update registrar NS records to Azure's nameservers
6. Monitor for 48 hours, then decommission old zone

**From Cloudflare**:
- If using Cloudflare's CDN/proxy (orange cloud), migrating DNS means losing that proxy. Replace with Azure Front Door or Azure CDN if needed.
- Export DNS records from Cloudflare (via API or UI)
- Import to Azure as above
- Update NS delegation at registrar

**Domain transfer vs. NS delegation**: You almost never need to transfer domain registration to move DNS hosting. Just update nameservers at your existing registrar. Keeps things simple and avoids registration lock periods.

### Azure Service Integration

**Azure App Services**: Map custom domains to web apps by creating:
- A/CNAME record pointing to `{app}.azurewebsites.net`
- TXT record for domain verification (Azure provides the verification token)
- Consider using Azure DNS alias records to reference App Gateway or Traffic Manager directly

**Azure Kubernetes Service (AKS)**: Use **ExternalDNS** controller to auto-manage DNS:
- Watches Kubernetes Ingress and Service resources
- Creates/deletes Azure DNS records based on annotations
- No manual DNS changes needed when services scale or move
- Example: Deploy Ingress for `api.dev.example.com`, ExternalDNS creates the A record automatically

**Azure Private Endpoints**: When enabling Private Endpoints for PaaS services (Storage, SQL, etc.):
- Create private DNS zone (e.g., `privatelink.blob.core.windows.net`)
- Link zone to VNets that need to resolve the private endpoint
- Azure can auto-create DNS records when you provision the endpoint (or manage via IaC)

## Planton's Approach

Planton's `AzureDnsZone` API provides a **cloud-agnostic, protobuf-defined interface** for managing Azure DNS zones as code. The implementation uses **Pulumi with Azure Native provider** for several strategic reasons:

**Why Pulumi?**
1. **Real programming languages**: Write DNS logic in Go, TypeScript, Python. Use loops, conditionals, data structures – not limited to DSL constraints.
2. **Azure Native provider**: Auto-generated from Azure ARM APIs. Always current with latest features (DNSSEC, new record types).
3. **Stack outputs**: Export nameservers and zone metadata for consumption by other infrastructure components.
4. **Type safety**: Compile-time validation of configurations. Catch errors before deployment.

**Minimal API philosophy**:
The protobuf spec exposes only essential fields (zone name, resource group, records). This achieves:
- **Simplicity**: 80% of users need 20% of DNS features. Focus on the common path.
- **Portability**: A minimal, standard DNS schema can map to other cloud providers (Route 53, Cloud DNS) with similar protobuf definitions.
- **Extensibility**: Advanced users can layer Azure-specific features (alias records, DNSSEC) via Pulumi code or post-deployment configuration.

**Deployment workflow**:
1. Define `AzureDnsZone` resource in protobuf/YAML
2. Planton CLI generates Pulumi stack
3. Pulumi creates Azure DNS zone and records
4. Stack outputs return nameserver addresses
5. Update domain registrar with those nameservers (manual step, one-time)

**For advanced scenarios** (private zones, auto-registration, ExternalDNS integration), see:
- [Pulumi Implementation Guide](../iac/pulumi/README.md)
- [Terraform Alternative](../iac/tf/README.md)

## Conclusion

Azure DNS has evolved from a basic managed DNS offering to a strategic cloud-native component. For Azure-centric workloads, it provides seamless integration, powerful features like alias records and private zones, and now production-grade DNSSEC support.

The path to production-ready DNS management is clear:
- **Avoid** manual portal management (except for exploration)
- **Use** declarative infrastructure-as-code (Terraform, Pulumi, Bicep)
- **Embrace** GitOps workflows (pull requests for DNS changes, automated pipelines)
- **Focus** on the 20% of DNS features that solve 80% of problems (A, AAAA, CNAME, TXT, MX)
- **Monitor** and back up (zone exports, IaC as source of truth)

Planton's `AzureDnsZone` API distills these lessons into a minimal, multi-cloud interface backed by production-proven Pulumi automation. Whether you're hosting a single domain or orchestrating DNS across hundreds of zones, the principles remain: treat DNS as code, automate relentlessly, and keep configurations simple.

---

**Next Steps**:
- Review the [Pulumi implementation details](../iac/pulumi/README.md)
- Explore [example configurations](../iac/pulumi/examples.md)
- Understand [stack outputs and integration patterns](../iac/pulumi/overview.md)


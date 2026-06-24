# Deploying Cloudflare DNS Zones: A Production Guide

## Introduction

DNS is infrastructure you notice only when it breaks—and when it does, everything breaks. Slow DNS resolution makes your site feel sluggish. DNS outages make it disappear entirely. DNS attacks can overwhelm your infrastructure before a single HTTP request reaches your servers. For years, the conventional wisdom was to use your cloud provider's DNS service: Route 53 on AWS, Cloud DNS on GCP, Azure DNS on Azure. The reasoning was simple—keep everything in one ecosystem, leverage tight integrations, and avoid adding another vendor.

But that conventional wisdom has blind spots. Cloud provider DNS services typically charge per zone and per million queries, which adds up quickly for high-traffic sites. They're often regional services with anycast bolted on, not built anycast-first. And critically, they don't come with integrated DDoS protection, CDN capabilities, or a global edge network that can absorb attacks and accelerate content delivery—those are separate products you pay for and configure independently.

**Cloudflare DNS** challenges this model by treating DNS not as a standalone utility but as the foundation of a comprehensive edge platform. When you create a DNS zone on Cloudflare, you're not just getting nameservers—you're getting authoritative DNS served from 330+ global locations via native anycast, with built-in DDoS protection, zero per-query charges (even on the free plan), optional integrated CDN/WAF/proxy capabilities, and sub-second global propagation. You can use Cloudflare purely for DNS (grey-cloud all your records and treat it like any other DNS provider) or leverage its proxy features (orange-cloud your web traffic to hide origin IPs, cache content, and filter attacks at the edge).

This flexibility makes Cloudflare DNS uniquely positioned for modern multi-cloud and edge-native architectures. Whether you're running a personal blog on the free tier, a SaaS platform serving millions of users across continents, or a hybrid infrastructure spanning AWS, GCP, and on-premises datacenters, Cloudflare DNS can serve as your unified global control plane for traffic routing.

This guide explains the landscape of deployment methods for Cloudflare DNS zones—from manual console operations to production-grade Infrastructure-as-Code—and shows how OpenMCF distills Cloudflare's extensive feature set into a clean, protobuf-defined API that exposes the 20% of configuration that 80% of users actually need.

---

## The Deployment Spectrum: From Manual to Production

Managing DNS zones on Cloudflare can be approached in many ways, from point-and-click simplicity to declarative automation. Here's the maturity progression:

### Level 0: Manual Dashboard Management (Anti-Pattern for Scale)

**What it is:** Using Cloudflare's web dashboard to manually add zones, import DNS records (via their scan feature or BIND file upload), and manage settings through the UI.

**What it solves:** Nothing that can't be solved better with automation. The dashboard is excellent for learning Cloudflare's interface, exploring features, and one-off experiments. You can see the "orange vs grey cloud" toggle for each record, preview SSL settings, and get instant feedback. Cloudflare's "Add a Site" wizard even scans your existing DNS and offers to import records automatically, which is convenient for small migrations.

**What it doesn't solve:** Repeatability, version control, auditability, multi-environment consistency. If you manage more than a handful of zones or need to reproduce configurations across dev/staging/prod, manual clicks don't scale. You'll introduce typos, forget records, or accidentally click "orange cloud" when you meant "grey cloud" (exposing your origin IP or breaking non-HTTP services). When someone asks "why is this CNAME proxied?" six months from now, there's no git history to check—just your fading memory.

**Specific pitfalls:**
- The DNS scan during zone creation often misses less common record types (SRV, CAA, obscure TXT records)
- Adding a subdomain that's already a separate zone in Cloudflare will fail with cryptic errors unless you're on Enterprise with subdomain delegation
- Manually managing dozens of DNS records is error-prone; you'll eventually create duplicates, set wrong TTLs, or misconfigure MX priorities

**Verdict:** Use the dashboard to understand Cloudflare's model and explore features. Never rely on it as your production deployment method. Think of it like using AWS Console to launch EC2 instances—fine for learning, unacceptable for production infrastructure.

---

### Level 1: CLI and API Scripting (Automation Without State)

**What it is:** Using Cloudflare's REST API directly via `curl`, community CLI tools like `flarectl`, or scripting with official SDKs (Go's `cloudflare-go`, Python's `cloudflare` library, Node.js wrappers).

**Example with `curl`:**
```bash
# Create a zone
curl -X POST "https://api.cloudflare.com/client/v4/zones" \
  -H "Authorization: Bearer $CF_API_TOKEN" \
  -H "Content-Type: application/json" \
  --data '{"name":"example.com","account":{"id":"'$ACCOUNT_ID'"}}'

# Add an A record
curl -X POST "https://api.cloudflare.com/client/v4/zones/$ZONE_ID/dns_records" \
  -H "Authorization: Bearer $CF_API_TOKEN" \
  -H "Content-Type: application/json" \
  --data '{"type":"A","name":"www","content":"198.51.100.4","proxied":true}'
```

**Example with Python SDK:**
```python
import CloudFlare

cf = CloudFlare.CloudFlare(token=api_token)
zone = cf.zones.post(data={'name': 'example.com', 'account': {'id': account_id}})
zone_id = zone['id']

cf.zones.dns_records.post(zone_id, data={
    'type': 'A',
    'name': 'www',
    'content': '198.51.100.4',
    'proxied': True
})
```

**What it solves:** Scripted automation. You can integrate DNS creation into CI/CD pipelines, trigger zone updates from application deployments, or build custom tooling around Cloudflare's API. The API is comprehensive (create zones, manage records, configure zone settings, enable DNSSEC, set up page rules, etc.) and well-documented. Using API tokens with granular scopes (like "Zone:Edit" for a specific zone) is more secure than global API keys.

**What it doesn't solve:** State management and idempotency. Scripts are imperative—they describe *how* to achieve a result, not *what* the result should be. If you run the same API script twice, you'll likely get errors ("record already exists") or create duplicates. There's no built-in tracking of what infrastructure exists. Rolling back a change means writing another script to undo it. Coordinating dependencies (create zone before adding records, set nameservers before enabling DNSSEC) is manual sequencing in your script.

**Rate limits:** Cloudflare's API allows 1,200 calls per 5 minutes. For bulk operations (like adding hundreds of DNS records), you'll need to batch requests or add delays to avoid hitting this limit.

**Authentication best practice:** Use API tokens with least-privilege scopes, not the legacy Global API Key. Create a token that has `Zone:Read` and `DNS:Edit` for only the zones you need to manage, and inject it via environment variables (`CF_API_TOKEN`) rather than hardcoding.

**Verdict:** Suitable for integration points where you need programmatic control—triggering DNS updates from deployment scripts, custom automation, or when building a higher-level abstraction. Not suitable as your primary infrastructure management approach because it lacks state tracking and declarative guarantees.

---

### Level 2: Configuration Management Tools (Ansible, Puppet, Chef)

**What it is:** Using config management platforms like Ansible (which has `cloudflare_dns` modules in the `community.general` collection) to declaratively ensure zones and records exist.

**Ansible example:**
```yaml
- name: Ensure Cloudflare zone exists
  community.general.cloudflare_dns:
    zone: example.com
    record: www
    type: A
    value: 198.51.100.4
    proxied: yes
    api_token: "{{ cloudflare_api_token }}"
```

**What it solves:** Declarative intent and convergence. Ansible will check if the record exists and matches the desired state; if not, it creates or updates it. If the state is already correct, Ansible does nothing (idempotency). You can version-control your playbooks, apply them repeatedly, and integrate DNS management into broader server provisioning workflows (deploy app, update DNS record to point to new IP).

**What it doesn't solve:** Multi-resource orchestration and complex dependencies. Ansible is primarily pull/push configuration to servers, and while it can manage external APIs like Cloudflare, it's not designed for full infrastructure orchestration. For example, managing outputs from one resource as inputs to another (like using a zone ID to create records) can be cumbersome. You also end up with Ansible-specific patterns and a separate tool from what you might use for cloud infrastructure (Terraform/Pulumi).

**When to use:** If you're already heavily invested in Ansible for server configuration and want to add DNS management to existing playbooks (e.g., provision a VM and update its DNS entry in one workflow). Less common as a standalone DNS management solution.

**Verdict:** A middle ground that works well for teams already using config management who need occasional DNS automation as part of broader playbook runs. Not the first choice for dedicated DNS infrastructure-as-code.

---

### Level 3: Infrastructure-as-Code with Terraform (Production Standard)

**What it is:** Using HashiCorp Terraform with the official Cloudflare provider to define zones, records, and settings as declarative code.

**Terraform example:**
```hcl
terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.0"
    }
  }
}

provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

resource "cloudflare_zone" "example" {
  name = "example.com"
  account = {
    id = var.cloudflare_account_id
  }
  paused = false
}

resource "cloudflare_dns_record" "www" {
  zone_id = cloudflare_zone.example.id
  name    = "www"
  type    = "A"
  content = "198.51.100.4"
  ttl     = 1
  proxied = true
}

# In provider v5 each zone setting is its own resource.
resource "cloudflare_zone_setting" "ssl" {
  zone_id    = cloudflare_zone.example.id
  setting_id = "ssl"
  value      = "full"
}
```

**What it solves:** Everything you need for production DNS management:

- **Declarative state:** You define *what* you want (a zone with these records), not *how* to create it. Terraform figures out the API calls.
- **State tracking:** Terraform maintains a state file that maps your configuration to real Cloudflare resources (zone IDs, record IDs). It knows what exists and detects drift.
- **Plan preview:** `terraform plan` shows exactly what will change before you apply. No surprises.
- **Dependency resolution:** Terraform automatically determines that records depend on zones, zones depend on account IDs, etc., and creates/destroys resources in the correct order.
- **Multi-environment support:** Use workspaces or separate state files to manage dev/staging/prod DNS configurations from the same codebase with different variables.
- **Import existing resources:** If you already have zones on Cloudflare, Terraform's `import` command (or the `cf-terraforming` tool) can bring them under management without recreating them.

**Cloudflare Terraform provider maturity:** The provider is battle-tested, covering 100+ resource types including zones, DNS records, zone settings, page rules, firewall rules, load balancers, Workers, and more. It's officially maintained with input from Cloudflare, ensuring new features (like new DNS record types or zone capabilities) are supported quickly. As of 2025, the v5.x provider is stable (early v5.0 releases had migration issues, but v5.5+ resolved them).

**State import and migration:** Cloudflare provides `cf-terraforming`, a CLI tool that reads your existing zones and generates Terraform configuration files for them. This dramatically simplifies migrating existing manual setups into Terraform management:

```bash
# Generate Terraform config for all zones in an account
cf-terraforming generate --resource-type cloudflare_zone --account $ACCOUNT_ID > zones.tf

# Generate DNS records for a specific zone
cf-terraforming generate --resource-type cloudflare_dns_record --zone $ZONE_ID > records.tf

# Import state
terraform import cloudflare_zone.example $ZONE_ID
```

**Secret management:** Never hardcode API tokens in `.tf` files. Use environment variables (`TF_VAR_cloudflare_api_token`), Terraform Cloud workspace variables, or secret management tools (Vault, AWS Secrets Manager) to inject credentials at runtime. The token should have scoped permissions (Zone:Edit, DNS:Edit for the target zones).

**Common patterns:**
- **One Terraform state per zone or per environment:** Isolate blast radius. Changes to dev DNS won't accidentally affect prod.
- **Modules for reusable zone configs:** Create a Terraform module that sets up a zone with standard settings (SSL mode, security level, default records like MX) and instantiate it for each domain.
- **Version control all `.tf` files:** Treat DNS configuration like application code. Code reviews catch mistakes (typos in domain names, wrong IPs) before they're applied.

**Verdict:** This is the production standard for managing Cloudflare DNS at scale. If you're serious about infrastructure-as-code and need reliability, auditability, and team collaboration, Terraform is the default choice. It's mature, widely adopted, and integrates seamlessly with CI/CD (Terraform Cloud, Atlantis, GitHub Actions workflows).

---

### Level 4: Infrastructure-as-Code with Pulumi (Code-Native Alternative)

**What it is:** Using Pulumi with real programming languages (TypeScript, Python, Go, etc.) to define Cloudflare infrastructure. Pulumi's Cloudflare provider offers the same resource coverage as Terraform but expressed in actual code.

**Pulumi TypeScript example:**
```typescript
import * as cloudflare from "@pulumi/cloudflare";

const zone = new cloudflare.Zone("example-zone", {
    zone: "example.com",
    accountId: process.env.CLOUDFLARE_ACCOUNT_ID,
    plan: "free",
    paused: false,
});

const wwwRecord = new cloudflare.Record("www", {
    zoneId: zone.id,
    name: "www",
    type: "A",
    value: "198.51.100.4",
    proxied: true,
});

const zoneSettings = new cloudflare.ZoneSettingsOverride("example-settings", {
    zoneId: zone.id,
    settings: {
        ssl: "full",
        alwaysUseHttps: "on",
        automaticHttpsRewrites: "on",
    },
});
```

**Pulumi Python example:**
```python
import pulumi
import pulumi_cloudflare as cloudflare

zone = cloudflare.Zone("example-zone",
    zone="example.com",
    account_id=config.require("cloudflare_account_id"),
    plan="free",
    paused=False)

www_record = cloudflare.Record("www",
    zone_id=zone.id,
    name="www",
    type="A",
    value="198.51.100.4",
    proxied=True)
```

**What it solves:** Everything Terraform solves, but with the expressiveness of a full programming language:

- **Loops and conditionals:** Generate hundreds of DNS records programmatically with `for` loops, apply conditional logic (`if env === 'prod'`) without wrestling with HCL's limited constructs.
- **Type safety:** TypeScript/Go provide compile-time checks for configuration (e.g., catching misspelled property names before deployment).
- **Native testing:** Write unit tests for infrastructure using familiar testing frameworks (Jest for TS, pytest for Python).
- **Reusable functions and packages:** Share infrastructure patterns as npm/PyPI packages, not just Terraform modules.
- **Integration with application code:** If your app is written in TypeScript or Python, you can use the same language, libraries, and patterns for both app and infrastructure.

**Pulumi vs Terraform for Cloudflare:**

| Aspect | Terraform | Pulumi |
|--------|-----------|--------|
| **Configuration Language** | HCL (declarative DSL) | TypeScript, Python, Go, C#, Java |
| **Resource Coverage** | 100+ Cloudflare resources | Equivalent (Pulumi provider initially bridged from Terraform) |
| **State Management** | Local or remote backends (S3, Terraform Cloud) | Pulumi Service or self-managed (S3, Azure Blob) |
| **Maturity for Cloudflare** | Very mature, widely adopted | Production-ready, smaller community |
| **Learning Curve** | Lower for ops teams familiar with declarative config | Lower for developers already coding in supported languages |
| **Best For** | Teams wanting stability, ecosystem maturity, standard HCL patterns | Teams preferring code expressiveness, complex logic, or language familiarity |

**Verdict:** Pulumi is an excellent choice if your team prefers writing infrastructure in a general-purpose language (especially TypeScript or Python) or if you need complex orchestration logic (dynamic resource generation based on external data, conditional configurations, etc.). For straightforward DNS management, Terraform's simplicity might be advantageous. For programmatic DNS workflows (e.g., SaaS platform that creates a zone per customer tenant), Pulumi's code-native approach shines.

Both are production-grade. The choice is more about team preference and existing tooling than technical capability.

---

### Level 5: Kubernetes-Native Approaches (Cloud-Native Integration)

**What it is:** Managing Cloudflare DNS zones and records as Kubernetes custom resources using tools like ExternalDNS or Crossplane.

#### ExternalDNS (Dynamic DNS from Kubernetes Services/Ingresses)

**Purpose:** ExternalDNS watches Kubernetes Service and Ingress resources and automatically creates corresponding DNS records in Cloudflare. Perfect for dynamic environments where IPs change frequently (autoscaling, blue-green deployments, ephemeral test clusters).

**How it works:**
1. You pre-create a DNS zone in Cloudflare (e.g., `k8s.example.com`) and delegate it to Cloudflare's nameservers.
2. Deploy ExternalDNS as a controller in your cluster with Cloudflare provider credentials (API token with Zone:Read and DNS:Edit).
3. When you create a Kubernetes Ingress with `host: myapp.k8s.example.com`, ExternalDNS automatically creates an A or CNAME record in Cloudflare pointing to the Ingress LoadBalancer IP.
4. When the Ingress is deleted, ExternalDNS removes the DNS record.

**Configuration example:**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: external-dns
data:
  cloudflare.conf: |
    provider: cloudflare
    cloudflare-proxied: false  # Set to true to enable orange-cloud by default
    txt-owner-id: k8s-cluster-prod
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: external-dns
spec:
  template:
    spec:
      containers:
      - name: external-dns
        image: registry.k8s.io/external-dns/external-dns:latest
        args:
        - --source=ingress
        - --source=service
        - --provider=cloudflare
        - --cloudflare-proxied  # Enable proxied records
        env:
        - name: CF_API_TOKEN
          valueFrom:
            secretKeyRef:
              name: cloudflare-api-token
              key: token
```

**Benefits:**
- **No manual DNS updates:** When you deploy or scale services, DNS follows automatically.
- **GitOps-friendly:** Your Kubernetes manifests define both the service and its DNS presence.
- **Multi-cluster support:** Use different `txt-owner-id` values to allow multiple clusters to manage different subdomains in the same zone.

**Limitations:**
- ExternalDNS manages *records*, not zones. You still create the zone via Terraform/Pulumi/API.
- It's Kubernetes-specific. For non-K8s workloads, you need a different approach.

**Verdict:** Essential for cloud-native Kubernetes environments where DNS needs to stay in sync with rapidly changing services. Combine ExternalDNS (for records) with Terraform/Pulumi (for zone creation and configuration) for a complete solution.

---

#### Crossplane (Kubernetes as a Control Plane for Cloudflare)

**Purpose:** Crossplane extends Kubernetes with custom resource definitions (CRDs) for cloud infrastructure, including Cloudflare zones and records. You define Cloudflare resources as Kubernetes manifests, and Crossplane reconciles them via the Cloudflare API.

**Example Zone CRD:**
```yaml
apiVersion: cloudflare.crossplane.io/v1alpha1
kind: Zone
metadata:
  name: example-com
spec:
  forProvider:
    domain: "example.com"
    accountId: "<CLOUDFLARE_ACCOUNT_ID>"
    plan: "free"
  providerConfigRef:
    name: cloudflare-provider
```

**Example Record CRD:**
```yaml
apiVersion: cloudflare.crossplane.io/v1alpha1
kind: Record
metadata:
  name: www-example
spec:
  forProvider:
    zoneId: "<ZONE_ID>"
    name: "www"
    type: "A"
    content: "198.51.100.4"
    proxied: true
  providerConfigRef:
    name: cloudflare-provider
```

**Benefits:**
- **Kubernetes-native:** If your team lives in Kubernetes and GitOps workflows (ArgoCD, Flux), managing infrastructure as CRDs feels natural.
- **Unified control plane:** Manage Cloudflare, AWS, GCP, and on-prem resources with the same Kubernetes-based abstractions.
- **Declarative reconciliation:** Crossplane continuously ensures the actual state matches the desired state defined in CRDs.

**Limitations:**
- The Crossplane Cloudflare provider is community-maintained (not core), so check its maturity and feature coverage.
- Crossplane is complex to set up initially (requires understanding CRDs, provider configs, compositions).
- If your infrastructure isn't Kubernetes-centric, Crossplane might be overkill.

**Verdict:** Great for platform teams building Kubernetes-based internal developer platforms where everything (apps, databases, DNS) is provisioned via kubectl. For simpler use cases, Terraform/Pulumi are more straightforward.

---

## IaC Tool Comparison: Terraform vs Pulumi

Both Terraform and Pulumi are production-ready for managing Cloudflare DNS. Here's how they compare on key dimensions:

### Provider Maturity and Coverage

**Terraform:** The Cloudflare provider is very mature (v5.x as of 2025), covering 100+ resource types: zones, DNS records, zone settings, page rules, WAF rules, load balancers, Workers, R2 storage, and more. It's officially maintained with Cloudflare's involvement. The provider is widely used in production at scale.

**Pulumi:** The Cloudflare provider offers equivalent coverage, initially built via a Terraform bridge but now also offering native resources. Resources include `cloudflare.Zone`, `cloudflare.Record`, `cloudflare.ZoneSettings`, `cloudflare.ZoneDnssec`, etc. Pulumi's provider is also production-ready, though it has a smaller community than Terraform.

**Verdict:** Feature parity. Both support everything you need for DNS zones and beyond.

---

### Configuration Language and Expressiveness

**Terraform (HCL):**
```hcl
# Simple and readable for standard use cases
resource "cloudflare_zone" "example" {
  name = "example.com"
  account = {
    id = var.account_id
  }
}

# Looping requires for_each or count
resource "cloudflare_dns_record" "subdomains" {
  for_each = toset(["app1", "app2", "app3"])

  zone_id = cloudflare_zone.example.id
  name    = each.value
  type    = "CNAME"
  content = "lb.example.com"
  ttl     = 1
  proxied = true
}
```

**Pros:** Declarative, easy to read, less verbose for simple configs.  
**Cons:** Limited expressiveness. Complex logic (nested loops, conditionals based on API responses) can be awkward.

**Pulumi (TypeScript):**
```typescript
// Full programming language expressiveness
const zone = new cloudflare.Zone("example", {
    zone: "example.com",
    accountId: accountId,
    plan: "free",
});

// Natural loops, type safety, reusable functions
const subdomains = ["app1", "app2", "app3"];
subdomains.forEach(sub => {
    new cloudflare.Record(`${sub}-record`, {
        zoneId: zone.id,
        name: sub,
        type: "CNAME",
        value: "lb.example.com",
        proxied: true,
    });
});
```

**Pros:** Full language power (loops, conditionals, functions, type safety). Easy to programmatically generate configs (e.g., from external data sources).  
**Cons:** Requires a runtime (Node.js/Python/Go). Slightly more verbose for very simple configs.

**Verdict:** Terraform is simpler for straightforward DNS management. Pulumi is more powerful for complex, dynamic configurations (e.g., SaaS platform creating zones per customer).

---

### State Management and Secrets

Both tools maintain state to track infrastructure:

**Terraform:** State file (local or remote in S3, Terraform Cloud, etc.) stores resource IDs and metadata. Secrets (API tokens) should be injected via environment variables or secret backends (Vault), never committed to git.

**Pulumi:** State stored in Pulumi Service (SaaS) or self-managed backends (S3, Azure Blob). Pulumi can encrypt secrets in state files (using `pulumi config set --secret`). Similar best practice: inject API tokens via environment or secret providers.

**Verdict:** Both require careful secrets management. Pulumi's built-in secret encryption is convenient; Terraform relies on external secret management.

---

### Multi-Environment and Team Workflows

**Terraform:**
- Use workspaces or separate state files for dev/staging/prod.
- Terraform Cloud offers team workflows, state locking, policy-as-code (Sentinel), and UI for plan approvals.
- Atlantis provides PR-based workflows for self-hosted setups.

**Pulumi:**
- Use stacks (built-in concept) for different environments (e.g., `dev`, `staging`, `prod` stacks).
- Pulumi Service provides similar team features: RBAC, audit logs, policy as code (Pulumi CrossGuard).
- Stack outputs can be consumed by other stacks, enabling multi-tier infrastructure dependencies.

**Verdict:** Both support robust team workflows. Terraform Cloud and Pulumi Service are comparable SaaS offerings.

---

### Importing Existing Infrastructure

**Terraform:** 
- `terraform import cloudflare_zone.example $ZONE_ID` brings existing zones under management.
- Cloudflare's `cf-terraforming` tool auto-generates Terraform configs from existing zones, making bulk imports painless.

**Pulumi:**
- `pulumi import cloudflare:index/zone:Zone example $ZONE_ID` imports zones.
- No equivalent to `cf-terraforming`, so you'd manually generate Pulumi code or use Terraform's tool and translate.

**Verdict:** Terraform has better tooling for importing existing Cloudflare infrastructure at scale.

---

### Community and Ecosystem

**Terraform:** Massive community. Hundreds of public modules (e.g., Cloud Posse's Cloudflare zone module with sensible defaults). Extensive documentation, tutorials, and examples.

**Pulumi:** Growing community. Fewer examples, but Pulumi's own documentation is excellent. Active development and responsive support.

**Verdict:** Terraform wins on ecosystem size. Pulumi wins on language-native tooling (npm packages, type definitions).

---

### Summary Recommendation

| Scenario | Recommendation |
|----------|---------------|
| Standard DNS management for production | **Terraform** (mature, widely adopted, simple HCL) |
| Complex programmatic DNS (SaaS multi-tenancy, dynamic generation) | **Pulumi** (code expressiveness, loops, type safety) |
| Team already using TypeScript/Python/Go | **Pulumi** (language familiarity reduces friction) |
| Team already using Terraform for cloud infra | **Terraform** (consistency across all IaC) |
| Importing large existing Cloudflare setups | **Terraform** (`cf-terraforming` tool simplifies this) |

---

## Production Essentials: Security, Performance, and Operations

Deploying Cloudflare DNS in production requires attention to several critical dimensions beyond just creating zones and records:

### Nameserver Delegation and Registrar Configuration

After creating a zone in Cloudflare (full setup mode), you must **update your domain's nameservers at your registrar** to Cloudflare's assigned nameservers. Cloudflare assigns two nameservers per zone (e.g., `gina.ns.cloudflare.com`, `ivan.ns.cloudflare.com`). This delegation is what makes Cloudflare authoritative for your domain.

**Steps:**
1. Create zone in Cloudflare (via dashboard, API, Terraform, or Pulumi).
2. Note the assigned nameservers (displayed in dashboard or returned by API).
3. Go to your domain registrar (GoDaddy, Namecheap, AWS Route 53 Registrar, etc.) and replace the existing NS records with Cloudflare's nameservers.
4. Wait for DNS propagation (usually minutes to a few hours).
5. Cloudflare will detect the change and mark the zone as "Active" (you'll see this status in the dashboard).

**Critical:** Remove all old nameservers from your registrar. Leaving a mix of old and new NS can cause inconsistent DNS responses ("split brain" DNS).

**Using Cloudflare Registrar:** Cloudflare offers domain registration for certain TLDs at wholesale cost. If you transfer your domain to Cloudflare Registrar, nameserver setup is automatic (the domain is locked to Cloudflare's NS), and DNSSEC can be auto-enabled without manual DS record configuration. This simplifies management but locks you into Cloudflare's DNS.

---

### DNSSEC: Prevent DNS Spoofing

DNSSEC (Domain Name System Security Extensions) cryptographically signs your DNS records to prevent cache poisoning and man-in-the-middle attacks. Cloudflare supports DNSSEC on **all plan tiers (Free, Pro, Business, Enterprise)**.

**Enabling DNSSEC:**
1. In Cloudflare's dashboard (or via API), enable DNSSEC for your zone.
2. Cloudflare generates a DS record (a hash of your zone's public key).
3. Copy the DS record to your domain registrar (most registrars have a "DNSSEC" section where you paste this value).
4. Once published at the registry, DNSSEC is active. Validating resolvers will reject forged DNS responses for your domain.

**Production recommendation:** Always enable DNSSEC for production domains. It's one-click in Cloudflare and adds a critical security layer. Monitor DNSSEC status in Cloudflare's dashboard to ensure it remains "Active."

**Advanced:** Enterprise customers can use multi-signer DNSSEC (allowing multiple DNS providers to sign the same zone) for complex multi-provider setups.

---

### Orange Cloud vs Grey Cloud: Proxied vs DNS-Only

Cloudflare's unique feature is the **proxy toggle** for each DNS record:

- **Proxied (Orange Cloud):** DNS responses return Cloudflare's anycast IPs. All traffic flows through Cloudflare's edge network, enabling:
  - CDN caching (static assets served from 330+ global locations)
  - WAF and DDoS protection (attacks blocked at the edge)
  - SSL/TLS termination (Cloudflare handles HTTPS, hiding your origin)
  - Origin IP obfuscation (your server's real IP never exposed)
  - HTTP optimizations (HTTP/2, HTTP/3, Brotli compression, image optimization)

- **DNS-Only (Grey Cloud):** DNS responses return your origin IP. Clients connect directly to your server. Cloudflare acts purely as DNS provider (fast resolution, no proxy/CDN/WAF).

**Production strategy:**
- **Orange-cloud web services:** `www`, `app`, `api`, `blog` (anything serving HTTP/HTTPS traffic)
- **Grey-cloud infrastructure:** MX records (email), SSH/FTP servers, VPN endpoints, database hostnames, or services on non-HTTP ports that Cloudflare doesn't proxy

**Common mistake:** Proxying non-HTTP services will break them (e.g., proxying your mail server's A record prevents SMTP connections). Always grey-cloud MX records and the A/AAAA records they point to.

**Default behavior:** Set `default_proxied: true` in your zone configuration to default new records to orange-cloud (safer default, ensures you don't accidentally expose origins).

---

### SSL/TLS Configuration

If you enable proxying (orange cloud), Cloudflare terminates HTTPS at the edge and connects to your origin. **SSL mode** controls this origin connection:

| Mode | Client ↔ Cloudflare | Cloudflare ↔ Origin | Security |
|------|---------------------|---------------------|----------|
| **Off** | HTTP only | HTTP | ❌ Insecure (never use) |
| **Flexible** | HTTPS | HTTP | ❌ Insecure (exposed backend) |
| **Full** | HTTPS | HTTPS (any cert) | ⚠️ Vulnerable to MITM if origin cert is self-signed |
| **Full (Strict)** | HTTPS | HTTPS (valid cert) | ✅ **Recommended for production** |

**Production best practice:** Always use **Full (Strict)** mode. Your origin must have a valid TLS certificate (can be a free Cloudflare Origin CA certificate or Let's Encrypt). This ensures end-to-end encryption and prevents man-in-the-middle attacks between Cloudflare and your origin.

**Never use Flexible SSL in production.** It encrypts traffic between clients and Cloudflare but sends HTTP to your origin, exposing traffic within your network.

**Additional settings:**
- **Always Use HTTPS:** Redirect all HTTP requests to HTTPS (enable this via zone settings).
- **Automatic HTTPS Rewrites:** Cloudflare rewrites HTTP links in HTML to HTTPS where possible.
- **HSTS (HTTP Strict Transport Security):** Force browsers to only connect via HTTPS (enable with appropriate max-age value).

---

### Caching and Performance Optimizations

Cloudflare's CDN caches static content by default when proxied. Fine-tune with:

- **Cache Rules / Page Rules:** Control what gets cached, cache TTLs, and bypass cache for dynamic content (e.g., `/api/*` bypass cache).
- **Argo Smart Routing:** Paid add-on that routes traffic over Cloudflare's private backbone for lower latency (useful for API calls, not just static content).
- **HTTP/2 and HTTP/3:** Automatically enabled for proxied traffic.
- **Brotli Compression:** Enable in zone settings for smaller text payloads.
- **Image Optimization (Polish):** Compress images on-the-fly (Pro plan and above).

**Pro tip:** Monitor cache hit ratios in Cloudflare Analytics. Aim for high cache hit rates for static assets (images, CSS, JS) to reduce origin load.

---

### Security Best Practices

**1. Firewall Your Origin:**  
If you're proxying traffic (orange cloud), **only allow connections from Cloudflare's IP ranges** to your origin servers. Attackers shouldn't be able to bypass Cloudflare by hitting your origin IP directly. Cloudflare publishes their IP ranges—whitelist these in your firewall (iptables, security groups, cloud firewall rules).

**2. Enable WAF (Web Application Firewall):**  
Available on Pro and above. Enable managed rulesets to protect against OWASP Top 10 attacks (SQL injection, XSS, etc.). Create custom firewall rules for rate limiting, geo-blocking, or bot filtering.

**3. Hide Origin IPs:**  
Avoid creating DNS-only (grey cloud) records that expose your origin IP alongside proxied records. Attackers can find exposed IPs and bypass Cloudflare. If you need origin access for admin purposes, use Cloudflare Tunnel or VPN instead of exposing a public DNS record.

**4. Use API Tokens with Least Privilege:**  
Never use the Global API Key in production. Create scoped API tokens (e.g., `Zone:Edit` for specific zones only). Rotate tokens regularly.

**5. Monitor DNS Changes:**  
Use audit logs (Enterprise) or version control (if managing via Terraform/Pulumi) to track who changed what. Unexpected DNS changes can indicate compromise.

---

### Common Anti-Patterns to Avoid

**❌ Leaving zones in "Paused" state unintentionally:**  
Paused mode disables all proxying and security features (DNS-only mode for entire zone). Only use temporarily for debugging.

**❌ Mixing multiple IaC tools on the same zone:**  
Don't manage a zone with both Terraform and manual API calls. Pick one source of truth to avoid conflicts and drift.

**❌ Forgetting to lower TTLs before migration:**  
When migrating DNS to Cloudflare from another provider, lower TTLs on old records a day in advance. This ensures quick propagation when you switch nameservers.

**❌ Not testing DNS before going live:**  
Before changing nameservers at your registrar, query Cloudflare's assigned nameservers directly to verify records are correct:
```bash
dig @gina.ns.cloudflare.com example.com A
```

**❌ Ignoring MX record caveats:**  
Always grey-cloud MX records and the A/AAAA records mail servers point to. Proxying email breaks SMTP.

---

## The 80/20 Configuration: What Most Users Actually Need

Cloudflare offers hundreds of settings across zones, DNS, SSL, caching, firewall, Workers, and more. But for most deployments, you need only a handful of core configurations to get 80% of the value.

### Core Fields for Zone Creation

**`CloudflareDnsZoneSpec` (OpenMCF's protobuf):**

```protobuf
message CloudflareDnsZoneSpec {
  // The domain name (e.g., "example.com"). Required.
  string zone_name = 1;

  // Cloudflare account ID. Required.
  string account_id = 2;

  // Plan tier: FREE (default), PRO, BUSINESS, ENTERPRISE.
  CloudflareDnsZonePlan plan = 3;

  // Paused state: if true, zone is DNS-only with no proxy/WAF/CDN. Default: false.
  bool paused = 4;

  // Default proxied: if true, new DNS records default to orange-cloud. Default: false.
  bool default_proxied = 5;
}
```

### Why These Five Fields?

| Field | Why It Matters | Default/Typical Value |
|-------|---------------|----------------------|
| **zone_name** | The domain you're managing (e.g., `example.com`) | *Required* (no default) |
| **account_id** | Cloudflare account context (for API token scope, billing) | *Required* (no default) |
| **plan** | Free vs paid features (WAF, support, SLA) | `FREE` (can upgrade to PRO/BUSINESS) |
| **paused** | Whether zone is active (proxying enabled) or paused (DNS-only) | `false` (active by default) |
| **default_proxied** | Default new DNS records to orange-cloud (safer) vs grey-cloud | `false` (can set to `true` for security) |

### What's Not Included (and Why)

**Zone type (full vs partial):**  
Partial (CNAME) setup is Business/Enterprise only and is an advanced use case (needed when you can't delegate nameservers). For 80% of users, full setup (standard nameserver delegation) is the only mode.

**Jump-start (DNS scan on creation):**  
Convenience feature that scans existing DNS and imports records. Useful for migrations but not critical when defining zones as code (you'll define records explicitly).

**DNSSEC settings:**  
Typically enabled post-creation via separate API call or dashboard toggle (not part of initial zone spec).

**Advanced nameservers (vanity NS):**  
Business/Enterprise feature for branding (e.g., `ns1.yourdomain.com` instead of Cloudflare's NS). Rare need.

**Zone settings (SSL mode, caching rules, WAF):**  
Managed separately via the per-setting `cloudflare_zone_setting` resource in Terraform or dedicated API calls. Not part of the zone creation spec—these are post-creation configurations.

---

### Example Configurations

#### Development: Minimal Free Zone

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareDnsZone
metadata:
  name: dev-example-zone
spec:
  zone_name: "dev.example.com"
  account_id: "abc123..."
  plan: FREE
  paused: false
  default_proxied: true  # Orange-cloud new records by default
```

**Use case:** Developer testing environment. Free plan, proxying enabled by default to protect origin IPs even in dev.

---

#### Staging: Pro Plan with Proxying

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareDnsZone
metadata:
  name: staging-zone
spec:
  zone_name: "staging.example.com"
  account_id: "abc123..."
  plan: PRO
  paused: false
  default_proxied: true
```

**Use case:** Staging environment with Pro-level WAF and SSL features. Orange-cloud by default to mirror production config.

---

#### Production: Business Plan, Explicit Control

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareDnsZone
metadata:
  name: production-zone
spec:
  zone_name: "example.com"
  account_id: "xyz789..."
  plan: BUSINESS
  paused: false
  default_proxied: false  # Explicitly set proxy per record in prod
```

**Use case:** Production site on Business plan (SLA, 24/7 support, custom nameservers). `default_proxied: false` because operations team wants to explicitly control which records are proxied (some internal APIs stay grey-cloud).

---

## OpenMCF's Approach: Abstraction with Pragmatism

OpenMCF provides a **protobuf-defined API** that abstracts Cloudflare DNS zone creation into a clean, multi-cloud-consistent interface. The philosophy: expose the essential 20% of configuration, hide complexity, and make DNS deployment feel like deploying any other cloud resource.

### What We Abstract

**Unified API across cloud providers:**  
Whether you're deploying AWS Route 53, GCP Cloud DNS, or Cloudflare DNS, OpenMCF's resource specs follow similar patterns (zone name, account context, plan tier). This makes multi-cloud DNS management predictable.

**Intelligent defaults:**
- `plan` defaults to `FREE` (most users start here)
- `paused` defaults to `false` (zones are active by default)
- `default_proxied` defaults to `false` (but can be set to `true` for orange-cloud-by-default behavior)

**Separation of concerns:**  
Zone creation is separate from DNS record management, zone settings configuration, and DNSSEC enablement. This prevents overwhelming users with dozens of options when all they need is "create a zone for `example.com`."

### Under the Hood: Pulumi

OpenMCF currently uses **Pulumi (Go SDK)** to provision Cloudflare DNS zones. Why Pulumi over Terraform?

- **Language flexibility:** Pulumi's Go SDK integrates naturally into OpenMCF's multi-cloud orchestration layer (also written in Go).
- **Programmatic control:** Pulumi's code-native model simplifies conditional logic (e.g., "if Enterprise plan, enable advanced features").
- **Equivalent coverage:** Pulumi's Cloudflare provider (bridged from Terraform) supports all necessary zone operations.

**Note:** The choice of Pulumi vs Terraform is an implementation detail. The protobuf API (`CloudflareDnsZone`) remains the same regardless of the underlying provisioner. Users interact with the abstraction, not Pulumi/Terraform directly.

### Opinionated but Flexible

**Opinionated defaults:**
- We default to full zone setup (standard nameserver delegation), not partial CNAME setup, because it's simpler and suitable for 90% of use cases.
- We don't auto-enable jump-start (DNS scan) because infrastructure-as-code should define records explicitly, not rely on auto-discovery.

**Flexibility where it matters:**
- Users can choose Free, Pro, Business, or Enterprise plans via the `plan` field.
- Users can set `default_proxied: true` to enforce orange-cloud defaults (security-conscious teams prefer this).
- Users can pause zones (`paused: true`) for maintenance or debugging.

---

## Migrating DNS to Cloudflare: Zero-Downtime Strategy

Migrating DNS from another provider (AWS Route 53, GCP Cloud DNS, GoDaddy, etc.) to Cloudflare requires careful sequencing to avoid downtime:

### Pre-Migration Checklist

**1. Audit existing DNS records:**  
Export a zone file or list all records from your current provider. Ensure you capture:
- A/AAAA records (IPv4/IPv6)
- MX records (email)
- TXT records (SPF, DKIM, DMARC, domain verification)
- SRV records (services like SIP, LDAP)
- CAA records (certificate authority authorization)

**2. Lower TTLs 24-48 hours before migration:**  
Reduce TTLs on all records (especially NS, A, CNAME) to 300 seconds (5 minutes). This ensures DNS caches expire quickly after you switch nameservers.

**3. Create zone in Cloudflare:**  
Use Terraform, Pulumi, or the API to create the zone. Cloudflare will assign nameservers (e.g., `gina.ns.cloudflare.com`, `ivan.ns.cloudflare.com`). Don't update your registrar yet.

**4. Import DNS records into Cloudflare:**  
- **Option A:** Upload a BIND zone file via Cloudflare's dashboard or API.
- **Option B:** Use Cloudflare's "scan" feature during zone creation (attempts to auto-discover records).
- **Option C:** Define all records as code in Terraform/Pulumi (recommended for repeatability).

**5. Validate records on Cloudflare's nameservers (before delegation):**  
Query Cloudflare's assigned nameservers directly to verify records are correct:
```bash
dig @gina.ns.cloudflare.com example.com A
dig @gina.ns.cloudflare.com example.com MX
dig @gina.ns.cloudflare.com www.example.com CNAME
```
Check that all critical records (website, email, API endpoints) return correct values.

---

### Migration Day

**1. Update nameservers at your registrar:**  
Replace all existing nameservers with Cloudflare's assigned NS. Remove old NS entirely (don't leave a mix).

**2. Monitor propagation:**  
- Use [WhatsMyDNS.net](https://www.whatsmydns.net/) to check global propagation.
- Cloudflare's dashboard will detect the NS change and mark the zone as "Active" (usually within minutes to an hour).

**3. Verify traffic is flowing through Cloudflare:**  
- Visit your website and check HTTP response headers for Cloudflare (`CF-RAY`, `CF-Cache-Status`).
- Test critical services: email, APIs, subdomains.
- Run DNS queries from multiple locations to confirm Cloudflare is responding.

**4. Enable DNSSEC (post-migration):**  
Once the zone is stable, enable DNSSEC in Cloudflare and add the DS record to your registrar.

---

### Post-Migration

**1. Restore TTLs:**  
Increase TTLs back to reasonable values (e.g., 3600 = 1 hour, or 86400 = 1 day for stable records).

**2. Enable orange-cloud (proxying) for web services:**  
If you migrated with all records grey-clouded (DNS-only) for safety, gradually enable proxying (orange cloud) for `www`, `app`, `api`, etc., and monitor for issues.

**3. Configure zone settings:**  
Set SSL mode to **Full (Strict)**, enable **Always Use HTTPS**, configure WAF rules, set up caching rules.

**4. Decommission old DNS provider:**  
Once you've confirmed everything works for 24-48 hours, you can delete the zone from your old provider (to avoid paying for it).

---

### Rollback Plan

If something goes wrong:
- Change nameservers back to the old provider at your registrar.
- Cloudflare's zone remains intact (no data loss), but traffic routes back to the old DNS.
- Debug the issue, fix records in Cloudflare, and attempt migration again.

**Pro tip:** Keep the old DNS provider's zone intact for 7 days after migration as a safety net.

---

## Key Takeaways

1. **Cloudflare DNS offers a unique value proposition:** Authoritative DNS served from 330+ global locations, zero per-query charges (even on Free plan), integrated DDoS protection, and optional CDN/WAF/proxy capabilities. It's not just a DNS provider—it's an edge platform.

2. **Manual management doesn't scale.** Use Infrastructure-as-Code (Terraform or Pulumi) for production. Both are mature, production-ready, and support full lifecycle management of zones, records, and settings.

3. **The 80/20 configuration is zone name, account ID, plan, paused state, and default proxy behavior.** These five fields cover the vast majority of zone creation needs. Advanced features (partial zones, vanity nameservers, zone transfers) are Enterprise-only edge cases.

4. **Orange cloud vs grey cloud is a critical decision.** Proxy (orange) web services to get CDN, WAF, and DDoS protection. DNS-only (grey) infrastructure like email, SSH, and non-HTTP services.

5. **Always use Full (Strict) SSL mode in production.** This ensures end-to-end encryption and prevents man-in-the-middle attacks between Cloudflare and your origin.

6. **DNSSEC is a one-click security win.** Enable it on all production zones to prevent DNS spoofing.

7. **For Kubernetes, combine Terraform/Pulumi (zone creation) with ExternalDNS (dynamic record management)** to get the best of both worlds: declarative zone config and automatic DNS updates for ephemeral services.

8. **OpenMCF abstracts Cloudflare's complexity** into a clean protobuf API, making multi-cloud DNS management consistent while respecting Cloudflare's unique strengths (proxying, security, global network).

---

## Further Reading

- **Cloudflare DNS Documentation:** [Cloudflare Docs - DNS](https://developers.cloudflare.com/dns/)
- **Terraform Cloudflare Provider:** [GitHub - cloudflare/terraform-provider-cloudflare](https://github.com/cloudflare/terraform-provider-cloudflare)
- **Pulumi Cloudflare Provider:** [Pulumi Registry - Cloudflare](https://www.pulumi.com/registry/packages/cloudflare/)
- **ExternalDNS Cloudflare Tutorial:** [Kubernetes ExternalDNS](https://kubernetes-sigs.github.io/external-dns/v0.14.2/tutorials/cloudflare/)
- **Crossplane Cloudflare Provider:** [GitHub - crossplane-contrib/provider-cloudflare](https://github.com/crossplane-contrib/provider-cloudflare)
- **Cloudflare API Reference:** [Cloudflare API - Zones](https://developers.cloudflare.com/api/operations/zones-post)
- **Cloudflare for SaaS (Custom Hostnames):** [Cloudflare Docs - SaaS](https://developers.cloudflare.com/cloudflare-for-platforms/cloudflare-for-saas/)

---

**Bottom Line:** Cloudflare DNS is more than DNS—it's a global edge platform that gives you authoritative DNS, DDoS protection, CDN capabilities, and a WAF in one unified service, with zero per-query fees and sub-second propagation. Manage it with Terraform or Pulumi for production reliability, enable DNSSEC for security, proxy web traffic (orange cloud) for protection and performance, and leverage Kubernetes integrations like ExternalDNS for dynamic environments. OpenMCF makes this simple with a protobuf API that focuses on the essential configuration while hiding unnecessary complexity.


---
title: "DNS Zone: Managing DNS in a Multi-Cloud World"
description: "DNS Zone: Managing DNS in a Multi-Cloud World deployment documentation"
icon: "package"
order: 100
componentName: "civodnszone"
---

# Civo DNS Zone: Managing DNS in a Multi-Cloud World

## Introduction

DNS management often feels like an afterthought in cloud infrastructure—until a misconfigured record takes down your production site. For teams running workloads on Civo, there's a compelling option that's frequently overlooked: **Civo's built-in DNS service**.

Unlike heavyweight DNS providers with complex feature sets and per-query pricing, Civo DNS takes a refreshingly simple approach. It's included free with your Civo account, integrates seamlessly with Civo's API and tooling, and covers the essential DNS needs of most applications without the bloat. But is it production-ready? When should you use it versus established providers like Route 53 or Cloudflare? And how do you manage it declaratively as part of your infrastructure as code?

This document answers those questions. We'll explore the landscape of DNS management options, from manual console changes to GitOps-driven automation, and explain why OpenMCF defaults to a minimal, focused API that captures the 80% of DNS configuration that matters to 80% of users.

## The DNS Management Maturity Spectrum

### Level 0: Manual Console Changes (The Fragile Approach)

The easiest way to get started with Civo DNS is through the web dashboard. Click "DNS" in the navigation, add your domain name, and start filling in records through a point-and-click interface. Each record type (A, CNAME, MX, TXT) has a simple form where you enter the name, value, and TTL.

**The problem:** This works fine for your first domain or a quick test, but it doesn't scale. Changes made in a web UI aren't tracked in version control. There's no audit trail of who changed what and why. Rolling back a mistake means remembering what the old values were. And if you need to replicate a DNS setup across environments (dev, staging, prod), you're stuck manually copying values—a recipe for drift and errors.

**Verdict:** Acceptable for initial exploration or emergency fixes, but not a production pattern.

### Level 1: CLI Scripts (Repeatable but Imperative)

The Civo CLI (`civo`) provides commands for DNS management that can be scripted. You can create zones with `civo domain create example.com` and add records with `civo domain record create`. These commands can be checked into a script, giving you repeatability and some version control.

**What it solves:** You now have auditable changes (the script is in Git) and can replay the setup in new environments. For dynamic DNS scenarios (like updating an A record when your home IP changes), a cron job running the CLI works well.

**What it doesn't solve:** This is still imperative—you're giving commands to execute, not declaring desired state. If someone makes a manual change in the console, your script won't detect the drift. Deleting a record requires remembering to remove it from the script *and* run a delete command. There's no reconciliation loop ensuring your DNS matches your intent.

**Verdict:** A step up from manual, useful for scripting one-off tasks, but lacks the safety and consistency of declarative IaC.

### Level 2: Infrastructure as Code (Declarative State)

The game-changer for DNS management is treating it as code with tools like **Terraform, Pulumi, or OpenTofu**. Civo maintains an official Terraform provider (mature, version 1.x) that exposes DNS zones and records as resources. In Pulumi, the Civo provider uses the same underlying implementation, so feature parity is excellent.

Here's what this looks like in Terraform:

```hcl
resource "civo_dns_domain" "my_zone" {
  name = "example.com"
}

resource "civo_dns_domain_record" "www" {
  domain_id = civo_dns_domain.my_zone.id
  name      = "www"
  type      = "CNAME"
  value     = "example.com"
  ttl       = 3600
}

resource "civo_dns_domain_record" "mail" {
  domain_id = civo_dns_domain.my_zone.id
  name      = "@"
  type      = "MX"
  value     = "mail.example.com"
  priority  = 10
  ttl       = 3600
}
```

**What this approach delivers:**

- **Declarative state management:** You describe *what* DNS should look like, not *how* to get there. Terraform handles creation, updates, and (critically) deletions when you remove records from code.
- **Version control and review:** DNS changes go through pull requests. You can see exactly what will change with `terraform plan` before applying.
- **Idempotency:** Running `terraform apply` multiple times with the same config is safe—it converges to the desired state rather than duplicating records.
- **Multi-environment consistency:** Use Terraform workspaces or separate directories to manage dev/staging/prod zones with the same code but different values.

**Limitations to acknowledge:** IaC tools introduce state management complexity. You need to store Terraform state securely (use remote backends like S3) and prevent concurrent modifications. There's also a learning curve for teams new to IaC. But for production systems, this discipline is essential.

**Verdict:** The production-ready baseline for managing DNS. If you're serious about infrastructure automation, start here.

### Level 3: Kubernetes-Native DNS (GitOps Integration)

For teams running applications on Kubernetes, two powerful integrations elevate DNS management to a GitOps workflow:

#### ExternalDNS: Automatic Service Discovery

**ExternalDNS** is a Kubernetes controller that watches Service and Ingress resources and automatically creates DNS records pointing to them. When you deploy a service with a LoadBalancer type or create an Ingress with a hostname annotation, ExternalDNS detects it and calls the Civo DNS API to create the corresponding A or CNAME record.

This is transformative for Kubernetes workflows. Instead of manually updating DNS every time you deploy or scale a service, DNS becomes a side-effect of your application deployment. The records are always in sync with your actual services.

Civo DNS has been supported in ExternalDNS since version 0.13.5. Setup requires:
- Deploying ExternalDNS in your cluster with `--provider=civo`
- Providing your Civo API token as an environment variable
- Setting `--domain-filter` to scope which domains ExternalDNS manages

**Example use case:** You deploy an Ingress for `app.example.com`. ExternalDNS sees the Ingress, checks that `example.com` is in the allowed domains, and creates an A record for `app.example.com` pointing to your ingress controller's IP. When you delete the Ingress, ExternalDNS removes the DNS record. Zero manual DNS work.

#### cert-manager DNS-01 Challenges: Automated TLS Certificates

Getting wildcard TLS certificates from Let's Encrypt requires DNS-01 challenges—proving domain ownership by creating specific TXT records. **cert-manager**, the standard Kubernetes solution for certificate automation, can integrate with Civo DNS through a webhook.

The community-maintained [cert-manager webhook for Civo DNS](https://github.com/okteto/cert-manager-webhook-civo) handles this. When cert-manager needs to validate `*.example.com`, it calls the webhook, which uses the Civo API to create the `_acme-challenge.example.com` TXT record with Let's Encrypt's token. After validation, it cleans up the record.

**What this enables:** Fully automated wildcard certificate issuance and renewal for Kubernetes Ingress, with zero manual DNS steps.

**Verdict:** For Kubernetes-native teams, this is the pinnacle—DNS and TLS management fully integrated into your GitOps workflow. Your application manifests are the source of truth, and DNS/certificates automatically follow.

### Level 4: Multi-Cloud Abstraction (OpenMCF's Approach)

OpenMCF takes DNS management one step further: a **provider-agnostic API** that abstracts Civo DNS (and other providers) behind a unified Protobuf schema. Instead of writing Terraform HCL or Pulumi code specific to Civo, you define a `CivoDnsZone` resource in a standard format, and OpenMCF handles the provider-specific implementation.

This matters in multi-cloud environments. If you manage infrastructure across Civo, AWS, and GCP, you can use a consistent API pattern for DNS zones regardless of provider. The abstraction also enforces best practices—our schema focuses on the essential 80% of configuration (domain name, records with name/type/value/TTL), avoiding the complexity of rarely-used fields.

**How it works:** OpenMCF's `CivoDnsZone` CRD (Custom Resource Definition) lets you declare:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoDnsZone
metadata:
  name: example-zone
spec:
  domainName: example.com
  records:
    - name: "@"
      type: A
      values:
        - value: "198.51.100.42"
      ttlSeconds: 3600
    - name: "www"
      type: CNAME
      values:
        - value: "example.com"
      ttlSeconds: 3600
```

Behind the scenes, OpenMCF uses Pulumi or Terraform modules to provision the actual zone and records on Civo. But from your perspective, you're working with a clean, opinionated API that hides provider quirks.

**Verdict:** Ideal for teams standardizing on OpenMCF for multi-cloud infrastructure, or anyone who values API consistency over provider-specific features.

## When to Choose Civo DNS vs. Dedicated Providers

Civo DNS is **not** a one-size-fits-all solution, and that's okay. Understanding when to use it versus alternatives is key to making informed decisions.

### Use Civo DNS When:

- **Most of your infrastructure is on Civo.** Having DNS credentials and API access consolidated with your compute makes automation simpler.
- **Cost is a consideration.** Civo DNS is free with your account—no per-zone or per-query charges. Route 53 charges ~$0.50/month per hosted zone plus per-million-query fees.
- **Your DNS needs are straightforward.** You need A, CNAME, MX, and TXT records for web services, APIs, and email. You don't require advanced routing (geo-based, latency-based, health-check failover).
- **You want minimal setup.** For dev/test environments or internal tools, Civo DNS is quick to configure and "just works" without extra accounts or billing.

### Consider Dedicated DNS Providers (Route 53, Cloudflare, etc.) When:

- **You need advanced features.** Civo DNS lacks:
  - **DNSSEC** (domain signing for spoofing protection)
  - **IPv6 AAAA records** (currently unsupported in Civo's API)
  - **Health-checked failover** (automatic DNS updates based on endpoint health)
  - **Geographic or latency-based routing** (serving different IPs based on user location)
  
- **Multi-cloud neutrality is critical.** If you run workloads on Civo, AWS, and Azure, you might prefer a neutral DNS provider (like Cloudflare) that isn't tied to any single cloud. This avoids vendor lock-in and simplifies credential management.

- **You require extremely high query volumes or SLAs.** Civo DNS uses two authoritative nameservers (`ns0.civo.com` and `ns1.civo.com`), which is adequate for most use cases. Providers like Route 53 use four nameservers with massive global anycast networks. For mission-critical domains serving millions of queries per second, you might want that extra redundancy.

- **You need specialized DNS-based services.** Cloudflare bundles DNS with CDN, DDoS protection, and Web Application Firewall. If you're using those services, managing DNS there makes sense.

### The Hybrid Approach

Many teams use **both**. For example:
- Use Cloudflare DNS for your primary public-facing domain (to get CDN and security features)
- Use Civo DNS for internal service domains or development environments

There's no rule that you must consolidate all DNS in one provider. The key is having automation (IaC or ExternalDNS) so that you're not manually managing multiple consoles.

## The 80/20 Configuration Principle

When designing the OpenMCF API for Civo DNS, we analyzed what users *actually configure* in production. The Pareto principle holds true: **80% of DNS use cases require only 20% of possible configuration fields**.

### What Users Actually Need

**For the DNS zone:**
- **Domain name** (e.g., `example.com`) – That's it. Civo doesn't expose SOA tuning, DNSSEC settings, or custom nameserver configuration at the API level. Those are managed internally.

**For DNS records:**
- **Name** – The hostname or subdomain (e.g., `www`, `api`, `@` for the root)
- **Type** – A, CNAME, MX, TXT cover ~90% of real-world needs. AAAA (IPv6) and SRV (service records) are occasionally used. Exotic types like CAA or TLSA are rare.
- **Value(s)** – The IP address, target hostname, or text content. For round-robin A records, you might have multiple values.
- **TTL** – Time-to-live in seconds. Most users are fine with 3600 (1 hour) as a sensible default. Some tweak this for specific records (e.g., lower TTL before a migration).
- **Priority** (MX only) – For mail exchange records, the priority determines preference order. Default of 10 works for single-server setups.

### What Users Rarely Touch

- **Per-record priority/weight/port** (beyond MX) – SRV records have these, but SRV is niche (Active Directory, some VoIP setups).
- **Comments/descriptions** – Not part of DNS protocol; users add these in IaC code comments instead.
- **DNSSEC keys and signing** – Civo doesn't support DNSSEC, and many orgs still don't use it due to operational complexity.
- **Custom nameservers per record** – Not applicable; the zone's NS records are fixed by Civo.

### Example Configurations

**Basic web hosting:**
```yaml
domainName: example.com
records:
  - name: "@"          # Apex domain
    type: A
    values: ["198.51.100.42"]
    ttlSeconds: 3600
  - name: "www"
    type: CNAME
    values: ["example.com"]
    ttlSeconds: 3600
```

This gives you `example.com` and `www.example.com` both resolving to the same IP (via CNAME to the apex).

**Email setup with SPF:**
```yaml
domainName: example.com
records:
  - name: "@"
    type: MX
    values: ["10 mail.example.com"]
    ttlSeconds: 3600
  - name: "mail"
    type: A
    values: ["198.51.100.50"]
    ttlSeconds: 3600
  - name: "@"
    type: TXT
    values: ["v=spf1 mx ~all"]
    ttlSeconds: 3600
```

This routes mail to `mail.example.com` (which resolves to your mail server IP) and sets SPF policy to allow mail from your MX servers.

**Modern app with API and CDN:**
```yaml
domainName: example.com
records:
  - name: "api"
    type: A
    values: ["203.0.113.10"]
    ttlSeconds: 3600
  - name: "cdn"
    type: CNAME
    values: ["cdn-provider.example.net"]
    ttlSeconds: 3600
  - name: "_acme-challenge"
    type: TXT
    values: ["ACME_token_for_validation"]
    ttlSeconds: 600   # Lower TTL for validation speed
```

The API subdomain points to your Civo load balancer, the CDN subdomain aliases to an external CDN, and you have a TXT record for Let's Encrypt DNS-01 challenges (which cert-manager can manage automatically).

### API Design Implications

OpenMCF's `CivoDnsZoneSpec` captures exactly these fields:
- `domain_name` (required)
- `records[]` list with:
  - `name` (required)
  - `type` (enum: A, AAAA, CNAME, MX, TXT, SRV, etc.)
  - `values[]` (one or more, to support round-robin or multi-value records)
  - `ttl_seconds` (default: 3600)

By focusing on this minimal set, we keep the API approachable for new users while still covering advanced scenarios (like multiple A records for simple load distribution, or TXT records for DKIM/DMARC/ACME challenges).

## Production Best Practices

### 1. Always Use Infrastructure as Code

Treat DNS changes with the same rigor as application code. Use Terraform, Pulumi, or OpenMCF's declarative API. Check your DNS configuration into Git, review changes via pull requests, and apply through CI/CD pipelines. This prevents the "mystery record" problem where no one knows why a particular DNS entry exists or what will break if it's removed.

### 2. Set Nameservers at Your Registrar

After creating a DNS zone in Civo, you **must** update your domain's nameserver records at the registrar (where you purchased the domain) to point to Civo's nameservers: `ns0.civo.com` and `ns1.civo.com`. This delegates DNS authority to Civo. Forgetting this step is the #1 reason new users see "DNS not resolving"—Civo has your records, but the world is still asking the old DNS servers.

### 3. Plan TTL Values Strategically

- **Default TTL of 3600 (1 hour)** is a sensible balance for most records. It allows changes to propagate reasonably quickly without excessive DNS query load.
- **Lower TTL before migrations:** If you're planning to change an A record (e.g., migrating to a new IP), reduce the TTL to 300 (5 minutes) a day in advance. After the cutover, raise it back to 3600.
- **Higher TTL for stable records:** MX records, SPF/TXT records for email config, and other rarely-changed records can use 7200 (2 hours) or even 86400 (24 hours) if you prefer.

Civo enforces a minimum TTL of 600 seconds, so you can't go lower than that.

### 4. Email Records: Get Them Right

If your domain handles email (even if it's just forwarding to Google Workspace or Office 365), you need:
- **MX records** pointing to your mail provider's servers (with correct priority values)
- **SPF record** (a TXT record at the root: `v=spf1 include:_spf.google.com ~all`)
- **DKIM record** (a TXT record at a selector subdomain, e.g., `google._domainkey`)
- **DMARC record** (a TXT record at `_dmarc`: `v=DMARC1; p=quarantine; rua=mailto:postmaster@example.com`)

Mistakes here can cause email delivery failures or spoofing vulnerabilities. Test with tools like [MXToolbox](https://mxtoolbox.com/) after making changes.

### 5. Monitor DNS Resolution

Civo doesn't provide query analytics or built-in DNS monitoring. Use external monitoring tools (UptimeRobot, Pingdom, Datadog, etc.) to periodically resolve your critical hostnames (website, API endpoints) and alert if they fail to resolve or return unexpected IPs. This is especially important if you use automation (like ExternalDNS) that could accidentally delete records.

### 6. Backup Your DNS Configuration

While IaC code *is* your backup (you can always re-apply Terraform to recreate records), it's wise to periodically export your DNS zone as a safety net. Use the Civo API or CLI to fetch all records and store them:

```bash
civo domain record list example.com --output=json > example.com-backup.json
```

In a disaster scenario (like accidentally deleting the zone), you can restore from this.

### 7. Avoid Common Anti-Patterns

- **Don't use CNAME at the root domain.** DNS protocol forbids it. Use an A record for `@` (the apex).
- **Don't leave orphaned records.** When decommissioning a service, remove its DNS records. Old records pointing to unused IPs can be a security risk (someone else might claim that IP).
- **Don't over-rely on wildcard records.** Wildcards (`*.example.com`) are useful for catching all subdomains, but they can mask missing records and create unexpected behavior. Use them intentionally, not as a crutch.

## Multi-Cloud and Integration Patterns

### ExternalDNS for Kubernetes

If you run Kubernetes on Civo, ExternalDNS is a force multiplier. It turns DNS into a declarative extension of your Kubernetes manifests. Annotate a Service or Ingress with a hostname, and ExternalDNS creates the DNS record. Delete the resource, and the record is cleaned up.

**Setup is straightforward:**
1. Deploy ExternalDNS in your cluster with `--provider=civo` and your API token in a Secret.
2. Set `--domain-filter=example.com` to limit which domains it manages (safety measure).
3. Annotate your Ingress with `external-dns.alpha.kubernetes.io/hostname: app.example.com`.

ExternalDNS handles the rest. For multi-cloud, you can run ExternalDNS in clusters on different clouds (Civo and AWS), each managing their own subdomains or using label filters to avoid collisions.

### cert-manager DNS-01 Integration

Wildcard certificates require DNS-01 challenges. The [cert-manager webhook for Civo DNS](https://github.com/okteto/cert-manager-webhook-civo) enables this. Install the webhook, create an Issuer configured to use it, and cert-manager will automatically create/delete the `_acme-challenge` TXT records needed for Let's Encrypt validation.

This is essential for securing multiple subdomains with a single certificate (e.g., `*.example.com`).

### CI/CD Automation

For non-Kubernetes workflows, integrate DNS updates into your CI/CD pipeline:
- In GitHub Actions, run Terraform to apply DNS changes when merging to `main`.
- In GitLab CI, use Terraform Cloud or Atlantis for plan/apply workflows with approval gates for DNS changes.
- For blue/green deployments, your pipeline can update an A record to point to the new environment after health checks pass.

### Multi-Cloud DNS Strategies

If your architecture spans multiple clouds:
- **Centralized DNS on Civo:** Use Civo DNS for all domains, regardless of where workloads run. A records can point to IPs in any cloud.
- **Provider-per-cloud:** Use Civo DNS for Civo workloads, Route 53 for AWS, Cloud DNS for GCP. Manage with separate Terraform configs or a unified tool like OpenMCF.
- **Neutral provider:** Use Cloudflare or another neutral DNS provider for everything, decoupling DNS from any single cloud. This avoids provider lock-in but adds another service to manage.

There's no single "right" answer—choose based on your team's priorities (cost, simplicity, redundancy).

## Civo DNS vs. Other Providers: A Quick Comparison

| Feature | Civo DNS | AWS Route 53 | Cloudflare DNS | Google Cloud DNS |
|---------|----------|--------------|----------------|------------------|
| **Pricing** | Free | ~$0.50/zone/month + per-query fees | Free (unlimited queries) | ~$0.20/zone/month + per-query fees |
| **Record Types** | A, CNAME, MX, TXT, SRV | All standard types + ALIAS | All standard types | All standard types |
| **IPv6 (AAAA)** | Not supported (as of 2025) | Supported | Supported | Supported |
| **DNSSEC** | Not supported | Supported | Supported | Supported |
| **Health-based routing** | No | Yes (health checks + failover) | Yes (Load Balancing add-on) | No (use Traffic Director) |
| **Geo-routing** | No | Yes | Yes (with Load Balancing) | No |
| **Terraform Support** | Official provider | Official provider | Official provider | Official provider |
| **ExternalDNS** | Supported | Supported | Supported | Supported |
| **Best For** | Civo-centric workloads, cost-sensitive projects | AWS-heavy architectures, advanced routing | Public domains, CDN integration | GCP-heavy architectures |

**Takeaway:** Civo DNS trades advanced features for simplicity and cost. For the majority of use cases (web apps, APIs, email), it's perfectly sufficient. For edge cases requiring geo-routing or DNSSEC, consider a specialized provider.

## Why OpenMCF Chooses This Approach

OpenMCF's `CivoDnsZone` API is intentionally minimal. We focus on the **core fields that 80% of users need** (domain name, record name/type/value/TTL) and omit the rarely-used 20% (complex CAA syntax, exotic record types, per-record metadata).

This decision is grounded in real-world usage patterns. After analyzing production DNS configurations across hundreds of domains, the pattern is clear: most zones consist of a handful of A, CNAME, MX, and TXT records. Advanced features like SRV records or custom TTLs per record exist but are exceptions, not the norm.

By keeping the API simple:
- **New users aren't overwhelmed.** You can define a working DNS zone in a few lines of YAML.
- **The API remains stable.** We're not chasing every DNS protocol extension or provider-specific quirk.
- **Multi-cloud consistency is easier.** When we add support for other DNS providers, the API surface area is small and transferable.

For users who need advanced configuration, they can always drop down to Terraform or Pulumi. OpenMCF doesn't prevent that—it just provides a higher-level, opinionated path for common cases.

## Conclusion

DNS management has evolved from manual console edits and shell scripts to declarative IaC and GitOps-integrated automation. Civo DNS, while simpler than heavyweight providers, fits beautifully into modern infrastructure workflows when paired with tools like Terraform, ExternalDNS, and cert-manager.

The key insight is this: **most applications don't need the complexity of advanced DNS routing or DNSSEC**. They need reliable, cost-effective DNS that integrates seamlessly with their existing tooling. Civo DNS delivers that, and OpenMCF's minimal API makes it even easier to use.

Whether you're deploying a side project, a production SaaS, or a multi-cloud application, understanding the DNS management spectrum—from manual to fully automated—empowers you to choose the right approach for your context. Start simple, automate early, and layer in sophistication only when your requirements demand it.

For deeper implementation guides, see:
- [Setting up ExternalDNS with Civo](https://kubernetes-sigs.github.io/external-dns/v0.14.2/tutorials/civo/) (Kubernetes integration)
- [cert-manager webhook for Civo DNS](https://github.com/okteto/cert-manager-webhook-civo) (automated TLS)
- [Civo Terraform Provider Documentation](https://registry.terraform.io/providers/civo/civo/latest/docs) (IaC reference)

And remember: DNS is too important to wing it. Treat it like code, version it, review it, and automate it. Your future self—and your on-call team—will thank you.


# DigitalOcean DNS Zone Deployment: A Pragmatic Choice

## Introduction: Simplicity as a Strategic Asset

In the DNS provider landscape, DigitalOcean DNS represents a refreshing architectural choice: it is not designed to compete as a feature-rich, globally distributed DNS platform like AWS Route 53 or Cloudflare. Instead, its value proposition is **simplicity and ecosystem integration**.

DigitalOcean DNS is offered as a free, API-first service that allows you to consolidate DNS management on the same platform where your Droplets, Load Balancers, and Spaces object storage reside. This makes it an ideal choice for teams that:

- Run the majority of their infrastructure on DigitalOcean
- Value predictable, transparent pricing over feature complexity
- Prioritize ease of use and quick time-to-production
- Don't require advanced DNS features like DNSSEC, geo-routing, or weighted traffic policies

However, this simplicity comes with clear trade-offs. DigitalOcean DNS lacks DNSSEC support entirely and provides no zone import functionality, making large-scale migrations painful. Understanding these constraints is essential for making informed architectural decisions.

This document explores the evolution of DNS management methods—from manual control panel clicking to declarative Infrastructure as Code—and explains why Planton's approach balances flexibility with the platform's 80/20 design philosophy.

## The DNS Management Evolution

### Level 0: Manual Control Panel Management (The Starting Point)

For teams with a handful of domains, the DigitalOcean Control Panel provides a straightforward web interface for DNS management. You navigate to Networking → Domains, add a domain, and manually create records (A, CNAME, MX, TXT) through web forms.

**Why this works initially:**
- Zero learning curve for non-technical users
- Immediate visual feedback
- Integrated with Droplet selection (point records at your Droplets by name, not just IP)

**Why this breaks down:**
- **Manual errors:** Typos in record values, incorrect TTL settings, and forgotten delegation at the registrar are common pitfalls
- **No audit trail:** Changes are not version-controlled; tracking who changed what and when is impossible
- **Migration friction:** DigitalOcean lacks a zone import feature. Migrating an existing domain with 50+ records means manually recreating every single record—a process that is both tedious and error-prone
- **DNSSEC trap:** The most common production-impacting mistake is delegating a domain to DigitalOcean's nameservers (ns1.digitalocean.com, ns2.digitalocean.com, ns3.digitalocean.com) without first disabling DNSSEC at the registrar. Since DigitalOcean does not support DNSSEC, resolvers will fail validation and the domain becomes unreachable

**Verdict:** Acceptable for learning and prototyping with 1-3 simple domains. Not viable for production teams managing multiple environments or domains.

### Level 1: CLI and Scripting (Ad-Hoc Automation)

The next step is using DigitalOcean's official CLI, `doctl`, for record creation and management via shell scripts.

**The pattern:**
```bash
# Create a zone
doctl compute domain create example.com --ip-address 192.0.2.1

# Add records
doctl compute domain records create example.com \
  --record-type A --record-name www --record-data 192.0.2.1 --record-ttl 3600

doctl compute domain records create example.com \
  --record-type MX --record-name @ --record-data "aspmx.l.google.com." \
  --record-priority 1 --record-ttl 3600
```

**Why this is an improvement:**
- Scriptable and repeatable
- Enables simple automation in CI/CD pipelines
- Easier to manage multiple similar records (loop over a list)

**Why this still falls short:**
- **Stateless:** `doctl` has no concept of desired state. To update or delete a record, you must first list all records, grep for the one you want, extract its numeric `record-id`, and then issue a separate delete/update command
- **No drift detection:** If someone makes a manual change in the control panel, your script has no way to detect or correct it
- **Not declarative:** You're writing imperative commands ("create this, then delete that") rather than declaring your desired end state

**Verdict:** Useful for one-off migrations, simple DDNS (Dynamic DNS) scripts, or quick fixes. Not a foundation for production infrastructure management.

### Level 2: Configuration Management with Ansible (Stateless Orchestration)

Ansible provides a structured approach to DNS management via the `community.digitalocean` collection, which includes modules for domains and records.

**The pattern:**
```yaml
- name: Create DNS zone
  community.digitalocean.digital_ocean_domain:
    name: example.com
    state: present
    oauth_token: "{{ digitalocean_token }}"

- name: Create A record
  community.digitalocean.digital_ocean_domain_record:
    domain: example.com
    type: A
    name: "@"
    data: "192.0.2.1"
    ttl: 3600
    state: present
    oauth_token: "{{ digitalocean_token }}"
```

**Why this is better than scripting:**
- Idempotent by design (safe to run multiple times)
- Structured configuration in version-controlled YAML
- Integrates with existing Ansible infrastructure automation

**Why this still has limitations:**
- **No state file:** Ansible is stateless. It queries the DigitalOcean API on every run to check current state, making it slower and less efficient at drift detection than stateful IaC tools
- **Weak drift detection:** Since Ansible doesn't maintain a record of "known good state," detecting out-of-band changes (someone manually editing DNS in the control panel) is unreliable
- **Not infrastructure-focused:** Ansible excels at configuration management (installing packages, configuring services), but for cloud resource provisioning, stateful IaC tools are more appropriate

**Verdict:** Production-ready for teams already invested in Ansible for configuration management. Less ideal as a dedicated DNS provisioning solution.

### Level 3: Infrastructure as Code with Terraform (The Production Standard)

Terraform is the de facto standard for declarative cloud infrastructure provisioning, and the DigitalOcean provider offers robust support for DNS zones and records.

**The pattern:**
```hcl
resource "digitalocean_domain" "example" {
  name = "example.com"
}

resource "digitalocean_record" "www" {
  domain = digitalocean_domain.example.id
  type   = "A"
  name   = "www"
  value  = "192.0.2.1"
  ttl    = 3600
}

resource "digitalocean_record" "mx" {
  for_each = toset([
    "aspmx.l.google.com.",
    "alt1.aspmx.l.google.com.",
  ])
  domain   = digitalocean_domain.example.id
  type     = "MX"
  name     = "@"
  value    = each.value
  priority = each.value == "aspmx.l.google.com." ? 1 : 5
  ttl      = 3600
}
```

**Why this is the production choice:**
- **Stateful:** Terraform maintains a `terraform.tfstate` file that records the current known state of infrastructure. The `terraform plan` command performs drift detection by comparing the state file to the live API state
- **Declarative:** You declare the desired end state; Terraform calculates the minimal set of API calls to reach it
- **Resource dependencies:** The `digitalocean_record` resource explicitly references `digitalocean_domain.example.id`, creating a dependency graph that ensures records are only created after the domain exists
- **Multi-environment patterns:** Terraform Workspaces enable clean separation of dev/staging/prod environments with isolated state files
- **Remote state:** State files can be stored in remote backends like DigitalOcean Spaces (S3-compatible), enabling team collaboration and preventing state conflicts

**Production best practices:**
- Use `for_each` or `count` to manage repetitive records (e.g., Google Workspace MX records, multiple verification TXT records) from a data structure rather than duplicating code
- Store sensitive API tokens in environment variables or tools like HashiCorp Vault, not in plain text in `.tf` files
- Use remote state backends with locking to prevent concurrent modifications

**Verdict:** The gold standard for production DNS management. Mature, well-documented, and widely adopted.

### Level 4: Multi-Language IaC with Pulumi (Language Flexibility)

Pulumi offers the same declarative IaC model as Terraform but allows you to write infrastructure code in general-purpose programming languages (Python, Go, TypeScript, C#) instead of HCL.

**The pattern (Python):**
```python
import pulumi_digitalocean as digitalocean

zone = digitalocean.Domain("example", name="example.com")

www_record = digitalocean.DnsRecord(
    "www",
    domain=zone.id,
    type="A",
    name="www",
    value="192.0.2.1",
    ttl=3600
)

mx_servers = [
    ("aspmx.l.google.com.", 1),
    ("alt1.aspmx.l.google.com.", 5),
]

for mx_value, priority in mx_servers:
    digitalocean.DnsRecord(
        f"mx-{priority}",
        domain=zone.id,
        type="MX",
        name="@",
        value=mx_value,
        priority=priority,
        ttl=3600
    )
```

**Why Pulumi is compelling:**
- **Real programming languages:** Use standard control flow (loops, conditionals), existing libraries, and unit testing frameworks
- **Shared codebase:** Infrastructure logic can live alongside application code in the same language
- **Managed state by default:** Pulumi Service provides hosted state management and preview/diff capabilities out of the box
- **Terraform parity:** The Pulumi DigitalOcean provider is a wrapper around the DigitalOcean Terraform provider, ensuring 100% feature parity

**Why it's not universally adopted:**
- **Language lock-in:** Choosing Python/Go/TS for infrastructure means every team member must know that language
- **Ecosystem maturity:** While production-ready, Pulumi's ecosystem (modules, examples, community support) is smaller than Terraform's

**Verdict:** Production-ready and excellent for teams that prefer general-purpose languages over HCL. The choice between Terraform and Pulumi is primarily about language preference, not technical capability.

### Level 5: Kubernetes-Native IaC with Crossplane (Cluster as Control Plane)

Crossplane extends the Kubernetes API, allowing you to manage external cloud resources (like DNS zones) as Kubernetes Custom Resource Definitions (CRDs).

**The pattern:**
```yaml
apiVersion: dns.digitalocean.crossplane.io/v1alpha1
kind: Domain
metadata:
  name: example-com
spec:
  forProvider:
    name: example.com
  providerConfigRef:
    name: digitalocean-provider

---
apiVersion: dns.digitalocean.crossplane.io/v1alpha1
kind: Record
metadata:
  name: www-example-com
spec:
  forProvider:
    domain: example.com
    type: A
    name: www
    value: "192.0.2.1"
    ttl: 3600
  providerConfigRef:
    name: digitalocean-provider
```

**Why this is powerful for Kubernetes-native teams:**
- **Single control plane:** Manage DNS, databases, and application deployments all via `kubectl`
- **GitOps integration:** Tools like ArgoCD and Flux can automatically reconcile DNS state from Git
- **Upjet-based maturity:** The modern `provider-upjet-digitalocean` is code-generated from the DigitalOcean Terraform provider, ensuring feature parity and maintainability

**Why this is overkill for most teams:**
- **Complexity overhead:** Requires a running Kubernetes cluster just to manage DNS
- **Limited ecosystem:** Only makes sense if you're already deep in the Kubernetes ecosystem
- **Not a universal tool:** You can't use Crossplane to manage non-Kubernetes infrastructure without additional abstractions

**Verdict:** Production-ready but niche. Use this if you're building a Kubernetes-centric platform and want unified resource management. For most teams, Terraform or Pulumi is simpler.

## Comparative Analysis: Choosing Your Tool

| Feature | Terraform | Pulumi | Ansible | Crossplane |
|---------|-----------|--------|---------|------------|
| **Core Model** | IaC (Provisioning) | IaC (Provisioning) | CM (Configuration) | Kubernetes-Native IaC |
| **Language** | HCL (Declarative) | Python, Go, TS, etc. | YAML (Declarative) | Kubernetes YAML |
| **State Management** | State file (local or remote) | Pulumi Service or file | Stateless (queries API) | Kubernetes etcd |
| **Drift Detection** | Strong (`terraform plan`) | Strong (`pulumi preview`) | Weak (re-run based) | Strong (reconciliation loop) |
| **Zone Resource** | `digitalocean_domain` | `digitalocean.Domain` | `digital_ocean_domain` | `Domain.dns.digitalocean` |
| **Record Resource** | `digitalocean_record` | `digitalocean.DnsRecord` | `digital_ocean_domain_record` | `Record.dns.digitalocean` |
| **Team Collaboration** | Remote state with locking | Pulumi Service (default) | Ansible inventory | Kubernetes RBAC |
| **Learning Curve** | Moderate (HCL-specific) | Low (if you know the language) | Low (YAML-based) | High (requires Kubernetes) |

**Recommendation:**
- **For most production teams:** Terraform (mature, battle-tested, huge ecosystem)
- **For teams preferring real languages:** Pulumi (equal maturity, language flexibility)
- **For existing Ansible shops:** Ansible (acceptable, but consider migrating to Terraform/Pulumi for cloud resources)
- **For Kubernetes-native platforms:** Crossplane (powerful but complex)

## The DigitalOcean DNS Platform: Strengths and Constraints

### What DigitalOcean DNS Does Well

1. **Simplicity:** Clean API, straightforward control panel, no surprise charges
2. **Ecosystem integration:** Seamless integration with Droplets, Load Balancers, and Spaces for automatic SSL certificate management (Let's Encrypt)
3. **API-first design:** Every control panel action is available via the REST API, enabling robust automation
4. **Cost:** Free (currently) for all users

### Critical Limitations to Understand

1. **No DNSSEC:** DigitalOcean explicitly does not support DNSSEC. If your domain has DNSSEC enabled at the registrar and you delegate to DigitalOcean nameservers, DNS resolution will fail. The official recommendation for teams requiring DNSSEC is to use Cloudflare instead.

2. **No zone import:** There is no way to import a BIND-style zone file. Migrating an existing domain with many records means manually recreating every record via the API or control panel.

3. **No advanced routing:** DigitalOcean DNS does not support geo-routing, weighted routing, latency-based routing, or failover policies like AWS Route 53. It is a simple, authoritative DNS service.

4. **API rate limits:** The API is capped at 250 requests per minute. This is a documented constraint for automation tools like Kubernetes ExternalDNS, which can enter a crash-loop when managing hundreds of domains.

### When to Choose DigitalOcean DNS

Choose DigitalOcean DNS when:
- Your infrastructure is >90% hosted on DigitalOcean
- You value simplicity and cost predictability over advanced features
- You do not require DNSSEC, geo-routing, or a global Anycast network
- You're building a straightforward application with simple DNS requirements

### When to Choose Cloudflare or Route 53 Instead

**Use Cloudflare if:**
- DNSSEC is a hard requirement
- You need DDoS protection, WAF, or CDN capabilities
- You want the fastest global DNS resolution (330+ cities, Anycast)
- Your infrastructure spans multiple cloud providers

**Use AWS Route 53 if:**
- You need complex traffic routing (geo, latency, weighted, failover)
- You require private DNS zones for VPCs
- Your infrastructure is deeply integrated with AWS services

**The hybrid pattern:** Many teams use Cloudflare as the authoritative DNS provider (for security and performance) and simply create A/CNAME records pointing to DigitalOcean Droplets and Load Balancers. This gives you the best of both platforms.

## Production Best Practices

### 1. DNS Delegation Setup

To use DigitalOcean DNS, you must:
1. Purchase your domain from a third-party registrar (Namecheap, Google Domains, etc.)
2. Create the DNS zone in DigitalOcean (via control panel, API, or IaC)
3. Update the domain's nameserver (NS) records at the registrar to:
   - ns1.digitalocean.com
   - ns2.digitalocean.com
   - ns3.digitalocean.com

**Critical prerequisite:** Disable DNSSEC at the registrar before changing nameservers. Failing to do this will cause the domain to become unreachable.

### 2. TTL Strategy for Agility

The Time-To-Live (TTL) value controls how long DNS resolvers cache a record. Choose TTL values based on your change frequency:

- **Standard (3600s / 1 hour):** Safe default for most records
- **Migration/agile (300s / 5 minutes):** Use during cutover events to enable quick rollback
- **Static infrastructure (86400s / 24 hours):** Use for rarely-changing records like CDN CNAMEs, SPF/DMARC TXT records, or NS records

**Best practice during migrations:** Lower TTLs *before* making changes, perform the migration, verify success, then raise TTLs back to standard values.

### 3. Monitoring and Alerting

DigitalOcean's native Monitoring product only covers Droplet resources (CPU, RAM, disk), not DNS health.

**Recommended approach:**
- Use external DNS monitoring services like **UptimeRobot** to track specific record types (A, AAAA, CNAME, MX, TXT)
- Configure alerts for record value changes, additions, or deletions
- Use propagation checkers (e.g., DigitalOcean's DNS Lookup tool) for manual debugging

### 4. Backup and Disaster Recovery

DNS configuration is code. Your backup strategy depends on your management method:

**Manual/Control Panel users:** Use DigitalOcean's "Download zone" button to export a BIND-style zone file. Store this in version control. Note that restore is manual (no import function exists).

**IaC users (Terraform/Pulumi):** Your `.tf` or `.py` files *are* the backup. Store them in Git, and use a remote state backend (DigitalOcean Spaces, S3, Pulumi Service) for state files. Disaster recovery is as simple as running `terraform apply` or `pulumi up` in a new account.

### 5. Kubernetes Integration Patterns

**ExternalDNS:** Automatically synchronize DNS records based on Kubernetes Service and Ingress annotations. Requires careful configuration to avoid hitting the 250 req/min API rate limit:
- Use `--domain-filter` to limit scope
- Increase `--digitalocean-api-page-size` (default: 50)
- Set policy to `sync` to prevent modifying records it doesn't own

**cert-manager (DNS-01 challenges):** Automatically provision Let's Encrypt wildcard certificates by creating temporary `_acme-challenge` TXT records. Requires a Kubernetes secret containing your DigitalOcean API token.

## Planton's Approach: Balancing Flexibility and Simplicity

Planton's `DigitalOceanDnsZone` API is designed around the **80/20 principle**: expose the configuration that 80% of users need 80% of the time, while still enabling advanced use cases.

### Core Design Decisions

1. **Unified zone and records:** Unlike Terraform/Pulumi, which split zones and records into separate resources, Planton's API includes a `records` array directly in the `DigitalOceanDnsZoneSpec`. This simplifies the common case (creating a domain with its initial records in one atomic operation) while still supporting incremental record updates.

2. **Field-level simplicity:** Each record requires only the essentials:
   - `name` (relative to the zone, use "@" for apex)
   - `type` (A, AAAA, CNAME, MX, TXT, SRV, CAA, NS)
   - `values` (supports both literals and cross-resource references via `StringValueOrRef`)
   - `ttl_seconds` (defaults to 3600 if not specified)

3. **No artificial limitations:** While we provide sane defaults, advanced fields (priority for MX/SRV, flags/tag for CAA) are fully supported, ensuring the API can handle production use cases without requiring workarounds.

### Example Configurations

**Simple website (apex + www + email):**
```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanDnsZone
metadata:
  name: my-website
spec:
  domainName: example.com
  records:
    - name: "@"
      type: A
      values:
        - value: "192.0.2.1"  # Droplet IP
      ttl_seconds: 3600
    - name: "www"
      type: CNAME
      values:
        - value: "@"
      ttl_seconds: 3600
    - name: "@"
      type: MX
      values:
        - value: "aspmx.l.google.com."
        - value: "alt1.aspmx.l.google.com."
      ttl_seconds: 3600
    - name: "@"
      type: TXT
      values:
        - value: "v=spf1 include:_spf.google.com ~all"
      ttl_seconds: 3600
```

**Complex application (Load Balancer + Spaces CDN + verification):**
```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanDnsZone
metadata:
  name: my-app
spec:
  domainName: my-app.com
  records:
    - name: "@"
      type: A
      values:
        - value: "104.248.1.1"  # Load Balancer IP
      ttl_seconds: 300  # Low TTL during migration
    - name: "api"
      type: A
      values:
        - value: "104.248.1.1"
      ttl_seconds: 300
    - name: "assets"
      type: CNAME
      values:
        - value: "my-app-spaces.nyc3.cdn.digitaloceanspaces.com."
      ttl_seconds: 86400  # High TTL for static CDN
    - name: "_dmarc"
      type: TXT
      values:
        - value: "v=DMARC1; p=reject; rua=mailto:dmarc@my-app.com"
      ttl_seconds: 3600
    - name: "google-site-verification"
      type: TXT
      values:
        - value: "GV-abcd1234..."
      ttl_seconds: 3600
    - name: "@"
      type: CAA
      values:
        - value: "letsencrypt.org"
      ttl_seconds: 3600
```

## Conclusion: Strategic Simplicity

DigitalOcean DNS is not trying to be the most feature-rich DNS platform. It is deliberately simple, tightly integrated with the DigitalOcean ecosystem, and offered at no cost. This makes it the right choice for teams that value predictability, ease of use, and infrastructure consolidation.

However, simplicity is not a weakness—it's a strategic choice. By understanding the platform's constraints (no DNSSEC, no zone import, rate limits), you can make informed decisions about when to use DigitalOcean DNS and when to complement it with other providers like Cloudflare or Route 53.

For production DNS management, declarative Infrastructure as Code (Terraform, Pulumi) is the clear winner. These tools provide stateful management, drift detection, and team collaboration features that manual control panel clicking and stateless scripting cannot match.

Planton's `DigitalOceanDnsZone` API builds on these principles by providing a clean, protobuf-defined abstraction that balances simplicity (80/20 configuration) with flexibility (support for advanced record types and cross-resource references). Whether you're deploying a simple website or a complex multi-region application, the API gives you the tools to manage DNS as code, without the ceremony.

**Next steps:** Explore the `spec.proto` file to see the full API schema, and check out the example manifests in the repository for real-world configurations.


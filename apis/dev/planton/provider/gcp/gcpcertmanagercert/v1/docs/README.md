# GCP SSL/TLS Certificate Provisioning: Bridging Two Worlds

## Introduction

For years, the conventional wisdom around Google Cloud SSL certificates was simple: use the free "Google-managed SSL certificates" for your load balancers, point your DNS at the load balancer, and wait for the magic to happen. This approach worked—until it didn't.

The problems emerged at scale. Try to manage more than 15 certificates on a single load balancer? Hit a hard limit. Need a wildcard certificate for `*.example.com`? Not supported. Want to provision a certificate *before* your load balancer is live to avoid the dreaded "chicken-and-egg" validation problem? Impossible.

Enter **Google Certificate Manager**, a complete reimagining of how SSL/TLS certificates should be managed in the cloud. This newer service introduces flexible scoping (regional and global), DNS-based validation that decouples certificate provisioning from infrastructure, native wildcard support, and Certificate Maps—an abstraction layer that scales to thousands of certificates.

Yet the classic approach retains one compelling advantage: **it's free**. Certificate Manager charges $0.20 per certificate per month after the first 100 certificates.

This creates a strategic question: How do you design an automation framework that gives users the best of both worlds—the simplicity and cost savings of the classic approach when appropriate, and the power and flexibility of Certificate Manager when needed?

Planton's `GcpCertManagerCert` API answers this question by providing a **unified abstraction** over both services. Developers specify their requirements—domain names, DNS zone, desired certificate type—and the controller intelligently provisions the appropriate underlying resource. The API also automates the complex DNS validation workflow, transforming a multi-step orchestration into a single declarative specification.

This document explores the landscape of GCP certificate provisioning methods, compares the two services, and explains Planton's approach to making SSL/TLS certificate management simple, cost-effective, and production-ready.

## The Evolution: From Classic to Modern Certificate Management

Understanding the two certificate services requires understanding their architectural differences and the problems each was designed to solve.

### Level 0: The Classic Approach (Legacy, But Free)

Google's original certificate provisioning service is technically called **Google-managed SSL certificates**, accessed through the Compute Engine API as the `google_compute_managed_ssl_certificate` resource. In the Cloud Console, it appears as "Classic Certificates."

**How it works:**
1. Create a certificate resource specifying your domain(s)
2. Attach the certificate directly to a load balancer's target proxy
3. Point your domain's DNS to the load balancer's IP address
4. Google's CA validates ownership by probing the domain via the load balancer (HTTP-01 or TLS-ALPN-01 challenge)
5. Upon successful validation, the certificate is issued and automatically renewed

**The advantages:**
- **Cost: Free**—no charges for certificate provisioning or renewal
- Simple integration with Google Cloud Load Balancers
- Automatic renewal (90-day certificates, renewed automatically)

**The fundamental limitations:**
- **No wildcard support**—cannot issue certificates for `*.example.com`
- **Global scope only**—cannot create regional certificates for regional load balancers
- **Load balancer validation only**—certificate won't validate until DNS points to the load balancer
- **Scale limit**—maximum of 15 certificates per target proxy (direct attachment model)
- **Chicken-and-egg problem**—you need the load balancer live before the certificate validates, but you don't want to point users to an uncertified load balancer

**When to use it:**
For simple, non-wildcard, global load balancer deployments where cost is the primary concern. Think: a small blog with `example.com` and `www.example.com`, or a startup MVP where every dollar counts.

**The verdict:**
The classic approach is "good enough" for basic use cases but represents an architectural dead-end. The 15-certificate limit and lack of wildcard support make it unsuitable for any SaaS platform managing custom domains at scale.

### Level 1: The Modern Service (Certificate Manager)

**Google Certificate Manager** is a standalone service (`certificatemanager.googleapis.com`) introduced as the strategic, API-first platform for managing all TLS certificates in a GCP organization.

**How it works:**
1. Create a `DnsAuthorization` resource for each domain
2. The API returns a unique CNAME record (e.g., `_acme-challenge.example.com → xyz.authorize.certificatemanager.goog`)
3. Create the CNAME record in your Cloud DNS zone
4. Create a `Certificate` resource referencing the DNS authorization(s)
5. Google's CA validates ownership by checking for the CNAME record
6. Upon validation, the certificate becomes `ACTIVE` and can be used

**The architectural breakthroughs:**

1. **DNS-Based Validation**: Decouples certificate provisioning from infrastructure. You can fully provision a certificate to `ACTIVE` status *before your load balancer even exists*. This eliminates downtime and enables true zero-downtime migrations.

2. **Wildcard Support**: Native support for `*.example.com` wildcards, using a single DNS authorization for the apex domain.

3. **Flexible Scoping**: Supports global, regional, and `EDGE_CACHE` scopes. Regional certificates can be used with regional load balancers.

4. **Certificate Maps**: Instead of directly attaching certificates to load balancers (the classic model's bottleneck), Certificate Manager introduces `CertificateMap` resources. A map is a routing table: hostname → certificate. The load balancer references a single map, which can contain thousands of entries.

5. **Scalable Architecture**: The classic model supports 15 certificates per proxy. Certificate Manager direct attachment supports 100. Certificate Maps support thousands, shattering the scale ceiling for multi-tenant SaaS platforms.

**The trade-off:**
Certificate Manager is **paid** (free tier for first 100 certificates, then $0.20/cert/month). For large deployments, this cost is negligible compared to the operational complexity it eliminates. For small deployments, it's an unnecessary expense if the classic service meets requirements.

**When to use it:**
- When you need wildcard certificates
- When you need regional certificates
- When you need DNS-based validation for zero-downtime provisioning
- When managing certificates for thousands of domains (using Certificate Maps)
- When integrating with private CAs (Google Cloud CA Service)

**The verdict:**
Certificate Manager is the production-grade, at-scale solution. It's architecturally superior in every dimension except cost. Any modern infrastructure-as-code framework must support it as the default while retaining the option to fall back to the classic service for cost optimization.

### Comparative Analysis

The following table summarizes the key differences:

| Feature | Classic (Compute SSL) | Modern (Certificate Manager) |
|---------|----------------------|------------------------------|
| **Resource Type** | `google_compute_managed_ssl_certificate` | `google_certificate_manager_certificate` |
| **Cost** | **Free** | **Paid** ($0.20/cert/month after first 100) |
| **Wildcard Support** | ❌ No | ✅ Yes |
| **Validation Methods** | Load Balancer only | Load Balancer + DNS |
| **Scope** | Global only | Global, Regional, EDGE_CACHE |
| **Private PKI** | ❌ No | ✅ Yes (CA Service integration) |
| **Integration Model** | Direct proxy attachment | Certificate Maps (primary) or direct attachment |
| **Scale Limit** | 15 certs per proxy | 1 map with thousands of entries |
| **Zero-Downtime Migration** | Difficult (chicken-and-egg) | Native (DNS auth enables "make-before-break") |

## The Planton Choice: Intelligent, Unified Abstraction

Planton's `GcpCertManagerCert` API provides a **unified interface** over both services. The controller intelligently decides which underlying resource to provision based on user requirements.

### The Strategy: Best of Both Worlds

**Default Behavior:**
- If the user specifies `certificate_type: MANAGED` (the default), the controller provisions a `google_certificate_manager_certificate` with DNS authorization.
- If the user specifies `certificate_type: LOAD_BALANCER`, the controller provisions a `google_compute_managed_ssl_certificate` with load balancer authorization.

**Intelligent Provisioning Logic:**
The controller enforces compatibility constraints:
- **Wildcard certificates** (`*.example.com`) require `MANAGED` type (classic doesn't support wildcards)
- **Regional certificates** require `MANAGED` type (classic is global-only)
- **DNS validation** is only available with `MANAGED` type

If a user requests incompatible features (e.g., wildcard + `LOAD_BALANCER` type), the controller returns a validation error with a clear message explaining the constraint.

**Cost Optimization:**
For simple, non-wildcard, global certificates, users can explicitly choose `certificate_type: LOAD_BALANCER` to use the free classic service. For all other cases, the paid-but-powerful Certificate Manager is the default.

### The Automation "Magic": DNS Validation Orchestration

The most significant value-add of the Planton abstraction is **automated DNS validation**. Here's what the controller does behind the scenes:

**User provides (simplified API):**
```yaml
name: wildcard-cert
primary_domain_name: example.com
alternate_domain_names:
  - "*.example.com"
cloud_dns_zone_id: my-example-com-zone
certificate_type: MANAGED  # default
```

**Controller orchestrates (complex multi-step workflow):**
1. **Discover the Cloud DNS zone**: Look up the managed zone `my-example-com-zone` in GCP
2. **Create DNS Authorization**: Create a `google_certificate_manager_dns_authorization` resource for `example.com` (one auth covers both apex and wildcard)
3. **Extract CNAME data**: Read the `dns_resource_record` block from the authorization (e.g., `_acme-challenge.example.com → xyz.authorize.certificatemanager.goog`)
4. **Create DNS Record**: Create a `google_dns_record_set` in the Cloud DNS zone with the CNAME data
5. **Wait for propagation**: Ensure the DNS record exists before proceeding (explicit dependency)
6. **Create Certificate**: Create the `google_certificate_manager_certificate` resource referencing the DNS authorization
7. **Wait for ACTIVE**: Poll the certificate status until it reaches `ACTIVE` state

**What the user avoids:**
- Manually creating DNS authorization resources
- Looking up CNAME values from GCP API responses
- Creating DNS records in the correct zone
- Managing explicit dependencies between resources
- Dealing with race conditions (certificate creation before DNS record exists)

This transforms a 5-resource, order-dependent operation into a single declarative specification. The complexity is encapsulated; the user interface is simple.

### API Design: The 80/20 Principle

The `GcpCertManagerCertSpec` protobuf API focuses on the 20% of configuration that 80% of users need:

**Essential Fields (required or commonly set):**
- `gcp_project_id` — The GCP project
- `primary_domain_name` — The main domain (e.g., `example.com` or `*.example.com`)
- `alternate_domain_names` — Optional SANs for multi-domain certificates
- `cloud_dns_zone_id` — The Cloud DNS zone for automated validation
- `certificate_type` — `MANAGED` (default) or `LOAD_BALANCER`

**Advanced Fields (hidden in defaults):**
- `validation_method` — Defaults to `DNS` (only option currently supported)
- Location/scope — Defaults to `global` (can be extended for regional in the future)
- Labels, descriptions — Standard GCP metadata

**Example Configurations:**

**Simple (single domain):**
```yaml
gcp_project_id: my-project
primary_domain_name: app.example.com
cloud_dns_zone_id: example-com-zone
```

**Standard (wildcard):**
```yaml
gcp_project_id: my-project
primary_domain_name: example.com
alternate_domain_names:
  - "*.example.com"
cloud_dns_zone_id: example-com-zone
```

**Advanced (multi-domain SAN):**
```yaml
gcp_project_id: my-project
primary_domain_name: app.example.com
alternate_domain_names:
  - api.example.com
  - auth.example.com
cloud_dns_zone_id: example-com-zone
```

**Cost-optimized (classic, free):**
```yaml
gcp_project_id: my-project
primary_domain_name: www.example.com
cloud_dns_zone_id: example-com-zone
certificate_type: LOAD_BALANCER  # uses free classic service
```

## Implementation Landscape: How Certificates Are Provisioned

This section surveys the deployment methods across the industry, from manual workflows to fully automated IaC.

### Manual: Google Cloud Console

**Classic Certificates:**
Available under "Load Balancing → Certificates → Classic Certificates" (or within Certificate Manager UI). The workflow is a simple wizard:
1. Enter a certificate name
2. Enter domain names (comma-separated for SANs)
3. Attach to a load balancer target proxy

**Certificate Manager:**
A dedicated "Certificate Manager" section with step-by-step wizards:
1. Create DNS Authorization (provides CNAME to add to DNS)
2. Create Certificate (select DNS authorization)
3. Create Certificate Map (optional, for advanced routing)
4. Attach map to load balancer

**Verdict:** Useful for learning and one-off testing, but not viable for production-scale automation. The manual DNS record creation and attachment steps are error-prone.

### CLI: gcloud Commands

**Classic Certificates:**
```bash
gcloud compute ssl-certificates create my-cert \
  --domains=example.com,www.example.com \
  --global

gcloud compute target-https-proxies update my-proxy \
  --ssl-certificates=my-cert
```

**Certificate Manager:**
```bash
# Create DNS authorization
gcloud certificate-manager dns-authorizations create my-auth \
  --domain=example.com

# Get CNAME record to add to DNS (output must be manually parsed)
gcloud certificate-manager dns-authorizations describe my-auth

# (Manually create CNAME in Cloud DNS)

# Create certificate
gcloud certificate-manager certificates create my-cert \
  --domains=example.com,*.example.com \
  --dns-authorizations=my-auth

# Create certificate map and entry
gcloud certificate-manager maps create my-map
gcloud certificate-manager maps entries create my-entry \
  --map=my-map \
  --certificates=my-cert \
  --hostname=example.com

# Attach map to load balancer
gcloud compute target-https-proxies update my-proxy \
  --certificate-map=my-map
```

**Verdict:** More scriptable than the console, but still requires manual parsing of API outputs (CNAME data) and multi-step orchestration. Not declarative; hard to maintain desired state.

### IaC: Terraform/OpenTofu

Terraform provides first-class resources for both services.

**Classic:**
```hcl
resource "google_compute_managed_ssl_certificate" "classic" {
  name = "my-cert"
  managed {
    domains = ["example.com", "www.example.com"]
  }
}
```

**Certificate Manager (with DNS validation):**
```hcl
resource "google_certificate_manager_dns_authorization" "auth" {
  name   = "my-auth"
  domain = "example.com"
}

resource "google_dns_record_set" "auth_record" {
  name         = google_certificate_manager_dns_authorization.auth.dns_resource_record.0.name
  type         = google_certificate_manager_dns_authorization.auth.dns_resource_record.0.type
  ttl          = 300
  managed_zone = "my-dns-zone"
  rrdatas      = [google_certificate_manager_dns_authorization.auth.dns_resource_record.0.data]
}

resource "google_certificate_manager_certificate" "cert" {
  name = "my-cert"
  managed {
    domains = [
      google_certificate_manager_dns_authorization.auth.domain,
      "*.${google_certificate_manager_dns_authorization.auth.domain}"
    ]
    dns_authorizations = [google_certificate_manager_dns_authorization.auth.id]
  }
  depends_on = [google_dns_record_set.auth_record]
}
```

**Critical detail:** The `depends_on` is **mandatory**. Without it, Terraform might attempt to create the certificate *before* the DNS record exists, causing validation to fail. This is a common pitfall.

**Verdict:** Terraform is production-ready and widely used. The declarative model handles dependencies well. However, it still requires deep knowledge of the resource relationships and correct dependency ordering. The three-resource pattern (authorization → record → certificate) is verbose.

### IaC: Pulumi

Pulumi mirrors Terraform's resource model but in general-purpose programming languages (Python, TypeScript, Go, etc.).

**Example (Python):**
```python
import pulumi_gcp as gcp

auth = gcp.certificatemanager.DnsAuthorization("my-auth",
    domain="example.com")

record = gcp.dns.RecordSet("auth-record",
    name=auth.dns_resource_record.name,
    type=auth.dns_resource_record.type,
    ttl=300,
    managed_zone="my-dns-zone",
    rrdatas=[auth.dns_resource_record.data])

cert = gcp.certificatemanager.Certificate("my-cert",
    managed=gcp.certificatemanager.CertificateManagedArgs(
        domains=[auth.domain, f"*.{auth.domain}"],
        dns_authorizations=[auth.id]
    ),
    opts=pulumi.ResourceOptions(depends_on=[record]))
```

**Verdict:** Pulumi provides the same declarative benefits as Terraform with the added flexibility of a full programming language. The same three-resource pattern and dependency management are required.

### IaC: Ansible

Analysis of Ansible documentation reveals a gap: while modules exist for classic compute SSL certificates (`google.cloud.gcp_compute_ssl_certificate`), there are **no native modules** for the modern Certificate Manager API.

**Implication:** Ansible users are forced to either:
1. Use the classic (limited) service
2. Shell out to `gcloud` commands (breaking declarative idempotency)
3. Use raw HTTP API calls

**Verdict:** Ansible is **not recommended** for Certificate Manager automation. This represents a significant tooling gap that higher-level abstractions like Planton can fill.

### Cloud-Native: Crossplane

Crossplane's Upbound GCP provider includes resources for Certificate Manager:
- `certificatemanager.gcp.upbound.io/Certificate`
- `certificatemanager.gcp.upbound.io/CertificateMap`
- `certificatemanager.gcp.upbound.io/DnsAuthorization`

**Example (Kubernetes CRD):**
```yaml
apiVersion: certificatemanager.gcp.upbound.io/v1beta1
kind: Certificate
metadata:
  name: my-cert
spec:
  forProvider:
    managed:
      domains:
        - example.com
        - "*.example.com"
      dnsAuthorizations:
        - my-auth
```

**Verdict:** Crossplane enables Kubernetes-native certificate provisioning. It's well-suited for organizations already standardized on Kubernetes as a control plane. However, it still requires understanding the three-resource pattern and managing dependencies via Crossplane's reference model.

### Special Case: GKE Ingress and the "Wildcard Trap"

GKE provides a native Kubernetes CRD called `ManagedCertificate` (`networking.gke.io/v1`). Users create this object and reference it in an Ingress annotation:

```yaml
apiVersion: networking.gke.io/v1
kind: ManagedCertificate
metadata:
  name: my-cert
spec:
  domains:
    - example.com
    - www.example.com
```

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    networking.gke.io/managed-certificates: my-cert
spec:
  # ... ingress rules
```

**The trap:** This CRD is a wrapper around the **classic** compute SSL certificates service. It does **not** support wildcard domains. Attempting to use `*.example.com` will fail silently or with cryptic errors.

**The workaround:** To use wildcard Certificate Manager certificates with GKE Ingress, users must:
1. Manually create the Certificate Manager certificate and map via Terraform/gcloud
2. Create a *dummy* `ManagedCertificate` CRD (to force GKE to create the target proxy)
3. Use `gcloud compute target-https-proxies update` to manually swap the dummy certificate for the real Certificate Manager map

This is a painful, multi-step process that breaks Kubernetes-native workflows. It's a major opportunity for Planton to provide a better abstraction.

## Production Best Practices

### DNS Validation: Common Pitfalls

**FAILED_NOT_VISIBLE (Most Common Error):**
The certificate status shows `FAILED_NOT_VISIBLE`, meaning Google's multi-perspective validation system cannot find the required DNS record.

**Root causes:**
1. **Propagation delay**: The DNS record was just created and hasn't propagated globally yet. **Solution:** Wait 5-10 minutes and Google will retry automatically.
2. **Misconfiguration**: The CNAME record is incorrect (wrong name, wrong data, wrong zone). **Solution:** Use `dig _acme-challenge.example.com` from an external machine to verify what the public internet sees.
3. **GeoDNS/Split-Horizon DNS**: The DNS provider returns different responses based on the querier's location. **Solution:** Disable GeoDNS for validation records. Google's CAs query from multiple global locations and expect consistent responses.

**FAILED_CAA_FORBIDDEN:**
The domain has a DNS CAA (Certificate Authority Authorization) record that doesn't permit Google to issue certificates.

**Solution:** Add a CAA record permitting `pki.goog` (Google Trust Services):
```
example.com. CAA 0 issue "pki.goog"
```

### Lifecycle Management

**Renewal:**
All Google-managed certificates (both classic and Certificate Manager) are **automatically renewed** by Google. The certificates have a 90-day validity period, but Google renews them well before expiration. This is zero-touch; no user intervention is required.

**Monitoring:**
For **self-managed** certificates (uploaded to Certificate Manager), monitoring expiration is critical. Certificate Manager writes expiration events to Cloud Logging. Set up a logs-based metric and Cloud Monitoring alert:

```
logName = "projects/my-project/logs/certificatemanager.googleapis.com%2Fcertificates_expiry"
```

**Migration (Classic to Certificate Manager):**
The "make-before-break" pattern enabled by DNS authorization:
1. Domain is live with classic certificate attached to load balancer
2. Provision a *new* Certificate Manager certificate with DNS authorization (can happen while old cert is serving traffic)
3. Wait for new certificate to reach `ACTIVE` status
4. Create Certificate Map and Map Entry for new certificate
5. Update target proxy to use Certificate Map (atomic operation)
6. Delete old classic certificate

This approach achieves **zero downtime**.

### Integration with Load Balancers

**Certificate Maps (Recommended):**
For Certificate Manager, always use Certificate Maps rather than direct attachment. This provides:
- Scalability (thousands of certificates)
- Flexibility (update mappings without touching load balancer config)
- Granular control (different certificates for different hostnames on same load balancer)

**Attachment:**
```bash
gcloud compute target-https-proxies update my-proxy \
  --certificate-map=my-map \
  --clear-ssl-certificates  # removes any old classic certs
```

The `--clear-ssl-certificates` flag is critical when migrating from classic to Certificate Manager. A target proxy can have *either* a certificate list *or* a certificate map, but not both.

### Quota and Scale Considerations

**Classic:**
- Maximum 15 certificates per target proxy (hard limit)
- Maximum 100 domains per certificate (SANs)

**Certificate Manager:**
- Maximum 1 certificate map per target proxy
- A certificate map can contain thousands of entries
- Maximum 100 certificates via direct attachment (without map)
- API rate limits: 300 read/write requests per minute

**Recommendation:** For any deployment managing more than 10 certificates, standardize on Certificate Maps to future-proof against scale limits.

## Conclusion: Intelligent Abstraction for Modern Infrastructure

The landscape of GCP SSL/TLS certificate provisioning presents a choice: the free-but-limited classic service, or the powerful-but-paid Certificate Manager. For infrastructure automation, the answer is not either/or—it's both, intelligently.

Planton's `GcpCertManagerCert` API embodies this philosophy. By providing a unified interface that abstracts the complexity of DNS validation and intelligently selects the appropriate underlying service, it gives developers the best of both worlds: cost optimization when appropriate, and production-ready features when needed.

The shift from manual load balancer validation to automated DNS validation represents a paradigm change. DNS-based authorization decouples certificate lifecycle from infrastructure deployment, eliminating the chicken-and-egg problem and enabling true zero-downtime migrations. By automating the multi-step orchestration (authorization → DNS record → certificate), Planton transforms a complex, error-prone workflow into a simple, declarative specification.

Whether you're securing a single domain or managing thousands of certificates for a multi-tenant SaaS platform, the `GcpCertManagerCert` API provides a foundation for production-ready, scalable SSL/TLS management on Google Cloud.


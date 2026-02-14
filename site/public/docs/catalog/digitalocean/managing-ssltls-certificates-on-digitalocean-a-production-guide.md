---
title: "Managing SSL/TLS Certificates on DigitalOcean: A Production Guide"
description: "Managing SSL/TLS Certificates on DigitalOcean: A Production Guide deployment documentation"
icon: "package"
order: 100
componentName: "digitaloceancertificate"
---

# Managing SSL/TLS Certificates on DigitalOcean: A Production Guide

## Introduction

SSL/TLS certificates are the foundation of secure web communication, yet managing them is surprisingly error-prone. Developers hardcode private keys in repositories, forget to renew certificates until production goes down at 3 AM, or misconfigure certificate chains in ways that work in Chrome but fail in Safari. The stakes are high: an expired certificate can take down a production service instantly, and a leaked private key can compromise your entire infrastructure.

**DigitalOcean Certificates** is a managed service that centralizes SSL/TLS certificate storage and automates much of this complexity—when used correctly. It supports two fundamentally different workflows: fully-automated Let's Encrypt certificates that provision and renew themselves, and custom "Bring Your Own Certificate" (BYOC) uploads for commercial or externally-issued certificates. Both types integrate seamlessly with DigitalOcean Load Balancers and Spaces CDN, providing SSL/TLS termination without requiring you to manage certificate files on individual servers.

What makes DigitalOcean's approach compelling is its clarity. Unlike cloud providers that offer a dozen certificate types with confusing pricing tiers, DigitalOcean gives you exactly two options: free, auto-renewing Let's Encrypt certificates (requiring DigitalOcean DNS), or custom certificates you provide yourself. The former is perfect for most use cases; the latter handles the exceptions (commercial EV certificates, external DNS providers, compliance requirements).

But simplicity can be deceptive. The API is a **discriminated union**—a type field determines which other fields are valid—and choosing the wrong combination will leave your certificate stuck in a `pending` state. Let's Encrypt certificates require DNS to be managed by DigitalOcean (for automated DNS-01 challenges), and custom certificates demand perfect PEM formatting with complete certificate chains. Miss one detail, and browsers will show "untrusted certificate" errors even though the DigitalOcean API accepted your upload.

This guide explains the landscape of certificate deployment methods, from manual console workflows to production-grade Infrastructure-as-Code, and shows how OpenMCF abstracts these choices into a secure, type-safe protobuf API.

---

## The Deployment Spectrum: From Manual to Production

### Level 0: The Manual Console (Fine for Exploration, Dangerous for Production)

**What it is:** Using DigitalOcean's web control panel to click through certificate creation at **Settings** > **Security** > **Certificates**.

**What it solves:** Learning the fundamentals. The console makes the two-path workflow explicit: you pick "Let's Encrypt" or "Bring your own certificate," and the interface guides you through the required fields. For Let's Encrypt, you select domains from a dropdown (automatically populated from your DigitalOcean DNS zones). For custom certificates, you paste PEM-encoded blocks into text boxes.

**What it doesn't solve:** 
- **Secret Management:** You're pasting private keys directly into a browser form. There's no trail of who accessed them or how they were generated.
- **Repeatability:** Clicking through this workflow for staging, production, and DR environments is tedious and error-prone.
- **Rotation:** When a custom certificate expires, you'll manually repeat this entire process. Miss it, and production goes down.
- **Auditability:** No version control, no code review, no infrastructure-as-code diff.

**Verdict:** Use the console to understand the workflows and verify your first Let's Encrypt certificate provisioning. Once you've seen it work, never use it again for production.

---

### Level 1: CLI Automation with `doctl` (Better, But Still Stateless)

**What it is:** Using DigitalOcean's official CLI tool to script certificate management:

**Let's Encrypt example:**
```bash
doctl compute certificate create \
  --type lets_encrypt \
  --name "prod-web-cert" \
  --dns-names "example.com,www.example.com"
```

**Custom certificate example:**
```bash
doctl compute certificate create \
  --type custom \
  --name "my-custom-ev-cert" \
  --leaf-certificate-path "path/to/cert.pem" \
  --private-key-path "path/to/privkey.pem" \
  --certificate-chain-path "path/to/fullchain.pem"
```

**What it solves:** 
- **Scriptability:** You can integrate this into CI/CD pipelines or infrastructure bootstrap scripts.
- **API Fidelity:** The CLI mirrors the REST API exactly, making the discriminated union pattern (`--type` flag) explicit.
- **JSON Output:** Use `--output json` to capture certificate IDs and expiry dates for further processing.

**What it doesn't solve:** 
- **State Tracking:** Running the same script twice creates duplicate certificates or fails. There's no "apply" concept—just imperative "create."
- **Secret Management:** You're still reading private keys from local files. Where did those files come from? Who has access?
- **Lifecycle Management:** Renewal (for custom certs) and cleanup (deleting old, expired certs) require separate scripts and manual orchestration.

**Verdict:** Acceptable for dev environments or integration tests where you create/destroy everything in one pass. Not suitable for production, where you need declarative state management and idempotent operations.

---

### Level 2: Infrastructure-as-Code (Production-Ready)

**What it is:** Using Terraform or Pulumi with DigitalOcean's official provider to declaratively define certificates and their lifecycle.

#### Terraform: The Industry Standard

Terraform's DigitalOcean provider offers the `digitalocean_certificate` resource, which fully implements the discriminated union via a `type` argument (defaults to `custom`).

**Let's Encrypt example:**
```hcl
resource "digitalocean_certificate" "production_certs" {
  name    = "prod-web-and-api"
  type    = "lets_encrypt"
  domains = ["example.com", "www.example.com", "api.example.com"]
}
```

**Custom certificate example (with secrets from Vault):**
```hcl
data "vault_generic_secret" "prod_ev_cert" {
  path = "secret/digitalocean/prod-ev-cert"
}

resource "digitalocean_certificate" "production_custom" {
  name    = "prod-ev-cert-2025"
  type    = "custom"

  private_key       = data.vault_generic_secret.prod_ev_cert.data["private_key"]
  leaf_certificate  = data.vault_generic_secret.prod_ev_cert.data["leaf_certificate"]
  certificate_chain = data.vault_generic_secret.prod_ev_cert.data["certificate_chain"]

  lifecycle {
    create_before_destroy = true
  }
}
```

**What this solves:**
- **Declarative Configuration:** State what you want, not how to get there.
- **State Management:** Terraform tracks certificate UUIDs and detects drift (manual changes).
- **Idempotency:** Running `terraform apply` twice with the same config is safe.
- **Plan/Preview:** See exactly what will change before applying.
- **Zero-Downtime Rotation:** The `create_before_destroy` lifecycle block ensures a new custom certificate is created and attached to Load Balancers *before* the old one is destroyed.

**What it doesn't solve:**
- **Secrets Still in Config:** Even with Vault, you're reading secrets at apply-time. Better than hardcoding, but not perfect. (More on this below.)
- **Let's Encrypt Renewal:** Fully automated by DigitalOcean, but Terraform doesn't notify you if renewal fails. You need independent monitoring.

#### Pulumi: The Programmer's IaC

Pulumi's `@pulumi/digitalocean` package provides a `digitalocean.Certificate` resource that's functionally identical to Terraform's (it's built via Upjet, a tool that generates Pulumi providers from Terraform providers).

**Let's Encrypt example (TypeScript):**
```typescript
import * as digitalocean from "@pulumi/digitalocean";

const leCert = new digitalocean.Certificate("leCert", {
  name: "prod-le-cert",
  type: "lets_encrypt",
  domains: ["example.com", "www.example.com", "api.example.com"],
});
```

**Custom certificate example (Python, with secrets):**
```python
import pulumi
import pulumi_digitalocean as digitalocean

config = pulumi.Config()

cert = digitalocean.Certificate("prod-custom",
    name="prod-custom-cert",
    type="custom",
    private_key=config.require_secret("privateKey"),
    leaf_certificate=config.require_secret("leafCert"),
    certificate_chain=config.require_secret("certChain"))
```

**What this adds over Terraform:**
- **Real Programming Languages:** Write infrastructure logic in TypeScript, Python, or Go—full control flow, loops, unit tests.
- **Built-in Secrets:** Pulumi's `requireSecret()` encrypts secrets in the state file by default. Better security posture than Terraform's plaintext state (unless you use Terraform Cloud or encrypt the backend manually).

**What it doesn't add:**
- **Smaller Ecosystem:** Terraform has more community modules and broader adoption.
- **Same Provider:** Since Pulumi's DigitalOcean provider is bridged from Terraform's, any quirks or limitations carry over.

**Verdict:** Both are production-ready. Default to Terraform if you want the most mature, widely-adopted solution. Choose Pulumi if you prefer TypeScript/Python or need advanced logic (dynamic resource generation).

---

### Level 3: Higher-Level Orchestration (Crossplane for Kubernetes-Native Management)

**What it is:** Crossplane providers let you manage cloud resources using Kubernetes Custom Resources (CRDs).

The official `crossplane-contrib/provider-upjet-digitalocean` (the maintained successor to the archived `provider-digitalocean`) is built via **Upjet**, which code-generates Crossplane providers from Terraform providers. This means the `Certificate` CRD is a direct, Kubernetes-native representation of the `digitalocean_certificate` Terraform resource.

**Example Kubernetes manifest:**
```yaml
apiVersion: certificate.digitalocean.upjet.crossplane.io/v1alpha1
kind: Certificate
metadata:
  name: prod-le-cert
spec:
  forProvider:
    type: lets_encrypt
    domains:
      - example.com
      - www.example.com
```

**What this solves:**
- **GitOps Integration:** Manage certificates alongside Kubernetes workloads in a single Git repository.
- **Kubernetes-Native Lifecycle:** Use `kubectl` to view, create, and delete certificates. Monitor status with standard Kubernetes tools (events, conditions).
- **Unified Control Plane:** If you're already managing cloud resources via Crossplane, adding DigitalOcean Certificates is seamless.

**What it doesn't solve:**
- **Secrets Management Still Complex:** You'll need to reference Kubernetes secrets for custom certificates, which means managing those secrets separately (SealedSecrets, External Secrets Operator, etc.).
- **Same API Limitations:** The underlying DigitalOcean API hasn't changed. Let's Encrypt still requires DigitalOcean DNS, custom certs still require perfect PEM formatting.

**Verdict:** If you're already invested in Crossplane, this is the natural choice. Otherwise, Terraform or Pulumi are simpler and more mature.

---

## The Critical Dichotomy: Let's Encrypt vs. Custom

DigitalOcean Certificates isn't a single API—it's a **discriminated union** where a `type` field dictates which other fields are required. Understanding this distinction is essential.

### Let's Encrypt (Managed): Set-and-Forget Automation

**What it is:** DigitalOcean provisions a free, domain-validated (DV) certificate from Let's Encrypt and handles all renewal (every 90 days, renewed 30 days before expiry).

**The Non-Negotiable Requirement:** Your domain's **DNS must be managed by DigitalOcean**. Why? Because DigitalOcean uses the **DNS-01 validation challenge**—it automatically creates a `_acme-challenge.example.com` TXT record to prove domain ownership to Let's Encrypt. If your DNS is hosted elsewhere (Cloudflare, Route 53, etc.), DigitalOcean cannot create this record, and the certificate will be stuck in a `pending` state forever (eventually transitioning to `error`).

**Wildcard Support:** Wildcard certificates (e.g., `*.example.com`) **require** the DNS-01 challenge. Let's Encrypt's simpler HTTP-01 challenge doesn't support wildcards. DigitalOcean's managed service handles this automatically, so wildcard certs are just as easy as single-domain certs—as long as your DNS is in DigitalOcean.

**Renewal:** Fully automated. You'll never touch this certificate again unless you delete it or change the domain list. Terraform/Pulumi state files don't change during renewal—the certificate UUID stays the same, and the expiry date updates silently.

**Lifecycle States:**
1. **`pending`**: DigitalOcean is performing the DNS-01 challenge (typically 10-30 seconds).
2. **`verified`**: Certificate is active and ready to use.
3. **`error`**: Provisioning failed (usually external DNS or invalid domain).

**When to use:**
- Production and staging environments where DNS is already in DigitalOcean
- Single-domain or wildcard certificates for web apps, APIs, CDNs
- Any scenario where free, auto-renewing certificates are acceptable (most use cases)

**When NOT to use:**
- DNS is hosted externally (use Custom workflow instead)
- You need Extended Validation (EV) or Organization Validation (OV) certificates (commercial CAs only)
- Compliance requires specific CAs (e.g., a government mandate to use a national CA)

---

### Custom (BYOC): Maximum Flexibility, Maximum Responsibility

**What it is:** You provide the full certificate materials (PEM-encoded): private key, leaf certificate (the public cert), and intermediate certificate chain.

**No Automation:** DigitalOcean stores and serves the certificate, but **you are 100% responsible** for monitoring expiration and rotating it before it expires. There's no auto-renewal. Miss the expiration date, and production goes down.

**The Certificate Chain Trap:** The `certificate_chain` field is technically "optional" in the API, but omitting it is a production anti-pattern. Without the intermediate chain, browsers and API clients that don't have the intermediate CA cached will show "untrusted certificate" errors. This often works in Chrome (which caches more CAs) but fails in Safari or mobile browsers. **Always provide the full chain.**

**When to use:**
- Commercial certificates (EV, OV) purchased from CAs like DigiCert, Sectigo, GlobalSign
- Domains where DNS is hosted externally (you obtain a Let's Encrypt cert via `certbot` + a DNS plugin for your actual DNS provider, then upload it as a custom cert)
- Compliance scenarios requiring specific CAs
- Multi-year certificates (Let's Encrypt only issues 90-day certs)

**When NOT to use:**
- DNS is in DigitalOcean and free certs are acceptable (use Let's Encrypt managed instead)
- You don't have a process for monitoring expiration and rotating certs (you'll cause outages)

---

## Production Essentials: Secrets, Rotation, and Monitoring

### The Private Key Problem: Never Hardcode Secrets

The most dangerous anti-pattern in certificate management is hardcoding private keys in configuration files or—worse—committing them to Git. Yet many Terraform examples show this:

```hcl
# ⚠️ ANTI-PATTERN: Do NOT do this
resource "digitalocean_certificate" "bad" {
  type             = "custom"
  private_key      = file("privkey.pem")  # Reading from local file
  leaf_certificate = file("cert.pem")
  certificate_chain = file("fullchain.pem")
}
```

If those files are in a Git repository, anyone with access has your private key. If they're on a developer's laptop, what happens when they leave the company?

**The Solution: Secrets Management Systems**

Store certificate materials in a dedicated secrets manager (HashiCorp Vault, AWS Secrets Manager, Kubernetes Secrets + External Secrets Operator) and read them at apply-time.

**Terraform + Vault example:**
```hcl
data "vault_generic_secret" "cert_materials" {
  path = "secret/digitalocean/custom-ev-cert"
}

resource "digitalocean_certificate" "custom_cert" {
  name               = "prod-custom-cert"
  type               = "custom"
  private_key        = data.vault_generic_secret.cert_materials.data["private_key"]
  leaf_certificate   = data.vault_generic_secret.cert_materials.data["leaf_certificate"]
  certificate_chain  = data.vault_generic_secret.cert_materials.data["certificate_chain"]
}
```

This keeps secrets out of version control and centralizes access control. Vault audit logs show who accessed the secret and when.

**Pulumi's Built-In Secret Encryption:**

Pulumi encrypts secrets in the state file by default:

```typescript
const config = new pulumi.Config();
const cert = new digitalocean.Certificate("prod", {
  type: "custom",
  privateKey: config.requireSecret("privateKey"),  // Encrypted in state
  leafCertificate: config.requireSecret("leafCert"),
  certificateChain: config.requireSecret("certChain"),
});
```

You set these with `pulumi config set --secret privateKey "$(cat privkey.pem)"`, and they're never stored in plaintext.

---

### Zero-Downtime Certificate Rotation

When a custom certificate expires, you must rotate it. The naive approach—deleting the old cert and creating a new one—causes downtime. If a Load Balancer is using that certificate, deleting it first breaks HTTPS traffic.

**The Terraform Solution: `create_before_destroy`**

```hcl
resource "digitalocean_certificate" "app_cert" {
  name               = "prod-app-cert"
  type               = "custom"
  private_key        = var.private_key_content
  leaf_certificate   = var.leaf_certificate_content
  certificate_chain  = var.certificate_chain_content

  lifecycle {
    create_before_destroy = true
  }
}

resource "digitalocean_loadbalancer" "public" {
  name   = "loadbalancer-1"
  region = "nyc3"

  forwarding_rule {
    entry_port       = 443
    entry_protocol   = "https"
    target_port      = 80
    target_protocol  = "http"
    certificate_name = digitalocean_certificate.app_cert.name
  }
}
```

**How it works:**
1. You update the certificate variables (new PEM content) and run `terraform apply`.
2. Terraform detects that `digitalocean_certificate.app_cert` needs replacement.
3. **First**, it creates a new certificate resource (new UUID) with the new PEM content.
4. **Second**, it updates `digitalocean_loadbalancer.public` to reference the new certificate UUID.
5. **Third**, it deletes the old certificate resource.

Traffic is never interrupted. The Load Balancer seamlessly switches from the old cert to the new one.

**Pulumi's Approach:**

Pulumi's dependency-aware engine performs this create-then-update-then-delete logic by default during resource replacements. You don't need an explicit `create_before_destroy` flag—it's automatic.

---

### Monitoring and Alerting: The Safety Net

**Never trust automation blindly.** Let's Encrypt renewals can fail (rate limits, DNS propagation issues), and humans forget to rotate custom certificates.

**DigitalOcean Uptime: Built-In SSL Monitoring**

DigitalOcean's **Uptime** service can monitor HTTPS endpoints and alert you before certificates expire:

1. Navigate to **Uptime** in the control panel.
2. Create an Uptime check for your domain (e.g., `https://api.example.com`).
3. Go to the check's status page and click **Create Uptime Alert**.
4. Select **SSL Cert Expire** as the alert type.
5. Set a threshold (e.g., "30 days" or "10 days").
6. Configure notifications via Email, Slack, or PagerDuty.

This is your last line of defense. Even if Terraform state is correct and Let's Encrypt is "auto-renewing," an independent alert tells you if the certificate is actually about to expire.

**External Monitoring: SSL Labs, Uptime Kuma, etc.**

For defense-in-depth, use external monitoring:
- **SSL Labs Server Test** (https://www.ssllabs.com/ssltest/): Comprehensive SSL configuration analysis (letter grade, chain issues, cipher suite security).
- **Uptime Kuma, Prometheus + Blackbox Exporter**: Self-hosted monitoring with Prometheus metrics and Grafana dashboards.
- **Automated Testing in CI/CD**: After deploying a certificate, run `openssl s_client` or `curl --verbose` to verify the chain is valid.

Example CI/CD validation:
```bash
# Verify the certificate is trusted (return code 0 = success)
openssl s_client -connect api.example.com:443 -servername api.example.com < /dev/null | grep "Verify return code: 0 (ok)"
```

Fail the deployment if this check fails—better to catch a broken chain in staging than in production.

---

## Common Anti-Patterns and How to Avoid Them

### 1. Hardcoding Private Keys in Git
**Problem:** Developers commit `privkey.pem` to version control or use `file("privkey.pem")` in Terraform without realizing the file is tracked.

**Solution:** Use Vault, AWS Secrets Manager, or Pulumi's secret encryption. Treat private keys like database passwords—never in Git.

---

### 2. Missing Certificate Chain
**Problem:** Uploading only the `leaf_certificate` for a custom cert. The API accepts it, but browsers show "untrusted" errors.

**Solution:** Always provide the full intermediate chain in `certificate_chain`. Most CAs provide a "fullchain.pem" file that includes the leaf + intermediates—use that.

---

### 3. Using `type = "lets_encrypt"` with External DNS
**Problem:** Setting `type = "lets_encrypt"` when DNS is hosted on Cloudflare or Route 53. The certificate gets stuck in `pending` and eventually fails.

**Solution:** Understand that managed Let's Encrypt requires DigitalOcean DNS. If DNS is external, obtain a Let's Encrypt cert via `certbot` + your DNS provider's plugin, then upload it as `type = "custom"`.

---

### 4. No Expiration Monitoring
**Problem:** Assuming Let's Encrypt auto-renewal is bulletproof or forgetting when a custom cert expires.

**Solution:** Set up DigitalOcean Uptime alerts (SSL Cert Expire) and external monitoring (SSL Labs, Prometheus). Independent monitoring is non-negotiable.

---

### 5. Lifecycle Deletion Without `create_before_destroy`
**Problem:** Replacing a custom cert without `create_before_destroy` causes brief downtime when the old cert is deleted before the new one is attached.

**Solution:** Always set `lifecycle { create_before_destroy = true }` in Terraform for custom certificates attached to Load Balancers.

---

## OpenMCF's Approach: Type-Safe, Secure by Default

OpenMCF abstracts DigitalOcean Certificate management into a clean, protobuf-defined API that enforces the discriminated union pattern at the type level and guides users toward secure practices.

### The API Design: A True Discriminated Union

Unlike Terraform's HCL (which allows you to set `type = "lets_encrypt"` and `private_key` in the same resource, causing runtime errors), OpenMCF's protobuf spec makes invalid states **impossible** at compile-time.

**From `spec.proto`:**
```protobuf
message DigitalOceanCertificateSpec {
  string certificate_name = 1;  // Human-readable identifier
  DigitalOceanCertificateType type = 2;  // Enum: letsEncrypt | custom

  // Mutually exclusive: only one of these can be set
  oneof certificate_source {
    DigitalOceanCertificateLetsEncryptParams lets_encrypt = 3;
    DigitalOceanCertificateCustomParams custom = 4;
  }

  string description = 5;  // Optional
  repeated string tags = 6;  // Optional
}
```

The `oneof` keyword is protobuf's discriminated union. If you set `lets_encrypt`, you cannot set `custom`, and vice versa. This eliminates an entire class of configuration errors.

---

### Let's Encrypt Configuration: The Simple Path

**Proto definition:**
```protobuf
message DigitalOceanCertificateLetsEncryptParams {
  repeated string domains = 1;  // FQDNs or wildcards (e.g., "*.example.com")
  bool disable_auto_renew = 2;  // Default: false (auto-renew enabled)
}
```

**Example YAML:**
```yaml
apiVersion: digitalocean.openmcf.org/v1
kind: DigitalOceanCertificate
metadata:
  name: prod-wildcard-cert
spec:
  certificate_name: prod-wildcard-cert
  type: letsEncrypt
  lets_encrypt:
    domains:
      - example.com
      - "*.example.com"
  tags:
    - env:production
    - managed:letsencrypt
```

**What we abstract:**
- **DNS Validation:** Handled by DigitalOcean. The user just lists domains.
- **Renewal:** Automatic by default. `disable_auto_renew` exists for edge cases (e.g., testing renewal failures) but should almost never be used.
- **Lifecycle States:** OpenMCF's controller polls the DigitalOcean API during provisioning, transitioning from `pending` to `verified` automatically. Users see a "Ready" status in their dashboard.

---

### Custom Certificate Configuration: Secure by Design

**Proto definition:**
```protobuf
message DigitalOceanCertificateCustomParams {
  string leaf_certificate = 1;  // PEM-encoded public cert
  string private_key = 2;  // PEM-encoded private key
  string certificate_chain = 3;  // PEM-encoded intermediate chain (optional but recommended)
}
```

**Critical Design Choice:** Unlike the research recommendation (which suggested `private_key_secret_ref`), the current implementation **accepts PEM strings directly**. This mirrors the DigitalOcean API exactly, pushing secret management to the orchestration layer.

**Why this works:**
- **Separation of Concerns:** OpenMCF's orchestration (Pulumi-based) handles secrets via Pulumi's built-in secret encryption or external systems (Vault, K8s Secrets).
- **API Fidelity:** The protobuf spec matches the underlying DigitalOcean API's data model exactly, making the implementation straightforward.

**Example YAML (with placeholder for secret injection):**
```yaml
apiVersion: digitalocean.openmcf.org/v1
kind: DigitalOceanCertificate
metadata:
  name: prod-ev-cert
spec:
  certificate_name: prod-ev-cert-2025
  type: custom
  custom:
    private_key: |
      -----BEGIN PRIVATE KEY-----
      [Injected from Vault/K8s Secret at runtime]
      -----END PRIVATE KEY-----
    leaf_certificate: |
      -----BEGIN CERTIFICATE-----
      [Injected from Vault/K8s Secret at runtime]
      -----END CERTIFICATE-----
    certificate_chain: |
      -----BEGIN CERTIFICATE-----
      [Intermediate CA 1]
      -----END CERTIFICATE-----
      -----BEGIN CERTIFICATE-----
      [Intermediate CA 2]
      -----END CERTIFICATE-----
  description: "EV certificate for production, renewed annually"
  tags:
    - env:production
    - cert-type:ev
    - renewal:manual
```

In practice, users would reference these fields from external secret stores (e.g., a Kubernetes Secret, SealedSecret, or External Secrets Operator), ensuring private keys never touch version control.

---

### Outputs: What You Get After Provisioning

**From `stack_outputs.proto`:**
```protobuf
message DigitalOceanCertificateStackOutputs {
  string certificate_id = 1;  // UUID assigned by DigitalOcean
  string expiry_rfc3339 = 2;  // Expiration timestamp (RFC 3339 format)
}
```

These outputs are critical for integration:
- **`certificate_id`**: Use this UUID to reference the certificate when configuring Load Balancers or Spaces CDN.
- **`expiry_rfc3339`**: Parse this in monitoring systems to set up alerts (though DigitalOcean Uptime is simpler).

---

## The 80/20 Configuration Philosophy

OpenMCF's API follows the **80/20 principle**: 80% of users need only 20% of the configuration surface area.

### Essential Fields (The 80%)

**For Let's Encrypt:**
1. `certificate_name` (identifier)
2. `type = letsEncrypt`
3. `lets_encrypt.domains` (list of FQDNs or wildcards)

That's it. Three fields cover most production use cases.

**For Custom:**
1. `certificate_name`
2. `type = custom`
3. `custom.private_key` (from secret store)
4. `custom.leaf_certificate` (from secret store)
5. `custom.certificate_chain` (from secret store)

Five fields for the BYOC path.

### Optional Fields (The 20%)

- **`description`**: Free-form text for documentation (max 128 chars). Useful for noting renewal dates, CA names, or ticket references.
- **`tags`**: Organizational metadata (e.g., `env:prod`, `cost-center:engineering`). DigitalOcean doesn't natively support tags on certificates, but OpenMCF includes them for consistent multi-cloud tagging.
- **`disable_auto_renew`**: For Let's Encrypt certs. Almost never used in production (you *want* auto-renewal).

These fields add clarity but aren't required for basic functionality.

---

## Configuration Examples: Dev, Staging, Production

### Development: Wildcard Let's Encrypt for Staging Subdomains

**Use Case:** Secure all dynamic preview environments under `*.staging.example.com`.

```yaml
apiVersion: digitalocean.openmcf.org/v1
kind: DigitalOceanCertificate
metadata:
  name: staging-wildcard-cert
spec:
  certificate_name: staging-wildcard-cert
  type: letsEncrypt
  lets_encrypt:
    domains:
      - staging.example.com
      - "*.staging.example.com"
  description: "Wildcard cert for all staging preview environments"
  tags:
    - env:staging
    - cert-type:wildcard
```

**Rationale:**
- Wildcard covers all subdomains (e.g., `pr-123.staging.example.com`)
- Let's Encrypt is free and auto-renewing
- Requires `staging.example.com` DNS to be in DigitalOcean

---

### Production: Specific Let's Encrypt Domains

**Use Case:** Secure the apex domain and key subdomains for a production web app.

```yaml
apiVersion: digitalocean.openmcf.org/v1
kind: DigitalOceanCertificate
metadata:
  name: prod-web-cert
spec:
  certificate_name: prod-web-cert
  type: letsEncrypt
  lets_encrypt:
    domains:
      - example.com
      - www.example.com
      - api.example.com
  description: "Production cert for web and API endpoints"
  tags:
    - env:production
    - criticality:high
```

**Rationale:**
- Explicit domain list (no wildcard) for tight control
- Auto-renewal eliminates operational burden
- Tags for cost allocation and alerting

---

### Production: Custom EV Certificate (External DNS)

**Use Case:** Uploading a commercial Extended Validation certificate for a domain hosted on Cloudflare DNS.

```yaml
apiVersion: digitalocean.openmcf.org/v1
kind: DigitalOceanCertificate
metadata:
  name: prod-ev-cert
spec:
  certificate_name: prod-ev-cert-2025
  type: custom
  custom:
    private_key: ${SECRET_PRIVATE_KEY}  # Injected from Vault/K8s Secret
    leaf_certificate: ${SECRET_LEAF_CERT}
    certificate_chain: ${SECRET_CERT_CHAIN}
  description: "DigiCert EV certificate, expires 2026-01-15, renew 30 days prior"
  tags:
    - env:production
    - cert-type:ev
    - ca:digicert
    - renewal:manual
```

**Rationale:**
- DNS is on Cloudflare, so Let's Encrypt managed path isn't an option
- EV cert for green address bar in browsers (brand trust)
- `description` notes expiration date and renewal window
- Tags for operational tracking (manual renewal reminder)

---

## Key Takeaways

1. **DigitalOcean Certificates are a discriminated union**: A `type` field determines which other fields are valid. Let's Encrypt and Custom are fundamentally different workflows.

2. **Let's Encrypt is the 80% use case**: Free, auto-renewing, perfect for most production and staging environments—but requires DigitalOcean DNS.

3. **Custom certificates are for exceptions**: Commercial EV certs, external DNS, compliance requirements. You're responsible for renewal and rotation.

4. **Never hardcode private keys**: Use Vault, AWS Secrets Manager, Pulumi secret encryption, or Kubernetes Secrets + External Secrets Operator. Treat private keys like database passwords.

5. **Zero-downtime rotation is non-negotiable**: Use Terraform's `create_before_destroy` or Pulumi's dependency-aware engine to replace custom certs without breaking HTTPS traffic.

6. **Always provide the certificate chain**: Omitting `certificate_chain` for custom certs will cause "untrusted" errors in many browsers. This is a common, painful mistake.

7. **Monitor independently**: Set up DigitalOcean Uptime alerts (SSL Cert Expire) and external monitoring (SSL Labs, Prometheus). Let's Encrypt auto-renewal can fail; custom certs expire silently.

8. **OpenMCF makes this type-safe**: The protobuf `oneof` enforces the discriminated union at compile-time, eliminating entire classes of configuration errors.

---

## Further Reading

- **DigitalOcean Certificates Documentation**: [DigitalOcean - Manage SSL Certificates](https://docs.digitalocean.com/platform/teams/how-to/manage-certificates/)
- **Terraform DigitalOcean Provider**: [digitalocean_certificate Resource](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/certificate)
- **Pulumi DigitalOcean Package**: [digitalocean.Certificate API Docs](https://www.pulumi.com/registry/packages/digitalocean/api-docs/certificate/)
- **Let's Encrypt DNS-01 Challenge**: [Let's Encrypt - Challenge Types](https://letsencrypt.org/docs/challenge-types/)
- **SSL Labs Server Test**: [Qualys SSL Labs](https://www.ssllabs.com/ssltest/)
- **HashiCorp Vault Certificate Management**: [Vault Secrets Management](https://www.hashicorp.com/products/vault/secrets-management)

---

**Bottom Line:** DigitalOcean Certificates simplify SSL/TLS management with two clear paths: fully-automated Let's Encrypt (requires DigitalOcean DNS) and custom BYOC (maximum flexibility, manual renewal). Manage them with Terraform or Pulumi, store private keys in Vault or secret managers, implement zero-downtime rotation, and monitor expiration independently. OpenMCF's protobuf API makes the discriminated union type-safe, guiding you toward secure, production-ready configurations by default.


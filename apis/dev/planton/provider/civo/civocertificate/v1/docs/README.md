# TLS Certificate Management on Civo Cloud

## The Modern Certificate Landscape

HTTPS isn't optional anymore—it's table stakes. Yet for years, managing TLS certificates on cloud infrastructure remained surprisingly manual: renew before they expire, juggle private keys, chase down intermediate chains, and pray you don't misconfigure production at 2 AM.

The emergence of Let's Encrypt in 2015 fundamentally changed this game. Free, automated certificates with 90-day lifespans forced platforms to rethink certificate lifecycle management. Cloud providers responded by building managed certificate services that handle the ACME dance for you.

Civo Cloud's certificate management follows this modern pattern: **automated Let's Encrypt integration for zero-touch renewals, plus support for custom certificates when you need them**. This document explores how to deploy and manage certificates on Civo, from manual dashboard clicks to fully automated infrastructure-as-code.

## The Maturity Spectrum

### Level 0: No TLS (The Anti-Pattern)

Running production services over HTTP in 2025 is inexcusable. Browsers flag these as "Not Secure," users abandon transactions, and you're leaking credentials in plaintext. Modern browsers increasingly block features (geolocation, camera access, service workers) on non-HTTPS origins.

**Why it happens:** "We'll add HTTPS later" turns into "we never got around to it." Certificate management *seems* complex.

**Why it's wrong:** With Let's Encrypt automation, deploying HTTPS is often easier than configuring custom domain routing. There's no valid excuse.

### Level 1: Manual Dashboard Management

**What it looks like:** Logging into Civo's web console, clicking through certificate creation wizards, uploading PEM files, manually attaching certs to load balancers.

**When it's acceptable:** Quick prototypes, learning the platform, or truly one-off certificates that'll never be touched again.

**Pitfalls:**
- **Domain validation failures:** Let's Encrypt requires DNS pointing to Civo infrastructure before issuance. If your domain's A records aren't configured or you're using external DNS without proper setup, validation fails silently.
- **Incomplete certificate chains:** Uploading only the leaf certificate without intermediates causes trust errors in some browsers and API clients.
- **Forgotten renewals:** Custom certificates don't auto-renew. That cert you uploaded 364 days ago? It expires tomorrow, and you're on vacation.

**Verdict:** Fine for experiments, terrible for production. Manual processes scale linearly with team size and service count—meaning they don't scale at all.

### Level 2: CLI and Scripting

**What it looks like:** Using the `civo` CLI or `curl` scripts to call Civo's REST API, authenticating with `CIVO_TOKEN` environment variables, scripting certificate creation and attachment in bash or Python.

**Improvement over Level 1:** Repeatability. You can version control your scripts and re-run certificate creation deterministically.

**Limitations:**
- **State management:** Scripts don't remember what they created. Did you already create this certificate? Is it attached to the right load balancer? You'll need to query the API to find out.
- **Secret handling:** API keys and private keys end up in environment variables, CI logs, or (horrifyingly) committed to Git.
- **Error recovery:** If the script fails halfway through—certificate created, but load balancer attachment failed—you're left cleaning up manually.

**Verdict:** A step up for automation-minded teams, but not production-grade infrastructure management.

### Level 3: Infrastructure as Code (The Production Standard)

**What it looks like:** Terraform, Pulumi, or Crossplane definitions that declare certificates as resources, manage dependencies (certificate → load balancer attachment), and handle state tracking.

**Why it's the baseline for production:**
- **Declarative state:** "This certificate should exist with these properties" vs. "run this script to maybe create something."
- **Dependency graphs:** IaC tools understand that you can't attach a certificate to a load balancer until the certificate exists. They orchestrate the creation order automatically.
- **Secret management integration:** Pull private keys from HashiCorp Vault, AWS Secrets Manager, or Pulumi's secret system instead of hardcoding them.
- **Audit trails:** Every change is a Git commit. Who removed that cert? `git blame` knows.

**Trade-offs:**
- **Learning curve:** Teams new to Terraform/Pulumi face a steeper onboarding than clicking buttons.
- **State storage concerns:** Terraform state files contain sensitive data. Use encrypted remote backends (S3 with KMS, Terraform Cloud, etc.).

**Verdict:** This is the standard. If you're running production workloads on Civo, your certificates should be managed as code.

## Let's Encrypt vs Custom Certificates: The Decision Matrix

Civo supports two certificate types. Here's when to use each.

### Let's Encrypt (Managed)

**Use when:**
- You need publicly trusted certificates for standard domains
- You want zero-touch renewals every 90 days
- You're okay with domain validation (DV) level trust
- Cost matters (it's free)

**How it works:**
1. You provide a list of domains (e.g., `example.com`, `*.demo.example.com`)
2. Civo initiates ACME protocol with Let's Encrypt
3. For single domains, **HTTP-01** challenge: Let's Encrypt verifies you control the domain by checking `http://yourdomain/.well-known/acme-challenge/...`
4. For wildcards, **DNS-01** challenge: Civo creates `_acme-challenge` TXT records in Civo DNS to prove ownership
5. Certificate issued in minutes (typically under 60 seconds if DNS is configured)
6. Auto-renewal kicks in ~30 days before expiration

**Critical requirements:**
- **DNS pointing to Civo:** For HTTP-01, your domain's A/AAAA records must resolve to the Civo load balancer. For DNS-01 (wildcards), the domain must use Civo DNS or you must manually create TXT records.
- **Rate limits:** Let's Encrypt allows **50 certificates per domain per week** and **5 duplicate certificates** (same domain set) per week. Aggressive testing can hit these limits fast.

**Example configuration:**

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoCertificate
metadata:
  name: example-wildcard-cert
spec:
  certificate_name: example-wildcard-cert
  type: letsEncrypt
  lets_encrypt:
    domains:
      - "*.example.com"
      - "example.com"
    # disable_auto_renew: false (default)
```

This requests a wildcard cert covering both `*.example.com` and the apex `example.com`. Auto-renewal is enabled by default.

### Custom Certificates

**Use when:**
- You need Extended Validation (EV) or Organization Validation (OV) certificates
- You're using an internal Certificate Authority for private services
- You have compliance requirements mandating specific CAs
- You need certificates with lifespans longer than 90 days

**How it works:**
1. Obtain the certificate from your CA (Digicert, GlobalSign, internal PKI, etc.)
2. Provide Civo with:
   - **Leaf certificate:** Your domain's public cert (PEM format)
   - **Private key:** Corresponding private key (PEM, unencrypted)
   - **Certificate chain:** Intermediate CA certificates (optional but critical for trust)
3. Civo stores the certificate and makes it available to load balancers
4. **You** are responsible for monitoring expiration and renewing

**Critical requirements:**
- **PEM format:** Civo expects `-----BEGIN CERTIFICATE-----` blocks, not DER or PKCS#12
- **Unencrypted key:** Remove passphrases from private keys
- **Correct chain order:** Leaf → intermediate → root (but don't include the root itself)

**Example configuration:**

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoCertificate
metadata:
  name: corp-ev-cert
spec:
  certificate_name: corp-ev-cert
  type: custom
  custom:
    leaf_certificate: |
      -----BEGIN CERTIFICATE-----
      MIIFbjCCBFagAwIBAgIQDvO...
      (your certificate PEM)
      -----END CERTIFICATE-----
    private_key: |
      -----BEGIN RSA PRIVATE KEY-----
      MIIEpAIBAAKCAQEAv0...
      (your private key PEM)
      -----END RSA PRIVATE KEY-----
    certificate_chain: |
      -----BEGIN CERTIFICATE-----
      MIIEADCCAuigAwIBAgISA...
      (intermediate cert 1)
      -----END CERTIFICATE-----
      -----BEGIN CERTIFICATE-----
      (intermediate cert 2, if applicable)
      -----END CERTIFICATE-----
```

## Infrastructure as Code: Tool Comparison

### Terraform (and OpenTofu)

**Maturity:** Production-ready. Civo's Terraform provider is HashiCorp-certified and actively maintained.

**Certificate support:** The provider includes `civo_certificate` resources for both Let's Encrypt and custom certificates.

**Strengths:**
- Declarative HCL syntax familiar to most DevOps teams
- Strong ecosystem and community
- Excellent state management with remote backends

**Weaknesses:**
- **Secrets in state:** Private keys end up in Terraform state files. Use encrypted backends (S3 + KMS, Terraform Cloud) and mark variables as sensitive.
- **Lifecycle quirks:** Changing the domain list often forces resource replacement (delete then create), which can cause brief downtime if not handled with `create_before_destroy`.

**Example Terraform usage:**

```hcl
resource "civo_certificate" "wildcard" {
  name   = "wildcard-example-com"
  type   = "lets_encrypt"
  domains = ["*.example.com", "example.com"]
}

resource "civo_loadbalancer" "app_lb" {
  hostname      = "app.example.com"
  certificate_id = civo_certificate.wildcard.id
  # ... other config
}
```

**Verdict:** The default choice for most teams. Just ensure you're using remote state encryption.

### Pulumi

**Maturity:** Production-ready. Pulumi's Civo provider wraps the Terraform provider via their bridge system.

**Strengths:**
- **Real programming languages:** Write in TypeScript, Python, Go—use loops, conditionals, functions
- **Built-in secrets management:** Pulumi encrypts secrets in state by default
- **Type safety:** Catch configuration errors at compile time

**Weaknesses:**
- Smaller Civo-specific community compared to Terraform
- Requires learning Pulumi's programming model (though it's intuitive for developers)

**Example Pulumi usage (TypeScript):**

```typescript
import * as civo from "@pulumi/civo";

const cert = new civo.Certificate("wildcard-cert", {
  name: "wildcard-example-com",
  type: "lets_encrypt",
  domains: ["*.example.com", "example.com"],
});

const lb = new civo.LoadBalancer("app-lb", {
  hostname: "app.example.com",
  certificateId: cert.id,
});
```

**Verdict:** Excellent choice if your team prefers code over config, or if secret management is a top concern.

### Crossplane (and Planton)

**Maturity:** Emerging. The upstream Crossplane Civo provider doesn't yet include certificate resources, but **Planton's `CivoCertificate` custom resource fills this gap**.

**Strengths:**
- **Kubernetes-native:** Declare certificates as CRDs, manage them with `kubectl` and GitOps (Argo CD, Flux)
- **Unified control plane:** Manage Civo certs alongside AWS, GCP resources using the same patterns

**Weaknesses:**
- **Secret handling:** Private keys in YAML specs are visible in plaintext unless you integrate with external secret stores (Sealed Secrets, External Secrets Operator, etc.)
- **Smaller ecosystem:** Fewer examples and community resources compared to Terraform

**Verdict:** Best fit for teams already standardized on Crossplane or Kubernetes-centric workflows. Requires careful secret management.

## The 80/20 Configuration Rule

Most certificate configurations use a small subset of available fields. Here's what actually matters.

### Let's Encrypt Essentials

**Fields you'll always set:**
- `certificate_name`: Unique identifier (e.g., `prod-wildcard-cert`)
- `type`: `letsEncrypt`
- `domains`: List of FQDNs or wildcards (minimum 1)

**Fields you'll rarely touch:**
- `disable_auto_renew`: Almost always left `false` (auto-renew enabled)
- `description`: Optional metadata, typically empty
- `tags`: For billing/organization, but most teams skip it

**Minimal example:**

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoCertificate
metadata:
  name: simple-cert
spec:
  certificate_name: simple-cert
  type: letsEncrypt
  lets_encrypt:
    domains:
      - example.com
```

That's it. Three fields plus boilerplate.

### Custom Certificate Essentials

**Fields you'll always set:**
- `certificate_name`: Unique identifier
- `type`: `custom`
- `leaf_certificate`: Your public cert (PEM)
- `private_key`: Your private key (PEM)
- `certificate_chain`: Intermediate certs (PEM) — technically optional, but **always include it** for publicly-issued certs

**Minimal example:**

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoCertificate
metadata:
  name: custom-cert
spec:
  certificate_name: custom-cert
  type: custom
  custom:
    leaf_certificate: |
      -----BEGIN CERTIFICATE-----
      (paste your cert)
      -----END CERTIFICATE-----
    private_key: |
      -----BEGIN RSA PRIVATE KEY-----
      (paste your key)
      -----END RSA PRIVATE KEY-----
    certificate_chain: |
      -----BEGIN CERTIFICATE-----
      (paste intermediate certs)
      -----END CERTIFICATE-----
```

Five fields total. Everything else is optional noise.

## Production Best Practices

### Monitor Expiration Relentlessly

**Even with auto-renewal enabled**, things can fail:
- Domain ownership lapses
- DNS changes break validation
- Let's Encrypt has an outage
- Your custom cert renewal workflow has a bug

**Defense in depth:**
1. **Query Civo API daily:** Script a check that lists all certificates and flags any expiring within 15 days. The `CivoCertificateStatus.expiry_rfc3339` field provides this.
2. **External monitoring:** Use SSL Labs, UptimeRobot SSL monitors, or OpenSSL scripts that check your live HTTPS endpoints.
3. **Alerting:** Wire expiration warnings into PagerDuty, Slack, or your monitoring system.

**Don't rely on email:** Let's Encrypt stopped sending expiration emails in 2025. Civo may provide alerts, but verify this in their docs.

### Automate Everything (Except Custom Cert Procurement)

**Let's Encrypt on Civo:** Set it and forget it. Enable auto-renewal (default) and monitor.

**Custom certificates:** Automate as much as possible:
- Use ACME clients if your CA supports it (many commercial CAs now offer ACME endpoints)
- Script certificate updates via Civo's API or IaC tools
- Integrate with CI/CD pipelines to push renewed certs 30+ days before expiration

**Don't manually renew in production.** Humans forget. Calendars get ignored. Automation is reliable.

### Multi-Environment Strategy

**Development/Staging:**
- Use Let's Encrypt with unique subdomains (`dev.example.com`, `staging.example.com`)
- Consider self-signed certs if you're testing purely internal flows (but beware of trust errors)
- **Don't reuse production wildcard certs in dev** — security risk if dev is less hardened

**Production:**
- Separate certificates per security domain (wildcard for public services, distinct certs for admin panels)
- Use distinct Civo projects or accounts for isolation if needed
- Never let dev environments hit production Let's Encrypt rate limits

### TLS Termination: Load Balancer vs Kubernetes Ingress

**Two philosophies:**

1. **Terminate at Civo Load Balancer:**
   - Attach certificate to the LB
   - LB handles TLS, forwards HTTP to backend
   - **Pros:** Centralized cert management, offloads crypto from apps
   - **Cons:** Traffic between LB and pods is unencrypted (use a service mesh if this matters)

2. **TLS Pass-Through + Kubernetes Ingress:**
   - Civo LB forwards raw TLS traffic to cluster
   - cert-manager in Kubernetes handles Let's Encrypt via ACME
   - **Pros:** End-to-end encryption, per-service certificates
   - **Cons:** More complex setup, cert-manager learning curve

**Planton supports the first approach** — Civo-managed certificates attached to load balancers. If you prefer the second, you're better off using cert-manager with Civo DNS webhooks directly.

### Secure Private Key Handling

**Golden rules:**
1. **Never commit private keys to Git.** Use `.gitignore`, or better yet, don't write them to disk in repos at all.
2. **Use secret stores:** HashiCorp Vault, AWS Secrets Manager, Pulumi config secrets, Kubernetes Sealed Secrets.
3. **Rotate keys periodically:** Let's Encrypt does this automatically on each renewal. For custom certs, consider annual rotation even if the cert is valid longer.
4. **Delete old certificates:** Once replaced, remove unused certs from Civo to reduce clutter and confusion. Don't delete certs still attached to active load balancers (that'll break your service).

### Common Anti-Patterns to Avoid

**❌ Uploading only the leaf certificate (no chain)**
- **Result:** Browser trust errors, API clients reject connections
- **Fix:** Always include intermediate certs in `certificate_chain`

**❌ Hardcoding private keys in IaC files**
- **Result:** Keys leaked in Git history, compliance violations
- **Fix:** Use variables + secret stores

**❌ Ignoring Let's Encrypt rate limits during testing**
- **Result:** Hit 5-duplicate-cert limit, blocked for a week
- **Fix:** Use distinct subdomains for testing, or test against Let's Encrypt staging (if accessible)

**❌ Using the same certificate across 50 microservices**
- **Result:** If one service is compromised and cert revoked, all services break
- **Fix:** Use separate certs or scope wildcards appropriately

**❌ Forgetting to verify certificate chain after renewal**
- **Result:** Auto-renewal succeeded, but chain is incomplete (edge case, but it happens)
- **Fix:** Automated post-renewal checks that curl your endpoint and validate the chain

## The Planton Choice

Planton's `CivoCertificate` resource abstracts Civo's certificate API into a Kubernetes-style declarative interface. We support both Let's Encrypt and custom certificates because **production environments need both**:

- **Let's Encrypt for everything we can:** Free, automated, secure. This covers 90% of use cases.
- **Custom certificates when regulations or trust models demand it:** Extended Validation certs, internal CAs, compliance-mandated issuers.

Our design philosophy:
- **Minimal required configuration:** Only `certificate_name`, `type`, and the relevant parameters (domains or PEM data). Everything else is optional.
- **Secure by default:** Auto-renewal enabled, validation rules enforced via protobuf constraints.
- **Interoperable:** Works with Terraform, Pulumi, or as a Crossplane CRD.

We don't reinvent certificate lifecycle management — we integrate with Civo's platform capabilities and expose them through clean APIs.

## Civo-Specific Nuances

### Integration with Civo Services

**Primary use case:** Attaching certificates to **Civo Load Balancers** for HTTPS termination.

**Not supported:** Using these certificates directly with Kubernetes API endpoints (those use self-signed certs by default) or object storage (Civo doesn't offer S3-like cert attachment).

### Regional Considerations

Certificates appear to be globally available within your Civo account, but load balancers are region-specific. If you have multi-region deployments serving the same domain, you may need to:
- Create duplicate certificates per region (if Civo scopes them regionally behind the scenes)
- Or verify that a single cert can attach to LBs in multiple regions

Check current Civo documentation for regional scoping details.

### Cost Implications

- **Let's Encrypt certificates:** Free
- **Certificate management feature:** No additional charge
- **Civo Load Balancers:** ~$10/month (as of 2025)

**The cert itself costs nothing.** You pay for the infrastructure that uses it.

### Comparison to AWS/GCP/Azure

| Feature | Civo + Let's Encrypt | AWS ACM | GCP Managed Certs | Azure Key Vault |
|---------|---------------------|---------|------------------|-----------------|
| **Cost** | Free | Free (AWS-only) | Free | Cert storage free; issuance via Digicert costs $ |
| **CA** | Let's Encrypt | Amazon Trust Services | Let's Encrypt / Google Trust | Variable |
| **Auto-renewal** | Yes (90 days) | Yes (13 months) | Yes | Depends on integration |
| **Lifespan** | 90 days | 13 months | 90 days | Variable |
| **Private key access** | No (after upload) | No | No | Yes (in Vault) |
| **Wildcard support** | Yes | Yes | Yes | Yes |

**Civo's approach is closest to GCP's:** Simple, Let's Encrypt-based, auto-renewed. AWS's longer lifespan is convenient, but Civo's shorter renewal cycle arguably improves security through more frequent key rotation.

## What's Next

This document covered the fundamentals of certificate deployment on Civo. For deeper dives:

- **Terraform Provider Deep Dive:** Detailed resource configuration, state management patterns, and production workflows (create this as a separate guide)
- **Pulumi Implementation Guide:** TypeScript/Python examples, secret handling, stack organization (create this as a separate guide)
- **Crossplane Composition Patterns:** Building reusable `CivoCertificate` compositions, integrating with External Secrets Operator (create this as a separate guide)

## Conclusion

The shift from manual certificate management to automated lifecycle handling represents one of the clearest wins in modern DevOps. Let's Encrypt proved that free, automated certificates work at massive scale. Civo's integration brings this capability to their platform with minimal friction.

**The strategic choice:** Default to Let's Encrypt for everything. Use custom certificates only when you have a compelling reason (EV certs, internal CA, compliance). Automate via Terraform or Pulumi. Monitor expiration obsessively.

Done right, you'll never think about certificates again — which is exactly the point.


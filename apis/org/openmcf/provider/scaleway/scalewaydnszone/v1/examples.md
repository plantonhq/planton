# Scaleway DNS Zone Examples

This document provides YAML manifest examples for creating and managing Scaleway DNS zones using OpenMCF.

## Table of Contents

1. [Minimal Example: Empty Root Zone](#minimal-example-empty-root-zone)
2. [Root Zone with Email Configuration](#root-zone-with-email-configuration)
3. [Zone with Website Records](#zone-with-website-records)
4. [Subdomain Zone for Staging](#subdomain-zone-for-staging)
5. [CAA Records for Certificate Security](#caa-records-for-certificate-security)
6. [Complete Production Setup](#complete-production-setup)
7. [Cross-Resource References](#cross-resource-references)
8. [Infra Chart Integration](#infra-chart-integration)

---

## Minimal Example: Empty Root Zone

The simplest configuration creates a DNS zone with no records. Records can be added later via standalone ScalewayDnsRecord resources or infra chart templates.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsZone
metadata:
  name: my-domain-zone
  org: my-org
  env: production
spec:
  domain: example.com
```

**What this creates:**
- A root DNS zone for `example.com`
- Scaleway assigns nameservers (visible in `status.outputs.name_servers`)
- No DNS records (zone is empty until records are added)

---

## Root Zone with Email Configuration

A common pattern: create the zone with MX, SPF, and DMARC records for email services.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsZone
metadata:
  name: company-dns
  org: my-org
  env: production
spec:
  domain: mycompany.com
  records:
    # Google Workspace MX records
    - name: ""
      type: MX
      data:
        value: "aspmx.l.google.com."
      ttl: 3600
      priority: 1

    - name: ""
      type: MX
      data:
        value: "alt1.aspmx.l.google.com."
      ttl: 3600
      priority: 5

    - name: ""
      type: MX
      data:
        value: "alt2.aspmx.l.google.com."
      ttl: 3600
      priority: 5

    # SPF record
    - name: ""
      type: TXT
      data:
        value: "v=spf1 include:_spf.google.com ~all"
      ttl: 3600

    # DMARC record
    - name: "_dmarc"
      type: TXT
      data:
        value: "v=DMARC1; p=quarantine; rua=mailto:dmarc@mycompany.com"
      ttl: 3600
```

**Note:** Replace Google Workspace MX values with your email provider's records if using a different service.

---

## Zone with Website Records

DNS configuration for a website with apex and www records.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsZone
metadata:
  name: website-dns
  org: my-org
  env: production
spec:
  domain: example.com
  records:
    # Apex A record pointing to server IP
    - name: ""
      type: A
      data:
        value: "192.0.2.1"
      ttl: 3600

    # WWW as CNAME to apex
    - name: "www"
      type: CNAME
      data:
        value: "example.com."
      ttl: 3600

    # API subdomain
    - name: "api"
      type: A
      data:
        value: "192.0.2.2"
      ttl: 300
```

**Tip:** Use lower TTL (300s) for records that may change during deployments, and higher TTL (3600s+) for stable records.

---

## Subdomain Zone for Staging

Create a delegated subdomain zone for environment isolation.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsZone
metadata:
  name: staging-zone
  org: my-org
  env: staging
spec:
  domain: example.com
  subdomain: staging
  records:
    # Staging app endpoint
    - name: "app"
      type: A
      data:
        value: "10.0.1.10"
      ttl: 300

    # Staging API endpoint
    - name: "api"
      type: A
      data:
        value: "10.0.1.11"
      ttl: 300
```

**What this creates:**
- A subdomain zone for `staging.example.com`
- Records resolve as `app.staging.example.com` and `api.staging.example.com`
- Separate nameservers from the parent zone (subdomain delegation)

---

## CAA Records for Certificate Security

Restrict which Certificate Authorities can issue certificates for your domain.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsZone
metadata:
  name: secure-domain
  org: my-org
  env: production
spec:
  domain: secure-app.com
  records:
    # Only Let's Encrypt can issue certificates
    - name: ""
      type: CAA
      data:
        value: "0 issue \"letsencrypt.org\""
      ttl: 3600

    # Only Let's Encrypt can issue wildcard certificates
    - name: ""
      type: CAA
      data:
        value: "0 issuewild \"letsencrypt.org\""
      ttl: 3600

    # Report violations
    - name: ""
      type: CAA
      data:
        value: "0 iodef \"mailto:security@secure-app.com\""
      ttl: 3600
```

---

## Complete Production Setup

Comprehensive DNS zone with email, website, security, and monitoring records.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsZone
metadata:
  name: production-dns
  org: my-org
  env: production
spec:
  domain: production-app.com
  records:
    # ── Website ──────────────────────────────────────────
    - name: ""
      type: A
      data:
        value: "192.0.2.1"
      ttl: 3600

    - name: "www"
      type: CNAME
      data:
        value: "production-app.com."
      ttl: 3600

    # ── API ──────────────────────────────────────────────
    - name: "api"
      type: A
      data:
        value: "192.0.2.2"
      ttl: 3600

    # ── Email (Google Workspace) ─────────────────────────
    - name: ""
      type: MX
      data:
        value: "aspmx.l.google.com."
      ttl: 3600
      priority: 1

    - name: ""
      type: MX
      data:
        value: "alt1.aspmx.l.google.com."
      ttl: 3600
      priority: 5

    # SPF
    - name: ""
      type: TXT
      data:
        value: "v=spf1 include:_spf.google.com ~all"
      ttl: 3600

    # DMARC
    - name: "_dmarc"
      type: TXT
      data:
        value: "v=DMARC1; p=reject; rua=mailto:dmarc@production-app.com"
      ttl: 3600

    # ── Security (CAA) ───────────────────────────────────
    - name: ""
      type: CAA
      data:
        value: "0 issue \"letsencrypt.org\""
      ttl: 3600

    # ── Domain Verification ──────────────────────────────
    - name: ""
      type: TXT
      data:
        value: "google-site-verification=abc123def456"
      ttl: 3600
```

---

## Cross-Resource References

Inline records can reference outputs from other infrastructure resources using `StringValueOrRef`. This is useful when record values are determined by other resources.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsZone
metadata:
  name: dynamic-dns
  org: my-org
  env: production
spec:
  domain: myapp.com
  records:
    # A record pointing to a Scaleway Load Balancer's IP
    - name: ""
      type: A
      data:
        valueFromResourceOutput:
          resourceIdRef:
            name: my-load-balancer
          outputKey: ip_address
      ttl: 3600

    # CNAME pointing to a Kapsule cluster's wildcard DNS
    - name: "k8s"
      type: CNAME
      data:
        valueFromResourceOutput:
          resourceIdRef:
            name: my-kapsule-cluster
          outputKey: wildcard_dns
      ttl: 3600

    # Static email records alongside dynamic infrastructure records
    - name: ""
      type: MX
      data:
        value: "aspmx.l.google.com."
      ttl: 3600
      priority: 1
```

**Note:** For complex infra chart scenarios with many cross-resource references, prefer standalone ScalewayDnsRecord resources for better DAG visibility.

---

## Infra Chart Integration

When composing ScalewayDnsZone with other resources in infra charts, downstream resources reference the zone's outputs using `valueFrom`.

### ScalewayDnsRecord Referencing This Zone

```yaml
# In an infra chart template:
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsRecord
metadata:
  name: "{{ values.app_name }}-dns-record"
  org: "{{ values.org }}"
  env: "{{ values.env }}"
spec:
  zone_name:
    valueFrom:
      kind: ScalewayDnsZone
      name: "{{ values.zone_resource_name }}"
      fieldPath: status.outputs.zone_name
  name: "{{ values.subdomain }}"
  type: A
  data:
    valueFrom:
      kind: ScalewayLoadBalancer
      name: "{{ values.lb_name }}"
      fieldPath: status.outputs.ip_address
  ttl: 3600
```

### Registrar Delegation Output

After deployment, retrieve the nameservers for registrar configuration:

```bash
# Get nameservers from stack outputs
openmcf stack-outputs --manifest dns-zone.yaml --field name_servers

# Verify DNS resolution through Scaleway nameservers
dig @<scaleway-ns> example.com
```

---

## Tips and Best Practices

### TTL Strategy

- **3600s (1 hour)** -- Default for most records
- **300s (5 minutes)** -- During migrations, deployments, or cutover events
- **86400s (24 hours)** -- Static records that rarely change (NS delegation, CAA)

Lower TTLs **before** making critical DNS changes, then raise them after confirming success.

### Record Naming

- Use empty string `""` for the zone apex (root domain)
- Fully qualify CNAME/MX/NS targets with a trailing dot: `"example.com."` not `"example.com"`
- Use lowercase for record names

### When to Use Inline vs Standalone Records

**Inline (in zone spec):**
- MX records for email
- TXT records for SPF, DKIM, DMARC, domain verification
- CAA records for certificate security
- Static A/CNAME records with known values

**Standalone ScalewayDnsRecord:**
- Records pointing to dynamically provisioned infrastructure (LB IPs, cluster endpoints)
- Records that need independent lifecycle management
- Records managed by different teams or environments

### Testing

After applying DNS changes:
1. Wait for TTL to expire before testing
2. Use `dig @<ns> example.com` to query Scaleway nameservers directly
3. Use online DNS propagation checkers for global verification
4. Test email delivery with mail-tester.com for MX/SPF/DMARC records

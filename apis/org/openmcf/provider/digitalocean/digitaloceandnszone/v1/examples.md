# DigitalOcean DNS Zone Examples

This document provides practical examples of common DNS zone configurations using the DigitalOcean DNS Zone API resource.

## Table of Contents

1. [Simple Website (Apex + WWW)](#simple-website-apex--www)
2. [Email Configuration (Google Workspace)](#email-configuration-google-workspace)
3. [Application with Load Balancer](#application-with-load-balancer)
4. [CDN Integration with Spaces](#cdn-integration-with-spaces)
5. [Multi-Environment Setup](#multi-environment-setup)
6. [CAA Records for Let's Encrypt](#caa-records-for-lets-encrypt)
7. [SRV Records for Services](#srv-records-for-services)
8. [Complete Production Setup](#complete-production-setup)

---

## Simple Website (Apex + WWW)

Basic DNS configuration for a website hosted on a DigitalOcean Droplet or Load Balancer.

```yaml
apiVersion: code2cloud.planton.ai/v1
kind: DigitalOceanDnsZone
metadata:
  name: simple-website
spec:
  digitalOceanCredentialId: do-prod-cred
  domainName: example.com
  records:
    # Apex record pointing to Droplet/Load Balancer IP
    - name: "@"
      type: dns_record_type_a
      values:
        - value: "192.0.2.1"
      ttlSeconds: 3600
    
    # WWW subdomain as CNAME to apex
    - name: "www"
      type: dns_record_type_cname
      values:
        - value: "example.com."
      ttlSeconds: 3600
```

**Result**: Both `example.com` and `www.example.com` resolve to `192.0.2.1`.

---

## Email Configuration (Google Workspace)

Complete email setup with MX records and email authentication (SPF, DMARC, DKIM).

```yaml
apiVersion: code2cloud.planton.ai/v1
kind: DigitalOceanDnsZone
metadata:
  name: email-enabled-domain
spec:
  digitalOceanCredentialId: do-prod-cred
  domainName: mycompany.com
  records:
    # Google Workspace MX records (priority matters!)
    - name: "@"
      type: dns_record_type_mx
      priority: 1
      values:
        - value: "aspmx.l.google.com."
      ttlSeconds: 3600
    
    - name: "@"
      type: dns_record_type_mx
      priority: 5
      values:
        - value: "alt1.aspmx.l.google.com."
      ttlSeconds: 3600
    
    - name: "@"
      type: dns_record_type_mx
      priority: 5
      values:
        - value: "alt2.aspmx.l.google.com."
      ttlSeconds: 3600
    
    - name: "@"
      type: dns_record_type_mx
      priority: 10
      values:
        - value: "alt3.aspmx.l.google.com."
      ttlSeconds: 3600
    
    - name: "@"
      type: dns_record_type_mx
      priority: 10
      values:
        - value: "alt4.aspmx.l.google.com."
      ttlSeconds: 3600
    
    # SPF record (authorizes Google to send on behalf of domain)
    - name: "@"
      type: dns_record_type_txt
      values:
        - value: "v=spf1 include:_spf.google.com ~all"
      ttlSeconds: 3600
    
    # DMARC record (email authentication policy)
    - name: "_dmarc"
      type: dns_record_type_txt
      values:
        - value: "v=DMARC1; p=quarantine; rua=mailto:dmarc-reports@mycompany.com"
      ttlSeconds: 3600
    
    # DKIM record (replace with actual key from Google Workspace)
    - name: "google._domainkey"
      type: dns_record_type_txt
      values:
        - value: "v=DKIM1; k=rsa; p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC..."
      ttlSeconds: 3600
```

**Note**: Replace the DKIM public key with the actual value provided by Google Workspace.

---

## Application with Load Balancer

DNS configuration for an application with separate API and web frontends behind a DigitalOcean Load Balancer.

```yaml
apiVersion: code2cloud.planton.ai/v1
kind: DigitalOceanDnsZone
metadata:
  name: app-load-balanced
spec:
  digitalOceanCredentialId: do-prod-cred
  domainName: myapp.io
  records:
    # Apex points to Load Balancer
    - name: "@"
      type: dns_record_type_a
      values:
        - value: "104.248.1.1"
      ttlSeconds: 300  # Low TTL during migration
    
    # WWW as CNAME to apex
    - name: "www"
      type: dns_record_type_cname
      values:
        - value: "myapp.io."
      ttlSeconds: 300
    
    # API subdomain pointing to same Load Balancer
    - name: "api"
      type: dns_record_type_a
      values:
        - value: "104.248.1.1"
      ttlSeconds: 300
    
    # Admin panel on separate subdomain
    - name: "admin"
      type: dns_record_type_a
      values:
        - value: "104.248.1.1"
      ttlSeconds: 300
```

**Note**: Low TTL (300s) allows for quick updates during infrastructure changes. Increase to 3600s after stabilization.

---

## CDN Integration with Spaces

Pointing subdomains to DigitalOcean Spaces CDN endpoints for static assets.

```yaml
apiVersion: code2cloud.planton.ai/v1
kind: DigitalOceanDnsZone
metadata:
  name: cdn-enabled-site
spec:
  digitalOceanCredentialId: do-prod-cred
  domainName: mysite.com
  records:
    # Main site on Droplet
    - name: "@"
      type: dns_record_type_a
      values:
        - value: "192.0.2.1"
      ttlSeconds: 3600
    
    # Static assets subdomain pointing to Spaces CDN
    - name: "cdn"
      type: dns_record_type_cname
      values:
        - value: "mysite-assets.nyc3.cdn.digitaloceanspaces.com."
      ttlSeconds: 86400  # High TTL for CDN endpoints
    
    # Images subdomain also on Spaces
    - name: "images"
      type: dns_record_type_cname
      values:
        - value: "mysite-images.nyc3.cdn.digitaloceanspaces.com."
      ttlSeconds: 86400
```

**Benefit**: CDN endpoints have high TTL since they rarely change, reducing DNS query load.

---

## Multi-Environment Setup

Managing DNS for development, staging, and production environments using subdomains.

```yaml
apiVersion: code2cloud.planton.ai/v1
kind: DigitalOceanDnsZone
metadata:
  name: multi-env-dns
spec:
  digitalOceanCredentialId: do-prod-cred
  domainName: acme-corp.com
  records:
    # Production (apex)
    - name: "@"
      type: dns_record_type_a
      values:
        - value: "192.0.2.1"
      ttlSeconds: 3600
    
    # Staging environment
    - name: "staging"
      type: dns_record_type_a
      values:
        - value: "192.0.2.10"
      ttlSeconds: 600  # Lower TTL for frequent changes
    
    # Development environment
    - name: "dev"
      type: dns_record_type_a
      values:
        - value: "192.0.2.20"
      ttlSeconds: 300  # Very low TTL for rapid iteration
    
    # API endpoints per environment
    - name: "api"
      type: dns_record_type_a
      values:
        - value: "192.0.2.1"
      ttlSeconds: 3600
    
    - name: "api.staging"
      type: dns_record_type_a
      values:
        - value: "192.0.2.10"
      ttlSeconds: 600
    
    - name: "api.dev"
      type: dns_record_type_a
      values:
        - value: "192.0.2.20"
      ttlSeconds: 300
```

---

## CAA Records for Let's Encrypt

Authorizing certificate issuance for enhanced security and compliance.

```yaml
apiVersion: code2cloud.planton.ai/v1
kind: DigitalOceanDnsZone
metadata:
  name: caa-secured-domain
spec:
  digitalOceanCredentialId: do-prod-cred
  domainName: secure-app.com
  records:
    # Website A record
    - name: "@"
      type: dns_record_type_a
      values:
        - value: "192.0.2.1"
      ttlSeconds: 3600
    
    # CAA record: Only Let's Encrypt can issue certificates
    - name: "@"
      type: dns_record_type_caa
      flags: 0  # Non-critical
      tag: "issue"
      values:
        - value: "letsencrypt.org"
      ttlSeconds: 3600
    
    # CAA record: Only Let's Encrypt can issue wildcard certificates
    - name: "@"
      type: dns_record_type_caa
      flags: 0
      tag: "issuewild"
      values:
        - value: "letsencrypt.org"
      ttlSeconds: 3600
    
    # CAA record: Report violations to this URL
    - name: "@"
      type: dns_record_type_caa
      flags: 0
      tag: "iodef"
      values:
        - value: "mailto:security@secure-app.com"
      ttlSeconds: 3600
```

**Security Benefit**: CAA records prevent unauthorized certificate authorities from issuing certificates for your domain, reducing phishing risk.

---

## SRV Records for Services

Configuring SRV records for service discovery (e.g., SIP, XMPP, Minecraft).

```yaml
apiVersion: code2cloud.planton.ai/v1
kind: DigitalOceanDnsZone
metadata:
  name: srv-example
spec:
  digitalOceanCredentialId: do-prod-cred
  domainName: gaming-server.com
  records:
    # Minecraft server SRV record
    # Format: _service._proto.name
    - name: "_minecraft._tcp"
      type: dns_record_type_srv
      priority: 10
      weight: 60
      port: 25565
      values:
        - value: "mc1.gaming-server.com."
      ttlSeconds: 3600
    
    # Secondary Minecraft server (load distribution)
    - name: "_minecraft._tcp"
      type: dns_record_type_srv
      priority: 10
      weight: 40
      port: 25565
      values:
        - value: "mc2.gaming-server.com."
      ttlSeconds: 3600
    
    # Actual A records for the Minecraft servers
    - name: "mc1"
      type: dns_record_type_a
      values:
        - value: "192.0.2.10"
      ttlSeconds: 3600
    
    - name: "mc2"
      type: dns_record_type_a
      values:
        - value: "192.0.2.11"
      ttlSeconds: 3600
```

**How it works**: Clients query `_minecraft._tcp.gaming-server.com` and get both servers. The one with weight 60 is chosen 60% of the time.

---

## Complete Production Setup

Comprehensive DNS configuration combining all best practices.

```yaml
apiVersion: code2cloud.planton.ai/v1
kind: DigitalOceanDnsZone
metadata:
  name: production-complete
spec:
  digitalOceanCredentialId: do-prod-cred
  domainName: production-app.com
  records:
    # ====================
    # Web Traffic (Load Balanced)
    # ====================
    - name: "@"
      type: dns_record_type_a
      values:
        - value: "104.248.1.1"
      ttlSeconds: 3600
    
    - name: "www"
      type: dns_record_type_cname
      values:
        - value: "production-app.com."
      ttlSeconds: 3600
    
    # ====================
    # API Endpoints
    # ====================
    - name: "api"
      type: dns_record_type_a
      values:
        - value: "104.248.1.1"
      ttlSeconds: 3600
    
    - name: "api-v2"
      type: dns_record_type_a
      values:
        - value: "104.248.2.1"
      ttlSeconds: 3600
    
    # ====================
    # Static Assets (CDN)
    # ====================
    - name: "assets"
      type: dns_record_type_cname
      values:
        - value: "prod-assets.nyc3.cdn.digitaloceanspaces.com."
      ttlSeconds: 86400
    
    # ====================
    # Email (Google Workspace)
    # ====================
    - name: "@"
      type: dns_record_type_mx
      priority: 1
      values:
        - value: "aspmx.l.google.com."
      ttlSeconds: 3600
    
    - name: "@"
      type: dns_record_type_mx
      priority: 5
      values:
        - value: "alt1.aspmx.l.google.com."
      ttlSeconds: 3600
    
    # SPF record
    - name: "@"
      type: dns_record_type_txt
      values:
        - value: "v=spf1 include:_spf.google.com ~all"
      ttlSeconds: 3600
    
    # DMARC record
    - name: "_dmarc"
      type: dns_record_type_txt
      values:
        - value: "v=DMARC1; p=reject; rua=mailto:dmarc@production-app.com"
      ttlSeconds: 3600
    
    # ====================
    # Domain Verification
    # ====================
    - name: "@"
      type: dns_record_type_txt
      values:
        - value: "google-site-verification=abcd1234..."
      ttlSeconds: 3600
    
    # ====================
    # Security (CAA)
    # ====================
    - name: "@"
      type: dns_record_type_caa
      flags: 0
      tag: "issue"
      values:
        - value: "letsencrypt.org"
      ttlSeconds: 3600
    
    - name: "@"
      type: dns_record_type_caa
      flags: 0
      tag: "issuewild"
      values:
        - value: "letsencrypt.org"
      ttlSeconds: 3600
    
    # ====================
    # Monitoring/Status
    # ====================
    - name: "status"
      type: dns_record_type_cname
      values:
        - value: "production-app.statuspage.io."
      ttlSeconds: 3600
```

---

## Using Cross-Resource References

You can reference outputs from other resources using `StringValueOrRef`.

```yaml
apiVersion: code2cloud.planton.ai/v1
kind: DigitalOceanDnsZone
metadata:
  name: cross-ref-example
spec:
  digitalOceanCredentialId: do-prod-cred
  domainName: dynamic-app.com
  records:
    # Reference a Load Balancer IP from another resource
    - name: "@"
      type: dns_record_type_a
      values:
        - valueFromResourceOutput:
            resourceIdRef:
              name: my-load-balancer
            outputKey: ip_address
      ttlSeconds: 3600
    
    # Reference an output from a Droplet
    - name: "api"
      type: dns_record_type_a
      values:
        - valueFromResourceOutput:
            resourceIdRef:
              name: api-droplet
            outputKey: ipv4_address
      ttlSeconds: 3600
```

This allows for fully dynamic DNS configurations that update automatically when infrastructure changes.

---

## Tips and Best Practices

### TTL Selection

- **3600s (1 hour)**: Default for most records
- **300s (5 minutes)**: During migrations or cutover events
- **86400s (24 hours)**: For static records that rarely change (CDN, nameservers)

### Record Naming

- Use `"@"` for the apex/root domain
- Use `"www"` for the www subdomain (not `"www.example.com"`)
- Fully qualify CNAME targets with a trailing dot: `"example.com."` not `"example.com"`

### Priority for MX Records

- Lower numbers = higher priority
- Google Workspace convention: 1, 5, 5, 10, 10
- Office 365 convention: All 0 (equal priority)

### Testing

After applying DNS changes:

1. Wait for TTL to expire before testing
2. Use `dig @ns1.digitalocean.com example.com` to query DigitalOcean nameservers directly
3. Use online DNS propagation checkers to verify global propagation
4. Test email delivery with mail-tester.com for email records

### Delegation Checklist

Before going live:

- [ ] Disable DNSSEC at registrar (if enabled)
- [ ] Update nameservers to DigitalOcean's
- [ ] Wait 24-48 hours for full propagation
- [ ] Verify all records resolve correctly
- [ ] Test website and email functionality


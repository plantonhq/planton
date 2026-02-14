# DigitalOcean DNS Record Examples

Concrete, copy-and-paste examples for common DigitalOcean DNS record deployment scenarios.

## Table of Contents

- [A Record (IPv4)](#a-record-ipv4)
- [AAAA Record (IPv6)](#aaaa-record-ipv6)
- [CNAME Record](#cname-record)
- [MX Records (Email)](#mx-records-email)
- [TXT Records](#txt-records)
- [SRV Record](#srv-record)
- [CAA Record](#caa-record)
- [Root Domain (Apex)](#root-domain-apex)
- [Wildcard Subdomain](#wildcard-subdomain)
- [Complete Email Setup](#complete-email-setup)

---

## A Record (IPv4)

Map a subdomain to an IPv4 address:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: www-a-record
spec:
  domain:
    value: "example.com"
  name: "www"
  type: A
  value:
    value: "192.0.2.1"
  ttl_seconds: 3600
```

**Deploy:**
```bash
planton apply -f www-a-record.yaml
```

**Use Case:** Web servers, load balancers, any HTTP/HTTPS services.

---

## AAAA Record (IPv6)

Map a subdomain to an IPv6 address:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: www-aaaa-record
spec:
  domain:
    value: "example.com"
  name: "www"
  type: AAAA
  value:
    value: "2001:db8::1"
```

**Use Case:** Dual-stack deployments supporting both IPv4 and IPv6.

---

## CNAME Record

Create an alias pointing to another hostname:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: app-cname
spec:
  domain:
    value: "example.com"
  name: "app"
  type: CNAME
  value:
    value: "www.example.com"
```

**Use Case:** Aliases, CDN endpoints, SaaS integrations.

**Note:** CNAME cannot be used at the root domain (apex). Use A record instead.

---

## MX Records (Email)

### Primary Mail Server

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: mx-primary
spec:
  domain:
    value: "example.com"
  name: "@"
  type: MX
  value:
    value: "mail.example.com"
  priority: 10
```

### Backup Mail Server

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: mx-backup
spec:
  domain:
    value: "example.com"
  name: "@"
  type: MX
  value:
    value: "backup-mail.example.com"
  priority: 20
```

### Google Workspace MX Records

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: mx-google-1
spec:
  domain:
    value: "example.com"
  name: "@"
  type: MX
  value:
    value: "aspmx.l.google.com"
  priority: 1
---
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: mx-google-2
spec:
  domain:
    value: "example.com"
  name: "@"
  type: MX
  value:
    value: "alt1.aspmx.l.google.com"
  priority: 5
---
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: mx-google-3
spec:
  domain:
    value: "example.com"
  name: "@"
  type: MX
  value:
    value: "alt2.aspmx.l.google.com"
  priority: 5
```

---

## TXT Records

### SPF Record

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: spf-record
spec:
  domain:
    value: "example.com"
  name: "@"
  type: TXT
  value:
    value: "v=spf1 include:_spf.google.com ~all"
```

### DKIM Record

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: dkim-google
spec:
  domain:
    value: "example.com"
  name: "google._domainkey"
  type: TXT
  value:
    value: "v=DKIM1; k=rsa; p=MIGfMA0GCSqGSIb3DQEB..."
```

### DMARC Record

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: dmarc-record
spec:
  domain:
    value: "example.com"
  name: "_dmarc"
  type: TXT
  value:
    value: "v=DMARC1; p=quarantine; rua=mailto:dmarc@example.com"
```

### Domain Verification

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: google-verification
spec:
  domain:
    value: "example.com"
  name: "@"
  type: TXT
  value:
    value: "google-site-verification=abc123xyz..."
```

---

## SRV Record

Service location record for SIP:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: sip-srv
spec:
  domain:
    value: "example.com"
  name: "_sip._tcp"
  type: SRV
  value:
    value: "sipserver.example.com"
  priority: 10
  weight: 5
  port: 5060
```

### XMPP SRV Records

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: xmpp-client
spec:
  domain:
    value: "example.com"
  name: "_xmpp-client._tcp"
  type: SRV
  value:
    value: "xmpp.example.com"
  priority: 5
  weight: 0
  port: 5222
---
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: xmpp-server
spec:
  domain:
    value: "example.com"
  name: "_xmpp-server._tcp"
  type: SRV
  value:
    value: "xmpp.example.com"
  priority: 5
  weight: 0
  port: 5269
```

---

## CAA Record

Control which Certificate Authorities can issue certificates for your domain:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: caa-letsencrypt
spec:
  domain:
    value: "example.com"
  name: "@"
  type: CAA
  value:
    value: "letsencrypt.org"
  flags: 0
  tag: "issue"
```

### Allow Multiple CAs

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: caa-digicert
spec:
  domain:
    value: "example.com"
  name: "@"
  type: CAA
  value:
    value: "digicert.com"
  flags: 0
  tag: "issue"
```

### Allow Wildcard Certificates from Specific CA

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: caa-wildcard
spec:
  domain:
    value: "example.com"
  name: "@"
  type: CAA
  value:
    value: "letsencrypt.org"
  flags: 0
  tag: "issuewild"
```

---

## Root Domain (Apex)

### A Record at Root

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: apex-a-record
spec:
  domain:
    value: "example.com"
  name: "@"
  type: A
  value:
    value: "192.0.2.1"
```

---

## Wildcard Subdomain

Route all unmatched subdomains:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: wildcard-record
spec:
  domain:
    value: "example.com"
  name: "*"
  type: A
  value:
    value: "192.0.2.1"
```

**Use Case:** Multi-tenant applications, catch-all routing.

---

## Complete Email Setup

All records needed for professional email:

### MX Records

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: email-mx-1
spec:
  domain:
    value: "example.com"
  name: "@"
  type: MX
  value:
    value: "aspmx.l.google.com"
  priority: 1
---
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: email-mx-2
spec:
  domain:
    value: "example.com"
  name: "@"
  type: MX
  value:
    value: "alt1.aspmx.l.google.com"
  priority: 5
```

### SPF Record

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: email-spf
spec:
  domain:
    value: "example.com"
  name: "@"
  type: TXT
  value:
    value: "v=spf1 include:_spf.google.com ~all"
```

### DKIM Record

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: email-dkim
spec:
  domain:
    value: "example.com"
  name: "google._domainkey"
  type: TXT
  value:
    value: "v=DKIM1; k=rsa; p=YOUR_DKIM_PUBLIC_KEY"
```

### DMARC Record

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: email-dmarc
spec:
  domain:
    value: "example.com"
  name: "_dmarc"
  type: TXT
  value:
    value: "v=DMARC1; p=quarantine; pct=100; rua=mailto:dmarc-reports@example.com"
```

---

## CLI Commands

### Validate Manifest

```bash
planton validate -f record.yaml
```

### Deploy with Pulumi

```bash
planton pulumi up -f record.yaml
```

### Deploy with Terraform

```bash
planton tofu apply -f record.yaml
```

### View Outputs

```bash
planton output record_id
planton output hostname
```

---

## Support

For questions or issues:
- **OpenMCF**: [openmcf.org](https://openmcf.org)
- **DigitalOcean DNS Docs**: [docs.digitalocean.com/products/networking/dns](https://docs.digitalocean.com/products/networking/dns)

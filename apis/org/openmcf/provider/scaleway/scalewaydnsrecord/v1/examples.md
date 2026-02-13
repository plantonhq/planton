# Scaleway DNS Record Examples

This document provides YAML manifest examples for creating and managing standalone Scaleway DNS records using OpenMCF.

## Table of Contents

1. [A Record with Literal IP](#a-record-with-literal-ip)
2. [A Record Referencing a Load Balancer](#a-record-referencing-a-load-balancer)
3. [CNAME to Kapsule Cluster Endpoint](#cname-to-kapsule-cluster-endpoint)
4. [MX Records for Email](#mx-records-for-email)
5. [TXT Record for SPF](#txt-record-for-spf)
6. [CAA Record for Certificate Security](#caa-record-for-certificate-security)
7. [SRV Record](#srv-record)
8. [TLSA Record for DANE](#tlsa-record-for-dane)
9. [DNAME Record for Subtree Delegation](#dname-record-for-subtree-delegation)
10. [Infra Chart Integration](#infra-chart-integration)

---

## A Record with Literal IP

The simplest case: point a subdomain to a known IP address.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsRecord
metadata:
  name: www-a-record
  org: my-org
  env: production
spec:
  zone_name:
    value: "example.com"
  name: "www"
  type: A
  data:
    value: "192.0.2.1"
  ttl: 3600
```

**What this creates:**
- An A record for `www.example.com` pointing to `192.0.2.1`
- TTL of 1 hour
- Output: `fqdn` = `www.example.com`

---

## A Record Referencing a Load Balancer

Point a domain to a ScalewayLoadBalancer's IP using `StringValueOrRef`. This creates an explicit dependency edge in the infra chart DAG.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsRecord
metadata:
  name: app-lb-record
  org: my-org
  env: production
spec:
  zone_name:
    valueFrom:
      kind: ScalewayDnsZone
      name: my-dns-zone
      fieldPath: status.outputs.zone_name
  name: "app"
  type: A
  data:
    valueFrom:
      kind: ScalewayLoadBalancer
      name: my-load-balancer
      fieldPath: status.outputs.ip_address
  ttl: 300
```

**Key points:**
- `zone_name` references a ScalewayDnsZone resource (DAG edge 1)
- `data` references a ScalewayLoadBalancer resource (DAG edge 2)
- Lower TTL (300s) for infrastructure-bound records that may change during deployments

---

## CNAME to Kapsule Cluster Endpoint

Point a subdomain to a Kapsule cluster's wildcard DNS endpoint.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsRecord
metadata:
  name: k8s-cname
  org: my-org
  env: production
spec:
  zone_name:
    value: "example.com"
  name: "k8s"
  type: CNAME
  data:
    valueFrom:
      kind: ScalewayKapsuleCluster
      name: my-cluster
      fieldPath: status.outputs.wildcard_dns
  ttl: 3600
```

**Note:** CNAME targets should include a trailing dot in literal values (e.g., `"target.example.com."`), but when using `valueFrom` references, the value comes directly from the upstream resource's output.

---

## MX Records for Email

Route email to Google Workspace. Multiple MX records use separate ScalewayDnsRecord resources with different priorities.

```yaml
# Primary mail server
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsRecord
metadata:
  name: mx-primary
  org: my-org
  env: production
spec:
  zone_name:
    value: "example.com"
  name: ""
  type: MX
  data:
    value: "aspmx.l.google.com."
  ttl: 3600
  priority: 1
---
# Backup mail server
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsRecord
metadata:
  name: mx-backup
  org: my-org
  env: production
spec:
  zone_name:
    value: "example.com"
  name: ""
  type: MX
  data:
    value: "alt1.aspmx.l.google.com."
  ttl: 3600
  priority: 5
```

**Note:** Empty `name` targets the zone apex (root domain). Lower `priority` values are preferred.

---

## TXT Record for SPF

Add email authentication via SPF.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsRecord
metadata:
  name: spf-record
  org: my-org
  env: production
spec:
  zone_name:
    value: "example.com"
  name: ""
  type: TXT
  data:
    value: "v=spf1 include:_spf.google.com ~all"
  ttl: 3600
```

---

## CAA Record for Certificate Security

Restrict which Certificate Authorities can issue certificates. In Scaleway, CAA data includes flags, tag, and value in a single string.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsRecord
metadata:
  name: caa-letsencrypt
  org: my-org
  env: production
spec:
  zone_name:
    value: "example.com"
  name: ""
  type: CAA
  data:
    value: '0 issue "letsencrypt.org"'
  ttl: 86400
```

**Data format:** `"flags tag value"` where:
- `flags`: 0 (non-critical) or 128 (critical)
- `tag`: `issue`, `issuewild`, or `iodef`
- `value`: CA domain or reporting URL

---

## SRV Record

Service locator record for SIP. In Scaleway, SRV data includes weight, port, and target in the `data` field, while `priority` is a separate field.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsRecord
metadata:
  name: sip-srv
  org: my-org
  env: production
spec:
  zone_name:
    value: "example.com"
  name: "_sip._tcp"
  type: SRV
  data:
    value: "10 5060 sipserver.example.com."
  ttl: 3600
  priority: 10
```

**Data format:** `"weight port target"` with `priority` as a separate field.

---

## TLSA Record for DANE

DNS-based Authentication of Named Entities. Associates a TLS certificate with a domain name for SMTP security.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsRecord
metadata:
  name: smtp-tlsa
  org: my-org
  env: production
spec:
  zone_name:
    value: "example.com"
  name: "_25._tcp.mail"
  type: TLSA
  data:
    value: "3 1 1 a]b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"
  ttl: 3600
```

**Data format:** `"usage selector matching-type certificate-data"` where:
- `usage`: 0-3 (3 = DANE-EE)
- `selector`: 0 (full cert) or 1 (subject public key)
- `matching-type`: 0 (exact), 1 (SHA-256), 2 (SHA-512)
- `certificate-data`: hex-encoded hash or certificate data

---

## DNAME Record for Subtree Delegation

Redirect an entire DNS subtree to another domain. For example, redirect all queries for `*.old.example.com` to `*.new.example.com`.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsRecord
metadata:
  name: old-to-new-dname
  org: my-org
  env: production
spec:
  zone_name:
    value: "example.com"
  name: "old"
  type: DNAME
  data:
    value: "new.example.com."
  ttl: 3600
```

**What this does:** Any query for `x.old.example.com` is transparently rewritten to `x.new.example.com`. Useful for domain migrations and organizational restructuring.

---

## Infra Chart Integration

When composing ScalewayDnsRecord with other resources in infra charts, the record creates explicit dependency edges via `valueFrom` references.

### Kapsule Environment Chart Pattern

In a `kapsule-environment` infra chart template:

```yaml
# Zone (DAG Layer 0)
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsZone
metadata:
  name: "{{ values.env }}-zone"
spec:
  domain: "{{ values.domain }}"
  subdomain: "{{ values.env }}"
---
# Kapsule cluster (DAG Layer 2)
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayKapsuleCluster
metadata:
  name: "{{ values.env }}-cluster"
spec:
  private_network_id:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: "{{ values.env }}-network"
      fieldPath: status.outputs.private_network_id
  # ... cluster config ...
---
# DNS record pointing to cluster (DAG Layer 3)
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsRecord
metadata:
  name: "{{ values.env }}-app-dns"
spec:
  zone_name:
    valueFrom:
      kind: ScalewayDnsZone
      name: "{{ values.env }}-zone"
      fieldPath: status.outputs.zone_name
  name: "app"
  type: A
  data:
    valueFrom:
      kind: ScalewayKapsuleCluster
      name: "{{ values.env }}-cluster"
      fieldPath: status.outputs.apiserver_url
  ttl: 300
```

This creates a DAG: Zone -> Cluster -> DNS Record, deployed in topological order.

---

## Tips and Best Practices

### TTL Strategy

- **300s (5 min)** -- Infrastructure-bound records (LB IPs, cluster endpoints) that may change during deployments
- **3600s (1 hour)** -- Default for most records
- **86400s (24 hours)** -- Static records (CAA, NS delegation)

Lower TTLs **before** making critical DNS changes, then raise them after confirming success.

### Record Naming

- Use empty string `""` for the zone apex (root domain)
- Fully qualify targets with a trailing dot: `"example.com."` not `"example.com"`
- Use lowercase for record names
- For SRV: `"_service._protocol"` (e.g., `"_sip._tcp"`)
- For TLSA: `"_port._protocol.host"` (e.g., `"_25._tcp.mail"`)

### When to Use Standalone vs Inline Records

**Standalone ScalewayDnsRecord (this kind):**
- Records pointing to dynamically provisioned infrastructure (LB IPs, cluster endpoints)
- Records that need independent lifecycle management
- Records managed by different teams or environments
- Records needing explicit DAG edges in infra charts

**Inline ScalewayDnsZone records:**
- MX records for email
- TXT records for SPF, DKIM, DMARC, domain verification
- CAA records for certificate security
- Static A/CNAME records with known, stable values

### Testing

After applying DNS changes:

```bash
# Query Scaleway nameservers directly
dig @ns0.dom.scw.cloud www.example.com A

# Check MX records
dig @ns0.dom.scw.cloud example.com MX

# Verify TLSA records
dig @ns0.dom.scw.cloud _25._tcp.mail.example.com TLSA

# Get the FQDN output
openmcf stack-outputs --manifest dns-record.yaml --field fqdn
```

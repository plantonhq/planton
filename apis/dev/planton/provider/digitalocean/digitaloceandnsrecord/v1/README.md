# DigitalOcean DNS Record

Provision and manage individual DNS records in DigitalOcean domains using Planton's unified API.

## Overview

DigitalOcean DNS provides a simple and reliable managed DNS service for your domains. This component enables you to create individual DNS records within a DigitalOcean-managed domain (DNS zone).

DNS records are the fundamental building blocks of your domain's DNS configuration—mapping hostnames to IP addresses (A/AAAA records), creating aliases (CNAME), routing email (MX), and storing verification data (TXT).

This component provides a clean, protobuf-defined API for provisioning DNS records, following the **80/20 principle**: exposing only the essential configuration fields that cover the most common use cases.

## Key Features

- **All Major Record Types**: Support for A, AAAA, CNAME, MX, TXT, SRV, NS, and CAA records
- **TTL Control**: Configurable TTL (30-86400 seconds, default 1800)
- **Priority Support**: Required for MX and SRV records
- **SRV Fields**: Full support for weight, port configuration
- **CAA Fields**: Flags and tag support for certificate authority authorization
- **Validation**: Built-in validation for record types, TTL ranges, and cross-field rules

## Prerequisites

1. **DigitalOcean Domain**: An existing domain registered in DigitalOcean DNS
2. **API Token**: DigitalOcean API token with write permissions
3. **Planton CLI**: Install from [planton.dev](https://planton.dev)

## Quick Start

### A Record (IPv4)

Point a subdomain to an IPv4 address:

```yaml
apiVersion: digital-ocean.planton.dev/v1
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
```

Deploy:

```bash
planton apply -f record.yaml
```

### Reference to DNS Zone Resource

You can reference a DigitalOceanDnsZone resource for the domain:

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanDnsRecord
metadata:
  name: www-a-record
spec:
  domain:
    value_from:
      kind: DigitalOceanDnsZone
      name: my-dns-zone
      field_path: "status.outputs.zone_name"
  name: "www"
  type: A
  value:
    value: "192.0.2.1"
```

### CNAME Record

Create an alias to another hostname:

```yaml
apiVersion: digital-ocean.planton.dev/v1
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

### MX Record (Email)

Route email to your mail server:

```yaml
apiVersion: digital-ocean.planton.dev/v1
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

### TXT Record (SPF, DKIM, Verification)

Add SPF for email authentication:

```yaml
apiVersion: digital-ocean.planton.dev/v1
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

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `domain` | StringValueOrRef | DigitalOcean domain name (can be direct value or reference to DigitalOceanDnsZone) |
| `name` | string | Record name (e.g., "www", "@" for root, "api") |
| `type` | enum | DNS record type: A, AAAA, CNAME, MX, TXT, SRV, NS, CAA |
| `value` | StringValueOrRef | Record value (can be direct value or reference to another resource) |

### StringValueOrRef Format

The `domain` and `value` fields support either direct values or references to other resources:

**Direct value:**
```yaml
domain:
  value: "example.com"
```

**Reference to another resource:**
```yaml
domain:
  value_from:
    kind: DigitalOceanDnsZone
    name: my-dns-zone
    field_path: "status.outputs.zone_name"
```

### Optional Fields

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `ttl_seconds` | int32 | Time to live (30-86400 seconds) | 1800 |
| `priority` | int32 | Priority for MX/SRV records (0-65535) | 0 |
| `weight` | int32 | Weight for SRV records (0-65535) | 0 |
| `port` | int32 | Port for SRV records (0-65535) | 0 |
| `flags` | int32 | Flags for CAA records (0-255) | 0 |
| `tag` | string | Tag for CAA records (issue, issuewild, iodef) | "" |

### Record Types

| Type | Description | Example Value |
|------|-------------|---------------|
| **A** | IPv4 address | `192.0.2.1` |
| **AAAA** | IPv6 address | `2001:db8::1` |
| **CNAME** | Canonical name (alias) | `www.example.com` |
| **MX** | Mail exchange | `mail.example.com` |
| **TXT** | Text record | `v=spf1 include:...` |
| **SRV** | Service locator | `sipserver.example.com` |
| **NS** | Nameserver | `ns1.example.com` |
| **CAA** | Certificate Authority Authorization | `letsencrypt.org` |

## Outputs

After deployment, the following outputs are available:

- `record_id`: Unique identifier of the created DNS record
- `hostname`: Fully qualified hostname (e.g., "www.example.com")
- `record_type`: The DNS record type that was created
- `domain`: The domain name where the record was created
- `ttl_seconds`: The TTL applied to the record

Access outputs:

```bash
planton output record_id
planton output hostname
```

## Common Use Cases

### 1. Web Server

```yaml
spec:
  domain: "example.com"
  name: "www"
  type: A
  value: "192.0.2.1"
  ttl_seconds: 3600
```

### 2. API Endpoint

```yaml
spec:
  domain: "example.com"
  name: "api"
  type: CNAME
  value: "api-lb.example.com"
```

### 3. Root Domain (Apex)

```yaml
spec:
  domain: "example.com"
  name: "@"
  type: A
  value: "192.0.2.1"
```

### 4. SRV Record for SIP

```yaml
spec:
  domain: "example.com"
  name: "_sip._tcp"
  type: SRV
  value: "sipserver.example.com"
  priority: 10
  weight: 5
  port: 5060
```

### 5. CAA Record

Restrict certificate issuance:

```yaml
spec:
  domain: "example.com"
  name: "@"
  type: CAA
  value: "letsencrypt.org"
  flags: 0
  tag: "issue"
```

## Best Practices

1. **Use Appropriate TTLs**: Lower TTLs (60-300) for records that change frequently, higher (3600-86400) for stable records
2. **MX Priority Ordering**: Use 10, 20, 30 for primary, secondary, tertiary mail servers
3. **SPF Records**: Limit to one TXT record per domain with SPF
4. **Test Before Production**: Verify records resolve correctly with `dig` before relying on them

## Testing Records

```bash
# Test A record
dig @ns1.digitalocean.com www.example.com A

# Check MX records
dig @ns1.digitalocean.com example.com MX

# Verify TXT records
dig @ns1.digitalocean.com example.com TXT
```

## Examples

For detailed usage examples, see [examples.md](examples.md).

## Architecture Details

For in-depth architectural guidance and production best practices, see [docs/README.md](docs/README.md).

## Terraform and Pulumi

This component supports both Pulumi (default) and Terraform:

- **Pulumi**: `iac/pulumi/` - Go-based implementation
- **Terraform**: `iac/tf/` - HCL-based implementation

Both produce identical infrastructure. Choose based on your team's preference.

## Support

- **Documentation**: [docs/README.md](docs/README.md)
- **DigitalOcean DNS Docs**: [docs.digitalocean.com/products/networking/dns](https://docs.digitalocean.com/products/networking/dns)
- **Planton**: [planton.dev](https://planton.dev)

## License

This component is part of Planton and follows the same license.

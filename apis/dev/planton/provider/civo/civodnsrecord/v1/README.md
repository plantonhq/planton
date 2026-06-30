# Civo DNS Record

Provision and manage individual DNS records in Civo DNS zones using Planton's unified API.

## Overview

Civo DNS provides authoritative DNS services for domains managed through the Civo cloud platform. This component enables you to create individual DNS records within a Civo-managed DNS zone.

DNS records are the fundamental building blocks of your domain's DNS configuration—mapping hostnames to IP addresses (A/AAAA records), creating aliases (CNAME), routing email (MX), and storing verification data (TXT).

This component provides a clean, protobuf-defined API for provisioning DNS records, following the **80/20 principle**: exposing only the essential configuration fields that cover the most common use cases.

## Key Features

- **All Major Record Types**: Support for A, AAAA, CNAME, MX, TXT, SRV, and NS records
- **TTL Control**: Custom TTL settings (defaults to 3600 seconds)
- **Priority Support**: Required for MX records, optional for SRV
- **Validation**: Built-in validation for record types, TTL ranges, and cross-field rules
- **Zone Reference**: Link records to existing Civo DNS zones

## Prerequisites

1. **Civo DNS Zone**: An existing zone where records will be created (use CivoDnsZone component)
2. **Zone ID**: The Civo Zone ID (from CivoDnsZone outputs)
3. **API Key**: Civo API key with DNS management permissions
4. **Planton CLI**: Install from [planton.dev](https://planton.dev)

## Quick Start

### A Record (IPv4)

Point a subdomain to an IPv4 address:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoDnsRecord
metadata:
  name: www-a-record
spec:
  zone_id: "your-zone-id-here"
  name: "www"
  type: A
  value: "192.0.2.1"
```

Deploy:

```bash
planton apply -f record.yaml
```

### CNAME Record

Create an alias to another hostname:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoDnsRecord
metadata:
  name: app-cname
spec:
  zone_id: "your-zone-id-here"
  name: "app"
  type: CNAME
  value: "www.example.com"
```

### MX Record (Email)

Route email to your mail server:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoDnsRecord
metadata:
  name: mx-primary
spec:
  zone_id: "your-zone-id-here"
  name: "@"
  type: MX
  value: "mail.example.com"
  priority: 10
```

### TXT Record (SPF, DKIM, Verification)

Add SPF for email authentication:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoDnsRecord
metadata:
  name: spf-record
spec:
  zone_id: "your-zone-id-here"
  name: "@"
  type: TXT
  value: "v=spf1 include:_spf.google.com ~all"
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `zone_id` | string | Civo Zone ID where the record will be created |
| `name` | string | Record name (e.g., "www", "@" for root, "api") |
| `type` | enum | DNS record type: A, AAAA, CNAME, MX, TXT, SRV, NS |
| `value` | string | Record value (IP address, hostname, or text) |

### Optional Fields

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `ttl` | int32 | Time to live: 60-86400 seconds | 3600 |
| `priority` | int32 | Priority for MX/SRV records (0-65535) | 0 |

### Record Types

| Type | Description | Example Value |
|------|-------------|---------------|
| **A** | IPv4 address | `192.0.2.1` |
| **AAAA** | IPv6 address | `2001:db8::1` |
| **CNAME** | Canonical name (alias) | `www.example.com` |
| **MX** | Mail exchange | `mail.example.com` |
| **TXT** | Text record | `v=spf1 include:...` |
| **SRV** | Service locator | `0 5 5269 xmpp.example.com` |
| **NS** | Nameserver | `ns1.example.com` |

## Outputs

After deployment, the following outputs are available:

- `record_id`: Unique identifier of the created DNS record
- `hostname`: Fully qualified hostname (e.g., "www.example.com")
- `record_type`: The DNS record type that was created
- `account_id`: Civo account where the record was created

Access outputs:

```bash
planton output record_id
planton output hostname
```

## Common Use Cases

### 1. Web Server

```yaml
spec:
  zone_id: "zone-123"
  name: "www"
  type: A
  value: "192.0.2.1"
  ttl: 3600
```

### 2. API Endpoint

```yaml
spec:
  zone_id: "zone-123"
  name: "api"
  type: CNAME
  value: "api-lb.example.com"
```

### 3. Root Domain (Apex)

```yaml
spec:
  zone_id: "zone-123"
  name: "@"
  type: A
  value: "192.0.2.1"
```

### 4. Email Configuration

Primary MX record:
```yaml
spec:
  zone_id: "zone-123"
  name: "@"
  type: MX
  value: "mail.example.com"
  priority: 10
```

Backup MX record:
```yaml
spec:
  zone_id: "zone-123"
  name: "@"
  type: MX
  value: "backup.mail.example.com"
  priority: 20
```

### 5. Domain Verification

```yaml
spec:
  zone_id: "zone-123"
  name: "@"
  type: TXT
  value: "google-site-verification=abc123..."
```

## Best Practices

1. **Use Appropriate TTLs**: Lower TTLs (60-300) for records that change frequently, higher (3600+) for stable records
2. **MX Priority Ordering**: Use 10, 20, 30 for primary, secondary, tertiary mail servers
3. **SPF Records**: Limit to one TXT record per domain with SPF
4. **Test Before Production**: Verify records resolve correctly with `dig` before DNS propagation

## Testing Records

Before relying on DNS propagation:

```bash
# Test A record
dig www.example.com A

# Check MX records
dig example.com MX

# Verify TXT records
dig example.com TXT
```

## Troubleshooting

### "Record Already Exists" Error
The zone may have a conflicting record. Check existing records or import them.

### TTL Not Updating
DNS propagation can take up to the previous TTL value to complete.

### Email Not Working After Adding MX
Ensure SPF/DKIM/DMARC TXT records are also configured.

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
- **Civo DNS Docs**: [civo.com/docs/dns](https://www.civo.com/docs/dns)
- **Planton**: [planton.dev](https://planton.dev)

## License

This component is part of Planton and follows the same license.

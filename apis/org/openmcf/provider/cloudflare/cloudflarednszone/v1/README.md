# Cloudflare DNS Zone

Provision and manage Cloudflare DNS zones using OpenMCF's unified API.

## Overview

Cloudflare DNS provides authoritative DNS served from 330+ global locations via native anycast, with built-in DDoS protection, zero per-query charges (even on the free plan), and optional integrated CDN/WAF/proxy capabilities. When you create a DNS zone on Cloudflare, you're not just getting nameservers—you're getting a comprehensive edge platform.

This component provides a clean, protobuf-defined API for provisioning Cloudflare DNS zones, following the **80/20 principle**: exposing only the five essential configuration fields that 80% of users need.

## Key Features

- **Global Anycast DNS**: Authoritative DNS from 330+ locations worldwide
- **Zero Per-Query Charges**: Even on the free plan, unlimited DNS queries
- **Built-in DDoS Protection**: Attacks blocked at the edge
- **Optional Proxy/CDN**: Orange-cloud records for integrated CDN and WAF
- **Multiple Plan Tiers**: Free, Pro, Business, and Enterprise options
- **Simple Configuration**: Just `zone_name` and `account_id` to get started

## Prerequisites

1. **Cloudflare Account**: Active Cloudflare account
2. **API Token**: Cloudflare API token with Zone:Edit permissions
3. **OpenMCF CLI**: Install from [openmcf.org](https://openmcf.org)

## Quick Start

### Minimal Configuration

Create a DNS zone with the bare minimum configuration:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareDnsZone
metadata:
  name: my-zone
spec:
  zone_name: "example.com"
  account_id: "your-cloudflare-account-id"
```

Deploy:

```bash
planton apply -f zone.yaml
```

### With Plan Selection

Specify a plan tier:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareDnsZone
metadata:
  name: pro-zone
spec:
  zone_name: "example.com"
  account_id: "your-cloudflare-account-id"
  plan: PRO  # FREE, PRO, BUSINESS, or ENTERPRISE
```

### With Default Proxying

Enable orange-cloud by default for new DNS records:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareDnsZone
metadata:
  name: proxied-zone
spec:
  zone_name: "example.com"
  account_id: "your-cloudflare-account-id"
  default_proxied: true  # New records default to orange-cloud
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `zone_name` | string | Fully qualified domain name (e.g., "example.com") - Required |
| `account_id` | string | Cloudflare account ID - Required |

### Optional Fields

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `plan` | enum | Plan tier: FREE, PRO, BUSINESS, ENTERPRISE | FREE |
| `paused` | bool | If true, zone is DNS-only (no proxy/CDN/WAF) | false |
| `default_proxied` | bool | If true, new DNS records default to orange-cloud | false |

### Plan Tiers

| Plan | Features | Use Case |
|------|----------|----------|
| **FREE** | Basic CDN, Universal SSL, DDoS protection | Personal sites, small projects |
| **PRO** | Advanced SSL, WAF, Polish image optimization | Small businesses, blogs |
| **BUSINESS** | Custom SSL, SLA, Advanced DDoS, 24/7 support | E-commerce, professional sites |
| **ENTERPRISE** | Dedicated support, custom features, advanced security | Large organizations |

## Outputs

After deployment, the following outputs are available:

- `zone_id`: The unique identifier of the created zone
- `nameservers`: The Cloudflare nameservers assigned to this zone

Access outputs:

```bash
planton output zone_id
planton output nameservers
```

## Nameserver Configuration

After creating the zone, you must update your domain's nameservers at your registrar to point to the Cloudflare nameservers returned in the outputs.

**Example nameservers**:
- `gina.ns.cloudflare.com`
- `walt.ns.cloudflare.com`

**Steps**:
1. Deploy the zone: `planton apply -f zone.yaml`
2. Get nameservers: `planton output nameservers`
3. Log into your domain registrar (GoDaddy, Namecheap, etc.)
4. Update nameservers to the ones provided by Cloudflare
5. Wait for DNS propagation (typically 1-24 hours)

## Orange Cloud vs Grey Cloud

Cloudflare's unique feature is the **proxy toggle** for DNS records:

- **Orange Cloud (Proxied)**: Traffic flows through Cloudflare's edge network
  - Enables: CDN caching, WAF, DDoS protection, SSL/TLS termination
  - Use for: Web services (`www`, `app`, `api`)
  
- **Grey Cloud (DNS-Only)**: DNS returns your origin IP, clients connect directly
  - No proxy/CDN features, pure DNS resolution
  - Use for: Email (MX records), SSH, VPN, non-HTTP services

**Set `default_proxied: true`** to make new records orange-cloud by default (safer for protecting origin IPs).

## Common Use Cases

### 1. Personal Website (Free Plan)

```yaml
spec:
  zone_name: "myblog.com"
  account_id: "abc123..."
  plan: FREE
  default_proxied: true  # Protect origin IP
```

### 2. Business Website (Pro Plan)

```yaml
spec:
  zone_name: "mybusiness.com"
  account_id: "abc123..."
  plan: PRO
  default_proxied: true
```

### 3. Development Zone (Paused)

Create a zone in paused state (DNS-only, no proxy):

```yaml
spec:
  zone_name: "dev.example.com"
  account_id: "abc123..."
  paused: true  # DNS-only, no CDN/WAF
```

### 4. Enterprise Setup

```yaml
spec:
  zone_name: "enterprise.com"
  account_id: "xyz789..."
  plan: ENTERPRISE
  default_proxied: false  # Explicit control per record
```

## Best Practices

1. **Use Descriptive Names**: Name zones clearly: `my-app-prod-zone`, `staging-zone`
2. **Choose Plans Wisely**: Start with Free, upgrade to Pro for advanced features
3. **Enable Default Proxying**: Set `default_proxied: true` to protect origin IPs by default
4. **Separate Environments**: Create distinct zones for dev/staging/prod
5. **Version Control Configs**: Store zone manifests in git alongside application code
6. **Lower TTLs Before Migration**: If migrating from another DNS provider, lower TTLs 24 hours before to speed up propagation

## Migration from Another DNS Provider

### Pre-Migration Checklist

1. **Export existing DNS records** from your current provider
2. **Lower TTLs** on all records 24-48 hours before migration
3. **Note current nameservers** in case rollback is needed
4. **Backup DNS configuration** (screenshots or exports)

### Migration Steps

1. **Create Cloudflare Zone**:
   ```bash
   planton apply -f zone.yaml
   ```

2. **Get Cloudflare Nameservers**:
   ```bash
   planton output nameservers
   ```

3. **Add DNS Records** (via Terraform, Pulumi, or dashboard)

4. **Test DNS Before Going Live**:
   ```bash
   dig @gina.ns.cloudflare.com example.com A
   ```

5. **Update Nameservers** at your registrar

6. **Wait for Propagation** (check with `dig example.com NS`)

7. **Verify All Records** are resolving correctly

### Rollback Plan

If something goes wrong:
- Change nameservers back to the old provider at your registrar
- Cloudflare zone remains intact (no data loss)
- Fix issues in Cloudflare, then attempt migration again

**Pro tip**: Keep the old DNS provider's zone intact for 7 days after migration as a safety net.

## Zone Settings

After zone creation, you can configure additional settings:
- **SSL/TLS Mode**: Full (Strict) recommended for production
- **Always Use HTTPS**: Redirect HTTP to HTTPS
- **DNSSEC**: Enable for security (one-click in dashboard)
- **Caching Rules**: Configure cache behavior
- **WAF Rules**: Set up Web Application Firewall

These settings are managed separately via Cloudflare dashboard, the per-setting Terraform `cloudflare_zone_setting` resource, or Pulumi.

## Troubleshooting

### "Zone Already Exists" Error

The zone name must be unique within your Cloudflare account. If you get this error:
- Check if the zone already exists in your Cloudflare dashboard
- Import the existing zone: `planton import cloudflare-dns-zone <zone-id>`
- Or choose a different zone name

### DNS Not Resolving After Nameserver Update

DNS propagation can take up to 24-48 hours. Check propagation status:
```bash
dig example.com NS
```

If nameservers aren't updated, verify you changed them at your domain registrar (not just in Cloudflare).

### Can't Add Records to Zone

Ensure the zone is not paused. Paused zones accept DNS records but don't serve them until unpaused.

### DNSSEC Issues

If enabling DNSSEC:
1. Enable DNSSEC in Cloudflare dashboard
2. Copy DS record from Cloudflare
3. Add DS record to your domain registrar
4. Wait for propagation (can take 24 hours)

## Examples

For detailed usage examples, see [examples.md](examples.md).

## Architecture Details

For in-depth architectural guidance, deployment methods comparison, and production best practices, see [docs/README.md](docs/README.md).

## Terraform and Pulumi

This component supports both Pulumi (default) and Terraform:

- **Pulumi**: `iac/pulumi/` - Go-based implementation
- **Terraform**: `iac/tf/` - HCL-based implementation

Both produce identical infrastructure. Choose based on your team's preference.

## Support

- **Documentation**: [docs/README.md](docs/README.md)
- **Cloudflare DNS Docs**: [developers.cloudflare.com/dns](https://developers.cloudflare.com/dns)
- **OpenMCF**: [openmcf.org](https://openmcf.org)

## License

This component is part of OpenMCF and follows the same license.


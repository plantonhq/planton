# Civo DNS Zone

Manage DNS zones and records on Civo using a declarative, infrastructure-as-code approach.

## Overview

The `CivoDnsZone` resource allows you to create and manage DNS zones (domains) on Civo's built-in DNS service. This component provides a clean, minimal API focused on the essential DNS configuration that covers 80% of real-world use cases.

**Key features:**

- **Declarative configuration** - Define your DNS setup in YAML
- **Multiple record types** - A, AAAA, CNAME, MX, TXT, SRV support
- **Multi-value records** - Support for round-robin DNS and load distribution
- **Production-ready** - Validated with comprehensive test coverage
- **Free service** - Included with your Civo account at no extra cost

## Prerequisites

- Civo account with API access
- Civo API token ([get one here](https://dashboard.civo.com/security))
- Planton CLI installed
- Domain registered and ready to use

## Quick Start

### 1. Basic Web Hosting

Create a simple DNS zone for hosting a website:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoDnsZone
metadata:
  name: example-zone
spec:
  domainName: example.com
  records:
    - name: "@"
      type: A
      values:
        - value: "198.51.100.42"
      ttlSeconds: 3600
    - name: "www"
      type: CNAME
      values:
        - value: "example.com"
      ttlSeconds: 3600
```

This configures:
- Root domain (`example.com`) pointing to your server IP
- `www` subdomain as an alias to the root domain

### 2. Email Configuration

Set up email with MX and SPF records:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoDnsZone
metadata:
  name: email-zone
spec:
  domainName: example.com
  records:
    - name: "@"
      type: MX
      values:
        - value: "10 mail.example.com"
      ttlSeconds: 3600
    - name: "mail"
      type: A
      values:
        - value: "198.51.100.50"
      ttlSeconds: 3600
    - name: "@"
      type: TXT
      values:
        - value: "v=spf1 mx ~all"
      ttlSeconds: 3600
```

### 3. Load-Balanced API

Distribute traffic across multiple servers with round-robin DNS:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoDnsZone
metadata:
  name: api-zone
spec:
  domainName: example.com
  records:
    - name: "api"
      type: A
      values:
        - value: "203.0.113.10"
        - value: "203.0.113.11"
        - value: "203.0.113.12"
      ttlSeconds: 3600
```

## Deploy with Planton CLI

```bash
# Create the DNS zone
planton apply -f dns-zone.yaml

# Check status
planton get civodnszones

# View outputs (zone ID, nameservers)
planton outputs civodnszones/example-zone
```

## Important: Update Nameservers

After creating your DNS zone, you **must** update your domain's nameservers at your registrar to:

```
ns0.civo.com
ns1.civo.com
ns2.civo.com
```

Until you do this, your DNS records won't resolve. This typically takes 1-48 hours to propagate.

## Configuration Reference

### Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `domainName` | string | Yes | Fully-qualified domain name (e.g., `example.com`) |
| `records` | array | No | List of DNS records to create in the zone |

### Record Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | Yes | - | Record name (use `@` for root domain) |
| `type` | enum | Yes | - | Record type: A, AAAA, CNAME, MX, TXT, SRV |
| `values` | array | Yes | - | One or more values for the record |
| `ttlSeconds` | uint32 | No | 3600 | Time-to-live in seconds (min: 600) |

## Stack Outputs

After provisioning, the following outputs are available:

- `zone_name` - The domain name of the zone
- `zone_id` - Civo's unique zone identifier (UUID)
- `name_servers` - List of nameservers to configure at your registrar

Access outputs via:

```bash
planton outputs civodnszones/your-zone-name
```

## Common Use Cases

### Wildcard Certificate Validation

For Let's Encrypt DNS-01 challenges (required for wildcard certs):

```yaml
records:
  - name: "_acme-challenge"
    type: TXT
    values:
      - value: "ACME_validation_token_here"
    ttlSeconds: 600  # Lower TTL for faster validation
```

### CDN Integration

Point a subdomain to an external CDN:

```yaml
records:
  - name: "cdn"
    type: CNAME
    values:
      - value: "mycdn.provider.com"
    ttlSeconds: 3600
```

### Multiple Mail Servers

Configure backup MX records:

```yaml
records:
  - name: "@"
    type: MX
    values:
      - value: "10 mail1.example.com"
      - value: "20 mail2.example.com"
      - value: "30 mail3.example.com"
    ttlSeconds: 3600
```

## Best Practices

1. **Version Control** - Keep your DNS configuration in Git for audit trails and rollback capability

2. **TTL Strategy**
   - Use 3600s (1 hour) as default for most records
   - Lower to 300s (5 minutes) before planned migrations
   - Increase to 7200s+ for rarely-changed records (MX, SPF)

3. **Email Records** - For domains handling email, always configure:
   - MX records for mail routing
   - SPF (TXT) for sender validation
   - DKIM (TXT) for message signing
   - DMARC (TXT) for policy enforcement

4. **Avoid Common Mistakes**
   - Don't use CNAME at the root (`@`) - DNS protocol forbids it
   - Don't leave orphaned records when decommissioning services
   - Don't forget to update nameservers at your registrar

5. **Monitor Resolution** - Use external monitoring (UptimeRobot, Pingdom) to verify critical records resolve correctly

## Limitations

Civo DNS is designed for simplicity and covers most use cases, but has some limitations compared to advanced DNS providers:

- **No DNSSEC** - Domain signing not supported
- **No advanced routing** - No geo-routing, latency-based, or health-checked failover
- **Basic features only** - Focuses on standard record types and simple configurations

For advanced features, consider using AWS Route 53, Cloudflare DNS, or Google Cloud DNS.

## Integration with Kubernetes

### ExternalDNS

Automatically create DNS records from Kubernetes Services and Ingresses:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-app
  annotations:
    external-dns.alpha.kubernetes.io/hostname: app.example.com
spec:
  type: LoadBalancer
  # ... rest of service config
```

ExternalDNS will create the A record for `app.example.com` automatically.

### cert-manager

Use the [Civo DNS webhook](https://github.com/okteto/cert-manager-webhook-civo) for DNS-01 challenges to issue wildcard certificates:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: wildcard-cert
spec:
  secretName: wildcard-tls
  dnsNames:
    - "*.example.com"
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
```

## Troubleshooting

### DNS not resolving

1. Check nameservers are set correctly at your registrar
2. Wait up to 48 hours for propagation
3. Test with: `dig @ns0.civo.com example.com`

### Records not creating

1. Verify your Civo API token has correct permissions
2. Check domain name matches pattern: `^(?:[A-Za-z0-9-]+\.)+[A-Za-z]{2,}$`
3. Ensure record values are not empty

### Email delivery issues

1. Test MX records: `dig MX example.com`
2. Validate SPF: Use [MXToolbox SPF checker](https://mxtoolbox.com/spf.aspx)
3. Verify DKIM and DMARC records exist

## More Information

- **Deep Dive** - See [docs/README.md](docs/README.md) for comprehensive research, design decisions, and multi-cloud comparisons
- **Examples** - Check [examples.md](examples.md) for more real-world configuration scenarios
- **Pulumi Module** - See [iac/pulumi/README.md](iac/pulumi/README.md) for direct Pulumi usage
- **Civo DNS API** - [Official API documentation](https://www.civo.com/api/dns)

## Support

- Issues & Feature Requests: [Planton GitHub](https://github.com/plantonhq/planton/issues)
- Civo Support: [support@civo.com](mailto:support@civo.com)
- Community: [Planton Discord](#)


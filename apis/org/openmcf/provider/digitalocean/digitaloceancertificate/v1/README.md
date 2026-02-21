# DigitalOcean Certificate

Manage SSL/TLS certificates on DigitalOcean using a type-safe, protobuf-defined API with OpenMCF.

## Overview

**DigitalOceanCertificate** enables you to provision and manage SSL/TLS certificates on DigitalOcean through two distinct workflows:

1. **Let's Encrypt (Managed)**: Free, fully-automated certificates that auto-renew every 90 days—perfect for production workloads where DNS is managed by DigitalOcean. Supports wildcards.

2. **Custom (BYOC)**: Bring your own certificate for commercial EV certs, external DNS providers, or compliance requirements. You're responsible for renewal and rotation.

## Why Use This Component?

- **Type-Safe Configuration**: Protobuf-based API with compile-time validation prevents invalid certificate configurations
- **Discriminated Union Pattern**: The oneof constraint ensures Let's Encrypt and Custom parameters are mutually exclusive, eliminating runtime errors
- **Zero-Downtime Rotation**: Built-in lifecycle management for seamless custom certificate updates
- **Production-Ready Defaults**: Follows DigitalOcean best practices (auto-renewal enabled, complete certificate chains)

## Quick Start

### Let's Encrypt Certificate (Recommended)

For domains with DNS managed by DigitalOcean:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
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
  description: "Production certificate for web and API endpoints"
  tags:
    - env:production
    - team:platform
```

### Custom Certificate (BYOC)

For commercial certificates or external DNS providers:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanCertificate
metadata:
  name: prod-ev-cert
spec:
  certificate_name: prod-ev-cert-2025
  type: custom
  custom:
    private_key: |
      -----BEGIN PRIVATE KEY-----
      [... your private key ...]
      -----END PRIVATE KEY-----
    leaf_certificate: |
      -----BEGIN CERTIFICATE-----
      [... your certificate ...]
      -----END CERTIFICATE-----
    certificate_chain: |
      -----BEGIN CERTIFICATE-----
      [... intermediate CA chain ...]
      -----END CERTIFICATE-----
  description: "DigiCert EV certificate, expires 2026-01-15"
  tags:
    - env:production
    - cert-type:ev
```

## Key Features

### Let's Encrypt Benefits
- ✅ **Free**: No cost for DV (Domain Validated) certificates
- ✅ **Auto-Renewing**: Certificates renew 30 days before expiration
- ✅ **Wildcard Support**: Secure `*.example.com` with a single cert
- ⚠️ **Requires DigitalOcean DNS**: Uses DNS-01 challenge for domain validation

### Custom Certificate Benefits
- ✅ **Flexible**: Works with any DNS provider
- ✅ **EV/OV Support**: Upload commercial Extended/Organization Validation certs
- ✅ **Compliance-Ready**: Meet requirements for specific CAs
- ⚠️ **Manual Renewal**: You must monitor expiration and rotate before it expires

## Configuration Reference

### Core Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `certificate_name` | string | Yes | Unique identifier (1-64 chars) |
| `type` | enum | Yes | `letsEncrypt` or `custom` |
| `description` | string | No | Free-form text (max 128 chars) |
| `tags` | list | No | Organizational tags (must be unique) |

### Let's Encrypt Parameters

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `domains` | list(string) | Yes | FQDNs or wildcards (e.g., `*.example.com`) |
| `disable_auto_renew` | bool | No | Disable automatic renewal (default: false) |

**Domain Pattern**: `^(?:\*\.[A-Za-z0-9\-\.]+|[A-Za-z0-9\-\.]+\.[A-Za-z]{2,})$`

### Custom Certificate Parameters

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `leaf_certificate` | string | Yes | PEM-encoded public certificate |
| `private_key` | string | Yes | PEM-encoded private key |
| `certificate_chain` | string | No | PEM-encoded intermediate chain (highly recommended) |

**Security Note**: Always provide the complete certificate chain to prevent browser trust errors.

## Outputs

After successful provisioning, the following outputs are available:

| Output | Description |
|--------|-------------|
| `certificate_id` | DigitalOcean UUID for referencing in Load Balancers/CDN |
| `expiry_rfc3339` | Certificate expiration timestamp (RFC 3339 format) |

## Common Use Cases

### 1. Wildcard Certificate for Staging Environments

Secure all preview environments under `*.staging.example.com`:

```yaml
spec:
  certificate_name: staging-wildcard-cert
  type: letsEncrypt
  lets_encrypt:
    domains:
      - staging.example.com
      - "*.staging.example.com"
```

### 2. Multi-Domain Production Certificate

Secure apex domain and subdomains:

```yaml
spec:
  certificate_name: prod-multi-domain
  type: letsEncrypt
  lets_encrypt:
    domains:
      - example.com
      - www.example.com
      - api.example.com
      - app.example.com
```

### 3. Commercial EV Certificate (External DNS)

Upload a purchased Extended Validation certificate:

```yaml
spec:
  certificate_name: prod-ev-cert
  type: custom
  custom:
    # Reference secrets from Vault, K8s Secrets, or other secret managers
    private_key: ${SECRET_PRIVATE_KEY}
    leaf_certificate: ${SECRET_LEAF_CERT}
    certificate_chain: ${SECRET_CERT_CHAIN}
  description: "DigiCert EV, expires 2026-01-15, renew 30 days prior"
```

## Best Practices

### Security
1. **Never hardcode private keys** in YAML files or Git repositories
2. Use secret managers (Vault, K8s Secrets, AWS Secrets Manager) to inject certificate materials at runtime
3. Always provide the complete certificate chain for custom certs to prevent browser warnings

### Operations
1. **Monitor Expiration**: Set up DigitalOcean Uptime alerts (SSL Cert Expire) for custom certificates
2. **Zero-Downtime Rotation**: Leverage Terraform's `create_before_destroy` or Pulumi's dependency-aware engine
3. **Independent Monitoring**: Use SSL Labs or Prometheus + Blackbox Exporter as a second layer of defense

### Cost Optimization
1. Default to Let's Encrypt for internal/staging/production unless you specifically need:
   - Commercial EV/OV certificates (green address bar)
   - External DNS integration (Cloudflare, Route 53, etc.)
   - Compliance requirements for specific CAs

## Validation Rules

The protobuf spec enforces these constraints at compile-time:

- `certificate_name`: 1-64 characters
- `type`: Must be defined enum value (`letsEncrypt` or `custom`)
- `oneof certificate_source`: Exactly one of `lets_encrypt` or `custom` must be set
- `lets_encrypt.domains`: Required, unique, must match domain pattern
- `custom.leaf_certificate`: Required, min 1 char
- `custom.private_key`: Required, min 1 char
- `description`: Max 128 characters
- `tags`: Must be unique

## Integration

### With Load Balancers

Reference the certificate ID in DigitalOcean Load Balancer forwarding rules:

```yaml
forwarding_rule:
  entry_port: 443
  entry_protocol: https
  certificate_id: ${digitalocean_certificate.certificate_id}
```

### With Spaces CDN

Attach certificates to DigitalOcean Spaces CDN endpoints for secure object storage access.

## Troubleshooting

### Let's Encrypt Certificate Stuck in "Pending"

**Cause**: DNS is not managed by DigitalOcean, preventing DNS-01 challenge completion.

**Solution**: Either migrate DNS to DigitalOcean or use the Custom workflow (obtain a Let's Encrypt cert via `certbot` + your DNS provider's plugin, then upload it).

### Custom Certificate Shows "Untrusted" in Browsers

**Cause**: Missing or incomplete certificate chain.

**Solution**: Always provide the full intermediate chain in `certificate_chain`. Most CAs provide a `fullchain.pem` file—use that.

### Certificate Rotation Causes Brief Downtime

**Cause**: Old certificate was deleted before new one was attached to Load Balancer.

**Solution**: Use Terraform's `create_before_destroy` lifecycle block or Pulumi's automatic dependency management.

## Further Reading

- **Comprehensive Guide**: See [docs/README.md](./docs/README.md) for deep-dive coverage of deployment methods, anti-patterns, and production essentials
- **Examples**: See [examples.md](./examples.md) for copy-paste ready manifests
- **Pulumi Module**: See [iac/pulumi/README.md](./iac/pulumi/README.md) for standalone Pulumi usage
- **Terraform Module**: See [iac/tf/README.md](./iac/tf/README.md) for standalone Terraform usage

## Support

For issues, questions, or contributions, refer to the [OpenMCF documentation](https://openmcf.org) or file an issue in the repository.

---

**TL;DR**: Use Let's Encrypt for free, auto-renewing certificates when DNS is in DigitalOcean. Use Custom for commercial certs or external DNS. Always monitor expiration independently.


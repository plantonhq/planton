# CivoCertificate API

## Overview

The `CivoCertificate` API provides a declarative interface for managing TLS certificates on Civo Cloud. It supports both Let's Encrypt (automated, free) and custom (user-provided) certificates, following modern certificate management best practices.

**Important Note**: As of 2025, the Civo Pulumi/Terraform provider **does not expose certificate resources**. This API provides specification and validation but cannot provision certificates automatically. Certificates must be managed via the Civo dashboard or API until provider support is added. See: https://registry.terraform.io/providers/civo/civo/latest/docs

## API Structure

```protobuf
message CivoCertificate {
  string api_version = 1;                           // "civo.planton.dev/v1"
  string kind = 2;                                  // "CivoCertificate"
  CloudResourceMetadata metadata = 3;               // Name, labels, description
  CivoCertificateSpec spec = 4;                     // Certificate configuration
  CivoCertificateStatus status = 5;                 // Runtime outputs (when provisioned)
}
```

## Specification Fields

### `CivoCertificateSpec`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `certificate_name` | `string` | Yes | Unique identifier (1-64 chars) |
| `type` | `CivoCertificateType` | Yes | `letsEncrypt` or `custom` |
| `certificate_source` | `oneof` | Yes | Either `lets_encrypt` or `custom` params |
| `description` | `string` | No | Free-form description (≤ 128 chars) |
| `tags` | `repeated string` | No | Organizational tags (must be unique) |

### Certificate Types

#### Let's Encrypt (`letsEncrypt`)

Automated, free certificates with 90-day lifespans and auto-renewal:

**`CivoCertificateLetsEncryptParams`**:
- `domains` (required): List of FQDNs or wildcards (e.g., `["*.example.com", "example.com"]`)
  - Must match pattern: `^(?:\*\.[A-Za-z0-9\-\.]+|[A-Za-z0-9\-\.]+\.[A-Za-z]{2,})$`
  - Wildcard format: `*.domain.com`
  - Single FQDN format: `subdomain.domain.com`
  - Must be unique
- `disable_auto_renew` (optional): Set `true` to disable automatic renewal (default: `false`)

#### Custom Certificate (`custom`)

User-provided certificates for EV/OV certs or internal CAs:

**`CivoCertificateCustomParams`**:
- `leaf_certificate` (required): PEM-encoded public certificate
- `private_key` (required): PEM-encoded private key (unencrypted)
- `certificate_chain` (optional but recommended): PEM-encoded intermediate CA certificates

## Status and Outputs

### `CivoCertificateStatus`

After provisioning (when provider support is available), the `status` field contains:

```protobuf
message CivoCertificateStackOutputs {
  string certificate_id = 1;      // Unique identifier (UUID)
  string expiry_rfc3339 = 2;      // Expiration timestamp (RFC 3339 format)
}
```

## Quick Start

### Let's Encrypt - Single Domain

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoCertificate
metadata:
  name: simple-cert
spec:
  certificateName: simple-cert
  type: letsEncrypt
  letsEncrypt:
    domains:
      - example.com
```

### Let's Encrypt - Wildcard

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoCertificate
metadata:
  name: wildcard-cert
spec:
  certificateName: wildcard-cert
  type: letsEncrypt
  letsEncrypt:
    domains:
      - "*.example.com"
      - "example.com"  # Include apex domain if needed
```

### Custom Certificate

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoCertificate
metadata:
  name: custom-cert
spec:
  certificateName: custom-cert
  type: custom
  custom:
    leafCertificate: |
      -----BEGIN CERTIFICATE-----
      MIIFbjCCBFagAwIBAgIQDvO...
      (your certificate PEM)
      -----END CERTIFICATE-----
    privateKey: |
      -----BEGIN RSA PRIVATE KEY-----
      MIIEpAIBAAKCAQEAv0...
      (your private key PEM)
      -----END RSA PRIVATE KEY-----
    certificateChain: |
      -----BEGIN CERTIFICATE-----
      MIIEADCCAuigAwIBAgISA...
      (intermediate cert 1)
      -----END CERTIFICATE-----
```

## Validation Rules

### Certificate Name
- **Length**: 1-64 characters
- **Required**: Yes
- **Example**: `prod-wildcard-cert`, `api-cert-2025`

### Type
- **Values**: `letsEncrypt` (0) or `custom` (1)
- **Required**: Yes
- **Must match** the certificate_source branch chosen

### Domains (Let's Encrypt)
- **Format**: 
  - Standard FQDN: `subdomain.domain.tld`
  - Wildcard: `*.domain.tld`
- **Minimum**: 1 domain required
- **Uniqueness**: No duplicates allowed
- **Pattern validation**: Must be valid DNS name with TLD

### Custom Certificate Fields
- **Leaf Certificate**: Must be non-empty PEM format
- **Private Key**: Must be non-empty PEM format (unencrypted)
- **Certificate Chain**: Optional but highly recommended for publicly-issued certs

### Description
- **Length**: ≤ 128 characters
- **Optional**

### Tags
- **Uniqueness**: No duplicates allowed
- **Optional**

## Let's Encrypt Deep Dive

### How It Works

1. **Domain Validation**: Let's Encrypt uses ACME protocol to verify domain ownership
   - **HTTP-01 challenge**: For single domains (verifies `http://yourdomain/.well-known/acme-challenge/...`)
   - **DNS-01 challenge**: For wildcards (creates `_acme-challenge` TXT records)

2. **Requirements**:
   - DNS must point to Civo infrastructure before certificate issuance
   - For wildcards, domain must use Civo DNS or manual TXT record creation
   - Domain must be publicly accessible

3. **Auto-Renewal**:
   - Kicks in ~30 days before expiration
   - Default behavior: **enabled**
   - Can be disabled with `disable_auto_renew: true`

### Rate Limits

Let's Encrypt enforces:
- **50 certificates per domain per week**
- **5 duplicate certificates** (same domain set) per week

**Best Practice**: Use distinct subdomains for testing to avoid hitting limits.

## Custom Certificates Deep Dive

### When to Use Custom Certificates

- **Extended Validation (EV)** or **Organization Validation (OV)** certificates required
- Internal Certificate Authority for private services
- Compliance requirements mandating specific CAs
- Certificates with lifespans > 90 days

### PEM Format Requirements

**Leaf Certificate**:
```
-----BEGIN CERTIFICATE-----
MIIFbjCCBFagAwIBAgIQDvO...
(base64 encoded certificate)
-----END CERTIFICATE-----
```

**Private Key** (unencrypted):
```
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAv0...
(base64 encoded key)
-----END RSA PRIVATE KEY-----
```

**Certificate Chain** (intermediates only, no root):
```
-----BEGIN CERTIFICATE-----
(intermediate cert 1)
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
(intermediate cert 2, if applicable)
-----END CERTIFICATE-----
```

### Common Pitfalls

❌ **Missing certificate chain** → Browser trust errors  
✅ Always include intermediate certificates

❌ **Encrypted private key** → Civo rejects it  
✅ Remove passphrase: `openssl rsa -in encrypted.key -out unencrypted.key`

❌ **DER or PKCS#12 format** → Civo expects PEM  
✅ Convert: `openssl x509 -in cert.der -inform DER -out cert.pem -outform PEM`

## Security Best Practices

### Secret Management

1. **Never commit private keys to Git**
   - Use `.gitignore` for PEM files
   - Store in secret managers (Vault, AWS Secrets Manager, Pulumi secrets)

2. **Use separate certificates per environment**
   - Dev: `dev.example.com`
   - Staging: `staging.example.com`
   - Prod: `example.com`

3. **Rotate keys periodically**
   - Let's Encrypt does this automatically on renewal
   - For custom certs, rotate annually

### Monitoring Expiration

Even with auto-renewal, monitor expiration:
- Query Civo API daily for certificates expiring within 15 days
- Use external SSL monitors (SSL Labs, UptimeRobot)
- Set up alerts (PagerDuty, Slack)

**Don't rely on email**: Let's Encrypt stopped sending expiration emails in 2025.

## Current Limitations

### Provider Support

⚠️ **The Civo Pulumi/Terraform provider does not currently expose certificate resources.**

This means:
- The API specification is valid and validated
- Certificates must be managed manually via Civo dashboard or API
- Pulumi module logs warnings but cannot provision certificates
- Terraform module is non-functional

**Workarounds**:
1. **Manual provisioning**: Use Civo dashboard
2. **API automation**: Script certificate creation with `curl` and Civo API
3. **Wait for provider support**: Monitor Civo provider releases

### Tracking Provider Support

Check for updates:
- Terraform Provider: https://registry.terraform.io/providers/civo/civo/latest/docs
- Pulumi Provider: https://www.pulumi.com/registry/packages/civo/
- Civo API: https://www.civo.com/api/certificates

## Use Cases

### 1. Production HTTPS for Web App

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoCertificate
metadata:
  name: prod-web-cert
spec:
  certificateName: prod-web-cert
  type: letsEncrypt
  letsEncrypt:
    domains:
      - www.example.com
      - example.com
  description: Production website certificate
  tags:
    - env:prod
    - service:web
```

### 2. Wildcard for Microservices

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoCertificate
metadata:
  name: api-wildcard
spec:
  certificateName: api-wildcard
  type: letsEncrypt
  letsEncrypt:
    domains:
      - "*.api.example.com"
```

### 3. Internal CA Certificate

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoCertificate
metadata:
  name: internal-cert
spec:
  certificateName: internal-cert
  type: custom
  custom:
    leafCertificate: |
      -----BEGIN CERTIFICATE-----
      (from internal CA)
      -----END CERTIFICATE-----
    privateKey: |
      -----BEGIN RSA PRIVATE KEY-----
      (private key)
      -----END RSA PRIVATE KEY-----
    certificateChain: |
      -----BEGIN CERTIFICATE-----
      (internal CA intermediate)
      -----END CERTIFICATE-----
  tags:
    - internal:true
    - ca:company-pki
```

## Related Documentation

- **Examples**: See [examples.md](./examples.md) for more real-world scenarios
- **Research**: See [docs/README.md](./docs/README.md) for comprehensive certificate management guide
- **IaC Implementation**: See [iac/pulumi/](./iac/pulumi/) for module details

## Troubleshooting

### Issue: Validation fails for valid domain

**Cause**: Domain pattern doesn't match validation regex  
**Solution**: Ensure domain has TLD (e.g., `.com`, `.org`) and follows DNS naming rules

### Issue: Let's Encrypt issuance fails

**Cause**: Domain not pointing to Civo infrastructure  
**Solution**: Update DNS A/AAAA records to Civo load balancer before requesting certificate

### Issue: Custom certificate rejected

**Cause**: PEM format incorrect or private key encrypted  
**Solution**: Verify PEM headers/footers and remove key passphrase

### Issue: Browser shows "Certificate not trusted"

**Cause**: Missing intermediate certificates in chain  
**Solution**: Include full certificate chain in `certificate_chain` field

## Support

- **Planton Issues**: [GitHub Issues](https://github.com/plantonhq/planton/issues)
- **Civo Support**: [Civo Support Portal](https://www.civo.com/support)
- **Community**: [Planton Discussions](https://github.com/plantonhq/planton/discussions)

## Version History

- **v1**: Initial release with Let's Encrypt and custom certificate support


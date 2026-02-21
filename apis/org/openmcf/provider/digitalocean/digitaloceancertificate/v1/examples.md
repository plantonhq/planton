# DigitalOcean Certificate Examples

Complete, copy-paste ready YAML manifests for common certificate management scenarios.

---

## Example 1: Minimal Let's Encrypt Certificate

**Use Case**: Single domain certificate for a production website.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanCertificate
metadata:
  name: prod-web-cert
spec:
  certificateName: prod-web-cert
  type: letsEncrypt
  letsEncrypt:
    domains:
      - example.com
```

**Notes:**
- Requires DNS to be managed by DigitalOcean
- Auto-renews 30 days before expiration
- Suitable for production workloads

---

## Example 2: Multi-Domain Let's Encrypt Certificate

**Use Case**: Secure multiple subdomains (apex, www, API) with a single certificate.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanCertificate
metadata:
  name: prod-multi-domain-cert
spec:
  certificateName: prod-multi-domain-cert
  type: letsEncrypt
  letsEncrypt:
    domains:
      - example.com
      - www.example.com
      - api.example.com
      - app.example.com
  description: "Production certificate for web, API, and app endpoints"
  tags:
    - env:production
    - team:platform
    - managed:letsencrypt
```

**Notes:**
- All domains must have DNS managed by DigitalOcean
- Certificate covers up to 100 domains (DigitalOcean limit)
- Single expiration date for all domains

---

## Example 3: Wildcard Let's Encrypt Certificate

**Use Case**: Secure all subdomains under a base domain (e.g., dynamic preview environments).

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanCertificate
metadata:
  name: staging-wildcard-cert
spec:
  certificateName: staging-wildcard-cert
  type: letsEncrypt
  letsEncrypt:
    domains:
      - staging.example.com
      - "*.staging.example.com"
  description: "Wildcard cert for all staging preview environments"
  tags:
    - env:staging
    - cert-type:wildcard
```

**Notes:**
- Covers `pr-123.staging.example.com`, `feature-x.staging.example.com`, etc.
- Requires DNS-01 challenge (handled automatically by DigitalOcean)
- Perfect for CI/CD pipelines that spin up ephemeral environments

---

## Example 4: Let's Encrypt with Auto-Renewal Disabled

**Use Case**: Testing certificate expiration alerts or manual renewal workflows.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanCertificate
metadata:
  name: test-manual-renewal
spec:
  certificateName: test-manual-renewal
  type: letsEncrypt
  letsEncrypt:
    domains:
      - test.example.com
    disableAutoRenew: true
  description: "Test cert for validating expiration monitoring"
  tags:
    - env:test
    - auto-renew:disabled
```

**Notes:**
- **Not recommended for production** (you'll need to manually renew)
- Useful for testing monitoring and alerting systems
- Certificate will expire after 90 days unless renewed manually

---

## Example 5: Custom Certificate (Bring Your Own)

**Use Case**: Upload a commercial EV certificate or a cert for domains with external DNS.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanCertificate
metadata:
  name: prod-ev-cert
spec:
  certificateName: prod-ev-cert-2025
  type: custom
  custom:
    privateKey: |
      -----BEGIN PRIVATE KEY-----
      MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC7VJTUt9Us8cKj
      MzEfYyjiWA4R4/M2bS1+fWIcPm15A8d0bsI8Mz9XkPmfGQzJzPY7kCYSHl8e8tWN
      ... (truncated for brevity) ...
      -----END PRIVATE KEY-----
    leafCertificate: |
      -----BEGIN CERTIFICATE-----
      MIIDXTCCAkWgAwIBAgIJAKL0UG+mRKSzMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
      BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
      ... (truncated for brevity) ...
      -----END CERTIFICATE-----
    certificateChain: |
      -----BEGIN CERTIFICATE-----
      MIIEkjCCA3qgAwIBAgIQCgFBQgAAAVOFc2oLheynCDANBgkqhkiG9w0BAQsFADA/
      MSQwIgYDVQQKExtEaWdpdGFsIFNlY3VyaXR5IFRydXN0IENvLjEXMBUGA1UEAxMO
      ... (truncated for brevity) ...
      -----END CERTIFICATE-----
  description: "DigiCert EV certificate, expires 2026-01-15, renew 30 days prior"
  tags:
    - env:production
    - cert-type:ev
    - ca:digicert
    - renewal:manual
```

**Notes:**
- **Security**: Never commit private keys to Git. Use secret managers (Vault, K8s Secrets, AWS Secrets Manager).
- **Certificate Chain**: Always include the full intermediate chain to prevent browser trust warnings.
- **Monitoring**: Set up expiration alerts—you're responsible for renewal.

---

## Example 6: Custom Certificate with Secret References

**Use Case**: Inject certificate materials from a secret manager at runtime (secure best practice).

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanCertificate
metadata:
  name: prod-secure-cert
spec:
  certificateName: prod-secure-cert-2025
  type: custom
  custom:
    # These placeholders would be replaced by your orchestration layer
    # (e.g., Pulumi config secrets, Terraform vault data sources)
    privateKey: ${SECRET_PRIVATE_KEY}
    leafCertificate: ${SECRET_LEAF_CERT}
    certificateChain: ${SECRET_CERT_CHAIN}
  description: "Secure cert with secrets injected from Vault"
  tags:
    - env:production
    - cert-type:custom
    - secrets:vault
```

**Secret Manager Examples:**

### Pulumi (with secret encryption)
```typescript
import * as pulumi from "@pulumi/pulumi";
const config = new pulumi.Config();

const cert = new digitalocean.Certificate("prod", {
  type: "custom",
  privateKey: config.requireSecret("privateKey"),
  leafCertificate: config.requireSecret("leafCert"),
  certificateChain: config.requireSecret("certChain"),
});
```

Set secrets via CLI:
```bash
pulumi config set --secret privateKey "$(cat privkey.pem)"
pulumi config set --secret leafCert "$(cat cert.pem)"
pulumi config set --secret certChain "$(cat chain.pem)"
```

### Terraform (with Vault)
```hcl
data "vault_generic_secret" "cert" {
  path = "secret/digitalocean/prod-cert"
}

resource "digitalocean_certificate" "prod" {
  type              = "custom"
  private_key       = data.vault_generic_secret.cert.data["private_key"]
  leaf_certificate  = data.vault_generic_secret.cert.data["leaf_certificate"]
  certificate_chain = data.vault_generic_secret.cert.data["certificate_chain"]
}
```

---

## Example 7: Production-Ready Let's Encrypt with Full Metadata

**Use Case**: Complete production certificate with all optional fields for observability and governance.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanCertificate
metadata:
  name: prod-complete-cert
spec:
  certificateName: prod-complete-cert
  type: letsEncrypt
  letsEncrypt:
    domains:
      - example.com
      - www.example.com
      - api.example.com
    disableAutoRenew: false  # Explicit: auto-renewal enabled
  description: "Production certificate for web/API, auto-renews 30 days before expiry"
  tags:
    - env:production
    - team:platform
    - cost-center:engineering
    - managed:letsencrypt
    - criticality:high
    - compliance:pci-dss
```

**Benefits of Full Metadata:**
- **Cost Allocation**: Tags enable DigitalOcean billing breakdowns
- **Compliance**: Document compliance requirements (PCI-DSS, SOC 2, HIPAA)
- **Alerting**: Filter monitoring alerts by criticality or environment
- **Automation**: CI/CD pipelines can query resources by tags

---

## Example 8: Staging Environment with Wildcard and Explicit Apex

**Use Case**: Secure both the staging apex and all subdomains.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanCertificate
metadata:
  name: staging-full-wildcard
spec:
  certificateName: staging-full-wildcard
  type: letsEncrypt
  letsEncrypt:
    domains:
      - staging.example.com           # Apex (staging.example.com)
      - "*.staging.example.com"       # All subdomains (*.staging.example.com)
  description: "Staging wildcard cert for apex and all subdomains"
  tags:
    - env:staging
    - cert-type:wildcard
```

**Notes:**
- Without the apex domain, `staging.example.com` itself would **not** be covered
- Wildcard `*.staging.example.com` only covers one level (e.g., `app.staging.example.com`)
- Does **not** cover multi-level subdomains (e.g., `api.app.staging.example.com`)

---

## Common Patterns Summary

| Use Case | Type | Domains | Key Features |
|----------|------|---------|--------------|
| Single domain | letsEncrypt | `example.com` | Simplest, auto-renewing |
| Multi-domain | letsEncrypt | `example.com`, `www.example.com` | Single cert for multiple domains |
| Wildcard | letsEncrypt | `*.example.com` | Covers all subdomains |
| Commercial EV | custom | N/A (embedded in cert) | Extended validation, manual renewal |
| External DNS | custom | N/A | Cert obtained via `certbot` + DNS plugin |

---

## Validation Checklist

Before deploying, ensure:

- ✅ `certificate_name` is unique and 1-64 characters
- ✅ For Let's Encrypt: DNS is managed by DigitalOcean
- ✅ For Custom: Private key and leaf certificate are PEM-formatted
- ✅ For Custom: Certificate chain is included (prevents browser warnings)
- ✅ Tags are unique (no duplicates)
- ✅ Description is ≤ 128 characters
- ✅ Private keys are sourced from secret managers (not hardcoded)

---

## Next Steps

1. **Deploy**: Use `openmcf pulumi up` or `terraform apply` to provision the certificate
2. **Monitor**: Set up DigitalOcean Uptime alerts for SSL expiration
3. **Integrate**: Reference the certificate ID in Load Balancers or Spaces CDN
4. **Validate**: Run `openssl s_client` or SSL Labs test to verify the certificate chain

For more details, see:
- [README.md](./README.md) - Component overview and best practices
- [docs/README.md](./docs/README.md) - Comprehensive production guide
- [iac/pulumi/README.md](./iac/pulumi/README.md) - Pulumi module usage
- [iac/tf/README.md](./iac/tf/README.md) - Terraform module usage


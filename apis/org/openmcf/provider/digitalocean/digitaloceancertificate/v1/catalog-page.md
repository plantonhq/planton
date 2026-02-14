# DigitalOcean Certificate

Provisions an SSL/TLS certificate on DigitalOcean using either a fully-managed Let's Encrypt workflow or a custom (bring-your-own) certificate upload. The component implements the DigitalOcean Certificates API as a protobuf-enforced discriminated union: a `type` field selects one of two mutually exclusive parameter sets (`letsEncrypt` or `custom`), and the `oneof certificateSource` constraint in the spec makes it impossible to mix fields from the two paths. Both the Terraform and Pulumi IaC modules apply `create_before_destroy` semantics so that certificate replacements do not interrupt HTTPS traffic on attached Load Balancers or Spaces CDN endpoints.

## What Gets Created

When you deploy a DigitalOceanCertificate resource, OpenMCF provisions:

- **SSL/TLS Certificate** -- a `digitalocean_certificate` resource of type `lets_encrypt` or `custom`, depending on which branch of the `certificateSource` oneof is populated
- **Automatic DNS-01 Validation** (Let's Encrypt only) -- DigitalOcean creates the required `_acme-challenge` TXT records and handles renewal every 90 days; this path requires DNS to be managed by DigitalOcean
- **Certificate Storage** (Custom only) -- the provided PEM-encoded leaf certificate, private key, and optional intermediate chain are stored in DigitalOcean's certificate store, ready for attachment to Load Balancers or Spaces CDN

## Prerequisites

- **DigitalOcean credentials** configured via environment variables (`DIGITALOCEAN_TOKEN`) or OpenMCF provider config
- **DigitalOcean-managed DNS** for the target domain(s) if using the `letsEncrypt` path (required for the DNS-01 challenge)
- **Valid PEM-encoded certificate materials** if using the `custom` path: leaf certificate, private key, and (recommended) the intermediate certificate chain

## Quick Start

Create a file `cert.yaml`:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanCertificate
metadata:
  name: my-cert
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.DigitalOceanCertificate.my-cert
spec:
  certificateName: my-cert
  type: letsEncrypt
  letsEncrypt:
    domains:
      - example.com
```

Deploy:

```shell
openmcf apply -f cert.yaml
```

This requests a free, auto-renewing Let's Encrypt certificate for `example.com`. DigitalOcean performs DNS-01 validation automatically, and the certificate is ready to attach to a Load Balancer within seconds.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `certificateName` | `string` | Unique human-readable identifier for the certificate in DigitalOcean. | Required, 1--64 characters |
| `type` | `enum` | Certificate source type. Valid values: `letsEncrypt`, `custom`. Must match the branch set in `certificateSource`. | Required, must be a defined enum value |
| `certificateSource` | `oneof` | Exactly one of `letsEncrypt` or `custom` must be provided. | Required (protobuf `oneof` with validation) |

#### When `type` is `letsEncrypt`

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `letsEncrypt.domains` | `string[]` | FQDNs or wildcard domains (e.g., `*.example.com`) to include on the certificate. DNS for every listed domain must be managed by DigitalOcean. | Required, at least one entry, unique, must match FQDN or wildcard pattern |

#### When `type` is `custom`

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `custom.leafCertificate` | `string` | PEM-encoded public certificate (the leaf cert issued by your CA). | Required, non-empty |
| `custom.privateKey` | `string` | PEM-encoded private key corresponding to the leaf certificate. | Required, non-empty |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | `""` | Free-form description for the certificate. Useful for noting expiry dates, CA names, or ticket references. Max 128 characters. |
| `tags` | `string[]` | `[]` | Organizational tags. Must be unique and lowercase kebab-case. |
| `letsEncrypt.disableAutoRenew` | `bool` | `false` | When `true`, prevents DigitalOcean from automatically renewing the Let's Encrypt certificate before expiry. Almost never needed in production. |
| `custom.certificateChain` | `string` | `""` | PEM-encoded intermediate certificate chain. Technically optional, but omitting it causes "untrusted certificate" errors in browsers that do not have the intermediate CA cached. Always provide the full chain in production. |

## Examples

### Let's Encrypt -- Single Domain

The minimal Let's Encrypt configuration for a single domain:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanCertificate
metadata:
  name: app-cert
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.DigitalOceanCertificate.app-cert
spec:
  certificateName: app-cert
  type: letsEncrypt
  letsEncrypt:
    domains:
      - app.example.com
  tags:
    - env-production
```

### Let's Encrypt -- Multiple Domains and Wildcard (SAN)

A certificate covering the apex domain, a `www` subdomain, and a wildcard for all subdomains under `staging`:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanCertificate
metadata:
  name: multi-domain-cert
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.DigitalOceanCertificate.multi-domain-cert
spec:
  certificateName: multi-domain-cert
  type: letsEncrypt
  letsEncrypt:
    domains:
      - example.com
      - www.example.com
      - "*.staging.example.com"
  description: "Production and staging wildcard cert"
  tags:
    - env-production
    - cert-type-wildcard
```

All listed domains must have their DNS managed by DigitalOcean. Wildcard entries (e.g., `*.staging.example.com`) are validated via the same DNS-01 challenge and cover any single-level subdomain under the specified parent.

### Custom Certificate Upload

Uploading a commercially-issued certificate (e.g., an EV certificate from DigiCert) for a domain whose DNS is hosted externally:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanCertificate
metadata:
  name: ev-cert
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.DigitalOceanCertificate.ev-cert
spec:
  certificateName: ev-cert-2026
  type: custom
  custom:
    leafCertificate: |
      -----BEGIN CERTIFICATE-----
      MIIFjTCCA3Wg...
      -----END CERTIFICATE-----
    privateKey: |
      -----BEGIN PRIVATE KEY-----
      MIIEvgIBADANBg...
      -----END PRIVATE KEY-----
    certificateChain: |
      -----BEGIN CERTIFICATE-----
      MIIEtjCCA56g...
      -----END CERTIFICATE-----
  description: "DigiCert EV cert, expires 2027-01-15"
  tags:
    - env-production
    - cert-type-ev
    - ca-digicert
```

For custom certificates, store the PEM materials in a secrets manager (HashiCorp Vault, Kubernetes Secrets with External Secrets Operator, or Pulumi secret config) and inject them at apply-time. Never commit private keys to version control. The underlying IaC module uses a `create_before_destroy` lifecycle rule, so replacing an expiring custom certificate creates the new resource before destroying the old one, avoiding downtime on any attached Load Balancer.

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `certificate_id` | `string` | UUID of the certificate assigned by DigitalOcean. Use this to reference the certificate when configuring Load Balancers or Spaces CDN. |
| `expiry_rfc3339` | `string` | Expiration timestamp in RFC 3339 format (e.g., `2026-05-14T00:00:00Z`). For Let's Encrypt certificates, this updates silently on each auto-renewal. |

## Related Components

- [DigitalOceanLoadBalancer](/docs/catalog/digitalocean/digitaloceanloadbalancer) -- attach certificates to load balancers for SSL/TLS termination
- [DigitalOceanDnsZone](/docs/catalog/digitalocean/digitaloceandnszone) -- manage DNS zones in DigitalOcean, a prerequisite for Let's Encrypt certificates
- [DigitalOceanDnsRecord](/docs/catalog/digitalocean/digitaloceandnsrecord) -- manage individual DNS records under a DigitalOcean-managed domain
- [DigitalOceanBucket](/docs/catalog/digitalocean/digitaloceanbucket) -- Spaces object storage with CDN endpoints that can use certificates for custom domain HTTPS

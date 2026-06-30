---
title: "Certificate"
description: "Certificate deployment documentation"
icon: "package"
order: 100
componentName: "civocertificate"
---

# Civo Certificate

Manages TLS certificates on Civo Cloud, supporting both automated Let's Encrypt certificates and user-supplied custom certificates. The component validates the manifest and provisions the certificate resource, exposing its ID and expiry timestamp as stack outputs.

> **Provider limitation (as of 2025):** The Civo Pulumi/Terraform provider does not yet expose a certificate resource type. The module validates your specification and logs the intended configuration, but actual certificate provisioning must be performed via the Civo API or dashboard until upstream provider support is added. Tracked at the [Civo Terraform provider docs](https://registry.terraform.io/providers/civo/civo/latest/docs).

## What Gets Created

When you deploy a CivoCertificate resource, Planton processes the manifest and (once the upstream provider adds support) provisions:

- **Civo Certificate** --- a TLS certificate registered in your Civo account, either auto-managed via Let's Encrypt or uploaded as a custom PEM bundle
- **Labels** --- key-value metadata derived from `metadata.labels` and `spec.tags`, applied to the certificate for filtering and organization

## Prerequisites

- **Civo credentials** configured via environment variables or Planton provider config
- **A registered domain** (for Let's Encrypt certificates) with DNS pointing to Civo infrastructure
- **PEM-encoded certificate files** (for custom certificates) including the leaf certificate and private key

## Quick Start

Create a file `civo-certificate.yaml`:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoCertificate
metadata:
  name: my-cert
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.CivoCertificate.my-cert
spec:
  certificateName: my-cert
  type: letsEncrypt
  letsEncrypt:
    domains:
      - "example.com"
      - "www.example.com"
```

Deploy:

```shell
planton apply -f civo-certificate.yaml
```

This requests a Let's Encrypt certificate covering `example.com` and `www.example.com` with automatic renewal enabled by default.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `certificateName` | `string` | A unique, human-readable name for the certificate. | Required; 1--64 characters |
| `type` | `CivoCertificateType` | The certificate source type. Must match the branch chosen in the `certificateSource` oneof. | Required; must be `letsEncrypt` or `custom` |
| `letsEncrypt` _or_ `custom` | `object` | Exactly one of `letsEncrypt` or `custom` must be provided (mutually exclusive oneof). | Required (oneof) |

#### CivoCertificateLetsEncryptParams (when `type` is `letsEncrypt`)

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `domains` | `string[]` | FQDNs or wildcard domains to include on the certificate (e.g., `"example.com"`, `"*.example.com"`). | Required; at least one entry; entries must be unique; must match FQDN or wildcard pattern |

#### CivoCertificateCustomParams (when `type` is `custom`)

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `leafCertificate` | `string` | PEM-encoded public certificate. | Required; non-empty |
| `privateKey` | `string` | PEM-encoded private key. | Required; non-empty |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | `""` | Free-form description of the certificate. Maximum 128 characters. |
| `tags` | `string[]` | `[]` | Tags for filtering and grouping. Must be unique, lowercase kebab-case values. |
| `letsEncrypt.disableAutoRenew` | `bool` | `false` | When `true`, disables automatic renewal of the Let's Encrypt certificate. |
| `custom.certificateChain` | `string` | `""` | PEM-encoded intermediate certificate chain. Only applicable when `type` is `custom`. |

## Examples

### Let's Encrypt with Auto-Renewal

A minimal Let's Encrypt certificate for a single domain:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoCertificate
metadata:
  name: api-cert
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.CivoCertificate.api-cert
spec:
  certificateName: api-cert
  type: letsEncrypt
  letsEncrypt:
    domains:
      - "api.example.com"
  description: "TLS for public API endpoint"
  tags:
    - api
    - production
```

### Let's Encrypt Wildcard with Renewal Disabled

A wildcard certificate covering all subdomains, with automatic renewal turned off for manual control:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoCertificate
metadata:
  name: wildcard-cert
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.CivoCertificate.wildcard-cert
spec:
  certificateName: wildcard-cert
  type: letsEncrypt
  letsEncrypt:
    domains:
      - "*.staging.example.com"
      - "staging.example.com"
    disableAutoRenew: true
  description: "Wildcard cert for staging subdomains"
```

### Custom Certificate Upload

Upload an existing certificate and private key, including the intermediate chain:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoCertificate
metadata:
  name: custom-cert
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.CivoCertificate.custom-cert
spec:
  certificateName: custom-cert
  type: custom
  custom:
    leafCertificate: |
      -----BEGIN CERTIFICATE-----
      MIIFazCCA1OgAwIBAgIUE...
      -----END CERTIFICATE-----
    privateKey: |
      -----BEGIN PRIVATE KEY-----
      MIIEvgIBADANBgkqhkiG9...
      -----END PRIVATE KEY-----
    certificateChain: |
      -----BEGIN CERTIFICATE-----
      MIIFYjCCBEqgAwIBAgIQd...
      -----END CERTIFICATE-----
  description: "Enterprise CA issued certificate"
  tags:
    - enterprise
    - production
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `certificateId` | `string` | Unique identifier of the certificate, assigned by Civo |
| `expiryRfc3339` | `string` | Expiration timestamp of the certificate in RFC 3339 format |

## Related Components

- [CivoKubernetesCluster](/docs/catalog/civo/kubernetes-cluster) --- Kubernetes clusters that can reference certificates for ingress TLS termination
- [CivoFirewall](/docs/catalog/civo/firewall) --- firewall rules to restrict access to services using the certificate
- [CivoVpc](/docs/catalog/civo/vpc) --- private networks where certificate-protected services run

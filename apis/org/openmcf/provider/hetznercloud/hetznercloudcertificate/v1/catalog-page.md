# Hetzner Cloud Certificate

Deploys a TLS certificate to the Hetzner Cloud certificate store for use by load balancer HTTPS services. Supports two mutually exclusive types: **uploaded** (user-provided PEM certificate and private key) or **managed** (automatic Let's Encrypt issuance and renewal). Exactly one type must be specified per manifest.

## What Gets Created

- **Uploaded Certificate** — an `hcloud_uploaded_certificate` resource that stores a user-provided PEM certificate chain and private key. Created only when the `uploaded` variant is set.
- **Managed Certificate** — an `hcloud_managed_certificate` resource that requests a Let's Encrypt certificate for the specified domains, with automatic renewal. Created only when the `managed` variant is set.

Exactly one of the two resources is created per deployment.

## Prerequisites

- **Hetzner Cloud API token** configured via environment variable (`HCLOUD_TOKEN`) or OpenMCF provider config

For **managed** certificates:
- DNS A or AAAA records for each domain pointing to a Hetzner Cloud load balancer
- A load balancer HTTPS service referencing this certificate (for the ACME HTTP-01 challenge)

For **uploaded** certificates:
- A PEM-encoded certificate chain (server cert + intermediate CAs)
- A PEM-encoded private key matching the certificate

## Quick Start

Create a file `certificate.yaml`:

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudCertificate
metadata:
  name: my-cert
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.HetznerCloudCertificate.my-cert
spec:
  managed:
    domainNames:
      - example.com
```

Deploy:

```shell
openmcf apply -f certificate.yaml
```

This creates a managed Let's Encrypt certificate for `example.com`. The certificate's ID and metadata are available in the stack outputs.

## Configuration Reference

Exactly one of `uploaded` or `managed` must be set. Setting both or neither is a validation error.

### Required Fields (Uploaded Variant)

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `uploaded.certificate` | `string` | PEM-encoded TLS certificate chain. Server certificate first, intermediate CAs in order. Changing this value forces replacement. | min length: 1 |
| `uploaded.privateKey` | `string` | PEM-encoded private key for the certificate. Treated as sensitive (encrypted in state, masked in output). Changing this value forces replacement. | min length: 1 |

### Required Fields (Managed Variant)

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `managed.domainNames` | `string[]` | Domain names for the Let's Encrypt certificate. Hetzner Cloud issues a single SAN certificate covering all listed domains. Changing this list forces replacement. | min items: 1 |

### Optional Fields

This component has no optional fields. All fields within the selected variant are required.

## Examples

### Managed Certificate with Multiple Domains

A SAN certificate covering a root domain and subdomains. All domains must resolve to a load balancer with an HTTPS service referencing this certificate.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudCertificate
metadata:
  name: web-platform-cert
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: web-platform
    pulumi.openmcf.org/stack.name: production.HetznerCloudCertificate.web-platform-cert
spec:
  managed:
    domainNames:
      - example.com
      - www.example.com
      - app.example.com
```

### Uploaded Wildcard Certificate

A user-provided wildcard certificate and private key. Use when you need wildcard coverage, an EV/OV certificate, or a certificate from a specific CA.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudCertificate
metadata:
  name: wildcard-cert
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: web-platform
    pulumi.openmcf.org/stack.name: production.HetznerCloudCertificate.wildcard-cert
spec:
  uploaded:
    certificate: |
      -----BEGIN CERTIFICATE-----
      MIIFYDCCBEigAwIBAgISA1...
      -----END CERTIFICATE-----
      -----BEGIN CERTIFICATE-----
      MIIEdTCCA12gAwIBAgIJAN...
      -----END CERTIFICATE-----
    privateKey: |
      -----BEGIN PRIVATE KEY-----
      MIIEvgIBADANBgkqhkiG9w...
      -----END PRIVATE KEY-----
```

### Certificate with Load Balancer HTTPS Service

A managed certificate deployed alongside a load balancer that references it for HTTPS termination. The load balancer uses `valueFrom` to resolve the certificate ID automatically.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudCertificate
metadata:
  name: api-cert
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: api-platform
    pulumi.openmcf.org/stack.name: production.HetznerCloudCertificate.api-cert
spec:
  managed:
    domainNames:
      - api.example.com
```

The companion load balancer manifest references the certificate via `valueFrom`:

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudLoadBalancer
metadata:
  name: api-lb
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: api-platform
    pulumi.openmcf.org/stack.name: production.HetznerCloudLoadBalancer.api-lb
spec:
  loadBalancerType: lb11
  location: fsn1
  services:
    - protocol: https
      listenPort: 443
      destinationPort: 80
      http:
        certificateIds:
          - valueFrom:
              kind: HetznerCloudCertificate
              name: api-cert
              fieldPath: status.outputs.certificate_id
  targets:
    - type: server
      serverId:
        valueFrom:
          kind: HetznerCloudServer
          name: api-server
          fieldPath: status.outputs.server_id
```

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `certificate_id` | `string` | The Hetzner Cloud numeric ID of the created certificate. Referenced by load balancer HTTPS services. |
| `type` | `string` | Certificate type: `"uploaded"` or `"managed"`. Computed by Hetzner Cloud. |
| `fingerprint` | `string` | SHA256 fingerprint of the certificate. Computed by Hetzner Cloud. |
| `not_valid_before` | `string` | Point in time when the certificate becomes valid (ISO-8601). |
| `not_valid_after` | `string` | Point in time when the certificate stops being valid (ISO-8601). |

## Related Components

- [HetznerCloudLoadBalancer](/docs/catalog/hetznercloud/hetznercloudloadbalancer) — The primary consumer of certificates. Load balancer HTTPS services reference `certificate_id` to terminate TLS connections.

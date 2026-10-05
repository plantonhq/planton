# GCP SSL Certificate (Self-Managed) - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying a self-managed Compute Engine SSL certificate using Planton's `GcpSslCertificate` API. The module is written in Go and creates `compute.SSLCertificate` (global) or `compute.RegionSslCertificate` (regional) based on whether `spec.region` is set.

You bring the PEM certificate chain and private key; the load balancer presents them to clients. The certificate attaches to target HTTPS (and SSL) proxies exactly like a Google-managed certificate — but nothing renews itself: track `expire_time` and rotate create-before-destroy.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** with Compute Engine API enabled (the module enables it if needed)
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs
6. **PEM material**: a certificate chain (leaf first, then intermediates) and its matching unencrypted private key

## Directory Structure

```
iac/pulumi/
├── main.go                    # Pulumi program entry point
├── Pulumi.yaml                # Pulumi project configuration
├── Makefile                   # Build and deployment targets
├── README.md                  # This file
└── module/
    ├── main.go                # Module coordinator
    ├── ssl_certificate.go     # Global/regional certificate creation
    ├── locals.go              # Resolved resource + derived values
    └── outputs.go             # Output constants
```

## Quick Start

```bash
cd iac/pulumi
pulumi stack init dev
```

Provide a `iac-input.yaml`:

```yaml
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpSslCertificate
  metadata:
    name: prod-app-cert
  spec:
    project_id:
      value: my-gcp-project-123
    certificate: |
      -----BEGIN CERTIFICATE-----
      ...leaf, then intermediates...
      -----END CERTIFICATE-----
    private_key: |
      -----BEGIN PRIVATE KEY-----
      ...matching unencrypted key...
      -----END PRIVATE KEY-----
```

```bash
make build
pulumi preview
pulumi up
```

## Inputs

The module consumes `GcpSslCertificateIacInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpSslCertificate` spec (PEM chain + private key, optional region/name/description) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `self_link` | string | Self-link URI — the value a target HTTPS proxy references in `ssl_certificates` |
| `certificate_name` | string | Name of the SSL certificate in GCP |
| `certificate_id` | string | Server-assigned numeric ID of the certificate |
| `expire_time` | string | Expiry in RFC3339, parsed from the uploaded chain |
| `region` | string | Region of a regional certificate; empty for global |

## Behavior Notes

- **The private key is a Pulumi secret**: marked with `ToSecret`, encrypted in state, write-only in GCP, never exported. The certificate chain is public handshake material and is not treated as a secret.
- **Fully immutable**: every argument is ForceNew; rotation is create-before-destroy (GCP blocks deleting a certificate a proxy references).
- **One kind, two resources**: empty `region` creates the global certificate; a set region creates the regional one. Scope is permanent.
- **Shared namespace**: self-managed and Google-managed certificates share one name namespace per scope.
- **API enablement**: the module enables `compute.googleapis.com` (with `disable_on_destroy=false`) so a fresh project works first try.

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation

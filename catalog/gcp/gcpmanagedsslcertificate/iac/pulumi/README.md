# GCP Managed SSL Certificate - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying a Google-managed SSL certificate using Planton's `GcpManagedSslCertificate` API. The module is written in Go and creates `compute.ManagedSslCertificate`.

A Google-managed SSL certificate terminates TLS at a global external Application Load Balancer without you ever handling key material. Google provisions and renews the certificate once DNS for each domain points at the load balancer.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** with Compute Engine API enabled (the module enables it if needed)
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs

## Directory Structure

```
iac/pulumi/
├── main.go                    # Pulumi program entry point
├── Pulumi.yaml                # Pulumi project configuration
├── Makefile                   # Build and deployment targets
├── README.md                  # This file
└── module/
    ├── main.go                # Module coordinator
    ├── managed_ssl_certificate.go  # Certificate creation
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
  kind: GcpManagedSslCertificate
  metadata:
    name: app-cert
  spec:
    project_id:
      value: my-gcp-project-123
    domains:
      - app.example.com
    description: TLS for the production app load balancer
```

```bash
make build
pulumi preview
pulumi up
```

## Inputs

The module consumes `GcpManagedSslCertificateIacInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpManagedSslCertificate` spec (domains, optional name/description) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `self_link` | string | Self-link URI — the value a target HTTPS proxy references in `ssl_certificates` |
| `certificate_name` | string | Name of the SSL certificate in GCP |
| `certificate_id` | string | Server-assigned numeric ID of the certificate |
| `expire_time` | string | Expiry time in RFC3339 format; empty until provisioning completes |

## Behavior Notes

- **Fully immutable**: name and domains are ForceNew — any change destroys and recreates the certificate. Rotate create-before-destroy when a proxy still references the old cert.
- **DNS-gated provisioning**: the certificate object is created immediately but stays PROVISIONING until each domain's DNS points at the load balancer.
- **No wildcards**: Google-managed certificates do not support `*.` wildcard domains.
- **API enablement**: the module enables `compute.googleapis.com` (with `disable_on_destroy=false`) so a fresh project works first try.

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation

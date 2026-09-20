# GCP Target HTTPS Proxy - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying GCP Compute Engine target HTTPS proxies using Planton's `GcpTargetHttpsProxy` API. The module is written in Go and creates exactly one of `compute.TargetHttpsProxy` (global; `spec.region` empty) or `compute.RegionTargetHttpsProxy` (regional; `spec.region` set), the same switch the Terraform module makes with its count guards.

A target HTTPS proxy terminates TLS for an Application Load Balancer: it binds a forwarding rule (the VIP) to a URL map (routing) and owns certificates, SSL policy, QUIC negotiation, and TLS 1.3 early data.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** with Compute Engine API available
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs
6. **A URL map and a certificate** — a `GcpUrlMap` and (typically) a `GcpManagedSslCertificate` whose self-links you wire in

## Directory Structure

```
iac/pulumi/
├── main.go           # Pulumi program entry point
├── Pulumi.yaml       # Pulumi project configuration
├── Makefile          # Build and deployment targets
├── README.md         # This file
└── module/
    ├── main.go                 # Module coordinator
    ├── target_https_proxy.go   # Proxy creation and mapping
    ├── locals.go               # Resolved resource + derived values
    └── outputs.go              # Stack output constants
```

## Quick Start

### 1. Initialize Pulumi Stack

```bash
cd iac/pulumi
pulumi stack init dev
```

### 2. Create Input File

Provide a `stack-input.yaml` with the proxy specification:

```yaml
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpTargetHttpsProxy
  metadata:
    name: web-https-frontend
  spec:
    project_id:
      value: my-gcp-project-123
    url_map:
      value: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/global/urlMaps/web-routing
    ssl_certificates:
      - value: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/global/sslCertificates/web-cert
```

### 3. Build, Preview, and Deploy

```bash
make build
pulumi preview
pulumi up
```

### 4. View Outputs

```bash
pulumi stack output self_link
pulumi stack output proxy_name
```

## Inputs

The module consumes `GcpTargetHttpsProxyStackInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpTargetHttpsProxy` spec (URL map, certificate source, TLS behavior) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `self_link` | string | Self-link URI — the value a forwarding rule references (`regions/{region}` in place of `global` for a regional proxy) |
| `proxy_name` | string | Name of the proxy in GCP |
| `proxy_id` | string | Server-assigned numeric ID |
| `fingerprint` | string | Server-computed fingerprint (empty for a regional proxy) |
| `region` | string | Region of a regional proxy; empty for global |

## Behavior Notes

- **Certificate exclusivity**: exactly one of `ssl_certificates`, `certificate_manager_certificates`, or `certificate_map` (enforced pre-deploy by CEL).
- **In-place updates**: URL map, certificate wiring, SSL policy, server TLS policy, and QUIC override swap without downtime; `tls_early_data` and `proxy_bind` are ForceNew.
- **PROVISIONING certificates attach fine** — attachment is required before a managed certificate can activate.
- **Ambient project**: an empty `project_id` falls back to the provider's default project.
- **API enablement**: the module enables `compute.googleapis.com` before creating the proxy (`disable_on_destroy=false`).

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation

# GCP Target HTTP Proxy - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying GCP Compute Engine target HTTP proxies using Planton's `GcpTargetHttpProxy` API. The module is written in Go and creates exactly one of `compute.TargetHttpProxy` (global; `spec.region` empty) or `compute.RegionTargetHttpProxy` (regional; `spec.region` set), the same switch the Terraform module makes with its count guards.

A target HTTP proxy binds a forwarding rule (the VIP) to a URL map (the routing brain); its standard production role is serving the http→https redirect on port 80. A regional proxy's URL map and forwarding rule must be regional in the same region.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** with Compute Engine API available
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs
6. **A URL map** — a `GcpUrlMap` whose `self_link` you wire in

## Directory Structure

```
iac/pulumi/
├── main.go           # Pulumi program entry point
├── Pulumi.yaml       # Pulumi project configuration
├── Makefile          # Build and deployment targets
├── README.md         # This file
└── module/
    ├── main.go                # Module coordinator
    ├── target_http_proxy.go   # Proxy creation and mapping
    ├── locals.go              # Resolved resource + derived values
    └── outputs.go             # Stack output constants
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
  kind: GcpTargetHttpProxy
  metadata:
    name: web-http-frontend
  spec:
    project_id:
      value: my-gcp-project-123
    url_map:
      value: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/global/urlMaps/http-redirect
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

The module consumes `GcpTargetHttpProxyStackInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpTargetHttpProxy` spec (URL map, keep-alive, proxy bind) |
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

- **Mutability**: only `url_map` updates in place (`setUrlMap`); every other field is ForceNew.
- **Ambient project**: an empty `project_id` falls back to the provider's default project.
- **API enablement**: the module enables `compute.googleapis.com` before creating the proxy (`disable_on_destroy=false`).

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation

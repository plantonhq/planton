# GCP Health Check - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying GCP Compute Engine health checks using Planton's `GcpHealthCheck` API. The module is written in Go and creates exactly one of `compute.HealthCheck` (global, when `region` is empty) or `compute.RegionHealthCheck` (regional, when `region` is set).

A health check is the probe backend services consult before routing traffic and managed instance groups consult before auto-healing. One check is commonly shared by many backend services.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** — health checks are free; no billed infrastructure is created
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs

## Directory Structure

```
iac/pulumi/
├── main.go           # Pulumi program entry point
├── Pulumi.yaml       # Pulumi project configuration
├── Makefile          # Build and deployment targets
├── README.md         # This file
└── module/
    ├── main.go          # Module coordinator
    ├── health_check.go  # Global/regional health check creation
    ├── locals.go        # Resolved resource + derived values
    └── outputs.go       # Output constants
```

## Quick Start

### 1. Initialize Pulumi Stack

```bash
cd iac/pulumi
pulumi stack init dev
```

### 2. Create Input File

Provide a `iac-input.yaml` with the health check specification:

```yaml
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpHealthCheck
  metadata:
    name: web-backend-probe
  spec:
    project_id:
      value: my-gcp-project-123
    http:
      port_specification: USE_SERVING_PORT
      request_path: /healthz
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
pulumi stack output type
```

## Inputs

The module consumes `GcpHealthCheckIacInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpHealthCheck` spec (protocol block, cadence, scope, etc.) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

Spec fields: exactly one protocol block (`http`/`https`/`http2`/`tcp`/`ssl`/`grpc`/`grpc_tls`), optional `project_id` (falls back to the provider default project when empty), optional `health_check_name` (defaults to `metadata.name`), optional `region` (empty = global), cadence knobs (defaults 5/5/2/2), `enable_logging`, and global-only `source_regions`.

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `self_link` | string | Self-link URI — the value backend services reference |
| `health_check_name` | string | Name of the health check in GCP |
| `type` | string | Probe protocol GCP computed (HTTP, TCP, GRPC, ...) |
| `region` | string | Region of a regional check; empty for global |

## Behavior Notes

- **Scope switch**: an empty `region` creates the global resource; a set region creates the regional one. The scope is immutable.
- **Immutability**: `health_check_name` and `project_id` are ForceNew; all probe knobs update in place.
- **Ports fall through to API defaults** (http/tcp 80, https/http2/ssl 443) when unset — the module never hardcodes them.
- **Prober firewall**: instance-group backends need an ingress rule admitting `35.191.0.0/16` and `130.211.0.0/22`, or every backend shows unhealthy.

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation

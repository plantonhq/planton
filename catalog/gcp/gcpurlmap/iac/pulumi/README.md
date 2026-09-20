# GCP URL Map - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying GCP Compute Engine URL maps using Planton's `GcpUrlMap` API. The module is written in Go and creates exactly one of `compute.URLMap` (global; `spec.region` empty) or `compute.RegionUrlMap` (regional; `spec.region` set), the same switch the Terraform module makes with its count guards; the regional builders (`region_url_map.go`) mirror the global ones minus the surfaces the regional API lacks.

A URL map is the routing brain of an Application Load Balancer — it matches each request's host and path and decides whether to forward, split, rewrite, redirect, or return a custom error page.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** with Compute Engine API available
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs
6. **A backend service** (or bucket) to reference as the default target — typically a `GcpBackendService` whose `self_link` you wire in

## Directory Structure

```
iac/pulumi/
├── main.go           # Pulumi program entry point
├── Pulumi.yaml       # Pulumi project configuration
├── README.md         # This file
└── module/
    ├── main.go       # Module coordinator
    ├── url_map.go    # URL map creation and mapping
    ├── locals.go     # Resolved resource + derived values
    └── outputs.go    # Stack output constants
```

## Quick Start

### 1. Initialize Pulumi Stack

```bash
cd iac/pulumi
pulumi stack init dev
```

### 2. Create Input File

Provide a `stack-input.yaml` with the URL map specification:

```yaml
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpUrlMap
  metadata:
    name: web-routing
  spec:
    project_id:
      value: my-gcp-project-123
    default_service:
      value: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/global/backendServices/web-backend
```

### 3. Preview and Deploy

```bash
pulumi preview
pulumi up
```

### 4. View Outputs

```bash
pulumi stack output self_link
pulumi stack output url_map_name
```

## Inputs

The module consumes `GcpUrlMapStackInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpUrlMap` spec (default target, host/path routing, tests, etc.) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `self_link` | string | Self-link URI — the value target proxies reference (`regions/{region}` in place of `global` for a regional map) |
| `url_map_name` | string | Name of the URL map in GCP |
| `map_id` | string | Server-assigned numeric ID |
| `fingerprint` | string | Server-computed fingerprint |
| `region` | string | Region of a regional URL map; empty for global |

## Behavior Notes

- **Immutability**: `url_map_name` and `project_id` are ForceNew; routing tables update in place.
- **Default target**: exactly one of `default_service`, `default_url_redirect`, or `default_route_action` with weighted backends (enforced pre-deploy by CEL).
- **Path matcher rules**: each path matcher uses either `path_rules` or `route_rules`, not both.
- **route_action scope**: the full traffic-management surface is mapped at every routing level — weighted splits (with per-backend header actions), URL rewrites, timeout/retry policies, request mirroring, CORS, fault injection, stream-duration limits, and route-scoped CDN cache policies.
- **Destroy stance**: `deletion_policy` (DELETE/PREVENT/ABANDON) is client-side — PREVENT fails a destroy outright; ABANDON drops the map from state without deleting it.

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation

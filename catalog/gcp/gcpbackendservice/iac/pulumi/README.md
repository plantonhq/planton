# GCP Backend Service - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying GCP Compute Engine backend services using Planton's `GcpBackendService` API. The module is written in Go and creates exactly one of `compute.BackendService` (global; `spec.region` empty) or `compute.RegionBackendService` (regional; `spec.region` set, in `region_backend_service.go`), the same switch the Terraform module makes with its count guards, plus — on the global arm — one `compute.BackendServiceSignedUrlKey` per configured signing key.

A backend service is the hub of the L7 load balancing family: it owns the backend list, health checking, session affinity, Cloud CDN policy, IAP, Cloud Armor attachment, and request logging. URL maps route traffic to it by self-link.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **A health check** — required unless every backend is an internet or serverless NEG
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
    ├── main.go             # Module coordinator
    ├── backend_service.go  # Backend service + signed-URL key creation
    ├── locals.go           # Resolved resource + derived values
    └── outputs.go          # Stack output constants
```

## Quick Start

### 1. Initialize Pulumi Stack

```bash
cd iac/pulumi
pulumi stack init dev
```

### 2. Create Input File

Provide a `stack-input.yaml` with the backend service specification:

```yaml
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpBackendService
  metadata:
    name: web-backend
  spec:
    project_id:
      value: my-gcp-project-123
    health_check:
      value: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/global/healthChecks/web-hc
    backends:
      - group:
          value: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/zones/us-central1-a/instanceGroups/web-ig
        balancing_mode: UTILIZATION
        max_utilization: 0.8
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
```

## Inputs

The module consumes `GcpBackendServiceStackInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpBackendService` spec (backends, health check, CDN, IAP, etc.) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

Spec fields mirror the Terraform module: protocol/scheme/timeouts, the singular `health_check` reference, `backends` with balancing modes and capacity dials, session affinity (incl. the strong-affinity cookie), locality policies and consistent hash, `enable_cdn` + `cdn_policy` with the full cache-key policy, Cloud Armor references, `iap`, `log_config`, headers and compression, the Traffic Director traffic-policy blocks, `security_settings`/`tls_settings`, the migration controls, and up to 3 `signed_url_keys`.

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `self_link` | string | Self-link URI — the value URL maps and passthrough forwarding rules reference (`regions/{region}` in place of `global` for a regional service) |
| `backend_service_name` | string | Name of the backend service in GCP |
| `generated_id` | string | Server-assigned numeric ID |
| `fingerprint` | string | Optimistic-concurrency fingerprint |
| `region` | string | Region of a regional backend service; empty for global |

## Behavior Notes

- **Immutability**: `backend_service_name` and `project_id` are ForceNew; everything else — backends, CDN policy, affinity, IAP — updates in place.
- **The scheme is always sent**: an unset `load_balancing_scheme` is sent as `EXTERNAL` (the spec's default, the classic global external ALB) rather than omitted. The provider's own default is `EXTERNAL_MANAGED` and the scheme is immutable, so leaving the choice to the provider would replace an existing classic backend service on its next apply. The Terraform module does the same.
- **Secrets**: the IAP `oauth2_client_secret`, the SigV4 `access_key`, and every signed-URL `key_value` are marked secret in Pulumi state (`pulumi.ToSecret`) and never exported.
- **One health check**: GCP caps `health_checks` at one; the SDK flattens the one-element set to a plain string, matching the spec's singular reference.
- **TTL semantics**: a 0 TTL in the spec means "unset — let the GCP API default"; cache-mode/TTL and scheme-applicability coherence is enforced by the spec before deploy.

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation

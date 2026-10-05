# GCP SSL Policy - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying a Compute Engine SSL policy using Planton's `GcpSslPolicy` API. The module is written in Go and creates `compute.SSLPolicy` (global) or `compute.RegionSslPolicy` (regional) based on whether `spec.region` is set.

An SSL policy controls which TLS versions and cipher suites a load balancer accepts from clients. Target HTTPS (and SSL) proxies reference the policy's `self_link`; without one, GCP's permissive default applies (minimum TLS 1.0, COMPATIBLE ciphers).

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
    ├── ssl_policy.go          # Global/regional policy creation
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
  kind: GcpSslPolicy
  metadata:
    name: prod-tls-floor
  spec:
    project_id:
      value: my-gcp-project-123
    profile: MODERN
    min_tls_version: TLS_1_2
```

```bash
make build
pulumi preview
pulumi up
```

## Inputs

The module consumes `GcpSslPolicyIacInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpSslPolicy` spec (profile, min TLS version, optional custom ciphers, optional region) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `self_link` | string | Self-link URI — the value a target HTTPS (or SSL) proxy references in `ssl_policy` |
| `ssl_policy_name` | string | Name of the SSL policy in GCP |
| `enabled_features` | string[] | Cipher suites the policy actually enables, as computed by GCP |
| `region` | string | Region of a regional policy; empty for global |

## Behavior Notes

- **One kind, two resources**: empty `region` creates the global policy; a set region creates the regional one. Scope is permanent.
- **Fleet-wide hardening in place**: `profile`, `min_tls_version`, and `custom_features` update in place and apply to every referencing proxy on the next handshake.
- **CUSTOM pairing**: the CUSTOM profile requires `custom_features`, and `custom_features` is rejected on every other profile — validated before deploy.
- **API defaults preserved**: unset profile/min-TLS fall through to GCP's defaults (COMPATIBLE / TLS_1_0) rather than being pinned by the module.
- **API enablement**: the module enables `compute.googleapis.com` (with `disable_on_destroy=false`) so a fresh project works first try.

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation

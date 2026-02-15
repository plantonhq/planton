# GCP Global Address - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying GCP global addresses using OpenMCF's `GcpGlobalAddress` API. The module is written in Go and leverages the Pulumi GCP provider to create `compute.GlobalAddress` resources (backed by `google_compute_global_address`).

Global addresses reserve static external or internal IP addresses (or CIDR ranges) for use with load balancers, VPC peering, and Private Service Connect.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** with billing enabled and Compute Engine API active
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: `roles/compute.networkAdmin` on the target project

## Directory Structure

```
iac/pulumi/
├── main.go           # Pulumi program entry point
├── Pulumi.yaml       # Pulumi project configuration
├── Makefile          # Build and deployment targets
├── debug.sh          # Debug helper script
├── README.md         # This file
├── overview.md       # Architecture overview
└── module/
    ├── main.go       # Module coordinator
    ├── global_address.go  # Global address resource creation
    ├── locals.go     # Local values and labels
    └── outputs.go    # Stack output constants
```

## Quick Start

### 1. Initialize Pulumi Stack

```bash
cd iac/pulumi
pulumi stack init dev
```

### 2. Create Input File

Provide a `stack-input.yaml` with the global address specification:

```yaml
target:
  apiVersion: gcp.openmcf.org/v1
  kind: GcpGlobalAddress
  metadata:
    name: my-global-ip
  spec:
    project_id:
      value: my-gcp-project-123
    address_name: my-global-ip
    address_type: EXTERNAL
    ip_version: IPV4

providerConfig:
  gcpCredential:
    value: <base64-encoded-service-account-key>
```

### 3. Build, Preview, and Deploy

```bash
make build
pulumi preview
pulumi up
```

### 4. View Outputs

```bash
pulumi stack output address
pulumi stack output self_link
pulumi stack output creation_timestamp
```

## Inputs

The module consumes `GcpGlobalAddressStackInput`, which includes:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpGlobalAddress` spec (project_id, address_name, address_type, ip_version, etc.) |
| `providerConfig` | Yes | GCP provider configuration (credentials) |

Spec fields: `project_id`, `address_name`, `address_type` (EXTERNAL/INTERNAL), `ip_version` (IPV4/IPV6), optional `address`, `description`, `network`, `purpose`, `prefix_length`.

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `address` | string | Reserved IP address or start of reserved CIDR range |
| `self_link` | string | Full self-link URI of the global address resource |
| `creation_timestamp` | string | RFC 3339 creation timestamp |

## Makefile Targets

```bash
make deps          # Download and tidy dependencies
make vet           # Run go vet
make fmt           # Format code
make build         # Build (runs deps, vet, fmt)
make update-deps   # Update OpenMCF dependencies to latest
```

## Debugging

Use `debug.sh` for local runs with a sample stack input:

```bash
./debug.sh preview
./debug.sh up
```

Ensure `stack-input.yaml` exists in the current directory or is pointed to by the script. Enable verbose Pulumi logging with `PULUMI_DEBUG=1 pulumi preview`.

## Related

- [Architecture Overview](overview.md) — internal module design and data flow
- [Terraform Module](../tf/README.md) — Terraform implementation

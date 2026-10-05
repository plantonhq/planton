# GCP Regional Address - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying GCP regional addresses using Planton's `GcpAddress` API. The module is written in Go and leverages the Pulumi GCP provider to create `compute.Address` resources (backed by `google_compute_address`).

Regional addresses reserve static external or internal IP addresses (or CIDR ranges) at regional scope for use with Cloud NAT, regional load balancers, VM instances, internal LB VIPs, VPC peering, and IPsec interconnect.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** with billing enabled
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
    ├── main.go       # Module coordinator
    ├── address.go    # Regional address resource creation
    ├── locals.go     # Local values and labels
    └── outputs.go    # Output constants
```

## Quick Start

### 1. Initialize Pulumi Stack

```bash
cd iac/pulumi
pulumi stack init dev
```

### 2. Create Input File

Provide a `iac-input.yaml` with the regional address specification:

```yaml
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpAddress
  metadata:
    name: my-regional-ip
  spec:
    project_id:
      value: my-gcp-project-123
    address_name: my-regional-ip
    region: us-central1
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
pulumi stack output name
pulumi stack output region
```

## Inputs

The module consumes `GcpAddressIacInput`, which includes:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpAddress` spec (project_id, address_name, region, address_type, etc.) |
| `providerConfig` | Yes | GCP provider configuration (credentials) |

Spec fields: `project_id`, `address_name`, `region` (required), `address_type` (EXTERNAL/INTERNAL), `ip_version` (IPV4/IPV6), optional `address`, `description`, `network`, `subnetwork`, `network_tier`, `prefix_length`, `purpose`, `ipv6_endpoint_type`.

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `address` | string | Reserved IP address or start of reserved CIDR range |
| `self_link` | string | Full self-link URI of the regional address resource |
| `name` | string | Name of the regional address resource in GCP |
| `region` | string | Plain region name from the spec (e.g. `us-central1`) |

## Makefile Targets

```bash
make build         # Build the Pulumi program
make preview       # Build and preview changes
make up            # Build and deploy
make destroy       # Build and destroy the stack
```

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation

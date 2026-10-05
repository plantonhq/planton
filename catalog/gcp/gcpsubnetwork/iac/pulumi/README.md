# GCP Subnetwork - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying GCP subnetworks using Planton's `GcpSubnetwork` API. The module is written in Go and creates a `compute.Subnetwork` (backed by `google_compute_subnetwork`) in an existing custom-mode VPC, enabling the Compute Engine API first.

A subnetwork is the regional address space workloads live in: primary IPv4 range, secondary (alias) ranges for GKE pods/services, optional dual-stack IPv6, special-purpose roles (proxy-only, Private Service Connect), and VPC Flow Logs.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **An existing custom-mode VPC** — the parent network
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
    ├── main.go        # Module coordinator
    ├── subnetwork.go  # API enablement + subnetwork creation
    ├── locals.go      # Resolved resource holder
    └── outputs.go     # Output constants
```

## Quick Start

### 1. Initialize Pulumi Stack

```bash
cd iac/pulumi
pulumi stack init dev
```

### 2. Create Input File

Provide a `iac-input.yaml` with the subnetwork specification:

```yaml
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpSubnetwork
  metadata:
    name: app-subnet
  spec:
    project_id:
      value: my-gcp-project-123
    vpc_self_link:
      value: projects/my-gcp-project-123/global/networks/my-vpc
    subnetwork_name: app-subnet
    region: us-central1
    ip_cidr_range: 10.10.0.0/20
    private_ip_google_access: true
```

### 3. Build, Preview, and Deploy

```bash
make build
pulumi preview
pulumi up
```

### 4. View Outputs

```bash
pulumi stack output subnetwork_self_link
```

## Inputs

The module consumes `GcpSubnetworkIacInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpSubnetwork` spec (VPC, ranges, purpose, IPv6, flow logs) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

Spec fields: `vpc_self_link` (required), `subnetwork_name` (required), `region` (required), `ip_cidr_range` (required except IPv6-only), optional `project_id` (falls back to the provider default project when empty), `purpose`/`role`, `secondary_ip_ranges`, `private_ip_google_access` + `private_ipv6_google_access`, `stack_type` + `ipv6_access_type` + `external_ipv6_prefix`, `allow_subnet_cidr_routes_overlap`, `send_secondary_ip_range_if_empty`, `log_config`.

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `subnetwork_self_link` | string | Self-link — the value subnet consumers reference |
| `subnetwork_name` | string | Name in GCP |
| `region` / `ip_cidr_range` | string | Placement and primary range |
| `secondary_ranges` | list | Names + CIDRs of secondary ranges (exported per-index for the outputs transformer) |
| `gateway_address` | string | IPv4 address of the default gateway |
| `subnetwork_id` | string | Server-assigned numeric ID (stringified for cross-engine shape stability) |
| `internal_ipv6_prefix` / `external_ipv6_prefix` | string | Allocated IPv6 prefixes |

## Behavior Notes

- **Immutability**: name, project, region, network, and description are ForceNew. The primary range expands in place but never shrinks.
- **Secondary-range safety latch**: an empty range list does NOT remove existing ranges unless `send_secondary_ip_range_if_empty` is true.
- **Flow-log defaults** mirror the GCP API's own (5s aggregation, 50% sampling, all metadata) so an empty `log_config` behaves sanely.

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation

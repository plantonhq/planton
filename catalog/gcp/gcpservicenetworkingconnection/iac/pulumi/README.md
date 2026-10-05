# GCP Service Networking Connection - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying a private services access connection using Planton's `GcpServiceNetworkingConnection` API. The module is written in Go and creates `servicenetworking.Connection` — the VPC peering between one of your networks and a service producer's network.

For the default producer (`servicenetworking.googleapis.com`), this peering is what lets Cloud SQL, AlloyDB, Memorystore (PRIVATE_SERVICE_ACCESS mode), and Filestore hand out private IPs from `VPC_PEERING` address ranges reserved inside your VPC.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** with the Service Networking and Compute Engine APIs enabled (the module enables both if needed)
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
    ├── connection.go          # Connection creation + API enablement
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
  kind: GcpServiceNetworkingConnection
  metadata:
    name: prod-vpc-psa
  spec:
    network:
      value: projects/my-gcp-project-123/global/networks/prod-vpc
    reserved_peering_ranges:
      - value: prod-vpc-psa-range
```

```bash
make build
pulumi preview
pulumi up
```

## Inputs

The module consumes `GcpServiceNetworkingConnectionIacInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpServiceNetworkingConnection` spec (network, reserved ranges, optional service override) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `peering` | string | Name of the VPC peering created on the network (e.g. `servicenetworking-googleapis-com`) |
| `network` | string | The peered VPC network as the connection resolved it |

## Behavior Notes

- **One connection per (network, service) pair**: GCP rejects a second connection for the same pair — capacity grows by appending ranges to this resource, never by adding another connection.
- **Ranges are referenced by NAME**: `reserved_peering_ranges` carries `GcpGlobalAddress` names (INTERNAL, purpose `VPC_PEERING`), not self-links or CIDRs.
- **In-place range growth**: appending ranges updates the connection without disturbing service subnets the producer already provisioned.
- **Teardown ordering**: GCP refuses to delete the connection while the producer still holds subnets — destroy the private-IP service instances first.
- **API enablement**: the module enables `servicenetworking.googleapis.com` and `compute.googleapis.com` (with `disable_on_destroy=false`) so a fresh project works first try.

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation

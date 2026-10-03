# GCP Region Network Endpoint Group - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying a GCP regional network endpoint group using Planton's `GcpRegionNetworkEndpointGroup` API. The module is written in Go and creates `compute.RegionNetworkEndpointGroup`.

A regional NEG bridges a load balancer's backend service to a serverless workload (Cloud Run, Cloud Functions, App Engine), a Private Service Connect endpoint, or an external origin — instead of a group of VMs.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** — regional NEGs are free; no billed infrastructure is created
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
    ├── main.go                            # Module coordinator
    ├── region_network_endpoint_group.go   # NEG creation
    ├── locals.go                          # Resolved resource + derived values
    └── outputs.go                         # Output constants
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
  kind: GcpRegionNetworkEndpointGroup
  metadata:
    name: web-neg
  spec:
    project_id:
      value: my-gcp-project-123
    region: us-central1
    cloud_run:
      service:
        value: my-cloud-run-service
```

```bash
make build
pulumi preview
pulumi up
```

## Inputs

The module consumes `GcpRegionNetworkEndpointGroupIacInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpRegionNetworkEndpointGroup` spec (region, endpoint type, target block) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `self_link` | string | Self-link URI — the value a backend service references in `backends[].group` |
| `network_endpoint_group_name` | string | Name of the NEG in GCP |
| `network_endpoint_type` | string | The NEG's endpoint type |
| `region` | string | Region the NEG lives in |

## Behavior Notes

- **Fully immutable**: every field is ForceNew; any change destroys and recreates the NEG.
- **Serverless target need not exist at create time**: endpoints resolve at serving time.
- **API enablement**: the module enables `compute.googleapis.com` (with `disable_on_destroy=false`) so a fresh project works first try.

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation

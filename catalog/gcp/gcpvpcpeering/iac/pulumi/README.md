# GCP VPC Peering - Pulumi Module

## Overview

This directory contains the Pulumi implementation for managing one side of a VPC Network Peering using Planton's `GcpVpcPeering` API. The module is written in Go and creates exactly one of two resources through the bridged Google provider: `compute.NetworkPeering` when this side creates the peering, or `compute.NetworkPeeringRoutesConfig` when it manages the route exchange of a peering that already exists.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
4. **IAM permissions**: `roles/compute.networkAdmin` on the project of `network` (creating a peering also needs `compute.networks.addPeering` on the peer network's project, which `networkAdmin` there grants)

## Directory Structure

```
iac/pulumi/
├── main.go                      # Pulumi program entry point
├── Pulumi.yaml                  # Pulumi project configuration
├── README.md                    # This file
└── module/
    ├── main.go                  # Module coordinator: selects the form
    ├── network_peering.go       # CREATE form
    ├── peering_routes_config.go # ROUTES-CONFIG form
    ├── locals.go                # Peering name and form selection
    └── outputs.go               # Stack output constants
```

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --stack org/project/stack
planton pulumi destroy --manifest ../../e2e/manifest.yaml --stack org/project/stack
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`.

## What the module does

- **`peer_network` set (CREATE)**: creates the peering entry on `network` pointing at `peer_network`, named from `peering_name` or `metadata.name`. Always sends all four route-exchange flags (the public-IP pair with the spec's defaults — true / false — when unset, because they are immutable and the implied value must be on the wire from the first apply). Sends `stack_type`, `update_strategy`, and `deletion_policy` only when set. Exports `peering_name`, `network`, `state`, `state_details`.
- **`peer_network` empty (ROUTES-CONFIG)**: manages the route exchange of the existing peering named `peering_name` on `network` (for example `servicenetworking-googleapis-com`). Always sends all four route-exchange flags (the provider requires the custom-route pair; the public-IP pair carries the spec's defaults when unset); takes the project from the network's self link when it carries one. Exports `peering_name` and `network`; `state` and `state_details` are empty strings because this form does not own the peering entry.

## Parity with Terraform

Both engines select the same resource on the same condition, send the same arguments, and export the same four keys. There is no `PARITY-EXCEPTION` in this module.

## Notes

- **A peering is two half-entries.** Declare the other side as its own `GcpVpcPeering` (or have the other party create theirs); `state` reads `ACTIVE` only when both exist.
- **The CREATE form is nearly immutable**: name, both networks, and the public-IP flags recreate the peering.
- **Destroying the ROUTES-CONFIG form is a no-op in GCP** — the peering keeps its current flags.

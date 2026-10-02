# GCP HA VPN Gateway - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying one HA VPN gateway and the Cloud Router its tunnels terminate BGP on, using Planton's `GcpHaVpnGateway` API. The module is written in Go and creates `compute.HaVpnGateway` and `compute.Router` (plus `projects.Service` for API enablement) through the bridged Google provider. Sites connect to the gateway through `GcpHaVpnConnection` resources that reference it.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
4. **IAM permissions**: `roles/compute.networkAdmin` on the project

## Directory Structure

```
iac/pulumi/
├── main.go               # Pulumi program entry point
├── Pulumi.yaml           # Pulumi project configuration
├── README.md             # This file
└── module/
    ├── main.go           # Module coordinator
    ├── ha_vpn_gateway.go # API enablement, the gateway, the router
    ├── locals.go         # Names and the merged label set
    └── outputs.go        # Stack output constants
```

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --stack org/project/stack
planton pulumi destroy --manifest ../../e2e/manifest.yaml --stack org/project/stack
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`.

## What the module does

- Enables `compute.googleapis.com` (never disabled on destroy).
- Creates the HA VPN gateway named from `gateway_name` or `metadata.name`, in `region` on `network`, with the platform labels merged over the user's. Sends `description`, `gateway_ip_version`, `stack_type`, the Interconnect `vpn_interfaces`, `resource_manager_tags`, and `deletion_policy` only when set.
- Creates the Cloud Router named from `router.name` or the gateway's name, in the same region and network, with the required `bgp` block (ASN always; advertisement mode, groups, ranges, keepalive, identifier range when set), `encrypted_interconnect_router` always sent, tags and description when set, and the same `deletion_policy`.
- Exports `gateway_self_link`, `gateway_name`, `region`, `interface_0_ip_address`, `interface_1_ip_address` (picked by interface id from the read-back list), `router_name`, `router_self_link`, `router_asn`.

## Parity with Terraform

Both engines create the same three resources with the same arguments and export the same eight keys. There is no `PARITY-EXCEPTION` in this module.

## Notes

- **A recreated gateway has NEW public IPs.** Network, region, IP version, stack type, and interface pinning are immutable.
- **The router's ASN is immutable**; its description and advertisement change in place.
- **Deleting the gateway fails while any `GcpHaVpnConnection`'s tunnels reference it** — a chart's dependency order destroys the connections first.

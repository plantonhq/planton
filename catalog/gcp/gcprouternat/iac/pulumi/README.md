# GcpRouterNat - Pulumi Module

This Pulumi (Go) module provisions a Cloud Router (`compute.Router`) with a Cloud NAT gateway (`compute.RouterNat`) — managed egress for instances without external IPs. It is the Pulumi-side implementation of the Planton `GcpRouterNat` resource kind and has feature parity with the Terraform module.

## Overview

The module creates the router and its NAT as one unit (the one-NAT-per-router shape) and enables the Compute Engine API so a fresh project works first try. Manual NAT IPs are referenced `GcpAddress` reservations, never created here — the reservation is its own composable node, and the literal IP lives on that node's outputs.

`router_name`, `nat_name`, `region`, the network, `endpoint_types`, and `type` force replacement; everything else — IP allocation and draining, subnetwork scoping, port tuning, timeouts, rules, logging — updates in place, which is what makes NAT IP rotation and fleet-wide egress tuning zero-downtime operations.

Behavior notes carried in the module:

- **Allocation is derived, not declared**: a non-empty `nat_ips` list selects `MANUAL_ONLY`; an empty list selects `AUTO_ONLY`. Private NAT (`type: PRIVATE`) sets no allocation option at all — it draws from subnetwork ranges.
- **Scoping is derived the same way**: listing `subnetworks` implies `LIST_OF_SUBNETWORKS`; an empty list means every subnetwork in the region, all ranges.
- **Logging**: the `DISABLED` filter turns logging off (the API still requires a valid placeholder filter value); every other filter value enables it.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Build

```bash
cd catalog/gcp/gcprouternat/iac/pulumi
make build
```

## Outputs

| Name | Description |
|------|-------------|
| `name` | Name of the Cloud NAT gateway as created in GCP |
| `router_self_link` | Self-link URL of the Cloud Router carrying this NAT |
| `nat_ips` | Self links of the static IPs the NAT translates through (manual allocation only; empty for auto-allocation) |

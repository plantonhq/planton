# GcpRouterNat - Terraform Module

This Terraform module provisions a Cloud Router (`google_compute_router`) with a Cloud NAT gateway (`google_compute_router_nat`) — managed egress for instances without external IPs. It is the Terraform-side implementation of the Planton `GcpRouterNat` resource kind and has feature parity with the Pulumi module.

## Overview

The module creates the router and its NAT as one unit (the one-NAT-per-router shape) and enables the Compute Engine API so a fresh project works first try. Manual NAT IPs are referenced `GcpAddress` reservations, never created here — the reservation is its own composable node, and the literal IP lives on that node's outputs.

`router_name`, `nat_name`, `region`, the network, `endpoint_types`, `type`, `encrypted_interconnect_router`, and `resource_manager_tags` are ForceNew; everything else — IP allocation and draining, subnetwork scoping, NAT64 scope, port tuning, timeouts, rules, logging, BGP advertisement — updates in place, which is what makes NAT IP rotation and fleet-wide egress tuning zero-downtime operations. The module runs on the plain `google` provider: every modeled field is GA at the pinned release.

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml
planton tofu plan --manifest ../../e2e/manifest.yaml
planton tofu apply --manifest ../../e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --auto-approve
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Terraform Usage

```bash
cd catalog/gcp/gcprouternat/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpRouterNat spec | — |

The `spec` object includes: `router_name` + `nat_name` + `region` + `vpc_self_link` (required, all ForceNew), `type` (PUBLIC/PRIVATE), `nat_ips` + `drain_nat_ips` (GcpAddress self links; non-empty `nat_ips` selects MANUAL_ONLY, empty selects AUTO_ONLY), `auto_network_tier`, `source_subnetwork_ip_ranges_to_nat` + `subnetworks` (per-subnetwork range scoping), `source_subnetwork_ip_ranges_to_nat64` + `nat64_subnetworks` (IPv6-to-IPv4 translation scope), `min_ports_per_vm` / `max_ports_per_vm` / `enable_dynamic_port_allocation` / `enable_endpoint_independent_mapping`, `endpoint_types`, the five idle/wait timeouts, `rules` (CEL match + dedicated IPs or ranges), `log_filter`, `router_bgp` (ASN, advertise mode/groups/ranges, keepalive, identifier range), `router_description`, `encrypted_interconnect_router`, `resource_manager_tags`, `deletion_policy` (applied to both router and NAT), and `project_id` (empty falls back to the provider default project).

## Outputs

| Name | Description |
|------|-------------|
| `name` | Name of the Cloud NAT gateway as created in GCP |
| `router_self_link` | Self-link URL of the Cloud Router carrying this NAT |
| `nat_ips` | Self links of the static IPs the NAT translates through (manual allocation only; empty for auto-allocation) |

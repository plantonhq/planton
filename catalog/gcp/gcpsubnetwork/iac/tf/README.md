# GcpSubnetwork - Terraform Module

This Terraform module provisions a subnetwork in a custom-mode GCP VPC. It is the Terraform-side implementation of the Planton `GcpSubnetwork` resource kind and has feature parity with the Pulumi module.

## Overview

The module enables the Compute Engine API (never disabling it on destroy) and creates a `google_compute_subnetwork` with the full address plan: primary IPv4 range (literal or sourced from a Network Connectivity reserved internal range), secondary (alias) ranges, purpose/role (proxy-only, PSC), dual-stack IPv6 (including BYOIP via `ip_collection` and prefix pinning), Private Google Access (v4 and v6), the secondary-range removal latch, create-time Resource Manager tags, and VPC Flow Logs. Everything runs on the GA `hashicorp/google` provider (`allow_subnet_cidr_routes_overlap` is GA on the 8.x line).

`name`, `project`, `region`, `network`, and `description` are immutable (ForceNew). The primary range is asymmetric: expansion updates in place, shrinkage recreates.

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml
planton tofu plan --manifest ../../e2e/manifest.yaml
planton tofu apply --manifest ../../e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --auto-approve
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Terraform Usage

```bash
cd catalog/gcp/gcpsubnetwork/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpSubnetwork spec | — |

The `spec` object includes: `vpc_self_link` (required), `subnetwork_name` (required), `region` (required), `ip_cidr_range` (required unless `reserved_internal_range` supplies the primary range, or on IPv6-only subnets), `reserved_internal_range` (Network Connectivity internal range as the primary-range source), `project_id` (empty falls back to the provider default project), `purpose`/`role`, `secondary_ip_ranges` (each with `ip_cidr_range` OR `reserved_internal_range`), `private_ip_google_access` + `private_ipv6_google_access`, `stack_type` + `ipv6_access_type` + `external_ipv6_prefix` + `internal_ipv6_prefix` (ULA pinning) + `ip_collection` (BYOIP sub-PDP), `allow_subnet_cidr_routes_overlap`, `send_secondary_ip_range_if_empty`, `resource_manager_tags` (create-time; changing them replaces the subnet), `resolve_subnet_mask` (ARP resolution mode; immutable), `log_config` (flow logs), and `deletion_policy` (DELETE/PREVENT/ABANDON destroy behavior).

## Outputs

| Name | Description |
|------|-------------|
| `subnetwork_self_link` | Self-link — the value subnet consumers reference |
| `subnetwork_name` | Name in GCP |
| `region` / `ip_cidr_range` | Placement and primary range |
| `secondary_ranges` | Names + CIDRs of secondary ranges |
| `gateway_address` | IPv4 address of the default gateway |
| `subnetwork_id` | Server-assigned numeric ID (as a string) |
| `internal_ipv6_prefix` / `external_ipv6_prefix` | Allocated IPv6 prefixes |

# GcpVpcPeering - Terraform Module

This Terraform module manages one side of a VPC Network Peering: `google_compute_network_peering` when this side creates the peering, or `google_compute_network_peering_routes_config` when it manages the route exchange of a peering that already exists. It is the Terraform-side implementation of the Planton `GcpVpcPeering` resource kind and has feature parity with the Pulumi module.

## Overview

Two resources, exactly one per instance, count-gated on whether `spec.peer_network` is set. The CREATE form owns the peering entry on `network` and its route exchange; the ROUTES-CONFIG form owns only the route exchange of an existing peering named `peering_name` (the shape for Google-managed private services access peerings such as `servicenetworking-googleapis-com`). The module runs on the plain `google` provider — every modeled field is GA on the pinned 8.x line.

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
cd catalog/gcp/gcpvpcpeering/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpVpcPeering spec | — |

The `spec` object includes: `peering_name` (empty defaults to `metadata.name`), `network`, `peer_network` (set → CREATE form; empty → ROUTES-CONFIG form), `export_custom_routes`, `import_custom_routes`, `export_subnet_routes_with_public_ip`, `import_subnet_routes_with_public_ip`, `stack_type`, `update_strategy`, and `deletion_policy` (the last three CREATE form only).

`variables.tf` is generated from the proto contract (`planton tofu generate-variables GcpVpcPeering`) and formatted with `tofu fmt`; regenerate it when the spec changes rather than editing by hand.

## Outputs

| Name | Description |
|------|-------------|
| `peering_name` | The peering entry's name on this side's network |
| `network` | This side's network |
| `state` | `ACTIVE` / `INACTIVE` (CREATE form; empty on the ROUTES-CONFIG form) |
| `state_details` | GCP's explanation of the state (CREATE form; empty otherwise) |

## Resources Created

- `google_compute_network_peering` (count-gated on `peer_network` set) — all four route-exchange flags always sent (the public-IP pair with the spec's defaults when unset, because they are immutable); `stack_type`, `update_strategy`, `deletion_policy` sent only when set
- `google_compute_network_peering_routes_config` (count-gated on `peer_network` empty) — all four route-exchange flags always sent (the provider requires the custom-route pair; the public-IP pair carries the spec's defaults when unset); the project taken from the network's self link when it carries one

## Notes

- **A peering is two half-entries.** Declare the other side as its own `GcpVpcPeering` (or have the other party create theirs); `state` reads `ACTIVE` only when both exist. The provider serializes the two sides when one chart creates both.
- **The CREATE form is nearly immutable**: name, both networks, and the public-IP flags recreate the peering.
- **Destroying the ROUTES-CONFIG form is a no-op in GCP** — the peering keeps its current flags.

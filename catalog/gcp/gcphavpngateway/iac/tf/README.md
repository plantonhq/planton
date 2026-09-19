# GcpHaVpnGateway - Terraform Module

This Terraform module provisions one HA VPN gateway (`google_compute_ha_vpn_gateway`) and the Cloud Router its tunnels terminate BGP on (`google_compute_router`). It is the Terraform-side implementation of the Planton `GcpHaVpnGateway` resource kind and has feature parity with the Pulumi module.

## Overview

Three resources: API enablement, the gateway, and the router — the pair provisioned as one node because Google allows many routers per network and region, so a VPN router beside a NAT router is the normal topology. The gateway's two interfaces receive public IPs (or, pinned to Interconnect attachments, none); the router carries the ASN and default advertisement every `GcpHaVpnConnection` session on it inherits. The module runs on the plain `google` provider — every modeled field is GA on the pinned 8.x line.

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
cd catalog/gcp/gcphavpngateway/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpHaVpnGateway spec | — |

The `spec` object includes: `project_id`, `gateway_name` (empty defaults to `metadata.name`), `region`, `network`, `description`, `gateway_ip_version`, `stack_type`, `vpn_interfaces` (Interconnect pinning), `labels`, `resource_manager_tags`, `router` (`name` — empty defaults to the gateway's name — `description`, `bgp`, `encrypted_interconnect_router`, `resource_manager_tags`), and `deletion_policy`.

`variables.tf` is generated from the proto contract (`planton tofu generate-variables GcpHaVpnGateway`) and formatted with `tofu fmt`; regenerate it when the spec changes rather than editing by hand.

## Outputs

| Name | Description |
|------|-------------|
| `gateway_self_link` | What a `GcpHaVpnConnection`'s `gateway` references |
| `gateway_name` | The gateway's name in GCP |
| `region` | What a `GcpHaVpnConnection`'s `region` references |
| `interface_0_ip_address` | Public IP of interface 0 (empty when Interconnect-backed) |
| `interface_1_ip_address` | Public IP of interface 1 (empty when Interconnect-backed) |
| `router_name` | What a `GcpHaVpnConnection`'s `router` references |
| `router_self_link` | The Cloud Router's self link |
| `router_asn` | The ASN the router speaks as |

## Resources Created

- `google_project_service` — enables `compute.googleapis.com`; never disabled on destroy
- `google_compute_ha_vpn_gateway` — optional inputs sent only when set; labels merged (platform wins); `deletion_policy` sent only when set
- `google_compute_router` — the required `bgp` block with the ASN always and the rest when set; `encrypted_interconnect_router` always sent; the same `deletion_policy`

## Notes

- **A recreated gateway has NEW public IPs.** Network, region, IP version, stack type, and interface pinning are immutable.
- **The router's ASN is immutable**; its description and advertisement change in place.
- **Deleting the gateway fails while any `GcpHaVpnConnection`'s tunnels reference it** — a chart's dependency order destroys the connections first.
- **The gateway itself is free**; Google bills per tunnel-hour and for tunnel traffic, both declared on the `GcpHaVpnConnection`.

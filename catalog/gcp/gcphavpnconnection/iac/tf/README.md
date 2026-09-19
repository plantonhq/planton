# GcpHaVpnConnection - Terraform Module

This Terraform module connects one HA VPN gateway to one peer with one to four IPsec tunnels, each carrying a BGP session: an optional `google_compute_external_vpn_gateway` and, per tunnel, `google_compute_vpn_tunnel`, `google_compute_router_interface`, and `google_compute_router_peer` on the gateway's Cloud Router. It is the Terraform-side implementation of the Planton `GcpHaVpnConnection` resource kind and has feature parity with the Pulumi module.

## Overview

The external gateway is count-gated on `spec.peer.external_gateway`; the three per-tunnel resources are `for_each` over the tunnels keyed by name, so adding or removing a tunnel in the middle of the list never renumbers (and so recreates) its neighbours. The gateway trio (`gateway`, `router`, `region`) arrives as resolved references to the `GcpHaVpnGateway`. The module runs on the plain `google` provider — every modeled field is GA on the pinned 8.x line.

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
cd catalog/gcp/gcphavpnconnection/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpHaVpnConnection spec | — |

The `spec` object includes: `project_id`, `gateway`, `router`, `region` (all three referencing the `GcpHaVpnGateway`), `peer` (exactly one of `external_gateway { name, redundancy_type, interfaces[], description, labels }` or `gcp_gateway`), `tunnels[]` (`name`, `vpn_gateway_interface`, `peer_external_gateway_interface`, `shared_secret`, `ike_version`, `local_traffic_selector`, `remote_traffic_selector`, `cipher_suite`, `labels`, `description`, `bgp_session { name, interface_ip_range, ip_version, peer_asn, peer_ip_address, advertised_route_priority, advertise_mode, advertised_groups, advertised_ip_ranges, enable, enable_ipv4, enable_ipv6, the four nexthop addresses, custom_learned_ip_ranges, custom_learned_route_priority, bfd, md5_authentication_key, import_policies, export_policies }`), `resource_manager_tags`, and `deletion_policy`.

`variables.tf` is generated from the proto contract (`planton tofu generate-variables GcpHaVpnConnection`) and formatted with `tofu fmt`; regenerate it when the spec changes rather than editing by hand.

## Outputs

| Name | Description |
|------|-------------|
| `tunnel_self_links` | The tunnels' self links, in spec order |
| `tunnel_names` | The tunnels' names, in spec order |
| `router_interface_names` | The Cloud Router interfaces' names, in spec order |
| `bgp_peer_names` | The BGP peers' names, in spec order |
| `external_gateway_self_link` | The external VPN gateway's self link (empty for a Google peer) |
| `gateway_self_link` | The HA VPN gateway the tunnels leave from |
| `router_name` | The Cloud Router the sessions run on |

## Resources Created

- `google_compute_external_vpn_gateway` (count-gated on an external peer) — the device's interfaces, exactly one address each; labels merged (platform wins)
- `google_compute_vpn_tunnel` (`for_each` tunnel) — `shared_secret` (sensitive), `ike_version` always sent (2 when unset), exactly one of `peer_external_gateway` / `peer_gcp_gateway`, traffic selectors and cipher suite only when set
- `google_compute_router_interface` (`for_each` tunnel) — named from `bgp_session.name` or the tunnel's name; `ip_range` / `ip_version` only when set
- `google_compute_router_peer` (`for_each` tunnel) — `enable` always sent (true when unset), optional numerics and addresses only when set, BFD and the MD5 key (sensitive; name defaults to `<tunnel>-md5`) when declared

`deletion_policy` is fanned to every resource when set.

## Notes

- **Every tunnel field except labels is immutable.** Rotate a pre-shared key by adding a tunnel with the new key and removing the old one.
- **Two tunnels, one per gateway interface, to two peer addresses is the 99.99% shape.**
- **The MD5 key rides the peer**: the provider inserts it into the router's key table and Google requires each key to be used by exactly one session.
- **Tunnels bill per hour from creation**, whether or not the peer is up.

# CloudflareZeroTrustTunnel — Research & Design Notes

## Purpose

A Cloudflare Tunnel (`cloudflare_zero_trust_tunnel_cloudflared`) creates a persistent,
outbound-only connection between a private network and Cloudflare's global edge, run by
the `cloudflared` connector. It replaces inbound firewall holes and VPN appliances for
two use cases: publishing private HTTP/TCP/SSH/RDP services on public hostnames (ingress),
and making private IP ranges reachable to enrolled WARP devices (routes).

## 80/20 scope and the fold decision

The provider models tunnels across four resources: the tunnel, its remote configuration
(ingress), routes, and virtual networks. This component:

- **Folds the configuration** (`zero_trust_tunnel_cloudflared_config`) into the tunnel as
  `spec.ingress` + `spec.origin_request`. The configuration is 1:1 with the tunnel
  (keyed by `tunnel_id`) and has no independent lifecycle, so it is a field, not a kind —
  the same reasoning as a queue's consumer. The module still provisions it as a distinct
  provider resource, so toggling ingress never recreates the tunnel.
- **Keeps routes and virtual networks as separate kinds**
  (`CloudflareZeroTrustTunnelRoute`, `CloudflareZeroTrustTunnelVirtualNetwork`): both are
  account-scoped, independently versioned, subnet/segment-class primitives that compose
  across many tunnels.

The full `origin_request` surface (~15 fields, including Access enforcement) is modeled
and reused for both the tunnel-level default and per-ingress overrides.

## config_src default

The provider defaults `config_src` to `local`. This component defaults to `cloudflare`
(remote management) because that is the only mode in which ingress can be expressed as
desired state — the whole point of declarative IaC. `local` remains available for users
who manage ingress with a cloudflared YAML on the origin; in that mode only the tunnel
object and the run token are managed.

## The token output

The connector run token is not an attribute of the tunnel resource in either engine. It
is read from a dedicated data source:

- Terraform: `data "cloudflare_zero_trust_tunnel_cloudflared_token"`.
- Pulumi: `GetZeroTrustTunnelCloudflaredTokenOutput` (the Output-returning form, so the
  tunnel's computed id can be passed as an Input).

It is exported as `tunnel_token` and marked sensitive in both engines.

## Field-name nuance

The provider field `match_sn_ito_host` is a known upstream misspelling of "match SNI to
host". The spec exposes the corrected name `match_sni_to_host`; both modules map it to the
provider's `match_sn_ito_host` / `MatchSnItoHost`. If the provider ever corrects the
spelling, only the modules change — the spec already reads correctly.

## Engine parity

Both engines create the same resources for an identical spec and emit the same outputs.
Unset optional origin-request values are omitted on both sides (false booleans included),
keeping plans byte-for-byte equivalent. No `PARITY-EXCEPTION` is required at
pulumi-cloudflare v6.17.0 / provider v5.

## Gotchas

- The final ingress rule must be a catch-all (a `service` with no `hostname`, e.g.
  `http_status:404`); the spec enforces this with CEL.
- Ingress requires `config_src: cloudflare`; setting ingress on a `local` tunnel is
  rejected by CEL.
- A public hostname only resolves once a DNS record CNAMEs to `tunnel_cname`.

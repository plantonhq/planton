# CloudflareZeroTrustTunnel

Provision a Cloudflare Tunnel (cloudflared): a secure, outbound-only connection from a
private network to Cloudflare's edge. A tunnel exposes private HTTP/TCP/SSH/RDP services
through public hostnames (ingress rules) and/or makes private IP ranges reachable to
WARP clients (via `CloudflareZeroTrustTunnelRoute`) — without opening any inbound
firewall ports.

## Why a first-class tunnel

The tunnel is the anchor of a private-connectivity topology: DNS records CNAME to it,
routes attach private networks to it, and a connector authenticates with its token.
Modeling it as a node lets the resource graph wire all of those as explicit edges.

The ingress configuration is folded in (`ingress`, `originRequest`) rather than being a
separate kind: a tunnel has exactly one configuration, with no independent lifecycle. The
module still provisions the configuration as its own provider resource, so editing
ingress never recreates the tunnel.

## Quick start

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustTunnel
metadata:
  name: prod-tunnel
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: prod-tunnel
  ingress:
    - hostname: app.example.com
      service: http://localhost:8080
    - service: http_status:404   # required catch-all
```

## Configuration reference

| Field | Required | Description |
|---|---|---|
| `accountId` | yes | 32-char Cloudflare account ID |
| `name` | yes | User-friendly tunnel name |
| `configSrc` | no | `cloudflare` (default, remote-managed ingress) or `local` (origin YAML) |
| `tunnelSecret` | no | Base64 secret (≥32 bytes) for a locally-managed tunnel; sensitive |
| `ingress` | no | Public-hostname rules, top-to-bottom; the last must be a catch-all (service, no hostname). Requires `configSrc: cloudflare` |
| `ingress[].service` | yes | Local service, e.g. `http://localhost:8080`, `ssh://10.0.0.5:22`, `http_status:404` |
| `ingress[].originRequest` / `originRequest` | no | Per-rule / tunnel-wide origin connection settings (timeouts, TLS, Access) |
| `originRequest.access.audTag` | — | Access application AUD tags (literal or `CloudflareZeroTrustAccessApplication` refs) |

## Stack outputs

| Output | Description |
|---|---|
| `tunnel_id` | The tunnel UUID |
| `tunnel_cname` | CNAME target (`<id>.cfargotunnel.com`) for public hostnames |
| `tunnel_token` | Connector run token (sensitive) |
| `tunnel_status` | `inactive` / `degraded` / `healthy` / `down` |
| `account_tag` | Account tag |
| `created_on` | Creation timestamp |

## Composition

- Point a `CloudflareDnsRecord` CNAME at `status.outputs.tunnel_cname` to route a public
  hostname through the tunnel.
- Attach private networks with `CloudflareZeroTrustTunnelRoute` referencing
  `status.outputs.tunnel_id`.
- Protect ingress with Cloudflare Access by referencing a
  `CloudflareZeroTrustAccessApplication`'s `aud` output in `originRequest.access.audTag`.
- Run the connector with `status.outputs.tunnel_token`
  (`cloudflared tunnel run --token <token>`).

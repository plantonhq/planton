# Pulumi Module: Cloudflare Zero Trust Tunnel

Provisions a `cloudflare.ZeroTrustTunnelCloudflared` — a cloudflared tunnel
whose connector makes an outbound-only connection to Cloudflare's edge — plus,
when the tunnel is remotely managed and declares ingress rules, its companion
`cloudflare.ZeroTrustTunnelCloudflaredConfig` (the remote ingress
configuration, a separate provider resource so editing ingress never recreates
the tunnel).

## Layout

```
iac/pulumi/
├── main.go            # entrypoint (loads iac-input, calls module.Resources)
├── Pulumi.yaml
└── module/
    ├── main.go            # Resources(): provider setup + tunnel()
    ├── locals.go          # iac-input references
    ├── tunnel.go          # the tunnel + conditional config + mappers
    └── outputs.go         # output constant names
```

## Inputs

A `CloudflareZeroTrustTunnelIacInput` (target + provider config). Required:
`account_id` and `name`. `config_src` selects remote (`cloudflare`, the
default) vs `local` management; ingress and origin-request settings apply only
to a remotely-managed tunnel.

## Outputs

- `tunnel_id` — the tunnel's UUID (routes reference it).
- `tunnel_cname` — the `<id>.cfargotunnel.com` target DNS records CNAME to.
- `tunnel_token` — the connector run token (**sensitive**).
- `tunnel_status`, `account_tag`, `created_on`.

## Requirements

- API token with **Account → Cloudflare Tunnel → Edit**.

# Pulumi Module: Cloudflare Zero Trust Tunnel Route

Provisions a single `cloudflare.ZeroTrustTunnelCloudflaredRoute` — a private
CIDR made reachable to WARP clients through an existing cloudflared tunnel,
optionally inside a specific virtual network (omit for the account's default
network).

## Layout

```
iac/pulumi/
├── main.go            # entrypoint (loads iac-input, calls module.Resources)
├── Pulumi.yaml
└── module/
    ├── main.go            # Resources(): provider setup + route()
    ├── locals.go          # iac-input references
    ├── route.go           # the cloudflare.ZeroTrustTunnelCloudflaredRoute
    └── outputs.go         # output constant names
```

## Inputs

A `CloudflareZeroTrustTunnelRouteIacInput` (target + provider config).
Required: `account_id`, `network` (the CIDR — unique per virtual network), and
`tunnel_id`. Optional: `virtual_network_id` and `comment`.

## Outputs

- `route_id` — the route's UUID.
- `network` — the routed CIDR as recorded by Cloudflare.

## Requirements

- API token with **Account → Cloudflare Tunnel → Edit**.

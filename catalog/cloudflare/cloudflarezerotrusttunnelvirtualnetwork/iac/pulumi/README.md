# Pulumi Module: Cloudflare Zero Trust Tunnel Virtual Network

Provisions a single `cloudflare.ZeroTrustTunnelCloudflaredVirtualNetwork` — a
named routing namespace that lets overlapping private CIDRs coexist in one
account (each tunnel route may target a specific virtual network).

## Layout

```
iac/pulumi/
├── main.go            # entrypoint (loads iac-input, calls module.Resources)
├── Pulumi.yaml
└── module/
    ├── main.go              # Resources(): provider setup + virtualNetwork()
    ├── locals.go            # iac-input references
    ├── virtual_network.go   # the ..VirtualNetwork resource
    └── outputs.go           # output constant names
```

## Inputs

A `CloudflareZeroTrustTunnelVirtualNetworkIacInput` (target + provider
config). Required: `account_id` and `name` (unique within the account).
Optional: `comment` and `is_default_network` (promoting a network to the
account default demotes the current one — flip it deliberately).

## Outputs

- `virtual_network_id` — the network's UUID (routes reference it).
- `virtual_network_name`.

## Requirements

- API token with **Account → Cloudflare Tunnel → Edit**.

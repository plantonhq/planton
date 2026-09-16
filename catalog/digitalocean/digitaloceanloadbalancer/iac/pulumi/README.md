# DigitalOcean Load Balancer -- Pulumi Module

Deploys a `digitalocean:index/loadBalancer:LoadBalancer` from a `DigitalOceanLoadBalancer` stack input: regional and global types, sizing, forwarding rules with TLS termination or passthrough, health checks, sticky sessions, backend targeting by Droplet IDs or tag, VPC placement, firewall, HTTPS redirect, PROXY protocol, keepalive, idle-timeout, TLS cipher policy, project placement, and the global balancer's domains, targets, CDN, and failover. Bridge SDK pin is `pulumi-digitalocean/sdk/v4 v4.53.0` (the gaps below re-verified against that version's `LoadBalancerArgs`).

## Module structure

- `main.go` -- Pulumi program entry point reading the stack input
- `module/main.go` -- `Resources()`: locals, provider, load balancer
- `module/locals.go` -- stack-input references and the standard Planton label map
- `module/load_balancer.go` -- the balancer resource and stack-output exports
- `module/outputs.go` -- output key constants (the kind's outputs.proto contract)

## Outputs

Exactly the kind's stack-output contract, identical to the Terraform module: `load_balancer_id`, `ip`, `urn`, `ipv6`.

## Behavior notes

These spec fields are real and the Terraform module wires them. This module fails the apply with `PARITY-EXCEPTION` if they are meaningfully set (proto zero values pass), until the Pulumi DigitalOcean SDK grows the matching args:

- **`spec.subnet_uuid`** -- no subnet_uuid field on LoadBalancer
- **`spec.ip`** -- Ip is a computed output only; BYOIP cannot be set at create time

Re-evaluate each when the SDK exposes it.

- Health-check `path` and interval/threshold leaves are sent only when set, so a TCP check never carries an empty path and an omitted interval never arrives as `0` (outside the provider's 3–300 range).
- `droplet_ids` entries that are not numeric Droplet IDs fail the apply rather than being silently skipped.
- Firewall leaves are `Allows` / `Denies` on the SDK (plural) against the provider's `allow` / `deny`.
- See the kind [GUIDE](../../GUIDE.md).

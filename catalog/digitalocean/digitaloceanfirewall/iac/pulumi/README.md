# DigitalOcean Firewall -- Pulumi Module

Deploys a `digitalocean:index/firewall:Firewall` from a `DigitalOceanFirewall` stack input: the named rule set (both directions, all five source/destination classes), Droplet targeting by resolved reference, and tag targeting. Bridge SDK pin is `pulumi-digitalocean/sdk/v4 v4.79.1`, which carries the complete provider argument surface for this resource — no PARITY-EXCEPTION guards.

## Module structure

- `main.go` -- Pulumi program entry point reading the stack input
- `module/main.go` -- `Resources()`: locals, provider, firewall
- `module/locals.go` -- stack-input references and the standard Planton label map
- `module/firewall.go` -- the firewall resource, reference resolution, and the stack-output export
- `module/outputs.go` -- output key constants (the kind's outputs.proto contract)

## Behavior notes

- Reference fields (`droplet_ids` and the rule-level droplet/kubernetes/load-balancer lists) carry resolved literal values by the time the module runs; droplet entries convert to the API's integers and fail loudly on a non-numeric value.
- `port_range` stays unset on icmp rules and empty collections stay nil — matching the provider's read-back normalization so rule hashes stay stable (see the Terraform module's notes; both engines send identical shapes).

## Outputs

Exactly the kind's stack-output contract, identical to the Terraform module: `firewall_id`.

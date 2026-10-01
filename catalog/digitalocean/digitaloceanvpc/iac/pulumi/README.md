# DigitalOcean VPC -- Pulumi Module

Deploys a `digitalocean:index/vpc:Vpc` from a `DigitalOceanVpc` stack input: the VPC's name from `metadata.name`, the region, an optional description, and an optional immutable IP range. Bridge SDK pin is `pulumi-digitalocean/sdk/v4 v4.79.1`, which carries the complete provider argument surface -- no PARITY-EXCEPTION guards. (The SDK renames the VPC's `urn` attribute to `VpcUrn`; the module exports it under the contract's `urn` key.)

## Module structure

- `main.go` -- Pulumi program entry point reading the stack input
- `module/main.go` -- `Resources()`: locals, provider, vpc
- `module/locals.go` -- stack-input references and the standard Planton label map
- `module/vpc.go` -- the VPC resource and stack-output exports
- `module/outputs.go` -- output key constants (the kind's outputs.proto contract)

## Behavior notes

- An unset `ip_range_cidr` stays nil: DigitalOcean assigns a non-conflicting range, reported through the `ip_range` output.
- The range is immutable: an edit to a set range plans a full VPC REPLACEMENT.
- `name` and `description` update in place; `region` is create-only.

## Outputs

Exactly the kind's stack-output contract, identical to the Terraform module: `vpc_id`, `ip_range`, `urn`.

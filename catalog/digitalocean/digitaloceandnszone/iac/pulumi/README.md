# DigitalOcean DNS Zone -- Pulumi Module

Deploys a `digitalocean:index/domain:Domain` plus one `digitalocean:index/dnsRecord:DnsRecord` per managed record value from a `DigitalOceanDnsZone` stack input: the zone itself, the create-only `ip_address` convenience, and the inline records with their per-type fields on presence semantics. Bridge SDK pin is `pulumi-digitalocean/sdk/v4 v4.79.1`, which carries the complete provider argument surface for both resources — no PARITY-EXCEPTION guards. (The SDK renames the domain's `urn` attribute to `DomainUrn`; the module exports it under the contract's `urn` key.)

## Module structure

- `main.go` -- Pulumi program entry point reading the stack input
- `module/main.go` -- `Resources()`: locals, provider, zone
- `module/locals.go` -- stack-input references and the standard Planton label map
- `module/dns_zone.go` -- the domain resource, the per-value record fan-out, and stack-output exports
- `module/outputs.go` -- output key constants (the kind's outputs.proto contract)

## Behavior notes

- Each record value becomes its own record resource named `name-index-valueIndex` — identical fan-out to the Terraform module — and the records are chained with `DependsOn` so they are created (and destroyed) ONE AT A TIME: DigitalOcean's DNS API deadlocks on concurrent record writes to one domain (`422 Error 1213 Deadlock found`).
- `ip_address` is ForceNew and never read back by the API, so the domain carries `IgnoreChanges("ipAddress")` — adopting an existing zone whose manifest carries it never plans a zone replace (the Terraform module does the same).
- `priority`/`weight`/`port`/`flags` are set only when present; `ttl_seconds` 0/unset leaves the ttl Computed (DigitalOcean's 1800-second default); `ip_address` is sent only when non-empty.

## Outputs

Exactly the kind's stack-output contract, identical to the Terraform module: `zone_name`, `zone_id`, `name_servers`, `urn`, and `record_ids` (the inline records' numeric ids keyed by the resource name `name-index-valueIndex` — the second half of each record's `{domain},{record_id}` import id).

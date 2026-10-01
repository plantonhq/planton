# DigitalOcean DNS Record -- Pulumi Module

Deploys a `digitalocean:index/dnsRecord:DnsRecord` from a `DigitalOceanDnsRecord` stack input: every record type the API accepts (A, AAAA, CNAME, MX, TXT, SRV, NS, CAA, SOA) with the per-type fields carried on presence semantics. Bridge SDK pin is `pulumi-digitalocean/sdk/v4 v4.79.1`, which carries the complete provider argument surface for this resource — no PARITY-EXCEPTION guards. (The SDK's named RecordType enum omits SOA; the module passes types as raw strings, so SOA works regardless.)

## Module structure

- `main.go` -- Pulumi program entry point reading the stack input
- `module/main.go` -- `Resources()`: locals, provider, record
- `module/locals.go` -- stack-input references and the standard Planton label map
- `module/dns_record.go` -- the record resource and stack-output exports
- `module/outputs.go` -- output key constants (the kind's outputs.proto contract)

## Behavior notes

- `priority`/`weight`/`port`/`flags` are set only when present in the spec; `ttl` left unset is Computed — DigitalOcean applies its 1800-second default.
- Outputs come from the created resource (`Fqdn`, `Ttl`), never recomputed locally, so both engines export identical values.

## Outputs

Exactly the kind's stack-output contract, identical to the Terraform module: `record_id`, `hostname`, `record_type`, `domain`, `ttl_seconds`.

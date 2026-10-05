# DigitalOcean DNS Record -- Terraform Module

Deploys a `digitalocean_record` from a `DigitalOceanDnsRecord` spec: every record type the API accepts (A, AAAA, CNAME, MX, TXT, SRV, NS, CAA, SOA) with the per-type fields carried on presence semantics. Provider pin is `~> 2.99`.

`variables.tf` is generated (`planton tofu generate-variables DigitalOceanDnsRecord`). Do not hand-edit it. The API token lives in `credentials.tf`.

## Prerequisites

- OpenTofu or Terraform 1.5+
- DigitalOcean API token (`digitalocean_token`)
- An existing DigitalOcean-hosted zone (the `domain`)

## Usage

```hcl
module "dns_record" {
  source = "./path/to/module"

  metadata = {
    name = "app-a-record"
  }

  spec = {
    domain = "example.com"
    name   = "app"
    type   = "A"
    value  = "203.0.113.10"
  }

  digitalocean_token = var.digitalocean_token
}
```

## Behavior notes

- Reference fields (`domain`, `value`) arrive flattened as plain strings — the Planton orchestrator resolves `valueFrom` references before Terraform runs.
- The spec's enum value names ARE the provider's record types, so `type` wires through directly.
- `priority`/`weight`/`port`/`flags` pass through as null when unset (spec presence semantics); `ttl` left null is Computed — DigitalOcean applies its 1800-second default and the applied value reads back.
- Outputs come from the created resource (`fqdn`, `ttl`), never recomputed locally, so both engines export identical values.

## Outputs

Exactly the kind's output contract, identical to the Pulumi module: `record_id`, `hostname`, `record_type`, `domain`, `ttl_seconds`.

# CloudflareZeroTrustDnsLocation Pulumi Module

Pulumi (Go) IaC module for Gateway DNS locations.

## Architecture

```
main.go                   — Entrypoint loading the IaC input
module/main.go            — Resources(): provider setup, resource, outputs
module/locals.go          — Locals initialization
module/dns_location.go    — cloudflare.ZeroTrustDnsLocation
module/outputs.go         — output keys
```

## Behavior

Mirrors the Terraform module's contract exactly: a plain CRUD resource (real create/update/delete; only the account forces replacement). Update is a full replace at the API — `max_ttl` is sent whenever the spec declares it, and omitting it genuinely resets the TTL behavior to inherit. `dns_destination_ips_id` is only sent when set; unset lets Cloudflare auto-assign the shared IPv4 destination pair.

The module ALWAYS sends `max_ttl` and every networks list as known values — an unset `max_ttl` coalesces to the documented server default (`mode: inherit`) and absent lists are sent empty. At provider v5.23.0/v5.24.0 these computed-optional attributes carry raw Go model types that cannot hold "unknown", so a null crashes the apply-time conversion through the bridged provider too (measured live; unfixed on provider main). The coalesce changes nothing semantically. One residue no module can fix: a declared endpoint with an EMPTY networks list re-plans a cosmetic update forever on the Terraform side (Cloudflare drops empty lists on read) — declare real networks or accept the no-op diff.

## Outputs

| Name | Description |
|------|-------------|
| `location_id` | The Cloudflare-assigned UUID of the location |
| `doh_subdomain` | The location's unique DNS-over-HTTPS subdomain |
| `ip` | The IPv4 destination assigned to the plain-DNS endpoint |

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).

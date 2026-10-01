# CloudflareZeroTrustDnsLocation Terraform Module

Terraform IaC module for Gateway DNS locations.

## Architecture

```
provider.tf   — Cloudflare provider configuration (~> 5.23)
variables.tf  — Input variables mirroring CloudflareZeroTrustDnsLocationSpec (generated)
main.tf       — cloudflare_zero_trust_dns_location
outputs.tf    — location_id, doh_subdomain, ip
```

## Behavior

A plain CRUD resource: real create/update/delete; only the account forces replacement. Update is a full replace at the API — `max_ttl` is sent whenever the spec declares it, and omitting it genuinely resets the TTL behavior to inherit. The endpoints tree declares ALL FOUR types at once (spec validation mirrors the provider schema). `dns_destination_ips_id` is only sent when set; unset lets Cloudflare auto-assign the shared IPv4 destination pair. Import as `{account_id}/{location_id}`.

The module ALWAYS sends `max_ttl` and every networks list as known values — an unset `max_ttl` coalesces to the documented server default `{mode = "inherit"}` and absent lists are sent empty. At provider v5.23.0/v5.24.0 these computed-optional attributes carry raw Go model types that cannot hold "unknown", so a null in config crashes the apply-time conversion (measured live; unfixed on provider main). The coalesce changes nothing semantically. One residue no module can fix: a declared endpoint with an EMPTY networks list re-plans a cosmetic in-place update forever (Cloudflare drops empty lists on read) — declare real networks or accept the no-op diff.

## Outputs

| Name | Description |
|------|-------------|
| `location_id` | The Cloudflare-assigned UUID of the location |
| `doh_subdomain` | The location's unique DNS-over-HTTPS subdomain |
| `ip` | The IPv4 destination assigned to the plain-DNS endpoint |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.

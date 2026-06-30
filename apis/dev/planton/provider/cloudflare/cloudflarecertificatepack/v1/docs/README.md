# CloudflareCertificatePack — Research & Design Notes

## Purpose

An advanced certificate pack (`cloudflare_certificate_pack`, type `advanced`) is a
publicly-trusted edge TLS certificate that Cloudflare provisions and auto-renews for
a chosen set of hostnames, with a selectable certificate authority and validity
period — going beyond the zone's free Universal SSL certificate.

## Scope and immutability

Nearly every attribute (`certificate_authority`, `type`, `validation_method`,
`validity_days`, `hosts`, `cloudflare_branding`) is immutable upstream: changing any
of them re-orders (replaces) the pack. The spec models them as required/derived
accordingly; `type` defaults to the only supported value, `advanced`.

## Allowed values

`certificate_authority`, `validation_method`, and `type` are validated with CEL using
the provider's exact strings so the modules pass them through verbatim (no lossy
enum-to-string mapping). `validity_days` is an int constrained to {14, 30, 90, 365}.

## Domain control validation

For a zone using Cloudflare's nameservers, `txt` validation completes automatically —
Cloudflare manages the validation records, so no downstream DNS record is required.
For partial (CNAME) setups the DCV records would need to be published manually; that
flow is out of scope for the initial outputs (which expose `certificate_pack_id`,
`status`, and `primary_certificate`). Exposing per-host DCV records as structured
outputs is a future enhancement if a concrete composition need arises.

## Outputs

Only `repeated string`/scalar outputs are used (the established output grain for this
provider). The pack's computed certificate list and validation records are available
on the resource but are not surfaced as stack outputs.

## Engine parity

Both engines create the same `certificate_pack` and emit the same outputs
(`certificate_pack_id`, `status`, `primary_certificate`). `cloudflare_branding` is
sent only when true on both sides. No `PARITY-EXCEPTION` is required at
pulumi-cloudflare v6.17.0 / provider v5. (Pulumi types `zone_id` as an optional
pointer where Terraform requires it; the spec requires it and both modules always
send it, so behavior is identical.)

## Gotchas

- Requires Advanced Certificate Manager on the zone; ordering fails without it.
- `hosts` must include the zone apex and may not exceed 50 entries.

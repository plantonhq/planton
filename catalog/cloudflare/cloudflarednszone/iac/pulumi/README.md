# Pulumi Module: Cloudflare DNS Zone

Provisions a `cloudflare.Zone` together with its zone-singleton satellites —
`cloudflare.ZoneDnsSettings`, `cloudflare.ZoneDnssec`, `cloudflare.ZoneHold`,
`cloudflare.ZoneSubscription` — and an inline set of `cloudflare.DnsRecord`
resources whose lifecycle tracks the zone. Records with independent lifecycles
are better modeled as standalone `cloudflarednsrecord` resources; the inline
surface is identical in depth.

## Layout

```
iac/pulumi/
├── main.go            # entrypoint (loads iac-input, calls module.Resources)
├── Pulumi.yaml
└── module/
    ├── main.go            # Resources(): provider setup + zone()
    ├── locals.go          # iac-input references
    ├── zone.go            # cloudflare.Zone + satellite orchestration + outputs
    ├── records.go         # inline records + the typed-data builder
    ├── dns_settings.go    # cloudflare.ZoneDnsSettings
    ├── dnssec.go          # cloudflare.ZoneDnssec (status "active" when enabled)
    ├── hold.go            # cloudflare.ZoneHold (created only when enabled)
    ├── subscription.go    # cloudflare.ZoneSubscription (created only when set)
    └── outputs.go         # output constant names
```

## Inputs

A `CloudflareDnsZoneIacInput` (target + provider config). Required spec
fields: `zoneName`, `accountId`.

`spec.records[]` entries are either simple records (`content`) or structured
records whose typed oneof case (srv/caa/cert/…) is translated by
`buildRecordData` in [module/records.go](./module/records.go) into the
provider's single union `data` object. `ttl` 0 maps to 1 (automatic — the
provider requires ttl ≥ 1). Top-level `priority` is sent for MX from the
record's priority field, and for SRV/URI mirrored from their structured data
(Cloudflare reflects it there on read; omitting the mirror re-plans forever,
live-measured). HTTPS/SVCB carry priority only inside their data.

## Outputs

- `zone_id` — the zone ID (also the import identity of every zone-singleton
  satellite).
- `nameservers` — the Cloudflare-assigned nameservers.
- `status` — zone status (`pending` until the registrar delegates).
- `dnssec_*` — DS/DNSKEY material to enter at the registrar (empty strings when
  DNSSEC is off, so the output contract is always satisfied).
- `record_ids` — inline-record Cloudflare ids keyed by name-type-index (import
  recipes derive `{zone_id}/{dns_record_id}` from it; empty map without inline
  records).

## Requirements

- The Cloudflare provider is configured from the iac-input provider config /
  `CLOUDFLARE_API_TOKEN`. The token needs **Zone → Zone → Edit** and
  **Zone → DNS → Edit**; add **Zone → Zone Settings → Edit** for the hold, and
  **Billing → Edit** (Billing Write) when `spec.subscription` is set.
- `vanity_name_servers` needs a Business/Enterprise plan; `foundation_dns` is a
  paid add-on; `multi_provider` and `secondary_overrides` are plan-gated.

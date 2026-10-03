# Public DNS Zone

Creates a public Cloud DNS managed zone for an internet-facing domain. DNS records are composed separately via [GcpDnsRecord](/docs/catalog/gcp/gcpdnsrecord).

## When to Use

- New public domains that need authoritative nameservers from Google Cloud DNS
- Foundation zones for cert-manager, external-dns, or manual GcpDnsRecord resources

## Key Configuration Choices

- **visibility: public** — zone is exposed to the internet; configure returned nameservers at your registrar
- **dns_name omitted** — defaults to `metadata.name` + `.` (e.g. `example.com.`)

## Placeholders to Replace

| Placeholder | Description |
|---|---|
| `my-gcp-project-123` | GCP project ID |
| `example.com` | Your domain (metadata.name) |

## Related Presets

- **02-private-vpc** — internal service discovery on a VPC
- **03-private-dnssec** — public zone with DNSSEC enabled

## Related Kinds

- [GcpDnsRecord](/docs/catalog/gcp/gcpdnsrecord) — individual DNS records in this zone
- [GcpProject](/docs/catalog/gcp/gcpproject) — project that owns the zone

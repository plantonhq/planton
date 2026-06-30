---
title: "DNSSEC-Signed Zone"
description: "Creates a zone with DNSSEC enabled. Cloudflare signs the zone, and the DS record material (digest, key tag, algorithm, and the full DS record) is published as stack outputs for you to enter at your..."
type: "preset"
rank: "02"
presetSlug: "02-dnssec-signed"
componentSlug: "dns-zone"
componentTitle: "DNS Zone"
provider: "cloudflare"
icon: "package"
order: 2
---

# DNSSEC-Signed Zone

Creates a zone with DNSSEC enabled. Cloudflare signs the zone, and the DS record
material (digest, key tag, algorithm, and the full DS record) is published as
stack outputs for you to enter at your domain registrar to complete the chain of
trust.

## When to Use

- Hardening a domain against DNS spoofing/cache poisoning with DNSSEC
- Any zone where you will paste DS records into the registrar after provisioning

## Key Configuration Choices

- **dnssec.enabled: true** (`dnssec.enabled`) -- Turns on Cloudflare DNSSEC signing.
- **DS outputs** -- After apply, read `dnssec_ds` (and the individual digest/key-tag
  fields) from the stack outputs and enter them at your registrar.
- For multi-provider or secondary-DNS setups, also set `dnssec.multi_signer`,
  `dnssec.presigned`, or `dnssec.use_nsec3`.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<your-domain.com>` | Fully qualified domain for the zone | Your registered domain |
| `<cloudflare-account-id>` | Cloudflare account ID | Cloudflare Dashboard → Overview → Account ID (right sidebar) |

## Note

DNSSEC fully activates only once the zone is active (nameservers delegated) and
the DS records are accepted by your registrar.

## Related Presets

- **01-basic-zone** -- A plain zone without DNSSEC

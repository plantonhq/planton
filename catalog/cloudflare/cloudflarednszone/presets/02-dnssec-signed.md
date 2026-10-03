---
display_name: DNSSEC Signed
---

# DNSSEC-Signed Zone

Creates a zone with DNSSEC enabled. Cloudflare signs the zone, and the DS record
material (digest, key tag, algorithm, and the full DS record) is published as
outputs for you to enter at your domain registrar to complete the chain of
trust.

## Deploy in two phases — DNSSEC cannot be enabled on a brand-new zone

Cloudflare accepts the DNSSEC enable only on an ACTIVE zone (nameservers already
delegated at the registrar). A zone created by this preset starts PENDING, and
the enable fails with error 1017 "Invalid zone plan for action" until the
registrar delegates — so a single apply on a fresh domain cannot succeed. The
working sequence:

1. Create the zone WITHOUT DNSSEC (the **01-basic-zone** preset), or with
   `dnssec.enabled: false`.
2. At your registrar, delegate the domain to the nameservers in
   `status.outputs.nameservers` and wait for the zone to report `active`
   (typically 1–24h of registrar propagation).
3. Set `dnssec.enabled: true` (this preset's shape) and re-apply. Then enter the
   DS material from the outputs at the registrar.

Apply this preset directly only to a domain whose zone is already active on
your account.

## When to Use

- Hardening an already-delegated (active) domain against DNS spoofing/cache
  poisoning with DNSSEC
- Any zone where you will paste DS records into the registrar after provisioning

## Key Configuration Choices

- **dnssec.enabled: true** (`dnssec.enabled`) -- Turns on Cloudflare DNSSEC
  signing. Active zones only — see the two-phase sequence above.
- **DS outputs** -- After apply, read `dnssec_ds` (and the individual digest/key-tag
  fields) from the outputs and enter them at your registrar.
- For multi-provider or secondary-DNS setups, also set `dnssec.multi_signer`,
  `dnssec.presigned`, or `dnssec.use_nsec3`.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `replaceme.example.com` | Fully qualified domain for the zone | Your registered domain |
| `0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d` | Cloudflare account ID | Cloudflare Dashboard → Overview → Account ID (right sidebar) |

## Note

Enter the DS records at the registrar only AFTER the zone is signed — entering
DS material for an unsigned zone breaks resolution. The chain of trust
completes once the registrar accepts the DS records.

## Related Presets

- **01-basic-zone** -- A plain zone without DNSSEC (phase 1 of the sequence above)

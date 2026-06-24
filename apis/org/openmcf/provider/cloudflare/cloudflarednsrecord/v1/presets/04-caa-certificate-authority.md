# CAA Record to Restrict Certificate Issuance

Creates a CAA record that controls which certificate authorities may issue
certificates for your domain. CAA records are structured: their flags, tag, and
value are supplied through the `data.caa` block.

## When to Use

- Restricting certificate issuance to a specific CA (e.g., Let's Encrypt)
- Hardening a domain against mis-issuance by unauthorized CAs
- Specifying an `iodef` contact for issuance-policy violation reports

## Key Configuration Choices

- **type CAA** (`type: CAA`) -- Certification Authority Authorization; uses the `data.caa` block.
- **data.caa.tag** (`tag: issue`) -- `issue` (allow standard certs), `issuewild` (wildcard certs), or `iodef` (violation reporting URL).
- **data.caa.value** (`value: letsencrypt.org`) -- The authorized CA's domain, or an `iodef` URL.
- **data.caa.flags** (`flags: 0`) -- 128 marks the property critical; 0 is the common default.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<cloudflare-zone-id>` | Zone ID for the DNS zone | CloudflareDnsZone status.outputs.zone_id or Dashboard |

## Related Presets

- **03-srv-service** -- Another structured record type using a `data` block

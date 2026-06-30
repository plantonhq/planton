# Cloudflare Certificate Pack

Order an advanced edge certificate for a Cloudflare zone — a publicly-trusted TLS
certificate, auto-renewed by Cloudflare, covering the hostnames you choose.

## What Gets Created

- A `cloudflare_certificate_pack` (type `advanced`) for the zone's hostnames.

## Prerequisites

- A `CloudflareDnsZone` (or an existing zone ID).
- A Cloudflare API token with `SSL and Certificates` permission.
- Advanced Certificate Manager enabled on the zone.

## Configuration Reference

**Required**

- `zoneId` — the zone the certificate is ordered for.
- `certificateAuthority` — `google`, `lets_encrypt`, or `ssl_com`.
- `validationMethod` — `txt`, `http`, or `email`.
- `validityDays` — 14, 30, 90, or 365.
- `hosts` — covered hostnames (must include the zone apex).

**Optional**

- `type` — `advanced` (default).
- `cloudflareBranding` — add Cloudflare branding to the order.

## Stack Outputs

| Output | Description |
|---|---|
| `certificate_pack_id` | The certificate pack identifier |
| `status` | The order/issuance status |
| `primary_certificate` | The primary certificate identifier |

## Related Components

- `CloudflareDnsZone`

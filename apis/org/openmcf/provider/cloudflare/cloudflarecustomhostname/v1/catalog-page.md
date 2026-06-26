# Cloudflare Custom Hostname (Cloudflare for SaaS)

Extend Cloudflare's edge — TLS, caching, WAF — onto your customers' own domains.
With Cloudflare for SaaS, a customer points their hostname (e.g. `support.acme.com`)
at your SaaS zone and Cloudflare provisions and auto-renews a per-customer
certificate, so the customer's branded domain serves over HTTPS through your
infrastructure.

## How it works (a concrete example)

Say you run **Helpdesk.io** and a customer, Acme, wants the product on
`support.acme.com`:

1. **One-time, on your zone (`helpdesk.io`):** configure a fallback origin (see
   `CloudflareCustomHostnameFallbackOrigin`) pointing at your app, e.g.
   `origin.helpdesk.io`. Every custom hostname routes here by default.
2. **Per customer:** create a `CloudflareCustomHostname` for `support.acme.com` on
   your zone. Cloudflare returns ownership-verification records (a TXT name/value,
   and an HTTP alternative) in the stack outputs.
3. **The customer makes two DNS changes on `acme.com`:**
   - `support.acme.com  CNAME  <your SaaS CNAME target>` — routes their traffic to you.
   - the `ownership_verification` TXT record — proves they control the domain.
4. **Cloudflare does the rest:** validates control, issues and auto-renews a
   certificate for `support.acme.com`, terminates TLS at the edge, applies your
   zone's WAF/cache/rules, and forwards to your origin (override per-hostname with
   `customOriginServer`).

The result: `https://support.acme.com` is live, valid HTTPS, branded entirely as
Acme — and you never handled a certificate file. Repeat per customer; it is one
`CloudflareCustomHostname` node each.

## What Gets Created

- A `cloudflare_custom_hostname` on the SaaS zone, with the requested SSL settings.

## Prerequisites

- A SaaS zone (`CloudflareDnsZone`) with a fallback origin
  (`CloudflareCustomHostnameFallbackOrigin`).
- A Cloudflare API token with `SSL and Certificates` permission.

## Configuration Reference

**Required**

- `zoneId` — the SaaS zone.
- `hostname` — the customer's hostname.

**Optional**

- `customOriginServer` / `customOriginSni` — override the origin for this hostname.
- `customMetadata` — arbitrary key/value metadata (e.g. a tenant id).
- `ssl` — `bundleMethod` (`ubiquitous` default), `certificateAuthority`, `method`
  (`http`/`txt`/`email`), `type` (`dv`), `wildcard`, `cloudflareBranding`, an
  uploaded `customCertificate`/`customKey` or `customCertBundle`, and `settings`
  (`ciphers`, `earlyHints`, `http2`, `minTlsVersion`, `tls_1_3`).

## Stack Outputs

| Output                                             | Description                                    |
| -------------------------------------------------- | ---------------------------------------------- |
| `custom_hostname_id`                               | The custom hostname identifier                 |
| `status`                                           | Activation status                              |
| `ownership_verification_name` / `_type` / `_value` | DNS record the customer adds to verify control |
| `ownership_verification_http_url` / `_http_body`   | HTTP verification alternative                  |
| `verification_errors`                              | Any verification errors                        |
| `created_at`                                       | Creation timestamp                             |

## Pricing

Cloudflare for SaaS is available on Free, Pro, Business, and Enterprise plans.

|                               | Free / Pro / Business | Enterprise                |
| ----------------------------- | --------------------- | ------------------------- |
| Custom hostnames included     | 100, free             | Custom                    |
| Price per additional hostname | $0.10 / month         | Custom                    |
| Max hostnames                 | 50,000                | Unlimited (contact sales) |

Billing is usage-based and pro-rated: Cloudflare meters active custom hostnames per
day, so removing a churned customer's hostname stops the charge. Several `ssl`
options — uploaded `customCertificate`/`customCertBundle`, a selectable
`certificateAuthority`, `wildcard`, and mTLS — are Enterprise-only.

## Related Components

- `CloudflareCustomHostnameFallbackOrigin`, `CloudflareDnsZone`

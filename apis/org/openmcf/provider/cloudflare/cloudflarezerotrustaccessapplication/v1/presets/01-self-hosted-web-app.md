# Preset: Self-hosted web application

A self-hosted web app behind Cloudflare Access, protecting `dashboard.example.com`
with a referenced Access policy and a 24-hour session.

## When to use

- You front an internal web app/dashboard with Cloudflare and want Access in front.
- Authorization lives in a reusable `CloudflareZeroTrustAccessPolicy`.

## Key choices

- `domain`: the FQDN Access protects (required for `self_hosted`).
- `policies`: reference one or more `CloudflareZeroTrustAccessPolicy` resources by
  output ID; `precedence` sets evaluation order.
- `autoRedirectToIdentity`: skip the IdP chooser when a single IdP is configured.

## Placeholders

| Placeholder | Description |
|---|---|
| `REPLACE_WITH_ACCOUNT_ID` | 32-character Cloudflare account ID |

## Composition

Pair with a `CloudflareZeroTrustAccessPolicy` (the decision + rules) and, optionally,
reusable `CloudflareZeroTrustAccessGroup`s referenced from that policy.

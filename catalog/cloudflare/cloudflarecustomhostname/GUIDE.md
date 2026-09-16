# CloudflareCustomHostname guide

Operational judgment for SaaS custom hostnames. The README covers what each field is; this covers how the pieces interact.

## The account must be enrolled in Cloudflare for SaaS first

Every custom-hostname API call — creates AND reads — answers 400 code 1404 "No quota has been allocated for this zone or for this account" until Cloudflare for SaaS is enabled (measured live on Free and Pro zones alike; plan tier does not matter). Enrollment is a dashboard action (SSL/TLS → Custom Hostnames) with a payment method on file; the first 100 hostnames are included, and every hostname beyond that allotment bills per hostname per month. The quota lands ACCOUNT-WIDE (measured live 2026-08-29: after enrolling, a create succeeded on a freshly created, still-pending zone that was never individually enrolled), so one enrollment unlocks every zone on the account. Nothing in this kind works before that toggle.

## The certificate authority is not yours to pick (unless Enterprise)

Setting `ssl.certificate_authority` on any non-Enterprise plan is rejected with 400 code 1459 "Certificate Authority selection is only available on an Enterprise plan" (measured live 2026-08-29). When the field is left empty, Cloudflare assigns a CA at random per create — consecutive identical creates measured `ssl_com` then `google`. Because config can never mirror that random pick, both IaC modules deliberately ignore post-create drift on this one field (the upstream provider's own acceptance tests do exactly the same). Consequence: changing the CA of an existing hostname in the manifest is NOT applied in place — recreate the hostname to change CA (and only Enterprise accounts may set it at all).

## Ownership proof is the customer's job

Onboarding a hostname does not make it live. Cloudflare returns TXT (and sometimes HTTP) ownership-verification records; the customer must create them on *their* DNS (or serve the HTTP body) before the hostname activates. The stack outputs those records so a chart or a ticket can hand them over. Until then the hostname sits in `pending` / `pending_validation`.

## SSL is a nested lifecycle

The hostname resource owns its SSL settings. Changing validation method, certificate type, or swapping a custom cert is an SSL-settings edit, not a hostname replace — but a custom (BYO) certificate is Enterprise-gated and will 403 on a non-Enterprise account even though the proto accepts it. Keep BYO certs off the default path.

## The fallback origin is a different kind

Traffic for a custom hostname that has no more-specific origin goes to the zone's fallback origin (CloudflareCustomHostnameFallbackOrigin). That singleton is per-zone, not per-hostname. Set it once for the SaaS zone; do not try to express it on each hostname.

## Soft-delete still answers GET

A hostname in `pending_deletion` or `deleted` is still readable. An automation that treats "the id exists" as "the hostname is live" will lie after a destroy. Wait for a real 404 (or an absent-status) before considering it gone.

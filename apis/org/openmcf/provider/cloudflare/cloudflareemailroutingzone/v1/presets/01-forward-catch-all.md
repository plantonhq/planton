# Preset: Forward-All Email Routing

Enable Email Routing on a zone and forward all otherwise-unmatched mail to a
verified destination address. The simplest "catch everything → my inbox" setup.

## When to use

- A domain that should funnel all inbound mail to one (or a few) real mailboxes.

## Key choices

- `catchAll.type: forward` with `forwardTo` listing verified destination
  addresses (each must be a verified `CloudflareEmailRoutingAddress`).
- Add `CloudflareEmailRoutingRule`s for per-address routing that should take
  precedence over the catch-all.

## Placeholders

| Placeholder | Description |
|---|---|
| `<zone-name>` | Name of the CloudflareDnsZone |
| `<destination-email>` | A verified destination email address |

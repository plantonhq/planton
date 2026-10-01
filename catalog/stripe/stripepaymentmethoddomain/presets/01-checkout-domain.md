# Checkout Domain

This preset registers the domain your checkout page is served from, so Apple Pay, Google Pay, Link and the other wallets can show their buttons there. Stripe validates the domain for each wallet; `status.outputs` reports each one as active or inactive, with Stripe's reason when it is not.

## When to Use

- A checkout built with Stripe Elements or embedded Checkout on your own domain
- Any page where wallet buttons should appear

## Key Configuration Choices

- **Domain** (`domainName`) -- a hostname, no scheme or path; changing it registers the new one and leaves the old one registered in Stripe
- **Apple Pay** -- serve Stripe's domain association file from the domain before expecting it to go active
- **Destroy only forgets** -- the domain stays registered and enabled in Stripe; apply `enabled: false` first to turn wallets off

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.domainName` | The host your checkout page runs on | Your DNS |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **02-wallets-paused** -- a registration with wallets turned off

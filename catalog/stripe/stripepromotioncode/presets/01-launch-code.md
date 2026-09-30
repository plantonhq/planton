# Launch Code

This preset is the code customers type, LAUNCH25, to redeem a launch coupon: first-time customers only, 500 redemptions at most, and none after the end of 2026. The coupon is named by reference, so one chart declares the whole campaign.

## When to Use

- A public launch or campaign code on a coupon declared beside it
- Any code you want reviewed like code instead of typed into the Dashboard

## Key Configuration Choices

- **Code** (`code: LAUNCH25`) -- letters, digits and dashes; a typo like `LAUNCH 25` is refused before anything runs
- **Limits** (`maxRedemptions`, `expiresAt`, `firstTimeTransaction`) -- can't exceed the coupon's own limits; changing them creates a new code
- **Expiry** (`expiresAt`) -- Unix seconds; 1798761599 is 2026-12-31 23:59:59 UTC
- **Destroy deactivates** -- the code stops working, and the same code can be declared again later

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.coupon.valueFrom.name` | The name of your StripeCoupon | Its manifest's `metadata.name` |
| `spec.code` | What customers type | Your campaign |
| `spec.expiresAt` | When the code stops working, in Unix seconds | `date -u -d <date> +%s` on Linux |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **02-first-order-minimum** -- a code that needs a minimum order total

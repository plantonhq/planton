# Minimal

A named service token with Cloudflare's one-year default duration. Capture `client_id` and `client_secret` from the outputs in the same change -- the secret is never readable again except on rotation.

## When to Use

- CI jobs, deploy bots, and service-to-service calls into Access-protected apps
- First machine credential on an account
- A starting point before you decide on `forever` or a shorter duration

## Key Configuration Choices

- **No duration set** -- Cloudflare defaults to 8760h (one year). Add `duration: forever` only when you have a rotation plan.
- **No rotation pair** -- leaving both version and expiry unset is the normal non-rotating state (Cloudflare treats the secret as version 1).
- **Capture the secret now** -- import cannot recover it.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `account_id` | The Cloudflare account that owns this token | Cloudflare Dashboard -> Overview -> API section |

## Related Presets

- **02-rotation** -- mint a new secret without changing the token ID

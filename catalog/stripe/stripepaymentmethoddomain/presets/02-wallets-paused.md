# Wallets Paused

This preset keeps a domain registered while turning its wallet buttons off. It is the state to apply before deleting a registration, because deleting it in Planton leaves the domain registered and still enabled in Stripe.

## When to Use

- Retiring a checkout domain
- Pausing wallets on a domain while keeping its registration

## Key Configuration Choices

- **Enabled** (`enabled: false`) -- updates in place; set it back to true to resume
- **Retiring a domain** -- apply this, then delete the resource

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.domainName` | The host whose wallets pause | Your DNS |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-checkout-domain** -- a registration with wallets on

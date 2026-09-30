# Free Express

This preset is free shipping that arrives in 1 to 2 business days, for a promotion or a premium plan.

## When to Use

- A free-shipping option beside a paid standard one
- A promotion that ships everything free

## Key Configuration Choices

- **Free** (`amount: 0`) -- a zero amount is a real rate customers can choose
- **Window** (`deliveryEstimate`) -- shown in Checkout; changing it creates a new rate
- **Tax** -- left unset, so the account's default tax behavior applies

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.displayName` | The name customers see | Your shipping policy |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-standard-shipping** -- a paid option in two currencies

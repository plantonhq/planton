# Ten Off Once

This preset takes 10 dollars off a customer's first invoice, with the same discount in euros and pounds so customers paying in their own currency get it too.

## When to Use

- A first-order discount for new customers
- A goodwill credit your support team hands out through a promotion code

## Key Configuration Choices

- **Amount** (`amountOff: 1000`, `currency: usd`) -- in cents; changing either creates a new coupon
- **Other currencies** (`currencyOptions`) -- update in place
- **Once** (`duration: once`) -- the first invoice only
- **No limits** -- add `maxRedemptions` or `redeemBy` to cap it, or limit it per code with a StripePromotionCode

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.amountOff`, `spec.currencyOptions` | Your discount, in each currency's smallest unit | Your pricing page |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-25-percent-three-months** -- a percentage off a plan for three months

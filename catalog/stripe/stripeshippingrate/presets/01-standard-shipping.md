# Standard Shipping

This preset is a 5 dollar shipping option (4.50 euros) that arrives in 3 to 5 business days, with tax added on top at Stripe's shipping tax code.

## When to Use

- The default shipping option on a payment link or in Checkout
- Any shipping price you want the same in every environment

## Key Configuration Choices

- **Amount** (`fixedAmount`) -- in cents; changing it creates a new rate, and payment links that offer it are replaced
- **Window** (`deliveryEstimate`) -- shown to customers; changing it creates a new rate
- **Tax** (`taxBehavior: exclusive`, `taxCode: txcd_92010001`) -- tax is added on top; once set, the behavior can't change
- **Destroy deactivates** -- new purchases can't choose it; orders already placed keep it

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.fixedAmount` | Your shipping price, in each currency's smallest unit | Your shipping policy |
| `spec.deliveryEstimate` | Your delivery window | Your carrier's service level |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **02-free-express** -- free next-day shipping

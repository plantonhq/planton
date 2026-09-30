# Monthly Subscription

This preset charges 49 dollars a month for a plan, with euro and pound prices so customers pay in their own currency. Your application finds it by the lookup key `pro-monthly`, and `transferLookupKey` lets an amount change keep that key: Stripe prices can never change their amount, so a new amount creates a new price first, moves the key to it, and archives the old one. Existing subscribers keep the old amount.

## When to Use

- A per-seat or flat monthly plan sold through Checkout or the customer portal
- Any price your code looks up by a stable name rather than an id

## Key Configuration Choices

- **Amount** (`unitAmount: 4900`) -- in cents; changing it replaces the price
- **Lookup key** (`lookupKey`, `transferLookupKey`) -- the key moves to the replacement price in the same call
- **Currencies** (`currencyOptions`) -- update in place; Stripe does not report them back, so change them only here
- **Tax** (`taxBehavior: exclusive`) -- tax is added on top; once set, Stripe refuses to change it
- **Destroy archives** -- subscribers keep paying; new purchases are refused

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.product.valueFrom.name` | The name of your StripeProduct | Its manifest's `metadata.name` |
| `spec.unitAmount`, `spec.currencyOptions` | Your prices, in each currency's smallest unit | Your pricing page |
| `spec.lookupKey` | The name your application retrieves the price by | Your checkout code |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **02-graduated-usage** -- a tiered price that gets cheaper per unit as usage grows

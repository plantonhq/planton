# One-Time Purchase

This preset is a payment page for a one-time purchase: one price, promotion codes allowed, a customer record and an invoice for every buyer, and Stripe's confirmation page with your message. Your site reads the page's address from `status.outputs.url` by reference.

## When to Use

- Selling a download, a course or a single product from a static site
- A purchase page declared beside its product and price

## Key Configuration Choices

- **Takes payments on apply** -- anyone with the address can pay from the moment it is applied
- **Price** (`lineItems`) -- by reference; changing the price or quantity creates a new link with a new address, and the old one shows `inactiveMessage`
- **Invoices** (`invoiceCreation`) -- a paid invoice for every buyer
- **Destroy deactivates** -- the address shows `inactiveMessage`; payments already taken stay

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.lineItems[0].price.valueFrom.name` | The name of your StripePrice | Its manifest's `metadata.name` |
| `spec.afterCompletion.hostedConfirmation.customMessage` | What buyers see after paying | Your copy |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **02-subscription-with-trial** -- a subscription page with a free trial

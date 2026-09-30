# Subscription Lifecycle

This preset delivers every change to a subscription -- created, updated, cancelled, trial ending -- plus the paid and failed invoices that move it between states. It suits a billing service that mirrors each customer's subscription in its own database instead of asking Stripe on every request.

## When to Use

- A service that grants or revokes access as a subscription changes state
- A service that warns customers before a trial ends or after a payment fails

## Key Configuration Choices

- **Events** (`enabledEvents`) -- the subscription events plus `invoice.paid` and `invoice.payment_failed`, which carry the payment outcome a subscription change waits on
- **Separate endpoint** -- a second endpoint with its own signing secret, so the billing service and the rest of the application verify with different secrets
- **Destroy deletes** -- the endpoint and its secret are gone; a new endpoint gets a new secret

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.url` | Your billing service's webhook route | Your service's routing |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-billing-events** -- checkout, paid invoices and refunds for one application route

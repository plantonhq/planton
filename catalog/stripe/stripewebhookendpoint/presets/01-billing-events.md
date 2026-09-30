# Billing Events

This preset delivers the three events an application that sells through Stripe Checkout most often acts on: a completed checkout (fulfill the purchase), a paid invoice (extend or issue what a subscription buys), and a refunded charge (take back what was refunded). Nothing else is delivered, so the receiving route never has to acknowledge traffic it ignores.

## When to Use

- An application that sells one-off purchases or subscriptions through Checkout and keeps its own record of what each customer owns
- As the starting point for any endpoint: add events as the receiving code learns to handle them

## Key Configuration Choices

- **Events** (`enabledEvents`) -- exactly the events the route handles; `*` would deliver every event the account emits
- **Signing secret** (`status.outputs.secret`) -- captured at creation; the receiving service reads it by reference and verifies every delivery's `Stripe-Signature` header with it
- **API version** (`apiVersion`) -- left unset, so events render in the account's default version and the endpoint never needs replacing for it
- **Destroy deletes** -- the endpoint and its secret are gone; a new endpoint gets a new secret

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.url` | Your application's webhook route | Your application's routing |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **02-subscription-lifecycle** -- every change to a subscription, for applications that mirror subscription state

# Cancel at Once, with a Credit

This preset lets customers end their subscription the moment they ask, and credits the time they paid for and won't use to their next invoice, after telling you why. They can also update their details and payment methods and download past invoices.

## When to Use

- A product where access should stop when the customer cancels, and staying until the period ends would feel like being held
- Usage-light subscriptions where fairness matters more than keeping the last weeks of revenue

## Key Configuration Choices

- **Cancel at once** (`subscriptionCancel.mode: immediately`) -- the subscription ends when the customer confirms
- **Credit the unused time** (`prorationBehavior: create_prorations`) -- the difference lands on the customer's next invoice as a credit
- **Ask why** (`cancellationReason`) -- Stripe's fixed reasons, shown before the customer confirms
- **No subscription changes** -- letting customers switch plans or change quantities can't be declared yet: Stripe requires switchable products whenever that feature is on, and the pinned Stripe provider can't hold them, so validation refuses it
- **Destroy deactivates** -- Stripe keeps the configuration, inactive, forever

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.defaultReturnUrl` | Your application's billing page | Your application's routing |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-self-serve-at-period-end** -- the same portal, with cancellation taking effect when the paid period ends

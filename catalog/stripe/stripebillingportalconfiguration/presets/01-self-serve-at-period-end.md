# Self-Serve Portal, Cancel at Period End

This preset gives customers the portal most subscription businesses start with: they can update their email, address and tax id, manage payment methods, download past invoices, and cancel -- with the cancellation taking effect when the period they paid for ends, after telling you why. Subscription changes stay off: they can't be declared on the pinned provider.

## When to Use

- A subscription product where customers may cancel on their own but keep what they paid for
- As the portal every environment runs, so what a customer may do is the same in test and live

## Key Configuration Choices

- **Cancellation** (`subscriptionCancel`) -- `mode: at_period_end`, so no proration is needed; the reasons asked are a subset of Stripe's fixed list
- **Customer details** (`customerUpdate.allowedUpdates`) -- email, address and tax id; name and phone stay as your application set them
- **Return link** (`defaultReturnUrl`) -- where the portal's back link leads when the session names none
- **A configuration of its own** -- the application passes `status.outputs.id` as `configuration` when it opens a portal session; the account's default configuration is never touched
- **Destroy deactivates** -- Stripe keeps the configuration, inactive, forever

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.defaultReturnUrl` | Your application's billing page | Your application's routing |
| `spec.businessProfile.privacyPolicyUrl` | Your privacy policy | Your website |
| `spec.businessProfile.termsOfServiceUrl` | Your terms of service | Your website |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **02-cancel-at-once-with-credit** -- cancellation that takes effect at once, crediting the unused time

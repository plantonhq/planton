# Subscription With a Trial

This preset is a sign-up page for a monthly plan with a 14-day free trial: no card needed to start, promotion codes allowed, your terms accepted at checkout, and a redirect to your site afterwards.

## When to Use

- A pricing page's "Start free trial" button on a static site
- A sign-up page declared beside its plan and price

## Key Configuration Choices

- **Trial** (`subscriptionData.trialPeriodDays: 14`, `paymentMethodCollection: if_required`) -- no card at sign-up; a trial that ends without one is cancelled
- **Terms** (`consentCollection.termsOfService: required`) -- needs your terms of service URL in the Dashboard's public details first; changing consent collection creates a new link
- **Moving to a new price** -- creates a new link with a new address before the old one is deactivated; your site follows by reading `status.outputs.url` by reference
- **Redirect** (`afterCompletion.redirect.url`) -- `{CHECKOUT_SESSION_ID}` becomes the session's id

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.lineItems[0].price.valueFrom.name` | The name of your StripePrice | Its manifest's `metadata.name` |
| `spec.afterCompletion.redirect.url` | Your welcome page | Your site |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-one-time-purchase** -- a one-time purchase page

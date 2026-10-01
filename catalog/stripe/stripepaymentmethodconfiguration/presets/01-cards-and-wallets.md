# Cards and Wallets

This preset offers cards plus the three one-tap wallets -- Apple Pay, Google Pay and Link -- and leaves every other method to Stripe's default for the account. Wallets appear only on devices and browsers that support them, so a customer never sees a button that cannot work.

## When to Use

- A product selling worldwide by card that wants the fastest checkout on phones
- As the configuration a checkout and the customer portal share, so both offer the same methods

## Key Configuration Choices

- **Methods on** (`card`, `applePay`, `googlePay`, `link`) -- each set `"on"`, quoted so that YAML 1.1 tools (PyYAML, older linters), which read a bare `on` as a boolean, see the same value Planton does
- **Everything else** -- left out, so Stripe's default for the account applies; set a method to `"off"` to hide it everywhere this configuration is named
- **A configuration of its own** -- the application passes `status.outputs.id` as `payment_method_configuration` when it creates a Checkout Session; the account's default configuration is never touched
- **Destroy deactivates** -- Stripe keeps the configuration, inactive, forever

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **02-cards-only** -- cards and nothing else

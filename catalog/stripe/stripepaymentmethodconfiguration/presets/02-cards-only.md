# Cards Only

This preset offers cards and turns off the wallets and buy-now-pay-later methods Stripe would otherwise show for eligible payments. It suits businesses whose reconciliation, refunds or compliance process handles card payments only.

## When to Use

- A business that must reconcile every payment as a card transaction
- A regulated product where deferred-payment methods are not allowed

## Key Configuration Choices

- **Card on** (`card: "on"`) -- quoted so that YAML 1.1 tools (PyYAML, older linters), which read a bare `on` as a boolean, see the same value Planton does
- **Wallets and deferred payments off** (`applePay`, `googlePay`, `link`, `cashapp`, `klarna`, `affirm`, `afterpayClearpay`) -- set `"off"` explicitly, since a method left out keeps Stripe's default
- **Other local methods** -- left out; Stripe shows them only where the account has turned their capability on
- **Destroy deactivates** -- Stripe keeps the configuration, inactive, forever

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-cards-and-wallets** -- cards plus Apple Pay, Google Pay and Link

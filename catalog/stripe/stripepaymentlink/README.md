# StripePaymentLink

Declares a Stripe-hosted payment page at a public address: the prices it sells, what it collects from the buyer, and where the buyer goes after paying. See Stripe's [Payment Links](https://docs.stripe.com/payment-links).

## When to Use

- **A sales page next to its prices**: the link names its StripePrices by reference, and your site reads the page's address from `status.outputs.url` by reference.
- **Selling from a static site** without writing Checkout code.

Declare a link here only if nothing else owns it. When your application or the Dashboard already manages your payment links, leave them there.

**The link takes payments from the moment it is applied**: anyone with its address can pay for the declared prices. Applying moves no money and creates no customer; buyers do both, later.

## Quick Start

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripePaymentLink
metadata:
  name: pro-monthly-link
  org: acme-corp
  env: production
spec:
  lineItems:
    - price:
        valueFrom:
          name: pro-monthly
      quantity: 1
  allowPromotionCodes: true
  subscriptionData:
    trialPeriodDays: 14
```

## Fields

| Field | Description |
|---|---|
| `lineItems` | 1 to 20 prices sold, by reference to StripePrice, each with a quantity. A price or quantity change **replaces**; adjustable quantities change in place |
| `optionalItems` | Up to 10 extra prices the buyer may add |
| `afterCompletion` | `hostedConfirmation` (Stripe's page, with a message) or `redirect` (your site) |
| `allowPromotionCodes`, `consentCollection` | Promotion codes; terms and marketing consent (**replaces**) |
| `subscriptionData` | Trial days, trial-end behavior, invoice issuer, metadata; `description` **replaces** |
| `shippingAddressCollection`, `shippingOptions` | Shipping countries; shipping rates by reference to StripeShippingRate (**replaces**) |
| `customFields`, `customText` | Up to 3 extra questions; extra text on the page |
| `billingAddressCollection`, `nameCollection`, `phoneNumberCollection`, `taxIdCollection` | What the page collects |
| `automaticTax`, `invoiceCreation`, `customerCreation`, `paymentIntentData`, `paymentMethodCollection`, `paymentMethodOptions`, `paymentMethodTypes`, `submitType`, `currency` | Tax, invoices, customers and the payment itself |
| `restrictions` | Deactivate after a number of completed payments |
| `applicationFeeAmount`, `applicationFeePercent`, `onBehalfOf`, `transferData` | Connect: where buyers' money goes. **Replaces** |
| `active`, `inactiveMessage` | Whether it takes payments, and what its address shows when it doesn't |
| `metadata` | Key-value pairs stored on the link |

## Key Behaviors

- **A link can't change which prices it sells.** Changing a line item's price or quantity (or their order) creates a new link with a new address before the old one is deactivated. Your site follows by reading `status.outputs.url` by reference.
- **Destroy deactivates the link.** Its address shows `inactiveMessage`; payments already taken stay.
- **Connect fields decide where future buyers' money goes**: to a connected account, less the platform's fee.
- **Runs on OpenTofu only.** Planton refuses any other engine for this kind before anything runs.

## Outputs

| Output | Description |
|---|---|
| `id` | The payment link's Stripe id (`plink_...`); it changes when the link is replaced |
| `url` | The page's public address; it changes when the link is replaced |
| `active` | `false` once deactivated |

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

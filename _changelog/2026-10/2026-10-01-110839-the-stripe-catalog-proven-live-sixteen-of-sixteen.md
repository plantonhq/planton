# The Stripe Catalog, Proven Live: Sixteen of Sixteen

**Date**: October 1, 2026
**Type**: Feature
**Components**: StripeBillingPortalConfiguration, StripePromotionCode; `e2e/framework/runner` (one test fixture); `pkg/iac/importmap`

## Summary

**Every Stripe kind is proven live.** The customer portal configuration and the promotion code join the other fourteen, proven against the dedicated Stripe test sandbox on OpenTofu. The parity page reads 16 of 16 proven live.

**Two settings refused, honestly.** The pinned Stripe provider can't hold a portal's switchable products or a promotion code's minimums in other currencies: Stripe returns both only when a read expands them, the provider never expands, and so every create that sets them succeeds in Stripe and then fails. Validation now refuses them with a sentence that says why, instead of letting an apply fail every time:
- **Portal:** `subscriptionUpdate.products`, and the subscription-update feature itself. Stripe requires switchable products whenever that feature is on, even to change only quantities (verified live: "Missing required param: features[subscription_update][products]"). Customers can still cancel (at once or at period end, with a reason), update payment methods and details, and download invoices.
- **Promotion code:** `restrictions.currencyOptions`. A minimum in one currency still works.

A provider version that reads these fields back lets the rules go.

## What Changed

- **StripeBillingPortalConfiguration:**
  - the `subscription_update.products_not_held` and `subscription_update.enabled_not_held` rules replace `price_needs_products`;
  - the field and kind comments, README, GUIDE, catalog page, module README and presets say so; the plan-switching preset becomes **Cancel at Once, with a Credit**;
  - the `plan-switching` scenario becomes `every-feature` (every feature the provider can hold, no prerequisites);
  - `minimal` gains an in-place second act.
- **StripePromotionCode:**
  - the `restrictions.currency_options_not_held` rule replaces `currency_options_shape`;
  - the field comment, README, the first-order-minimum preset and the restricted scenario drop the euro minimum.
- **Spec tests:** each new rule is pinned by its id.
- **Runner fixture:** the price-on-product chain test reads the payment link's scenario.
- **Profiles:** both kinds `green` on `[tofu]`; the live-proven ledger and the parity page follow.

## Verification

- **Live:** both kinds' lanes against the sandbox, with the import round trip and the portal's second act.
- **Offline:**
  - `make protos`, including the Java stub build and the protovalidate-java CEL gate;
  - spec tests of both kinds, the runner, presetvalidity, catalogpage, providerparity and importmap;
  - `make generate-reference` and `make generate-provider-parity-report`.

# Stripe Discounts, Usage Metering, Shipping, Tax Rates and Payment Links

**Date**: September 30, 2026
**Type**: Feature
**Components**: StripeCoupon, StripePromotionCode, StripeShippingRate, StripeTaxRate, StripeBillingMeter, StripePaymentLink, StripePrice; `catalog/stripe/aa_e2e`, `catalog/stripe/aa_import`, `e2e/stripe`, `pkg/providerparity`, `pkg/catalogkindreflect`, `pkg/explain/refgen`, `pkg/iac/importmap`; the component forge rule and flow rules 012 and 014

## Summary

**The rest of a Stripe sales setup as files.** A platform engineer can now declare a discount and the codes customers type to redeem it, shipping options, manual tax rates, a usage meter with threshold alerts, and a Stripe-hosted payment page that sells declared prices. Every link between them is a reference, so nobody pastes an id: a code names its coupon, a coupon names the products it discounts, a metered price names its meter, and a payment page names its prices and shipping options.

**Usage pricing that can't drift.** A metered StripePrice now names its StripeBillingMeter by reference, and an application's deployment can read the meter's event name the same way, so the name the application sends and the name the meter counts are always the same. A metered price without a meter is now refused before anything runs; Stripe refuses one at every API version since `2025-03-31.basil`.

**A payment page whose address follows its prices.** Stripe can't change which prices a payment link sells, and the pinned provider never stores a line item's quantity. The module tracks both and replaces the link when either changes: the new link, with its new address, exists before the old one is deactivated, and a site that reads `status.outputs.url` by reference follows it on the same apply.

**Destroy truth for folded children.** The live-test harness now proves what destroy leaves for a kind's folded children too: a meter's alerts are still present (the provider only forgets them), and a product's feature links and a Radar list's items are gone.

## What Changed

- **StripeCoupon** (10102, `stpcoup`): `percentOff`, or `amountOff` with `currency` and `currencyOptions`; `duration` and `durationInMonths`; `maxRedemptions`, `redeemBy`; `appliesToProducts` by reference to StripeProduct; `name`, `metadata`. Everything but the name, metadata and currency options replaces it. Destroy deletes it; customers who already applied it keep their discount.
- **StripePromotionCode** (10103, `stppc`, prerequisite StripeCoupon): `coupon` by reference; `code` held to Stripe's letters, digits and dashes; one customer; expiry, cap, and restrictions (first-time customers, a minimum order with other currencies). Only `active` and `metadata` change in place. Destroy deactivates it.
- **StripeShippingRate** (10105, `stpsr`): `displayName`, `fixedAmount` (with other currencies), `deliveryEstimate`, tax behavior and code, `active`, `metadata`. Destroy deactivates it.
- **StripeTaxRate** (10106, `stptxr`): `displayName`, `percentage` and `inclusive` (both replace), country, state, jurisdiction, description, the provider's 14 tax types, `active`, `metadata`. Destroy deactivates it; it still applies wherever it is already used.
- **StripeBillingMeter** (10108, `stpbm`): display name (the only in-place field), event name, aggregation, customer mapping, value key, time window, and `alerts` keyed by title, each its own `stripe_billing_alert`. Outputs include `event_name` and `alert_ids`. Destroy deactivates the meter; alerts are only forgotten and stay active in Stripe.
- **StripePaymentLink** (10109, `stppl`, prerequisite StripePrice): 1 to 20 line items and up to 10 optional items by reference to StripePrice, shipping options by reference to StripeShippingRate, and every other block the provider offers, Connect fields included. Stripe's "type" discriminators are written by the module from the block that is set. A `terraform_data` tracker of each line item's price and quantity replaces the link on a change, with `create_before_destroy`; the tracker is judged `internal` in parity and recorded as not importable. Outputs include the page's `url`. Destroy deactivates it.
- **StripePrice:** `recurring.meter` is a reference to StripeBillingMeter, and the new rule `recurring.metered_needs_meter` refuses a metered price without one. The graduated-usage preset, GUIDE, README and catalog page name the meter by reference, and a new `metered-usage` scenario deploys a meter as its prerequisite.
- **Each new kind ships its full anatomy:** spec tests, the OpenTofu module pinned at `0.3.0`, README, catalog page, GUIDE, logo, cost and control profiles, a permissions manifest, an import map, a parity manifest at total accounting, two presets, and E2E assets (`pending_proof`), with prerequisite install profiles for the coupon, shipping rate and meter.
- **The Stripe harness:** verifiers for all six kinds; the deactivated check reads a meter's `status`; a kind that folds children proves each child's destroy truth from its map output (alerts forgotten, feature links and Radar items deleted); six entrypoints; and eight import-catalog rows, one of them the payment link's tracker recorded as not importable.
- **Catalog wiring:** seven resources leave the dispositions ledger; the registry snapshot, kind map, references and the Stripe parity page are regenerated.
- **Teaching:**
  - flow 014's Stripe id-prefix line: when two Stripe objects share initials, each takes Stripe's own id prefix;
  - flow 012, the kind forge rule and the import-map README: a change the provider accepts but can't send (an update with no API field, a write-only value) forces a replacement through a `terraform_data` tracker, judged internal and never imported;
  - the forge rule's verifier section: inactive read the vendor's way, and folded children proven like their parent;
  - two questions join the skill's eval bank: changing the price a payment link sells, and how an application learns its meter's event name.

## Verification

- **Offline:**
  - `buf lint`, `make protos` (including the Java stub build and the protovalidate-java rule gate), the kind map, the registry snapshot, `make generate-reference` and `make generate-provider-parity-report`;
  - spec tests for all seven kinds;
  - `validate-manifest` on every preset, E2E manifest, token-expanded scenario and prerequisite profile;
  - `tofu fmt`, `init` and `validate` for each module, and offline plans for every manifest and scenario shape;
  - for the payment link, offline plans against a seeded state: an unchanged manifest plans nothing, a changed price or quantity plans a replacement (create before destroy), a changed adjustable quantity or metadata plans an in-place update, and a state without the tracker (the shape after an import) plans only the tracker's creation;
  - `module verify --provisioner tofu` with engine validation for seven kinds; `secret-coverage --check`, `validate-refs --check`, `provider-parity --kind` for each kind and `--check`;
  - `go test` for the harness and its verifiers, the runner (including the catalog-wide fixture-integrity check), catalogkindreflect, providerparity, cataloglogo, catalogpage, presetvalidity, refcheck, secretcoverage, importmap, permissions, e2e/profile, outputs, cost and control profiles, protodocs, explain and refgen;
  - the E2E package compiles and vets under the `e2e` tag; `defspack`.
- **Red-proofed:** the payment link's replace trigger (without it, a price change plans an impossible in-place update and a quantity change never reaches the link), the child verifiers (a forgotten alert, a surviving feature link), the meter's status-form check, the harness's child wiring, and the new reference format and metered-price rules.
- **Not run live:** no Stripe lane has run; the profiles stay `pending_proof`. Whether Stripe accepts dashes in a meter's event name, and the restricted-key label for alerts, are confirmed on the first live run.
- **Known, not from this change:** the anatomy gate fails on `catalog/kubernetes/kubernetesplantonplatform/v1alpha1/sizing_gate_test.go`.

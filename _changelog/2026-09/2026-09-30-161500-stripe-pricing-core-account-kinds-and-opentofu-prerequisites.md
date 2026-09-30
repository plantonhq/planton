# Stripe Pricing Core, Four Account Kinds, and OpenTofu Prerequisites

**Date**: September 30, 2026
**Type**: Feature
**Components**: StripeEntitlementFeature, StripeProduct, StripePrice, StripeEventDestination, StripePaymentMethodDomain, StripeRadarValueList, StripeBillingPortalConfiguration; `e2e/framework/runner`, `catalog/stripe/aa_e2e`, `catalog/stripe/aa_import`, `e2e/stripe`, `pkg/providerparity`, `pkg/crkreflect`, `pkg/explain/refgen`; the component forge rule and flow rules 012 and 014

## Summary

**A Stripe catalog declared as files.** A platform engineer can now declare the capabilities a customer can be entitled to, the products that grant them, and the prices on those products -- flat, per unit, tiered, metered, customer-chosen, one-time or recurring, in several currencies. Changing a price's amount just works: Stripe never changes a price's amount, so Planton creates the new price first, moves the lookup key to it when asked, and archives the old one, while existing subscribers keep theirs. The provider's inline objects (a price created inside a product, a product inside a price) are not offered, because Planton could never track or archive them.

**Four account kinds.** Where v2 events go (a webhook, Amazon EventBridge or Azure Event Grid, thin or snapshot), which domains may show wallet buttons, and Radar block and allow lists whose items are checked against the list's type before Stripe sees them. Each kind states what Stripe does on delete, and the answers differ: archived and kept, deleted, or only forgotten while Stripe keeps it live.

**Prerequisites on their own engine.** The E2E harness now deploys each prerequisite on its kind's engine, so an OpenTofu-only kind can be a prerequisite. The portal's plan-switching scenario no longer needs a hand-made product and price: it names StripePrice, and the harness deploys a product and a price on it through OpenTofu.

## What Changed

- **StripeEntitlementFeature** (10104, `stpef`): `lookupKey` (replaces), `name`, `metadata`. Destroy archives it.
- **StripeProduct** (10100, `stpprod`): name, description, `active`, `type` (replaces), images, marketing features, package dimensions, shipping, statement descriptor and unit label (a service only), tax code, URL, metadata, and `features` -- references to StripeEntitlementFeature, each attached as its own link keyed by the feature id. Outputs include `product_feature_ids`, so a hand-built product imports with its feature links. Destroy archives it.
- **StripePrice** (10101, `stpprice`, prerequisite StripeProduct): `product` by reference, currency, one amount shape per billing scheme (unit amount, decimal amount or customer-chosen; or graduated or volume tiers ending in `inf`), `transformQuantity`, `recurring` (interval up to three years, usage type, meter, trial days), `currencyOptions` keyed by currency, `lookupKey` and `transferLookupKey`, nickname, tax behavior, `active`, metadata. The module uses `create_before_destroy`, so a replacement never leaves a gap. Destroy archives it.
- **StripeEventDestination** (10001, `stped`): exactly one destination block (`webhookEndpoint`, `amazonEventbridge`, `azureEventGrid`), which is the destination's type; thin or snapshot payload; events, routing and snapshot version. Outputs carry a webhook's sensitive signing secret, the EventBridge partner source's name (ready for an event bus) and status, and the Event Grid partner topic. Destroy deletes it.
- **StripePaymentMethodDomain** (10004, `stppmd`): `domainName`, `enabled`, and per-wallet status and error-message outputs. Destroy only forgets it; the GUIDE says to disable before deleting.
- **StripeRadarValueList** (10005, `stprvl`): alias, name, item type (replaces), items (each its own object, keyed by value) validated per type (country, email, IP address, card BIN, customer id, account id), metadata. Destroy deletes it.
- **StripeBillingPortalConfiguration:** the switchable products and prices are references to StripeProduct and StripePrice, so a replaced price reaches the portal on its next apply.
- **Each new kind ships its full anatomy:** spec tests, the OpenTofu module pinned at `0.3.0`, README, catalog page, GUIDE, logo, cost and control profiles, a permissions manifest, an import map, a parity manifest at total accounting, two presets, and E2E assets (`pending_proof`) with prerequisite install profiles.
- **E2E prerequisites deploy on their kind's engine:** Pulumi when the kind has a Pulumi module (every existing chain, unchanged), otherwise the HCL lane in a disposable working copy with local state. A failed apply is still torn down, and a failed destroy keeps the working copy and names it.
- **The Stripe harness:** reads Stripe with the same `Stripe-Context` and `Stripe-Version` headers the pinned provider sends (it previously sent `Stripe-Account` and no version); reads v1 and v2 objects; and gains a "forgotten" verifier that proves an object is still in Stripe after a destroy that makes no call.
- **Catalog wiring:** eight resources leave the dispositions ledger; the registry snapshot, kind map, references and the Stripe parity page are regenerated.
- **Teaching:**
  - the component forge rule and flow 012 gain the law that an argument creating an untracked inline object is never offered;
  - flow 014's Stripe id-prefix line gains the one-word rule;
  - the forge rule and the E2E README teach prerequisites on their own engine and the three destroy-verifier shapes;
  - two questions join the skill's eval bank: changing a price's amount, and who should own a product.

## Verification

- **Offline:**
  - `make protos` (including the Java stub build and the protovalidate-java rule gate), `make generate-cloud-resource-kind-map`, the registry snapshot, `make generate-reference` and `make generate-provider-parity-report`;
  - spec tests for all seven kinds;
  - `validate-manifest` on every preset, E2E manifest, token-expanded scenario and prerequisite profile;
  - `tofu fmt`, `init` and `validate` for each module, and offline plans for every manifest shape (tiered prices with currency options, a product granting two features, EventBridge and Event Grid destinations);
  - `module verify --provisioner tofu` with engine validation;
  - `secret-coverage --check`, `validate-refs --check`, and `provider-parity --kind` for each kind and `--check`;
  - `go test` for the runner (including the catalog-wide fixture-integrity check), the harness, crkreflect, provisioner, providerparity, anatomy, cataloglogo, catalogpage, presetvalidity, refcheck, secretcoverage, importmap, permissions, e2e/profile, outputs, cost and control profiles, protodocs, explain and refgen;
  - the E2E package compiles and vets under the `e2e` tag; `defspack`.
- **Red-proofed:** the engine choice, a failed apply staying tracked, mixed-engine teardown, the product-to-price chain resolving its reference, the harness's headers, the forgotten verifier, and the reference-safe product format rule.
- **Not run live:** no Stripe lane has run; the profiles stay `pending_proof`.
- **Known, not from this change:** the anatomy gate fails on `catalog/kubernetes/kubernetesplantonplatform/v1alpha1/sizing_gate_test.go`.

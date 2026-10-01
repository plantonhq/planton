# Every Stripe Kind Proven Live, and Two Waiting on the Provider

**Date**: October 1, 2026
**Type**: Feature
**Components**: 14 Stripe kinds' modules, docs and E2E assets; `catalog/stripe/aa_e2e`, `catalog/stripe/aa_import`; `e2e/framework/runner`, `e2e/framework/provider`; `pkg/iac/importmap`; `.github/workflows/e2e-stripe.yaml`; the component forge rule, flow rule 012 and the E2E README

## Summary

**Proven against Stripe itself.** Fourteen of the sixteen Stripe kinds now run live on OpenTofu against a dedicated test sandbox: each is created, read back from Stripe, re-planned clean, imported blind with a zero diff, changed the way its guide says it changes (in place, or replaced with the old object left exactly as the kind's destroy leaves it), and destroyed with Stripe keeping, deactivating or deleting what the guide promises. The parity page reads 14 proven live. The billing portal configuration and the promotion code each prove their minimal scenario, but stay unproven as kinds: the pinned provider can't hold a portal's switchable products or a code's minimums in other currencies.

**What the live runs corrected.** Stripe leaves some fields out of what it returns unless a read expands them, rewrites some values, and fills some defaults, and each of those made a module plan wrong or the provider reject its own result. Every one is fixed at the root and taught where its reader lives:
- a coupon now carries its products and other-currency amounts in a tracker, so an import adopts it untouched and a change still replaces it; changing `currencyOptions` replaces a coupon, because Stripe refuses a new amount for a currency it already has;
- a Radar list sends string, email and country values in lower case, the way Stripe stores them;
- a billing meter always sends its payload keys;
- a graduated price never sends a zero unit amount beside a flat fee.

**Two acts every provider can use.** A second act now declares whether an upgrade keeps the object or replaces it, judged against the vendor's own ids, and the runner prepares it the way it prepares the first (tokens and references included). A new act deletes the object in the vendor's console and proves the GUIDE's recovery. Its first run caught a wrong sentence: a Radar list deleted in the Dashboard takes its items with it, so the recovery forgets both. Dates come from the lane's clock (`${E2E_UNIX_TIME_PLUS:<offset>}`), because Stripe refuses an expiry or a registration start more than five years ahead.

## What Changed

- **The runner:**
  - `planton.dev/e2e-out-of-band-delete: recreates | forget:<addresses>` adds OUT-OF-BAND-DELETE, DRIFT-PLAN, RECOVER and VERIFY-RECOVERED, with `provider.OutOfBandDeleter`; HCL lanes only;
  - a second act is found beside the authored scenario, its tokens expand to the first act's run id, scenario slug and lane clock, and its references resolve against the same prerequisites; the context carries `provider.FirstActManifestPathKey`, and `planton.dev/e2e-expect-upgrade: in-place | replaced` is harness vocabulary;
  - `${E2E_UNIX_TIME_PLUS:<offset>}` expands from each lane's own clock;
  - annotations are read from the YAML document, not the typed message, and the fixture-integrity gate expands run-clock tokens first, so a token in a number field never blocks either.
- **The Stripe harness:**
  - it judges every second act (it refuses one that declares nothing);
  - a replaced object must meet its kind's delete truth;
  - it checks signing secrets by shape and by digest, never by value;
  - it deletes outright-deletable kinds for the out-of-band act, through the client's one write;
  - a deleted object's `deleted: true` stub reads as absent.
- **Kinds:**
  - **Coupon:** the products and other-currency tracker; changing `currencyOptions` replaces it.
  - **Radar value list:** case-folded values; a precondition refuses values that differ only by case; the recovery forgets the items too.
  - **Billing meter:** both payload keys are always sent.
  - **Price:** a flat-fee tier's zero unit amount is dropped.
  - **Promotion code and tax registration:** field comments state Stripe's five-year limit; examples and presets use 1 January 2028.
  - **Tax registration:** the guide names Texas's one election.
  - **Portal:** the guide teaches that an account's first configuration becomes its default and can't be deactivated.
- **Import catalog:** write-normalized `currency_options` on prices, `fixed_amount.currency_options` on shipping rates, and `webhook_endpoint.url` on event destinations; the tracker row covers the coupon. The live-proven ledger gains the Stripe rows.
- **Scenarios:**
  - second acts for the webhook endpoint (in place, and an API-version replacement with a new secret), coupon (in place, and a currency replacement), payment-method configuration, price (replacement moving the lookup key), product, payment link (quantity replacement) and tax registration (expiry in place);
  - out-of-band scenarios for the webhook endpoint, event destination, coupon and Radar list;
  - tax registrations that start a minute after the lane and end seconds later, since Stripe holds one registration per place until it has ended.
- **Profiles:** 14 kinds `green` on `[tofu]`; the portal and promotion code stay `pending_proof` with the reason.
- **CI:** `e2e-stripe.yaml` gains a dispatch-only live job with the import round trip, reading the repository secret `STRIPE_API_KEY`.

## Verification

- **Live, against the sandbox:** every scenario of the 14 kinds, with the import round trip, in one final pass of the committed code, plus the minimal scenarios of the portal and promotion code. Tax registration re-ran after its registrations ended and found the places free.
- **Offline:**
  - runner and harness unit tests, each new rule red-proofed by mutation;
  - spec tests of the three kinds whose protos changed; scoped Go regeneration of those protos; `make generate-reference`; `make generate-provider-parity-report`.
- **Not proven:**
  - least-privilege key labels (the sandbox key holds every permission);
  - a signed delivery to a real receiver (the lanes' URLs are `example.com`).


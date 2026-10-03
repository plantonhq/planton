# Stripe Tax Registrations, and the Stripe Catalog Complete

**Date**: September 30, 2026
**Type**: Feature
**Components**: StripeTaxRegistration; StripePromotionCode's E2E assets; `catalog/stripe/aa_e2e`, `catalog/stripe/aa_import`, `e2e/stripe`, `pkg/providerparity`, `pkg/catalogkindreflect`, `pkg/explain/refgen`; the component forge rule and flow rule 014

## Summary

**Tax registrations as files.** A platform engineer can now declare where their Stripe account is registered to collect tax with Stripe Tax -- VAT in Germany under the EU's One-Stop Shop, sales tax in Texas -- one resource per registration, the same in every environment. The spec is flat: a country, a type, two dates, and the few options a country takes. Validation knows which of the 101 supported countries take which types, and refuses a province, state, jurisdiction or place-of-supply scheme where it does not belong.

**Honest about what Stripe keeps.** Stripe never deletes a registration, and the provider only forgets it, so the kind teaches the one way to stop collecting: set `expiresAt` and apply. Only the two dates change in place; any other change creates a new registration while the old one keeps collecting, so the guide teaches expiring the old one first. Removing an expiry from the file does not clear it, and a new registration's start must be now or later.

**The Stripe catalog is complete.** Every Stripe resource that is a declarable setting is now a kind or folds into one: 16 kinds. Every other resource carries its reason in the dispositions ledger. The card reader, reader locations, reader settings, file uploads and printed card designs join the money-moving resources as not offered: enrolling hardware with a single-use code can never be re-applied, in-store fleets and uploaded files are not settings a platform engineer declares, and the pinned provider cannot upload the logo a card design needs.

**No Stripe lane goes stale.** Dates a vendor requires in the future are literals far enough ahead: Stripe's E2E assets use 1 January 2100. The promotion code's lane carried 31 December 2026, which would have failed on New Year's Day.

## What Changed

- **StripeTaxRegistration** (10107, `stptxrg`): `country`, `type` (standard, simplified, the EU's `oss_union`, `oss_non_union` and `ioss`, Canada's `province_standard`, and the US state and local types), `activeFrom`, `expiresAt`, `placeOfSupplyScheme`, `province`, `state`, `jurisdiction` and `stateSalesTaxElections`. Sixteen CEL rules hold the fields' formats, and each type and option to the countries the pinned provider defines for it. The module writes the provider's single `country_options` block for the declared country, with only the parts the type uses. Outputs: `id` and `status` (scheduled, active or expired). Destroy only forgets the registration.
- **Full anatomy:** spec tests pinning each rule by its id, the OpenTofu module pinned at `0.3.0`, README, catalog page, GUIDE with the country list, logo, cost and control profiles, a permissions manifest, an import map, a parity manifest at total accounting, two presets (EU One-Stop Shop, US state sales tax) and E2E assets (`pending_proof`) with a verifier that proves the registration is still in Stripe after destroy.
- **Dispositions:** `stripe_terminal_reader`, `stripe_terminal_location`, `stripe_terminal_configuration`, `stripe_file`, `stripe_file_link` and `stripe_issuing_personalization_design` are deferred with their reasons; `stripe_tax_registration` leaves the ledger as modeled. The Stripe accounting reads 19 modeled, 27 deferred, 3 excluded; 16 of 16 kinds at total accounting.
- **The kind enum:** the Stripe band drops the Terminal and Issuing sub-bands, and the tax prefix comment names both "tr" objects.
- **StripePromotionCode:** its E2E manifest and `restricted` scenario expire on 1 January 2100.
- **StripeTaxRate:** its guide points accounts that use Stripe Tax to StripeTaxRegistration.
- **Teaching:**
  - flow 014: when two Stripe objects share initials, each takes Stripe's own id prefix with its vowels dropped (`stptxr` from `txr_`, `stptxrg` from `taxreg_`);
  - the kind forge rule: the base E2E manifest round-trips into its typed message, so no run token can stand in a number field, and a date a vendor requires in the future is a literal far enough ahead;
  - one question joins the skill's eval bank: changing and ending a tax registration.

## Verification

- **Offline:**
  - `buf lint`, `make protos` (including the Java stub build and the protovalidate-java rule gate), the kind map, the registry snapshot, `make generate-reference` and `make generate-provider-parity-report`;
  - spec tests; `validate-manifest` on every preset, E2E manifest and token-expanded scenario;
  - `tofu fmt`, `init` and `validate`, and offline plans for all five country shapes (an EU OSS registration, a standard registration with a place-of-supply scheme, a simplified registration, a Canadian province, and US state and local registrations with elections);
  - offline plans against a seeded state: an unchanged manifest plans nothing; a changed type plans a replacement; a set expiry and a moved start plan in-place updates; an expiry removed from the manifest plans nothing;
  - `module verify --provisioner tofu`; `secret-coverage --check`, `validate-refs --check`, `provider-parity --kind` and `--check`;
  - `go test` for the harness and its verifiers, the runner, catalogkindreflect, providerparity, anatomy, cataloglogo, catalogpage, presetvalidity, refcheck, secretcoverage, importmap, permissions, e2e/profile, outputs, cost and control profiles, protodocs, explain, refgen, skills and the tofu generators; the E2E package compiles and vets under the `e2e` tag; `defspack`; the catalog bundle and schemas build and verify.
- **Not run live:** no Stripe lane has run; the profiles stay `pending_proof`. The first live run confirms whether the sandbox needs Stripe Tax turned on, whether Stripe accepts a second scheduled registration for one place, the restricted-key label, and whether Stripe accepts dates in 2100 for promotion-code expiries and registration starts.

# Stripe Webhook Endpoint, Portal and Payment-Method Kinds, With a Live-Test Harness

**Date**: September 30, 2026
**Type**: Feature
**Components**: StripeWebhookEndpoint, StripeBillingPortalConfiguration, StripePaymentMethodConfiguration; the permissions schema (`iac/componentpermissions/v1`), `pkg/iac/provider/stripe/stripekey`, `pkg/iac/stackinput/providerenvvars`, `pkg/providerparity`, `pkg/catalogbundle`, `pkg/cataloglogo`, `catalog/stripe/aa_e2e`, `catalog/stripe/aa_import`, `e2e/stripe`, `e2e/framework/runner`; the component forge rule and flow rules 014 and 021

## Summary

**The first three Stripe kinds.** A platform engineer can now declare, as files, where a Stripe account delivers its events, what the customer portal lets customers do, and which payment methods checkout offers. Each kind runs on OpenTofu only and says in plain words what Stripe does on delete: an endpoint is deleted, while a portal or payment-method configuration is deactivated and kept by Stripe forever. The webhook endpoint's signing secret is captured at creation into a sensitive output, since Stripe returns it only then.

**A live-test harness that cannot touch real money.** The Stripe lanes are authored and offline-verified; they run later against a dedicated test sandbox. The harness refuses any key that is not a test-mode key before its first API call, and it checks destroy honestly: an endpoint must be gone, while a configuration must still read back inactive.

## What Changed

- **StripeWebhookEndpoint** (10000, `stpwh`): `url`, `enabledEvents` (Stripe's dotted spelling or `*`), `description`, `metadata`, and the replace-triggering `apiVersion` and `connect`. Its outputs are `id`, the sensitive `secret`, `status`, `url` and `application`.
- **StripeBillingPortalConfiguration** (10002, `stpbpc`): typed features with an enum for every closed set (cancellation mode and reasons, proration, billing-cycle anchor, trial behavior, deferred-change conditions), up to ten products with their prices and quantity range, the business profile, and the login page. Cross-field rules refuse prorations on a cancellation at period end, and price changes with no products to switch between. `paymentMethodUpdate.paymentMethodConfiguration` references a StripePaymentMethodConfiguration.
- **StripePaymentMethodConfiguration** (10003, `stppmc`): a typed preference (`on`, `off`, `none`) for each of the provider's 59 methods, `parent` for Connect, and an `availablePaymentMethods` output of the methods Stripe reports available.
- **Each kind ships its full anatomy:** spec tests, the OpenTofu module (the provider pinned exactly at `0.3.0`), README, catalog page, GUIDE, logo, cost and control profiles, a permissions manifest, an import map, a parity manifest at total accounting, two presets, and E2E assets (`pending_proof`).
- **Permissions:** the schema gains `StripePermissions` (field 8). Each group lists restricted-key permissions as a Dashboard resource label plus `read` or `write`. Stripe publishes no machine-readable permission inventory, so the conformance arm checks structure and spelling, and a live run under a least-privilege key upgrades entries from derived to proven.
- **One Stripe key-mode reader** (`pkg/iac/provider/stripe/stripekey`), shared by the credential loader and the harness.
- **The harness:**
  - `catalog/stripe/aa_e2e`: a REST client, the harness with its live-key refusal, and deleted and deactivated verifiers;
  - `catalog/stripe/aa_import/catalog.yaml`, for plain-id imports;
  - `e2e/stripe` with `TestStripe<Kind>_Tofu` entrypoints;
  - `make e2e-test-stripe` and a Stripe arm in `make e2e-test-component`;
  - a build-check workflow, `e2e-stripe.yaml`.
- **The E2E runner honors a kind's declared engines:** a lane whose binary the kind does not run on (`PLANTON_E2E_TF_BINARY=terraform` for an OpenTofu-only kind) is refused before any phase. `TerraformBinary()` is now the one place the HCL binary is chosen.
- **"Proven" follows the declaration:** a kind counts as proven when its green live runs exercised every module it ships. That means both for the ordinary kind, and only the HCL module for an OpenTofu-only kind. Every existing green profile validates both engines, so no count changes. The parity pages say "every IaC engine the kind runs on", and all five were regenerated.
- **Catalog wiring:**
  - the three kinds leave the dispositions ledger, where they are now counted as modeled;
  - Stripe joins the Terraform module release matrix and the logo gate's judged providers;
  - Stripe's public parity page is enrolled;
  - references and proto docs are regenerated.
- **Teaching:**
  - flow 014 gains the Stripe id-prefix pattern and the `provisioners` field;
  - flow 021 gains Stripe's brand terms, palette and object vocabulary;
  - the component forge rule teaches the `_Tofu` entrypoint name for OpenTofu-only kinds;
  - `architecture/component.md` lists the Stripe band.

## Verification

- **Offline:**
  - `make protos`, `make generate-cloud-resource-kind-map`, the registry snapshot, `make generate-reference` and `make generate-provider-parity-report`;
  - spec tests for all three kinds;
  - `validate-manifest` on every E2E manifest and preset;
  - `tofu fmt -check`, `init` and `validate`, plus an offline `tofu plan` from each E2E manifest (one object to add each);
  - `module verify --provisioner tofu` with engine validation;
  - `secret-coverage --check` and `validate-refs --check`;
  - `provider-parity --kind` for each kind (6/6, 7/7 and 62/62 arguments accounted) and `--check`;
  - `go test` for crkreflect, permissions, importmap, e2e/profile, secretcoverage, refcheck, catalogbundle, catalogpage, presetvalidity, outputs, cost profiles, control profiles, specpath, actioninventory, providerparity, moduleverify's secret outputs, protodocs, refgen, tofumodule, providerenvvars, the runner, and the harness;
  - the E2E package compiles and vets under the `e2e` tag;
  - `defspack`.
- **Red-proofed:** the harness's live-key refusal (without it, a fake live key reached Stripe's API), the runner's engine refusal, and the permissions arm.
- **Not run live:** no Stripe lane has run; the profiles stay `pending_proof`.
- **Known, not from this change:** the anatomy gate fails on `catalog/kubernetes/kubernetesplantonplatform/v1alpha1/sizing_gate_test.go` (a test file in a version directory, from an earlier commit).

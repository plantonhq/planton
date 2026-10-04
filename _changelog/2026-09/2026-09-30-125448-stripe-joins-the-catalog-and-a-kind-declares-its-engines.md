# Stripe Joins the Catalog, and a Kind Declares the Engines It Runs On

**Date**: September 30, 2026
**Type**: Feature
**Components**: the provider enum and kind band (`shared/catalogkind`), `catalog/stripe`, `pkg/iac/iacinput/providerenvvars`, `pkg/iac/iacinput/providerdetect`, `pkg/iac/iacinput/iacinputproviderconfig`, `pkg/providerparity`, `pkg/catalogkindreflect`, `pkg/anatomy`, `pkg/iac/provisioner`, `pkg/iac/tofu/tofumodule`, `pkg/iac/pulumi/pulumimodule`, `pkg/setdeploy`, `pkg/iac/moduleverify`, `pkg/iac/eject`, `pkg/e2e/profile`, the CLI's engine prompt; OpenFgaStore, OpenFgaAuthorizationModel, OpenFgaRelationshipTuple; the provider and catalog-kind forge rules

## Summary

**Stripe is provider 29.** Its config takes an API key, an optional Connect account, and a required mode: test or live. Stripe's Terraform provider applies whatever key it is handed to that key's account and has no guard of its own, so every deploy now refuses a key whose prefix belongs to the other mode, before the engine starts. Every one of the provider's 49 resources has a recorded disposition: 25 are planned kinds or planned compositions, 3 are superseded, and 21 are deliberately not offered because applying them moves money or creates a record of business that destroy does not undo. No Stripe kinds ship yet.

**A kind can now declare the engines it runs on.** Until now every kind was assumed to run on Pulumi, OpenTofu and Terraform. OpenFGA has no Pulumi provider, so its kinds shipped Pulumi modules that reported success while creating nothing. A kind now lists its engines in `kind_meta.provisioners`, and every surface that picks or runs an engine honors the list: it uses the kind's only engine when a manifest names none, and it refuses any other engine before anything runs, with one sentence that names the engines the kind runs on. OpenFGA declares OpenTofu and Terraform, and its three placeholder Pulumi modules are gone.

## What Changed

- **`stripe = 29`** with the band `10000–10999`. `StripeProviderConfig` has `api_key` (required; its secrecy is stated in the comment), `mode` (`StripeMode`, required) and `stripe_account` (optional). The credential loader emits `STRIPE_API_KEY` and `STRIPE_ACCOUNT`. When a key is present, it refuses an undeclared mode, a key of the other mode and a key that is neither secret nor restricted, naming the key's prefix and never the key. The provider detection, display name and credentials-required arms are wired.
- **Provider parity:** `stripe/stripe` is pinned exactly at `0.3.0`, and its schema artifact comes from OpenTofu's registry. The ledger `dispositions/stripe.yaml` records every resource with its reason, including what each planned kind's destroy does in Stripe (deleted, deactivated, or only removed from state). `catalog/stripe/provider-config-parity.yaml` holds the provider block at total accounting. The artifact test's provider-block floor is "at least one argument" (Stripe's provider takes two); the guard still catches a dropped block.
- **`kind_meta.provisioners`** (field 12, `repeated string` of IacProvisioner names, because the kind proto cannot import `shared/iac.proto` without a package cycle). `catalogkindreflect.Provisioners` and `RunsOn` read it, and a registry test holds every value to a real, non-duplicated engine name.
- **The anatomy gate holds the tree to the declaration:** a declared kind carries exactly the modules its engines run (`iac/pulumi` for pulumi, `iac/tf` for tofu or terraform), and a module for an undeclared engine is the new `module-for-undeclared-engine` violation. Undeclared kinds still owe both modules.
- **One policy, in `pkg/iac/provisioner`:** `ForManifest`, `Require`, `Allowed` and `ModuleFamily`.
  - The CLI's single-manifest commands and `planton init` resolve through it. A kind with one engine is not prompted for; a kind with several is prompted among only those.
  - The multi-manifest preflight refuses a label naming an undeclared engine.
  - The two execution entries refuse too: `tofumodule.Init`, where every OpenTofu and Terraform run starts, including the platform runner's, and `pulumimodule.GetPath`, where every Pulumi run the CLI makes resolves its module.
  - `planton module verify`, `planton module eject` and E2E profile discovery refuse an undeclared engine.
  - The duplicated engine-to-module-family switch in the CLI is now `ModuleFamily()`.
- **OpenFGA:** its kinds declare `["tofu", "terraform"]`. Their Pulumi placeholders are deleted and `openfga` leaves the Pulumi module release matrix. Its protos, READMEs and catalog page describe the engines it runs on and how to choose, instead of a `--provisioner` flag that does not exist.
- **Teaching:** the provider forge rule's stale steps are fixed (`ProviderConfigProto`, the display-name and credentials-required arms, Bazel deps, `putIfSet` and the empty-config test, where a credential guard belongs). The kind forge rule, flow rule 014, `architecture/`, `MODULE_PARITY.md`, the catalog bundle's comment and the multi-cloud-catalog skill describe kinds that declare fewer engines. The proto-docs index and the reference tree are regenerated; the Kubernetes workload references list OpenFGA's new enum comment.
- **Generator output:** gazelle, run by `make protos`, adds a missing test dependency to AzureFrontDoorOrigin's `v1alpha1/BUILD.bazel`.

## Verification

- **Offline:** `make protos`, `make generate-reference`; `go test` for providerenvvars, providerdetect, catalogkindreflect, providerparity, anatomy, provisioner, tofumodule, pulumimodule, setdeploy, moduleverify, eject, e2e/profile, catalogbundle, protodocs and refgen; `defspack`. The mode guard, the anatomy rule and every refusal were red-proofed: each test fails with its check removed.
- **The accounting:** `planton provider-parity --provider stripe --ga-schema stripe` reads 25 model-planned, 21 deferred, 3 excluded-deprecated, provider block 2 of 2 matched.
- **Not run live:** no Stripe kind exists yet; the refusals are proven in tests against OpenFGA's real declaration.
- **Known, not from this change:** the anatomy gate fails on `catalog/kubernetes/kubernetesplantonplatform/v1alpha1/sizing_gate_test.go` (a test file in a version directory, from an earlier commit).

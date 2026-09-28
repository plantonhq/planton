# A Sign-In Domain Proven End to End, and a Tenant That Sets Only Its Default Domain

**Date**: September 28, 2026
**Type**: Fix
**Components**: Auth0TenantSettings, Auth0PromptScreenPartials, the Auth0 and Cloudflare live-test harnesses, the Pulumi module rule

## Summary

**Every Auth0 kind is now proven live on both engines, the sign-in-domain chain included.** The chain is a custom domain, its CNAME in a real Cloudflare zone, and Auth0 reporting the domain verified, about four minutes end to end. The blind-import round-trip reaches zero diff on OpenTofu. All 17 Auth0 profiles are green.

The runs found one more defect in a kind and three in the harnesses, each fixed at its source.

## What Changed

- **`Auth0TenantSettings` with only a default domain now deploys.** The modules used to declare the tenant resource even when the spec set no tenant setting. The provider then sent Auth0 an empty tenant update, which Auth0 refuses (400 "Too few properties defined (0), minimum 1").
  - Both engines now declare `auth0_tenant` only when the spec declares a setting beyond the default domain. Otherwise they read the tenant's settings through the provider's tenant data source, which writes nothing, so the outputs still report what the tenant carries.
  - A `moved` block keeps an installed tenant's state at its new address, so it plans no replacement.
  - A module test runs the Pulumi program under mocks and asserts which resources each spec declares.
- **Screen partials need a page template as well as a custom domain** (measured live). Auth0 refuses them with a 403 for each that is missing. The GUIDE had said partials without a template are "stored and never shown"; the field comment and GUIDE now say what Auth0 does. The lanes bring both: a custom domain (it need not be verified), and a branding that sets the smallest page template Auth0 accepts.
- **The live-test harnesses:**
  - **A fixture that names a real zone by id no longer drags in a throwaway zone.** A scenario declares the zone resident (`planton.dev/e2e-resident-prerequisites`), and the rule's doc now covers account-level objects as well as cluster residents.
  - **The Cloudflare harness accepts a token scoped to zones.** The token proves its reach by seeing one of the account's zones when it cannot read the account itself. The custom-domain lanes' token is DNS edit on one zone, the least they need.
- **The Pulumi module rule** no longer says an out-of-range `Index` returns the zero value. That holds for the SDK's built-in arrays only. A provider-generated array output panics on an empty list, as a tenant lookup with no flags block showed, so an element of a list that can be empty is read with one bounds-checked applier.

## Follow-up

The catalog has 25 `Index(pulumi.Int(...))` calls on provider outputs. Each is safe only if its list is never empty after an apply, and each needs checking against that rule.

## Verification

- **Live, on both engines, one lane at a time:** Auth0CustomDomain, Auth0CustomDomainVerification, Auth0TenantSettings (minimal, rich, default-domain) and Auth0PromptScreenPartials (minimal, multi-screen), with the import round-trip on OpenTofu.
- **Offline:** `go test` for the Auth0 catalog, the Cloudflare harness, the e2e runner and `pkg/iac/importmap`; the Bazel build of the touched packages; `make generate-reference`.

# Every Auth0 Kind Proven Live, and One Spelling for a Zone's Name Servers

**Date**: September 28, 2026
**Type**: Fix, Breaking (output rename)
**Components**: The Auth0 catalog and its live tests, the Terraform variables generator, the live-test harness, AzureDnsZone, DigitalOceanDnsZone

## Summary

**The Auth0 kinds were run live, on both engines, against a dedicated test tenant, and the runs found six defects that no offline check could.** Each is fixed at its source with a test that fails on the old code. Seven kinds now carry a green profile, and blind-import round-trips reach zero diff for every Auth0 kind with an import map. **A DNS zone's name servers now have one output name across the catalog**: `nameservers`, which AzureDnsZone and DigitalOceanDnsZone adopt.

## Breaking change

`AzureDnsZone` and `DigitalOceanDnsZone` rename their `name_servers` output to `nameservers`, the name `AwsRoute53Zone`, `GcpDnsZone` and `CloudflareDnsZone` already use. The field numbers are unchanged. A reference to `status.outputs.name_servers` on either kind becomes `status.outputs.nameservers`.

## What the live runs found

- **Every `Auth0CustomDomain` deploy crashed on Pulumi.** The program chained a typed function onto the untyped output `pulumi.All` returns, which panics when the program registers its outputs. Each record output is now derived from the provider's values directly. A new module test runs the whole program under Pulumi's mocks, and that test fails on the old code with the same panic.
- **Every `Auth0ResourceServer` without a signing secret failed on OpenTofu.** The module's `variables.tf` marked `signing_secret` required. The root cause was in the variables generator, which read a length rule on an `optional` field as required presence. A field with explicit presence is now always optional in the generated `variables.tf`, whatever rules constrain its value.
- **An `Auth0ResourceServer` with no token dialect failed on OpenTofu.** The Terraform module sent an empty `token_dialect` (and `signing_alg`), which the provider refuses. It now sends null, exactly where the Pulumi module sends nothing.
- **Adopting an `Auth0EmailProvider` always planned a change,** because Auth0 never returns a provider's credentials. The import catalog now declares them config-only.
- **The harness refused a branding without a theme.** A verifier can now mark its id output optional.
- **The harness copied a local lock file that no release ships,** so a stale one made OpenTofu refuse to initialize. The lane's copy now skips it.

## What the live runs measured

- **A custom domain needs a verified card on the tenant.** Auth0 refuses one (403) on a tenant without a verified credit card on file, even on a trial. Auth0's documentation says the card is for verification and is not charged.
- **Screen partials need a custom domain.** Auth0 refuses them (403) on a tenant without one; the documentation had said they are stored and never shown. The field comment and GUIDE now say what Auth0 does.
- **Two adoption facts, measured by blind import:**
  - An imported `Auth0ClientFromMetadataDocument` plans to clear the description its document seeded unless the spec declares it; a fresh registration keeps it.
  - An imported `Auth0TenantSettings` plans to remove a `default_redirection_uri` the spec leaves unset.

  Both are the adoption rules the field comments state. Each such scenario opts out of the round-trip with its reason, and a sibling scenario that declares the value proves the recipe.

## Profiles

- **Green on both engines:**
  - newly: Auth0Branding, Auth0Prompt, Auth0PromptCustomText, Auth0EmailProvider, Auth0EmailTemplate, Auth0ClientFromMetadataDocument, Auth0ResourceServer;
  - re-proven on provider 1.58.0: Auth0Action, Auth0Client, Auth0Connection, Auth0EventStream, Auth0Role, Auth0User.
- **Waiting on the card:** Auth0CustomDomain, Auth0CustomDomainVerification, Auth0PromptScreenPartials, and Auth0TenantSettings' default-domain scenario.

## The rules learned

- The Pulumi module rule names the `pulumi.All` trap, and asks for a module test that runs the program under mocks.
- The Terraform module rule: a field with explicit presence is optional in `variables.tf`, and the Terraform module sends exactly what the Pulumi module sends.
- The forge rule:
  - write-only credentials are config-only for import;
  - a scenario that measures an adoption rule opts out of the round-trip with its reason;
  - an object with no identity of its own belongs to the kind it serves;
  - a local lane never inherits the shell's provider credentials.
- The update rule: extending a kind that manages live objects keeps unset unmanaged, and a provider argument with no computed value is found in the schema, measured by blind import, and named where adopters read.

## Verification

- The live lanes on both engines, one at a time, on the test tenant, and the blind-import round-trip on the Terraform engine.
- `go test` for the generator, the harness, the Auth0 catalog, both DNS zones, `pkg/outputs` and `pkg/iac/importmap`.
- `make generate-reference`.

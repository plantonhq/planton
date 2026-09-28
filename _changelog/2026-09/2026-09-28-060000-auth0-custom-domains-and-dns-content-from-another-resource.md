# Auth0 Custom Domains, and DNS Content From Another Resource

**Date**: September 28, 2026
**Type**: Feature
**Components**: API Definitions, Provider Framework, IAC Stack Runner, Testing Framework, Parity Accounting

## Summary

An Auth0 tenant can now sign people in on a domain its owner controls, declared in one install:
- `Auth0CustomDomain` creates the domain and reports the DNS record that proves control of it;
- a DNS record kind publishes that record, reading the domain's output;
- `Auth0CustomDomainVerification`, ordered after the record, waits until Auth0 reports the domain ready;
- `Auth0TenantSettings.default_custom_domain` then makes the verified domain the one the tenant's emails link to.

To make that composition possible, a Cloudflare DNS record's `content` (standalone and inline in a zone) takes a literal or another resource's output. A reference can also name one element of a list output, such as a zone's `status.outputs.nameservers.0`. The E2E runner learns to deploy a prerequisite of another provider with that provider's own module and harness.

## Motivation

A stranger's first screen is the sign-in page. Without a custom domain it sits on the tenant's `auth0.com` address, and Universal Login's page template cannot be set at all, since Auth0 requires a custom domain first. The chain also could not be written from files: the verification record is an output of the domain, and a DNS record's content was a plain string, so it took two installs and a hand copy. That is the same gap a subdomain delegation had with a child zone's name servers.

## What Changed

- **`catalog/auth0/auth0customdomain/` (8008, `a0cd`).** The complete component, built for 100% parity with `auth0_custom_domain`:
  - `domain`, `type`, `custom_client_ip_header`, `tls_policy` (`recommended` only, since Auth0 retired `compatible`, and valid only with Auth0-managed certificates), `domain_metadata` (ten pairs, 255-character values) and `relying_party_identifier`;
  - outputs `id`, `domain`, `status`, `origin_domain_name` and `dns_record_name` / `dns_record_type` / `dns_record_value`, named as `GcpCertManagerDnsAuthorization` names its validation record;
  - both modules pick the record with one rule (prefer the CNAME method, otherwise the first; a TXT method's `domain` is its name, and a CNAME's name is the domain itself), pinned by a Pulumi unit test and the outputs conformance case;
  - three presets, the permissions, cost and controls sidecars, an import map, a logo, a pending-proof E2E profile and a run-scoped scenario.
- **`catalog/auth0/auth0customdomainverification/` (8009, `a0cdv`, prerequisite `Auth0CustomDomain`).** Built for 100% parity with `auth0_custom_domain_verification`:
  - `custom_domain_id` is a reference to the domain;
  - outputs are the verified `domain` (read back by id on both engines), `origin_domain_name` and `cname_api_key` (sensitive; `pulumi.ToSecret`, a sensitive OpenTofu output);
  - its permission set is `create:custom_domains` (the Management API files verify under create) and `read:custom_domains`;
  - its E2E scenario composes the domain, a Cloudflare CNAME fixture in a real delegated zone, and the verification.
- **`Auth0TenantSettings.default_custom_domain`.** A reference to the verification's `domain` output, mapping `auth0_custom_domain_default`:
  - declared only when set;
  - counted by the at-least-one-setting rule;
  - its delete only forgets the default, since Auth0 cannot unset one;
  - it adds a permission group, a provider-parity mapping, a preset and a default-domain E2E scenario.
  - The kind's logo becomes its own glyph (it had worn `Auth0Role`'s), and its guide and cost profile now say that page templates are gated on a custom domain, not on a paid plan.
- **Cloudflare DNS content as value-or-reference.** `CloudflareDnsRecord.content` and `CloudflareDnsZone.records[].content` become `StringValueOrRef`:
  - the content/data rules read `has(this.content)`;
  - both Pulumi modules read `.GetValue()`, and the OpenTofu variable is unchanged, because the tfvars flattener writes the resolved string;
  - every manifest, preset and example writes `content: {value: ...}`;
  - a new `05-subdomain-delegation` preset reads a child zone's `nameservers.0`.
- **`pkg/refcheck`: a reference names one element of a list of plain values.**
  - The index ends the path, and the element must be a string.
  - A reference to the whole list, which used to pass offline and resolve to nothing at deploy, is refused with the fix in its message.
  - All 1,673 references in the repository's charts, presets and scenarios still resolve.
- **E2E runner: prerequisites of another provider.**
  - A prerequisite's provider is read from its kind, and its install profile and module come from that provider's catalog.
  - A suite lends the runner another provider's harness with `runner.RegisterDependencyHarness`; it is set up on first use and torn down with `runner.TeardownDependencyHarnesses`.
  - A prerequisite of an unregistered provider is refused, naming the call to add.
  - The Auth0 suite registers the Cloudflare harness, honors `planton.dev/e2e-required-env`, and gains entrypoints for the three kinds.
- **Auth0 E2E verifiers.**
  - The Management API client can read a resource's body.
  - The custom domain uses the path verifier.
  - The verification requires the domain `ready` after deploy and still `ready` after destroy (a no-op destroy).
  - Tenant settings, which have no identifier, are verified as the tenant's settings answering after deploy and destroy.
- **Auth0 provider at 1.58.0.**
  - The distilled schema is committed (`pkg/providerparity/schemas/auth0-1.58.0.json.gz`, 79 resources), and every Auth0 Terraform pin moves to `~> 1.58`.
  - `pulumi-auth0` moves from v3.35.0 to v3.54.0, the release that bridges 1.58.0; it requires Go 1.26.6, so `go.mod`, `go.work` and Bazel's Go SDK move to 1.26.6.
  - The dispositions ledger judges all 79 resources: none is deferred, and only deprecated or superseded ones are excluded.
- **The Management API scope snapshot** is refreshed with `tenant_settings` and `custom_domains`, so `TestAuth0ScopesExist` passes again.
- **Forge rule and E2E README.** Four lessons are recorded:
  - list-element references;
  - prerequisites of another provider, and the order of path entries;
  - measuring a new kind's parity before its provider enrolls;
  - confirming which tenant the ambient credentials reach before a live lane of a one-per-tenant kind.

## Design Notes

- **Verification is its own kind** because it cannot be folded into the domain: the record needs the domain's output, and verification needs the record, so one resource would loop. The verification orders itself after the record with `metadata.relationships` (`depends_on`), which the platform's dependency graph honors.
- **The default domain lives on tenant settings**, as `AwsIamAccountSettings` holds several per-account resources: it is a per-tenant setting, can be the canonical domain, and has no lifecycle of its own.
- **Auth0 is measured, not enrolled.** Enrolling the provider in the parity gate would add its eight older kinds to a baseline that stands at zero and never grows. New kinds are held to total accounting through `--kind`, and the provider enrolls when its older kinds catch up.

## Validation

- Spec tests for both new kinds, tenant settings and both Cloudflare DNS kinds (new tests red first).
- `pkg/refcheck` and the runner's dependency tests, each red first then green; the catalog-wide fixture integrity gate is green.
- `planton module verify`: all four new modules and both tenant-settings modules conform.
- `planton provider-parity --kind` reads total accounting for both new kinds, and no Auth0 resource is left undispositioned.
- Every new or changed manifest and preset validates.
- `make protos` passes, with its Java compile gate and its protovalidate-java rule gate. `make bazel-build-cli` passes, and all eight existing Auth0 kinds build on the new SDK.
- These gates pass: action inventory, permissions, cost, compliance, certification, secret coverage, logo, and import-map conformance.
- **Not run live:** both engines' lanes wait for a dedicated test tenant, because each kind's objects are one per tenant.

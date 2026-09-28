# A Rebuilt Database Archives Into Its Own Series, and a Tenant Declares Its Sign-In Page and Emails

**Date**: September 28, 2026
**Type**: Feature
**Components**: KubernetesPostgres (both engines, e2e), CloudNativePG typed bindings, six new Auth0 kinds, DNS kinds of AWS, Azure, GCP, Cloudflare and DigitalOcean, the Auth0 live-test verifiers, the forge rule, the disaster-recovery pattern

## Summary

**A database recreated from the same declaration keeps backing up.** `KubernetesPostgres` used to file every install's archive under the cluster's name. A recreate therefore wrote into its predecessor's history, which Barman refuses quietly: the database stayed healthy while `ContinuousArchiving` stayed false and WAL filled the data volume until the database stopped. Each install now archives into a backup **series** of its own, `<name>-<first 8 characters of the backup ObjectStore's UID>`, and every new series starts with a base backup. The series is the new `backup_server_name` output. The new `backup.server_name` names a series explicitly, for continuing a known one.

**A tenant's sign-in page and emails are declared from files.** Six Auth0 kinds, each at total parity accounting against provider 1.58.0:

- `Auth0Branding`: the logo, favicon, colors, font and page template, with the no-code theme folded in;
- `Auth0Prompt`: the login flow;
- `Auth0PromptCustomText`: the words of each prompt in each language;
- `Auth0PromptScreenPartials`: the HTML fragments on a prompt's screens;
- `Auth0EmailProvider`: the sending service;
- `Auth0EmailTemplate`: each email.

**Every DNS kind takes a reference wherever a value is a hostname or an address.** Any record field whose value another resource can produce is now a value or a reference.

## What Changed

- **KubernetesPostgres**
  - `backup.server_name` is rendered as the Cluster archiver plugin's `serverName` in both engines. The ObjectStore CRD forbids it on the store.
  - **Default series.** When `server_name` is unset, the series comes from the ObjectStore's UID: Pulumi reads `Metadata.Uid()`, and OpenTofu reads `kubectl_manifest.uid`. A destroy and recreate gets a fresh series. An import keeps the live ObjectStore, and so its series.
  - **Series-start backup.** An on-demand `Backup`, `<name>-series-start`, carries the series as the annotation `planton.ai/backup-series`. Any change replaces it: Pulumi `ReplaceOnChanges(*)` with `DeleteBeforeReplace`, OpenTofu `force_new`. So a new series is restorable from its first minute, and an unchanged one never re-runs it.
  - **Output and recovery.** The new `backup_server_name` output is what `bootstrap.recovery.source_server_name` names. A recovered cluster may now archive beside its source under the same path.
  - **CEL.** `server_name` must be a DNS label of at most 63 characters. A declared `server_name` equal to the recovery source's series under the same path is refused.
  - **Docs.** The field comments, a "Backups across rebuilds" section in the guide, the README, the catalog page, the presets and the import map are updated.
  - **E2E.** The in-cluster S3 fixture moves from MinIO, whose images no longer pull, to SeaweedFS (`fixture-s3.yaml`). The GKE recovery sources name their series.
- **CloudNativePG typed bindings** gain `Backup` (`gen-cloudnative-pg` includes the Backups CRD).
- **Auth0 kinds 8010 to 8015**, each through the full forge: protos with CEL, spec tests, both engines, presets, permissions, cost, controls, import maps, a provider-parity manifest, docs and logo.
  - `Auth0Branding` sends a declared theme whole, filling each unset field with Auth0's own default.
  - `Auth0PromptCustomText` renders its screens into the same sorted-key JSON bytes in both engines.
  - `Auth0EmailProvider` gives each sending service its own arm, and sends Resend over its SMTP interface: Auth0's API accepts `resend`, but provider 1.58.0 does not.
  - `Auth0EmailTemplate` states Auth0's refusal of a custom redirect on non-Enterprise tenants created on or after May 5, 2026.
  - **Around the kinds:**
    - one verifier per kind, each stating its destroy contract, with unit tests;
    - live-test entrypoints for both engines;
    - the import catalog;
    - dispositions (the single-screen partial is composed);
    - the Management API scope snapshot gains `branding`, `prompts`, `email_provider` and `email_templates`.
  - Their live profiles stay `pending_proof` until a dedicated test tenant exists.
- **DNS**, each field now a value or a reference:
  - `AwsRoute53DnsRecord.values`.
  - `AzureDnsRecord` A and AAAA addresses, NS, PTR, MX exchanges and SRV targets. Its IPv4 and IPv6 checks move to CEL on each literal.
  - `AzurePrivateDnsRecord`: the same fields.
  - `GcpDnsRecord` routed values (weighted and geolocation).
  - `CloudflareDnsRecord` and `CloudflareDnsZone` SRV, HTTPS, SVCB and URI targets.
  - `DigitalOceanDnsZone.ip_address`.
  - Pulumi reads the resolved value. OpenTofu already receives the plain string, so its modules do not change. Every manifest, preset and doc example is rewritten to the reference form.
- **Forge rule** gains two design rules: a value another resource produces takes a reference, and a per-install identity comes from an object the stack creates earlier in the same run. The disaster-recovery pattern states the per-install archive identity.

## Verification

- **Series lane (live, on kind).** `KubernetesPostgres` `with-backup` now reinstalls into the same bucket path. It is green on Pulumi and on OpenTofu.
  - Each install archived into its own series: `...-178dcaae`, then `...-32ca6294` on Pulumi; `...-f10c5159`, then `...-13724352` on OpenTofu.
  - Each install showed `ContinuousArchiving=True` and a completed base backup.
  - The red case is the recreate this lane replays, observed live: a rebuilt database archiving under its predecessor's name read "Expected empty archive" for four days, until its volume filled.
- **Spec tests.** Every touched kind's spec tests pass, with new cases for each rule. The Auth0 module unit tests and verifier tests pass.
- **Modules.** `planton module verify` confirms conformance for both engines of the six Auth0 kinds, the seven DNS kinds and `KubernetesPostgres`.
- **Parity.** `planton provider-parity --provider auth0 --ga-schema auth0 --kind <Kind>` reads total accounting, with zero unaccounted, for all six kinds.
- **Gates.** These pass:
  - `go test` for `pkg/crkreflect`, `pkg/refcheck`, `pkg/outputs`, `pkg/iac/actioninventory`, `pkg/providerparity`, `pkg/secretcoverage`, `pkg/cataloglogo`, `pkg/presetvalidity`, `pkg/certification`, `pkg/iac/importmap`, `pkg/iac/permissions`, `pkg/finops`, `pkg/compliance`;
  - `TestCatalogFixtureIntegrity`;
  - Bazel builds of every touched package.
- **Anatomy.** The anatomy gate's only finding is an existing one in `kubernetesplantonplatform`, outside this change.

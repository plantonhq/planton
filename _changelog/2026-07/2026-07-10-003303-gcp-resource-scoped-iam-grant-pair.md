# GCP Resource-Scoped IAM Grant Pair: GcpServiceAccountIamMember + GcpKmsKeyIamMember

**Date**: 2026-07-10
**Type**: Feature
**Scope**: `gcp-serviceaccount-iam`, `gcp-kms-iam`
**Impact**: Two new GCP resource kinds closing the two resource-scoped IAM gaps in the catalog — service-account-scoped grants (the keyless-impersonation hop) and key-scoped CMEK grants (least privilege plus a real ordering edge for encrypted resources)

## Summary

The GCP catalog modeled IAM grants only at project scope (`GcpProjectIamMember`). Two real composition needs sit below the project:

1. **Who may use a service account?** Granting `roles/iam.workloadIdentityUser` (federated impersonation — the terminal hop of keyless CI/CD), `roles/iam.serviceAccountTokenCreator` (short-lived token minting), or `roles/iam.serviceAccountUser` (actAs for Cloud Run/GCE/Functions deploys) belongs ON the account. A project-wide grant of those roles allows acting as EVERY account in the project.
2. **Who may use a crypto key?** Every CMEK consumer's service agent needs `roles/cloudkms.cryptoKeyEncrypterDecrypter` on the key it encrypts with. A project-wide grant over-privileges, and it carries no dependency edge — so a first CMEK deploy can race IAM propagation.

This change forges both kinds as full kinds with additive-only semantics (`*_iam_binding` / `*_iam_policy` authoritative modes remain deliberately unmodeled — clobber semantics are hostile to composition).

## What Was Built

### GcpServiceAccountIamMember (enum 703, `gcpsaim`)

- **Spec**: `service_account_id` (required; FK → `GcpServiceAccount.status.outputs.name`, the fully-qualified `projects/.../serviceAccounts/...` resource name — no separate project field, the account name embeds it), `role` (required; FK → `GcpIamCustomRole.status.outputs.name`), `member` (required; FK → `GcpServiceAccount.status.outputs.member`), optional IAM `condition`. Everything immutable, mirroring the API (grants have no update).
- **Outputs**: the resolved grant tuple + the policy `etag`.
- **Docs teach the boundary**: this kind is the general primitive for any principal (GitHub principalSet, cross-SA impersonation, users, groups); `GcpGkeWorkloadIdentityBinding` stays as the GKE convenience that derives the workload-identity principal from cluster coordinates.
- **Presets**: GitHub workload-identity impersonation (principalSet member assembled from project number + pool), cross-account token creator, deployer actAs.

### GcpKmsKeyIamMember (enum 692, `gcpkim`)

- **Spec**: `crypto_key_id` (required; FK → `GcpKmsKey.status.outputs.key_id` — byte-exact one of the three identifier formats the provider accepts), `role` (required), `member` (required), optional IAM `condition`. Everything immutable.
- **Outputs**: the resolved grant tuple + the policy `etag`; `crypto_key_id` echoes the configured identifier on both engines (the provider normalizes only on import).
- **Docs teach the ordering win**: referencing the key makes the grant a DAG edge, so encrypted resources deploy strictly after the permission exists — closing the first-CMEK-deploy IAM-propagation race. Ring-level inheritance and the lazily-created service-agent sharp edge are documented.
- **Presets**: Cloud Storage service-agent CMEK grant, workload key user (both sides referenced), conditional time-bound key access.

### Both kinds

- Both engines (Terraform on `google ~> 6.0`, Pulumi on pulumi-gcp v9) at 100% behavioral parity: identical deploy-time validation (member format with `deleted:` rejection; identifier shape checks), identical condition handling, identical outputs — verified by the `pkg/outputs` conformance cases, `planton validate-outputs`, and live E2E output checks (4/4 proto fields populated per engine).
- Design substance verified against the released provider tag (v6.50.0), including the shared IAM member base schema; the KMS IAM `condition` block is GA in the released source (the resource doc page's Beta annotation is stale) and is proven live.
- Kind registry entries with prerequisites (`[GcpServiceAccount]`; `[GcpKmsKeyRing, GcpKmsKey, GcpServiceAccount]`), catalogkindreflect kind-map regeneration, spec tests (14 cases each), 3 presets each, hack manifests, full doc sets, site catalog pages.

## E2E

- Two new verifiers probe the target resource's IAM policy at version 3 (conditioned bindings listed distinctly) and assert the exact (role, member) pair present after deploy and absent after destroy.
- `GcpKmsKey` now publishes an E2E prerequisite profile (a key composed on the ring prerequisite with a run-scoped name), enabling any future CMEK consumer chain.
- **SA-IAM scenario** (`token-creator-self`): one GcpServiceAccount prerequisite feeds BOTH the grant target (its `name` output) and the grantee (its `member` output) — both FK paths proven with one prerequisite. Live dual-engine green (~1m10s/engine).
- **KMS-IAM scenario** (`conditional-cmek-grant`): ring → key → service account chain, grant carries an IAM condition — the conditioned-binding arm proven live. Live dual-engine green (~1m55s/engine incl. the chain).
- Zero orphans (the run-scoped ring/key prerequisites remain as inert, free objects per the undeletable-resource contract).

## Validation

- Spec tests green (14 + 14 cases); per-kind `go build` incl. release-equivalent Pulumi binary builds; `tofu validate` + real-converter `planton tofu plan` green on both hack manifests.
- `validate-refs --check` green (all FK annotations resolve); `secret-coverage --check` green (no secret-bearing fields; the GCP baseline stays empty); `pkg/outputs` conformance green; manifest validation green for every new YAML (run-scoped-token manifests validated with the token substituted).
- Live dual-engine E2E 4/4 green on the test project.
- Audits: both kinds Fully Complete, PARITY ✅ (`docs/audit/2026-07-10-002907.md` in each kind).

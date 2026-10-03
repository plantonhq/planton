# GCP Vertex AI: Notebook and Endpoint Rebuilt to the Released Floor with a Live-Proven Deterministic Endpoint ID

**Date**: July 8, 2026
**Type**: Feature
**Components**: GCP Provider, API Definitions, Testing Framework

## Summary

The Vertex AI pair — the catalog's ML serving and experimentation surface — now stands at the released `google ~> 6.x` floor. `GcpVertexAiNotebook` (670) gained user labels, Confidential Computing, reservation affinity, managed EUC, third-party identity, the external-IP arm, and declarative stop/start; `GcpVertexAiEndpoint` (671) gained user labels, BigQuery request/response logging, and IAM-authorized PSC — and its probable Pulumi deploy blocker (the empty-name path) was closed with a deterministic, cross-engine-identical endpoint ID derivation that the live run proved byte-identical in the strongest way possible: the two engines collided on the same derived ID. Both kinds carry the family's first E2E coverage, live-proven 6/6 on both engines.

## Problem Statement / Motivation

### Pain Points

- **A probable deploy blocker on the endpoint's empty-name path.** Vertex AI requires a numeric endpoint ID (1–10 digits, no leading zero) and never generates one. Terraform satisfied it with `random_integer` (non-deterministic across re-applies from scratch); the Pulumi module omitted `Name` entirely and leaned on bridge auto-naming — which synthesizes a non-numeric name the API rejects. The family had zero E2E, consistent with a defect that had never been caught.
- **Floor gaps on the notebook** vs the released `google_workbench_instance`: no user labels (the recurring TF zero-labels live parity class), no `confidential_instance_config`, `reservation_affinity`, `enable_managed_euc`, `enable_third_party_identity`, external-IP `access_configs`, or `desired_state`; outputs missing `health_state`/`update_time`.
- **Floor gaps on the endpoint**: no user labels, no `predict_request_response_logging_config`, and the PSC block omitted `enable_secure_private_service_connect`.
- **Systemic conformance classes on both kinds**: no API enablement, required `project_id` instead of ambient, stale `object({ value = string })` ref typing, stale `Pulumi.yaml binary:`, no registry prerequisites, unpinned bridged-provider flags.
- **Zero validation surface**: no verifiers, scenarios, test entrypoints, or `pkg/outputs` cases for either kind.
- **Stale docs**: internal-plan breadcrumb sections; image-family guidance pointing at deep-learning-VM families GCP has since retired.

## Solution / What's New

### The deterministic endpoint ID (the session's core fix)

When `spec.endpoint_name` is empty, both engines derive the numeric ID identically: sha256 of `"{org}/{env}/{name}"` → first 12 hex chars (48 bits) → mapped into `[1000000000, 9999999999]`. Implemented in `endpoint_name.go` (Pulumi) and `locals.tf` (Terraform), locked by a pinned-value unit test (three independently computed values), exact in HCL (48-bit values are precise in cty numbers), and `Name` is now always sent on both engines. The TF module's `random` provider dependency is gone.

**The live run turned a failure into the proof.** Pulumi created endpoint `1853927074` for the E2E identity; Terraform, deriving the ID independently minutes later, collided with 409 ALREADY_EXISTS on GCP's deleted-ID reservation — the collision itself demonstrated byte-identical cross-engine derivation against the real API. It also settled the pre-flight reservation question (GCP reserves deleted endpoint IDs: GET 404s while create 409s), now documented on the spec field, in `docs/README.md`, and encoded as a learn-once rule: identity-derived cloud IDs need run-scoped `metadata.name` in E2E scenarios (a recorded exception to the fixed-name rule, legal only for leaf kinds).

A second live-discovered fix rode along: the provider resolves the Vertex AI regional API host from `region`, never `location` — both modules now pin `region = location` (matching comments), keeping a single honest field in the spec.

### `GcpVertexAiNotebook` (670) — depth rebuilt

- **New spec surfaces at the released floor** (each verified GA on the released line AND expressible in the bridged Pulumi SDK before modeling): user `labels`, `confidential_instance_config` (AMD SEV), `reservation_affinity` (all three consume types with a key/values coherence CEL), `enable_managed_euc`, `enable_third_party_identity`, external-IP `access_configs`, `desired_state` (declarative stop/start). Outputs extend-only: `health_state`, `update_time`. 56-case spec test.
- The singular `boot_disk`/`data_disk`/`accelerator_config`/`network_interface`/`service_account` sub-messages verified provider-authentic (the released schema caps each at one) — no restructuring, inbound refs stable.
- **Shielded-VM false-value semantics unified**: the API enables vTPM and integrity monitoring by default, so an explicit `false` actively disables them; both engines now omit false flags, keeping server defaults intact for unset flags.
- Registry `prerequisites: [GcpVpcNetwork, GcpSubnetwork, GcpServiceAccount]` — exercised live end-to-end, not just declared.
- **Image guidance corrected against live reality**: GCP retired the deep-learning-VM notebook families (`common-cpu-notebooks` et al. no longer resolve). Spec comments, presets, hack manifest, catalog page, and the E2E scenario now prefer the service's default Workbench image (omit `vmImage`) or pin `cloud-notebooks-managed/workbench-instances`.

### `GcpVertexAiEndpoint` (671) — conformed and deepened

- **New spec surfaces**: user `labels`, `request_response_logging_config` (BigQuery destination as an honest plain string — the `bq://` scheme has no matching output to reference), `enable_secure_private_service_connect`. Two message-level CELs move the provider's `ConflictsWith` rejections pre-deploy (network⇔PSC, dedicated⇔PSC); the `endpoint_name` CEL encodes the exact numeric contract. Output extend-only: `endpoint_name` (the value model-deployment tooling consumes). 37-case spec test.
- `enable_secure_private_service_connect` verified **GA on the released 6.x line** via the installed provider schema — with a recorded PARITY/UPGRADE NOTE that the v7 major drops it from GA (the catalog-wide bump decision must drop the field or move to google-beta).
- Recorded skips with reasons: `traffic_split` (keyed by deployed-model IDs that only exist after out-of-band model deployment — modeling it invites perma-diff), `region` (module plumbing, not a spec field).

### Module defect closure (both engines, both kinds)

- API enablement (`aiplatform.googleapis.com`; `notebooks` + `compute` for the notebook) with `disable_on_destroy=false` and dependency gating; ambient `project_id`; identical user-beneath-platform labels merge; flat-string ref typing; canonical `Pulumi.yaml`; bridged client-side `deletion_policy` pinned to `DELETE` with PARITY comments on both kinds.

### E2E: first coverage for the family

- Two REST-probe verifiers (no typed Vertex/Workbench client in the pinned `google.golang.org/api` line — the established `Services.RestClient` pattern): both assert the `planton-ai_resource` label live (permanent guard on the labels class); the endpoint verifier additionally asserts the `endpoint_name` output matches the live numeric ID (the determinism contract) and the dedicated-DNS output matches live; the notebook verifier asserts operational state.
- Endpoint: minimal + dedicated-endpoint leaf scenarios (run-scoped identities per the reservation finding). Notebook: minimal on the full VPC → subnetwork → service-account chain via consumer-scoped prerequisite overrides. `pkg/outputs` conformance cases ×2.

## Validation

- Offline (all green): spec tests 56 + 37 + the pinned derivation test; per-kind Go builds + release-equivalent Pulumi builds; `tofu init/validate` ×2; offline `planton tofu plan` through the real converter ×2 (labels, PSC, logging, shielded, derived name inspected); `secret-coverage --check`; `validate-refs --check`; `validate-outputs` on both module dirs ×2; two new `pkg/outputs` conformance cases; every preset/hack/scenario/prerequisite manifest through a freshly built `planton validate`; gazelle + `make build-go`.
- **Live dual-engine E2E green on `planton-e2e` (6/6)**: endpoint minimal + dedicated 4/4 (~65s/engine; label, ID-determinism, and dedicated-DNS posture verified live), notebook composed chain 2/2 (Pulumi ~7 min, Terraform ~10 min — inside the live-run threshold). Post-run sweep: zero endpoints, instances, networks, or service accounts remaining.
- Both kinds audited **Fully Complete — PARITY ✅** (`docs/audit/2026-07-08-121500.md` per kind); site catalog regenerated (notebook preset pages created; endpoint pages refreshed).

## Deliberately Not Modeled (recorded reasons)

- Endpoint: `traffic_split`, spec-level `region`, `psc_automation_configs` (not in the released line), `model_deployment_monitoring_job` output (always empty from IaC).
- Notebook: `instance_id` (vestigial — the create call derives the instance ID from `name`; a second identity field invites drift).
- The vector-search trio (`index`/`index_endpoint`/`deployed_index`), Model Garden deployment, and feature-store generation — recorded as the strongest candidates for a follow-up AI-family session.
- Charts (`charts/gcp/ml-notebook-environment`) stay frozen; may be stale against the new spec (accepted).

## Impact

The catalog's ML pair is floor-complete, parity-proven, and live-verified for the first time — and the endpoint's deterministic ID contract means the same manifest now yields the same endpoint on either engine, the property every future model-deployment kind will build on. Three learn-once rules landed in `e2e/README.md`: identity-derived IDs need run-scoped metadata names, offline manifest validation belongs in the gate (it catches dead-on-arrival scenarios in seconds), and named image families are a staleness trap.

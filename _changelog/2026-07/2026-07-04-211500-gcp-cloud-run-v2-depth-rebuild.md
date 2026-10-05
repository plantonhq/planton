# GCP Cloud Run v2 Deep Rebuild

**Date**: July 4, 2026
**Type**: Enhancement
**Components**: API Definitions, GCP Provider, IaC Modules, E2E Framework, Presets, Documentation

## Summary

Deep-rebuilt the `GcpCloudRun` kind from a flat 38-field single-container
abstraction to the released google-provider v2 floor: provider-authentic
multi-container shape with sidecars, startup ordering, startup and liveness probes,
secret/GCS/NFS/in-memory/Cloud SQL volumes, revision traffic splitting,
direct VPC egress, session affinity, CMEK, binary authorization, and GPU
inference. Removed the bundled custom-domain black box (hidden DomainMapping +
RecordSet resources). Fixed a verified Terraform parity break where the entire
Cloud SQL surface was silently dropped. Closed the standing secret-coverage gap
for Secret Manager env references. Proven live on both engines with a minimal
public-service scenario, zero orphans.

## What Changed

### Spec redesign (`spec.proto`, contiguous renumbering)

- **Multi-container**: `containers[]` (min 1 by CEL) replaces the single
  `container`; each container carries image, command/args, env (literal XOR
  Secret Manager `secret_key_ref`), ports, provider-authentic cpu/memory limits,
  startup and liveness probes (readiness_probe is main-branch-only on the
  provider and recorded as a skip), volume mounts, working directory,
  `depends_on` startup ordering, and name.
- **Volumes**: `volumes[]` with exactly-one-source CEL per volume — cloud_sql
  (repeated refs → `GcpCloudSql` connection names), secret, empty_dir, gcs
  (ref → `GcpGcsBucket`), nfs.
- **Networking**: `vpc_access` with connector ref XOR direct
  `network_interfaces` (refs → `GcpVpcNetwork`/`GcpSubnetwork`) enforced by CEL;
  egress enum.
- **Service level**: scaling (service + per-revision), `traffic[]` (percent/revision/tag/type with sum-to-100 CEL), ingress, `allow_unauthenticated`, custom audiences, session affinity, execution environment, launch stage, `encryption_key` (ref → `GcpKmsKey`), binary authorization, GPU (`node_selector.accelerator` + zonal redundancy), timeout, max concurrency, labels/annotations, description.
- **Safety defaults**: `deletion_protection` defaults TRUE (provider default; replaces the old recommended-false `delete_protection`).
- **`service_account`**: plain string → `StringValueOrRef` → `GcpServiceAccount` (`status.outputs.email`).
- **Dns block removed**: the spec's `dns.enabled` toggle and both engines' hidden
  `google_cloud_run_domain_mapping` + DNS record set resources deleted — silent
  data loss on repeated hostnames fixed by removal; custom domains taught via
  the composed LB family + `GcpDnsRecord`.
- 12 CEL coherence rules; 81-case spec test (from ~15 cases).

### Both engines, one contract

- **Cloud SQL parity break fixed**: Terraform previously implemented none of the
  spec's `cloud_sql` block while Pulumi wired it — both engines now use native
  Cloud SQL volume mounts.
- **Pulumi invoker flags fixed**: `allow_unauthenticated` and
  `invoker_iam_disabled` were conflated; now separate.
- Converter-contract `variables.tf` (plain-string refs), `~> 6.0` float (was
  hardcoded `6.19.0`), `run.googleapis.com` API enablement with
  `disable_on_destroy=false`, ambient-project fallback, canonical
  `iac/hack/manifest.yaml` (wrong-kind stray manifests deleted).
- Outputs extended (+`uri`, `urls`, `latest_ready_revision`, `location`,
  `uid`); `pkg/outputs` conformance case added.
- Secret coverage: Secret Manager secret name fields annotated with
  `sensitive_exempt_reason` — gap removed from baseline.

### Registry, presets, docs

- Three rewritten presets: public API service (minimal), private VPC-connected
  backend (Cloud SQL volume + secret env), GPU inference (L4 sidecar pattern).
- README, catalog page, and docs rewritten with the 90/10 table and the
  recorded-skips ledger (verified against released 6.50.0 schema dump).
- Site catalog copies refreshed.

### E2E

- New `verify/cloud_run.go` (Run Admin API v2 — service READY + URL + revision +
  traffic posture assertions).
- `minimal` scenario (public hello-service, both engines).
- Direct VPC egress scenario authored but excluded from live matrix: GCP holds
  serverless address reservations 1–2 hours post-delete (documented in
  `e2e/README.md`).
- **Live dual-engine proof on `planton-e2e`: 2/2 scenario-runs green** — Pulumi
  68s, Terraform 72s. Zero orphans.

### Workflow uplift

- `e2e/README.md`: hours-scale async-release exclusions for cloud-side holds
  that outlive ephemeral teardown (serverless-ipv4-* subnetwork reservations).
- `_rules/catalog-kind/forge/flow/013-terraform-module.mdc`: offline
  `planton tofu plan` must use absolute paths for both manifest and `--module-dir`
  (relative paths silently fall back to a stale staging clone).

## Validation

- Offline: spec tests 81/81; Pulumi build; `tofu validate`; `secret-coverage --check` (GcpCloudRun gap closed); `validate-refs --check`; `validate-outputs` 6/6 both engines; `pkg/outputs` conformance; preset/hack/scenario manifests through `planton validate`.
- Live: 2/2 dual-engine scenario-runs green, zero orphans.
- Audit: Fully Complete — PARITY ✅ (`docs/audit/2026-07-04-205001.md`).

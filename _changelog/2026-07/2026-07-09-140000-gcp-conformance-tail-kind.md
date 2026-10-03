# GCP Conformance Tail: Five Kinds Brought to the Released Floor + Certificate Manager DNS Authorization Forged

**Date**: July 9, 2026
**Type**: Feature
**Components**: GCP Provider, API Definitions, E2E Framework, Site Catalog

## Summary

Kind #29 closes the catalog's last untouched GCP kinds — `GcpDnsRecord`, `GcpGkeWorkloadIdentityBinding`, `GcpCloudArmorPolicy`, `GcpCertManagerCert`, and `GcpProject` — plus the newly forged `GcpCertManagerDnsAuthorization` (716). All six stand at the released `google ~> 6.x` floor with PARITY ✅, secret-coverage green, outputs conformance, and live dual-engine E2E on `planton-e2e` (GcpProject plan-only by design). The Certificate Manager family is honestly decomposed: DNS authorizations and validation records are first-class composable nodes instead of a bundled black box inside the cert module.

## Problem Statement / Motivation

- **Five kinds never received the 90/10 rebuild** — stale pins, parity breaks, missing provider surfaces, wrong docs, and zero E2E.
- **Certificate Manager was dishonestly bundled** — the cert module silently created DNS authorizations and records, overlapping `GcpDnsRecord` and hiding `PER_PROJECT_RECORD` authorizations.
- **GcpCertManagerCert duplicated `GcpManagedSslCertificate`** via a LOAD_BALANCER arm creating `google_compute_managed_ssl_certificate`.
- **GcpProject carried Layer-0 defects** — bundled owner IAM grant, nondeterministic `add_suffix` per engine, bool `delete_protection` hiding ABANDON, label-key parity breaks.
- **Live-blocking defects** — WIB passed bare SA email instead of the required service-account link; Cloud Armor TF dropped labels and request-body inspection size.

## Solution / What's New

### GcpDnsRecord (618)
- Full `routing_policy` depth (WRR / geo / primary-backup + health-checked targets) with values-XOR-routing CEL mirroring the provider's ExactlyOneOf.
- `name` and `values` remodeled as `StringValueOrRef` so validation records compose from `GcpCertManagerDnsAuthorization` outputs via E2E/production FK resolution.
- Ambient `project_id`; free-string `type`; registry prerequisite `GcpDnsZone`.

### GcpGkeWorkloadIdentityBinding (615)
- Both engines construct the full `projects/{project}/serviceAccounts/{email}` link (fixes live deploy failure).
- IAM `condition` (619 shape); RFC-1123 validation on KSA namespace/name; honest docs (no false dual-write claim).

### GcpCloudArmorPolicy (622)
- TF parity breaks closed (labels, `request_body_inspection_size`).
- GA depth: reCAPTCHA options, multi-key `enforce_on_key_configs`, `json_custom_config`, adaptive-protection `threshold_configs`.
- Default-rule contract encoded as CEL (priority 2147483647 when rules non-empty) — no module-side injection.

### GcpCertManagerDnsAuthorization (716, new)
- First-class DNS authorization with validation-record outputs (`dns_record_name/type/data`) for `GcpDnsRecord` composition.
- Enum 716 in the 710–719 block; full anatomy (protos, modules, presets, E2E, audit).

### GcpCertManagerCert (616)
- Pure Certificate Manager kind: managed XOR self_managed (PEM sensitive), `location`/`scope`, FK `dns_authorizations` → auth kind.
- LOAD_BALANCER arm removed; real outputs (`managed_state`, `san_dnsnames`, etc.); honest E2E posture (PROVISIONING, never ACTIVE).

### GcpProject (609)
- Owner grant removed; `add_suffix` removed; `deletion_policy` (DELETE/PREVENT/ABANDON); `tags`; `auto_create_network`; `display_name`.
- Label-key and output parity closed; plan-only E2E (org-level Project Creator required for live create/delete).

## Validation

- Spec tests: all six kinds green.
- `pkg/refcheck`, `pkg/secretcoverage` gate, `pkg/outputs` conformance ×6: green.
- Targeted Pulumi builds ×6: green.
- Live dual-engine E2E on `planton-e2e`: DnsRecord, WIB, CloudArmor, DnsAuthorization, CertManagerCert — all green both engines; GcpProject plan-only.
- Audits ×6: Fully Complete PARITY ✅; site catalog regenerated (`npm run copy-docs` + `generate-structure`).

## E2E Learnings Encoded

- Certificate Manager rejects unrecognized TLDs (`.local` fails); E2E domains use `example.com`.
- Cloud Armor spec has no `labels` field — E2E scenarios must not set it.
- Composed cert chain: zone → auth → record (valueFrom on name/values) → cert; verifier asserts PROVISIONING only.

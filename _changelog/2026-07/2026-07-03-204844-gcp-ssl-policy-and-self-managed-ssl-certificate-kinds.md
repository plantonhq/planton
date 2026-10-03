# GCP SSL Policy and Self-Managed SSL Certificate Kinds

**Date**: July 3, 2026
**Type**: Feature
**Components**: GCP Provider, API Definitions, Pulumi Modules, Terraform Modules, E2E Framework, Documentation

## Summary

Adds the two TLS-side leaves of the GCP L7 load-balancing family: `GcpSslPolicy` (711) — TLS version and cipher-suite hardening for load balancer frontends — and `GcpSslCertificate` (712) — self-managed certificates where the user brings the PEM chain and private key. The target HTTPS proxy's `ssl_policy` reference gains its default kind, closing the family's last un-defaulted TLS reference. Both kinds fold global and regional scopes into one kind (the schemas are field-for-field identical apart from `region`), so the future regional load-balancer wave needs no new TLS kinds.

## Problem Statement / Motivation

The HTTPS frontend kinds shipped with two open TLS ends:

- `GcpTargetHttpsProxy.ssl_policy` was an un-defaulted reference — without an SSL policy kind, users could not harden TLS versions/ciphers through composition, and GCP's permissive default (TLS 1.0, COMPATIBLE ciphers) applied silently.
- Only Google-managed certificates existed as a kind. Wildcard domains, private/corporate CA issuance, internal load balancers (no public DNS for managed validation), and pre-DNS-cutover TLS all require self-managed certificates.

## Solution / What's New

### GcpSslPolicy (enum 711, `gcpsslpol`)

- Full released 6.x GA surface: `profile` (COMPATIBLE/MODERN/RESTRICTED/CUSTOM), `min_tls_version` (TLS_1_0/1_1/1_2), `custom_features` cipher allowlist, dual-scope `region`.
- Message-level CEL enforces the provider's CustomizeDiff pairing both ways pre-deploy: CUSTOM requires `custom_features`; `custom_features` requires CUSTOM.
- Outputs include `enabled_features` — the cipher list GCP actually computed, the artifact a compliance auditor asks for.
- Mutability documented as the feature it is: profile/version/ciphers update in place and apply fleet-wide to every referencing proxy on the next handshake.
- Deliberately unmodeled (recorded in `docs/README.md`): `post_quantum_key_exchange`, `FIPS_202205`, `TLS_1_3` minimum — all unreleased-main only, absent from the released schema on both GA and beta.

### GcpSslCertificate (enum 712, `gcpsslcert`)

- Self-managed upload: PEM `certificate` chain + PEM `private_key`, dual-scope `region`, shared name namespace with managed certificates documented.
- `private_key` is `(sensitive)`: encrypted in Pulumi state via `ToSecret`, write-only in GCP, never in outputs. The `certificate` chain is deliberately NOT sensitive — it is public handshake material presented to every client (the TF provider's sensitive flag on it is a plan-display concern, not a secrecy contract).
- PEM-framing CEL catches swap mistakes pre-deploy (private key pasted into the certificate field and vice versa).
- Full immutability + the create-before-destroy rotation sequence documented in spec comments, README, docs, and a dedicated rotation preset.
- `expire_time` output is the rotation clock — nothing renews itself.

### One kind, two scopes (both kinds)

Empty `region` creates the global resource; a set region creates the regional one — `google_compute_ssl_policy`/`google_compute_region_ssl_policy` and `google_compute_ssl_certificate`/`google_compute_region_ssl_certificate` expose identical surfaces (verified against the released schema on GA and beta). Both engines branch with the count-guard pattern.

### Target HTTPS proxy follow-ups

- `ssl_policy` gains `default_kind = GcpSslPolicy` (+ `self_link` field path) — referencing a policy is now `valueFrom: {name: ...}`.
- The `ssl_certificates` comment documents that self-managed certificates attach via an explicit `valueFrom.kind: GcpSslCertificate` (the list's default kind stays the managed certificate).
- Registry prerequisites gain both new kinds, serving the new composed E2E scenario.

### E2E

- Verifiers for both kinds (existence + posture assertions: non-empty enabled-features on the policy; non-empty expire-time on the certificate — proof GCP parsed real chain material).
- Scenarios: SSL policy global + regional (both module branches), certificate leaf (uploads a THROWAWAY self-signed keypair checked into the manifest — CN=e2e-test.invalid, trusted by nothing), and a new `hardened-frontend` HTTPS-proxy scenario composing routing + a self-managed certificate + an SSL policy live — the full production TLS posture on one proxy, with the policy resolved through the field's new default kind.

## Validation

- Offline: `make protos` ×2 + kind-map regen; spec tests 21 + 20 green; release-equivalent Pulumi builds; `tofu validate` + offline `planton tofu plan` through the real tfvars converter (plans inspected; both PEM fields render sensitive in the certificate plan); `secret-coverage --check`; `validate-refs --check`; `validate-outputs` dry-runs fully populated on BOTH module dirs for both kinds; two new `pkg/outputs` conformance cases (incl. repeated `enabled_features`); every preset/hack/scenario/prerequisite manifest through `planton validate`; `make build-go`; `make reset-gazelle`; framework tests (outputs/refcheck/catalogkindreflect/runner).
- Live (project `planton-e2e`, dual-engine, create → verify → destroy): SSL policy 113s/87s (global + regional scenarios), certificate 35s/50s, and the hardened-frontend + minimal HTTPS-proxy chains 13m11s/12m09s (7 transitive prerequisites each). Zero orphans after per-type sweeps across all LB resource types.
- Audits: both kinds Fully Complete with cross-engine PARITY, zero parity exceptions (`docs/audit/2026-07-03-183500.md` in each kind).

## Related Cleanup

- Retired-kind link sweep: dead links to the removed Secrets Manager and Cloud CDN catalog pages dropped from the GCP catalog index, cloud-function, and project site pages; the GCS-bucket static-website preset now points at the composed `GcpBackendBucket` + HTTPS load balancer path instead of a retired kind's preset.
- The delete-kind rule now mandates an inbound-link sweep (site catalog + sibling presets/docs) so removing a kind can never strand dead links again.
- `site/scripts/copy-kind-docs.ts` title map gains SSL Policy / SSL Certificate entries.

## Impact

The global external HTTPS load balancer family is now complete end to end, including its TLS controls: VIP, TLS termination with both certificate ownership models, handshake hardening, routing, backends, and probes — all first-class, composable, dual-engine kinds. Regional TLS resources come free with the folded scopes when the regional-LB wave lands.

---

**Status**: ✅ Production Ready

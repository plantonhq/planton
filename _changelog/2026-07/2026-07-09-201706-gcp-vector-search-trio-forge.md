# GCP Vector Search Trio: Index, Index Endpoint, and Deployed Index Forged

**Date**: July 9, 2026
**Type**: Feature
**Components**: GCP Provider, API Definitions, E2E Framework, Site Catalog

## Summary

Kind #30 forges the Vertex AI vector-search trio — `GcpVertexAiIndex` (672), `GcpVertexAiIndexEndpoint` (673), and `GcpVertexAiDeployedIndex` (674) — to the released `google ~> 6.50.0` floor with full dual-engine parity, REST-probe E2E, and audits. This completes the AI family and freezes the Phase-1 kind catalog for the chart wave (RoE §17).

## Problem Statement / Motivation

Vector Search (Matching Engine) is a three-resource GCP surface — storage (index), serving surface (index endpoint), and placement (deployed index) — that had no catalog representation. The online-prediction `GcpVertexAiEndpoint` (671) is a different GCP resource; conflating them would mislead users wiring RAG retrieval.

## Solution / What's New

### GcpVertexAiIndex (672, `gcpvaidx`)
- Full index geometry at the 6.50.0 floor: BATCH_UPDATE vs STREAM_UPDATE regimes, `contents_delta_uri` + overwrite semantics, immutable `config` (dimensions, algorithm arms tree-AH XOR brute-force, shard size, distance measure, feature norm).
- CEL enforces what the provider's empty `ExactlyOneOf` does not: algorithm XOR and tree-AH requires `approximate_neighbors_count`.
- 43-case spec test; 5 outputs including `index_id` (the deployed index's composition key).

### GcpVertexAiIndexEndpoint (673, `gcpvaiep`)
- Three mutually exclusive connectivity arms (public / VPC-peered / PSC) with pairwise CEL exclusion.
- Network self-link → relative form normalization on both engines (session-022 SCP precedent).
- Explicit docs distinguishing this from online-prediction `GcpVertexAiEndpoint` (671).

### GcpVertexAiDeployedIndex (674, `gcpvdidx`)
- Catalog's first kind FK-referencing two same-session siblings (index + index endpoint).
- Honest no-labels/no-project documentation — the GCP API carries neither; none is faked.
- `reserved_ip_ranges` → `GcpGlobalAddress.name` (Vertex peering ranges are global INTERNAL VPC_PEERING addresses).
- Automatic XOR dedicated sizing; JWT auth config; deployment group ↔ reserved-ranges pairing contract taught.

### Shared infrastructure
- Registry 672/673/674 with deployed-index prerequisites; kind-map regen; `pkg/outputs` conformance ×3.
- REST-probe verifiers ×3 (deployed index located inside endpoint GET's `deployedIndexes[]`).
- E2E runner: `${E2E_RUN_ID_UNDERSCORE}` token for hyphen-forbidding identifier classes.

## Live E2E (RoE §18 ladder)

| Rung | Scenario | Pulumi | Terraform |
|------|----------|--------|-----------|
| Index | `stream-minimal` (empty STREAM_UPDATE) | ✅ ~57s | ✅ ~76s |
| Index endpoint | `public` | ✅ ~35s | ✅ ~33s |
| Deployed index | `automatic-public` (composed chain) | ✅ ~34 min | ✅ ~17 min deploy + chain |

Live-found class: GCP holds a failed/undeploying `deployed_index_id` project-wide — the hold survives deleting the parent endpoint. Fixed with run-scoped `${E2E_RUN_ID_UNDERSCORE}` in the scenario.

## Validation

- Spec tests: 43 + 23 + 34 cases, all green.
- `validate-refs --check`, `secret-coverage --check`, `validate-outputs` ×6 module dirs, `pkg/outputs` conformance ×3: green.
- Offline converter `tofu plan` ×3: green.
- Audits ×3: Fully Complete PARITY ✅; site catalog regenerated (`npm run copy-docs`).

## Deliberate Exclusions (7.x-only at pinned tag)

- `encryption_spec` (CMEK) on index + endpoint
- `deletion_policy` on all three (Pulumi pins `DELETE` for parity)
- `psc_automation_configs` on the endpoint

## Workflow Uplift

- `e2e/README.md`: underscore-token contract + sub-resource-ID hold class.
- `forge-catalog-kind.mdc`: same learnings for future forges.

# Rename GcpVpc to GcpVpcNetwork

**Date**: July 4, 2026
**Type**: Refactor
**Components**: API Definitions, GCP Provider, Registry, E2E Framework, InfraCharts, Rename Workflow

## Summary

Renamed `GcpVpc` (610) to `GcpVpcNetwork` with folder/E2E slug `gcpvpcnetwork`; id_prefix stays `gcpvpc` (abbreviations don't track kind names — registry uniqueness is the only requirement). Zero spec or module behavior change — naming alignment with GCP's VPC network grain before GKE/AlloyDB depth work. Live dual-engine E2E proves leaf + subnetwork FK composition on both engines.

## Problem Statement / Motivation

The kind name `GcpVpc` was an internal abbreviation that diverged from GCP's native "VPC network" vocabulary and from the catalog grain used by sibling kinds. With ~20 sibling `default_kind` consumers, six infra charts, and every E2E PSA chain referencing the root network kind, deferring the rename widened fan-in cost every session.

## Solution

### Registry

- Enum `GcpVpc = 610` → `GcpVpcNetwork = 610` (value preserved)
- id_prefix `gcpvpc` unchanged (kept deliberately — short, unique, matches the registry's abbreviation convention)
- Prerequisites on `GcpSubnetwork`, `GcpRouterNat`, `GcpServiceNetworkingConnection`, `GcpAddress` updated

### Component tree

- Folder `gcpvpc/` → `gcpvpcnetwork/`; nested types (`GcpVpcNetworkSpec`, routing enums, etc.)
- Both IaC modules unchanged in behavior; TF `planton-ai_kind` label uses slug `gcpvpcnetwork`

### Repo-wide fan-in

- 20 sibling proto FK sites, ~45 explicit `kind:` YAML hits, six chart templates
- ~38 sibling catalog slug links, four consumer prerequisite file renames
- E2E harness slugs, verifier map, `dependencies_test.go`
- `pkg/outputs/conformance_test.go` + regenerated kind map

### Workflow uplift

Extended `_rules/catalog-kind/rename/rename-catalog-kind.mdc` with high-fan-in checklist, four-way naming table, safe replace order, script scope honesty, and live E2E minimum (leaf + one FK consumer).

## Validation

Offline green: 17/17 spec tests on `gcpvpcnetwork/v1`, `validate-refs --check`, `secret-coverage --check`, `pkg/outputs` + `pkg/refcheck`, release-equivalent Pulumi build, `tofu validate`, hack manifest + presets validated, locally built CLI `validate-outputs` on both module dirs.

Live E2E (project `planton-e2e`, zero orphans):

| Kind | Scenario | Pulumi | Terraform |
|------|----------|--------|-----------|
| GcpVpcNetwork | minimal | ✅ ~65s | ✅ ~73s |
| GcpSubnetwork | minimal (composed FK) | ✅ ~2m52s | ✅ ~2m45s |

## Breaking change

User manifests must use `kind: GcpVpcNetwork`. No backward-compat alias (`AliasMap` empty). Phase 2 web integration deferred until OSS release + `make upgrade-planton`.

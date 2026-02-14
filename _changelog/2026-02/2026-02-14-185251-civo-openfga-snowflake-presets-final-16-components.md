# Civo, OpenFGA, and Snowflake Presets -- Final 16 Components (213/213)

**Date**: February 14, 2026
**Type**: Feature
**Components**: Presets, Civo Provider, OpenFGA Provider, Snowflake Provider, API Definitions

## Summary

Created production-quality presets for the final 16 OpenMCF deployment components (Civo 12, OpenFGA 3, Snowflake 1), bringing the presets system to 100% coverage across all 213 components and 11 providers. Also fixed two CivoVpc proto issues and updated the OpenFGA RelationshipTuple hack manifest to use the structured proto-correct form.

## Problem Statement / Motivation

The presets system was at 192/213 (90%) coverage after T01-T08a completed presets for AWS, GCP, Azure, Kubernetes, OpenStack, Scaleway, DigitalOcean, and Cloudflare. Three providers remained without presets: Civo (12 components), OpenFGA (3 components), and Snowflake (1 component).

### Pain Points

- 16 components had no presets directory, leaving gaps in the "every component gets a starting point" promise
- CivoVpc had two proto anomalies: a `civo_credential_id` field (unique across all 213 components) and a `string region` instead of the `CivoRegion` enum used by all other Civo components
- OpenFGA RelationshipTuple hack manifest used flat strings instead of the structured nested messages defined in spec.proto

## Solution / What's New

### 30 Presets Across 16 Components

| Provider | Components | Presets | Files |
|----------|-----------|---------|-------|
| Civo | 12 | 22 | 44 |
| OpenFGA | 3 | 5 | 10 |
| Snowflake | 1 | 3 | 6 |
| **Total** | **16** | **30** | **60** |

### Proto Fixes

- **CivoVpc**: Removed `civo_credential_id` field (field 1) -- credentials should be handled externally, not in the spec
- **CivoVpc**: Changed `region` from `string` to `CivoRegion` enum for consistency with all other Civo components

### Hack Manifest Fix

- **OpenFGA RelationshipTuple**: Updated from flat string format (`user: "user:anne"`) to structured proto-correct format with nested `user`/`object` messages

## Implementation Details

### Civo Presets (22 presets across 12 components)

- **Foundation**: CivoVpc (1), CivoIpAddress (1), CivoVolume (2), CivoBucket (2), CivoFirewall (2) -- 8 presets
- **Compute/K8s**: CivoComputeInstance (2), CivoKubernetesCluster (2), CivoKubernetesNodePool (2) -- 6 presets
- **Data/DNS**: CivoDatabase (2), CivoDnsZone (2), CivoDnsRecord (2) -- 6 presets
- **Certificate**: CivoCertificate (1) -- Let's Encrypt preset with IaC-pending warning (Civo provider does not yet support certificates in Terraform/Pulumi)

Key design decisions:
- Engine-specific database presets (PostgreSQL production, MySQL dev) consistent with AWS/GCP/DO approach
- CivoDnsZone includes inline records (consistent with DigitalOcean) since Civo's zone model supports this as the primary pattern
- CivoFirewall uses distinct inbound rules per tier (web: 80/443/22, database: 5432/3306) with restricted CIDRs
- CivoComputeInstance presets include cloud-init for production, skip it for development

### OpenFGA Presets (5 presets across 3 components)

- **OpenFgaStore**: 1 preset (single required field)
- **OpenFgaAuthorizationModel**: 2 presets (RBAC DSL, hierarchical document access DSL) -- DSL format chosen over JSON for human readability
- **OpenFgaRelationshipTuple**: 2 presets (direct user-document access, group membership) -- structured proto-correct form with nested user/object messages

### Snowflake Presets (3 presets for 1 component)

- **01-production**: 30-day Time Travel, WARN logging, persistent (Fail-safe enabled)
- **02-development**: Transient, 1-day retention, DEBUG logging, console output enabled
- **03-iceberg-analytics**: Snowflake catalog, external volume, OPTIMIZED serialization policy

## Benefits

- **100% preset coverage** -- all 213 OpenMCF components across 11 providers now have at least one preset
- **375 total presets** -- up from 345, providing ready-to-deploy configurations for every cloud resource kind
- **Proto consistency** -- CivoVpc now uses the same `CivoRegion` enum as all other Civo components
- **Proto hygiene** -- removed the only credential ID field found in any spec across the entire framework

## Impact

- **Users**: Every OpenMCF component now has a "30-second starting point" -- no component is left without a preset
- **Developers**: The presets system is complete; future work focuses on quality improvements and additional presets for high-variety components
- **Platform**: The `civo_credential_id` removal and `region` type fix in CivoVpc are breaking changes for the IaC module (to be fixed separately)

## Related Work

- T01: Foundation (convention, rules, Forge integration, pilot preset)
- T02: AWS presets (25 components, 49 presets)
- T03: GCP presets (19 components, 36 presets)
- T04: Azure presets (29 components, 55 presets)
- T05: Kubernetes presets (51 components, 83 presets)
- T06: OpenStack presets (27 components, 44 presets)
- T07: Scaleway presets (18 components, 35 presets)
- T08a: DigitalOcean + Cloudflare presets (23 components, 43 presets)
- **T08b: This changelog** -- Civo + OpenFGA + Snowflake presets (16 components, 30 presets)

---

**Status**: Production Ready
**Timeline**: Single session (~2 hours)

# GcpBigQueryDataset Deployment Component

**Date**: February 15, 2026
**Type**: Feature
**Components**: API Definitions, GCP Provider, Pulumi Module, Terraform Module, Documentation

## Summary

Added GcpBigQueryDataset as the first Data & Analytics resource in the GCP provider expansion (R05 of 21). This complete deployment component provides BigQuery dataset provisioning with location control, CMEK encryption, authoritative access management, storage billing model selection, and time travel configuration -- all with full Pulumi and Terraform feature parity.

## Problem Statement / Motivation

OpenMCF's GCP coverage lacked any data analytics resources. BigQuery is Google Cloud's flagship analytics service, and the dataset is the foundational infrastructure boundary -- controlling data location, encryption, access, and lifecycle. Without it, OpenMCF couldn't compose data analytics infra charts or serve teams building BigQuery-based data platforms.

### Pain Points

- No BigQuery support in OpenMCF prevented data analytics infra chart composition
- Teams couldn't provision BigQuery datasets with cross-resource CMEK wiring
- No declarative access control management for BigQuery datasets
- Missing storage billing model control (LOGICAL vs PHYSICAL) for cost optimization

## Solution / What's New

### Complete Deployment Component

A full GcpBigQueryDataset deployment component following the established OpenMCF patterns, delivering:

- **Proto API** (4 proto files) with 14 spec fields, 2 sub-messages, 4 CEL validations
- **Pulumi module** (4 Go files) with full field mapping including access entries
- **Terraform module** (6 HCL files) with dynamic access blocks and feature parity
- **31 validation tests** (19 positive, 12 negative) covering all validation rules
- **3 presets** (basic-analytics, cmek-encrypted, team-shared)
- **Production documentation** (README, examples with 6 scenarios, research docs, catalog page)

## Implementation Details

### Spec Design (Corrections from T01 Plan)

Seven corrections were applied to the original T01 plan guidance after deep study of the Terraform provider and Pulumi SDK:

1. **`kms_key_id` renamed to `kms_key_name`** -- matches GCP's native terminology
2. **Added `storage_billing_model`** -- LOGICAL vs PHYSICAL billing, significant cost lever missing from plan
3. **`max_time_travel_hours` as `int32`** -- semantically correct despite TF/Pulumi using string
4. **`dataset_id` validation** -- regex `^[0-9A-Za-z_]+$`, max 1024 chars (no hyphens!)
5. **`default_table_expiration_ms` minimum** -- CEL validation for >= 3600000 (1 hour)
6. **Replaced `etag` output with `project`** -- etag has no infra-chart composition value
7. **Explicit access entry sub-message** -- `GcpBigQueryDatasetAccessEntry` with 7 identity types

### Access Control Model

The `access` field maps BigQuery's authoritative access model. Each entry supports:
- Role-based: user_by_email, group_by_email, domain, special_group, iam_member
- View-based: authorized view references (no role required)

Advanced features deliberately excluded: condition (CEL-based conditional access), routine (authorized routines), dataset (authorized datasets).

### Key Outputs for Composition

| Output | Infra Chart Use |
|--------|----------------|
| `dataset_id` | SQL queries, job configs, downstream references |
| `self_link` | API references, audit trails |
| `project` | Cross-project dataset references |
| `creation_time` | Metadata and auditing |

### StringValueOrRef Fields

- `project_id` -- default_kind: GcpProject
- `kms_key_name` -- default_kind: GcpKmsKey (references `status.outputs.key_id`)

## Benefits

- **Data analytics infra charts unblocked** -- BigQuery Dataset is Layer 1 in data-analytics-environment, ml-notebook-environment, and event-pipeline compositions
- **CMEK composability** -- seamless `valueFrom` wiring to GcpKmsKey for encryption
- **Cost control** -- storage billing model selection (PHYSICAL can save 60-80%)
- **Team collaboration** -- declarative, authoritative access control with single source of truth
- **Quality bar maintained** -- 31 passing tests, both IaC implementations validated, production-quality documentation

## Impact

- **OpenMCF GCP coverage**: 20 resource kinds (from 19) -- 5 of 21 expansion resources complete
- **Infra chart readiness**: BigQuery Dataset unlocks 3 planned infra charts (data-analytics-environment, ml-notebook-environment, event-pipeline)
- **Users**: Any GCP user can now provision BigQuery datasets through OpenMCF with full lifecycle management

## Related Work

- Part of **20260215.01.sp.gcp-resource-expansion** (R05 of 21)
- Builds on R03 GcpKmsKeyRing and R04 GcpKmsKey for CMEK composition
- Next resource: R06 GcpPubSubTopic (Messaging)

---

**Status**: Production Ready
**Files Created**: 35 files across proto, Go, HCL, YAML, and Markdown
**Tests**: 31 passing (19 positive + 12 negative)

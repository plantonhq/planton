# GCP Global Address Resource Kind

**Date**: February 15, 2026
**Type**: Feature
**Components**: API Definitions, GCP Provider, Pulumi Module, Terraform Module

## Summary

Added `GcpGlobalAddress` as a new catalog kind to Planton, covering static IP address reservations at global scope in Google Cloud Platform. The kind supports external static IPs for HTTP(S) load balancers, internal IP ranges for VPC peering (Cloud SQL, Redis, AlloyDB private networking), and Private Service Connect addresses. Includes proto API with CEL cross-field validation, dual IaC backends (Pulumi + Terraform), 20 passing validation tests, 3 presets, and production-quality documentation.

## Problem Statement / Motivation

Platform engineers provisioning GCP infrastructure through Planton could create VPCs, firewall rules, GKE clusters, and DNS records, but had no way to reserve static global IP addresses. This gap meant:

### Pain Points

- No way to reserve static IPs for HTTP(S) load balancers through Planton
- VPC peering for managed services (Cloud SQL, Redis, AlloyDB) required manual address range reservation outside Planton
- Private Service Connect endpoints couldn't be provisioned as part of an Planton-managed environment
- Infra charts composing load-balanced environments had no `GcpGlobalAddress` to reference via `StringValueOrRef`

## Solution / What's New

A complete `GcpGlobalAddress` catalog kind covering the `google_compute_global_address` GCP resource with full lifecycle management.

### Kind Structure

```
apis/dev/planton/provider/gcp/gcpglobaladdress/v1/
├── spec.proto              # 9 fields, 3 CEL cross-field rules
├── outputs.proto     # address, self_link, creation_timestamp
├── api.proto               # KRM envelope (GcpGlobalAddress + Status)
├── iac_input.proto       # target + GcpProviderConfig
├── spec_test.go            # 20 tests (7 positive, 13 negative)
├── *.pb.go                 # Generated Go stubs
├── README.md               # User-facing overview
├── examples.md             # 6 copy-paste YAML examples
├── catalog-page.md         # Kind catalog entry
├── docs/README.md          # Comprehensive research document
├── presets/                 # 3 presets (YAML + MD each)
│   ├── 01-external-static-ip.*
│   ├── 02-internal-vpc-peering-range.*
│   └── 03-private-service-connect.*
└── iac/
    ├── hack/manifest.yaml  # Test manifest
    ├── pulumi/             # Go Pulumi module (4 files + support)
    │   ├── module/{main,locals,global_address,outputs}.go
    │   ├── main.go, Pulumi.yaml, Makefile, debug.sh
    │   ├── README.md, overview.md
    └── tf/                 # Terraform module (5 files + README)
        ├── provider.tf, variables.tf, locals.tf, main.tf, outputs.tf
        └── README.md
```

## Implementation Details

### Proto API Design

**Spec fields (9 total)**:

| Field | Type | Notes |
|-------|------|-------|
| `project_id` | StringValueOrRef | Required, defaults to GcpProject |
| `address_name` | string | Required, RFC1035 pattern validated |
| `address` | string | Optional, GCP auto-assigns if omitted |
| `address_type` | optional string | Default "EXTERNAL", CEL: EXTERNAL/INTERNAL |
| `description` | string | Optional human-readable description |
| `ip_version` | optional string | Default "IPV4", CEL: IPV4/IPV6 |
| `network` | StringValueOrRef | Optional, defaults to GcpVpc, required for INTERNAL |
| `prefix_length` | optional int32 | Range 8-29, required for VPC_PEERING |
| `purpose` | string | CEL: empty/VPC_PEERING/PRIVATE_SERVICE_CONNECT |

**Cross-field CEL validations**:
1. `purpose_requires_internal` — purpose can only be set when address_type is INTERNAL
2. `vpc_peering_requires_prefix_length` — prefix_length required for VPC_PEERING
3. `internal_requires_network` — network required when address_type is INTERNAL

### Corrections from Planning Phase

The implementation refined the T01 plan spec in 6 ways:
1. **Added `address_name`** — GCP requires an explicit RFC1035-compliant name; the plan omitted it
2. **Added `address` field** — allows users to reserve specific IPs (BYOIP)
3. **Added `description` field** — standard GCP resource metadata
4. **Removed `address_id` output** — not a real GCP computed output; `self_link` serves this purpose
5. **Added `creation_timestamp` output** — free computed metadata from the API
6. **Added 3 CEL cross-field rules** — catches invalid combinations before hitting the GCP API

### Validation Tests (20 total)

- 7 positive cases: minimal external, full external, IPv6, VPC peering, PSC, prefix bounds
- 13 negative cases: missing project_id, missing address_name, invalid name patterns, invalid address_type, invalid ip_version, invalid purpose, purpose+EXTERNAL, purpose without address_type, VPC_PEERING without prefix_length, prefix bounds, INTERNAL without network

### Enum Registration

Registered as `GcpGlobalAddress = 621` with id_prefix `gcpgip` in `catalog_kind.proto` (GCP range 600-799).

## Benefits

- **Infra-chart composable**: `StringValueOrRef` on project_id and network enables DAG-based deployment in composed environments
- **Early validation**: CEL rules catch 3 common misconfiguration patterns before deployment
- **Dual IaC backends**: Feature parity between Pulumi and Terraform implementations
- **Production-ready presets**: 3 presets cover the most common address reservation patterns
- **Complete documentation**: Research doc, user docs, examples, catalog page

## Impact

- **GCP resource count**: 19 → 20 (second new resource in the expansion project)
- **Downstream enablement**: Future resources like load balancers and CDN can reference global addresses via `StringValueOrRef`
- **Infra charts**: `gcp-gke-environment` and `gcp-serverless-api-backend` charts can now include static IP provisioning

## Related Work

- **R01 GcpFirewallRule** (completed 2026-02-15): First resource in the GCP expansion project; established forge patterns used here
- **GCP Resource Expansion sub-project** (20260215.01.sp.gcp-resource-expansion): This is R02 of 21 planned GCP resource kinds
- **Parent project** (20260212.01.planton-cloud-provider-expansion): Cross-provider expansion initiative

---

**Status**: ✅ Production Ready
**Timeline**: Single session (~2 hours)

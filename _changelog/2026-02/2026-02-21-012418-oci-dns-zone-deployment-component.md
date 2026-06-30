# OCI DNS Zone Deployment Component

**Date**: February 21, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi CLI Integration, Provider Framework

## Summary

Implemented the OciDnsZone deployment component -- OCI's managed authoritative DNS zone supporting both public (GLOBAL) and private resolution scopes, PRIMARY and SECONDARY zone types, zone transfers via external masters/downstreams, and DNSSEC signing. First resource of Phase 10 (DNS and Certificates).

## Problem Statement / Motivation

The Planton Oracle Cloud provider needs DNS infrastructure components to enable domain management for OCI workloads. DNS zones are the foundational building block -- without them, users have no declarative way to provision public or private DNS resolution for their OCI services.

### Pain Points

- No DNS component existed in the OCI provider catalog
- Teams deploying OCI workloads had no way to declaratively manage DNS zones
- Phase 10 (DNS and Certificates) was entirely unstarted
- Infra charts requiring DNS (e.g., OKE Environment) could not include zone provisioning

## Solution / What's New

A complete deployment component (`OciDnsZone`) with proto API definitions, Pulumi module (Go), and Terraform module (HCL), registered as CloudResourceKind 3390.

### Key Design Decisions

**Scope as optional enum with unspecified=GLOBAL**: Most DNS zones are public. Omitting scope lets OCI default to GLOBAL, reducing YAML ceremony for the common case. The `scope_unspecified` zero-value avoids collision with `ZoneType.unspecified`. Private zones are the exception and require explicit `scope: private`.

**Shared ExternalServer message**: `external_masters` and `external_downstreams` have identical structure (address, port, tsig_key_id). A single nested message avoids duplication while serving both inbound (SECONDARY) and outbound (PRIMARY) zone transfer configurations.

**DNSSEC as optional bool**: `is_dnssec_enabled: true` is cleaner YAML than `dnssecState: enabled`. The underlying ENABLED/DISABLED string conversion is hidden in IaC modules. Nil means OCI default (disabled).

**view_id without default_kind**: OCI DNS Views are VCN-internal constructs, not standalone deployable resources in the Planton catalog. Using StringValueOrRef without `default_kind` allows passing OCIDs directly.

**tsig_key_id as plain string**: TSIG keys are not modeled as Planton components. They are pre-existing DNS infrastructure typically managed outside IaC.

**4 CEL validation rules**: zone_type required, private requires view_id, no private+secondary (OCI limitation), secondary requires external_masters. These catch invalid configurations at schema validation time rather than at provider API call time.

## Implementation Details

### Proto API (4 files)

- **spec.proto**: 7 fields, 2 embedded enums (ZoneType, Scope), 1 nested message (ExternalServer), 4 CEL rules
- **api.proto**: Standard KRM wiring (OciDnsZone, OciDnsZoneStatus)
- **stack_input.proto**: OciDnsZoneStackInput with target + provider config
- **stack_outputs.proto**: 2 outputs (`zone_id`, `nameservers`)

### Spec Fields

| Field | Type | Notes |
|-------|------|-------|
| `compartment_id` | StringValueOrRef (required) | default_kind: OciCompartment |
| `zone_type` | ZoneType enum (required) | CEL: != unspecified; ForceNew |
| `scope` | Scope enum (optional) | unspecified = GLOBAL; ForceNew |
| `view_id` | StringValueOrRef (optional) | Required when scope=private; ForceNew |
| `is_dnssec_enabled` | optional bool | ENABLED/DISABLED; nil = OCI default |
| `external_masters` | repeated ExternalServer | Required for SECONDARY zones |
| `external_downstreams` | repeated ExternalServer | Only for PRIMARY+GLOBAL zones |

### Validation Tests

24 Ginkgo/Gomega tests (12 valid, 12 invalid scenarios) covering minimal primary global zone, explicit global scope, private zone with view_id, secondary zone with external masters, external downstreams, DNSSEC toggle, multiple masters with TSIG keys, combined masters+downstreams, valueFrom refs for compartment_id and view_id, full configuration, and all CEL validation rules (zone_type required, private requires view_id, no private+secondary, secondary requires external_masters, empty address rejected).

### Pulumi Module (5 files)

- `main.go`: Entry point with stack input loading
- `module/main.go`: Resources orchestrator with OCI provider setup
- `module/locals.go`: Locals struct with freeform tags from metadata labels
- `module/outputs.go`: Output constants (`zone_id`, `nameservers`)
- `module/zone.go`: `dnsZone()` creating `dns.NewZone()` with conditional scope/DNSSEC/external servers; `buildExternalMasters()` and `buildExternalDownstreams()` helpers; nameserver hostname extraction via `ApplyT` with comma-join

### Terraform Module (5 files)

- `main.tf`: `oci_dns_zone.this` with dynamic `external_masters` and `external_downstreams` blocks
- `locals.tf`: Freeform tags, `zone_type_map`, scope/DNSSEC ternary conversion
- `outputs.tf`: `zone_id` and `nameservers` (join from nameserver objects)
- `variables.tf`: Metadata and spec type definitions with optional fields
- `provider.tf`: OCI provider requirement (>= 5.0)

### Kind Registration

`OciDnsZone = 3390` registered under new "DNS and Certificates" section in `cloud_resource_kind.proto`, `kind_map_gen.go` regenerated.

## Benefits

- Enables declarative provisioning of public and private DNS zones on OCI
- Supports enterprise DNS scenarios: SECONDARY zone replication from on-prem masters, outbound zone transfers to external secondaries
- DNSSEC toggle provides one-line security hardening for public zones
- Two outputs (`zone_id`, `nameservers`) enable composability with OciDnsRecord (R35) and domain registrar configuration
- StringValueOrRef on compartment_id enables infra-chart composability

## Impact

- **Users**: Can now declaratively manage DNS zones covering public hosting, private VCN resolution, hybrid DNS with zone transfers, and DNSSEC signing
- **Platform**: Phase 10 (DNS and Certificates) started -- 1/2 resources done
- **Infra Charts**: OKE Environment and Serverless Stack charts can now incorporate DNS zone provisioning

## Related Work

- **OciDnsRecord** (R35): Record sets within a zone, next resource of Phase 10
- **OciCompartment** (R04): Compartment referenced via `compartment_id`
- **OciVcn** (R01): Private DNS zones resolve within VCN context

---

**Status**: Production Ready

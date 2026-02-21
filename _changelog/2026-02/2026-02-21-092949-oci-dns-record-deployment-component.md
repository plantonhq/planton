# OCI DNS Record Deployment Component

**Date**: February 21, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi CLI Integration, Provider Framework

## Summary

Implemented the OciDnsRecord deployment component -- OCI's DNS record set (RRSet) resource for managing all DNS records of a given type for a specific domain within a zone. This is the 37th and final OCI resource kind, completing Phase 10 (DNS and Certificates) and the entire OCI resource implementation queue.

## Problem Statement / Motivation

OciDnsZone (R34) provides authoritative DNS zones, but without record management there is no way to declaratively create A, AAAA, CNAME, MX, TXT, or other DNS records within those zones. This was the last missing piece of the OCI resource catalog.

### Pain Points

- DNS zones were deployable but records could not be managed declaratively
- Phase 10 (DNS and Certificates) was incomplete at 1/2 resources
- The full 37-resource OCI catalog could not be marked complete
- Infra charts requiring DNS records (OKE Environment, Serverless Stack) lacked the building block

## Solution / What's New

A complete deployment component (`OciDnsRecord`) wrapping `oci_dns_rrset` with proto API definitions, Pulumi module (Go), and Terraform module (HCL), registered as CloudResourceKind 3391.

### Key Design Decisions

**No compartment_id (departure from convention)**: Both Terraform and Pulumi providers mark `compartment_id` as deprecated on `oci_dns_rrset` -- it is inferred from the target zone. This is the first OCI component without `compartment_id` as field 1. Including it would add YAML ceremony for zero benefit.

**No freeform_tags**: DNS record sets do not support OCI tagging. Unlike zone-level resources, individual RRSets are not taggable objects in the OCI data model.

**Simplified record items**: The provider requires each item to carry redundant `domain` and `rtype` fields (must match the top-level values). The OpenMCF spec strips items to just `rdata` + `ttl`, and IaC modules inject domain and rtype from the spec-level fields. This eliminates YAML redundancy.

**rtype as plain string (not enum)**: DNS record types (A, AAAA, CNAME, MX, TXT, NS, SRV, PTR, CAA) are IETF-standardized uppercase tokens. Proto enum convention requires lowercase values, which would force users to write `rtype: a` instead of the natural `rtype: "A"`. A plain string passes through to OCI without conversion.

**Empty stack outputs**: DNS record sets do not produce an OCID or composable identifier. The resource is identified by its (zone, domain, rtype) tuple, all of which are inputs. No downstream component references an RRSet by ID.

**zone_name_or_id naming**: Preserves the provider's dual-use field name rather than simplifying to `zone_id`, because the OCI DNS API genuinely accepts either a zone name or OCID.

## Implementation Details

### Proto API (4 files)

- **spec.proto**: 5 fields, 1 nested message (RecordItem), no enums, no CEL rules
- **api.proto**: Standard KRM wiring (OciDnsRecord, OciDnsRecordStatus)
- **stack_input.proto**: OciDnsRecordStackInput with target + provider config
- **stack_outputs.proto**: Empty (no composable outputs)

### Spec Fields

| Field | Type | Notes |
|-------|------|-------|
| `zone_name_or_id` | StringValueOrRef (required) | default_kind: OciDnsZone; ForceNew |
| `domain` | string (required) | FQDN within the zone; ForceNew |
| `rtype` | string (required) | DNS record type (A, AAAA, etc.); ForceNew |
| `view_id` | StringValueOrRef (optional) | For private zone access by name; ForceNew |
| `items` | repeated RecordItem (min 1) | rdata + ttl per record |

### Validation Tests

23 Ginkgo/Gomega tests (12 valid, 11 invalid scenarios) covering minimal A record, multiple A records, AAAA, CNAME, MX, TXT, view_id, valueFrom refs, TTL edge cases (1s, 86400s), zone name vs OCID, and all field validation rules (missing zone_name_or_id, empty domain, empty rtype, empty items, empty rdata, zero ttl, negative ttl).

### Pulumi Module (4 files)

- `module/main.go`: Resources orchestrator with OCI provider setup
- `module/locals.go`: Locals struct (no freeform tags -- rrsets don't support tagging)
- `module/rrset.go`: `rrset()` creating `dns.NewRrset()` with `dns.RrsetItemArray` built from simplified spec items, injecting domain and rtype per item; conditional `ViewId`
- `iac/pulumi/main.go`: Entrypoint with stack input loading

### Terraform Module (5 files)

- `main.tf`: `oci_dns_rrset.this` with dynamic `items` block injecting domain/rtype from spec level
- `locals.tf`: resource_id only (no freeform_tags)
- `outputs.tf`: No outputs (comment explaining why)
- `variables.tf`: Metadata and spec type definitions
- `provider.tf`: OCI provider requirement (>= 5.0)

### Kind Registration

`OciDnsRecord = 3391` registered under "DNS and Certificates" section with `id_prefix: "ocidnsr"` in `cloud_resource_kind.proto`, `kind_map_gen.go` regenerated.

## Benefits

- Enables declarative management of DNS records (A, AAAA, CNAME, MX, TXT, NS, SRV, PTR, CAA) within OCI DNS zones
- Simplified YAML UX: items contain only rdata + ttl, not redundant domain/rtype
- Atomic record set management: updates replace the entire RRSet for predictable state
- StringValueOrRef on zone_name_or_id enables infra-chart composability with OciDnsZone
- Completes the entire 37-resource OCI catalog

## Impact

- **Users**: Can now declaratively manage DNS records for any record type within OCI DNS zones
- **Platform**: Phase 10 (DNS and Certificates) now 100% complete; all 37 OCI resource kinds implemented
- **Infra Charts**: All 5 planned infra charts now have their full building block set available
- **Project**: The resource implementation queue (T02) is fully complete; project transitions to post-resource phases (infra charts)

## Related Work

- **OciDnsZone** (R34): Parent zone resource referenced via `zone_name_or_id`
- **Infra Charts**: OKE Environment and Serverless Stack charts can now include DNS record provisioning
- **Post-Resource Phase**: All 37 resources done; 5 infra charts and sub-project completion remain

---

**Status**: Production Ready

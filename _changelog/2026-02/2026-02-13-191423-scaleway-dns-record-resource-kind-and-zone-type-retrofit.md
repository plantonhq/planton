# Scaleway DNS Record Resource Kind and Zone Record Type Retrofit

**Date**: February 13, 2026
**Type**: Feature
**Components**: API Definitions, Protobuf Schemas, Pulumi CLI Integration, Provider Framework

## Summary

Implemented ScalewayDnsRecord (R16), the sixteenth Scaleway resource kind, completing the DNS tier alongside ScalewayDnsZone (R15). Simultaneously retrofitted ScalewayDnsZone to use a local `RecordType` enum replacing the shared `DnsRecordType`, adding DNAME and TLSA support to both components. This establishes the design principle that DNS record type enums are component-local, not shared across providers or even between zone and record components of the same provider.

## Problem Statement / Motivation

The Scaleway DNS tier needed a standalone DNS record kind for DAG-friendly record management in infra charts. While ScalewayDnsZone's inline records are convenient for static records known at zone creation time, records whose values come from other infrastructure resources (A records pointing to Load Balancer IPs, CNAMEs to Kapsule cluster endpoints) need to be separate resources with explicit dependency edges.

### Pain Points

- No standalone record kind for Scaleway DNS -- infra charts couldn't express record-level dependencies
- ScalewayDnsZone used the shared `DnsRecordType` enum which lacked DNAME and TLSA support
- The shared enum pattern created unnecessary coupling between providers and components
- Documentation falsely stated "use standalone ScalewayDnsRecord for DNAME/TLSA" before that kind existed

## Solution / What's New

### Part A: ScalewayDnsZone Record Type Retrofit

Replaced the shared `DnsRecordType` import with a local `RecordType` enum nested inside `ScalewayDnsZoneRecord`. Added DNAME and TLSA as the 12th and 13th record types. Updated both Pulumi Go and Terraform HCL type maps, and removed all documentation caveats about DNAME/TLSA being unsupported in inline records.

### Part B: ScalewayDnsRecord Resource Kind

Implemented as a standalone (non-composite) resource wrapping a single `scaleway_domain_record` Terraform resource. Two `StringValueOrRef` inputs create infra chart dependency edges:

- `zone_name` -> ScalewayDnsZone's `status.outputs.zone_name`
- `data` -> any resource's output (no `default_kind`, since record values can reference Load Balancers, Instances, Kapsule clusters, etc.)

Key design choices:
- **Local RecordType enum** with all 13 Scaleway-supported types (matching the zone's local enum but independent)
- **Simpler than DigitalOcean** -- Scaleway embeds SRV weight/port and CAA flags/tag in the `data` field using standard DNS format, so no `weight`, `port`, `flags`, or `tag` fields needed
- **FQDN output** -- Scaleway computes the fully qualified domain name, eliminating manual string concatenation
- **No tags** -- Scaleway DNS API does not support tags (consistent with ScalewayDnsZone and ScalewayContainerRegistry)

## Implementation Details

### Discovery: `keep_empty_zone` Not Available in Current SDKs

The Scaleway Terraform provider docs list `keep_empty_zone` as an argument on `scaleway_domain_record`, but the installed provider version (v2.69.0) rejects it as unsupported. The Pulumi SDK similarly does not expose this field. The spec proto field is preserved for forward compatibility, with comments explaining the SDK gap. Both IaC modules skip the field with documentation for future enablement.

### Files Created/Modified

**Modified (ScalewayDnsZone retrofit):**
- `scalewaydnszone/v1/spec.proto` -- Replaced shared enum import with local RecordType (13 types)
- `scalewaydnszone/v1/iac/pulumi/module/dns_zone.go` -- Added DNAME + TLSA to type map
- `scalewaydnszone/v1/iac/tf/locals.tf` -- Added DNAME + TLSA to type map
- `scalewaydnszone/v1/README.md` -- Removed DNAME/TLSA caveats

**Created (ScalewayDnsRecord -- 17 new files):**
- 4 proto schemas: `api.proto`, `spec.proto`, `stack_outputs.proto`, `stack_input.proto`
- 6 Pulumi Go files: `Pulumi.yaml`, `main.go`, `module/main.go`, `module/locals.go`, `module/dns_record.go`, `module/outputs.go`
- 5 Terraform HCL files: `provider.tf`, `variables.tf`, `locals.tf`, `main.tf`, `outputs.tf`
- 2 documentation files: `README.md`, `examples.md`
- Plus auto-generated proto stubs and BUILD.bazel files

## Benefits

- **DAG-friendly DNS management** -- Infra charts can now express record-level dependencies with explicit edges
- **Complete Scaleway DNS type coverage** -- All 13 Scaleway record types (A, AAAA, ALIAS, CAA, CNAME, DNAME, MX, NS, PTR, SOA, SRV, TXT, TLSA) available in both zone and record kinds
- **Design principle established** -- Record type enums are component-local, enabling each kind to evolve its type surface independently
- **Simpler spec surface** -- 7 fields vs DigitalOcean's 10, thanks to Scaleway's self-contained data format

## Impact

- **Scaleway DNS tier complete** -- Both zone and record kinds implemented
- **16 of 19 Scaleway resource kinds done** (84%)
- **Infra chart readiness** -- The DNS record kind enables the `kapsule-environment` and `serverless-environment` infra charts to wire DNS records to dynamically provisioned infrastructure
- **Design precedent** -- The "no shared DNS enums" decision will guide future provider implementations

## Related Work

- **R15: ScalewayDnsZone** -- The companion zone kind that this record kind references
- **R05: ScalewayLoadBalancer** -- Primary upstream for A record data references
- **R07: ScalewayKapsuleCluster** -- Primary upstream for CNAME data references
- **DigitalOceanDnsRecord** -- Reference implementation (local RecordType enum pattern)

---

**Status**: Production Ready
**Timeline**: ~45 minutes implementation + verification

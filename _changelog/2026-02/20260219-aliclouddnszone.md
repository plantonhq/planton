# AliCloudDnsZone Component Added

**Date**: 2026-02-19
**Component**: AliCloudDnsZone
**Enum**: 3040
**ID Prefix**: acdns

## Summary

Added the AliCloudDnsZone deployment component -- manages DNS domains in the Alibaba Cloud Alidns service. This is the prerequisite for creating DNS records (A, AAAA, CNAME, MX, TXT, etc.) via the AliCloudDnsRecord component.

Registering a domain in Alidns does not purchase or transfer it -- it creates a hosted zone so that DNS records can be managed. Users point their domain registrar's NS records to the DNS servers returned in the stack outputs.

## What Was Created

### API Definition
- `apis/org/openmcf/provider/alicloud/aliclouddnszone/v1/` -- Full proto API (spec, api, stack_input, stack_outputs)
- Registered `AliCloudDnsZone = 3040` in `CloudResourceKind` enum under the DNS category

### IaC Modules
- **Pulumi** (Go): Creates alicloud provider and a single `dns.AlidnsDomain` resource with all spec fields mapped
- **Terraform** (HCL): Single `alicloud_alidns_domain` resource with matching variables, outputs, and tag merging

### Tests
- Ginkgo/Gomega spec validation tests: 11 specs covering valid inputs (minimal, full config, subdomain, tags only), missing required fields (region, domain_name), wrong api_version/kind, missing metadata, missing spec, and domain_name max length

### Documentation
- README.md with configuration reference, output reference, and related components
- examples.md with 3 YAML examples (minimal, with tags/resource group, with group assignment)
- catalog-page.md with full configuration reference, quick start, and examples
- docs/README.md with comprehensive research documentation

### Presets
- 01-standard: Minimal domain registration
- 02-organizational: Domain with resource group, tags, and remarks

## Spec Corrections from T02

- **`group_name` -> `group_id`**: T02 listed `group_name` as an input, but the provider takes `group_id`. `group_name` is a computed output.
- **Added `remark`**: Provider supports a remark field for domain description. Not in T02 but useful.
- **Added `tags`**: Consistent with all other Alibaba Cloud components.
- **Added `region`**: Alidns is global, but the provider requires a region for initialization.

## Fields Excluded for v1

- `lang` -- controls API response language; not meaningful for IaC users

## Verification

- `go build ./...` -- PASS
- `go vet ./...` -- PASS
- `go test ./...` -- PASS (11/11 specs)
- `terraform init` -- PASS (alicloud provider v1.271.0)
- `terraform validate` -- PASS

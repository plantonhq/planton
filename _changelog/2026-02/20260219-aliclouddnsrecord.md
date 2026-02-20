# AlicloudDnsRecord Component Added

**Date**: 2026-02-19
**Component**: AlicloudDnsRecord
**Enum**: 3041
**ID Prefix**: acdr

## Summary

Added the AlicloudDnsRecord deployment component -- manages DNS records within an Alibaba Cloud Alidns-hosted domain. Supports all standard record types (A, AAAA, CNAME, MX, TXT, NS, SRV, CAA) with configurable TTL, priority, DNS resolution lines, and record status.

The parent domain must already exist in Alidns, either managed by the AlicloudDnsZone component or added manually. This is a leaf resource -- nothing downstream depends on its outputs.

## What Was Created

### API Definition
- `apis/org/openmcf/provider/alicloud/aliclouddnsrecord/v1/` -- Full proto API (spec, api, stack_input, stack_outputs)
- Registered `AlicloudDnsRecord = 3041` in `CloudResourceKind` enum under the DNS category

### IaC Modules
- **Pulumi** (Go): Creates alicloud provider and a single `dns.AlidnsRecord` resource with all spec fields mapped. No tag computation (resource does not support tags).
- **Terraform** (HCL): Single `alicloud_alidns_record` resource with matching variables, outputs, and input validations for type and status.

### Tests
- Ginkgo/Gomega spec validation tests: 20 specs covering valid inputs (A record minimal, all-optional MX, CNAME, apex @, wildcard *, TXT/SPF, DISABLE status, CAA) and invalid inputs (missing region/domain_name/rr/type/value, invalid type, wrong api_version/kind, missing metadata/spec, domain_name max length, invalid status)

### Documentation
- README.md with configuration reference, output reference, and related components
- examples.md with 6 YAML examples (A, CNAME, MX with priority, TXT/SPF, wildcard, disabled record)
- catalog-page.md with full configuration reference, quick start, and examples
- docs/README.md with comprehensive research documentation including record type reference table

### Presets
- 01-a-record: Standard A record mapping subdomain to IPv4
- 02-cname-record: CNAME alias to another domain
- 03-mx-record: Mail exchange record with priority and higher TTL

## Spec Corrections from T02

- **`host_record` -> `rr`**: T02 used `host_record` but the provider field is `rr`. Per coding guidelines, field names match the provider.
- **Added `region`**: Needed for provider initialization; not in T02 spec.
- **Added `line`**: DNS resolution line for ISP/geo-based routing; not in T02.
- **Added `status`**: Record enable/disable; not in T02.
- **Added `remark`**: Description field; not in T02.
- **Expanded `type` values**: T02 listed 7 types; provider supports 10 (added CAA, REDIRECT_URL, FORWORD_URL).

## Fields Excluded for v1

- `lang` -- controls API response language; not meaningful for IaC users
- `user_client_ip` -- client IP for API tracking; not meaningful for IaC users
- `tags` -- resource does not support tags (unlike `alicloud_alidns_domain`)

## Verification

- `go build ./...` -- PASS
- `go vet ./...` -- PASS
- `go test ./...` -- PASS (20/20 specs)
- `terraform init` -- PASS (alicloud provider v1.271.0)
- `terraform validate` -- PASS

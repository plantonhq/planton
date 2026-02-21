# AlicloudCdnDomain Component Added

**Date**: 2026-02-20
**Component**: AlicloudCdnDomain
**Enum**: 3100
**ID Prefix**: accdn

## Summary

Added the AlicloudCdnDomain deployment component -- manages CDN accelerated domains in the Alibaba Cloud CDN service. A CDN domain maps a user-facing domain name to one or more origin servers; edge nodes worldwide cache and serve content, reducing latency for end users. After deployment, create a CNAME record at your DNS provider pointing to the `cname` stack output.

## What Was Created

### API Definition
- `apis/org/openmcf/provider/alicloud/alicloudcdndomain/v1/` -- Full proto API (spec, api, stack_input, stack_outputs)
- Registered `AlicloudCdnDomain = 3100` in `CloudResourceKind` enum under a new CDN category

### IaC Modules
- **Pulumi** (Go): Creates alicloud provider and a single `cdn.DomainNew` resource with all spec fields mapped, including sources array and optional certificate config
- **Terraform** (HCL): Single `alicloud_cdn_domain_new` resource with dynamic `sources` block, conditional `certificate_config` block, and tag merging

### Tests
- Ginkgo/Gomega spec validation tests: 23 specs covering valid inputs (minimal, full config, all cdn_types, all scopes, all source types, upload cert, CAS cert, cert status off), invalid inputs (missing required fields, invalid cdn_type, invalid scope, invalid source type, empty source content, invalid cert_type, invalid cert_status, wrong api_version/kind, missing metadata, missing spec, domain_name max length)

### Documentation
- README.md with configuration reference tables for all fields (spec, source, certificate), output reference, and related components
- examples.md with 4 YAML examples (minimal web CDN, multiple origins with failover, HTTPS with CAS cert, OSS bucket origin)
- catalog-page.md with full catalog documentation including quick start, prerequisites, and 3 deployment examples

## Design Decisions (Deviations from T02)

- **Dropped `alicloud_cdn_domain_config`**: T02 listed it as a bundled resource, but its schema (`function_name` + arbitrary key-value `function_args`) is too generic to model meaningfully in proto. The CDN domain is fully functional without it per DD07.
- **Added `certificate_config`**: Not in T02, but HTTPS is essential for production CDN. Included as optional nested message supporting `upload`, `cas`, and `free` certificate types.
- **Added `tags`, `resource_group_id`**: Standard per established pattern and DD05.
- **Added `check_url`**: Origin health check URL, simple and useful.
- **Added `weight` to sources**: Present in TF provider for load balancing across origins, omitted from T02.
- **Added source type `common`**: TF provider supports 4 types (`ipaddr`, `domain`, `oss`, `common`); T02 listed only 3.
- **Skipped `env`**: Grayscale testing feature, extremely niche.
- **Skipped `status`**: Lifecycle concern; deploy/destroy controls online/offline state.

## Verification

- `go build ./...` -- PASS
- `go vet ./...` -- PASS
- `go test ./...` -- PASS (23/23 specs)
- `terraform init` -- PASS (alicloud provider v1.271.0)
- `terraform validate` -- PASS

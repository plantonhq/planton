# AlicloudStorageBucket Component Added

**Date**: 2026-02-19
**Component**: AlicloudStorageBucket
**Enum**: 3050
**ID Prefix**: acoss

## Summary

Added the AlicloudStorageBucket deployment component -- the first Storage-tier resource in the Alibaba Cloud catalog. This component manages an Alibaba Cloud OSS bucket with configurable access control, storage class, zone redundancy, versioning, server-side encryption, lifecycle management, CORS rules, and access logging.

## What Was Created

### API Definition
- `apis/org/openmcf/provider/alicloud/alicloudstoragebucket/v1/` -- Full proto API (spec, api, stack_input, stack_outputs)
- Registered `AlicloudStorageBucket = 3050` in `CloudResourceKind` enum under the Storage category
- 5 nested messages: `AlicloudStorageBucketEncryption`, `AlicloudStorageBucketLifecycleRule`, `AlicloudStorageBucketLifecycleTransition`, `AlicloudStorageBucketCorsRule`, `AlicloudStorageBucketLogging`

### IaC Modules
- **Pulumi** (Go): Creates alicloud provider and a single `oss.Bucket` resource with conditional versioning, encryption, lifecycle rules, CORS rules, and logging blocks
- **Terraform** (HCL): Single `alicloud_oss_bucket` resource with dynamic blocks for all optional features, matching variables, outputs, and tag merging

### Tests
- Ginkgo/Gomega spec validation tests: 19 specs covering valid inputs (minimal, full config, AES256 encryption, lifecycle with versioning), missing required fields (region, bucket_name), bucket name length limits, invalid enum values (acl, storage_class, redundancy_type, sse_algorithm, transition storage_class), invalid CORS rules, missing logging target, wrong api_version/kind, and missing metadata

### Documentation
- README.md with configuration reference, lifecycle rule fields, output reference, and related components
- examples.md with 5 YAML examples (minimal, production versioned+encrypted, archive lifecycle, CORS browser access, KMS encrypted with logging)
- catalog-page.md with full configuration reference and examples
- docs/README.md with comprehensive research documentation
- 3 presets: private-standard, versioned-encrypted, archive-lifecycle

## Spec Design Decisions

- **`acl` included despite provider deprecation**: The `acl` field is deprecated in provider v1.220.0 in favor of `alicloud_oss_bucket_acl`, but it remains functional and is the most ergonomic way to set bucket access control. Forcing users to create a separate resource for a fundamental property is bad UX. Terraform validates with a deprecation warning, which is acceptable.
- **`redundancy_type` added (not in T02)**: This is an immutable-at-creation-time choice between LRS and ZRS. Omitting it would force all buckets to LRS with no way to get ZRS durability without a re-creation.
- **`versioning_enabled` as bool**: Simpler than the provider's `Enabled`/`Suspended` enum. The `Suspended` state only matters for buckets that previously had versioning -- a runtime concern, not an IaC provisioning decision.
- **Lifecycle rules -- 80/20 cut**: Exposed days-based expiration, transitions, abort multipart upload, and noncurrent version expiration. Excluded advanced features (date-based expiration, access-time-based transitions, filter exclusions) that serve niche use cases.
- **No dependencies**: OSS buckets are standalone resources with zero upstream dependencies -- no VPC, no subnet.

## Verification

- `go build ./...` -- PASS
- `go vet ./...` -- PASS
- `go test ./...` -- PASS (19/19 specs)
- `terraform init` -- PASS (alicloud provider v1.271.0)
- `terraform validate` -- PASS (with expected `acl` deprecation warning)

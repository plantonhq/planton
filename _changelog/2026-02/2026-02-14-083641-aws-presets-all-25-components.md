# AWS Presets: All 25 Components

**Date**: February 14, 2026
**Type**: Feature
**Components**: Presets, AWS Provider

## Summary

Created production-quality presets for all 25 AWS deployment components, delivering 49 presets (98 YAML + markdown files) across foundation infrastructure, networking, compute, orchestration, serverless, data, and IAM. Every AWS component now has at least one ranked, deployable preset with a companion documentation file following the conventions established in the presets system foundation (T01).

## Problem Statement / Motivation

The presets system foundation was established in an earlier session (convention document, Cursor rules, Forge integration, and a single AwsAlb pilot preset). However, only 1 of 25 AWS components had actual presets. Users browsing AWS components had no ready-to-deploy starting points for the remaining 24 components.

### Pain Points

- Platform users still faced the "blank page" problem for 24 AWS resource types
- The presets convention was validated with a single pilot but not yet proven at scale
- No presets existed for critical infrastructure like VPC, EKS, RDS, Lambda, or IAM roles

## Solution / What's New

Created 49 presets across all 25 AWS components, organized into 4 logical batches:

### Batch 1: Foundation and Simple (8 components, 13 presets)
- AwsKmsKey, AwsSecretsManager, AwsEcrRepo, AwsCertManagerCert, AwsRoute53Zone, AwsS3ObjectSet, AwsIamUser, AwsRoute53DnsRecord

### Batch 2: Networking and Compute (6 components, 13 presets)
- AwsVpc, AwsSecurityGroup, AwsEc2Instance, AwsCloudFront, AwsClientVpn, AwsEcsCluster

### Batch 3: Orchestration and Serverless (5 components, 11 presets)
- AwsEksCluster, AwsEksNodeGroup, AwsEcsService, AwsLambda, AwsS3Bucket

### Batch 4: Data and IAM Roles (5 components, 12 presets)
- AwsRdsCluster, AwsRdsInstance, AwsDocumentDb, AwsDynamodb, AwsIamRole

## Implementation Details

Each preset follows the conventions defined in `architecture/presets.md`:
- Complete KRM manifests with `apiVersion`, `kind`, `metadata`, `spec`
- `StringValueOrRef` fields use the proto-correct `value:` wrapper
- Angle-bracket placeholders for user-specific values
- Real values with YAML comments for sensible defaults
- Companion markdown with: Title, Description, When to Use, Key Configuration Choices, Placeholders to Replace, and Related Presets

### Notable Design Decisions

- **Engine-specific database presets**: AwsRdsCluster has separate presets for Aurora PostgreSQL, Aurora MySQL, and Aurora Serverless v2 rather than generic placeholders, because the engine is a required field and configurations differ significantly
- **IAM roles with real policies**: AwsIamRole presets include full `google.protobuf.Struct` trust policies and managed policy ARNs (Lambda, ECS, EC2+SSM) rather than placeholders, providing immediately usable starting points
- **Zone-only Route53Zone presets**: Zone presets do not include inline DNS records, favoring composability with standalone `AwsRoute53DnsRecord` resources
- **AwsSecretsManager minimal preset**: Component only has `secret_names` as its spec; one preset with descriptive example names is sufficient

### Preset Distribution

| Presets | Components |
|---------|-----------|
| 1 preset | AwsKmsKey, AwsSecretsManager, AwsS3ObjectSet |
| 2 presets | AwsCertManagerCert, AwsRoute53Zone, AwsIamUser, AwsRoute53DnsRecord, AwsVpc, AwsEc2Instance, AwsCloudFront, AwsClientVpn, AwsEcsCluster, AwsEksCluster, AwsEksNodeGroup, AwsLambda, AwsS3Bucket, AwsRdsInstance, AwsDocumentDb, AwsDynamodb |
| 3 presets | AwsEcsService, AwsRdsCluster, AwsIamRole, AwsSecurityGroup |

## Benefits

- **24 more components** with ready-to-deploy presets (was 1, now 25)
- **49 total presets** providing the most common real-world configurations
- **98 files** (YAML + markdown pairs) following a consistent convention
- Users can now pick an AWS preset, replace angle-bracket placeholders, and deploy immediately
- Convention validated at scale across components of varying complexity (from 1-field AwsSecretsManager to 30+-field AwsRdsCluster)

## Impact

- All 25 AWS components now have presets, completing T02 of the presets project
- Sets the pattern and confidence level for T03-T08 (remaining 188 components across GCP, Azure, Kubernetes, OpenStack, Scaleway, and others)
- Presets rules (`create-planton-preset`, `audit-planton-presets`, `validate-planton-presets`) are now proven against real content

## Related Work

- **Presets System Foundation** (`2026-02-14-075740-presets-system-foundation.md`) -- Convention document, AI reference, Cursor rules, and Forge integration that this work builds on
- **T03-T08** -- Remaining provider presets (GCP, Azure, Kubernetes, OpenStack, Scaleway, others) planned as subsequent tasks

---

**Status**: Production Ready
**Timeline**: Single session

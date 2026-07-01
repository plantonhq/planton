# Fix Template Validation Errors Across All Non-AliCloud Providers

**Date**: March 31, 2026
**Type**: Fix
**Provider**: Multi-Provider
**Chart(s)**: 24 charts across AWS, Azure, GCP, OCI, DigitalOcean, Civo, Scaleway

## Summary

Resolved `planton chart build` validation failures across all non-AliCloud InfraCharts by correcting field names, enum values, and message structure in templates to match their protobuf spec definitions. Took the pass rate from 25/49 charts to 43/43 (excluding 6 AliCloud charts for separate work).

## Problem Statement / Motivation

After the initial bulk creation of InfraCharts across multiple cloud providers, many chart templates used field names, enum values, or message structures that didn't match the actual protobuf definitions in the planton repo. These charts failed `planton chart build` validation, making them unusable for infrastructure provisioning.

### Pain Points

- 24 of 49 charts failed validation, across 8 providers
- Field names from Terraform/Pulumi conventions were used instead of protobuf field names (e.g., `awsRegion` vs `region`)
- Boolean fields were used where protobuf defines enums (e.g., `isVersioningEnabled: true` vs `versioning: enabled`)
- Enum values were bare words instead of fully-qualified protobuf names (e.g., `public` vs `endpoint_type_public`)
- Non-existent fields like `displayName` or `name` were used where the resource derives its name from `metadata.name`
- Kubernetes addon templates had structural issues: a `group` field in relationships and a `kubernetesClusterSelector` wrapper that don't exist in the proto

## Solution / What's New

Cross-referenced every failing template against its protobuf spec in the planton repo and applied targeted fixes. The work fell into several systemic categories, each addressed in bulk.

### Fix Categories

#### 1. Kubernetes Addon Relationships (54 files, 6 providers)

All Kubernetes environment charts (EKS, AKS, GKE, DOKS, Civo, Kapsule) shared the same addon template structure with two issues:

- **`group` field in relationships**: `CloudResourceRelationship` only has `kind`, `name`, `type`, and `env` — no `group`. Removed from all 54 addon files.
- **`kubernetesClusterSelector` wrapper**: `targetCluster` IS a `KubernetesClusterSelector` directly — it shouldn't contain a nested `kubernetesClusterSelector`. Flattened `clusterKind` and `clusterName` up one level.

Before:

```yaml
spec:
  targetCluster:
    kubernetesClusterSelector:
      clusterKind: AwsEksCluster
      clusterName: "{{ values.cluster_name }}"
```

After:

```yaml
spec:
  targetCluster:
    clusterKind: AwsEksCluster
    clusterName: "{{ values.cluster_name }}"
```

#### 2. AWS S3 Bucket Region (5 charts)

`AwsS3BucketSpec` uses `region`, not `awsRegion`. Fixed in: data-analytics, ml-workbench, pulumi-backend, static-website, terraform-backend.

#### 3. AWS Resource-Specific Fixes (4 charts)

| Chart | Resource | Fix |
|-------|----------|-----|
| container-app | AwsAppRunnerService | Removed `imageConfiguration` block (not in ImageSource proto) |
| event-driven-pipeline | AwsSqsQueue | `queueType` → `fifoQueue` (bool), `deadLetterQueue` → `deadLetterConfig` |
| serverless-api | AwsDynamodb | `hashKey` → `attributeDefinitions` + `keySchema` (proper DynamoDB schema) |
| ml-workbench | AwsSagemakerDomain | Removed `domainName` (derived from metadata) |

#### 4. Azure Fixes (3 charts)

| Chart | Resource | Fix |
|-------|----------|-----|
| container-apps-environment | AzureContainerApp | Flattened `scale` block → `minReplicas`/`maxReplicas` at spec level |
| enterprise-network-foundation | AzureNatGateway | Removed `name` and `publicIpId` (not in spec), added `subnetId` |
| web-app-environment | AzureLinuxWebApp | Replaced `valueFrom` in `appSettings` with literal string (plain `map<string,string>`) |

#### 5. OCI Fixes (2 charts)

| Chart | Resource | Fix |
|-------|----------|-----|
| data-platform | OciObjectStorageBucket | `isVersioningEnabled` → `versioning: enabled`, `isAutoTieringEnabled` → `autoTiering: infrequent_access` |
| data-platform | OciStreamPool | Removed `displayName`, fixed `autoCreateTopicsEnabled` → `autoCreateTopicsEnable` |
| data-platform | OciLogGroup | Removed `displayName` |
| serverless-stack | OciObjectStorageBucket | `isVersioningEnabled` → `versioning: enabled` |
| serverless-stack | OciApiGateway | `endpointType: public` → `endpoint_type_public` |
| serverless-stack | OciLogGroup | Removed `displayName` |

#### 6. GCP Fix (1 chart)

- **serverless-api-backend**: `GcpCloudRun` `container.image` is a `GcpCloudRunContainerImage` message (`repo` + `tag`), not a plain string.

#### 7. DigitalOcean Fixes (2 charts)

- **app-platform-environment**: Registry tier `BASIC` → `basic`
- **doks-environment**: Registry tier `STARTER` → `starter`

#### 8. Scaleway Fix (1 chart)

- **serverless-environment**: Container privacy `public` → `privacy_public`

## Implementation Details

### Code Metrics

- **71 template files changed** across 10 providers
- **54 Kubernetes addon files** systematically fixed via Python regex
- **17 provider-specific template files** fixed individually
- **142 lines added, 251 lines removed** (net simplification)

### Approach

1. Ran `planton chart build` on all 49 charts to identify failures
2. Categorized errors by root cause (field mismatch, enum mismatch, structural mismatch)
3. Cross-referenced each failing resource against its `spec.proto` in planton
4. Applied fixes in bulk for systemic issues (group, kubernetesClusterSelector) and individually for resource-specific issues
5. Iteratively re-ran builds until all 43 non-AliCloud charts passed

## Benefits

- **43/43 non-AliCloud charts pass validation** — ready for deployment
- Charts now faithfully reflect the protobuf contract, preventing runtime surprises
- Kubernetes environment charts across all 6 providers share a consistent, validated addon structure
- DynamoDB, SQS, and S3 templates use correct AWS protobuf conventions

## Impact

All infrastructure teams using InfraCharts for AWS, Azure, GCP, OCI, DigitalOcean, Civo, Scaleway, HetznerCloud, and OpenStack environments can now build and deploy their charts without validation errors.

### Verification

```bash
# Build any chart to verify
planton chart build aws/eks-environment
planton chart build azure/aks-environment
planton chart build gcp/gke-environment
planton chart build oci/data-platform
planton chart build digital-ocean/doks-environment
```

## Known Limitations

- **6 AliCloud charts remain unfixed** — excluded from this pass for separate work. Errors include `enableAutoScaling`, `namespaces` type mismatch, `serviceName`/`functionName`/`displayName` on AliCloud-specific resources, and `healthCheck` on ALB.

## Related Work

- Previous changelog: `2026-03-31-125036-oci-infra-charts-five-environment-charts.md` (original OCI chart creation)
- Previous changelog: `2026-03-31-161443-oci-data-platform-serverless-stack-template-fixes.md` (initial OCI-only fixes, now superseded by this broader fix)

---

**Status**: ✅ Production Ready
**Timeline**: Single session

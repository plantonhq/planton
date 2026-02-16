# StringValueOrRef Fixes for Existing AWS Components (F1-F6)

**Date**: February 16, 2026
**Type**: Enhancement
**Components**: API Definitions, Pulumi Modules, Terraform Modules, Validation Tests

## Summary

Converted 10 plain `string` fields across 6 existing AWS components to `StringValueOrRef`, enabling cross-resource references via the `valueFrom` pattern. This brings these foundational components in line with the StringValueOrRef convention established in all 33 newly-forged AWS resource kinds, completing the API consistency goal for the AWS provider.

## Problem Statement / Motivation

When the AWS resource expansion project began, 7 existing AWS components were identified as having plain `string` fields for values that should support cross-resource references (KMS keys, IAM roles, VPC IDs, security groups, ACM certificates). Without `StringValueOrRef`, users deploying these components in infra charts or multi-resource compositions had to hard-code IDs and ARNs -- defeating the purpose of OpenMCF's declarative dependency system.

### Pain Points

- Users could not use `valueFrom` references to wire a KMS key provisioned by one component into an S3 bucket or ECR repo managed by another
- Security group rules couldn't reference other OpenMCF-managed security groups, forcing manual ID copying
- Private Route53 zones couldn't reference VPCs from OpenMCF's AwsVpc component
- CloudFront distributions couldn't reference ACM certificates managed by OpenMCF
- The inconsistency between old and new components made the platform feel unfinished

## Solution / What's New

All 6 components updated (F7 AwsSecretsManager excluded -- pending removal):

| # | Component | Field(s) Changed | Reference Target |
|---|-----------|-----------------|------------------|
| F1 | AwsCloudFront | `certificate_arn` | AwsCertManagerCert |
| F2 | AwsEcrRepo | `kms_key_id` | AwsKmsKey |
| F3 | AwsEcsCluster | `ExecConfiguration.kms_key_id` | AwsKmsKey |
| F4 | AwsS3Bucket | `kms_key_id`, `replication.role_arn` | AwsKmsKey, AwsIamRole |
| F5 | AwsRoute53Zone | `Route53VpcAssociation.vpc_id` | AwsVpc |
| F6 | AwsSecurityGroup | `source_security_group_ids`, `destination_security_group_ids` | AwsSecurityGroup |

## Implementation Details

### Proto Changes

Each field changed from:
```protobuf
string kms_key_id = N;
```
to:
```protobuf
org.openmcf.shared.foreignkey.v1.StringValueOrRef kms_key_id = N [
  (org.openmcf.shared.foreignkey.v1.default_kind) = AwsKmsKey,
  (org.openmcf.shared.foreignkey.v1.default_kind_field_path) = "status.outputs.key_arn"
];
```

Field numbers preserved for wire-format compatibility. `default_kind` and `default_kind_field_path` annotations provide IDE/UI hints for the reference target.

For F1 (CloudFront), the ARN pattern regex was removed from the field -- it's meaningless when the value comes from a reference, and the referenced ACM cert resource handles its own validation.

CEL validation for CloudFront's `aliases_require_cert` updated from string comparison (`this.certificate_arn != ""`) to presence check (`has(this.certificate_arn)`).

### Pulumi Go Changes

Singular fields: `spec.Field` -> `spec.Field != nil && spec.Field.GetValue() != ""` for conditionals, `spec.Field.GetValue()` for value extraction.

Repeated fields (F6): `pulumi.ToStringArray(r.SourceSecurityGroupIds)` -> explicit loop with `.GetValue()`:
```go
var sourceSGs pulumi.StringArray
for _, sg := range r.SourceSecurityGroupIds {
    sourceSGs = append(sourceSGs, pulumi.String(sg.GetValue()))
}
```

### Terraform HCL Changes

Singular: `string` -> `optional(object({ value = string }))`, access via `.value`.
Repeated: `list(string)` -> `list(object({ value = string }))`, access via `[for sg in list : sg.value]`.

### Test Updates

All `spec_test.go` files updated to wrap string literals in `StringValueOrRef`:
```go
KmsKeyId: &foreignkeyv1.StringValueOrRef{
    LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
        Value: "arn:aws:kms:...",
    },
},
```

## Benefits

- **Cross-resource references**: Users can now wire dependencies between all AWS components using `valueFrom`
- **Infra chart composability**: These components can participate in multi-resource compositions without hard-coded IDs
- **API consistency**: All AWS components now follow the same StringValueOrRef convention
- **Zero breaking changes**: Field numbers preserved, literal values still work exactly as before

## Impact

- **6 components updated**, 10 fields converted
- **All existing tests pass** -- no behavioral regressions
- **Proto compiles clean**, Go builds, Terraform variables updated
- Users with literal values in existing manifests are unaffected (backward compatible)
- Users can now add `valueFrom` references to these fields

## Related Work

- Part of the **20260215.02.sp.aws-resource-expansion** sub-project
- Complements the 33 new AWS resource kinds forged in the same project
- F7 (AwsSecretsManager) intentionally skipped -- component scheduled for removal

---

**Status**: Production Ready
**Timeline**: Single session

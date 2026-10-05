# AwsConfigConformancePack

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsConfigConformancePackSpec defines the desired configuration for
an AWS Config conformance pack - a template bundle that deploys a
set of Config rules (and optional remediations) as one unit.

A pack references no standalone Config rules: the template CREATES
its own rules, prefixed with the pack name and managed by the pack
service-linked role. Deploying a pack DOES require a Config
recorder to already be running in the region (AWS rejects the
deployment without one) - pair it with an AwsConfigRecorder.

The pack deploys at one of two scopes: this account only (the
default), or every account in the AWS Organization
(organization_scope = true, run from the management account or a
delegated Config administrator). The pack's name is metadata.name
(letters, digits, hyphens, starting with a letter; AWS caps account
packs at 256 characters and ORGANIZATION packs at 128).

Destroying this component deletes the pack and every rule it
created (a real delete; organization packs unwind from all member
accounts, which can take several minutes).

## Example

```yaml
# Canonical AwsConfigConformancePack example (hack/dev manifest and
# refgen Example source): an account-scoped pack deploying two managed
# S3 rules from an inline template, one parameterized. Literals stand
# in so the offline `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsConfigConformancePack
metadata:
  name: s3-best-practices
  id: s3-best-practices
  org: test-org
  env: dev
spec:
  region: us-west-2
  inputParameters:
    - parameterName: S3BucketPublicReadProhibitedParamCheckPeriod
      parameterValue: TwentyFour_Hours
  templateBody: |
    Parameters:
      S3BucketPublicReadProhibitedParamCheckPeriod:
        Type: String
    Resources:
      S3BucketPublicReadProhibited:
        Type: AWS::Config::ConfigRule
        Properties:
          ConfigRuleName: s3-bucket-public-read-prohibited
          Source:
            Owner: AWS
            SourceIdentifier: S3_BUCKET_PUBLIC_READ_PROHIBITED
          MaximumExecutionFrequency:
            Ref: S3BucketPublicReadProhibitedParamCheckPeriod
      S3BucketVersioningEnabled:
        Type: AWS::Config::ConfigRule
        Properties:
          ConfigRuleName: s3-bucket-versioning-enabled
          Source:
            Owner: AWS
            SourceIdentifier: S3_BUCKET_VERSIONING_ENABLED
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.organizationScope` | `bool` |  |  |  |
| `spec.deliveryS3Bucket` | `string \| valueFrom` |  |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.deliveryS3KeyPrefix` | `string` |  |  |  |
| `spec.templateBody` | `string` |  |  |  |
| `spec.templateS3Uri` | `string` |  |  |  |
| `spec.inputParameters` | `[]AwsConfigConformancePackInputParameter` |  |  |  |
| `spec.inputParameters[].parameterName` | `string` | yes |  |  |
| `spec.inputParameters[].parameterValue` | `string` | yes |  |  |
| `spec.excludedAccounts` | `[]string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the pack deploys in.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.organizationScope

`bool`

Deploy to every account in the AWS Organization instead of just
this one. Requires the management account or a delegated Config
administrator, with the Config service enabled for the
organization.

### spec.deliveryS3Bucket

`string | valueFrom`

S3 bucket where AWS Config stores the pack's compliance
evaluation results. Unset = AWS keeps results queryable through
the Config APIs only. ORGANIZATION packs require the bucket name
to start with "awsconfigconforms" (an AWS naming contract the
deploy enforces).

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.deliveryS3KeyPrefix

`string`

Key prefix for the delivered results inside the bucket.

- rule: {"string":{"maxLen":"1024"}}

### spec.templateBody

`string`

The pack template, inline (YAML or JSON, the CloudFormation-like
conformance pack schema). Mutually exclusive with
template_s3_uri on organization packs; account packs accept both
set at once, in which case AWS prefers the S3 template. AWS never
reports the template back, so imports re-assert it on the first
apply.

- rule: {"string":{"maxLen":"51200"}}

### spec.templateS3Uri

`string`

The pack template as an S3 object ("s3://bucket/key"). The
deploying principal must be able to read it.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"1024","pattern":"^s3://"}}

### spec.inputParameters

`[]AwsConfigConformancePackInputParameter`

Values for the template's parameters (each template declares its
own). At most 60.

- rule: {"repeated":{"maxItems":"60"}}

### spec.inputParameters[].parameterName

`string` · required

The parameter name, as the template declares it.

- rule: {"required":true}

### spec.inputParameters[].parameterValue

`string` · required

The parameter value.

- rule: {"required":true}

### spec.excludedAccounts

`[]string`

Organization-scope only: member accounts to SKIP when deploying
the pack. At most 1000.

- rule: {"repeated":{"maxItems":"1000","unique":true,"items":{"string":{"pattern":"^[0-9]{12}$"}}}}

## Validation Rules

- `template_required`: set template_body or template_s3_uri
- `org_template_exactly_one`: organization packs accept template_body or template_s3_uri, not both
- `excluded_accounts_org_only`: excluded_accounts requires organization_scope = true

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsConfigConformancePack, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.pack_name` | `string` | The pack's name (also the provider's import ID, at either scope). |
| `status.outputs.pack_arn` | `string` | The pack's ARN (the account-scope or organization-scope ARN, whichever this instance deployed). |
| `status.outputs.region` | `string` | The AWS region the pack lives in. Conformance packs are region-scoped and addressed by region + name, so any consumer (or verifier) reaching the pack off the ambient region needs this alongside pack_name. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.deliveryS3Bucket` | AwsS3Bucket | `status.outputs.bucket_id` |

## See Also

- [Overview](../README.md)

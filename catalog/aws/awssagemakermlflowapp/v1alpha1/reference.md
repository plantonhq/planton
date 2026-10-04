# AwsSagemakerMlflowApp

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsSagemakerMlflowAppSpec defines the desired configuration for an
Amazon SageMaker AI MLflow app - the SERVERLESS managed MLflow
deployment (MLflow 3.x) that succeeds the hourly-billed tracking
server: no capacity to size, billed per use. The app's AWS name
derives from metadata.name.

The app is standalone - it does NOT attach to a tracking server. It
associates with SageMaker domains (`default_domain_ids`) so Studio
users in those domains get it as their default MLflow, and it can be
the account-wide default (`account_default_status`). Everything
updates in place except the role.

## Example

```yaml
# Canonical AwsSagemakerMlflowApp example (hack/dev manifest and refgen
# Example source): an account-default serverless MLflow app exercising
# every arm - domain associations, auto-registration, and the
# maintenance window. Literal ARNs and domain IDs stand in for composed
# references so the offline `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSagemakerMlflowApp
metadata:
  name: ml-experiments
  id: ml-experiments
  org: test-org
  env: dev
spec:
  region: us-west-2
  artifactStoreUri: s3://my-mlflow/artifacts
  roleArn:
    value: arn:aws:iam::123456789012:role/mlflow-app
  accountDefaultStatus: ENABLED
  defaultDomainIds:
    - value: d-abcdef123456
  modelRegistrationMode: AutoModelRegistrationEnabled
  weeklyMaintenanceWindowStart: Sun:03:00
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.artifactStoreUri` | `string` | yes |  |  |
| `spec.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.accountDefaultStatus` | `string` |  |  |  |
| `spec.defaultDomainIds` | `[]string \| valueFrom` |  |  | AwsSagemakerDomain (`status.outputs.domain_id`) |
| `spec.modelRegistrationMode` | `string` |  |  |  |
| `spec.weeklyMaintenanceWindowStart` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the MLflow app will be created.
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.artifactStoreUri

`string` · required

S3 location where MLflow artifacts (model files, run outputs) are
stored. Example: "s3://my-mlflow/artifacts". Updates in place.

- rule: {"string":{"minLen":"1","pattern":"^(https|s3)://([^/]+)/?(.*)$"}}

### spec.roleArn

`string | valueFrom` · required

IAM role the app uses to read and write the artifact store. Must
trust sagemaker.amazonaws.com with S3 access to the artifact
bucket. The ONE replace-on-change field - everything else updates
in place.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.accountDefaultStatus

`string`

Make this app the account's default MLflow: "ENABLED" or
"DISABLED". Omitted = AWS default (not the account default). Only
one app per account can be the default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ENABLED","DISABLED"]}}

### spec.defaultDomainIds

`[]string | valueFrom`

SageMaker domains for which this app is the default MLflow (Studio
users in these domains track to it automatically).

- references: AwsSagemakerDomain (`status.outputs.domain_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSagemakerDomain, name: <that resource's name>, fieldPath: status.outputs.domain_id}} -- a bare string does not parse

### spec.modelRegistrationMode

`string`

Auto-register models logged to MLflow into the SageMaker Model
Registry: "AutoModelRegistrationEnabled" or
"AutoModelRegistrationDisabled". Omitted = AWS default (disabled).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["AutoModelRegistrationEnabled","AutoModelRegistrationDisabled"]}}

### spec.weeklyMaintenanceWindowStart

`string`

Weekly maintenance window start, UTC 24-hour "Ddd:HH:MM" (e.g.
"Sun:03:00"). The day token is MIXED-CASE by AWS's server-side
contract - the MLflow APIs reject "SUN:03:00" with a
ValidationException naming the exact regex
(Mon|Tue|Wed|Thu|Fri|Sat|Sun), live-caught 2026-08-25 on the
sibling tracking-server kind. Omitted = AWS picks.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(Mon|Tue|Wed|Thu|Fri|Sat|Sun):([01]\\d|2[0-3]):[0-5]\\d$"}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsSagemakerMlflowApp, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.app_arn` | `string` | The Amazon Resource Name of the MLflow app (the AWS identity). |
| `status.outputs.app_name` | `string` | The app name. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.defaultDomainIds` | AwsSagemakerDomain | `status.outputs.domain_id` |

## See Also

- [Overview](../README.md)

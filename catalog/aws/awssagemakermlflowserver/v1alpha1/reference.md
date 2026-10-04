# AwsSagemakerMlflowServer

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsSagemakerMlflowServerSpec defines the desired configuration for an
Amazon SageMaker AI MLflow tracking server - the classic managed
MLflow deployment (experiments, runs, model tracking) billed per
hour while it runs. The server's AWS name derives from metadata.name.

Operational facts worth planning around: creation takes ~25 minutes
and deletion is similar (AWS provisions dedicated capacity), and the
server bills hourly from Created onward, whether or not anyone is
tracking - the size ladder scales the hourly rate. For
the serverless successor product, see AwsSagemakerMlflowApp.

## Example

```yaml
# Canonical AwsSagemakerMlflowServer example (hack/dev manifest and
# refgen Example source): a Medium server exercising every arm -
# version pin, auto-registration, and the maintenance window. Literal
# ARNs stand in for composed references so the offline `tofu plan`
# renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSagemakerMlflowServer
metadata:
  name: ml-experiments
  id: ml-experiments
  org: test-org
  env: dev
spec:
  region: us-west-2
  artifactStoreUri: s3://my-mlflow/artifacts
  roleArn:
    value: arn:aws:iam::123456789012:role/mlflow-server
  size: Medium
  mlflowVersion: "3.0"
  automaticModelRegistration: true
  weeklyMaintenanceWindowStart: Tue:03:30
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.artifactStoreUri` | `string` | yes |  |  |
| `spec.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.size` | `string` |  |  |  |
| `spec.mlflowVersion` | `string` |  |  |  |
| `spec.automaticModelRegistration` | `bool` |  |  |  |
| `spec.weeklyMaintenanceWindowStart` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the tracking server will be created.
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.artifactStoreUri

`string` · required

S3 location where MLflow artifacts (model files, run outputs) are
stored. Example: "s3://my-mlflow/artifacts". Updates in place.

- rule: {"string":{"minLen":"1","maxLen":"1024","pattern":"^(https|s3)://([^/]+)/?(.*)$"}}

### spec.roleArn

`string | valueFrom` · required

IAM role the server uses to read and write the artifact store. Must
trust sagemaker.amazonaws.com with S3 access to the artifact
bucket. Changing it replaces the server.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.size

`string`

Server size: "Small" (AWS default - up to ~25 users), "Medium"
(~50), or "Large" (~100). Resizes in place (a maintenance-window
style operation).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Small","Medium","Large"]}}

### spec.mlflowVersion

`string`

MLflow version as "major.minor" (e.g. "3.0") - AWS normalizes away
the patch, so a patch-level value here would drift forever. Omitted
= AWS picks the latest. Changing it replaces the server.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^\\d+\\.\\d+$"}}

### spec.automaticModelRegistration

`bool`

Auto-register models logged to MLflow into the SageMaker Model
Registry. UPSTREAM TRAP: the provider cannot turn this back OFF
(a true-to-false change is silently not transmitted) - disabling
requires replacing the server or an out-of-band API call.

### spec.weeklyMaintenanceWindowStart

`string`

Weekly maintenance window start, UTC 24-hour "Ddd:HH:MM" (e.g.
"Tue:03:30"). The day token is MIXED-CASE by AWS's server-side
contract - CreateMlflowTrackingServer rejects "TUE:03:30" with a
ValidationException naming the exact regex
(Mon|Tue|Wed|Thu|Fri|Sat|Sun), live-caught 2026-08-25. Omitted =
AWS picks.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(Mon|Tue|Wed|Thu|Fri|Sat|Sun):([01]\\d|2[0-3]):[0-5]\\d$"}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsSagemakerMlflowServer, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.tracking_server_name` | `string` | The tracking server name (the AWS identity). |
| `status.outputs.tracking_server_arn` | `string` | The Amazon Resource Name of the tracking server. |
| `status.outputs.tracking_server_url` | `string` | URL of the MLflow UI served by this tracking server. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.roleArn` | AwsIamRole | `status.outputs.role_arn` |

## See Also

- [Overview](../README.md)

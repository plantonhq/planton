# AwsApiGatewayAccountSettings

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsApiGatewayAccountSettingsSpec defines the desired API Gateway
account-level settings for one AWS region.

This is a SETTINGS SINGLETON: the resource's identity is the
account+region pair, not a name - AWS keeps exactly one API Gateway
account object per region, and this component manages it. Deploy at
most one instance per region; two instances targeting the same
region fight over the same settings object. metadata.name never
reaches AWS - it is Planton-side identity only.

The one configurable lever is the CloudWatch Logs role: API Gateway
uses it to push execution and access logs for every REST API stage
in the region. Stage-level logging on AwsRestApiGateway silently
does nothing until this role is set - that cross-component contract
is why this kind exists.

Destroying this component RESETS the role to none (the region
reverts to no API Gateway logging); the account object itself
always exists and cannot be deleted.

## Example

```yaml
# Canonical AwsApiGatewayAccountSettings example (hack/dev manifest and
# refgen Example source): the region's API Gateway account object with
# a CloudWatch logging role set. Literal ARN stands in for the composed
# AwsIamRole reference so the offline `tofu plan` renders the resource.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsApiGatewayAccountSettings
metadata:
  name: apigw-account-us-west-2
  id: apigw-account-us-west-2
  org: test-org
  env: dev
spec:
  region: us-west-2
  cloudwatchRoleArn:
    value: arn:aws:iam::123456789012:role/apigw-cloudwatch-logs
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.cloudwatchRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |

## Field Details

### spec.region

`string` · required

The AWS region whose API Gateway account settings this instance
manages. The region IS the resource identity - one instance per
region.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.cloudwatchRoleArn

`string | valueFrom`

The IAM role API Gateway assumes to write execution/access logs
to CloudWatch for every REST API in the region. The role must
trust "apigateway.amazonaws.com" and carry CloudWatch Logs write
permissions (the AWS-managed policy
AmazonAPIGatewayPushToCloudWatchLogs is the canonical grant) -
AWS validates both at apply and rejects a role it cannot use
("The role ARN does not have required permissions"). IAM
propagation can lag role creation by a few seconds; both engines'
providers retry through that window.

Leave unset to manage the region WITHOUT a logging role (the
explicit "no API Gateway logging" posture - applying an unset
value clears any role previously set by anyone).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsApiGatewayAccountSettings, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.account_id` | `string` | The 12-digit AWS account ID the settings belong to (also the provider's import ID for this singleton). |
| `status.outputs.api_key_version` | `string` | The API key version the account is on (AWS-managed). |
| `status.outputs.features` | `[]string` | Feature flags AWS reports enabled on the account (for example "UsagePlans"). Informational - AWS manages the set. |
| `status.outputs.throttle_burst_limit` | `int32` | The account-level throttle ceiling AWS reports: maximum burst of concurrent requests any stage may serve before 429s. |
| `status.outputs.throttle_rate_limit` | `double` | The account-level steady-state request rate ceiling (requests per second) AWS reports. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.cloudwatchRoleArn` | AwsIamRole | `status.outputs.role_arn` |

## See Also

- [Overview](../README.md)

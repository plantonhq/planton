# AwsCloudwatchLogResourcePolicy

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsCloudwatchLogResourcePolicySpec defines one CloudWatch Logs
resource policy: the IAM policy document that grants AWS services
permission to write logs - Route53 query logging, EventBridge,
Elasticsearch slow logs, and every other service that asks for a
"CloudWatch Logs resource policy" in its setup docs.

The policy has exactly one scope: ACCOUNT (a named policy applying
account-wide in the region - the common shape) or RESOURCE (a policy
pinned to one log group's ARN). AWS guards updates with a revision
ID; the modules pass the tracked revision so concurrent edits fail
loudly instead of overwriting each other.

## Example

```yaml
# Canonical AwsCloudwatchLogResourcePolicy example (hack/dev manifest
# and refgen Example source): the account-scope grant letting Route53
# write query logs.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudwatchLogResourcePolicy
metadata:
  name: route53-query-logging
  id: route53-query-logging
  org: test-org
  env: dev
spec:
  region: us-east-1
  policyName: route53-query-logging
  policyDocument:
    Version: "2012-10-17"
    Statement:
      - Effect: Allow
        Principal:
          Service: route53.amazonaws.com
        Action:
          - logs:CreateLogStream
          - logs:PutLogEvents
        Resource: arn:aws:logs:us-east-1:123456789012:log-group:/aws/route53/*
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.policyName` | `string` |  |  |  |
| `spec.resourceArn` | `string \| valueFrom` |  |  | AwsCloudwatchLogGroup (`status.outputs.log_group_arn`) |
| `spec.policyDocument` | `object` | yes |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the policy applies in. Example: "us-east-1".

- rule: {"string":{"minLen":"1"}}

### spec.policyName

`string`

The account-scope policy's name (up to 255 characters, no colons
or asterisks). Identity for the account scope - changing it
replaces the policy. Mutually exclusive with resource_arn.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"255","pattern":"^[^:*]*$"}}

### spec.resourceArn

`string | valueFrom`

The resource-scope target: one log group's ARN. Reference an
AwsCloudwatchLogGroup log_group_arn output or pass a literal ARN.
Identity for the resource scope - changing it replaces the policy.
Mutually exclusive with policy_name.

- references: AwsCloudwatchLogGroup (`status.outputs.log_group_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCloudwatchLogGroup, name: <that resource's name>, fieldPath: status.outputs.log_group_arn}} -- a bare string does not parse

### spec.policyDocument

`object` · required

The IAM policy document granting services log-write permission.
Statements typically allow a service principal
(route53.amazonaws.com, events.amazonaws.com, ...) the
logs:PutLogEvents and logs:CreateLogStream actions on the target
log groups. The engines diff it semantically, so formatting never
causes drift.

- rule: {"required":true}

## Validation Rules

- `spec.exactly_one_scope`: set exactly one of policy_name (account scope) and resource_arn (resource scope)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsCloudwatchLogResourcePolicy, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.policy_id` | `string` | The policy's identity: the policy name (account scope) or the target resource ARN (resource scope) - also the provider's import ID (an ARN-shaped import ID selects the resource scope). |
| `status.outputs.policy_scope` | `string` | The scope AWS recorded: ACCOUNT or RESOURCE. |
| `status.outputs.revision_id` | `string` | AWS's revision ID after the last apply - the optimistic-concurrency token guarding the next update. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceArn` | AwsCloudwatchLogGroup | `status.outputs.log_group_arn` |

## See Also

- [Overview](../README.md)

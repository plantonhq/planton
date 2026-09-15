# AwsConfigRule

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

AwsConfigRuleSpec defines the desired configuration for one AWS
Config rule - a compliance check evaluated against recorded
resource configurations - plus its optional auto-remediation.

The rule name is metadata.name (AWS caps account-scoped rule names
at 128 characters; ORGANIZATION-scoped rule names at 64).

One of three rule sources must be chosen:
  - managed: an AWS-authored rule referenced by identifier (e.g.
    "S3_BUCKET_VERSIONING_ENABLED") - the zero-code path.
  - custom_lambda: your Lambda function evaluates resources.
  - custom_policy: a CloudFormation-Guard policy evaluates
    resources - custom logic without a function.

Setting `organization` deploys the rule across EVERY account in the
AWS Organization instead of just this one - run it from the
management account or the Config delegated administrator. The
region needs a running configuration recorder (AwsConfigRecorder)
or every evaluation reports nothing.

Destroying this component deletes the rule (and its remediation
configuration); recorded evaluation history ages out with AWS
Config retention.

## Example

```yaml
# Canonical AwsConfigRule example (hack/dev manifest and refgen Example
# source): a managed rule scoped to S3 buckets with non-automatic SSM
# remediation. Literal values stand in for composed references so the
# offline `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsConfigRule
metadata:
  name: s3-bucket-versioning-enabled
  id: s3-bucket-versioning-enabled
  org: test-org
  env: dev
spec:
  region: us-west-2
  description: Every S3 bucket must have versioning enabled
  managed:
    ruleIdentifier: S3_BUCKET_VERSIONING_ENABLED
  scope:
    complianceResourceTypes:
      - AWS::S3::Bucket
  evaluationModes:
    - DETECTIVE
  remediation:
    automatic: false
    targetId: AWS-ConfigureS3BucketVersioning
    resourceType: AWS::S3::Bucket
    parameters:
      - name: BucketName
        resourceValue: RESOURCE_ID
      - name: AutomationAssumeRole
        staticValue: arn:aws:iam::123456789012:role/config-remediation
    concurrentExecutionRatePercentage: 10
    errorPercentage: 50
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.inputParameters` | `string` |  |  |  |
| `spec.maximumExecutionFrequency` | `string` |  |  |  |
| `spec.managed` | `AwsConfigRuleManagedSource` |  |  |  |
| `spec.managed.ruleIdentifier` | `string` | yes |  |  |
| `spec.customLambda` | `AwsConfigRuleCustomLambdaSource` |  |  |  |
| `spec.customLambda.functionArn` | `string \| valueFrom` | yes |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.customLambda.sourceDetails` | `[]AwsConfigRuleSourceDetail` |  |  |  |
| `spec.customLambda.sourceDetails[].messageType` | `string` |  |  |  |
| `spec.customLambda.sourceDetails[].maximumExecutionFrequency` | `string` |  |  |  |
| `spec.customPolicy` | `AwsConfigRuleCustomPolicySource` |  |  |  |
| `spec.customPolicy.policyRuntime` | `string` |  |  |  |
| `spec.customPolicy.policyText` | `string` | yes |  |  |
| `spec.customPolicy.enableDebugLogDelivery` | `bool` |  |  |  |
| `spec.scope` | `AwsConfigRuleScope` |  |  |  |
| `spec.scope.complianceResourceId` | `string` |  |  |  |
| `spec.scope.complianceResourceTypes` | `[]string` |  |  |  |
| `spec.scope.tagKey` | `string` |  |  |  |
| `spec.scope.tagValue` | `string` |  |  |  |
| `spec.evaluationModes` | `[]string` |  |  |  |
| `spec.organization` | `AwsConfigRuleOrganization` |  |  |  |
| `spec.organization.excludedAccounts` | `[]string` |  |  |  |
| `spec.organization.triggerTypes` | `[]string` |  |  |  |
| `spec.organization.debugLogDeliveryAccounts` | `[]string` |  |  |  |
| `spec.organization.enabled` | `bool` |  | `true` |  |
| `spec.remediation` | `AwsConfigRuleRemediation` |  |  |  |
| `spec.remediation.automatic` | `bool` |  |  |  |
| `spec.remediation.targetId` | `string` | yes |  |  |
| `spec.remediation.targetVersion` | `string` |  |  |  |
| `spec.remediation.resourceType` | `string` |  |  |  |
| `spec.remediation.parameters` | `[]AwsConfigRuleRemediationParameter` |  |  |  |
| `spec.remediation.parameters[].name` | `string` | yes |  |  |
| `spec.remediation.parameters[].resourceValue` | `string` |  |  |  |
| `spec.remediation.parameters[].staticValue` | `string` |  |  |  |
| `spec.remediation.parameters[].staticValues` | `[]string` |  |  |  |
| `spec.remediation.maximumAutomaticAttempts` | `int32` |  |  |  |
| `spec.remediation.retryAttemptSeconds` | `int64` |  |  |  |
| `spec.remediation.concurrentExecutionRatePercentage` | `int32` |  |  |  |
| `spec.remediation.errorPercentage` | `int32` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the rule is created in.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

What this rule checks (shown in the AWS console).

- rule: {"string":{"maxLen":"256"}}

### spec.inputParameters

`string`

Rule parameters as a JSON object string (e.g.
'{"desiredInstanceType":"t3.micro"}' for DESIRED_INSTANCE_TYPE).
AWS validates the JSON and each key against the rule's schema at
apply.

- rule: {"string":{"maxLen":"2048"}}

### spec.maximumExecutionFrequency

`string`

How often a PERIODIC rule evaluates (change-triggered rules
ignore it). Unset = the rule's default (TwentyFour_Hours).

- rule: {"string":{"in":["","One_Hour","Three_Hours","Six_Hours","Twelve_Hours","TwentyFour_Hours"]}}

### spec.managed

`AwsConfigRuleManagedSource`

An AWS-managed rule, by identifier. The zero-code path - AWS
maintains the evaluation logic.

### spec.managed.ruleIdentifier

`string` · required

The managed rule identifier (e.g.
"S3_BUCKET_VERSIONING_ENABLED", "REQUIRED_TAGS",
"DESIRED_INSTANCE_TYPE") - the full list lives in the AWS Config
managed-rules reference.

- rule: {"string":{"minLen":"1","maxLen":"256"}}

### spec.customLambda

`AwsConfigRuleCustomLambdaSource`

A custom rule backed by your Lambda function.

### spec.customLambda.functionArn

`string | valueFrom` · required

The Lambda function that evaluates resources.

- references: AwsLambda (`status.outputs.function_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.customLambda.sourceDetails

`[]AwsConfigRuleSourceDetail`

When the function is invoked (account-scoped rules; organization
rules declare triggers on the organization message instead). Unset
= AWS derives from the function's evaluation logic at apply.

- rule: {"repeated":{"maxItems":"25"}}

### spec.customLambda.sourceDetails[].messageType

`string`

The notification that triggers evaluation.

- rule: {"string":{"in":["ConfigurationItemChangeNotification","OversizedConfigurationItemChangeNotification","ScheduledNotification","ConfigurationSnapshotDeliveryCompleted"]}}

### spec.customLambda.sourceDetails[].maximumExecutionFrequency

`string`

For ScheduledNotification triggers: how often. Unset =
TwentyFour_Hours.

- rule: {"string":{"in":["","One_Hour","Three_Hours","Six_Hours","Twelve_Hours","TwentyFour_Hours"]}}

### spec.customPolicy

`AwsConfigRuleCustomPolicySource`

A custom rule backed by a CloudFormation-Guard policy.

### spec.customPolicy.policyRuntime

`string`

The Guard runtime version. Only "guard-2.x.x" exists today.

- rule: {"string":{"pattern":"^guard\\-2\\.x\\.x$"}}

### spec.customPolicy.policyText

`string` · required

The Guard policy source (rules written in the Guard DSL).

- rule: {"string":{"minLen":"1","maxLen":"10000"}}

### spec.customPolicy.enableDebugLogDelivery

`bool`

Deliver Guard debug logs to CloudWatch (account-scoped rules; for
organization rules use organization.debug_log_delivery_accounts).

### spec.scope

`AwsConfigRuleScope`

Which resources the rule evaluates. Unset = every recorded
resource the rule's logic applies to.

- rule: compliance_resource_id requires exactly one compliance_resource_types entry
- rule: tag_value requires tag_key

### spec.scope.complianceResourceId

`string`

Evaluate only this one resource (requires exactly one entry in
compliance_resource_types naming its type).

- rule: {"string":{"maxLen":"768"}}

### spec.scope.complianceResourceTypes

`[]string`

Evaluate only these resource types (e.g. "AWS::EC2::Instance").

- rule: {"repeated":{"maxItems":"100","unique":true}}

### spec.scope.tagKey

`string`

Evaluate only resources carrying this tag key (optionally with
tag_value).

- rule: {"string":{"maxLen":"128"}}

### spec.scope.tagValue

`string`

The tag value paired with tag_key.

- rule: {"string":{"maxLen":"256"}}

### spec.evaluationModes

`[]string`

Evaluation modes: DETECTIVE (after deployment, the default) and/or
PROACTIVE (before provisioning, via the Config proactive APIs).
Account-scoped rules only - AWS has no proactive organization
rules.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["DETECTIVE","PROACTIVE"]}}}}

### spec.organization

`AwsConfigRuleOrganization`

Deploy the rule organization-wide. Declaring this message (with its
switch on, the default) makes it an ORGANIZATION rule (name cap drops
to 64 characters; run from the management or delegated-admin account);
`organization.enabled: false` keeps the organization settings in the
manifest while the rule deploys account-scoped.

### spec.organization.excludedAccounts

`[]string`

Member accounts to EXCLUDE from the rule.

- rule: {"repeated":{"maxItems":"1000","unique":true,"items":{"string":{"pattern":"^[0-9]{12}$"}}}}

### spec.organization.triggerTypes

`[]string`

When custom rules evaluate (required for custom_lambda /
custom_policy organization rules; managed rules derive their own).

- rule: {"repeated":{"maxItems":"3","unique":true,"items":{"string":{"in":["ConfigurationItemChangeNotification","OversizedConfigurationItemChangeNotification","ScheduledNotification"]}}}}

### spec.organization.debugLogDeliveryAccounts

`[]string`

Accounts allowed to receive Guard debug logs (custom_policy rules
only).

- rule: {"repeated":{"maxItems":"1000","unique":true,"items":{"string":{"pattern":"^[0-9]{12}$"}}}}

### spec.organization.enabled

`bool` · optional (explicit presence)

Whether the rule deploys organization-wide. Unset means yes: declaring
the block has always meant an organization rule, and this switch lets
a manifest say the opposite out loud.

- default: `true`

### spec.remediation

`AwsConfigRuleRemediation`

Auto-remediation: the SSM document AWS Config runs against
non-compliant resources. Account-scoped rules only.

- rule: automatic remediation requires maximum_automatic_attempts and retry_attempt_seconds

### spec.remediation.automatic

`bool`

Remediate automatically when a resource turns non-compliant.
False = remediation is available in the console but a human
triggers it.

### spec.remediation.targetId

`string` · required

The SSM document to run (e.g. "AWS-DisableS3BucketPublicReadWrite"
or your own document's name). The only target type AWS supports is
SSM_DOCUMENT - both engines send it implicitly.

- rule: {"string":{"minLen":"1","maxLen":"256"}}

### spec.remediation.targetVersion

`string`

The document version to run. Unset = the document's default
version.

### spec.remediation.resourceType

`string`

The resource type this remediation applies to (e.g.
"AWS::S3::Bucket") - required by AWS when the rule covers multiple
types.

### spec.remediation.parameters

`[]AwsConfigRuleRemediationParameter`

Parameters passed to the SSM document.

- rule: {"repeated":{"maxItems":"25"}}
- rule: set exactly one of resource_value, static_value, static_values

### spec.remediation.parameters[].name

`string` · required

The SSM document parameter name.

- rule: {"string":{"minLen":"1"}}

### spec.remediation.parameters[].resourceValue

`string`

Pass the non-compliant resource's ID. The only value AWS accepts
today is "RESOURCE_ID".

- rule: {"string":{"in":["","RESOURCE_ID"]}}

### spec.remediation.parameters[].staticValue

`string`

One fixed value.

### spec.remediation.parameters[].staticValues

`[]string`

Multiple fixed values (for list-typed document parameters).

### spec.remediation.maximumAutomaticAttempts

`int32`

How many times AWS retries an automatic remediation within
retry_attempt_seconds before marking it failed (1-25). Required
when automatic.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":25,"gte":1}}

### spec.remediation.retryAttemptSeconds

`int64`

The retry window in seconds (1-2678000). Required when automatic.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int64":{"lte":"2678000","gte":"1"}}

### spec.remediation.concurrentExecutionRatePercentage

`int32`

Throttle: what percentage of non-compliant resources remediate
concurrently (1-100). Unset = the AWS default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":100,"gte":1}}

### spec.remediation.errorPercentage

`int32`

Circuit breaker: stop remediating when this percentage of
executions fail (1-100). Unset = the AWS default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":100,"gte":1}}

## Validation Rules

- `exactly_one_source`: set exactly one of managed, custom_lambda, custom_policy
- `org_trigger_types_by_source`: organization.trigger_types is required for custom_lambda/custom_policy organization rules and must be empty for managed ones
- `org_custom_policy_no_scheduled_trigger`: organization custom_policy rules do not accept the ScheduledNotification trigger
- `debug_accounts_custom_policy_only`: organization.debug_log_delivery_accounts applies only to custom_policy rules
- `org_lambda_uses_trigger_types`: organization custom_lambda rules declare triggers via organization.trigger_types, not custom_lambda.source_details
- `remediation_account_scope_only`: remediation is not supported on organization rules
- `evaluation_modes_account_scope_only`: evaluation_modes apply only to account-scoped rules

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsConfigRule, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.rule_arn` | `string` | The rule's ARN. For organization rules this is the organization rule ARN. |
| `status.outputs.rule_name` | `string` | The rule's name (metadata.name echoed) - the key remediation attachments and aggregator queries address rules by. |
| `status.outputs.rule_id` | `string` | The rule's AWS-assigned ID (account-scoped rules; empty for organization rules, which have no rule_id). |
| `status.outputs.remediation_arn` | `string` | The remediation configuration's ARN (set only when spec.remediation is configured). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.customLambda.functionArn` | AwsLambda | `status.outputs.function_arn` |

## See Also

- [Overview](../README.md)

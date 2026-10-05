# AwsCloudwatchLogAccountPolicy

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsCloudwatchLogAccountPolicySpec defines one CloudWatch Logs
account-level policy: a policy document AWS applies account-wide in
the region, to every log group (or the subset selection_criteria
narrows to). One policy object exists per (name, type) pair - two
instances sharing both would fight over one AWS object.

The five policy types carry different documents: DATA_PROTECTION
masks sensitive data account-wide; SUBSCRIPTION_FILTER forwards
matching events account-wide; FIELD_INDEX declares indexed fields;
TRANSFORMER rewrites events at ingest; METRIC_EXTRACTION derives
metrics from log fields. The document's schema is the policy type's
own - AWS validates it server-side at Put time.

## Example

```yaml
# Canonical AwsCloudwatchLogAccountPolicy example (hack/dev manifest
# and refgen Example source): an account-wide field-index policy.
# Field-index policies always apply account-wide - AWS accepts
# selection criteria only on SUBSCRIPTION_FILTER_POLICY, as an
# exact-name exclusion list (LogGroupName NOT IN [...]).
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudwatchLogAccountPolicy
metadata:
  name: api-field-index
  id: api-field-index
  org: test-org
  env: dev
spec:
  region: us-west-2
  policyName: api-field-index
  policyType: FIELD_INDEX_POLICY
  policyDocument:
    Fields:
      - requestId
      - sourceIp
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.policyName` | `string` | yes |  |  |
| `spec.policyType` | `string` |  |  |  |
| `spec.policyDocument` | `object` | yes |  |  |
| `spec.selectionCriteria` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the policy applies in (account policies are
per-region objects). Example: "us-east-1".

- rule: {"string":{"minLen":"1"}}

### spec.policyName

`string` · required

The policy's name in AWS. Together with policy_type it is the
object's identity - changing either replaces the policy. The
provider imports the pair as "policy_name:policy_type".

- rule: {"string":{"minLen":"1","maxLen":"256"}}

### spec.policyType

`string`

Which account-wide capability this policy configures. Changing the
type replaces the policy.

- rule: {"string":{"in":["DATA_PROTECTION_POLICY","SUBSCRIPTION_FILTER_POLICY","FIELD_INDEX_POLICY","TRANSFORMER_POLICY","METRIC_EXTRACTION_POLICY"]}}

### spec.policyDocument

`object` · required

The policy document, in the chosen type's own JSON schema (the
CloudWatch Logs console's per-type editor shows the exact shape;
for DATA_PROTECTION it is the data-protection statement document,
for SUBSCRIPTION_FILTER a {destination_arn, filter_pattern, ...}
object, and so on). AWS validates it server-side; the engines diff
it semantically so formatting never causes drift.

- rule: {"required":true}

### spec.selectionCriteria

`string`

Excludes log groups from a SUBSCRIPTION_FILTER_POLICY, in AWS's
one supported grammar: "LogGroupName NOT IN [\"name1\", \"name2\"]"
(an exact-name exclusion list - no prefix form, no IN form). AWS
accepts selection criteria ONLY on subscription-filter policies:
PutAccountPolicy rejects any criteria string on the other four
types with "Invalid selection criteria provided", so every other
policy type applies account-wide, period. Changing it replaces the
policy. The provider's scope argument is deliberately not modeled:
ALL is its only legal value at the pin, so the modules pin it (a
recorded exclusion).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

## Validation Rules

- `spec.selection_criteria_subscription_filter_only`: selection_criteria is accepted by AWS only when policy_type is SUBSCRIPTION_FILTER_POLICY

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsCloudwatchLogAccountPolicy, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.policy_name` | `string` | The policy's name. |
| `status.outputs.policy_type` | `string` | The policy's type - with the name, the provider's import ID ("policy_name:policy_type"). |

## See Also

- [Overview](../README.md)

# AwsConfigAggregator

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsConfigAggregatorSpec defines the desired configuration for AWS
Config aggregation - the cross-account, cross-region rollup of
Config data into one queryable view.

Aggregation has two sides, and this component models both as arms:

  - aggregation: the AGGREGATOR itself, deployed in the account
    that collects. It references no Config recorder - aggregation
    works in an account with zero recorders (the collected data
    comes from the SOURCE accounts' recorders).
  - authorizations: the reciprocal GRANTS, deployed in each SOURCE
    account, naming the aggregator account+region allowed to
    collect from it. Organization-sourced aggregators need no
    grants; account-list aggregators need one in every source
    account outside the deployer's own.

A same-account rollup needs only the aggregation arm. A
cross-account topology deploys this component twice: once in the
aggregator account (aggregation arm) and once per source account
(authorizations arm). The aggregator's name is metadata.name.

Destroying this component is a real delete of whichever arms it
manages; aggregated data disappears from the view but source
accounts' Config data is untouched.

## Example

```yaml
# Canonical AwsConfigAggregator example (hack/dev manifest and refgen
# Example source): an account playing both aggregation roles -- it
# collects two accounts' Config data across two regions AND grants a
# sibling account's aggregator permission to collect from it. Literal
# account ids stand in so the offline `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsConfigAggregator
metadata:
  name: org-config-rollup
  id: org-config-rollup
  org: test-org
  env: dev
spec:
  region: us-west-2
  aggregation:
    accountSource:
      accountIds:
        - "123456789012"
        - "210987654321"
      regions:
        - us-west-2
        - us-east-1
  authorizations:
    - accountId: "310987654321"
      authorizedAwsRegion: us-west-2
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.aggregation` | `AwsConfigAggregatorAggregation` |  |  |  |
| `spec.aggregation.accountSource` | `AwsConfigAggregatorAccountSource` |  |  |  |
| `spec.aggregation.accountSource.accountIds` | `[]string` | yes |  |  |
| `spec.aggregation.accountSource.allRegions` | `bool` |  |  |  |
| `spec.aggregation.accountSource.regions` | `[]string` |  |  |  |
| `spec.aggregation.organizationSource` | `AwsConfigAggregatorOrganizationSource` |  |  |  |
| `spec.aggregation.organizationSource.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.aggregation.organizationSource.allRegions` | `bool` |  |  |  |
| `spec.aggregation.organizationSource.regions` | `[]string` |  |  |  |
| `spec.authorizations` | `[]AwsConfigAggregatorAuthorization` |  |  |  |
| `spec.authorizations[].accountId` | `string` | yes |  |  |
| `spec.authorizations[].authorizedAwsRegion` | `string` | yes |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the managed resources live in: the aggregator's
home region, and the region the grants are created in (a grant
lives in the source account's region of the GRANTING side; the
aggregator region it authorizes is named per grant).
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.aggregation

`AwsConfigAggregatorAggregation`

The aggregator arm - deploy in the account that collects.

- rule: set exactly one of account_source and organization_source

### spec.aggregation.accountSource

`AwsConfigAggregatorAccountSource`

Collect from an explicit list of accounts. Mutually exclusive
with organization_source.

- rule: list regions or set all_regions true

### spec.aggregation.accountSource.accountIds

`[]string` · required

The 12-digit source account IDs. The deployer's own account may
appear (no grant needed for it); every OTHER account needs an
authorizations-arm deployment on its side before data flows.

- rule: {"repeated":{"minItems":"1","unique":true,"items":{"string":{"pattern":"^[0-9]{12}$"}}}}

### spec.aggregation.accountSource.allRegions

`bool`

Collect from every AWS region. Unset = false: list the regions
explicitly instead. AWS wants one of the two shapes - regions
listed, or all_regions true.

### spec.aggregation.accountSource.regions

`[]string`

The regions to collect from (used when all_regions is false).

- rule: {"repeated":{"unique":true,"items":{"string":{"minLen":"1"}}}}

### spec.aggregation.organizationSource

`AwsConfigAggregatorOrganizationSource`

Collect from every account in the AWS Organization. Requires
running in the management account or a delegated administrator
for Config. Mutually exclusive with account_source.

- rule: list regions or set all_regions true

### spec.aggregation.organizationSource.roleArn

`string | valueFrom` · required

The IAM role Config assumes to read the organization structure.
It needs the AWSConfigRoleForOrganizations managed policy (or
equivalent organizations:List*/Describe* permissions) and a trust
policy for "config.amazonaws.com".

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.aggregation.organizationSource.allRegions

`bool`

Collect from every AWS region. Unset = false: list the regions
explicitly instead.

### spec.aggregation.organizationSource.regions

`[]string`

The regions to collect from (used when all_regions is false).

- rule: {"repeated":{"unique":true,"items":{"string":{"minLen":"1"}}}}

### spec.authorizations

`[]AwsConfigAggregatorAuthorization`

The grant arm - deploy in each source account, one entry per
aggregator allowed to collect from it. Duplicate
account+region pairs configure nothing and AWS rejects them.

### spec.authorizations[].accountId

`string` · required

The 12-digit account ID of the AGGREGATOR (the collector, not
this account). Changing it replaces the grant.

- rule: {"required":true,"string":{"pattern":"^[0-9]{12}$"}}

### spec.authorizations[].authorizedAwsRegion

`string` · required

The region the aggregator lives in (the aggregator's home
region, not necessarily this deployment's region). Changing it
replaces the grant.

- rule: {"required":true,"string":{"minLen":"1"}}

## Validation Rules

- `at_least_one_arm`: set aggregation (the collector) or authorizations (the source-account grants) - or both for an account playing both roles

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsConfigAggregator, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.aggregator_name` | `string` | The aggregator's name (also the provider's import ID). Set only when spec.aggregation is configured. |
| `status.outputs.aggregator_arn` | `string` | The aggregator's ARN. Set only when spec.aggregation is configured. |
| `status.outputs.authorization_arns` | `map<string, string>` | The grants' ARNs, keyed "{account_id}:{authorized_aws_region}" (each key is also that grant's provider import ID). One entry per spec.authorizations element. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.aggregation.organizationSource.roleArn` | AwsIamRole | `status.outputs.role_arn` |

## See Also

- [Overview](../README.md)

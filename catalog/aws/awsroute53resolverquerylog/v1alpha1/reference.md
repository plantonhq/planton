# AwsRoute53ResolverQueryLog

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsRoute53ResolverQueryLogSpec defines one Resolver query logging
configuration - the pipeline that records the DNS queries VPCs make
through Route 53 Resolver - with its VPC associations managed
in-line.

This logs RESOLVER queries (everything a VPC asks, including
queries the resolver answers from cache or forwards on-prem). It is
a different surface from hosted-zone query logging, which records
only what Route 53 answers for one public zone and lives on
AwsRoute53Zone.

Both the destination and the name are fixed for life at the
provider - changing either replaces the configuration (existing log
data stays in the destination).

## Example

```yaml
# Canonical AwsRoute53ResolverQueryLog example (hack/dev manifest and
# refgen Example source): resolver queries from one VPC logged to a
# CloudWatch log group.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRoute53ResolverQueryLog
metadata:
  name: vpc-dns-queries
  id: vpc-dns-queries
  org: test-org
  env: dev
spec:
  region: us-west-2
  destinationArn:
    value: arn:aws:logs:us-west-2:123456789012:log-group:resolver-queries
  vpcIds:
    - value: vpc-0123456789abcdef0
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.destinationArn` | `string \| valueFrom` | yes |  | AwsCloudwatchLogGroup (`status.outputs.log_group_arn`) |
| `spec.vpcIds` | `[]string \| valueFrom` |  |  | AwsVpc (`status.outputs.vpc_id`) |

## Field Details

### spec.region

`string` · required

The AWS region whose resolver queries are logged. Example:
"us-west-2".

- rule: {"string":{"minLen":"1"}}

### spec.destinationArn

`string | valueFrom` · required

Where the logs go: a CloudWatch log group ARN, an S3 bucket ARN
(optionally with a prefix, "arn:...:my-bucket/prefix"), or a
Kinesis Data Firehose delivery stream ARN. Each destination type
needs the documented resource policy or permissions in place -
an association against a destination the resolver cannot write
to fails asynchronously after apply. Reference an
AwsCloudwatchLogGroup log_group_arn output or pass a literal ARN.
Fixed for life.

- references: AwsCloudwatchLogGroup (`status.outputs.log_group_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCloudwatchLogGroup, name: <that resource's name>, fieldPath: status.outputs.log_group_arn}} -- a bare string does not parse

### spec.vpcIds

`[]string | valueFrom`

The VPCs whose resolver queries are logged to the destination.
Reference AwsVpc vpc_id outputs or pass literal vpc-... ids.

- references: AwsVpc (`status.outputs.vpc_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsRoute53ResolverQueryLog, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.query_log_config_id` | `string` | The configuration's id (rqlc-...) - the provider's import ID. |
| `status.outputs.query_log_config_arn` | `string` | The configuration's ARN. |
| `status.outputs.share_status` | `string` | Whether the configuration is shared via RAM (NOT_SHARED / SHARED_BY_ME / SHARED_WITH_ME). |
| `status.outputs.association_ids` | `map<string, string>` | AWS-generated association IDs (rqlca-...) keyed by the resolved VPC id. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.destinationArn` | AwsCloudwatchLogGroup | `status.outputs.log_group_arn` |
| `spec.vpcIds` | AwsVpc | `status.outputs.vpc_id` |

## See Also

- [Overview](../README.md)

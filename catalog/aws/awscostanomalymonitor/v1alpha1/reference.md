# AwsCostAnomalyMonitor

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsCostAnomalyMonitorSpec defines one Cost Explorer anomaly
monitor - the ML-driven watcher that flags unusual spend - together
with its folded alert SUBSCRIPTIONS (a subscription's monitor list
is the structural edge that makes it this monitor's satellite).

A monitor is one of two shapes, chosen by monitor_type:
DIMENSIONAL segments spend by ONE built-in dimension (the common
"by service" monitor); CUSTOM watches the slice a Cost Explorer
expression selects (linked accounts, tags, cost categories).
Both shape arms are create-only - changing the shape replaces the
monitor; only the name updates in place.

Cost Explorer is account-global (served from us-east-1); the spec's
region is the provider endpoint region. Anomaly detection itself is
free - AWS bills nothing for monitors or subscriptions.

## Example

```yaml
# Canonical AwsCostAnomalyMonitor example (hack/dev manifest and
# refgen Example source): the recommended by-service dimensional
# monitor with a daily email summary above a $100 absolute impact
# threshold.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCostAnomalyMonitor
metadata:
  name: service-spend-monitor
  id: service-spend-monitor
  org: test-org
  env: dev
spec:
  region: us-east-1
  monitorName: Service Spend Monitor
  monitorType: DIMENSIONAL
  monitorDimension: SERVICE
  subscriptions:
    - name: finops-daily-summary
      frequency: DAILY
      subscribers:
        - address:
            value: finops@example.com
          type: EMAIL
      thresholdExpression:
        dimension:
          key: ANOMALY_TOTAL_IMPACT_ABSOLUTE
          matchOptions:
            - GREATER_THAN_OR_EQUAL
          values:
            - "100"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.monitorName` | `string` | yes |  |  |
| `spec.monitorType` | `string` |  |  |  |
| `spec.monitorDimension` | `string` |  |  |  |
| `spec.monitorSpecification` | `object` |  |  |  |
| `spec.subscriptions` | `[]AwsCostAnomalyMonitorSubscription` |  |  |  |
| `spec.subscriptions[].name` | `string` | yes |  |  |
| `spec.subscriptions[].frequency` | `string` |  |  |  |
| `spec.subscriptions[].subscribers` | `[]AwsCostAnomalyMonitorSubscriber` | yes |  |  |
| `spec.subscriptions[].subscribers[].address` | `string \| valueFrom` | yes |  | AwsSnsTopic (`status.outputs.topic_arn`) |
| `spec.subscriptions[].subscribers[].type` | `string` |  |  |  |
| `spec.subscriptions[].thresholdExpression` | `AwsCostAnomalyMonitorExpression` |  |  |  |
| `spec.subscriptions[].thresholdExpression.dimension` | `AwsCostAnomalyMonitorExpressionDimension` |  |  |  |
| `spec.subscriptions[].thresholdExpression.dimension.key` | `string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.dimension.matchOptions` | `[]string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.dimension.values` | `[]string` | yes |  |  |
| `spec.subscriptions[].thresholdExpression.tag` | `AwsCostAnomalyMonitorExpressionTag` |  |  |  |
| `spec.subscriptions[].thresholdExpression.tag.key` | `string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.tag.matchOptions` | `[]string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.tag.values` | `[]string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.costCategory` | `AwsCostAnomalyMonitorExpressionCostCategory` |  |  |  |
| `spec.subscriptions[].thresholdExpression.costCategory.key` | `string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.costCategory.values` | `[]string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.and` | `[]AwsCostAnomalyMonitorExpressionLeaf` |  |  |  |
| `spec.subscriptions[].thresholdExpression.and[].dimension` | `AwsCostAnomalyMonitorExpressionDimension` |  |  |  |
| `spec.subscriptions[].thresholdExpression.and[].dimension.key` | `string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.and[].dimension.matchOptions` | `[]string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.and[].dimension.values` | `[]string` | yes |  |  |
| `spec.subscriptions[].thresholdExpression.and[].tag` | `AwsCostAnomalyMonitorExpressionTag` |  |  |  |
| `spec.subscriptions[].thresholdExpression.and[].tag.key` | `string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.and[].tag.matchOptions` | `[]string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.and[].tag.values` | `[]string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.and[].costCategory` | `AwsCostAnomalyMonitorExpressionCostCategory` |  |  |  |
| `spec.subscriptions[].thresholdExpression.and[].costCategory.key` | `string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.and[].costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.and[].costCategory.values` | `[]string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.or` | `[]AwsCostAnomalyMonitorExpressionLeaf` |  |  |  |
| `spec.subscriptions[].thresholdExpression.or[].dimension` | `AwsCostAnomalyMonitorExpressionDimension` |  |  |  |
| `spec.subscriptions[].thresholdExpression.or[].dimension.key` | `string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.or[].dimension.matchOptions` | `[]string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.or[].dimension.values` | `[]string` | yes |  |  |
| `spec.subscriptions[].thresholdExpression.or[].tag` | `AwsCostAnomalyMonitorExpressionTag` |  |  |  |
| `spec.subscriptions[].thresholdExpression.or[].tag.key` | `string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.or[].tag.matchOptions` | `[]string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.or[].tag.values` | `[]string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.or[].costCategory` | `AwsCostAnomalyMonitorExpressionCostCategory` |  |  |  |
| `spec.subscriptions[].thresholdExpression.or[].costCategory.key` | `string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.or[].costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.or[].costCategory.values` | `[]string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.not` | `AwsCostAnomalyMonitorExpressionLeaf` |  |  |  |
| `spec.subscriptions[].thresholdExpression.not.dimension` | `AwsCostAnomalyMonitorExpressionDimension` |  |  |  |
| `spec.subscriptions[].thresholdExpression.not.dimension.key` | `string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.not.dimension.matchOptions` | `[]string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.not.dimension.values` | `[]string` | yes |  |  |
| `spec.subscriptions[].thresholdExpression.not.tag` | `AwsCostAnomalyMonitorExpressionTag` |  |  |  |
| `spec.subscriptions[].thresholdExpression.not.tag.key` | `string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.not.tag.matchOptions` | `[]string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.not.tag.values` | `[]string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.not.costCategory` | `AwsCostAnomalyMonitorExpressionCostCategory` |  |  |  |
| `spec.subscriptions[].thresholdExpression.not.costCategory.key` | `string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.not.costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.subscriptions[].thresholdExpression.not.costCategory.values` | `[]string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region used by the provider while managing the monitor.
Cost Explorer is account-global (served from us-east-1) - every
API call still needs a regional endpoint. Example: "us-east-1".

- rule: {"string":{"minLen":"1"}}

### spec.monitorName

`string` · required

The monitor's display name - an explicit field because monitor
names legally carry spaces and mixed case metadata.name cannot.
Up to 1024 characters; the only field that updates in place.

- rule: {"string":{"minLen":"1","maxLen":"1024"}}

### spec.monitorType

`string`

The monitor's shape. DIMENSIONAL segments spend by one built-in
dimension; CUSTOM watches a Cost Explorer expression's slice.
Changing the shape replaces the monitor.

DIMENSIONAL is an account SINGLETON: AWS permits exactly one
services monitor per account, and AUTO-CREATES it
("Default-Services-Monitor") for every account that enabled Cost
Explorer on or after 2023-03-27 - so on most accounts creating a
DIMENSIONAL monitor fails at apply with "ValidationException:
Limit exceeded on dimensional spend monitor creation"
(server-verified 2026-08-25). If your account already carries the
default monitor, import it instead of creating, or use CUSTOM
(up to 500 custom monitors per account).

- rule: {"string":{"in":["DIMENSIONAL","CUSTOM"]}}

### spec.monitorDimension

`string`

For DIMENSIONAL monitors: the segmentation dimension. SERVICE is
AWS's recommended default posture (one anomaly stream per
service) - and exactly the shape AWS auto-provisions as
"Default-Services-Monitor" on post-2023 Cost Explorer accounts
(see monitor_type: one per account, import rather than create
when it already exists). Create-only.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["SERVICE","LINKED_ACCOUNT","TAG","COST_CATEGORY"]}}

### spec.monitorSpecification

`object`

For CUSTOM monitors: the Cost Explorer Expression (as free-form
JSON, the AWS Expression document verbatim) selecting the spend
slice to watch - e.g. {"Dimensions": {"Key": "LINKED_ACCOUNT",
"Values": ["123456789012"]}}. Create-only.

Author the document in its CANONICAL stored form or every
subsequent plan proposes a replacement (both contracts
server-verified 2026-08-25): (1) every root member present -
And, CostCategories, Dimensions, Not, Or, Tags - with unused
ones explicitly null (the provider re-marshals the SDK's
Expression struct on read, emitting ALL members; null-vs-absent
is not JSON-equivalent to its diff suppression); (2) tag keys
carry Cost Explorer's canonical prefix - "user:<key>" for
user-defined cost-allocation tags (e.g. "user:team"),
"aws:<key>" for AWS-generated ones - because CE accepts an
unprefixed key at create but echoes it back prefixed. The
upstream provider's own tests spell exactly this form.

### spec.subscriptions

`[]AwsCostAnomalyMonitorSubscription`

Alert subscriptions - the folded satellites: each entry is one
aws_ce_anomaly_subscription bound to this monitor, keyed by its
name. A subscription decides who hears about anomalies, how often,
and above what impact threshold.

- rule: IMMEDIATE subscriptions deliver individual alerts via SNS - every subscriber must be type SNS
- rule: DAILY and WEEKLY summary subscriptions deliver via email - every subscriber must be type EMAIL

### spec.subscriptions[].name

`string` · required

The subscription's name - the key the modules use for the
for_each entry and the outputs map, and the AlertSubscription
name at AWS. Up to 1024 characters.

- rule: {"string":{"minLen":"1","maxLen":"1024"}}

### spec.subscriptions[].frequency

`string`

How alerts deliver: IMMEDIATE (individual alerts, SNS),
DAILY or WEEKLY (email summaries).

- rule: {"string":{"in":["IMMEDIATE","DAILY","WEEKLY"]}}

### spec.subscriptions[].subscribers

`[]AwsCostAnomalyMonitorSubscriber` · required

Who receives the alerts (at least one).

- rule: {"repeated":{"minItems":"1"}}

### spec.subscriptions[].subscribers[].address

`string | valueFrom` · required

An email address (type EMAIL) or an SNS topic ARN (type SNS -
reference an AwsSnsTopic's topic_arn output or pass a literal
ARN; the topic's policy must allow costalerts.amazonaws.com to
publish).

- references: AwsSnsTopic (`status.outputs.topic_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSnsTopic, name: <that resource's name>, fieldPath: status.outputs.topic_arn}} -- a bare string does not parse

### spec.subscriptions[].subscribers[].type

`string`

How the subscriber is reached.

- rule: {"string":{"in":["EMAIL","SNS"]}}

### spec.subscriptions[].thresholdExpression

`AwsCostAnomalyMonitorExpression`

The impact threshold an anomaly must exceed before this
subscription alerts - a Cost Explorer expression over the
ANOMALY_TOTAL_IMPACT_ABSOLUTE or ANOMALY_TOTAL_IMPACT_PERCENTAGE
dimensions (e.g. "absolute impact >= 100 USD"). Unset = every
anomaly the monitor flags alerts.

### spec.subscriptions[].thresholdExpression.dimension

`AwsCostAnomalyMonitorExpressionDimension`

A dimension leaf. For threshold expressions the impact dimensions
(ANOMALY_TOTAL_IMPACT_ABSOLUTE / ANOMALY_TOTAL_IMPACT_PERCENTAGE)
with match option GREATER_THAN_OR_EQUAL are the meaningful keys.

### spec.subscriptions[].thresholdExpression.dimension.key

`string`

The dimension to filter on.

- rule: {"string":{"in":["AZ","INSTANCE_TYPE","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","OPERATION","PURCHASE_TYPE","REGION","SERVICE","SERVICE_CODE","USAGE_TYPE","USAGE_TYPE_GROUP","RECORD_TYPE","OPERATING_SYSTEM","TENANCY","SCOPE","PLATFORM","SUBSCRIPTION_ID","LEGAL_ENTITY_NAME","INVOICING_ENTITY","DEPLOYMENT_OPTION","DATABASE_ENGINE","CACHE_ENGINE","INSTANCE_TYPE_FAMILY","BILLING_ENTITY","RESERVATION_ID","RESOURCE_ID","RIGHTSIZING_TYPE","SAVINGS_PLANS_TYPE","SAVINGS_PLAN_ARN","PAYMENT_OPTION","AGREEMENT_END_DATE_TIME_AFTER","AGREEMENT_END_DATE_TIME_BEFORE","PAYER_ACCOUNT","ANOMALY_TOTAL_IMPACT_ABSOLUTE","ANOMALY_TOTAL_IMPACT_PERCENTAGE"]}}

### spec.subscriptions[].thresholdExpression.dimension.matchOptions

`[]string`

How values match. Unset = EQUALS. Impact-threshold dimensions use
GREATER_THAN_OR_EQUAL.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.subscriptions[].thresholdExpression.dimension.values

`[]string` · required

The values to match (for impact thresholds: the numeric threshold
as a string, e.g. "100").

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.subscriptions[].thresholdExpression.tag

`AwsCostAnomalyMonitorExpressionTag`

A cost-allocation-tag leaf.

### spec.subscriptions[].thresholdExpression.tag.key

`string`

The tag key.

- rule: {"string":{"maxLen":"1024"}}

### spec.subscriptions[].thresholdExpression.tag.matchOptions

`[]string`

How values match (see the dimension leaf).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.subscriptions[].thresholdExpression.tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.subscriptions[].thresholdExpression.costCategory

`AwsCostAnomalyMonitorExpressionCostCategory`

A cost-category leaf.

### spec.subscriptions[].thresholdExpression.costCategory.key

`string`

The cost category's name.

- rule: {"string":{"maxLen":"50"}}

### spec.subscriptions[].thresholdExpression.costCategory.matchOptions

`[]string`

How values match (see the dimension leaf).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.subscriptions[].thresholdExpression.costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.subscriptions[].thresholdExpression.and

`[]AwsCostAnomalyMonitorExpressionLeaf`

Leaf children that must ALL match (e.g. absolute AND percentage
impact thresholds together).

### spec.subscriptions[].thresholdExpression.and[].dimension

`AwsCostAnomalyMonitorExpressionDimension`

A dimension leaf.

### spec.subscriptions[].thresholdExpression.and[].dimension.key

`string`

The dimension to filter on.

- rule: {"string":{"in":["AZ","INSTANCE_TYPE","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","OPERATION","PURCHASE_TYPE","REGION","SERVICE","SERVICE_CODE","USAGE_TYPE","USAGE_TYPE_GROUP","RECORD_TYPE","OPERATING_SYSTEM","TENANCY","SCOPE","PLATFORM","SUBSCRIPTION_ID","LEGAL_ENTITY_NAME","INVOICING_ENTITY","DEPLOYMENT_OPTION","DATABASE_ENGINE","CACHE_ENGINE","INSTANCE_TYPE_FAMILY","BILLING_ENTITY","RESERVATION_ID","RESOURCE_ID","RIGHTSIZING_TYPE","SAVINGS_PLANS_TYPE","SAVINGS_PLAN_ARN","PAYMENT_OPTION","AGREEMENT_END_DATE_TIME_AFTER","AGREEMENT_END_DATE_TIME_BEFORE","PAYER_ACCOUNT","ANOMALY_TOTAL_IMPACT_ABSOLUTE","ANOMALY_TOTAL_IMPACT_PERCENTAGE"]}}

### spec.subscriptions[].thresholdExpression.and[].dimension.matchOptions

`[]string`

How values match. Unset = EQUALS. Impact-threshold dimensions use
GREATER_THAN_OR_EQUAL.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.subscriptions[].thresholdExpression.and[].dimension.values

`[]string` · required

The values to match (for impact thresholds: the numeric threshold
as a string, e.g. "100").

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.subscriptions[].thresholdExpression.and[].tag

`AwsCostAnomalyMonitorExpressionTag`

A cost-allocation-tag leaf.

### spec.subscriptions[].thresholdExpression.and[].tag.key

`string`

The tag key.

- rule: {"string":{"maxLen":"1024"}}

### spec.subscriptions[].thresholdExpression.and[].tag.matchOptions

`[]string`

How values match (see the dimension leaf).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.subscriptions[].thresholdExpression.and[].tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.subscriptions[].thresholdExpression.and[].costCategory

`AwsCostAnomalyMonitorExpressionCostCategory`

A cost-category leaf.

### spec.subscriptions[].thresholdExpression.and[].costCategory.key

`string`

The cost category's name.

- rule: {"string":{"maxLen":"50"}}

### spec.subscriptions[].thresholdExpression.and[].costCategory.matchOptions

`[]string`

How values match (see the dimension leaf).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.subscriptions[].thresholdExpression.and[].costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.subscriptions[].thresholdExpression.or

`[]AwsCostAnomalyMonitorExpressionLeaf`

Leaf children of which AT LEAST ONE must match.

### spec.subscriptions[].thresholdExpression.or[].dimension

`AwsCostAnomalyMonitorExpressionDimension`

A dimension leaf.

### spec.subscriptions[].thresholdExpression.or[].dimension.key

`string`

The dimension to filter on.

- rule: {"string":{"in":["AZ","INSTANCE_TYPE","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","OPERATION","PURCHASE_TYPE","REGION","SERVICE","SERVICE_CODE","USAGE_TYPE","USAGE_TYPE_GROUP","RECORD_TYPE","OPERATING_SYSTEM","TENANCY","SCOPE","PLATFORM","SUBSCRIPTION_ID","LEGAL_ENTITY_NAME","INVOICING_ENTITY","DEPLOYMENT_OPTION","DATABASE_ENGINE","CACHE_ENGINE","INSTANCE_TYPE_FAMILY","BILLING_ENTITY","RESERVATION_ID","RESOURCE_ID","RIGHTSIZING_TYPE","SAVINGS_PLANS_TYPE","SAVINGS_PLAN_ARN","PAYMENT_OPTION","AGREEMENT_END_DATE_TIME_AFTER","AGREEMENT_END_DATE_TIME_BEFORE","PAYER_ACCOUNT","ANOMALY_TOTAL_IMPACT_ABSOLUTE","ANOMALY_TOTAL_IMPACT_PERCENTAGE"]}}

### spec.subscriptions[].thresholdExpression.or[].dimension.matchOptions

`[]string`

How values match. Unset = EQUALS. Impact-threshold dimensions use
GREATER_THAN_OR_EQUAL.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.subscriptions[].thresholdExpression.or[].dimension.values

`[]string` · required

The values to match (for impact thresholds: the numeric threshold
as a string, e.g. "100").

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.subscriptions[].thresholdExpression.or[].tag

`AwsCostAnomalyMonitorExpressionTag`

A cost-allocation-tag leaf.

### spec.subscriptions[].thresholdExpression.or[].tag.key

`string`

The tag key.

- rule: {"string":{"maxLen":"1024"}}

### spec.subscriptions[].thresholdExpression.or[].tag.matchOptions

`[]string`

How values match (see the dimension leaf).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.subscriptions[].thresholdExpression.or[].tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.subscriptions[].thresholdExpression.or[].costCategory

`AwsCostAnomalyMonitorExpressionCostCategory`

A cost-category leaf.

### spec.subscriptions[].thresholdExpression.or[].costCategory.key

`string`

The cost category's name.

- rule: {"string":{"maxLen":"50"}}

### spec.subscriptions[].thresholdExpression.or[].costCategory.matchOptions

`[]string`

How values match (see the dimension leaf).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.subscriptions[].thresholdExpression.or[].costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.subscriptions[].thresholdExpression.not

`AwsCostAnomalyMonitorExpressionLeaf`

A leaf child that must NOT match.

### spec.subscriptions[].thresholdExpression.not.dimension

`AwsCostAnomalyMonitorExpressionDimension`

A dimension leaf.

### spec.subscriptions[].thresholdExpression.not.dimension.key

`string`

The dimension to filter on.

- rule: {"string":{"in":["AZ","INSTANCE_TYPE","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","OPERATION","PURCHASE_TYPE","REGION","SERVICE","SERVICE_CODE","USAGE_TYPE","USAGE_TYPE_GROUP","RECORD_TYPE","OPERATING_SYSTEM","TENANCY","SCOPE","PLATFORM","SUBSCRIPTION_ID","LEGAL_ENTITY_NAME","INVOICING_ENTITY","DEPLOYMENT_OPTION","DATABASE_ENGINE","CACHE_ENGINE","INSTANCE_TYPE_FAMILY","BILLING_ENTITY","RESERVATION_ID","RESOURCE_ID","RIGHTSIZING_TYPE","SAVINGS_PLANS_TYPE","SAVINGS_PLAN_ARN","PAYMENT_OPTION","AGREEMENT_END_DATE_TIME_AFTER","AGREEMENT_END_DATE_TIME_BEFORE","PAYER_ACCOUNT","ANOMALY_TOTAL_IMPACT_ABSOLUTE","ANOMALY_TOTAL_IMPACT_PERCENTAGE"]}}

### spec.subscriptions[].thresholdExpression.not.dimension.matchOptions

`[]string`

How values match. Unset = EQUALS. Impact-threshold dimensions use
GREATER_THAN_OR_EQUAL.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.subscriptions[].thresholdExpression.not.dimension.values

`[]string` · required

The values to match (for impact thresholds: the numeric threshold
as a string, e.g. "100").

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.subscriptions[].thresholdExpression.not.tag

`AwsCostAnomalyMonitorExpressionTag`

A cost-allocation-tag leaf.

### spec.subscriptions[].thresholdExpression.not.tag.key

`string`

The tag key.

- rule: {"string":{"maxLen":"1024"}}

### spec.subscriptions[].thresholdExpression.not.tag.matchOptions

`[]string`

How values match (see the dimension leaf).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.subscriptions[].thresholdExpression.not.tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.subscriptions[].thresholdExpression.not.costCategory

`AwsCostAnomalyMonitorExpressionCostCategory`

A cost-category leaf.

### spec.subscriptions[].thresholdExpression.not.costCategory.key

`string`

The cost category's name.

- rule: {"string":{"maxLen":"50"}}

### spec.subscriptions[].thresholdExpression.not.costCategory.matchOptions

`[]string`

How values match (see the dimension leaf).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.subscriptions[].thresholdExpression.not.costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

## Validation Rules

- `spec.dimensional_requires_dimension`: DIMENSIONAL monitors segment by monitor_dimension - set it (and leave monitor_specification unset)
- `spec.custom_requires_specification`: CUSTOM monitors watch the slice monitor_specification selects - set it (and leave monitor_dimension unset)
- `spec.subscription_names_unique`: subscriptions entries must have unique names

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsCostAnomalyMonitor, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.monitor_arn` | `string` | The monitor's ARN (also the provider's import ID). |
| `status.outputs.subscription_arns` | `map<string, string>` | Subscription ARNs keyed by subscription name (each subscription imports by its ARN). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.subscriptions[].subscribers[].address` | AwsSnsTopic | `status.outputs.topic_arn` |

## See Also

- [Overview](../README.md)

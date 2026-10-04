# AwsBudget

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsBudgetSpec defines one AWS Budgets budget - a spend or usage
threshold AWS evaluates continuously - together with its alert
notifications and its folded budget ACTIONS (each action exists only
on its budget and fires an IAM-policy application, an SCP
attachment, or SSM instance stops when its threshold breaches).

Budgets is an account-global service served from us-east-1; the
spec's region is the provider endpoint region, never the budget's
location. The budget's scope is chosen by exactly ONE funding shape:
a fixed limit, per-period planned limits, or auto-adjustment from
history/forecast.

Filtering has two generations: the legacy cost_filters name/values
pairs, and the modern filter_expression tree paired with a metric.
AWS accepts one generation per budget, never both.

## Example

```yaml
# Canonical AwsBudget example (hack/dev manifest and refgen Example
# source): a monthly cost budget with the modern metric +
# filter-expression generation, a forecast alert, and an automatic
# IAM-guardrail action. Literal ARNs stand in for composed references
# so the offline `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBudget
metadata:
  name: engineering-monthly
  id: engineering-monthly
  org: test-org
  env: dev
spec:
  region: us-east-1
  budgetName: Engineering Monthly Spend
  budgetType: COST
  timeUnit: MONTHLY
  limit:
    amount: "5000"
    unit: USD
  metric: UnblendedCost
  filterExpression:
    and:
      - dimension:
          key: LINKED_ACCOUNT
          values:
            - "123456789012"
      - not:
          tag:
            key: environment
            values:
              - sandbox
  notifications:
    - comparisonOperator: GREATER_THAN
      notificationType: FORECASTED
      threshold: 80
      thresholdType: PERCENTAGE
      subscriberEmailAddresses:
        - finops@example.com
  actions:
    - name: freeze-dev-provisioning
      actionType: APPLY_IAM_POLICY
      approvalModel: AUTOMATIC
      notificationType: ACTUAL
      executionRoleArn:
        value: arn:aws:iam::123456789012:role/budget-actions
      actionThreshold:
        actionThresholdType: PERCENTAGE
        actionThresholdValue: 100
      subscribers:
        - address:
            value: finops@example.com
          subscriptionType: EMAIL
      iamActionDefinition:
        policyArn:
          value: arn:aws:iam::aws:policy/AWSDenyAll
        groups:
          - value: developers
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.budgetName` | `string` | yes |  |  |
| `spec.budgetType` | `string` |  |  |  |
| `spec.timeUnit` | `string` |  |  |  |
| `spec.accountId` | `string` |  |  |  |
| `spec.billingViewArn` | `string` |  |  |  |
| `spec.limit` | `AwsBudgetLimit` |  |  |  |
| `spec.limit.amount` | `string` |  |  |  |
| `spec.limit.unit` | `string` | yes |  |  |
| `spec.plannedLimits` | `[]AwsBudgetPlannedLimit` |  |  |  |
| `spec.plannedLimits[].startTime` | `string` |  |  |  |
| `spec.plannedLimits[].amount` | `string` |  |  |  |
| `spec.plannedLimits[].unit` | `string` | yes |  |  |
| `spec.autoAdjust` | `AwsBudgetAutoAdjust` |  |  |  |
| `spec.autoAdjust.autoAdjustType` | `string` |  |  |  |
| `spec.autoAdjust.budgetAdjustmentPeriod` | `int32` |  |  |  |
| `spec.timePeriodStart` | `string` |  |  |  |
| `spec.timePeriodEnd` | `string` |  |  |  |
| `spec.costTypes` | `AwsBudgetCostTypes` |  |  |  |
| `spec.costTypes.includeCredit` | `bool` |  |  |  |
| `spec.costTypes.includeDiscount` | `bool` |  |  |  |
| `spec.costTypes.includeOtherSubscription` | `bool` |  |  |  |
| `spec.costTypes.includeRecurring` | `bool` |  |  |  |
| `spec.costTypes.includeRefund` | `bool` |  |  |  |
| `spec.costTypes.includeSubscription` | `bool` |  |  |  |
| `spec.costTypes.includeSupport` | `bool` |  |  |  |
| `spec.costTypes.includeTax` | `bool` |  |  |  |
| `spec.costTypes.includeUpfront` | `bool` |  |  |  |
| `spec.costTypes.useAmortized` | `bool` |  |  |  |
| `spec.costTypes.useBlended` | `bool` |  |  |  |
| `spec.costFilters` | `[]AwsBudgetCostFilter` |  |  |  |
| `spec.costFilters[].name` | `string` | yes |  |  |
| `spec.costFilters[].values` | `[]string` | yes |  |  |
| `spec.metric` | `string` |  |  |  |
| `spec.filterExpression` | `AwsBudgetFilterExpression` |  |  |  |
| `spec.filterExpression.dimension` | `AwsBudgetFilterDimension` |  |  |  |
| `spec.filterExpression.dimension.key` | `string` |  |  |  |
| `spec.filterExpression.dimension.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.dimension.values` | `[]string` | yes |  |  |
| `spec.filterExpression.tag` | `AwsBudgetFilterTag` |  |  |  |
| `spec.filterExpression.tag.key` | `string` |  |  |  |
| `spec.filterExpression.tag.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.tag.values` | `[]string` |  |  |  |
| `spec.filterExpression.costCategory` | `AwsBudgetFilterCostCategory` |  |  |  |
| `spec.filterExpression.costCategory.key` | `string` |  |  |  |
| `spec.filterExpression.costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.costCategory.values` | `[]string` |  |  |  |
| `spec.filterExpression.and` | `[]AwsBudgetFilterExpressionNode` |  |  |  |
| `spec.filterExpression.and[].dimension` | `AwsBudgetFilterDimension` |  |  |  |
| `spec.filterExpression.and[].dimension.key` | `string` |  |  |  |
| `spec.filterExpression.and[].dimension.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.and[].dimension.values` | `[]string` | yes |  |  |
| `spec.filterExpression.and[].tag` | `AwsBudgetFilterTag` |  |  |  |
| `spec.filterExpression.and[].tag.key` | `string` |  |  |  |
| `spec.filterExpression.and[].tag.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.and[].tag.values` | `[]string` |  |  |  |
| `spec.filterExpression.and[].costCategory` | `AwsBudgetFilterCostCategory` |  |  |  |
| `spec.filterExpression.and[].costCategory.key` | `string` |  |  |  |
| `spec.filterExpression.and[].costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.and[].costCategory.values` | `[]string` |  |  |  |
| `spec.filterExpression.and[].and` | `[]AwsBudgetFilterExpressionLeaf` |  |  |  |
| `spec.filterExpression.and[].and[].dimension` | `AwsBudgetFilterDimension` |  |  |  |
| `spec.filterExpression.and[].and[].dimension.key` | `string` |  |  |  |
| `spec.filterExpression.and[].and[].dimension.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.and[].and[].dimension.values` | `[]string` | yes |  |  |
| `spec.filterExpression.and[].and[].tag` | `AwsBudgetFilterTag` |  |  |  |
| `spec.filterExpression.and[].and[].tag.key` | `string` |  |  |  |
| `spec.filterExpression.and[].and[].tag.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.and[].and[].tag.values` | `[]string` |  |  |  |
| `spec.filterExpression.and[].and[].costCategory` | `AwsBudgetFilterCostCategory` |  |  |  |
| `spec.filterExpression.and[].and[].costCategory.key` | `string` |  |  |  |
| `spec.filterExpression.and[].and[].costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.and[].and[].costCategory.values` | `[]string` |  |  |  |
| `spec.filterExpression.and[].or` | `[]AwsBudgetFilterExpressionLeaf` |  |  |  |
| `spec.filterExpression.and[].or[].dimension` | `AwsBudgetFilterDimension` |  |  |  |
| `spec.filterExpression.and[].or[].dimension.key` | `string` |  |  |  |
| `spec.filterExpression.and[].or[].dimension.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.and[].or[].dimension.values` | `[]string` | yes |  |  |
| `spec.filterExpression.and[].or[].tag` | `AwsBudgetFilterTag` |  |  |  |
| `spec.filterExpression.and[].or[].tag.key` | `string` |  |  |  |
| `spec.filterExpression.and[].or[].tag.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.and[].or[].tag.values` | `[]string` |  |  |  |
| `spec.filterExpression.and[].or[].costCategory` | `AwsBudgetFilterCostCategory` |  |  |  |
| `spec.filterExpression.and[].or[].costCategory.key` | `string` |  |  |  |
| `spec.filterExpression.and[].or[].costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.and[].or[].costCategory.values` | `[]string` |  |  |  |
| `spec.filterExpression.and[].not` | `AwsBudgetFilterExpressionLeaf` |  |  |  |
| `spec.filterExpression.and[].not.dimension` | `AwsBudgetFilterDimension` |  |  |  |
| `spec.filterExpression.and[].not.dimension.key` | `string` |  |  |  |
| `spec.filterExpression.and[].not.dimension.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.and[].not.dimension.values` | `[]string` | yes |  |  |
| `spec.filterExpression.and[].not.tag` | `AwsBudgetFilterTag` |  |  |  |
| `spec.filterExpression.and[].not.tag.key` | `string` |  |  |  |
| `spec.filterExpression.and[].not.tag.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.and[].not.tag.values` | `[]string` |  |  |  |
| `spec.filterExpression.and[].not.costCategory` | `AwsBudgetFilterCostCategory` |  |  |  |
| `spec.filterExpression.and[].not.costCategory.key` | `string` |  |  |  |
| `spec.filterExpression.and[].not.costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.and[].not.costCategory.values` | `[]string` |  |  |  |
| `spec.filterExpression.or` | `[]AwsBudgetFilterExpressionNode` |  |  |  |
| `spec.filterExpression.or[].dimension` | `AwsBudgetFilterDimension` |  |  |  |
| `spec.filterExpression.or[].dimension.key` | `string` |  |  |  |
| `spec.filterExpression.or[].dimension.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.or[].dimension.values` | `[]string` | yes |  |  |
| `spec.filterExpression.or[].tag` | `AwsBudgetFilterTag` |  |  |  |
| `spec.filterExpression.or[].tag.key` | `string` |  |  |  |
| `spec.filterExpression.or[].tag.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.or[].tag.values` | `[]string` |  |  |  |
| `spec.filterExpression.or[].costCategory` | `AwsBudgetFilterCostCategory` |  |  |  |
| `spec.filterExpression.or[].costCategory.key` | `string` |  |  |  |
| `spec.filterExpression.or[].costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.or[].costCategory.values` | `[]string` |  |  |  |
| `spec.filterExpression.or[].and` | `[]AwsBudgetFilterExpressionLeaf` |  |  |  |
| `spec.filterExpression.or[].and[].dimension` | `AwsBudgetFilterDimension` |  |  |  |
| `spec.filterExpression.or[].and[].dimension.key` | `string` |  |  |  |
| `spec.filterExpression.or[].and[].dimension.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.or[].and[].dimension.values` | `[]string` | yes |  |  |
| `spec.filterExpression.or[].and[].tag` | `AwsBudgetFilterTag` |  |  |  |
| `spec.filterExpression.or[].and[].tag.key` | `string` |  |  |  |
| `spec.filterExpression.or[].and[].tag.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.or[].and[].tag.values` | `[]string` |  |  |  |
| `spec.filterExpression.or[].and[].costCategory` | `AwsBudgetFilterCostCategory` |  |  |  |
| `spec.filterExpression.or[].and[].costCategory.key` | `string` |  |  |  |
| `spec.filterExpression.or[].and[].costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.or[].and[].costCategory.values` | `[]string` |  |  |  |
| `spec.filterExpression.or[].or` | `[]AwsBudgetFilterExpressionLeaf` |  |  |  |
| `spec.filterExpression.or[].or[].dimension` | `AwsBudgetFilterDimension` |  |  |  |
| `spec.filterExpression.or[].or[].dimension.key` | `string` |  |  |  |
| `spec.filterExpression.or[].or[].dimension.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.or[].or[].dimension.values` | `[]string` | yes |  |  |
| `spec.filterExpression.or[].or[].tag` | `AwsBudgetFilterTag` |  |  |  |
| `spec.filterExpression.or[].or[].tag.key` | `string` |  |  |  |
| `spec.filterExpression.or[].or[].tag.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.or[].or[].tag.values` | `[]string` |  |  |  |
| `spec.filterExpression.or[].or[].costCategory` | `AwsBudgetFilterCostCategory` |  |  |  |
| `spec.filterExpression.or[].or[].costCategory.key` | `string` |  |  |  |
| `spec.filterExpression.or[].or[].costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.or[].or[].costCategory.values` | `[]string` |  |  |  |
| `spec.filterExpression.or[].not` | `AwsBudgetFilterExpressionLeaf` |  |  |  |
| `spec.filterExpression.or[].not.dimension` | `AwsBudgetFilterDimension` |  |  |  |
| `spec.filterExpression.or[].not.dimension.key` | `string` |  |  |  |
| `spec.filterExpression.or[].not.dimension.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.or[].not.dimension.values` | `[]string` | yes |  |  |
| `spec.filterExpression.or[].not.tag` | `AwsBudgetFilterTag` |  |  |  |
| `spec.filterExpression.or[].not.tag.key` | `string` |  |  |  |
| `spec.filterExpression.or[].not.tag.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.or[].not.tag.values` | `[]string` |  |  |  |
| `spec.filterExpression.or[].not.costCategory` | `AwsBudgetFilterCostCategory` |  |  |  |
| `spec.filterExpression.or[].not.costCategory.key` | `string` |  |  |  |
| `spec.filterExpression.or[].not.costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.or[].not.costCategory.values` | `[]string` |  |  |  |
| `spec.filterExpression.not` | `AwsBudgetFilterExpressionNode` |  |  |  |
| `spec.filterExpression.not.dimension` | `AwsBudgetFilterDimension` |  |  |  |
| `spec.filterExpression.not.dimension.key` | `string` |  |  |  |
| `spec.filterExpression.not.dimension.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.not.dimension.values` | `[]string` | yes |  |  |
| `spec.filterExpression.not.tag` | `AwsBudgetFilterTag` |  |  |  |
| `spec.filterExpression.not.tag.key` | `string` |  |  |  |
| `spec.filterExpression.not.tag.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.not.tag.values` | `[]string` |  |  |  |
| `spec.filterExpression.not.costCategory` | `AwsBudgetFilterCostCategory` |  |  |  |
| `spec.filterExpression.not.costCategory.key` | `string` |  |  |  |
| `spec.filterExpression.not.costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.not.costCategory.values` | `[]string` |  |  |  |
| `spec.filterExpression.not.and` | `[]AwsBudgetFilterExpressionLeaf` |  |  |  |
| `spec.filterExpression.not.and[].dimension` | `AwsBudgetFilterDimension` |  |  |  |
| `spec.filterExpression.not.and[].dimension.key` | `string` |  |  |  |
| `spec.filterExpression.not.and[].dimension.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.not.and[].dimension.values` | `[]string` | yes |  |  |
| `spec.filterExpression.not.and[].tag` | `AwsBudgetFilterTag` |  |  |  |
| `spec.filterExpression.not.and[].tag.key` | `string` |  |  |  |
| `spec.filterExpression.not.and[].tag.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.not.and[].tag.values` | `[]string` |  |  |  |
| `spec.filterExpression.not.and[].costCategory` | `AwsBudgetFilterCostCategory` |  |  |  |
| `spec.filterExpression.not.and[].costCategory.key` | `string` |  |  |  |
| `spec.filterExpression.not.and[].costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.not.and[].costCategory.values` | `[]string` |  |  |  |
| `spec.filterExpression.not.or` | `[]AwsBudgetFilterExpressionLeaf` |  |  |  |
| `spec.filterExpression.not.or[].dimension` | `AwsBudgetFilterDimension` |  |  |  |
| `spec.filterExpression.not.or[].dimension.key` | `string` |  |  |  |
| `spec.filterExpression.not.or[].dimension.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.not.or[].dimension.values` | `[]string` | yes |  |  |
| `spec.filterExpression.not.or[].tag` | `AwsBudgetFilterTag` |  |  |  |
| `spec.filterExpression.not.or[].tag.key` | `string` |  |  |  |
| `spec.filterExpression.not.or[].tag.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.not.or[].tag.values` | `[]string` |  |  |  |
| `spec.filterExpression.not.or[].costCategory` | `AwsBudgetFilterCostCategory` |  |  |  |
| `spec.filterExpression.not.or[].costCategory.key` | `string` |  |  |  |
| `spec.filterExpression.not.or[].costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.not.or[].costCategory.values` | `[]string` |  |  |  |
| `spec.filterExpression.not.not` | `AwsBudgetFilterExpressionLeaf` |  |  |  |
| `spec.filterExpression.not.not.dimension` | `AwsBudgetFilterDimension` |  |  |  |
| `spec.filterExpression.not.not.dimension.key` | `string` |  |  |  |
| `spec.filterExpression.not.not.dimension.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.not.not.dimension.values` | `[]string` | yes |  |  |
| `spec.filterExpression.not.not.tag` | `AwsBudgetFilterTag` |  |  |  |
| `spec.filterExpression.not.not.tag.key` | `string` |  |  |  |
| `spec.filterExpression.not.not.tag.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.not.not.tag.values` | `[]string` |  |  |  |
| `spec.filterExpression.not.not.costCategory` | `AwsBudgetFilterCostCategory` |  |  |  |
| `spec.filterExpression.not.not.costCategory.key` | `string` |  |  |  |
| `spec.filterExpression.not.not.costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.filterExpression.not.not.costCategory.values` | `[]string` |  |  |  |
| `spec.notifications` | `[]AwsBudgetNotification` |  |  |  |
| `spec.notifications[].comparisonOperator` | `string` |  |  |  |
| `spec.notifications[].notificationType` | `string` |  |  |  |
| `spec.notifications[].threshold` | `double` |  |  |  |
| `spec.notifications[].thresholdType` | `string` |  |  |  |
| `spec.notifications[].subscriberEmailAddresses` | `[]string` |  |  |  |
| `spec.notifications[].subscriberSnsTopicArns` | `[]string \| valueFrom` |  |  | AwsSnsTopic (`status.outputs.topic_arn`) |
| `spec.actions` | `[]AwsBudgetAction` |  |  |  |
| `spec.actions[].name` | `string` | yes |  |  |
| `spec.actions[].actionType` | `string` |  |  |  |
| `spec.actions[].approvalModel` | `string` |  |  |  |
| `spec.actions[].notificationType` | `string` |  |  |  |
| `spec.actions[].executionRoleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.actions[].actionThreshold` | `AwsBudgetActionThreshold` | yes |  |  |
| `spec.actions[].actionThreshold.actionThresholdType` | `string` |  |  |  |
| `spec.actions[].actionThreshold.actionThresholdValue` | `double` |  |  |  |
| `spec.actions[].subscribers` | `[]AwsBudgetActionSubscriber` | yes |  |  |
| `spec.actions[].subscribers[].address` | `string \| valueFrom` | yes |  | AwsSnsTopic (`status.outputs.topic_arn`) |
| `spec.actions[].subscribers[].subscriptionType` | `string` |  |  |  |
| `spec.actions[].iamActionDefinition` | `AwsBudgetIamActionDefinition` |  |  |  |
| `spec.actions[].iamActionDefinition.policyArn` | `string \| valueFrom` | yes |  | AwsIamPolicy (`status.outputs.policy_arn`) |
| `spec.actions[].iamActionDefinition.groups` | `[]string \| valueFrom` |  |  | AwsIamGroup (`status.outputs.group_name`) |
| `spec.actions[].iamActionDefinition.roles` | `[]string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_name`) |
| `spec.actions[].iamActionDefinition.users` | `[]string \| valueFrom` |  |  | AwsIamUser (`status.outputs.user_name`) |
| `spec.actions[].scpActionDefinition` | `AwsBudgetScpActionDefinition` |  |  |  |
| `spec.actions[].scpActionDefinition.policyId` | `string \| valueFrom` | yes |  | AwsOrganizationPolicy (`status.outputs.policy_id`) |
| `spec.actions[].scpActionDefinition.targetIds` | `[]string` | yes |  |  |
| `spec.actions[].ssmActionDefinition` | `AwsBudgetSsmActionDefinition` |  |  |  |
| `spec.actions[].ssmActionDefinition.actionSubType` | `string` |  |  |  |
| `spec.actions[].ssmActionDefinition.region` | `string` | yes |  |  |
| `spec.actions[].ssmActionDefinition.instanceIds` | `[]string \| valueFrom` | yes |  | AwsEc2Instance (`status.outputs.instance_id`) |

## Field Details

### spec.region

`string` · required

The AWS region used by the provider while managing the budget.
Budgets is account-global (served from us-east-1) - every API call
still needs a regional endpoint. Example: "us-east-1".

- rule: {"string":{"minLen":"1"}}

### spec.budgetName

`string` · required

The budget's name in AWS - an explicit field because budget names
legally carry spaces and mixed case metadata.name cannot
("Monthly AWS Spend"). Any character except ":" and "\", up to
100. Changing the name replaces the budget.

- rule: {"string":{"minLen":"1","maxLen":"100","pattern":"^[^:\\\\]+$"}}

### spec.budgetType

`string`

What the budget tracks. COST tracks spend in a currency; USAGE
tracks usage quantities; the RI_* and SAVINGS_PLANS_* types track
reservation/Savings Plans utilization or coverage as a PERCENTAGE
(their limit unit). Changing the type replaces the budget.

- rule: {"string":{"in":["COST","USAGE","RI_UTILIZATION","RI_COVERAGE","SAVINGS_PLANS_UTILIZATION","SAVINGS_PLANS_COVERAGE"]}}

### spec.timeUnit

`string`

The budget's reset period. CUSTOM budgets run once between
time_period_start and time_period_end without resetting.

- rule: {"string":{"in":["DAILY","MONTHLY","QUARTERLY","ANNUALLY","CUSTOM"]}}

### spec.accountId

`string`

The 12-digit account the budget belongs to. Unset = the deploying
account. Set it from a management/payer account to manage a member
account's budget. Changing it replaces the budget.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]{12}$"}}

### spec.billingViewArn

`string`

ARN of a billing view to scope the budget to (AWS Billing
Conductor / custom billing views). Unset = the account's primary
billing view.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^arn:aws[a-zA-Z-]*:billing::[0-9]{12}:billingview/.+$"}}

### spec.limit

`AwsBudgetLimit`

The fixed spend/usage ceiling - the common funding shape. Mutually
exclusive with planned_limits and auto_adjust.

### spec.limit.amount

`string`

The ceiling as a decimal string (AWS compares budgets as
decimals - "100.0" and "100" are the same limit).

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?$"}}

### spec.limit.unit

`string` · required

The limit's unit: a currency code for COST budgets ("USD"), a
usage unit for USAGE budgets ("GB", "Hrs"), or "PERCENTAGE" for
the RI/Savings Plans types.

- rule: {"string":{"minLen":"1"}}

### spec.plannedLimits

`[]AwsBudgetPlannedLimit`

Per-period planned limits (a growth plan: each period gets its own
ceiling). AWS keys them by period start time; periods must align
with time_unit. Mutually exclusive with limit and auto_adjust.

### spec.plannedLimits[].startTime

`string`

The period this limit applies to, as "YYYY-MM-DD_HH:MM" (UTC).
Must align with the budget's time_unit period boundaries.

- rule: {"string":{"pattern":"^[0-9]{4}-[0-9]{2}-[0-9]{2}_[0-9]{2}:[0-9]{2}$"}}

### spec.plannedLimits[].amount

`string`

The period's ceiling as a decimal string.

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?$"}}

### spec.plannedLimits[].unit

`string` · required

The period's unit (see AwsBudgetLimit.unit).

- rule: {"string":{"minLen":"1"}}

### spec.autoAdjust

`AwsBudgetAutoAdjust`

Auto-adjustment: AWS recomputes the limit from history or
forecast. When set, AWS owns the limit - fixed and planned limits
are mutually exclusive with it.

### spec.autoAdjust.autoAdjustType

`string`

HISTORICAL adjusts from a lookback window of past periods;
FORECAST adjusts from AWS's spend forecast.

- rule: {"string":{"in":["HISTORICAL","FORECAST"]}}

### spec.autoAdjust.budgetAdjustmentPeriod

`int32`

For HISTORICAL: how many past periods (1-60) feed the
adjustment. AWS ignores it for FORECAST.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":60,"gte":1}}

### spec.timePeriodStart

`string`

When the budget starts tracking, as "YYYY-MM-DD_HH:MM" (AWS's
budgets timestamp format, UTC). Unset = AWS starts at the current
period's beginning.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]{4}-[0-9]{2}-[0-9]{2}_[0-9]{2}:[0-9]{2}$"}}

### spec.timePeriodEnd

`string`

When the budget stops tracking, same format. Unset = the
provider's far-future default (2087-06-15_00:00) - effectively
"never".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]{4}-[0-9]{2}-[0-9]{2}_[0-9]{2}:[0-9]{2}$"}}

### spec.costTypes

`AwsBudgetCostTypes`

Which cost components count toward a COST budget (legacy
generation - mutually exclusive with metric). Unset leaves every
AWS default in place: all include_* toggles default TRUE except
use_blended/use_amortized which default FALSE.

### spec.costTypes.includeCredit

`bool` · optional (explicit presence)

Count credits toward the budget. Unset = AWS default (true).

### spec.costTypes.includeDiscount

`bool` · optional (explicit presence)

Count discounts. Unset = AWS default (true).

### spec.costTypes.includeOtherSubscription

`bool` · optional (explicit presence)

Count non-RI subscription costs. Unset = AWS default (true).

### spec.costTypes.includeRecurring

`bool` · optional (explicit presence)

Count recurring fees (e.g. monthly RI fees). Unset = AWS default
(true).

### spec.costTypes.includeRefund

`bool` · optional (explicit presence)

Count refunds. Unset = AWS default (true).

### spec.costTypes.includeSubscription

`bool` · optional (explicit presence)

Count subscription costs. Unset = AWS default (true).

### spec.costTypes.includeSupport

`bool` · optional (explicit presence)

Count support fees. Unset = AWS default (true).

### spec.costTypes.includeTax

`bool` · optional (explicit presence)

Count taxes. Unset = AWS default (true).

### spec.costTypes.includeUpfront

`bool` · optional (explicit presence)

Count upfront RI costs. Unset = AWS default (true).

### spec.costTypes.useAmortized

`bool` · optional (explicit presence)

Measure amortized costs instead of cash-basis. Unset = AWS
default (false).

### spec.costTypes.useBlended

`bool` · optional (explicit presence)

Measure blended costs instead of unblended. Unset = AWS default
(false).

### spec.costFilters

`[]AwsBudgetCostFilter`

Legacy name/values filters (e.g. name "Service", values
["Amazon Elastic Compute Cloud - Compute"]). Multiple entries AND
together. Mutually exclusive with filter_expression.

### spec.costFilters[].name

`string` · required

The filter dimension name (e.g. "Service", "LinkedAccount",
"TagKeyValue" for "user:key$value" entries).

- rule: {"string":{"minLen":"1"}}

### spec.costFilters[].values

`[]string` · required

The values the dimension must match.

- rule: {"repeated":{"minItems":"1"}}

### spec.metric

`string`

The modern generation's measure. Must be paired with
filter_expression; mutually exclusive with cost_types.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["BlendedCost","UnblendedCost","AmortizedCost","NetUnblendedCost","NetAmortizedCost","UsageQuantity","NormalizedUsageAmount","Hours"]}}

### spec.filterExpression

`AwsBudgetFilterExpression`

The modern filter tree: dimensions, tag keys, and cost categories
composed with and/or/not. The leveled shape (root -> node -> leaf)
is exactly the nesting AWS accepts - two composition levels below
the root. Must be paired with metric.

### spec.filterExpression.dimension

`AwsBudgetFilterDimension`

A dimension leaf (service, region, account, purchase type, ...).

### spec.filterExpression.dimension.key

`string`

The dimension to filter on.

- rule: {"string":{"in":["AZ","INSTANCE_TYPE","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","OPERATION","PURCHASE_TYPE","REGION","SERVICE","SERVICE_CODE","USAGE_TYPE","USAGE_TYPE_GROUP","RECORD_TYPE","OPERATING_SYSTEM","TENANCY","SCOPE","PLATFORM","SUBSCRIPTION_ID","LEGAL_ENTITY_NAME","INVOICING_ENTITY","DEPLOYMENT_OPTION","DATABASE_ENGINE","CACHE_ENGINE","INSTANCE_TYPE_FAMILY","BILLING_ENTITY","RESERVATION_ID","RESOURCE_ID","RIGHTSIZING_TYPE","SAVINGS_PLANS_TYPE","SAVINGS_PLAN_ARN","PAYMENT_OPTION","RESERVATION_MODIFIED","TAG_KEY","COST_CATEGORY_NAME"]}}

### spec.filterExpression.dimension.matchOptions

`[]string`

How values match. Unset = EQUALS. ABSENT is deliberately not
offered: the provider rejects it for budgets at plan time.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.dimension.values

`[]string` · required

The values to match (each up to 1024 characters).

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.tag

`AwsBudgetFilterTag`

A cost-allocation-tag leaf.

### spec.filterExpression.tag.key

`string`

The tag key (up to 1024 characters).

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.tag.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.costCategory

`AwsBudgetFilterCostCategory`

A cost-category leaf.

### spec.filterExpression.costCategory.key

`string`

The cost category's name.

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.costCategory.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.and

`[]AwsBudgetFilterExpressionNode`

Children that must ALL match.

### spec.filterExpression.and[].dimension

`AwsBudgetFilterDimension`

A dimension leaf.

### spec.filterExpression.and[].dimension.key

`string`

The dimension to filter on.

- rule: {"string":{"in":["AZ","INSTANCE_TYPE","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","OPERATION","PURCHASE_TYPE","REGION","SERVICE","SERVICE_CODE","USAGE_TYPE","USAGE_TYPE_GROUP","RECORD_TYPE","OPERATING_SYSTEM","TENANCY","SCOPE","PLATFORM","SUBSCRIPTION_ID","LEGAL_ENTITY_NAME","INVOICING_ENTITY","DEPLOYMENT_OPTION","DATABASE_ENGINE","CACHE_ENGINE","INSTANCE_TYPE_FAMILY","BILLING_ENTITY","RESERVATION_ID","RESOURCE_ID","RIGHTSIZING_TYPE","SAVINGS_PLANS_TYPE","SAVINGS_PLAN_ARN","PAYMENT_OPTION","RESERVATION_MODIFIED","TAG_KEY","COST_CATEGORY_NAME"]}}

### spec.filterExpression.and[].dimension.matchOptions

`[]string`

How values match. Unset = EQUALS. ABSENT is deliberately not
offered: the provider rejects it for budgets at plan time.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.and[].dimension.values

`[]string` · required

The values to match (each up to 1024 characters).

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.and[].tag

`AwsBudgetFilterTag`

A cost-allocation-tag leaf.

### spec.filterExpression.and[].tag.key

`string`

The tag key (up to 1024 characters).

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.and[].tag.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.and[].tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.and[].costCategory

`AwsBudgetFilterCostCategory`

A cost-category leaf.

### spec.filterExpression.and[].costCategory.key

`string`

The cost category's name.

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.and[].costCategory.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.and[].costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.and[].and

`[]AwsBudgetFilterExpressionLeaf`

Leaf children that must ALL match.

### spec.filterExpression.and[].and[].dimension

`AwsBudgetFilterDimension`

A dimension leaf.

### spec.filterExpression.and[].and[].dimension.key

`string`

The dimension to filter on.

- rule: {"string":{"in":["AZ","INSTANCE_TYPE","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","OPERATION","PURCHASE_TYPE","REGION","SERVICE","SERVICE_CODE","USAGE_TYPE","USAGE_TYPE_GROUP","RECORD_TYPE","OPERATING_SYSTEM","TENANCY","SCOPE","PLATFORM","SUBSCRIPTION_ID","LEGAL_ENTITY_NAME","INVOICING_ENTITY","DEPLOYMENT_OPTION","DATABASE_ENGINE","CACHE_ENGINE","INSTANCE_TYPE_FAMILY","BILLING_ENTITY","RESERVATION_ID","RESOURCE_ID","RIGHTSIZING_TYPE","SAVINGS_PLANS_TYPE","SAVINGS_PLAN_ARN","PAYMENT_OPTION","RESERVATION_MODIFIED","TAG_KEY","COST_CATEGORY_NAME"]}}

### spec.filterExpression.and[].and[].dimension.matchOptions

`[]string`

How values match. Unset = EQUALS. ABSENT is deliberately not
offered: the provider rejects it for budgets at plan time.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.and[].and[].dimension.values

`[]string` · required

The values to match (each up to 1024 characters).

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.and[].and[].tag

`AwsBudgetFilterTag`

A cost-allocation-tag leaf.

### spec.filterExpression.and[].and[].tag.key

`string`

The tag key (up to 1024 characters).

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.and[].and[].tag.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.and[].and[].tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.and[].and[].costCategory

`AwsBudgetFilterCostCategory`

A cost-category leaf.

### spec.filterExpression.and[].and[].costCategory.key

`string`

The cost category's name.

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.and[].and[].costCategory.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.and[].and[].costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.and[].or

`[]AwsBudgetFilterExpressionLeaf`

Leaf children of which AT LEAST ONE must match.

### spec.filterExpression.and[].or[].dimension

`AwsBudgetFilterDimension`

A dimension leaf.

### spec.filterExpression.and[].or[].dimension.key

`string`

The dimension to filter on.

- rule: {"string":{"in":["AZ","INSTANCE_TYPE","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","OPERATION","PURCHASE_TYPE","REGION","SERVICE","SERVICE_CODE","USAGE_TYPE","USAGE_TYPE_GROUP","RECORD_TYPE","OPERATING_SYSTEM","TENANCY","SCOPE","PLATFORM","SUBSCRIPTION_ID","LEGAL_ENTITY_NAME","INVOICING_ENTITY","DEPLOYMENT_OPTION","DATABASE_ENGINE","CACHE_ENGINE","INSTANCE_TYPE_FAMILY","BILLING_ENTITY","RESERVATION_ID","RESOURCE_ID","RIGHTSIZING_TYPE","SAVINGS_PLANS_TYPE","SAVINGS_PLAN_ARN","PAYMENT_OPTION","RESERVATION_MODIFIED","TAG_KEY","COST_CATEGORY_NAME"]}}

### spec.filterExpression.and[].or[].dimension.matchOptions

`[]string`

How values match. Unset = EQUALS. ABSENT is deliberately not
offered: the provider rejects it for budgets at plan time.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.and[].or[].dimension.values

`[]string` · required

The values to match (each up to 1024 characters).

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.and[].or[].tag

`AwsBudgetFilterTag`

A cost-allocation-tag leaf.

### spec.filterExpression.and[].or[].tag.key

`string`

The tag key (up to 1024 characters).

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.and[].or[].tag.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.and[].or[].tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.and[].or[].costCategory

`AwsBudgetFilterCostCategory`

A cost-category leaf.

### spec.filterExpression.and[].or[].costCategory.key

`string`

The cost category's name.

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.and[].or[].costCategory.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.and[].or[].costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.and[].not

`AwsBudgetFilterExpressionLeaf`

A leaf child that must NOT match.

### spec.filterExpression.and[].not.dimension

`AwsBudgetFilterDimension`

A dimension leaf.

### spec.filterExpression.and[].not.dimension.key

`string`

The dimension to filter on.

- rule: {"string":{"in":["AZ","INSTANCE_TYPE","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","OPERATION","PURCHASE_TYPE","REGION","SERVICE","SERVICE_CODE","USAGE_TYPE","USAGE_TYPE_GROUP","RECORD_TYPE","OPERATING_SYSTEM","TENANCY","SCOPE","PLATFORM","SUBSCRIPTION_ID","LEGAL_ENTITY_NAME","INVOICING_ENTITY","DEPLOYMENT_OPTION","DATABASE_ENGINE","CACHE_ENGINE","INSTANCE_TYPE_FAMILY","BILLING_ENTITY","RESERVATION_ID","RESOURCE_ID","RIGHTSIZING_TYPE","SAVINGS_PLANS_TYPE","SAVINGS_PLAN_ARN","PAYMENT_OPTION","RESERVATION_MODIFIED","TAG_KEY","COST_CATEGORY_NAME"]}}

### spec.filterExpression.and[].not.dimension.matchOptions

`[]string`

How values match. Unset = EQUALS. ABSENT is deliberately not
offered: the provider rejects it for budgets at plan time.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.and[].not.dimension.values

`[]string` · required

The values to match (each up to 1024 characters).

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.and[].not.tag

`AwsBudgetFilterTag`

A cost-allocation-tag leaf.

### spec.filterExpression.and[].not.tag.key

`string`

The tag key (up to 1024 characters).

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.and[].not.tag.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.and[].not.tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.and[].not.costCategory

`AwsBudgetFilterCostCategory`

A cost-category leaf.

### spec.filterExpression.and[].not.costCategory.key

`string`

The cost category's name.

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.and[].not.costCategory.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.and[].not.costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.or

`[]AwsBudgetFilterExpressionNode`

Children of which AT LEAST ONE must match.

### spec.filterExpression.or[].dimension

`AwsBudgetFilterDimension`

A dimension leaf.

### spec.filterExpression.or[].dimension.key

`string`

The dimension to filter on.

- rule: {"string":{"in":["AZ","INSTANCE_TYPE","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","OPERATION","PURCHASE_TYPE","REGION","SERVICE","SERVICE_CODE","USAGE_TYPE","USAGE_TYPE_GROUP","RECORD_TYPE","OPERATING_SYSTEM","TENANCY","SCOPE","PLATFORM","SUBSCRIPTION_ID","LEGAL_ENTITY_NAME","INVOICING_ENTITY","DEPLOYMENT_OPTION","DATABASE_ENGINE","CACHE_ENGINE","INSTANCE_TYPE_FAMILY","BILLING_ENTITY","RESERVATION_ID","RESOURCE_ID","RIGHTSIZING_TYPE","SAVINGS_PLANS_TYPE","SAVINGS_PLAN_ARN","PAYMENT_OPTION","RESERVATION_MODIFIED","TAG_KEY","COST_CATEGORY_NAME"]}}

### spec.filterExpression.or[].dimension.matchOptions

`[]string`

How values match. Unset = EQUALS. ABSENT is deliberately not
offered: the provider rejects it for budgets at plan time.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.or[].dimension.values

`[]string` · required

The values to match (each up to 1024 characters).

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.or[].tag

`AwsBudgetFilterTag`

A cost-allocation-tag leaf.

### spec.filterExpression.or[].tag.key

`string`

The tag key (up to 1024 characters).

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.or[].tag.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.or[].tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.or[].costCategory

`AwsBudgetFilterCostCategory`

A cost-category leaf.

### spec.filterExpression.or[].costCategory.key

`string`

The cost category's name.

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.or[].costCategory.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.or[].costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.or[].and

`[]AwsBudgetFilterExpressionLeaf`

Leaf children that must ALL match.

### spec.filterExpression.or[].and[].dimension

`AwsBudgetFilterDimension`

A dimension leaf.

### spec.filterExpression.or[].and[].dimension.key

`string`

The dimension to filter on.

- rule: {"string":{"in":["AZ","INSTANCE_TYPE","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","OPERATION","PURCHASE_TYPE","REGION","SERVICE","SERVICE_CODE","USAGE_TYPE","USAGE_TYPE_GROUP","RECORD_TYPE","OPERATING_SYSTEM","TENANCY","SCOPE","PLATFORM","SUBSCRIPTION_ID","LEGAL_ENTITY_NAME","INVOICING_ENTITY","DEPLOYMENT_OPTION","DATABASE_ENGINE","CACHE_ENGINE","INSTANCE_TYPE_FAMILY","BILLING_ENTITY","RESERVATION_ID","RESOURCE_ID","RIGHTSIZING_TYPE","SAVINGS_PLANS_TYPE","SAVINGS_PLAN_ARN","PAYMENT_OPTION","RESERVATION_MODIFIED","TAG_KEY","COST_CATEGORY_NAME"]}}

### spec.filterExpression.or[].and[].dimension.matchOptions

`[]string`

How values match. Unset = EQUALS. ABSENT is deliberately not
offered: the provider rejects it for budgets at plan time.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.or[].and[].dimension.values

`[]string` · required

The values to match (each up to 1024 characters).

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.or[].and[].tag

`AwsBudgetFilterTag`

A cost-allocation-tag leaf.

### spec.filterExpression.or[].and[].tag.key

`string`

The tag key (up to 1024 characters).

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.or[].and[].tag.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.or[].and[].tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.or[].and[].costCategory

`AwsBudgetFilterCostCategory`

A cost-category leaf.

### spec.filterExpression.or[].and[].costCategory.key

`string`

The cost category's name.

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.or[].and[].costCategory.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.or[].and[].costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.or[].or

`[]AwsBudgetFilterExpressionLeaf`

Leaf children of which AT LEAST ONE must match.

### spec.filterExpression.or[].or[].dimension

`AwsBudgetFilterDimension`

A dimension leaf.

### spec.filterExpression.or[].or[].dimension.key

`string`

The dimension to filter on.

- rule: {"string":{"in":["AZ","INSTANCE_TYPE","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","OPERATION","PURCHASE_TYPE","REGION","SERVICE","SERVICE_CODE","USAGE_TYPE","USAGE_TYPE_GROUP","RECORD_TYPE","OPERATING_SYSTEM","TENANCY","SCOPE","PLATFORM","SUBSCRIPTION_ID","LEGAL_ENTITY_NAME","INVOICING_ENTITY","DEPLOYMENT_OPTION","DATABASE_ENGINE","CACHE_ENGINE","INSTANCE_TYPE_FAMILY","BILLING_ENTITY","RESERVATION_ID","RESOURCE_ID","RIGHTSIZING_TYPE","SAVINGS_PLANS_TYPE","SAVINGS_PLAN_ARN","PAYMENT_OPTION","RESERVATION_MODIFIED","TAG_KEY","COST_CATEGORY_NAME"]}}

### spec.filterExpression.or[].or[].dimension.matchOptions

`[]string`

How values match. Unset = EQUALS. ABSENT is deliberately not
offered: the provider rejects it for budgets at plan time.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.or[].or[].dimension.values

`[]string` · required

The values to match (each up to 1024 characters).

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.or[].or[].tag

`AwsBudgetFilterTag`

A cost-allocation-tag leaf.

### spec.filterExpression.or[].or[].tag.key

`string`

The tag key (up to 1024 characters).

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.or[].or[].tag.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.or[].or[].tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.or[].or[].costCategory

`AwsBudgetFilterCostCategory`

A cost-category leaf.

### spec.filterExpression.or[].or[].costCategory.key

`string`

The cost category's name.

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.or[].or[].costCategory.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.or[].or[].costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.or[].not

`AwsBudgetFilterExpressionLeaf`

A leaf child that must NOT match.

### spec.filterExpression.or[].not.dimension

`AwsBudgetFilterDimension`

A dimension leaf.

### spec.filterExpression.or[].not.dimension.key

`string`

The dimension to filter on.

- rule: {"string":{"in":["AZ","INSTANCE_TYPE","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","OPERATION","PURCHASE_TYPE","REGION","SERVICE","SERVICE_CODE","USAGE_TYPE","USAGE_TYPE_GROUP","RECORD_TYPE","OPERATING_SYSTEM","TENANCY","SCOPE","PLATFORM","SUBSCRIPTION_ID","LEGAL_ENTITY_NAME","INVOICING_ENTITY","DEPLOYMENT_OPTION","DATABASE_ENGINE","CACHE_ENGINE","INSTANCE_TYPE_FAMILY","BILLING_ENTITY","RESERVATION_ID","RESOURCE_ID","RIGHTSIZING_TYPE","SAVINGS_PLANS_TYPE","SAVINGS_PLAN_ARN","PAYMENT_OPTION","RESERVATION_MODIFIED","TAG_KEY","COST_CATEGORY_NAME"]}}

### spec.filterExpression.or[].not.dimension.matchOptions

`[]string`

How values match. Unset = EQUALS. ABSENT is deliberately not
offered: the provider rejects it for budgets at plan time.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.or[].not.dimension.values

`[]string` · required

The values to match (each up to 1024 characters).

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.or[].not.tag

`AwsBudgetFilterTag`

A cost-allocation-tag leaf.

### spec.filterExpression.or[].not.tag.key

`string`

The tag key (up to 1024 characters).

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.or[].not.tag.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.or[].not.tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.or[].not.costCategory

`AwsBudgetFilterCostCategory`

A cost-category leaf.

### spec.filterExpression.or[].not.costCategory.key

`string`

The cost category's name.

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.or[].not.costCategory.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.or[].not.costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.not

`AwsBudgetFilterExpressionNode`

A child that must NOT match.

### spec.filterExpression.not.dimension

`AwsBudgetFilterDimension`

A dimension leaf.

### spec.filterExpression.not.dimension.key

`string`

The dimension to filter on.

- rule: {"string":{"in":["AZ","INSTANCE_TYPE","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","OPERATION","PURCHASE_TYPE","REGION","SERVICE","SERVICE_CODE","USAGE_TYPE","USAGE_TYPE_GROUP","RECORD_TYPE","OPERATING_SYSTEM","TENANCY","SCOPE","PLATFORM","SUBSCRIPTION_ID","LEGAL_ENTITY_NAME","INVOICING_ENTITY","DEPLOYMENT_OPTION","DATABASE_ENGINE","CACHE_ENGINE","INSTANCE_TYPE_FAMILY","BILLING_ENTITY","RESERVATION_ID","RESOURCE_ID","RIGHTSIZING_TYPE","SAVINGS_PLANS_TYPE","SAVINGS_PLAN_ARN","PAYMENT_OPTION","RESERVATION_MODIFIED","TAG_KEY","COST_CATEGORY_NAME"]}}

### spec.filterExpression.not.dimension.matchOptions

`[]string`

How values match. Unset = EQUALS. ABSENT is deliberately not
offered: the provider rejects it for budgets at plan time.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.not.dimension.values

`[]string` · required

The values to match (each up to 1024 characters).

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.not.tag

`AwsBudgetFilterTag`

A cost-allocation-tag leaf.

### spec.filterExpression.not.tag.key

`string`

The tag key (up to 1024 characters).

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.not.tag.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.not.tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.not.costCategory

`AwsBudgetFilterCostCategory`

A cost-category leaf.

### spec.filterExpression.not.costCategory.key

`string`

The cost category's name.

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.not.costCategory.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.not.costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.not.and

`[]AwsBudgetFilterExpressionLeaf`

Leaf children that must ALL match.

### spec.filterExpression.not.and[].dimension

`AwsBudgetFilterDimension`

A dimension leaf.

### spec.filterExpression.not.and[].dimension.key

`string`

The dimension to filter on.

- rule: {"string":{"in":["AZ","INSTANCE_TYPE","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","OPERATION","PURCHASE_TYPE","REGION","SERVICE","SERVICE_CODE","USAGE_TYPE","USAGE_TYPE_GROUP","RECORD_TYPE","OPERATING_SYSTEM","TENANCY","SCOPE","PLATFORM","SUBSCRIPTION_ID","LEGAL_ENTITY_NAME","INVOICING_ENTITY","DEPLOYMENT_OPTION","DATABASE_ENGINE","CACHE_ENGINE","INSTANCE_TYPE_FAMILY","BILLING_ENTITY","RESERVATION_ID","RESOURCE_ID","RIGHTSIZING_TYPE","SAVINGS_PLANS_TYPE","SAVINGS_PLAN_ARN","PAYMENT_OPTION","RESERVATION_MODIFIED","TAG_KEY","COST_CATEGORY_NAME"]}}

### spec.filterExpression.not.and[].dimension.matchOptions

`[]string`

How values match. Unset = EQUALS. ABSENT is deliberately not
offered: the provider rejects it for budgets at plan time.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.not.and[].dimension.values

`[]string` · required

The values to match (each up to 1024 characters).

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.not.and[].tag

`AwsBudgetFilterTag`

A cost-allocation-tag leaf.

### spec.filterExpression.not.and[].tag.key

`string`

The tag key (up to 1024 characters).

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.not.and[].tag.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.not.and[].tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.not.and[].costCategory

`AwsBudgetFilterCostCategory`

A cost-category leaf.

### spec.filterExpression.not.and[].costCategory.key

`string`

The cost category's name.

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.not.and[].costCategory.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.not.and[].costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.not.or

`[]AwsBudgetFilterExpressionLeaf`

Leaf children of which AT LEAST ONE must match.

### spec.filterExpression.not.or[].dimension

`AwsBudgetFilterDimension`

A dimension leaf.

### spec.filterExpression.not.or[].dimension.key

`string`

The dimension to filter on.

- rule: {"string":{"in":["AZ","INSTANCE_TYPE","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","OPERATION","PURCHASE_TYPE","REGION","SERVICE","SERVICE_CODE","USAGE_TYPE","USAGE_TYPE_GROUP","RECORD_TYPE","OPERATING_SYSTEM","TENANCY","SCOPE","PLATFORM","SUBSCRIPTION_ID","LEGAL_ENTITY_NAME","INVOICING_ENTITY","DEPLOYMENT_OPTION","DATABASE_ENGINE","CACHE_ENGINE","INSTANCE_TYPE_FAMILY","BILLING_ENTITY","RESERVATION_ID","RESOURCE_ID","RIGHTSIZING_TYPE","SAVINGS_PLANS_TYPE","SAVINGS_PLAN_ARN","PAYMENT_OPTION","RESERVATION_MODIFIED","TAG_KEY","COST_CATEGORY_NAME"]}}

### spec.filterExpression.not.or[].dimension.matchOptions

`[]string`

How values match. Unset = EQUALS. ABSENT is deliberately not
offered: the provider rejects it for budgets at plan time.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.not.or[].dimension.values

`[]string` · required

The values to match (each up to 1024 characters).

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.not.or[].tag

`AwsBudgetFilterTag`

A cost-allocation-tag leaf.

### spec.filterExpression.not.or[].tag.key

`string`

The tag key (up to 1024 characters).

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.not.or[].tag.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.not.or[].tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.not.or[].costCategory

`AwsBudgetFilterCostCategory`

A cost-category leaf.

### spec.filterExpression.not.or[].costCategory.key

`string`

The cost category's name.

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.not.or[].costCategory.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.not.or[].costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.not.not

`AwsBudgetFilterExpressionLeaf`

A leaf child that must NOT match.

### spec.filterExpression.not.not.dimension

`AwsBudgetFilterDimension`

A dimension leaf.

### spec.filterExpression.not.not.dimension.key

`string`

The dimension to filter on.

- rule: {"string":{"in":["AZ","INSTANCE_TYPE","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","OPERATION","PURCHASE_TYPE","REGION","SERVICE","SERVICE_CODE","USAGE_TYPE","USAGE_TYPE_GROUP","RECORD_TYPE","OPERATING_SYSTEM","TENANCY","SCOPE","PLATFORM","SUBSCRIPTION_ID","LEGAL_ENTITY_NAME","INVOICING_ENTITY","DEPLOYMENT_OPTION","DATABASE_ENGINE","CACHE_ENGINE","INSTANCE_TYPE_FAMILY","BILLING_ENTITY","RESERVATION_ID","RESOURCE_ID","RIGHTSIZING_TYPE","SAVINGS_PLANS_TYPE","SAVINGS_PLAN_ARN","PAYMENT_OPTION","RESERVATION_MODIFIED","TAG_KEY","COST_CATEGORY_NAME"]}}

### spec.filterExpression.not.not.dimension.matchOptions

`[]string`

How values match. Unset = EQUALS. ABSENT is deliberately not
offered: the provider rejects it for budgets at plan time.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.not.not.dimension.values

`[]string` · required

The values to match (each up to 1024 characters).

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.not.not.tag

`AwsBudgetFilterTag`

A cost-allocation-tag leaf.

### spec.filterExpression.not.not.tag.key

`string`

The tag key (up to 1024 characters).

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.not.not.tag.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.not.not.tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.filterExpression.not.not.costCategory

`AwsBudgetFilterCostCategory`

A cost-category leaf.

### spec.filterExpression.not.not.costCategory.key

`string`

The cost category's name.

- rule: {"string":{"maxLen":"1024"}}

### spec.filterExpression.not.not.costCategory.matchOptions

`[]string`

How values match (see AwsBudgetFilterDimension.match_options;
ABSENT is rejected by the provider for budgets).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.filterExpression.not.not.costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.notifications

`[]AwsBudgetNotification`

Threshold alerts on the budget. Each notification compares actual
or forecasted spend against a threshold and alerts its
subscribers; AWS requires at least one subscriber per
notification.

- rule: a notification needs at least one subscriber - add an email address or an SNS topic

### spec.notifications[].comparisonOperator

`string`

How spend compares against the threshold.

- rule: {"string":{"in":["GREATER_THAN","LESS_THAN","EQUAL_TO"]}}

### spec.notifications[].notificationType

`string`

Alert on ACTUAL spend or AWS's FORECASTED spend.

- rule: {"string":{"in":["ACTUAL","FORECASTED"]}}

### spec.notifications[].threshold

`double`

The threshold value: a percentage of the budget limit
(threshold_type PERCENTAGE, e.g. 80) or an absolute amount
(ABSOLUTE_VALUE).

- rule: {"double":{"gte":0}}

### spec.notifications[].thresholdType

`string`

How the threshold is interpreted.

- rule: {"string":{"in":["PERCENTAGE","ABSOLUTE_VALUE"]}}

### spec.notifications[].subscriberEmailAddresses

`[]string`

Email addresses to alert.

- rule: {"repeated":{"items":{"string":{"email":true}}}}

### spec.notifications[].subscriberSnsTopicArns

`[]string | valueFrom`

SNS topics to alert. Reference an AwsSnsTopic's topic_arn output
or pass a literal topic ARN. The topic's policy must allow
budgets.amazonaws.com to publish or alerts silently never arrive.

- references: AwsSnsTopic (`status.outputs.topic_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSnsTopic, name: <that resource's name>, fieldPath: status.outputs.topic_arn}} -- a bare string does not parse

### spec.actions

`[]AwsBudgetAction`

Budget ACTIONS - the folded satellites: each entry is one
aws_budgets_budget_action keyed by its name. An action executes
(or stages for approval) a response when its threshold breaches:
applying a restrictive IAM policy, attaching an SCP, or stopping
EC2/RDS instances via SSM.

- rule: iam_action_definition is set exactly when action_type is APPLY_IAM_POLICY
- rule: scp_action_definition is set exactly when action_type is APPLY_SCP_POLICY
- rule: ssm_action_definition is set exactly when action_type is RUN_SSM_DOCUMENTS

### spec.actions[].name

`string` · required

The action's name within this budget - the key the modules use
for the for_each entry and the outputs map. Planton-side only;
AWS identifies actions by a generated action ID.

- rule: {"string":{"minLen":"1","maxLen":"63"}}

### spec.actions[].actionType

`string`

What the action does when triggered: apply a restrictive IAM
policy to groups/roles/users, attach an SCP to organization
targets, or run SSM documents that stop EC2/RDS instances.
Changing the type replaces the action.

- rule: {"string":{"in":["APPLY_IAM_POLICY","APPLY_SCP_POLICY","RUN_SSM_DOCUMENTS"]}}

### spec.actions[].approvalModel

`string`

AUTOMATIC executes the action on breach; MANUAL stages it for a
human to approve in the console.

- rule: {"string":{"in":["AUTOMATIC","MANUAL"]}}

### spec.actions[].notificationType

`string`

Trigger on ACTUAL or FORECASTED spend.

- rule: {"string":{"in":["ACTUAL","FORECASTED"]}}

### spec.actions[].executionRoleArn

`string | valueFrom` · required

The role Budgets assumes to execute the action. Its trust policy
must allow budgets.amazonaws.com to assume it, and it needs the
permissions the definition arm implies (iam:AttachGroupPolicy,
organizations:AttachPolicy, ssm:StartAutomationExecution, ...).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.actions[].actionThreshold

`AwsBudgetActionThreshold` · required

The threshold that triggers the action.

- rule: {"required":true}

### spec.actions[].actionThreshold.actionThresholdType

`string`

PERCENTAGE of the budget limit, or ABSOLUTE_VALUE spend.

- rule: {"string":{"in":["PERCENTAGE","ABSOLUTE_VALUE"]}}

### spec.actions[].actionThreshold.actionThresholdValue

`double`

The threshold value.

- rule: {"double":{"lte":40000000000,"gte":0}}

### spec.actions[].subscribers

`[]AwsBudgetActionSubscriber` · required

Who is told when the action triggers/executes. AWS caps
subscribers at 11 per action.

- rule: {"repeated":{"minItems":"1","maxItems":"11"}}

### spec.actions[].subscribers[].address

`string | valueFrom` · required

An email address (subscription_type EMAIL) or an SNS topic ARN
(subscription_type SNS - reference an AwsSnsTopic's topic_arn
output or pass a literal ARN).

- references: AwsSnsTopic (`status.outputs.topic_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSnsTopic, name: <that resource's name>, fieldPath: status.outputs.topic_arn}} -- a bare string does not parse

### spec.actions[].subscribers[].subscriptionType

`string`

How the subscriber is reached.

- rule: {"string":{"in":["EMAIL","SNS"]}}

### spec.actions[].iamActionDefinition

`AwsBudgetIamActionDefinition`

The IAM arm: attach a restrictive policy to principals.

- rule: name at least one group, role, or user for the policy to attach to

### spec.actions[].iamActionDefinition.policyArn

`string | valueFrom` · required

The managed policy to attach (typically a deny-heavy guardrail
policy). Reference an AwsIamPolicy's policy_arn output or pass a
literal ARN (AWS-managed policies included).

- references: AwsIamPolicy (`status.outputs.policy_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamPolicy, name: <that resource's name>, fieldPath: status.outputs.policy_arn}} -- a bare string does not parse

### spec.actions[].iamActionDefinition.groups

`[]string | valueFrom`

Group names the policy attaches to (up to 100). Reference an
AwsIamGroup's group_name output or pass literal names.

- references: AwsIamGroup (`status.outputs.group_name`)
- rule: {"repeated":{"maxItems":"100"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamGroup, name: <that resource's name>, fieldPath: status.outputs.group_name}} -- a bare string does not parse

### spec.actions[].iamActionDefinition.roles

`[]string | valueFrom`

Role names the policy attaches to (up to 100).

- references: AwsIamRole (`status.outputs.role_name`)
- rule: {"repeated":{"maxItems":"100"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_name}} -- a bare string does not parse

### spec.actions[].iamActionDefinition.users

`[]string | valueFrom`

User names the policy attaches to (up to 100).

- references: AwsIamUser (`status.outputs.user_name`)
- rule: {"repeated":{"maxItems":"100"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamUser, name: <that resource's name>, fieldPath: status.outputs.user_name}} -- a bare string does not parse

### spec.actions[].scpActionDefinition

`AwsBudgetScpActionDefinition`

The SCP arm: attach a service control policy to org targets.

### spec.actions[].scpActionDefinition.policyId

`string | valueFrom` · required

The SCP to attach. Reference an AwsOrganizationPolicy's policy_id
output or pass a literal policy ID ("p-...").

- references: AwsOrganizationPolicy (`status.outputs.policy_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsOrganizationPolicy, name: <that resource's name>, fieldPath: status.outputs.policy_id}} -- a bare string does not parse

### spec.actions[].scpActionDefinition.targetIds

`[]string` · required

The roots, OUs, or accounts the SCP attaches to (up to 100).

- rule: {"repeated":{"minItems":"1","maxItems":"100"}}

### spec.actions[].ssmActionDefinition

`AwsBudgetSsmActionDefinition`

The SSM arm: stop EC2 or RDS instances.

### spec.actions[].ssmActionDefinition.actionSubType

`string`

Stop EC2 or RDS instances.

- rule: {"string":{"in":["STOP_EC2_INSTANCES","STOP_RDS_INSTANCES"]}}

### spec.actions[].ssmActionDefinition.region

`string` · required

The region the instances live in (SSM actions are regional even
though the budget is global).

- rule: {"string":{"minLen":"1"}}

### spec.actions[].ssmActionDefinition.instanceIds

`[]string | valueFrom` · required

The instances to stop (up to 100). For EC2, reference an
AwsEc2Instance's instance_id output or pass literal IDs.

- references: AwsEc2Instance (`status.outputs.instance_id`)
- rule: {"repeated":{"minItems":"1","maxItems":"100"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsEc2Instance, name: <that resource's name>, fieldPath: status.outputs.instance_id}} -- a bare string does not parse

## Validation Rules

- `spec.exactly_one_funding_shape`: configure exactly one of limit, planned_limits, and auto_adjust - AWS budgets carry one funding shape
- `spec.one_filter_generation`: cost_filters (legacy) and filter_expression (modern) are mutually exclusive - pick one filtering generation
- `spec.metric_pairs_with_filter_expression`: metric and filter_expression must be set together - the modern filtering generation is a pair
- `spec.metric_conflicts_with_cost_types`: metric and cost_types are mutually exclusive - the modern metric already selects what the budget measures
- `spec.action_names_unique`: actions entries must have unique names

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsBudget, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.budget_name` | `string` | The budget's name (with the account ID, the provider's import ID: "account_id:budget_name"). |
| `status.outputs.budget_arn` | `string` | The budget's ARN. |
| `status.outputs.account_id` | `string` | The account the budget belongs to (the deploying account unless spec.account_id targeted a member account). |
| `status.outputs.action_ids` | `map<string, string>` | AWS-generated action IDs keyed by action name (each action imports as "account_id:action_id:budget_name"). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.notifications[].subscriberSnsTopicArns` | AwsSnsTopic | `status.outputs.topic_arn` |
| `spec.actions[].executionRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.actions[].subscribers[].address` | AwsSnsTopic | `status.outputs.topic_arn` |
| `spec.actions[].iamActionDefinition.policyArn` | AwsIamPolicy | `status.outputs.policy_arn` |
| `spec.actions[].iamActionDefinition.groups` | AwsIamGroup | `status.outputs.group_name` |
| `spec.actions[].iamActionDefinition.roles` | AwsIamRole | `status.outputs.role_name` |
| `spec.actions[].iamActionDefinition.users` | AwsIamUser | `status.outputs.user_name` |
| `spec.actions[].scpActionDefinition.policyId` | AwsOrganizationPolicy | `status.outputs.policy_id` |
| `spec.actions[].ssmActionDefinition.instanceIds` | AwsEc2Instance | `status.outputs.instance_id` |

## See Also

- [Overview](../README.md)

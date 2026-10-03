# AWS Budget

Deploys an AWS Budgets budget: a spend or usage threshold AWS evaluates continuously, with alerts when actual or forecasted spend crosses it and optional budget actions — apply a restrictive IAM policy, attach an SCP, or stop EC2/RDS instances — when it breaches. The budget carries exactly one funding shape (a fixed limit, per-period planned limits, or an auto-adjusting limit AWS recomputes from history or forecast) and scopes what it watches through a filter tree over services, regions, accounts, tags, and cost categories. Budgets is an account-global service; the spec's region is only the provider's API endpoint.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Budget** — the tracked type (cost, usage, or RI/Savings Plans utilization and coverage), its reset period, funding shape, filters, and threshold notifications.
- **Budget Actions** — one per `actions` entry: a staged or automatic response bound to its own threshold, each getting an AWS-generated action ID (echoed in the `action_ids` output). Created only when `actions` is set.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with Budgets permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **An execution role** (only for `actions`) — a role whose trust policy allows `budgets.amazonaws.com` to assume it, carrying the permissions its action arm implies; AWS's managed `AWSBudgetsActionsWithAWSResourceControlAccess` policy is the canonical grant.
- **Activated cost-allocation tags** (only for tag-based filters) — a tag filter expression matches nothing until the tag key is activated in the Billing console.
- **An organization with SCPs enabled** (only for the SCP action arm) — a management-account capability.

## Deploy

### Console

Open the deployment store, find **AWS Budget**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields covering the budget type, funding shape, filters, alerts, and actions. Start from the **Monthly Cost Guardrail** preset in the [Presets](#presets) tab for the first budget every account should carry.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBudget
metadata:
  name: monthly-cost-guardrail
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  budgetName: Monthly Cost Guardrail
  budgetType: COST
  timeUnit: MONTHLY
  limit:
    amount: "1000"
    unit: USD
  notifications:
    - comparisonOperator: GREATER_THAN
      notificationType: ACTUAL
      threshold: 80
      thresholdType: PERCENTAGE
      subscriberEmailAddresses:
        - finops@acme-corp.com
    - comparisonOperator: GREATER_THAN
      notificationType: FORECASTED
      threshold: 100
      thresholdType: PERCENTAGE
      subscriberEmailAddresses:
        - finops@acme-corp.com
```

```shell
planton apply -f budget.yaml
```

This creates a monthly cost budget with a fixed 1000 USD ceiling, alerting the FinOps team at 80% of actual spend and again when AWS forecasts the month will breach. An Infra Job tracks the provisioning in real time.

### InfraChart

When alerts should fan out through an SNS topic deployed in the same chart, wire the subscriber via ValueFromRef:

```yaml
spec:
  region: us-east-1
  budgetName: Platform Monthly Spend
  budgetType: COST
  timeUnit: MONTHLY
  limit:
    amount: "5000"
    unit: USD
  notifications:
    - comparisonOperator: GREATER_THAN
      notificationType: ACTUAL
      threshold: 90
      thresholdType: PERCENTAGE
      subscriberSnsTopicArns:
        - valueFrom:
            kind: AwsSnsTopic
            name: cost-alerts
            fieldPath: status.outputs.topic_arn
```

The InfraPipeline resolves the dependency graph, deploys the topic first, then creates the budget publishing to it.

## Key Configuration

These are the most important decisions when configuring a budget. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Pick one funding shape** — a fixed `limit`, per-period `plannedLimits` (a growth plan where each period gets its own ceiling), or `autoAdjust` (AWS recomputes the limit from a historical lookback or its forecast). Exactly one is legal, validated at manifest time. Auto-adjustment means AWS owns the number — alerts stay meaningful as spend grows, but the budget stops encoding an intentional ceiling.

**Two filter generations, never both** — the modern `metric` + `filterExpression` pair (dimensions, tag keys, and cost categories composed with and/or/not) or the legacy `costFilters`/`costTypes`. Prefer the modern pair: it selects the measure (`UnblendedCost`, `AmortizedCost`, ...) explicitly and composes filters the legacy generation cannot. The tree is leveled exactly as AWS accepts it — composition at most two levels below the root — so deeper nesting is inexpressible rather than rejected at apply.

**Tag filters need activated tags** — a `filterExpression` tag leaf only matches cost-allocation tag keys activated in Billing. An unactivated key does not error; it silently matches nothing, and the budget tracks zero.

**FORECASTED alerts buy reaction time** — an ACTUAL threshold fires after the money is spent; a FORECASTED threshold at 100% fires when AWS projects the breach, usually days earlier. The guardrail pattern uses both: actual at 80% as the early warning, forecasted at 100% as the escalation.

**SNS subscribers can silently fail** — the topic's policy must allow `budgets.amazonaws.com` to publish, or alerts never arrive and nothing reports the loss. Email subscribers have no such trap.

**AUTOMATIC actions execute and reverse; MANUAL actions stage** — an AUTOMATIC action fires on breach and reverses when spend drops back under the threshold (IAM policies and SCPs detach; SSM instance stops do NOT restart instances). MANUAL stages the action in the console for a human to approve. Start MANUAL in production accounts; go AUTOMATIC where a hard stop beats a paged human.

**RI and Savings Plans budgets take PERCENTAGE limits** — a dollar limit on `RI_UTILIZATION` is rejected by AWS; the limit unit is `PERCENTAGE` (typically 100).

**The name is identity** — `budgetName` is an explicit field because budget names legally carry spaces and mixed case; changing it replaces the budget and its alert history.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSnsTopic** | `notifications[].subscriberSnsTopicArns` | `status.outputs.topic_arn` |
| **AwsIamRole** | `actions[].executionRoleArn` | `status.outputs.role_arn` |
| **AwsIamPolicy** | `actions[].iamActionDefinition.policyArn` | `status.outputs.policy_arn` |
| **AwsOrganizationPolicy** | `actions[].scpActionDefinition.policyId` | `status.outputs.policy_id` |
| **AwsEc2Instance** | `actions[].ssmActionDefinition.instanceIds` | `status.outputs.instance_id` |

The IAM action arm's principal lists (`groups`, `roles`, `users`) and action subscribers also accept references to the matching IAM and SNS kinds.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `budget_name` | The budget's name (with the account ID, the provider's import ID) | Addressing the budget in CLI/API operations |
| `budget_arn` | The budget's ARN | IAM policy statements scoping who may edit the budget or its actions |

`account_id` and `action_ids` are also echoed — the owning account and each action's AWS-generated ID keyed by action name. They exist for import addressing rather than as composition inputs.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**The monthly guardrail** — a fixed monthly ceiling, actual-spend alert at 80%, forecast alert at 100%, humans on the list. Every account's first budget; scope later with a filter expression as team budgets emerge. Start from the **Monthly Cost Guardrail** preset.

**The self-enforcing sandbox** — a tag-scoped budget whose AUTOMATIC action applies a deny policy to the sandbox users group at 100%: provisioning freezes until spend resets or a human reverses it. The trade is bluntness — the freeze hits everything the policy's principals do, so scope the principal lists to genuinely expendable identities. Start from the **Enforced Sandbox Budget** preset.

**Utilization watchdogs** — `RI_UTILIZATION` or `SAVINGS_PLANS_UTILIZATION` budgets with PERCENTAGE limits alerting when committed spend sits idle. These catch the quiet waste that cost budgets structurally cannot: money already committed but not being used.

## Works With

- [**AWS SNS Topic**](/infra-catalog/aws-sns-topic) — programmatic fan-out for threshold alerts and action notices
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the execution role Budgets assumes to run actions
- [**AWS IAM Policy**](/infra-catalog/aws-iam-policy) — the restrictive policy the IAM action arm attaches on breach
- [**AWS Organization Policy**](/infra-catalog/aws-organization-policy) — the SCP the SCP action arm attaches to organization targets
- [**AWS EC2 Instance**](/infra-catalog/aws-ec2-instance) — the instances the SSM action arm stops
- [**AWS Cost Category**](/infra-catalog/aws-cost-category) — team/project groupings the filter expression can scope budgets to
- [**AWS Cost Anomaly Monitor**](/infra-catalog/aws-cost-anomaly-monitor) — the complementary detector for unusual spend patterns a fixed threshold misses

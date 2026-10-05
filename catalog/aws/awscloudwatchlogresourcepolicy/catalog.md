# AWS CloudWatch Logs Resource Policy

Deploys a CloudWatch Logs resource policy — the IAM document AWS services need before they can write logs into your account: Route53 query logging, EventBridge delivery to log groups, OpenSearch slow logs, and every other service whose setup docs ask for a "CloudWatch Logs resource policy". The policy has exactly one scope: account (a named policy applying account-wide in the region — the common shape) or resource (pinned to one log group's ARN). Updates are guarded by AWS's revision ID, so concurrent edits fail loudly instead of silently overwriting each other.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **CloudWatch Logs Resource Policy** — the policy document granting service principals (`route53.amazonaws.com`, `events.amazonaws.com`, …) `logs:CreateLogStream` and `logs:PutLogEvents` on the target log groups, at the account scope (named by `policyName`) or the resource scope (pinned by `resourceArn`).

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with credentials for the target AWS account, carrying `logs:PutResourcePolicy`, `logs:DescribeResourcePolicies`, and `logs:DeleteResourcePolicy`. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **Quota headroom** — AWS caps account-scope resource policies at 10 per region, and the quota fills fast when every stack ships its own grant. Prefer one shared policy per service class over many near-duplicates.
- **The target log group** (only for resource scope) — the group whose ARN pins the policy; reference an AwsCloudwatchLogGroup or pass a literal ARN.

## Deploy

### Console

Open the deployment store, find **AWS CloudWatch Logs Resource Policy**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec fields — region, scope, and the policy document. Start from the **Route53 Query Logging Grant** or **EventBridge Log Delivery Grant** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudwatchLogResourcePolicy
metadata:
  name: route53-query-logging
  org: acme-corp
  env: prod
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

```shell
planton apply -f log-resource-policy.yaml
```

This creates the account-scope grant letting Route53 write query logs into any group under `/aws/route53/` — one policy covers every zone that logs there (Route53 query logs land only in us-east-1, so this policy deploys there). An Infra Job tracks the provisioning in real time.

### InfraChart

When a resource-scoped policy deploys alongside its log group in one chart, wire the group reference via ValueFromRef:

```yaml
spec:
  region: us-east-1
  resourceArn:
    valueFrom:
      kind: AwsCloudwatchLogGroup
      name: events-audit-log
      fieldPath: status.outputs.log_group_arn
  policyDocument:
    Version: "2012-10-17"
    Statement:
      - Effect: Allow
        Principal:
          Service: events.amazonaws.com
        Action:
          - logs:CreateLogStream
          - logs:PutLogEvents
        Resource: "*"
```

The InfraPipeline resolves the dependency graph, deploys the log group first, then attaches the policy to it.

## Key Configuration

These are the most important decisions when configuring a resource policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Scope choice** — set exactly one of `policyName` (account scope) and `resourceArn` (resource scope). Account scope is right for service grants covering a path pattern (`/aws/route53/*`); resource scope is for a grant that must live and die with a single group — rare, and it makes the group's ARN the policy's identity. Both fields are identity: changing either replaces the policy. Only the document updates in place.

**Consolidate against the quota** — with 10 account-scope policies per region, ship one policy per service class (`route53-query-logging`, `eventbridge-delivery`) with Resource patterns wide enough for the class, owned by one instance — not one grant per stack.

**The revision guard is your friend** — updates send the last-seen revision ID; if someone edited the policy in the console since, the apply fails with a revision mismatch instead of overwriting their change. Re-plan (refreshing state) and apply again to take ownership of the merged truth.

**Resource scope needs its revision to delete** — AWS refuses a resource-scoped delete without the current revision ID, which lives in state. Never hand-clear state for this kind; a lost revision means deleting the policy with the AWS CLI before the next apply.

**The grant takes effect immediately** — once applied, the named service principal can create log streams and write events; the service feature that needed the grant (query logging, rule targets) starts working without further action.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsCloudwatchLogGroup** (resource scope only) | `resourceArn` | `status.outputs.log_group_arn` |

### What This Kind Provides

After provisioning, `status.outputs` records the policy's operational identity: `policy_id` (the name or the target ARN — also the provider's import ID), `policy_scope` (ACCOUNT or RESOURCE as AWS recorded it), and `revision_id` (the optimistic-concurrency token guarding the next update). These are audit and import records — nothing downstream composes on a resource policy via ValueFromRef.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Route53 query logging grant** — the account-scope policy Route53 requires before zone query logging works, deployed in us-east-1 (the only region query logs land in), covering every zone under `/aws/route53/*`. Start from the **Route53 Query Logging Grant** preset.

**EventBridge log delivery grant** — lets EventBridge rules target log groups under `/aws/events/*` — the grant the "target a log group" feature silently requires. Start from the **EventBridge Log Delivery Grant** preset.

**One grant per service class** — a single account-scope policy whose Resource pattern covers the class's whole log path, owned by one instance and reused by every stack that needs the service. The shape the 10-per-region quota forces — and the right one anyway.

## Works With

- [**AWS CloudWatch Log Group**](/infra-catalog/aws-cloudwatch-log-group) — the groups the granted services write into; resource-scoped policies pin to one group's ARN
- [**AWS Route 53 Zone**](/infra-catalog/aws-route53-zone) — zone query logging is the classic consumer of this grant

# AWS REST API Usage Plan

Deploys an API Gateway usage plan with its API keys — the metering and throttling layer for REST API consumers. Each plan covers one or more REST API stages, sets quota (requests per day, week, or month) and throttle ceilings including per-method throttles, and admits callers through API keys created and attached by the same resource. A plan spans APIs and stages, which is why it is its own component rather than part of the REST API; routes opt in with `apiKeyRequired` on the AWS REST API Gateway component.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Usage Plan** — covering the named REST API stages, with optional quota over a calendar period and plan-wide throttle ceilings; per-method throttles tighten specific paths within a covered stage
- **API Key** — one per `apiKeys` entry, created in AWS and attached to the plan in the same apply, so the attachment cannot be forgotten. Key values are secrets and are not exported as outputs

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with API Gateway usage-plan permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- The REST APIs and stages the plan will cover, already deployed — referencing the AWS REST API Gateway `stage_name` output wires the ordering
- `apiKeyRequired: true` on the routes that should demand a key (set on the AWS REST API Gateway component)

## Deploy

### Console

Open the deployment store, find **AWS REST API Usage Plan**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, the covered stages, and the quota, throttle, and key settings. Start from the **Metered API Keys** preset in the [Presets](#presets) tab for a daily-quota consumer tier, or the **Throttled Stages** preset for rate ceilings with a tighter cap on the expensive method.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRestApiUsagePlan
metadata:
  name: partner-tier
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  description: Partner tier -- 1000 requests/day
  apiStages:
    - restApiId:
        valueFrom:
          kind: AwsRestApiGateway
          name: orders-api
          fieldPath: status.outputs.rest_api_id
      stageName:
        value: prod
  quota:
    limit: 1000
    period: DAY
  apiKeys:
    - name: partner-mobile-app
      description: The partner's mobile app
      enabled: true
```

```shell
planton apply -f usage-plan.yaml
```

This creates a usage plan covering the orders API's `prod` stage with a 1,000-requests-per-day quota and one AWS-generated API key attached — the key value is read from AWS when distributing it, never from outputs. An Infra Job tracks the provisioning in real time.

### InfraChart

When the plan deploys alongside the API it meters in one chart, wire the stage coverage via ValueFromRef:

```yaml
spec:
  region: us-west-2
  apiStages:
    - restApiId:
        valueFrom:
          kind: AwsRestApiGateway
          name: orders-api
          fieldPath: status.outputs.rest_api_id
      stageName:
        valueFrom:
          kind: AwsRestApiGateway
          name: orders-api
          fieldPath: status.outputs.stage_name
  quota:
    limit: 1000
    period: DAY
```

The InfraPipeline resolves the dependency graph, deploys the API and its stage first, then attaches the plan to it.

## Key Configuration

These are the most important decisions when configuring a usage plan. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Keys are not authentication** — an API key identifies a consumer for metering; anyone who sees the key is that consumer. Pair `apiKeyRequired` routes with IAM, Cognito, or a Lambda authorizer on the API — the key meters, the authorizer authenticates.

**Start with a daily quota** — weekly and monthly periods hide burn until the window rolls; DAY is the honest unit for a new plan. Quota and throttle are independent: omit `quota` for throttle-only plans, omit `throttle` to inherit the account ceilings. The quota `offset` exists for mid-period migrations — requests already counted against the first period.

**Per-method throttles are for the expensive paths** — put the plan-wide ceiling on `throttle` and a tighter cap on the costly method (`/search/GET` and friends, addressed as `{resource-path}/{METHOD}`). A plan-wide limit sized for cheap endpoints lets one expensive path exhaust the backend within it.

**Rotate by adding a key, then removing the old one** — there is no in-place value rotation; treat keys as replaceable identities. To cut a consumer off without regenerating anything, set the key's `enabled: false` instead of deleting it.

**Let AWS generate key values** — omitting `value` is the recommended path. Set it only when a consumer's key must survive re-creation, source it as the secret it is, and remember key names are unique account-wide.

**Key values never appear in outputs** — deliberately. `api_key_ids` and `api_key_arns` identify keys for rotation and IAM; the value is read from the AWS API or console at distribution time.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsRestApiGateway** | `apiStages[].restApiId` | `status.outputs.rest_api_id` |
| **AwsRestApiGateway** | `apiStages[].stageName` | `status.outputs.stage_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `api_key_ids` | API key IDs keyed by each `apiKeys` entry's name | Key rotation tooling; identifying which consumer a key belongs to |
| `api_key_arns` | API key ARNs keyed the same way | IAM policies scoping key management |
| `usage_plan_id` / `usage_plan_arn` | The plan's identifiers | Operational tooling addressing the plan |

Key values are deliberately absent from this table — they are secrets, read from AWS at distribution time.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Consumer tiers** — one plan per tier (free, partner, internal), each covering the same stage with different quotas and its own keys. Moving a consumer between tiers is moving their key between plans, with no API change. Start from the **Metered API Keys** preset.

**Protecting the expensive path** — a plan-wide steady-state ceiling with a much tighter per-method throttle on the costly endpoint, so a burst of cheap calls cannot mask a flood of expensive ones. Start from the **Throttled Stages** preset.

**Marketplace metering** — `productCode` on the plan and `customerId` on each key link consumers to an AWS Marketplace SaaS listing, so subscriber usage is metered through Marketplace billing.

## Works With

- [**AWS REST API Gateway**](/infra-catalog/aws-rest-api-gateway) — the APIs and stages the plan covers; routes demand keys via `apiKeyRequired`
- [**AWS REST API Domain**](/infra-catalog/aws-rest-api-domain) — the custom hostname consumers call the metered stages through

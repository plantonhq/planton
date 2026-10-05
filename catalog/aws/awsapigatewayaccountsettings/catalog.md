# AWS API Gateway Account Settings

Manages the region's API Gateway account object, whose one configurable lever is the CloudWatch Logs role every REST API stage in the region logs through. Stage-level execution and access logging on a REST API silently does nothing until this role is set — that cross-component contract is why this kind exists. This is a settings singleton: AWS keeps exactly one API Gateway account object per region, its identity is the account+region pair, and destroying the component resets the role (region-wide API Gateway logging stops) while the account object itself always exists.

## What Gets Created

Nothing is created at AWS — the API Gateway account object already exists in every region and cannot be deleted. The module adopts and configures:

- **API Gateway Account** — the region's account object, setting (or explicitly clearing) the CloudWatch Logs role API Gateway assumes to push execution and access logs for every REST API stage in the region. Deploy at most one instance per region — two instances targeting the same region fight over the same settings object

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with API Gateway account permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- An IAM role trusting `apigateway.amazonaws.com` with CloudWatch Logs write permissions — AWS's managed `AmazonAPIGatewayPushToCloudWatchLogs` policy is the canonical grant. AWS validates the role at apply time and rejects one it cannot use ("The role ARN does not have required permissions"); custom policies that miss `logs:CreateLogGroup` fail this validation.

## Deploy

### Console

Open the deployment store, find **AWS API Gateway Account Settings**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the logging role. Start from the **CloudWatch Logging Role** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsApiGatewayAccountSettings
metadata:
  name: apigw-account-settings
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  cloudwatchRoleArn:
    valueFrom:
      kind: AwsIamRole
      name: apigw-logging-role
      fieldPath: status.outputs.role_arn
```

```shell
planton apply -f apigw-account-settings.yaml
```

This adopts the region's API Gateway account object and sets the referenced role as its CloudWatch logging role — stage-level logging on every REST API in the region works from here on. An Infra Job tracks the provisioning in real time.

### InfraChart

When the settings deploy alongside the logging role in one chart, wire the role reference via ValueFromRef:

```yaml
spec:
  region: us-east-1
  cloudwatchRoleArn:
    valueFrom:
      kind: AwsIamRole
      name: apigw-logging-role
      fieldPath: status.outputs.role_arn
```

The InfraPipeline resolves the dependency graph, creates the role first, then configures the account object with it.

## Key Configuration

These are the most important decisions when configuring API Gateway account settings. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Set this before enabling stage logging** — Stage-level execution/access logging on REST APIs fails silently without the account role: the stages deploy fine and log nothing, with no error anywhere. When API Gateway logs are mysteriously absent, this kind is the first place to look.

**One role serves every API in the region** — The role is region-wide, not per-API. Use the managed `AmazonAPIGatewayPushToCloudWatchLogs` policy rather than a hand-rolled one; AWS validates the role's permissions at apply and rejects incomplete grants.

**Unset is a real posture, not an omission** — An instance without `cloudwatchRoleArn` explicitly manages "no API Gateway logging": applying it clears any role previously set by anyone. Use that shape deliberately (the no-logging posture), never as a default you forgot to fill.

**Destroy stops logging region-wide** — Destroying this component resets the role, silencing execution and access logs for EVERY REST API in the region at once. Treat the destroy as a region-wide logging outage and drain expectations accordingly.

**Singleton discipline is yours to keep** — One instance per region; `metadata.name` never reaches AWS. A role created seconds before this setting can race IAM propagation — both engines retry through the transient "role ARN does not have required permissions" window, so a slow first apply is normal.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `cloudwatchRoleArn` | `status.outputs.role_arn` |

### What This Kind Provides

The outputs are AWS-reported echoes of the account object, not composition inputs — no catalog kind consumes them via ValueFromRef. `account_id` is the provider's import ID for the singleton; `api_key_version` and `features` are AWS-managed informational fields; `throttle_burst_limit` and `throttle_rate_limit` report the account-level throttle ceilings every stage in the region shares — useful when sizing per-stage throttling, but read-only here.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**The logging foundation, once per region** — one instance with the logging role, deployed before any REST API that wants stage logging. Pair it with the role in a chart so a fresh region bootstraps in one apply. Start from the **CloudWatch Logging Role** preset.

**Explicit no-logging posture** — an instance with `cloudwatchRoleArn` unset, making "this region does not log API Gateway traffic" a reviewed, version-controlled decision instead of an accident — and clearing any role a console user set by hand. Start from the **No-Logging Posture** preset.

## Works With

- [**AWS REST API Gateway**](/infra-catalog/aws-rest-api-gateway) — the stages whose execution/access logging depends on this account role
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the region-wide logging role API Gateway assumes
- [**AWS CloudWatch Log Group**](/infra-catalog/aws-cloudwatch-log-group) — where the execution and access logs land once the role is set

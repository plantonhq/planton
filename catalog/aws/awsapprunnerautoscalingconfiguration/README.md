# AwsAppRunnerAutoScalingConfiguration

Deploy and manage an AWS App Runner auto scaling configuration using Planton -- the reusable, versioned scaling policy that controls how App Runner scales a service's instance count in response to incoming request concurrency.

## Overview

App Runner scales on concurrency: when the number of simultaneous requests routed to one instance exceeds `maxConcurrency`, another instance launches (up to `maxSize`); when traffic drops, instances above `minSize` stop. Instances at the warm floor are billed for memory only while idle, so `minSize` is the cold-start-versus-cost dial.

The configuration is deliberately its own resource rather than fields on the service:

- **Shared by many services** -- each service references the configuration by ARN, so a fleet adopts one scaling posture tuned in one place.
- **Versioned by AWS** -- every value is create-time immutable; changing any of them registers a NEW revision under the same name. The exported ARN carries the revision, so a change here rolls referencing services through the resource graph on their next deployment.

## When to Use

- Give production APIs a warm floor (`minSize` 2-3) so bursts never wait on cold starts.
- Cap costs on internal tools with a small `maxSize`.
- Standardize one scaling posture across a fleet of services instead of tuning each service separately.

## Minimal Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsAppRunnerAutoScalingConfiguration
metadata:
  name: prod-api-scaling
  org: my-org
  env: prod
  id: prod-api-scaling-prod
spec:
  region: us-east-1
  maxConcurrency: 50
  maxSize: 15
  minSize: 3
```

## Spec Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `region` | string | Yes | AWS region for the configuration. |
| `maxConcurrency` | int | No (default 100) | Concurrent requests per instance before scale-out (1-200). Lower = more headroom per instance, higher cost. |
| `maxSize` | int | No (default 25) | Scale-out ceiling. AWS caps at 25 by default (quota-adjustable). |
| `minSize` | int | No (default 1) | Warm floor kept provisioned at all times (memory-only billing while idle). |
| `setAsAccountDefault` | bool | No (default false) | Claim this configuration as the account/region default for App Runner services created WITHOUT an explicit configuration. One default per account/region; claiming displaces the previous holder; only future services are affected; one-way at AWS (destroy never restores the previous default). |

## Outputs

| Output | Description |
|--------|-------------|
| `configuration_arn` | The revision-carrying ARN services reference via `autoScalingConfigurationArn`. |
| `configuration_revision` | The revision this deployment registered. |
| `latest` | Whether this revision is the latest under the configuration name. |
| `is_default` | Whether this configuration holds the account/region default designation (true when `setAsAccountDefault` claimed it). |

## Composing into a Service

```yaml
# In an AwsAppRunnerService spec:
autoScalingConfigurationArn:
  valueFrom:
    kind: AwsAppRunnerAutoScalingConfiguration
    name: prod-api-scaling
    fieldPath: status.outputs.configuration_arn
```

## Deliberately Omitted

- **Per-kind tags**: identity tags derive from metadata; custom user tags are a platform-wide concern.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

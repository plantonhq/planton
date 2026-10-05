# AWS App Runner Auto Scaling Configuration

Deploys an App Runner auto scaling configuration — the reusable scaling policy that controls how App Runner scales a service's instance count in response to request concurrency. It is deliberately its own resource: any number of [App Runner services](/infra-catalog/aws-app-runner-service) reference one configuration by ARN, so a fleet adopts a common scaling posture that is tuned in one place. AWS versions these configurations — a change registers a new revision under the same name, and the revision-carrying ARN rolls referencing services on their next deployment.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **App Runner Auto Scaling Configuration** -- a named, versioned scaling policy; the resource name is the configuration name and each value change registers a new revision under it
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **Max size quota** -- AWS caps a configuration's max size at 25 instances by default; request a service-quota increase before setting it higher.
- **Same-region services** -- App Runner services can only reference configurations in their own region; a multi-region fleet keeps one configuration per region.

## Deploy

### Console

Open the deployment store, find **AWS App Runner Auto Scaling Configuration**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Latency-Sensitive API** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsAppRunnerAutoScalingConfiguration
metadata:
  name: latency-sensitive-api
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  maxConcurrency: 50
  maxSize: 15
  minSize: 3
```

```shell
planton apply -f app-runner-auto-scaling.yaml
```

This registers a scaling posture with three warm instances and a lowered concurrency ceiling. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring an auto scaling configuration. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Max concurrency is the scale-out trigger** -- when a single instance's concurrent requests exceed it (1–200; AWS default 100), App Runner adds instances. Lower values give each instance more headroom (better tail latency) at higher cost; `1` dedicates an instance per request — the serverless-function posture.

**Min size is the warm floor** -- instances kept provisioned at all times, serving without cold-start latency and billed for memory only while idle (AWS default 1). Raise it for latency-sensitive services worth the standing memory charge.

**Max size is the cost ceiling** -- the hard scale-out cap during traffic spikes for every service using this configuration (AWS default 25). It can never sit below the warm floor.

**Every change is a new revision** -- values are create-time immutable at AWS; editing them here registers a new revision under the same configuration name, and referencing services roll onto it on their next deployment. That is the intended fleet-tuning workflow: tune once, the fleet follows.

**Blank dials keep AWS defaults** -- an omitted value records no opinion, so AWS's own default applies and a future AWS default change flows through naturally.

**The account default is claimable** -- `setAsAccountDefault: true` designates this configuration as the account/region default: App Runner services created WITHOUT an explicit `autoScalingConfigurationArn` adopt it. Three truths before claiming: one default exists per account/region (claiming displaces the previous holder), only services created AFTERWARDS are affected, and the claim is one-way at AWS -- destroying the resource never restores the previous default; the designation stays until another configuration claims it. Keep exactly one claimant per account/region; `status.outputs.is_default` confirms the claim.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies — it is a leaf resource: App Runner services reference its ARN, never the other way around.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `configuration_arn` | Revision-carrying ARN of the configuration | App Runner service `autoScalingConfigurationArn` |
| `configuration_revision` | The revision number this deployment registered | Audit and rollout tracking |
| `latest` | Whether this revision is the configuration name's latest | Fleet-tuning verification |
| `is_default` | Whether this configuration holds the account/region default designation | Governance verification of the claimed baseline |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Latency-sensitive API** -- three warm instances absorb bursts without cold starts, and a 50-request concurrency ceiling keeps per-instance headroom. Start from the **Latency-Sensitive API** preset.

**Conservative internal tool** -- one warm instance, dense request packing, and a tight scale-out cap so a traffic anomaly can never triple the bill. Start from the **Cost-Conscious** preset.

**Per-posture fleets** -- name configurations for postures (`latency-sensitive-api`, `batch-conservative`), never for individual services; every service that shares a posture references the same configuration and is retuned in one edit.

## Works With

- [**AWS App Runner Service**](/infra-catalog/aws-app-runner-service) -- adopts this scaling posture via `autoScalingConfigurationArn` (consumes `configuration_arn`)

# AWS SageMaker Endpoint

Deploys an Amazon SageMaker real-time inference endpoint together with its endpoint configuration as one declarative resource. Each of the 1–10 production variants serves one SageMaker model on either serverless compute (per-inference billing, nothing while idle) or dedicated instances with managed scaling, routing strategy, and startup timeouts; shadow variants mirror production traffic for silent testing. Because AWS makes endpoint configurations immutable, every capacity change rolls a new name-suffixed configuration and repoints the endpoint — the modules own that choreography, and a `deployment` policy with CloudWatch-alarm auto-rollback shapes the crossing for production fleets. Data capture to S3 and async inference with SNS notifications round out the surface.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **SageMaker Endpoint Configuration** — the immutable capacity definition: variants, serverless or instance sizing, data capture, async inference, and KMS volume encryption. Configurations are name-suffixed (`<name>-cfg-…`) and a new one is created before the old is destroyed, so the endpoint never references a deleted configuration.
- **SageMaker Endpoint** — the invocable resource, named from `metadata.name`. That name never changes across configuration rolls, so clients keep invoking the same identity while capacity evolves underneath. When a `deployment` policy is set, UpdateEndpoint rolls blue/green or in batches with alarm-guarded rollback.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with SageMaker control-plane permissions (`sagemaker:CreateEndpoint`, `sagemaker:CreateEndpointConfig`, and their update/delete siblings). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- A SageMaker model for each variant to serve — or, for inference-component endpoints where variants omit `model`, an IAM role wired via `executionRoleArn` instead.
- For instance-backed variants: a granted "for endpoint usage" Service Quota for your instance type. Fresh accounts default to ZERO for nearly every family (ml.m5.large included), and CreateEndpoint fails with ResourceLimitExceeded until the increase lands. Serverless variants need no per-type quota.
- The S3 buckets (and SNS topics) that `dataCapture`, `asyncInference`, or `coreDump` point at (only for those features).

## Deploy

### Console

Open the deployment store, find **AWS SageMaker Endpoint**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, region, and the variant list — model, compute form, and traffic weight per variant. Start from the **Serverless Endpoint** preset in the [Presets](#presets) tab for the start-cheap shape, or the **Production Canary Endpoint** preset for a weighted two-variant fleet with guarded rollouts.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSagemakerEndpoint
metadata:
  name: churn-scoring
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  productionVariants:
    - variantName: AllTraffic
      model:
        valueFrom:
          kind: AwsSagemakerModel
          name: churn-model
          fieldPath: status.outputs.model_name
      serverless:
        maxConcurrency: 20
        memorySizeMb: 2048
```

```shell
planton apply -f sagemaker-endpoint.yaml
```

This creates a serverless endpoint serving the referenced model at up to 20 concurrent invocations in 2 GB environments — no instances, no idle cost. An Infra Job tracks the provisioning in real time.

### InfraChart

When the endpoint deploys alongside its model in one chart, wire the model reference via ValueFromRef:

```yaml
spec:
  region: us-east-1
  productionVariants:
    - variantName: AllTraffic
      model:
        valueFrom:
          kind: AwsSagemakerModel
          name: churn-model
          fieldPath: status.outputs.model_name
      serverless:
        maxConcurrency: 20
        memorySizeMb: 2048
```

The InfraPipeline resolves the dependency graph, creates the model first, then stands up the endpoint serving it.

## Key Configuration

These are the most important decisions when configuring an endpoint. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Serverless first, instances when the numbers say so** — A serverless variant needs no capacity math and no quota request, and costs nothing idle; graduate to `instanceType` when latency floors, GPU needs, or provisioned-concurrency spend justify dedicated capacity. The swap is a configuration roll the modules handle. Request instance quotas BEFORE that first instance-backed deploy — and size them for rollouts, not steady state: blue/green transiently doubles the fleet.

**Shape the crossing before you need it** — Every capacity change is a configuration roll, so production endpoints deserve a `deployment` policy with `autoRollbackAlarmNames`: a bad model version backs itself out instead of paging you. Blue/green provisions a full parallel fleet and holds it through each bake step; rolling replaces batches in place with no parallel-fleet cost — pick by budget and blast radius. One hard rule: AWS rejects a rolling policy on a single-instance fleet, so a one-instance endpoint uses blue/green or omits `deployment`.

**Weight, don't flip** — New model versions ride a second variant at a small `initialVariantWeight`; traffic splits proportionally to weight over the sum. A weight of 0 keeps a variant deployed but takes no traffic — an instant rollback target that costs its instances but saves the redeploy.

**Capture from day one** — `dataCapture` is the Model Monitor feed, and flipping it on later is itself a configuration roll. Sample a low percentage early rather than retrofitting under incident pressure.

**Async for large or slow inference** — `asyncInference` queues requests and delivers responses to S3 instead of holding the connection: the shape for multi-hundred-MB payloads or minutes-long inference, with SNS topics notifying success and failure separately.

**The KMS caveat** — `kmsKeyArn` encrypts ML storage volumes, but nitro-local-storage families (ml.g5/g6/p4d/p5 and similar) encrypt locally by default and reject a custom key — as do serverless-only endpoints. Set it only on EBS-backed instance fleets.

**A healthy endpoint needs a loadable model** — A model whose container has no artifacts to load passes CreateModel but can never answer the endpoint's ping health checks; the endpoint parks at `Failed`. A failed create can also strand the endpoint object outside IaC state — delete it explicitly before retrying the same name.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSagemakerModel** | `productionVariants[].model`, `shadowVariants[].model` | `status.outputs.model_name` |
| **AwsIamRole** | `executionRoleArn` | `status.outputs.role_arn` |
| **AwsKmsKey** | `kmsKeyArn` (also under `dataCapture`, `asyncInference`, `coreDump`) | `status.outputs.key_arn` |
| **AwsSnsTopic** | `asyncInference.successTopicArn`, `asyncInference.errorTopicArn` | `status.outputs.topic_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `endpoint_name` | The endpoint's AWS identity — what clients pass to InvokeEndpoint | Application configuration; Application Auto Scaling target registration |
| `endpoint_arn` | Amazon Resource Name of the endpoint | IAM policies scoping `sagemaker:InvokeEndpoint` to this endpoint |

`endpoint_config_name` and `endpoint_config_arn` are also exported, but they echo whichever configuration is currently in service — the modules roll a new one on every capacity change, so treat them as audit values, not composition inputs.

## Common Patterns

**Start-cheap serverless endpoint** — one serverless variant, no capacity planning, no idle charge: the right first endpoint for any model, before traffic patterns are known. Raise `maxConcurrency` and `memorySizeMb` as traffic and model size grow, and add `provisionedConcurrency` only when cold starts measurably hurt (it bills while provisioned). Start from the **Serverless Endpoint** preset.

**Guarded canary fleet** — two weighted instance-backed variants (stable at 90, candidate at 10), data capture feeding Model Monitor, and a blue/green canary policy watching a CloudWatch alarm. Promote by shifting weights; every weight change is itself a guarded roll. Start from the **Production Canary Endpoint** preset.

**Shadow testing** — exactly one production and one shadow variant (AWS's required shape, enforced at manifest time): the shadow receives a copy of live traffic but its responses never reach callers. The zero-risk way to compare a candidate's behavior on real inputs before it earns a traffic weight.

## Works With

- [**AWS SageMaker Model**](/infra-catalog/aws-sagemaker-model) — the model each variant serves, wired via the variant's `model` reference
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the execution role for inference-component endpoints whose variants omit a model
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — customer-managed encryption for ML volumes, captured data, and async outputs
- [**AWS SNS Topic**](/infra-catalog/aws-sns-topic) — success and error notifications for async inference
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) — destinations for data capture, async responses, and core dumps, referenced by URI

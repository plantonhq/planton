# AWS Bedrock Provisioned Throughput

Purchases dedicated, guaranteed Amazon Bedrock model serving capacity in model units — the required serving path for fine-tuned custom models and an option for high-volume foundation-model workloads. This is the catalog's most financially consequential small resource: capacity bills from the moment it is created, whether used or not — per hour without a commitment, or for the full term with a 1-month or 6-month commitment that cannot be canceled (or even destroyed) until the term lapses. Every spec field is create-time-immutable, so resizing or re-targeting is a replacement.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Provisioned Model** — reserved throughput for the referenced model (typically a custom model's output ARN; a foundation-model ARN where AWS allows direct provisioning), sized by `modelUnits` and addressable by its own ARN, which applications invoke exactly like a model ID

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with Bedrock provisioned-throughput permissions (`bedrock:CreateProvisionedModelThroughput` and its read/update/delete siblings). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **The model to serve** — typically an AwsBedrockCustomModel whose `custom_model_arn` this purchase serves; fine-tuned models cannot serve on-demand at all.

### AWS Account

- **A no-commitment model-unit quota above zero** — the Service Quotas default is often 0, so the first no-commitment purchase commonly fails on quota; raise it deliberately (only for no-commitment purchases — commitments follow their own path).
- **A cost sign-off** — capacity bills from creation, and committed terms bill in full regardless of usage.

## Deploy

### Console

Open the deployment store, find **AWS Bedrock Provisioned Throughput**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: region, the model reference, units, and commitment. Start from the **No-Commitment Custom Model Capacity** preset in the [Presets](#presets) tab — hourly billing, deletable any time.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockProvisionedThroughput
metadata:
  name: support-model-capacity
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  modelArn:
    valueFrom:
      kind: AwsBedrockCustomModel
      name: support-titan-ft
      fieldPath: status.outputs.custom_model_arn
  modelUnits: 1
```

```shell
planton apply -f provisioned-throughput.yaml
```

This purchases one no-commitment model unit for the referenced fine-tuned model — hourly billing starts at creation, and the purchase can be deleted at any time. An Infra Job tracks the provisioning in real time.

### InfraChart

When the purchase deploys alongside the custom model it serves, wire the reference via ValueFromRef:

```yaml
spec:
  region: us-east-1
  modelArn:
    valueFrom:
      kind: AwsBedrockCustomModel
      name: support-titan-ft
      fieldPath: status.outputs.custom_model_arn
  modelUnits: 1
```

The InfraPipeline resolves the dependency graph, waits for the custom model, then buys capacity for it.

## Key Configuration

These are the most important decisions when configuring a throughput purchase. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**No-commitment first** — validate the integration on hourly billing before committing to a term. Omitting `commitmentDuration` IS the no-commitment purchase (deletable any time); `OneMonth`/`SixMonths` buy discounted rates in exchange for a term that bills in full and refuses deletion until it lapses. Treat committed terms like reserved instances: a procurement decision recorded in a manifest, not infrastructure to iterate on.

**Quota before purchase** — the account's no-commitment model-unit quota often defaults to 0, so a create failing on quota is the GOOD failure mode: it means the guardrail worked. Raise the quota via Service Quotas before the deploy that needs it.

**A model unit is model-specific** — the tokens-per-minute a unit buys differs per model; AWS quotes each model's per-unit throughput in the console. Size `modelUnits` against measured peak demand, not intuition — the cost drivers are units × the model's per-unit rate × the commitment.

**Every change is a replacement** — resizing units or changing the model destroys and recreates the purchase, and a committed purchase blocks that replacement until its term ends. Plan resizes as capacity overlaps: buy the new purchase, move consumers to its ARN, then delete the old one.

**Watch utilization** — provisioned capacity bills whether traffic flows or not. CloudWatch's provisioned-throughput utilization metrics are the signal for when to resize or release.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsBedrockCustomModel** | `modelArn` | `status.outputs.custom_model_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `provisioned_model_arn` | ARN of the provisioned model — the modelId applications pass to InvokeModel/Converse | An AwsBedrockAgent alias's `routing.provisionedThroughput`; application configuration consuming the dedicated capacity |

`provisioned_model_name` echoes `metadata.name` — a record, not a composition input.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Serve a fine-tuned model** — the custom model produces `custom_model_arn`, this purchase makes it servable, and applications (or an agent alias's routing) consume `provisioned_model_arn`. One no-commitment unit is the standard integration-validation shape. Start from the **No-Commitment Custom Model Capacity** preset.

**Committed production capacity** — once demand is measured and steady, a OneMonth or SixMonths commitment at the measured unit count buys the discounted rate. Enter it knowing the term is irreversible — the manifest cannot delete or resize the purchase until the term ends. Start from the **Committed Production Capacity** preset.

**Capacity behind an agent alias** — wire `provisioned_model_arn` into an AwsBedrockAgent alias's `routing.provisionedThroughput` so the alias serves production traffic through reserved capacity while the draft and other aliases stay on on-demand.

## Works With

- [**AWS Bedrock Custom Model**](/infra-catalog/aws-bedrock-custom-model) — the fine-tuned model this purchase serves, wired via `modelArn`
- [**AWS Bedrock Agent**](/infra-catalog/aws-bedrock-agent) — agent aliases route production traffic through this capacity via `routing.provisionedThroughput`
- [**AWS Bedrock Model Access**](/infra-catalog/aws-bedrock-model-access) — the agreement required before provisioning capacity for a marketplace foundation model

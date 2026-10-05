# AWS Bedrock Inference Profile

Deploys an Amazon Bedrock application inference profile — a named handle over a foundation model (or an AWS system-defined cross-region profile) that carries its own ARN and tags, so each application's or tenant's model usage can be tracked, cost-allocated, and IAM-scoped independently. Invocations through the profile bill at the underlying model's rates; the profile is the attribution and access-control layer, not a capacity or routing change of its own. Every spec field is create-time-immutable — changing one replaces the profile and its ARN.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Application Inference Profile** — routing to the model source named by `sourceArn`: a foundation-model ARN for single-region use, or an AWS system-defined inference-profile ARN to inherit cross-region routing. The profile's tags carry the cost-allocation identity.

This kind creates APPLICATION profiles only — SYSTEM_DEFINED profiles are AWS-owned and are consumed by reference (as a `sourceArn`), never managed.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with Bedrock inference-profile permissions (`bedrock:CreateInferenceProfile` and its read/delete siblings). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **Access to the underlying model** — auto-enabled AWS models work immediately; models requiring a marketplace agreement need it in place first (only when `sourceArn` routes to one).

## Deploy

### Console

Open the deployment store, find **AWS Bedrock Inference Profile**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: region, description, and the model source. Start from the **Per-Application Tracking** preset in the [Presets](#presets) tab and name the resource after the consuming application.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockInferenceProfile
metadata:
  name: checkout-nova
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  description: Cost tracking for the checkout service
  sourceArn: arn:aws:bedrock:us-west-2:123456789012:inference-profile/us.amazon.nova-micro-v1:0
```

```shell
planton apply -f inference-profile.yaml
```

This creates a profile for the checkout service over the US cross-region Nova Micro profile — its ARN is what the service invokes and what IAM scopes. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring an inference profile. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One profile per consumer** — name the profile after the service, team, or tenant it attributes; the whole point is per-consumer cost and usage attribution, and the profile's tags carry that identity into billing.

**Scope IAM to the profile ARN, not the model** — granting `bedrock:InvokeModel` on `inference_profile_arn` is what turns attribution from advisory into enforceable: the application can invoke through ITS handle only, and usage lands on its cost line.

**The `sourceArn` shape depends on the model's inference types** — the foundation-model ARN form works only for models that support ON_DEMAND inference. Profile-only models — the Nova family and most 2025+ releases — must be referenced through their system-defined inference-profile ARN form (`arn:...:inference-profile/us.<model-id>`); a direct foundation-model reference to one is rejected at create with "The provided foundation model does not support On Demand inference".

**Cross-region routing is inherited, not configured** — source the profile from AWS's system-defined geo profile and invocations ride AWS's cross-region capacity pools. There is no routing knob on this component; the choice of `sourceArn` IS the routing decision.

**Replacement changes the ARN** — every spec field is create-time-immutable, so any change destroys and recreates the profile under a new ARN. Roll the new ARN to consumers like a credential rotation: deploy the replacement, move consumers, then retire the old profile.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies — `sourceArn` names an AWS foundation model or system-defined profile, which are AWS-owned resources referenced as ARN strings rather than platform-managed components.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `inference_profile_arn` | The profile's ARN — the modelId applications pass to InvokeModel/Converse | An AwsBedrockAgent's `foundationModel`; IAM policies granting invoke on exactly this handle |
| `inference_profile_id` | The profile's unique identifier | Application configuration and usage queries |

`status` (ACTIVE when usable) and `type` (always APPLICATION for profiles this component creates) are state echoes, not composition inputs.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Per-application cost tracking** — one profile per service over the model it uses, IAM scoped to each profile's ARN. Model spend then decomposes by consumer in Cost Explorer without any application-side instrumentation. Start from the **Per-Application Tracking** preset.

**Cross-region resilience** — source the profile from an AWS geo profile (`us.`, `eu.`, ...) so invocations ride AWS's cross-region capacity pools during regional load spikes, while your applications keep a single stable ARN. This is also the required shape for profile-only models like the Nova family. Start from the **Cross-Region Routing** preset.

## Works With

- [**AWS Bedrock Agent**](/infra-catalog/aws-bedrock-agent) — accepts this profile's ARN as its `foundationModel` for per-agent cost attribution
- [**AWS Bedrock Prompt**](/infra-catalog/aws-bedrock-prompt) — prompt variants can execute through a profile instead of a bare model ID
- [**AWS Bedrock Flow**](/infra-catalog/aws-bedrock-flow) — inline prompt nodes accept a profile ID/ARN as their model
- [**AWS Bedrock Model Access**](/infra-catalog/aws-bedrock-model-access) — the marketplace agreement the underlying model may require

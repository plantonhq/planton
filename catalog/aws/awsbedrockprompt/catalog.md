# AWS Bedrock Prompt

Deploys an Amazon Bedrock prompt — a reusable, versionable prompt definition in Bedrock Prompt Management, carrying up to 3 named variants for A/B comparison instead of formulations scattered through application code. Each variant is a text template with `{{variables}}` or a multi-turn chat template (system context, tool catalogs with auto/any/forced tool choice, prompt-caching checkpoints), and executes against a model — ID, ARN, or inference profile — or a Bedrock agent alias, with its own inference settings. This resource manages the prompt's mutable DRAFT; the cost driver is the target model's per-invocation billing when the prompt executes.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Bedrock Prompt** — the prompt with every `variants` entry rendered onto its draft: the template shape (text or chat, with AWS's discriminator derived from which one is set), the execution target (model XOR agent alias, likewise derived), inference configuration, metadata annotations, and model-specific request fields
- **Prompt Encryption** — configured only when `customerEncryptionKeyArn` is set; without it AWS encrypts with a Bedrock-managed key

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with Bedrock Prompt Management permissions (`bedrock:CreatePrompt` and its read/update/delete siblings). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **Bedrock Prompt Management available in the target region.**
- **Access to the variant models** — auto-enabled AWS models need nothing; marketplace models need an access agreement first (only for variants targeting one).
- **The agent alias deployed** — only for agent-targeted variants: the referenced alias must exist before the prompt can be created.

## Deploy

### Console

Open the deployment store, find **AWS Bedrock Prompt**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: region, the variants, and the default. Start from the **Single Text Prompt** preset in the [Presets](#presets) tab for the one-variant starting shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockPrompt
metadata:
  name: support-answer
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  defaultVariant: chat
  variants:
    - name: text
      modelId: amazon.nova-micro-v1:0
      text:
        text: "Answer concisely: {{question}}"
        inputVariables:
          - question
    - name: chat
      modelId: amazon.nova-lite-v1:0
      chat:
        system:
          - text: You are a concise support assistant.
        messages:
          - role: user
            text: "Answer the question: {{question}}"
        inputVariables:
          - question
```

```shell
planton apply -f prompt.yaml
```

This creates a prompt with two candidate formulations — a Nova Micro text variant and a Nova Lite chat variant — serving the chat variant by default. An Infra Job tracks the provisioning in real time.

### InfraChart

When a variant executes through a Bedrock agent, wire the alias reference via ValueFromRef:

```yaml
spec:
  region: us-west-2
  variants:
    - name: agent-backed
      agentAliasArn:
        valueFrom:
          kind: AwsBedrockAgent
          name: support-agent
          fieldPath: status.outputs.alias_arns.live
      text:
        text: "Handle this customer request: {{request}}"
        inputVariables:
          - request
```

The InfraPipeline resolves the dependency graph, deploys the agent (and its alias) first, then creates the prompt that targets it.

## Key Configuration

These are the most important decisions when configuring a prompt. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The prompt is the interface, variants are implementations** — applications invoke the prompt ID; `defaultVariant` decides what runs. Promote a candidate formulation by flipping the default — no application deploy, no consumer change. AWS caps a prompt at 3 variants, so retire losers rather than accumulating them.

**The DRAFT moves; published versions do not** — this component manages the draft only. When a formulation ships, publish a numbered version (console or API) and pin critical consumers to it; the draft keeps evolving underneath.

**Declare every `{{variable}}`** — AWS matches template placeholders against `inputVariables` at invocation, not at deploy. An undeclared variable surfaces as a runtime invocation error — the one failure mode this kind cannot catch at apply.

**Tools describe, models decide** — in chat variants, the tool catalog's descriptions are the model's only signal for tool selection: write them like API documentation. Use `toolChoice.any` to force SOME tool call (the structured-output extraction trick) and `toolName` to force a specific one.

**Cache points need model support** — `cachePoint: true` marks a prompt-caching checkpoint in templates, messages, system blocks, and tool lists; the repeated-prefix token savings only materialize on models that support caching. Verify before designing the prompt around it.

**`additionalModelRequestFields` is the escape hatch** — model-specific parameters outside the standard inference set (e.g. `top_k` for Anthropic models) pass through as JSON, unvalidated until invocation.

**Model target or agent target** — a variant executes against a model (`modelId`: foundation model, ARN, or inference profile) XOR an agent alias (`agentAliasArn`). Agent-targeted variants reference the alias, never the agent itself — read the ARN from the agent's `alias_arns` output map.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsBedrockAgent** | `variants[].agentAliasArn` | `status.outputs.alias_arns.<alias-name>` |
| **AwsKmsKey** | `customerEncryptionKeyArn` | `status.outputs.key_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `prompt_id` | The unique prompt identifier | Application configuration for invoking the prompt |
| `prompt_arn` | The prompt's ARN | An AwsBedrockFlow prompt node's `promptArn`; IAM policies |

`draft_version` is the constant "DRAFT" — a record of the mutable working version, not a composition input.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Single managed prompt** — one text variant, one model: the smallest step from prompts-in-code to prompts-as-infrastructure, giving the formulation review, history, and an ARN that flows can reference. Start from the **Single Text Prompt** preset.

**Chat assistant with tools** — a chat variant carrying system context, a described tool catalog, and tool choice, for assistants whose behavior is defined declaratively rather than assembled per-request in application code. Start from the **Chat with Tools** preset.

**A/B by default flip** — two variants of the same intent (different template, model, or inference settings) on one prompt; consumers follow `defaultVariant`, so the comparison and the promotion are both one-field edits. Use variant `metadata` to annotate what each candidate changes.

## Works With

- [**AWS Bedrock Flow**](/infra-catalog/aws-bedrock-flow) — prompt nodes execute this prompt by `prompt_arn`
- [**AWS Bedrock Agent**](/infra-catalog/aws-bedrock-agent) — the execution target for agent-backed variants, referenced through its alias ARN
- [**AWS Bedrock Inference Profile**](/infra-catalog/aws-bedrock-inference-profile) — an alternative `modelId` for per-application cost attribution
- [**AWS Bedrock Model Access**](/infra-catalog/aws-bedrock-model-access) — the agreement a marketplace variant model requires
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — customer-managed encryption via `customerEncryptionKeyArn`

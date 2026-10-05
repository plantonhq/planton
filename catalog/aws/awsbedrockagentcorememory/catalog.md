# AWS Bedrock AgentCore Memory

Deploys an Amazon Bedrock AgentCore memory — a managed store that gives agents short-term memory (raw session events kept for a retention window you set) and long-term memory (facts, summaries, preferences, and episodes that extraction strategies distill from those events). Strategies are declarative pipelines: built-in SEMANTIC, SUMMARIZATION, USER_PREFERENCE, and EPISODIC extraction run fully AWS-managed, while CUSTOM overrides the prompts and models. Records partition by namespace templates, can be encrypted with your KMS key, and can stream to Kinesis as they are written.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Memory** — the store itself, named by `memoryName`, with a 7–365 day short-term window (`eventExpiryDays`), optional customer-managed KMS encryption, indexed metadata keys for filtered retrieval, and optional Kinesis delivery of long-term records
- **Memory Strategy** — one per `strategies` entry: an extraction pipeline with its namespace templates, and for CUSTOM entries the prompt/model overrides per pipeline step. AWS serializes all strategy changes through the parent memory, and the modules order them accordingly

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with AgentCore control-plane permissions (`bedrock-agentcore:CreateMemory` and its siblings). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- Bedrock AgentCore available in the target region
- An IAM role trusting `bedrock-agentcore.amazonaws.com` with model-invoke or stream-write access, wired as `executionRoleArn` (only for CUSTOM strategies or `kinesisDelivery` — built-in strategies run entirely AWS-managed)
- A Kinesis data stream the role may write to (only for `kinesisDelivery`)

## Deploy

### Console

Open the deployment store, find **AWS Bedrock AgentCore Memory**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, the retention window, and the extraction strategies. Start from the **Assistant Memory** preset in the [Presets](#presets) tab for the facts-plus-summaries shape most assistants need, or the **Episodic Memory with Streaming** preset for experience episodes with Kinesis analytics.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockAgentCoreMemory
metadata:
  name: assistant-memory
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  memoryName: acme_assistant_memory
  description: Facts and summaries for the support assistant
  eventExpiryDays: 30
  strategies:
    - name: facts
      type: SEMANTIC
      description: Standalone facts from conversations
      namespaceTemplates:
        - /facts/{actorId}
    - name: summaries
      type: SUMMARIZATION
      namespaceTemplates:
        - /summaries/{actorId}/{sessionId}
```

```shell
planton apply -f agentcore-memory.yaml
```

This creates a memory with a 30-day event window and two built-in extraction strategies — semantic facts per actor and per-session summaries — running fully AWS-managed with no execution role. An Infra Job tracks the provisioning in real time.

### InfraChart

When the memory streams records to a Kinesis stream deployed in the same chart, wire the stream reference via ValueFromRef:

```yaml
spec:
  region: us-west-2
  memoryName: acme_assistant_memory
  eventExpiryDays: 30
  executionRoleArn:
    valueFrom:
      kind: AwsIamRole
      name: memory-delivery-role
      fieldPath: status.outputs.role_arn
  kinesisDelivery:
    dataStreamArn:
      valueFrom:
        kind: AwsKinesisStream
        name: memory-analytics
        fieldPath: status.outputs.stream_arn
    contentLevel: METADATA_ONLY
```

The InfraPipeline resolves the dependency graph, deploys the role and stream first, then provisions the memory delivering into them.

## Key Configuration

These are the most important decisions when configuring a memory. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**memoryName is an explicit field with a strict charset** — AWS requires a letter first, then letters, digits, and underscores; hyphens are rejected, which is why the name is not derived from `metadata.name`. Changing it replaces the memory and everything in it.

**Size the window to the conversation, not the archive** — short-term events exist to feed extraction; long-term records outlive them. A 30-day `eventExpiryDays` suits most assistants — retention of raw events is the storage cost driver, so pay for it only where raw replay matters.

**Namespaces are your query API** — design templates around retrieval (`/facts/{actorId}`, `/preferences/{actorId}`) before data lands. Changing them later strands existing records under old paths; there is no migration.

**Batch strategy edits** — each strategy change serializes through the parent memory (the provider allows 45 minutes per strategy operation; a four-strategy attach has run about 13 minutes per direction in live timing, individual attaches 2–7.5 minutes). Five separate applies take five serialized waits — group strategy changes into one apply.

**Reflection namespaces must stay inside the episodic namespace** — AWS requires each `reflectionNamespaceTemplates` entry to equal, or be a whole-segment prefix of, a `namespaceTemplates` entry (`/episodes` prefixes `/episodes/{actorId}`; `/epi` does not). The manifest validation rejects the contradiction before AWS's 400 does. Leave the field unset to mirror the episodic namespaces exactly.

**Indexed keys and the KMS key are create-time structure** — both replace the memory when changed. Choose the metadata keys you will filter retrieval by, and the encryption posture, up front.

**CUSTOM strategies need the execution role** — with invoke access to the override models; built-in strategies run without one. The same role covers Kinesis delivery writes.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKmsKey** | `encryptionKeyArn` | `status.outputs.key_arn` |
| **AwsIamRole** | `executionRoleArn` | `status.outputs.role_arn` |
| **AwsKinesisStream** | `kinesisDelivery.dataStreamArn` | `status.outputs.stream_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `memory_arn` | The memory's ARN — the canonical key for IAM policies | AgentCore Evaluation harness `memory.memoryArn`; IAM policies scoping data-plane access |
| `memory_id` | The unique memory identifier | Agent code writing events and querying records through the AgentCore data plane |
| `strategy_ids` | Strategy IDs keyed by each `strategies` entry's name | Scoping harness memory retrieval to one strategy (`retrieval.strategyId`) |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Assistant memory** — SEMANTIC facts per actor plus SUMMARIZATION per session, a 30-day window, no execution role. The default shape for a conversational assistant that should remember users across sessions. Start from the **Assistant Memory** preset.

**Episodic memory with analytics** — an EPISODIC strategy capturing experience episodes with reflection, indexed keys for filtered retrieval, and Kinesis delivery streaming records to your analytics pipeline as they are written. Start from the **Episodic Memory with Streaming** preset.

**Custom extraction** — a CUSTOM strategy overriding a built-in shape's prompts and models (`SEMANTIC_OVERRIDE` and siblings) when the managed extraction misses your domain's vocabulary. Requires the execution role with invoke access to the override models; EPISODIC_OVERRIDE additionally requires the `reflection` override, and SUMMARY_OVERRIDE takes no `extraction` step.

## Works With

- [**AWS Bedrock AgentCore Evaluation**](/infra-catalog/aws-bedrock-agent-core-evaluation) — harnesses read and write this memory during evaluation runs via `memory_arn`
- [**AWS Bedrock AgentCore Runtime**](/infra-catalog/aws-bedrock-agent-core-runtime) — agents hosted on a runtime write events and query records through the AgentCore data plane
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the execution role for custom-strategy model invocation and Kinesis delivery
- [**AWS Kinesis Data Stream**](/infra-catalog/aws-kinesis-stream) — receives long-term memory records as they are written
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — customer-managed encryption for the memory at rest

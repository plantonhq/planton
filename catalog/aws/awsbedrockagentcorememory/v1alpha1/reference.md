# AwsBedrockAgentCoreMemory

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsBedrockAgentCoreMemorySpec defines the desired configuration for an
Amazon Bedrock AgentCore memory - a managed store that gives agents
both short-term memory (raw session events, kept for
`event_expiry_days`) and long-term memory (structured records a
STRATEGY extracts from those events: facts, summaries, preferences,
episodes).

Each `strategies` entry is one extraction pipeline. Built-in strategy
types (SEMANTIC, SUMMARIZATION, USER_PREFERENCE, EPISODIC) run
AWS-managed extraction; CUSTOM lets you override the extraction/
consolidation prompts and models. AWS serializes all strategy changes
through the parent memory (the modules order them; strategy operations
can take tens of minutes).

## Example

```yaml
# Canonical AwsBedrockAgentCoreMemory example (hack/dev manifest and
# refgen Example source): a memory exercising every arm -- encryption,
# the execution role, indexed keys, Kinesis stream delivery, built-in
# strategies of every type, and a CUSTOM strategy overriding extraction
# and consolidation. Literal ARNs stand in for composed references so the
# offline `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockAgentCoreMemory
metadata:
  name: support-agent-memory
  id: support-agent-memory
  org: test-org
  env: dev
spec:
  region: us-west-2
  memoryName: support_memory
  description: Long-term memory for the support agent
  eventExpiryDays: 30
  encryptionKeyArn:
    value: arn:aws:kms:us-west-2:123456789012:key/1234abcd-12ab-34cd-56ef-1234567890ab
  executionRoleArn:
    value: arn:aws:iam::123456789012:role/agentcore-memory-role
  indexedKeys:
    - key: customer_id
      type: STRING
    - key: priority
      type: NUMBER
  kinesisDelivery:
    dataStreamArn:
      value: arn:aws:kinesis:us-west-2:123456789012:stream/memory-records
    contentLevel: FULL_CONTENT
  strategies:
    - name: facts
      type: SEMANTIC
      description: Extract standalone facts from conversations
      namespaceTemplates:
        - /facts/{actorId}
    - name: summaries
      type: SUMMARIZATION
      namespaceTemplates:
        - /summaries/{actorId}/{sessionId}
    - name: preferences
      type: USER_PREFERENCE
      namespaceTemplates:
        - /preferences/{actorId}
    - name: episodes
      type: EPISODIC
      namespaceTemplates:
        - /episodes/{actorId}
      # AWS requires each reflection namespace to equal, or be a
      # hierarchical prefix of, an episodic namespace above; unrelated
      # roots are rejected at UpdateMemory. Leave unset to mirror the
      # episodic namespaces exactly.
      reflectionNamespaceTemplates:
        - /episodes
    - name: tuned_prefs
      type: CUSTOM
      namespaceTemplates:
        - /preferences/{actorId}
      custom:
        type: USER_PREFERENCE_OVERRIDE
        extraction:
          appendToPrompt: Focus on stated product preferences.
          modelId: anthropic.claude-3-5-sonnet-20241022-v2:0
        consolidation:
          appendToPrompt: Merge overlapping preferences into one record.
          modelId: anthropic.claude-3-5-sonnet-20241022-v2:0
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.memoryName` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.eventExpiryDays` | `int32` |  |  |  |
| `spec.encryptionKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.executionRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.indexedKeys` | `[]AwsBedrockAgentCoreMemoryIndexedKey` |  |  |  |
| `spec.indexedKeys[].key` | `string` | yes |  |  |
| `spec.indexedKeys[].type` | `string` |  |  |  |
| `spec.kinesisDelivery` | `AwsBedrockAgentCoreMemoryKinesisDelivery` |  |  |  |
| `spec.kinesisDelivery.dataStreamArn` | `string \| valueFrom` | yes |  | AwsKinesisStream (`status.outputs.stream_arn`) |
| `spec.kinesisDelivery.contentLevel` | `string` |  |  |  |
| `spec.strategies` | `[]AwsBedrockAgentCoreMemoryStrategy` |  |  |  |
| `spec.strategies[].name` | `string` | yes |  |  |
| `spec.strategies[].type` | `string` |  |  |  |
| `spec.strategies[].description` | `string` |  |  |  |
| `spec.strategies[].namespaceTemplates` | `[]string` | yes |  |  |
| `spec.strategies[].custom` | `AwsBedrockAgentCoreMemoryCustomStrategy` |  |  |  |
| `spec.strategies[].custom.type` | `string` |  |  |  |
| `spec.strategies[].custom.extraction` | `AwsBedrockAgentCoreMemoryPromptOverride` |  |  |  |
| `spec.strategies[].custom.extraction.appendToPrompt` | `string` | yes |  |  |
| `spec.strategies[].custom.extraction.modelId` | `string` | yes |  |  |
| `spec.strategies[].custom.consolidation` | `AwsBedrockAgentCoreMemoryPromptOverride` |  |  |  |
| `spec.strategies[].custom.consolidation.appendToPrompt` | `string` | yes |  |  |
| `spec.strategies[].custom.consolidation.modelId` | `string` | yes |  |  |
| `spec.strategies[].custom.reflection` | `AwsBedrockAgentCoreMemoryReflectionOverride` |  |  |  |
| `spec.strategies[].custom.reflection.appendToPrompt` | `string` | yes |  |  |
| `spec.strategies[].custom.reflection.modelId` | `string` | yes |  |  |
| `spec.strategies[].custom.reflection.namespaceTemplates` | `[]string` | yes |  |  |
| `spec.strategies[].reflectionNamespaceTemplates` | `[]string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the memory will be created.
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.memoryName

`string` · required

Memory name in AWS (1-48 characters; must start with a letter, then
letters, digits, underscore - AWS rejects hyphens here, so the name
is an explicit field rather than metadata.name). Changing it replaces
the memory.

- rule: {"string":{"minLen":"1","pattern":"^[a-zA-Z][a-zA-Z0-9_]{0,47}$"}}

### spec.description

`string`

Human-readable description shown in the AgentCore console (1-4096
characters when set). Updates in place.

- rule: {"string":{"maxLen":"4096"}}

### spec.eventExpiryDays

`int32`

Days AWS retains raw session events - the short-term memory window
(7-365). Long-term records extracted by strategies outlive it.

- rule: {"int32":{"lte":365,"gte":7}}

### spec.encryptionKeyArn

`string | valueFrom`

Customer-managed KMS key encrypting the memory at rest. Without it,
AWS uses a service-managed key. Changing it replaces the memory.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.executionRoleArn

`string | valueFrom`

IAM role AWS assumes for memory operations that touch YOUR resources
(custom-strategy model invocation, Kinesis stream delivery). The role
must trust bedrock-agentcore.amazonaws.com. Required in practice when
`strategies` has a CUSTOM entry or `kinesis_delivery` is set.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.indexedKeys

`[]AwsBedrockAgentCoreMemoryIndexedKey`

Metadata keys indexed for filtered retrieval (1-10 entries when set).
Changing the set replaces the memory.

- rule: {"repeated":{"maxItems":"10"}}

### spec.indexedKeys[].key

`string` · required

The metadata key (1-128 characters; letters, digits, spaces, and
. _ : / = + @ -).

- rule: {"string":{"minLen":"1","maxLen":"128","pattern":"^[a-zA-Z0-9\\s._:/=+@-]*$"}}

### spec.indexedKeys[].type

`string`

The key's value type: STRING, STRINGLIST, or NUMBER.

- rule: {"string":{"in":["STRING","STRINGLIST","NUMBER"]}}

### spec.kinesisDelivery

`AwsBedrockAgentCoreMemoryKinesisDelivery`

Stream long-term memory records to a Kinesis data stream as they are
written (build your own analytics or replication on top).

### spec.kinesisDelivery.dataStreamArn

`string | valueFrom` · required

The Kinesis data stream records are delivered to. The memory's
execution role must be allowed to write to it.

- references: AwsKinesisStream (`status.outputs.stream_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKinesisStream, name: <that resource's name>, fieldPath: status.outputs.stream_arn}} -- a bare string does not parse

### spec.kinesisDelivery.contentLevel

`string`

How much of each record is delivered: METADATA_ONLY or FULL_CONTENT.
Omitted = AWS default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["METADATA_ONLY","FULL_CONTENT"]}}

### spec.strategies

`[]AwsBedrockAgentCoreMemoryStrategy`

Long-term extraction pipelines.

- rule: custom is required when type is CUSTOM and forbidden otherwise
- rule: reflection_namespace_templates requires type EPISODIC
- rule: each reflection_namespace_templates entry must equal, or be a hierarchical (whole-segment) prefix of, a namespace_templates entry

### spec.strategies[].name

`string` · required

Strategy name in AWS (1-48 characters; letter first, then letters,
digits, underscore). The for_each key on both engines and the key in
the `strategy_ids` output map.

- rule: {"string":{"minLen":"1","pattern":"^[a-zA-Z][a-zA-Z0-9_]{0,47}$"}}

### spec.strategies[].type

`string`

What AWS extracts: SEMANTIC (facts), SUMMARIZATION (session
summaries), USER_PREFERENCE (preferences), EPISODIC (experience
episodes; supports `reflection_namespace_templates`), or CUSTOM
(your own prompts/models via `custom`). Changing the type replaces
the strategy.

- rule: {"string":{"in":["SEMANTIC","SUMMARIZATION","USER_PREFERENCE","EPISODIC","CUSTOM"]}}

### spec.strategies[].description

`string`

Human-readable description.

### spec.strategies[].namespaceTemplates

`[]string` · required

Namespace templates that partition extracted records (e.g.
"/facts/{actorId}" -- at least one). The provider REQUIRES an
explicit set: its deprecated `namespaces` twin and this field are an
exactly-one pair, and the living surface is this one.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.strategies[].custom

`AwsBedrockAgentCoreMemoryCustomStrategy`

Prompt/model overrides - required when type is CUSTOM, forbidden
otherwise.

- rule: reflection is required when type is EPISODIC_OVERRIDE and forbidden otherwise
- rule: extraction cannot be set when type is SUMMARY_OVERRIDE

### spec.strategies[].custom.type

`string`

Which built-in shape is being overridden: SEMANTIC_OVERRIDE,
SUMMARY_OVERRIDE, USER_PREFERENCE_OVERRIDE, EPISODIC_OVERRIDE
(requires `reflection`), or SELF_MANAGED. Changing it replaces the
strategy.

- rule: {"string":{"in":["SEMANTIC_OVERRIDE","SUMMARY_OVERRIDE","USER_PREFERENCE_OVERRIDE","EPISODIC_OVERRIDE","SELF_MANAGED"]}}

### spec.strategies[].custom.extraction

`AwsBedrockAgentCoreMemoryPromptOverride`

Override the extraction step (how raw events become candidate
records). AWS rejects it on SUMMARY_OVERRIDE (summaries have no
extraction step).

### spec.strategies[].custom.extraction.appendToPrompt

`string` · required

Text appended to AWS's built-in prompt for this step.

- rule: {"string":{"minLen":"1"}}

### spec.strategies[].custom.extraction.modelId

`string` · required

Bedrock model that runs this step (a foundation-model ID like
"anthropic.claude-3-5-sonnet-20241022-v2:0").

- rule: {"string":{"minLen":"1"}}

### spec.strategies[].custom.consolidation

`AwsBedrockAgentCoreMemoryPromptOverride`

Override the consolidation step (how candidates merge into stored
records).

### spec.strategies[].custom.consolidation.appendToPrompt

`string` · required

Text appended to AWS's built-in prompt for this step.

- rule: {"string":{"minLen":"1"}}

### spec.strategies[].custom.consolidation.modelId

`string` · required

Bedrock model that runs this step (a foundation-model ID like
"anthropic.claude-3-5-sonnet-20241022-v2:0").

- rule: {"string":{"minLen":"1"}}

### spec.strategies[].custom.reflection

`AwsBedrockAgentCoreMemoryReflectionOverride`

Override the EPISODIC reflection step - required exactly when type
is EPISODIC_OVERRIDE.

### spec.strategies[].custom.reflection.appendToPrompt

`string` · required

Text appended to AWS's built-in reflection prompt.

- rule: {"string":{"minLen":"1"}}

### spec.strategies[].custom.reflection.modelId

`string` · required

Bedrock model that runs reflection.

- rule: {"string":{"minLen":"1"}}

### spec.strategies[].custom.reflection.namespaceTemplates

`[]string` · required

Namespace templates reflection records land in (at least one).

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.strategies[].reflectionNamespaceTemplates

`[]string`

Namespace templates for EPISODIC reflection records - only legal on
EPISODIC strategies. AWS requires each reflection namespace to be
"the same as or a hierarchical prefix of the episodic namespace" in
`namespace_templates` (UpdateMemory rejects unrelated roots like
"/reflections/..." server-side; live-caught 2026-08-14). Leave unset
to let the provider mirror the episodic namespaces exactly.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

## Validation Rules

- `strategy_names_unique`: strategies entries must have unique names
- `indexed_keys_unique`: indexed_keys entries must have unique keys

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsBedrockAgentCoreMemory, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.memory_id` | `string` | The unique memory identifier (e.g. "my_memory-AbC1dEf2Gh"). |
| `status.outputs.memory_arn` | `string` | The Amazon Resource Name of the memory - the canonical key for IAM policies and harness memory configuration. |
| `status.outputs.strategy_ids` | `map<string, string>` | Strategy IDs keyed by each `strategies` entry's name. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.encryptionKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.executionRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.kinesisDelivery.dataStreamArn` | AwsKinesisStream | `status.outputs.stream_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsBedrockAgentCoreEvaluation | `spec.harnesses[].memory.memoryArn` | `status.outputs.memory_arn` |

## See Also

- [Overview](../README.md)

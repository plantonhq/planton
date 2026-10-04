# AwsBedrockPrompt

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsBedrockPromptSpec defines the desired configuration for an Amazon
Bedrock prompt - a reusable, versionable prompt definition in Bedrock
Prompt Management. A prompt carries one or more VARIANTS (candidate
formulations, e.g. for A/B comparison); each variant targets a model (or
an agent) with a text or chat template plus inference settings.

The prompt's name is taken from `metadata.name` (letters, digits,
hyphen, underscore; AWS rejects spaces and dots).

This resource manages the prompt's mutable working draft (version
"DRAFT"). Prompts are free to create - model invocations bill when the
prompt is executed.

## Example

```yaml
# Canonical AwsBedrockPrompt example (hack/dev manifest and refgen
# Example source): a two-variant prompt -- a text variant and a chat
# variant with system context, tools, and prompt caching -- with the chat
# variant as the default. Literal ARNs stand in for composed references
# so the offline `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockPrompt
metadata:
  name: support-answer
  id: support-answer
  org: test-org
  env: dev
spec:
  region: us-west-2
  description: Support answer prompt with tool-assisted chat variant
  defaultVariant: chat
  variants:
    - name: text
      modelId: amazon.nova-micro-v1:0
      text:
        text: "Answer the customer question concisely: {{question}}"
        inputVariables:
          - question
      inferenceConfiguration:
        maxTokens: 512
        temperature: 0
        topP: 0.9
      metadata:
        team: support
    - name: chat
      modelId: amazon.nova-lite-v1:0
      chat:
        system:
          - text: You are a concise support assistant.
          - cachePoint: true
        messages:
          - role: user
            text: "Answer the question: {{question}}"
          - role: assistant
            text: "Certainly:"
        inputVariables:
          - question
        toolConfiguration:
          tools:
            - spec:
                name: lookup_order
                description: Look up an order by its ID.
                inputSchema:
                  type: object
                  properties:
                    order_id:
                      type: string
                  required:
                    - order_id
          toolChoice:
            auto: true
      additionalModelRequestFields:
        inferenceConfig:
          topK: 200
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.customerEncryptionKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.defaultVariant` | `string` |  |  |  |
| `spec.variants` | `[]AwsBedrockPromptVariant` | yes |  |  |
| `spec.variants[].name` | `string` | yes |  |  |
| `spec.variants[].modelId` | `string` |  |  |  |
| `spec.variants[].agentAliasArn` | `string \| valueFrom` |  |  |  |
| `spec.variants[].text` | `AwsBedrockPromptTextTemplate` |  |  |  |
| `spec.variants[].text.text` | `string` | yes |  |  |
| `spec.variants[].text.inputVariables` | `[]string` |  |  |  |
| `spec.variants[].text.cachePoint` | `bool` |  |  |  |
| `spec.variants[].chat` | `AwsBedrockPromptChatTemplate` |  |  |  |
| `spec.variants[].chat.messages` | `[]AwsBedrockPromptMessage` | yes |  |  |
| `spec.variants[].chat.messages[].role` | `string` |  |  |  |
| `spec.variants[].chat.messages[].text` | `string` |  |  |  |
| `spec.variants[].chat.messages[].cachePoint` | `bool` |  |  |  |
| `spec.variants[].chat.system` | `[]AwsBedrockPromptSystemBlock` |  |  |  |
| `spec.variants[].chat.system[].text` | `string` |  |  |  |
| `spec.variants[].chat.system[].cachePoint` | `bool` |  |  |  |
| `spec.variants[].chat.inputVariables` | `[]string` |  |  |  |
| `spec.variants[].chat.toolConfiguration` | `AwsBedrockPromptToolConfiguration` |  |  |  |
| `spec.variants[].chat.toolConfiguration.tools` | `[]AwsBedrockPromptTool` | yes |  |  |
| `spec.variants[].chat.toolConfiguration.tools[].spec` | `AwsBedrockPromptToolSpec` |  |  |  |
| `spec.variants[].chat.toolConfiguration.tools[].spec.name` | `string` | yes |  |  |
| `spec.variants[].chat.toolConfiguration.tools[].spec.description` | `string` |  |  |  |
| `spec.variants[].chat.toolConfiguration.tools[].spec.inputSchema` | `object` |  |  |  |
| `spec.variants[].chat.toolConfiguration.tools[].cachePoint` | `bool` |  |  |  |
| `spec.variants[].chat.toolConfiguration.toolChoice` | `AwsBedrockPromptToolChoice` |  |  |  |
| `spec.variants[].chat.toolConfiguration.toolChoice.auto` | `bool` |  |  |  |
| `spec.variants[].chat.toolConfiguration.toolChoice.any` | `bool` |  |  |  |
| `spec.variants[].chat.toolConfiguration.toolChoice.toolName` | `string` |  |  |  |
| `spec.variants[].inferenceConfiguration` | `AwsBedrockPromptInferenceConfiguration` |  |  |  |
| `spec.variants[].inferenceConfiguration.maxTokens` | `int32` |  |  |  |
| `spec.variants[].inferenceConfiguration.stopSequences` | `[]string` |  |  |  |
| `spec.variants[].inferenceConfiguration.temperature` | `double` |  |  |  |
| `spec.variants[].inferenceConfiguration.topP` | `double` |  |  |  |
| `spec.variants[].additionalModelRequestFields` | `object` |  |  |  |
| `spec.variants[].metadata` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the prompt will be created.
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

Human-readable description shown in the Bedrock console. Updates in
place.

- rule: {"string":{"maxLen":"200"}}

### spec.customerEncryptionKeyArn

`string | valueFrom`

Customer-managed KMS key ARN for encrypting the prompt at rest.
Without it, AWS uses a Bedrock-managed key.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.defaultVariant

`string`

Which variant executes when the prompt is invoked without naming one.
Must match a `variants` entry's name.

### spec.variants

`[]AwsBedrockPromptVariant` · required

The prompt's candidate formulations (at least one; AWS caps a prompt
at 3 variants).

- rule: {"repeated":{"minItems":"1"}}
- rule: set exactly one of model_id or agent_alias_arn
- rule: set exactly one of text or chat

### spec.variants[].name

`string` · required

Variant name (1-100 characters; letters, digits, hyphen, underscore).

- rule: {"string":{"minLen":"1","maxLen":"100","pattern":"^([0-9a-zA-Z][_-]?){1,100}$"}}

### spec.variants[].modelId

`string`

The model that executes this variant: a foundation-model ID
("amazon.nova-micro-v1:0"), model ARN, or inference-profile ID/ARN.

### spec.variants[].agentAliasArn

`string | valueFrom`

OR: execute this variant through a Bedrock agent - the ARN of one of
the agent's aliases. Reference an AwsBedrockAgent's `alias_arns`
output map entry, e.g. valueFrom: {kind: AwsBedrockAgent,
name: <agent>, fieldPath: "status.outputs.alias_arns.<alias-name>"}.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.variants[].text

`AwsBedrockPromptTextTemplate`

Single-string template with {{variable}} placeholders.

### spec.variants[].text.text

`string` · required

The template body. Reference variables as {{name}} and declare them
in input_variables.

- rule: {"string":{"minLen":"1"}}

### spec.variants[].text.inputVariables

`[]string`

Names of the {{variables}} the template expects.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.variants[].text.cachePoint

`bool`

Cache everything up to this point of the prompt across invocations
(prompt caching; the model must support it).

### spec.variants[].chat

`AwsBedrockPromptChatTemplate`

Multi-turn chat template (messages, system context, tools).

### spec.variants[].chat.messages

`[]AwsBedrockPromptMessage` · required

The conversation turns (at least one).

- rule: {"repeated":{"minItems":"1"}}
- rule: set exactly one of text or cache_point

### spec.variants[].chat.messages[].role

`string`

Who speaks this turn.

- rule: {"string":{"in":["user","assistant"]}}

### spec.variants[].chat.messages[].text

`string`

The turn's text (may contain {{variables}}).

### spec.variants[].chat.messages[].cachePoint

`bool`

OR: mark a prompt-caching checkpoint at this turn instead of text.

### spec.variants[].chat.system

`[]AwsBedrockPromptSystemBlock`

System-level context blocks prepended to the conversation.

- rule: set exactly one of text or cache_point

### spec.variants[].chat.system[].text

`string`

System instruction text.

### spec.variants[].chat.system[].cachePoint

`bool`

OR: mark a prompt-caching checkpoint.

### spec.variants[].chat.inputVariables

`[]string`

Names of the {{variables}} used across the messages (at most 20).

- rule: {"repeated":{"maxItems":"20","items":{"string":{"minLen":"1"}}}}

### spec.variants[].chat.toolConfiguration

`AwsBedrockPromptToolConfiguration`

Tools the model may call while executing this prompt.

### spec.variants[].chat.toolConfiguration.tools

`[]AwsBedrockPromptTool` · required

The tool catalog (at least one entry).

- rule: {"repeated":{"minItems":"1"}}
- rule: set exactly one of spec or cache_point

### spec.variants[].chat.toolConfiguration.tools[].spec

`AwsBedrockPromptToolSpec`

The tool's callable specification.

### spec.variants[].chat.toolConfiguration.tools[].spec.name

`string` · required

Tool name the model calls.

- rule: {"string":{"minLen":"1"}}

### spec.variants[].chat.toolConfiguration.tools[].spec.description

`string`

What the tool does, shown to the model.

### spec.variants[].chat.toolConfiguration.tools[].spec.inputSchema

`object`

The tool's input parameters as a JSON Schema document.

### spec.variants[].chat.toolConfiguration.tools[].cachePoint

`bool`

OR: mark a prompt-caching checkpoint in the tool list.

### spec.variants[].chat.toolConfiguration.toolChoice

`AwsBedrockPromptToolChoice`

How the model chooses among the tools. Omitted = the model decides
freely.

- rule: set exactly one of auto, any, or tool_name

### spec.variants[].chat.toolConfiguration.toolChoice.auto

`bool`

The model decides whether and which tool to call.

### spec.variants[].chat.toolConfiguration.toolChoice.any

`bool`

The model MUST call some tool (its choice which).

### spec.variants[].chat.toolConfiguration.toolChoice.toolName

`string`

The model MUST call this named tool.

### spec.variants[].inferenceConfiguration

`AwsBedrockPromptInferenceConfiguration`

Model sampling parameters for this variant.

### spec.variants[].inferenceConfiguration.maxTokens

`int32` · optional (explicit presence)

Maximum number of tokens in the model response.

- rule: {"int32":{"gte":1}}

### spec.variants[].inferenceConfiguration.stopSequences

`[]string`

Sequences that stop generation when the model emits them.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.variants[].inferenceConfiguration.temperature

`double` · optional (explicit presence)

Sampling temperature (0 = deterministic, 1 = most random). Bedrock
stores it as a 32-bit float: a value that is not float32-exact (0.9,
0.7) reads back slightly widened (0.8999999761581421) from the API -
harmless for deploys (state keeps your value), visible only when
importing a pre-existing prompt.

- rule: {"double":{"lte":1,"gte":0}}

### spec.variants[].inferenceConfiguration.topP

`double` · optional (explicit presence)

Nucleus sampling - consider tokens covering the top P probability
mass (0-1). Stored by Bedrock as a 32-bit float; see temperature for
the read-back precision note.

- rule: {"double":{"lte":1,"gte":0}}

### spec.variants[].additionalModelRequestFields

`object`

Model-specific request parameters outside the standard inference set,
as a JSON document (e.g. {"top_k": 200} for Anthropic models).

### spec.variants[].metadata

`map<string, string>`

Arbitrary key-value annotations stored on the variant (at most 50;
keys 1-128 characters, values up to 1024).

- rule: {"map":{"maxPairs":"50","keys":{"string":{"minLen":"1","maxLen":"128"}},"values":{"string":{"maxLen":"1024"}}}}

## Validation Rules

- `variant_names_unique`: variants entries must have unique names
- `default_variant_exists`: default_variant must match one of the variants' names

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsBedrockPrompt, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.prompt_id` | `string` | The unique prompt identifier (e.g. "1A2BC3DEFG"). |
| `status.outputs.prompt_arn` | `string` | The Amazon Resource Name of the prompt - the value a flow's prompt node consumes. |
| `status.outputs.draft_version` | `string` | The prompt's mutable working version - always the literal "DRAFT". |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.customerEncryptionKeyArn` | AwsKmsKey | `status.outputs.key_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsBedrockFlow | `spec.definition.nodes[].prompt.promptArn` | `status.outputs.prompt_arn` |

## See Also

- [Overview](../README.md)

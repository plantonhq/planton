# AwsBedrockFlow

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsBedrockFlowSpec defines the desired configuration for an Amazon
Bedrock flow - a node graph that orchestrates prompts, agents, knowledge
bases, Lambda functions, and control-flow logic into one invocable
generative-AI pipeline.

The flow's name is taken from `metadata.name` (letters, digits, hyphen,
underscore; AWS rejects spaces and dots).

A flow is a directed graph: `definition.nodes` declares the steps (every
flow needs one Input and at least one Output node) and
`definition.connections` wires node outputs to node inputs. AWS
validates the graph server-side (unreachable nodes, type mismatches,
missing connections) at create/update time. Flows are free to create -
the nodes' model/agent invocations bill at runtime.

## Example

```yaml
# Canonical AwsBedrockFlow example (hack/dev manifest and refgen Example
# source): a routing pipeline exercising the node classes -- input,
# inline prompt, condition, knowledge base, inline code, retrieval, and
# outputs -- with data and conditional connections. Literal ARNs/ids
# stand in for composed references so the offline `tofu plan` renders
# every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockFlow
metadata:
  name: support-router
  id: support-router
  org: test-org
  env: dev
spec:
  region: us-west-2
  description: Classify a request, answer from docs or summarize inline
  executionRoleArn:
    value: arn:aws:iam::123456789012:role/bedrock-flow-role
  definition:
    nodes:
      - name: FlowInput
        type: Input
        outputs:
          - name: document
            type: String
      - name: Classify
        type: Prompt
        inputs:
          - name: input
            expression: $.data
            type: String
        outputs:
          - name: modelCompletion
            type: String
        prompt:
          inline:
            modelId: amazon.nova-micro-v1:0
            text:
              text: "Classify this request as billing or docs: {{input}}"
              inputVariables:
                - input
            inferenceConfiguration:
              temperature: 0
      - name: Route
        type: Condition
        inputs:
          - name: category
            expression: $.data
            type: String
        condition:
          conditions:
            - name: billing
              expression: category == "billing"
            - name: default
      - name: AskKb
        type: KnowledgeBase
        inputs:
          - name: retrievalQuery
            expression: $.data
            type: String
        outputs:
          - name: outputText
            type: String
        knowledgeBase:
          knowledgeBaseId:
            value: KBEXAMPLE01
          modelId: amazon.nova-micro-v1:0
          numberOfResults: 3
      - name: Summarize
        type: InlineCode
        inputs:
          - name: doc
            expression: $.data
            type: String
        outputs:
          - name: summary
            type: String
        inlineCode:
          code: |
            def handler(doc):
                return doc[:200]
      - name: FetchContext
        type: Retrieval
        inputs:
          - name: retrievalPath
            expression: $.data
            type: String
        outputs:
          - name: retrievedData
            type: String
        retrieval:
          bucketName:
            value: support-context-bucket
      - name: FlowOutput
        type: Output
        inputs:
          - name: document
            expression: $.data
            type: String
      - name: RawOutput
        type: Output
        inputs:
          - name: document
            expression: $.data
            type: String
    connections:
      - name: InToClassify
        source: FlowInput
        target: Classify
        data:
          sourceOutput: document
          targetInput: input
      - name: ClassifyToRoute
        source: Classify
        target: Route
        data:
          sourceOutput: modelCompletion
          targetInput: category
      - name: ClassifyToKb
        source: Classify
        target: AskKb
        data:
          sourceOutput: modelCompletion
          targetInput: retrievalQuery
      - name: RouteBilling
        source: Route
        target: AskKb
        conditional:
          condition: billing
      - name: ClassifyToRaw
        source: Classify
        target: RawOutput
        data:
          sourceOutput: modelCompletion
          targetInput: document
      - name: RouteDefault
        source: Route
        target: RawOutput
        conditional:
          condition: default
      - name: KbToOut
        source: AskKb
        target: FlowOutput
        data:
          sourceOutput: outputText
          targetInput: document
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.executionRoleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.customerEncryptionKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.definition` | `AwsBedrockFlowDefinition` |  |  |  |
| `spec.definition.nodes` | `[]AwsBedrockFlowNode` | yes |  |  |
| `spec.definition.nodes[].name` | `string` |  |  |  |
| `spec.definition.nodes[].type` | `string` |  |  |  |
| `spec.definition.nodes[].inputs` | `[]AwsBedrockFlowNodeInput` |  |  |  |
| `spec.definition.nodes[].inputs[].name` | `string` |  |  |  |
| `spec.definition.nodes[].inputs[].expression` | `string` | yes |  |  |
| `spec.definition.nodes[].inputs[].type` | `string` |  |  |  |
| `spec.definition.nodes[].inputs[].category` | `string` |  |  |  |
| `spec.definition.nodes[].outputs` | `[]AwsBedrockFlowNodeOutput` |  |  |  |
| `spec.definition.nodes[].outputs[].name` | `string` |  |  |  |
| `spec.definition.nodes[].outputs[].type` | `string` |  |  |  |
| `spec.definition.nodes[].agent` | `AwsBedrockFlowAgentNode` |  |  |  |
| `spec.definition.nodes[].agent.agentAliasArn` | `string \| valueFrom` | yes |  |  |
| `spec.definition.nodes[].prompt` | `AwsBedrockFlowPromptNode` |  |  |  |
| `spec.definition.nodes[].prompt.promptArn` | `string \| valueFrom` |  |  | AwsBedrockPrompt (`status.outputs.prompt_arn`) |
| `spec.definition.nodes[].prompt.inline` | `AwsBedrockFlowInlinePrompt` |  |  |  |
| `spec.definition.nodes[].prompt.inline.modelId` | `string` | yes |  |  |
| `spec.definition.nodes[].prompt.inline.text` | `AwsBedrockFlowPromptTextTemplate` |  |  |  |
| `spec.definition.nodes[].prompt.inline.text.text` | `string` | yes |  |  |
| `spec.definition.nodes[].prompt.inline.text.inputVariables` | `[]string` |  |  |  |
| `spec.definition.nodes[].prompt.inline.text.cachePoint` | `bool` |  |  |  |
| `spec.definition.nodes[].prompt.inline.chat` | `AwsBedrockFlowPromptChatTemplate` |  |  |  |
| `spec.definition.nodes[].prompt.inline.chat.messages` | `[]AwsBedrockFlowPromptMessage` | yes |  |  |
| `spec.definition.nodes[].prompt.inline.chat.messages[].role` | `string` |  |  |  |
| `spec.definition.nodes[].prompt.inline.chat.messages[].text` | `string` |  |  |  |
| `spec.definition.nodes[].prompt.inline.chat.messages[].cachePoint` | `bool` |  |  |  |
| `spec.definition.nodes[].prompt.inline.chat.system` | `[]AwsBedrockFlowPromptSystemBlock` |  |  |  |
| `spec.definition.nodes[].prompt.inline.chat.system[].text` | `string` |  |  |  |
| `spec.definition.nodes[].prompt.inline.chat.system[].cachePoint` | `bool` |  |  |  |
| `spec.definition.nodes[].prompt.inline.chat.inputVariables` | `[]string` |  |  |  |
| `spec.definition.nodes[].prompt.inline.chat.toolConfiguration` | `AwsBedrockFlowPromptToolConfiguration` |  |  |  |
| `spec.definition.nodes[].prompt.inline.chat.toolConfiguration.tools` | `[]AwsBedrockFlowPromptTool` | yes |  |  |
| `spec.definition.nodes[].prompt.inline.chat.toolConfiguration.tools[].spec` | `AwsBedrockFlowPromptToolSpec` |  |  |  |
| `spec.definition.nodes[].prompt.inline.chat.toolConfiguration.tools[].spec.name` | `string` | yes |  |  |
| `spec.definition.nodes[].prompt.inline.chat.toolConfiguration.tools[].spec.description` | `string` |  |  |  |
| `spec.definition.nodes[].prompt.inline.chat.toolConfiguration.tools[].spec.inputSchema` | `object` |  |  |  |
| `spec.definition.nodes[].prompt.inline.chat.toolConfiguration.tools[].cachePoint` | `bool` |  |  |  |
| `spec.definition.nodes[].prompt.inline.chat.toolConfiguration.toolChoice` | `AwsBedrockFlowPromptToolChoice` |  |  |  |
| `spec.definition.nodes[].prompt.inline.chat.toolConfiguration.toolChoice.auto` | `bool` |  |  |  |
| `spec.definition.nodes[].prompt.inline.chat.toolConfiguration.toolChoice.any` | `bool` |  |  |  |
| `spec.definition.nodes[].prompt.inline.chat.toolConfiguration.toolChoice.toolName` | `string` |  |  |  |
| `spec.definition.nodes[].prompt.inline.inferenceConfiguration` | `AwsBedrockFlowInferenceConfiguration` |  |  |  |
| `spec.definition.nodes[].prompt.inline.inferenceConfiguration.maxTokens` | `int32` |  |  |  |
| `spec.definition.nodes[].prompt.inline.inferenceConfiguration.stopSequences` | `[]string` |  |  |  |
| `spec.definition.nodes[].prompt.inline.inferenceConfiguration.temperature` | `double` |  |  |  |
| `spec.definition.nodes[].prompt.inline.inferenceConfiguration.topP` | `double` |  |  |  |
| `spec.definition.nodes[].prompt.inline.additionalModelRequestFields` | `object` |  |  |  |
| `spec.definition.nodes[].prompt.guardrail` | `AwsBedrockFlowGuardrail` |  |  |  |
| `spec.definition.nodes[].prompt.guardrail.guardrailId` | `string \| valueFrom` | yes |  | AwsBedrockGuardrail (`status.outputs.guardrail_id`) |
| `spec.definition.nodes[].prompt.guardrail.version` | `string` | yes |  |  |
| `spec.definition.nodes[].knowledgeBase` | `AwsBedrockFlowKnowledgeBaseNode` |  |  |  |
| `spec.definition.nodes[].knowledgeBase.knowledgeBaseId` | `string \| valueFrom` | yes |  | AwsBedrockKnowledgeBase (`status.outputs.knowledge_base_id`) |
| `spec.definition.nodes[].knowledgeBase.modelId` | `string` |  |  |  |
| `spec.definition.nodes[].knowledgeBase.numberOfResults` | `int32` |  |  |  |
| `spec.definition.nodes[].knowledgeBase.guardrail` | `AwsBedrockFlowGuardrail` |  |  |  |
| `spec.definition.nodes[].knowledgeBase.guardrail.guardrailId` | `string \| valueFrom` | yes |  | AwsBedrockGuardrail (`status.outputs.guardrail_id`) |
| `spec.definition.nodes[].knowledgeBase.guardrail.version` | `string` | yes |  |  |
| `spec.definition.nodes[].knowledgeBase.inferenceConfiguration` | `AwsBedrockFlowInferenceConfiguration` |  |  |  |
| `spec.definition.nodes[].knowledgeBase.inferenceConfiguration.maxTokens` | `int32` |  |  |  |
| `spec.definition.nodes[].knowledgeBase.inferenceConfiguration.stopSequences` | `[]string` |  |  |  |
| `spec.definition.nodes[].knowledgeBase.inferenceConfiguration.temperature` | `double` |  |  |  |
| `spec.definition.nodes[].knowledgeBase.inferenceConfiguration.topP` | `double` |  |  |  |
| `spec.definition.nodes[].lambdaFunction` | `AwsBedrockFlowLambdaNode` |  |  |  |
| `spec.definition.nodes[].lambdaFunction.lambdaArn` | `string \| valueFrom` | yes |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.definition.nodes[].lex` | `AwsBedrockFlowLexNode` |  |  |  |
| `spec.definition.nodes[].lex.botAliasArn` | `string` |  |  |  |
| `spec.definition.nodes[].lex.localeId` | `string` | yes |  |  |
| `spec.definition.nodes[].condition` | `AwsBedrockFlowConditionNode` |  |  |  |
| `spec.definition.nodes[].condition.conditions` | `[]AwsBedrockFlowCondition` | yes |  |  |
| `spec.definition.nodes[].condition.conditions[].name` | `string` |  |  |  |
| `spec.definition.nodes[].condition.conditions[].expression` | `string` | yes |  |  |
| `spec.definition.nodes[].inlineCode` | `AwsBedrockFlowInlineCodeNode` |  |  |  |
| `spec.definition.nodes[].inlineCode.code` | `string` | yes |  |  |
| `spec.definition.nodes[].retrieval` | `AwsBedrockFlowRetrievalNode` |  |  |  |
| `spec.definition.nodes[].retrieval.bucketName` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.definition.nodes[].storage` | `AwsBedrockFlowStorageNode` |  |  |  |
| `spec.definition.nodes[].storage.bucketName` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.definition.connections` | `[]AwsBedrockFlowConnection` |  |  |  |
| `spec.definition.connections[].name` | `string` |  |  |  |
| `spec.definition.connections[].source` | `string` |  |  |  |
| `spec.definition.connections[].target` | `string` |  |  |  |
| `spec.definition.connections[].data` | `AwsBedrockFlowDataConnection` |  |  |  |
| `spec.definition.connections[].data.sourceOutput` | `string` |  |  |  |
| `spec.definition.connections[].data.targetInput` | `string` |  |  |  |
| `spec.definition.connections[].conditional` | `AwsBedrockFlowConditionalConnection` |  |  |  |
| `spec.definition.connections[].conditional.condition` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the flow will be created.
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

Human-readable description shown in the Bedrock console (1-200
characters when set). Updates in place.

- rule: {"string":{"maxLen":"200"}}

### spec.executionRoleArn

`string | valueFrom` · required

IAM role the Bedrock service assumes to run the flow (invoke models,
agents, knowledge bases, and Lambdas the nodes reference). The role
must trust bedrock.amazonaws.com.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.customerEncryptionKeyArn

`string | valueFrom`

Customer-managed KMS key ARN for encrypting the flow at rest.
Without it, AWS uses a Bedrock-managed key.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.definition

`AwsBedrockFlowDefinition`

The node graph. A flow created without a definition is an empty shell
you fill in later (the console flow builder writes the same shape).

- rule: node names must be unique
- rule: connection names must be unique

### spec.definition.nodes

`[]AwsBedrockFlowNode` · required

The graph's steps.

- rule: {"repeated":{"minItems":"1"}}
- rule: configure exactly the arm matching the node type (Agent -> agent, Prompt -> prompt, KnowledgeBase -> knowledge_base, LambdaFunction -> lambda_function, Lex -> lex, Condition -> condition, InlineCode -> inline_code, Retrieval -> retrieval, Storage -> storage); Input/Output/Iterator/Collector/Loop nodes take no configuration arm

### spec.definition.nodes[].name

`string`

Node name (1-50 characters; must start with a letter;
letters/digits/single underscores) - the identity connections
reference.

- rule: {"string":{"pattern":"^[a-zA-Z]([_]?[0-9a-zA-Z]){0,49}$"}}

### spec.definition.nodes[].type

`string`

The node class.

- rule: {"string":{"in":["Input","Output","Prompt","Agent","KnowledgeBase","LambdaFunction","Lex","Condition","InlineCode","Iterator","Collector","Retrieval","Storage","Loop","LoopInput","LoopController"]}}

### spec.definition.nodes[].inputs

`[]AwsBedrockFlowNodeInput`

The node's input sockets (at most 20). Each maps an upstream value
into this node via an expression.

- rule: {"repeated":{"maxItems":"20"}}

### spec.definition.nodes[].inputs[].name

`string`

Socket name (1-50 characters; must start with a letter;
letters/digits/single underscores).

- rule: {"string":{"pattern":"^[a-zA-Z]([_]?[0-9a-zA-Z]){0,49}$"}}

### spec.definition.nodes[].inputs[].expression

`string` · required

Expression selecting the value from the connected upstream output
(1-64 characters; "$.data" passes it through unchanged).

- rule: {"string":{"minLen":"1","maxLen":"64"}}

### spec.definition.nodes[].inputs[].type

`string`

The value type this socket expects.

- rule: {"string":{"in":["String","Number","Boolean","Object","Array"]}}

### spec.definition.nodes[].inputs[].category

`string`

Loop-specific role of this input (only on Loop-family nodes):
LoopCondition, ReturnValueToLoopStart, or ExitLoop.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["LoopCondition","ReturnValueToLoopStart","ExitLoop"]}}

### spec.definition.nodes[].outputs

`[]AwsBedrockFlowNodeOutput`

The node's output sockets (at most 5).

- rule: {"repeated":{"maxItems":"5"}}

### spec.definition.nodes[].outputs[].name

`string`

Socket name (1-50 characters; must start with a letter;
letters/digits/single underscores).

- rule: {"string":{"pattern":"^[a-zA-Z]([_]?[0-9a-zA-Z]){0,49}$"}}

### spec.definition.nodes[].outputs[].type

`string`

The value type this socket produces.

- rule: {"string":{"in":["String","Number","Boolean","Object","Array"]}}

### spec.definition.nodes[].agent

`AwsBedrockFlowAgentNode`

Agent node: delegate to a Bedrock agent.

### spec.definition.nodes[].agent.agentAliasArn

`string | valueFrom` · required

ARN of the agent's alias to invoke. Reference an AwsBedrockAgent's
`alias_arns` output map entry, e.g. valueFrom: {kind: AwsBedrockAgent,
name: <agent>, fieldPath: "status.outputs.alias_arns.<alias-name>"}.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.definition.nodes[].prompt

`AwsBedrockFlowPromptNode`

Prompt node: run a prompt (from Prompt Management or inline).

- rule: set exactly one of prompt_arn or inline

### spec.definition.nodes[].prompt.promptArn

`string | valueFrom`

Run a prompt resource from Prompt Management.

- references: AwsBedrockPrompt (`status.outputs.prompt_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsBedrockPrompt, name: <that resource's name>, fieldPath: status.outputs.prompt_arn}} -- a bare string does not parse

### spec.definition.nodes[].prompt.inline

`AwsBedrockFlowInlinePrompt`

OR: define the prompt inline on the node.

- rule: set exactly one of text or chat

### spec.definition.nodes[].prompt.inline.modelId

`string` · required

The model that executes the prompt: a foundation-model ID, model ARN,
or inference-profile ID/ARN.

- rule: {"string":{"minLen":"1"}}

### spec.definition.nodes[].prompt.inline.text

`AwsBedrockFlowPromptTextTemplate`

Single-string template with {{variable}} placeholders.

### spec.definition.nodes[].prompt.inline.text.text

`string` · required

The template body. Reference variables as {{name}} and declare them
in input_variables.

- rule: {"string":{"minLen":"1"}}

### spec.definition.nodes[].prompt.inline.text.inputVariables

`[]string`

Names of the {{variables}} the template expects.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.definition.nodes[].prompt.inline.text.cachePoint

`bool`

Cache everything up to this point of the prompt across invocations
(prompt caching; the model must support it).

### spec.definition.nodes[].prompt.inline.chat

`AwsBedrockFlowPromptChatTemplate`

Multi-turn chat template.

### spec.definition.nodes[].prompt.inline.chat.messages

`[]AwsBedrockFlowPromptMessage` · required

The conversation turns (at least one).

- rule: {"repeated":{"minItems":"1"}}
- rule: set exactly one of text or cache_point

### spec.definition.nodes[].prompt.inline.chat.messages[].role

`string`

Who speaks this turn.

- rule: {"string":{"in":["user","assistant"]}}

### spec.definition.nodes[].prompt.inline.chat.messages[].text

`string`

The turn's text (may contain {{variables}}).

### spec.definition.nodes[].prompt.inline.chat.messages[].cachePoint

`bool`

OR: mark a prompt-caching checkpoint at this turn instead of text.

### spec.definition.nodes[].prompt.inline.chat.system

`[]AwsBedrockFlowPromptSystemBlock`

System-level context blocks prepended to the conversation.

- rule: set exactly one of text or cache_point

### spec.definition.nodes[].prompt.inline.chat.system[].text

`string`

System instruction text.

### spec.definition.nodes[].prompt.inline.chat.system[].cachePoint

`bool`

OR: mark a prompt-caching checkpoint.

### spec.definition.nodes[].prompt.inline.chat.inputVariables

`[]string`

Names of the {{variables}} used across the messages (at most 20).

- rule: {"repeated":{"maxItems":"20","items":{"string":{"minLen":"1"}}}}

### spec.definition.nodes[].prompt.inline.chat.toolConfiguration

`AwsBedrockFlowPromptToolConfiguration`

Tools the model may call while executing this prompt.

### spec.definition.nodes[].prompt.inline.chat.toolConfiguration.tools

`[]AwsBedrockFlowPromptTool` · required

The tool catalog (at least one entry).

- rule: {"repeated":{"minItems":"1"}}
- rule: set exactly one of spec or cache_point

### spec.definition.nodes[].prompt.inline.chat.toolConfiguration.tools[].spec

`AwsBedrockFlowPromptToolSpec`

The tool's callable specification.

### spec.definition.nodes[].prompt.inline.chat.toolConfiguration.tools[].spec.name

`string` · required

Tool name the model calls.

- rule: {"string":{"minLen":"1"}}

### spec.definition.nodes[].prompt.inline.chat.toolConfiguration.tools[].spec.description

`string`

What the tool does, shown to the model.

### spec.definition.nodes[].prompt.inline.chat.toolConfiguration.tools[].spec.inputSchema

`object`

The tool's input parameters as a JSON Schema document.

### spec.definition.nodes[].prompt.inline.chat.toolConfiguration.tools[].cachePoint

`bool`

OR: mark a prompt-caching checkpoint in the tool list.

### spec.definition.nodes[].prompt.inline.chat.toolConfiguration.toolChoice

`AwsBedrockFlowPromptToolChoice`

How the model chooses among the tools. Omitted = the model decides
freely.

- rule: set exactly one of auto, any, or tool_name

### spec.definition.nodes[].prompt.inline.chat.toolConfiguration.toolChoice.auto

`bool`

The model decides whether and which tool to call.

### spec.definition.nodes[].prompt.inline.chat.toolConfiguration.toolChoice.any

`bool`

The model MUST call some tool (its choice which).

### spec.definition.nodes[].prompt.inline.chat.toolConfiguration.toolChoice.toolName

`string`

The model MUST call this named tool.

### spec.definition.nodes[].prompt.inline.inferenceConfiguration

`AwsBedrockFlowInferenceConfiguration`

Model sampling parameters.

### spec.definition.nodes[].prompt.inline.inferenceConfiguration.maxTokens

`int32` · optional (explicit presence)

Maximum number of tokens in the model response.

- rule: {"int32":{"gte":1}}

### spec.definition.nodes[].prompt.inline.inferenceConfiguration.stopSequences

`[]string`

Sequences that stop generation when the model emits them.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.definition.nodes[].prompt.inline.inferenceConfiguration.temperature

`double` · optional (explicit presence)

Sampling temperature (0 = deterministic, 1 = most random). Bedrock
stores it as a 32-bit float: a value that is not float32-exact (0.2,
0.9) reads back slightly widened (0.20000000298023224) from the API -
harmless for deploys (state keeps your value), visible only when
importing a pre-existing flow.

- rule: {"double":{"lte":1,"gte":0}}

### spec.definition.nodes[].prompt.inline.inferenceConfiguration.topP

`double` · optional (explicit presence)

Nucleus sampling - consider tokens covering the top P probability
mass (0-1). Same float32 storage as temperature.

- rule: {"double":{"lte":1,"gte":0}}

### spec.definition.nodes[].prompt.inline.additionalModelRequestFields

`object`

Model-specific request parameters outside the standard inference set,
as a JSON document.

### spec.definition.nodes[].prompt.guardrail

`AwsBedrockFlowGuardrail`

Apply a guardrail to the prompt's model I/O.

### spec.definition.nodes[].prompt.guardrail.guardrailId

`string | valueFrom` · required

The guardrail to apply.

- references: AwsBedrockGuardrail (`status.outputs.guardrail_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsBedrockGuardrail, name: <that resource's name>, fieldPath: status.outputs.guardrail_id}} -- a bare string does not parse

### spec.definition.nodes[].prompt.guardrail.version

`string` · required

The guardrail version to pin: "DRAFT" or a published number.

- rule: {"string":{"minLen":"1"}}

### spec.definition.nodes[].knowledgeBase

`AwsBedrockFlowKnowledgeBaseNode`

KnowledgeBase node: query a knowledge base.

### spec.definition.nodes[].knowledgeBase.knowledgeBaseId

`string | valueFrom` · required

The knowledge base to query.

- references: AwsBedrockKnowledgeBase (`status.outputs.knowledge_base_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsBedrockKnowledgeBase, name: <that resource's name>, fieldPath: status.outputs.knowledge_base_id}} -- a bare string does not parse

### spec.definition.nodes[].knowledgeBase.modelId

`string`

Model that generates the answer from retrieved chunks. Omitted = the
node returns the retrieved chunks themselves.

### spec.definition.nodes[].knowledgeBase.numberOfResults

`int32`

How many chunks to retrieve (1-100). Omitted = AWS default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":100,"gte":1}}

### spec.definition.nodes[].knowledgeBase.guardrail

`AwsBedrockFlowGuardrail`

Apply a guardrail to the generation.

### spec.definition.nodes[].knowledgeBase.guardrail.guardrailId

`string | valueFrom` · required

The guardrail to apply.

- references: AwsBedrockGuardrail (`status.outputs.guardrail_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsBedrockGuardrail, name: <that resource's name>, fieldPath: status.outputs.guardrail_id}} -- a bare string does not parse

### spec.definition.nodes[].knowledgeBase.guardrail.version

`string` · required

The guardrail version to pin: "DRAFT" or a published number.

- rule: {"string":{"minLen":"1"}}

### spec.definition.nodes[].knowledgeBase.inferenceConfiguration

`AwsBedrockFlowInferenceConfiguration`

Model sampling parameters for the generation.

### spec.definition.nodes[].knowledgeBase.inferenceConfiguration.maxTokens

`int32` · optional (explicit presence)

Maximum number of tokens in the model response.

- rule: {"int32":{"gte":1}}

### spec.definition.nodes[].knowledgeBase.inferenceConfiguration.stopSequences

`[]string`

Sequences that stop generation when the model emits them.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.definition.nodes[].knowledgeBase.inferenceConfiguration.temperature

`double` · optional (explicit presence)

Sampling temperature (0 = deterministic, 1 = most random). Bedrock
stores it as a 32-bit float: a value that is not float32-exact (0.2,
0.9) reads back slightly widened (0.20000000298023224) from the API -
harmless for deploys (state keeps your value), visible only when
importing a pre-existing flow.

- rule: {"double":{"lte":1,"gte":0}}

### spec.definition.nodes[].knowledgeBase.inferenceConfiguration.topP

`double` · optional (explicit presence)

Nucleus sampling - consider tokens covering the top P probability
mass (0-1). Same float32 storage as temperature.

- rule: {"double":{"lte":1,"gte":0}}

### spec.definition.nodes[].lambdaFunction

`AwsBedrockFlowLambdaNode`

LambdaFunction node: call a Lambda.

### spec.definition.nodes[].lambdaFunction.lambdaArn

`string | valueFrom` · required

The Lambda to invoke.

- references: AwsLambda (`status.outputs.function_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.definition.nodes[].lex

`AwsBedrockFlowLexNode`

Lex node: classify the input with an Amazon Lex bot.

### spec.definition.nodes[].lex.botAliasArn

`string`

ARN of the Lex bot ALIAS to invoke.

- rule: {"string":{"pattern":"^arn:aws[a-z-]*:lex:[a-z0-9-]+:[0-9]{12}:bot-alias/.+$"}}

### spec.definition.nodes[].lex.localeId

`string` · required

Locale the bot converses in (e.g. "en_US").

- rule: {"string":{"minLen":"1"}}

### spec.definition.nodes[].condition

`AwsBedrockFlowConditionNode`

Condition node: branch on input values.

- rule: condition names must be unique

### spec.definition.nodes[].condition.conditions

`[]AwsBedrockFlowCondition` · required

The branch conditions (1-5). Conditional connections reference these
by name; an edge with condition "default" fires when none match.

- rule: {"repeated":{"minItems":"1","maxItems":"5"}}

### spec.definition.nodes[].condition.conditions[].name

`string`

Condition name (1-50 characters; must start with a letter).

- rule: {"string":{"pattern":"^[a-zA-Z]([_]?[0-9a-zA-Z]){0,49}$"}}

### spec.definition.nodes[].condition.conditions[].expression

`string` · required

Boolean expression over the node's inputs (1-64 characters, e.g.
"categoryLetter == \"A\""). Omit for the default (else) condition.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"1","maxLen":"64"}}

### spec.definition.nodes[].inlineCode

`AwsBedrockFlowInlineCodeNode`

InlineCode node: run a code snippet.

### spec.definition.nodes[].inlineCode.code

`string` · required

The code body (up to 5,000,000 characters).

- rule: {"string":{"minLen":"1","maxLen":"5000000"}}

### spec.definition.nodes[].retrieval

`AwsBedrockFlowRetrievalNode`

Retrieval node: fetch data from S3.

### spec.definition.nodes[].retrieval.bucketName

`string | valueFrom` · required

Bucket to read from.

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.definition.nodes[].storage

`AwsBedrockFlowStorageNode`

Storage node: write data to S3.

### spec.definition.nodes[].storage.bucketName

`string | valueFrom` · required

Bucket to write to.

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.definition.connections

`[]AwsBedrockFlowConnection`

Directed edges wiring node outputs to node inputs.

- rule: set exactly one of data or conditional

### spec.definition.connections[].name

`string`

Connection name (1-100 characters; must start with a letter;
letters/digits/single underscores).

- rule: {"string":{"pattern":"^[a-zA-Z]([_]?[0-9a-zA-Z]){1,100}$"}}

### spec.definition.connections[].source

`string`

Source node name.

- rule: {"string":{"pattern":"^[a-zA-Z]([_]?[0-9a-zA-Z]){1,50}$"}}

### spec.definition.connections[].target

`string`

Target node name.

- rule: {"string":{"pattern":"^[a-zA-Z]([_]?[0-9a-zA-Z]){1,50}$"}}

### spec.definition.connections[].data

`AwsBedrockFlowDataConnection`

Data edge: move source's output into target's input.

### spec.definition.connections[].data.sourceOutput

`string`

Output socket on the source node.

- rule: {"string":{"pattern":"^[a-zA-Z]([_]?[0-9a-zA-Z]){1,50}$"}}

### spec.definition.connections[].data.targetInput

`string`

Input socket on the target node.

- rule: {"string":{"pattern":"^[a-zA-Z]([_]?[0-9a-zA-Z]){1,50}$"}}

### spec.definition.connections[].conditional

`AwsBedrockFlowConditionalConnection`

Conditional edge: activate when the source Condition node's named
condition fires.

### spec.definition.connections[].conditional.condition

`string`

The condition name on the source Condition node (use "default" for
the else-branch).

- rule: {"string":{"pattern":"^[a-zA-Z]([_]?[0-9a-zA-Z]){1,50}$"}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsBedrockFlow, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.flow_id` | `string` | The unique flow identifier (e.g. "ABCDEFGHIJ"). |
| `status.outputs.flow_arn` | `string` | The Amazon Resource Name of the flow - the canonical key for IAM policies and flow invocations. |
| `status.outputs.draft_version` | `string` | The flow's mutable working version - always the literal "DRAFT". |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.executionRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.customerEncryptionKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.definition.nodes[].prompt.promptArn` | AwsBedrockPrompt | `status.outputs.prompt_arn` |
| `spec.definition.nodes[].prompt.guardrail.guardrailId` | AwsBedrockGuardrail | `status.outputs.guardrail_id` |
| `spec.definition.nodes[].knowledgeBase.knowledgeBaseId` | AwsBedrockKnowledgeBase | `status.outputs.knowledge_base_id` |
| `spec.definition.nodes[].knowledgeBase.guardrail.guardrailId` | AwsBedrockGuardrail | `status.outputs.guardrail_id` |
| `spec.definition.nodes[].lambdaFunction.lambdaArn` | AwsLambda | `status.outputs.function_arn` |
| `spec.definition.nodes[].retrieval.bucketName` | AwsS3Bucket | `status.outputs.bucket_id` |
| `spec.definition.nodes[].storage.bucketName` | AwsS3Bucket | `status.outputs.bucket_id` |

## See Also

- [Overview](../README.md)

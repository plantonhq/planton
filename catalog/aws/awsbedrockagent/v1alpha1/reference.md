# AwsBedrockAgent

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

AwsBedrockAgentSpec defines the desired configuration for an Amazon
Bedrock agent - a foundation-model-powered assistant that reasons over a
user request, calls the tools exposed through its action groups, retrieves
context from associated knowledge bases, and (in multi-agent mode)
delegates to collaborator agents.

The agent's name is taken from `metadata.name` (letters, digits, hyphen,
underscore; AWS caps it at 100 characters and rejects spaces/dots).

An agent always has a mutable working version, the literal "DRAFT".
Action groups, collaborators, and knowledge-base associations attach to
the draft; each `aliases` entry snapshots the draft into an immutable
numbered version and points at it. Editing the spec changes only the
draft - live traffic through an alias is unaffected until the alias is
re-pointed or recreated. Agents are free to create; AWS bills per model
invocation at runtime.

After every change the modules let AWS "prepare" the agent (compile the
draft for serving) and wait until it reaches PREPARED - the deploy is not
done while the agent is still compiling.

## Example

```yaml
# Canonical AwsBedrockAgent example (hack/dev manifest and refgen Example
# source): a supervisor agent exercising every folded satellite -- action
# groups (custom function-schema plus a reserved capability), a published
# alias, a collaborator, a knowledge-base association -- plus guardrail,
# memory, and a prompt override. Literal ARNs/ids stand in for composed
# references so the offline `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockAgent
metadata:
  name: support-agent
  id: support-agent
  org: test-org
  env: dev
spec:
  region: us-west-2
  description: Customer support agent with order tools and docs retrieval
  foundationModel: amazon.nova-micro-v1:0
  agentResourceRoleArn:
    value: arn:aws:iam::123456789012:role/bedrock-agent-role
  instruction: >-
    You are a customer support agent. Answer order questions using the
    order tools, retrieve product documentation when asked how something
    works, and delegate billing questions to the billing collaborator.
  idleSessionTtlSeconds: 900
  agentCollaboration: SUPERVISOR
  guardrail:
    guardrailId:
      value: gr-examplegr01
    version: "1"
  memory:
    storageDays: 30
    maxRecentSessions: 5
  promptOverride:
    promptConfigurations:
      - promptType: ORCHESTRATION
        basePromptTemplate: "$instruction$ $question$ $agent_scratchpad$"
        promptState: ENABLED
        inferenceConfiguration:
          maxLength: 2048
          temperature: 0
          topP: 1
  actionGroups:
    - name: orders
      description: Look up and update customer orders.
      executor:
        returnControl: true
      functionSchema:
        functions:
          - name: get_order
            description: Look up one order by its ID.
            parameters:
              - name: order_id
                type: string
                description: The order identifier.
                required: true
    - name: user-input
      parentActionGroupSignature: AMAZON.UserInput
  aliases:
    - name: live
      description: Production serving endpoint
  collaborators:
    - name: billing
      collaborationInstruction: Handle all billing and invoice questions.
      collaboratorAliasArn:
        value: arn:aws:bedrock:us-west-2:123456789012:agent-alias/AGENT12345/ALIAS12345
      relayConversationHistory: TO_COLLABORATOR
  knowledgeBaseAssociations:
    - name: docs
      knowledgeBaseId:
        value: KBEXAMPLE01
      description: Product documentation for troubleshooting answers.
      state: ENABLED
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.foundationModel` | `string` | yes |  |  |
| `spec.agentResourceRoleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.instruction` | `string` | yes |  |  |
| `spec.idleSessionTtlSeconds` | `int32` |  |  |  |
| `spec.customerEncryptionKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.agentCollaboration` | `string` |  |  |  |
| `spec.guardrail` | `AwsBedrockAgentGuardrail` |  |  |  |
| `spec.guardrail.guardrailId` | `string \| valueFrom` | yes |  | AwsBedrockGuardrail (`status.outputs.guardrail_id`) |
| `spec.guardrail.version` | `string` | yes |  |  |
| `spec.memory` | `AwsBedrockAgentMemory` |  |  |  |
| `spec.memory.enabled` | `bool` |  | `true` |  |
| `spec.memory.storageDays` | `int32` |  |  |  |
| `spec.memory.maxRecentSessions` | `int32` |  |  |  |
| `spec.promptOverride` | `AwsBedrockAgentPromptOverride` |  |  |  |
| `spec.promptOverride.overrideLambda` | `string \| valueFrom` |  |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.promptOverride.promptConfigurations` | `[]AwsBedrockAgentPromptConfiguration` | yes |  |  |
| `spec.promptOverride.promptConfigurations[].promptType` | `string` |  |  |  |
| `spec.promptOverride.promptConfigurations[].basePromptTemplate` | `string` | yes |  |  |
| `spec.promptOverride.promptConfigurations[].parserMode` | `string` |  |  |  |
| `spec.promptOverride.promptConfigurations[].promptState` | `string` |  |  |  |
| `spec.promptOverride.promptConfigurations[].inferenceConfiguration` | `AwsBedrockAgentInferenceConfiguration` |  |  |  |
| `spec.promptOverride.promptConfigurations[].inferenceConfiguration.maxLength` | `int32` |  |  |  |
| `spec.promptOverride.promptConfigurations[].inferenceConfiguration.stopSequences` | `[]string` |  |  |  |
| `spec.promptOverride.promptConfigurations[].inferenceConfiguration.temperature` | `double` |  |  |  |
| `spec.promptOverride.promptConfigurations[].inferenceConfiguration.topK` | `int32` |  |  |  |
| `spec.promptOverride.promptConfigurations[].inferenceConfiguration.topP` | `double` |  |  |  |
| `spec.actionGroups` | `[]AwsBedrockAgentActionGroup` |  |  |  |
| `spec.actionGroups[].name` | `string` | yes |  |  |
| `spec.actionGroups[].description` | `string` |  |  |  |
| `spec.actionGroups[].state` | `string` |  |  |  |
| `spec.actionGroups[].parentActionGroupSignature` | `string` |  |  |  |
| `spec.actionGroups[].executor` | `AwsBedrockAgentActionGroupExecutor` |  |  |  |
| `spec.actionGroups[].executor.lambda` | `string \| valueFrom` |  |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.actionGroups[].executor.returnControl` | `bool` |  |  |  |
| `spec.actionGroups[].apiSchema` | `AwsBedrockAgentApiSchema` |  |  |  |
| `spec.actionGroups[].apiSchema.payload` | `string` |  |  |  |
| `spec.actionGroups[].apiSchema.s3` | `AwsBedrockAgentApiSchemaS3` |  |  |  |
| `spec.actionGroups[].apiSchema.s3.bucketName` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.actionGroups[].apiSchema.s3.objectKey` | `string` | yes |  |  |
| `spec.actionGroups[].functionSchema` | `AwsBedrockAgentFunctionSchema` |  |  |  |
| `spec.actionGroups[].functionSchema.functions` | `[]AwsBedrockAgentFunction` | yes |  |  |
| `spec.actionGroups[].functionSchema.functions[].name` | `string` | yes |  |  |
| `spec.actionGroups[].functionSchema.functions[].description` | `string` |  |  |  |
| `spec.actionGroups[].functionSchema.functions[].parameters` | `[]AwsBedrockAgentFunctionParameter` |  |  |  |
| `spec.actionGroups[].functionSchema.functions[].parameters[].name` | `string` | yes |  |  |
| `spec.actionGroups[].functionSchema.functions[].parameters[].type` | `string` |  |  |  |
| `spec.actionGroups[].functionSchema.functions[].parameters[].description` | `string` |  |  |  |
| `spec.actionGroups[].functionSchema.functions[].parameters[].required` | `bool` |  |  |  |
| `spec.aliases` | `[]AwsBedrockAgentAlias` |  |  |  |
| `spec.aliases[].name` | `string` | yes |  |  |
| `spec.aliases[].description` | `string` |  |  |  |
| `spec.aliases[].routing` | `AwsBedrockAgentAliasRouting` |  |  |  |
| `spec.aliases[].routing.agentVersion` | `string` |  |  |  |
| `spec.aliases[].routing.provisionedThroughput` | `string \| valueFrom` |  |  | AwsBedrockProvisionedThroughput (`status.outputs.provisioned_model_arn`) |
| `spec.collaborators` | `[]AwsBedrockAgentCollaborator` |  |  |  |
| `spec.collaborators[].name` | `string` | yes |  |  |
| `spec.collaborators[].collaborationInstruction` | `string` | yes |  |  |
| `spec.collaborators[].collaboratorAliasArn` | `string \| valueFrom` | yes |  |  |
| `spec.collaborators[].relayConversationHistory` | `string` |  |  |  |
| `spec.knowledgeBaseAssociations` | `[]AwsBedrockAgentKnowledgeBaseAssociation` |  |  |  |
| `spec.knowledgeBaseAssociations[].name` | `string` | yes |  |  |
| `spec.knowledgeBaseAssociations[].knowledgeBaseId` | `string \| valueFrom` | yes |  | AwsBedrockKnowledgeBase (`status.outputs.knowledge_base_id`) |
| `spec.knowledgeBaseAssociations[].description` | `string` | yes |  |  |
| `spec.knowledgeBaseAssociations[].state` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the agent will be created.
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

Human-readable description shown in the Bedrock console (1-200
characters when set). Updates in place.

- rule: {"string":{"maxLen":"200"}}

### spec.foundationModel

`string` · required

The model that powers the agent. Accepts a foundation-model ID
("amazon.nova-micro-v1:0"), a foundation-model ARN, or an inference
profile ID/ARN (per-application cost attribution or cross-region
routing). The account must have access to the model in this region -
auto-enabled AWS models (Amazon, Mistral, Meta) work immediately;
marketplace models need an AwsBedrockModelAccess agreement first.

- rule: {"string":{"minLen":"1"}}

### spec.agentResourceRoleArn

`string | valueFrom` · required

IAM role the Bedrock service assumes to operate the agent (invoke the
model, call action-group Lambdas, query knowledge bases). The role
must trust bedrock.amazonaws.com. Changing the role replaces the
agent.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.instruction

`string` · required

Natural-language instructions that tell the agent what it does and how
to behave (40-20000 characters). This is the agent's system-level
briefing - write it the way you would onboard a new team member.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"40","maxLen":"20000"}}

### spec.idleSessionTtlSeconds

`int32`

How long (seconds) AWS keeps an idle conversation session alive before
discarding its context (60-5400). Omitted = AWS default (600).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":5400,"gte":60}}

### spec.customerEncryptionKeyArn

`string | valueFrom`

Customer-managed KMS key ARN for encrypting the agent's resources at
rest. Without it, AWS uses a Bedrock-managed key.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.agentCollaboration

`string`

Multi-agent collaboration mode. SUPERVISOR lets this agent plan and
delegate to its `collaborators`; SUPERVISOR_ROUTER routes each request
to exactly one collaborator without a planning step; DISABLED (the AWS
default when omitted) makes this a standalone agent.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["SUPERVISOR","SUPERVISOR_ROUTER","DISABLED"]}}

### spec.guardrail

`AwsBedrockAgentGuardrail`

Attach a Bedrock guardrail so every model input/output the agent
produces is evaluated against its content-safety policies.

### spec.guardrail.guardrailId

`string | valueFrom` · required

The guardrail to apply to every model input/output of this agent.

- references: AwsBedrockGuardrail (`status.outputs.guardrail_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsBedrockGuardrail, name: <that resource's name>, fieldPath: status.outputs.guardrail_id}} -- a bare string does not parse

### spec.guardrail.version

`string` · required

The guardrail version to pin: "DRAFT" or a published number ("1",
"2", ...). Production agents should pin a published version so
guardrail draft edits never change live behavior.

- rule: {"string":{"minLen":"1"}}

### spec.memory

`AwsBedrockAgentMemory`

Enable conversation memory - AWS summarizes each session and carries
the summaries into later sessions with the same memory id.

### spec.memory.enabled

`bool` · optional (explicit presence)

Whether session-summary memory is on. Unset means on: declaring the
block has always meant enabling memory, and this switch lets a manifest
say the opposite out loud.

- default: `true`

### spec.memory.storageDays

`int32`

How many days AWS retains session summaries (AWS documents 0-365;
omitted = AWS default, 30).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":365,"gte":1}}

### spec.memory.maxRecentSessions

`int32`

How many of the most recent sessions AWS summarizes into the memory
context. Omitted = AWS default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.promptOverride

`AwsBedrockAgentPromptOverride`

Override the prompt templates the agent uses at specific steps of its
orchestration (advanced tuning; most agents never need this).

- rule: each prompt_type may be overridden at most once
- rule: override_lambda is required when any prompt configuration sets parser_mode = OVERRIDDEN

### spec.promptOverride.overrideLambda

`string | valueFrom`

Lambda that parses the raw model output for every step whose
`parser_mode` is OVERRIDDEN. Required when any configuration overrides
its parser.

- references: AwsLambda (`status.outputs.function_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.promptOverride.promptConfigurations

`[]AwsBedrockAgentPromptConfiguration` · required

One override per orchestration step (at least one). Only steps listed
here are overridden; the agent keeps AWS defaults for the rest (the
modules mark each entry as an OVERRIDDEN template - authoring an entry
IS the override).

- rule: {"repeated":{"minItems":"1"}}

### spec.promptOverride.promptConfigurations[].promptType

`string`

The orchestration step to override.

- rule: {"string":{"in":["PRE_PROCESSING","ORCHESTRATION","POST_PROCESSING","KNOWLEDGE_BASE_RESPONSE_GENERATION","MEMORY_SUMMARIZATION"]}}

### spec.promptOverride.promptConfigurations[].basePromptTemplate

`string` · required

The replacement prompt template. Use AWS's placeholder variables
(e.g. $question$, $instruction$) - AWS validates the template
server-side at prepare time.

- rule: {"string":{"minLen":"1"}}

### spec.promptOverride.promptConfigurations[].parserMode

`string`

How the model's raw output for this step is parsed: DEFAULT uses
AWS's parser; OVERRIDDEN routes it through `override_lambda`. Omitted
= AWS default (DEFAULT).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["DEFAULT","OVERRIDDEN"]}}

### spec.promptOverride.promptConfigurations[].promptState

`string`

Whether this step runs at all: ENABLED or DISABLED (e.g. disable
PRE_PROCESSING entirely). Omitted = AWS default (ENABLED).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ENABLED","DISABLED"]}}

### spec.promptOverride.promptConfigurations[].inferenceConfiguration

`AwsBedrockAgentInferenceConfiguration`

Model sampling parameters for this step.

### spec.promptOverride.promptConfigurations[].inferenceConfiguration.maxLength

`int32` · optional (explicit presence)

Maximum number of tokens in the model response.

- rule: {"int32":{"gte":1}}

### spec.promptOverride.promptConfigurations[].inferenceConfiguration.stopSequences

`[]string`

Sequences that stop generation when the model emits them.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.promptOverride.promptConfigurations[].inferenceConfiguration.temperature

`double` · optional (explicit presence)

Sampling temperature (0 = deterministic, 1 = most random).

- rule: {"double":{"lte":1,"gte":0}}

### spec.promptOverride.promptConfigurations[].inferenceConfiguration.topK

`int32` · optional (explicit presence)

Sample only from the K most likely next tokens.

- rule: {"int32":{"gte":0}}

### spec.promptOverride.promptConfigurations[].inferenceConfiguration.topP

`double` · optional (explicit presence)

Nucleus sampling - consider tokens covering the top P probability
mass (0-1).

- rule: {"double":{"lte":1,"gte":0}}

### spec.actionGroups

`[]AwsBedrockAgentActionGroup`

Tools the agent can call, grouped by purpose. Each group exposes either
a Lambda-backed API (OpenAPI or function schema) or a reserved AWS
capability (user-input elicitation, code interpreter).

- rule: description cannot be set on a reserved group (parent_action_group_signature)
- rule: custom groups require executor plus exactly one of api_schema/function_schema; reserved groups (parent_action_group_signature) take neither

### spec.actionGroups[].name

`string` · required

Stable name for this action group (1-100 characters; letters, digits,
hyphen, underscore). The for_each key on both engines and the key in
the `action_group_ids` output map.

- rule: {"string":{"minLen":"1","maxLen":"100","pattern":"^([0-9a-zA-Z][_-]?){1,100}$"}}

### spec.actionGroups[].description

`string`

What the group's operations do, shown to the model as tool context
(1-200 characters). AWS forbids it on reserved groups.

- rule: {"string":{"maxLen":"200"}}

### spec.actionGroups[].state

`string`

ENABLED (default) or DISABLED - a disabled group stays attached but
the model cannot call it. Omitted = AWS default (ENABLED).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ENABLED","DISABLED"]}}

### spec.actionGroups[].parentActionGroupSignature

`string`

Reserved AWS capability signature. AMAZON.UserInput lets the agent ask
the user clarifying questions; AMAZON.CodeInterpreter lets it run
code; the ANTHROPIC.* signatures enable Claude computer-use tools.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["AMAZON.UserInput","AMAZON.CodeInterpreter","ANTHROPIC.Computer","ANTHROPIC.Bash","ANTHROPIC.TextEditor"]}}

### spec.actionGroups[].executor

`AwsBedrockAgentActionGroupExecutor`

Where custom operations execute.

- rule: executor must set exactly one of lambda or return_control

### spec.actionGroups[].executor.lambda

`string | valueFrom`

Lambda function that fulfills the group's operations.

- references: AwsLambda (`status.outputs.function_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.actionGroups[].executor.returnControl

`bool`

Skip server-side execution entirely - the agent returns the tool call
to YOUR application, which executes it and continues the conversation
(AWS's RETURN_CONTROL method, the only custom-control method AWS
defines; the modules send that constant).

### spec.actionGroups[].apiSchema

`AwsBedrockAgentApiSchema`

Describe operations with an OpenAPI schema (inline or in S3).

- rule: api_schema must set exactly one of payload or s3

### spec.actionGroups[].apiSchema.payload

`string`

Inline OpenAPI 3 schema body (JSON or YAML).

### spec.actionGroups[].apiSchema.s3

`AwsBedrockAgentApiSchemaS3`

S3 object holding the OpenAPI schema.

### spec.actionGroups[].apiSchema.s3.bucketName

`string | valueFrom` · required

Bucket holding the schema object.

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.actionGroups[].apiSchema.s3.objectKey

`string` · required

Object key of the schema document.

- rule: {"string":{"minLen":"1"}}

### spec.actionGroups[].functionSchema

`AwsBedrockAgentFunctionSchema`

Describe operations as typed function signatures (simpler than
OpenAPI for direct tool definitions).

- rule: function names must be unique

### spec.actionGroups[].functionSchema.functions

`[]AwsBedrockAgentFunction` · required

The callable functions (at least one).

- rule: {"repeated":{"minItems":"1"}}
- rule: parameter names must be unique

### spec.actionGroups[].functionSchema.functions[].name

`string` · required

Function name the model calls (1-100 characters; letters, digits,
hyphen, underscore).

- rule: {"string":{"minLen":"1","maxLen":"100","pattern":"^([0-9a-zA-Z][_-]?){1,100}$"}}

### spec.actionGroups[].functionSchema.functions[].description

`string`

What the function does, shown to the model (1-1200 characters when
set) - the better the description, the better the model's tool use.

- rule: {"string":{"maxLen":"1200"}}

### spec.actionGroups[].functionSchema.functions[].parameters

`[]AwsBedrockAgentFunctionParameter`

The function's parameters.

### spec.actionGroups[].functionSchema.functions[].parameters[].name

`string` · required

Parameter name (1-100 characters; letters, digits, hyphen,
underscore).

- rule: {"string":{"minLen":"1","maxLen":"100","pattern":"^([0-9a-zA-Z][_-]?){1,100}$"}}

### spec.actionGroups[].functionSchema.functions[].parameters[].type

`string`

Parameter type.

- rule: {"string":{"in":["string","number","integer","boolean","array"]}}

### spec.actionGroups[].functionSchema.functions[].parameters[].description

`string`

What the parameter means, shown to the model (1-500 characters when
set).

- rule: {"string":{"maxLen":"500"}}

### spec.actionGroups[].functionSchema.functions[].parameters[].required

`bool`

Whether the model must supply this parameter.

### spec.aliases

`[]AwsBedrockAgentAlias`

Immutable serving endpoints. Each entry snapshots the current draft
into a new numbered agent version at create time (or pins an explicit
version through `routing`) and exposes it under a stable alias ID.

### spec.aliases[].name

`string` · required

Stable name for this alias (1-100 characters; letters, digits, hyphen,
underscore). The for_each key on both engines and the key in the
`alias_ids`/`alias_arns` output maps; also the alias name in AWS.

- rule: {"string":{"minLen":"1","maxLen":"100","pattern":"^([0-9a-zA-Z][_-]?){1,100}$"}}

### spec.aliases[].description

`string`

Human-readable description (1-200 characters when set).

- rule: {"string":{"maxLen":"200"}}

### spec.aliases[].routing

`AwsBedrockAgentAliasRouting`

Pin the alias to an existing agent version and/or provisioned
throughput. Omitted = AWS snapshots the current DRAFT into a NEW
numbered version at alias creation (the common case).

- rule: routing must set agent_version and/or provisioned_throughput

### spec.aliases[].routing.agentVersion

`string`

The numbered agent version this alias serves ("1", "2", ...).

### spec.aliases[].routing.provisionedThroughput

`string | valueFrom`

Provisioned-throughput capacity the alias serves traffic through
(reserved model units bought for this agent's model).

- references: AwsBedrockProvisionedThroughput (`status.outputs.provisioned_model_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsBedrockProvisionedThroughput, name: <that resource's name>, fieldPath: status.outputs.provisioned_model_arn}} -- a bare string does not parse

### spec.collaborators

`[]AwsBedrockAgentCollaborator`

Agents this supervisor delegates to. Requires `agent_collaboration` to
be SUPERVISOR or SUPERVISOR_ROUTER. Each collaborator is another
Bedrock agent, referenced through one of ITS alias ARNs.

### spec.collaborators[].name

`string` · required

Collaborator name the supervisor uses in its plans (1-100 characters;
letters, digits, hyphen, underscore). The for_each key on both engines
and the key in the `collaborator_ids` output map.

- rule: {"string":{"minLen":"1","maxLen":"100","pattern":"^([0-9a-zA-Z][_-]?){1,100}$"}}

### spec.collaborators[].collaborationInstruction

`string` · required

When the supervisor should delegate to this collaborator (1-4000
characters) - describe its specialty the way you would brief a
dispatcher.

- rule: {"string":{"minLen":"1","maxLen":"4000"}}

### spec.collaborators[].collaboratorAliasArn

`string | valueFrom` · required

ARN of the collaborating agent's ALIAS (not the agent itself) - the
published endpoint this supervisor calls. Reference another
AwsBedrockAgent's `alias_arns` output map entry, e.g.
valueFrom: {kind: AwsBedrockAgent, name: <agent>,
fieldPath: "status.outputs.alias_arns.<alias-name>"}.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.collaborators[].relayConversationHistory

`string`

Whether the conversation history is relayed to the collaborator:
TO_COLLABORATOR or DISABLED. Omitted = AWS default (DISABLED).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["TO_COLLABORATOR","DISABLED"]}}

### spec.knowledgeBaseAssociations

`[]AwsBedrockAgentKnowledgeBaseAssociation`

Knowledge bases the agent queries for retrieval-augmented answers.

### spec.knowledgeBaseAssociations[].name

`string` · required

Stable local key for this entry (1-100 characters; letters, digits,
hyphen, underscore). The for_each key on both engines and the key in
the `associated_knowledge_base_ids` output map - never sent to AWS.

- rule: {"string":{"minLen":"1","maxLen":"100","pattern":"^([0-9a-zA-Z][_-]?){1,100}$"}}

### spec.knowledgeBaseAssociations[].knowledgeBaseId

`string | valueFrom` · required

The knowledge base to associate.

- references: AwsBedrockKnowledgeBase (`status.outputs.knowledge_base_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsBedrockKnowledgeBase, name: <that resource's name>, fieldPath: status.outputs.knowledge_base_id}} -- a bare string does not parse

### spec.knowledgeBaseAssociations[].description

`string` · required

When the agent should query this knowledge base (1-200 characters,
REQUIRED by AWS) - the model reads this to decide when to retrieve.

- rule: {"string":{"minLen":"1","maxLen":"200"}}

### spec.knowledgeBaseAssociations[].state

`string`

ENABLED (default) or DISABLED - a disabled association stays attached
but is never queried. Omitted = AWS default (ENABLED).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ENABLED","DISABLED"]}}

## Validation Rules

- `action_group_names_unique`: action_groups entries must have unique names
- `alias_names_unique`: aliases entries must have unique names
- `collaborator_names_unique`: collaborators entries must have unique names
- `kb_association_names_unique`: knowledge_base_associations entries must have unique names
- `collaborators_require_supervisor_mode`: collaborators require agent_collaboration to be SUPERVISOR or SUPERVISOR_ROUTER

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsBedrockAgent, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.agent_id` | `string` | The unique agent identifier (e.g. "GGRRAED6JP"). |
| `status.outputs.agent_arn` | `string` | The Amazon Resource Name of the agent - the canonical key for IAM policies. |
| `status.outputs.draft_version` | `string` | The agent's mutable working version - always the literal "DRAFT". Serving traffic goes through aliases, never the draft. |
| `status.outputs.alias_ids` | `map<string, string>` | Alias IDs keyed by each `aliases` entry's name. Example: {"live": "66IVY0GUTF"}. |
| `status.outputs.alias_arns` | `map<string, string>` | Alias ARNs keyed by each `aliases` entry's name - the value a supervisor agent's collaborator entry or a flow's agent node consumes. |
| `status.outputs.action_group_ids` | `map<string, string>` | Action group IDs keyed by each `action_groups` entry's name. |
| `status.outputs.collaborator_ids` | `map<string, string>` | Collaborator IDs keyed by each `collaborators` entry's name. |
| `status.outputs.associated_knowledge_base_ids` | `map<string, string>` | Associated knowledge base IDs keyed by each `knowledge_base_associations` entry's name. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.agentResourceRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.customerEncryptionKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.guardrail.guardrailId` | AwsBedrockGuardrail | `status.outputs.guardrail_id` |
| `spec.promptOverride.overrideLambda` | AwsLambda | `status.outputs.function_arn` |
| `spec.actionGroups[].executor.lambda` | AwsLambda | `status.outputs.function_arn` |
| `spec.actionGroups[].apiSchema.s3.bucketName` | AwsS3Bucket | `status.outputs.bucket_id` |
| `spec.aliases[].routing.provisionedThroughput` | AwsBedrockProvisionedThroughput | `status.outputs.provisioned_model_arn` |
| `spec.knowledgeBaseAssociations[].knowledgeBaseId` | AwsBedrockKnowledgeBase | `status.outputs.knowledge_base_id` |

## See Also

- [Overview](../README.md)

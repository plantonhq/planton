# AWS Bedrock Flow

Deploys an Amazon Bedrock flow — a node graph that orchestrates prompts, agents, knowledge bases, Lambda functions, and control-flow logic into one invocable generative-AI pipeline, declared as YAML instead of drawn in the visual builder. The graph wires Input/Output boundaries, inline or referenced prompt nodes, agent delegation by alias, knowledge-base queries with optional generation, Lambda and Lex calls, condition-based branching, inline Python, and S3 retrieval/storage; data connections move typed values between node sockets and conditional connections branch on a Condition node's expressions. The cost drivers are the nodes' invocations at runtime — model tokens, agent calls, Lambda duration — not the flow object itself.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Bedrock Flow** — the flow with its DRAFT definition: every node from `definition.nodes` and every edge from `definition.connections`, validated server-side by AWS at create/update time
- **Flow Encryption** — configured only when `customerEncryptionKeyArn` is set; without it AWS encrypts the flow with a Bedrock-managed key

The module creates the DRAFT definition only — flows do not auto-prepare, so a NotPrepared status after deploy is healthy; preparing and versioning happen at invocation setup.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with Bedrock flow permissions (`bedrock:CreateFlow` and its read/update/delete siblings, plus `iam:PassRole` on the execution role). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **Bedrock flows available in the target region.**
- **An IAM role trusting `bedrock.amazonaws.com`** — referenced by `executionRoleArn`, with invoke permissions on everything the nodes touch: models for prompt and knowledge-base nodes, agent aliases for agent nodes, knowledge bases, and Lambdas. One role aggregates all of it.
- **The referenced resources deployed first** — agent aliases, knowledge bases, prompts, and Lambdas a node names must exist before the flow's server-side validation passes (only for graphs using those node classes).

## Deploy

### Console

Open the deployment store, find **AWS Bedrock Flow**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: region, execution role, and the node graph. Start from the **Summarize Pipeline** preset in the [Presets](#presets) tab for the minimal Input→Prompt→Output skeleton.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockFlow
metadata:
  name: summarizer
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  executionRoleArn:
    valueFrom:
      kind: AwsIamRole
      name: bedrock-flow-role
      fieldPath: status.outputs.role_arn
  definition:
    nodes:
      - name: FlowInput
        type: Input
        outputs:
          - name: document
            type: String
      - name: Summarize
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
              text: "Summarize in one sentence: {{input}}"
              inputVariables:
                - input
      - name: FlowOutput
        type: Output
        inputs:
          - name: document
            expression: $.data
            type: String
    connections:
      - name: InToPrompt
        source: FlowInput
        target: Summarize
        data:
          sourceOutput: document
          targetInput: input
      - name: PromptToOut
        source: Summarize
        target: FlowOutput
        data:
          sourceOutput: modelCompletion
          targetInput: document
```

```shell
planton apply -f flow.yaml
```

This creates a three-node summarization pipeline: input document in, one Nova Micro prompt node, summary out. An Infra Job tracks the provisioning in real time.

### InfraChart

When the flow deploys alongside the resources its nodes reference, wire them via ValueFromRef — here a prompt node consuming a managed prompt:

```yaml
spec:
  region: us-west-2
  executionRoleArn:
    valueFrom:
      kind: AwsIamRole
      name: bedrock-flow-role
      fieldPath: status.outputs.role_arn
  definition:
    nodes:
      - name: Summarize
        type: Prompt
        inputs:
          - name: input
            expression: $.data
            type: String
        outputs:
          - name: modelCompletion
            type: String
        prompt:
          promptArn:
            valueFrom:
              kind: AwsBedrockPrompt
              name: summarize-prompt
              fieldPath: status.outputs.prompt_arn
```

The InfraPipeline resolves the dependency graph, deploys the role and prompt first, then creates the flow that references them.

## Key Configuration

These are the most important decisions when configuring a flow. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Build minimal, extend validated** — AWS validates the topology server-side with precise, named error classes (UnreachableNode, MismatchedNodeInputType, MissingConnectionConfiguration). Deploy the Input→Prompt→Output skeleton first, then grow the graph a node at a time — debugging a twenty-node graph's first create is much slower than growing a working one.

**Socket types are contracts** — a String output feeding an Object-typed input fails validation at create. Align the node IO types before reaching for expressions; `$.data` passes the whole upstream value through, and narrower JSONPath expressions select fields from Object-typed sockets.

**Inline prompts or referenced prompts** — a prompt node either defines its template inline or references an AwsBedrockPrompt by `promptArn`. Inline keeps the flow self-contained; a referenced prompt is versioned and shared across flows, so template edits ship without touching the graph. The inline tree deliberately mirrors the prompt kind's — moving between the two is a mechanical change.

**The execution role aggregates every node's permissions** — one role covers model invocation, agent-alias invocation, knowledge-base retrieval, and Lambda invoke for the entire graph. Audit it whenever you add a node class; a missing permission surfaces at run time, not at deploy.

**Condition nodes need a `default` arm** — a Condition node declares 1–5 named conditions; conditional connections reference them by name, and the edge whose condition is `default` is the else-branch. Branches that never reach an Output node fail validation.

**The Loop family is typed but not configurable yet** — AWS accepts Loop, LoopInput, and LoopController nodes, but the pinned provider cannot express their configuration beyond the structural markers. Graphs needing loop bodies wait on the provider, not on this component.

**Pin node guardrail versions** — prompt and knowledge-base nodes accept a guardrail at a version; pin a published number in production so guardrail draft edits never change the flow's live behavior.

**Temperature reads back widened on import only** — Bedrock stores node `temperature`/`topP` as 32-bit floats, so importing a pre-existing flow with a value like 0.2 shows a one-time reconcile back to your manifest's value. Normal deploys never see this.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `executionRoleArn` | `status.outputs.role_arn` |
| **AwsKmsKey** | `customerEncryptionKeyArn` | `status.outputs.key_arn` |
| **AwsBedrockAgent** | `definition.nodes[].agent.agentAliasArn` | `status.outputs.alias_arns.<alias-name>` |
| **AwsBedrockPrompt** | `definition.nodes[].prompt.promptArn` | `status.outputs.prompt_arn` |
| **AwsBedrockKnowledgeBase** | `definition.nodes[].knowledgeBase.knowledgeBaseId` | `status.outputs.knowledge_base_id` |
| **AwsBedrockGuardrail** | `definition.nodes[].prompt.guardrail.guardrailId`, `definition.nodes[].knowledgeBase.guardrail.guardrailId` | `status.outputs.guardrail_id` |
| **AwsLambda** | `definition.nodes[].lambdaFunction.lambdaArn` | `status.outputs.function_arn` |
| **AwsS3Bucket** | `definition.nodes[].retrieval.bucketName`, `definition.nodes[].storage.bucketName` | `status.outputs.bucket_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `flow_id` | The unique flow identifier | `InvokeFlow` calls from applications |
| `flow_arn` | The flow's ARN — the canonical key for IAM policies and invocations | Policies scoping who can invoke or manage the flow |

`draft_version` is the constant "DRAFT" — a record of the mutable working version, not a composition input.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Single-step pipeline** — Input→Prompt→Output with an inline template: the smallest deployable flow and the right validation baseline before growing the graph. Start from the **Summarize Pipeline** preset.

**Classify and route** — a Prompt node classifies the input, a Condition node branches on the classification, and each branch ends in its own Output (with a `default` arm catching everything unmatched). This is the shape for triage pipelines where different categories deserve different models or handlers. Start from the **Classify and Route** preset.

**Composed AI stack** — the flow as the orchestration layer over independently deployed pieces: managed prompts for versioned templates, an agent alias for tool-using steps, a knowledge base for retrieval. Each piece ships and versions on its own; the chart's reference wiring orders the deployments.

## Works With

- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the execution role the flow assumes, wired via `executionRoleArn`
- [**AWS Bedrock Prompt**](/infra-catalog/aws-bedrock-prompt) — versioned templates prompt nodes reference via `promptArn`
- [**AWS Bedrock Agent**](/infra-catalog/aws-bedrock-agent) — agents flow nodes delegate to through their alias ARNs
- [**AWS Bedrock Knowledge Base**](/infra-catalog/aws-bedrock-knowledge-base) — retrieval sources knowledge-base nodes query
- [**AWS Bedrock Guardrail**](/infra-catalog/aws-bedrock-guardrail) — content-safety policies pinned onto prompt and knowledge-base nodes
- [**AWS Lambda**](/infra-catalog/aws-lambda) — functions LambdaFunction nodes invoke
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) — buckets Retrieval and Storage nodes read from and write to
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — customer-managed encryption via `customerEncryptionKeyArn`

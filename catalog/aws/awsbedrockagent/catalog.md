# AWS Bedrock Agent

Deploys an Amazon Bedrock agent — a foundation-model-powered assistant that reasons over requests, calls tools, retrieves knowledge, and delegates to collaborator agents — with its action groups, aliases, collaborators, and knowledge-base associations folded into one component. Every spec edit changes the agent's mutable DRAFT version; each `aliases` entry snapshots the draft into an immutable numbered version and serves it, so live traffic never shifts until an alias is created or re-pointed. The cost driver is model invocation at serving time — token usage on the foundation model, plus knowledge-base retrieval and guardrail evaluation when those are attached.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Bedrock Agent** — the agent on the model you name (a foundation-model ID or ARN, or an inference-profile ID or ARN), with optional guardrail attachment, session-summary memory, and per-step prompt-template overrides
- **Action Groups** — created only when `actionGroups` entries exist: Lambda-backed or return-control tools described by an OpenAPI or function schema, or reserved AWS capabilities (`AMAZON.UserInput`, `AMAZON.CodeInterpreter`, the `ANTHROPIC.*` computer-use signatures)
- **Collaborators** — created only when `collaborators` entries exist: delegation links from this supervisor to other agents' aliases
- **Knowledge Base Associations** — created only when `knowledgeBaseAssociations` entries exist: retrieval attachments the agent queries for grounded answers
- **Aliases** — created only when `aliases` entries exist: immutable serving endpoints, created after every other satellite so each snapshot captures the fully assembled draft

Every satellite change triggers an AWS prepare cycle; the deploy completes only when the agent reaches PREPARED.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with Bedrock agent permissions (`bedrock:CreateAgent` and its satellite read/update/delete siblings) plus `iam:PassRole` on the agent's role. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **The CreateAgent allowlist** — AWS put Bedrock Agents Classic into maintenance mode: accounts that did not call `CreateAgent` or `InvokeInlineAgent` in the 12 months before 2026-07-30 get `AccessDeniedException` on new creates, with no exception process. Existing agents and every other API (including invocation and satellite management) stay available to all accounts.
- **An IAM role trusting `bedrock.amazonaws.com`** — with model-invocation permissions, and Lambda invoke permissions when action groups use Lambda executors (referenced by `agentResourceRoleArn`). AWS validates the role when the agent prepares, so a missing permission fails the deploy, not the first invocation.
- **Access to the foundation model** — auto-enabled AWS models (Nova, Titan, Mistral, Meta) work immediately; Anthropic models need the account's use-case form on file; marketplace models need an agreement first (only when `foundationModel` names one).

## Deploy

### Console

Open the deployment store, find **AWS Bedrock Agent**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: region and model, the agent role, instructions, and the satellite collections. Start from the **Support Agent with Tools** preset in the [Presets](#presets) tab for the workhorse shape: one tool group plus a serving alias.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockAgent
metadata:
  name: support-agent
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  foundationModel: amazon.nova-micro-v1:0
  agentResourceRoleArn:
    valueFrom:
      kind: AwsIamRole
      name: bedrock-agent-role
      fieldPath: status.outputs.role_arn
  instruction: >-
    You are a customer support agent. Answer order questions using the
    order tools and keep responses short and accurate.
  actionGroups:
    - name: orders
      executor:
        returnControl: true
      functionSchema:
        functions:
          - name: get_order
            parameters:
              - name: order_id
                type: string
                required: true
  aliases:
    - name: live
```

```shell
planton apply -f agent.yaml
```

This creates a Nova Micro support agent with one return-control tool group and a `live` alias that snapshots the assembled draft into version 1. An Infra Job tracks the provisioning in real time.

### InfraChart

When the agent deploys alongside its role and knowledge base in one chart, wire the references via ValueFromRef:

```yaml
spec:
  region: us-west-2
  foundationModel: amazon.nova-micro-v1:0
  agentResourceRoleArn:
    valueFrom:
      kind: AwsIamRole
      name: bedrock-agent-role
      fieldPath: status.outputs.role_arn
  knowledgeBaseAssociations:
    - name: docs
      description: Product documentation for grounded answers about features and setup
      knowledgeBaseId:
        valueFrom:
          kind: AwsBedrockKnowledgeBase
          name: product-docs
          fieldPath: status.outputs.knowledge_base_id
  aliases:
    - name: live
```

The InfraPipeline resolves the dependency graph, deploys the role and knowledge base first, then assembles the agent on top of them.

## Key Configuration

These are the most important decisions when configuring an agent. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Serve through aliases, never the draft** — Every spec edit changes the DRAFT; an alias pins the numbered snapshot taken at its creation, so production behavior only changes when you create or re-point an alias. Treat `aliases` entries like release channels — `live` for production, a second entry for canary — and re-point consumers between aliases as the rollback path.

**The model choice gates account setup** — `foundationModel` accepts a foundation-model ID or ARN, or an inference-profile ID or ARN (for per-application cost attribution or cross-region routing). Auto-enabled AWS models need nothing; marketplace models need an access agreement in place before the agent can prepare — compose the agreement first and let the chart order the dependency.

**The role is load-bearing at prepare time** — AWS validates assumability and model permissions when the agent prepares: a role missing `bedrock:InvokeModel` fails the deploy itself. Changing `agentResourceRoleArn` later replaces the agent.

**Collaborators need supervisor mode and alias ARNs** — Set `agentCollaboration` to SUPERVISOR (plan-and-delegate) or SUPERVISOR_ROUTER (route each request to exactly one collaborator) before adding `collaborators`, and reference each collaborating agent's ALIAS (its `alias_arns` output), never the agent itself. Expect deletes of collaborator-carrying supervisors to take extra prepare cycles — AWS refuses to prepare a supervisor losing its last collaborator, and the provider works around it service-side.

**Action groups take one of two shapes** — A custom group needs `executor` plus exactly one of `apiSchema`/`functionSchema`; a reserved group names only `parentActionGroupSignature` (no executor, schema, or description). Within `executor`, `returnControl: true` hands the tool call back to your application to execute; a `lambda` reference makes AWS invoke the function directly.

**Pin the guardrail version** — `guardrail.version` accepts DRAFT or a published number. Production agents should pin a published version so guardrail draft edits never change live behavior mid-conversation.

**Entry names are identity** — Each satellite entry's `name` is the for_each key on both engines and the key in the output maps (`alias_arns`, `action_group_ids`, ...). Renaming an entry destroys and recreates that satellite; renaming an alias in particular drops its snapshot and takes a fresh one.

**Memory is session summaries only** — AWS defines one memory type; the presence of `memory` enables it, and `storageDays` bounds how long summaries persist (1–365 days). It carries context ACROSS sessions with the same memory ID — within-session context is governed by `idleSessionTtlSeconds` instead.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `agentResourceRoleArn` | `status.outputs.role_arn` |
| **AwsKmsKey** | `customerEncryptionKeyArn` | `status.outputs.key_arn` |
| **AwsBedrockGuardrail** | `guardrail.guardrailId` | `status.outputs.guardrail_id` |
| **AwsBedrockKnowledgeBase** | `knowledgeBaseAssociations[].knowledgeBaseId` | `status.outputs.knowledge_base_id` |
| **AwsLambda** | `actionGroups[].executor.lambda`, `promptOverride.overrideLambda` | `status.outputs.function_arn` |
| **AwsS3Bucket** | `actionGroups[].apiSchema.s3.bucketName` | `status.outputs.bucket_id` |
| **AwsBedrockAgent** | `collaborators[].collaboratorAliasArn` | `status.outputs.alias_arns.<alias-name>` |
| **AwsBedrockProvisionedThroughput** | `aliases[].routing.provisionedThroughput` | `status.outputs.provisioned_model_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `agent_id` | The unique agent identifier | `InvokeAgent` calls from applications, paired with an alias ID |
| `agent_arn` | The agent's ARN — the canonical key for IAM policies | Policies scoping who can invoke or manage the agent |
| `alias_ids` | Alias IDs keyed by each `aliases` entry's name | Application configuration for `InvokeAgent` against a specific release channel |
| `alias_arns` | Alias ARNs keyed by each `aliases` entry's name | A supervisor agent's `collaboratorAliasArn`; a flow's agent node |

`draft_version` is the constant "DRAFT", and the per-satellite ID maps (`action_group_ids`, `collaborator_ids`, `associated_knowledge_base_ids`) echo the identifiers of what this component itself created — they are records, not composition inputs.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Tool-using assistant** — One agent, one or two action groups, one `live` alias. Choose `returnControl` while your tool execution lives inside your application (you keep full control and audit of side effects); switch to a Lambda executor when the tool should run server-side without your application in the loop. Start from the **Support Agent with Tools** preset.

**Supervisor with specialist collaborators** — A SUPERVISOR agent that plans and delegates to purpose-built agents, each referenced through its alias ARN with a `collaborationInstruction` written like a dispatcher's briefing. The trade: multi-agent quality on complex requests against more model invocations (and cost) per user turn. Start from the **Supervisor with Collaborators** preset.

**Retrieval-grounded agent** — Associate a knowledge base and spend effort on the association's `description` — the model reads exactly that text to decide when to retrieve, so "Product documentation for setup and feature questions" outperforms a generic label. Pair with a pinned guardrail version when answers reach end users.

## Works With

- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the service role the agent assumes, wired via `agentResourceRoleArn`
- [**AWS Bedrock Knowledge Base**](/infra-catalog/aws-bedrock-knowledge-base) — retrieval sources associated through `knowledgeBaseAssociations`
- [**AWS Bedrock Guardrail**](/infra-catalog/aws-bedrock-guardrail) — content-safety policies attached via `guardrail`, pinned to a published version
- [**AWS Lambda**](/infra-catalog/aws-lambda) — action-group executors and the optional prompt-override parser
- [**AWS Bedrock Model Access**](/infra-catalog/aws-bedrock-model-access) — the marketplace-model agreement that must exist before such a model can power the agent
- [**AWS Bedrock Inference Profile**](/infra-catalog/aws-bedrock-inference-profile) — an alternative `foundationModel` value for cost attribution or cross-region routing
- [**AWS Bedrock Provisioned Throughput**](/infra-catalog/aws-bedrock-provisioned-throughput) — reserved capacity an alias serves through via `routing.provisionedThroughput`
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — customer-managed encryption via `customerEncryptionKeyArn`
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) — hosts OpenAPI schema documents for action groups

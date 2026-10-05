# AWS Bedrock AgentCore Evaluation

Deploys a bundle of Amazon Bedrock AgentCore Evaluations resources — evaluators that score agent behavior, harnesses that run repeatable agent test benches, and online evaluation configs that continuously score sampled production sessions from CloudWatch logs. Each arm is optional (the bundle needs at least one), entries are name-keyed collections, and none of the three requires an agent runtime to exist — the Evaluations capability deploys standalone. Charges follow evaluation runs: model tokens for LLM judges, Lambda invocations for code evaluators, and sampled-session scoring for online configs.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions, per named entry:

- **Evaluator** — one scoring definition per `evaluators` entry: an LLM-as-a-judge with a categorical or numerical rating scale, or a code-based evaluator backed by your Lambda function
- **Harness** — one agent test bench per `harnesses` entry: the model under test (Bedrock, Gemini, or OpenAI), system prompts, tools, and an optional memory and runtime environment. When `memory` or `runtimeEnvironment` is omitted, AWS auto-provisions managed ones — the harness still has both, they are just AWS-owned
- **Online Evaluation Config** — one continuous-evaluation rule per `onlineEvaluationConfigs` entry, sampling production agent sessions from CloudWatch log groups and scoring them with the listed evaluators

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with Bedrock AgentCore control-plane permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- An IAM role trusting `bedrock-agentcore.amazonaws.com`, wired as `executionRoleArn` (only for harnesses and online evaluation configs — evaluators take no role)
- Bedrock model access in the region (only for LLM-judge evaluators and Bedrock-model harnesses)
- A Lambda function whose resource policy allows `bedrock-agentcore` to invoke it (only for code-based evaluators)
- CloudWatch log groups that AgentCore observability actually writes to (only for online evaluation configs — sampling an empty group scores nothing)
- A Secrets Manager secret holding the vendor API key (only for Gemini or OpenAI harness models)

## Deploy

### Console

Open the deployment store, find **AWS Bedrock AgentCore Evaluation**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the three bundle arms — evaluators, harnesses, and online evaluation configs. Start from the **Code Evaluator** preset in the [Presets](#presets) tab for the cheapest first deploy, or the **LLM Judge and Harness** preset for a judge plus a test bench.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockAgentCoreEvaluation
metadata:
  name: agent-scoring
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  evaluators:
    - name: order_accuracy
      description: Scores order-handling traces with a Lambda scorer
      level: TRACE
      codeBased:
        lambdaArn:
          valueFrom:
            kind: AwsLambda
            name: order-accuracy-scorer
            fieldPath: status.outputs.function_arn
```

```shell
planton apply -f agentcore-evaluation.yaml
```

This creates one TRACE-level code-based evaluator that scores runs with the referenced Lambda function — create does not invoke the function, so the first deploy is independent of Bedrock model access. An Infra Job tracks the provisioning in real time.

### InfraChart

When the evaluation bundle deploys alongside its scorer function in one chart, wire the Lambda reference via ValueFromRef:

```yaml
spec:
  region: us-west-2
  evaluators:
    - name: order_accuracy
      level: TRACE
      codeBased:
        lambdaArn:
          valueFrom:
            kind: AwsLambda
            name: order-accuracy-scorer
            fieldPath: status.outputs.function_arn
```

The InfraPipeline resolves the dependency graph, deploys the Lambda first, then creates the evaluator against it.

## Key Configuration

These are the most important decisions when configuring an evaluation bundle. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Entry names are permanent identity** — Each evaluator, harness, and online config name is the for_each key on both IaC engines and the key in the corresponding output map, and AWS exposes no rename: changing a name replaces the resource. Harness names additionally reject hyphens (a letter, then letters/digits/underscores) — `support_bench`, never `support-bench`.

**Exactly one scoring arm per evaluator** — `llmAsAJudge` and `codeBased` are mutually exclusive. Start with a code evaluator: create does not invoke the Lambda, so the first evaluator is cheap to stand up and independent of Bedrock model access. Judges cost model tokens on every run.

**Judge models must resolve through the region's inference set** — CreateEvaluator validates the judge's `modelId` against the region's INFERENCE set and rejects models it cannot invoke there with "not available in region", even when the bare foundation-model ID is regionally listed. Use the cross-region inference-profile form (`us.amazon.nova-2-lite-v1:0`). The harness's model field has no such create-time gate — a bad harness model surfaces only when a run executes.

**SESSION judges need placeholders in their instructions** — a SESSION-level judge whose instructions embed none of `{available_tools}`, `{context}`, `{actual_tool_trajectory}`, `{expected_tool_trajectory}`, or `{assertions}` is rejected at create. The manifest validation front-loads this so the failure never reaches AWS.

**samplingPercentage is the cost lever** — every sampled session runs every evaluator in the online config's list, so the sampling rate multiplies directly into judge-token or Lambda-invocation spend. Production configs usually sample low single digits; 100 is for short calibration windows, not steady state.

**An evaluator in use by an active online config is locked** — editing or destroying it waits on AWS's conflict retries and then fails. Disable or delete the referencing online config first; the day-two symptom is an apply that hangs on the evaluator and errors with a conflict.

**Applies can legitimately take up to 30 minutes per resource** — all three resource types carry 30-minute create/update/delete waiters. Harnesses finish at READY; evaluators and online configs at ACTIVE — a stuck deploy is diagnosed against the right terminal word per arm.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsLambda** | `evaluators[].codeBased.lambdaArn` | `status.outputs.function_arn` |
| **AwsKmsKey** | `evaluators[].kmsKeyArn` | `status.outputs.key_arn` |
| **AwsIamRole** | `harnesses[].executionRoleArn`, `onlineEvaluationConfigs[].executionRoleArn` | `status.outputs.role_arn` |
| **AwsSecretsManagerSecret** | `harnesses[].model.gemini.apiKeyArn`, `harnesses[].model.openai.apiKeyArn` | `status.outputs.secret_arn` |
| **AwsBedrockAgentCoreGateway** | `harnesses[].tools[].agentcoreGateway.gatewayArn` | `status.outputs.gateway_arn` |
| **AwsBedrockAgentCoreMemory** | `harnesses[].memory.memoryArn` | `status.outputs.memory_arn` |
| **AwsBedrockAgentCoreRuntime** | `harnesses[].runtimeEnvironment.agentRuntimeArn` | `status.outputs.agent_runtime_arn` |
| **AwsCloudwatchLogGroup** | `onlineEvaluationConfigs[].dataSource.logGroupNames` | `status.outputs.log_group_name` |

Harness runtime-environment networking and private endpoints also take references — **AwsVpc**, **AwsSubnet**, **AwsSecurityGroup**, and **AwsEfsAccessPoint** — through the VPC config, managed-endpoint, and filesystem fields.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `evaluator_ids` / `evaluator_arns` | Evaluator identifiers keyed by each `evaluators` entry's name | Online evaluation configs in other bundles referencing custom evaluators by ID |
| `harness_ids` / `harness_arns` | Harness identifiers keyed by each `harnesses` entry's name | Starting evaluation runs against a named test bench |
| `online_evaluation_output_log_groups` | The CloudWatch log group each online config writes its results to (server-assigned), keyed by the entry's name | Dashboards and metric alarms over evaluation scores |

`online_evaluation_config_ids` and `online_evaluation_config_arns` are also exported, keyed the same way; they identify the configs themselves rather than feeding a composition, so they mostly serve operational tooling.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Code evaluator first** — one TRACE-level evaluator backed by your Lambda. Create does not invoke the function, so this shape proves the pipeline before any Bedrock model access exists, and the scoring logic stays fully yours. Start from the **Code Evaluator** preset.

**Judge plus bench** — an LLM judge with a categorical scale alongside a harness running the agent's model and system prompts. Runs execute the harness and the judge scores the results — the repeatable pre-release regression bench. Start from the **LLM Judge and Harness** preset.

**Continuous production scoring** — an online evaluation config pointed at the log groups AgentCore observability writes, sampling a low single-digit percentage and scoring with AWS builtins (`Builtin.Helpfulness` and siblings) or this bundle's own evaluators referenced by name — the modules resolve the name to the created ID and order the create behind it. Filters narrow scoring to the sessions that matter.

## Works With

- [**AWS Lambda**](/infra-catalog/aws-lambda) — the scoring function behind code-based evaluators, wired via `lambdaArn`
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the execution role harnesses and online configs assume; must trust `bedrock-agentcore.amazonaws.com`
- [**AWS CloudWatch Log Group**](/infra-catalog/aws-cloudwatch-log-group) — the session-trace source online configs sample from
- [**AWS Secrets Manager Secret**](/infra-catalog/aws-secrets-manager-secret) — holds the API key for Gemini and OpenAI harness models
- [**AWS Bedrock AgentCore Gateway**](/infra-catalog/aws-bedrock-agent-core-gateway) — a tool front door harness agents call through `agentcoreGateway` tools
- [**AWS Bedrock AgentCore Memory**](/infra-catalog/aws-bedrock-agent-core-memory) — explicit memory a harness reads and writes during runs
- [**AWS Bedrock AgentCore Runtime**](/infra-catalog/aws-bedrock-agent-core-runtime) — pins a harness to an explicit runtime environment instead of the AWS-managed default
- [**AWS Bedrock AgentCore Tools**](/infra-catalog/aws-bedrock-agent-core-tools) — browsers and code interpreters harness tools reference by ARN
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — customer-managed encryption for evaluator data at rest

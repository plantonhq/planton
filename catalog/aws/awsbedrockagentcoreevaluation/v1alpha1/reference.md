# AwsBedrockAgentCoreEvaluation

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

AwsBedrockAgentCoreEvaluationSpec defines the desired configuration for
a bundle of Amazon Bedrock AgentCore Evaluations resources - the
capability that scores agent behavior:

  - `evaluators`: scoring definitions - an LLM judge with a rating
    scale, or your own Lambda function;
  - `harnesses`: repeatable agent test benches (model, tools, prompts)
    that evaluation runs execute against;
  - `online_evaluation_configs`: continuous evaluation of PRODUCTION
    traffic - sampling agent sessions from CloudWatch logs and scoring
    them with evaluators.

Every arm is optional; author the ones this bundle owns (at least
one). None of the three requires an agent runtime to exist - the
Evaluations capability deploys standalone, and a harness references a
runtime only through its optional runtime_environment arm.

Creating evaluation objects is free; AWS bills per evaluation run
(model tokens for LLM judges, Lambda invocations for code evaluators,
sampled-session scoring for online configs).

## Example

```yaml
# Canonical AwsBedrockAgentCoreEvaluation example (hack/dev manifest and
# refgen Example source): a bundle exercising every arm -- an LLM-judge
# evaluator with a numerical scale, a code-based evaluator, a Bedrock
# harness, and an online evaluation config sampling CloudWatch logs
# with a builtin evaluator. Literal ARNs stand in for composed
# references so the offline `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockAgentCoreEvaluation
metadata:
  name: support-eval
  id: support-eval
  org: test-org
  env: dev
spec:
  region: us-west-2
  evaluators:
    - name: helpfulness
      description: Scores whether the agent was helpful
      level: SESSION
      llmAsAJudge:
        # SESSION-level instructions must embed one of the level's
        # single-brace placeholders ({context} here).
        instructions: Score how helpful the agent's response was given the {context}.
        # Judge models resolve through the region's inference set; the
        # inference-profile id form is what CreateEvaluator accepts.
        model:
          modelId: us.amazon.nova-2-lite-v1:0
          inference:
            temperature: 0
            maxTokens: 256
        ratingScale:
          numerical:
            - label: poor
              definition: Unhelpful or incorrect
              value: 1
            - label: excellent
              definition: Fully helpful and correct
              value: 5
    - name: custom_score
      description: Lambda scorer
      level: TRACE
      codeBased:
        lambdaArn:
          value: arn:aws:lambda:us-west-2:123456789012:function:score-trace
        timeoutSeconds: 60
  harnesses:
    # Harness names take letters/digits/underscores only (no hyphens).
    - name: support_bench
      executionRoleArn:
        value: arn:aws:iam::123456789012:role/agentcore-eval-role
      model:
        bedrock:
          modelId: anthropic.claude-3-haiku-20240307-v1:0
          temperature: 0.2
          maxTokens: 512
      systemPrompts:
        - text: You are a support agent under evaluation.
      tools:
        - name: search
          type: inline_function
          inlineFunction:
            description: Search the knowledge base
            inputSchema: '{"type":"object","properties":{"query":{"type":"string"}}}'
  onlineEvaluationConfigs:
    - name: prod_sampling
      description: Sample 5 percent of production sessions
      executionRoleArn:
        value: arn:aws:iam::123456789012:role/agentcore-eval-role
      enabled: true
      dataSource:
        logGroupNames:
          - value: /aws/bedrock/agentcore/support
        serviceNames:
          - support-agent
      evaluatorIds:
        - Builtin.Helpfulness
        - helpfulness
      rule:
        samplingPercentage: 5
        sessionTimeoutMinutes: 15
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.evaluators` | `[]AwsBedrockAgentCoreEvaluator` |  |  |  |
| `spec.evaluators[].name` | `string` | yes |  |  |
| `spec.evaluators[].description` | `string` |  |  |  |
| `spec.evaluators[].level` | `string` |  |  |  |
| `spec.evaluators[].kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.evaluators[].llmAsAJudge` | `AwsBedrockAgentCoreLlmJudge` |  |  |  |
| `spec.evaluators[].llmAsAJudge.instructions` | `string` (sensitive) | yes |  |  |
| `spec.evaluators[].llmAsAJudge.model` | `AwsBedrockAgentCoreJudgeModel` | yes |  |  |
| `spec.evaluators[].llmAsAJudge.model.modelId` | `string` | yes |  |  |
| `spec.evaluators[].llmAsAJudge.model.additionalModelRequestFields` | `object` |  |  |  |
| `spec.evaluators[].llmAsAJudge.model.inference` | `AwsBedrockAgentCoreJudgeInference` |  |  |  |
| `spec.evaluators[].llmAsAJudge.model.inference.maxTokens` | `int32` |  |  |  |
| `spec.evaluators[].llmAsAJudge.model.inference.stopSequences` | `[]string` |  |  |  |
| `spec.evaluators[].llmAsAJudge.model.inference.temperature` | `double` |  |  |  |
| `spec.evaluators[].llmAsAJudge.model.inference.topP` | `double` |  |  |  |
| `spec.evaluators[].llmAsAJudge.ratingScale` | `AwsBedrockAgentCoreRatingScale` | yes |  |  |
| `spec.evaluators[].llmAsAJudge.ratingScale.categorical` | `[]AwsBedrockAgentCoreCategoricalRating` |  |  |  |
| `spec.evaluators[].llmAsAJudge.ratingScale.categorical[].label` | `string` | yes |  |  |
| `spec.evaluators[].llmAsAJudge.ratingScale.categorical[].definition` | `string` | yes |  |  |
| `spec.evaluators[].llmAsAJudge.ratingScale.numerical` | `[]AwsBedrockAgentCoreNumericalRating` |  |  |  |
| `spec.evaluators[].llmAsAJudge.ratingScale.numerical[].label` | `string` | yes |  |  |
| `spec.evaluators[].llmAsAJudge.ratingScale.numerical[].definition` | `string` | yes |  |  |
| `spec.evaluators[].llmAsAJudge.ratingScale.numerical[].value` | `double` |  |  |  |
| `spec.evaluators[].codeBased` | `AwsBedrockAgentCoreCodeEvaluator` |  |  |  |
| `spec.evaluators[].codeBased.lambdaArn` | `string \| valueFrom` | yes |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.evaluators[].codeBased.timeoutSeconds` | `int32` |  |  |  |
| `spec.harnesses` | `[]AwsBedrockAgentCoreHarness` |  |  |  |
| `spec.harnesses[].name` | `string` | yes |  |  |
| `spec.harnesses[].executionRoleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.harnesses[].model` | `AwsBedrockAgentCoreHarnessModel` | yes |  |  |
| `spec.harnesses[].model.bedrock` | `AwsBedrockAgentCoreHarnessBedrockModel` |  |  |  |
| `spec.harnesses[].model.bedrock.modelId` | `string` | yes |  |  |
| `spec.harnesses[].model.bedrock.maxTokens` | `int32` |  |  |  |
| `spec.harnesses[].model.bedrock.temperature` | `double` |  |  |  |
| `spec.harnesses[].model.bedrock.topP` | `double` |  |  |  |
| `spec.harnesses[].model.gemini` | `AwsBedrockAgentCoreHarnessGeminiModel` |  |  |  |
| `spec.harnesses[].model.gemini.apiKeyArn` | `string \| valueFrom` | yes |  | AwsSecretsManagerSecret (`status.outputs.secret_arn`) |
| `spec.harnesses[].model.gemini.modelId` | `string` | yes |  |  |
| `spec.harnesses[].model.gemini.maxTokens` | `int32` |  |  |  |
| `spec.harnesses[].model.gemini.temperature` | `double` |  |  |  |
| `spec.harnesses[].model.gemini.topP` | `double` |  |  |  |
| `spec.harnesses[].model.gemini.topK` | `int32` |  |  |  |
| `spec.harnesses[].model.openai` | `AwsBedrockAgentCoreHarnessOpenAiModel` |  |  |  |
| `spec.harnesses[].model.openai.apiKeyArn` | `string \| valueFrom` | yes |  | AwsSecretsManagerSecret (`status.outputs.secret_arn`) |
| `spec.harnesses[].model.openai.modelId` | `string` | yes |  |  |
| `spec.harnesses[].model.openai.maxTokens` | `int32` |  |  |  |
| `spec.harnesses[].model.openai.temperature` | `double` |  |  |  |
| `spec.harnesses[].model.openai.topP` | `double` |  |  |  |
| `spec.harnesses[].systemPrompts` | `[]AwsBedrockAgentCoreHarnessSystemPrompt` |  |  |  |
| `spec.harnesses[].systemPrompts[].text` | `string` (sensitive) | yes |  |  |
| `spec.harnesses[].tools` | `[]AwsBedrockAgentCoreHarnessTool` |  |  |  |
| `spec.harnesses[].tools[].name` | `string` |  |  |  |
| `spec.harnesses[].tools[].type` | `string` |  |  |  |
| `spec.harnesses[].tools[].remoteMcp` | `AwsBedrockAgentCoreHarnessRemoteMcpTool` |  |  |  |
| `spec.harnesses[].tools[].remoteMcp.url` | `string` (sensitive) | yes |  |  |
| `spec.harnesses[].tools[].remoteMcp.headers` | `map<string, string>` (sensitive) |  |  |  |
| `spec.harnesses[].tools[].agentcoreBrowser` | `AwsBedrockAgentCoreHarnessBrowserTool` |  |  |  |
| `spec.harnesses[].tools[].agentcoreBrowser.browserArn` | `string \| valueFrom` |  |  |  |
| `spec.harnesses[].tools[].agentcoreGateway` | `AwsBedrockAgentCoreHarnessGatewayTool` |  |  |  |
| `spec.harnesses[].tools[].agentcoreGateway.gatewayArn` | `string \| valueFrom` | yes |  | AwsBedrockAgentCoreGateway (`status.outputs.gateway_arn`) |
| `spec.harnesses[].tools[].agentcoreGateway.outboundAuth` | `AwsBedrockAgentCoreHarnessGatewayOutboundAuth` |  |  |  |
| `spec.harnesses[].tools[].agentcoreGateway.outboundAuth.awsIam` | `bool` |  |  |  |
| `spec.harnesses[].tools[].agentcoreGateway.outboundAuth.noAuth` | `bool` |  |  |  |
| `spec.harnesses[].tools[].agentcoreGateway.outboundAuth.oauth` | `AwsBedrockAgentCoreHarnessGatewayOAuth` |  |  |  |
| `spec.harnesses[].tools[].agentcoreGateway.outboundAuth.oauth.providerArn` | `string \| valueFrom` | yes |  |  |
| `spec.harnesses[].tools[].agentcoreGateway.outboundAuth.oauth.scopes` | `[]string` | yes |  |  |
| `spec.harnesses[].tools[].agentcoreGateway.outboundAuth.oauth.customParameters` | `map<string, string>` |  |  |  |
| `spec.harnesses[].tools[].agentcoreGateway.outboundAuth.oauth.defaultReturnUrl` | `string` |  |  |  |
| `spec.harnesses[].tools[].agentcoreGateway.outboundAuth.oauth.grantType` | `string` |  |  |  |
| `spec.harnesses[].tools[].inlineFunction` | `AwsBedrockAgentCoreHarnessInlineFunctionTool` |  |  |  |
| `spec.harnesses[].tools[].inlineFunction.description` | `string` | yes |  |  |
| `spec.harnesses[].tools[].inlineFunction.inputSchema` | `string` (sensitive) | yes |  |  |
| `spec.harnesses[].tools[].agentcoreCodeInterpreter` | `AwsBedrockAgentCoreHarnessCodeInterpreterTool` |  |  |  |
| `spec.harnesses[].tools[].agentcoreCodeInterpreter.codeInterpreterArn` | `string \| valueFrom` |  |  |  |
| `spec.harnesses[].skillPaths` | `[]string` |  |  |  |
| `spec.harnesses[].memory` | `AwsBedrockAgentCoreHarnessMemory` |  |  |  |
| `spec.harnesses[].memory.memoryArn` | `string \| valueFrom` | yes |  | AwsBedrockAgentCoreMemory (`status.outputs.memory_arn`) |
| `spec.harnesses[].memory.actorId` | `string` |  |  |  |
| `spec.harnesses[].memory.messagesCount` | `int32` |  |  |  |
| `spec.harnesses[].memory.retrieval` | `AwsBedrockAgentCoreHarnessMemoryRetrieval` |  |  |  |
| `spec.harnesses[].memory.retrieval.namespace` | `string` | yes |  |  |
| `spec.harnesses[].memory.retrieval.relevanceScore` | `double` |  |  |  |
| `spec.harnesses[].memory.retrieval.strategyId` | `string` |  |  |  |
| `spec.harnesses[].memory.retrieval.topK` | `int32` |  |  |  |
| `spec.harnesses[].environmentVariables` | `map<string, string>` (sensitive) |  |  |  |
| `spec.harnesses[].runtimeEnvironment` | `AwsBedrockAgentCoreHarnessRuntimeEnvironment` |  |  |  |
| `spec.harnesses[].runtimeEnvironment.agentRuntimeArn` | `string \| valueFrom` |  |  | AwsBedrockAgentCoreRuntime (`status.outputs.agent_runtime_arn`) |
| `spec.harnesses[].runtimeEnvironment.filesystems` | `[]AwsBedrockAgentCoreHarnessFilesystem` |  |  |  |
| `spec.harnesses[].runtimeEnvironment.filesystems[].mountPath` | `string` | yes |  |  |
| `spec.harnesses[].runtimeEnvironment.filesystems[].efsAccessPointArn` | `string \| valueFrom` |  |  | AwsEfsAccessPoint (`status.outputs.access_point_arn`) |
| `spec.harnesses[].runtimeEnvironment.filesystems[].s3AccessPointArn` | `string \| valueFrom` |  |  |  |
| `spec.harnesses[].runtimeEnvironment.filesystems[].sessionStorage` | `bool` |  |  |  |
| `spec.harnesses[].runtimeEnvironment.lifecycle` | `AwsBedrockAgentCoreHarnessLifecycle` |  |  |  |
| `spec.harnesses[].runtimeEnvironment.lifecycle.idleRuntimeSessionTimeoutSeconds` | `int32` |  |  |  |
| `spec.harnesses[].runtimeEnvironment.lifecycle.maxLifetimeSeconds` | `int32` |  |  |  |
| `spec.harnesses[].runtimeEnvironment.network` | `AwsBedrockAgentCoreHarnessNetwork` |  |  |  |
| `spec.harnesses[].runtimeEnvironment.network.mode` | `string` |  |  |  |
| `spec.harnesses[].runtimeEnvironment.network.vpcConfig` | `AwsBedrockAgentCoreHarnessVpcConfig` |  |  |  |
| `spec.harnesses[].runtimeEnvironment.network.vpcConfig.subnets` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.harnesses[].runtimeEnvironment.network.vpcConfig.securityGroups` | `[]string \| valueFrom` | yes |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.harnesses[].runtimeEnvironment.network.vpcConfig.requireServiceS3Endpoint` | `bool` |  |  |  |
| `spec.harnesses[].containerImageUri` | `string` |  |  |  |
| `spec.harnesses[].customJwtAuthorizer` | `AwsBedrockAgentCoreJwtAuthorizer` |  |  |  |
| `spec.harnesses[].customJwtAuthorizer.discoveryUrl` | `string` | yes |  |  |
| `spec.harnesses[].customJwtAuthorizer.allowedAudience` | `[]string` |  |  |  |
| `spec.harnesses[].customJwtAuthorizer.allowedClients` | `[]string` |  |  |  |
| `spec.harnesses[].customJwtAuthorizer.allowedScopes` | `[]string` |  |  |  |
| `spec.harnesses[].customJwtAuthorizer.allowedWorkloads` | `AwsBedrockAgentCoreAllowedWorkloads` |  |  |  |
| `spec.harnesses[].customJwtAuthorizer.allowedWorkloads.workloadIdentities` | `[]string` | yes |  |  |
| `spec.harnesses[].customJwtAuthorizer.allowedWorkloads.hostingEnvironmentArns` | `[]string` | yes |  |  |
| `spec.harnesses[].customJwtAuthorizer.customClaims` | `[]AwsBedrockAgentCoreCustomClaim` |  |  |  |
| `spec.harnesses[].customJwtAuthorizer.customClaims[].claimName` | `string` | yes |  |  |
| `spec.harnesses[].customJwtAuthorizer.customClaims[].valueType` | `string` |  |  |  |
| `spec.harnesses[].customJwtAuthorizer.customClaims[].matchOperator` | `string` |  |  |  |
| `spec.harnesses[].customJwtAuthorizer.customClaims[].matchValue` | `string` |  |  |  |
| `spec.harnesses[].customJwtAuthorizer.customClaims[].matchValues` | `[]string` |  |  |  |
| `spec.harnesses[].customJwtAuthorizer.privateEndpoint` | `AwsBedrockAgentCorePrivateEndpoint` |  |  |  |
| `spec.harnesses[].customJwtAuthorizer.privateEndpoint.managedVpc` | `AwsBedrockAgentCoreManagedVpcEndpoint` |  |  |  |
| `spec.harnesses[].customJwtAuthorizer.privateEndpoint.managedVpc.vpcId` | `string \| valueFrom` | yes |  | AwsVpc (`status.outputs.vpc_id`) |
| `spec.harnesses[].customJwtAuthorizer.privateEndpoint.managedVpc.subnetIds` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.harnesses[].customJwtAuthorizer.privateEndpoint.managedVpc.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.harnesses[].customJwtAuthorizer.privateEndpoint.managedVpc.endpointIpAddressType` | `string` |  |  |  |
| `spec.harnesses[].customJwtAuthorizer.privateEndpoint.managedVpc.routingDomain` | `string` | yes |  |  |
| `spec.harnesses[].customJwtAuthorizer.privateEndpoint.managedVpc.tags` | `map<string, string>` |  |  |  |
| `spec.harnesses[].customJwtAuthorizer.privateEndpoint.selfManagedLattice` | `AwsBedrockAgentCoreLatticeEndpoint` |  |  |  |
| `spec.harnesses[].customJwtAuthorizer.privateEndpoint.selfManagedLattice.resourceConfigurationId` | `string` | yes |  |  |
| `spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides` | `[]AwsBedrockAgentCorePrivateEndpointOverride` |  |  |  |
| `spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].domain` | `string` | yes |  |  |
| `spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint` | `AwsBedrockAgentCorePrivateEndpoint` | yes |  |  |
| `spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc` | `AwsBedrockAgentCoreManagedVpcEndpoint` |  |  |  |
| `spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.vpcId` | `string \| valueFrom` | yes |  | AwsVpc (`status.outputs.vpc_id`) |
| `spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.subnetIds` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.endpointIpAddressType` | `string` |  |  |  |
| `spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.routingDomain` | `string` | yes |  |  |
| `spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.tags` | `map<string, string>` |  |  |  |
| `spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.selfManagedLattice` | `AwsBedrockAgentCoreLatticeEndpoint` |  |  |  |
| `spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.selfManagedLattice.resourceConfigurationId` | `string` | yes |  |  |
| `spec.harnesses[].allowedTools` | `[]string` |  |  |  |
| `spec.harnesses[].maxIterations` | `int32` |  |  |  |
| `spec.harnesses[].maxTokens` | `int32` |  |  |  |
| `spec.harnesses[].timeoutSeconds` | `int32` |  |  |  |
| `spec.harnesses[].truncation` | `AwsBedrockAgentCoreHarnessTruncation` |  |  |  |
| `spec.harnesses[].truncation.strategy` | `string` |  |  |  |
| `spec.harnesses[].truncation.slidingWindow` | `AwsBedrockAgentCoreHarnessSlidingWindow` |  |  |  |
| `spec.harnesses[].truncation.slidingWindow.messagesCount` | `int32` |  |  |  |
| `spec.harnesses[].truncation.summarization` | `AwsBedrockAgentCoreHarnessSummarization` |  |  |  |
| `spec.harnesses[].truncation.summarization.summaryRatio` | `double` |  |  |  |
| `spec.harnesses[].truncation.summarization.preserveRecentMessages` | `int32` |  |  |  |
| `spec.harnesses[].truncation.summarization.summarizationSystemPrompt` | `string` (sensitive) |  |  |  |
| `spec.onlineEvaluationConfigs` | `[]AwsBedrockAgentCoreOnlineEvaluationConfig` |  |  |  |
| `spec.onlineEvaluationConfigs[].name` | `string` | yes |  |  |
| `spec.onlineEvaluationConfigs[].description` | `string` |  |  |  |
| `spec.onlineEvaluationConfigs[].executionRoleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.onlineEvaluationConfigs[].enabled` | `bool` |  |  |  |
| `spec.onlineEvaluationConfigs[].dataSource` | `AwsBedrockAgentCoreOnlineEvaluationDataSource` | yes |  |  |
| `spec.onlineEvaluationConfigs[].dataSource.logGroupNames` | `[]string \| valueFrom` | yes |  | AwsCloudwatchLogGroup (`status.outputs.log_group_name`) |
| `spec.onlineEvaluationConfigs[].dataSource.serviceNames` | `[]string` | yes |  |  |
| `spec.onlineEvaluationConfigs[].evaluatorIds` | `[]string` | yes |  |  |
| `spec.onlineEvaluationConfigs[].rule` | `AwsBedrockAgentCoreOnlineEvaluationRule` | yes |  |  |
| `spec.onlineEvaluationConfigs[].rule.samplingPercentage` | `double` |  |  |  |
| `spec.onlineEvaluationConfigs[].rule.filters` | `[]AwsBedrockAgentCoreOnlineEvaluationFilter` |  |  |  |
| `spec.onlineEvaluationConfigs[].rule.filters[].key` | `string` | yes |  |  |
| `spec.onlineEvaluationConfigs[].rule.filters[].operator` | `string` |  |  |  |
| `spec.onlineEvaluationConfigs[].rule.filters[].stringValue` | `string` |  |  |  |
| `spec.onlineEvaluationConfigs[].rule.filters[].booleanValue` | `bool` |  |  |  |
| `spec.onlineEvaluationConfigs[].rule.filters[].doubleValue` | `double` |  |  |  |
| `spec.onlineEvaluationConfigs[].rule.sessionTimeoutMinutes` | `int32` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the evaluation resources will be created.
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.evaluators

`[]AwsBedrockAgentCoreEvaluator`

Scoring definitions: LLM-as-a-judge or Lambda-backed evaluators.

- rule: evaluator must set exactly one of llm_as_a_judge or code_based
- rule: SESSION-level llm_as_a_judge instructions must embed at least one placeholder: {available_tools}, {context}, {actual_tool_trajectory}, {expected_tool_trajectory}, {assertions}

### spec.evaluators[].name

`string` · required

Evaluator name in AWS (a letter, then up to 47 letters/digits/
underscores). The for_each key on both engines and the key in the
`evaluator_ids` output map. AWS derives the evaluator ID from it
("<name>-<10 chars>") and exposes NO rename - changing it replaces
the evaluator.

- rule: {"string":{"minLen":"1","pattern":"^[a-zA-Z][a-zA-Z0-9_]{0,47}$"}}

### spec.evaluators[].description

`string`

What this evaluator scores (1-200 characters).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"200"}}

### spec.evaluators[].level

`string`

The granularity the evaluator scores at: a single TOOL_CALL, one
TRACE (a full request/response cycle), or a whole SESSION.

- rule: {"string":{"in":["TOOL_CALL","TRACE","SESSION"]}}

### spec.evaluators[].kmsKeyArn

`string | valueFrom`

Customer-managed KMS key encrypting the evaluator's data at rest.
Omitted = the AWS-owned key.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.evaluators[].llmAsAJudge

`AwsBedrockAgentCoreLlmJudge`

A Bedrock model judges against instructions and a rating scale.

### spec.evaluators[].llmAsAJudge.instructions

`string` · required · sensitive

The judge prompt - what to assess and how to decide the score.
AWS substitutes single-brace placeholders (e.g. "{context}") with
run data, and each evaluator level REQUIRES at least one of its
allowed placeholders: for SESSION the set is {available_tools},
{context}, {actual_tool_trajectory}, {expected_tool_trajectory},
{assertions} (CreateEvaluator names the level's set when it
rejects). Treated as sensitive: prompts routinely embed
proprietary evaluation criteria.

- rule: {"required":true}

### spec.evaluators[].llmAsAJudge.model

`AwsBedrockAgentCoreJudgeModel` · required

The Bedrock model that judges.

- rule: {"required":true}

### spec.evaluators[].llmAsAJudge.model.modelId

`string` · required

Bedrock model identifier the judge runs on. Prefer a cross-region
inference profile ID ("us.amazon.nova-2-lite-v1:0") - CreateEvaluator
validates the judge model against the region's INFERENCE set and
rejects models it cannot invoke there with "not available in region",
even when the bare foundation-model ID is regionally listed (the
harness model field has no such create-time gate). A model ARN also
works; the account must have access to the model.

- rule: {"string":{"minLen":"1"}}

### spec.evaluators[].llmAsAJudge.model.additionalModelRequestFields

`object`

Model-specific request fields passed through verbatim (the shape the
model's own InvokeModel API documents), for parameters the typed
inference config does not cover.

### spec.evaluators[].llmAsAJudge.model.inference

`AwsBedrockAgentCoreJudgeInference`

Sampling controls for the judge.

### spec.evaluators[].llmAsAJudge.model.inference.maxTokens

`int32`

Maximum tokens the judge may generate. Omitted = the model default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.evaluators[].llmAsAJudge.model.inference.stopSequences

`[]string`

Sequences that stop generation (max 2500).

- rule: {"repeated":{"maxItems":"2500","items":{"string":{"minLen":"1"}}}}

### spec.evaluators[].llmAsAJudge.model.inference.temperature

`double` · optional (explicit presence)

Sampling temperature (0-1). Omitted = the model default. Judges
usually run at 0 for deterministic scoring.

- rule: {"double":{"lte":1,"gte":0}}

### spec.evaluators[].llmAsAJudge.model.inference.topP

`double` · optional (explicit presence)

Nucleus-sampling cap (0-1). Omitted = the model default.

- rule: {"double":{"lte":1,"gte":0}}

### spec.evaluators[].llmAsAJudge.ratingScale

`AwsBedrockAgentCoreRatingScale` · required

The allowed scores - exactly one scale shape.

- rule: {"required":true}
- rule: rating scale must define exactly one of categorical or numerical entries

### spec.evaluators[].llmAsAJudge.ratingScale.categorical

`[]AwsBedrockAgentCoreCategoricalRating`

Named categories ("helpful", "unhelpful") - the judge picks one.

### spec.evaluators[].llmAsAJudge.ratingScale.categorical[].label

`string` · required

The category label (1-100 characters).

- rule: {"string":{"minLen":"1","maxLen":"100"}}

### spec.evaluators[].llmAsAJudge.ratingScale.categorical[].definition

`string` · required

When the judge should assign this category.

- rule: {"string":{"minLen":"1"}}

### spec.evaluators[].llmAsAJudge.ratingScale.numerical

`[]AwsBedrockAgentCoreNumericalRating`

Numbered scores (1 = poor ... 5 = excellent) - the judge picks one
value.

### spec.evaluators[].llmAsAJudge.ratingScale.numerical[].label

`string` · required

The score label (1-100 characters), e.g. "excellent".

- rule: {"string":{"minLen":"1","maxLen":"100"}}

### spec.evaluators[].llmAsAJudge.ratingScale.numerical[].definition

`string` · required

When the judge should assign this score.

- rule: {"string":{"minLen":"1"}}

### spec.evaluators[].llmAsAJudge.ratingScale.numerical[].value

`double`

The numeric value (>= 0) reported for this score.

- rule: {"double":{"gte":0}}

### spec.evaluators[].codeBased

`AwsBedrockAgentCoreCodeEvaluator`

Your Lambda function computes the score.

### spec.evaluators[].codeBased.lambdaArn

`string | valueFrom` · required

The Lambda function that computes the score. The function receives
the evaluation payload and returns the result; the evaluator's IAM
integration invokes it, so the function's resource policy must allow
bedrock-agentcore.

- references: AwsLambda (`status.outputs.function_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.evaluators[].codeBased.timeoutSeconds

`int32`

Seconds the evaluator waits for the Lambda (1-300). Omitted = 60.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":300,"gte":1}}

### spec.harnesses

`[]AwsBedrockAgentCoreHarness`

Agent test benches evaluation runs execute against.

- rule: each allowed_tools entry must match the name of a tool configured in tools

### spec.harnesses[].name

`string` · required

Harness name in AWS (a letter, then up to 39 letters/digits/
underscores - hyphens are rejected; CreateHarness names this exact
regex server-side). The for_each key on both engines and the key in
the `harness_ids` output map. AWS exposes no rename - changing it
replaces the harness.

- rule: {"string":{"minLen":"1","pattern":"^[a-zA-Z][a-zA-Z0-9_]{0,39}$"}}

### spec.harnesses[].executionRoleArn

`string | valueFrom` · required

IAM role the harness assumes to run the agent under test (invoking
models, calling tools, reading memory). Must trust
bedrock-agentcore.amazonaws.com. Updatable in place.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.harnesses[].model

`AwsBedrockAgentCoreHarnessModel` · required

The model the harness drives - exactly one vendor arm.

- rule: {"required":true}
- rule: harness model must set exactly one of bedrock, gemini, or openai

### spec.harnesses[].model.bedrock

`AwsBedrockAgentCoreHarnessBedrockModel`

A Bedrock-hosted model.

### spec.harnesses[].model.bedrock.modelId

`string` · required

Bedrock model identifier (foundation model ID, inference profile ID,
or model ARN). The harness's execution role must be able to invoke
it.

- rule: {"string":{"minLen":"1"}}

### spec.harnesses[].model.bedrock.maxTokens

`int32`

Maximum tokens per response. Omitted = the model default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.harnesses[].model.bedrock.temperature

`double` · optional (explicit presence)

Sampling temperature (0-2). Omitted = the model default.

- rule: {"double":{"lte":2,"gte":0}}

### spec.harnesses[].model.bedrock.topP

`double` · optional (explicit presence)

Nucleus-sampling cap (0-1). Omitted = the model default.

- rule: {"double":{"lte":1,"gte":0}}

### spec.harnesses[].model.gemini

`AwsBedrockAgentCoreHarnessGeminiModel`

A Google Gemini model (API key from Secrets Manager).

### spec.harnesses[].model.gemini.apiKeyArn

`string | valueFrom` · required

Secrets Manager secret holding the Gemini API key. The ARN is a
reference, not the key itself - the harness reads the secret at run
time with its execution role.

- references: AwsSecretsManagerSecret (`status.outputs.secret_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecretsManagerSecret, name: <that resource's name>, fieldPath: status.outputs.secret_arn}} -- a bare string does not parse

### spec.harnesses[].model.gemini.modelId

`string` · required

Gemini model identifier (e.g. "gemini-2.0-flash").

- rule: {"string":{"minLen":"1"}}

### spec.harnesses[].model.gemini.maxTokens

`int32`

Maximum tokens per response. Omitted = the model default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.harnesses[].model.gemini.temperature

`double` · optional (explicit presence)

Sampling temperature (0-2). Omitted = the model default.

- rule: {"double":{"lte":2,"gte":0}}

### spec.harnesses[].model.gemini.topP

`double` · optional (explicit presence)

Nucleus-sampling cap (0-1). Omitted = the model default.

- rule: {"double":{"lte":1,"gte":0}}

### spec.harnesses[].model.gemini.topK

`int32` · optional (explicit presence)

Top-K sampling cutoff (0-500). Omitted = the model default.

- rule: {"int32":{"lte":500,"gte":0}}

### spec.harnesses[].model.openai

`AwsBedrockAgentCoreHarnessOpenAiModel`

An OpenAI model (API key from Secrets Manager).

### spec.harnesses[].model.openai.apiKeyArn

`string | valueFrom` · required

Secrets Manager secret holding the OpenAI API key. The ARN is a
reference, not the key itself - the harness reads the secret at run
time with its execution role.

- references: AwsSecretsManagerSecret (`status.outputs.secret_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecretsManagerSecret, name: <that resource's name>, fieldPath: status.outputs.secret_arn}} -- a bare string does not parse

### spec.harnesses[].model.openai.modelId

`string` · required

OpenAI model identifier (e.g. "gpt-4o").

- rule: {"string":{"minLen":"1"}}

### spec.harnesses[].model.openai.maxTokens

`int32`

Maximum tokens per response. Omitted = the model default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.harnesses[].model.openai.temperature

`double` · optional (explicit presence)

Sampling temperature (0-2). Omitted = the model default.

- rule: {"double":{"lte":2,"gte":0}}

### spec.harnesses[].model.openai.topP

`double` · optional (explicit presence)

Nucleus-sampling cap (0-1). Omitted = the model default.

- rule: {"double":{"lte":1,"gte":0}}

### spec.harnesses[].systemPrompts

`[]AwsBedrockAgentCoreHarnessSystemPrompt`

System prompts prepended to every harness run, in order.

### spec.harnesses[].systemPrompts[].text

`string` · required · sensitive

The prompt text. Treated as sensitive: system prompts routinely
embed proprietary agent behavior.

- rule: {"required":true}

### spec.harnesses[].tools

`[]AwsBedrockAgentCoreHarnessTool`

Tools the agent under test may call.

- rule: set exactly the config arm matching type (remote_mcp, agentcore_gateway, and inline_function require theirs; agentcore_browser and agentcore_code_interpreter may omit theirs to use the AWS default tool)

### spec.harnesses[].tools[].name

`string`

Tool name (1-64 characters) the agent addresses it by, and the name
`allowed_tools` filters against.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"64"}}

### spec.harnesses[].tools[].type

`string`

The tool kind.

- rule: {"string":{"in":["remote_mcp","agentcore_browser","agentcore_gateway","inline_function","agentcore_code_interpreter"]}}

### spec.harnesses[].tools[].remoteMcp

`AwsBedrockAgentCoreHarnessRemoteMcpTool`

A remote MCP server the agent calls over HTTP.

### spec.harnesses[].tools[].remoteMcp.url

`string` · required · sensitive

The MCP server URL. Treated as sensitive: MCP URLs routinely embed
tokens or tenant identifiers.

- rule: {"required":true}

### spec.harnesses[].tools[].remoteMcp.headers

`map<string, string>` · sensitive

HTTP headers sent on every MCP call. Treated as sensitive: the
common use is an Authorization header.

### spec.harnesses[].tools[].agentcoreBrowser

`AwsBedrockAgentCoreHarnessBrowserTool`

An AgentCore managed browser.

### spec.harnesses[].tools[].agentcoreBrowser.browserArn

`string | valueFrom`

The browser to use. Accepts a browser ARN or a reference to an
AwsBedrockAgentCoreTools resource's `browser_arns` output entry
(reference the map key explicitly). Omitted = the AWS default
browser (aws.browser.v1).

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.harnesses[].tools[].agentcoreGateway

`AwsBedrockAgentCoreHarnessGatewayTool`

An AgentCore gateway (MCP tool front door).

### spec.harnesses[].tools[].agentcoreGateway.gatewayArn

`string | valueFrom` · required

The gateway fronting the tools.

- references: AwsBedrockAgentCoreGateway (`status.outputs.gateway_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsBedrockAgentCoreGateway, name: <that resource's name>, fieldPath: status.outputs.gateway_arn}} -- a bare string does not parse

### spec.harnesses[].tools[].agentcoreGateway.outboundAuth

`AwsBedrockAgentCoreHarnessGatewayOutboundAuth`

How the harness authenticates to the gateway. Omitted = the
gateway's default (AWS IAM).

- rule: set exactly one of aws_iam, no_auth, or oauth

### spec.harnesses[].tools[].agentcoreGateway.outboundAuth.awsIam

`bool`

Sign gateway calls with AWS IAM (SigV4).

### spec.harnesses[].tools[].agentcoreGateway.outboundAuth.noAuth

`bool`

Call the gateway unauthenticated.

### spec.harnesses[].tools[].agentcoreGateway.outboundAuth.oauth

`AwsBedrockAgentCoreHarnessGatewayOAuth`

Authenticate with OAuth through an AgentCore credential provider.

### spec.harnesses[].tools[].agentcoreGateway.outboundAuth.oauth.providerArn

`string | valueFrom` · required

The AgentCore OAuth2 credential provider. Accepts a provider ARN or
a reference to an AwsBedrockAgentCoreIdentity resource's
`oauth2_provider_arns` output entry (reference the map key
explicitly).

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.harnesses[].tools[].agentcoreGateway.outboundAuth.oauth.scopes

`[]string` · required

OAuth scopes requested for the token.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.harnesses[].tools[].agentcoreGateway.outboundAuth.oauth.customParameters

`map<string, string>`

Extra token-request parameters the provider requires.

### spec.harnesses[].tools[].agentcoreGateway.outboundAuth.oauth.defaultReturnUrl

`string`

Return URL for the AUTHORIZATION_CODE flow.

### spec.harnesses[].tools[].agentcoreGateway.outboundAuth.oauth.grantType

`string`

The OAuth grant used to obtain tokens. Omitted = the provider
default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["CLIENT_CREDENTIALS","AUTHORIZATION_CODE","TOKEN_EXCHANGE"]}}

### spec.harnesses[].tools[].inlineFunction

`AwsBedrockAgentCoreHarnessInlineFunctionTool`

A function the harness defines inline.

### spec.harnesses[].tools[].inlineFunction.description

`string` · required

What the function does (the agent reads this to decide when to call
it).

- rule: {"string":{"minLen":"1"}}

### spec.harnesses[].tools[].inlineFunction.inputSchema

`string` · required · sensitive

JSON Schema of the function's input, as a JSON document. Treated as
sensitive: schemas routinely describe proprietary internal APIs.

- rule: {"required":true}

### spec.harnesses[].tools[].agentcoreCodeInterpreter

`AwsBedrockAgentCoreHarnessCodeInterpreterTool`

An AgentCore managed code interpreter.

### spec.harnesses[].tools[].agentcoreCodeInterpreter.codeInterpreterArn

`string | valueFrom`

The code interpreter to use. Accepts an interpreter ARN or a
reference to an AwsBedrockAgentCoreTools resource's
`code_interpreter_arns` output entry (reference the map key
explicitly). Omitted = the AWS default interpreter
(aws.codeinterpreter.v1).

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.harnesses[].skillPaths

`[]string`

Skill bundle paths loaded into the harness.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.harnesses[].memory

`AwsBedrockAgentCoreHarnessMemory`

AgentCore memory the harness reads/writes during runs. Omitted =
AWS auto-provisions a managed memory for the harness (the common
case) - the harness still HAS a memory, it is just AWS-owned.

### spec.harnesses[].memory.memoryArn

`string | valueFrom` · required

The AgentCore memory the harness reads/writes.

- references: AwsBedrockAgentCoreMemory (`status.outputs.memory_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsBedrockAgentCoreMemory, name: <that resource's name>, fieldPath: status.outputs.memory_arn}} -- a bare string does not parse

### spec.harnesses[].memory.actorId

`string`

The memory actor the harness runs as. Omitted = AWS assigns one per
run.

### spec.harnesses[].memory.messagesCount

`int32`

Short-term conversation turns loaded into context. Omitted = the
AWS default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.harnesses[].memory.retrieval

`AwsBedrockAgentCoreHarnessMemoryRetrieval`

Long-term retrieval against one memory namespace.

### spec.harnesses[].memory.retrieval.namespace

`string` · required

The memory namespace to retrieve from (e.g.
"/strategies/{memoryStrategyId}/actors/{actorId}").

- rule: {"string":{"minLen":"1"}}

### spec.harnesses[].memory.retrieval.relevanceScore

`double` · optional (explicit presence)

Minimum relevance score (0-1) a record needs to be retrieved.
Omitted = the AWS default.

- rule: {"double":{"lte":1,"gte":0}}

### spec.harnesses[].memory.retrieval.strategyId

`string`

Restrict retrieval to one memory strategy. Omitted = all strategies
in the namespace.

### spec.harnesses[].memory.retrieval.topK

`int32`

Maximum records retrieved per query. Omitted = the AWS default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.harnesses[].environmentVariables

`map<string, string>` · sensitive

Environment variables injected into the harness runtime. Treated as
sensitive: env maps routinely carry tokens and connection strings.

### spec.harnesses[].runtimeEnvironment

`AwsBedrockAgentCoreHarnessRuntimeEnvironment`

Explicit AgentCore runtime environment the harness executes in.
Omitted = AWS provisions and manages a default runtime environment
(the common case); set it to pin the harness to an existing runtime
or to control networking/filesystems/lifecycle.

### spec.harnesses[].runtimeEnvironment.agentRuntimeArn

`string | valueFrom`

The agent runtime hosting the harness's environment. Omitted = AWS
provisions a managed environment with the settings below. (AWS also
accepts the runtime's ID or name; the ARN is the canonical form and
the only one the modules send.)

- references: AwsBedrockAgentCoreRuntime (`status.outputs.agent_runtime_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsBedrockAgentCoreRuntime, name: <that resource's name>, fieldPath: status.outputs.agent_runtime_arn}} -- a bare string does not parse

### spec.harnesses[].runtimeEnvironment.filesystems

`[]AwsBedrockAgentCoreHarnessFilesystem`

Filesystems mounted into the environment.

- rule: each filesystem must set exactly one of efs_access_point_arn, s3_access_point_arn, or session_storage

### spec.harnesses[].runtimeEnvironment.filesystems[].mountPath

`string` · required

Where the filesystem mounts inside the environment.

- rule: {"string":{"minLen":"1"}}

### spec.harnesses[].runtimeEnvironment.filesystems[].efsAccessPointArn

`string | valueFrom`

Mount an EFS access point.

- references: AwsEfsAccessPoint (`status.outputs.access_point_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsEfsAccessPoint, name: <that resource's name>, fieldPath: status.outputs.access_point_arn}} -- a bare string does not parse

### spec.harnesses[].runtimeEnvironment.filesystems[].s3AccessPointArn

`string | valueFrom`

Mount an S3 access point.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.harnesses[].runtimeEnvironment.filesystems[].sessionStorage

`bool`

Mount ephemeral per-session storage.

### spec.harnesses[].runtimeEnvironment.lifecycle

`AwsBedrockAgentCoreHarnessLifecycle`

Session lifecycle tuning.

### spec.harnesses[].runtimeEnvironment.lifecycle.idleRuntimeSessionTimeoutSeconds

`int32`

Seconds an idle session survives before reclamation. Omitted = the
AWS default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.harnesses[].runtimeEnvironment.lifecycle.maxLifetimeSeconds

`int32`

Maximum seconds any session lives. Omitted = the AWS default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.harnesses[].runtimeEnvironment.network

`AwsBedrockAgentCoreHarnessNetwork`

Environment networking.

- rule: vpc_config is required when mode is VPC and forbidden otherwise

### spec.harnesses[].runtimeEnvironment.network.mode

`string`

PUBLIC gives the environment AWS-managed outbound internet; VPC
attaches it to your subnets.

- rule: {"string":{"in":["PUBLIC","VPC"]}}

### spec.harnesses[].runtimeEnvironment.network.vpcConfig

`AwsBedrockAgentCoreHarnessVpcConfig`

VPC placement - required when mode is VPC.

### spec.harnesses[].runtimeEnvironment.network.vpcConfig.subnets

`[]string | valueFrom` · required

Subnets the environment's network interfaces attach to (at least
one).

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.harnesses[].runtimeEnvironment.network.vpcConfig.securityGroups

`[]string | valueFrom` · required

Security groups on those interfaces (at least one).

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.harnesses[].runtimeEnvironment.network.vpcConfig.requireServiceS3Endpoint

`bool`

Require S3 traffic to ride a VPC endpoint instead of the internet
path.

### spec.harnesses[].containerImageUri

`string`

Container image for the harness's environment artifact - run the
agent under test from your own image instead of the managed
environment's default.

### spec.harnesses[].customJwtAuthorizer

`AwsBedrockAgentCoreJwtAuthorizer`

Inbound JWT authorization for invoking the harness. Omitted = AWS
IAM (SigV4) only.

### spec.harnesses[].customJwtAuthorizer.discoveryUrl

`string` · required

The provider's OIDC discovery URL (must serve
/.well-known/openid-configuration).

- rule: {"string":{"minLen":"1"}}

### spec.harnesses[].customJwtAuthorizer.allowedAudience

`[]string`

Accepted "aud" claim values. A token must match at least one entry
of at least one configured allow-list.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.harnesses[].customJwtAuthorizer.allowedClients

`[]string`

Accepted OAuth client IDs (the "client_id"/"azp" claim).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.harnesses[].customJwtAuthorizer.allowedScopes

`[]string`

Accepted OAuth scopes.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.harnesses[].customJwtAuthorizer.allowedWorkloads

`AwsBedrockAgentCoreAllowedWorkloads`

Restrict which AgentCore workload identities may present tokens.

### spec.harnesses[].customJwtAuthorizer.allowedWorkloads.workloadIdentities

`[]string` · required

Workload identity names allowed to call (1-10).

- rule: {"repeated":{"minItems":"1","maxItems":"10","items":{"string":{"minLen":"1"}}}}

### spec.harnesses[].customJwtAuthorizer.allowedWorkloads.hostingEnvironmentArns

`[]string` · required

Hosting environment ARNs allowed to call (1-10).

- rule: {"repeated":{"minItems":"1","maxItems":"10","items":{"string":{"minLen":"1"}}}}

### spec.harnesses[].customJwtAuthorizer.customClaims

`[]AwsBedrockAgentCoreCustomClaim`

Additional claim-matching rules a token must satisfy beyond the
standard audience/client/scope checks.

- rule: custom claim must set exactly one of match_value or match_values

### spec.harnesses[].customJwtAuthorizer.customClaims[].claimName

`string` · required

The inbound token claim to inspect (1-255 characters; letters,
digits, and _ . - :).

- rule: {"string":{"minLen":"1","maxLen":"255","pattern":"^[A-Za-z0-9_.:-]+$"}}

### spec.harnesses[].customJwtAuthorizer.customClaims[].valueType

`string`

Whether the claim's value is a single STRING or a STRING_ARRAY.

- rule: {"string":{"in":["STRING","STRING_ARRAY"]}}

### spec.harnesses[].customJwtAuthorizer.customClaims[].matchOperator

`string`

How the claim value is compared: EQUALS (exact), CONTAINS (the
value appears), or CONTAINS_ANY (any of the expected values
appears).

- rule: {"string":{"in":["EQUALS","CONTAINS","CONTAINS_ANY"]}}

### spec.harnesses[].customJwtAuthorizer.customClaims[].matchValue

`string`

Expected value when matching a single string.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"255","pattern":"^[A-Za-z0-9_.:-]+$"}}

### spec.harnesses[].customJwtAuthorizer.customClaims[].matchValues

`[]string`

Expected values when matching against a list.

- rule: {"repeated":{"items":{"string":{"minLen":"1","maxLen":"255","pattern":"^[A-Za-z0-9_.:-]+$"}}}}

### spec.harnesses[].customJwtAuthorizer.privateEndpoint

`AwsBedrockAgentCorePrivateEndpoint`

Reach a PRIVATE OIDC provider through your VPC instead of the
public internet.

- rule: private endpoint must set exactly one of managed_vpc or self_managed_lattice

### spec.harnesses[].customJwtAuthorizer.privateEndpoint.managedVpc

`AwsBedrockAgentCoreManagedVpcEndpoint`

AWS manages VPC endpoints in your subnets.

### spec.harnesses[].customJwtAuthorizer.privateEndpoint.managedVpc.vpcId

`string | valueFrom` · required

The VPC to route through.

Containment-exempt: a managed endpoint is a private PATH through the
VPC to a private provider or backend; the harness is not deployed inside
it. On a diagram it stands where its own network places it, with a
line into the VPC -- the verdict a Lambda's VPC subnets and an
EventBridge pipe's task subnets already carry.

- references: AwsVpc (`status.outputs.vpc_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.harnesses[].customJwtAuthorizer.privateEndpoint.managedVpc.subnetIds

`[]string | valueFrom` · required

Subnets for the managed endpoint's network interfaces (at least
one).
Containment-exempt for the same reason as the VPC above.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.harnesses[].customJwtAuthorizer.privateEndpoint.managedVpc.securityGroupIds

`[]string | valueFrom`

Security groups on the endpoint interfaces (max 5).

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"repeated":{"maxItems":"5"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.harnesses[].customJwtAuthorizer.privateEndpoint.managedVpc.endpointIpAddressType

`string`

Whether the endpoint answers IPV4 or IPV6. AWS requires the choice
on the harness's managed endpoint.

- rule: {"string":{"in":["IPV4","IPV6"]}}

### spec.harnesses[].customJwtAuthorizer.privateEndpoint.managedVpc.routingDomain

`string` · required

Domain the endpoint routes (3-255 characters). Omitted = derived
from the target.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"3","maxLen":"255"}}

### spec.harnesses[].customJwtAuthorizer.privateEndpoint.managedVpc.tags

`map<string, string>`

Tags applied to the AWS-managed endpoint resources (the module
always adds the Planton identity tags).

### spec.harnesses[].customJwtAuthorizer.privateEndpoint.selfManagedLattice

`AwsBedrockAgentCoreLatticeEndpoint`

You bring a VPC Lattice resource configuration.

### spec.harnesses[].customJwtAuthorizer.privateEndpoint.selfManagedLattice.resourceConfigurationId

`string` · required

The Lattice resource-configuration identifier.

- rule: {"string":{"minLen":"1"}}

### spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides

`[]AwsBedrockAgentCorePrivateEndpointOverride`

Per-domain overrides of the private endpoint (max 5) - route
specific issuer domains through different private paths.

- rule: {"repeated":{"maxItems":"5"}}

### spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].domain

`string` · required

The domain this override captures (1-253 characters).

- rule: {"string":{"minLen":"1","maxLen":"253"}}

### spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint

`AwsBedrockAgentCorePrivateEndpoint` · required

The private path for that domain.

- rule: {"required":true}
- rule: private endpoint must set exactly one of managed_vpc or self_managed_lattice

### spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc

`AwsBedrockAgentCoreManagedVpcEndpoint`

AWS manages VPC endpoints in your subnets.

### spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.vpcId

`string | valueFrom` · required

The VPC to route through.

Containment-exempt: a managed endpoint is a private PATH through the
VPC to a private provider or backend; the harness is not deployed inside
it. On a diagram it stands where its own network places it, with a
line into the VPC -- the verdict a Lambda's VPC subnets and an
EventBridge pipe's task subnets already carry.

- references: AwsVpc (`status.outputs.vpc_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.subnetIds

`[]string | valueFrom` · required

Subnets for the managed endpoint's network interfaces (at least
one).
Containment-exempt for the same reason as the VPC above.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.securityGroupIds

`[]string | valueFrom`

Security groups on the endpoint interfaces (max 5).

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"repeated":{"maxItems":"5"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.endpointIpAddressType

`string`

Whether the endpoint answers IPV4 or IPV6. AWS requires the choice
on the harness's managed endpoint.

- rule: {"string":{"in":["IPV4","IPV6"]}}

### spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.routingDomain

`string` · required

Domain the endpoint routes (3-255 characters). Omitted = derived
from the target.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"3","maxLen":"255"}}

### spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.tags

`map<string, string>`

Tags applied to the AWS-managed endpoint resources (the module
always adds the Planton identity tags).

### spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.selfManagedLattice

`AwsBedrockAgentCoreLatticeEndpoint`

You bring a VPC Lattice resource configuration.

### spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.selfManagedLattice.resourceConfigurationId

`string` · required

The Lattice resource-configuration identifier.

- rule: {"string":{"minLen":"1"}}

### spec.harnesses[].allowedTools

`[]string`

Restrict runs to these tool names (a subset of `tools`). Omitted =
all configured tools are allowed.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.harnesses[].maxIterations

`int32`

Maximum agent iterations (tool-call loops) per run. Omitted = the
AWS default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.harnesses[].maxTokens

`int32`

Maximum tokens per run. Omitted = the AWS default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.harnesses[].timeoutSeconds

`int32`

Seconds before a run times out. Omitted = the AWS default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.harnesses[].truncation

`AwsBedrockAgentCoreHarnessTruncation`

How conversation history is truncated when it outgrows the context
window.

- rule: set exactly the config matching strategy: sliding_window with 'sliding_window', summarization with 'summarization', neither with 'none'

### spec.harnesses[].truncation.strategy

`string`

The truncation strategy.

- rule: {"string":{"in":["sliding_window","summarization","none"]}}

### spec.harnesses[].truncation.slidingWindow

`AwsBedrockAgentCoreHarnessSlidingWindow`

Keep only the most recent messages - pairs with strategy
"sliding_window".

### spec.harnesses[].truncation.slidingWindow.messagesCount

`int32`

Messages kept in the window.

- rule: {"int32":{"gte":1}}

### spec.harnesses[].truncation.summarization

`AwsBedrockAgentCoreHarnessSummarization`

Summarize older history - pairs with strategy "summarization".

### spec.harnesses[].truncation.summarization.summaryRatio

`double` · optional (explicit presence)

Fraction of the context (0-1) the summary may occupy.

- rule: {"double":{"lte":1,"gte":0}}

### spec.harnesses[].truncation.summarization.preserveRecentMessages

`int32`

Recent messages preserved verbatim (never summarized).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.harnesses[].truncation.summarization.summarizationSystemPrompt

`string` · sensitive

Prompt driving the summarizer. Treated as sensitive like the other
prompt surfaces.

### spec.onlineEvaluationConfigs

`[]AwsBedrockAgentCoreOnlineEvaluationConfig`

Continuous scoring of sampled production sessions.

### spec.onlineEvaluationConfigs[].name

`string` · required

Config name in AWS (a letter, then up to 47 letters/digits/
underscores). The for_each key on both engines and the key in the
`online_evaluation_config_ids` output map. AWS exposes no rename -
changing it replaces the config.

- rule: {"string":{"minLen":"1","pattern":"^[a-zA-Z][a-zA-Z0-9_]{0,47}$"}}

### spec.onlineEvaluationConfigs[].description

`string`

What this config evaluates (1-200 characters).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"200"}}

### spec.onlineEvaluationConfigs[].executionRoleArn

`string | valueFrom` · required

IAM role the evaluation service assumes to read the source logs and
invoke evaluators. Must trust bedrock-agentcore.amazonaws.com.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.onlineEvaluationConfigs[].enabled

`bool` · optional (explicit presence)

Whether continuous evaluation runs. Omitted = enabled. The modules
map this onto AWS's create-time intent AND the post-create
execution status, so the field stays declarative across the
lifecycle.

### spec.onlineEvaluationConfigs[].dataSource

`AwsBedrockAgentCoreOnlineEvaluationDataSource` · required

Where the agent session traces come from.

- rule: {"required":true}

### spec.onlineEvaluationConfigs[].dataSource.logGroupNames

`[]string | valueFrom` · required

CloudWatch log groups holding the session traces (at least one) -
the log groups AgentCore observability writes to. Neither the
provider nor the AWS API documents an upper bound, so none is
imposed here.

- references: AwsCloudwatchLogGroup (`status.outputs.log_group_name`)
- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCloudwatchLogGroup, name: <that resource's name>, fieldPath: status.outputs.log_group_name}} -- a bare string does not parse

### spec.onlineEvaluationConfigs[].dataSource.serviceNames

`[]string` · required

Service names (the OTel service.name resource attribute) whose
traces are evaluated (at least one).

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.onlineEvaluationConfigs[].evaluatorIds

`[]string` · required

Evaluators to run over sampled sessions (1-10). Entries are AWS
builtins ("Builtin.Helpfulness"), full custom evaluator IDs, or
names of evaluators defined in this bundle. AWS treats the list as
a SET - duplicates would silently collapse and the 1-10 bound
would measure the wrong count, so uniqueness is enforced here.

- rule: {"repeated":{"minItems":"1","maxItems":"10","unique":true,"items":{"string":{"minLen":"1"}}}}

### spec.onlineEvaluationConfigs[].rule

`AwsBedrockAgentCoreOnlineEvaluationRule` · required

What to sample and score.

- rule: {"required":true}

### spec.onlineEvaluationConfigs[].rule.samplingPercentage

`double`

Percentage of sessions sampled for evaluation (0.01-100). Every
sampled session incurs evaluator cost - production configs usually
sample low single digits.

- rule: {"double":{"lte":100,"gte":0.01}}

### spec.onlineEvaluationConfigs[].rule.filters

`[]AwsBedrockAgentCoreOnlineEvaluationFilter`

Only sessions matching ALL filters are evaluated (max 5).

- rule: {"repeated":{"maxItems":"5"}}
- rule: each filter must set exactly one of string_value, boolean_value, or double_value

### spec.onlineEvaluationConfigs[].rule.filters[].key

`string` · required

The session attribute to match (1-256 characters).

- rule: {"string":{"minLen":"1","maxLen":"256"}}

### spec.onlineEvaluationConfigs[].rule.filters[].operator

`string`

The comparison operator.

- rule: {"string":{"in":["Equals","NotEquals","GreaterThan","LessThan","GreaterThanOrEqual","LessThanOrEqual","Contains","NotContains"]}}

### spec.onlineEvaluationConfigs[].rule.filters[].stringValue

`string`

Match against a string (1-1024 characters).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"1024"}}

### spec.onlineEvaluationConfigs[].rule.filters[].booleanValue

`bool` · optional (explicit presence)

Match against a boolean.

### spec.onlineEvaluationConfigs[].rule.filters[].doubleValue

`double` · optional (explicit presence)

Match against a number.

### spec.onlineEvaluationConfigs[].rule.sessionTimeoutMinutes

`int32`

Minutes of inactivity after which a session is considered complete
and eligible for evaluation (1-60). Omitted = the AWS default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":60,"gte":1}}

## Validation Rules

- `at_least_one_evaluation_resource`: set at least one of evaluators, harnesses, or online_evaluation_configs
- `evaluator_names_unique`: evaluators entries must have unique names
- `harness_names_unique`: harnesses entries must have unique names
- `online_evaluation_config_names_unique`: online_evaluation_configs entries must have unique names
- `online_config_evaluators_resolvable`: each online_evaluation_configs evaluator entry must be an AWS builtin (Builtin.*), a full custom evaluator ID (name-XXXXXXXXXX), or the name of an evaluator defined in this spec

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsBedrockAgentCoreEvaluation, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.evaluator_ids` | `map<string, string>` | Evaluator IDs keyed by each `evaluators` entry's name. |
| `status.outputs.evaluator_arns` | `map<string, string>` | Evaluator ARNs keyed by each `evaluators` entry's name. |
| `status.outputs.harness_ids` | `map<string, string>` | Harness IDs keyed by each `harnesses` entry's name. |
| `status.outputs.harness_arns` | `map<string, string>` | Harness ARNs keyed by each `harnesses` entry's name. |
| `status.outputs.online_evaluation_config_ids` | `map<string, string>` | Online evaluation config IDs keyed by each `online_evaluation_configs` entry's name. |
| `status.outputs.online_evaluation_config_arns` | `map<string, string>` | Online evaluation config ARNs keyed by each `online_evaluation_configs` entry's name. |
| `status.outputs.online_evaluation_output_log_groups` | `map<string, string>` | CloudWatch log group each online config writes its evaluation results to (server-assigned), keyed by the entry's name. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.evaluators[].kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.evaluators[].codeBased.lambdaArn` | AwsLambda | `status.outputs.function_arn` |
| `spec.harnesses[].executionRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.harnesses[].model.gemini.apiKeyArn` | AwsSecretsManagerSecret | `status.outputs.secret_arn` |
| `spec.harnesses[].model.openai.apiKeyArn` | AwsSecretsManagerSecret | `status.outputs.secret_arn` |
| `spec.harnesses[].tools[].agentcoreGateway.gatewayArn` | AwsBedrockAgentCoreGateway | `status.outputs.gateway_arn` |
| `spec.harnesses[].memory.memoryArn` | AwsBedrockAgentCoreMemory | `status.outputs.memory_arn` |
| `spec.harnesses[].runtimeEnvironment.agentRuntimeArn` | AwsBedrockAgentCoreRuntime | `status.outputs.agent_runtime_arn` |
| `spec.harnesses[].runtimeEnvironment.filesystems[].efsAccessPointArn` | AwsEfsAccessPoint | `status.outputs.access_point_arn` |
| `spec.harnesses[].runtimeEnvironment.network.vpcConfig.subnets` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.harnesses[].runtimeEnvironment.network.vpcConfig.securityGroups` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.harnesses[].customJwtAuthorizer.privateEndpoint.managedVpc.vpcId` | AwsVpc | `status.outputs.vpc_id` |
| `spec.harnesses[].customJwtAuthorizer.privateEndpoint.managedVpc.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.harnesses[].customJwtAuthorizer.privateEndpoint.managedVpc.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.vpcId` | AwsVpc | `status.outputs.vpc_id` |
| `spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.onlineEvaluationConfigs[].executionRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.onlineEvaluationConfigs[].dataSource.logGroupNames` | AwsCloudwatchLogGroup | `status.outputs.log_group_name` |

## See Also

- [Overview](../README.md)

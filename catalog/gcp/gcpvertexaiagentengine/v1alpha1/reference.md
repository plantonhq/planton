# GcpVertexAiAgentEngine

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpVertexAiAgentEngineSpec defines a Vertex AI Agent Engine instance
(`google_vertex_ai_reasoning_engine`) -- the managed runtime an AI agent
runs in. Agent Engine builds the agent from source (or runs a container
you bring), hosts it as an autoscaled service with its own identity, and
optionally gives it a Memory Bank of long-term memories.

The shape a manifest usually takes: `spec.source_code_spec` with an
inline source archive and a `python_spec` naming the ADK root agent;
`spec.deployment_spec` for environment, secrets, and instance bounds;
`context_spec.memory_bank_config` when the agent should remember across
sessions. Deployments from source run a Cloud Build in the project.

Immutable: location and the encryption key. Everything else -- the
source, the deployment shape, the memory bank -- updates in place; a
new source archive redeploys the agent's code.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiAgentEngine
metadata:
  name: support-agent
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  displayName: Support agent
  description: Answers support questions from the docs
  spec:
    agentFramework: google-adk
    # Build from source: a base64 .tar.gz of the agent's root (agent.py and
    # requirements.txt -- the readable copy is e2e/fixtures/agent-source/),
    # imported as root_agent from the agent module.
    sourceCodeSpec:
      inlineSource:
        sourceArchive: H4sIAGxTsmoAA+1U24rbMBDdZ3/F4JfuQuw414eFFNJtH0q3UMpSKKUErT22p5UlryTnsvTjO5ITCnlIobC9kYODHM9FR2c0IypULm13F0+ILMvm0yn4lXG8ZqPJNIPRbDybTmej2WQO2Wg8n8wuIHtKUgd01gnDVIzW7pTfpkaUJ+zHh/tHEMfxXY1gGyElWgdLfx/gJa5R6rbx72/IgQhfe9srVZFC2JCUcN+RLECoAkyn0ii6q8kCP86n1J3JEe6xJrb7L6SkjxQmr2nt/wKOcdgIRSVvne5EI0Mu9o1sjkoY0vYaBFhSlUS4vX27Z7IhV4PS4LSWdsBbhfwt17AE3KLJyWJgEQWGSWl0k+wJtYJjL2+k7gp4EfgzkX34F8zdAD6gcbiF5etnFt7tXK1Vn+bqwM6f1lGDgYfuHBTYoiqYJWjFTjsm4dAoIcGiWVOOrM177NXy8QcJRMlugAU5jr2OgMG3EfLHEhJIbiAXTkhdDau89b91ICYoiIChDkMvYUlb1xm0w2A4HDSFb3AvLM6nEVc5irwIUGnNUqai+JoGZy5X02qzL24U+TZY9SIv+m+XgZYSDS7iVgrltFrxpr1TPAjWRhcoF3GFDSlKxuksKaWw9d5aoM0NtY60WsRLZTdo7I+qPnRcfAq6FWCdbm26jyNlnenyPu6j7lg2lgxYMsurtcSty8ML+pS+jJqvlw3asOSc5Sr60/31t8PgQ0cGfafb1G1PjsBfxU/mv7cezX9+m5zn/+9APw+S3I/DRBA3uCu1aT6F7l71M8YOeFx8fr4YpaMsS7NzT51xxhln/Af4DhtSz9UADgAA
      pythonSpec:
        version: "3.12"
        entrypointModule: agent
        entrypointObject: root_agent
    deploymentSpec:
      minInstances: 1
      maxInstances: 5
      env:
        - name: LOG_LEVEL
          value: info
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.spec` | `GcpVertexAiAgentEngineSpecConfig` |  |  |  |
| `spec.spec.agentFramework` | `string` |  |  |  |
| `spec.spec.classMethods` | `string` |  |  |  |
| `spec.spec.identityType` | `string` |  |  |  |
| `spec.spec.serviceAccount` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.spec.containerSpec` | `GcpVertexAiAgentEngineContainerSpec` |  |  |  |
| `spec.spec.containerSpec.imageUri` | `string` | yes |  |  |
| `spec.spec.containerSpec.port` | `int32` |  |  |  |
| `spec.spec.sourceCodeSpec` | `GcpVertexAiAgentEngineSourceCodeSpec` |  |  |  |
| `spec.spec.sourceCodeSpec.inlineSource` | `GcpVertexAiAgentEngineInlineSource` |  |  |  |
| `spec.spec.sourceCodeSpec.inlineSource.sourceArchive` | `string` | yes |  |  |
| `spec.spec.sourceCodeSpec.developerConnectSource` | `GcpVertexAiAgentEngineDeveloperConnectSource` |  |  |  |
| `spec.spec.sourceCodeSpec.developerConnectSource.config` | `GcpVertexAiAgentEngineDeveloperConnectSourceConfig` | yes |  |  |
| `spec.spec.sourceCodeSpec.developerConnectSource.config.gitRepositoryLink` | `string` | yes |  |  |
| `spec.spec.sourceCodeSpec.developerConnectSource.config.dir` | `string` | yes |  |  |
| `spec.spec.sourceCodeSpec.developerConnectSource.config.revision` | `string` | yes |  |  |
| `spec.spec.sourceCodeSpec.agentConfigSource` | `GcpVertexAiAgentEngineAgentConfigSource` |  |  |  |
| `spec.spec.sourceCodeSpec.agentConfigSource.adkConfig` | `GcpVertexAiAgentEngineAdkConfig` | yes |  |  |
| `spec.spec.sourceCodeSpec.agentConfigSource.adkConfig.jsonConfig` | `string` | yes |  |  |
| `spec.spec.sourceCodeSpec.agentConfigSource.inlineSource` | `GcpVertexAiAgentEngineInlineSource` |  |  |  |
| `spec.spec.sourceCodeSpec.agentConfigSource.inlineSource.sourceArchive` | `string` | yes |  |  |
| `spec.spec.sourceCodeSpec.pythonSpec` | `GcpVertexAiAgentEnginePythonSpec` |  |  |  |
| `spec.spec.sourceCodeSpec.pythonSpec.version` | `string` |  |  |  |
| `spec.spec.sourceCodeSpec.pythonSpec.entrypointModule` | `string` |  |  |  |
| `spec.spec.sourceCodeSpec.pythonSpec.entrypointObject` | `string` |  |  |  |
| `spec.spec.sourceCodeSpec.pythonSpec.requirementsFile` | `string` |  |  |  |
| `spec.spec.sourceCodeSpec.imageSpec` | `GcpVertexAiAgentEngineImageSpec` |  |  |  |
| `spec.spec.sourceCodeSpec.imageSpec.buildArgs` | `map<string, string>` |  |  |  |
| `spec.spec.packageSpec` | `GcpVertexAiAgentEnginePackageSpec` |  |  |  |
| `spec.spec.packageSpec.pickleObjectGcsUri` | `string` |  |  |  |
| `spec.spec.packageSpec.dependencyFilesGcsUri` | `string` |  |  |  |
| `spec.spec.packageSpec.requirementsGcsUri` | `string` |  |  |  |
| `spec.spec.packageSpec.pythonVersion` | `string` |  |  |  |
| `spec.spec.buildSpec` | `GcpVertexAiAgentEngineBuildSpec` |  |  |  |
| `spec.spec.buildSpec.workerPool` | `string` |  |  |  |
| `spec.spec.deploymentSpec` | `GcpVertexAiAgentEngineDeploymentSpec` |  |  |  |
| `spec.spec.deploymentSpec.env` | `[]GcpVertexAiAgentEngineEnvVar` |  |  |  |
| `spec.spec.deploymentSpec.env[].name` | `string` | yes |  |  |
| `spec.spec.deploymentSpec.env[].value` | `string` | yes |  |  |
| `spec.spec.deploymentSpec.secretEnv` | `[]GcpVertexAiAgentEngineSecretEnvVar` |  |  |  |
| `spec.spec.deploymentSpec.secretEnv[].name` | `string` | yes |  |  |
| `spec.spec.deploymentSpec.secretEnv[].secretRef` | `GcpVertexAiAgentEngineSecretRef` | yes |  |  |
| `spec.spec.deploymentSpec.secretEnv[].secretRef.secret` | `string \| valueFrom` | yes |  | GcpSecretManagerSecret (`status.outputs.secret_id`) |
| `spec.spec.deploymentSpec.secretEnv[].secretRef.version` | `string` |  |  |  |
| `spec.spec.deploymentSpec.minInstances` | `int32` |  |  |  |
| `spec.spec.deploymentSpec.maxInstances` | `int32` |  |  |  |
| `spec.spec.deploymentSpec.containerConcurrency` | `int32` |  |  |  |
| `spec.spec.deploymentSpec.resourceLimits` | `map<string, string>` |  |  |  |
| `spec.spec.deploymentSpec.pscInterfaceConfig` | `GcpVertexAiAgentEnginePscInterfaceConfig` |  |  |  |
| `spec.spec.deploymentSpec.pscInterfaceConfig.networkAttachment` | `string` |  |  |  |
| `spec.spec.deploymentSpec.pscInterfaceConfig.dnsPeeringConfigs` | `[]GcpVertexAiAgentEngineDnsPeeringConfig` |  |  |  |
| `spec.spec.deploymentSpec.pscInterfaceConfig.dnsPeeringConfigs[].domain` | `string` | yes |  |  |
| `spec.spec.deploymentSpec.pscInterfaceConfig.dnsPeeringConfigs[].targetProject` | `string \| valueFrom` | yes |  | GcpProject (`status.outputs.project_id`) |
| `spec.spec.deploymentSpec.pscInterfaceConfig.dnsPeeringConfigs[].targetNetwork` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_name`) |
| `spec.spec.deploymentSpec.agentGatewayConfig` | `GcpVertexAiAgentEngineAgentGatewayConfig` |  |  |  |
| `spec.spec.deploymentSpec.agentGatewayConfig.clientToAgentConfig` | `GcpVertexAiAgentEngineGatewayTarget` |  |  |  |
| `spec.spec.deploymentSpec.agentGatewayConfig.clientToAgentConfig.agentGateway` | `string` | yes |  |  |
| `spec.spec.deploymentSpec.agentGatewayConfig.agentToAnywhereConfig` | `GcpVertexAiAgentEngineGatewayTarget` |  |  |  |
| `spec.spec.deploymentSpec.agentGatewayConfig.agentToAnywhereConfig.agentGateway` | `string` | yes |  |  |
| `spec.contextSpec` | `GcpVertexAiAgentEngineContextSpec` |  |  |  |
| `spec.contextSpec.memoryBankConfig` | `GcpVertexAiAgentEngineMemoryBankConfig` |  |  |  |
| `spec.contextSpec.memoryBankConfig.generationConfig` | `GcpVertexAiAgentEngineGenerationConfig` |  |  |  |
| `spec.contextSpec.memoryBankConfig.generationConfig.model` | `string` | yes |  |  |
| `spec.contextSpec.memoryBankConfig.generationConfig.generationTriggerConfig` | `GcpVertexAiAgentEngineGenerationTriggerConfig` |  |  |  |
| `spec.contextSpec.memoryBankConfig.generationConfig.generationTriggerConfig.generationRule` | `GcpVertexAiAgentEngineGenerationRule` |  |  |  |
| `spec.contextSpec.memoryBankConfig.generationConfig.generationTriggerConfig.generationRule.eventCount` | `int32` |  |  |  |
| `spec.contextSpec.memoryBankConfig.generationConfig.generationTriggerConfig.generationRule.fixedInterval` | `string` |  |  |  |
| `spec.contextSpec.memoryBankConfig.generationConfig.generationTriggerConfig.generationRule.idleDuration` | `string` |  |  |  |
| `spec.contextSpec.memoryBankConfig.generationConfig.generationTriggerConfig.generationRule.overlapEventCount` | `int32` |  |  |  |
| `spec.contextSpec.memoryBankConfig.similaritySearchConfig` | `GcpVertexAiAgentEngineSimilaritySearchConfig` |  |  |  |
| `spec.contextSpec.memoryBankConfig.similaritySearchConfig.embeddingModel` | `string` | yes |  |  |
| `spec.contextSpec.memoryBankConfig.ttlConfig` | `GcpVertexAiAgentEngineTtlConfig` |  |  |  |
| `spec.contextSpec.memoryBankConfig.ttlConfig.defaultTtl` | `string` |  |  |  |
| `spec.contextSpec.memoryBankConfig.ttlConfig.granularTtlConfig` | `GcpVertexAiAgentEngineGranularTtlConfig` |  |  |  |
| `spec.contextSpec.memoryBankConfig.ttlConfig.granularTtlConfig.createTtl` | `string` |  |  |  |
| `spec.contextSpec.memoryBankConfig.ttlConfig.granularTtlConfig.generateCreatedTtl` | `string` |  |  |  |
| `spec.contextSpec.memoryBankConfig.ttlConfig.granularTtlConfig.generateUpdatedTtl` | `string` |  |  |  |
| `spec.contextSpec.memoryBankConfig.ttlConfig.memoryRevisionDefaultTtl` | `string` |  |  |  |
| `spec.contextSpec.memoryBankConfig.disableMemoryRevisions` | `bool` |  |  |  |
| `spec.contextSpec.memoryBankConfig.structuredMemoryConfigs` | `[]GcpVertexAiAgentEngineStructuredMemoryConfig` |  |  |  |
| `spec.contextSpec.memoryBankConfig.structuredMemoryConfigs[].scopeKeys` | `[]string` |  |  |  |
| `spec.contextSpec.memoryBankConfig.structuredMemoryConfigs[].schemaConfigs` | `[]GcpVertexAiAgentEngineSchemaConfig` |  |  |  |
| `spec.contextSpec.memoryBankConfig.structuredMemoryConfigs[].schemaConfigs[].id` | `string` | yes |  |  |
| `spec.contextSpec.memoryBankConfig.structuredMemoryConfigs[].schemaConfigs[].memorySchema` | `string` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs` | `[]GcpVertexAiAgentEngineCustomizationConfig` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].scopeKeys` | `[]string` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].memoryTopics` | `[]GcpVertexAiAgentEngineMemoryTopic` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].memoryTopics[].customMemoryTopic` | `GcpVertexAiAgentEngineCustomMemoryTopic` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].memoryTopics[].customMemoryTopic.label` | `string` | yes |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].memoryTopics[].customMemoryTopic.description` | `string` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].memoryTopics[].managedMemoryTopic` | `GcpVertexAiAgentEngineManagedMemoryTopic` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].memoryTopics[].managedMemoryTopic.managedTopicEnum` | `string` | yes |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples` | `[]GcpVertexAiAgentEngineGenerateMemoriesExample` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource` | `GcpVertexAiAgentEngineConversationSource` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events` | `[]GcpVertexAiAgentEngineConversationEvent` | yes |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content` | `GcpVertexAiAgentEngineContent` | yes |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.role` | `string` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts` | `[]GcpVertexAiAgentEngineContentPart` | yes |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].text` | `string` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].thought` | `bool` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].inlineData` | `GcpVertexAiAgentEngineInlineData` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].inlineData.mimeType` | `string` | yes |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].inlineData.data` | `string` | yes |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].fileData` | `GcpVertexAiAgentEngineFileData` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].fileData.mimeType` | `string` | yes |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].fileData.fileUri` | `string` | yes |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].functionCall` | `GcpVertexAiAgentEngineFunctionCall` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].functionCall.id` | `string` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].functionCall.name` | `string` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].functionCall.args` | `string` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].functionResponse` | `GcpVertexAiAgentEngineFunctionResponse` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].functionResponse.id` | `string` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].functionResponse.name` | `string` | yes |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].functionResponse.response` | `string` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].executableCode` | `GcpVertexAiAgentEngineExecutableCode` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].executableCode.id` | `string` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].executableCode.language` | `string` | yes |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].executableCode.code` | `string` | yes |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].codeExecutionResult` | `GcpVertexAiAgentEngineCodeExecutionResult` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].codeExecutionResult.id` | `string` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].codeExecutionResult.outcome` | `string` | yes |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].codeExecutionResult.output` | `string` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].videoMetadata` | `GcpVertexAiAgentEngineVideoMetadata` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].videoMetadata.startOffset` | `string` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].videoMetadata.endOffset` | `string` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].generatedMemories` | `[]GcpVertexAiAgentEngineGeneratedMemory` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].generatedMemories[].fact` | `string` | yes |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].generatedMemories[].topics` | `[]GcpVertexAiAgentEngineGeneratedMemoryTopic` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].generatedMemories[].topics[].customMemoryTopicLabel` | `string` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].generatedMemories[].topics[].managedMemoryTopic` | `string` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].consolidationConfig` | `GcpVertexAiAgentEngineConsolidationConfig` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].consolidationConfig.revisionsPerCandidateCount` | `int32` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].disableNaturalLanguageMemories` | `bool` |  |  |  |
| `spec.contextSpec.memoryBankConfig.customizationConfigs[].enableThirdPersonMemories` | `bool` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the agent lives in: a literal project ID or a
GcpProject reference. If omitted, the provider's default project is
used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The Vertex AI location (region) the agent runs in, e.g. "us-central1".
Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.displayName

`string`

Human-readable name of the agent. Defaults to metadata.name.

### spec.description

`string`

Free-text description of the agent.

### spec.labels

`map<string, string>`

Labels on the agent.

### spec.kmsKeyName

`string | valueFrom`

Customer-managed encryption key protecting the agent's data: a
GcpKmsKey reference or a literal
projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}.
Omit to use Google-managed encryption. Immutable.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.spec

`GcpVertexAiAgentEngineSpecConfig`

The agent: its code, identity, and deployment shape.

- rule: container_spec and source_code_spec cannot both be set
- rule: service_account must not be set when identity_type is AGENT_IDENTITY

### spec.spec.agentFramework

`string`

The open-source framework the agent is built with (e.g. "google-adk",
"langchain", "langgraph", "llama-index", "ag2"); tells Agent Engine
which framework integration to load.

### spec.spec.classMethods

`string`

Declarations of the agent object's class methods, as ONE OpenAPI JSON
string (write it compact; Google normalizes it). Required by Google
when deploying through infrastructure-as-code rather than the SDK,
which infers them from the object.

### spec.spec.identityType

`string`

Which identity the agent runs as: "" or SERVICE_ACCOUNT uses
service_account when set and the project's Vertex AI Reasoning Engine
service agent otherwise; AGENT_IDENTITY gives the agent its own
Agent Identity (service_account must then be unset).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["SERVICE_ACCOUNT","AGENT_IDENTITY"]}}

### spec.spec.serviceAccount

`string | valueFrom`

Custom service account the agent runs as: a GcpServiceAccount
reference (resolving to its email) or a literal email. Empty runs as
the project's Vertex AI Reasoning Engine service agent.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.spec.containerSpec

`GcpVertexAiAgentEngineContainerSpec`

Run a prebuilt container image. Mutually exclusive with
source_code_spec.

### spec.spec.containerSpec.imageUri

`string` · required

Container image URI in Artifact Registry, e.g.
us-central1-docker.pkg.dev/{project}/{repo}/agent:latest. The image
must implement the Agent Engine serving contract.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.spec.containerSpec.port

`int32` · optional (explicit presence)

Port the container listens on (Google's default 8080). Sent only when
set.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.spec.sourceCodeSpec

`GcpVertexAiAgentEngineSourceCodeSpec`

Build the agent from source. Mutually exclusive with container_spec.

- rule: set exactly one of inline_source, developer_connect_source, or agent_config_source
- rule: set exactly one of python_spec or image_spec

### spec.spec.sourceCodeSpec.inlineSource

`GcpVertexAiAgentEngineInlineSource`

Source uploaded inline as a base64 .tar.gz.

### spec.spec.sourceCodeSpec.inlineSource.sourceArchive

`string` · required

The source archive: a gzip-compressed tarball of the source root,
base64-encoded (`tar czf - -C agent-source . | base64`).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.spec.sourceCodeSpec.developerConnectSource

`GcpVertexAiAgentEngineDeveloperConnectSource`

Source pulled from a Developer Connect repository.

### spec.spec.sourceCodeSpec.developerConnectSource.config

`GcpVertexAiAgentEngineDeveloperConnectSourceConfig` · required

The repository, directory, and ref.

- rule: {"required":true}

### spec.spec.sourceCodeSpec.developerConnectSource.config.gitRepositoryLink

`string` · required

The Developer Connect Git repository link, as
projects/*/locations/*/connections/*/gitRepositoryLinks/*. A literal
today: no catalog block produces one yet.

- rule: {"required":true,"string":{"pattern":"^projects/[^/]+/locations/[^/]+/connections/[^/]+/gitRepositoryLinks/[^/]+$"}}

### spec.spec.sourceCodeSpec.developerConnectSource.config.dir

`string` · required

Directory, relative to the repository root, that is the source root.

- rule: {"required":true}

### spec.spec.sourceCodeSpec.developerConnectSource.config.revision

`string` · required

The Git ref to fetch: a branch, a tag, a commit SHA, or any ref.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.spec.sourceCodeSpec.agentConfigSource

`GcpVertexAiAgentEngineAgentConfigSource`

An ADK agent described by configuration.

### spec.spec.sourceCodeSpec.agentConfigSource.adkConfig

`GcpVertexAiAgentEngineAdkConfig` · required

The ADK agent config.

- rule: {"required":true}

### spec.spec.sourceCodeSpec.agentConfigSource.adkConfig.jsonConfig

`string` · required

The ADK agent config as a JSON string (write it compact; Google
normalizes it).

- rule: {"required":true,"string":{"minLen":"2"}}

### spec.spec.sourceCodeSpec.agentConfigSource.inlineSource

`GcpVertexAiAgentEngineInlineSource`

Supporting source files (tools, callbacks) the config refers to, as an
inline archive.

### spec.spec.sourceCodeSpec.agentConfigSource.inlineSource.sourceArchive

`string` · required

The source archive: a gzip-compressed tarball of the source root,
base64-encoded (`tar czf - -C agent-source . | base64`).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.spec.sourceCodeSpec.pythonSpec

`GcpVertexAiAgentEnginePythonSpec`

Build with Vertex AI's Python build (requirements + entrypoint).

### spec.spec.sourceCodeSpec.pythonSpec.version

`string`

Python version: 3.9, 3.10 (Google's default), 3.11, 3.12, or 3.13.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["3.9","3.10","3.11","3.12","3.13"]}}

### spec.spec.sourceCodeSpec.pythonSpec.entrypointModule

`string`

Fully qualified module that defines the agent, relative to the source
root (which is on sys.path), e.g. "path.to.agent". Google's default
"agent".

### spec.spec.sourceCodeSpec.pythonSpec.entrypointObject

`string`

The callable in entrypoint_module that IS the agent. Google's default
"root_agent" (the ADK convention).

### spec.spec.sourceCodeSpec.pythonSpec.requirementsFile

`string`

Path of the requirements file relative to the source root. Google's
default "requirements.txt".

### spec.spec.sourceCodeSpec.imageSpec

`GcpVertexAiAgentEngineImageSpec`

Build from the Dockerfile at the source root.

### spec.spec.sourceCodeSpec.imageSpec.buildArgs

`map<string, string>`

Build arguments passed as --build-arg flags.

### spec.spec.packageSpec

`GcpVertexAiAgentEnginePackageSpec`

The legacy pickled-object package.

### spec.spec.packageSpec.pickleObjectGcsUri

`string`

Cloud Storage URI (gs://...) of the pickled Python object.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^gs://.+"}}

### spec.spec.packageSpec.dependencyFilesGcsUri

`string`

Cloud Storage URI of the dependency files, as a .tar.gz.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^gs://.+"}}

### spec.spec.packageSpec.requirementsGcsUri

`string`

Cloud Storage URI of the requirements.txt.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^gs://.+"}}

### spec.spec.packageSpec.pythonVersion

`string`

Python version: 3.8 through 3.13 (Google's default 3.10).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["3.8","3.9","3.10","3.11","3.12","3.13"]}}

### spec.spec.buildSpec

`GcpVertexAiAgentEngineBuildSpec`

Cloud Build settings for the source build.

### spec.spec.buildSpec.workerPool

`string`

The Cloud Build private worker pool the build runs in, as
projects/{project}/locations/{location}/workerPools/{pool}. A literal
today: the catalog's Cloud Build blocks arrive with their own family.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^projects/[^/]+/locations/[^/]+/workerPools/[^/]+$"}}

### spec.spec.deploymentSpec

`GcpVertexAiAgentEngineDeploymentSpec`

Instances, resources, environment, secrets, networking, gateway.

- rule: max_instances must be at least min_instances
- rule: resource_limits accepts only the keys cpu and memory

### spec.spec.deploymentSpec.env

`[]GcpVertexAiAgentEngineEnvVar`

Literal environment variables.

### spec.spec.deploymentSpec.env[].name

`string` · required

Variable name.

- rule: {"required":true,"string":{"pattern":"^[A-Za-z_][A-Za-z0-9_.-]*$"}}

### spec.spec.deploymentSpec.env[].value

`string` · required

Literal value. Never a credential -- use secret_env for those.

- rule: {"required":true}

### spec.spec.deploymentSpec.secretEnv

`[]GcpVertexAiAgentEngineSecretEnvVar`

Environment variables filled from Secret Manager.

### spec.spec.deploymentSpec.secretEnv[].name

`string` · required

Variable name.

- rule: {"required":true,"string":{"pattern":"^[A-Za-z_][A-Za-z0-9_.-]*$"}}

### spec.spec.deploymentSpec.secretEnv[].secretRef

`GcpVertexAiAgentEngineSecretRef` · required

The secret version the value comes from.

- rule: {"required":true}

### spec.spec.deploymentSpec.secretEnv[].secretRef.secret

`string | valueFrom` · required

The secret, by name in the agent's project: a GcpSecretManagerSecret
reference (resolving to its secret_id output) or a literal short name.
The agent's identity needs roles/secretmanager.secretAccessor on it.

- references: GcpSecretManagerSecret (`status.outputs.secret_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSecretManagerSecret, name: <that resource's name>, fieldPath: status.outputs.secret_id}} -- a bare string does not parse

### spec.spec.deploymentSpec.secretEnv[].secretRef.version

`string`

The version to resolve: a version number or "latest" (Google's
default when empty).

### spec.spec.deploymentSpec.minInstances

`int32` · optional (explicit presence)

Instances kept running at all times (0-10; Google's default 1). Zero
scales the agent to nothing between requests at the cost of cold
starts. Sent only when set.

- rule: {"int32":{"lte":10,"gte":0}}

### spec.spec.deploymentSpec.maxInstances

`int32` · optional (explicit presence)

Most instances the agent may scale to (1-1000; 1-100 when VPC Service
Controls or a PSC interface is on; Google's default 100). Sent only
when set.

- rule: {"int32":{"lte":1000,"gte":1}}

### spec.spec.deploymentSpec.containerConcurrency

`int32` · optional (explicit presence)

Concurrent requests one instance handles (Google's default 9;
recommended 2 x cpu + 1). Sent only when set.

- rule: {"int32":{"gte":1}}

### spec.spec.deploymentSpec.resourceLimits

`map<string, string>`

Per-container limits, keys "cpu" (1, 2, 4, 6, 8) and "memory" (1Gi
... 32Gi); Google's default {cpu: "4", memory: "4Gi"}. Sent only when
set.

### spec.spec.deploymentSpec.pscInterfaceConfig

`GcpVertexAiAgentEnginePscInterfaceConfig`

Private Service Connect interface into a VPC.

### spec.spec.deploymentSpec.pscInterfaceConfig.networkAttachment

`string`

Bare name of a Compute Engine network attachment in the agent's region
and project, created beforehand.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[a-z]([-a-z0-9]*[a-z0-9])?$"}}

### spec.spec.deploymentSpec.pscInterfaceConfig.dnsPeeringConfigs

`[]GcpVertexAiAgentEngineDnsPeeringConfig`

Private DNS zones of other projects the agent may resolve.

### spec.spec.deploymentSpec.pscInterfaceConfig.dnsPeeringConfigs[].domain

`string` · required

DNS suffix of the peered zone, ending with a dot, e.g.
"my-internal-domain.corp.".

- rule: {"required":true,"string":{"pattern":"^([a-z0-9-]+\\.)+$"}}

### spec.spec.deploymentSpec.pscInterfaceConfig.dnsPeeringConfigs[].targetProject

`string | valueFrom` · required

The project hosting the Cloud DNS zone: a GcpProject reference or a
literal project ID. The Vertex AI service agent needs roles/dns.peer
there.

- references: GcpProject (`status.outputs.project_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.spec.deploymentSpec.pscInterfaceConfig.dnsPeeringConfigs[].targetNetwork

`string | valueFrom` · required

The VPC network in target_project where the zone is visible, by bare
name: a GcpVpcNetwork reference (resolving to its network_name output)
or a literal.

- references: GcpVpcNetwork (`status.outputs.network_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_name}} -- a bare string does not parse

### spec.spec.deploymentSpec.agentGatewayConfig

`GcpVertexAiAgentEngineAgentGatewayConfig`

Route traffic through Agent Gateway.

### spec.spec.deploymentSpec.agentGatewayConfig.clientToAgentConfig

`GcpVertexAiAgentEngineGatewayTarget`

Gateway for traffic targeting the agent.

### spec.spec.deploymentSpec.agentGatewayConfig.clientToAgentConfig.agentGateway

`string` · required

The Agent Gateway resource name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.spec.deploymentSpec.agentGatewayConfig.agentToAnywhereConfig

`GcpVertexAiAgentEngineGatewayTarget`

Gateway for traffic originating from the agent.

### spec.spec.deploymentSpec.agentGatewayConfig.agentToAnywhereConfig.agentGateway

`string` · required

The Agent Gateway resource name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.contextSpec

`GcpVertexAiAgentEngineContextSpec`

Context services -- the Memory Bank.

### spec.contextSpec.memoryBankConfig

`GcpVertexAiAgentEngineMemoryBankConfig`

The Memory Bank.

### spec.contextSpec.memoryBankConfig.generationConfig

`GcpVertexAiAgentEngineGenerationConfig`

How memories are generated.

### spec.contextSpec.memoryBankConfig.generationConfig.model

`string` · required

The generation model, as
projects/{project}/locations/{location}/publishers/google/models/{model}.

- rule: {"required":true,"string":{"pattern":"^projects/[^/]+/locations/[^/]+/publishers/[^/]+/models/[^/]+$"}}

### spec.contextSpec.memoryBankConfig.generationConfig.generationTriggerConfig

`GcpVertexAiAgentEngineGenerationTriggerConfig`

When buffered events are turned into memories.

### spec.contextSpec.memoryBankConfig.generationConfig.generationTriggerConfig.generationRule

`GcpVertexAiAgentEngineGenerationRule`

The active rule; omit to flush immediately.

- rule: set at most one of event_count, fixed_interval, or idle_duration

### spec.contextSpec.memoryBankConfig.generationConfig.generationTriggerConfig.generationRule.eventCount

`int32` · optional (explicit presence)

Generate when this many events have accumulated.

- rule: {"int32":{"gte":1}}

### spec.contextSpec.memoryBankConfig.generationConfig.generationTriggerConfig.generationRule.fixedInterval

`string`

Generate at a fixed interval, as a duration string with minute
granularity (e.g. "300s").

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+(\\.[0-9]+)?s$"}}

### spec.contextSpec.memoryBankConfig.generationConfig.generationTriggerConfig.generationRule.idleDuration

`string`

Generate when the stream has been idle this long after its last
event, as a duration string with minute granularity.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+(\\.[0-9]+)?s$"}}

### spec.contextSpec.memoryBankConfig.generationConfig.generationTriggerConfig.generationRule.overlapEventCount

`int32` · optional (explicit presence)

Re-include the last N already-processed events in the next window
for context continuity.

- rule: {"int32":{"gte":0}}

### spec.contextSpec.memoryBankConfig.similaritySearchConfig

`GcpVertexAiAgentEngineSimilaritySearchConfig`

How memories are looked up.

### spec.contextSpec.memoryBankConfig.similaritySearchConfig.embeddingModel

`string` · required

The embedding model used to find similar memories, as
projects/{project}/locations/{location}/publishers/google/models/{model}.

- rule: {"required":true,"string":{"pattern":"^projects/[^/]+/locations/[^/]+/publishers/[^/]+/models/[^/]+$"}}

### spec.contextSpec.memoryBankConfig.ttlConfig

`GcpVertexAiAgentEngineTtlConfig`

Automatic expiry.

- rule: set exactly one of default_ttl or granular_ttl_config

### spec.contextSpec.memoryBankConfig.ttlConfig.defaultTtl

`string`

Default lifetime of every memory, as a duration (e.g. "7776000s").

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+(\\.[0-9]+)?s$"}}

### spec.contextSpec.memoryBankConfig.ttlConfig.granularTtlConfig

`GcpVertexAiAgentEngineGranularTtlConfig`

Lifetimes by how the memory came to be.

### spec.contextSpec.memoryBankConfig.ttlConfig.granularTtlConfig.createTtl

`string`

Lifetime of memories uploaded through CreateMemory, as a duration.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+(\\.[0-9]+)?s$"}}

### spec.contextSpec.memoryBankConfig.ttlConfig.granularTtlConfig.generateCreatedTtl

`string`

Lifetime of memories newly generated by GenerateMemories.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+(\\.[0-9]+)?s$"}}

### spec.contextSpec.memoryBankConfig.ttlConfig.granularTtlConfig.generateUpdatedTtl

`string`

Lifetime reset for memories updated by GenerateMemories.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+(\\.[0-9]+)?s$"}}

### spec.contextSpec.memoryBankConfig.ttlConfig.memoryRevisionDefaultTtl

`string`

Default lifetime of memory REVISIONS (the history behind each memory).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+(\\.[0-9]+)?s$"}}

### spec.contextSpec.memoryBankConfig.disableMemoryRevisions

`bool`

Do not keep memory revisions (the history behind each memory).

### spec.contextSpec.memoryBankConfig.structuredMemoryConfigs

`[]GcpVertexAiAgentEngineStructuredMemoryConfig`

Generate memories into fixed schemas.

### spec.contextSpec.memoryBankConfig.structuredMemoryConfigs[].scopeKeys

`[]string`

Scope keys (e.g. "user_id") this config applies to.

### spec.contextSpec.memoryBankConfig.structuredMemoryConfigs[].schemaConfigs

`[]GcpVertexAiAgentEngineSchemaConfig`

The schemas memories are generated into.

### spec.contextSpec.memoryBankConfig.structuredMemoryConfigs[].schemaConfigs[].id

`string` · required

The schema's identifier.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.contextSpec.memoryBankConfig.structuredMemoryConfigs[].schemaConfigs[].memorySchema

`string`

The memory schema as an OpenAPI Schema Object JSON string (write it
compact).

### spec.contextSpec.memoryBankConfig.customizationConfigs

`[]GcpVertexAiAgentEngineCustomizationConfig`

Per-scope generation tuning.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].scopeKeys

`[]string`

Scope keys (e.g. "user_id", "session_id") this config applies to.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].memoryTopics

`[]GcpVertexAiAgentEngineMemoryTopic`

Topics memories should be associated with.

- rule: set exactly one of custom_memory_topic or managed_memory_topic

### spec.contextSpec.memoryBankConfig.customizationConfigs[].memoryTopics[].customMemoryTopic

`GcpVertexAiAgentEngineCustomMemoryTopic`

An operator-defined topic.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].memoryTopics[].customMemoryTopic.label

`string` · required

The topic's label.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.contextSpec.memoryBankConfig.customizationConfigs[].memoryTopics[].customMemoryTopic.description

`string`

What the topic covers, for the generation model.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].memoryTopics[].managedMemoryTopic

`GcpVertexAiAgentEngineManagedMemoryTopic`

A Google-managed topic.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].memoryTopics[].managedMemoryTopic.managedTopicEnum

`string` · required

USER_PERSONAL_INFO, USER_PREFERENCES, KEY_CONVERSATION_DETAILS, or
EXPLICIT_INSTRUCTIONS.

- rule: {"required":true,"string":{"in":["USER_PERSONAL_INFO","USER_PREFERENCES","KEY_CONVERSATION_DETAILS","EXPLICIT_INSTRUCTIONS"]}}

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples

`[]GcpVertexAiAgentEngineGenerateMemoriesExample`

Worked examples for the generation model.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource

`GcpVertexAiAgentEngineConversationSource`

The input conversation.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events

`[]GcpVertexAiAgentEngineConversationEvent` · required

The conversation, in order.

- rule: {"repeated":{"minItems":"1"}}

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content

`GcpVertexAiAgentEngineContent` · required

The turn's content.

- rule: {"required":true}

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.role

`string`

Who produced the turn: "user" or "model" (Google's default "user").

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["user","model"]}}

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts

`[]GcpVertexAiAgentEngineContentPart` · required

The turn's parts.

- rule: {"repeated":{"minItems":"1"}}
- rule: a part carries exactly one of text, inline_data, file_data, function_call, function_response, executable_code, or code_execution_result

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].text

`string`

Plain text.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].thought

`bool`

Marks the part as the model's reasoning rather than its answer.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].inlineData

`GcpVertexAiAgentEngineInlineData`

Raw media inline.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].inlineData.mimeType

`string` · required

IANA MIME type of the data.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].inlineData.data

`string` · required

The bytes, base64-encoded.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].fileData

`GcpVertexAiAgentEngineFileData`

Media in Cloud Storage.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].fileData.mimeType

`string` · required

IANA MIME type of the file.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].fileData.fileUri

`string` · required

The file's gs:// URI.

- rule: {"required":true,"string":{"pattern":"^gs://.+"}}

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].functionCall

`GcpVertexAiAgentEngineFunctionCall`

A tool call the model made.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].functionCall.id

`string`

Correlates the call with its response.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].functionCall.name

`string`

The function's name.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].functionCall.args

`string`

The arguments as a JSON object string.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].functionResponse

`GcpVertexAiAgentEngineFunctionResponse`

A tool's response.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].functionResponse.id

`string`

The id of the call this answers.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].functionResponse.name

`string` · required

The function's name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].functionResponse.response

`string`

The response as a JSON object string.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].executableCode

`GcpVertexAiAgentEngineExecutableCode`

Code the model produced to run.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].executableCode.id

`string`

Identifier the execution result refers back to.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].executableCode.language

`string` · required

PYTHON or BASH.

- rule: {"required":true,"string":{"in":["PYTHON","BASH"]}}

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].executableCode.code

`string` · required

The code.

- rule: {"required":true}

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].codeExecutionResult

`GcpVertexAiAgentEngineCodeExecutionResult`

The result of running code.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].codeExecutionResult.id

`string`

The executable_code part this result is for.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].codeExecutionResult.outcome

`string` · required

OUTCOME_OK, OUTCOME_FAILED, or OUTCOME_DEADLINE_EXCEEDED.

- rule: {"required":true,"string":{"in":["OUTCOME_OK","OUTCOME_FAILED","OUTCOME_DEADLINE_EXCEEDED"]}}

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].codeExecutionResult.output

`string`

Captured stdout (on success) or stderr (on failure).

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].videoMetadata

`GcpVertexAiAgentEngineVideoMetadata`

Clip bounds when the payload is a video.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].videoMetadata.startOffset

`string`

Start offset into the video, as a duration (e.g. "3.5s").

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+(\\.[0-9]+)?s$"}}

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].conversationSource.events[].content.parts[].videoMetadata.endOffset

`string`

End offset into the video, as a duration.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+(\\.[0-9]+)?s$"}}

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].generatedMemories

`[]GcpVertexAiAgentEngineGeneratedMemory`

The memories expected from it.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].generatedMemories[].fact

`string` · required

The fact the memory records.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].generatedMemories[].topics

`[]GcpVertexAiAgentEngineGeneratedMemoryTopic`

Topics the memory belongs to.

- rule: set exactly one of custom_memory_topic_label or managed_memory_topic

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].generatedMemories[].topics[].customMemoryTopicLabel

`string`

The label of a custom topic declared in memory_topics.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].generateMemoriesExamples[].generatedMemories[].topics[].managedMemoryTopic

`string`

A managed topic: USER_PERSONAL_INFO, USER_PREFERENCES,
KEY_CONVERSATION_DETAILS, or EXPLICIT_INSTRUCTIONS.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["USER_PERSONAL_INFO","USER_PREFERENCES","KEY_CONVERSATION_DETAILS","EXPLICIT_INSTRUCTIONS"]}}

### spec.contextSpec.memoryBankConfig.customizationConfigs[].consolidationConfig

`GcpVertexAiAgentEngineConsolidationConfig`

Revision-merging behavior.

### spec.contextSpec.memoryBankConfig.customizationConfigs[].consolidationConfig.revisionsPerCandidateCount

`int32` · optional (explicit presence)

Revisions considered per memory candidate when consolidating.

- rule: {"int32":{"gte":1}}

### spec.contextSpec.memoryBankConfig.customizationConfigs[].disableNaturalLanguageMemories

`bool`

Turn off natural-language memory generation (structured memories
only).

### spec.contextSpec.memoryBankConfig.customizationConfigs[].enableThirdPersonMemories

`bool`

Write memories in the third person ("The user prefers...").

### spec.deletionPolicy

`string`

What happens to the agent when this resource is destroyed:
  "" / "DELETE" -- the agent and its memories are deleted
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the agent leaves management and keeps running

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpVertexAiAgentEngine, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name of the agent: projects/{project}/locations/{location}/reasoningEngines/{id} -- what the Vertex AI SDK's `agent_engines.get()` and a `:query` call take. |
| `status.outputs.reasoning_engine_id` | `string` | The numeric ID Vertex AI assigned to the agent (the last segment of name). |
| `status.outputs.location` | `string` | The location the agent runs in, for rebuilding resource paths and the regional API host. |
| `status.outputs.create_time` | `string` | RFC 3339 timestamp of the agent's creation. |
| `status.outputs.update_time` | `string` | RFC 3339 timestamp of the agent's last update. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |
| `spec.spec.serviceAccount` | GcpServiceAccount | `status.outputs.email` |
| `spec.spec.deploymentSpec.secretEnv[].secretRef.secret` | GcpSecretManagerSecret | `status.outputs.secret_id` |
| `spec.spec.deploymentSpec.pscInterfaceConfig.dnsPeeringConfigs[].targetProject` | GcpProject | `status.outputs.project_id` |
| `spec.spec.deploymentSpec.pscInterfaceConfig.dnsPeeringConfigs[].targetNetwork` | GcpVpcNetwork | `status.outputs.network_name` |

## See Also

- [Overview](../README.md)

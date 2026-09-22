# GCP Vertex AI Agent Engine

The managed runtime an AI agent runs in. Agent Engine takes your agent -- built with the Agent Development Kit, LangChain, LangGraph, LlamaIndex, or your own framework -- as source code it builds for you or as a container you bring, hosts it as an autoscaled service with its own identity, environment, and secrets, and optionally gives it a Memory Bank: long-term memories generated from conversations, scoped by keys such as `user_id`, retrieved by similarity, expiring on a schedule you set. Everything but the location and the encryption key updates in place; a new source archive redeploys the agent's code.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `aiplatform.googleapis.com` on the project (never disabled on destroy)
- **Agent Engine instance** -- a `vertex_ai_reasoning_engine` with its code source, deployment shape, and memory bank

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/aiplatform.admin` on the project (and `iam.serviceAccounts.actAs` on a custom service account when one is named).
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpServiceAccount`** -- a custom identity for the agent (`spec.serviceAccount`).
- **`GcpSecretManagerSecret`** -- secrets injected as environment variables (`spec.deploymentSpec.secretEnv[]`).
- **`GcpKmsKey`** -- customer-managed encryption (`kmsKeyName`). Immutable.
- **`GcpProject`, `GcpVpcNetwork`** -- the DNS peering targets of a Private Service Connect interface.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiAgentEngine
metadata:
  name: support-agent
spec:
  location: us-central1
  spec:
    agentFramework: google-adk
    sourceCodeSpec:
      inlineSource:
        sourceArchive: <base64 of a .tar.gz containing agent.py and requirements.txt>
      pythonSpec:
        version: "3.12"
        entrypointModule: agent
        entrypointObject: root_agent
```

```shell
planton apply -f agent-engine.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The Vertex AI location (region). Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project, as a literal or a `GcpProject` reference. |
| `displayName` | `string` | `metadata.name` | Human-readable name. |
| `description`, `labels` | | | Descriptive metadata. |
| `kmsKeyName` | `StringValueOrRef` | Google-managed | A `GcpKmsKey` reference or literal key path. Immutable. |
| `spec.agentFramework` | `string` | none | `google-adk`, `langchain`, `langgraph`, `llama-index`, `ag2`, or your own. |
| `spec.classMethods` | `string` | none | The agent object's class methods as one OpenAPI JSON string. |
| `spec.identityType` | `string` | service agent | `SERVICE_ACCOUNT` (uses `serviceAccount` when set) or `AGENT_IDENTITY` (the agent's own identity; `serviceAccount` must be unset). |
| `spec.serviceAccount` | `StringValueOrRef` | Vertex AI service agent | A `GcpServiceAccount` reference or literal email. |
| `spec.containerSpec` | `object` | none | `imageUri`, `port`. Mutually exclusive with `sourceCodeSpec`. |
| `spec.sourceCodeSpec` | `object` | none | Exactly one source (`inlineSource.sourceArchive`, `developerConnectSource.config`, `agentConfigSource.adkConfig`) and exactly one build (`pythonSpec { version, entrypointModule, entrypointObject, requirementsFile }` or `imageSpec { buildArgs }`). |
| `spec.packageSpec` | `object` | none | The legacy pickled-object package: `pickleObjectGcsUri`, `dependencyFilesGcsUri`, `requirementsGcsUri`, `pythonVersion`. |
| `spec.buildSpec.workerPool` | `string` | none | A Cloud Build private worker pool for the source build. |
| `spec.deploymentSpec` | `object` | Google's defaults | `env[]`, `secretEnv[]` (`GcpSecretManagerSecret` references), `minInstances` (0-10), `maxInstances` (1-1000), `containerConcurrency`, `resourceLimits { cpu, memory }`, `pscInterfaceConfig { networkAttachment, dnsPeeringConfigs[] }`, `agentGatewayConfig { clientToAgentConfig, agentToAnywhereConfig }`. |
| `contextSpec.memoryBankConfig` | `object` | none | `generationConfig { model, generationTriggerConfig.generationRule }`, `similaritySearchConfig.embeddingModel`, `ttlConfig`, `disableMemoryRevisions`, `structuredMemoryConfigs[]`, `customizationConfigs[]` (scope keys, topics, worked examples, consolidation, generation flags). |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- `containerSpec` and `sourceCodeSpec` cannot both be set; a source spec has exactly one source and exactly one build recipe.
- `serviceAccount` must be unset under `AGENT_IDENTITY`.
- `maxInstances` is at least `minInstances`; `resourceLimits` accepts only `cpu` and `memory`.
- A conversation part in a memory-bank example carries exactly one payload; a memory topic is exactly custom or managed; a TTL is exactly a default or a granular set; at most one generation trigger.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/reasoningEngines/{id}` |
| `reasoning_engine_id` | `string` | The numeric ID Vertex AI assigned |
| `location` | `string` | The agent's location |
| `create_time`, `update_time` | `string` | RFC 3339 timestamps |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Two knobs are held out on both engines** until pulumi-gcp bridges provider 8.x: `buildSpec.serviceAccount` (the Cloud Build builder's identity) and the `audioTranscription` payload of a memory-bank example part. Neither can be sent from the pinned Pulumi SDK, and an argument one engine cannot send is never a one-engine field.
- **A source build is a Cloud Build in the project.** The Cloud Build API must be enabled and Google's Vertex AI service agent needs its default roles; a build failure naming the requirements file is a source defect.
- **The inline archive is input-only.** Google never reads it back; a changed archive is detected by the manifest diff.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`; the readable source of the sample agent is `e2e/fixtures/agent-source/`.

## Related Components

- **GcpServiceAccount** -- the agent's custom identity
- **GcpSecretManagerSecret** -- secrets injected as environment variables
- **GcpKmsKey** -- customer-managed encryption
- **GcpVertexAiModelGardenDeployment** -- a model the agent calls
- **GcpVectorSearchCollection** -- a vector store the agent's tools search

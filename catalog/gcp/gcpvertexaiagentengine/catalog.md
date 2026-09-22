# GCP Vertex AI Agent Engine

Runs your AI agent as a managed service. Hand Agent Engine the agent's source (it builds it) or a container (it runs it), and Vertex AI hosts the agent behind an authenticated endpoint with its own identity, autoscaling between the instance bounds you set, reading secrets from Secret Manager and reaching private resources over Private Service Connect when you ask. Turn on the Memory Bank and the agent remembers each user across sessions: memories generated from conversations by Gemini, found by similarity, organized by topic, expiring on your schedule. It is the runtime behind agents built with the Agent Development Kit, LangChain, LangGraph, or LlamaIndex.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `aiplatform.googleapis.com` on the project
- **Agent Engine instance** -- a `vertex.AiReasoningEngine` with its code source, identity, deployment shape, and optional Memory Bank

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/aiplatform.admin` on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **GcpServiceAccount** -- a custom identity for the agent
- **GcpSecretManagerSecret** -- secrets injected as environment variables
- **GcpKmsKey** -- customer-managed encryption

## Deploy

### Console

Open the deployment store, find **GCP Vertex AI Agent Engine**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **ADK Agent from Source** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiAgentEngine
metadata:
  name: support-agent
  org: acme-corp
  env: prod
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
    deploymentSpec:
      minInstances: 1
      maxInstances: 5
```

```shell
planton apply -f agent-engine.yaml
```

This builds the agent from its source and runs it between one and five instances. A Stack Job tracks the provisioning in real time; the build is a Cloud Build in the project.

### InfraChart

When deploying as part of a multi-resource environment, reference a `GcpServiceAccount` as the agent's identity, `GcpSecretManagerSecret`s for its secrets, and a `GcpKmsKey` for encryption; the application that talks to the agent reads its `name` output.

## Key Configuration

These are the most important decisions when configuring an agent. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Source or container** -- `sourceCodeSpec` has Vertex AI build the agent: an inline archive, a Developer Connect repository, or an ADK config, built with the Python build (`requirements.txt` plus an entrypoint) or a Dockerfile. `containerSpec` runs an image you built. Never both.

**Identity** -- by default the agent runs as the project's Vertex AI service agent; `serviceAccount` names a custom `GcpServiceAccount`, and `identityType: AGENT_IDENTITY` gives the agent its own identity instead.

**Memory Bank** -- `contextSpec.memoryBankConfig` turns on long-term memory: the Gemini model that generates memories and when it runs, the embedding model that finds them, TTLs, and per-scope topics and worked examples.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId`, `spec.deploymentSpec.pscInterfaceConfig.dnsPeeringConfigs[].targetProject` | `status.outputs.project_id` |
| **GcpServiceAccount** | `spec.serviceAccount` | `status.outputs.email` |
| **GcpSecretManagerSecret** | `spec.deploymentSpec.secretEnv[].secretRef.secret` | `status.outputs.secret_id` |
| **GcpKmsKey** | `kmsKeyName` | `status.outputs.key_id` |
| **GcpVpcNetwork** | `spec.deploymentSpec.pscInterfaceConfig.dnsPeeringConfigs[].targetNetwork` | `status.outputs.network_name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The agent's full resource name | SDK clients; `:query` calls |
| `reasoning_engine_id` | The numeric ID | Dashboards; IAM bindings |
| `location` | The agent's location | Regional clients |
| `create_time`, `update_time` | Timestamps | Audit |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**ADK agent from source** -- Build from an inline archive with the Python build. Start from the **ADK Agent from Source** preset.

**Container agent with Memory Bank** -- Your own image, a custom identity, a secret from Secret Manager, and long-term memory. Start from the **Container Agent with Memory Bank** preset.

## Works With

- [**GCP Service Account**](/cloud-catalog/gcp-service-account) -- the agent's identity
- [**GCP Secret Manager Secret**](/cloud-catalog/gcp-secret-manager-secret) -- secrets as environment variables
- [**GCP KMS Key**](/cloud-catalog/gcp-kms-key) -- customer-managed encryption
- [**GCP Vertex AI Model Garden Deployment**](/cloud-catalog/gcp-vertex-ai-model-garden-deployment) -- a model the agent calls
- [**GCP Vector Search Collection**](/cloud-catalog/gcp-vector-search-collection) -- a vector store the agent searches

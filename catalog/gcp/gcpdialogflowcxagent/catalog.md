# GCP Dialogflow CX Agent

Stands up a conversational agent for chat, voice, and telephony on Google's Dialogflow CX, together with everything its conversations rely on: the webhooks that look up orders or book appointments, the tools a generative playbook calls, frozen releases of your flows served from production and staging environments, and the generative AI settings per language. Design the conversations in the Dialogflow console; declare the agent and its moving parts here, so they deploy, version, and tear down like the rest of your stack.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `dialogflow.googleapis.com` on the project
- **Agent** -- a `diagflow.CxAgent`, with its default start flow and default playbook
- **Webhooks, tools, and tool versions** -- one `diagflow.CxWebhook`, `diagflow.CxTool`, or `diagflow.CxToolVersion` per declared entry
- **Flow versions and environments** -- one `diagflow.CxVersion` and `diagflow.CxEnvironment` per declared entry
- **Generative settings** -- one `diagflow.CxGenerativeSettings` per declared language

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Dialogflow admin permissions on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **GcpDialogflowCxSecuritySettings** -- redaction and retention, referenced by `securitySettings`.
- **GcpVertexAiSearchDataStore** -- the stores a data store tool searches.
- **GcpVertexAiSearchEngine** -- the Vertex AI Search engine linked to the agent.
- **GcpServiceAccount** -- the identity a webhook authenticates as.

## Deploy

### Console

Open the deployment store, find **GCP Dialogflow CX Agent**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Flow Agent** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDialogflowCxAgent
metadata:
  name: support-agent
  org: acme-corp
  env: prod
spec:
  location: global
  defaultLanguageCode: en
  timeZone: America/New_York
  webhooks:
    - displayName: order-lookup
      genericWebService:
        uri: https://orders.example.com/dialogflow
        serviceAgentAuth: ID_TOKEN
```

```shell
planton apply -f dialogflow-cx-agent.yaml
```

This creates a global agent with one webhook its flows can call for order lookups. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, reference a `GcpDialogflowCxSecuritySettings` from `securitySettings`, point data store tools at `GcpVertexAiSearchDataStore` blocks, authenticate webhooks as a `GcpServiceAccount`, and let a `GcpVertexAiSearchEngine` chat engine link the agent by its `name` output.

## Key Configuration

These are the most important decisions when configuring an agent. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Flows or playbooks** -- a flow agent routes each turn through state machines you design; `startWithDefaultPlaybook` makes a generative agent that follows natural-language instructions and calls tools. Playbook turns bill at a higher rate.

**Location** -- `global` needs no setup; a region such as `us-central1` keeps data in that region but needs Google's one-time location settings in the console for a project's first regional agent. Fixed at creation.

**Releases** -- declare `versions` of your flows and `environments` that pin them, so production answers from a frozen release while the draft keeps changing. An environment needs a version for every flow its start flow can reach.

**Privacy** -- attach security settings before turning on logging, and keep credentials in Secret Manager (`secretVersionFor...` fields) rather than in the spec.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpDialogflowCxSecuritySettings** | `securitySettings` | `status.outputs.name` |
| **GcpVertexAiSearchEngine** | `genAppBuilderSettings.engine` | `status.outputs.name` |
| **GcpVertexAiSearchDataStore** | `tools[].dataStoreSpec.dataStoreConnections[].dataStore` | `status.outputs.name` |
| **GcpServiceAccount** | `webhooks[].genericWebService.serviceAccount` | `status.outputs.email` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The agent's full resource name | A chat engine's `dialogflowAgentToLink` |
| `agent_id` | The id Google assigned | Console links, API clients |
| `location` | The agent's location | Regional clients |
| `start_flow` | The default start flow's full name | Console content, API clients |
| `webhook_names`, `tool_names`, `tool_version_names` | The declared webhooks', tools', and tool versions' full names | Flows, pages, and playbooks authored in the console |
| `version_names`, `environment_names` | The declared releases and environments | Detect-intent calls addressed to an environment |
| `generative_settings_names` | The declared generative settings | Diagnostics |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Flow agent** -- A global flow-first agent with spell correction and interaction history. Start from the **Flow Agent** preset.

**Playbook agent with tools** -- A generative agent with an OpenAPI tool, a data store tool, and a knowledge-connector persona. Start from the **Playbook Agent with Tools** preset.

**Production release** -- A regional agent with security settings, a Cloud Run webhook, a frozen release served from production and staging, and `PREVENT` on destroy. Start from the **Production Release** preset.

## Works With

- [**GCP Dialogflow CX Security Settings**](/cloud-catalog/gcp-dialogflow-cx-security-settings) -- redaction, retention, and exports for every conversation
- [**GCP Vertex AI Search Data Store**](/cloud-catalog/gcp-vertex-ai-search-data-store) -- the stores data store tools answer from
- [**GCP Vertex AI Search Engine**](/cloud-catalog/gcp-vertex-ai-search-engine) -- a chat engine that answers through the agent
- [**GCP Cloud Run**](/cloud-catalog/gcp-cloud-run) -- a common home for webhook fulfillment code

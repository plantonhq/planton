# GCP Dialogflow CX Agent

A Dialogflow CX conversational agent -- a virtual agent for chat, voice, and telephony -- together with the infrastructure its conversation content uses: webhooks for fulfillment, tools for generative playbooks (with frozen tool versions), flow versions and the environments that pin them, and generative settings per language. The conversation content itself (flows, pages, intents, entity types, playbooks, generators) is authored in the Dialogflow console or restored from GitHub, and refers to these by the names in the outputs.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `dialogflow.googleapis.com` on the project (never disabled on destroy)
- **Agent** -- a `dialogflow_cx_agent` with its languages, time zone, speech, logging, and integration settings; Google creates its default start flow and default playbook
- **Webhooks** -- one `dialogflow_cx_webhook` per `webhooks[]` entry, keyed by `displayName`
- **Tools** -- one `dialogflow_cx_tool` per `tools[]` entry, keyed by `displayName`, and one `dialogflow_cx_tool_version` per `tools[].versions[]` entry
- **Flow versions** -- one `dialogflow_cx_version` per `versions[]` entry, keyed by flow and `displayName`
- **Environments** -- one `dialogflow_cx_environment` per `environments[]` entry, keyed by `displayName`
- **Generative settings** -- one `dialogflow_cx_generative_settings` per `generativeSettings[]` entry, keyed by `languageCode`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Dialogflow admin permissions on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpDialogflowCxSecuritySettings`** -- redaction and retention for every conversation (`securitySettings`), in the agent's location.
- **`GcpVertexAiSearchDataStore`** -- the stores a data store tool searches (`tools[].dataStoreSpec.dataStoreConnections[].dataStore`).
- **`GcpVertexAiSearchEngine`** -- the Vertex AI Search engine linked to the agent (`genAppBuilderSettings.engine`).
- **`GcpServiceAccount`** -- the identity a webhook authenticates as (`webhooks[].genericWebService.serviceAccount`); the Dialogflow service agent needs `roles/iam.serviceAccountTokenCreator` on it.
- **Secret Manager secret versions** -- webhook and tool credentials named by `projects/{p}/secrets/{s}/versions/{v}` instead of written into the spec.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDialogflowCxAgent
metadata:
  name: support-agent
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

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | `global` or a Dialogflow region. Immutable. |
| `defaultLanguageCode` | `string` | The agent's default language, e.g. `en`. Immutable. |
| `timeZone` | `string` | IANA time zone, e.g. `America/New_York`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project, as a literal or a `GcpProject` reference. |
| `displayName` | `string` | `metadata.name` | Unique within the location. |
| `description`, `avatarUri` | `string` | | Descriptive settings. |
| `supportedLanguageCodes` | `[]string` | none | Languages other than the default. |
| `enableMultiLanguageTraining`, `enableSpellCorrection`, `locked` | `bool` | `false` | Model training, spell correction, and change lock. |
| `securitySettings` | `StringValueOrRef` | none | A `GcpDialogflowCxSecuritySettings` reference or literal name. |
| `startWithDefaultPlaybook` | `bool` | `false` | Begin conversations in the default playbook instead of the start flow. |
| `advancedSettings` | `object` | Google's | `audioExportGcsDestination`, `dtmfSettings`, `loggingSettings`, `speechSettings`. |
| `enableAnswerFeedback`, `enableSpeechAdaptation` | `bool` | `false` | Answer feedback (needs interaction logging) and speech adaptation. |
| `defaultEndUserMetadata`, `synthesizeSpeechConfigs` | `string` | none | JSON object strings. |
| `clientCertificateSettings` | `object` | none | `sslCertificate` (PEM) with `privateKey` and `passphrase` as Secret Manager version names. |
| `genAppBuilderSettings` | `object` | Google's | `engine`: a `GcpVertexAiSearchEngine` reference or literal engine name. |
| `deleteChatEngineOnDestroy` | `bool` | `false` | Delete the linked engine on destroy (literal engine names only). |
| `gitIntegrationSettings.githubSettings` | `object` | none | Repository, branches, and access token (sensitive). |
| `webhooks[]` | `[]object` | none | `displayName`, `disabled`, `timeout`, and exactly one of `genericWebService` or `serviceDirectory { service, genericWebService }`. |
| `tools[]` | `[]object` | none | `displayName`, `description`, exactly one of `openApiSpec`, `dataStoreSpec`, `functionSpec`, and `versions[]` (`displayName`, `tool` snapshot). |
| `versions[]` | `[]object` | none | `flowId` (empty is the start flow), `displayName`, `description`. |
| `environments[]` | `[]object` | none | `displayName`, `description`, `versionConfigs[]` (`flowId`, and `version` -- a declared version's display name -- or `versionId`). |
| `generativeSettings[]` | `[]object` | none | `languageCode`, `fallbackSettings`, `generativeSafetySettings`, `knowledgeConnectorSettings`, `llmModelSettings`. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`; fanned to every webhook, tool, tool version, version, and environment. |

### Validation Rules

- `supportedLanguageCodes` never includes `defaultLanguageCode` (Google's rule).
- A webhook is exactly one of `genericWebService` or `serviceDirectory`; a web service URI is https.
- A tool (and a tool version's snapshot) is exactly one of `openApiSpec`, `dataStoreSpec`, `functionSpec`; an OpenAPI tool uses at most one authentication method.
- Webhook, tool, and environment display names are unique within the agent (Google's rule); flow versions are unique per flow and tool versions per tool (the kind keys them by display name); generative settings are declared once per language.
- Every version an environment names by display name is declared in `versions` for the same flow; a version config names exactly one of `version` or `versionId`.
- `deleteChatEngineOnDestroy` is refused when the engine is a kind reference -- that block owns the engine.
- Secret Manager version names, Service Directory services, and the audio export URI must be well formed; credentials carry no format rule.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/agents/{agent_id}` -- what a chat engine's `dialogflowAgentToLink` takes |
| `agent_id` | `string` | The id Google assigned |
| `location` | `string` | The agent's location |
| `start_flow` | `string` | The default start flow's full name |
| `webhook_names` | `[]string` | Full names of the declared webhooks, in manifest order |
| `tool_names` | `[]string` | Full names of the declared tools, in manifest order |
| `tool_version_names` | `[]string` | Full names of the declared tool versions, tool by tool |
| `version_names` | `[]string` | Full names of the declared flow versions, in manifest order |
| `environment_names` | `[]string` | Full names of the declared environments, in manifest order |
| `generative_settings_names` | `[]string` | `{agent}/generativeSettings?languageCode={language}` per declared language |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **The display name is a child's identity.** Google renames webhooks, tools, versions, and environments in place, but this block keys them by display name, so a rename replaces the child and gives it a new resource name; console content that refers to it must be re-pointed.
- **Flows are named by id.** A version's `flowId` and an environment's `flowId` take the flow's id, because a full flow path contains the agent's id, which does not exist before the first apply. Empty is the start flow.
- **An environment needs a version for every reachable flow.** Google rejects an environment that leaves a flow reachable from the start flow without a version.
- **Generative settings outlive destroy.** Google keeps one set per agent and language; destroying the entry only stops managing it.
- **Values Google never returns.** The start playbook, answer feedback, logging settings, credentials, and prompt templates are not read back, so a change made in the console to one of them is not detected.
- **Tool versions are frozen.** Every field is immutable; a version's snapshot is written out in full, never copied from its tool.
- **Integration Connectors tools** (`connector_spec`) exist only in Google's beta provider and are not offered.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpDialogflowCxSecuritySettings** -- redaction, retention, and exports for the agent's conversations
- **GcpVertexAiSearchDataStore** -- the stores data store tools answer from
- **GcpVertexAiSearchEngine** -- a chat engine that answers through the agent
- **GcpCloudRun** -- a common home for webhook fulfillment code

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

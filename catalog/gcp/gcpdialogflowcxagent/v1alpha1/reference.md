# GcpDialogflowCxAgent

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpDialogflowCxAgentSpec defines a Dialogflow CX agent
(`google_dialogflow_cx_agent`) -- a virtual agent that holds
conversations over chat, voice, and telephony -- together with the
infrastructure its conversation content uses:
  - webhooks (`google_dialogflow_cx_webhook`) its flows call for
    fulfillment
  - tools (`google_dialogflow_cx_tool`) its generative playbooks call,
    each with frozen versions (`google_dialogflow_cx_tool_version`)
  - flow versions (`google_dialogflow_cx_version`) and the environments
    (`google_dialogflow_cx_environment`) that pin them
  - generative settings per language
    (`google_dialogflow_cx_generative_settings`)
Each is folded in because it lives under exactly one agent and has no
other root. The conversation content itself -- flows, pages, intents,
entity types, playbooks, generators -- is authored in the Dialogflow
console (or restored from GitHub) and refers to these by the names in
the outputs.

Google creates every agent with a default start flow and a default
playbook; start_with_default_playbook chooses which one conversations
begin in. Security settings (redaction, retention, audio export) are a
block of their own, GcpDialogflowCxSecuritySettings, shared by agents in
the same location.

Immutable: project_id, location, and default_language_code (a change
replaces the agent and everything in it). Everything else updates in
place.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDialogflowCxAgent
metadata:
  name: support-agent
spec:
  projectId:
    value: my-gcp-project
  location: global
  displayName: Support agent
  description: Answers order and shipping questions
  defaultLanguageCode: en
  timeZone: America/New_York
  enableSpellCorrection: true
  advancedSettings:
    loggingSettings:
      enableInteractionLogging: true
  # Fulfillment the agent's console-authored flows call by the webhook's
  # full name (the webhook_names output).
  webhooks:
    - displayName: order-lookup
      timeout: 10s
      genericWebService:
        uri: https://orders.example.com/dialogflow
        serviceAgentAuth: ID_TOKEN
  # A tool the agent's playbooks can call, frozen once as v1.
  tools:
    - displayName: order-status
      description: Looks up an order's status by order number.
      functionSpec:
        inputSchema: '{"type": "object", "properties": {"order_id": {"type": "string"}}}'
      versions:
        - displayName: v1
          tool:
            displayName: order-status
            description: Looks up an order's status by order number.
            functionSpec:
              inputSchema: '{"type": "object", "properties": {"order_id": {"type": "string"}}}'
  # Freeze the start flow and serve the frozen copy from production.
  versions:
    - displayName: release-1
  environments:
    - displayName: production
      versionConfigs:
        - version: release-1
  generativeSettings:
    - languageCode: en
      knowledgeConnectorSettings:
        business: Acme
        agentIdentity: AI assistant
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.defaultLanguageCode` | `string` | yes |  |  |
| `spec.timeZone` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.avatarUri` | `string` |  |  |  |
| `spec.supportedLanguageCodes` | `[]string` |  |  |  |
| `spec.enableMultiLanguageTraining` | `bool` |  |  |  |
| `spec.enableSpellCorrection` | `bool` |  |  |  |
| `spec.locked` | `bool` |  |  |  |
| `spec.securitySettings` | `string \| valueFrom` |  |  | GcpDialogflowCxSecuritySettings (`status.outputs.name`) |
| `spec.startWithDefaultPlaybook` | `bool` |  |  |  |
| `spec.advancedSettings` | `GcpDialogflowCxAgentAdvancedSettings` |  |  |  |
| `spec.advancedSettings.audioExportGcsDestination` | `GcpDialogflowCxAgentAudioExportGcsDestination` |  |  |  |
| `spec.advancedSettings.audioExportGcsDestination.uri` | `string` |  |  |  |
| `spec.advancedSettings.dtmfSettings` | `GcpDialogflowCxAgentDtmfSettings` |  |  |  |
| `spec.advancedSettings.dtmfSettings.enabled` | `bool` |  |  |  |
| `spec.advancedSettings.dtmfSettings.finishDigit` | `string` |  |  |  |
| `spec.advancedSettings.dtmfSettings.maxDigits` | `int32` |  |  |  |
| `spec.advancedSettings.loggingSettings` | `GcpDialogflowCxAgentLoggingSettings` |  |  |  |
| `spec.advancedSettings.loggingSettings.enableConsentBasedRedaction` | `bool` |  |  |  |
| `spec.advancedSettings.loggingSettings.enableInteractionLogging` | `bool` |  |  |  |
| `spec.advancedSettings.loggingSettings.enableStackdriverLogging` | `bool` |  |  |  |
| `spec.advancedSettings.speechSettings` | `GcpDialogflowCxAgentSpeechSettings` |  |  |  |
| `spec.advancedSettings.speechSettings.endpointerSensitivity` | `int32` |  |  |  |
| `spec.advancedSettings.speechSettings.models` | `map<string, string>` |  |  |  |
| `spec.advancedSettings.speechSettings.noSpeechTimeout` | `string` |  |  |  |
| `spec.advancedSettings.speechSettings.useTimeoutBasedEndpointing` | `bool` |  |  |  |
| `spec.enableAnswerFeedback` | `bool` |  |  |  |
| `spec.clientCertificateSettings` | `GcpDialogflowCxAgentClientCertificateSettings` |  |  |  |
| `spec.clientCertificateSettings.sslCertificate` | `string` | yes |  |  |
| `spec.clientCertificateSettings.privateKey` | `string` |  |  |  |
| `spec.clientCertificateSettings.passphrase` | `string` |  |  |  |
| `spec.genAppBuilderSettings` | `GcpDialogflowCxAgentGenAppBuilderSettings` |  |  |  |
| `spec.genAppBuilderSettings.engine` | `string \| valueFrom` | yes |  | GcpVertexAiSearchEngine (`status.outputs.name`) |
| `spec.deleteChatEngineOnDestroy` | `bool` |  |  |  |
| `spec.gitIntegrationSettings` | `GcpDialogflowCxAgentGitIntegrationSettings` |  |  |  |
| `spec.gitIntegrationSettings.githubSettings` | `GcpDialogflowCxAgentGithubSettings` |  |  |  |
| `spec.gitIntegrationSettings.githubSettings.displayName` | `string` |  |  |  |
| `spec.gitIntegrationSettings.githubSettings.repositoryUri` | `string` |  |  |  |
| `spec.gitIntegrationSettings.githubSettings.trackingBranch` | `string` |  |  |  |
| `spec.gitIntegrationSettings.githubSettings.branches` | `[]string` |  |  |  |
| `spec.gitIntegrationSettings.githubSettings.accessToken` | `string` (sensitive) |  |  |  |
| `spec.defaultEndUserMetadata` | `string` |  |  |  |
| `spec.enableSpeechAdaptation` | `bool` |  |  |  |
| `spec.synthesizeSpeechConfigs` | `string` |  |  |  |
| `spec.webhooks` | `[]GcpDialogflowCxAgentWebhook` |  |  |  |
| `spec.webhooks[].displayName` | `string` | yes |  |  |
| `spec.webhooks[].disabled` | `bool` |  |  |  |
| `spec.webhooks[].timeout` | `string` |  |  |  |
| `spec.webhooks[].genericWebService` | `GcpDialogflowCxAgentGenericWebService` |  |  |  |
| `spec.webhooks[].genericWebService.uri` | `string` |  |  |  |
| `spec.webhooks[].genericWebService.webhookType` | `string` |  |  |  |
| `spec.webhooks[].genericWebService.httpMethod` | `string` |  |  |  |
| `spec.webhooks[].genericWebService.requestBody` | `string` |  |  |  |
| `spec.webhooks[].genericWebService.parameterMapping` | `map<string, string>` |  |  |  |
| `spec.webhooks[].genericWebService.requestHeaders` | `map<string, string>` |  |  |  |
| `spec.webhooks[].genericWebService.secretVersionsForRequestHeaders` | `[]GcpDialogflowCxAgentSecretHeader` |  |  |  |
| `spec.webhooks[].genericWebService.secretVersionsForRequestHeaders[].key` | `string` | yes |  |  |
| `spec.webhooks[].genericWebService.secretVersionsForRequestHeaders[].secretVersion` | `string` |  |  |  |
| `spec.webhooks[].genericWebService.secretVersionForUsernamePassword` | `string` |  |  |  |
| `spec.webhooks[].genericWebService.oauthConfig` | `GcpDialogflowCxAgentOauthConfig` |  |  |  |
| `spec.webhooks[].genericWebService.oauthConfig.clientId` | `string` | yes |  |  |
| `spec.webhooks[].genericWebService.oauthConfig.tokenEndpoint` | `string` | yes |  |  |
| `spec.webhooks[].genericWebService.oauthConfig.clientSecret` | `string` (sensitive) |  |  |  |
| `spec.webhooks[].genericWebService.oauthConfig.scopes` | `[]string` |  |  |  |
| `spec.webhooks[].genericWebService.oauthConfig.secretVersionForClientSecret` | `string` |  |  |  |
| `spec.webhooks[].genericWebService.serviceAgentAuth` | `string` |  |  |  |
| `spec.webhooks[].genericWebService.serviceAccount` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.webhooks[].genericWebService.allowedCaCerts` | `[]string` |  |  |  |
| `spec.webhooks[].serviceDirectory` | `GcpDialogflowCxAgentServiceDirectory` |  |  |  |
| `spec.webhooks[].serviceDirectory.service` | `string` |  |  |  |
| `spec.webhooks[].serviceDirectory.genericWebService` | `GcpDialogflowCxAgentGenericWebService` |  |  |  |
| `spec.webhooks[].serviceDirectory.genericWebService.uri` | `string` |  |  |  |
| `spec.webhooks[].serviceDirectory.genericWebService.webhookType` | `string` |  |  |  |
| `spec.webhooks[].serviceDirectory.genericWebService.httpMethod` | `string` |  |  |  |
| `spec.webhooks[].serviceDirectory.genericWebService.requestBody` | `string` |  |  |  |
| `spec.webhooks[].serviceDirectory.genericWebService.parameterMapping` | `map<string, string>` |  |  |  |
| `spec.webhooks[].serviceDirectory.genericWebService.requestHeaders` | `map<string, string>` |  |  |  |
| `spec.webhooks[].serviceDirectory.genericWebService.secretVersionsForRequestHeaders` | `[]GcpDialogflowCxAgentSecretHeader` |  |  |  |
| `spec.webhooks[].serviceDirectory.genericWebService.secretVersionsForRequestHeaders[].key` | `string` | yes |  |  |
| `spec.webhooks[].serviceDirectory.genericWebService.secretVersionsForRequestHeaders[].secretVersion` | `string` |  |  |  |
| `spec.webhooks[].serviceDirectory.genericWebService.secretVersionForUsernamePassword` | `string` |  |  |  |
| `spec.webhooks[].serviceDirectory.genericWebService.oauthConfig` | `GcpDialogflowCxAgentOauthConfig` |  |  |  |
| `spec.webhooks[].serviceDirectory.genericWebService.oauthConfig.clientId` | `string` | yes |  |  |
| `spec.webhooks[].serviceDirectory.genericWebService.oauthConfig.tokenEndpoint` | `string` | yes |  |  |
| `spec.webhooks[].serviceDirectory.genericWebService.oauthConfig.clientSecret` | `string` (sensitive) |  |  |  |
| `spec.webhooks[].serviceDirectory.genericWebService.oauthConfig.scopes` | `[]string` |  |  |  |
| `spec.webhooks[].serviceDirectory.genericWebService.oauthConfig.secretVersionForClientSecret` | `string` |  |  |  |
| `spec.webhooks[].serviceDirectory.genericWebService.serviceAgentAuth` | `string` |  |  |  |
| `spec.webhooks[].serviceDirectory.genericWebService.serviceAccount` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.webhooks[].serviceDirectory.genericWebService.allowedCaCerts` | `[]string` |  |  |  |
| `spec.tools` | `[]GcpDialogflowCxAgentTool` |  |  |  |
| `spec.tools[].displayName` | `string` | yes |  |  |
| `spec.tools[].description` | `string` | yes |  |  |
| `spec.tools[].openApiSpec` | `GcpDialogflowCxAgentToolOpenApiSpec` |  |  |  |
| `spec.tools[].openApiSpec.textSchema` | `string` | yes |  |  |
| `spec.tools[].openApiSpec.authentication` | `GcpDialogflowCxAgentToolAuthentication` |  |  |  |
| `spec.tools[].openApiSpec.authentication.apiKeyConfig` | `GcpDialogflowCxAgentToolApiKeyConfig` |  |  |  |
| `spec.tools[].openApiSpec.authentication.apiKeyConfig.keyName` | `string` | yes |  |  |
| `spec.tools[].openApiSpec.authentication.apiKeyConfig.requestLocation` | `string` | yes |  |  |
| `spec.tools[].openApiSpec.authentication.apiKeyConfig.apiKey` | `string` (sensitive) |  |  |  |
| `spec.tools[].openApiSpec.authentication.apiKeyConfig.secretVersionForApiKey` | `string` |  |  |  |
| `spec.tools[].openApiSpec.authentication.bearerTokenConfig` | `GcpDialogflowCxAgentToolBearerTokenConfig` |  |  |  |
| `spec.tools[].openApiSpec.authentication.bearerTokenConfig.token` | `string` (sensitive) |  |  |  |
| `spec.tools[].openApiSpec.authentication.bearerTokenConfig.secretVersionForToken` | `string` |  |  |  |
| `spec.tools[].openApiSpec.authentication.oauthConfig` | `GcpDialogflowCxAgentToolOauthConfig` |  |  |  |
| `spec.tools[].openApiSpec.authentication.oauthConfig.clientId` | `string` | yes |  |  |
| `spec.tools[].openApiSpec.authentication.oauthConfig.oauthGrantType` | `string` | yes |  |  |
| `spec.tools[].openApiSpec.authentication.oauthConfig.tokenEndpoint` | `string` | yes |  |  |
| `spec.tools[].openApiSpec.authentication.oauthConfig.clientSecret` | `string` (sensitive) |  |  |  |
| `spec.tools[].openApiSpec.authentication.oauthConfig.scopes` | `[]string` |  |  |  |
| `spec.tools[].openApiSpec.authentication.oauthConfig.secretVersionForClientSecret` | `string` |  |  |  |
| `spec.tools[].openApiSpec.authentication.serviceAgentAuthConfig` | `GcpDialogflowCxAgentToolServiceAgentAuthConfig` |  |  |  |
| `spec.tools[].openApiSpec.authentication.serviceAgentAuthConfig.serviceAgentAuth` | `string` |  |  |  |
| `spec.tools[].openApiSpec.serviceDirectoryConfig` | `GcpDialogflowCxAgentToolServiceDirectoryConfig` |  |  |  |
| `spec.tools[].openApiSpec.serviceDirectoryConfig.service` | `string` |  |  |  |
| `spec.tools[].openApiSpec.tlsConfig` | `GcpDialogflowCxAgentToolTlsConfig` |  |  |  |
| `spec.tools[].openApiSpec.tlsConfig.caCerts` | `[]GcpDialogflowCxAgentToolCaCert` | yes |  |  |
| `spec.tools[].openApiSpec.tlsConfig.caCerts[].displayName` | `string` | yes |  |  |
| `spec.tools[].openApiSpec.tlsConfig.caCerts[].cert` | `string` | yes |  |  |
| `spec.tools[].dataStoreSpec` | `GcpDialogflowCxAgentToolDataStoreSpec` |  |  |  |
| `spec.tools[].dataStoreSpec.dataStoreConnections` | `[]GcpDialogflowCxAgentToolDataStoreConnection` | yes |  |  |
| `spec.tools[].dataStoreSpec.dataStoreConnections[].dataStore` | `string \| valueFrom` |  |  | GcpVertexAiSearchDataStore (`status.outputs.name`) |
| `spec.tools[].dataStoreSpec.dataStoreConnections[].dataStoreType` | `string` |  |  |  |
| `spec.tools[].dataStoreSpec.dataStoreConnections[].documentProcessingMode` | `string` |  |  |  |
| `spec.tools[].functionSpec` | `GcpDialogflowCxAgentToolFunctionSpec` |  |  |  |
| `spec.tools[].functionSpec.inputSchema` | `string` |  |  |  |
| `spec.tools[].functionSpec.outputSchema` | `string` |  |  |  |
| `spec.tools[].versions` | `[]GcpDialogflowCxAgentToolVersion` |  |  |  |
| `spec.tools[].versions[].displayName` | `string` | yes |  |  |
| `spec.tools[].versions[].tool` | `GcpDialogflowCxAgentToolSnapshot` | yes |  |  |
| `spec.tools[].versions[].tool.displayName` | `string` | yes |  |  |
| `spec.tools[].versions[].tool.description` | `string` | yes |  |  |
| `spec.tools[].versions[].tool.openApiSpec` | `GcpDialogflowCxAgentToolOpenApiSpec` |  |  |  |
| `spec.tools[].versions[].tool.openApiSpec.textSchema` | `string` | yes |  |  |
| `spec.tools[].versions[].tool.openApiSpec.authentication` | `GcpDialogflowCxAgentToolAuthentication` |  |  |  |
| `spec.tools[].versions[].tool.openApiSpec.authentication.apiKeyConfig` | `GcpDialogflowCxAgentToolApiKeyConfig` |  |  |  |
| `spec.tools[].versions[].tool.openApiSpec.authentication.apiKeyConfig.keyName` | `string` | yes |  |  |
| `spec.tools[].versions[].tool.openApiSpec.authentication.apiKeyConfig.requestLocation` | `string` | yes |  |  |
| `spec.tools[].versions[].tool.openApiSpec.authentication.apiKeyConfig.apiKey` | `string` (sensitive) |  |  |  |
| `spec.tools[].versions[].tool.openApiSpec.authentication.apiKeyConfig.secretVersionForApiKey` | `string` |  |  |  |
| `spec.tools[].versions[].tool.openApiSpec.authentication.bearerTokenConfig` | `GcpDialogflowCxAgentToolBearerTokenConfig` |  |  |  |
| `spec.tools[].versions[].tool.openApiSpec.authentication.bearerTokenConfig.token` | `string` (sensitive) |  |  |  |
| `spec.tools[].versions[].tool.openApiSpec.authentication.bearerTokenConfig.secretVersionForToken` | `string` |  |  |  |
| `spec.tools[].versions[].tool.openApiSpec.authentication.oauthConfig` | `GcpDialogflowCxAgentToolOauthConfig` |  |  |  |
| `spec.tools[].versions[].tool.openApiSpec.authentication.oauthConfig.clientId` | `string` | yes |  |  |
| `spec.tools[].versions[].tool.openApiSpec.authentication.oauthConfig.oauthGrantType` | `string` | yes |  |  |
| `spec.tools[].versions[].tool.openApiSpec.authentication.oauthConfig.tokenEndpoint` | `string` | yes |  |  |
| `spec.tools[].versions[].tool.openApiSpec.authentication.oauthConfig.clientSecret` | `string` (sensitive) |  |  |  |
| `spec.tools[].versions[].tool.openApiSpec.authentication.oauthConfig.scopes` | `[]string` |  |  |  |
| `spec.tools[].versions[].tool.openApiSpec.authentication.oauthConfig.secretVersionForClientSecret` | `string` |  |  |  |
| `spec.tools[].versions[].tool.openApiSpec.authentication.serviceAgentAuthConfig` | `GcpDialogflowCxAgentToolServiceAgentAuthConfig` |  |  |  |
| `spec.tools[].versions[].tool.openApiSpec.authentication.serviceAgentAuthConfig.serviceAgentAuth` | `string` |  |  |  |
| `spec.tools[].versions[].tool.openApiSpec.serviceDirectoryConfig` | `GcpDialogflowCxAgentToolServiceDirectoryConfig` |  |  |  |
| `spec.tools[].versions[].tool.openApiSpec.serviceDirectoryConfig.service` | `string` |  |  |  |
| `spec.tools[].versions[].tool.openApiSpec.tlsConfig` | `GcpDialogflowCxAgentToolTlsConfig` |  |  |  |
| `spec.tools[].versions[].tool.openApiSpec.tlsConfig.caCerts` | `[]GcpDialogflowCxAgentToolCaCert` | yes |  |  |
| `spec.tools[].versions[].tool.openApiSpec.tlsConfig.caCerts[].displayName` | `string` | yes |  |  |
| `spec.tools[].versions[].tool.openApiSpec.tlsConfig.caCerts[].cert` | `string` | yes |  |  |
| `spec.tools[].versions[].tool.dataStoreSpec` | `GcpDialogflowCxAgentToolDataStoreSpec` |  |  |  |
| `spec.tools[].versions[].tool.dataStoreSpec.dataStoreConnections` | `[]GcpDialogflowCxAgentToolDataStoreConnection` | yes |  |  |
| `spec.tools[].versions[].tool.dataStoreSpec.dataStoreConnections[].dataStore` | `string \| valueFrom` |  |  | GcpVertexAiSearchDataStore (`status.outputs.name`) |
| `spec.tools[].versions[].tool.dataStoreSpec.dataStoreConnections[].dataStoreType` | `string` |  |  |  |
| `spec.tools[].versions[].tool.dataStoreSpec.dataStoreConnections[].documentProcessingMode` | `string` |  |  |  |
| `spec.tools[].versions[].tool.functionSpec` | `GcpDialogflowCxAgentToolFunctionSpec` |  |  |  |
| `spec.tools[].versions[].tool.functionSpec.inputSchema` | `string` |  |  |  |
| `spec.tools[].versions[].tool.functionSpec.outputSchema` | `string` |  |  |  |
| `spec.versions` | `[]GcpDialogflowCxAgentVersion` |  |  |  |
| `spec.versions[].flowId` | `string` |  |  |  |
| `spec.versions[].displayName` | `string` | yes |  |  |
| `spec.versions[].description` | `string` |  |  |  |
| `spec.environments` | `[]GcpDialogflowCxAgentEnvironment` |  |  |  |
| `spec.environments[].displayName` | `string` | yes |  |  |
| `spec.environments[].description` | `string` |  |  |  |
| `spec.environments[].versionConfigs` | `[]GcpDialogflowCxAgentEnvironmentVersionConfig` | yes |  |  |
| `spec.environments[].versionConfigs[].flowId` | `string` |  |  |  |
| `spec.environments[].versionConfigs[].version` | `string` |  |  |  |
| `spec.environments[].versionConfigs[].versionId` | `string` |  |  |  |
| `spec.generativeSettings` | `[]GcpDialogflowCxAgentGenerativeSettings` |  |  |  |
| `spec.generativeSettings[].languageCode` | `string` | yes |  |  |
| `spec.generativeSettings[].fallbackSettings` | `GcpDialogflowCxAgentFallbackSettings` |  |  |  |
| `spec.generativeSettings[].fallbackSettings.selectedPrompt` | `string` |  |  |  |
| `spec.generativeSettings[].fallbackSettings.promptTemplates` | `[]GcpDialogflowCxAgentPromptTemplate` |  |  |  |
| `spec.generativeSettings[].fallbackSettings.promptTemplates[].displayName` | `string` |  |  |  |
| `spec.generativeSettings[].fallbackSettings.promptTemplates[].frozen` | `bool` |  |  |  |
| `spec.generativeSettings[].fallbackSettings.promptTemplates[].promptText` | `string` |  |  |  |
| `spec.generativeSettings[].generativeSafetySettings` | `GcpDialogflowCxAgentGenerativeSafetySettings` |  |  |  |
| `spec.generativeSettings[].generativeSafetySettings.defaultBannedPhraseMatchStrategy` | `string` |  |  |  |
| `spec.generativeSettings[].generativeSafetySettings.bannedPhrases` | `[]GcpDialogflowCxAgentBannedPhrase` |  |  |  |
| `spec.generativeSettings[].generativeSafetySettings.bannedPhrases[].languageCode` | `string` | yes |  |  |
| `spec.generativeSettings[].generativeSafetySettings.bannedPhrases[].text` | `string` | yes |  |  |
| `spec.generativeSettings[].knowledgeConnectorSettings` | `GcpDialogflowCxAgentKnowledgeConnectorSettings` |  |  |  |
| `spec.generativeSettings[].knowledgeConnectorSettings.agent` | `string` |  |  |  |
| `spec.generativeSettings[].knowledgeConnectorSettings.agentIdentity` | `string` |  |  |  |
| `spec.generativeSettings[].knowledgeConnectorSettings.agentScope` | `string` |  |  |  |
| `spec.generativeSettings[].knowledgeConnectorSettings.business` | `string` |  |  |  |
| `spec.generativeSettings[].knowledgeConnectorSettings.businessDescription` | `string` |  |  |  |
| `spec.generativeSettings[].knowledgeConnectorSettings.disableDataStoreFallback` | `bool` |  |  |  |
| `spec.generativeSettings[].llmModelSettings` | `GcpDialogflowCxAgentLlmModelSettings` |  |  |  |
| `spec.generativeSettings[].llmModelSettings.model` | `string` |  |  |  |
| `spec.generativeSettings[].llmModelSettings.promptText` | `string` |  |  |  |
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

The Dialogflow CX location: "global" or a region such as "us-central1"
(data residency). A project's first regional agent needs Google's
one-time location settings, which only the Dialogflow console can set;
"global" needs none. Immutable.

- rule: {"required":true,"string":{"pattern":"^(global|[a-z]+(-[a-z]+[0-9]+)?)$"}}

### spec.displayName

`string`

Human-readable name, unique within the location. Defaults to
metadata.name. Mutable in place.

### spec.defaultLanguageCode

`string` · required

The agent's default language as a language tag, e.g. "en". Immutable.

- rule: {"string":{"minLen":"1"}}

### spec.timeZone

`string` · required

The agent's time zone from the IANA database, e.g. "America/New_York".

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

What the agent is for, at most 500 characters.

- rule: {"string":{"maxLen":"500"}}

### spec.avatarUri

`string`

The avatar shown in the console and the web demo integration.

### spec.supportedLanguageCodes

`[]string`

The agent's languages other than default_language_code, e.g. ["es",
"fr"].

### spec.enableMultiLanguageTraining

`bool`

Train one multilingual model over every supported language instead of
one per language.

### spec.enableSpellCorrection

`bool`

Correct spelling in end-user input before matching.

### spec.locked

`bool`

Lock the agent: Google rejects every change except a restore. Unlock
(set false) before changing anything else.

### spec.securitySettings

`string | valueFrom`

The security settings applied to every conversation: a
GcpDialogflowCxSecuritySettings reference or a literal
projects/{project}/locations/{location}/securitySettings/{id} in the
agent's location.

- references: GcpDialogflowCxSecuritySettings (`status.outputs.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpDialogflowCxSecuritySettings, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.startWithDefaultPlaybook

`bool`

Begin conversations in the agent's default playbook (a generative,
instruction-driven agent) instead of its default start flow (a
state-machine agent). Google allows only the default playbook as a
start playbook. Google does not read this back.

### spec.advancedSettings

`GcpDialogflowCxAgentAdvancedSettings`

Agent-level advanced settings: audio export, touch-tone, logging,
speech.

### spec.advancedSettings.audioExportGcsDestination

`GcpDialogflowCxAgentAudioExportGcsDestination`

Export incoming audio to Cloud Storage.

### spec.advancedSettings.audioExportGcsDestination.uri

`string`

The Cloud Storage URI, gs://bucket/object-name-or-prefix. Whether it is
a full object name or a prefix depends on the Dialogflow operation.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^gs://[^/]+(/.*)?$"}}

### spec.advancedSettings.dtmfSettings

`GcpDialogflowCxAgentDtmfSettings`

Touch-tone detection.

### spec.advancedSettings.dtmfSettings.enabled

`bool`

Process incoming audio for key presses: a caller pressing "3" becomes
an event the agent's flows can route on.

### spec.advancedSettings.dtmfSettings.finishDigit

`string`

The digit that ends a digit sequence, e.g. "#".

### spec.advancedSettings.dtmfSettings.maxDigits

`int32`

The longest digit sequence collected.

- rule: {"int32":{"gte":0}}

### spec.advancedSettings.loggingSettings

`GcpDialogflowCxAgentLoggingSettings`

Conversation logging.

### spec.advancedSettings.loggingSettings.enableConsentBasedRedaction

`bool`

Redact end-user input when the session parameter
$session.params.conversation-redaction is true -- consent-based
redaction.

### spec.advancedSettings.loggingSettings.enableInteractionLogging

`bool`

Log interactions to Dialogflow's interaction history (needed for answer
feedback and conversation history in the console).

### spec.advancedSettings.loggingSettings.enableStackdriverLogging

`bool`

Log conversation queries to Cloud Logging.

### spec.advancedSettings.speechSettings

`GcpDialogflowCxAgentSpeechSettings`

Speech-to-text detection.

### spec.advancedSettings.speechSettings.endpointerSensitivity

`int32`

How eagerly the end of speech is detected, 0 (least) to 100 (most).
Read as seconds of timeout when use_timeout_based_endpointing is set.

- rule: {"int32":{"lte":100,"gte":0}}

### spec.advancedSettings.speechSettings.models

`map<string, string>`

The Speech-to-Text model per language, e.g. {"en": "phone_call"}.

### spec.advancedSettings.speechSettings.noSpeechTimeout

`string`

How long to wait for speech before a no-speech event: seconds with up
to nine fractional digits and a trailing "s", e.g. "3.5s".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+(\\.[0-9]{1,9})?s$"}}

### spec.advancedSettings.speechSettings.useTimeoutBasedEndpointing

`bool`

Interpret endpointer_sensitivity as a timeout in seconds instead of a
sensitivity score.

### spec.enableAnswerFeedback

`bool`

Let end users rate responses. Works only with interaction logging on
(advanced_settings.logging_settings.enable_interaction_logging). Google
does not read this back.

### spec.clientCertificateSettings

`GcpDialogflowCxAgentClientCertificateSettings`

A custom client certificate for the agent's outgoing calls.

### spec.clientCertificateSettings.sslCertificate

`string` · required

The certificate, PEM-encoded, including the BEGIN and END lines.

- rule: {"string":{"minLen":"1"}}

### spec.clientCertificateSettings.privateKey

`string`

The Secret Manager secret version holding the PEM private key:
projects/{project}/secrets/{secret}/versions/{version}. Dialogflow
reads the key from Secret Manager; the key never appears here.

- rule: {"string":{"pattern":"^projects/[^/]+/secrets/[^/]+/versions/[^/]+$"}}

### spec.clientCertificateSettings.passphrase

`string`

The Secret Manager secret version holding the key's passphrase, same
format. Leave empty when the private key is not encrypted.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^projects/[^/]+/secrets/[^/]+/versions/[^/]+$"}}

### spec.genAppBuilderSettings

`GcpDialogflowCxAgentGenAppBuilderSettings`

The Vertex AI Search engine linked to the agent.

### spec.genAppBuilderSettings.engine

`string | valueFrom` · required

The engine: a GcpVertexAiSearchEngine reference or a literal
projects/{project}/locations/{location}/collections/{collection}/engines/{engine}.

- references: GcpVertexAiSearchEngine (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVertexAiSearchEngine, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.deleteChatEngineOnDestroy

`bool`

Delete the linked Vertex AI Search engine (gen_app_builder_settings)
when the agent is destroyed. Google creates that engine when a data
store is connected to the agent, outside any declaration; this is the
only way it goes away with the agent. Refused when the engine is a
kind reference, because that block owns the engine's lifecycle.

### spec.gitIntegrationSettings

`GcpDialogflowCxAgentGitIntegrationSettings`

Git integration for exporting and restoring agent content.

### spec.gitIntegrationSettings.githubSettings

`GcpDialogflowCxAgentGithubSettings`

The GitHub integration.

### spec.gitIntegrationSettings.githubSettings.displayName

`string`

The repository's display name in Dialogflow, unique among the agent's
repositories.

### spec.gitIntegrationSettings.githubSettings.repositoryUri

`string`

The repository URI, e.g. https://api.github.com/repos/acme/support-bot.

### spec.gitIntegrationSettings.githubSettings.trackingBranch

`string`

The branch Dialogflow tracks for the agent.

### spec.gitIntegrationSettings.githubSettings.branches

`[]string`

Every branch Dialogflow may use.

### spec.gitIntegrationSettings.githubSettings.accessToken

`string` · sensitive

The GitHub access token Dialogflow authenticates with. Google returns
it redacted, so the declared value is always what the engines send.

### spec.defaultEndUserMetadata

`string`

Default end-user metadata merged into every DetectIntent request, as a
JSON object string. Prefer templates over constants, e.g.
{"age": "$session.params.age"}.

### spec.enableSpeechAdaptation

`bool`

Use speech adaptation (phrase hints from the agent's content) in speech
recognition.

### spec.synthesizeSpeechConfigs

`string`

Speech synthesis per language, as a JSON object string mapping a
language code to a SynthesizeSpeechConfig, e.g.
{"en": {"voice": {"name": "en-US-Neural2-C"}}}. Applies to the phone
gateway and to DetectIntent responses that request audio.

### spec.webhooks

`[]GcpDialogflowCxAgentWebhook`

Webhooks, keyed by display name.

- rule: a webhook is exactly one of generic_web_service or service_directory

### spec.webhooks[].displayName

`string` · required

The webhook's human-readable name, unique within the agent.

- rule: {"string":{"minLen":"1"}}

### spec.webhooks[].disabled

`bool`

Keep the webhook but stop calling it.

### spec.webhooks[].timeout

`string`

How long Dialogflow waits for the webhook, e.g. "5s". Empty leaves
Google's default.

### spec.webhooks[].genericWebService

`GcpDialogflowCxAgentGenericWebService`

Call an HTTPS endpoint directly.

### spec.webhooks[].genericWebService.uri

`string`

The endpoint, https only.

- rule: {"string":{"pattern":"^https://.+"}}

### spec.webhooks[].genericWebService.webhookType

`string`

What Dialogflow sends and expects:
  STANDARD -- Dialogflow's WebhookRequest / WebhookResponse over POST
  FLEXIBLE -- any method and body you define (http_method,
              request_body), with parameter_mapping lifting response
              fields into session parameters
Empty leaves Google's default (STANDARD).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["STANDARD","FLEXIBLE"]}}

### spec.webhooks[].genericWebService.httpMethod

`string`

The HTTP method of a FLEXIBLE webhook; a STANDARD webhook always POSTs.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["POST","GET","HEAD","PUT","DELETE","PATCH","OPTIONS"]}}

### spec.webhooks[].genericWebService.requestBody

`string`

The JSON request body of a FLEXIBLE webhook; session parameters may be
referenced as $session.params.<name>.

### spec.webhooks[].genericWebService.parameterMapping

`map<string, string>`

FLEXIBLE webhooks: session parameter name -> field path in the webhook
response, e.g. {"order_status": "$.status"}.

### spec.webhooks[].genericWebService.requestHeaders

`map<string, string>`

Plain request headers sent with every call. Put secrets in
secret_versions_for_request_headers instead -- these values are stored
in the agent as written.

### spec.webhooks[].genericWebService.secretVersionsForRequestHeaders

`[]GcpDialogflowCxAgentSecretHeader`

Request headers whose values live in Secret Manager. A header named
here and in request_headers takes this value.

### spec.webhooks[].genericWebService.secretVersionsForRequestHeaders[].key

`string` · required

The header name, e.g. "X-Api-Key".

- rule: {"string":{"minLen":"1"}}

### spec.webhooks[].genericWebService.secretVersionsForRequestHeaders[].secretVersion

`string`

The Secret Manager secret version holding the header value:
projects/{project}/secrets/{secret}/versions/{version}.

- rule: {"string":{"pattern":"^projects/[^/]+/secrets/[^/]+/versions/[^/]+$"}}

### spec.webhooks[].genericWebService.secretVersionForUsernamePassword

`string`

The Secret Manager secret version holding "username:password" for HTTP
Basic authentication: projects/{project}/secrets/{secret}/versions/{version}.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^projects/[^/]+/secrets/[^/]+/versions/[^/]+$"}}

### spec.webhooks[].genericWebService.oauthConfig

`GcpDialogflowCxAgentOauthConfig`

Authenticate with the OAuth client-credentials flow.

### spec.webhooks[].genericWebService.oauthConfig.clientId

`string` · required

The client id the third party issued.

- rule: {"string":{"minLen":"1"}}

### spec.webhooks[].genericWebService.oauthConfig.tokenEndpoint

`string` · required

The endpoint Dialogflow exchanges the client credentials at for an
access token.

- rule: {"string":{"minLen":"1"}}

### spec.webhooks[].genericWebService.oauthConfig.clientSecret

`string` · sensitive

The client secret. Ignored when secret_version_for_client_secret is set
-- prefer that, so the secret stays in Secret Manager. Google never
returns it, so the declared value is always what the engines send.

### spec.webhooks[].genericWebService.oauthConfig.scopes

`[]string`

The OAuth scopes to request.

### spec.webhooks[].genericWebService.oauthConfig.secretVersionForClientSecret

`string`

The Secret Manager secret version holding the client secret:
projects/{project}/secrets/{secret}/versions/{version}. Wins over
client_secret.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^projects/[^/]+/secrets/[^/]+/versions/[^/]+$"}}

### spec.webhooks[].genericWebService.serviceAgentAuth

`string`

Have the Dialogflow service agent mint a token for the Authorization
header:
  NONE         -- no token
  ID_TOKEN     -- a Google ID token (for Cloud Run and Cloud Functions
                  behind IAM)
  ACCESS_TOKEN -- a Google OAuth access token (for Google APIs)
Empty leaves Google's default (NONE).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["NONE","ID_TOKEN","ACCESS_TOKEN"]}}

### spec.webhooks[].genericWebService.serviceAccount

`string | valueFrom`

Send an access token of this service account in the Authorization
header: a GcpServiceAccount reference or a literal email. The Dialogflow
service agent needs roles/iam.serviceAccountTokenCreator on it.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.webhooks[].genericWebService.allowedCaCerts

`[]string`

Custom CA certificates (DER, base64) trusted for this endpoint in place
of Google's default trust store. The server certificate must carry a
subject alternative name.

### spec.webhooks[].serviceDirectory

`GcpDialogflowCxAgentServiceDirectory`

Call an endpoint registered in Service Directory.

### spec.webhooks[].serviceDirectory.service

`string`

The service:
projects/{project}/locations/{location}/namespaces/{namespace}/services/{service}.

- rule: {"string":{"pattern":"^projects/[^/]+/locations/[^/]+/namespaces/[^/]+/services/[^/]+$"}}

### spec.webhooks[].serviceDirectory.genericWebService

`GcpDialogflowCxAgentGenericWebService`

How the endpoint behind the service is called.

### spec.webhooks[].serviceDirectory.genericWebService.uri

`string`

The endpoint, https only.

- rule: {"string":{"pattern":"^https://.+"}}

### spec.webhooks[].serviceDirectory.genericWebService.webhookType

`string`

What Dialogflow sends and expects:
  STANDARD -- Dialogflow's WebhookRequest / WebhookResponse over POST
  FLEXIBLE -- any method and body you define (http_method,
              request_body), with parameter_mapping lifting response
              fields into session parameters
Empty leaves Google's default (STANDARD).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["STANDARD","FLEXIBLE"]}}

### spec.webhooks[].serviceDirectory.genericWebService.httpMethod

`string`

The HTTP method of a FLEXIBLE webhook; a STANDARD webhook always POSTs.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["POST","GET","HEAD","PUT","DELETE","PATCH","OPTIONS"]}}

### spec.webhooks[].serviceDirectory.genericWebService.requestBody

`string`

The JSON request body of a FLEXIBLE webhook; session parameters may be
referenced as $session.params.<name>.

### spec.webhooks[].serviceDirectory.genericWebService.parameterMapping

`map<string, string>`

FLEXIBLE webhooks: session parameter name -> field path in the webhook
response, e.g. {"order_status": "$.status"}.

### spec.webhooks[].serviceDirectory.genericWebService.requestHeaders

`map<string, string>`

Plain request headers sent with every call. Put secrets in
secret_versions_for_request_headers instead -- these values are stored
in the agent as written.

### spec.webhooks[].serviceDirectory.genericWebService.secretVersionsForRequestHeaders

`[]GcpDialogflowCxAgentSecretHeader`

Request headers whose values live in Secret Manager. A header named
here and in request_headers takes this value.

### spec.webhooks[].serviceDirectory.genericWebService.secretVersionsForRequestHeaders[].key

`string` · required

The header name, e.g. "X-Api-Key".

- rule: {"string":{"minLen":"1"}}

### spec.webhooks[].serviceDirectory.genericWebService.secretVersionsForRequestHeaders[].secretVersion

`string`

The Secret Manager secret version holding the header value:
projects/{project}/secrets/{secret}/versions/{version}.

- rule: {"string":{"pattern":"^projects/[^/]+/secrets/[^/]+/versions/[^/]+$"}}

### spec.webhooks[].serviceDirectory.genericWebService.secretVersionForUsernamePassword

`string`

The Secret Manager secret version holding "username:password" for HTTP
Basic authentication: projects/{project}/secrets/{secret}/versions/{version}.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^projects/[^/]+/secrets/[^/]+/versions/[^/]+$"}}

### spec.webhooks[].serviceDirectory.genericWebService.oauthConfig

`GcpDialogflowCxAgentOauthConfig`

Authenticate with the OAuth client-credentials flow.

### spec.webhooks[].serviceDirectory.genericWebService.oauthConfig.clientId

`string` · required

The client id the third party issued.

- rule: {"string":{"minLen":"1"}}

### spec.webhooks[].serviceDirectory.genericWebService.oauthConfig.tokenEndpoint

`string` · required

The endpoint Dialogflow exchanges the client credentials at for an
access token.

- rule: {"string":{"minLen":"1"}}

### spec.webhooks[].serviceDirectory.genericWebService.oauthConfig.clientSecret

`string` · sensitive

The client secret. Ignored when secret_version_for_client_secret is set
-- prefer that, so the secret stays in Secret Manager. Google never
returns it, so the declared value is always what the engines send.

### spec.webhooks[].serviceDirectory.genericWebService.oauthConfig.scopes

`[]string`

The OAuth scopes to request.

### spec.webhooks[].serviceDirectory.genericWebService.oauthConfig.secretVersionForClientSecret

`string`

The Secret Manager secret version holding the client secret:
projects/{project}/secrets/{secret}/versions/{version}. Wins over
client_secret.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^projects/[^/]+/secrets/[^/]+/versions/[^/]+$"}}

### spec.webhooks[].serviceDirectory.genericWebService.serviceAgentAuth

`string`

Have the Dialogflow service agent mint a token for the Authorization
header:
  NONE         -- no token
  ID_TOKEN     -- a Google ID token (for Cloud Run and Cloud Functions
                  behind IAM)
  ACCESS_TOKEN -- a Google OAuth access token (for Google APIs)
Empty leaves Google's default (NONE).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["NONE","ID_TOKEN","ACCESS_TOKEN"]}}

### spec.webhooks[].serviceDirectory.genericWebService.serviceAccount

`string | valueFrom`

Send an access token of this service account in the Authorization
header: a GcpServiceAccount reference or a literal email. The Dialogflow
service agent needs roles/iam.serviceAccountTokenCreator on it.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.webhooks[].serviceDirectory.genericWebService.allowedCaCerts

`[]string`

Custom CA certificates (DER, base64) trusted for this endpoint in place
of Google's default trust store. The server certificate must carry a
subject alternative name.

### spec.tools

`[]GcpDialogflowCxAgentTool`

Tools, keyed by display name.

- rule: a tool is exactly one of open_api_spec, data_store_spec, function_spec
- rule: tool version display names must be unique within the tool -- the kind keys versions by them

### spec.tools[].displayName

`string` · required

The tool's human-readable name, unique within the agent.

- rule: {"string":{"minLen":"1"}}

### spec.tools[].description

`string` · required

What the tool does and when to use it -- the model reads this to decide
when to call the tool, so write it for the model.

- rule: {"string":{"minLen":"1"}}

### spec.tools[].openApiSpec

`GcpDialogflowCxAgentToolOpenApiSpec`

An OpenAPI tool.

### spec.tools[].openApiSpec.textSchema

`string` · required

The OpenAPI 3 schema as text (YAML or JSON).

- rule: {"string":{"minLen":"1"}}

### spec.tools[].openApiSpec.authentication

`GcpDialogflowCxAgentToolAuthentication`

How calls authenticate. Omit for an unauthenticated API.

- rule: set at most one of api_key_config, bearer_token_config, oauth_config, service_agent_auth_config -- Google accepts one authentication method

### spec.tools[].openApiSpec.authentication.apiKeyConfig

`GcpDialogflowCxAgentToolApiKeyConfig`

An API key in a header or query parameter.

### spec.tools[].openApiSpec.authentication.apiKeyConfig.keyName

`string` · required

The header or query parameter the key travels in, e.g. "X-Api-Key".

- rule: {"string":{"minLen":"1"}}

### spec.tools[].openApiSpec.authentication.apiKeyConfig.requestLocation

`string` · required

Where the key goes: "HEADER" or "QUERY_STRING" (Google's
RequestLocation values).

- rule: {"string":{"minLen":"1"}}

### spec.tools[].openApiSpec.authentication.apiKeyConfig.apiKey

`string` · sensitive

The key itself. Ignored when secret_version_for_api_key is set --
prefer that. Google never returns it.

### spec.tools[].openApiSpec.authentication.apiKeyConfig.secretVersionForApiKey

`string`

The Secret Manager secret version holding the key:
projects/{project}/secrets/{secret}/versions/{version}. Wins over
api_key.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^projects/[^/]+/secrets/[^/]+/versions/[^/]+$"}}

### spec.tools[].openApiSpec.authentication.bearerTokenConfig

`GcpDialogflowCxAgentToolBearerTokenConfig`

A bearer token.

### spec.tools[].openApiSpec.authentication.bearerTokenConfig.token

`string` · sensitive

The token. A session parameter reference such as
$session.params.user-token passes it per conversation. Ignored when
secret_version_for_token is set. Google never returns it.

### spec.tools[].openApiSpec.authentication.bearerTokenConfig.secretVersionForToken

`string`

The Secret Manager secret version holding the token:
projects/{project}/secrets/{secret}/versions/{version}. Wins over
token.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^projects/[^/]+/secrets/[^/]+/versions/[^/]+$"}}

### spec.tools[].openApiSpec.authentication.oauthConfig

`GcpDialogflowCxAgentToolOauthConfig`

OAuth.

### spec.tools[].openApiSpec.authentication.oauthConfig.clientId

`string` · required

The client id the OAuth provider issued.

- rule: {"string":{"minLen":"1"}}

### spec.tools[].openApiSpec.authentication.oauthConfig.oauthGrantType

`string` · required

The grant type: "CLIENT_CREDENTIAL" (Google's OauthGrantType values).

- rule: {"string":{"minLen":"1"}}

### spec.tools[].openApiSpec.authentication.oauthConfig.tokenEndpoint

`string` · required

The provider's token endpoint.

- rule: {"string":{"minLen":"1"}}

### spec.tools[].openApiSpec.authentication.oauthConfig.clientSecret

`string` · sensitive

The client secret. Ignored when secret_version_for_client_secret is
set -- prefer that. Google never returns it.

### spec.tools[].openApiSpec.authentication.oauthConfig.scopes

`[]string`

The OAuth scopes to request.

### spec.tools[].openApiSpec.authentication.oauthConfig.secretVersionForClientSecret

`string`

The Secret Manager secret version holding the client secret:
projects/{project}/secrets/{secret}/versions/{version}. Wins over
client_secret.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^projects/[^/]+/secrets/[^/]+/versions/[^/]+$"}}

### spec.tools[].openApiSpec.authentication.serviceAgentAuthConfig

`GcpDialogflowCxAgentToolServiceAgentAuthConfig`

A token the Dialogflow service agent mints.

### spec.tools[].openApiSpec.authentication.serviceAgentAuthConfig.serviceAgentAuth

`string`

The token type: "ID_TOKEN" or "ACCESS_TOKEN" (Google's ServiceAgentAuth
values). Empty leaves Google's default.

### spec.tools[].openApiSpec.serviceDirectoryConfig

`GcpDialogflowCxAgentToolServiceDirectoryConfig`

Reach the server through Service Directory.

### spec.tools[].openApiSpec.serviceDirectoryConfig.service

`string`

The service, in the agent's location:
projects/{project}/locations/{location}/namespaces/{namespace}/services/{service}.

- rule: {"string":{"pattern":"^projects/[^/]+/locations/[^/]+/namespaces/[^/]+/services/[^/]+$"}}

### spec.tools[].openApiSpec.tlsConfig

`GcpDialogflowCxAgentToolTlsConfig`

Trust custom CA certificates for the server.

### spec.tools[].openApiSpec.tlsConfig.caCerts

`[]GcpDialogflowCxAgentToolCaCert` · required

The trusted certificates.

- rule: {"repeated":{"minItems":"1"}}

### spec.tools[].openApiSpec.tlsConfig.caCerts[].displayName

`string` · required

A name that tells the certificates apart.

- rule: {"string":{"minLen":"1"}}

### spec.tools[].openApiSpec.tlsConfig.caCerts[].cert

`string` · required

The certificate, DER and base64-encoded. It replaces Google's default
trust store; the server certificate must carry a subject alternative
name.

- rule: {"string":{"minLen":"1"}}

### spec.tools[].dataStoreSpec

`GcpDialogflowCxAgentToolDataStoreSpec`

A data store tool.

### spec.tools[].dataStoreSpec.dataStoreConnections

`[]GcpDialogflowCxAgentToolDataStoreConnection` · required

The data stores searched.

- rule: {"repeated":{"minItems":"1"}}

### spec.tools[].dataStoreSpec.dataStoreConnections[].dataStore

`string | valueFrom`

The data store: a GcpVertexAiSearchDataStore reference or a literal
projects/{project}/locations/{location}/collections/{collection}/dataStores/{store}
(or the form without collections/).

- references: GcpVertexAiSearchDataStore (`status.outputs.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVertexAiSearchDataStore, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.tools[].dataStoreSpec.dataStoreConnections[].dataStoreType

`string`

The kind of store: "PUBLIC_WEB", "UNSTRUCTURED", or "STRUCTURED"
(Google's DataStoreType values).

### spec.tools[].dataStoreSpec.dataStoreConnections[].documentProcessingMode

`string`

For PUBLIC_WEB and UNSTRUCTURED stores: "DOCUMENTS" (Google's default)
or "CHUNKS" (Google's DocumentProcessingMode values).

### spec.tools[].functionSpec

`GcpDialogflowCxAgentToolFunctionSpec`

A client-executed function tool.

### spec.tools[].functionSpec.inputSchema

`string`

The JSON schema of the function's input, as a JSON object string.

### spec.tools[].functionSpec.outputSchema

`string`

The JSON schema of the function's output, as a JSON object string.

### spec.tools[].versions

`[]GcpDialogflowCxAgentToolVersion`

Frozen versions of this tool, keyed by display name.

### spec.tools[].versions[].displayName

`string` · required

The version's display name. Google does not require it unique; the
kind keys a tool's versions by it, so it must be unique within the
tool. A rename replaces the version.

- rule: {"string":{"minLen":"1"}}

### spec.tools[].versions[].tool

`GcpDialogflowCxAgentToolSnapshot` · required

The frozen definition.

- rule: {"required":true}
- rule: a tool is exactly one of open_api_spec, data_store_spec, function_spec

### spec.tools[].versions[].tool.displayName

`string` · required

The tool's display name at the time of the snapshot.

- rule: {"string":{"minLen":"1"}}

### spec.tools[].versions[].tool.description

`string` · required

What the tool does and when to use it -- the model reads this.

- rule: {"string":{"minLen":"1"}}

### spec.tools[].versions[].tool.openApiSpec

`GcpDialogflowCxAgentToolOpenApiSpec`

An OpenAPI tool.

### spec.tools[].versions[].tool.openApiSpec.textSchema

`string` · required

The OpenAPI 3 schema as text (YAML or JSON).

- rule: {"string":{"minLen":"1"}}

### spec.tools[].versions[].tool.openApiSpec.authentication

`GcpDialogflowCxAgentToolAuthentication`

How calls authenticate. Omit for an unauthenticated API.

- rule: set at most one of api_key_config, bearer_token_config, oauth_config, service_agent_auth_config -- Google accepts one authentication method

### spec.tools[].versions[].tool.openApiSpec.authentication.apiKeyConfig

`GcpDialogflowCxAgentToolApiKeyConfig`

An API key in a header or query parameter.

### spec.tools[].versions[].tool.openApiSpec.authentication.apiKeyConfig.keyName

`string` · required

The header or query parameter the key travels in, e.g. "X-Api-Key".

- rule: {"string":{"minLen":"1"}}

### spec.tools[].versions[].tool.openApiSpec.authentication.apiKeyConfig.requestLocation

`string` · required

Where the key goes: "HEADER" or "QUERY_STRING" (Google's
RequestLocation values).

- rule: {"string":{"minLen":"1"}}

### spec.tools[].versions[].tool.openApiSpec.authentication.apiKeyConfig.apiKey

`string` · sensitive

The key itself. Ignored when secret_version_for_api_key is set --
prefer that. Google never returns it.

### spec.tools[].versions[].tool.openApiSpec.authentication.apiKeyConfig.secretVersionForApiKey

`string`

The Secret Manager secret version holding the key:
projects/{project}/secrets/{secret}/versions/{version}. Wins over
api_key.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^projects/[^/]+/secrets/[^/]+/versions/[^/]+$"}}

### spec.tools[].versions[].tool.openApiSpec.authentication.bearerTokenConfig

`GcpDialogflowCxAgentToolBearerTokenConfig`

A bearer token.

### spec.tools[].versions[].tool.openApiSpec.authentication.bearerTokenConfig.token

`string` · sensitive

The token. A session parameter reference such as
$session.params.user-token passes it per conversation. Ignored when
secret_version_for_token is set. Google never returns it.

### spec.tools[].versions[].tool.openApiSpec.authentication.bearerTokenConfig.secretVersionForToken

`string`

The Secret Manager secret version holding the token:
projects/{project}/secrets/{secret}/versions/{version}. Wins over
token.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^projects/[^/]+/secrets/[^/]+/versions/[^/]+$"}}

### spec.tools[].versions[].tool.openApiSpec.authentication.oauthConfig

`GcpDialogflowCxAgentToolOauthConfig`

OAuth.

### spec.tools[].versions[].tool.openApiSpec.authentication.oauthConfig.clientId

`string` · required

The client id the OAuth provider issued.

- rule: {"string":{"minLen":"1"}}

### spec.tools[].versions[].tool.openApiSpec.authentication.oauthConfig.oauthGrantType

`string` · required

The grant type: "CLIENT_CREDENTIAL" (Google's OauthGrantType values).

- rule: {"string":{"minLen":"1"}}

### spec.tools[].versions[].tool.openApiSpec.authentication.oauthConfig.tokenEndpoint

`string` · required

The provider's token endpoint.

- rule: {"string":{"minLen":"1"}}

### spec.tools[].versions[].tool.openApiSpec.authentication.oauthConfig.clientSecret

`string` · sensitive

The client secret. Ignored when secret_version_for_client_secret is
set -- prefer that. Google never returns it.

### spec.tools[].versions[].tool.openApiSpec.authentication.oauthConfig.scopes

`[]string`

The OAuth scopes to request.

### spec.tools[].versions[].tool.openApiSpec.authentication.oauthConfig.secretVersionForClientSecret

`string`

The Secret Manager secret version holding the client secret:
projects/{project}/secrets/{secret}/versions/{version}. Wins over
client_secret.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^projects/[^/]+/secrets/[^/]+/versions/[^/]+$"}}

### spec.tools[].versions[].tool.openApiSpec.authentication.serviceAgentAuthConfig

`GcpDialogflowCxAgentToolServiceAgentAuthConfig`

A token the Dialogflow service agent mints.

### spec.tools[].versions[].tool.openApiSpec.authentication.serviceAgentAuthConfig.serviceAgentAuth

`string`

The token type: "ID_TOKEN" or "ACCESS_TOKEN" (Google's ServiceAgentAuth
values). Empty leaves Google's default.

### spec.tools[].versions[].tool.openApiSpec.serviceDirectoryConfig

`GcpDialogflowCxAgentToolServiceDirectoryConfig`

Reach the server through Service Directory.

### spec.tools[].versions[].tool.openApiSpec.serviceDirectoryConfig.service

`string`

The service, in the agent's location:
projects/{project}/locations/{location}/namespaces/{namespace}/services/{service}.

- rule: {"string":{"pattern":"^projects/[^/]+/locations/[^/]+/namespaces/[^/]+/services/[^/]+$"}}

### spec.tools[].versions[].tool.openApiSpec.tlsConfig

`GcpDialogflowCxAgentToolTlsConfig`

Trust custom CA certificates for the server.

### spec.tools[].versions[].tool.openApiSpec.tlsConfig.caCerts

`[]GcpDialogflowCxAgentToolCaCert` · required

The trusted certificates.

- rule: {"repeated":{"minItems":"1"}}

### spec.tools[].versions[].tool.openApiSpec.tlsConfig.caCerts[].displayName

`string` · required

A name that tells the certificates apart.

- rule: {"string":{"minLen":"1"}}

### spec.tools[].versions[].tool.openApiSpec.tlsConfig.caCerts[].cert

`string` · required

The certificate, DER and base64-encoded. It replaces Google's default
trust store; the server certificate must carry a subject alternative
name.

- rule: {"string":{"minLen":"1"}}

### spec.tools[].versions[].tool.dataStoreSpec

`GcpDialogflowCxAgentToolDataStoreSpec`

A data store tool.

### spec.tools[].versions[].tool.dataStoreSpec.dataStoreConnections

`[]GcpDialogflowCxAgentToolDataStoreConnection` · required

The data stores searched.

- rule: {"repeated":{"minItems":"1"}}

### spec.tools[].versions[].tool.dataStoreSpec.dataStoreConnections[].dataStore

`string | valueFrom`

The data store: a GcpVertexAiSearchDataStore reference or a literal
projects/{project}/locations/{location}/collections/{collection}/dataStores/{store}
(or the form without collections/).

- references: GcpVertexAiSearchDataStore (`status.outputs.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVertexAiSearchDataStore, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.tools[].versions[].tool.dataStoreSpec.dataStoreConnections[].dataStoreType

`string`

The kind of store: "PUBLIC_WEB", "UNSTRUCTURED", or "STRUCTURED"
(Google's DataStoreType values).

### spec.tools[].versions[].tool.dataStoreSpec.dataStoreConnections[].documentProcessingMode

`string`

For PUBLIC_WEB and UNSTRUCTURED stores: "DOCUMENTS" (Google's default)
or "CHUNKS" (Google's DocumentProcessingMode values).

### spec.tools[].versions[].tool.functionSpec

`GcpDialogflowCxAgentToolFunctionSpec`

A client-executed function tool.

### spec.tools[].versions[].tool.functionSpec.inputSchema

`string`

The JSON schema of the function's input, as a JSON object string.

### spec.tools[].versions[].tool.functionSpec.outputSchema

`string`

The JSON schema of the function's output, as a JSON object string.

### spec.versions

`[]GcpDialogflowCxAgentVersion`

Flow versions, keyed by flow and display name.

### spec.versions[].flowId

`string`

The id of the flow to snapshot, the last segment of its resource name.
Empty is the agent's start flow (00000000-0000-0000-0000-000000000000),
the one flow every agent has; other flows are authored in the console
and named by id here, because a full flow path contains the agent's
id, which does not exist before the first apply.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[A-Za-z0-9-]+$"}}

### spec.versions[].displayName

`string` · required

The version's name, at most 64 characters. Google does not require it
unique; the kind keys a flow's versions by it, so it must be unique per
flow. A rename replaces the version.

- rule: {"string":{"minLen":"1","maxLen":"64"}}

### spec.versions[].description

`string`

What changed in this version, at most 500 characters.

- rule: {"string":{"maxLen":"500"}}

### spec.environments

`[]GcpDialogflowCxAgentEnvironment`

Environments, keyed by display name.

### spec.environments[].displayName

`string` · required

The environment's name, unique within the agent, at most 64
characters.

- rule: {"string":{"minLen":"1","maxLen":"64"}}

### spec.environments[].description

`string`

What the environment is for, at most 500 characters.

- rule: {"string":{"maxLen":"500"}}

### spec.environments[].versionConfigs

`[]GcpDialogflowCxAgentEnvironmentVersionConfig` · required

The version pinned for each flow. Google requires a version for every
flow reachable from the start flow; a missing one fails the apply.

- rule: {"repeated":{"minItems":"1"}}
- rule: a version config names exactly one of version (a declared version's display name) or version_id

### spec.environments[].versionConfigs[].flowId

`string`

The flow, by id. Empty is the agent's start flow.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[A-Za-z0-9-]+$"}}

### spec.environments[].versionConfigs[].version

`string`

The display name of a version declared in spec.versions for this flow.

### spec.environments[].versionConfigs[].versionId

`string`

The numeric id of a version made outside this spec (in the console or
by another tool), the last segment of its resource name.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+$"}}

### spec.generativeSettings

`[]GcpDialogflowCxAgentGenerativeSettings`

Generative settings, one per language.

### spec.generativeSettings[].languageCode

`string` · required

The language these settings apply to -- the agent's default language
or one of its supported languages. Keyed by it: changing it declares
settings for another language and leaves the old language's settings
in place.

- rule: {"string":{"minLen":"1"}}

### spec.generativeSettings[].fallbackSettings

`GcpDialogflowCxAgentFallbackSettings`

Generative fallback.

### spec.generativeSettings[].fallbackSettings.selectedPrompt

`string`

The display name of the prompt template in use.

### spec.generativeSettings[].fallbackSettings.promptTemplates

`[]GcpDialogflowCxAgentPromptTemplate`

The stored prompt templates.

### spec.generativeSettings[].fallbackSettings.promptTemplates[].displayName

`string`

The prompt's name, e.g. "conservative" or "chatty".

### spec.generativeSettings[].fallbackSettings.promptTemplates[].frozen

`bool`

Freeze the prompt against edits in the console.

### spec.generativeSettings[].fallbackSettings.promptTemplates[].promptText

`string`

The prompt sent to the model on a no-match, with placeholders Google
fills, e.g. "Here is a conversation $conversation, a response is: ".

### spec.generativeSettings[].generativeSafetySettings

`GcpDialogflowCxAgentGenerativeSafetySettings`

Filters on generated text.

### spec.generativeSettings[].generativeSafetySettings.defaultBannedPhraseMatchStrategy

`string`

How banned phrases match: "PARTIAL_MATCH" or "WORD_MATCH" (Google's
PhraseMatchStrategy values). Empty leaves Google's default.

### spec.generativeSettings[].generativeSafetySettings.bannedPhrases

`[]GcpDialogflowCxAgentBannedPhrase`

Phrases generated text must never contain.

### spec.generativeSettings[].generativeSafetySettings.bannedPhrases[].languageCode

`string` · required

The phrase's language code, e.g. "en".

- rule: {"string":{"minLen":"1"}}

### spec.generativeSettings[].generativeSafetySettings.bannedPhrases[].text

`string` · required

The phrase.

- rule: {"string":{"minLen":"1"}}

### spec.generativeSettings[].knowledgeConnectorSettings

`GcpDialogflowCxAgentKnowledgeConnectorSettings`

The knowledge connector's persona and scope.

### spec.generativeSettings[].knowledgeConnectorSettings.agent

`string`

The virtual agent's name, used in the prompt. May be empty.

### spec.generativeSettings[].knowledgeConnectorSettings.agentIdentity

`string`

What the agent is, e.g. "virtual agent" or "AI assistant".

### spec.generativeSettings[].knowledgeConnectorSettings.agentScope

`string`

Where the agent operates, e.g. "Example company website".

### spec.generativeSettings[].knowledgeConnectorSettings.business

`string`

The company or organization the agent represents -- used in the prompt
and in knowledge search.

### spec.generativeSettings[].knowledgeConnectorSettings.businessDescription

`string`

A description of the business, e.g. "a family company selling freshly
roasted coffee beans".

### spec.generativeSettings[].knowledgeConnectorSettings.disableDataStoreFallback

`bool`

Stop falling back to raw data store search results when the model
cannot pick an answer (the fallback is on by default).

### spec.generativeSettings[].llmModelSettings

`GcpDialogflowCxAgentLlmModelSettings`

The model and prompt.

### spec.generativeSettings[].llmModelSettings.model

`string`

The model, by the id Dialogflow lists for generative features (a
Gemini model). Empty leaves Google's default.

### spec.generativeSettings[].llmModelSettings.promptText

`string`

A custom prompt for the model.

### spec.deletionPolicy

`string`

What happens to the agent and everything folded into it when this
resource is destroyed:
  "" / "DELETE" -- deleted, conversation content included
  "PREVENT"     -- destroy fails
  "ABANDON"     -- everything leaves management and stays in GCP
Generative settings are never deleted in Google either way.

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `spec.supported_languages_exclude_default`: supported_language_codes lists the agent's languages other than default_language_code -- Google's rule
- `spec.delete_chat_engine_only_for_a_literal_engine`: delete_chat_engine_on_destroy cannot delete an engine referenced by kind -- that GcpVertexAiSearchEngine owns its own lifecycle; name the engine literally or leave the flag off
- `spec.unique_webhook_display_names`: webhook display names must be unique within the agent -- Google's rule
- `spec.unique_tool_display_names`: tool display names must be unique within the agent -- Google's rule
- `spec.unique_environment_display_names`: environment display names must be unique within the agent -- Google's rule
- `spec.unique_version_display_names_per_flow`: version display names must be unique per flow -- the kind keys versions by them
- `spec.unique_generative_settings_languages`: generative settings are declared once per language
- `spec.environment_versions_declared`: every version an environment names must be declared in versions for the same flow

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpDialogflowCxAgent, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name -- what a GcpVertexAiSearchEngine's dialogflow_agent_to_link takes: projects/{project}/locations/{location}/agents/{agent_id}. |
| `status.outputs.agent_id` | `string` | The id Google assigned at creation (the last segment of name). |
| `status.outputs.location` | `string` | The location the agent lives in. |
| `status.outputs.start_flow` | `string` | Full resource name of the agent's default start flow. |
| `status.outputs.webhook_names` | `[]string` | Full resource names of the declared webhooks, in manifest order. |
| `status.outputs.tool_names` | `[]string` | Full resource names of the declared tools, in manifest order. |
| `status.outputs.tool_version_names` | `[]string` | Full resource names of the declared tool versions, tool by tool in manifest order. |
| `status.outputs.version_names` | `[]string` | Full resource names of the declared flow versions, in manifest order. |
| `status.outputs.environment_names` | `[]string` | Full resource names of the declared environments, in manifest order. |
| `status.outputs.generative_settings_names` | `[]string` | Resource names of the declared generative settings, in manifest order ({agent}/generativeSettings?languageCode={language}). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.securitySettings` | GcpDialogflowCxSecuritySettings | `status.outputs.name` |
| `spec.genAppBuilderSettings.engine` | GcpVertexAiSearchEngine | `status.outputs.name` |
| `spec.webhooks[].genericWebService.serviceAccount` | GcpServiceAccount | `status.outputs.email` |
| `spec.webhooks[].serviceDirectory.genericWebService.serviceAccount` | GcpServiceAccount | `status.outputs.email` |
| `spec.tools[].dataStoreSpec.dataStoreConnections[].dataStore` | GcpVertexAiSearchDataStore | `status.outputs.name` |
| `spec.tools[].versions[].tool.dataStoreSpec.dataStoreConnections[].dataStore` | GcpVertexAiSearchDataStore | `status.outputs.name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpVertexAiSearchEngine | `spec.chatEngineConfig.dialogflowAgentToLink` | `status.outputs.name` |

## See Also

- [Overview](../README.md)

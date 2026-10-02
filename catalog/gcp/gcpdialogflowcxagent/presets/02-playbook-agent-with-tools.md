# Playbook Agent with Tools

## Use Case

A generative agent: conversations begin in the default playbook, whose natural-language instructions (written in the console) decide when to call an OpenAPI tool for live order status and a data store tool that answers from a policy handbook indexed in Vertex AI Search.

## When to Use

- Open-ended assistance where scripting every path is not practical
- Answers that must come from your own documents, not the model's memory
- Actions against your APIs, authenticated without a stored secret

## What This Creates

- A global agent that starts in its default playbook
- An OpenAPI tool calling your orders API with an ID token the Dialogflow service agent mints
- A data store tool over a `GcpVertexAiSearchDataStore`
- English generative settings with the knowledge connector's persona and a banned phrase

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `tools[].openApiSpec.authentication` | service-agent ID token | `apiKeyConfig`, `bearerTokenConfig`, or `oauthConfig` for APIs outside Google Cloud; keep keys in Secret Manager (`secretVersionFor...`). |
| `tools[].dataStoreSpec` | one policy store | Add stores; set `documentProcessingMode: CHUNKS` for long unstructured documents. |
| `generativeSettings` | English persona and safety | Declare one entry per language the agent serves. |

Every playbook turn bills at the Playbooks rate -- see the component's cost profile.

# Chat Agent and Assistant

## Use Case

A support chat app: a CHAT engine that has Google create its Dialogflow CX agent over a document store, plus a Gemini Enterprise assistant with a content policy (banned phrases), Model Armor sanitization of prompts and responses that fails closed, Google Search grounding, and a system instruction that keeps answers on Acme's documents.

## When to Use

- A conversational support or help-desk experience over your own documents
- When the answers need guardrails: banned topics, prompt and response sanitization
- When Gemini Enterprise seats are licensed on the project

## What This Creates

- A CHAT engine in `global` over the referenced `CONTENT_REQUIRED` store (enrolled in `SOLUTION_TYPE_CHAT`), which creates a Dialogflow CX agent for Acme in English, New York time
- One assistant, `default_assistant`, with a banned phrase, two Model Armor templates in `FAIL_CLOSED` mode, Google Search grounding, and an additional system instruction

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `chatEngineConfig` | create an agent | `dialogflowAgentToLink` (with `allowCrossRegion` if the agent lives elsewhere) to reuse an existing Dialogflow CX agent instead. |
| `assistants[].customerPolicy.modelArmorConfig` | two templates | Your Model Armor templates (`projects/{project}/locations/{location}/templates/{id}`); `FAIL_OPEN` to answer even when sanitization fails. |
| `assistants[].webGroundingType` | Google Search | `WEB_GROUNDING_TYPE_ENTERPRISE_WEB_SEARCH` for enterprise web search, `WEB_GROUNDING_TYPE_DISABLED` to ground on your documents alone. |
| `assistants` | one | Remove the block when the project has no Gemini Enterprise license. |

The chat engine config is immutable: a change replaces the engine (and the agent it created). Data stores for a chat engine must be enrolled in `SOLUTION_TYPE_CHAT`; the module enables the Dialogflow API. The Dialogflow CX agent bills Dialogflow's per-request rates.

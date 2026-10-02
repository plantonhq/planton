# Chat Over a Managed Agent

## Use Case

A chat engine that answers from your data stores through a Dialogflow CX agent your team already runs as its own block, with its own webhooks, tools, flow versions, and environments.

## When to Use

- The conversation design lives in a `GcpDialogflowCxAgent` you manage, and search is one capability it gains
- Several teams own the agent and the search data separately
- You want the agent to outlive any one chat engine

## What This Creates

- A global chat engine over one chat-solution data store, linked to the `support-agent` Dialogflow CX agent by reference

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `chatEngineConfig.dialogflowAgentToLink` | the `support-agent` block | Point at your agent, or write its full resource name for an agent managed elsewhere. |
| `chatEngineConfig.allowCrossRegion` | off | Set it when the agent lives in a different location from the engine. |
| `dataStoreIds` | one chat store | Every store must be enrolled in `SOLUTION_TYPE_CHAT` and live in the engine's location. |

Link from one side: leave the agent's own `genAppBuilderSettings` unset.

# Container Agent with Memory Bank

## Use Case

A production agent you build and ship as a container, running as its own service account, reading a partner API key from Secret Manager, and remembering each user across sessions through Agent Engine's Memory Bank: memories generated every five events by Gemini, looked up by embedding similarity, expiring after ninety days, organized under Google's managed topics and one custom travel topic.

## When to Use

- Agents built outside the ADK Python build (any language, any framework) delivered as an image
- Assistants that must recall a user's preferences and instructions between conversations
- Production deployments that need a dedicated identity and secrets from Secret Manager

## What This Creates

- An Agent Engine instance in `us-central1` running the `concierge:1.4.0` image on port 8080 as the referenced `GcpServiceAccount`
- One to twenty instances at 4 CPU / 8 GiB, concurrency 9, with `PARTNER_API_KEY` from the referenced `GcpSecretManagerSecret`
- A Memory Bank on `gemini-2.5-flash` and `text-embedding-005`, scoped by `user_id`, with two managed topics and one custom topic, third-person memories, and a 90-day TTL
- `deletionPolicy: PREVENT`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `spec.containerSpec.imageUri` | `concierge:1.4.0` | Your image; it must implement the Agent Engine serving contract. |
| `spec.serviceAccount` | `concierge-agent-sa` | The identity that needs `roles/secretmanager.secretAccessor` on the secret and access to your APIs. |
| `spec.deploymentSpec.resourceLimits` | 4 CPU / 8 GiB | `cpu` in 1, 2, 4, 6, 8 and `memory` up to 32Gi. |
| `contextSpec.memoryBankConfig.ttlConfig.defaultTtl` | `7776000s` (90 days) | How long memories live; `granularTtlConfig` sets lifetimes by origin. |
| `contextSpec.memoryBankConfig.customizationConfigs[].memoryTopics` | two managed + `travel` | The topics memories are organized under. |
| `deletionPolicy` | `PREVENT` | `DELETE` for a disposable environment (memories go with the agent). |

The memory bank updates in place; only the location and the encryption key are immutable. Memory generation bills the generation model's tokens and the embedding calls.

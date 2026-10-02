# Prompt Guard

## Use Case

The first safety screen for any AI application: detect prompt-injection and jailbreak attempts and links to malicious sites, and report the verdicts without blocking while you learn what your traffic looks like.

## When to Use

- A new chatbot, agent, or RAG app before it meets real users
- Measuring how often your traffic trips the filters before turning blocking on
- The template a Vertex AI Search assistant screens prompts through

## What This Creates

- A template in `us-central1` running prompt-injection and jailbreak detection at `MEDIUM_AND_ABOVE` plus malicious URL detection, in inspect-only mode with sanitize logging on

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `us-central1` | Where your traffic may be inspected (`us`, `eu`, or another region). |
| `templateMetadata.enforcementType` | `INSPECT_ONLY` | `INSPECT_AND_BLOCK` once the false-positive rate is acceptable. |
| `filterConfig.piAndJailbreakFilterSettings.confidenceLevel` | `MEDIUM_AND_ABOVE` | `LOW_AND_ABOVE` to catch more, `HIGH` to flag only clear attempts. |
| `templateMetadata.logSanitizeOperations` | `true` | Off once tuning is done; the logs carry the screened text. |

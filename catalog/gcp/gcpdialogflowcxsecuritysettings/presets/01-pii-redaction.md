# PII Redaction

## Use Case

The privacy baseline for agents that talk to customers: Sensitive Data Protection scrubs personal data from everything Dialogflow stores, and conversation history is purged after 30 days.

## When to Use

- Any agent whose users may type or say names, emails, card numbers, or addresses
- Before turning on Cloud Logging or interaction history on an agent
- Shared by every agent of a team in one location

## What This Creates

- Security settings in `us-central1` that redact with Sensitive Data Protection's default detectors on disk storage and keep conversation history for 30 days

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `us-central1` | Must match the agents that use the settings. Fixed at creation. |
| `inspectTemplate` / `deidentifyTemplate` | Google's defaults | Your own Sensitive Data Protection templates (same region) to choose the detectors and how values are replaced. |
| `retentionWindowDays` | `30` | Shorter keeps less; `retentionStrategy: REMOVE_AFTER_CONVERSATION` keeps nothing past the conversation (and turns off audio and Insights export). |

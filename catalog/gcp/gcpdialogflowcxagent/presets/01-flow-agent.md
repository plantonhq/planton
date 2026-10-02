# Flow Agent

## Use Case

A conversational agent whose conversations are designed as flows and pages in the Dialogflow console: deterministic routing, intents, and forms, with spell correction and interaction history on.

## When to Use

- The first agent in a project
- Support or booking conversations that follow a known path
- Teams that design in the console and want the agent itself managed as code

## What This Creates

- A global Dialogflow CX agent in English on New York time, with its default start flow, spell correction, and interaction logging

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `global` | A region such as `us-central1` for data residency (a project's first regional agent needs Google's one-time location settings in the console). Fixed at creation. |
| `defaultLanguageCode` | `en` | The agent's main language; fixed at creation. Add others in `supportedLanguageCodes`. |
| `timeZone` | `America/New_York` | The time zone dates and times in conversations resolve against. |
| `advancedSettings.loggingSettings` | interaction logging | Add `enableStackdriverLogging` to send conversations to Cloud Logging -- attach security settings with redaction first. |

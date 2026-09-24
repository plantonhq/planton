# GcpDialogflowCxSecuritySettings Guide

The judgment this guide protects: security settings are the privacy contract of every conversation an agent holds. Decide them before an agent logs a single conversation, share one set across the agents that serve the same users, and remember that retention and redaction change what Google keeps, not what the agent can do.

## One set, many agents, one location

Settings belong to a project and a location, and an agent can only use settings in its own location -- a `global` agent needs `global` settings. Because an agent points at settings by name, one set can govern every agent a team runs in that location, and changing it changes all of them at once. That is why the settings are their own block instead of part of an agent.

## Redaction

`redactionStrategy: REDACT_WITH_SERVICE` has Sensitive Data Protection scrub what Dialogflow persists; `redactionScope: REDACT_DISK_STORAGE` says that means everything written to durable storage. Both are needed -- a strategy without a scope redacts nothing. Without templates, Google's default detectors run and matches become `[redacted]`. An inspect template chooses the detectors (card numbers, custom dictionaries, likelihood thresholds), and a de-identify template chooses the replacement (masking, tokenization). Both must live in the settings' region and are named by their full resource names.

## Retention

Without a rule, Dialogflow keeps conversation data for its default TTL (365 days; 30 for Agent Assist traffic). `retentionWindowDays` shortens it -- a value above the default is ignored. `retentionStrategy: REMOVE_AFTER_CONVERSATION` keeps nothing past the conversation, and also switches off audio and Insights export regardless of those fields. Google accepts one of the two; the spec refuses both. `purgeDataTypes: [DIALOGFLOW_HISTORY]` says the history is what the rule removes.

## Audio and Insights export

Audio export records telephony calls into your Cloud Storage bucket. Setting the bucket makes Google grant the Dialogflow service agent object-creator access on it, so the principal that applies needs `storage.buckets.setIamPolicy` there; the recordings then follow the bucket's own retention and lifecycle rules. `enableAudioRedaction` redacts sensitive speech in the recording too. `enableInsightsExport` sends each finished conversation to Conversational Insights, which bills its analysis separately.

## Destroy

`deletionPolicy` defaults to `DELETE`. An agent that references the settings depends on them, so in one environment the agent is destroyed first. `PREVENT` makes destroy fail -- the right choice for settings a production agent relies on.

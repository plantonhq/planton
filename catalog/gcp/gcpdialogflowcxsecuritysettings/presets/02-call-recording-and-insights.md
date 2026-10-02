# Call Recording and Insights

## Use Case

A voice contact center's policy: telephony audio is recorded into your own bucket with sensitive speech redacted, finished conversations flow into Conversational Insights for analysis, and history is kept for 90 days.

## When to Use

- Voice agents where calls must be reviewable
- Quality programs that analyze conversations in Conversational Insights
- Settings a destroy must never remove by accident

## What This Creates

- Security settings in `us-central1` with redaction, 90-day retention, MP3 audio export into a `GcpGcsBucket` (Google grants the Dialogflow service agent object-creator access on it), and Insights export
- `PREVENT` on destroy

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `audioExportSettings.gcsBucket` | the `call-recordings` bucket | Your bucket; give it the retention and lifecycle rules the recordings need. Applying needs `storage.buckets.setIamPolicy` on it. |
| `audioExportSettings.audioFormat` | `MP3` | `MULAW` keeps the telephone-native format; `OGG` for Vorbis. |
| `enableInsightsExport` | on | Off when conversations must not leave Dialogflow. |

# GCP Dialogflow CX Security Settings

Sets the privacy policy your conversational agents run under: personal data scrubbed from everything Dialogflow stores, conversation history kept only as long as you say, telephony calls recorded into your own bucket, and finished conversations sent to Conversational Insights for analysis. Declare it once per location and attach it to every agent that should follow it.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `dialogflow.googleapis.com` on the project
- **Security settings** -- a `diagflow.CxSecuritySettings` with redaction, retention, audio export, and Insights export

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Dialogflow admin permissions on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **GcpGcsBucket** -- for audio export, referenced by `audioExportSettings.gcsBucket`.

## Deploy

### Console

Open the deployment store, find **GCP Dialogflow CX Security Settings**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **PII Redaction** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDialogflowCxSecuritySettings
metadata:
  name: pii-redaction
  org: acme-corp
  env: prod
spec:
  location: global
  redactionStrategy: REDACT_WITH_SERVICE
  redactionScope: REDACT_DISK_STORAGE
  retentionWindowDays: 30
```

```shell
planton apply -f dialogflow-cx-security-settings.yaml
```

This creates settings in the global location that redact personal data and keep conversation history for 30 days. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, reference these settings from each `GcpDialogflowCxAgent`'s `securitySettings`, and reference a `GcpGcsBucket` from `audioExportSettings.gcsBucket` when calls are recorded.

## Key Configuration

These are the most important decisions when configuring security settings. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Redaction** -- `redactionStrategy: REDACT_WITH_SERVICE` with `redactionScope: REDACT_DISK_STORAGE` scrubs personal data before anything is stored. Your own Sensitive Data Protection templates decide what counts as sensitive and how it is replaced.

**Retention** -- `retentionWindowDays` keeps data for a set number of days; `retentionStrategy: REMOVE_AFTER_CONVERSATION` keeps nothing past the conversation and switches off audio and Insights export.

**Exports** -- audio export records telephony calls into your bucket; Insights export feeds Conversational Insights. Both follow the retention rule.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpGcsBucket** | `audioExportSettings.gcsBucket` | `status.outputs.bucket_name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The settings' full resource name | An agent's `securitySettings` |
| `security_settings_id` | The id Google assigned | Console links |
| `location` | The settings' location | Matching agents |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**PII redaction** -- Redaction with Google's default detectors and 30-day retention. Start from the **PII Redaction** preset.

**Call recording and insights** -- Redacted MP3 recordings into your bucket, Insights export, 90-day retention, protected by `PREVENT`. Start from the **Call Recording and Insights** preset.

## Works With

- [**GCP Dialogflow CX Agent**](/cloud-catalog/gcp-dialogflow-cx-agent) -- the agents that apply these settings
- [**GCP GCS Bucket**](/cloud-catalog/gcp-gcs-bucket) -- where exported audio lands

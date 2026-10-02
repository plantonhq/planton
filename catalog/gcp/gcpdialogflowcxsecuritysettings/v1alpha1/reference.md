# GcpDialogflowCxSecuritySettings

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpDialogflowCxSecuritySettingsSpec defines Dialogflow CX security
settings (`google_dialogflow_cx_security_settings`) -- the redaction,
retention, audio-export, and Insights-export policy that one or more
Dialogflow CX agents in the same project and location apply to every
conversation. A GcpDialogflowCxAgent attaches by referencing the `name`
output from its security_settings field; several agents may share one
set, which is why the settings are a block of their own rather than part
of an agent.

What it governs:
  - redaction: PII found by Sensitive Data Protection (DLP) is scrubbed
    from everything Dialogflow persists, optionally shaped by your own
    inspect and de-identify templates
  - retention: how long conversation history is kept (a window in days,
    or removed when the conversation ends)
  - audio export: telephony audio recorded into your Cloud Storage bucket
  - Insights export: finished conversations sent to Conversational
    Insights for its analyzers

Immutable: project_id and location (a change replaces the settings).
Everything else updates in place.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDialogflowCxSecuritySettings
metadata:
  name: pii-redaction
spec:
  projectId:
    value: my-gcp-project
  location: global
  displayName: PII redaction
  # Scrub personal data with Sensitive Data Protection's default detectors
  # from everything Dialogflow writes to disk.
  redactionStrategy: REDACT_WITH_SERVICE
  redactionScope: REDACT_DISK_STORAGE
  # Keep conversation history for 30 days, then purge it.
  retentionWindowDays: 30
  purgeDataTypes:
    - DIALOGFLOW_HISTORY
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.redactionStrategy` | `string` |  |  |  |
| `spec.redactionScope` | `string` |  |  |  |
| `spec.inspectTemplate` | `string` |  |  |  |
| `spec.deidentifyTemplate` | `string` |  |  |  |
| `spec.purgeDataTypes` | `[]string` |  |  |  |
| `spec.retentionStrategy` | `string` |  |  |  |
| `spec.retentionWindowDays` | `int32` |  |  |  |
| `spec.audioExportSettings` | `GcpDialogflowCxSecuritySettingsAudioExport` |  |  |  |
| `spec.audioExportSettings.gcsBucket` | `string \| valueFrom` |  |  | GcpGcsBucket (`status.outputs.bucket_name`) |
| `spec.audioExportSettings.audioExportPattern` | `string` |  |  |  |
| `spec.audioExportSettings.audioFormat` | `string` |  |  |  |
| `spec.audioExportSettings.enableAudioRedaction` | `bool` |  |  |  |
| `spec.enableInsightsExport` | `bool` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the settings live in: a literal project ID or a
GcpProject reference. If omitted, the provider's default project is
used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The Dialogflow CX location, e.g. "global" or "us-central1". The
settings apply only to agents in this same location, and the inspect
and de-identify templates must live in this region too. Immutable.

- rule: {"required":true,"string":{"pattern":"^(global|[a-z]+(-[a-z]+[0-9]+)?)$"}}

### spec.displayName

`string`

Human-readable name, unique within the location. Defaults to
metadata.name. Mutable in place.

### spec.redactionStrategy

`string`

How Dialogflow redacts sensitive data before it persists anything:
  REDACT_WITH_SERVICE -- call Sensitive Data Protection to scrub the
                         data (the inspect and de-identify templates
                         below shape what is found and how it is
                         replaced)
Empty means no redaction.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["REDACT_WITH_SERVICE"]}}

### spec.redactionScope

`string`

Which data redaction applies to:
  REDACT_DISK_STORAGE -- everything written to disk or other durable
                         storage, temporary files included
Empty means nothing is redacted, even with a redaction strategy set.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["REDACT_DISK_STORAGE"]}}

### spec.inspectTemplate

`string`

A Sensitive Data Protection inspect template deciding WHAT counts as
sensitive (info types, likelihood, custom detectors):
projects/{project}/locations/{location}/inspectTemplates/{template} or
organizations/{org}/locations/{location}/inspectTemplates/{template},
in the same region as these settings. Empty uses Google's default
inspect configuration.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(projects|organizations)/[^/]+/locations/[^/]+/inspectTemplates/[^/]+$"}}

### spec.deidentifyTemplate

`string`

A Sensitive Data Protection de-identify template deciding HOW found
data is replaced (masking, tokenization, bucketing):
projects/{project}/locations/{location}/deidentifyTemplates/{template}
or the organizations/ form, in the same region. Empty replaces
sensitive values with "[redacted]".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(projects|organizations)/[^/]+/locations/[^/]+/deidentifyTemplates/[^/]+$"}}

### spec.purgeDataTypes

`[]string`

What the retention rule purges when it fires. DIALOGFLOW_HISTORY
(conversation history) is the only data type Google defines.

- rule: {"repeated":{"items":{"string":{"in":["DIALOGFLOW_HISTORY"]}}}}

### spec.retentionStrategy

`string`

Retain sensitive conversation data only while the conversation lasts:
  REMOVE_AFTER_CONVERSATION -- removed when the conversation ends (a
                               session without an explicit
                               conversation ends with the session)
This also turns off audio export and Insights export. Set this or
retention_window_days, never both.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["REMOVE_AFTER_CONVERSATION"]}}

### spec.retentionWindowDays

`int32`

Keep sensitive conversation data for this many days. Only values below
Dialogflow's default TTL (365 days; 30 for Agent Assist traffic) take
effect -- a larger value is ignored and the default applies, as does 0
or unset. Set this or retention_strategy, never both.

- rule: {"int32":{"gte":0}}

### spec.audioExportSettings

`GcpDialogflowCxSecuritySettingsAudioExport`

Record telephony audio into a Cloud Storage bucket.

### spec.audioExportSettings.gcsBucket

`string | valueFrom`

The bucket the audio lands in: a GcpGcsBucket reference or a literal
bucket name. Setting it makes Google grant the Dialogflow service agent
roles/storage.objectCreator on the bucket, so whoever applies this
needs storage.buckets.setIamPolicy there. Objects then follow the
bucket's own retention and lifecycle rules.

- references: GcpGcsBucket (`status.outputs.bucket_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_name}} -- a bare string does not parse

### spec.audioExportSettings.audioExportPattern

`string`

The object-name pattern for exported audio files. Empty uses Google's
default naming.

### spec.audioExportSettings.audioFormat

`string`

The file format of exported audio -- telephony recordings only today:
  MULAW -- G.711 mu-law PCM at 8 kHz (the telephone-native format)
  MP3   -- MP3
  OGG   -- OGG Vorbis
Empty leaves Google's default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["MULAW","MP3","OGG"]}}

### spec.audioExportSettings.enableAudioRedaction

`bool`

Redact sensitive spoken content in the exported audio as well as in the
transcripts.

### spec.enableInsightsExport

`bool`

Send each finished conversation to Conversational Insights and run its
analyzers. Ignored under REMOVE_AFTER_CONVERSATION.

### spec.deletionPolicy

`string`

What happens to the settings when this resource is destroyed:
  "" / "DELETE" -- deleted (an agent that references the settings is
                   destroyed first, because it depends on them)
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the settings leave management and stay in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `spec.one_retention_rule`: set at most one of retention_strategy and retention_window_days -- Google accepts only one retention rule

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpDialogflowCxSecuritySettings, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name -- what an agent's security_settings field takes: projects/{project}/locations/{location}/securitySettings/{id}. |
| `status.outputs.security_settings_id` | `string` | The id Google assigned at creation (the last segment of name). |
| `status.outputs.location` | `string` | The location the settings live in; only agents here can use them. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.audioExportSettings.gcsBucket` | GcpGcsBucket | `status.outputs.bucket_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpDialogflowCxAgent | `spec.securitySettings` | `status.outputs.name` |

## See Also

- [Overview](../README.md)

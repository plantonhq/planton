# GCP Dialogflow CX Security Settings

Dialogflow CX security settings -- the redaction, retention, audio-export, and Insights-export policy that the Dialogflow CX agents in one project and location apply to every conversation. An agent attaches by referencing the settings' `name` output from its `securitySettings` field; several agents can share one set.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `dialogflow.googleapis.com` on the project (never disabled on destroy)
- **Security settings** -- a `dialogflow_cx_security_settings` with its redaction, retention, audio-export, and Insights-export policy

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Dialogflow admin permissions on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpGcsBucket`** -- the bucket telephony audio is exported to (`audioExportSettings.gcsBucket`). Google grants the Dialogflow service agent `roles/storage.objectCreator` on it, which needs `storage.buckets.setIamPolicy` on the bucket from whoever applies.
- **Sensitive Data Protection templates** -- inspect and de-identify templates in the settings' region (`inspectTemplate`, `deidentifyTemplate`), named by their full resource names.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDialogflowCxSecuritySettings
metadata:
  name: pii-redaction
spec:
  location: global
  redactionStrategy: REDACT_WITH_SERVICE
  redactionScope: REDACT_DISK_STORAGE
  retentionWindowDays: 30
```

```shell
planton apply -f dialogflow-cx-security-settings.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | `global` or a Dialogflow region; must match the agents that use the settings. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project, as a literal or a `GcpProject` reference. |
| `displayName` | `string` | `metadata.name` | Unique within the location. |
| `redactionStrategy` | `string` | no redaction | `REDACT_WITH_SERVICE` scrubs sensitive data with Sensitive Data Protection. |
| `redactionScope` | `string` | nothing redacted | `REDACT_DISK_STORAGE`: everything written to durable storage. |
| `inspectTemplate` / `deidentifyTemplate` | `string` | Google's defaults | Sensitive Data Protection templates (project or organization) in the same region. |
| `purgeDataTypes` | `[]string` | none | `DIALOGFLOW_HISTORY`. |
| `retentionStrategy` | `string` | none | `REMOVE_AFTER_CONVERSATION`; also turns off audio and Insights export. |
| `retentionWindowDays` | `int32` | Google's default TTL | Days to keep sensitive data; only values below the default (365, or 30 for Agent Assist) take effect. |
| `audioExportSettings` | `object` | off | `gcsBucket` (`GcpGcsBucket` reference or bucket name), `audioExportPattern`, `audioFormat` (`MULAW`, `MP3`, `OGG`), `enableAudioRedaction`. |
| `enableInsightsExport` | `bool` | `false` | Export finished conversations to Conversational Insights. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- At most one of `retentionStrategy` and `retentionWindowDays` (Google accepts one retention rule).
- `redactionStrategy`, `redactionScope`, `purgeDataTypes`, and `audioFormat` take only the values Google lists.
- The two templates must be full Sensitive Data Protection template names of the matching kind.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/securitySettings/{id}` -- what an agent's `securitySettings` takes |
| `security_settings_id` | `string` | The id Google assigned |
| `location` | `string` | The settings' location |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Same location only.** Settings apply to agents in their own location; a `global` agent needs `global` settings.
- **Removal after the conversation switches exports off.** `REMOVE_AFTER_CONVERSATION` disables audio export and Insights export whatever those fields say.
- **Audio export needs the grant to land.** Google makes the object-creator grant to the Dialogflow service agent when the bucket is set; the applying principal needs `storage.buckets.setIamPolicy` on the bucket.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpDialogflowCxAgent** -- the agents that apply these settings
- **GcpGcsBucket** -- the bucket exported audio lands in

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

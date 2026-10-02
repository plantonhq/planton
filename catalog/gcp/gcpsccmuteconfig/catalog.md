# GCP SCC Mute Config

Silences Security Command Center findings you have decided not to act on -- accepted risks, noisy detectors, sandbox projects -- for a project, a folder, or the whole organization. Muted findings stay on record but drop out of triage, filtered notifications, and exports, so responders see only what needs attention.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `securitycenter.googleapis.com` on a project rule's project
- **Mute config** -- the scope's `securitycenter.V2*MuteConfig`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Security Command Center mute-config admin permissions at the scope. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

## Deploy

### Console

Open the deployment store, find **GCP SCC Mute Config**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Sandbox Project Noise** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSccMuteConfig
metadata:
  name: sandbox-public-buckets
  org: acme-corp
  env: prod
spec:
  muteConfigId: sandbox-public-buckets
  description: Sandbox buckets are public by design
  filter: category = "PUBLIC_BUCKET_ACL" AND resource.project_display_name = "sandbox"
  type: DYNAMIC
```

```shell
planton apply -f scc-mute-config.yaml
```

This mutes public-bucket findings in the sandbox project. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference a `GcpFolder` from `scope.folderId` to apply one rule to every project in a folder.

## Key Configuration

These are the most important decisions when configuring a mute rule. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Dynamic or static** -- `DYNAMIC` mutes existing and future matches and is reversible; `STATIC` permanently mutes future matches.

**Filter** -- which findings, written for the rule's scope.

**Scope** -- a project, a folder, or the organization.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `scope.projectId` | `status.outputs.project_id` |
| **GcpFolder** | `scope.folderId` | `status.outputs.folder_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The rule's resource name | Audits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Sandbox project noise** -- mute a category in one sandbox project. Start from the **Sandbox Project Noise** preset.

**Folder accepted risk** -- mute a low-severity category across a folder. Start from the **Folder Accepted Risk** preset.

## Works With

- [**GCP SCC Notification Config**](/cloud-catalog/gcp-scc-notification-config) -- streams unmuted findings
- [**GCP SCC BigQuery Export**](/cloud-catalog/gcp-scc-bigquery-export) -- exports unmuted findings
- [**GCP Folder**](/cloud-catalog/gcp-folder) -- a folder-wide rule

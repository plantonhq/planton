# GCP SCC Notification Config

Sends Security Command Center findings to Pub/Sub the moment they appear or change, for a project, a folder, or the whole organization. Point it at the topic your alerting, ticketing, or SOAR pipeline listens on, and filter it to the findings your responders act on.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `securitycenter.googleapis.com` on a project config's project
- **Notification config** -- the scope's `securitycenter.V2*NotificationConfig`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Security Command Center notification-config admin permissions at the scope. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

## Deploy

### Console

Open the deployment store, find **GCP SCC Notification Config**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Project High Findings** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSccNotificationConfig
metadata:
  name: high-findings
  org: acme-corp
  env: prod
spec:
  configId: high-findings
  pubsubTopic:
    value: projects/security-ops/topics/scc-findings
  filter: state = "ACTIVE" AND severity = "HIGH"
```

```shell
planton apply -f scc-notification-config.yaml
```

This streams every active high-severity finding in the project to the topic. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference a `GcpPubSubTopic` from `pubsubTopic` and grant the config's `status.outputs.service_account` publish rights on the topic; reference a `GcpFolder` from `scope.folderId` to cover a folder.

## Key Configuration

These are the most important decisions when configuring notifications. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Scope** -- a project, a folder (every project beneath it), or the organization.

**Filter** -- which finding events are streamed; empty streams everything.

**Topic** -- where the findings go; the config's service account needs publish rights there.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `scope.projectId` | `status.outputs.project_id` |
| **GcpFolder** | `scope.folderId` | `status.outputs.folder_id` |
| **GcpPubSubTopic** | `pubsubTopic` | `status.outputs.topic_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `service_account` | The publisher Security Command Center uses | The topic's publisher grant |
| `name` | The config's resource name | Audits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Project high findings** -- active high and critical findings of one project to a topic. Start from the **Project High Findings** preset.

**Organization critical findings** -- every critical finding in the organization to the SOC's topic. Start from the **Organization Critical Findings** preset.

## Works With

- [**GCP Pub/Sub Topic**](/cloud-catalog/gcp-pub-sub-topic) -- the destination
- [**GCP SCC Mute Config**](/cloud-catalog/gcp-scc-mute-config) -- silence accepted findings first
- [**GCP SCC BigQuery Export**](/cloud-catalog/gcp-scc-bigquery-export) -- the findings history

# GCP Cloud Build Trigger

Starts builds automatically. A Cloud Build trigger watches for an event -- a push or pull request on a repository, a message on a Pub/Sub topic, a call to a webhook -- and runs a build when it arrives. The build is either declared right in the trigger or read from a `cloudbuild.yaml` file in your repository.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- the Cloud Build API on the trigger's project
- **Trigger** -- the trigger, with its event source and its build

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with permission to manage Cloud Build triggers in the target project, and to act as the trigger's service account when it sets one. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Account

- **A linked repository** -- for triggers on GitHub, GitLab, or Bitbucket events: a Cloud Build connection and repository (GCP Cloud Build Connection and GCP Cloud Build Repository).

## Deploy

### Console

Open the deployment store, find **GCP Cloud Build Trigger**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Push to Main** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudBuildTrigger
metadata:
  name: orders-main
  org: acme-corp
  env: prod
spec:
  projectId:
    value: acme-ci
  location: us-central1
  repositoryEventConfig:
    repository:
      valueFrom:
        kind: GcpCloudBuildRepository
        name: orders
        fieldPath: status.outputs.name
    push:
      branch: ^main$
  filename: cloudbuild.yaml
```

```shell
planton apply -f cloud-build-trigger.yaml
```

This builds every push to main with the repository's `cloudbuild.yaml`. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference a repository's `status.outputs.name` from `repositoryEventConfig.repository`, a service account's `status.outputs.name` from `serviceAccount`, a topic's `status.outputs.topic_id` from `pubsubConfig.topic`, and a worker pool's `status.outputs.name` from `build.options.workerPool`.

## Key Configuration

These are the most important decisions when configuring a trigger. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**When it fires** -- a repository event (push or pull request), a Pub/Sub message, a webhook call, or only when run by hand.

**What it builds** -- an inline build, a `cloudbuild.yaml` in the triggering repository, or a `cloudbuild.yaml` from a named repository.

**Who it runs as** -- a dedicated service account with only the access the build needs, or the project's default Cloud Build account.

**Where it runs** -- the global location or a region; a trigger on a regional repository or private pool lives in that region.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpServiceAccount** | `serviceAccount` | `status.outputs.name` |
| **GcpServiceAccount** | `pubsubConfig.serviceAccountEmail` | `status.outputs.email` |
| **GcpCloudBuildRepository** | `repositoryEventConfig.repository`, `sourceToBuild.repository`, `gitFileSource.repository` | `status.outputs.name` |
| **GcpPubSubTopic** | `pubsubConfig.topic` | `status.outputs.topic_id` |
| **GcpSecretManagerSecret** | `webhookConfig.secret`, `build.availableSecrets.secretManager[].versionName` | `status.outputs.latest_version_name` |
| **GcpCloudBuildWorkerPool** | `build.options.workerPool` | `status.outputs.name` |
| **GcpGcsBucket** | `build.logsBucket` | `status.outputs.url` |
| **GcpKmsKey** | `build.secrets[].kmsKeyName` | `status.outputs.key_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `id` | The trigger's full resource ID | Audits, tooling |
| `trigger_id` | Google's generated trigger ID | Running the trigger from scripts or schedulers |
| `name` | The trigger's name | Console and log lookups |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Push to main** -- build every push to the main branch with the repository's `cloudbuild.yaml`. Start from the **Push to Main** preset.

**Pull request checks** -- build every pull request against main, gated by a `/gcbrun` comment for outside contributors. Start from the **Pull Request Checks** preset.

**Event-driven build** -- run an inline build whenever a message lands on a Pub/Sub topic. Start from the **Pub/Sub Build** preset.

## Works With

- [**GCP Cloud Build Repository**](/cloud-catalog/gcp-cloud-build-repository) -- the repository a trigger builds from
- [**GCP Cloud Build Worker Pool**](/cloud-catalog/gcp-cloud-build-worker-pool) -- private machines builds run on
- [**GCP Service Account**](/cloud-catalog/gcp-service-account) -- the identity builds run as
- [**GCP Pub/Sub Topic**](/cloud-catalog/gcp-pub-sub-topic) -- the topic a Pub/Sub trigger listens on
- [**GCP Secret Manager Secret**](/cloud-catalog/gcp-secret-manager-secret) -- webhook keys and build secrets

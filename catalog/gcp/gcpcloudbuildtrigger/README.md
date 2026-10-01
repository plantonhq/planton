# GCP Cloud Build Trigger

Declares a Cloud Build trigger: the rule that starts a build when something happens, and the build it starts. A trigger fires on pushes or pull requests to a repository, on a Pub/Sub message, on a webhook call, or on demand; it builds either an inline build declared in the spec or a `cloudbuild.yaml` file from a repository.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `cloudbuild.googleapis.com` on the trigger's project (never disabled on destroy)
- **Trigger** -- one `cloudbuild_trigger`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/cloudbuild.builds.editor` (or the permissions in `iac/permissions.yaml`) on the trigger's project, plus `roles/iam.serviceAccountUser` on the trigger's service account when `serviceAccount` is set.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpServiceAccount`** -- the identity builds run as (`serviceAccount`), and the push identity of a Pub/Sub trigger (`pubsubConfig.serviceAccountEmail`).
- **`GcpCloudBuildRepository`** -- a repository linked through a Cloud Build connection (`repositoryEventConfig.repository`, `sourceToBuild.repository`, `gitFileSource.repository`).
- **`GcpPubSubTopic`** -- the topic a Pub/Sub trigger listens on (`pubsubConfig.topic`).
- **`GcpSecretManagerSecret`** -- a webhook key (`webhookConfig.secret`) or a build secret (`build.availableSecrets.secretManager[].versionName`).
- **`GcpCloudBuildWorkerPool`** -- a private pool for an inline build (`build.options.workerPool`).
- **`GcpGcsBucket`** -- a logs bucket (`build.logsBucket`); **`GcpKmsKey`** -- the key for KMS-encrypted build secrets (`build.secrets[].kmsKeyName`).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudBuildTrigger
metadata:
  name: orders-main
spec:
  projectId:
    value: acme-ci
  location: us-central1
  serviceAccount:
    valueFrom:
      kind: GcpServiceAccount
      name: orders-builder
      fieldPath: status.outputs.name
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

## Configuration Reference

### Exactly One Build Configuration

| Field | Type | Description |
|-------|------|-------------|
| `build` | object | The build inline: `steps` (at least one), `timeout`, `queueTtl`, `images`, `logsBucket`, `substitutions`, `tags`, `artifacts`, `options`, `source`, `availableSecrets`, `secrets`. |
| `filename` | `string` | A `cloudbuild.yaml` path in the triggering repository (repository-event and Cloud Source Repository triggers). |
| `gitFileSource` | object | A `cloudbuild.yaml` path in a named repository and revision (Pub/Sub, webhook, and manual triggers): `path`, `repoType`, `repository` or `uri`, `revision`. |

### Event Sources (at most one in practice; none is a manual trigger)

| Field | Description |
|-------|-------------|
| `repositoryEventConfig` | `repository` plus exactly one of `pullRequest` or `push`. |
| `github` | `owner`, `name`, `enterpriseConfigResourceName`, exactly one of `pullRequest` (branch required) or `push`. Not with `triggerTemplate`. |
| `bitbucketServerTriggerConfig` | `bitbucketServerConfigResource`, `projectKey`, `repoSlug`, exactly one of `pullRequest` (branch required) or `push`. |
| `developerConnectEventConfig` | `gitRepositoryLink`, exactly one of `pullRequest` or `push`. |
| `triggerTemplate` | A Cloud Source Repository: `repoName`, `projectId`, `dir`, exactly one of `branchName`, `tagName`, or `commitSha`, `invertRegex`. Not with `github`. |
| `pubsubConfig` | `topic` (required), `serviceAccountEmail`. |
| `webhookConfig` | `secret` (a secret version, required). |

A `pullRequest` filter takes `branch`, `commentControl`, and `invertRegex`; a `push` filter takes exactly one of `branch` or `tag`, and `invertRegex`.

### Other Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `string` / ref | provider project | The trigger's project (`GcpProject` ref). Immutable. |
| `location` | `string` | `global` | The Cloud Build location. Immutable. |
| `triggerName` | `string` | `metadata.name` | The trigger's name; renames in place. |
| `description` | `string` | -- | Human-readable description. |
| `disabled` | `bool` | `false` | Stop the trigger from starting builds. |
| `tags` | list | -- | Tags on the trigger. |
| `substitutions` | `map` | -- | User substitutions (`_NAME` keys). |
| `serviceAccount` | `string` / ref | default Cloud Build account | `projects/{p}/serviceAccounts/{email}`. |
| `approvalConfig.approvalRequired` | `bool` | `false` | Builds wait for an approver. |
| `sourceToBuild` | object | -- | `repository` or `uri`, `ref` (`refs/...`), `repoType`, and the enterprise or Bitbucket config. |
| `filter` | `string` | -- | A CEL filter on the event (Pub/Sub and webhook). |
| `ignoredFiles` / `includedFiles` | list | -- | File globs that decide whether a change builds. |
| `includeBuildLogs` | `string` | -- | `INCLUDE_BUILD_LOGS_WITH_STATUS` (GitHub triggers only). |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- Exactly one of `build`, `filename`, or `gitFileSource`; `github` and `triggerTemplate` are mutually exclusive.
- Each repository event source takes exactly one of `pullRequest` or `push`; a push filter takes exactly one of `branch` or `tag`; GitHub and Bitbucket Server pull requests need `branch`.
- `triggerTemplate` and `build.source.repoSource` take exactly one of a branch, tag, or commit; `build.source` takes exactly one of `repoSource` or `storageSource`.
- `sourceToBuild` takes exactly one of `repository` or `uri`; `gitFileSource` takes at most one.
- A step with `script` sets neither `entrypoint` nor `args`.
- `triggerName` is 1-64 letters, digits, or dashes; substitution keys match `^_[A-Z0-9_]+$`; literal secret versions, topics, repositories, service accounts, and KMS keys use their full resource names.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `id` | `string` | `projects/{p}/locations/{l}/triggers/{trigger_id}` (`projects/{p}/triggers/{trigger_id}` when global) |
| `trigger_id` | `string` | Google's generated trigger ID |
| `name` | `string` | The trigger's name |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Name and ID are different things.** `triggerName` is yours; `trigger_id` is the ID Google generates, the one `gcloud builds triggers run` takes.
- **Region.** A trigger on a repository linked through a regional connection, or whose build uses a private pool, must be in that region.
- **Service accounts and logs.** A build running as a user-specified account needs `build.logsBucket` or `build.options.logging` set to `CLOUD_LOGGING_ONLY` or `NONE`.
- **Webhook secrets.** Grant Google's Cloud Build service agent `roles/secretmanager.secretAccessor` on the webhook secret.
- **Private pools.** `build.options.workerPool` is the field the provider offers, and Google's API marks it deprecated in favor of `options.pool.name`; builds from `filename` or `gitFileSource` set `options.pool.name` in their `cloudbuild.yaml`.
- **Destroy** deletes the trigger; builds it already started run to completion and stay in the history.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpCloudBuildRepository** -- repositories the trigger builds from
- **GcpCloudBuildWorkerPool** -- private pools builds run on
- **GcpServiceAccount** -- the identity builds run as
- **GcpPubSubTopic** -- the topic a Pub/Sub trigger listens on
- **GcpSecretManagerSecret** -- webhook keys and build secrets

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

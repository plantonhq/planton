# GCP Workflow

Creates a Cloud Workflows workflow — a serverless orchestrator that executes a sequence of steps (HTTP calls, GCP connector calls, conditionals, retries) defined in YAML or JSON. The glue between services: an Eventarc trigger fires an execution, the workflow calls Cloud Run / BigQuery / Pub/Sub in order, handles errors, and records the result. Every source, env-var, or service-account change deploys a new revision; executions already running finish on the revision they started with.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Workflows API enablement** -- the module enables `workflows.googleapis.com` on the target project (never disabled on destroy), so a fresh project works without manual API setup
- **Workflow** -- a Cloud Workflows workflow with the configured source, service account, CMEK, call-log and execution-history levels, environment variables, and deletion guards
- **GCP Labels** -- resource metadata labels merged with any user labels for tracking and governance

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** to host the workflow. Provide the project ID directly, reference a GcpProject Cloud Resource via ValueFromRef, or omit `projectId` to use the provider connection's default project. The module enables the Workflows API itself.
- **IAM** -- the deploying identity needs `roles/workflows.editor` or broader, plus `iam.serviceAccounts.actAs` on the workflow's service account when one is set.
- **CMEK grant** (only for `cryptoKey`) -- grant the Workflows service agent `roles/cloudkms.cryptoKeyEncrypterDecrypter` on the key BEFORE deploying, or the deploy fails.

## Deploy

### Console

Open the deployment store, find **GCP Workflow**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields — the workflow source goes in as YAML or JSON. Start from the **HTTP Orchestrator** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpWorkflow
metadata:
  name: order-orchestrator
  org: acme-corp
  env: prod
spec:
  region: us-central1
  sourceContents: |
    main:
      params: [input]
      steps:
        - charge:
            call: http.post
            args:
              url: https://payments.internal/charge
              auth:
                type: OIDC
            result: charge_result
        - done:
            return: ${charge_result.body}
  serviceAccount:
    value: order-workflow@acme-prod-12345.iam.gserviceaccount.com
```

```shell
planton apply -f workflow.yaml
```

This deploys a two-step HTTP orchestration in `us-central1` running as a dedicated service account, with deletion protection on by default. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the workflow to resources deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  serviceAccount:
    valueFrom:
      kind: GcpServiceAccount
      name: order-workflow-sa
      fieldPath: status.outputs.email
```

The InfraPipeline resolves the dependency graph, deploys the project and service account first, then provisions the workflow — and a GcpEventarcTrigger node can reference this workflow's `workflow_id` as its destination, completing the event-driven backbone in one chart.

## Key Configuration

These are the most important decisions when configuring a workflow. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Workflow source** -- `sourceContents` is the definition executed on each run, capped at 128KB by the API (enforced at manifest time). Every change deploys a NEW revision; running executions finish on the revision they started with, so a mid-flight deploy never breaks an in-progress orchestration.

**Execution identity** -- `serviceAccount` is the identity the workflow's HTTP and connector calls carry. Grant a dedicated account only the roles the steps need — the default compute account is fine for experiments, wrong for production. Changing it deploys a new revision.

**Observability levels** -- `callLogLevel` and `executionHistoryLevel` set how much call and step detail lands in Cloud Logging and execution history. `LOG_ALL_CALLS` plus `EXECUTION_HISTORY_DETAILED` while developing; scale back for cost in steady state — all-calls logging on a busy workflow is a real log bill.

**Environment variables** -- `userEnvVars` are visible to the source via `sys.get_env()`: at most 20 entries, and keys must not start with `GOOGLE` or `WORKFLOWS` (reserved prefixes the API rejects). An env-var change deploys a new revision. Secrets go in `secretEnvVars`: each value is stored in its own Secret Manager secret, readable only by the workflow's service account, and the env var of the same name holds the version's resource name for the source to read with the Secret Manager connector. The 20-entry limit counts both maps together, and the two maps cannot share a key.

**Resource manager tags** -- `resourceManagerTags` bind org-policy/cost tags at creation and are ForceNew: a tag change REPLACES the workflow, and the replacement starts with a fresh, empty execution history. Plan tag changes deliberately.

**Deletion guards** -- `deletionProtection` is ON by default: destroy fails until you explicitly set it `false`. The workflow's execution HISTORY dies with the workflow, so the guard is protecting your audit trail, not just the definition. `deletionPolicy: ABANDON` is the alternative when the workflow should keep running outside Planton management.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpServiceAccount** (optional) | `serviceAccount` | `status.outputs.email` |
| **GcpKmsKey** (optional) | `cryptoKey` | `status.outputs.key_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `workflow_id` | Full resource name (`projects/{p}/locations/{region}/workflows/{name}`) | A GcpEventarcTrigger's `destination.workflow`; a GcpEventarcMessageBus pipeline's `destination.workflow` |
| `workflow_name` | Short name | Display, gcloud commands |
| `revision_id` | Deployed revision (e.g. `000001-a4d`) | Deploy tracking — compare across applies to confirm a deploy actually rolled |
| `state` | `ACTIVE` / `UNAVAILABLE` | Health checks (`UNAVAILABLE` signals an unusable CMEK key) |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**HTTP orchestrator** -- call services in order with per-step auth and retries; the workflow owns the sequencing and error handling that would otherwise live in application code. Start from the **HTTP Orchestrator** preset.

**Event-driven step function** -- an Eventarc trigger starts an execution per event; the workflow fans out the work. Start from the **Event Handler** preset.

## Works With

- [**GCP Eventarc Trigger**](/cloud-catalog/gcp-eventarc-trigger) -- starts an execution per matching event
- [**GCP Eventarc Message Bus**](/cloud-catalog/gcp-eventarc-message-bus) -- pipelines deliver messages into executions
- [**GCP Service Account**](/cloud-catalog/gcp-service-account) -- the identity the steps run as
- [**GCP KMS Key**](/cloud-catalog/gcp-kms-key) -- CMEK for workflow and execution data
- [**GCP Project**](/cloud-catalog/gcp-project) -- provides the GCP project where the workflow is created

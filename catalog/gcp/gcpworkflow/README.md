# GCP Workflow

Creates a Cloud Workflows workflow — a serverless orchestrator that executes a sequence of steps (HTTP calls, GCP connector calls, conditionals, retries) defined in YAML or JSON. The glue between services: an Eventarc trigger fires an execution, the workflow calls Cloud Run / BigQuery / Pub/Sub in order, handles errors, and records the result.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Workflow** -- a `workflows.Workflow` with the configured source, service account, CMEK, logging levels, and env vars
- **Secret Manager secrets** -- for each `secretEnvVars` entry, one secret in the workflow's region holding the value as a pinned version, readable only by the workflow's service account; the workflow's env var holds the version's resource name, never the value
- **Workflows API enablement** -- `workflows.googleapis.com` enabled in the target project (never disabled on destroy)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** -- required when using Runner-based credential delivery.

### GCP Project

- **A GCP project** to host the workflow (directly or via a GcpProject reference).
- **IAM**: see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying identity needs.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpWorkflow
metadata:
  name: order-orchestrator
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
    value: order-workflow@my-project.iam.gserviceaccount.com
```

```shell
planton apply -f workflow.yaml
```

Every source change deploys a NEW revision; running executions finish on the revision they started with.

## Outputs

| Output | Description |
|--------|-------------|
| `workflow_id` | Full resource name — the value Eventarc destinations consume |
| `workflow_name` | The short workflow name |
| `revision_id` | The deployed revision |
| `state` | Workflow state (`ACTIVE` when executable) |

## Works With

- **GcpEventarcTrigger** -- `destination.workflow` starts an execution per matching event
- **GcpEventarcMessageBus** -- a pipeline's `destination.workflow` starts an execution per delivered message
- **GcpServiceAccount** -- the identity the workflow's steps run as
- **GcpKmsKey** -- CMEK for workflow and execution data
- **GcpProject** -- provides the GCP project the workflow lives in

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

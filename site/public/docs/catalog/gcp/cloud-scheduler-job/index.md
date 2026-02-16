---
title: "Cloud Scheduler Job"
description: "Cloud Scheduler Job deployment documentation"
icon: "package"
order: 100
componentName: "gcpcloudschedulerjob"
---

# GCP Cloud Scheduler Job

Deploys a Google Cloud Scheduler job that executes on a unix-cron schedule, dispatching to an HTTP endpoint, Pub/Sub topic, or App Engine handler with optional authentication and configurable retry behavior.

## What Gets Created

When you deploy a GcpCloudSchedulerJob resource, OpenMCF provisions:

- **Cloud Scheduler Job** — a `google_cloud_scheduler_job` resource configured with the specified schedule, target, and retry policy
- **Authentication Token** (optional) — OIDC or OAuth2 token generation for secure HTTP target invocation, using the specified service account

## Prerequisites

- **GCP credentials** configured via environment variables or OpenMCF provider config
- **A GCP project** with the Cloud Scheduler API enabled
- **A target endpoint** — an HTTP URL, a Pub/Sub topic, or an App Engine handler
- **A service account** with `iam.serviceAccounts.actAs` permission if using OIDC/OAuth authentication
- **Cloud Run invoker role** (`roles/run.invoker`) on the service account if targeting Cloud Run

## Quick Start

Create a file `scheduler-job.yaml`:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudSchedulerJob
metadata:
  name: my-cron-job
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpCloudSchedulerJob.my-cron-job
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  schedule: "0 9 * * 1-5"
  httpTarget:
    uri: https://example.com/api/trigger
    httpMethod: GET
```

Deploy:

```shell
openmcf apply -f scheduler-job.yaml
```

This creates a Cloud Scheduler job that sends an HTTP GET to `https://example.com/api/trigger` every weekday at 9:00 AM UTC.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `StringValueOrRef` | GCP project where the job is created | Required |
| `projectId.value` | `string` | Direct project ID value | — |
| `projectId.valueFrom` | `object` | Foreign key reference to a GcpProject resource | Default kind: `GcpProject` |
| `location` | `string` | GCP region for the job (e.g., `us-central1`) | Required, immutable |
| `schedule` | `string` | Unix-cron schedule expression (e.g., `"0 9 * * 1"`) | Required |
| One of: `httpTarget`, `pubsubTarget`, `appEngineHttpTarget` | `object` | Exactly one target type must be specified | CEL: exactly one set |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `jobName` | `string` | `metadata.name` | Explicit GCP job name. Immutable after creation. Must start with a letter. |
| `timeZone` | `string` | `Etc/UTC` | Timezone for schedule interpretation (tz database name). |
| `description` | `string` | — | Human-readable description. Maximum 500 characters. |
| `attemptDeadline` | `string` | `180s` | Request timeout as a duration string (e.g., `"600s"`). HTTP: 15s-30min. App Engine: 15s-24h. Ignored for Pub/Sub. |
| `paused` | `bool` | `false` | If `true`, job is created in paused state (won't execute until resumed). |
| `retryConfig.retryCount` | `int` | GCP default | Number of retry attempts. Max 5. |
| `retryConfig.maxRetryDuration` | `string` | GCP default | Maximum retry window (e.g., `"3600s"`). `"0s"` for unlimited. |
| `retryConfig.minBackoffDuration` | `string` | GCP default | Minimum wait between retries (e.g., `"5s"`). |
| `retryConfig.maxBackoffDuration` | `string` | GCP default | Maximum wait between retries (e.g., `"3600s"`). |
| `retryConfig.maxDoublings` | `int` | GCP default | Number of times the retry interval doubles before becoming linear. |

### HTTP Target Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `httpTarget.uri` | `string` | — | Full URI of the HTTP endpoint. **Required** when using HTTP target. |
| `httpTarget.httpMethod` | `string` | `POST` | HTTP method: `POST`, `GET`, `HEAD`, `PUT`, `DELETE`, `PATCH`, `OPTIONS`. |
| `httpTarget.body` | `string` | — | Base64-encoded request body. Only for POST, PUT, PATCH methods. |
| `httpTarget.headers` | `map(string)` | `{}` | HTTP request headers. Cannot set `Content-Length` or `X-Google-*` headers. |
| `httpTarget.oidcToken.serviceAccountEmail` | `StringValueOrRef` | — | Service account for OIDC token. Mutually exclusive with `oauthToken`. |
| `httpTarget.oidcToken.audience` | `string` | target URI | OIDC audience. Defaults to the target URI if not set. |
| `httpTarget.oauthToken.serviceAccountEmail` | `StringValueOrRef` | — | Service account for OAuth2 token. Mutually exclusive with `oidcToken`. |
| `httpTarget.oauthToken.scope` | `string` | `cloud-platform` | OAuth2 scope for the access token. |

### Pub/Sub Target Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `pubsubTarget.topicName` | `StringValueOrRef` | — | Fully qualified Pub/Sub topic name. **Required** when using Pub/Sub target. |
| `pubsubTarget.topicName.value` | `string` | — | Direct topic name (e.g., `projects/my-project/topics/my-topic`). |
| `pubsubTarget.topicName.valueFrom` | `object` | — | Foreign key reference to a GcpPubSubTopic resource. Default kind: `GcpPubSubTopic`. |
| `pubsubTarget.data` | `string` | — | Base64-encoded message payload. |
| `pubsubTarget.attributes` | `map(string)` | `{}` | Message attributes (key-value metadata). |

### App Engine Target Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `appEngineHttpTarget.relativeUri` | `string` | — | URI path starting with `/`. **Required** when using App Engine target. Max 2083 chars. |
| `appEngineHttpTarget.httpMethod` | `string` | `POST` | HTTP method. Same values as HTTP target. |
| `appEngineHttpTarget.body` | `string` | — | Base64-encoded request body. Only for POST and PUT methods. |
| `appEngineHttpTarget.headers` | `map(string)` | `{}` | HTTP request headers. |
| `appEngineHttpTarget.appEngineRouting.service` | `string` | default | App Engine service name. |
| `appEngineHttpTarget.appEngineRouting.version` | `string` | default | App Engine version. |
| `appEngineHttpTarget.appEngineRouting.instance` | `string` | — | Specific App Engine instance. |

## Examples

### OIDC-Authenticated Cloud Run Trigger

Securely invoke a Cloud Run service every weekday at 9am Eastern:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudSchedulerJob
metadata:
  name: daily-report
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpCloudSchedulerJob.daily-report
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  schedule: "0 9 * * 1-5"
  timeZone: America/New_York
  description: Triggers daily report generation
  attemptDeadline: "600s"
  httpTarget:
    uri: https://report-service-abc123.run.app/generate
    httpMethod: POST
    body: eyJhY3Rpb24iOiAiZGFpbHlfcmVwb3J0In0=
    headers:
      Content-Type: application/json
    oidcToken:
      serviceAccountEmail:
        value: invoker@my-gcp-project.iam.gserviceaccount.com
      audience: https://report-service-abc123.run.app
  retryConfig:
    retryCount: 3
    maxRetryDuration: "1800s"
    minBackoffDuration: "5s"
    maxBackoffDuration: "600s"
    maxDoublings: 3
```

### Pub/Sub Scheduled Publisher

Publish a message to a Pub/Sub topic every 5 minutes:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudSchedulerJob
metadata:
  name: pipeline-trigger
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpCloudSchedulerJob.pipeline-trigger
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  schedule: "*/5 * * * *"
  description: Triggers data pipeline every 5 minutes
  pubsubTarget:
    topicName:
      value: projects/my-gcp-project/topics/pipeline-trigger
    data: eyJwaXBlbGluZSI6ICJkYWlseS1ldGwifQ==
    attributes:
      source: cloud-scheduler
      pipeline: daily-etl
  retryConfig:
    retryCount: 5
    maxDoublings: 3
```

### Using Foreign Key References

Wire dependencies from other OpenMCF-managed resources:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudSchedulerJob
metadata:
  name: composed-scheduler
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpCloudSchedulerJob.composed-scheduler
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      field: status.outputs.project_id
  location: us-central1
  schedule: "0 8 * * *"
  pubsubTarget:
    topicName:
      valueFrom:
        kind: GcpPubSubTopic
        name: events-topic
        field: status.outputs.topic_id
    data: eyJ0cmlnZ2VyIjogInNjaGVkdWxlZCJ9
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `job_id` | `string` | Fully qualified job ID in the format `projects/{project}/locations/{location}/jobs/{name}` |
| `job_name` | `string` | Short job name (matches `jobName` or `metadata.name`) |
| `state` | `string` | Current job state: `ENABLED`, `PAUSED`, `DISABLED`, or `UPDATE_FAILED` |

## Related Components

- [GcpPubSubTopic](/docs/catalog/gcp/pubsub-topic) — provides the target topic for Pub/Sub scheduled publishing
- [GcpCloudTasksQueue](/docs/catalog/gcp/cloud-tasks-queue) — asynchronous task dispatch (complementary to cron-based scheduling)
- [GcpCloudRun](/docs/catalog/gcp/cloud-run) — common HTTP target for scheduled job invocation
- [GcpServiceAccount](/docs/catalog/gcp/service-account) — provides the identity for OIDC/OAuth authentication
- [GcpProject](/docs/catalog/gcp/project) — project hosting the scheduler job and target resources

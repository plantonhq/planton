# GCP Cloud Tasks Queue

Deploys a GCP Cloud Tasks queue with configurable dispatch rate limits, retry policies, and optional queue-level HTTP target configuration with OIDC or OAuth authentication. Cloud Tasks queues do not support GCP labels.

## What Gets Created

When you deploy a GcpCloudTasksQueue resource, OpenMCF provisions:

- **Cloud Tasks Queue** — a `google_cloud_tasks_queue` resource in the specified project and region with the configured dispatch and retry settings
- **Rate Limits** — created only when `rateLimits` is specified, controls dispatch rate and concurrency
- **Retry Config** — created only when `retryConfig` is specified, controls exponential backoff and max attempts
- **HTTP Target** — created only when `httpTarget` is specified, configures queue-level HTTP method, authentication, URI override, and header defaults for all tasks
- **Stackdriver Logging** — created only when `stackdriverLoggingConfig` is specified, enables dispatch operation logging at the configured sampling ratio

## Prerequisites

- **GCP credentials** configured via environment variables or OpenMCF provider config
- **A GCP project** with the Cloud Tasks API enabled
- **A service account** with `iam.serviceAccounts.actAs` permission if configuring queue-level OIDC or OAuth authentication
- **The target service** (e.g., Cloud Run, Cloud Functions) deployed if configuring `httpTarget.uriOverride`

## Quick Start

Create a file `cloud-tasks-queue.yaml`:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudTasksQueue
metadata:
  name: my-queue
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpCloudTasksQueue.my-queue
spec:
  projectId:
    value: my-gcp-project
  queueName: my-task-queue
  location: us-central1
```

Deploy:

```shell
openmcf apply -f cloud-tasks-queue.yaml
```

This creates a Cloud Tasks queue in `us-central1` with GCP-managed defaults for rate limits and retry behavior.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `StringValueOrRef` | GCP project where the queue will be created. Can reference a GcpProject resource via `valueFrom`. | Required |
| `queueName` | `string` | Name of the queue. Immutable after creation. | 1-63 chars, starts with letter, `^[a-zA-Z][a-zA-Z0-9-]*$` |
| `location` | `string` | GCP region (e.g., `us-central1`). Immutable after creation. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `desiredState` | `string` | `RUNNING` | Queue dispatch state. `RUNNING` dispatches tasks normally. `PAUSED` holds tasks without dispatching. Valid: `RUNNING`, `PAUSED`. |
| `httpTarget.httpMethod` | `string` | — | HTTP method override for all tasks. Valid: `POST`, `GET`, `HEAD`, `PUT`, `DELETE`, `PATCH`, `OPTIONS`. |
| `httpTarget.headerOverrides` | `object[]` | `[]` | HTTP headers to set on all tasks. Each entry has `key` and `value`. |
| `httpTarget.oauthToken.serviceAccountEmail` | `StringValueOrRef` | — | Service account for OAuth2 token generation. Can reference GcpServiceAccount via `valueFrom`. Required when `oauthToken` is set. Mutually exclusive with `oidcToken`. |
| `httpTarget.oauthToken.scope` | `string` | `cloud-platform` | OAuth scope for the access token. |
| `httpTarget.oidcToken.serviceAccountEmail` | `StringValueOrRef` | — | Service account for OIDC token generation. Can reference GcpServiceAccount via `valueFrom`. Required when `oidcToken` is set. Mutually exclusive with `oauthToken`. |
| `httpTarget.oidcToken.audience` | `string` | target URI | Audience for the OIDC token. |
| `httpTarget.uriOverride.scheme` | `string` | — | URI scheme override. Valid: `HTTP`, `HTTPS`. |
| `httpTarget.uriOverride.host` | `string` | — | URI host override. Replaces the host part of all task URLs. |
| `httpTarget.uriOverride.port` | `string` | — | URI port override. Positive integer as string. |
| `httpTarget.uriOverride.path` | `string` | — | URI path override. Replaces the path of all task URLs. |
| `httpTarget.uriOverride.queryParams` | `string` | — | URI query override (e.g., `key=val&key2=val2`). |
| `httpTarget.uriOverride.enforceMode` | `string` | `ALWAYS` | When to apply overrides. `ALWAYS` replaces task URIs. `IF_NOT_EXISTS` only applies if the task lacks the component. |
| `rateLimits.maxDispatchesPerSecond` | `double` | GCP default | Maximum tasks dispatched per second. |
| `rateLimits.maxConcurrentDispatches` | `int` | GCP default | Maximum concurrent task executions. |
| `retryConfig.maxAttempts` | `int` | GCP default | Maximum attempts per task. Set to `-1` for unlimited. |
| `retryConfig.maxRetryDuration` | `string` | GCP default | Maximum total retry time (e.g., `86400s`). `0s` for unlimited. |
| `retryConfig.minBackoff` | `string` | GCP default | Minimum wait between retries (e.g., `0.100s`). |
| `retryConfig.maxBackoff` | `string` | GCP default | Maximum wait between retries (e.g., `3600s`). |
| `retryConfig.maxDoublings` | `int` | GCP default | Number of times the backoff interval doubles before becoming linear. |
| `stackdriverLoggingConfig.samplingRatio` | `double` | `0.0` | Fraction of dispatch operations to log. Must be 0.0-1.0. |

## Examples

### Rate-Limited Background Processing

Queue with explicit rate limits and retry configuration for production background processing:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudTasksQueue
metadata:
  name: background-processor
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpCloudTasksQueue.background-processor
spec:
  projectId:
    value: my-gcp-project
  queueName: background-processor
  location: us-central1
  rateLimits:
    maxDispatchesPerSecond: 500
    maxConcurrentDispatches: 100
  retryConfig:
    maxAttempts: 5
    minBackoff: "1s"
    maxBackoff: "3600s"
    maxDoublings: 16
  stackdriverLoggingConfig:
    samplingRatio: 0.1
```

### Cloud Run Target with OIDC Authentication

Queue configured to dispatch all tasks to a Cloud Run service with automatic OIDC token generation. This is the recommended pattern for serverless task processing:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudTasksQueue
metadata:
  name: cloud-run-dispatcher
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpCloudTasksQueue.cloud-run-dispatcher
spec:
  projectId:
    value: my-gcp-project
  queueName: cloud-run-dispatcher
  location: us-central1
  httpTarget:
    httpMethod: POST
    oidcToken:
      serviceAccountEmail:
        value: task-invoker@my-gcp-project.iam.gserviceaccount.com
      audience: https://my-service-abc123-uc.a.run.app
    uriOverride:
      scheme: HTTPS
      host: my-service-abc123-uc.a.run.app
      path: /v1/tasks/process
      enforceMode: ALWAYS
    headerOverrides:
      - key: Content-Type
        value: application/json
  rateLimits:
    maxDispatchesPerSecond: 100
    maxConcurrentDispatches: 50
  retryConfig:
    maxAttempts: 3
    minBackoff: "5s"
    maxBackoff: "300s"
    maxDoublings: 4
  stackdriverLoggingConfig:
    samplingRatio: 1.0
```

### Full-Featured Queue with Foreign Key References

Production queue referencing other OpenMCF-managed resources via `valueFrom` for infra-chart composition:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudTasksQueue
metadata:
  name: composed-queue
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpCloudTasksQueue.composed-queue
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      fieldPath: status.outputs.project_id
  queueName: composed-task-queue
  location: us-central1
  desiredState: RUNNING
  httpTarget:
    httpMethod: POST
    oidcToken:
      serviceAccountEmail:
        valueFrom:
          kind: GcpServiceAccount
          name: task-invoker-sa
          fieldPath: status.outputs.email
      audience: https://my-service.run.app
    uriOverride:
      scheme: HTTPS
      host: my-service.run.app
      enforceMode: ALWAYS
  rateLimits:
    maxDispatchesPerSecond: 200
    maxConcurrentDispatches: 50
  retryConfig:
    maxAttempts: 5
    minBackoff: "1s"
    maxBackoff: "3600s"
    maxDoublings: 16
    maxRetryDuration: "86400s"
  stackdriverLoggingConfig:
    samplingRatio: 0.5
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `queue_id` | `string` | Fully qualified queue path: `projects/{project}/locations/{location}/queues/{name}` |
| `queue_name` | `string` | Short queue name (same as `spec.queueName`) |
| `state` | `string` | Current queue state: `RUNNING`, `PAUSED`, or `DISABLED` (Pulumi only) |

## Related Components

- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project for queue creation
- [GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount) — provides the service account for OIDC/OAuth authentication in `httpTarget`
- [GcpCloudSchedulerJob](/docs/catalog/gcp/gcpcloudschedulerjob) — creates scheduled tasks that target Cloud Tasks queues
- [GcpPubSubTopic](/docs/catalog/gcp/gcppubsubtopic) — alternative messaging pattern for event fan-out (vs task dispatch)

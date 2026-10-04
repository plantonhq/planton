# GcpCloudTasksQueue

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpCloudTasksQueueSpec defines the configuration for a GCP Cloud Tasks queue.

Cloud Tasks is a fully managed service for executing, dispatching, and
delivering tasks asynchronously. A queue manages the dispatch rate, retry
behavior, and routing of tasks to their target handlers.

Cloud Tasks supports two task target types:
  - HTTP tasks (modern): Tasks dispatched to any HTTP endpoint. Configure
    queue-level authentication and routing via http_target.
  - App Engine tasks (legacy): Tasks dispatched to App Engine handlers.
    Not supported by this component (use application-level configuration).

Important behavioral notes:

  - The queue_name and location fields are immutable after creation.
    Changing them requires destroying and recreating the queue. Queue names
    are deliberately explicit (never derived from metadata) because the
    Cloud Tasks API reserves a deleted queue's ID for up to 7 days after
    deletion — an accidental name change burns the old identifier for that
    window.

  - Cloud Tasks queues do NOT support GCP labels.

  - Pause/resume is declarative here: editing desired_state and applying
    pauses or resumes the queue. But the provider tracks the field as a
    config-only (virtual) value — it never reads the queue's live dispatch
    state back into it — so an out-of-band gcloud pause is INVISIBLE to an
    apply whose desired_state did not change (live-verified: the apply
    plans zero changes and the queue stays paused). To recover a queue
    paused out-of-band, either resume it out-of-band or flip the spec
    PAUSED → apply → RUNNING → apply (the value must change for the
    provider to issue the resume call).

  - Rate limits and retry config have GCP-computed defaults when not specified.
    For most workloads, the defaults are reasonable starting points.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudTasksQueue
metadata:
  name: test-cloud-tasks-queue
spec:
  projectId:
    value: my-gcp-project
  queueName: my-task-queue
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
  # Destroy really destroys in E2E: the live lanes prove the full lifecycle.
  deletionPolicy: DELETE
  # Explicit dispatch state: the queue runs (and the module reconciles the
  # value on every apply -- pause/resume is declarative).
  desiredState: RUNNING
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.queueName` | `string` | yes |  |  |
| `spec.location` | `string` | yes |  |  |
| `spec.httpTarget` | `GcpCloudTasksQueueHttpTarget` |  |  |  |
| `spec.httpTarget.httpMethod` | `string` |  |  |  |
| `spec.httpTarget.headerOverrides` | `[]GcpCloudTasksQueueHttpHeaderOverride` |  |  |  |
| `spec.httpTarget.headerOverrides[].key` | `string` | yes |  |  |
| `spec.httpTarget.headerOverrides[].value` | `string` | yes |  |  |
| `spec.httpTarget.oauthToken` | `GcpCloudTasksQueueOAuthToken` |  |  |  |
| `spec.httpTarget.oauthToken.serviceAccountEmail` | `string \| valueFrom` | yes |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.httpTarget.oauthToken.scope` | `string` |  |  |  |
| `spec.httpTarget.oidcToken` | `GcpCloudTasksQueueOidcToken` |  |  |  |
| `spec.httpTarget.oidcToken.serviceAccountEmail` | `string \| valueFrom` | yes |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.httpTarget.oidcToken.audience` | `string` |  |  |  |
| `spec.httpTarget.uriOverride` | `GcpCloudTasksQueueUriOverride` |  |  |  |
| `spec.httpTarget.uriOverride.scheme` | `string` |  |  |  |
| `spec.httpTarget.uriOverride.host` | `string` |  |  |  |
| `spec.httpTarget.uriOverride.port` | `string` |  |  |  |
| `spec.httpTarget.uriOverride.path` | `string` |  |  |  |
| `spec.httpTarget.uriOverride.queryParams` | `string` |  |  |  |
| `spec.httpTarget.uriOverride.enforceMode` | `string` |  |  |  |
| `spec.appEngineRoutingOverride` | `GcpCloudTasksQueueAppEngineRoutingOverride` |  |  |  |
| `spec.appEngineRoutingOverride.service` | `string` |  |  |  |
| `spec.appEngineRoutingOverride.version` | `string` |  |  |  |
| `spec.appEngineRoutingOverride.instance` | `string` |  |  |  |
| `spec.rateLimits` | `GcpCloudTasksQueueRateLimits` |  |  |  |
| `spec.rateLimits.maxDispatchesPerSecond` | `double` |  |  |  |
| `spec.rateLimits.maxConcurrentDispatches` | `int32` |  |  |  |
| `spec.retryConfig` | `GcpCloudTasksQueueRetryConfig` |  |  |  |
| `spec.retryConfig.maxAttempts` | `int32` |  |  |  |
| `spec.retryConfig.maxRetryDuration` | `string` |  |  |  |
| `spec.retryConfig.minBackoff` | `string` |  |  |  |
| `spec.retryConfig.maxBackoff` | `string` |  |  |  |
| `spec.retryConfig.maxDoublings` | `int32` |  |  |  |
| `spec.stackdriverLoggingConfig` | `GcpCloudTasksQueueLoggingConfig` |  |  |  |
| `spec.stackdriverLoggingConfig.samplingRatio` | `double` |  |  |  |
| `spec.desiredState` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

GCP project where the queue will be created.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.queueName

`string` · required

Name of the Cloud Tasks queue.
Must start with a letter, contain only letters, numbers, and hyphens,
and be between 1 and 63 characters.
Immutable after creation, and deliberately required: a deleted queue's
ID is reserved by the API for up to 7 days, so the name deserves an
explicit, stable choice rather than a derived default.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"63","pattern":"^[a-zA-Z][a-zA-Z0-9-]*$"}}

### spec.location

`string` · required

GCP region where the queue will be created (e.g., "us-central1").
Immutable after creation.

- rule: {"required":true}

### spec.httpTarget

`GcpCloudTasksQueueHttpTarget`

Queue-level HTTP target configuration. When set, these settings apply
to all HTTP tasks dispatched from this queue, overriding task-level
HTTP configuration.

This is the recommended pattern for microservices: configure auth and
routing at the queue level, then enqueue tasks with just a request body.

- rule: only one of oauth_token or oidc_token can be set, not both

### spec.httpTarget.httpMethod

`string`

HTTP method override for all tasks in this queue.
When specified, overrides the method on individual tasks.
Note: if set to GET, the task body is ignored at execution time.
Valid values: "POST", "GET", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS",
or the API's explicit "HTTP_METHOD_UNSPECIFIED" sentinel (equivalent to
leaving the override unset).

- rule: http_method must be one of: HTTP_METHOD_UNSPECIFIED, POST, GET, HEAD, PUT, DELETE, PATCH, OPTIONS

### spec.httpTarget.headerOverrides

`[]GcpCloudTasksQueueHttpHeaderOverride`

HTTP headers to set on all tasks dispatched from this queue.
These headers override any task-level headers with the same key.
Header size must be less than 80KB total.

### spec.httpTarget.headerOverrides[].key

`string` · required

The header field name.

- rule: {"required":true}

### spec.httpTarget.headerOverrides[].value

`string` · required

The header field value.

- rule: {"required":true}

### spec.httpTarget.oauthToken

`GcpCloudTasksQueueOAuthToken`

OAuth2 access token configuration for authenticating HTTP task requests.
Use for calling Google APIs on *.googleapis.com.
Mutually exclusive with oidc_token.

### spec.httpTarget.oauthToken.serviceAccountEmail

`string | valueFrom` · required

Service account email to generate the OAuth token.
The service account must be within the same project as the queue.
The caller must have iam.serviceAccounts.actAs on this service account.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.httpTarget.oauthToken.scope

`string`

OAuth scope for the generated access token.
If not specified, defaults to "https://www.googleapis.com/auth/cloud-platform".

### spec.httpTarget.oidcToken

`GcpCloudTasksQueueOidcToken`

OIDC token configuration for authenticating HTTP task requests.
Use for calling Cloud Run, Cloud Functions, or custom endpoints.
Mutually exclusive with oauth_token.

### spec.httpTarget.oidcToken.serviceAccountEmail

`string | valueFrom` · required

Service account email to generate the OIDC token.
The service account must be within the same project as the queue.
The caller must have iam.serviceAccounts.actAs on this service account.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.httpTarget.oidcToken.audience

`string`

Audience for the generated OIDC token.
If not specified, the URI specified in the target will be used.

### spec.httpTarget.uriOverride

`GcpCloudTasksQueueUriOverride`

URI override settings. When specified, modifies the URI of all tasks
dispatched from this queue before dispatch.

### spec.httpTarget.uriOverride.scheme

`string`

Scheme override. Replaces the task URI scheme with HTTP or HTTPS.
Valid values: "HTTP", "HTTPS".

- rule: scheme must be HTTP or HTTPS

### spec.httpTarget.uriOverride.host

`string`

Host override. Replaces the host part of the task URL.
For example, if the task URL is "https://www.google.com" and host is
"example.net", the overridden URI becomes "https://example.net".
Must not be empty when set (INVALID_ARGUMENT).

### spec.httpTarget.uriOverride.port

`string`

Port override. Replaces the port part of the task URI.
Must be a positive integer. Setting to "0" clears the URI port.

### spec.httpTarget.uriOverride.path

`string`

Path override. Replaces the existing path of the task URL.
Setting to an empty string clears the URI path segment.

### spec.httpTarget.uriOverride.queryParams

`string`

Query parameters override. Replaces the query part of the task URI.
For example: "qparam1=123&qparam2=456".
Setting to an empty string clears the URI query segment.

### spec.httpTarget.uriOverride.enforceMode

`string`

URI Override Enforce Mode.
ALWAYS: Always override the task URI (default).
IF_NOT_EXISTS: Only apply the override if the task does not already have
the corresponding URI component set.

- rule: enforce_mode must be ALWAYS or IF_NOT_EXISTS

### spec.appEngineRoutingOverride

`GcpCloudTasksQueueAppEngineRoutingOverride`

App Engine routing override for App Engine tasks in this queue.
When set, overrides each task's own App Engine routing so the whole
queue targets one service/version/instance. Only relevant for queues
dispatching App Engine tasks; ignored for HTTP tasks.

### spec.appEngineRoutingOverride.service

`string`

The App Engine service to route queue tasks to.
If not specified, the task is sent to the service that is the default
service when the task is attempted.

### spec.appEngineRoutingOverride.version

`string`

The App Engine version to route queue tasks to.
If not specified, the task is sent to the version that is the default
version when the task is attempted.

### spec.appEngineRoutingOverride.instance

`string`

The App Engine instance to route queue tasks to.
If not specified, the task is sent to an instance which is available
when the task is attempted (subject to the service/version's scaling).

### spec.rateLimits

`GcpCloudTasksQueueRateLimits`

Rate limits for task dispatches. Controls how fast and how many tasks
are dispatched concurrently.

### spec.rateLimits.maxDispatchesPerSecond

`double`

Maximum rate at which tasks are dispatched from this queue (tasks/second).
If unspecified, Cloud Tasks picks a default based on queue configuration.

### spec.rateLimits.maxConcurrentDispatches

`int32`

Maximum number of concurrent tasks that Cloud Tasks allows to be
dispatched for this queue. After this threshold, Cloud Tasks stops
dispatching until the concurrency drops.
If unspecified, Cloud Tasks picks a default.

### spec.retryConfig

`GcpCloudTasksQueueRetryConfig`

Retry configuration for failed task attempts. Controls backoff behavior,
maximum attempts, and retry duration.

### spec.retryConfig.maxAttempts

`int32`

Number of attempts per task. Includes the first attempt.
Must be >= -1. Set to -1 for unlimited attempts.
If unspecified, Cloud Tasks picks a default.

### spec.retryConfig.maxRetryDuration

`string`

Maximum time limit for retrying a failed task, measured from the first attempt.
Once elapsed, no further attempts are made regardless of max_attempts.
Set to "0s" for unlimited retry duration.
Format: duration string (e.g., "3600s" for 1 hour).

### spec.retryConfig.minBackoff

`string`

Minimum wait time between retry attempts.
Format: duration string (e.g., "0.100s" for 100ms).

### spec.retryConfig.maxBackoff

`string`

Maximum wait time between retry attempts.
Format: duration string (e.g., "3600s" for 1 hour).

### spec.retryConfig.maxDoublings

`int32`

The number of times the retry interval doubles before becoming constant.
The retry interval starts at min_backoff, doubles max_doublings times,
then increases linearly until reaching max_backoff.

### spec.stackdriverLoggingConfig

`GcpCloudTasksQueueLoggingConfig`

Cloud Logging configuration for task dispatch operations.

### spec.stackdriverLoggingConfig.samplingRatio

`double`

Fraction of operations to log. Must be between 0.0 and 1.0 inclusive.
0.0 means no logging (default), 1.0 means log every dispatch operation.

- rule: sampling_ratio must be between 0.0 and 1.0 inclusive

### spec.desiredState

`string`

Dispatch state the queue is held in:
  ""        -- same as "RUNNING" (the provider default)
  "RUNNING" -- tasks are dispatched to their targets
  "PAUSED"  -- tasks accumulate in the queue but none are dispatched —
               the safe holding state during target maintenance or
               incident response
Declarative for spec-driven transitions: editing this value and
applying pauses/resumes the queue (live-verified both directions).
NOT drift-correcting: the provider treats this as a config-only
(virtual) field and never reads the live dispatch state back, so an
out-of-band gcloud pause survives applies whose spec value is
unchanged. Recover by resuming out-of-band, or by flipping this field
PAUSED → apply → RUNNING → apply so the value change triggers the
provider's resume call.

- rule: desired_state must be RUNNING or PAUSED

### spec.deletionPolicy

`string`

What destroying this resource does to the queue:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the queue and every task still in it are deleted; the
               queue's ID is reserved by the API for up to 7 days
  "PREVENT" -- destroy FAILS; protects a queue whose backlog must not
               be lost
  "ABANDON" -- the queue is removed from management but keeps running
               (and dispatching) in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpCloudTasksQueue, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.queue_id` | `string` | The fully qualified queue ID. Format: projects/{project}/locations/{location}/queues/{name} This is the value downstream resources use to reference this queue. |
| `status.outputs.queue_name` | `string` | The short queue name (same as the spec's queue_name input). |
| `status.outputs.max_burst_size` | `int32` | The effective max burst size of the queue's rate limits. Computed by GCP from max_dispatches_per_second — it is not configurable. Max burst size caps how fast queued tasks are processed when many tasks are enqueued in a short period at a high dispatch rate. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.httpTarget.oauthToken.serviceAccountEmail` | GcpServiceAccount | `status.outputs.email` |
| `spec.httpTarget.oidcToken.serviceAccountEmail` | GcpServiceAccount | `status.outputs.email` |

## See Also

- [Overview](../README.md)

# GcpCloudSchedulerJob

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpCloudSchedulerJobSpec defines the configuration for a GCP Cloud Scheduler job.

Cloud Scheduler is a fully managed cron job scheduler. It allows you to
schedule virtually any job, including batch, big data, cloud infrastructure
operations, and other recurring workloads. Each job runs on a user-defined
schedule (unix-cron format) and dispatches to one of three target types:
HTTP endpoints, Pub/Sub topics, or App Engine handlers.

Important behavioral notes:

  - The job_name and location fields are immutable after creation.
    Changing them requires destroying and recreating the job.

  - Cloud Scheduler jobs do NOT support GCP labels.

  - Exactly one target type must be specified: http_target, pubsub_target,
    or app_engine_http_target. Multiple targets are not supported.

  - The schedule field uses unix-cron format (e.g., "0 9 * * 1" for
    every Monday at 9:00 AM). The time_zone field controls which time
    zone the schedule is interpreted in.

  - For HTTP targets, authentication via OAuth or OIDC tokens ensures
    secure invocation of Cloud Run, Cloud Functions, and other endpoints.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudSchedulerJob
metadata:
  name: test-cloud-scheduler-job
spec:
  projectId:
    value: my-gcp-project
  jobName: daily-report-trigger
  location: us-central1
  schedule: "0 9 * * 1-5"
  timeZone: America/New_York
  description: Triggers daily report generation on weekdays at 9am ET
  attemptDeadline: "300s"
  httpTarget:
    uri: https://my-service-abc123.run.app/api/report
    httpMethod: POST
    body: eyJhY3Rpb24iOiAiZ2VuZXJhdGVfcmVwb3J0In0=
    headers:
      Content-Type: application/json
    oidcToken:
      serviceAccountEmail:
        value: invoker@my-gcp-project.iam.gserviceaccount.com
  retryConfig:
    retryCount: 3
    maxRetryDuration: "1800s"
    minBackoffDuration: "5s"
    maxBackoffDuration: "600s"
    maxDoublings: 3
  # Destroy really destroys in E2E: the live lanes prove the full lifecycle.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.jobName` | `string` |  |  |  |
| `spec.location` | `string` | yes |  |  |
| `spec.schedule` | `string` | yes |  |  |
| `spec.timeZone` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.attemptDeadline` | `string` |  |  |  |
| `spec.paused` | `bool` |  |  |  |
| `spec.httpTarget` | `GcpCloudSchedulerJobHttpTarget` |  |  |  |
| `spec.httpTarget.uri` | `string` | yes |  |  |
| `spec.httpTarget.httpMethod` | `string` |  |  |  |
| `spec.httpTarget.body` | `string` |  |  |  |
| `spec.httpTarget.headers` | `map<string, string>` |  |  |  |
| `spec.httpTarget.oauthToken` | `GcpCloudSchedulerJobOAuthToken` |  |  |  |
| `spec.httpTarget.oauthToken.serviceAccountEmail` | `string \| valueFrom` | yes |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.httpTarget.oauthToken.scope` | `string` |  |  |  |
| `spec.httpTarget.oidcToken` | `GcpCloudSchedulerJobOidcToken` |  |  |  |
| `spec.httpTarget.oidcToken.serviceAccountEmail` | `string \| valueFrom` | yes |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.httpTarget.oidcToken.audience` | `string` |  |  |  |
| `spec.pubsubTarget` | `GcpCloudSchedulerJobPubsubTarget` |  |  |  |
| `spec.pubsubTarget.topicName` | `string \| valueFrom` | yes |  | GcpPubSubTopic (`status.outputs.topic_id`) |
| `spec.pubsubTarget.data` | `string` |  |  |  |
| `spec.pubsubTarget.attributes` | `map<string, string>` |  |  |  |
| `spec.appEngineHttpTarget` | `GcpCloudSchedulerJobAppEngineHttpTarget` |  |  |  |
| `spec.appEngineHttpTarget.relativeUri` | `string` | yes |  |  |
| `spec.appEngineHttpTarget.httpMethod` | `string` |  |  |  |
| `spec.appEngineHttpTarget.body` | `string` |  |  |  |
| `spec.appEngineHttpTarget.headers` | `map<string, string>` |  |  |  |
| `spec.appEngineHttpTarget.appEngineRouting` | `GcpCloudSchedulerJobAppEngineRouting` |  |  |  |
| `spec.appEngineHttpTarget.appEngineRouting.service` | `string` |  |  |  |
| `spec.appEngineHttpTarget.appEngineRouting.version` | `string` |  |  |  |
| `spec.appEngineHttpTarget.appEngineRouting.instance` | `string` |  |  |  |
| `spec.retryConfig` | `GcpCloudSchedulerJobRetryConfig` |  |  |  |
| `spec.retryConfig.retryCount` | `int32` |  |  |  |
| `spec.retryConfig.maxRetryDuration` | `string` |  |  |  |
| `spec.retryConfig.minBackoffDuration` | `string` |  |  |  |
| `spec.retryConfig.maxBackoffDuration` | `string` |  |  |  |
| `spec.retryConfig.maxDoublings` | `int32` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

GCP project where the scheduler job will be created.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.jobName

`string`

Name of the Cloud Scheduler job.
If not specified, defaults to metadata.name.
Immutable after creation.

- rule: job_name must start with a letter and contain only letters, numbers, hyphens, and underscores

### spec.location

`string` · required

GCP region where the scheduler job will be created (e.g., "us-central1").
Immutable after creation.

- rule: {"required":true}

### spec.schedule

`string` · required

Cron schedule on which the job will be executed.
Uses unix-cron format (e.g., "*/5 * * * *" for every 5 minutes,
"0 9 * * 1" for every Monday at 9:00 AM).
The schedule is interpreted in the time zone specified by time_zone.

- rule: {"required":true}

### spec.timeZone

`string`

Time zone name from the tz database (e.g., "America/New_York", "Europe/London").
If not specified, defaults to "Etc/UTC".
See: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones

### spec.description

`string`

Human-readable description of the job.
Maximum 500 characters.

- rule: {"string":{"maxLen":"500"}}

### spec.attemptDeadline

`string`

The deadline for job attempts. If the request handler does not respond
by this deadline, the request is cancelled and the attempt is marked
as a DEADLINE_EXCEEDED failure.

For HTTP targets: between 15 seconds and 30 minutes.
For App Engine targets: between 15 seconds and 24 hours 15 seconds.
For Pub/Sub targets: this field is ignored.

Format: duration string (e.g., "180s", "30m").
If not specified, defaults to "180s" (3 minutes).

### spec.paused

`bool`

If true, the job will be created in a paused state (will not execute
on schedule until resumed). Defaults to false (job starts enabled).

### spec.httpTarget

`GcpCloudSchedulerJobHttpTarget`

HTTP target configuration. Dispatches the job to an HTTP endpoint.
This is the most common target type, used for triggering Cloud Run
services, Cloud Functions, webhooks, or any HTTP-accessible endpoint.
Exactly one target must be specified.

- rule: only one of oauth_token or oidc_token can be set, not both

### spec.httpTarget.uri

`string` · required

The full URI of the HTTP target.
Required. Must be a valid HTTP or HTTPS URL.

- rule: {"required":true}

### spec.httpTarget.httpMethod

`string`

HTTP request method.
If not specified, defaults to POST.
Valid values: "POST", "GET", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS".

- rule: http_method must be one of: POST, GET, HEAD, PUT, DELETE, PATCH, OPTIONS

### spec.httpTarget.body

`string`

HTTP request body.
A request body is allowed only if the HTTP method is POST, PUT, or PATCH.
Must be base64-encoded.

### spec.httpTarget.headers

`map<string, string>`

HTTP request headers.
The following headers cannot be set: Content-Length, User-Agent,
and headers matching X-Google-* or X-AppEngine-*.

### spec.httpTarget.oauthToken

`GcpCloudSchedulerJobOAuthToken`

OAuth2 access token configuration for authenticating requests.
Use for calling Google APIs on *.googleapis.com.
Mutually exclusive with oidc_token.

### spec.httpTarget.oauthToken.serviceAccountEmail

`string | valueFrom` · required

Service account email to generate the OAuth token.
The service account must be within the same project as the job.
The caller must have iam.serviceAccounts.actAs on this service account.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.httpTarget.oauthToken.scope

`string`

OAuth scope for the generated access token.
If not specified, defaults to "https://www.googleapis.com/auth/cloud-platform".

### spec.httpTarget.oidcToken

`GcpCloudSchedulerJobOidcToken`

OIDC token configuration for authenticating requests.
Use for calling Cloud Run, Cloud Functions, or custom endpoints.
Mutually exclusive with oauth_token.

### spec.httpTarget.oidcToken.serviceAccountEmail

`string | valueFrom` · required

Service account email to generate the OIDC token.
The service account must be within the same project as the job.
The caller must have iam.serviceAccounts.actAs on this service account.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.httpTarget.oidcToken.audience

`string`

Audience for the generated OIDC token.
If not specified, the URI of the HTTP target will be used.

### spec.pubsubTarget

`GcpCloudSchedulerJobPubsubTarget`

Pub/Sub target configuration. Publishes a message to a Pub/Sub topic
when the job executes. Use this for event-driven architectures where
downstream consumers process messages asynchronously.
Exactly one target must be specified.

### spec.pubsubTarget.topicName

`string | valueFrom` · required

The fully qualified Pub/Sub topic name to publish to.
Format: projects/{project}/topics/{topic}

- references: GcpPubSubTopic (`status.outputs.topic_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpPubSubTopic, name: <that resource's name>, fieldPath: status.outputs.topic_id}} -- a bare string does not parse

### spec.pubsubTarget.data

`string`

The message payload for the Pub/Sub message.
Must be base64-encoded.
The message must contain either non-empty data OR at least one attribute.

### spec.pubsubTarget.attributes

`map<string, string>`

Attributes for the Pub/Sub message.
Key-value pairs attached to the message as metadata.
The message must contain either non-empty data OR at least one attribute.

### spec.appEngineHttpTarget

`GcpCloudSchedulerJobAppEngineHttpTarget`

App Engine HTTP target configuration. Dispatches the job to an
App Engine handler within the same project. Use this when the
target handler runs on App Engine.
Exactly one target must be specified.

### spec.appEngineHttpTarget.relativeUri

`string` · required

The relative URI of the App Engine handler.
Must begin with "/" and have a maximum length of 2083 characters.

- rule: relative_uri must begin with '/'
- rule: {"required":true}

### spec.appEngineHttpTarget.httpMethod

`string`

HTTP request method.
If not specified, defaults to POST.
Valid values: "POST", "GET", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS".

- rule: http_method must be one of: POST, GET, HEAD, PUT, DELETE, PATCH, OPTIONS

### spec.appEngineHttpTarget.body

`string`

HTTP request body.
A request body is allowed only if the HTTP method is POST or PUT.
Must be base64-encoded.

### spec.appEngineHttpTarget.headers

`map<string, string>`

HTTP request headers.
The following headers cannot be set: Content-Length, Host, User-Agent,
and headers matching X-Google-* or X-AppEngine-*.

### spec.appEngineHttpTarget.appEngineRouting

`GcpCloudSchedulerJobAppEngineRouting`

App Engine routing configuration.
Controls which App Engine service, version, and instance handles
the request. If not set, the default service and version are used.

### spec.appEngineHttpTarget.appEngineRouting.service

`string`

The App Engine service to route the request to.
If not specified, the default service is used.

### spec.appEngineHttpTarget.appEngineRouting.version

`string`

The App Engine version to route the request to.
If not specified, the default version is used.

### spec.appEngineHttpTarget.appEngineRouting.instance

`string`

The App Engine instance to route the request to.
If not specified, the request is routed according to the service/version
traffic splitting configuration.

### spec.retryConfig

`GcpCloudSchedulerJobRetryConfig`

Retry configuration for failed job attempts.
Controls exponential backoff behavior, maximum attempts, and
retry duration limits.

### spec.retryConfig.retryCount

`int32`

The number of attempts that the system will make to run a job using
the exponential backoff procedure. Values greater than 5 and negative
values are not allowed.

### spec.retryConfig.maxRetryDuration

`string`

The time limit for retrying a failed job, measured from when the job
was first attempted. Once elapsed, no further attempts are made.
Format: duration string (e.g., "3600s" for 1 hour).
Set to "0s" for unlimited retry duration.

### spec.retryConfig.minBackoffDuration

`string`

The minimum amount of time to wait before retrying a job after it fails.
Format: duration string (e.g., "5s" for 5 seconds).

### spec.retryConfig.maxBackoffDuration

`string`

The maximum amount of time to wait before retrying a job after it fails.
Format: duration string (e.g., "3600s" for 1 hour).

### spec.retryConfig.maxDoublings

`int32`

The number of times that the retry interval doubles before becoming
constant. The retry interval starts at min_backoff_duration, then
doubles max_doublings times, and increases linearly thereafter.

### spec.deletionPolicy

`string`

What destroying this resource does to the job:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the job is deleted and its schedule stops firing
  "PREVENT" -- destroy FAILS; protects a job whose missed runs would
               break downstream systems
  "ABANDON" -- the job is removed from management but keeps firing on
               schedule in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `exactly_one_target`: exactly one of http_target, pubsub_target, or app_engine_http_target must be set

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpCloudSchedulerJob, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.job_id` | `string` | The fully qualified job ID. Format: projects/{project}/locations/{location}/jobs/{name} This is the value downstream resources use to reference this job. |
| `status.outputs.job_name` | `string` | The short job name (same as the spec's job_name input, or metadata.name if job_name was not specified). |
| `status.outputs.state` | `string` | The current state of the job. Possible values: ENABLED, PAUSED, DISABLED, UPDATE_FAILED. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.httpTarget.oauthToken.serviceAccountEmail` | GcpServiceAccount | `status.outputs.email` |
| `spec.httpTarget.oidcToken.serviceAccountEmail` | GcpServiceAccount | `status.outputs.email` |
| `spec.pubsubTarget.topicName` | GcpPubSubTopic | `status.outputs.topic_id` |

## See Also

- [Overview](../README.md)

# GCP Cloud Scheduler Job

Deploys a Cloud Scheduler cron job that dispatches to an HTTP endpoint, Pub/Sub topic, or App Engine handler on a user-defined schedule. The job supports OAuth and OIDC token authentication for secure invocation of Cloud Run and Cloud Functions targets, configurable retry behavior with exponential backoff, and time zone-aware scheduling. The GCP project, the token-minting service account, and the Pub/Sub topic all wire by ValueFromRef, so the job drops into an InfraChart beside the workload it triggers.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Cloud Scheduler Job** -- a managed cron job in the specified GCP project and region, configured with the chosen schedule (unix-cron format), time zone, target type, and retry policy
- **HTTP Target Configuration** -- created only when `httpTarget` is specified; configures the URI, HTTP method, request body, headers, and optional OAuth or OIDC token authentication
- **Pub/Sub Target Configuration** -- created only when `pubsubTarget` is specified; configures the topic name, message data, and message attributes
- **App Engine HTTP Target Configuration** -- created only when `appEngineHttpTarget` is specified; configures the relative URI, HTTP method, and App Engine routing
- **Retry Configuration** -- created only when `retryConfig` is specified; configures exponential backoff with retry count, min/max backoff durations, and max doublings

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the Cloud Scheduler job will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **Cloud Scheduler API** enabled in the target project.
- **GCP service account** (if using authenticated HTTP targets) -- the service account must exist in the same project, and the deploying principal must have `iam.serviceAccounts.actAs` permission on it. Reference via ValueFromRef to a GcpServiceAccount Infra Component.
- **Pub/Sub topic** (if using Pub/Sub target) -- the topic must exist before deploying the scheduler job. Reference via ValueFromRef to a GcpPubSubTopic Infra Component.

## Deploy

### Console

Open the deployment store, find **GCP Cloud Scheduler Job**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Basic HTTP Job** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudSchedulerJob
metadata:
  name: nightly-cleanup
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  location: us-central1
  schedule: "0 2 * * *"
  timeZone: America/New_York
  httpTarget:
    uri: https://api.example.com/cleanup
    httpMethod: POST
```

```shell
planton apply -f scheduler-job.yaml
```

This creates a Cloud Scheduler job that sends an HTTP POST to the specified endpoint every day at 2:00 AM Eastern; no authentication, retry configuration, or request body is configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the scheduler job to a GCP project, service account, and Pub/Sub topic deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  httpTarget:
    uri: https://my-service-url.run.app/process
    oidcToken:
      serviceAccountEmail:
        valueFrom:
          kind: GcpServiceAccount
          name: scheduler-invoker
          fieldPath: status.outputs.email
```

The InfraPipeline resolves the dependency graph, deploys the project and service account first, then provisions the scheduler job with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Cloud Scheduler job. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Schedule and time zone** -- The `schedule` field uses unix-cron format (e.g., `"0 9 * * 1"` for every Monday at 9:00 AM). Set `timeZone` to a tz database name (e.g., `America/New_York`); defaults to `Etc/UTC` if not specified. The `location` field is immutable after creation.

**Target type** -- Exactly one of `httpTarget`, `pubsubTarget`, or `appEngineHttpTarget` must be specified. HTTP targets are the most common, used for Cloud Run, Cloud Functions, and external webhooks. Pub/Sub targets decouple scheduling from processing.

**Authentication** -- For HTTP targets calling Cloud Run or Cloud Functions, configure `oidcToken` with a service account email. For targets calling Google APIs, use `oauthToken` instead. These are mutually exclusive. Unauthenticated HTTP targets require no token configuration.

**Retry behavior** -- Configure `retryConfig` to control failure handling. Set `retryCount` (max 5), `minBackoffDuration` and `maxBackoffDuration` for exponential backoff timing, and `maxDoublings` to control the transition from exponential to linear backoff.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpServiceAccount** (optional) | `httpTarget.oauthToken.serviceAccountEmail` | `status.outputs.email` |
| **GcpServiceAccount** (optional) | `httpTarget.oidcToken.serviceAccountEmail` | `status.outputs.email` |
| **GcpPubSubTopic** (optional) | `pubsubTarget.topicName` | `status.outputs.topic_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `job_id` | Fully qualified job ID (`projects/{project}/locations/{location}/jobs/{name}`) | Monitoring dashboards, audit log filters |
| `job_name` | Short job name | Reference in application configuration |
| `state` | Current job state (ENABLED, PAUSED, DISABLED) | Health monitoring, automation triggers |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Basic HTTP job** -- An unauthenticated HTTP GET on a cron schedule. Suitable for health checks, public webhooks, or simple periodic triggers that do not require authentication. Start from the **Basic HTTP Job** preset.

**Pub/Sub publisher** -- Publishes a message with payload and attributes to a Pub/Sub topic on a schedule. Suitable for event-driven architectures where downstream consumers process messages asynchronously. Start from the **Pub/Sub Publisher** preset.

**Secure Cloud Run trigger** -- An OIDC-authenticated HTTP POST to a Cloud Run service with custom retry configuration and exponential backoff. Suitable for production workloads that require authenticated, reliable invocation of private Cloud Run endpoints. Start from the **Secure Cloud Run Trigger** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the scheduler job is created
- [**GCP Service Account**](/infra-catalog/gcp-service-account) -- provides the identity for OAuth or OIDC token generation
- [**GCP Pub/Sub Topic**](/infra-catalog/gcp-pub-sub-topic) -- provides the Pub/Sub topic for message publishing targets
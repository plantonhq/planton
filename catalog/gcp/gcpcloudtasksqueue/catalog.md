# GCP Cloud Tasks Queue

Deploys a Cloud Tasks queue with configurable rate limits, retry behavior, queue-level HTTP target settings with OAuth or OIDC authentication, URI overrides, and dispatch logging. The queue manages asynchronous task dispatch to HTTP endpoints with controlled concurrency and automatic retry with exponential backoff. The GCP project and the token-minting service account wire by ValueFromRef, so the queue composes into an InfraChart beside the service it dispatches to.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Cloud Tasks Queue** -- a managed task queue in the specified GCP project and region, configured with rate limits, retry behavior, and optional queue-level dispatch defaults
- **Queue-Level HTTP Target** -- created only when `httpTarget` is configured; sets queue-wide authentication (OAuth or OIDC), HTTP method override, header overrides, and URI overrides applied to all dispatched tasks
- **Rate Limits** -- created only when `rateLimits` is configured; controls maximum dispatch rate (tasks/second) and maximum concurrent dispatches
- **Retry Configuration** -- created only when `retryConfig` is configured; controls retry attempts, backoff durations, and the exponential-to-linear backoff transition
- **Logging Configuration** -- created only when `stackdriverLoggingConfig` is configured; controls the fraction of dispatch operations written to Cloud Logging

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the Cloud Tasks queue will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **Cloud Tasks API** enabled in the target project.
- **GCP service account** (if using queue-level authentication) -- the service account must exist in the same project, and the deploying principal must have `iam.serviceAccounts.actAs` permission on it. Reference via ValueFromRef to a GcpServiceAccount Infra Component.
- **HTTP endpoint** (Cloud Run, Cloud Functions, or external) that the queue will dispatch tasks to.

## Deploy

### Console

Open the deployment store, find **GCP Cloud Tasks Queue**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Basic Queue** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudTasksQueue
metadata:
  name: background-jobs
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  queueName: background-jobs
  location: us-central1
```

```shell
planton apply -f cloud-tasks-queue.yaml
```

This creates a Cloud Tasks queue with GCP-managed defaults for rate limits and retry behavior; no queue-level HTTP target, authentication, or logging is configured -- individual tasks define their own targets. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the queue to a GCP project and service account deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  httpTarget:
    oidcToken:
      serviceAccountEmail:
        valueFrom:
          kind: GcpServiceAccount
          name: task-invoker
          fieldPath: status.outputs.email
```

The InfraPipeline resolves the dependency graph, deploys the project and service account first, then provisions the Cloud Tasks queue with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Cloud Tasks queue. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Queue name is a burned identifier** -- `queueName` and `location` are immutable, and the Cloud Tasks API reserves a deleted queue's ID for up to 7 days after deletion — an accidental rename burns the old name for that window. Choose a stable, explicit name up front.

**Rate limits** -- Set `rateLimits.maxDispatchesPerSecond` to control how fast tasks are dispatched and `rateLimits.maxConcurrentDispatches` to limit parallel task execution. Use rate limits to protect downstream services from being overwhelmed. GCP provides reasonable defaults when not specified.

**Retry behavior** -- Configure `retryConfig` with `maxAttempts` (set to -1 for unlimited), `minBackoff` and `maxBackoff` for exponential backoff timing, and `maxDoublings` to control the transition from exponential to linear backoff. Tasks that exhaust all retries are dropped.

**Queue-level HTTP target** -- Configure `httpTarget` with authentication, URI overrides, and method overrides that apply to all tasks in the queue. This is the recommended pattern for microservices: set auth and routing at the queue level, then enqueue tasks with just a request body.

**Authentication** -- For queues targeting Cloud Run or Cloud Functions, configure `httpTarget.oidcToken` with a service account email. For queues targeting Google APIs, use `httpTarget.oauthToken` instead. These are mutually exclusive.

**Pause and resume** -- Set `desiredState: PAUSED` and apply to hold every task in the queue while producers keep enqueueing; flip it back to `RUNNING` and apply to resume dispatch. One boundary to know: the field only acts when its value changes between applies -- a queue paused out-of-band with `gcloud tasks queues pause` is not resumed by re-applying an unchanged spec (resume it the way it was paused, or flip the spec value through `PAUSED` and back).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpServiceAccount** (optional) | `httpTarget.oauthToken.serviceAccountEmail` | `status.outputs.email` |
| **GcpServiceAccount** (optional) | `httpTarget.oidcToken.serviceAccountEmail` | `status.outputs.email` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `queue_id` | Fully qualified queue ID (`projects/{project}/locations/{location}/queues/{name}`) | Task enqueue operations, monitoring dashboards |
| `queue_name` | Short queue name | Application configuration, reference in task creation |
| `max_burst_size` | GCP-computed burst bucket derived from the rate limits | Capacity planning, dispatch-behavior tuning |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Basic queue** -- A minimal queue with GCP-managed defaults for rate limits and retry behavior. Tasks define their own HTTP targets individually. Suitable for simple background processing where individual tasks manage their own routing. Start from the **Basic Queue** preset.

**Rate-limited processing** -- A queue with explicit rate limits (500 tasks/second, 100 concurrent) and retry configuration with exponential backoff. Suitable for workloads that need controlled dispatch rates to protect downstream services. Start from the **Rate-Limited Processing** preset.

**Secure Cloud Run target** -- A queue with OIDC authentication, HTTPS URI override pointing to a Cloud Run service, and controlled dispatch rates. Suitable for microservices architectures where all tasks in the queue target the same authenticated Cloud Run endpoint. Start from the **Secure Cloud Run Target** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the queue is created
- [**GCP Service Account**](/infra-catalog/gcp-service-account) -- provides the identity for OAuth or OIDC token generation on task dispatch
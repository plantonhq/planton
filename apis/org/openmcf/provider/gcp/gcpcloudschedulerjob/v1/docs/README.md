# GcpCloudSchedulerJob: Research & Design Documentation

## Service Overview

Google Cloud Scheduler is a fully managed, enterprise-grade cron job service. It allows you to schedule arbitrary jobs -- including batch, big data jobs, cloud infrastructure operations, and other workloads -- with guaranteed delivery and configurable retry behavior. Under the hood, Cloud Scheduler is built on the same infrastructure as App Engine Cron.

### Key Characteristics

- **Managed cron**: No servers to provision, no crontabs to manage
- **Three target types**: HTTP, Pub/Sub, App Engine
- **At-least-once delivery**: Jobs may fire more than once in rare cases (design for idempotency)
- **Configurable retry**: Exponential backoff with doublings, count limits, duration limits
- **IAM authentication**: OIDC and OAuth2 tokens for secure endpoint invocation
- **Regional**: Jobs run in a specific region (latency considerations for targets)

## Deployment Landscape

### How Cloud Scheduler Jobs Are Provisioned Today

| Method | Usage | Pros | Cons |
|--------|-------|------|------|
| **Google Cloud Console** | Ad-hoc, small teams | Visual, quick setup | No version control, no repeatability |
| **gcloud CLI** | Scripts, CI/CD | Scriptable, familiar | Imperative, no drift detection |
| **Terraform** (`google_cloud_scheduler_job`) | IaC teams | Declarative, drift detection | HCL syntax, state management |
| **Pulumi** (`cloudscheduler.Job`) | Developer-centric IaC | Go/TypeScript, programmatic | Requires Pulumi runtime |
| **OpenMCF GcpCloudSchedulerJob** | Multi-cloud platform | Kubernetes-style YAML, composable | Requires OpenMCF CLI |

### Target Type Comparison

| Target | Use Case | Auth | Payload |
|--------|----------|------|---------|
| **HTTP** | Cloud Run, Cloud Functions, any URL | OIDC/OAuth/None | Base64 body + headers |
| **Pub/Sub** | Event-driven pipelines, decoupled processing | IAM (publish permission) | Base64 data + attributes |
| **App Engine** | App Engine handlers in same project | App Engine IAM | Base64 body + headers |

**Recommendation**: For new deployments, HTTP target with OIDC authentication is the modern pattern. Pub/Sub is ideal for event-driven architectures. App Engine target is for legacy workloads.

## 80/20 Scoping Rationale

### What We Include (Covers 95%+ of Use Cases)

1. **All three target types** (HTTP, Pub/Sub, App Engine) -- core scheduling functionality
2. **OIDC and OAuth authentication** on HTTP target -- essential for secure Cloud Run/Functions invocation
3. **Retry configuration** -- critical for reliability in production
4. **Paused state control** -- operational necessity for staged deployments
5. **App Engine routing** (service/version/instance) -- complete App Engine integration
6. **Custom headers and body** -- enables real-world API integrations
7. **Attempt deadline** -- timeout control for long-running targets

### What We Exclude (Niche, Deferred to v2)

| Feature | Why Excluded |
|---------|-------------|
| **`deadline` field** | Deprecated alias for `attempt_deadline` |
| **Inline schema validation** | GCP validates cron expressions, timezones, and URLs at API level |
| **Job management operations** | Pause/resume/run-now are operational, not infrastructure definition |
| **Cross-project targets** | Rare; requires complex IAM; add if demanded |

## GCP API Constraints

### Immutable Fields
- `name` -- Job name cannot be changed (ForceNew in Terraform)
- `region` -- Region cannot be changed (ForceNew in Terraform)
- `project` -- Project cannot be changed (ForceNew in Terraform)

### Labels
Cloud Scheduler jobs **do not support GCP labels**. This is a GCP API limitation, not an OpenMCF decision.

### State Management
- `state` is **computed-only** (read from API, not settable)
- `paused` is the input mechanism: `true` creates the job paused, `false` (default) creates it enabled
- Terraform manages pause/resume via separate API endpoints (POST `:pause` / `:resume`)

### Target Mutual Exclusion
Exactly one of `http_target`, `pubsub_target`, or `app_engine_http_target` must be specified. The GCP API enforces this, and we additionally validate at the proto level with CEL.

### Authentication
- **OAuth token**: For calling Google APIs (*.googleapis.com). Generates an OAuth2 access token.
- **OIDC token**: For calling Cloud Run, Cloud Functions, or custom endpoints. Generates a JWT.
- Mutually exclusive within a single HTTP target.

### Body Encoding
The `body` field on HTTP and App Engine targets, and the `data` field on Pub/Sub targets, must be **base64-encoded**. This matches the GCP API and both Terraform/Pulumi providers.

### Attempt Deadline
- HTTP targets: 15 seconds to 30 minutes (default: 180s)
- App Engine targets: 15 seconds to 24 hours 15 seconds (default: 180s)
- Pub/Sub targets: field is **ignored** (setting it introduces unresolvable diff in Terraform)

## StringValueOrRef Design

Four `StringValueOrRef` fields enable infra-chart composability:

| Field | Default Kind | Field Path | Purpose |
|-------|-------------|------------|---------|
| `project_id` | GcpProject | `status.outputs.project_id` | Project reference |
| `http_target.oauth_token.service_account_email` | GcpServiceAccount | `status.outputs.email` | OAuth identity |
| `http_target.oidc_token.service_account_email` | GcpServiceAccount | `status.outputs.email` | OIDC identity |
| `pubsub_target.topic_name` | GcpPubSubTopic | `status.outputs.topic_id` | Pub/Sub topic reference |

## Corrections from T01 Plan

| Change | Rationale |
|--------|-----------|
| Added `job_name` | Consistent with R01-R17 naming pattern; GCP requires explicit name |
| Added `paused` boolean | Maps to TF `paused` field; operational control for ENABLED/PAUSED state |
| `region` -> `location` | Consistency with R17 (GcpCloudTasksQueue) |
| `body` fields documented as base64 | Both TF and Pulumi expect base64-encoded body |
| `pubsub_target.topic_name` as StringValueOrRef | Infra-chart composability with GcpPubSubTopic |
| OAuth/OIDC `service_account_email` as StringValueOrRef | Infra-chart composability with GcpServiceAccount |
| No GCP labels | Cloud Scheduler jobs don't support labels (GCP API limitation) |
| `app_engine_http_target` included | Core target type -- 1 of 3 fundamental scheduling targets |

## Infra Chart Composition

GcpCloudSchedulerJob composes naturally with:

- **GcpCloudRun** + **GcpServiceAccount**: Scheduler triggers authenticated Cloud Run endpoints
- **GcpPubSubTopic** + **GcpPubSubSubscription**: Scheduler publishes to topics consumed by subscribers
- **GcpCloudFunction** + **GcpServiceAccount**: Scheduler triggers Cloud Functions via HTTP
- **GcpBigQueryDataset**: Scheduler triggers ETL jobs that populate BigQuery

Typical infra chart pattern (serverless-api-backend):
```
GcpProject -> GcpServiceAccount -> GcpCloudRun + GcpCloudSchedulerJob(HTTP+OIDC)
```

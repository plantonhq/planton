# GCP Cloud Function

Deploys a Cloud Functions (Gen 2) function — source-based serverless compute built on Cloud Run and Eventarc. You ship a source archive; Cloud Build containerizes it with buildpacks and Cloud Run serves it, so every function is backed by a real Cloud Run service. For container-image workloads, use [GcpCloudRun](/docs/catalog/gcp/gcpcloudrun) directly; for batch, [GcpCloudRunJob](/docs/catalog/gcp/gcpcloudrunjob).

## What Gets Created

When you deploy a GcpCloudFunction resource, Planton provisions:

- **Cloud Functions (Gen 2) function** — a `google_cloudfunctions2_function`; the Cloud Functions, Cloud Build, Cloud Run, Artifact Registry, and Eventarc APIs are enabled automatically
- **A Cloud Build build** — containerizes the source archive at deploy time
- **The underlying Cloud Run service** — serves the function (its ID and URI are exported)
- **An Eventarc trigger** — for event-driven functions
- **A public-invoker IAM grant** — `run.invoker` for `allUsers` on the underlying service, only when `allowUnauthenticated: true`
- **Platform attribution** — organization, environment, and resource labels on the function object

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **A source archive in GCS** ([GcpGcsBucket](/docs/catalog/gcp/gcpgcsbucket)) — a .zip of the code and dependency manifest
- **A service account** ([GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount)) for least-privilege runtime identity
- **A Serverless VPC Access connector** ([GcpServerlessVpcConnector](/docs/catalog/gcp/gcpserverlessvpcconnector)) to reach private VPC resources
- **A Pub/Sub topic** ([GcpPubSubTopic](/docs/catalog/gcp/gcppubsubtopic)) for Pub/Sub event triggers

## Quick Start

Create a file `function.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudFunction
metadata:
  name: hello-api
spec:
  region: us-central1
  buildConfig:
    runtime: nodejs22
    entryPoint: helloHttp
    source:
      storageSource:
        bucket:
          value: my-source-bucket
        object: functions/hello-v1.zip
  serviceConfig:
    allowUnauthenticated: true
```

Deploy:

```shell
planton apply -f function.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Region the function is deployed in. Immutable. | Required |
| `buildConfig.runtime` | `string` | Gen 2 runtime, e.g. `python312`, `nodejs22`, `go123`. Any current runtime GCP publishes. | Required |
| `buildConfig.entryPoint` | `string` | Function name in source code. | Required |
| `buildConfig.source` | — | Exactly one of `storageSource` (GCS zip) or `repoSource` (Cloud Source Repositories). | Enforced pre-deploy |

### Identity & Metadata

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default | GCP project. Can reference GcpProject. |
| `functionName` | `string` | `metadata.name` | Function name in GCP (1–63 chars). Immutable. |
| `description` | `string` | — | Human-readable description. |
| `labels` | `map` | — | User labels; merged beneath platform attribution. |
| `kmsKeyName` | `StringValueOrRef` | — | CMEK for image + source artifacts. Can reference GcpKmsKey. Requires a customer-managed `dockerRepository`. |
| `deletionPolicy` | `string` | `DELETE` | `PREVENT` fails the destroy; `ABANDON` leaves the function serving and consuming events unmanaged. |

### Build

| Field | Type | Description |
|-------|------|-------------|
| `source.storageSource` | block | `bucket` (can reference GcpGcsBucket), `object`, optional `generation` pin. |
| `source.repoSource` | block | Cloud Source Repositories: `repoName` + exactly one of `branchName`/`tagName`/`commitSha` (CSR is deprecated for new customers — prefer GCS). |
| `buildEnvironmentVariables` | `map` | Build-time only (buildpack knobs). Plain text; there is no secret path for build variables. |
| `serviceAccount` | `StringValueOrRef` | Build identity — FULLY-QUALIFIED SA resource name. Can reference GcpServiceAccount. |
| `workerPool` | `StringValueOrRef` | Cloud Build private worker pool for private-perimeter builds (`projects/<p>/locations/<r>/workerPools/<name>`). Can reference GcpCloudBuildWorkerPool. |
| `dockerRepository` | `StringValueOrRef` | User-managed Artifact Registry repo for the built image (fully-qualified path). Required for CMEK. |
| `updatePolicy` | enum | `AUTOMATIC` (default: continuous runtime security updates) or `ON_DEPLOY` (pin until next deploy). |

### Service (runtime shape)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `serviceAccountEmail` | `StringValueOrRef` | Compute default SA | Runtime identity (bare email). Can reference GcpServiceAccount. |
| `availableMemory` | `string` | `256M` | Quantity string: `256M`, `1Gi`, `16Gi`. |
| `availableCpu` | `string` | derived from memory | e.g. `"1"`, `"0.5"`. Concurrency > 1 needs ≥ 1 CPU. |
| `timeoutSeconds` | `int` | 60 | Up to 3600 for HTTP; events cap at 540. |
| `maxInstanceRequestConcurrency` | `int` | 1 | Requests per instance (1–1000). |
| `environmentVariables` | `map` | — | Plain-text configuration. |
| `secretEnvironmentVariables` | `list` | — | Each entry is either a reference to an existing secret (`key`/`secret`/`version`/`projectId`) or a `value` the module stores in its own Secret Manager secret. Either way the function reads it natively at start. |
| `secretVolumes` | `list` | — | Secret versions projected as files under `mountPath`. |
| `vpcConnector` | `StringValueOrRef` | — | Serverless VPC Access connector for private egress. Can reference GcpServerlessVpcConnector. |
| `vpcConnectorEgressSettings` | enum | `PRIVATE_RANGES_ONLY` | Or `ALL_TRAFFIC` (static egress IPs via Cloud NAT). |
| `directVpcNetworkInterface` | object | — | Direct VPC egress — attach straight to a `network`/`subnetwork` (refs to GcpVpcNetwork/GcpSubnetwork) + firewall `tags`; no connector to size or pay for. Mutually exclusive with `vpcConnector`. |
| `directVpcEgress` | enum | `PRIVATE_RANGES_ONLY` | Or `ALL_TRAFFIC`; which egress takes the direct-VPC path. |
| `ingressSettings` | enum | `ALLOW_ALL` | `ALLOW_INTERNAL_ONLY`, `ALLOW_INTERNAL_AND_GCLB` (for LB fronting). |
| `scaling.minInstanceCount` | `int` | 0 | > 0 eliminates cold starts at idle cost. |
| `scaling.maxInstanceCount` | `int` | 100 | Cost / backpressure ceiling (≤ 3000). |
| `allTrafficOnLatestRevision` | `bool` | `true` | `false` holds traffic for manual canary/rollback on the underlying service. |
| `binaryAuthorizationPolicy` | `string` | — | Policy checked before instances start. |
| `allowUnauthenticated` | `bool` | `false` | Grants `run.invoker` to `allUsers` (HTTP functions). |

### Trigger

| Field | Type | Description |
|-------|------|-------------|
| `triggerType` | enum | `HTTP` (default) or `EVENT_TRIGGER`. |
| `eventTrigger.eventType` | `string` | CloudEvents type, e.g. `google.cloud.pubsub.topic.v1.messagePublished`, `google.cloud.storage.object.v1.finalized`. |
| `eventTrigger.pubsubTopic` | `StringValueOrRef` | Topic for Pub/Sub triggers. Can reference GcpPubSubTopic. |
| `eventTrigger.eventFilters` | `list` | `attribute`/`value` (+ optional `match-path-pattern` operator). |
| `eventTrigger.triggerRegion` | `string` | Defaults to the function's region; multi-region sources use `us`/`eu`. |
| `eventTrigger.retryPolicy` | enum | `RETRY_POLICY_DO_NOT_RETRY` (default) or `RETRY_POLICY_RETRY` (at-least-once; idempotent handlers). |
| `eventTrigger.serviceAccountEmail` | `StringValueOrRef` | Eventarc's invoking identity (needs `run.invoker`). |

## Outputs

| Output | Description |
|--------|-------------|
| `functionId` | Fully qualified resource name |
| `name` | Bare function name — the handle serverless NEGs reference |
| `functionUrl` | HTTPS URL of the function |
| `uri` | URI of the underlying Cloud Run service (`*.run.app`) |
| `cloudRunServiceId` | The underlying Cloud Run service |
| `serviceAccountEmail` | Runtime identity |
| `eventarcTriggerId` | Eventarc trigger (event functions) |
| `state` | `ACTIVE`, `OFFLINE`, ... |
| `environment` | e.g. `GEN_2` |
| `updateTime` | Last update timestamp |

## Presets

- [HTTP API — basic](presets/01-http-api.yaml)
- [Pub/Sub event processor](presets/02-pubsub-event.yaml)
- [Private VPC egress — database-backed](presets/03-private-vpc-egress.yaml)

## See Also

- [GcpServerlessVpcConnector](/docs/catalog/gcp/gcpserverlessvpcconnector) — private VPC egress
- [GcpRegionNetworkEndpointGroup](/docs/catalog/gcp/gcpregionnetworkendpointgroup) — put the function behind an external Application Load Balancer
- [GcpCloudRun](/docs/catalog/gcp/gcpcloudrun) / [GcpCloudRunJob](/docs/catalog/gcp/gcpcloudrunjob) — container-image serving and batch siblings

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

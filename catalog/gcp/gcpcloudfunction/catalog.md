# GCP Cloud Function

Deploys a Cloud Functions (Gen 2) function — source-based serverless compute built on Cloud Run and Eventarc. You ship a source archive; Cloud Build containerizes it with buildpacks and Cloud Run serves it, so every function is backed by a real Cloud Run service. The spec mirrors the API's two-config split: `buildConfig` owns how the source becomes a container, `serviceConfig` owns how it runs, and the trigger decides what invokes it — HTTPS requests or a CloudEvent delivered by Eventarc. The spec wires by reference to other Planton resources at every seam, from the GCP project down to the Pub/Sub topic that triggers it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Cloud Functions (Gen 2) function** -- a `google_cloudfunctions2_function` in the specified project and region, with the chosen runtime, entry point, resources, scaling bounds, and timeout
- **A Cloud Build build** -- containerizes the source archive (or Cloud Source Repositories revision) at deploy time, optionally as a custom identity, in a private worker pool, into a customer-managed Artifact Registry repository
- **The underlying Cloud Run service** -- serves the function; its ID and `*.run.app` URI are exported
- **An Eventarc trigger** -- created only when `trigger.triggerType` is `EVENT_TRIGGER`; wires the function to Pub/Sub, Cloud Storage, Firestore, or audit-log events with filters and a retry policy
- **A public-invoker IAM grant** -- created only when `serviceConfig.allowUnauthenticated` is true; grants `roles/run.invoker` to `allUsers` on the underlying Cloud Run service
- **API enablement** -- the Cloud Functions, Cloud Build, Cloud Run, Artifact Registry, and Eventarc APIs are enabled on the project before the function deploys, and never disabled on destroy
- **Platform attribution** -- organization, environment, and resource labels merged beneath your own `labels`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the function is created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef. The Cloud Functions, Cloud Build, Cloud Run, Artifact Registry, and Eventarc APIs are enabled automatically.
- **A source archive in GCS** -- a .zip of the code and dependency manifest, readable by the build identity. Reference a GcpGcsBucket Infra Component for the bucket.
- **A Pub/Sub topic** (for messagePublished triggers) -- reference a GcpPubSubTopic Infra Component.
- **A Serverless VPC Access connector** (to reach private resources) -- reference a GcpServerlessVpcConnector Infra Component; it must live in the function's region.
- **A customer-managed Artifact Registry repository** (required for CMEK) -- reference a GcpArtifactRegistryRepo Infra Component.

## Deploy

### Console

Open the deployment store, find **GCP Cloud Function**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **HTTP API — basic** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudFunction
metadata:
  name: hello-api
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  region: us-central1
  buildConfig:
    runtime: nodejs22
    entryPoint: helloHttp
    source:
      storageSource:
        bucket:
          value: acme-functions-source
        object: functions/hello-api-v1.0.0.zip
  serviceConfig:
    allowUnauthenticated: true
```

```shell
planton apply -f cloud-function.yaml
```

This creates a public HTTP function on Node.js 22 with GCP's defaults — 256M memory, derived CPU, a 60-second timeout, and scale-to-zero. Shipping a new release is uploading a new archive and pointing `object` at it: a changed object name is what makes the deploy roll. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the function to resources deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  buildConfig:
    source:
      storageSource:
        bucket:
          valueFrom:
            kind: GcpGcsBucket
            name: functions-source
            fieldPath: status.outputs.bucket_id
        object: functions/hello-api-v1.0.0.zip
```

The InfraPipeline resolves the dependency graph, deploys the project and bucket first, then provisions the function with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Cloud Function. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Trigger type** -- `trigger.triggerType: HTTP` (the default) gives the function a URL; `EVENT_TRIGGER` wires it to Eventarc and requires the `eventTrigger` block (the CloudEvents `eventType`, a `pubsubTopic` for messagePublished, `eventFilters` for Storage/Firestore/audit-log sources, and a `retryPolicy` — retry is at-least-once, so handlers must be idempotent). Event-driven functions are capped at a 540-second timeout by Eventarc's delivery window.

**Source** -- exactly one of `storageSource` (a GCS zip archive — the standard CI/CD path; pin an exact object version with `generation`) or `repoSource` (a Cloud Source Repositories revision, pinned to exactly one of branch/tag/commit — note GCP deprecated CSR for new customers in June 2024).

**Runtime and entry point** -- `buildConfig.runtime` accepts any Gen 2 runtime GCP currently publishes (`gcloud functions runtimes list`); `buildConfig.entryPoint` names the handler function in your source, in the language's own casing.

**Resources and scaling** -- `serviceConfig.availableMemory` is a quantity string (`256M`, `1Gi`); CPU derives from memory unless `availableCpu` sets it explicitly — and concurrency above 1 requires at least 1 CPU. `scaling.minInstanceCount` above 0 eliminates cold starts at idle cost; `maxInstanceCount` is the cost and backpressure ceiling for event storms.

**Secrets** -- `secretEnvironmentVariables` entries either name an existing Secret Manager secret or carry a `value`. A `value` is stored by the module in its own secret, in the function's region, with access granted to the runtime identity on that secret alone; GCP resolves it at instance start. Named secrets and `secretVolumes` need `roles/secretmanager.secretAccessor` granted by you.

**Access control** -- `allowUnauthenticated: true` grants public invocation (deliberate for webhooks); `ingressSettings` restricts network-level reachability (`ALLOW_INTERNAL_ONLY` for event consumers, `ALLOW_INTERNAL_AND_GCLB` when fronting with an external Application Load Balancer). The runtime `serviceAccountEmail` is the identity your code exercises — production functions get a dedicated least-privilege GcpServiceAccount.

**Private egress** -- two mutually exclusive paths into a VPC. `vpcConnector` routes egress through a Serverless VPC Access connector to reach Cloud SQL private IP, Memorystore, and internal load balancers. `directVpcNetworkInterface` attaches the function straight to a network or subnet — no connector to size, pay for, or saturate — but instances draw IPs from the subnet, so its free range caps how far the function scales. On either path, the `ALL_TRAFFIC` egress setting enables static egress IPs via Cloud NAT.

**Build hardening** -- a custom build `serviceAccount` (fully-qualified resource name), a private `workerPool`, a customer-managed `dockerRepository`, and `updatePolicy: ON_DEPLOY` to pin the runtime between deploys. `kmsKeyName` (CMEK) encrypts the image and source artifacts and REQUIRES the customer-managed `dockerRepository`.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpKmsKey** | `kmsKeyName` | `status.outputs.key_id` |
| **GcpServiceAccount** | `buildConfig.serviceAccount` | `status.outputs.name` |
| **GcpArtifactRegistryRepo** | `buildConfig.dockerRepository` | `status.outputs.repository_path` |
| **GcpGcsBucket** | `buildConfig.source.storageSource.bucket` | `status.outputs.bucket_id` |
| **GcpServiceAccount** | `serviceConfig.serviceAccountEmail` | `status.outputs.email` |
| **GcpServerlessVpcConnector** | `serviceConfig.vpcConnector` | `status.outputs.self_link` |
| **GcpVpcNetwork** | `serviceConfig.directVpcNetworkInterface.network` | `status.outputs.network_name` |
| **GcpSubnetwork** | `serviceConfig.directVpcNetworkInterface.subnetwork` | `status.outputs.subnetwork_name` |
| **GcpPubSubTopic** | `trigger.eventTrigger.pubsubTopic` | `status.outputs.topic_id` |
| **GcpServiceAccount** | `trigger.eventTrigger.serviceAccountEmail` | `status.outputs.email` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `function_id` | Fully qualified function resource name | IAM bindings, monitoring filters |
| `name` | Bare function name (the last segment of `function_id`) | Serverless NEG (`GcpRegionNetworkEndpointGroup`) references, gcloud commands |
| `function_url` | HTTPS URL of the function (HTTP triggers only) | Webhook registration, Cloud Scheduler targets |
| `uri` | URI of the underlying Cloud Run service (`*.run.app`) — populated for every Gen 2 function | Service-to-service calls, uptime checks |
| `cloud_run_service_id` | The underlying Cloud Run service | Traffic splitting (canary/rollback), Cloud Run API access |
| `service_account_email` | The runtime identity | IAM policy grants, secret accessor bindings |
| `eventarc_trigger_id` | The Eventarc trigger (event functions only) | Event routing configuration, monitoring |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**HTTP API** -- a public HTTPS endpoint for webhooks, small REST APIs, and glue code, built from a versioned GCS archive. Start from the **HTTP API — basic** preset.

**Pub/Sub event processor** -- an event-driven worker consuming a topic through Eventarc, with internal-only ingress and a bounded instance ceiling as backpressure. Start from the **Pub/Sub event processor** preset.

**Private VPC egress** -- a function reaching a private-IP database through a Serverless VPC Access connector, running as a dedicated least-privilege service account, reading its credential from Secret Manager at instance start. Start from the **Private VPC egress — database-backed function** preset.

**Behind a load balancer** -- set `ingressSettings: ALLOW_INTERNAL_AND_GCLB` and reference the function's `name` output from a serverless network endpoint group to front it with an external Application Load Balancer.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the project the function is created in
- [**GCP GCS Bucket**](/infra-catalog/gcp-gcs-bucket) -- holds the versioned source archives
- [**GCP Pub/Sub Topic**](/infra-catalog/gcp-pub-sub-topic) -- the event source for messagePublished triggers
- [**GCP Serverless VPC Connector**](/infra-catalog/gcp-serverless-vpc-connector) -- private VPC egress through a connector
- [**GCP VPC Network**](/infra-catalog/gcp-vpc-network) / [**GCP Subnetwork**](/infra-catalog/gcp-subnetwork) -- direct VPC egress without a connector
- [**GCP Artifact Registry Repo**](/infra-catalog/gcp-artifact-registry-repo) -- customer-managed image storage (required for CMEK)
- [**GCP Region Network Endpoint Group**](/infra-catalog/gcp-region-network-endpoint-group) -- puts the function behind an external Application Load Balancer
- [**GCP Cloud Run**](/infra-catalog/gcp-cloud-run) / [**GCP Cloud Run Job**](/infra-catalog/gcp-cloud-run-job) -- the container-image serving and batch siblings

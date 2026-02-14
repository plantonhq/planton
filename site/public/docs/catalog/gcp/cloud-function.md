---
title: "Cloud Function"
description: "Cloud Function deployment documentation"
icon: "package"
order: 100
componentName: "gcpcloudfunction"
---

# GCP Cloud Function

Deploys a Google Cloud Function (Gen 2) from source code stored in a GCS bucket, with full control over runtime, compute resources, scaling, networking, secrets, and trigger configuration. Gen 2 functions are built on Cloud Run and Eventarc, supporting both HTTP and event-driven invocation patterns.

## What Gets Created

When you deploy a GcpCloudFunction resource, OpenMCF provisions:

- **Cloud Function (Gen 2)** — a `cloudfunctionsv2.Function` in the specified project and region, built from a source archive in GCS, with the configured runtime, entry point, service config, and labels applied
- **Service Configuration** — compute resources (memory, timeout, concurrency), environment variables, secret references, VPC connector, ingress/egress settings, and scaling limits applied to the underlying Cloud Run service
- **Event Trigger** — an Eventarc trigger created when `trigger.triggerType` is `EVENT_TRIGGER`, with event type, filters, Pub/Sub topic, retry policy, and trigger service account configured
- **IAM Policy Binding** — a `cloudrunv2.ServiceIamMember` granting `roles/run.invoker` to `allUsers` when `serviceConfig.allowUnauthenticated` is `true` (HTTP triggers only)

## Prerequisites

- **GCP credentials** configured via environment variables or OpenMCF provider config
- **A GCP project** where the Cloud Function will be created
- **A GCS bucket** containing the source code archive (`.zip` file with function code and dependencies)
- **Cloud Functions API and Cloud Build API** enabled in the target project
- **A VPC connector** if connecting the function to a VPC network (optional)
- **Secret Manager secrets** if using `secretEnvironmentVariables` (optional)

## Quick Start

Create a file `cloud-function.yaml`:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudFunction
metadata:
  name: my-http-handler
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpCloudFunction.my-http-handler
spec:
  projectId: my-gcp-project-123
  region: us-central1
  buildConfig:
    runtime: python312
    entryPoint: hello_http
    source:
      bucket: my-functions-source
      object: functions/hello-v1.0.0.zip
```

Deploy:

```shell
openmcf apply -f cloud-function.yaml
```

This creates a Gen 2 Cloud Function with an HTTP trigger, 256 MB memory, 60-second timeout, and default scaling (0-100 instances).

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `string` or `valueFrom` | GCP project ID where the function is created. Can be a literal value or a reference to a GcpProject resource. | Required |
| `region` | `string` | Region where the function is deployed (e.g., `us-central1`, `europe-west1`). | Required, pattern `^[a-z]+-[a-z]+[0-9]$` |
| `buildConfig` | `GcpCloudFunctionBuildConfig` | Build configuration for the function. | Required |
| `buildConfig.runtime` | `string` | Runtime environment. Examples: `python312`, `nodejs22`, `go122`, `java21`, `dotnet8`. | Required, must be a supported Gen 2 runtime |
| `buildConfig.entryPoint` | `string` | Name of the function in source code that is executed. | Required, 1-128 chars |
| `buildConfig.source` | `GcpCloudFunctionSource` | Source code location in GCS. | Required |
| `buildConfig.source.bucket` | `string` | GCS bucket name containing the source archive. | Required, pattern `^[a-z0-9][a-z0-9._-]{1,61}[a-z0-9]$` |
| `buildConfig.source.object` | `string` | Object path of the source archive in the bucket (e.g., `functions/my-func-v1.zip`). | Required, min 1 char |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `functionName` | `string` | `metadata.name` | Override the Cloud Function name. Must be 1-63 chars, lowercase, start with a letter. |
| `buildConfig.source.generation` | `int64` | latest | Specific GCS object generation number. If omitted, uses the latest version. |
| `buildConfig.buildEnvironmentVariables` | `map<string, string>` | `{}` | Environment variables available during the build process. |
| `serviceConfig` | `GcpCloudFunctionServiceConfig` | see defaults | Runtime service configuration. |
| `serviceConfig.serviceAccountEmail` | `string` | default compute SA | Service account the function runs as. Format: `{id}@{project}.iam.gserviceaccount.com`. |
| `serviceConfig.availableMemoryMb` | `int32` | `256` | Memory in MB. Valid: 128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768. |
| `serviceConfig.timeoutSeconds` | `int32` | `60` | Execution timeout in seconds. Range: 1-3600. |
| `serviceConfig.maxInstanceRequestConcurrency` | `int32` | `80` | Max concurrent requests per instance. Range: 1-1000. |
| `serviceConfig.environmentVariables` | `map<string, string>` | `{}` | Plain-text environment variables injected at runtime. |
| `serviceConfig.secretEnvironmentVariables` | `map<string, string>` | `{}` | Secret Manager references as `KEY: secret_name`. Version `latest` is used automatically. |
| `serviceConfig.vpcConnector` | `string` | `""` | VPC connector resource path. Format: `projects/{project}/locations/{region}/connectors/{name}`. |
| `serviceConfig.vpcConnectorEgressSettings` | `enum` | `PRIVATE_RANGES_ONLY` | VPC egress routing. One of: `PRIVATE_RANGES_ONLY`, `ALL_TRAFFIC`. |
| `serviceConfig.ingressSettings` | `enum` | `ALLOW_ALL` | Ingress control. One of: `ALLOW_ALL`, `ALLOW_INTERNAL_ONLY`, `ALLOW_INTERNAL_AND_GCLB`. |
| `serviceConfig.scaling` | `GcpCloudFunctionScalingConfig` | `{}` | Scaling limits. |
| `serviceConfig.scaling.minInstanceCount` | `int32` | `0` | Minimum warm instances. Range: 0-100. Set > 0 to eliminate cold starts. |
| `serviceConfig.scaling.maxInstanceCount` | `int32` | `100` | Maximum instances. Range: 1-3000. |
| `serviceConfig.allowUnauthenticated` | `bool` | `false` | If true, grants `roles/run.invoker` to `allUsers` (HTTP triggers only). |
| `trigger` | `GcpCloudFunctionTrigger` | HTTP | Trigger configuration. Defaults to HTTP trigger if omitted. |
| `trigger.triggerType` | `enum` | `HTTP` | Trigger type. One of: `HTTP`, `EVENT_TRIGGER`. |
| `trigger.eventTrigger` | `GcpCloudFunctionEventTrigger` | — | Event trigger config. Required when `triggerType` is `EVENT_TRIGGER`. |
| `trigger.eventTrigger.eventType` | `string` | — | CloudEvents event type (e.g., `google.cloud.pubsub.topic.v1.messagePublished`). Required for event triggers. |
| `trigger.eventTrigger.pubsubTopic` | `string` | `""` | Pub/Sub topic resource name. Format: `projects/{project}/topics/{topic}`. |
| `trigger.eventTrigger.eventFilters` | `GcpCloudFunctionEventFilter[]` | `[]` | Event attribute filters. |
| `trigger.eventTrigger.eventFilters[].attribute` | `string` | — | Attribute name (e.g., `bucket` for Storage events). |
| `trigger.eventTrigger.eventFilters[].value` | `string` | — | Value to match. Supports wildcards. |
| `trigger.eventTrigger.eventFilters[].operator` | `string` | exact match | Match operator. Use `match-path-pattern` for Firestore document paths. |
| `trigger.eventTrigger.triggerRegion` | `string` | same as `region` | Region where the event trigger listens. |
| `trigger.eventTrigger.retryPolicy` | `enum` | `RETRY_POLICY_DO_NOT_RETRY` | Retry behavior. One of: `RETRY_POLICY_DO_NOT_RETRY`, `RETRY_POLICY_RETRY`. |
| `trigger.eventTrigger.serviceAccountEmail` | `string` | default Eventarc SA | Service account used by Eventarc to invoke the function. |

## Examples

### HTTP Function with Custom Memory and Timeout

A Python function with increased memory, a longer timeout, and a dedicated service account:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudFunction
metadata:
  name: image-resizer
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpCloudFunction.image-resizer
spec:
  projectId: my-gcp-project-123
  region: us-central1
  buildConfig:
    runtime: python312
    entryPoint: resize_image
    source:
      bucket: my-functions-source
      object: functions/image-resizer-v2.1.0.zip
  serviceConfig:
    serviceAccountEmail: image-resizer@my-gcp-project-123.iam.gserviceaccount.com
    availableMemoryMb: 1024
    timeoutSeconds: 300
    maxInstanceRequestConcurrency: 10
    environmentVariables:
      OUTPUT_BUCKET: processed-images-dev
      MAX_WIDTH: "1920"
    scaling:
      minInstanceCount: 1
      maxInstanceCount: 50
```

### Pub/Sub Event-Driven Function

A Node.js function triggered by messages published to a Pub/Sub topic, with retry enabled:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudFunction
metadata:
  name: order-processor
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpCloudFunction.order-processor
spec:
  projectId: my-gcp-project-123
  region: us-central1
  buildConfig:
    runtime: nodejs22
    entryPoint: processOrder
    source:
      bucket: my-functions-source
      object: functions/order-processor-v3.0.1.zip
  serviceConfig:
    serviceAccountEmail: order-processor@my-gcp-project-123.iam.gserviceaccount.com
    availableMemoryMb: 512
    timeoutSeconds: 120
    environmentVariables:
      DB_HOST: 10.0.0.5
    secretEnvironmentVariables:
      DB_PASSWORD: order-db-password
    scaling:
      minInstanceCount: 2
      maxInstanceCount: 200
  trigger:
    triggerType: EVENT_TRIGGER
    eventTrigger:
      eventType: google.cloud.pubsub.topic.v1.messagePublished
      pubsubTopic: projects/my-gcp-project-123/topics/new-orders
      retryPolicy: RETRY_POLICY_RETRY
```

### Cloud Storage Event Function with VPC and Secrets

A Go function triggered when objects are created in a GCS bucket, connected to a VPC for private database access, with secrets from Secret Manager:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudFunction
metadata:
  name: file-indexer
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpCloudFunction.file-indexer
spec:
  projectId: my-gcp-project-123
  region: us-central1
  functionName: prod-file-indexer
  buildConfig:
    runtime: go122
    entryPoint: IndexFile
    source:
      bucket: my-functions-source
      object: functions/file-indexer-v1.4.0.zip
      generation: 1708012345678
    buildEnvironmentVariables:
      GOOGLE_BUILDPACKS_GO_BUILD_FLAGS: "-ldflags=-s -w"
  serviceConfig:
    serviceAccountEmail: file-indexer@my-gcp-project-123.iam.gserviceaccount.com
    availableMemoryMb: 2048
    timeoutSeconds: 540
    maxInstanceRequestConcurrency: 1
    environmentVariables:
      INDEX_TABLE: file_metadata
      LOG_LEVEL: info
    secretEnvironmentVariables:
      DB_CONNECTION_STRING: file-indexer-db-conn
      API_KEY: file-indexer-api-key
    vpcConnector: projects/my-gcp-project-123/locations/us-central1/connectors/main-connector
    vpcConnectorEgressSettings: PRIVATE_RANGES_ONLY
    ingressSettings: ALLOW_INTERNAL_ONLY
    scaling:
      minInstanceCount: 0
      maxInstanceCount: 500
    allowUnauthenticated: false
  trigger:
    triggerType: EVENT_TRIGGER
    eventTrigger:
      eventType: google.cloud.storage.object.v1.finalized
      eventFilters:
        - attribute: bucket
          value: incoming-documents-prod
      retryPolicy: RETRY_POLICY_RETRY
      serviceAccountEmail: eventarc-trigger@my-gcp-project-123.iam.gserviceaccount.com
```

### Public HTTP Function with Foreign Key Reference

A publicly accessible function that references an OpenMCF-managed GcpProject for the project ID:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudFunction
metadata:
  name: webhook-receiver
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpCloudFunction.webhook-receiver
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      fieldPath: status.outputs.project_id
  region: europe-west1
  buildConfig:
    runtime: java21
    entryPoint: com.example.WebhookHandler
    source:
      bucket: my-functions-source
      object: functions/webhook-receiver-v2.0.0.zip
  serviceConfig:
    availableMemoryMb: 512
    timeoutSeconds: 30
    maxInstanceRequestConcurrency: 200
    ingressSettings: ALLOW_ALL
    scaling:
      minInstanceCount: 1
      maxInstanceCount: 100
    allowUnauthenticated: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `functionId` | `string` | Fully qualified resource name of the function. Format: `projects/{project}/locations/{region}/functions/{name}` |
| `functionUrl` | `string` | HTTPS URL of the function. Only populated for HTTP-triggered functions; empty for event-driven functions. |
| `serviceAccountEmail` | `string` | Email of the service account the function runs as. |
| `state` | `string` | Current state of the function. Possible values: `ACTIVE`, `OFFLINE`, `DEPLOY_IN_PROGRESS`, `DELETE_IN_PROGRESS`, `UNKNOWN`. |
| `cloudRunServiceId` | `string` | Cloud Run service name backing this Gen 2 function. Format: `projects/{project}/locations/{region}/services/{name}` |
| `eventarcTriggerId` | `string` | Eventarc trigger ID. Only populated for event-driven functions; empty for HTTP-triggered functions. |

## Related Components

- [GcpProject](/docs/catalog/gcp/project) — provides the GCP project where the function is created
- [GcpServiceAccount](/docs/catalog/gcp/service-account) — creates service accounts for `serviceConfig.serviceAccountEmail` and event trigger identity
- [GcpSecretsManager](/docs/catalog/gcp/secrets-manager) — manages secrets referenced in `serviceConfig.secretEnvironmentVariables`
- [GcpGcsBucket](/docs/catalog/gcp/gcs-bucket) — stores the function source code archive and can be an event trigger source
- [GcpVpc](/docs/catalog/gcp/vpc) — network for VPC connector used by `serviceConfig.vpcConnector`

# GcpCloudFunction API - Examples

Here are examples of how to create and configure a **GcpCloudFunction** API resource using the Planton CLI. The examples cover basic HTTP functions, event-driven functions, and production configurations with VPC, secrets, and scaling.

## Create using CLI

First, create a YAML file using the examples provided below. After the YAML file is created, you can apply the configuration using the following command:

```shell
planton apply -f <yaml-path>
```

## Basic HTTP Function (Development)

This basic example demonstrates a simple HTTP-triggered Cloud Function for development environments. It uses default settings for memory (256MB), timeout (60s), and scaling.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudFunction
metadata:
  name: hello-http-dev
spec:
  projectId:
    value: my-gcp-project
  region: us-central1
  buildConfig:
    runtime: python311
    entryPoint: hello_http
    source:
      bucket: my-code-bucket
      object: functions/hello-http-v1.0.0.zip
```

## HTTP Function with Project Reference (ValueFrom)

This example demonstrates using a reference to a GcpProject resource instead of a hardcoded project ID. This enables cross-resource dependencies where the project ID is dynamically resolved from another resource's outputs.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudFunction
metadata:
  name: hello-http-with-ref
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: main-project
      fieldPath: status.outputs.project_id
  region: us-central1
  buildConfig:
    runtime: python311
    entryPoint: hello_http
    source:
      bucket: my-code-bucket
      object: functions/hello-http-v1.0.0.zip
```

## HTTP Function with Environment Variables

This example shows an HTTP function with custom configuration: increased memory, custom timeout, and environment variables for non-sensitive configuration.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudFunction
metadata:
  name: api-gateway-dev
spec:
  projectId:
    value: my-gcp-project
  region: us-central1
  buildConfig:
    runtime: nodejs20
    entryPoint: handleRequest
    source:
      bucket: my-code-bucket
      object: functions/api-gateway-v2.1.3.zip
  serviceConfig:
    availableMemoryMb: 512
    timeoutSeconds: 120
    environmentVariables:
      LOG_LEVEL: "debug"
      API_VERSION: "v2"
      ENVIRONMENT: "development"
```

## Production HTTP Function with Secrets and VPC

This example demonstrates a production-ready HTTP function with:
- Secrets from Google Secret Manager
- VPC connectivity for private database access
- Custom service account for least-privilege access
- Min instances for warm starts (eliminates cold starts)
- Public access enabled

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudFunction
metadata:
  name: api-gateway-prod
  env: prod
  org: my-org
spec:
  projectId:
    value: my-gcp-project-prod
  region: us-east1
  buildConfig:
    runtime: nodejs22
    entryPoint: handleRequest
    source:
      bucket: my-prod-code-bucket
      object: functions/api-gateway-v3.4.1.zip
  serviceConfig:
    serviceAccountEmail: api-gateway@my-gcp-project-prod.iam.gserviceaccount.com
    availableMemoryMb: 1024
    timeoutSeconds: 60
    maxInstanceRequestConcurrency: 80
    environmentVariables:
      LOG_LEVEL: "info"
      API_VERSION: "v3"
      ENVIRONMENT: "production"
    secretEnvironmentVariables:
      JWT_SECRET: jwt-secret
      DB_CONNECTION_STRING: db-connection
      API_KEY: external-api-key
    vpcConnector: projects/my-gcp-project-prod/locations/us-east1/connectors/vpc-connector
    vpcConnectorEgressSettings: 0  # PRIVATE_RANGES_ONLY
    ingressSettings: 0  # ALLOW_ALL (public)
    scaling:
      minInstanceCount: 2  # 2 warm instances for HA
      maxInstanceCount: 100
    allowUnauthenticated: true
```

## Private HTTP Function (Internal Only)

This example shows a private HTTP function accessible only from within the VPC or same GCP project. Public access is disabled.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudFunction
metadata:
  name: internal-api
spec:
  projectId:
    value: my-gcp-project
  region: us-central1
  buildConfig:
    runtime: python312
    entryPoint: process_request
    source:
      bucket: my-code-bucket
      object: functions/internal-api-v1.2.0.zip
  serviceConfig:
    serviceAccountEmail: internal-api@my-gcp-project.iam.gserviceaccount.com
    availableMemoryMb: 512
    timeoutSeconds: 180
    vpcConnector: projects/my-gcp-project/locations/us-central1/connectors/vpc-connector
    ingressSettings: 1  # ALLOW_INTERNAL_ONLY (private)
    allowUnauthenticated: false
```

## Event-Driven Function (Pub/Sub Trigger)

This example demonstrates an event-driven function triggered by Pub/Sub messages. It processes asynchronous jobs from a Pub/Sub topic.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudFunction
metadata:
  name: pubsub-worker
spec:
  projectId:
    value: my-gcp-project
  region: us-central1
  buildConfig:
    runtime: go122
    entryPoint: ProcessMessage
    source:
      bucket: my-code-bucket
      object: functions/pubsub-worker-v1.0.0.zip
  serviceConfig:
    serviceAccountEmail: pubsub-worker@my-gcp-project.iam.gserviceaccount.com
    availableMemoryMb: 512
    timeoutSeconds: 300
    secretEnvironmentVariables:
      DB_PASSWORD: database-password
    vpcConnector: projects/my-gcp-project/locations/us-central1/connectors/vpc-connector
    scaling:
      minInstanceCount: 0
      maxInstanceCount: 50
  trigger:
    triggerType: 1  # EVENT_TRIGGER
    eventTrigger:
      eventType: google.cloud.pubsub.topic.v1.messagePublished
      pubsubTopic: projects/my-gcp-project/topics/job-queue
      triggerRegion: us-central1
      retryPolicy: 1  # RETRY_POLICY_RETRY (automatic retries)
```

## Event-Driven Function (Cloud Storage Trigger)

This example shows a function triggered by Cloud Storage events (object creation). It processes new files uploaded to a GCS bucket.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudFunction
metadata:
  name: image-processor
spec:
  projectId:
    value: my-gcp-project
  region: us-central1
  buildConfig:
    runtime: python311
    entryPoint: process_image
    source:
      bucket: my-code-bucket
      object: functions/image-processor-v2.0.0.zip
  serviceConfig:
    serviceAccountEmail: image-processor@my-gcp-project.iam.gserviceaccount.com
    availableMemoryMb: 2048
    timeoutSeconds: 540  # 9 minutes for image processing
    environmentVariables:
      OUTPUT_BUCKET: processed-images
      IMAGE_FORMAT: "webp"
    scaling:
      minInstanceCount: 0
      maxInstanceCount: 10
  trigger:
    triggerType: 1  # EVENT_TRIGGER
    eventTrigger:
      eventType: google.cloud.storage.object.v1.finalized
      eventFilters:
        - attribute: bucket
          value: upload-bucket
      triggerRegion: us-central1
      retryPolicy: 0  # RETRY_POLICY_DO_NOT_RETRY (at-most-once)
```

## Notes

- **Project ID**: Supports both literal values (`projectId: {value: "my-project"}`) and references to other resources (`projectId: {valueFrom: {kind: GcpProject, name: "main-project", fieldPath: "status.outputs.project_id"}}`)
- **Runtimes**: Use current, non-deprecated runtimes (python311+, nodejs20+, go121+, etc.)
- **Source Code**: Must be uploaded to a GCS bucket as a ZIP file
- **Service Account**: Always use a dedicated service account with least-privilege permissions for production
- **Secrets**: Use `secret_environment_variables` for sensitive data like API keys, passwords, connection strings
- **VPC**: Use `vpc_connector` to access private resources (Cloud SQL, Memorystore, internal APIs)
- **Scaling**: Set `min_instance_count > 0` to eliminate cold starts for latency-sensitive workloads
- **Triggers**: HTTP (default, trigger_type=0) or Event-driven (trigger_type=1)
- **Authentication**: Set `allow_unauthenticated: true` only for public APIs; keep false for private functions

These examples demonstrate various configurations for deploying Cloud Functions Gen 2 on GCP, from simple HTTP endpoints to complex event-driven workflows.

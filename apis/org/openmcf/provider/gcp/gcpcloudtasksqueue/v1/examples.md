# GcpCloudTasksQueue Examples

## Minimal Queue

A basic queue with default settings. Cloud Tasks provides reasonable defaults for rate limits and retry behavior.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudTasksQueue
metadata:
  name: basic-queue
spec:
  projectId:
    value: my-gcp-project
  queueName: basic-task-queue
  location: us-central1
```

## Rate-Limited Background Processing

Queue with explicit rate limits and retry configuration for reliable background processing.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudTasksQueue
metadata:
  name: background-processing
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
    maxRetryDuration: "86400s"
  stackdriverLoggingConfig:
    samplingRatio: 0.1
```

## Cloud Run Target with OIDC Authentication

Queue configured to dispatch all tasks to a Cloud Run service with automatic OIDC token generation. This is the recommended pattern for microservices.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudTasksQueue
metadata:
  name: cloud-run-tasks
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
```

## Google API Target with OAuth Token

Queue targeting a Google API endpoint with OAuth2 authentication.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudTasksQueue
metadata:
  name: google-api-tasks
spec:
  projectId:
    value: my-gcp-project
  queueName: api-dispatcher
  location: us-east1
  httpTarget:
    httpMethod: POST
    oauthToken:
      serviceAccountEmail:
        value: api-caller@my-gcp-project.iam.gserviceaccount.com
      scope: https://www.googleapis.com/auth/cloud-platform
    uriOverride:
      scheme: HTTPS
      host: bigquery.googleapis.com
  retryConfig:
    maxAttempts: 3
    minBackoff: "5s"
    maxBackoff: "300s"
    maxDoublings: 4
```

## Paused Queue

Queue created in PAUSED state. Tasks can be enqueued but won't be dispatched until the queue is resumed.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudTasksQueue
metadata:
  name: maintenance-queue
spec:
  projectId:
    value: my-gcp-project
  queueName: maintenance-tasks
  location: us-central1
  desiredState: PAUSED
  rateLimits:
    maxDispatchesPerSecond: 10
    maxConcurrentDispatches: 5
```

## Infra-Chart Composition with ValueFrom

Queue that references a GCP project and service account from other OpenMCF resources, enabling dependency-aware deployment.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudTasksQueue
metadata:
  name: composed-queue
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      fieldPath: status.outputs.project_id
  queueName: composed-task-queue
  location: us-central1
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
  rateLimits:
    maxDispatchesPerSecond: 200
    maxConcurrentDispatches: 50
```

## Unlimited Retries

Queue with unlimited retry attempts for critical tasks that must eventually succeed.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudTasksQueue
metadata:
  name: critical-tasks
spec:
  projectId:
    value: my-gcp-project
  queueName: critical-task-queue
  location: us-central1
  retryConfig:
    maxAttempts: -1
    minBackoff: "10s"
    maxBackoff: "3600s"
    maxDoublings: 10
  stackdriverLoggingConfig:
    samplingRatio: 1.0
```

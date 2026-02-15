# GcpCloudSchedulerJob Examples

## 1. Simple HTTP GET on a Schedule

Trigger a public endpoint every hour:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudSchedulerJob
metadata:
  name: hourly-ping
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  schedule: "0 * * * *"
  httpTarget:
    uri: https://example.com/api/health
    httpMethod: GET
```

## 2. OIDC-Authenticated Cloud Run Trigger

Securely invoke a Cloud Run service on weekdays at 9am:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudSchedulerJob
metadata:
  name: daily-report
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  schedule: "0 9 * * 1-5"
  timeZone: America/New_York
  description: Triggers daily report generation
  attemptDeadline: "600s"
  httpTarget:
    uri: https://report-service-abc123.run.app/generate
    httpMethod: POST
    body: eyJhY3Rpb24iOiAiZGFpbHlfcmVwb3J0In0=
    headers:
      Content-Type: application/json
    oidcToken:
      serviceAccountEmail:
        value: invoker@my-gcp-project.iam.gserviceaccount.com
      audience: https://report-service-abc123.run.app
  retryConfig:
    retryCount: 3
    maxRetryDuration: "1800s"
    minBackoffDuration: "5s"
    maxBackoffDuration: "600s"
    maxDoublings: 3
```

## 3. Pub/Sub Scheduled Publisher

Publish a message to a Pub/Sub topic every 5 minutes:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudSchedulerJob
metadata:
  name: pipeline-trigger
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  schedule: "*/5 * * * *"
  timeZone: Etc/UTC
  description: Triggers data pipeline every 5 minutes
  pubsubTarget:
    topicName:
      value: projects/my-gcp-project/topics/pipeline-trigger
    data: eyJwaXBlbGluZSI6ICJkYWlseS1ldGwifQ==
    attributes:
      source: cloud-scheduler
      pipeline: daily-etl
  retryConfig:
    retryCount: 5
    maxDoublings: 3
```

## 4. App Engine Cron Job

Schedule a nightly cleanup task on an App Engine service:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudSchedulerJob
metadata:
  name: nightly-cleanup
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  schedule: "0 2 * * *"
  timeZone: Asia/Tokyo
  description: Runs nightly cleanup on the maintenance service
  attemptDeadline: "900s"
  appEngineHttpTarget:
    relativeUri: /tasks/cleanup
    httpMethod: POST
    body: eyJtb2RlIjogImZ1bGwifQ==
    headers:
      Content-Type: application/json
    appEngineRouting:
      service: maintenance
      version: v2
```

## 5. OAuth-Authenticated Google API Call

Call a Google API endpoint using OAuth2 authentication:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudSchedulerJob
metadata:
  name: api-sync
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  schedule: "0 */6 * * *"
  timeZone: Etc/UTC
  description: Syncs data via Google API every 6 hours
  httpTarget:
    uri: https://sheets.googleapis.com/v4/spreadsheets/SPREADSHEET_ID/values:batchUpdate
    httpMethod: POST
    body: eyJkYXRhIjogW119
    headers:
      Content-Type: application/json
    oauthToken:
      serviceAccountEmail:
        value: api-caller@my-gcp-project.iam.gserviceaccount.com
      scope: https://www.googleapis.com/auth/spreadsheets
```

## 6. Paused Job (Created but Not Active)

Create a job in paused state for testing or staged deployment:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudSchedulerJob
metadata:
  name: staged-job
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  schedule: "0 12 * * *"
  paused: true
  description: Staged job - enable when ready for production
  httpTarget:
    uri: https://staging-service.run.app/process
    httpMethod: POST
    oidcToken:
      serviceAccountEmail:
        value: invoker@my-gcp-project.iam.gserviceaccount.com
```

## 7. Pub/Sub Target with ValueFrom Reference

Wire the Pub/Sub topic from another OpenMCF component using `valueFrom`:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudSchedulerJob
metadata:
  name: composed-scheduler
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      fieldPath: status.outputs.project_id
  location: us-central1
  schedule: "0 8 * * *"
  pubsubTarget:
    topicName:
      valueFrom:
        kind: GcpPubSubTopic
        name: events-topic
        fieldPath: status.outputs.topic_id
    data: eyJ0cmlnZ2VyIjogInNjaGVkdWxlZCJ9
```

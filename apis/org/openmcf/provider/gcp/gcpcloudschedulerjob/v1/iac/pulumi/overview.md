# GcpCloudSchedulerJob Pulumi Module Overview

## Resource Flow

```
GcpCloudSchedulerJobStackInput
  ├── target (GcpCloudSchedulerJob)
  │   ├── metadata.name
  │   └── spec
  │       ├── project_id (StringValueOrRef)
  │       ├── job_name / location / schedule / time_zone
  │       ├── description / attempt_deadline / paused
  │       ├── http_target (+ oauth_token | oidc_token)
  │       ├── pubsub_target (+ topic_name StringValueOrRef)
  │       ├── app_engine_http_target (+ app_engine_routing)
  │       └── retry_config
  └── provider_config (GcpProviderConfig)

               │
               ▼

  cloudscheduler.NewJob()
    └── Exports: job_id, job_name, state
```

## Conditional Logic

The module uses nil-checks to conditionally build nested configuration:

1. **Target selection**: Exactly one of `http_target`, `pubsub_target`, `app_engine_http_target` is set
2. **Auth tokens**: Within `http_target`, at most one of `oauth_token` or `oidc_token`
3. **Retry config**: Only set if `retry_config` is non-nil; individual fields set only when non-zero/non-empty
4. **Job name**: Falls back to `metadata.name` if `job_name` is empty
5. **Optional strings**: Set via `pulumi.StringPtr()` only when non-empty

## Key Patterns

- **StringValueOrRef resolution**: Uses `.GetValue()` for `project_id`, `service_account_email`, `topic_name`
- **Map handling**: `headers` and `attributes` are converted to `pulumi.StringMap`
- **No labels**: Cloud Scheduler jobs don't support GCP labels (documented in `locals.go`)

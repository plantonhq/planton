# Pub/Sub Build

## Use Case

Run an inline build whenever a message lands on a Pub/Sub topic. Here the topic is Artifact Registry's `gcr` notifications topic, and the build runs for every newly pushed image tag.

## When to Use

- Builds that react to events from other Google services (Artifact Registry, Cloud Storage, Cloud Scheduler)
- Platform-owned builds that should not live in an application repository

## What This Creates

- The Cloud Build API on the project
- A Pub/Sub trigger that subscribes to the topic and runs a one-step inline build, filtered to new tags

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `pubsubConfig.topic` | a `GcpPubSubTopic` reference | The topic whose messages start builds. |
| `substitutions` | the image tag and action | Bind the payload fields your build reads. |
| `filter` | `_ACTION == "INSERT"` | Keep only the events you want to build on. |
| `build.steps` | one `cloud-sdk` step | The commands to run. |
| `build.options.logging` | `CLOUD_LOGGING_ONLY` | Required with a user-specified service account (or set `build.logsBucket`). |

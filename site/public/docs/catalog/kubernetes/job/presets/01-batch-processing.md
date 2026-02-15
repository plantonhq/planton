---
title: "Batch Processing Job"
description: "This preset creates a one-shot Kubernetes Job for batch processing. The job runs a single pod to completion and automatically cleans up after 1 hour."
type: "preset"
rank: "01"
presetSlug: "01-batch-processing"
componentSlug: "job"
componentTitle: "Job"
provider: "kubernetes"
icon: "package"
order: 1
---

# Batch Processing Job

This preset creates a one-shot Kubernetes Job for batch processing. The job runs a single pod to completion and automatically cleans up after 1 hour.

## When to Use

- One-time batch processing tasks (data transformation, ETL, report generation)
- Tasks that need to run once and exit
- Ad-hoc operations that should not be scheduled

## Key Configuration Choices

- **Single completion** (`completions: 1`, `parallelism: 1`) -- one pod runs to completion; increase for parallel batch processing
- **Backoff limit** (`3`) -- retries up to 3 times on failure before marking the job as failed
- **Restart policy** (`Never`) -- failed pods are replaced, not restarted; preserves logs for debugging
- **TTL cleanup** (`3600` seconds) -- the job and its pods are automatically deleted 1 hour after completion

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |
| `<your-container-registry>/<your-image>` | Container image for the batch task | Your container registry |
| `<your-image-tag>` | Image tag or version | Your CI/CD pipeline output |

## Related Presets

- **02-data-migration** -- Job configured for database migrations with longer timeouts

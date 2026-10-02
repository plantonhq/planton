# Autoscaled Worker

## Use Case

A worker whose load rises and falls -- a media transcoder, a report generator, a batch enricher -- run between an instance floor and ceiling by a signal you drive, with in-memory scratch space per instance and a startup probe that waits for the worker to warm up.

## When to Use

- A queue whose depth varies by hour or by day
- Work that benefits from many instances at peak but should shrink at night
- A worker that needs a scratch volume per instance

## What This Creates

- A worker pool in `us-central1` with a 2 CPU / 2 GiB transcoder container
- A 1 GiB in-memory `emptyDir` scratch volume mounted at `/tmp/work`
- A TCP startup probe on 8081 with a 60-second window
- `AUTOMATIC` scaling between 1 and 20 instances, a dedicated runtime identity

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `scaling.minInstanceCount` / `maxInstanceCount` | `1` / `20` | The floor is the committed spend; the ceiling is the cost circuit breaker. |
| `volumes[].emptyDir.sizeLimit` | `1Gi` | Scratch counts against the instance's memory limit on the MEMORY medium; `DISK` for larger scratch. |
| `containers[].startupProbe` | TCP 8081, 6 x 10 s | Match the worker's warm-up time; the window is capped at 240 seconds. |
| `containers[].resources` | `2` / `2Gi` | Size for one unit of work; sidecars add to the instance total. |

The scaling signal is yours to build -- Cloud Run does not watch a queue on its own. Wire a queue-depth metric through Cloud Monitoring to the pool's instance count.

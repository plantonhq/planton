# Scratch Volume Worker

## Use Case

A worker that needs room to work on each item -- a media transcoder, a report generator, a batch enricher -- with in-memory scratch space per instance and a startup probe that waits for the worker to warm up. It runs a count you set and change as the queue grows or shrinks.

## When to Use

- Work that writes intermediate files before uploading a result
- A worker that takes a while to warm up before it can take work
- A queue whose steady depth calls for a handful of instances

## What This Creates

- A worker pool in `us-central1` with a 2 CPU / 2 GiB transcoder container
- A 1 GiB in-memory `emptyDir` scratch volume mounted at `/tmp/work`
- A TCP startup probe on 8081 with a 60-second window
- `MANUAL` scaling at four instances, a dedicated runtime identity

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `scaling.manualInstanceCount` | `4` | Every instance bills around the clock; raise it for a deeper queue, lower it at quiet hours, `0` parks the pool. |
| `volumes[].emptyDir.sizeLimit` | `1Gi` | Scratch counts against the instance's memory limit on the MEMORY medium; `DISK` for larger scratch. |
| `containers[].startupProbe` | TCP 8081, 6 x 10 s | Match the worker's warm-up time; the window is capped at 240 seconds. |
| `containers[].resources` | `2` / `2Gi` | Size for one unit of work; sidecars add to the instance total. |

Cloud Run does not change the count on its own. To follow a queue, change `manualInstanceCount` from whatever watches the queue's depth.

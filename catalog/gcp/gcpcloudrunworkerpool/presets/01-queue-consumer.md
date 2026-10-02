# Queue Consumer

## Use Case

A fixed number of always-running workers pulling from a Pub/Sub subscription and writing to a private database: the classic background worker. The container has no port to serve; it opens outbound connections to the queue and the database and reports health on a listener the liveness probe names.

## When to Use

- Event processing from Pub/Sub, a Cloud Tasks queue, or a Redis stream
- A steady workload where a fixed instance count is the right size
- Any process that should run forever and find its own work

## What This Creates

- A worker pool in `us-central1` with one container pulling from `orders-events`
- A Secret Manager env var for the database password, a liveness probe on port 8081
- A dedicated runtime service account and direct VPC egress into `workers-subnet`
- `MANUAL` scaling at exactly two instances, deletion protection on, `deletionPolicy: PREVENT`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `containers[].image` | `orders-worker:1.4.0` | Your worker image (pin a digest or immutable tag). Leave blank for the pipeline to inject. |
| `scaling.manualInstanceCount` | `2` | Size for the queue's throughput; `0` parks the pool without deleting it. |
| `containers[].resources` | `1` CPU / `512Mi` | More for heavier per-message work; every instance bills its allocation around the clock. |
| `vpcAccess.egress` | `PRIVATE_RANGES_ONLY` | `ALL_TRAFFIC` to route public egress through the VPC (a NAT gateway is then needed). |
| `containers[].livenessProbe` | HTTP `/healthz` on 8081 | Match your health listener; remove it if the worker exposes none. |

Every instance bills instance-based for its whole lifetime, serving or idle -- there is no scale-to-zero on a worker pool except parking it.

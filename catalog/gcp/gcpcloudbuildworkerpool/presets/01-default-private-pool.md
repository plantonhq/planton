# Default Private Pool

## Use Case

Give a team's builds dedicated, larger machines than Google's default pool, without any private networking.

## When to Use

- Builds that need more CPU, memory, or disk than the default pool offers
- A first private pool before networking requirements are known

## What This Creates

- The Cloud Build API on the project
- A private pool of e2-standard-4 workers with 200 GB disks in us-central1

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `us-central1` | The region your builds and triggers run in. |
| `workerConfig.machineType` | `e2-standard-4` | Bigger machines for heavy builds; smaller to lower the per-minute rate. |
| `workerConfig.diskSizeGb` | `200` | Large container images and caches need room. |

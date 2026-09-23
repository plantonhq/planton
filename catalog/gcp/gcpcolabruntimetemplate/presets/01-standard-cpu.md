# Standard CPU

## Use Case

The everyday notebook machine: a mid-size CPU runtime with a balanced disk and internet access, shut down after an hour idle.

## When to Use

- Exploration, data wrangling, and light modeling
- The default template for a team new to Colab Enterprise

## What This Creates

- A template in `us-central1` on `e2-standard-4` with a 100 GB balanced disk, internet access, and a one-hour idle shutdown

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `machineSpec.machineType` | `e2-standard-4` | Larger machines for heavier work (a new template per size). |
| `idleTimeout` | `3600s` | Shorter for cost, longer for long-running interactive sessions. |
| `dataPersistentDiskSpec.diskSizeGb` | `100` | More room for local data. |

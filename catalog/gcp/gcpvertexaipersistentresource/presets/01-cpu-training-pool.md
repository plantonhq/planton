# CPU Training Pool

## Use Case

Keep two CPU machines warm so frequent small training jobs -- hyperparameter trials, nightly retrains, pipeline steps -- start in seconds instead of waiting for provisioning.

## When to Use

- Many short training jobs a day where startup time dominates
- Interactive iteration where waiting minutes per run hurts
- A first persistent resource to measure the readiness trade

## What This Creates

- A persistent resource `cpu-training-pool` in `us-central1` with one pool `cpu-workers` of two `n1-standard-8` machines on Google's default boot disk

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `resourcePools[].machineSpec.machineType` | `n1-standard-8` | The machine your jobs need (fixed at creation). |
| `resourcePools[].replicaCount` | `2` | How many jobs run at once; mutable, and every replica bills around the clock. |
| `resourcePools[].id` | `cpu-workers` | The pool id jobs' worker pools refer to. |
| `deletionPolicy` | `DELETE` | `PREVENT` when jobs in production depend on it. |

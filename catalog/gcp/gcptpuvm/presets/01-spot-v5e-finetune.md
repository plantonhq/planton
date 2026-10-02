# Spot v5e Fine-Tuning

## Use Case

Fine-tune or experiment on an 8-chip TPU v5e slice at spot prices, with JAX installed at boot.

## When to Use

- Fine-tuning open models where a reclaimed slice only costs a resume
- Research and batch experiments

## What This Creates

- A `v5litepod-8` slice in `us-west4-a` on the v5e runtime, on spot capacity, with a startup script installing JAX for TPU

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `zone` | `us-west4-a` | A zone where your project has v5e quota. |
| `acceleratorType` | `v5litepod-8` | Larger slices (`v5litepod-16`, ...) or another generation (with its runtime). |
| `schedulingConfig.spot` | `true` | Remove for on-demand capacity that is not reclaimed. |

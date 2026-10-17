# GPU L4 With Fallback

A class for GPU inference on NVIDIA L4s. It takes Spot L4 capacity first, falls back to on-demand L4s, and never falls back to a machine without a GPU.

## When to Use

- Model serving and batch inference that fits one L4 (24 GB) per replica.
- A region where L4 Spot capacity comes and goes.

## Key Configuration Choices

- **`machineType: g2-standard-8`.** G2 machines carry their L4s built in, so the machine type fixes the GPU count and no `gpu` block is needed.
- **Explicit zones.** Not every zone has L4s, and GKE accepts only zones of the cluster's own region. Listing the ones that do keeps GKE from trying zones that can never satisfy the rule.
- **`whenUnsatisfiable: DoNotScaleUp`.** A CPU-only node is useless to these pods, so they wait for a GPU rather than land somewhere they cannot run.
- **`cpuManagerPolicy: static`.** Guaranteed pods with whole-CPU requests get exclusive cores, which steadies latency for the model server's CPU work.
- **`gpuConsolidationThreshold: 100`.** A node whose GPU is not fully used becomes a candidate for removal, so idle GPUs do not linger.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<l4-zone-1>`, `<l4-zone-2>` | Zones of the cluster's region with L4 capacity (GKE refuses a zone outside the cluster's region) | `gcloud compute accelerator-types list --filter="name=nvidia-l4"` |

## Related Presets

- **02-spot-first-on-demand-fallback**: the same Spot-first idea for CPU work.

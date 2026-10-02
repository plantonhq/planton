# GPU Inference Worker

## Use Case

A queue-fed inference worker with one NVIDIA L4 GPU per instance: embeddings, transcription, image generation -- any model that should run continuously against a stream of work rather than behind an HTTP endpoint. Model weights mount read-only from a Cloud Storage bucket; a long startup window covers model loading.

## When to Use

- Batch or streaming inference where a GPU is idle-cheap enough to keep warm
- Embedding pipelines fed from Pub/Sub
- Any GPU workload that does not need a request path

## What This Creates

- A worker pool in `us-central1` with a 4 CPU / 16 GiB container and one `nvidia-l4` GPU per instance
- A read-only GCS FUSE mount of the `ml-models` bucket at `/models`
- A TCP startup probe with a four-minute window (the maximum)
- Single-zone GPU capacity (`gpuZonalRedundancyDisabled: true`) for the lower rate; one instance; deletion protection off

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `nodeSelector.accelerator` | `nvidia-l4` | The GPU type your region offers and your model needs. |
| `gpuZonalRedundancyDisabled` | `true` | `false` for zonal redundancy at the higher GPU rate. |
| `scaling.manualInstanceCount` | `1` | More instances for more throughput; each carries a GPU for its whole lifetime. |
| `volumes[].gcs.bucket` | `ml-models` | Your model bucket; keep `readOnly: true`. |
| `deletionProtection` | `false` | `true` once the worker is load-bearing. |

The GPU is the dominant meter -- roughly ten times the CPU meter of a 4-vCPU instance -- and bills for the instance's whole lifetime. Park the pool (`manualInstanceCount: 0`) between batches.

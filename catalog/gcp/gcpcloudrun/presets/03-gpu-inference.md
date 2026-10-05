# GPU Inference Service

This preset creates a GPU-backed model-serving endpoint: one NVIDIA L4 per instance, scale-to-zero so idle GPUs cost nothing, instance-based billing so the model stays resident, and IAM-authenticated access.

## When to Use

- LLM, embedding, image, and speech model serving with bursty demand
- Inference workloads that fit one GPU per instance and tolerate cold starts
- Teams that want managed GPU serving without operating a GKE GPU pool

## Key Configuration Choices

- **`nodeSelector.accelerator: nvidia-l4`** — every instance gets one L4; Cloud Run manages the driver and runtime
- **`cpu: "4"` / `memory: 16Gi`** — Cloud Run's GPU minimums; size up for larger models
- **`cpuIdle: false`** — instance-based billing keeps CPU (and the model in GPU memory) alive between requests
- **`gpuZonalRedundancyDisabled: true`** — single-zone serving buys cheaper GPU capacity at zonal-failure risk; remove it for zonal redundancy
- **Scale-to-zero with `maxInstanceCount: 4`** — GPU spend is bounded on both ends
- **Generous startup probe** — instances receive traffic only after model weights load
- **Low concurrency (`4`)** — inference is compute-bound; short queues beat head-of-line blocking

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `us-docker.pkg.dev/my-project/my-repo/inference:1.0.0` | Your model-serving image | Artifact Registry |

## Related Presets

- **01-public-api-service** — the standard CPU service shape
- **02-private-vpc-service** — add VPC egress if the model reads from private stores

## Related Kinds

- [GcpGkeNodePool](/docs/catalog/gcp/gcpgkenodepool) — GPU node pools, when the workload outgrows one-GPU-per-instance serving

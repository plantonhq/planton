# Gemma on L4 with vLLM

## Use Case

A production deployment of a Model Garden publisher model on an explicit GPU shape: Gemma 2B served by vLLM on a `g2-standard-12` with one NVIDIA L4, autoscaling between one and two Spot replicas on accelerator duty cycle, behind a dedicated endpoint with its own DNS name. The serving container, its routes, and all three probes are declared so the deployment is reproducible to the byte.

## When to Use

- Serving an open model at a known cost per hour with a known machine
- Traffic that needs its own endpoint DNS isolated from the shared regional host
- Any deployment where "let Model Garden pick" is not an acceptable answer

## What This Creates

- An uploaded Model (`gemma-2b`) served by Model Garden's vLLM image on `/generate`, health on `/ping`
- Dedicated resources: `g2-standard-12` + `NVIDIA_L4`, 1-2 Spot replicas, scaling at 70% accelerator duty cycle
- A dedicated endpoint (`{endpoint}.us-central1-{project}.prediction.vertexai.goog`)
- `deletionPolicy: PREVENT`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `publisherModelName` | `gemma@gemma-1.1-2b-it` | Another publisher model and version; the container args and shared memory follow the model's size. |
| `deployConfig.dedicatedResources.machineSpec` | `g2-standard-12` + 1 `NVIDIA_L4` | A larger GPU (`NVIDIA_TESLA_A100`, `NVIDIA_H100_80GB`) for larger models; check regional availability. |
| `deployConfig.dedicatedResources.spot` | `true` | `false` for on-demand capacity that is never preempted. |
| `deployConfig.dedicatedResources.maxReplicaCount` | `2` | Your traffic ceiling; every replica carries a GPU. |
| `endpointConfig.dedicatedEndpointEnabled` | `true` | `false` to serve from the shared regional host. |
| `endpointConfig.privateServiceConnectConfig` | none | Expose the endpoint over Private Service Connect to allowlisted projects. |

Every field is immutable; change anything and the deployment is replaced (the endpoint ID changes). The GPU bills from the moment the replica is ready; `minReplicaCount` is the committed spend.

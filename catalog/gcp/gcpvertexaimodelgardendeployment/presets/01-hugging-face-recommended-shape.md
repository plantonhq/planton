# Hugging Face on the Recommended Shape

## Use Case

Deploy an open Hugging Face model to a Vertex AI endpoint in one step and let Model Garden choose the machine: the fastest way from a model ID to a serving endpoint. Model Garden uploads the model, picks the container and the recommended machine shape (a small GPU for this model), creates the endpoint, and deploys one replica.

## When to Use

- Trying a model before committing to a machine shape
- Small open models where Model Garden's recommendation is the right answer
- Development and evaluation endpoints

## What This Creates

- An uploaded Model and a Vertex AI Endpoint in `us-central1` named `qwen-small`
- One replica of `Qwen/Qwen3-0.6B` on Model Garden's recommended machine shape
- License acceptance recorded (`acceptEula: true`)

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `huggingFaceModelId` | `Qwen/Qwen3-0.6B` | Any Hugging Face model Model Garden can serve; gated models also need `modelConfig.huggingFaceAccessToken`. |
| `location` | `us-central1` | A region with the accelerator the model's recommended shape uses. |
| `endpointConfig.endpointDisplayName` | `qwen-small` | The endpoint's display name; Google names it when empty. |
| `deletionPolicy` | `DELETE` | `PREVENT` for an endpoint clients depend on. |

Every field is immutable: a different model, machine, or endpoint setting replaces the whole deployment and the endpoint ID changes. The replica bills its machine and accelerator hours from the moment it is ready.

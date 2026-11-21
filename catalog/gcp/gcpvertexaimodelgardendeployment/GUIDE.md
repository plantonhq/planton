# GcpVertexAiModelGardenDeployment Guide

The judgment this guide protects: this block is a frozen deployment, not a
serving endpoint you tune. Only its replica bounds and autoscaling change
in place, the replicas bill from the moment they are ready, and most open
models need a GPU. Decide
the model, the compute, and the endpoint once; roll forward by declaring
again.

## One step, three resources

Google's one-step deployment uploads the model, creates a Vertex AI
endpoint, and deploys the model to it. The block exports the endpoint the
same way `GcpVertexAiEndpoint` does (`endpoint_id`, `endpoint_name`), so
whatever calls the model reads the endpoint without caring which block
made it. Under `dedicatedResources`, `minReplicaCount`, `maxReplicaCount`,
`requiredReplicaCount`, and `autoscalingMetricSpecs` change in place and
the endpoint keeps its ID. Anything else -- a changed model version,
machine, token, or endpoint setting -- undeploys, deletes the endpoint, and
redeploys, and the endpoint ID changes. Clients pinned to the old ID break.
Roll forward with a second block and cut traffic over, then remove the
first. Destroy undeploys the model and deletes the endpoint; Google's
provider makes no call that deletes the uploaded Model resource.

## Which model

Exactly one of `publisherModelName` -- a Model Garden entry with its
version, `publishers/google/models/gemma@gemma-1.1-2b-it`, or a Hugging
Face model Model Garden lists as `publishers/hf-{author}/models/{model}@001`
-- or `huggingFaceModelId`, a hub id like `Qwen/Qwen3-0.6B`. Licensed and
gated models refuse to deploy until `acceptEula` is true; gated Hugging
Face models also need a read token, held as a secret. Inside a VPC Service
Controls perimeter, `huggingFaceCacheEnabled` deploys from Google's copy.

## Which compute, and what it costs

Omit `deployConfig` and Model Garden picks the model's recommended machine
and one replica -- the right answer for trying a model. Decide this up
front: under Terraform, removing `deployConfig` from a deployed model
changes nothing, so returning to the recommended shape takes a new
deployment; and `fastTryoutEnabled` or `systemLabels` need
`dedicatedResources` beside them. For production, declare
`dedicatedResources`: the machine type (required there -- left blank,
Google falls back to a small CPU machine) and accelerator (an
`NVIDIA_L4` on `g2-standard-12` serves 2B-class models; A100s and H100s
serve larger ones), `minReplicaCount` (never below one; there is no
scale-to-zero), `maxReplicaCount`, `spot` for preemptible capacity at a
steep discount, and the metric autoscaling follows (`accelerator/duty_cycle`
on a GPU). Every replica bills its machine and accelerator hours from
ready until undeployed; `minReplicaCount` is the committed spend and the
accelerator is usually the whole bill. The project needs quota for the
accelerator in the region before the deploy will succeed, and Google
reserves it at the ceiling: `maxReplicaCount` times the machine's cores
and accelerators.

## The serving container

Model Garden models ship a serving container; leave `containerSpec` out
for them. Declare it for a custom image or to pin vLLM's options: the
image, its command and args, the port Vertex AI sends predictions to
(the first of `ports`), the predict and health routes, shared memory for
multi-GPU loading, and the startup, liveness, and health probes (each
exactly one of exec, gRPC, HTTP GET, or TCP).

## The endpoint

`dedicatedEndpointEnabled` gives the endpoint its own DNS name isolated
from the shared regional host -- once on, the shared host stops serving
it. `privateServiceConnectConfig` removes the public host altogether: the
endpoint is published only to the allowlisted `GcpProject`s, and
`pscAutomationConfig` has Vertex AI create the consumer endpoint in one
named project and network.

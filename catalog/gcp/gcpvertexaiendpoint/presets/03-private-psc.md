# Private PSC Endpoint

Prediction serving exposed through Private Service Connect: per-project
access control and IAM-authorized connections, without VPC peering.

## What this preset creates

An endpoint named `Partner Inference` in `us-central1`, reachable only
through PSC forwarding rules created from the two allowlisted consumer
projects, and CMEK-encrypted under the referenced `GcpKmsKey`
(`inference-key`). Secure PSC (IAM authorization on top of network
reachability) is not offered here: the GA provider does not expose it.

## When to use

- Serving predictions to specific consumer projects (internal platform
  teams or external partners) without sharing a network
- The strongest isolation posture Vertex AI serving offers
- Multi-tenant architectures where each consumer connects from its own
  VPC

## Constraints

- PSC is mutually exclusive with both VPC peering (`network`) and the
  dedicated DNS (`dedicatedEndpointEnabled`) — the spec enforces both
  pre-deploy.

## Remix ideas

- Leave `projectAllowlist` empty to allow any project in the same
  organization to connect.

# GcpServiceConnectionPolicy - Pulumi Module

This Pulumi (Go) module provisions a service connection policy (`networkconnectivity.ServiceConnectionPolicy`) — the per-network authorization for Google's service connectivity automation to create PSC endpoints on a producer's behalf. It is the Pulumi-side implementation of the Planton `GcpServiceConnectionPolicy` resource kind and has feature parity with the Terraform module.

## Overview

The module creates one policy per (network, service class, region) triple. PSC-first managed services (Memorystore for Valkey, Redis Cluster) refuse to create instances on a network until a policy for their service class exists in that region — this module is that prerequisite.

`location`, `network`, `serviceClass`, and the policy name are ForceNew; the `pscConfig` contents, description, and labels update in place, so subnet growth and limit raises never recreate the policy. The Service Connectivity API requires relative resource paths — the module normalizes `https://` self-link URLs for `network` and `subnetworks` identically to the Terraform module. The module enables the Network Connectivity and Compute Engine APIs so a fresh project works first try.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Pulumi Usage

```bash
cd catalog/gcp/gcpserviceconnectionpolicy/iac/pulumi
make build
pulumi up --stack dev
```

## Module Layout

- `main.go` — entrypoint; loads the IaC input and calls the module
- `module/main.go` — provider setup and resource orchestration
- `module/locals.go` — metadata-derived values, label merge, self-link normalization
- `module/service_connection_policy.go` — API enablement + the policy resource
- `module/outputs.go` — output keys (must match `outputs.proto`)

## Outputs

| Name | Description |
|------|-------------|
| `policy_id` | Fully qualified policy resource path |
| `name` | Short policy name |
| `infrastructure` | Underlying connectivity mechanism (PSC) |
| `etag` | Server-computed etag |

## Lifecycle Notes

Deploy the policy before the first instance of its service class in the region, and keep it alive as long as any instance depends on it — deleting the policy strands existing endpoints and blocks new ones.

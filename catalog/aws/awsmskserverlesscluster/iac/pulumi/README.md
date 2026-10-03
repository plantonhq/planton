# AwsMskServerlessCluster — Pulumi IaC Module

Pulumi (Go) module for provisioning an Amazon MSK Serverless cluster using the Planton `AwsMskServerlessClusterSpec`.

## Overview

This module creates a single `msk.ServerlessCluster`: network interfaces in the referenced subnets, the referenced security groups attached, and SASL/IAM client authentication enabled unconditionally (the only scheme serverless MSK supports — AWS requires it, so it is not a spec field).

The resource is effectively immutable: everything except tags is create-time (ForceNew).

## Usage

The module is invoked from the entry point in `main.go`, which loads an `AwsMskServerlessClusterIacInput` and calls `module.Resources()`.

### IaC Input

- `target` — the `AwsMskServerlessCluster` resource (metadata + spec).
- `provider_config` — AWS credentials (static keys, keyless web identity, or ambient chain), resolved by the shared provider builder.

### Outputs

```bash
pulumi stack output cluster_arn
pulumi stack output bootstrap_brokers_sasl_iam
```

## File Structure

| File | Purpose |
|------|---------|
| `Pulumi.yaml` | Pulumi project metadata (name: `aws-msk-serverless-cluster`, runtime: Go) |
| `main.go` | Entry point — loads IaC input, runs the Pulumi program |
| `module/main.go` | Orchestrator — provider setup, cluster creation, output exports |
| `module/locals.go` | Naming basis (metadata.name) and resource-identity tags |
| `module/cluster.go` | The MSK Serverless cluster resource |
| `module/outputs.go` | Output key constants |

## Prerequisites

- Go 1.21+
- Pulumi CLI v3+
- `pulumi-aws` plugin v7
- AWS credentials (ambient or via IaC input)

# AwsLambdaEventSourceMapping — Pulumi IaC Module

Pulumi (Go) module for provisioning a Lambda event source mapping using the Planton `AwsLambdaEventSourceMappingSpec`.

## Overview

This module creates a single `lambda.EventSourceMapping`: the managed poller that reads an SQS queue, a Kinesis or DynamoDB stream, a Kafka topic (MSK or self-managed), an Amazon MQ queue, or a DocumentDB change stream and invokes a Lambda function with batched records.

The event source (ARN or self-managed Kafka bootstrap servers) is create-time immutable; batching, filters, failure handling, and the target function edit in place.

## Usage

The module is invoked from the entry point in `main.go`, which loads an `AwsLambdaEventSourceMappingIacInput` and calls `module.Resources()`.

### IaC Input

- `target` — the `AwsLambdaEventSourceMapping` resource (metadata + spec).
- `provider_config` — AWS credentials (static keys, keyless web identity, or ambient chain), resolved by the shared provider builder.

### Outputs

```bash
pulumi stack output uuid
pulumi stack output mapping_arn
```

## File Structure

| File | Purpose |
|------|---------|
| `Pulumi.yaml` | Pulumi project metadata (name: `aws-lambda-event-source-mapping`, runtime: Go) |
| `main.go` | Entry point — loads IaC input, runs the Pulumi program |
| `module/main.go` | Orchestrator — provider setup, mapping creation, output exports |
| `module/locals.go` | Planton identity tags |
| `module/mapping.go` | The event source mapping resource |
| `module/outputs.go` | Output key constants |

## Prerequisites

- Go 1.21+
- Pulumi CLI v3+
- `pulumi-aws` plugin v7
- AWS credentials (ambient or via IaC input)

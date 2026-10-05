# AwsMwaaEnvironment — Pulumi IaC Module

Pulumi module for provisioning AWS MWAA (Managed Workflows for Apache Airflow) environments using the Planton `AwsMwaaEnvironmentSpec`.

## Overview

This module creates:
- An MWAA Environment (`mwaa.Environment`) with configurable Airflow version, S3 source, IAM execution role, VPC networking, encryption, sizing, logging, maintenance, and worker replacement strategy settings.

Network ingress is composed, never embedded: the environment attaches the referenced `securityGroupIds` directly, and the rules MWAA needs (self-referencing all-traffic, HTTPS 443 ingress, egress) live on those first-class security group nodes.

## Usage

### As a Pulumi program

The module is designed to be invoked from the entry point in `main.go`, which loads an `AwsMwaaEnvironmentIacInput` and calls `module.Resources()`:

```go
package main

import (
    awsmwaaenvironmentv1 "github.com/plantonhq/planton/catalog/aws/awsmwaaenvironment/v1alpha1"
    "github.com/plantonhq/planton/catalog/aws/awsmwaaenvironment/iac/pulumi/module"
    "github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        iacInput := &awsmwaaenvironmentv1.AwsMwaaEnvironmentIacInput{}
        if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
            return err
        }
        return module.Resources(ctx, iacInput)
    })
}
```

### IaC Input

The IaC input is an `AwsMwaaEnvironmentIacInput` protobuf message containing:
- `target` — the `AwsMwaaEnvironment` resource (metadata + spec).
- `provider_config` — optional AWS credentials (region, access key, secret key, session token).

### Outputs

The module exports the outputs declared in `AwsMwaaEnvironmentOutputs` (see `module/outputs.go` for keys). Access them via `pulumi stack output`:

```bash
pulumi stack output environment_arn
pulumi stack output webserver_url
```

## File Structure

| File | Purpose |
|------|---------|
| `Pulumi.yaml` | Pulumi project metadata (name: `aws-mwaa-environment`, runtime: Go) |
| `main.go` | Entry point — loads IaC input, runs Pulumi program |
| `module/main.go` | Orchestrator — resource creation flow + output exports |
| `module/locals.go` | Locals initialization (identity tags, naming basis, resolved target) |
| `module/environment.go` | MWAA Environment resource with all configuration blocks |
| `module/outputs.go` | Output key constants |

## Prerequisites

- Go 1.21+
- Pulumi CLI v3+
- AWS credentials (ambient or via IaC input)
- `pulumi-aws` plugin v7

## Running Locally

```bash
# Navigate to the Pulumi module directory
cd catalog/aws/awsmwaaenvironment/iac/pulumi

# Set stack configuration (or provide via IaC input JSON)
pulumi stack init dev
pulumi config set aws:region us-east-1

# Preview changes
pulumi preview

# Apply changes
pulumi up

# View outputs
pulumi stack output

# Destroy resources
pulumi destroy
```

### Debug

If the environment gets stuck in `CREATING` status:

1. Check CloudWatch Logs for the environment (if logging was enabled).
2. Verify the execution role has the required S3, CloudWatch Logs, and SQS permissions.
3. Verify the VPC subnets are private (no internet gateway route) and have NAT gateway access.
4. Verify the referenced security group has the self-referencing inbound rule — without it MWAA components cannot communicate.
5. Check that the S3 bucket has versioning enabled and contains DAG files at the specified `dagS3Path`.

Common failure modes:
- **"Access denied" during creation:** Execution role missing S3 or SQS permissions.
- **Environment stuck in CREATING (30+ min):** Security group missing self-referencing rule, or subnets are public.
- **"Invalid plugins.zip":** The zip file structure doesn't match Airflow's expected plugin layout.
- **"Requirements installation failed":** Conflicting Python package versions or packages requiring system-level dependencies (use startup script).

## Related

- [Spec reference](../../README.md)

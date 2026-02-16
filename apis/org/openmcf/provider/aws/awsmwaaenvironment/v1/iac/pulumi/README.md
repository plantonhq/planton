# AwsMwaaEnvironment — Pulumi IaC Module

Pulumi module for provisioning AWS MWAA (Managed Workflows for Apache Airflow) environments using the OpenMCF `AwsMwaaEnvironmentSpec`.

## Overview

This module creates:
- An MWAA Environment (`mwaa.Environment`) with configurable Airflow version, S3 source, IAM execution role, VPC networking, encryption, sizing, logging, and maintenance settings.
- A managed EC2 Security Group (`ec2.SecurityGroup`) with a self-referencing inbound rule and HTTPS (443) ingress from source security groups and/or CIDR blocks — conditional on `securityGroupIds` or `allowedCidrBlocks` being provided.

## Usage

### As a Pulumi program

The module is designed to be invoked from the entry point in `main.go`, which loads an `AwsMwaaEnvironmentStackInput` and calls `module.Resources()`:

```go
package main

import (
    awsmwaaenvironmentv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsmwaaenvironment/v1"
    "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsmwaaenvironment/v1/iac/pulumi/module"
    "github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/stackinput"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        stackInput := &awsmwaaenvironmentv1.AwsMwaaEnvironmentStackInput{}
        if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
            return err
        }
        return module.Resources(ctx, stackInput)
    })
}
```

### Stack Input

The stack input is an `AwsMwaaEnvironmentStackInput` protobuf message containing:
- `target` — the `AwsMwaaEnvironment` resource (metadata + spec).
- `provider_config` — optional AWS credentials (region, access key, secret key, session token).

### Outputs

The module exports 8 stack outputs (see `module/outputs.go` for keys). Access them via `pulumi stack output`:

```bash
pulumi stack output environment_arn
pulumi stack output webserver_url
pulumi stack output security_group_id
```

## File Structure

| File | Purpose |
|------|---------|
| `Pulumi.yaml` | Pulumi project metadata (name: `aws-mwaa-environment`, runtime: Go) |
| `main.go` | Entry point — loads stack input, runs Pulumi program |
| `module/main.go` | Orchestrator — resource creation flow + output exports |
| `module/locals.go` | Locals initialization (labels, resolved target) |
| `module/security_group.go` | Managed security group with self-referencing + HTTPS ingress rules |
| `module/environment.go` | MWAA Environment resource with all configuration blocks |
| `module/outputs.go` | Output key constants (8 total) |

## Prerequisites

- Go 1.21+
- Pulumi CLI v3+
- AWS credentials (ambient or via stack input)
- `pulumi-aws` plugin v7

## Running Locally

```bash
# Navigate to the Pulumi module directory
cd apis/org/openmcf/provider/aws/awsmwaaenvironment/v1/iac/pulumi

# Set stack configuration (or provide via stack input JSON)
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
4. Verify the security group has the self-referencing inbound rule (the managed SG handles this automatically).
5. Check that the S3 bucket has versioning enabled and contains DAG files at the specified `dagS3Path`.

Common failure modes:
- **"Access denied" during creation:** Execution role missing S3 or SQS permissions.
- **Environment stuck in CREATING (30+ min):** Security group missing self-referencing rule, or subnets are public.
- **"Invalid plugins.zip":** The zip file structure doesn't match Airflow's expected plugin layout.
- **"Requirements installation failed":** Conflicting Python package versions or packages requiring system-level dependencies (use startup script).

## Related

- [Spec reference](../../README.md)
- [Module architecture](./overview.md)
- [Examples](../../examples.md)

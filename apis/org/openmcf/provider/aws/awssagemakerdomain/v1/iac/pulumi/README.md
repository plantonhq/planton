# AwsSagemakerDomain — Pulumi IaC Module

Pulumi module for provisioning Amazon SageMaker Domains using the OpenMCF `AwsSagemakerDomainSpec`.

## Overview

This module creates:
- A SageMaker Domain (`aws.sagemaker.Domain`) with configurable authentication, VPC networking, user settings, Docker access, and encryption.
- Default user settings including JupyterLab configuration, KernelGateway configuration, idle timeout, sharing settings, and space storage.

## Usage

### As a Pulumi program

The module is designed to be invoked from the entry point in `main.go`, which loads an `AwsSagemakerDomainStackInput` and calls `module.Resources()`:

```go
package main

import (
    awssagemakerdomainv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awssagemakerdomain/v1"
    "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awssagemakerdomain/v1/iac/pulumi/module"
    "github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/stackinput"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        stackInput := &awssagemakerdomainv1.AwsSagemakerDomainStackInput{}
        if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
            return err
        }
        return module.Resources(ctx, stackInput)
    })
}
```

### Stack Input

The stack input is an `AwsSagemakerDomainStackInput` protobuf message containing:
- `target` — the `AwsSagemakerDomain` resource (metadata + spec).
- `provider_config` — optional AWS credentials (region, access key, secret key, session token).

### Outputs

The module exports 6 stack outputs (see `module/outputs.go` for keys). Access them via `pulumi stack output`:

```bash
pulumi stack output domain_id
pulumi stack output domain_arn
pulumi stack output domain_url
pulumi stack output home_efs_file_system_id
pulumi stack output security_group_id_for_domain_boundary
pulumi stack output single_sign_on_application_arn
```

## File Structure

| File | Purpose |
|------|---------|
| `main.go` | Entry point — loads stack input, runs Pulumi program |
| `module/main.go` | Orchestrator — resource creation flow + output exports |
| `module/locals.go` | Locals initialization (labels, resolved target) |
| `module/domain.go` | SageMaker Domain resource creation |
| `module/outputs.go` | Output key constants |

## Prerequisites

- Go 1.21+
- Pulumi CLI v3+
- AWS credentials (ambient or via stack input)
- `pulumi-aws` plugin v7

## Build

```bash
cd iac/pulumi
go build -o /dev/null ./...
```

## Related

- [Spec reference](../../README.md)
- [Module architecture](./overview.md)
- [Examples](../../examples.md)

# AwsStepFunction — Terraform Module

## Overview

This Terraform module provisions an AWS Step Functions state machine with optional CloudWatch logging, X-Ray tracing, and customer-managed KMS encryption.

## Module Structure

```
main.tf       — State machine resource with dynamic config blocks
locals.tf     — Tag computation, definition serialization, defaults
outputs.tf    — state_machine_arn, state_machine_name
variables.tf  — Input variables (metadata, spec)
provider.tf   — AWS provider configuration (v5.82.0)
```

## Usage

```hcl
module "step_function" {
  source = "./path/to/module"

  metadata = {
    name = "my-workflow"
    org  = "my-org"
    env  = "dev"
    id   = "my-workflow-dev"
  }

  spec = {
    type = "STANDARD"
    role_arn = {
      value = "arn:aws:iam::123456789012:role/StepFunctionsExecRole"
    }
    definition = {
      StartAt = "Hello"
      States = {
        Hello = {
          Type   = "Pass"
          Result = "Hello, World!"
          End    = true
        }
      }
    }
  }
}
```

## Outputs

| Output | Description |
|--------|-------------|
| `state_machine_arn` | ARN of the state machine |
| `state_machine_name` | Name of the state machine |

## Implementation Notes

- **Definition**: The `definition` arrives as a map and is serialized to JSON using `jsonencode()`.
- **Dynamic blocks**: Tracing, logging, and encryption configurations use `dynamic` blocks to conditionally include them only when specified.
- **Log destination suffix**: Auto-appends `:*` to CloudWatch Log Group ARNs (AWS requirement).
- **Encryption type**: Automatically set to `CUSTOMER_MANAGED_KMS_KEY` when a KMS key is provided.

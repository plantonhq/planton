---
title: "Dual IaC Engines"
description: "How Planton supports both Pulumi and OpenTofu/Terraform as deployment engines, giving teams the freedom to choose their provisioner without sacrificing component coverage"
icon: "gear"
order: 40
---

# Dual IaC Engines

Every deployment component in Planton ships with two IaC module implementations: a Pulumi module written in Go and an OpenTofu/Terraform module written in HCL. Both implementations receive the same input (the manifest's metadata, spec, and provider credentials), create the same cloud resources, and produce the same outputs.

This is not an abstraction layer that wraps one engine with another. These are independent, native implementations for each engine. The Pulumi module uses the Pulumi Go SDK directly. The Terraform module uses standard HCL configuration. You choose which engine to use, and Planton handles the rest.

## How Each Engine Works

### The Pulumi Path

When you run a Pulumi deployment:

```bash
planton pulumi up -f postgres.yaml --stack my-org/my-project/production
```

The CLI:

1. Validates the manifest against the protobuf schema
2. Resolves the Pulumi module for the component's kind (e.g., `KubernetesPostgres`)
3. Constructs the stack input from your manifest and provider config
4. Exports the stack input as a Pulumi config value
5. Runs the Pulumi Go program, which loads the stack input and provisions resources

The Pulumi module entry point is a Go program:

```go
func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        stackInput := &kubernetespostgresv1.KubernetesPostgresStackInput{}

        if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
            return errors.Wrap(err, "failed to load stack-input")
        }

        return module.Resources(ctx, stackInput)
    })
}
```

The `module.Resources` function is where the actual resource creation happens -- creating namespaces, deploying operators, configuring services, and exporting outputs.

### The OpenTofu/Terraform Path

When you run an OpenTofu or Terraform deployment:

```bash
planton tofu apply -f postgres.yaml
```

The CLI:

1. Validates the manifest against the protobuf schema
2. Resolves the Terraform module for the component's kind
3. Generates `terraform.tfvars.json` from the manifest's metadata and spec
4. Writes the `backend.tf` file based on manifest labels
5. Runs `tofu init` followed by `tofu apply`

The Terraform module uses standard HCL:

```hcl
# variables.tf
variable "metadata" {
  type = object({
    name = string
    org  = optional(string)
    env  = optional(string)
  })
}

variable "spec" {
  type = object({
    namespace        = object({ value = string })
    create_namespace = optional(bool, false)
    container = optional(object({
      replicas  = optional(number, 1)
      disk_size = optional(string, "1Gi")
    }))
  })
}
```

```hcl
# provider.tf
terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.35"
    }
  }
}

provider "kubernetes" {}
```

The main.tf file uses these variables to create the same resources that the Pulumi module creates.

### Terraform Support

Planton also supports HashiCorp Terraform directly:

```bash
planton terraform apply -f postgres.yaml
```

The Terraform path works identically to the OpenTofu path. The same HCL modules are used -- OpenTofu and Terraform are compatible at the module level.

## Feature Parity

Both engines receive the same stack input structure:

```protobuf
message KubernetesPostgresStackInput {
    KubernetesPostgres target = 1;
    KubernetesProviderConfig provider_config = 2;
}
```

The `target` field contains the full manifest (apiVersion, kind, metadata, spec). The `provider_config` field contains the credentials needed to authenticate with the cloud provider. Both engines receive exactly the same data.

Both engines produce the same outputs. A `KubernetesPostgres` deployment produces a namespace, service name, port-forward command, Kubernetes endpoint, and secret references -- regardless of whether Pulumi or Terraform created the resources.

## Module Structure Comparison

Every component's IaC directory contains both implementations side by side:

```text
apis/dev/planton/provider/kubernetes/kubernetespostgres/v1/iac/
|-- pulumi/
|   |-- main.go                  # Entry point: load stack input, call module
|   |-- Pulumi.yaml              # Pulumi project definition
|   |-- module/
|   |   |-- main.go              # Resource creation
|   |   |-- namespace.go         # Namespace management
|   |   |-- outputs.go           # Stack outputs
|   |   |-- locals.go            # Derived values
|   |   \-- variables.go         # Constants
|   \-- Makefile
\-- tf/
    |-- main.tf                  # Resource creation
    |-- variables.tf             # Input variables (mirrors spec.proto)
    |-- provider.tf              # Provider configuration
    |-- outputs.tf               # Stack outputs
    \-- locals.tf                # Derived values
```

The Pulumi modules are intentionally designed to be straightforward. They use simple, linear code with minimal abstraction -- making them readable by engineers who are more familiar with Terraform-style infrastructure code.

## Choosing an Engine

Both engines are fully supported. Your choice depends on your team's preferences and existing infrastructure:

| Consideration | Pulumi | OpenTofu/Terraform |
|--------------|--------|-------------------|
| **Language** | Go (compiled programs) | HCL (declarative configuration) |
| **State backends** | Pulumi Cloud, S3, GCS, Azure Blob, local | S3, GCS, Azure Storage, local |
| **Team familiarity** | Teams using Go or programmatic IaC | Teams with existing Terraform expertise |
| **Ecosystem** | Pulumi provider ecosystem | Terraform/OpenTofu provider ecosystem |
| **Complex logic** | Native Go conditionals, loops, error handling | HCL `for_each`, `count`, `dynamic` blocks |

### Unified Commands

If you do not want to specify the engine on every command, set the `planton.dev/provisioner` label in your manifest's metadata:

```yaml
metadata:
  labels:
    planton.dev/provisioner: pulumi  # or: tofu, terraform
```

Then use the unified commands, which delegate to the correct engine automatically:

```bash
# These read the provisioner label and call the right engine
planton apply -f postgres.yaml
planton plan -f postgres.yaml
planton destroy -f postgres.yaml
```

This is equivalent to running `planton pulumi up`, `planton pulumi preview`, or `planton pulumi destroy` -- but the engine selection comes from the manifest rather than the command.

## What's Next

- **[Module System](module-system)** -- How IaC modules are resolved, cached, and versioned
- **[State Management](state-management)** -- How each engine manages deployment state
- **[Deployment Components](deployment-components)** -- The anatomy of a component including both IaC implementations

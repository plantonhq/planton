# Kubernetes Secret - Terraform Module

## Overview

This Terraform module creates and manages a Kubernetes Secret with type-safe data variants. It supports all five common Kubernetes secret types: Opaque, TLS, DockerConfigJson, BasicAuth, and SSHAuth.

## Architecture

```
iac/tf/
├── provider.tf     # Terraform and Kubernetes provider requirements
├── variables.tf    # Input variables mirroring spec.proto
├── locals.tf       # Type mapping: determines secret type and data from variant
├── main.tf         # Creates kubernetes_secret_v1 resource
├── outputs.tf      # Exports secret_name, secret_namespace, secret_type
└── README.md       # This file
```

## How It Works

1. **Variable Input**: The `spec` variable accepts exactly one of the five secret type blocks
2. **Type Determination**: `locals.tf` inspects which block is non-null and sets the Kubernetes secret type
3. **Data Mapping**: The same locals block constructs the data map with correct keys per type
4. **Resource Creation**: `main.tf` creates a single `kubernetes_secret_v1` resource
5. **Output Export**: Secret name, namespace, and type are exported

## Type Mapping

| Variable Block | Kubernetes Type | Data Keys |
|---------------|----------------|-----------|
| `opaque` | `Opaque` | User-defined keys from `data` map |
| `tls` | `kubernetes.io/tls` | `tls.crt`, `tls.key` |
| `docker_config_json` | `kubernetes.io/dockerconfigjson` | `.dockerconfigjson` (constructed JSON) |
| `basic_auth` | `kubernetes.io/basic-auth` | `username`, `password` |
| `ssh_auth` | `kubernetes.io/ssh-auth` | `ssh-privatekey` |

## Usage

```hcl
module "secret" {
  source = "./iac/tf"

  metadata = {
    name = "my-secret"
  }

  spec = {
    name      = "my-secret"
    namespace = "production"

    opaque = {
      data = {
        "api-key" = "supersecret"
      }
    }
  }
}
```

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| `metadata` | Resource metadata (name, org, env) | object | yes |
| `spec` | Secret specification with one type variant | object | yes |

## Outputs

| Name | Description |
|------|-------------|
| `secret_name` | Name of the created secret |
| `secret_namespace` | Namespace of the created secret |
| `secret_type` | Kubernetes secret type string |

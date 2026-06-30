# Kubernetes Secret - Pulumi Module

## Overview

This Pulumi module creates and manages a Kubernetes Secret with type-safe data variants. It supports all five common Kubernetes secret types: Opaque, TLS, DockerConfigJson, BasicAuth, and SSHAuth.

## Architecture

```
iac/pulumi/
├── main.go          # Entrypoint: loads stack input, calls module
├── Pulumi.yaml      # Pulumi project configuration
├── Makefile         # Make targets for preview/up/down/refresh
└── module/
    ├── main.go      # Orchestrator: provider init, resource creation, output export
    ├── locals.go    # Derived values: labels, annotations, secret type + data mapping
    ├── secret.go    # Creates kubernetes.core.v1.Secret resource
    └── outputs.go   # Exports secret_name, secret_namespace, secret_type
```

## How It Works

1. **Stack Input Loading**: The entrypoint loads `KubernetesSecretStackInput` from Pulumi config
2. **Locals Initialization**: `locals.go` computes:
   - Standard Planton labels merged with user labels
   - User annotations
   - Kubernetes secret type string from the `oneof secret_data` variant
   - Secret data map with correctly-keyed entries per type
3. **Provider Creation**: Kubernetes provider is initialized from `provider_config`
4. **Secret Creation**: A single `kubernetes.core.v1.Secret` is created with the computed type, data, labels, annotations, and immutability flag
5. **Output Export**: Secret name, namespace, and type are exported as stack outputs

## Type Mapping

| Proto Variant | Kubernetes Type | Data Keys |
|--------------|----------------|-----------|
| `opaque` | `Opaque` | User-defined keys from `data` map |
| `tls` | `kubernetes.io/tls` | `tls.crt`, `tls.key` |
| `docker_config_json` | `kubernetes.io/dockerconfigjson` | `.dockerconfigjson` (constructed JSON) |
| `basic_auth` | `kubernetes.io/basic-auth` | `username`, `password` |
| `ssh_auth` | `kubernetes.io/ssh-auth` | `ssh-privatekey` |

## Usage

```bash
# Preview changes
make preview manifest=../../hack/manifest.yaml

# Deploy
make up manifest=../../hack/manifest.yaml

# Destroy
make down manifest=../../hack/manifest.yaml
```

## Debug

```bash
# Build the module
go build ./module/...

# Build the entrypoint
go build .

# Run tests
go test ./...
```

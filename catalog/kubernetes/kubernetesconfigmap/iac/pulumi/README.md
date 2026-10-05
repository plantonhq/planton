# Kubernetes ConfigMap - Pulumi Module

## Overview

This Pulumi module creates and manages a Kubernetes ConfigMap. It supports the complete ConfigMap surface: UTF-8 `data` entries, base64-encoded `binaryData` entries, and the `immutable` flag.

## Architecture

```
iac/pulumi/
├── main.go          # Entrypoint: loads IaC input, calls module
├── Pulumi.yaml      # Pulumi project configuration
├── Makefile         # Make targets for preview/up/down/refresh
└── module/
    ├── main.go      # Orchestrator: provider init, resource creation, output export
    ├── locals.go    # Derived values: labels, annotations, namespace default, data maps
    ├── configmap.go # Creates kubernetes.core.v1.ConfigMap resource
    └── outputs.go   # Exports configmap_name, namespace
```

## How It Works

1. **IaC Input Loading**: The entrypoint loads `KubernetesConfigMapIacInput` from Pulumi config
2. **Locals Initialization**: `locals.go` computes:
   - Standard Planton labels merged with user labels
   - User annotations
   - The target namespace (foreign-key references are pre-resolved; falls back to `default` when omitted)
   - The `data` and `binary_data` maps (binary values are already base64-encoded and pass through unchanged)
3. **Provider Creation**: Kubernetes provider is initialized from `provider_config`
4. **ConfigMap Creation**: A single `kubernetes.core.v1.ConfigMap` is created with the data, binary data, labels, annotations, and immutability flag
5. **Output Export**: ConfigMap name and namespace are exported as outputs

## Data Handling

| Spec Field | ConfigMap Field | Encoding |
|------------|----------------|----------|
| `data` | `data` | Plain UTF-8 strings, passed through |
| `binary_data` | `binaryData` | Base64 strings, passed through (Kubernetes stores binaryData as base64) |

## Usage

```bash
# Preview changes
make preview manifest=../../e2e/manifest.yaml

# Deploy
make up manifest=../../e2e/manifest.yaml

# Destroy
make down manifest=../../e2e/manifest.yaml
```

## Debug

```bash
# Build the module
go build ./module/...

# Build the entrypoint
go build .
```

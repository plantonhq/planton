# KubernetesService Pulumi Module

## Overview

This Pulumi module creates and manages Kubernetes Service resources based on the `KubernetesServiceSpec` protobuf schema. It supports all four Kubernetes service types (ClusterIP, NodePort, LoadBalancer, ExternalName) plus headless services.

## Usage

### Via OpenMCF CLI (Recommended)

```bash
# Preview changes
openmcf pulumi preview --manifest manifest.yaml

# Deploy
openmcf pulumi up --manifest manifest.yaml

# Tear down
openmcf pulumi down --manifest manifest.yaml
```

### Via Makefile

```bash
make preview manifest=path/to/manifest.yaml
make up manifest=path/to/manifest.yaml
make down manifest=path/to/manifest.yaml
```

### Via debug.sh (Local Development)

```bash
export MANIFEST_PATH=../hack/manifest.yaml
./debug.sh
```

## Stack Input

The module receives a `KubernetesServiceStackInput` containing:

- `target` -- The complete `KubernetesService` resource (metadata + spec)
- `provider_config` -- Kubernetes cluster connection details and credentials

The stack input is loaded from the `stack-input` Pulumi config key, which the OpenMCF CLI sets automatically from the manifest.

## Module Structure

| File | Purpose |
|------|---------|
| `main.go` | Entrypoint, loads stack input and calls module |
| `module/main.go` | Orchestrates provider setup, service creation, and output export |
| `module/locals.go` | Transforms spec fields into Pulumi-friendly values |
| `module/service.go` | Creates the `kubernetes.core/v1.Service` resource |
| `module/outputs.go` | Exports stack outputs (cluster IP, LB address, DNS name) |

## Outputs

| Output Key | Description |
|------------|-------------|
| `service_name` | The created service name |
| `namespace` | The namespace where the service was created |
| `type` | The service type (ClusterIP, NodePort, LoadBalancer, ExternalName) |
| `cluster_ip` | The allocated cluster IP (None for headless, empty for ExternalName) |
| `load_balancer_ingress` | The LB hostname/IP (only for LoadBalancer type) |
| `internal_dns_name` | Fully qualified cluster DNS name |

## Required Pulumi Plugins

- `pulumi-kubernetes` v4.x

## Troubleshooting

- **Provider error**: Ensure `provider_config` contains valid cluster credentials
- **Namespace not found**: The target namespace must already exist in the cluster
- **NodePort conflict**: If a specific node port is in use, let Kubernetes auto-allocate by omitting `node_port`

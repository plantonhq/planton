# KubernetesJob Pulumi Module

This Pulumi module deploys a KubernetesJob to a Kubernetes cluster.

## Overview

The module creates the following Kubernetes resources:

1. **Namespace** (optional) - Created if `createNamespace: true`
2. **ConfigMaps** - From `spec.configMaps`
3. **Secret** - For environment secrets with direct values
4. **Image Pull Secret** (optional) - If Docker credentials are provided
5. **Job** - The main batch workload

## Usage

### Standalone Usage

```bash
# Set the stack input (base64 encoded manifest)
export STACK_INPUT=$(cat manifest.yaml | base64)

# Initialize and deploy
pulumi stack init dev
pulumi up
```

### With Planton CLI

```bash
# Preview changes
planton pulumi preview --manifest job.yaml

# Deploy
planton pulumi up --manifest job.yaml
```

## Required Environment Variables

| Variable | Description |
|----------|-------------|
| `STACK_INPUT` | Base64-encoded KubernetesJobStackInput |

The stack input includes:
- `target` - The KubernetesJob resource definition
- `provider_config` - Kubernetes provider configuration (kubeconfig, context)
- `kubernetes_namespace` - Resolved namespace name
- `docker_config_json` - Docker credentials for private registries (optional)

## Pulumi Plugins

This module requires the following Pulumi plugins:

- `kubernetes` v4.18.1 or later

Install with:
```bash
make install-pulumi-plugins
```

## Module Structure

```
pulumi/
├── main.go          # Entry point
├── Pulumi.yaml      # Project configuration
├── Makefile         # Build and test automation
├── README.md        # This file
├── overview.md      # Architecture overview
└── module/
    ├── main.go           # Resource orchestrator
    ├── locals.go         # Local variables and configuration
    ├── outputs.go        # Stack output exports
    ├── vars.go           # Output variable names
    ├── namespace.go      # Namespace creation
    ├── secret.go         # Secret management
    ├── image_pull_secret.go # Image pull secret
    ├── configmap.go      # ConfigMap creation
    ├── volumes.go        # Volume helper functions
    └── job.go            # Job resource creation
```

## Outputs

| Output | Description |
|--------|-------------|
| `namespace` | The Kubernetes namespace where the job is deployed |
| `job_name` | The name of the created job |

## Troubleshooting

### Job Not Starting

1. Check if the namespace exists (if `createNamespace: false`)
2. Verify image pull credentials are correct
3. Check resource quotas in the namespace

### Job Failing

1. Check pod logs: `kubectl logs job/<job-name> -n <namespace>`
2. Describe the job: `kubectl describe job <job-name> -n <namespace>`
3. Check events: `kubectl get events -n <namespace> --field-selector involvedObject.name=<job-name>`

### Image Pull Errors

1. Verify the image repository and tag
2. Check if image pull secret is correctly configured
3. Ensure the secret has access to the registry

## Development

```bash
# Build the module
make build

# Run tests with hack manifest
make test

# Format code
make fmt

# Run linter
make lint
```

# KubernetesJob Pulumi Module

This Pulumi module deploys a KubernetesJob to a Kubernetes cluster.

## Overview

The module creates the following Kubernetes resources:

1. **Namespace** (optional) - Created if `createNamespace: true`
2. **ConfigMaps** - From `spec.configMaps`
3. **Secret** - For environment secrets with direct values
4. **Image Pull Secret** (optional) - From the registry logins declared on `spec.pod.imageRegistries`
5. **Job** - The main batch workload

## Usage

### Standalone Usage

```bash
# Set the IaC input: a KubernetesJobIacInput YAML file with the manifest under `target`
export IAC_INPUT_YAML_FILE=iac-input.yaml

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

## IaC Input

The module reads a `KubernetesJobIacInput` from the Pulumi config key `planton:iac-input`, or else from one of these environment variables (the planton CLI sets `IAC_INPUT_YAML_FILE` for you):

| Variable | Description |
|----------|-------------|
| `IAC_INPUT_YAML` | `KubernetesJobIacInput` as YAML content |
| `IAC_INPUT_YAML_FILE` | Path to a YAML file holding the `KubernetesJobIacInput` |

The IaC input includes:
- `target` - The KubernetesJob resource definition
- `provider_config` - Kubernetes provider configuration (kubeconfig, context)
- `kubernetes_namespace` - Resolved namespace name

## Pulumi Plugins

This module requires the following Pulumi plugins:

- `kubernetes` v4.18.1 or later

Install with:
```bash
pulumi plugin install resource kubernetes
```

## Module Structure

```
pulumi/
├── main.go          # Entry point
├── Pulumi.yaml      # Project configuration
├── BUILD.bazel      # Bazel build target
├── README.md        # This file
└── module/
    ├── main.go           # Resource orchestrator
    ├── locals.go         # Local variables and configuration
    ├── outputs.go        # Output exports
    ├── namespace.go      # Namespace creation
    ├── secret.go         # Secret management
    ├── image_pull_secret.go # Image pull secret
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
2. Check the login declared on `spec.pod.imageRegistries` (server, username, and the referenced secret), or the Secret named in `spec.pod.imagePullSecrets`
3. Ensure that login has read access to the registry -- or, on a same-cloud registry, that the node identity does

## Development

```bash
# Build the module
go build .

# Preview against the test manifest
planton pulumi preview --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack>

# Format code
gofmt -w .

# Vet
go vet ./...
```

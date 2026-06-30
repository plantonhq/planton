# Kubernetes DaemonSet Pulumi Module

This Pulumi module deploys a DaemonSet to Kubernetes using the KubernetesDaemonSet API resource.

## Overview

The module creates the following Kubernetes resources:

- **Namespace** (optional): Created when `create_namespace` is true
- **DaemonSet**: The main workload controller
- **Secret**: Contains environment secrets
- **Secret**: Image pull secret for private registries (optional)

## Usage

### Prerequisites

1. A Kubernetes cluster
2. Kubernetes credentials configured in the stack input
3. Pulumi installed and configured

### Running Locally

1. Navigate to this directory:
   ```bash
   cd apis/dev/planton/provider/kubernetes/kubernetesdaemonset/v1/iac/pulumi
   ```

2. Install dependencies:
   ```bash
   make deps
   ```

3. Create a stack input file (e.g., `stack-input.yaml`):
   ```yaml
   target:
     apiVersion: kubernetes.planton.dev/v1
     kind: KubernetesDaemonSet
     metadata:
       name: my-daemonset
     spec:
       namespace:
         value: my-namespace
       create_namespace: true
       container:
         app:
           image:
             repo: nginx
             tag: latest
           resources:
             limits:
               cpu: "500m"
               memory: "512Mi"
             requests:
               cpu: "100m"
               memory: "128Mi"
   provider_config:
     kubeconfig_path: ~/.kube/config
   ```

4. Run Pulumi:
   ```bash
   pulumi up
   ```

### Using with Planton CLI

```bash
planton apply -m manifest.yaml
```

## Module Structure

```
.
├── main.go           # Pulumi entrypoint
├── Makefile          # Build and development commands
├── Pulumi.yaml       # Pulumi project configuration
└── module/
    ├── main.go           # Main resource orchestration
    ├── locals.go         # Local variables initialization
    ├── outputs.go        # Output constants
    ├── daemonset.go      # DaemonSet resource creation
    ├── secret.go         # Environment secrets
    └── image_pull_secret.go  # Image pull secret
```

## Outputs

| Name | Description |
|------|-------------|
| namespace | The namespace where resources are deployed |
| daemonset_name | The name of the created DaemonSet |

## Debug Mode

To run with debug logging, uncomment the binary option in `Pulumi.yaml`:

```yaml
runtime:
  name: go
  options:
    binary: ./debug.sh
```

Then create a `debug.sh` script with your debugging configuration.


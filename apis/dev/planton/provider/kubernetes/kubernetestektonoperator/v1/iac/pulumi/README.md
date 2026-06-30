# Kubernetes Tekton Operator - Pulumi Module

This directory contains the Pulumi implementation for deploying the Tekton Operator on Kubernetes.

## Overview

The Pulumi module installs the Tekton Operator using official release manifests. The operator manages the lifecycle of Tekton components (Pipelines, Triggers, Dashboard) via the TektonConfig CRD.

## Module Structure

```
iac/pulumi/
├── main.go                     # Pulumi program entrypoint
├── Pulumi.yaml                 # Project configuration
├── Makefile                    # Build automation
├── debug.sh                    # Debugging helper script
└── module/
    ├── main.go                 # Module orchestration
    ├── locals.go               # Local variables and label management
    ├── outputs.go              # Stack output constants
    ├── vars.go                 # Tekton operator constants
    └── tekton_operator.go      # Operator installation implementation
```

## Prerequisites

- **Pulumi CLI**: Install from [pulumi.com](https://www.pulumi.com/docs/get-started/install/)
- **Go**: Version 1.21 or later
- **Kubernetes Cluster**: Target cluster with kubectl access

## Quick Start

### 1. Configure Stack

Create a new Pulumi stack and configure it:

```bash
cd iac/pulumi
pulumi stack init dev
pulumi config set-secret kubeconfig "$(cat ~/.kube/config)"
```

### 2. Set Input Variables

Create a `stack-input.json` file:

```json
{
  "metadata": {
    "name": "tekton-operator",
    "id": "tekton-op-dev"
  },
  "spec": {
    "target_cluster": {
      "kubernetes_credential_id": "my-cluster-cred"
    },
    "container": {
      "resources": {
        "requests": {
          "cpu": "100m",
          "memory": "128Mi"
        },
        "limits": {
          "cpu": "500m",
          "memory": "512Mi"
        }
      }
    },
    "components": {
      "pipelines": true,
      "triggers": true,
      "dashboard": true
    }
  }
}
```

### 3. Deploy

```bash
pulumi up
```

### 4. Verify Deployment

```bash
kubectl get pods -n tekton-operator
kubectl get tektonconfig
kubectl get pods -n tekton-pipelines
```

## Module Components

### Tekton Operator Installation

Deploys the Tekton Operator from official release manifests using `yaml/v2`:

```go
yamlv2.NewConfigFile(ctx, "tekton-operator", &yamlv2.ConfigFileArgs{
    File: pulumi.String(locals.OperatorReleaseURL),
}, pulumi.Provider(k8sProvider))
```

> **Note**: We use `yaml/v2` instead of the older `yaml` package because v2 provides better CRD ordering and await behavior. This prevents the "no matches for kind TektonConfig" error that can occur when the TektonConfig CRD hasn't been registered yet.

### TektonConfig Configuration

Creates a TektonConfig CRD to configure which components to install:

```go
tektonConfigYAML := `apiVersion: operator.tekton.dev/v1alpha1
kind: TektonConfig
metadata:
  name: config
spec:
  profile: all  # or: basic, lite
  targetNamespace: tekton-pipelines`
```

> **Note**: Do not set fields that the operator manages automatically (e.g., `pipeline.enable-api-fields`) to avoid Server-Side Apply field conflicts.

### Component Profiles

The module selects a profile based on enabled components:

- **all**: Pipelines + Triggers + Dashboard
- **basic**: Pipelines + Triggers
- **lite**: Pipelines only

## Constants (vars.go)

Tekton Operator configuration constants:

```go
var vars = struct {
    OperatorNamespace        string
    ComponentsNamespace      string
    OperatorReleaseURLFormat string
    TektonConfigName         string
}{
    OperatorNamespace:        "tekton-operator",
    ComponentsNamespace:      "tekton-pipelines",
    // Using infra.tekton.dev as per current Tekton operator documentation
    OperatorReleaseURLFormat: "https://infra.tekton.dev/tekton-releases/operator/previous/%s/release.yaml",
    TektonConfigName:         "config",
}
```

The operator version is read from the spec input (`spec.operator_version`) and defaults to `v0.78.0` via proto field options.

> **Note**: The release URL uses `infra.tekton.dev` as per the current [Tekton Operator installation docs](https://tekton.dev/docs/operator/install/).

## Stack Outputs

The module exports the following outputs:

- **namespace**: Kubernetes namespace where Tekton components run (`tekton-pipelines`)
- **tekton_config_name**: Name of the TektonConfig resource
- **pipelines_controller_service**: Pipelines controller service name (if enabled)
- **triggers_controller_service**: Triggers controller service name (if enabled)
- **dashboard_service**: Dashboard service name (if enabled)
- **dashboard_port_forward_command**: Port-forward command for dashboard access

Access outputs:

```bash
pulumi stack output namespace
pulumi stack output dashboard_port_forward_command
```

## Debugging

Use the included debug script to inspect the stack:

```bash
./debug.sh
```

View Pulumi logs:

```bash
pulumi logs
```

## Cleanup

Remove the Tekton Operator deployment:

```bash
pulumi destroy
```

> **Warning**: This will remove the operator AND all Tekton components it manages.

## Common Issues

### Operator Pod Not Starting

**Symptom**: Tekton operator pod remains in Pending state.

**Solution**: Check resource availability:

```bash
kubectl describe pod -n tekton-operator -l app=tekton-operator
kubectl top nodes
```

### TektonConfig Not Ready

**Symptom**: TektonConfig status shows not ready.

**Solution**: Check operator logs:

```bash
kubectl logs -n tekton-operator -l app=tekton-operator
kubectl describe tektonconfig config
```

### Dashboard Not Accessible

**Symptom**: Dashboard service not reachable.

**Solution**: Use port-forwarding:

```bash
kubectl port-forward svc/tekton-dashboard -n tekton-pipelines 9097:9097
```

Then access: http://localhost:9097

### CRD Registration Timing Issues

**Symptom**: Error "no matches for kind TektonConfig in version operator.tekton.dev/v1alpha1"

**Solution**: This is fixed by using `yaml/v2` which handles CRD registration timing properly. If you still encounter this, check that the operator manifests were applied successfully:

```bash
kubectl get crds | grep tekton
kubectl get tektonconfigs
```

## References

- [Pulumi Kubernetes Provider](https://www.pulumi.com/registry/packages/kubernetes/)
- [Pulumi Kubernetes YAML v2](https://www.pulumi.com/blog/kubernetes-yaml-v2/) - Better CRD ordering
- [Tekton Documentation](https://tekton.dev/docs/)
- [Tekton Operator GitHub](https://github.com/tektoncd/operator)
- [Tekton Operator Installation](https://tekton.dev/docs/operator/install/)

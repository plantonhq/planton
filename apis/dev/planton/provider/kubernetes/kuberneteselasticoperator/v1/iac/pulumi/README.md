# Kubernetes Elastic Operator - Pulumi Module

This directory contains the Pulumi implementation for deploying the Elastic Cloud on Kubernetes (ECK) operator.

## Overview

The Pulumi module installs the ECK operator using the official Elastic Helm chart. The operator extends Kubernetes with Custom Resource Definitions (CRDs) for managing Elasticsearch, Kibana, APM Server, and other Elastic Stack components.

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
    ├── outputs.go              # Stack outputs
    ├── vars.go                 # ECK operator constants
    └── kubernetes_elastic_operator.go  # Helm release implementation
```

## Prerequisites

- **Pulumi CLI**: Install from [pulumi.com](https://www.pulumi.com/docs/get-started/install/)
- **Go**: Version 1.21 or later
- **Kubernetes Cluster**: Target cluster with kubectl access
- **Helm**: Helm 3.x (used by Pulumi Kubernetes provider)

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
    "name": "eck-operator",
    "id": "eck-op-dev"
  },
  "spec": {
    "target_cluster": {
      "kubernetes_credential_id": "my-cluster-cred"
    },
    "namespace": {
      "value": "elastic-system"
    },
    "create_namespace": true,
    "container": {
      "resources": {
        "requests": {
          "cpu": "50m",
          "memory": "100Mi"
        },
        "limits": {
          "cpu": "1000m",
          "memory": "1Gi"
        }
      }
    }
  }
}
```

**Namespace Configuration:**
- `namespace.value`: The name of the namespace (default: `elastic-system`)
- `create_namespace`: Set to `true` to create the namespace, or `false` to use an existing one

### 3. Deploy

```bash
pulumi up
```

### 4. Verify Deployment

```bash
kubectl get pods -n elastic-system
kubectl get crds | grep elastic
```

## Module Components

### Namespace Management

The module supports two modes of namespace management:

**1. Creating the Namespace (create_namespace: true)**

When `spec.create_namespace` is `true`, the module creates the namespace with Planton labels:

```go
if locals.KubernetesElasticOperator.Spec.CreateNamespace {
    ns, err := corev1.NewNamespace(ctx, namespace, &corev1.NamespaceArgs{
        Metadata: metav1.ObjectMetaPtrInput(&metav1.ObjectMetaArgs{
            Name:   pulumi.String(namespace),
            Labels: pulumi.ToStringMap(locals.KubeLabels),
        }),
    }, pulumi.Provider(k8sProvider))
}
```

**2. Using Existing Namespace (create_namespace: false)**

When `spec.create_namespace` is `false`, the module assumes the namespace already exists and only references it by name. This is useful when:
- The namespace is managed by another component (e.g., KubernetesNamespace)
- The namespace has specific resource quotas or policies managed externally
- You want to deploy multiple operators into a shared namespace

### Helm Release

Deploys the ECK operator via Helm chart:

```go
helm.NewRelease(ctx, "kubernetes-elastic-operator", &helm.ReleaseArgs{
    Name:            pulumi.String(vars.HelmChartName),
    Namespace:       ns.Metadata.Name(),
    Chart:           pulumi.String(vars.HelmChartName),
    Version:         pulumi.String(vars.HelmChartVersion),
    RepositoryOpts:  helm.RepositoryOptsArgs{Repo: pulumi.String(vars.HelmChartRepo)},
    CreateNamespace: pulumi.Bool(false),
    Atomic:          pulumi.Bool(false),
    CleanupOnFail:   pulumi.Bool(true),
    WaitForJobs:     pulumi.Bool(true),
    Timeout:         pulumi.Int(180),
    Values:          values,
}, pulumi.Parent(ns))
```

### Label Inheritance

Configures ECK to propagate Planton labels to all managed resources:

```go
values := pulumi.Map{
    "configKubernetes": pulumi.Map{
        "inherited_labels": pulumi.ToStringArray([]string{
            kuberneteslabelkeys.Resource,
            kuberneteslabelkeys.Organization,
            kuberneteslabelkeys.Environment,
            kuberneteslabelkeys.ResourceKind,
            kuberneteslabelkeys.ResourceId,
        }),
    },
}
```

### Resource Configuration

Passes container resource specifications to Helm chart:

```go
if cr := locals.KubernetesElasticOperator.Spec.GetContainer().GetResources(); cr != nil {
    res := pulumi.Map{}
    if lim := cr.GetLimits(); lim != nil {
        res["limits"] = pulumi.StringMap{
            "cpu":    pulumi.String(lim.Cpu),
            "memory": pulumi.String(lim.Memory),
        }
    }
    if req := cr.GetRequests(); req != nil {
        res["requests"] = pulumi.StringMap{
            "cpu":    pulumi.String(req.Cpu),
            "memory": pulumi.String(req.Memory),
        }
    }
    if len(res) > 0 {
        values["resources"] = res
    }
}
```

## Constants (vars.go)

ECK operator configuration constants:

```go
var vars = struct {
    Namespace        string
    HelmChartName    string
    HelmChartRepo    string
    HelmChartVersion string
}{
    Namespace:        "elastic-system",
    HelmChartName:    "eck-operator",
    HelmChartRepo:    "https://helm.elastic.co",
    HelmChartVersion: "2.14.0",
}
```

### Upgrading ECK Version

To upgrade the ECK operator version, edit `vars.go` and update `HelmChartVersion`:

```go
HelmChartVersion: "2.15.0",  // new version
```

Then run:

```bash
pulumi up
```

## Stack Outputs

The module exports the following outputs:

- **namespace**: Kubernetes namespace where ECK operator is deployed (`elastic-system`)

Access outputs:

```bash
pulumi stack output namespace
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

Remove the ECK operator deployment:

```bash
pulumi destroy
```

> **Warning**: This will remove the operator but NOT the Elastic Stack resources it manages. Delete Elasticsearch, Kibana, and other custom resources manually before destroying the operator.

## Common Issues

### CRDs Not Installing

**Symptom**: Elasticsearch/Kibana resources fail to create with "CRD not found" errors.

**Solution**: Verify CRDs are installed:

```bash
kubectl get crds | grep elastic
```

If missing, the Helm chart may have failed. Check Pulumi logs:

```bash
pulumi logs | grep CRD
```

### Operator Pod Not Starting

**Symptom**: ECK operator pod remains in Pending or CrashLoopBackOff state.

**Solution**: Check resource availability and pod events:

```bash
kubectl describe pod -n elastic-system -l control-plane=elastic-operator
kubectl top nodes
```

### Permission Issues

**Symptom**: Operator logs show permission denied errors.

**Solution**: Verify RBAC resources were created:

```bash
kubectl get clusterrole elastic-operator
kubectl get clusterrolebinding elastic-operator
kubectl get serviceaccount -n elastic-system elastic-operator
```

## Development

### Running Locally

```bash
cd iac/pulumi
pulumi up --stack dev
```

### Testing

```bash
# Deploy to test cluster
pulumi up --stack test

# Verify deployment
kubectl get all -n elastic-system

# Check operator logs
kubectl logs -n elastic-system -l control-plane=elastic-operator

# Cleanup
pulumi destroy --stack test
```

## References

- [Pulumi Kubernetes Provider](https://www.pulumi.com/registry/packages/kubernetes/)
- [Pulumi Helm Release](https://www.pulumi.com/registry/packages/kubernetes/api-docs/helm/v3/release/)
- [Elastic ECK Documentation](https://www.elastic.co/guide/en/cloud-on-k8s/current/index.html)
- [ECK Helm Chart](https://github.com/elastic/cloud-on-k8s/tree/main/deploy/eck-operator)


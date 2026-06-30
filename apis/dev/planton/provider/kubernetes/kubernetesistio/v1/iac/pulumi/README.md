# Kubernetes Istio - Pulumi Module

This Pulumi module deploys Istio service mesh on Kubernetes clusters using the official Istio Helm charts with proper component ordering and resource management.

## Overview

The module provides a simplified interface for deploying Istio with:

- **Complete Component Stack**: Automatic deployment of base, istiod, and ingress gateway
- **Resource Configuration**: Tunable CPU and memory for the control plane
- **Proper Ordering**: Ensures correct installation sequence (base → istiod → gateway)
- **Namespace Management**: Creates dedicated namespaces for control plane and gateway
- **Production Ready**: Uses atomic Helm releases with automatic rollback

## Key Features

1. **Automated Component Deployment**
   - Deploys Istio base (CRDs and foundational resources)
   - Installs istiod control plane with resource customization
   - Creates ingress gateway for external traffic

2. **Helm Chart Based**
   - Uses official Istio Helm charts
   - Pinned chart version for reproducibility (1.22.3)
   - Atomic releases with rollback on failure

3. **Resource Management**
   - Configurable CPU and memory for istiod
   - Default values suitable for production
   - Scalable from development to enterprise

4. **Dual Namespace Architecture**
   - `istio-system` for control plane components
   - `istio-ingress` for ingress gateway (security best practice)
   - Configurable namespace creation via `CreateNamespace` flag

## Module Structure

- `main.go` - Primary resource orchestration and Helm deployments
- `vars.go` - Chart repositories, versions, and namespace constants
- `outputs.go` - Stack output constants

## Usage

### Basic Example

```go
import (
    kubernetesistiov1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesistio/v1"
    "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesistio/v1/iac/pulumi/module"
    "github.com/plantonhq/planton/apis/dev/planton/shared"
    "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        istioSpec := &kubernetesistiov1.KubernetesIstioSpec{
            TargetCluster: &kubernetes.KubernetesClusterSelector{
                ClusterName: "my-gke-cluster",
            },
            Namespace: &foreignkeyv1.StringValueOrRef{
                LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
                    Value: "istio-system",
                },
            },
            CreateNamespace: true,
            Container: &kubernetesistiov1.KubernetesIstioSpecContainer{
                Resources: &kubernetes.ContainerResources{
                    Requests: &kubernetes.CpuMemory{
                        Cpu:    "50m",
                        Memory: "100Mi",
                    },
                    Limits: &kubernetes.CpuMemory{
                        Cpu:    "1000m",
                        Memory: "1Gi",
                    },
                },
            },
        }

        istio := &kubernetesistiov1.KubernetesIstio{
            ApiVersion: "kubernetes.planton.dev/v1",
            Kind:       "KubernetesIstio",
            Metadata: &shared.CloudResourceMetadata{
                Name: "main-istio",
            },
            Spec: istioSpec,
        }

        stackInput := &kubernetesistiov1.KubernetesIstioStackInput{
            Target: istio,
        }

        if err := module.Resources(ctx, stackInput); err != nil {
            return err
        }

        return nil
    })
}
```

### Production Configuration

```go
istioSpec := &kubernetesistiov1.KubernetesIstioSpec{
    TargetCluster: &kubernetes.KubernetesClusterSelector{
        ClusterName: "prod-gke-cluster",
    },
    Namespace: &foreignkeyv1.StringValueOrRef{
        LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
            Value: "istio-system",
        },
    },
    CreateNamespace: true,
    Container: &kubernetesistiov1.KubernetesIstioSpecContainer{
        Resources: &kubernetes.ContainerResources{
            Requests: &kubernetes.CpuMemory{
                Cpu:    "500m",
                Memory: "512Mi",
            },
            Limits: &kubernetes.CpuMemory{
                Cpu:    "2000m",
                Memory: "2Gi",
            },
        },
    },
}
```

### High-Availability Configuration

```go
istioSpec := &kubernetesistiov1.KubernetesIstioSpec{
    TargetCluster: &kubernetes.KubernetesClusterSelector{
        ClusterName: "prod-ha-gke-cluster",
    },
    Namespace: &foreignkeyv1.StringValueOrRef{
        LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
            Value: "istio-system",
        },
    },
    CreateNamespace: true,
    Container: &kubernetesistiov1.KubernetesIstioSpecContainer{
        Resources: &kubernetes.ContainerResources{
            Requests: &kubernetes.CpuMemory{
                Cpu:    "1000m",
                Memory: "1Gi",
            },
            Limits: &kubernetes.CpuMemory{
                Cpu:    "4000m",
                Memory: "8Gi",
            },
        },
    },
}
```

## Stack Outputs

The module exports the following outputs:

- `namespace` - Namespace where Istio control plane is deployed (istio-system)
- `service` - Name of the istiod service
- `port_forward_command` - Command to port-forward to istiod for debugging
- `kube_endpoint` - Kubernetes service endpoint for istiod
- `ingress_endpoint` - Kubernetes service endpoint for ingress gateway

Access outputs after deployment:

```bash
pulumi stack output namespace
pulumi stack output service
pulumi stack output port_forward_command
```

Use port-forward command:

```bash
# Get the command
FORWARD_CMD=$(pulumi stack output port_forward_command)

# Execute it
eval $FORWARD_CMD

# Access debug interface at http://localhost:15014/debug
```

## Debugging

Use the provided debug script:

```bash
cd iac/pulumi
./debug.sh
```

Or manually:

```bash
cd iac/pulumi
export PULUMI_CONFIG_PASSPHRASE=""
pulumi stack select dev
pulumi preview
```

To see detailed Helm release information:

```bash
pulumi preview --diff
```

## Prerequisites

- Pulumi CLI installed
- Go 1.21 or later
- Access to Kubernetes cluster
- kubectl configured
- Sufficient cluster resources for Istio components

## Cluster Requirements

### Minimum Requirements (Development)
- 2 CPUs available
- 4GB RAM available
- Kubernetes 1.25+

### Recommended (Production)
- 4+ CPUs available
- 8GB+ RAM available
- Kubernetes 1.26+
- Multiple nodes for HA

### Enterprise Scale
- 8+ CPUs available
- 16GB+ RAM available
- Kubernetes 1.27+
- Dedicated node pool for Istio components

## Deployment Flow

The module orchestrates the following deployment sequence:

1. **Namespace Creation**
   - Creates `istio-system` namespace
   - Creates `istio-ingress` namespace
   - Applies resource labels

2. **Base Installation**
   - Deploys Istio base Helm chart
   - Installs CRDs
   - Creates foundational resources
   - Waits for completion (timeout: 180s)

3. **Control Plane Deployment**
   - Deploys istiod Helm chart
   - Configures pilot resources from spec
   - Ensures atomic installation
   - Depends on base completion

4. **Gateway Deployment**
   - Deploys ingress gateway Helm chart
   - Configures as ClusterIP service
   - Depends on istiod completion

5. **Output Export**
   - Exports all stack outputs
   - Generates utility commands

## Verification

After deployment, verify Istio installation:

```bash
# Check control plane
kubectl get pods -n istio-system
kubectl get svc -n istio-system

# Check ingress gateway
kubectl get pods -n istio-ingress
kubectl get svc -n istio-ingress

# View Helm releases
helm list -n istio-system
helm list -n istio-ingress

# Check Istio version
kubectl get pods -n istio-system -l app=istiod -o jsonpath='{.items[0].spec.containers[0].image}'
```

Verify with istioctl (if installed):

```bash
istioctl version
istioctl verify-install
```

## Namespace Management

The module manages two Istio namespaces:

- **istio-system**: Control plane namespace (istiod, base CRDs)
- **istio-ingress**: Ingress gateway namespace

### Automatic Namespace Creation

Set `CreateNamespace: true` (default behavior):

```go
istioSpec := &kubernetesistiov1.KubernetesIstioSpec{
    CreateNamespace: true,  // Creates both namespaces automatically
    // ... other fields
}
```

The module will create both `istio-system` and `istio-ingress` namespaces.

### Using Existing Namespaces

Set `CreateNamespace: false` to use pre-existing namespaces:

```go
istioSpec := &kubernetesistiov1.KubernetesIstioSpec{
    CreateNamespace: false,  // Use existing namespaces
    // ... other fields
}
```

**Prerequisites:**
- Both `istio-system` and `istio-ingress` namespaces must already exist
- Namespaces should have appropriate labels and RBAC configured

**Verification:**

Before deploying with `CreateNamespace: false`, verify namespaces exist:

```bash
kubectl get namespace istio-system istio-ingress
```

## Troubleshooting

### Namespace Not Found Error

If you see errors like "namespace not found" during deployment:

1. Check if `CreateNamespace` is set correctly:
   ```bash
   # If CreateNamespace: false, verify namespaces exist
   kubectl get namespace istio-system istio-ingress
   ```

2. If namespaces don't exist, either:
   - Set `CreateNamespace: true` in your spec
   - Or create the namespaces manually:
     ```bash
     kubectl create namespace istio-system
     kubectl create namespace istio-ingress
     ```

### Control Plane Not Starting

If istiod pods fail to start:

1. Check resource constraints:
   ```bash
   kubectl describe pod -n istio-system <istiod-pod>
   ```

2. Review events:
   ```bash
   kubectl get events -n istio-system --sort-by='.lastTimestamp'
   ```

3. Check logs:
   ```bash
   kubectl logs -n istio-system -l app=istiod
   ```

4. Increase resources in spec and redeploy

### Helm Release Failed

If Helm release fails:

```bash
# Check Helm release status
helm status base -n istio-system
helm status istiod -n istio-system
helm status gateway -n istio-ingress

# Get release history
helm history istiod -n istio-system
```

The module uses `atomic: true` so failed releases are automatically rolled back.

### Gateway Not Accessible

If ingress gateway isn't working:

```bash
# Check gateway pods
kubectl get pods -n istio-ingress
kubectl logs -n istio-ingress -l app=istio-gateway

# Verify service
kubectl get svc -n istio-ingress

# Check for configuration issues
kubectl describe svc -n istio-ingress istio-gateway
```

### Resource Quota Issues

If deployment fails due to resource quotas:

```bash
# Check cluster resource availability
kubectl top nodes

# Check namespace quotas
kubectl get resourcequota -n istio-system
kubectl get resourcequota -n istio-ingress
```

## Post-Deployment Configuration

### Enable Automatic Sidecar Injection

```bash
kubectl label namespace default istio-injection=enabled
kubectl label namespace production istio-injection=enabled
```

### Deploy Sample Application

```bash
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.22/samples/bookinfo/platform/kube/bookinfo.yaml
```

### Configure Traffic Management

Create a Gateway:

```yaml
apiVersion: networking.istio.io/v1beta1
kind: Gateway
metadata:
  name: bookinfo-gateway
  namespace: default
spec:
  selector:
    istio: gateway
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - "*"
```

Apply configuration:

```bash
kubectl apply -f gateway.yaml
```

## Upgrading Istio

To upgrade Istio to a newer version:

1. Update chart version in `vars.go`
2. Test in non-production environment
3. Apply upgrade:
   ```bash
   pulumi up
   ```

The module handles the upgrade sequence automatically.

## Additional Resources

- [Pulumi Kubernetes Documentation](https://www.pulumi.com/registry/packages/kubernetes/)
- [Istio Official Documentation](https://istio.io/latest/docs/)
- [Istio Helm Charts](https://github.com/istio/istio/tree/master/manifests/charts)
- [Istio Performance and Scalability](https://istio.io/latest/docs/ops/deployment/performance-and-scalability/)
- [Module Overview](overview.md) - Detailed architecture and design decisions


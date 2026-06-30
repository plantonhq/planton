# Kubernetes Ingress NGINX - Pulumi Module

This Pulumi module deploys the official NGINX Ingress Controller on Kubernetes clusters with cloud-specific optimizations for GKE, EKS, and AKS.

## Overview

The module provides a simplified interface for deploying ingress controllers with:

- **Multi-Cloud Support**: Automatic configuration for GCP, AWS, and Azure
- **Load Balancer Integration**: Native cloud load balancer setup
- **Internal/External Control**: Single flag to toggle private vs public access
- **Static IP Support**: Cloud-specific static IP assignment
- **Security Integration**: Security groups (AWS), managed identities (Azure)

## Key Features

1. **Automated Cloud Configuration**
   - Detects cloud provider from spec
   - Applies correct annotations automatically
   - Configures load balancer type (internal/external)

2. **Helm Chart Deployment**
   - Uses official NGINX Ingress Controller Helm chart
   - Configurable chart version
   - Default to tested stable version

3. **Service Setup**
   - Creates LoadBalancer service
   - Configures ingress class as default
   - Watches ingresses without explicit class

4. **Namespace Management**
   - **Flexible namespace control**: Create new or use existing namespaces
   - **User-specified names**: Configure custom namespace names
   - **Label propagation**: Applies resource labels for organization
   - **RBAC integration**: Works with existing namespace policies

## Module Structure

- `main.go` - Primary resource orchestration
- `locals.go` - Local variable initialization and transformations
- `outputs.go` - Stack output constants
- `vars.go` - Deployment constants (namespace, chart repo, versions)

## Usage

### Basic Example

```go
import (
    kubernetesingressnginxv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesingressnginx/v1"
    "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesingressnginx/v1/iac/pulumi/module"
    "github.com/plantonhq/planton/apis/dev/planton/shared"
    foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
    "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        ingressSpec := &kubernetesingressnginxv1.KubernetesIngressNginxSpec{
            TargetCluster: &kubernetes.KubernetesClusterSelector{
                ClusterName: "my-cluster",
            },
            Namespace: &foreignkeyv1.StringValueOrRef{
                LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
                    Value: "ingress-nginx",
                },
            },
            CreateNamespace: true,
            ChartVersion:    "4.11.1",
            Internal:        false,
        }

        ingress := &kubernetesingressnginxv1.KubernetesIngressNginx{
            ApiVersion: "kubernetes.planton.dev/v1",
            Kind:       "KubernetesIngressNginx",
            Metadata: &shared.CloudResourceMetadata{
                Name: "my-ingress",
            },
            Spec: ingressSpec,
        }

        stackInput := &kubernetesingressnginxv1.KubernetesIngressNginxStackInput{
            Target: ingress,
        }

        if err := module.Resources(ctx, stackInput); err != nil {
            return err
        }

        return nil
    })
}
```

### GKE with Static IP

```go
ingressSpec := &kubernetesingressnginxv1.KubernetesIngressNginxSpec{
    TargetCluster: &kubernetes.KubernetesAddonTargetCluster{
        CredentialSource: &kubernetes.KubernetesAddonTargetCluster_KubernetesCredentialId{
            KubernetesCredentialId: "gke-cluster-credential",
        },
    },
    ChartVersion: "4.11.1",
    Internal:     false,
    ProviderConfig: &kubernetesingressnginxv1.KubernetesIngressNginxSpec_Gke{
        Gke: &kubernetesingressnginxv1.KubernetesIngressNginxGkeConfig{
            StaticIpName: "my-ingress-static-ip",
        },
    },
}
```

### EKS Internal with Subnets

```go
ingressSpec := &kubernetesingressnginxv1.KubernetesIngressNginxSpec{
    TargetCluster: &kubernetes.KubernetesAddonTargetCluster{
        CredentialSource: &kubernetes.KubernetesAddonTargetCluster_KubernetesCredentialId{
            KubernetesCredentialId: "eks-cluster-credential",
        },
    },
    ChartVersion: "4.11.1",
    Internal:     true,
    ProviderConfig: &kubernetesingressnginxv1.KubernetesIngressNginxSpec_Eks{
        Eks: &kubernetesingressnginxv1.KubernetesIngressNginxEksConfig{
            SubnetIds: []*foreignkeyv1.StringValueOrRef{
                {Value: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-abc123"}},
                {Value: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-def456"}},
            },
        },
    },
}
```

## Stack Outputs

The module exports the following outputs:

- `namespace` - Kubernetes namespace where ingress controller is deployed
- `release_name` - Helm release name
- `service_name` - Kubernetes service name for the controller
- `service_type` - Service type (LoadBalancer)

Access outputs after deployment:

```bash
pulumi stack output namespace
pulumi stack output service_name
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

## Prerequisites

- Pulumi CLI installed
- Go 1.21 or later
- Access to Kubernetes cluster
- kubectl configured

## Cloud-Specific Prerequisites

### GKE
- Reserved static IP (if using `static_ip_name`)
- GKE cluster with workload identity enabled

### EKS
- IAM role for IRSA (if using `irsa_role_arn_override`)
- Security groups created
- Subnets configured

### AKS
- User-assigned managed identity (if using)
- Public IP resource (if using `public_ip_name`)

## Verification

After deployment, verify the ingress controller:

```bash
# Check pods
kubectl get pods -n kubernetes-ingress-nginx

# Check service and load balancer
kubectl get svc -n kubernetes-ingress-nginx

# View Helm release
helm list -n kubernetes-ingress-nginx

# Check ingress class
kubectl get ingressclass
```

## Troubleshooting

### Load Balancer Pending

If the service stays in Pending state:

1. Check cloud provider quotas
2. Verify subnet configuration (for internal LBs)
3. Check security groups allow ingress traffic
4. Review controller logs:
   ```bash
   kubectl logs -n kubernetes-ingress-nginx -l app.kubernetes.io/component=controller
   ```

### Static IP Not Assigned

For GKE static IP issues:

```bash
# Verify IP exists
gcloud compute addresses list

# Check IP is in correct region
gcloud compute addresses describe <ip-name> --global
```

## Additional Resources

- [Pulumi Kubernetes Documentation](https://www.pulumi.com/registry/packages/kubernetes/)
- [NGINX Ingress Controller](https://kubernetes.github.io/ingress-nginx/)
- [Helm Chart Values](https://github.com/kubernetes/ingress-nginx/blob/main/charts/ingress-nginx/values.yaml)


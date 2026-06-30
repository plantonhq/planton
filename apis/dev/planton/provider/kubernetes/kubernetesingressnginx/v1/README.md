# Kubernetes Ingress NGINX

## Overview

The **KubernetesIngressNginx** API resource provides a unified interface for deploying the official NGINX Ingress Controller across different Kubernetes platforms. This resource abstracts away cloud-specific configuration complexities, allowing you to deploy ingress controllers with consistent behavior across GKE, EKS, AKS, and other Kubernetes clusters.

## Why We Created This API Resource

Managing ingress controllers across multiple cloud providers presents several challenges:

1. **Cloud-Specific Configuration**: Each cloud provider (GCP, AWS, Azure) requires different annotations, settings, and configurations for load balancers
2. **Inconsistent Behavior**: Default deployments often result in different networking setups across clouds
3. **Complex Setup**: Setting up static IPs, internal load balancers, and cloud-native integrations requires deep platform knowledge
4. **Maintenance Burden**: Keeping Helm values files synchronized across environments is error-prone
5. **Lack of Standardization**: No unified approach to deploying the same ingress controller across hybrid/multi-cloud setups

The KubernetesIngressNginx resource solves these problems by providing a single, consistent API that:

- **Abstracts Cloud Differences**: Automatically applies correct annotations and configurations for each cloud provider
- **Simplifies Deployment**: Define your requirements once, deploy anywhere
- **Integrates Seamlessly**: Leverages cloud-native features (static IPs, managed identities, security groups) through simple configuration
- **Ensures Consistency**: Same ingress behavior across all your Kubernetes clusters
- **Reduces Complexity**: No need to manage different Helm value files for different clouds

## Key Features

### Multi-Cloud Support

Deploy ingress controllers with native cloud integrations:

- **Google Kubernetes Engine (GKE)**
  - Static IP reservation and assignment
  - Internal/external load balancer configuration
  - Subnetwork specification for internal LBs
  
- **Amazon Elastic Kubernetes Service (EKS)**
  - Network Load Balancer (NLB) integration
  - Security group management
  - Subnet placement control
  - IAM Roles for Service Accounts (IRSA) support
  
- **Azure Kubernetes Service (AKS)**
  - Azure Load Balancer integration
  - Workload Identity support
  - Public IP resource reuse
  - Internal load balancer configuration

- **Generic Kubernetes**
  - Works with any Kubernetes distribution
  - Standard LoadBalancer service type
  - Cloud-agnostic deployment

### Internal vs External Load Balancers

Control load balancer exposure with a single flag:

- **External** (default): Public load balancer accessible from the internet
- **Internal**: Private load balancer accessible only within your network

This setting automatically applies the correct cloud-specific annotations and configurations.

### Namespace Management

The component provides flexible namespace management:

- **Automatic Creation**: Set `create_namespace: true` to let the module create the namespace
- **Use Existing**: Set `create_namespace: false` to use a pre-existing namespace
- **Custom Names**: Specify any namespace name via the `namespace` field

This flexibility allows you to:
- Use namespace-scoped RBAC and policies
- Deploy to shared namespaces managed by platform teams
- Integrate with namespace lifecycle management tools

### Version Management

- Specify exact Helm chart versions for reproducible deployments
- Default to tested stable versions when version is not specified
- Easy upgrades by changing the chart version

### Credential Management

Flexible authentication options:

- **Direct Credential**: Use Kubernetes cluster credentials directly
- **Cluster Selector**: Reference a cluster in the same environment

## How It Works

The KubernetesIngressNginx resource:

1. **Accepts Your Configuration**: You provide basic settings (cluster, internal/external, cloud-specific options)
2. **Resolves Cloud Settings**: Automatically determines the correct annotations and Helm values for your target cloud
3. **Deploys via Helm**: Uses the official NGINX Ingress Controller Helm chart
4. **Configures Load Balancer**: Sets up the LoadBalancer service with appropriate cloud integrations
5. **Exports Outputs**: Provides namespace, service name, and other deployment details

## Benefits

### For Platform Engineers

- **Unified API**: One resource type for all clouds
- **Reduced Complexity**: No need to learn cloud-specific ingress configurations
- **Faster Deployments**: Pre-configured templates for common patterns
- **Better Governance**: Consistent security and networking policies

### For Development Teams

- **Simple Interface**: Minimal configuration required
- **Cloud Agnostic**: Same YAML works across environments
- **Production Ready**: Built-in best practices
- **Well Documented**: Comprehensive examples and guides

### For Organizations

- **Multi-Cloud Strategy**: Deploy consistently across AWS, GCP, Azure
- **Cost Optimization**: Internal load balancers reduce data transfer costs
- **Security**: Fine-grained control over network exposure
- **Compliance**: Standardized configurations meet security requirements

## Quick Start

### Basic External Ingress Controller

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesIngressNginx
metadata:
  name: my-ingress
spec:
  target_cluster:
    cluster_name: my-cluster
  namespace:
    value: ingress-nginx
  create_namespace: true
  chart_version: "4.11.1"
  internal: false
```

### Internal Ingress Controller on GKE

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesIngressNginx
metadata:
  name: internal-ingress
spec:
  target_cluster:
    cluster_name: gke-prod-cluster
  namespace:
    value: ingress-nginx
  create_namespace: true
  chart_version: "4.11.1"
  internal: true
  gke:
    subnetwork_self_link: projects/my-project/regions/us-west1/subnetworks/my-subnet
```

### EKS with Specific Subnets

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesIngressNginx
metadata:
  name: eks-ingress
spec:
  target_cluster:
    cluster_name: eks-prod-cluster
  namespace:
    value: ingress-nginx
  create_namespace: true
  chart_version: "4.11.1"
  eks:
    subnet_ids:
      - value: subnet-abc123
      - value: subnet-def456
```

## Use Cases

1. **API Gateways**: Route traffic to microservices
2. **Web Applications**: Expose web apps with SSL/TLS termination
3. **Multi-Tenant**: Separate internal and external ingress controllers
4. **Hybrid Cloud**: Consistent ingress across on-prem and cloud
5. **Development/Staging**: Quick ingress setup for non-production environments
6. **Production**: HA ingress with cloud-native load balancer features

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Internet / VPN                          │
└──────────────────────────┬──────────────────────────────────┘
                           │
                  ┌────────▼────────┐
                  │  Cloud Load     │  (GCP LB / AWS NLB / Azure LB)
                  │   Balancer      │
                  └────────┬────────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                                     │
┌───────▼────────┐                  ┌────────▼────────┐
│  NGINX Ingress │                  │  NGINX Ingress  │
│  Controller    │                  │  Controller     │
│   (Pod 1)      │                  │   (Pod 2)       │
└───────┬────────┘                  └────────┬────────┘
        │                                     │
        └──────────────────┬──────────────────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
┌───────▼────────┐ ┌──────▼──────┐ ┌────────▼────────┐
│   Service 1    │ │  Service 2  │ │   Service 3     │
│   (Backend)    │ │  (Backend)  │ │   (Backend)     │
└────────────────┘ └─────────────┘ └─────────────────┘
```

## Next Steps

1. Review [examples.md](./examples.md) for comprehensive deployment scenarios
2. Check [docs/README.md](./docs/README.md) for detailed architecture and design decisions
3. Explore [iac/pulumi/](./iac/pulumi/) for Pulumi deployment module
4. Explore [iac/tf/](./iac/tf/) for Terraform deployment module

## Additional Resources

- [NGINX Ingress Controller Official Documentation](https://kubernetes.github.io/ingress-nginx/)
- [NGINX Ingress Controller Helm Chart](https://github.com/kubernetes/ingress-nginx/tree/main/charts/ingress-nginx)
- [Kubernetes Ingress Documentation](https://kubernetes.io/docs/concepts/services-networking/ingress/)


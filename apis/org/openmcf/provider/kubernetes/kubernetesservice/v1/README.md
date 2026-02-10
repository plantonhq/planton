# KubernetesService

## Overview

**KubernetesService** is an OpenMCF deployment component that provides declarative management of Kubernetes Service resources. It wraps the Kubernetes Service primitive as a first-class deployment unit, enabling platform engineers and application developers to manage service discovery, load balancing, and external access through a standardized, version-controlled manifest.

While workload components like `KubernetesDeployment` and `KubernetesStatefulSet` automatically create Services for their pods, `KubernetesService` addresses use cases where standalone service management is required -- services decoupled from workload lifecycle, services pointing to external endpoints, or services targeting pods managed by other tools.

## Purpose

KubernetesService simplifies the creation and management of Kubernetes Services by:

- **Standardizing service definitions** -- Consistent manifest format across all environments
- **Enabling GitOps workflows** -- Services defined in version control, deployed through CI/CD
- **Abstracting cloud-specific complexity** -- Annotations for AWS NLB, GCP NEG, Azure ILB are just manifest fields
- **Providing dual IaC support** -- Deploy with Pulumi or Terraform, same manifest either way

## Key Features

- **All service types** -- ClusterIP, NodePort, LoadBalancer, and ExternalName
- **Headless services** -- Direct pod addressing for StatefulSets and custom DNS resolution
- **External traffic policy** -- Control source IP preservation vs even load distribution
- **Session affinity** -- Sticky sessions via ClientIP for stateful applications
- **Load balancer source ranges** -- IP-based access control for LoadBalancer services
- **Cloud-provider annotations** -- Full support for cloud-specific load balancer configuration
- **Named and numeric target ports** -- Flexible pod targeting by port name or number

## Example Usage

### Basic ClusterIP Service

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesService
metadata:
  name: api-service
spec:
  namespace: production
  name: api-service
  selector:
    app: api
    tier: backend
  ports:
    - name: http
      port: 80
      target_port: "8080"
```

Deploy with:

```bash
openmcf pulumi up --manifest service.yaml
```

### LoadBalancer with AWS NLB

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesService
metadata:
  name: web-lb
spec:
  namespace: production
  name: web-lb
  type: load_balancer
  selector:
    app: web
  ports:
    - name: https
      port: 443
      target_port: "8443"
  external_traffic_policy: local
  load_balancer_source_ranges:
    - "10.0.0.0/8"
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: "nlb"
```

## Best Practices

- **Use ClusterIP** for internal-only services (the default and most common type)
- **Use `external_traffic_policy: local`** for LoadBalancer services when source IP preservation is important
- **Use headless services** for StatefulSet DNS resolution rather than creating per-pod services
- **Always name ports** when a service exposes more than one port
- **Use annotations** for cloud-specific load balancer configuration rather than separate resources
- **Keep selectors minimal** -- match on `app` and optionally `version` or `tier`

## Further Reading

- [Examples](examples.md) -- Complete, copy-paste ready examples for all service types
- [Research Documentation](docs/README.md) -- Deep dive into Kubernetes Service landscape and design decisions

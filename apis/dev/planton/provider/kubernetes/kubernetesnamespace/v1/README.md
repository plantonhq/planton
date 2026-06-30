# Kubernetes Namespace

## Overview

**KubernetesNamespace** is a Planton deployment component that implements a "Namespace-as-a-Service" pattern for creating and managing production-ready Kubernetes namespaces. Rather than creating a bare namespace, this component provisions a complete, secure, multi-tenant environment with resource quotas, network policies, access controls, and optional service mesh integration pre-configured according to best practices.

## Purpose

A Kubernetes namespace by itself is merely a logical partition. To function as a true platform primitive for multi-tenancy, it requires careful configuration of:

- **Resource Quotas** to prevent resource contention and cost runaway
- **LimitRanges** to provide sensible defaults for containers
- **NetworkPolicies** to enforce zero-trust networking
- **RBAC** to control access
- **Pod Security Standards** to enforce security posture
- **Service Mesh Integration** for observability and traffic management

This component abstracts all that complexity into a simple, declarative API that follows the 80/20 principle: exposing the 20% of configuration options that deliver 80% of the value.

## Key Features

### 1. Resource Profiles

Instead of manually configuring ResourceQuotas and LimitRanges, choose from T-shirt sized profiles:

- **SMALL**: 2-4 CPU cores, 4-8Gi memory (dev/test)
- **MEDIUM**: 4-8 CPU cores, 8-16Gi memory (staging)
- **LARGE**: 8-16 CPU cores, 16-32Gi memory (production)
- **XLARGE**: 16-32 CPU cores, 32-64Gi memory (high-scale production)

Or specify custom quotas for precise control over CPU, memory, and object counts (pods, services, configmaps, secrets).

### 2. Network Isolation

Implements "Default Deny" security pattern with:

- **Ingress Isolation**: Block all incoming traffic by default
- **Egress Restriction**: Allow only DNS and Kubernetes API by default
- **Explicit Allow Lists**: Configure specific namespaces, CIDR blocks, or domains that can communicate

### 3. Service Mesh Integration

Automatic sidecar injection support for:

- **Istio**: With revision tags for safe canary upgrades
- **Linkerd**: Lightweight service mesh
- **Consul Connect**: HashiCorp service mesh

### 4. Pod Security Standards

Enforce Kubernetes-native security policies:

- **Privileged**: Unrestricted (system workloads only)
- **Baseline**: Prevents known privilege escalations
- **Restricted**: Production-grade pod hardening

### 5. Cost Allocation & Governance

- Automatic label application for cost tracking
- Support for janitor/TTL annotations for ephemeral environments
- Custom labels and annotations for organizational requirements

## Essential Configuration Fields

### Required

- **`spec.name`**: The namespace name (DNS-compliant: lowercase alphanumeric and hyphens)

### Common

- **`spec.resource_profile`**: Choose preset (SMALL/MEDIUM/LARGE/XLARGE) or define custom quotas
- **`spec.network_config`**: Enable network isolation (ingress/egress controls)
- **`spec.service_mesh_config`**: Configure service mesh (type, revision tag)
- **`spec.pod_security_standard`**: Set security enforcement level
- **`spec.labels`**: Additional labels for cost allocation and governance
- **`spec.annotations`**: Annotations for mesh injection, TTL, node selection

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

- **`namespace`**: The created namespace name
- **`resource_quotas_applied`**: Whether quotas were configured
- **`limit_ranges_applied`**: Whether default limits were set
- **`network_policies_applied`**: Whether network isolation is active
- **`service_mesh_enabled`**: Whether service mesh injection is enabled
- **`service_mesh_type`**: The configured mesh type (if enabled)
- **`pod_security_standard`**: The enforced security level
- **`labels_json`**: JSON representation of applied labels
- **`annotations_json`**: JSON representation of applied annotations

## How It Works

This component includes both **Pulumi** (Go) and **Terraform** (HCL) modules that:

1. Create the Kubernetes Namespace resource
2. Apply ResourceQuota objects based on the selected profile
3. Configure LimitRange for default container resource requests/limits
4. Create NetworkPolicy objects for ingress/egress control
5. Apply service mesh labels/annotations for automatic sidecar injection
6. Configure Pod Security Standards via namespace labels
7. Export observable outputs for service discovery

Both IaC implementations have feature parity and follow the same logic, ensuring consistent behavior regardless of which tool you use.

## When to Use

Use **KubernetesNamespace** when you need:

- ✅ **Multi-tenant Kubernetes clusters** with resource isolation
- ✅ **Cost allocation** by team, project, or environment
- ✅ **Zero-trust networking** with default-deny policies
- ✅ **Service mesh integration** with automatic sidecar injection
- ✅ **Pod security enforcement** at the namespace level
- ✅ **Ephemeral environments** (with TTL annotations for automatic cleanup)
- ✅ **Standardized namespace configuration** across environments

## Use Cases

1. **Team Namespaces**: Allocate dedicated namespaces to development teams with resource quotas to prevent noisy neighbor issues
2. **Environment Isolation**: Create separate namespaces for dev, staging, and production with appropriate resource allocations and security controls
3. **Microservices Deployment**: Pre-configure namespaces with service mesh and network policies before deploying applications
4. **Cost Management**: Track cloud costs by namespace using labels and quotas
5. **Compliance**: Enforce pod security standards and network isolation for regulated workloads

## Prerequisites

- **Kubernetes Cluster**: Access to a Kubernetes cluster (any distribution: GKE, EKS, AKS, self-hosted)
- **Credentials**: Kubernetes cluster credentials (kubeconfig)
- **CNI Plugin** (optional): For network policies, ensure your cluster has a CNI that supports NetworkPolicy (Calico, Cilium, AWS VPC CNI, etc.)
- **Service Mesh** (optional): If using service mesh integration, the mesh control plane must be pre-installed

## Best Practices

1. **Start with Preset Profiles**: Use SMALL/MEDIUM/LARGE profiles initially, customize only when needed
2. **Enable Network Isolation**: For production, always enable `isolate_ingress` and `restrict_egress`
3. **Use Meaningful Labels**: Add `team`, `environment`, `cost-center` labels for governance
4. **Service Mesh Revision Tags**: For Istio, use revision tags (e.g., "prod-stable") instead of hardcoding versions
5. **TTL for Ephemeral Environments**: Add `janitor/ttl: "24h"` annotation for PR environments
6. **Test Security Standards**: Start with BASELINE, move to RESTRICTED after validating workload compatibility

## References

- [Kubernetes Namespaces Documentation](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/)
- [Resource Quotas](https://kubernetes.io/docs/concepts/policy/resource-quotas/)
- [Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
- [Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
- [Multi-Tenancy Best Practices](https://kubernetes.io/docs/concepts/security/multi-tenancy/)
- [Istio Sidecar Injection](https://istio.io/latest/docs/setup/additional-setup/sidecar-injection/)
- [Namespace-as-a-Service Pattern](https://docs.rafay.co/template_catalog/get_started/namespace_asaservice/)

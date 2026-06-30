# KubernetesService: Research Documentation

## Introduction

The Kubernetes Service is one of the most fundamental networking primitives in the Kubernetes ecosystem. It provides a stable network identity and load balancing for a set of pods, abstracting away the ephemeral nature of individual pod IPs. Services are the standard mechanism for service discovery, inter-service communication, and external access to workloads running in a Kubernetes cluster.

This document provides a comprehensive analysis of the Kubernetes Service landscape, explains why Planton offers a standalone Service component, and justifies the design decisions behind the `KubernetesServiceSpec` schema.

## The Kubernetes Service Primitive

### What is a Kubernetes Service?

A Kubernetes Service is an abstraction that defines a logical set of pods and a policy by which to access them. Services enable:

- **Service Discovery**: Pods can find each other by service name through cluster DNS (CoreDNS)
- **Load Balancing**: Traffic is distributed across all healthy pods matching the service selector
- **Stable Endpoints**: A service maintains a consistent IP and DNS name regardless of pod churn
- **External Access**: LoadBalancer and NodePort types expose services outside the cluster

### Service Types

Kubernetes defines four service types, each serving different network topology needs:

| Type | Description | Cluster IP | External Access | Use Case |
|------|-------------|------------|-----------------|----------|
| ClusterIP | Internal only | Allocated | No | Inter-service communication |
| NodePort | Static port on nodes | Allocated | Via node IP + port | Dev/test, bare metal |
| LoadBalancer | Cloud provider LB | Allocated | Via LB IP/hostname | Production external access |
| ExternalName | DNS CNAME alias | None | N/A (DNS redirect) | External service proxying |

Additionally, the **headless** variant (clusterIP: None) provides direct pod IP resolution through DNS, which is essential for StatefulSets and custom service discovery.

### Why a Standalone Service Component?

Workload components in Planton (`KubernetesDeployment`, `KubernetesStatefulSet`) automatically create Services for their pods. However, a standalone `KubernetesService` component addresses scenarios that fall outside the workload lifecycle:

1. **ExternalName services**: Proxying to external DNS names (e.g., RDS endpoints, external APIs) without any pods
2. **Services without selectors**: Manually managed Endpoints pointing to non-Kubernetes backends
3. **Headless services**: For StatefulSets or custom DNS-based discovery mechanisms managed by separate components
4. **Cross-tool integration**: Services for pods managed by Helm charts, operators, or other tools outside Planton
5. **LoadBalancer services**: Adding external access to existing internal workloads without modifying the workload component
6. **Service migration**: Gradually migrating traffic between different backends by managing the service independently

## Deployment Methods

### Manual (kubectl)

The most basic approach. Services are created via YAML manifests and `kubectl apply`. Simple but lacks drift detection, version control integration, and automation.

```bash
kubectl apply -f service.yaml
```

**Limitations**: No state management, no dependency tracking, difficult to audit or roll back.

### Helm Charts

Helm wraps service definitions within chart templates. Services are typically a supporting resource within a larger application chart, not managed independently.

**Limitations**: Tight coupling to the chart lifecycle, templating complexity, hard to manage standalone services.

### Kustomize

Kustomize overlays allow environment-specific service customization. Good for multi-environment deployments but requires understanding of the overlay system.

**Limitations**: No state management, limited cross-resource dependency handling.

### Terraform (hashicorp/kubernetes provider)

Terraform manages Kubernetes resources declaratively with state tracking. The `kubernetes_service_v1` resource provides full service configuration.

**Strengths**: State management, plan/apply workflow, integration with cloud provider resources.

### Pulumi (pulumi-kubernetes)

Pulumi uses real programming languages (Go, TypeScript, Python) to create Kubernetes resources. The `kubernetes.core.v1.Service` resource mirrors the Kubernetes API.

**Strengths**: Type safety, IDE support, conditional logic, reusable components.

### Planton Approach

Planton provides a unified manifest format that works with both Pulumi and Terraform, plus:
- Protocol Buffer schema with compile-time validation
- Consistent KRM structure across all components
- Built-in credential management and provider configuration
- Dual IaC support with feature parity

## 80/20 Scoping Decision

### In-Scope Fields (Covers 80%+ of Use Cases)

| Field | Rationale |
|-------|-----------|
| `type` | Fundamental -- every service has a type |
| `selector` | Core routing mechanism -- matches pods by label |
| `ports` | Essential -- defines what the service exposes |
| `headless` | Common pattern for StatefulSets and service discovery |
| `external_dns_name` | Required for ExternalName type |
| `external_traffic_policy` | Critical for source IP preservation in production |
| `session_affinity` | Common for stateful applications |
| `load_balancer_source_ranges` | Security best practice for LoadBalancer |
| `annotations` | Essential for cloud-specific LB configuration |
| `labels` | Standard Kubernetes governance mechanism |

### Out-of-Scope Fields

| Field | Rationale for Exclusion |
|-------|------------------------|
| `ipFamilies` / `ipFamilyPolicy` | Dual-stack networking is used by <5% of clusters |
| `allocateLoadBalancerNodePorts` | Niche optimization, defaults work for 99% of cases |
| `loadBalancerClass` | Emerging feature, not widely adopted yet |
| `internalTrafficPolicy` | Rarely changed from default (Cluster) |
| `topologyKeys` (deprecated) | Removed in Kubernetes 1.22+ |
| `publishNotReadyAddresses` | Very specialized use case (Consul, custom DNS) |
| `healthCheckNodePort` | Auto-managed when externalTrafficPolicy=Local |

### Design Decisions

**`headless` as boolean vs raw `clusterIP`**: We use a boolean `headless` flag rather than exposing the raw `clusterIP` field. This provides a cleaner UX (`headless: true` vs `cluster_ip: "None"`) and prevents users from accidentally setting invalid ClusterIP values. The boolean clearly communicates intent.

**`target_port` as string**: Kubernetes allows `targetPort` to be either a number or a named port reference. We use `string` to support both formats (`"8080"` or `"http"`), matching Kubernetes API behavior.

**`external_dns_name` instead of `external_name`**: The field was named `external_dns_name` instead of `external_name` to avoid a protobuf scoping conflict with the `external_name` enum value (protobuf uses C++ scoping where enum values exist in the enclosing scope). The name is also more descriptive -- it explicitly communicates that this is a DNS name.

**Enum conventions**: Service types use lowercase values (`cluster_ip`, `node_port`, `load_balancer`, `external_name`) following Planton enum guidelines. Protocol values (`TCP`, `UDP`, `SCTP`) use uppercase as they represent external standards where uppercase is conventional.

## Production Best Practices

1. **Default to ClusterIP**: Only use NodePort or LoadBalancer when external access is genuinely needed
2. **Use `external_traffic_policy: local`** for LoadBalancer services when source IP preservation matters (security logging, geo-routing)
3. **Name all ports** when exposing multiple ports -- required by Kubernetes and improves readability
4. **Use annotations for LB configuration** rather than separate cloud resources -- keeps configuration co-located
5. **Set `load_balancer_source_ranges`** for any public-facing LoadBalancer to restrict access
6. **Use headless services for StatefulSets** to enable per-pod DNS resolution
7. **Use ExternalName services** instead of hardcoding external endpoints in application config

## Common Pitfalls

1. **Forgetting `external_traffic_policy: local`**: Source IP is masked with the default `cluster` policy, breaking IP-based security rules
2. **NodePort range conflicts**: Node ports must be in 30000-32767; let Kubernetes auto-allocate when possible
3. **ExternalName and selectors**: ExternalName services must not have a selector; they only provide DNS CNAME records
4. **Headless + NodePort/LoadBalancer**: Headless services are incompatible with these types -- there's no ClusterIP to front
5. **Missing port names**: When a service has multiple ports, all ports must be named -- Kubernetes requires this

## Conclusion

The `KubernetesService` component provides a clean, validated interface for managing Kubernetes Services as standalone deployment units. By focusing on the 80/20 of service configuration, it delivers a practical tool that covers the vast majority of real-world use cases while maintaining the simplicity and consistency that Planton is built on. The cross-field validation rules (ExternalName requires DNS name, headless incompatible with NodePort/LoadBalancer, ports required for non-ExternalName) catch configuration errors at schema validation time, long before they reach the cluster.

# Overview

The **KubernetesManifest** API resource provides a generic, flexible way to deploy any raw Kubernetes manifests onto a Kubernetes cluster. This deployment component is designed for scenarios where you need to deploy arbitrary Kubernetes resources—including CRDs, Custom Resources, multi-resource applications, or any YAML that doesn't fit neatly into more specialized deployment components.

## Purpose

While Planton provides specialized deployment components for common workloads (Deployments, StatefulSets, DaemonSets, Helm releases), there are many scenarios where you need to deploy custom or third-party Kubernetes manifests:

- **Third-party operator CRDs and resources**: Deploy operator Custom Resources without wrapping them in Helm charts
- **One-off configurations**: ConfigMaps, Secrets, ServiceAccounts, RBAC resources
- **Multi-resource applications**: Deploy a complete application stack defined in a single YAML file
- **Vendor-provided manifests**: Apply manifests from third-party vendors exactly as provided
- **Testing and prototyping**: Quickly deploy raw YAML during development

The KubernetesManifest API resource aims to:

- **Provide Maximum Flexibility**: Accept any valid Kubernetes YAML, single or multi-document
- **Maintain Consistency**: Use the same target cluster and namespace patterns as other deployment components
- **Simplify Operations**: One API resource to manage complex multi-resource deployments
- **Support GitOps**: Integrate seamlessly with CI/CD pipelines and GitOps workflows

## Key Features

### Raw YAML Support

- **Single Manifests**: Deploy a single Kubernetes resource
- **Multi-Document YAML**: Deploy multiple resources separated by `---` in a single API call
- **Any Resource Type**: Deployments, Services, ConfigMaps, CRDs, Custom Resources, and more

### Target Cluster Selection

- **Environment-Aware**: Deploy to specific Kubernetes clusters in your environment
- **Multi-Cloud Support**: Works with GKE, EKS, AKS, DigitalOcean, Civo, and any Kubernetes cluster

### Namespace Management

- **Namespace Configuration**: Specify the default namespace for resources without explicit namespaces
- **Namespace Creation Control**: Use the `create_namespace` flag to control namespace lifecycle:
  - **`create_namespace: true`**: Creates the namespace before applying manifests
  - **`create_namespace: false`**: Uses an existing namespace (must exist)

### Smart Resource Ordering

- **CRD-Aware**: Uses Pulumi's yamlv2 which provides intelligent resource ordering
- **Dependency Resolution**: CRDs are applied before Custom Resources that depend on them
- **Await Behavior**: Properly waits for resources to be ready before proceeding

## Benefits

- **Zero Abstraction Overhead**: Deploy exactly what you specify—no transformation or interpretation
- **Escape Hatch**: When specialized components don't fit your use case, use raw manifests
- **Operator Integration**: Perfect for deploying operator CRDs and Custom Resources
- **Rapid Prototyping**: Test new configurations quickly without creating new components
- **Vendor Compatibility**: Apply vendor-provided manifests without modification

## Use Cases

- **Operator Deployments**: Install operators and their Custom Resources
- **Infrastructure Resources**: NetworkPolicies, ResourceQuotas, LimitRanges
- **RBAC Configuration**: Roles, RoleBindings, ClusterRoles, ClusterRoleBindings
- **Custom Applications**: Deploy complete application stacks from raw YAML
- **Migration Assistance**: Migrate existing YAML-based deployments to Planton
- **Testing**: Quickly deploy test resources during development

## Comparison with Other Components

| Feature | KubernetesManifest | KubernetesHelmRelease | KubernetesDeployment |
|---------|-------------------|----------------------|---------------------|
| **Input Format** | Raw YAML | Helm Chart | Structured API |
| **Templating** | None | Helm values | None (API defaults) |
| **Best For** | Custom/one-off resources | Packaged applications | Microservices |
| **Flexibility** | Maximum | High | Medium |
| **Abstraction** | None | Some | High |


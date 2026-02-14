---
title: "Jenkins"
description: "Jenkins deployment documentation"
icon: "package"
order: 100
componentName: "kubernetesjenkins"
---

# Kubernetes Jenkins

Deploys Jenkins on Kubernetes using the official Jenkins Helm chart. Provisions admin credentials automatically, supports resource tuning via container limits/requests, allows arbitrary Helm value overrides, and optionally exposes Jenkins externally through Istio Gateway API ingress with TLS termination and HTTP-to-HTTPS redirect.

## What Gets Created

When you deploy a KubernetesJenkins resource, OpenMCF provisions:

- **Kubernetes Namespace** — created if `createNamespace` is `true`
- **Admin Credentials Secret** — a Kubernetes Secret containing a randomly generated 12-character admin password (includes uppercase, lowercase, numeric, and special characters)
- **Jenkins Helm Release** — the official `jenkins` chart (v5.1.5) from `https://charts.jenkins.io`, which creates:
  - A Jenkins controller pod running image tag `2.454-jdk17`
  - Kubernetes Service for cluster-internal access on port 8080
  - Persistent storage and configuration managed by the chart
- **Ingress Resources** (when `ingress.enabled` is `true`):
  - cert-manager Certificate for TLS, issued by a ClusterIssuer matching the ingress domain
  - Gateway API Gateway with HTTPS (port 443) and HTTP (port 80) listeners
  - HTTPRoute for HTTPS traffic forwarding to the Jenkins service
  - HTTPRoute for HTTP-to-HTTPS 301 redirect

## Prerequisites

- **A Kubernetes cluster** with kubectl configured for access
- **Istio ingress gateway** installed (only if using ingress)
- **cert-manager** with a ClusterIssuer matching your ingress domain (only if using ingress)
- **Gateway API CRDs** installed in the cluster (only if using ingress)

## Quick Start

Create a file `jenkins.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesJenkins
metadata:
  name: my-jenkins
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesJenkins.my-jenkins
spec:
  namespace:
    value: jenkins-dev
  createNamespace: true
```

Deploy:

```shell
openmcf apply -f jenkins.yaml
```

This creates a Jenkins instance with default resources (1 CPU / 1Gi memory limit, 50m CPU / 100Mi memory request) in the `jenkins-dev` namespace. An admin user is created automatically with a generated password stored in a Kubernetes Secret.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `StringValueOrRef` | Kubernetes namespace for the Jenkins deployment. Use `value` for a direct string or `valueFrom` to reference a KubernetesNamespace resource. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `createNamespace` | `bool` | `false` | Create the namespace if it does not exist. |
| `containerResources.limits.cpu` | `string` | `"1000m"` | CPU limit for the Jenkins controller container. |
| `containerResources.limits.memory` | `string` | `"1Gi"` | Memory limit for the Jenkins controller container. |
| `containerResources.requests.cpu` | `string` | `"50m"` | CPU request for the Jenkins controller container. |
| `containerResources.requests.memory` | `string` | `"100Mi"` | Memory request for the Jenkins controller container. |
| `helmValues` | `map<string, string>` | `{}` | Additional Helm chart values for customization. See the [Jenkins Helm chart values](https://github.com/jenkinsci/helm-charts/blob/main/charts/jenkins/values.yaml) for available options. |
| `ingress.enabled` | `bool` | `false` | Enable external access via Istio Gateway API ingress with TLS. |
| `ingress.hostname` | `string` | -- | Full hostname for external access (e.g., `jenkins.example.com`). Required when `ingress.enabled` is `true`. |

## Examples

### Jenkins with Custom Resources

Increase CPU and memory allocations for a busier Jenkins instance:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesJenkins
metadata:
  name: ci-jenkins
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesJenkins.ci-jenkins
spec:
  namespace:
    value: ci-tools
  createNamespace: true
  containerResources:
    limits:
      cpu: "2000m"
      memory: "4Gi"
    requests:
      cpu: "500m"
      memory: "1Gi"
```

### Jenkins with Helm Value Overrides

Use `helmValues` to configure plugins, JVM options, or any chart setting:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesJenkins
metadata:
  name: custom-jenkins
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.KubernetesJenkins.custom-jenkins
spec:
  namespace:
    value: jenkins-staging
  createNamespace: true
  containerResources:
    limits:
      cpu: "2000m"
      memory: "4Gi"
    requests:
      cpu: "250m"
      memory: "512Mi"
  helmValues:
    controller.javaOpts: "-Xms512m -Xmx2g"
    controller.numExecutors: "4"
    controller.installPlugins: "git:latest,pipeline-stage-view:latest,blueocean:latest"
```

### Full-Featured with Ingress

External access over HTTPS with automatic TLS and HTTP redirect:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesJenkins
metadata:
  name: prod-jenkins
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesJenkins.prod-jenkins
spec:
  namespace:
    value: production
  createNamespace: true
  containerResources:
    limits:
      cpu: "4000m"
      memory: "8Gi"
    requests:
      cpu: "1000m"
      memory: "2Gi"
  helmValues:
    controller.javaOpts: "-Xms1g -Xmx4g"
    controller.numExecutors: "8"
    persistence.size: "50Gi"
  ingress:
    enabled: true
    hostname: jenkins.example.com
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where Jenkins was created |
| `service` | `string` | Name of the Kubernetes service for Jenkins |
| `port_forward_command` | `string` | Ready-to-run `kubectl port-forward` command for local access on port 8080 |
| `kube_endpoint` | `string` | Cluster-internal endpoint (e.g., `my-jenkins.jenkins-dev.svc.cluster.local`) |
| `external_hostname` | `string` | External hostname when ingress is enabled (e.g., `jenkins.example.com`) |
| `internal_hostname` | `string` | Internal hostname for private ingress (e.g., `internal-jenkins.example.com`) |
| `username` | `string` | Jenkins admin username (default: `admin`) |
| `password_secret` | `KubernetesSecretKey` | Reference to the Kubernetes Secret containing the admin password (`name` = secret name, `key` = `jenkins-admin-password`) |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/namespace) — pre-create a namespace to reference via `valueFrom`
- [KubernetesPostgres](/docs/catalog/kubernetes/postgres) — deploy PostgreSQL for Jenkins pipeline data or external storage
- [KubernetesRedis](/docs/catalog/kubernetes/redis) — deploy Redis for caching in CI/CD pipelines

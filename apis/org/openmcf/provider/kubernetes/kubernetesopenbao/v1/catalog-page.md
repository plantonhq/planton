# Kubernetes OpenBao

Deploys OpenBao on Kubernetes using the official OpenBao Helm chart. OpenBao is an open-source secrets management solution forked from HashiCorp Vault, providing secure secret storage, dynamic secrets, data encryption, leasing/renewal, and revocation. Supports standalone and high-availability (HA) deployment modes with Raft integrated storage, optional ingress for external access, agent injector for automatic sidecar secret injection, and a built-in web UI.

## What Gets Created

When you deploy a KubernetesOpenBao resource, OpenMCF provisions:

- **Kubernetes Namespace** — created if `createNamespace` is `true`
- **OpenBao Helm Release** — the `openbao` chart from `openbao.github.io/openbao-helm`, which creates:
  - StatefulSet running OpenBao server pods (standalone or HA mode)
  - Persistent Volume Claims sized to `serverContainer.dataStorageSize` for data storage
  - Kubernetes Service for cluster-internal access on port 8200
  - ConfigMap with the HCL listener and storage configuration
- **HA Raft Cluster** — when `highAvailability.enabled` is `true`, deploys multiple replicas using Raft consensus for leader election and replicated storage
- **Agent Injector** — when `injector.enabled` is `true`, deploys a mutating webhook that automatically injects OpenBao Agent sidecar containers into annotated pods
- **Ingress** — when `ingress.enabled` is `true`, creates an Ingress resource for external access with optional TLS termination
- **Web UI** — enabled by default, accessible at the service endpoint on port 8200

## Prerequisites

- **A Kubernetes cluster** with a default StorageClass for persistent volumes
- **kubectl** configured to access the target cluster
- **Ingress controller** (e.g., nginx, traefik) installed in the cluster if using ingress
- **cert-manager** or a pre-provisioned TLS secret if enabling TLS on ingress

## Quick Start

Create a file `openbao.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesOpenBao
metadata:
  name: my-openbao
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesOpenBao.my-openbao
spec:
  namespace:
    value: openbao-dev
  createNamespace: true
```

Deploy:

```shell
openmcf apply -f openbao.yaml
```

This creates a single-replica OpenBao instance in standalone mode with 10Gi storage, the web UI enabled, and TLS disabled, running in the `openbao-dev` namespace.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `StringValueOrRef` | Kubernetes namespace for the OpenBao deployment. Use `value` for a direct string or `valueFrom` to reference a KubernetesNamespace resource. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `createNamespace` | `bool` | `false` | Create the namespace if it does not exist. |
| `helmChartVersion` | `string` | `"0.23.3"` | Override the OpenBao Helm chart version. |
| `serverContainer.replicas` | `int` | `1` | Number of OpenBao server replicas. Range: 1-10. For HA mode use 3 or more (odd numbers recommended). |
| `serverContainer.resources.limits.cpu` | `string` | `"500m"` | CPU limit for the OpenBao server container. |
| `serverContainer.resources.limits.memory` | `string` | `"256Mi"` | Memory limit for the OpenBao server container. |
| `serverContainer.resources.requests.cpu` | `string` | `"100m"` | CPU request for the OpenBao server container. |
| `serverContainer.resources.requests.memory` | `string` | `"128Mi"` | Memory request for the OpenBao server container. |
| `serverContainer.dataStorageSize` | `string` | `"10Gi"` | Persistent volume size for OpenBao data. Must match pattern like `10Gi`, `50Gi`, `100Gi`. |
| `highAvailability.enabled` | `bool` | `false` | Enable HA mode with Raft integrated storage for leader election and replicated data. |
| `highAvailability.replicas` | `int` | `3` | Number of HA replicas. Must be an odd number (3, 5, 7) for Raft consensus. Range: 3-10. |
| `ingress.enabled` | `bool` | `false` | Expose OpenBao externally via an Ingress resource. |
| `ingress.hostname` | `string` | — | Hostname for external access (e.g., `openbao.example.com`). Required when `ingress.enabled` is `true`. |
| `ingress.ingressClassName` | `string` | — | Ingress class name (e.g., `nginx`, `traefik`). Uses the cluster default if not set. |
| `ingress.tlsEnabled` | `bool` | `false` | Enable TLS termination at the ingress controller. |
| `ingress.tlsSecretName` | `string` | — | Kubernetes secret containing the TLS certificate. Required when `ingress.tlsEnabled` is `true` unless cert-manager annotations are used. |
| `uiEnabled` | `bool` | `true` | Enable the OpenBao web UI. |
| `injector.enabled` | `bool` | `false` | Deploy the OpenBao Agent Injector mutating webhook for automatic sidecar injection. |
| `injector.replicas` | `int` | `1` | Number of injector replicas. Range: 1-5. |
| `tlsEnabled` | `bool` | `false` | Enable TLS for OpenBao server listeners. When `false`, the Helm chart sets `tlsDisable: true`. |

## Examples

### Standalone with Custom Resources

A single-replica deployment with tuned CPU/memory and larger storage:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesOpenBao
metadata:
  name: dev-openbao
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesOpenBao.dev-openbao
spec:
  namespace:
    value: secrets
  createNamespace: true
  serverContainer:
    replicas: 1
    resources:
      limits:
        cpu: "1000m"
        memory: "512Mi"
      requests:
        cpu: "200m"
        memory: "256Mi"
    dataStorageSize: "20Gi"
```

### High Availability with Raft

A production HA deployment with 5 Raft replicas and the agent injector enabled:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesOpenBao
metadata:
  name: prod-openbao
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesOpenBao.prod-openbao
spec:
  namespace:
    value: openbao-prod
  createNamespace: true
  serverContainer:
    replicas: 5
    resources:
      limits:
        cpu: "2000m"
        memory: "1Gi"
      requests:
        cpu: "500m"
        memory: "512Mi"
    dataStorageSize: "50Gi"
  highAvailability:
    enabled: true
    replicas: 5
  injector:
    enabled: true
    replicas: 2
```

### Full-Featured with Ingress and TLS

External access with TLS termination at the ingress, the web UI enabled, and the agent injector running:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesOpenBao
metadata:
  name: full-openbao
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesOpenBao.full-openbao
spec:
  namespace:
    value: openbao-production
  createNamespace: true
  serverContainer:
    replicas: 3
    resources:
      limits:
        cpu: "2000m"
        memory: "1Gi"
      requests:
        cpu: "500m"
        memory: "256Mi"
    dataStorageSize: "100Gi"
  highAvailability:
    enabled: true
    replicas: 3
  ingress:
    enabled: true
    hostname: openbao.example.com
    ingressClassName: nginx
    tlsEnabled: true
    tlsSecretName: openbao-tls
  uiEnabled: true
  injector:
    enabled: true
    replicas: 2
  tlsEnabled: false
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where OpenBao was deployed |
| `service` | `string` | Name of the Kubernetes service for OpenBao |
| `portForwardCommand` | `string` | Ready-to-run `kubectl port-forward` command for local access on port 8200 |
| `kubeEndpoint` | `string` | Cluster-internal FQDN endpoint (e.g., `my-openbao.openbao-dev.svc.cluster.local:8200`) |
| `externalHostname` | `string` | External hostname when ingress is enabled |
| `rootTokenSecret` | `KubernetesSecretKey` | Reference to the Kubernetes secret containing the OpenBao root token |
| `unsealKeysSecret` | `KubernetesSecretKey` | Reference to the Kubernetes secret containing the OpenBao unseal keys |
| `clusterAddress` | `string` | Internal cluster communication address for HA mode (e.g., `https://my-openbao-0.my-openbao-internal:8201`) |
| `apiAddress` | `string` | OpenBao API address (e.g., `http://my-openbao.openbao-dev.svc.cluster.local:8200`) |
| `haEnabled` | `string` | Whether the deployment is running in HA mode (`true` or `false`) |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — pre-create a namespace to reference via `valueFrom`
- [KubernetesSecret](/docs/catalog/kubernetes/kubernetessecret) — manage Kubernetes secrets that OpenBao can inject into pods
- [KubernetesHelmRelease](/docs/catalog/kubernetes/kuberneteshelmrelease) — deploy additional Helm charts alongside OpenBao

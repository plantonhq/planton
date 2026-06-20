---
title: "StatefulSet"
description: "StatefulSet deployment documentation"
icon: "package"
order: 100
componentName: "kubernetesstatefulset"
---

# Kubernetes StatefulSet

Deploys a stateful application to Kubernetes as a StatefulSet with a headless service for stable pod identity, per-pod persistent volume claims via volume claim templates, an optional ClusterIP client service, configurable ingress via Gateway API, environment variable and secret management, and pod disruption budget controls.

## What Gets Created

When you deploy a KubernetesStatefulSet resource, OpenMCF provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **ServiceAccount** — a dedicated service account for the StatefulSet's pods
- **Headless Service** — a ClusterIP `None` service named `{metadata.name}-headless` for stable network identity; each pod is addressable at `{pod-name}.{headless-service}.{namespace}.svc.cluster.local`
- **StatefulSet** — a Kubernetes StatefulSet with the specified container image, resource limits, probes, volume mounts, volume claim templates, and pod management policy
- **Client Service** — a ClusterIP Service named `{metadata.name}` providing load-balanced access to StatefulSet pods, created only when at least one port is defined in `container.app.ports`
- **ConfigMaps** — one ConfigMap per entry in `configMaps`, created before the StatefulSet to allow volume mount references
- **Secret** — an Opaque Secret containing environment secrets provided as direct string values, created only when `container.app.env.secrets` includes direct values
- **Image Pull Secret** — a `kubernetes.io/dockerconfigjson` Secret for pulling from private registries, created only when a Docker config JSON is provided
- **Certificate** — a cert-manager Certificate for TLS termination on ingress hostnames, created only when ingress is enabled
- **External Gateway** — a Gateway API Gateway for traffic from outside the VPC, with HTTPS (port 443) and HTTP (port 80) listeners, created only when ingress is enabled
- **Internal Gateway** — a Gateway API Gateway for traffic from within the VPC, with HTTPS and HTTP listeners on an `internal-` prefixed hostname, created only when ingress is enabled
- **HTTPRoutes** — four HTTPRoute resources handling external HTTP-to-HTTPS redirect, external HTTPS routing, internal HTTP-to-HTTPS redirect, and internal HTTPS routing to the client Service, created only when ingress is enabled
- **PodDisruptionBudget** — ensures minimum pod availability during voluntary disruptions, created only when `availability.podDisruptionBudget.enabled` is `true`

## Prerequisites

- **Kubernetes credentials** configured via environment variables or OpenMCF provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **A container image** accessible from the cluster (public registry or with a configured image pull secret)
- **cert-manager with a ClusterIssuer** if enabling ingress — the ClusterIssuer name must match the domain extracted from the ingress hostname
- **Gateway API CRDs and an Istio-based gateway controller** if enabling ingress

## Quick Start

Create a file `statefulset.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesStatefulSet
metadata:
  name: my-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesStatefulSet.my-db
spec:
  namespace: my-namespace
  createNamespace: true
  container:
    app:
      image:
        repo: postgres
        tag: "16"
      ports:
        - name: postgres
          containerPort: 5432
          networkProtocol: TCP
          appProtocol: tcp
          servicePort: 5432
  volumeClaimTemplates:
    - name: data
      size: 10Gi
```

Deploy:

```shell
openmcf apply -f statefulset.yaml
```

This creates a single-replica PostgreSQL StatefulSet with a headless service, a client service on port 5432, and a 10Gi persistent volume per pod in the `my-namespace` namespace, using default resource limits (1000m CPU, 1Gi memory).

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `string` | Kubernetes namespace for the StatefulSet. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |
| `container.app.image.repo` | `string` | Container image repository (e.g., `postgres`, `gcr.io/project/image`). | Required, non-empty |
| `container.app.image.tag` | `string` | Container image tag (e.g., `16`, `latest`). | Required, non-empty |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying resources. |
| `container.app.resources.limits.cpu` | `string` | `1000m` | Maximum CPU allocation. |
| `container.app.resources.limits.memory` | `string` | `1Gi` | Maximum memory allocation. |
| `container.app.resources.requests.cpu` | `string` | `50m` | Minimum guaranteed CPU. |
| `container.app.resources.requests.memory` | `string` | `100Mi` | Minimum guaranteed memory. |
| `container.app.env.variables` | `ContainerEnvVariable[]` | `[]` | Environment variables as a list. Each entry has a `name` and either a direct string `value` or a `valueFrom` reference to another resource. |
| `container.app.env.secrets` | `ContainerEnvSecret[]` | `[]` | Secret environment variables as a list. Each entry has a `name` and either a direct string `value` (auto-stored in a Kubernetes Secret) or a `secretRef` referencing an existing Kubernetes Secret. |
| `container.app.ports` | `object[]` | `[]` | Container port definitions. Each port requires `name` (lowercase alphanumeric and hyphens), `containerPort`, `networkProtocol` (`TCP`, `UDP`, or `SCTP`), `appProtocol`, and `servicePort`. |
| `container.app.ports[].isIngressPort` | `bool` | `false` | When `true`, ingress traffic routes to this port's `servicePort`. |
| `container.app.livenessProbe` | `Probe` | — | Periodic probe of container liveness. Restarts the container on failure. Supports `httpGet`, `tcpSocket`, `grpc`, and `exec` handlers. |
| `container.app.readinessProbe` | `Probe` | — | Periodic probe of service readiness. Removes the pod from Service endpoints on failure. |
| `container.app.startupProbe` | `Probe` | — | Startup probe. All other probes are disabled until this succeeds. Useful for slow-starting containers. |
| `container.app.volumeMounts` | `VolumeMount[]` | `[]` | Volume mounts supporting ConfigMap, Secret, HostPath, EmptyDir, and PVC sources. PVC mounts referencing a `volumeClaimTemplates` entry are handled automatically by the StatefulSet controller. |
| `container.app.command` | `string[]` | `[]` | Overrides the container image's ENTRYPOINT. |
| `container.app.args` | `string[]` | `[]` | Overrides the container image's CMD. |
| `container.app.image.pullSecretName` | `string` | — | Name of an existing image pull secret in the namespace. |
| `container.sidecars` | `Container[]` | `[]` | Sidecar containers deployed alongside the main application container. |
| `ingress.enabled` | `bool` | `false` | Enables Gateway API ingress with automatic TLS via cert-manager. Creates both external and internal gateways with HTTP-to-HTTPS redirect. |
| `ingress.hostname` | `string` | — | Full hostname for external access (e.g., `myapp.example.com`). Required when `ingress.enabled` is `true`. An internal hostname `internal-{hostname}` is created automatically. |
| `availability.replicas` | `int32` | `1` | Number of pod replicas to maintain. |
| `availability.podDisruptionBudget.enabled` | `bool` | `false` | Creates a PodDisruptionBudget to protect pods during voluntary disruptions (node maintenance, cluster upgrades). |
| `availability.podDisruptionBudget.minAvailable` | `string` | `1` | Minimum available pods during disruptions. Can be an absolute number or percentage. Cannot be used with `maxUnavailable`. |
| `availability.podDisruptionBudget.maxUnavailable` | `string` | — | Maximum unavailable pods during disruptions. Can be an absolute number or percentage. Cannot be used with `minAvailable`. |
| `podManagementPolicy` | `string` | `OrderedReady` | Pod management policy. `OrderedReady` creates pods sequentially, waiting for each to be ready. `Parallel` creates and deletes all pods simultaneously. |
| `volumeClaimTemplates` | `object[]` | `[]` | Persistent volume claim templates. Each pod gets its own PVC based on these templates. |
| `volumeClaimTemplates[].name` | `string` | — | Name of the volume claim template. Referenced in `container.app.volumeMounts` PVC entries. Required. |
| `volumeClaimTemplates[].size` | `string` | — | Requested storage size (e.g., `10Gi`, `100Gi`). Must be a valid Kubernetes quantity. Required. |
| `volumeClaimTemplates[].storageClass` | `string` | — | Storage class for the PVC. Uses the cluster default if not specified. |
| `volumeClaimTemplates[].accessModes` | `string[]` | `["ReadWriteOnce"]` | Access modes for the PVC. Valid values: `ReadWriteOnce`, `ReadOnlyMany`, `ReadWriteMany`, `ReadWriteOncePod`. |
| `configMaps` | `map<string, string>` | `{}` | ConfigMaps to create alongside the StatefulSet. Key is the ConfigMap name, value is the content. These ConfigMaps can be referenced in volume mounts. |

## Examples

### Basic Database with Persistent Storage

A single-replica PostgreSQL instance with a 20Gi persistent volume and health check:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesStatefulSet
metadata:
  name: postgres
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesStatefulSet.postgres
spec:
  namespace: databases
  createNamespace: true
  container:
    app:
      image:
        repo: postgres
        tag: "16"
      ports:
        - name: postgres
          containerPort: 5432
          networkProtocol: TCP
          appProtocol: tcp
          servicePort: 5432
      env:
        variables:
          - name: POSTGRES_DB
            value: "appdb"
        secrets:
          - name: POSTGRES_PASSWORD
            value: "change-me-in-production"
      readinessProbe:
        exec:
          command:
            - pg_isready
            - -U
            - postgres
        initialDelaySeconds: 5
        periodSeconds: 10
      volumeMounts:
        - name: data
          mountPath: /var/lib/postgresql/data
          pvc:
            claimName: data
  volumeClaimTemplates:
    - name: data
      size: 20Gi
```

### Distributed Cache with Environment References and ConfigMap

A three-replica Redis cluster referencing configuration from a ConfigMap and receiving connection details from another OpenMCF resource:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesStatefulSet
metadata:
  name: redis-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.KubernetesStatefulSet.redis-cluster
spec:
  namespace: caching
  container:
    app:
      image:
        repo: redis
        tag: "7.2"
      resources:
        limits:
          cpu: "2000m"
          memory: "4Gi"
        requests:
          cpu: "500m"
          memory: "1Gi"
      ports:
        - name: redis
          containerPort: 6379
          networkProtocol: TCP
          appProtocol: tcp
          servicePort: 6379
        - name: gossip
          containerPort: 16379
          networkProtocol: TCP
          appProtocol: tcp
          servicePort: 16379
      command:
        - redis-server
      args:
        - /etc/redis/redis.conf
        - --cluster-enabled
        - "yes"
      env:
        variables:
          - name: SENTINEL_HOST
            valueFrom:
              kind: KubernetesRedis
              name: sentinel
              field: status.outputs.service
      readinessProbe:
        tcpSocket:
          portNumber: 6379
        initialDelaySeconds: 5
        periodSeconds: 10
      livenessProbe:
        tcpSocket:
          portNumber: 6379
        initialDelaySeconds: 30
        periodSeconds: 15
      volumeMounts:
        - name: data
          mountPath: /data
          pvc:
            claimName: data
        - name: redis-config
          mountPath: /etc/redis/redis.conf
          subPath: redis.conf
          configMap:
            name: redis-config
            key: redis-config
  availability:
    replicas: 3
  podManagementPolicy: Parallel
  volumeClaimTemplates:
    - name: data
      size: 50Gi
      storageClass: ssd
      accessModes:
        - ReadWriteOnce
  configMaps:
    redis-config: |
      maxmemory 3gb
      maxmemory-policy allkeys-lru
      appendonly yes
      appendfsync everysec
```

### Production StatefulSet with Ingress and High Availability

A production-grade application with Gateway API ingress, multiple replicas, a pod disruption budget, startup probe, secret references, and multiple volume claim templates:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesStatefulSet
metadata:
  name: event-store
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesStatefulSet.event-store
spec:
  namespace: production
  container:
    app:
      image:
        repo: gcr.io/my-project/event-store
        tag: "v4.0.0"
      resources:
        limits:
          cpu: "4000m"
          memory: "8Gi"
        requests:
          cpu: "2000m"
          memory: "4Gi"
      ports:
        - name: http
          containerPort: 8080
          networkProtocol: TCP
          appProtocol: http
          servicePort: 80
          isIngressPort: true
        - name: grpc
          containerPort: 9090
          networkProtocol: TCP
          appProtocol: grpc
          servicePort: 9090
      env:
        variables:
          - name: CLUSTER_SIZE
            value: "5"
          - name: DATABASE_HOST
            valueFrom:
              kind: KubernetesPostgres
              name: prod-postgres
              field: status.outputs.service
        secrets:
          - name: DATABASE_PASSWORD
            secretRef:
              name: prod-db-credentials
              key: password
          - name: API_KEY
            secretRef:
              name: prod-api-keys
              key: event-store
      readinessProbe:
        grpc:
          port: 9090
        initialDelaySeconds: 10
        periodSeconds: 5
      livenessProbe:
        httpGet:
          path: /healthz
          portNumber: 8080
        initialDelaySeconds: 30
        periodSeconds: 15
      startupProbe:
        httpGet:
          path: /healthz
          portNumber: 8080
        failureThreshold: 30
        periodSeconds: 10
      volumeMounts:
        - name: data
          mountPath: /var/lib/event-store/data
          pvc:
            claimName: data
        - name: wal
          mountPath: /var/lib/event-store/wal
          pvc:
            claimName: wal
  ingress:
    enabled: true
    hostname: events.example.com
  availability:
    replicas: 5
    podDisruptionBudget:
      enabled: true
      minAvailable: "3"
  podManagementPolicy: OrderedReady
  volumeClaimTemplates:
    - name: data
      size: 200Gi
      storageClass: ssd
      accessModes:
        - ReadWriteOnce
    - name: wal
      size: 50Gi
      storageClass: ssd
      accessModes:
        - ReadWriteOnce
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where the StatefulSet is deployed |
| `headless_service` | `string` | Headless service name for stable pod network identity (e.g., `my-db-headless`). Pod DNS: `{pod-name}.{headless-service}.{namespace}.svc.cluster.local` |
| `service` | `string` | ClusterIP service name for load-balanced client access (matches `metadata.name`) |
| `port_forward_command` | `string` | kubectl port-forward command for local access |
| `kube_endpoint` | `string` | Cluster-internal FQDN (e.g., `my-db.my-namespace.svc.cluster.local`) |
| `external_hostname` | `string` | Public hostname for external access, only set when ingress is enabled |
| `internal_hostname` | `string` | VPC-internal hostname (e.g., `internal-myapp.example.com`), only set when ingress is enabled |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/namespace) — provides the target namespace via `valueFrom` reference
- [KubernetesDeployment](/docs/catalog/kubernetes/deployment) — alternative for stateless workloads that do not require stable identity or persistent storage per pod
- [KubernetesPostgres](/docs/catalog/kubernetes/postgres) — commonly referenced for database connection environment variables
- [KubernetesRedis](/docs/catalog/kubernetes/redis) — commonly referenced for cache connection environment variables
- [KubernetesCertManager](/docs/catalog/kubernetes/kubernetescertmanager) — provides the ClusterIssuer for ingress TLS certificates
- [KubernetesIstio](/docs/catalog/kubernetes/istio) — provides the Gateway API controller for ingress routing

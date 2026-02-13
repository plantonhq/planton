# Kubernetes DaemonSet

Deploys a containerized application to Kubernetes as a DaemonSet with support for node selection, tolerations, rolling update strategy, ConfigMap and Secret management, RBAC policy creation, volume mounts (including HostPath for node-level access), and container security contexts.

## What Gets Created

When you deploy a KubernetesDaemonSet resource, OpenMCF provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **ServiceAccount** — a dedicated service account for the DaemonSet pods, created only when `createServiceAccount` is `true`
- **DaemonSet** — a Kubernetes DaemonSet with the specified container image, resource limits, probes, volume mounts, node selector, tolerations, and update strategy
- **ConfigMaps** — one ConfigMap per entry in `configMaps`, created in the target namespace
- **Secret** — an Opaque Secret containing environment secrets provided as direct string values, created only when `container.app.env.secrets` includes direct values
- **Image Pull Secret** — a `kubernetes.io/dockerconfigjson` Secret for pulling from private registries, created only when a Docker config JSON is provided
- **ClusterRole and ClusterRoleBinding** — cluster-wide RBAC permissions for the ServiceAccount, created only when `rbac.clusterRules` is specified and `createServiceAccount` is `true`
- **Role and RoleBinding** — namespace-scoped RBAC permissions for the ServiceAccount, created only when `rbac.namespaceRules` is specified and `createServiceAccount` is `true`

## Prerequisites

- **Kubernetes credentials** configured via environment variables or OpenMCF provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **A container image** accessible from the cluster (public registry or with a configured image pull secret)

## Quick Start

Create a file `daemonset.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesDaemonSet
metadata:
  name: my-agent
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesDaemonSet.my-agent
spec:
  namespace: monitoring
  createNamespace: true
  container:
    app:
      image:
        repo: fluent/fluent-bit
        tag: "3.0"
```

Deploy:

```shell
openmcf apply -f daemonset.yaml
```

This creates a Fluent Bit DaemonSet running on every node in the `monitoring` namespace, using default resource limits (1000m CPU, 1Gi memory).

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `StringValueOrRef` | Kubernetes namespace for the DaemonSet. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |
| `container.app.image.repo` | `string` | Container image repository (e.g., `fluent/fluent-bit`, `gcr.io/project/image`). | Required, non-empty |
| `container.app.image.tag` | `string` | Container image tag (e.g., `latest`, `3.0`). | Required, non-empty |

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
| `container.app.env.variables` | `map<string, StringValueOrRef>` | `{}` | Environment variables. Each value can be a direct string via `value` or a reference to another resource via `valueFrom`. |
| `container.app.env.secrets` | `map<string, KubernetesSensitiveValue>` | `{}` | Secret environment variables. Each value can be a direct string via `value` (auto-stored in a Kubernetes Secret) or a reference to an existing Kubernetes Secret via `secretRef`. |
| `container.app.ports` | `object[]` | `[]` | Container port definitions. Each port requires `name` (lowercase alphanumeric and hyphens), `containerPort`, and `networkProtocol` (`TCP`, `UDP`, or `SCTP`). Optional `hostPort` exposes the port on the host node. |
| `container.app.volumeMounts` | `VolumeMount[]` | `[]` | Volume mounts supporting ConfigMap, Secret, HostPath, EmptyDir, and PVC sources. |
| `container.app.livenessProbe` | `Probe` | — | Periodic probe of container liveness. Restarts the container on failure. Supports `httpGet`, `tcpSocket`, `grpc`, and `exec` handlers. |
| `container.app.readinessProbe` | `Probe` | — | Periodic probe of service readiness. Removes the pod from Service endpoints on failure. |
| `container.app.startupProbe` | `Probe` | — | Startup probe. All other probes are disabled until this succeeds. Useful for slow-starting containers. |
| `container.app.command` | `string[]` | `[]` | Overrides the container image's ENTRYPOINT. |
| `container.app.args` | `string[]` | `[]` | Overrides the container image's CMD. |
| `container.app.securityContext.privileged` | `bool` | `false` | Run the container in privileged mode. |
| `container.app.securityContext.runAsUser` | `int64` | — | Run the container as a specific user ID. |
| `container.app.securityContext.runAsGroup` | `int64` | — | Run the container as a specific group ID. |
| `container.app.securityContext.runAsNonRoot` | `bool` | `false` | Require the container to run as a non-root user. |
| `container.app.securityContext.readOnlyRootFilesystem` | `bool` | `false` | Mount the root filesystem as read-only. |
| `container.app.securityContext.capabilities.add` | `string[]` | `[]` | Linux capabilities to add (e.g., `SYS_PTRACE`, `NET_ADMIN`). |
| `container.app.securityContext.capabilities.drop` | `string[]` | `[]` | Linux capabilities to drop (e.g., `ALL`). |
| `container.app.image.pullSecretName` | `string` | — | Name of an existing image pull secret in the namespace. |
| `container.sidecars` | `Container[]` | `[]` | Sidecar containers deployed alongside the main application container. |
| `nodeSelector` | `map<string, string>` | `{}` | Key-value pairs that must match labels on nodes for the DaemonSet pods to be scheduled. |
| `tolerations` | `object[]` | `[]` | Tolerations allowing pods to be scheduled on nodes with matching taints. Each toleration accepts `key`, `operator` (`Equal` or `Exists`), `value`, `effect` (`NoSchedule`, `PreferNoSchedule`, or `NoExecute`), and `tolerationSeconds`. |
| `updateStrategy.type` | `string` | — | Update strategy type. `RollingUpdate` progressively replaces pods; `OnDelete` waits for manual pod deletion. |
| `updateStrategy.rollingUpdate.maxUnavailable` | `string` | `1` | Maximum pods unavailable during a rolling update. Absolute number or percentage. |
| `updateStrategy.rollingUpdate.maxSurge` | `string` | `0` | Maximum extra pods created during a rolling update. Absolute number or percentage. |
| `minReadySeconds` | `int32` | `0` | Seconds a new pod must be ready without crashing before it is considered available. |
| `createServiceAccount` | `bool` | `false` | When `true`, creates a ServiceAccount for the DaemonSet pods. |
| `serviceAccountName` | `string` | `metadata.name` | Name of the ServiceAccount. If `createServiceAccount` is `false`, references an existing ServiceAccount. |
| `configMaps` | `map<string, string>` | `{}` | ConfigMaps to create alongside the DaemonSet. Key is the ConfigMap name, value is the content. |
| `rbac.clusterRules` | `RbacRule[]` | `[]` | Cluster-wide RBAC policy rules. Creates a ClusterRole and ClusterRoleBinding. Only used when `createServiceAccount` is `true`. Each rule requires `apiGroups`, `resources`, and `verbs`. Optional `resourceNames` restricts to specific resources. |
| `rbac.namespaceRules` | `RbacRule[]` | `[]` | Namespace-scoped RBAC policy rules. Creates a Role and RoleBinding. Only used when `createServiceAccount` is `true`. Same structure as `clusterRules`. |

## Examples

### Node Log Collector

A Fluent Bit DaemonSet that mounts host log directories to collect and forward container logs from every node:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesDaemonSet
metadata:
  name: log-collector
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesDaemonSet.log-collector
spec:
  namespace: logging
  createNamespace: true
  container:
    app:
      image:
        repo: fluent/fluent-bit
        tag: "3.0"
      resources:
        limits:
          cpu: "500m"
          memory: "256Mi"
        requests:
          cpu: "100m"
          memory: "128Mi"
      ports:
        - name: metrics
          containerPort: 2020
          networkProtocol: TCP
      volumeMounts:
        - name: varlog
          mountPath: /var/log
          readOnly: true
          hostPath:
            path: /var/log
        - name: containers
          mountPath: /var/lib/docker/containers
          readOnly: true
          hostPath:
            path: /var/lib/docker/containers
      readinessProbe:
        httpGet:
          path: /api/v1/health
          portNumber: 2020
        initialDelaySeconds: 5
        periodSeconds: 10
```

### Node Monitor with Tolerations and RBAC

A Prometheus Node Exporter DaemonSet that runs on all nodes (including control-plane nodes via tolerations), with a dedicated ServiceAccount and cluster-wide read permissions:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesDaemonSet
metadata:
  name: node-exporter
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.KubernetesDaemonSet.node-exporter
spec:
  namespace: monitoring
  createServiceAccount: true
  container:
    app:
      image:
        repo: prom/node-exporter
        tag: "v1.8.0"
      resources:
        limits:
          cpu: "250m"
          memory: "180Mi"
        requests:
          cpu: "100m"
          memory: "128Mi"
      args:
        - "--path.procfs=/host/proc"
        - "--path.sysfs=/host/sys"
        - "--path.rootfs=/host/root"
        - "--collector.filesystem.mount-points-exclude=^/(dev|proc|sys|var/lib/docker/.+)($|/)"
      ports:
        - name: metrics
          containerPort: 9100
          networkProtocol: TCP
          hostPort: 9100
      volumeMounts:
        - name: proc
          mountPath: /host/proc
          readOnly: true
          hostPath:
            path: /proc
        - name: sys
          mountPath: /host/sys
          readOnly: true
          hostPath:
            path: /sys
        - name: root
          mountPath: /host/root
          readOnly: true
          hostPath:
            path: /
            type: Directory
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
        readOnlyRootFilesystem: true
        capabilities:
          drop:
            - ALL
      livenessProbe:
        httpGet:
          path: /
          portNumber: 9100
        initialDelaySeconds: 10
        periodSeconds: 15
  nodeSelector:
    kubernetes.io/os: linux
  tolerations:
    - operator: Exists
  rbac:
    clusterRules:
      - apiGroups:
          - ""
        resources:
          - nodes
          - nodes/metrics
        verbs:
          - get
          - list
          - watch
```

### Production Log Pipeline with ConfigMap, Secrets, and Rolling Updates

A Vector log agent with a custom configuration file loaded from a ConfigMap, environment-referenced pipeline endpoints, secret credentials from an existing Kubernetes Secret, and a controlled rolling update strategy:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesDaemonSet
metadata:
  name: vector-agent
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesDaemonSet.vector-agent
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: logging-ns
      fieldPath: spec.name
  createServiceAccount: true
  serviceAccountName: vector-agent
  container:
    app:
      image:
        repo: timberio/vector
        tag: "0.38.0-alpine"
      resources:
        limits:
          cpu: "1000m"
          memory: "512Mi"
        requests:
          cpu: "200m"
          memory: "256Mi"
      command:
        - vector
      args:
        - "--config-dir"
        - "/etc/vector"
      env:
        variables:
          VECTOR_SELF_NODE_NAME:
            value: "${K8S_NODE_NAME}"
          SINK_ENDPOINT:
            valueFrom:
              kind: KubernetesDeployment
              name: log-aggregator
              fieldPath: status.outputs.kubeEndpoint
        secrets:
          SINK_API_KEY:
            secretRef:
              name: vector-credentials
              key: api-key
      ports:
        - name: metrics
          containerPort: 9598
          networkProtocol: TCP
      volumeMounts:
        - name: config
          mountPath: /etc/vector
          configMap:
            name: vector-config
        - name: varlog
          mountPath: /var/log
          readOnly: true
          hostPath:
            path: /var/log
        - name: data
          mountPath: /var/lib/vector
          emptyDir:
            sizeLimit: "1Gi"
      securityContext:
        privileged: false
        readOnlyRootFilesystem: true
        capabilities:
          add:
            - DAC_READ_SEARCH
          drop:
            - ALL
      readinessProbe:
        httpGet:
          path: /health
          portNumber: 9598
        initialDelaySeconds: 5
        periodSeconds: 10
      livenessProbe:
        httpGet:
          path: /health
          portNumber: 9598
        initialDelaySeconds: 15
        periodSeconds: 20
  nodeSelector:
    kubernetes.io/os: linux
  tolerations:
    - key: node-role.kubernetes.io/control-plane
      operator: Exists
      effect: NoSchedule
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: "1"
      maxSurge: "0"
  minReadySeconds: 10
  configMaps:
    vector-config: |
      data_dir: /var/lib/vector
      sources:
        kubernetes_logs:
          type: kubernetes_logs
      sinks:
        aggregator:
          type: vector
          inputs:
            - kubernetes_logs
          address: "${SINK_ENDPOINT}:6000"
  rbac:
    clusterRules:
      - apiGroups:
          - ""
        resources:
          - pods
          - namespaces
          - nodes
        verbs:
          - get
          - list
          - watch
    namespaceRules:
      - apiGroups:
          - ""
        resources:
          - configmaps
        verbs:
          - get
          - list
          - watch

```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where the DaemonSet is created |
| `daemonsetName` | `string` | Name of the created DaemonSet |
| `desiredNumberScheduled` | `string` | Number of nodes that should be running the daemon pod |
| `currentNumberScheduled` | `string` | Number of nodes running at least one daemon pod |
| `numberReady` | `string` | Number of nodes running the daemon pod with at least one container ready |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — provides the target namespace via `valueFrom` reference
- [KubernetesDeployment](/docs/catalog/kubernetes/kubernetesdeployment) — commonly referenced for service endpoint environment variables
- [KubernetesSecret](/docs/catalog/kubernetes/kubernetessecret) — provides pre-existing secrets that can be referenced via `secretRef`
- [KubernetesService](/docs/catalog/kubernetes/kubernetesservice) — can expose DaemonSet pod ports as a Service for cluster-internal access

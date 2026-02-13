# Kubernetes Deployment

Deploys a containerized application to Kubernetes as a Deployment with automatic Service creation, configurable ingress via Gateway API, environment variable and secret management, and availability controls including autoscaling, rolling update strategy, and pod disruption budgets.

## What Gets Created

When you deploy a KubernetesDeployment resource, OpenMCF provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **ServiceAccount** — a dedicated service account for the deployment's pods
- **Deployment** — a Kubernetes Deployment with the specified container image, resource limits, probes, volume mounts, and rolling update strategy
- **Service** — a ClusterIP Service mapping service ports to container ports, created only when at least one port is defined in `container.app.ports`
- **ConfigMaps** — one ConfigMap per entry in `configMaps`, each prefixed with `metadata.name` to avoid namespace conflicts
- **Secret** — an Opaque Secret containing environment secrets provided as direct string values, created only when `container.app.env.secrets` includes direct values
- **Image Pull Secret** — a `kubernetes.io/dockerconfigjson` Secret for pulling from private registries, created only when a Docker config JSON is provided
- **Certificate** — a cert-manager Certificate for TLS termination on ingress hostnames, created only when ingress is enabled
- **External Gateway** — a Gateway API Gateway for traffic from outside the VPC, with HTTPS (port 443) and HTTP (port 80) listeners, created only when ingress is enabled
- **Internal Gateway** — a Gateway API Gateway for traffic from within the VPC, with HTTPS and HTTP listeners on an `internal-` prefixed hostname, created only when ingress is enabled
- **HTTPRoutes** — four HTTPRoute resources handling external HTTP-to-HTTPS redirect, external HTTPS routing, internal HTTP-to-HTTPS redirect, and internal HTTPS routing to the Service, created only when ingress is enabled
- **PodDisruptionBudget** — ensures minimum pod availability during voluntary disruptions, created only when `availability.podDisruptionBudget.enabled` is `true`

## Prerequisites

- **Kubernetes credentials** configured via environment variables or OpenMCF provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **A container image** accessible from the cluster (public registry or with a configured image pull secret)
- **cert-manager with a ClusterIssuer** if enabling ingress — the ClusterIssuer name must match the domain extracted from the ingress hostname
- **Gateway API CRDs and an Istio-based gateway controller** if enabling ingress

## Quick Start

Create a file `deployment.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesDeployment
metadata:
  name: my-app
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesDeployment.my-app
spec:
  namespace: my-namespace
  createNamespace: true
  container:
    app:
      image:
        repo: nginx
        tag: "1.25"
      ports:
        - name: http
          containerPort: 80
          networkProtocol: TCP
          appProtocol: http
          servicePort: 80
```

Deploy:

```shell
openmcf apply -f deployment.yaml
```

This creates a single-replica nginx Deployment with a ClusterIP Service on port 80 in the `my-namespace` namespace, using default resource limits (1000m CPU, 1Gi memory).

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `string` | Kubernetes namespace for the deployment. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |
| `container.app.image.repo` | `string` | Container image repository (e.g., `nginx`, `gcr.io/project/image`). | Required, non-empty |
| `container.app.image.tag` | `string` | Container image tag (e.g., `latest`, `1.25`). | Required, non-empty |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying resources. |
| `version` | `string` | — | Deployment version identifier (e.g., `main`, `review-42`). 1–30 characters, lowercase alphanumeric and hyphens only, must not end with a hyphen. |
| `container.app.resources.limits.cpu` | `string` | `1000m` | Maximum CPU allocation. |
| `container.app.resources.limits.memory` | `string` | `1Gi` | Maximum memory allocation. |
| `container.app.resources.requests.cpu` | `string` | `50m` | Minimum guaranteed CPU. |
| `container.app.resources.requests.memory` | `string` | `100Mi` | Minimum guaranteed memory. |
| `container.app.env.variables` | `map<string, StringValueOrRef>` | `{}` | Environment variables. Each value can be a direct string via `value` or a reference to another resource via `valueFrom`. |
| `container.app.env.secrets` | `map<string, KubernetesSensitiveValue>` | `{}` | Secret environment variables. Each value can be a direct string via `value` (auto-stored in a Kubernetes Secret) or a reference to an existing Kubernetes Secret via `secretRef`. |
| `container.app.ports` | `object[]` | `[]` | Container port definitions. Each port requires `name` (lowercase alphanumeric and hyphens), `containerPort`, `networkProtocol` (`TCP`, `UDP`, or `SCTP`), `appProtocol`, and `servicePort`. |
| `container.app.ports[].isIngressPort` | `bool` | `false` | When `true`, ingress traffic routes to this port's `servicePort`. |
| `container.app.livenessProbe` | `Probe` | — | Periodic probe of container liveness. Restarts the container on failure. Supports `httpGet`, `tcpSocket`, `grpc`, and `exec` handlers. |
| `container.app.readinessProbe` | `Probe` | — | Periodic probe of service readiness. Removes the pod from Service endpoints on failure. |
| `container.app.startupProbe` | `Probe` | — | Startup probe. All other probes are disabled until this succeeds. Useful for slow-starting containers. |
| `container.app.volumeMounts` | `VolumeMount[]` | `[]` | Volume mounts supporting ConfigMap, Secret, HostPath, EmptyDir, and PVC sources. |
| `container.app.command` | `string[]` | `[]` | Overrides the container image's ENTRYPOINT. |
| `container.app.args` | `string[]` | `[]` | Overrides the container image's CMD. |
| `container.app.image.pullSecretName` | `string` | — | Name of an existing image pull secret in the namespace. |
| `container.sidecars` | `Container[]` | `[]` | Sidecar containers deployed alongside the main application container. |
| `ingress.enabled` | `bool` | `false` | Enables Gateway API ingress with automatic TLS via cert-manager. Creates both external and internal gateways with HTTP-to-HTTPS redirect. |
| `ingress.hostname` | `string` | — | Full hostname for external access (e.g., `myapp.example.com`). Required when `ingress.enabled` is `true`. An internal hostname `internal-{hostname}` is created automatically. |
| `availability.minReplicas` | `int32` | `1` | Minimum number of pod replicas. |
| `availability.horizontalPodAutoscaling.isEnabled` | `bool` | `false` | Enables horizontal pod autoscaling. |
| `availability.horizontalPodAutoscaling.targetCpuUtilizationPercent` | `double` | — | CPU utilization percentage that triggers autoscaling (e.g., `70`). |
| `availability.horizontalPodAutoscaling.targetMemoryUtilization` | `string` | — | Memory utilization that triggers autoscaling (e.g., `1Gi`). |
| `availability.deploymentStrategy.maxUnavailable` | `string` | `25%` | Maximum pods unavailable during a rolling update. Set to `0` for zero-downtime deployments. Cannot be `0` if `maxSurge` is also `0`. |
| `availability.deploymentStrategy.maxSurge` | `string` | `25%` | Maximum pods created above desired count during a rolling update. Cannot be `0` if `maxUnavailable` is also `0`. |
| `availability.podDisruptionBudget.enabled` | `bool` | `false` | Creates a PodDisruptionBudget to protect pods during voluntary disruptions (node maintenance, cluster upgrades). |
| `availability.podDisruptionBudget.minAvailable` | `string` | `1` | Minimum available pods during disruptions. Can be an absolute number or percentage. Cannot be used with `maxUnavailable`. |
| `availability.podDisruptionBudget.maxUnavailable` | `string` | — | Maximum unavailable pods during disruptions. Can be an absolute number or percentage. Cannot be used with `minAvailable`. |
| `configMaps` | `map<string, string>` | `{}` | ConfigMaps to create alongside the deployment. Key is the logical name, value is the content. Each ConfigMap is created with the name `{metadata.name}-{key}`. |

## Examples

### Basic Web Server with Health Checks

A single-replica web server with HTTP readiness and liveness probes:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesDeployment
metadata:
  name: web-server
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesDeployment.web-server
spec:
  namespace: web
  createNamespace: true
  container:
    app:
      image:
        repo: nginx
        tag: "1.25"
      ports:
        - name: http
          containerPort: 80
          networkProtocol: TCP
          appProtocol: http
          servicePort: 80
      readinessProbe:
        httpGet:
          path: /
          portNumber: 80
        initialDelaySeconds: 5
        periodSeconds: 10
      livenessProbe:
        httpGet:
          path: /
          portNumber: 80
        initialDelaySeconds: 15
        periodSeconds: 20
```

### Microservice with Environment Variables and Secrets

A backend API service referencing other OpenMCF-managed resources for database and cache connections, with secrets stored in an external Kubernetes Secret:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesDeployment
metadata:
  name: api-server
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.KubernetesDeployment.api-server
spec:
  namespace: backend
  version: main
  container:
    app:
      image:
        repo: gcr.io/my-project/api-server
        tag: "v2.1.0"
      resources:
        limits:
          cpu: "2000m"
          memory: "2Gi"
        requests:
          cpu: "500m"
          memory: "512Mi"
      ports:
        - name: http
          containerPort: 8080
          networkProtocol: TCP
          appProtocol: http
          servicePort: 80
        - name: grpc
          containerPort: 9090
          networkProtocol: TCP
          appProtocol: grpc
          servicePort: 9090
      env:
        variables:
          DATABASE_HOST:
            valueFrom:
              kind: KubernetesPostgres
              name: my-postgres
              field: status.outputs.service
          REDIS_HOST:
            valueFrom:
              kind: KubernetesRedis
              name: my-redis
              field: status.outputs.service
          LOG_LEVEL:
            value: "info"
        secrets:
          DATABASE_PASSWORD:
            secretRef:
              name: postgres-credentials
              key: password
      readinessProbe:
        grpc:
          port: 9090
        initialDelaySeconds: 5
        periodSeconds: 10
```

### Production Deployment with Ingress and High Availability

Full-featured production deployment with Gateway API ingress, autoscaling, zero-downtime rolling updates, and a pod disruption budget:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesDeployment
metadata:
  name: prod-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesDeployment.prod-api
spec:
  namespace: production
  version: main
  container:
    app:
      image:
        repo: gcr.io/my-project/api
        tag: "v3.0.0"
      resources:
        limits:
          cpu: "4000m"
          memory: "4Gi"
        requests:
          cpu: "1000m"
          memory: "1Gi"
      ports:
        - name: http
          containerPort: 8080
          networkProtocol: TCP
          appProtocol: http
          servicePort: 80
          isIngressPort: true
      readinessProbe:
        httpGet:
          path: /healthz
          portNumber: 8080
        initialDelaySeconds: 5
        periodSeconds: 10
      livenessProbe:
        httpGet:
          path: /healthz
          portNumber: 8080
        initialDelaySeconds: 15
        periodSeconds: 20
      startupProbe:
        httpGet:
          path: /healthz
          portNumber: 8080
        failureThreshold: 30
        periodSeconds: 10
  ingress:
    enabled: true
    hostname: api.example.com
  availability:
    minReplicas: 3
    horizontalPodAutoscaling:
      isEnabled: true
      targetCpuUtilizationPercent: 70
    deploymentStrategy:
      maxUnavailable: "0"
      maxSurge: "1"
    podDisruptionBudget:
      enabled: true
      minAvailable: "2"
```

### Deployment with ConfigMaps and Volume Mounts

A worker process with a custom command that mounts a ConfigMap as a configuration file:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesDeployment
metadata:
  name: worker
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesDeployment.worker
spec:
  namespace: workers
  createNamespace: true
  container:
    app:
      image:
        repo: my-registry/worker
        tag: "latest"
      command:
        - /bin/sh
        - -c
      args:
        - "python /app/worker.py --config /etc/config/settings.yaml"
      volumeMounts:
        - name: config
          mountPath: /etc/config/settings.yaml
          subPath: settings.yaml
          configMap:
            name: worker-settings
            key: settings
  configMaps:
    settings: |
      workers: 4
      timeout: 30
      queue: default
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where the deployment is created |
| `service` | `string` | Kubernetes Service name (matches `metadata.name`) |
| `port_forward_command` | `string` | kubectl port-forward command for local access |
| `kube_endpoint` | `string` | Cluster-internal FQDN (e.g., `my-app.my-namespace.svc.cluster.local`) |
| `external_hostname` | `string` | Public hostname for external access, only set when ingress is enabled |
| `internal_hostname` | `string` | VPC-internal hostname (e.g., `internal-myapp.example.com`), only set when ingress is enabled |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — provides the target namespace via `valueFrom` reference
- [KubernetesPostgres](/docs/catalog/kubernetes/kubernetespostgres) — commonly referenced for database connection environment variables
- [KubernetesRedis](/docs/catalog/kubernetes/kubernetesredis) — commonly referenced for cache connection environment variables
- [KubernetesCertManager](/docs/catalog/kubernetes/kubernetescertmanager) — provides the ClusterIssuer for ingress TLS certificates
- [KubernetesIstio](/docs/catalog/kubernetes/kubernetesistio) — provides the Gateway API controller for ingress routing

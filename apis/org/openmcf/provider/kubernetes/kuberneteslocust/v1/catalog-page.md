# Kubernetes Locust

Deploys a Locust distributed load testing cluster on Kubernetes using the Delivery Hero Locust Helm chart. Provisions master and worker nodes with configurable replicas and resource limits, injects test scripts and library files via ConfigMaps, supports extra pip packages, allows arbitrary Helm value overrides, and optionally exposes the Locust web UI externally through Istio Gateway API ingress with TLS termination and HTTP-to-HTTPS redirect.

## What Gets Created

When you deploy a KubernetesLocust resource, OpenMCF provisions:

- **Kubernetes Namespace** — created if `createNamespace` is `true`
- **Main Script ConfigMap** — a ConfigMap containing your `main.py` Locust test file
- **Library Files ConfigMap** — a ConfigMap containing additional Python files referenced by the main script
- **Locust Helm Release** — the `locust` chart (v0.31.5) from `https://charts.deliveryhero.io`, which creates:
  - A Locust master pod serving the web UI and coordinating workers
  - One or more Locust worker pods that generate simulated user traffic
  - Kubernetes Service for cluster-internal access to the master on port 8080
- **Ingress Resources** (when `ingress.enabled` is `true`):
  - cert-manager Certificate for TLS, issued by a ClusterIssuer matching the ingress domain
  - Gateway API Gateway with HTTPS (port 443) and HTTP (port 80) listeners
  - HTTPRoute for HTTPS traffic forwarding to the Locust master service
  - HTTPRoute for HTTP-to-HTTPS 301 redirect

## Prerequisites

- **A Kubernetes cluster** with kubectl configured for access
- **Istio ingress gateway** installed (only if using ingress)
- **cert-manager** with a ClusterIssuer matching your ingress domain (only if using ingress)
- **Gateway API CRDs** installed in the cluster (only if using ingress)

## Quick Start

Create a file `locust.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesLocust
metadata:
  name: my-locust
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesLocust.my-locust
spec:
  namespace:
    value: locust-dev
  createNamespace: true
  loadTest:
    name: smoke-test
    mainPyContent: |
      from locust import HttpUser, task, between

      class QuickstartUser(HttpUser):
          wait_time = between(1, 3)

          @task
          def index(self):
              self.client.get("/")
    libFilesContent: {}
```

Deploy:

```shell
openmcf apply -f locust.yaml
```

This creates a Locust cluster with one master and one worker using default resources (1 CPU / 1Gi memory limit, 50m CPU / 100Mi memory request each) in the `locust-dev` namespace.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `StringValueOrRef` | Kubernetes namespace for the Locust deployment. Use `value` for a direct string or `valueFrom` to reference a KubernetesNamespace resource. | Required |
| `loadTest.name` | `string` | Unique identifier for this load test configuration. | Required |
| `loadTest.mainPyContent` | `string` | Python source code for the main Locust test script. Defines simulated user behavior. | Required |
| `loadTest.libFilesContent` | `map<string, string>` | Map of filename to Python source code for additional library files used by the main script. Pass an empty map `{}` when no extra files are needed. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `createNamespace` | `bool` | `false` | Create the namespace if it does not exist. |
| `masterContainer.replicas` | `int32` | `1` | Number of Locust master replicas. |
| `masterContainer.resources.limits.cpu` | `string` | `"1000m"` | CPU limit for the master container. |
| `masterContainer.resources.limits.memory` | `string` | `"1Gi"` | Memory limit for the master container. |
| `masterContainer.resources.requests.cpu` | `string` | `"50m"` | CPU request for the master container. |
| `masterContainer.resources.requests.memory` | `string` | `"100Mi"` | Memory request for the master container. |
| `workerContainer.replicas` | `int32` | `1` | Number of Locust worker replicas. Increase for higher load generation concurrency. |
| `workerContainer.resources.limits.cpu` | `string` | `"1000m"` | CPU limit for each worker container. |
| `workerContainer.resources.limits.memory` | `string` | `"1Gi"` | Memory limit for each worker container. |
| `workerContainer.resources.requests.cpu` | `string` | `"50m"` | CPU request for each worker container. |
| `workerContainer.resources.requests.memory` | `string` | `"100Mi"` | Memory request for each worker container. |
| `loadTest.pipPackages` | `repeated string` | `[]` | Extra Python pip packages to install in the Locust environment for custom dependencies. |
| `helmValues` | `map<string, string>` | `{}` | Additional Helm chart values for customization. See the [Locust Helm chart values](https://github.com/deliveryhero/helm-charts/tree/master/stable/locust#values) for available options. |
| `ingress.enabled` | `bool` | `false` | Enable external access to the Locust web UI via Istio Gateway API ingress with TLS. |
| `ingress.hostname` | `string` | -- | Full hostname for external access (e.g., `locust.example.com`). Required when `ingress.enabled` is `true`. |

## Examples

### Locust with Scaled Workers

Scale workers up and allocate more resources for higher throughput testing:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesLocust
metadata:
  name: load-test
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesLocust.load-test
spec:
  namespace:
    value: load-testing
  createNamespace: true
  masterContainer:
    replicas: 1
    resources:
      limits:
        cpu: "2000m"
        memory: "2Gi"
      requests:
        cpu: "250m"
        memory: "256Mi"
  workerContainer:
    replicas: 5
    resources:
      limits:
        cpu: "2000m"
        memory: "2Gi"
      requests:
        cpu: "500m"
        memory: "512Mi"
  loadTest:
    name: api-stress
    mainPyContent: |
      from locust import HttpUser, task, between

      class ApiUser(HttpUser):
          wait_time = between(0.5, 2)

          @task(3)
          def list_items(self):
              self.client.get("/api/items")

          @task(1)
          def create_item(self):
              self.client.post("/api/items", json={"name": "test"})
    libFilesContent: {}
```

### Locust with Library Files and Pip Packages

Supply helper modules and install additional Python packages for complex test scenarios:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesLocust
metadata:
  name: advanced-test
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.KubernetesLocust.advanced-test
spec:
  namespace:
    value: perf-testing
  createNamespace: true
  workerContainer:
    replicas: 3
    resources:
      limits:
        cpu: "1000m"
        memory: "1Gi"
      requests:
        cpu: "200m"
        memory: "256Mi"
  loadTest:
    name: e2e-flow
    mainPyContent: |
      from locust import HttpUser, task, between
      from lib.auth import get_token

      class AuthenticatedUser(HttpUser):
          wait_time = between(1, 5)

          def on_start(self):
              self.token = get_token(self.client)

          @task
          def get_dashboard(self):
              self.client.get("/api/dashboard",
                  headers={"Authorization": f"Bearer {self.token}"})
    libFilesContent:
      auth.py: |
        def get_token(client):
            resp = client.post("/auth/login",
                json={"user": "loadtest", "pass": "secret"})
            return resp.json()["token"]
    pipPackages:
      - faker
      - pytz
  helmValues:
    master.environment.LOCUST_HOST: "https://api.staging.example.com"
```

### Full-Featured with Ingress

Expose the Locust web UI over HTTPS with automatic TLS and HTTP redirect:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesLocust
metadata:
  name: prod-loadtest
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesLocust.prod-loadtest
spec:
  namespace:
    value: loadtest-prod
  createNamespace: true
  masterContainer:
    replicas: 1
    resources:
      limits:
        cpu: "4000m"
        memory: "4Gi"
      requests:
        cpu: "500m"
        memory: "1Gi"
  workerContainer:
    replicas: 10
    resources:
      limits:
        cpu: "2000m"
        memory: "2Gi"
      requests:
        cpu: "500m"
        memory: "512Mi"
  loadTest:
    name: production-soak
    mainPyContent: |
      from locust import HttpUser, task, between, events
      from lib.scenarios import browse_catalog, checkout_flow

      class ProductionUser(HttpUser):
          wait_time = between(2, 10)

          @task(5)
          def browse(self):
              browse_catalog(self.client)

          @task(1)
          def checkout(self):
              checkout_flow(self.client)
    libFilesContent:
      scenarios.py: |
        import random

        def browse_catalog(client):
            client.get("/products")
            product_id = random.randint(1, 100)
            client.get(f"/products/{product_id}")

        def checkout_flow(client):
            client.post("/cart", json={"productId": 1, "qty": 1})
            client.post("/checkout", json={"method": "credit"})
    pipPackages:
      - faker
  helmValues:
    master.environment.LOCUST_HOST: "https://api.example.com"
    master.environment.LOCUST_USERS: "500"
    master.environment.LOCUST_SPAWN_RATE: "10"
  ingress:
    enabled: true
    hostname: locust.example.com
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where the Locust cluster was created |
| `service` | `string` | Name of the Kubernetes service for the Locust master |
| `port_forward_command` | `string` | Ready-to-run `kubectl port-forward` command for local access on port 8080 |
| `kube_endpoint` | `string` | Cluster-internal endpoint (e.g., `my-locust.locust-dev.svc.cluster.local`) |
| `external_hostname` | `string` | External hostname when ingress is enabled (e.g., `locust.example.com`) |
| `internal_hostname` | `string` | Internal hostname for private ingress (e.g., `internal-locust.example.com`) |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — pre-create a namespace to reference via `valueFrom`
- [KubernetesDeployment](/docs/catalog/kubernetes/kubernetesdeployment) — deploy the target application to load test against
- [KubernetesRedis](/docs/catalog/kubernetes/kubernetesredis) — deploy Redis as a backend or caching layer for the system under test

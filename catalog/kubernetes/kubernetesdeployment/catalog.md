# Kubernetes Deployment

Deploys a containerized application on any Kubernetes cluster as an apps/v1 Deployment fronted by a Kubernetes Service. Supports custom container images, environment variables with cross-resource references, secret variables, sidecar and init containers, liveness/readiness/startup probes, volume mounts from ConfigMaps, Secrets, HostPath, EmptyDir, or PVCs, horizontal pod autoscaling, pod disruption budgets, zero-downtime rolling update strategies, pod scheduling (affinity, tolerations, topology spread), and the full restricted-profile security hardening surface. External exposure is composed with first-class kinds (KubernetesIngress, Gateway API routes) referencing the exported Service. Credentials are delivered through a Kubernetes Provider Connection or Runner-based delivery.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **Kubernetes Deployment** -- the core workload resource with the configured container image, resource requests/limits, environment variables, probes, volume mounts, command/args overrides, lifecycle hooks, security contexts, scheduling rules, sidecar containers, and init containers
- **Kubernetes Service** -- routes traffic to deployment pods on the configured service ports; created only when `container.app.ports` are defined
- **Kubernetes Secret** -- created only when `container.app.env.secrets` carry literal values; stores them and mounts them as environment variables
- **Horizontal Pod Autoscaler** -- created only when `availability.horizontalPodAutoscaling.enabled` is `true`; scales pods between the replica floor and `maxReplicas` based on CPU or memory utilization
- **PodDisruptionBudget** -- created only when `availability.podDisruptionBudget.enabled` is `true`; enforces minimum availability during voluntary disruptions
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **A container registry** accessible from the cluster for pulling the application image. If the registry is private and the cluster's own identity does not reach it, declare the login on `pod.imageRegistries` (the module materializes the pull Secret), or name a Secret declared beside the workload in `pod.imagePullSecrets` (or attach it to the referenced ServiceAccount).

## Deploy

### Console

Open the deployment store, find **Kubernetes Deployment**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Web Service Deployment** preset in the [Presets](#presets) tab to pre-populate a working configuration for HTTP services.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesDeployment
metadata:
  name: api-server
  org: acme-corp
  env: prod
spec:
  namespace:
    value: "apps"
  createNamespace: true
  version: "main"
  container:
    app:
      image:
        repo: ghcr.io/acme-corp/api-server
        tag: v1.0.0
      ports:
        - name: http
          containerPort: 8080
          networkProtocol: TCP
          appProtocol: http
          servicePort: 80
```

```shell
planton apply -f deployment.yaml
```

This creates a single-replica Deployment with a Kubernetes Service on port 80 and cluster-internal access. Autoscaling, probes, secrets, and hardening are not configured -- add them as your workload needs them. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the deployment to a namespace managed by another Infra Component:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: apps-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline resolves the dependency graph, deploys the namespace first, then provisions the Deployment with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Kubernetes Deployment. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Container image and deployment track** -- The `container.app.image` field specifies the repository and tag for the application container; deployment pipelines inject the freshly built tag through exactly these paths. The `version` field (e.g., `"main"` or `"review-42"`) is the deployment track, stamped as a pod label so multiple tracks of one app coexist in a namespace with disjoint traffic -- pipelines set it from the git branch.

**Environment variables and secrets** -- Use `container.app.env.variables` for plain configuration and `container.app.env.secrets` for sensitive values. Variables support every Kubernetes-native source (ConfigMap keys, pod fields, container resource quantities) plus ValueFromRef to resolve values from other Infra Components at deploy time. Secret literals are materialized into an auto-created Kubernetes Secret; existing Secrets are referenced by name and key.

**Availability and autoscaling** -- Set `availability.replicas` for the baseline pod count. Enable `availability.horizontalPodAutoscaling` with `maxReplicas` and a CPU or memory target to auto-scale under load. Configure `availability.strategy` with `maxUnavailable: "0"` and `maxSurge: "1"` for zero-downtime rolling updates. Add a `podDisruptionBudget` to protect availability during node drains and cluster upgrades.

**Health probes** -- Configure `livenessProbe`, `readinessProbe`, and `startupProbe` on the application container. Readiness probes are essential for zero-downtime deployments -- they prevent traffic from routing to pods that are not yet ready to serve requests.

**Pod identity and hardening** -- Reference a `KubernetesServiceAccount` resource in `pod.serviceAccount` so workload identity and pull secrets live on the composed identity, and disable `pod.automountServiceAccountToken` for apps that never call the Kubernetes API. The pod- and container-level security contexts express the full restricted Pod Security Standard: non-root, read-only root filesystem, dropped capabilities, and seccomp.

**External exposure** -- Composed, never embedded: the workload exports its Service name and selector labels, and first-class exposure kinds (KubernetesIngress, KubernetesHttpRoute and the other Gateway API routes, with certificates) reference them -- so every piece of exposure infrastructure is a visible node in the resource graph.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesServiceAccount** | `pod.serviceAccount` | `status.outputs.service_account_name` |
| **KubernetesSecret** | `pod.imagePullSecrets` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Kubernetes namespace where the Deployment runs | Other workloads deploying into the same namespace |
| `deployment_name` | Name of the Deployment object as created in the cluster | `kubectl` operations and workload references |
| `service` | Kubernetes Service name (empty when the app exposes no ports) | Ingress and Gateway API route backends; service-to-service communication |
| `selector_labels` | The pod selector labels as a `k=v,k=v` string | NetworkPolicy pod selectors, `kubectl get pods -l`, pod-affinity terms in sibling workloads |
| `port_forward_command` | Ready-to-run `kubectl port-forward` command | Local development access without any external exposure |
| `kube_endpoint` | Cluster-internal FQDN (`{service}.{namespace}.svc.cluster.local`) | Application connection strings within the same cluster |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Web service** -- A single-replica HTTP application fronted by a ClusterIP Service, with a readiness probe so traffic only reaches ready pods. Start from the **Web Service Deployment** preset.

**Production web service with autoscaling** -- CPU-based autoscaling between 2 and 10 replicas, a zero-downtime rollout strategy, a pod disruption budget, and a pre-stop connection drain. Start from the **Web Service with Autoscaling and Zero-Downtime Rollouts** preset.

**Background worker** -- A long-running process without ports, such as a queue consumer or event processor. No Kubernetes Service is created. Start from the **Background Worker** preset.

**Hardened production service** -- Passes the restricted Pod Security Standard: non-root with a pinned UID, read-only root filesystem, all capabilities dropped, runtime-default seccomp, no API token mount, and zone-spread replicas with composed identity. Start from the **Hardened Production Service** preset.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the target namespace for the Deployment
- [**Kubernetes ServiceAccount**](/infra-catalog/kubernetes-service-account) -- the composed identity pods run as, carrying workload-identity bindings and pull secrets
- [**Kubernetes Ingress**](/infra-catalog/kubernetes-ingress) -- exposes the Deployment's Service externally with host/path routing and TLS
- [**Kubernetes HTTPRoute**](/infra-catalog/kubernetes-http-route) -- Gateway API exposure referencing the exported Service

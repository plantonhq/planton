# Kubernetes StatefulSet

Deploys a stateful application on any Kubernetes cluster as an apps/v1 StatefulSet: every replica gets a stable name (`{name}-0`, `{name}-1`, ...), stable per-replica DNS through a headless governing Service, and its own PersistentVolumeClaim stamped from volume claim templates. This is the kind for databases, message brokers, and consensus systems — anything where replicas are NOT interchangeable. Supports the full shared workload surface (sidecars and init containers, all environment variable sources, probes, lifecycle hooks, scheduling, security contexts), plus StatefulSet-specific orchestration: update strategy with canary-by-partition, pod management policy, PVC retention policy, and ordinal numbering.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **Headless Governing Service** -- provides stable DNS names for each pod (e.g., `{pod-name}.{service}.{namespace}.svc.cluster.local`); always created as required by the StatefulSet controller
- **Kubernetes StatefulSet** -- the core workload resource with ordered pod management, persistent volume claim templates, container image, resources, environment, probes, and volume mounts
- **ClusterIP Service** -- created only when `container.app.ports` are defined; provides load-balanced access to StatefulSet pods for client traffic
- **Kubernetes Secret** -- created only when `container.app.env.secrets` carry literal values; stores them and mounts them as environment variables
- **PodDisruptionBudget** -- created only when `availability.podDisruptionBudget.enabled` is `true`; enforces minimum availability during voluntary disruptions
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

External exposure is composed, never embedded: this kind exports its governing Service and selector labels, and first-class exposure kinds (KubernetesIngress, KubernetesHttpRoute and the other Gateway API route kinds, with certificates) reference them — so every piece of exposure infrastructure is a visible node in the resource graph.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **A storage class** available for persistent volumes when using `volumeClaimTemplates`. The cluster must support dynamic PV provisioning for the configured storage sizes; pick a class with volume expansion enabled if the data may grow.

## Deploy

### Console

Open the deployment store, find **Kubernetes StatefulSet**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Single-Instance Database** preset for a single-replica database with persistent storage, or **Highly Available Quorum Cluster** for a 3-member quorum system, in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesStatefulSet
metadata:
  name: cache-cluster
  org: acme-corp
  env: prod
spec:
  namespace:
    value: "stateful-apps"
  createNamespace: true
  container:
    app:
      image:
        repo: ghcr.io/acme-corp/cache-node
        tag: v2.0.0
      ports:
        - name: tcp
          containerPort: 6379
          networkProtocol: TCP
          appProtocol: tcp
          servicePort: 6379
  availability:
    replicas: 3
```

```shell
planton apply -f statefulset.yaml
```

This creates a 3-replica StatefulSet with a headless service for stable pod DNS and a ClusterIP service on port 6379. No persistent volumes or pod disruption budget are configured — members keep only in-memory state. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the StatefulSet to a namespace managed by another Infra Component:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: stateful-apps-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline resolves the dependency graph, deploys the namespace first, then provisions the StatefulSet with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Kubernetes StatefulSet. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Persistent volume claim templates** -- Define `volumeClaimTemplates` to give each replica its own persistent volume. Each template specifies a `name`, `size` (e.g., `"10Gi"`), optional `storageClass`, `accessModes`, and `volumeMode`. Mount a template in `container.app.volumeMounts` with a PVC source whose claim name equals the template's name — replica 0's claim becomes `{template}-{statefulset}-0`. Templates are fixed at creation (the Kubernetes API rejects changing them on a live StatefulSet).

**PVC retention** -- `pvcRetentionPolicy.whenDeleted` and `.whenScaled` control what happens to the stamped volumes. The Kubernetes default for both is `Retain`: data outlives the workload, re-creating the set with the same name re-adopts it, and scale-up rejoins members with their data. `Delete` trades that safety for automatic cleanup — right for reproducible caches, wrong for the only copy of a database.

**Replicas and pod management policy** -- Set `availability.replicas` for the member count. `podManagementPolicy` controls scale operations: `"OrderedReady"` (default) proceeds one pod at a time, waiting for readiness — what most clustered systems need for safe bootstrap; `"Parallel"` launches all pods at once for systems that coordinate membership themselves. There is deliberately no autoscaling: stateful members join and leave through application-aware procedures, not HPA.

**Update strategy and canary** -- `updateStrategy.type` is `"RollingUpdate"` (default, one ordinal at a time from the highest down) or `"OnDelete"` (operator-driven). `updateStrategy.partition` canaries by ordinal: only members with ordinal ≥ partition receive the new template — validate one member, then lower the partition step by step.

**Pod disruption budget** -- Enable `availability.podDisruptionBudget` with `minAvailable` or `maxUnavailable` to protect availability during node drains and cluster upgrades. For a 3-member quorum system, `minAvailable: "2"` ensures a drain can never break quorum.

**Environment variables and secrets** -- Use `container.app.env.variables` for configuration and `container.app.env.secrets` for sensitive values. Variables support every Kubernetes source: literals, ValueFromRef to other Infra Components, ConfigMap keys, pod fields (a member can learn its own identity from `metadata.name`), and container resources.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| Kubernetes Namespace | `spec.namespace` | `spec.name` |
| Kubernetes ServiceAccount | `spec.pod.serviceAccount` | `status.outputs.service_account_name` |
| Kubernetes Secret | `spec.pod.imagePullSecrets` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Kubernetes namespace the workload was deployed into | Other workloads deploying into the same namespace |
| `stateful_set_name` | Name of the StatefulSet object as created in the cluster | Operational tooling and audits |
| `service` | The headless governing Service — stable per-pod DNS and load-balanced access | Client applications connecting to the system |
| `selector_labels` | Pod selector labels as a `k=v,k=v` string | NetworkPolicy podSelectors, `kubectl get pods -l`, pod-affinity terms |
| `port_forward_command` | Ready-to-run `kubectl port-forward` command | Local development access without external exposure |
| `kube_endpoint` | Cluster-internal FQDN (`{service}.{namespace}.svc.cluster.local`) | Application connection strings; exposure kinds reference it |
| `pod_dns_template` | Per-member DNS template (`{name}-{ordinal}.{service}.{namespace}.svc.cluster.local`) | Clustered clients building member lists |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Database** -- A single-replica database: one pod with a stable name, one PersistentVolumeClaim stamped from the `data` template that survives restarts and rescheduling, and TCP probes so the pod only receives connections once the engine accepts them. Start from the **Single-Instance Database** preset.

**HA quorum cluster** -- A 3-member quorum system with per-member storage, a pod disruption budget sized to the quorum, and zone spread. Start from the **Highly Available Quorum Cluster** preset.

**Hardened database** -- Passes the Kubernetes restricted Pod Security Standard while running persistent storage: non-root with a pinned UID, read-only root filesystem, all Linux capabilities dropped, the runtime-default seccomp filter, and no API token mount. Start from the **Hardened Database** preset.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the target namespace for the StatefulSet
- [**Kubernetes ServiceAccount**](/infra-catalog/kubernetes-service-account) -- the identity members run as; cloud access federates through its workload-identity configuration
- [**Kubernetes Secret**](/infra-catalog/kubernetes-secret) -- image pull secrets and referenced credential material
- [**Kubernetes Ingress**](/infra-catalog/kubernetes-ingress) -- composes external exposure by referencing the exported Service (never embedded in the workload)

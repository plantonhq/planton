# Kubernetes DaemonSet

Deploys a node agent on any Kubernetes cluster as an apps/v1 DaemonSet: exactly one pod runs on every node that matches the pod's scheduling rules, and pods are added or garbage-collected as nodes join and leave. This is the kind for log shippers, node monitors, storage daemons, and CNI components. Supports the full shared workload surface (sidecars and init containers, all environment variable sources, probes, lifecycle hooks, security contexts), node coverage via selectors/affinity/tolerations, host ports and host namespaces for node-level access, and fleet-safe update strategies with surge support. Credentials are delivered through a Kubernetes Provider Connection or Runner-based delivery.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **Kubernetes DaemonSet** -- the core workload resource that schedules one pod per matching node, with the configured container image, resources, environment, probes, volume mounts, and security contexts
- **Kubernetes Secret** -- created only when `container.app.env.secrets` carry literal values; stores them and mounts them as environment variables
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

There is no replica count — node membership IS the replica count — and no Service or ingress: clients that must reach an agent do so on its node via per-container host ports or the pod's host networking. Identity and permissions are composed, not bundled: `pod.serviceAccount` references a KubernetesServiceAccount, and API permissions come from KubernetesRbac grants targeting that identity.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Node access** for agents that need host-level resources. If the agent reads node logs or metrics, nodes must allow HostPath volume mounts.
- **Taint tolerations** configured under `pod.scheduling.tolerations` if the agent must run on control-plane or specially tainted nodes.

## Deploy

### Console

Open the deployment store, find **Kubernetes DaemonSet**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Log Collector** preset for log forwarding with host path mounts, or **Node Monitor** for a lightweight metrics agent, in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesDaemonSet
metadata:
  name: node-exporter
  org: acme-corp
  env: prod
spec:
  namespace:
    value: "monitoring"
  createNamespace: true
  container:
    app:
      image:
        repo: prom/node-exporter
        tag: v1.8.2
      ports:
        - name: metrics
          containerPort: 9100
          hostPort: 9100
          networkProtocol: TCP
  pod:
    scheduling:
      tolerations:
        - key: node-role.kubernetes.io/control-plane
          operator: Exists
          effect: NoSchedule
```

```shell
planton apply -f daemonset.yaml
```

This creates a DaemonSet running on every node including control-plane nodes, exposing node metrics on each node's own IP at port 9100. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the DaemonSet to a namespace managed by another Infra Component:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: monitoring-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline resolves the dependency graph, deploys the namespace first, then provisions the DaemonSet with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Kubernetes DaemonSet. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Node coverage** -- `pod.scheduling.nodeSelector` and `pod.scheduling.nodeAffinity` NARROW the node set (e.g., `node-type: gpu` for a GPU telemetry agent); `pod.scheduling.tolerations` UNLOCK tainted nodes -- tolerate `node-role.kubernetes.io/control-plane` with `operator: Exists` to cover control-plane nodes. With no rules at all, the agent runs on every schedulable node.

**Host ports and host namespaces** -- Expose node-local endpoints (metrics, health) through each port's `hostPort` -- DaemonSets have no Service. Network observers use `pod.hostNetwork` and process monitors `pod.hostPid`; host-network agents that resolve cluster services should also set `pod.dnsPolicy: ClusterFirstWithHostNet`.

**Security context** -- Agents that touch kernel interfaces need specific capabilities via `container.app.securityContext.capabilities.add` (prefer NET_ADMIN or SYS_ADMIN over `privileged: true`). Observability agents usually need no elevation: run non-root, read-only root filesystem, all capabilities dropped.

**Volume mounts** -- HostPath volume mounts are the DaemonSet's signature storage: node logs (`/var/log`), the container runtime socket, kernel interfaces. Mount them read-only wherever possible. ConfigMap mounts inject agent configuration files.

**Update strategy** -- `updateStrategy.type: "RollingUpdate"` (default) replaces pods node by node within `maxUnavailable`/`maxSurge`; surging gives gapless updates for agents that must never miss data, but cannot combine with exclusive host ports. `"OnDelete"` replaces pods only when deleted -- full manual control per node. Pair rollouts with `minReadySeconds` so a crashing agent build is caught before it covers the fleet.

**Identity** -- Reference a KubernetesServiceAccount via `pod.serviceAccount` for agents that read the Kubernetes API (log enrichment, node metadata); grant permissions with KubernetesRbac cluster-scope read rules -- never cluster-admin, since an agent's credential is the widest-deployed secret in the cluster.

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
| `namespace` | Kubernetes namespace the workload was deployed into | Other workloads deploying into the same namespace |
| `daemon_set_name` | Name of the DaemonSet object as created in the cluster | Operational tooling and audits |
| `selector_labels` | Pod selector labels as a `k=v,k=v` string | NetworkPolicy podSelectors, `kubectl get pods -l`, pod-affinity terms |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Log collector** -- A node-level log forwarder with host path mounts for `/var/log`, buffering headroom, and a control-plane toleration. Suitable for Fluent Bit, Fluentd, or Filebeat. Start from the **Log Collector** preset.

**Node monitor** -- A lightweight metrics agent on every node with a host-port scrape endpoint and minimal resources. Suitable for Prometheus node exporter and similar collectors. Start from the **Node Monitor** preset.

**Hardened agent** -- Passes the Kubernetes restricted Pod Security Standard: non-root, read-only root filesystem, all capabilities dropped, runtime-default seccomp, no API token mount. Start from the **Hardened Agent** preset.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the target namespace for the DaemonSet
- [**Kubernetes ServiceAccount**](/infra-catalog/kubernetes-service-account) -- the identity agents run as when they read the Kubernetes API
- [**Kubernetes RBAC**](/infra-catalog/kubernetes-rbac) -- grants the agent identity its API permissions (composed, never bundled)
- [**Kubernetes Secret**](/infra-catalog/kubernetes-secret) -- image pull secrets and referenced credential material

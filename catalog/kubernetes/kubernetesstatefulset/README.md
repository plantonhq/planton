# Kubernetes StatefulSet

## Overview

**KubernetesStatefulSet** is a catalog kind that deploys a stateful application to a Kubernetes cluster as an apps/v1 StatefulSet. Every replica gets a stable name (`<name>-0`, `<name>-1`, ...), a stable per-replica DNS name through a headless governing Service the module derives from the resource name, and its own PersistentVolumeClaim stamped from `volumeClaimTemplates`. This is the kind for databases, message brokers, and consensus systems — anything where replicas are NOT interchangeable.

For stateless services use **KubernetesDeployment**; for run-to-completion work use **KubernetesJob** / **KubernetesCronJob**; for one-pod-per-node agents use **KubernetesDaemonSet**.

## What Gets Created

- **StatefulSet** — the apps/v1 workload, with pod template assembled from `spec.container` (app + sidecars) and `spec.pod` (identity, scheduling, security, DNS, termination)
- **Headless governing Service** — shares the workload's name; gives each replica its stable DNS identity and carries load-balanced client traffic on the ports declared under `container.app.ports`
- **PersistentVolumeClaims** — one per replica per template in `volumeClaimTemplates`, named `<template>-<name>-<ordinal>`, surviving pod restarts and rescheduling
- **PodDisruptionBudget** — when `availability.podDisruptionBudget.enabled` is true; shares the workload's name
- **Env Secret** — `<name>-env-secrets`, materialized only when the spec carries literal secret env values
- **Namespace** — only when `createNamespace` is true

## Composition, Not Bundling

The workload deliberately owns nothing but the workload:

- **Identity is composed.** `pod.serviceAccount` references a **KubernetesServiceAccount** resource (or names an existing ServiceAccount literally). Workload-identity annotations, pull-secret attachment, and RBAC grants (via **KubernetesRbac**) live on the identity — the StatefulSet never creates ServiceAccounts or RBAC objects.
- **Exposure is composed.** The spec has no ingress block. The component exports its governing `service`, `kubeEndpoint`, and `selectorLabels`, and first-class exposure kinds (KubernetesHttpRoute and the other Gateway API route kinds, with certificates) reference those outputs. Every piece of exposure infrastructure stays a visible node in the resource graph.
- **Configuration is composed.** ConfigMaps are first-class **KubernetesConfigMap** resources that containers mount or import by name — the workload spec carries no inline ConfigMap definitions.
- **Member addressing is exported.** The `podDnsTemplate` output is the template for each replica's stable DNS name; member-aware clients substitute the ordinal to build their member lists.

## Spec Highlights

- **`container.app`** — the main container, using the shared workload container model: image (repo + tag, the deploy-pipeline injection point), ports (with `servicePort` driving the Service), env (variables, secrets, bulk `envFrom`), probes, volume mounts, lifecycle hooks, and a container security context. `container.sidecars` are full containers with the same surface.
- **`pod`** — the shared pod model: ServiceAccount reference, init containers, scheduling (node selection, tolerations, affinity, topology spread), pod security context (including `fsGroup` — essential for non-root writers on persistent volumes), DNS, host aliases, and `terminationGracePeriodSeconds` (size it to a clean member handoff).
- **`volumeClaimTemplates`** — per-replica storage; mount a template via a `pvc` volume mount whose `claimName` equals the template's `name`. Leave `storageClass` unset for the cluster default.
- **`availability`** — replica count, PodDisruptionBudget (set `minAvailable` to your quorum size), `minReadySeconds`, and `revisionHistoryLimit`. Deliberately no autoscaling: stateful members join and leave through application-aware procedures, not HPA.
- **`updateStrategy`** — `RollingUpdate` (highest ordinal down, one at a time) with `partition` for canary-by-ordinal rollouts, or `OnDelete` for operator-driven updates.
- **`podManagementPolicy`** — `OrderedReady` (default; safe sequential bootstrap) or `Parallel` (all pods at once, for systems that coordinate their own membership).
- **`pvcRetentionPolicy`** — what happens to the stamped PVCs `whenDeleted` and `whenScaled`; the default retains everything.
- **`ordinals.start`** — alternate ordinal base for numbering conventions or ordinal-range migrations.

## Outputs

After deployment, the following are available in `status.outputs`:

| Output | Description |
|--------|-------------|
| `namespace` | The namespace the workload was deployed into |
| `stateful_set_name` | The name of the StatefulSet object in the cluster |
| `service` | The headless governing Service — stable per-pod DNS and load-balanced client access |
| `selector_labels` | Pod selector labels as `k=v,k=v` — for NetworkPolicies, `kubectl -l`, and sibling anti-affinity |
| `port_forward_command` | Ready-to-run `kubectl port-forward` command for local access without exposure |
| `kube_endpoint` | In-cluster DNS endpoint (`<name>.<namespace>.svc.cluster.local`) — what exposure kinds and sibling workloads connect to |
| `pod_dns_template` | Template for each replica's stable DNS name (`<name>-<ordinal>.<service>.<namespace>.svc.cluster.local`) — how clustered clients address individual members |

## Example

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesStatefulSet
metadata:
  name: orders-db
spec:
  namespace:
    value: orders
  container:
    app:
      image:
        repo: postgres
        tag: "16.3"
      ports:
        - name: db
          containerPort: 5432
          servicePort: 5432
      readinessProbe:
        tcpSocket:
          portNumber: 5432
        periodSeconds: 5
      volumeMounts:
        - name: data
          mountPath: /var/lib/postgresql/data
          pvc:
            claimName: data
  pod:
    securityContext:
      fsGroup: 999
  availability:
    replicas: 3
    podDisruptionBudget:
      enabled: true
      minAvailable: "2"
  volumeClaimTemplates:
    - name: data
      size: 20Gi
      accessModes:
        - ReadWriteOnce
```

Deploy with `planton apply -f statefulset.yaml`. Replica 0 is then addressable at `orders-db-0.orders-db.orders.svc.cluster.local`, and the whole cluster load-balances through `orders-db.orders.svc.cluster.local`.

## When to Use

Use **KubernetesStatefulSet** when replicas need stable identity, stable per-replica DNS, or per-replica persistent storage: databases, brokers (Kafka, RabbitMQ), consensus stores (etcd, ZooKeeper), and replicated caches with persistence.

**Do NOT use** when:

- Replicas are interchangeable — use **KubernetesDeployment**; it rolls out faster and scales freely
- The work runs to completion — use **KubernetesJob** or **KubernetesCronJob**
- You need one pod per node — use **KubernetesDaemonSet**

## References

- [Kubernetes StatefulSets Documentation](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/)
- [StatefulSet Basics Tutorial](https://kubernetes.io/docs/tutorials/stateful-application/basic-stateful-set/)
- [PersistentVolumeClaim Retention](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#persistentvolumeclaim-retention)
- [StatefulSet API Reference](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/stateful-set-v1/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

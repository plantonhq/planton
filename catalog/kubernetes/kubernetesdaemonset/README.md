# Kubernetes DaemonSet

## Overview

**KubernetesDaemonSet** is a catalog kind that deploys a node agent to a Kubernetes cluster as an apps/v1 DaemonSet: exactly one pod runs on every node that matches the pod's scheduling rules, and pods are added or garbage-collected as nodes join and leave. This is the kind for log shippers, node monitors, storage daemons, and CNI components.

There is no replica count — node membership IS the replica count — and no Service or ingress: clients that must reach an agent do so on its node via per-container `hostPort` or `pod.hostNetwork`. For stateless services use **KubernetesDeployment**; for stateful members use **KubernetesStatefulSet**; for run-to-completion work use **KubernetesJob** / **KubernetesCronJob**.

## What Gets Created

- **DaemonSet** — the apps/v1 workload, with the pod template assembled from `spec.container` (agent + sidecars) and `spec.pod` (identity, scheduling, host access, security, DNS, termination)
- **Env Secret** — `<name>-env-secrets`, materialized only when the spec carries literal secret env values
- **Namespace** — only when `createNamespace` is true

Deliberately absent: Services, ingress, HPAs, and PDBs — none of them apply to the one-pod-per-node model.

## Composition, Not Bundling

- **Identity is composed.** `pod.serviceAccount` references a **KubernetesServiceAccount** resource (or names an existing ServiceAccount literally). API permissions come from **KubernetesRbac** grants targeting that identity — many node agents (kube-state-style collectors, CNI components) need cluster-scoped reads, and those grants are visible resources in the graph, not side effects of the workload.
- **Configuration is composed.** ConfigMaps are first-class **KubernetesConfigMap** resources that containers mount or import by name.
- **Node access composes from shared building blocks.** `pod.hostNetwork` / `pod.hostPid` for host namespaces, per-container port `hostPort` for node-IP exposure, HostPath volume mounts for the node filesystem, and the container `securityContext` for capabilities — each an explicit, reviewable line in the spec.

## Spec Highlights

- **`container.app`** — the main agent container, using the shared workload container model: image (repo + tag), env, probes, volume mounts (HostPath sources for node logs and runtime sockets are the classic DaemonSet pattern), lifecycle hooks, and a container security context. Node-local endpoints are exposed through each port's `hostPort` — DaemonSets have no Service. `container.sidecars` are full containers with the same surface.
- **`pod.scheduling`** — controls WHICH nodes run the agent: `nodeSelector` and `nodeAffinity` narrow the node set, `tolerations` unlock tainted nodes (control-plane nodes carry a `NoSchedule` taint — agents that should cover them need an explicit toleration).
- **`pod.hostNetwork` / `pod.hostPid`** — host namespaces for agents that observe the node itself; host-network agents that resolve cluster services should also set `dnsPolicy: ClusterFirstWithHostNet`.
- **`updateStrategy`** — `RollingUpdate` (node by node within `maxUnavailable`/`maxSurge` bounds) or `OnDelete` (per-node manual control). `maxSurge` gives gapless updates for agents that must never miss data, but requires old and new pods to coexist on a node — it cannot be combined with exclusive host ports.
- **`minReadySeconds` / `revisionHistoryLimit`** — rollout flap detection and rollback depth.

## Outputs

After deployment, the following are available in `status.outputs`:

| Output | Description |
|--------|-------------|
| `namespace` | The namespace the workload was deployed into |
| `daemon_set_name` | The name of the DaemonSet object in the cluster |
| `selector_labels` | Pod selector labels as `k=v,k=v` — for NetworkPolicies, `kubectl -l`, and pod-affinity terms in sibling workloads |

## Example

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesDaemonSet
metadata:
  name: log-collector
spec:
  namespace:
    value: observability
  container:
    app:
      image:
        repo: fluent/fluent-bit
        tag: "3.0.7"
      resources:
        requests:
          cpu: 50m
          memory: 128Mi
        limits:
          cpu: 500m
          memory: 512Mi
      volumeMounts:
        - name: varlog
          mountPath: /var/log
          readOnly: true
          hostPath:
            path: /var/log
            type: Directory
  pod:
    scheduling:
      tolerations:
        - key: node-role.kubernetes.io/control-plane
          operator: Exists
          effect: NoSchedule
```

Deploy with `planton apply -f daemonset.yaml`. One collector pod runs on every node, including control-plane nodes, reading node logs through the read-only HostPath mount.

## When to Use

Use **KubernetesDaemonSet** when the unit of deployment is "per node": log and metrics collection, node security agents, storage daemons, network plugins, per-node caches.

**Do NOT use** when:

- You need N interchangeable replicas — use **KubernetesDeployment**
- Replicas need identity and their own storage — use **KubernetesStatefulSet**
- The work runs to completion — use **KubernetesJob** or **KubernetesCronJob**

## References

- [Kubernetes DaemonSet Documentation](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/)
- [DaemonSet Rolling Updates](https://kubernetes.io/docs/tasks/manage-daemon/update-daemon-set/)
- [Taints and Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/)
- [DaemonSet API Reference](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/daemon-set-v1/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

# Kubernetes Deployment

## Overview

**KubernetesDeployment** is a catalog kind that runs a long-running, stateless application on a Kubernetes cluster as an apps/v1 Deployment, fronted by a ClusterIP Service. It is the workhorse workload kind: replicas are interchangeable, updates roll out gradually, and scaling is horizontal. For workloads needing stable identity or per-replica storage use **KubernetesStatefulSet**; for run-to-completion work use **KubernetesJob** / **KubernetesCronJob**; for one-pod-per-node agents use **KubernetesDaemonSet**.

A single manifest produces:

- **Deployment** (apps/v1) — the pods, their containers, and the rollout machinery
- **Service** (ClusterIP) — created when the app container declares ports; skipped entirely for port-less workers
- **HorizontalPodAutoscaler** — when `availability.horizontalPodAutoscaling.enabled` is true
- **PodDisruptionBudget** — when `availability.podDisruptionBudget.enabled` is true
- **Satellite Secrets** — one Opaque Secret materializing literal secret env values (collected across the app container, sidecars, and init containers), and one `kubernetes.io/dockerconfigjson` pull Secret built from the registry logins the pod declares on `pod.imageRegistries`
- **Namespace** — only when `createNamespace` is true

## Composition, Not Bundling

Two things this kind deliberately does **not** create:

- **Identity.** Pods run as the ServiceAccount referenced in `spec.pod.serviceAccount` — a literal name or a reference to a **KubernetesServiceAccount** resource, which owns workload-identity annotations and pull-secret attachment. Permissions attach to that identity through **KubernetesRbac** grants. The workload never creates ServiceAccounts or RBAC objects of its own, so identity and permissions are auditable resources in the graph rather than side effects of a deployment.
- **Exposure.** No ingress configuration exists anywhere in this spec. The workload exports its Service name (`service`), in-cluster DNS endpoint (`kube_endpoint`), and pod selector labels (`selector_labels`) as outputs, and first-class exposure kinds — Gateway API routes like **KubernetesHttpRoute**, gateways, certificates — reference those outputs. Every piece of exposure infrastructure is a visible, independently managed node in the resource graph, and a workload's exposure can change without touching the workload.

## Deploy-Target Contract

This kind is a deployment target for application delivery pipelines. Pipelines inject the freshly built artifact through exactly two stable paths:

- **`spec.version`** — the deployment track, set from the git branch (`main`, or a flattened branch name like `review-42`). Stamped as a pod label so multiple tracks of one app coexist in a namespace with disjoint traffic. The version label is deliberately excluded from the pod *selector* labels — selectors are immutable on Deployments, and the version changes on every pipeline run.
- **`spec.container.app.image`** — the image split into `repo` and `tag`, so a pipeline injects a freshly built tag without rewriting the whole reference.

These paths are part of the kind's public contract.

## Spec Highlights

### Container surface (`spec.container`)

The `app` container and every `sidecars` entry use the same fully-modeled container shape — anything expressible on the app container is equally expressible on a sidecar:

- **Image** — `repo` + `tag` (both required; pin versions, avoid `latest`), plus `imagePullPolicy`
- **Ports** — named ports with `containerPort`, optional `servicePort` (drives the Service mapping, e.g. container 8080 served on 80), `appProtocol` hints for meshes and L7 load balancers
- **Environment** — plain variables (literal values, Kubernetes `fieldRef`/`resourceFieldRef`/`configMapKeyRef` sources, and cross-resource references to other Planton resources' outputs), secret variables (literal values materialize into the managed env Secret; `secretRef` wires existing Secrets directly), and bulk `envFrom` imports
- **Probes** — liveness (restart wedged processes), readiness (the probe that makes rollouts zero-downtime), and startup (protects slow-booting apps from premature liveness kills)
- **Volume mounts** — each mount declares its path *and* carries its volume source (ConfigMap, Secret, EmptyDir, HostPath, PVC); the module derives the pod volume list from the union of all containers' mounts
- **Lifecycle hooks** — `postStart` and `preStop`, including the kubelet-native `sleep` action that needs no sleep binary in the image
- **Security context** — the complete restricted-profile checklist: `runAsNonRoot`, `readOnlyRootFilesystem`, `allowPrivilegeEscalation`, capability add/drop, seccomp profiles

### Pod configuration (`spec.pod`)

Shared by every replica: the ServiceAccount reference, `automountServiceAccountToken` (set false for apps that never call the Kubernetes API), init containers (full containers, run to completion in order), extra pod labels/annotations, scheduling (node selectors, tolerations, node/pod affinity, topology spread constraints), pod-level security context (`fsGroup`, supplemental groups, sysctls, pod-wide seccomp), termination grace period, DNS policy and config, host aliases, host network/PID, priority class, and runtime class.

### Availability (`spec.availability`)

- **`replicas`** — desired count; the floor when autoscaling is enabled (default 1)
- **`horizontalPodAutoscaling`** — CPU target as average utilization percent of requests, memory target as an absolute per-pod quantity (e.g. `1Gi`), scaling between `replicas` and `maxReplicas`. Prefer CPU for most services — memory rarely falls after scale-out, which makes it a poor scale-in signal
- **`strategy`** — `RollingUpdate` (default) with `maxUnavailable`/`maxSurge`, or `Recreate` for workloads where two versions cannot coexist (exclusive volume locks, incompatible schemas), at the cost of downtime. `maxUnavailable: "0"` with `maxSurge: "1"` plus a readiness probe is the zero-downtime pattern
- **`podDisruptionBudget`** — `minAvailable` or `maxUnavailable` (exactly one) guarding against voluntary disruptions like node drains and upgrades
- **Rollout tuning** — `minReadySeconds` (a cheap flap detector: new pods must stay ready this long before counting as available), `revisionHistoryLimit` (rollback depth), `progressDeadlineSeconds` (when a stuck rollout is marked failed), and `paused` (batch several spec changes into one rollout)

## Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Description |
|--------|-------------|
| `namespace` | The namespace the workload was deployed into |
| `deployment_name` | The name of the Deployment object as created in the cluster |
| `service` | The Kubernetes Service fronting the replicas. Empty when the app container exposes no ports (no Service is created) |
| `selector_labels` | The pod selector labels as a `k=v,k=v` string — the exact labels the Service selects on, ready for NetworkPolicy podSelectors, `kubectl get pods -l`, and pod-affinity terms in sibling workloads |
| `port_forward_command` | Ready-to-run `kubectl port-forward` command for reaching the workload from a developer machine without any external exposure |
| `kube_endpoint` | In-cluster DNS endpoint of the Service (e.g. `my-app.my-ns.svc.cluster.local`) — the handle exposure kinds and sibling workloads connect to |

## Complete Example

A production web service: autoscaled, zero-downtime rollouts, disruption-protected, running as a composed identity:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesDeployment
metadata:
  name: checkout
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: shop
  container:
    app:
      image:
        repo: ghcr.io/acme/checkout
        tag: v1.4.2
      resources:
        requests:
          cpu: 100m
          memory: 256Mi
        limits:
          cpu: 1000m
          memory: 1Gi
      ports:
        - name: http
          containerPort: 8080
          servicePort: 80
          appProtocol: http
      env:
        variables:
          - name: LOG_LEVEL
            value: info
        secrets:
          - name: STRIPE_API_KEY
            secretRef:
              name: payment-secrets
              key: stripe-api-key
      readinessProbe:
        httpGet:
          path: /healthz
          portNumber: 8080
        periodSeconds: 5
      lifecycle:
        preStop:
          sleep:
            seconds: 10
  pod:
    serviceAccount:
      valueFrom:
        kind: KubernetesServiceAccount
        name: checkout-identity
        fieldPath: status.outputs.service_account_name
    automountServiceAccountToken: false
    scheduling:
      topologySpreadConstraints:
        - maxSkew: 1
          topologyKey: topology.kubernetes.io/zone
          whenUnsatisfiable: ScheduleAnyway
  availability:
    replicas: 2
    horizontalPodAutoscaling:
      enabled: true
      maxReplicas: 10
      targetCpuUtilizationPercent: 70
    strategy:
      type: RollingUpdate
      maxUnavailable: "0"
      maxSurge: "1"
    podDisruptionBudget:
      enabled: true
      minAvailable: "1"
```

To publish this service at a hostname, deploy a Gateway API route (e.g. KubernetesHttpRoute) that references this workload's `service` / `kube_endpoint` outputs.

## How It Works

This kind includes both **Pulumi** (Go) and **Terraform** (HCL) modules that:

1. Resolve the target namespace (literal value or resolved reference) and optionally create it
2. Create the satellite Secrets before the Deployment — pods reference them by name at startup, and a pod that starts before its env Secret exists crashes
3. Create the Deployment with the full container, pod, and availability configuration; derive the pod volume list from the union of all containers' volume mounts
4. Create the ClusterIP Service when the app container declares ports, mapping each `servicePort` to its `containerPort`
5. Create the HPA and PDB when enabled, both bound to the workload's selector labels
6. Export the namespace, names, selector labels, port-forward command, and in-cluster endpoint for downstream composition

Both IaC implementations have feature parity and follow identical logic.

## When to Use

Use **KubernetesDeployment** when you need:

- Stateless HTTP/gRPC services and APIs — the standard microservice shape
- Background workers and queue consumers (declare no ports; no Service is created)
- Horizontal scaling, gradual rollouts, and disruption budgets on interchangeable replicas
- A pipeline deploy target with stable version and image injection paths

**Do NOT use** when:

- Replicas need stable network identity or per-replica persistent volumes — use **KubernetesStatefulSet**
- The work runs to completion — use **KubernetesJob** or **KubernetesCronJob**
- Exactly one pod must run on every node — use **KubernetesDaemonSet**

## Prerequisites

- **Kubernetes Cluster**: Access to a Kubernetes cluster (any distribution: GKE, EKS, AKS, self-hosted)
- **Credentials**: Kubernetes cluster credentials (kubeconfig)
- **Namespace**: The target namespace must exist, be referenced as a KubernetesNamespace resource, or be created via `createNamespace`
- **Metrics server**: Required in the cluster when enabling horizontal pod autoscaling
- **Identity resources**: A KubernetesServiceAccount (and KubernetesRbac grants) when pods need a non-default identity or Kubernetes API permissions

## Best Practices

1. **Always set a readiness probe on serving workloads**: It is the piece that makes rolling updates zero-downtime — traffic only reaches pods that report ready
2. **Use `maxUnavailable: "0"` + `maxSurge: "1"` + a preStop sleep for production services**: New pods come up before old ones go away, and terminating pods keep serving while endpoint removal propagates
3. **Set resource requests deliberately**: Requests drive scheduling and are the denominator for CPU-based autoscaling; limits are the runtime ceiling
4. **Scale on CPU, not memory**: Memory rarely falls after scale-out, so it is a poor scale-in signal
5. **Harden by default**: `runAsNonRoot`, `readOnlyRootFilesystem` with an EmptyDir for `/tmp`, drop ALL capabilities, `RuntimeDefault` seccomp, and `automountServiceAccountToken: false` for apps that never call the Kubernetes API
6. **Compose identity and exposure**: Reference a KubernetesServiceAccount for identity; attach ingress kinds to the exported outputs for exposure. Keep the workload manifest about the workload

## References

- [Kubernetes Deployments Documentation](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [Horizontal Pod Autoscaling](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/)
- [Pod Disruption Budgets](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/)
- [Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
- [Deployment API Reference](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/deployment-v1/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

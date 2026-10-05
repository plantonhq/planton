# Ray Cluster

Deploys a Ray cluster on Kubernetes -- one resource declares one `RayCluster` custom resource (`rayclusters.ray.io/v1`) that the KubeRay operator reconciles into a head pod, worker groups, and the Services that Ray clients, jobs, and dashboards connect to. PREREQUISITE: a KubernetesKubeRayOperator whose watch scope covers this namespace (cluster-wide with the operator's defaults) -- nothing reconciles this declaration without it. THE STATE TRUTH, READ THIS FIRST: the head pod's GCS (Global Control Store) holds the cluster's control state IN MEMORY -- without GCS fault tolerance, losing the head loses every job, actor, and worker registration; the operator rebuilds an EMPTY cluster. Uses a Kubernetes Provider Connection for cluster access.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **RayCluster CR** -- the declaration the KubeRay operator reconciles into:
  - A head pod running the GCS, the dashboard, the job submission server, and (with autoscaling) the autoscaler sidecar -- with `num-cpus: "0"` rendered when tasks are kept off the head (the default, the production posture)
  - Worker pods per group -- homogeneous sets with their own sizing, accelerator limits, Ray start parameters, and placement
  - The head Service (`<name>-head-svc`) carrying the CLIENT (10001), DASHBOARD/Job-API (8265), and GCS (6379) ports
  - In token auth mode (the catalog default): a bearer-token Secret named exactly after this resource (key `auth_token`), required by every API surface and exported as the credential handle
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

The apply is deliberately NON-blocking: cluster readiness depends on the operator (multi-GB Ray image pulls, sidecar injection, GCS startup) -- the CR applies and the operator reconciles; nothing blocks on a controller.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **A KubeRay operator watching this namespace** -- deploy KubernetesKubeRayOperator first. Under a fenced operator (non-empty watch namespaces), a declaration OUTSIDE the fence is ignored WITHOUT an error -- the cluster simply never starts.
- **A name within budget** -- keep `metadata.name` at 40 characters or fewer: the operator derives `<name>-head-svc` and per-group worker pod names against the 63-character Kubernetes cap. Both engines fail loudly over the budget; keep worker group names short (at most 24 characters) for the same reason.
- **A Valkey for GCS fault tolerance, if you enable it** -- deploy a KubernetesValkey first, IN THIS SAME NAMESPACE: the credential rides a secretKeyRef, which cannot cross namespaces (or replicate the auth Secret).
- **Accelerator nodes for GPU groups** -- GPU nodes are usually tainted; pair each accelerator group's extended resource limits with its node selector and tolerations, or the pods never land.
- **Registry access for ~2.5GB Ray images** -- expect long first-node pulls; pre-pull DaemonSets or baked node images are the production mitigations.

## Deploy

### Console

Open the deployment store, find **Ray Cluster**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Lab preset** for a single-node cluster with tasks on the head, or the **Production autoscaling preset** for the autoscaling CPU + GPU fleet with GCS fault tolerance in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesRayCluster
metadata:
  name: ray-lab
  org: acme-corp
  env: dev
spec:
  namespace:
    value: "ml-lab"
  createNamespace: true
  rayVersion: "2.52.0"
  head:
    resources:
      requests:
        cpu: "1"
        memory: 4Gi
      limits:
        cpu: "2"
        memory: 8Gi
    scheduleTasksOnHead: true
```

```shell
planton apply -f ray-cluster.yaml
```

This deploys a single-node lab cluster: the head is the only capacity and tasks schedule onto it (the deliberate lab arm -- production keeps work off the head). Token authentication rides by absence. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the cluster to its state store:

```yaml
spec:
  gcsFaultTolerance:
    enabled: true
    redisAddress:
      valueFrom:
        kind: KubernetesValkey
        name: ray-state
        fieldPath: status.outputs.kube_endpoint
    redisPasswordSecret:
      name:
        valueFrom:
          kind: KubernetesValkey
          name: ray-state
          fieldPath: status.outputs.password_secret.name
      key: default
```

The InfraPipeline deploys the Valkey first, then provisions the Ray cluster wired to its endpoint and auth Secret.

## Key Configuration

These are the most important decisions when configuring a Ray Cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The state truth decides everything else** -- The head's GCS holds control state IN MEMORY. `gcsFaultTolerance` moves it into an external Redis-protocol store so a replaced head RECOVERS jobs, actors, and workers instead of rebuilding an empty cluster. Compose a KubernetesValkey: the foreign-key defaults wire its write-Service endpoint (always the primary) and its `<name>-auth` Secret (the key is the ACL USERNAME -- `default` unless you declared users). `externalStorageNamespace` empty = derived from the CR's UID (safe); set it explicitly only when state must survive delete-and-recreate of the declaration itself.

**Version and image move in lockstep** -- `rayVersion` (required; `2.52.0` is the version KubeRay v1.6.x ships samples for) drives the default image (`rayproject/ray:<ray_version>`) and the operator's command shaping. The operator runs a custom `image` AS GIVEN -- a mismatch fails at RUNTIME, not at apply. Custom images (dependencies baked in, CUDA variants like `-gpu`) must keep the Ray inside identical to the declared version.

**The unloaded head is the production posture** -- `head` and `head.resources` are REQUIRED: an unsized head is the classic Ray outage (GCS + dashboard + scheduler share it; upstream starts production heads at 4 CPU / 8Gi). With `scheduleTasksOnHead` empty (= false), the modules render `num-cpus: "0"` so application work stays off the head -- a task-loaded head starves the GCS. Pin production heads to stable nodes: without GCS fault tolerance, losing the head's node loses the cluster's state.

**Worker groups are homogeneous; heterogeneity lives between them** -- A CPU group plus a GPU group is the classic two-group fleet. Each group's `resources` is required (Ray schedules against DECLARED capacity), names must be unique lowercase DNS labels, and sizing must order `min_replicas <= replicas <= max_replicas`. A cluster with no groups is LEGAL -- tasks run on the head (labs only).

**Accelerators ride `extraResourceLimits` -- limits only** -- `"nvidia.com/gpu": "1"` (or AMD/TPU keys) lands in the container LIMITS: Kubernetes rejects requests-without-limits for extended resources, and Ray discovers accelerators from the limits. Pair with the group's node selector and tolerations so pods land on (usually tainted) accelerator nodes.

**Ray's own autoscaler, not an HPA** -- `autoscaling` injects an application-aware sidecar into the head pod that scales each group between its bounds on SCHEDULER demand (queued tasks and actors), which CPU-watching autoscalers cannot see. Worker `replicas` become the INITIAL size only; a `replicas: 0` GPU group materializes the moment tasks demand accelerators. `idleTimeoutSeconds` (empty = 60) is the reclaim-idle-GPUs dial; `upscalingMode` uses the CR's own capitalized vocabulary (`Default`/`Aggressive`/`Conservative`).

**Token auth is the default -- because an open Ray cluster runs your code** -- With `auth.mode` empty or `token`, the operator provisions the bearer-token Secret and every surface requires it. The operator's OWN default is disabled, so the modules render the auth block ALWAYS -- absence never silently deploys the legacy open cluster, where anyone reaching port 8265 runs arbitrary code. Set `disabled` only for fenced labs; `existingTokenSecretName` brings your own token (key `auth_token`).

**Suspend is the idle-GPU cost lever** -- the operator deletes head and worker PODS but keeps the declaration and (with GCS fault tolerance) the external state; un-suspend to resume. The classic overnight-fleet move.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesValkey** | `gcsFaultTolerance.redisAddress` | `status.outputs.kube_endpoint` |
| **KubernetesValkey** | `gcsFaultTolerance.redisPasswordSecret.name` | `status.outputs.password_secret.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the cluster runs in | Locating the cluster for diagnostics |
| `head_service` | The head Service (`<name>-head-svc`) every endpoint rides | Service-level wiring |
| `client_endpoint` | In-cluster CLIENT endpoint (port 10001) | What `ray.init("ray://…")` dials from notebooks and applications |
| `dashboard_endpoint` | In-cluster DASHBOARD/Job-API endpoint (port 8265) | Job submission and the web dashboard; authenticated in token mode |
| `gcs_endpoint` | In-cluster GCS endpoint (port 6379) | What `ray start --address` joins |
| `auth_token_secret` | The bearer-token Secret (key `auth_token`); unset when auth is disabled | The credential handle for dashboard/job/client APIs |
| `port_forward_command` | kubectl port-forward one-liner | Reaching the dashboard from a workstation |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Lab cluster** -- A single sized head with tasks scheduled onto it, no worker groups, token auth by absence. Start from the **Lab preset**.

**Production ML fleet** -- An unloaded 4-CPU/8Gi head, a CPU group plus a zero-replica GPU group (accelerator limits paired with node selector and taints), the autoscaler enabled, and GCS fault tolerance composed from a co-located KubernetesValkey. Start from the **Production autoscaling preset**.

## Works With

- [**KubeRay Operator**](/infra-catalog/kubernetes-kube-ray-operator) -- the PREREQUISITE: the controller that reconciles this declaration; its watch scope must cover this namespace
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the namespace for the cluster
- [**Valkey**](/infra-catalog/kubernetes-valkey) -- the external state store for GCS fault tolerance (deploy it in the SAME namespace -- the credential secretKeyRef cannot cross namespaces)

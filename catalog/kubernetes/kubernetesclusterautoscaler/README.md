# Kubernetes Cluster Autoscaler

## When NOT to Use This

**One installation per cluster.** The autoscaler leader-elects and owns
the cluster-wide scaling decision — a second installation would fight
the first over every scale-up. The Helm release name is therefore fixed
to `cluster-autoscaler` and never derives from `metadata.name`.

Also not the right kind when:

- **The cluster is GKE or AKS with the managed autoscaler** — both
  platforms ship a MANAGED autoscaler configured as a toggle on the node
  pool itself (through the cluster kinds). Deploying this component
  there is the exception, not the rule: use it only where self-managed
  autoscaling is the real posture (AKS opting out of the managed
  autoscaler, self-managed clusters on VMSS or GCE).
- **You want right-sized machines launched on demand** — the autoscaler
  grows and shrinks EXISTING node groups; for AWS clusters that want
  instance types picked per pending pod instead of pre-defined groups,
  KubernetesKarpenter is the modern alternative.
- **You need production scaling on the kwok arm** — the KWOK simulation
  provider creates FAKE nodes for testing scaling policies; it is never
  a production arm.

## Overview

**KubernetesClusterAutoscaler** installs the Kubernetes Cluster
Autoscaler from the official Helm chart (`cluster-autoscaler` at
`https://kubernetes.github.io/autoscaler`). The autoscaler grows and
shrinks EXISTING node groups: when pods are unschedulable it raises the
desired size of a matching group (an EC2 Auto Scaling group, an Azure
VMSS, a Cluster API MachineDeployment, ...), and it scales groups back
down when nodes sit underutilized. It earns its keep on clusters whose
node capacity is organized as pre-defined groups — EKS with ASGs,
Cluster API / self-managed clusters, and providers without a managed
autoscaler.

The typed spec covers the chart's meaningful configuration surface;
`extra_args` carries any of the autoscaler's 100+ tuning flags beyond
the typed set (the chart's own extraArgs contract), and `helm_values`
remains as the escape hatch for chart values (merged last, Helm `-f`
semantics, identical on both engines).

**Key design points:**

- **Exactly one provider arm** — the autoscaler binary supports many
  providers, but one installation talks to exactly one: `aws`, `azure`,
  `gce`, `cluster_api`, `civo`, or `kwok` (a required oneof).
- **Tag-based auto-discovery is the recommended AWS mode** — the
  autoscaler manages every ASG carrying the standard tags
  (`k8s.io/cluster-autoscaler/enabled` +
  `k8s.io/cluster-autoscaler/<cluster_name>`); new node groups enroll by
  tagging alone. Static `node_groups` lists are the alternative, on AWS
  and Azure alike (exactly one of the two).
- **The typed `scaling` block covers the flags every real installation
  tunes** — expander choice, scale-down thresholds and cool-downs,
  node-group balancing, scan interval — rendered into the chart's
  extraArgs; `extra_args` entries win over the typed block on key
  collision.
- **Version alignment matters** — keep the autoscaler's MINOR version
  aligned with the cluster's Kubernetes minor per upstream guidance
  (override the image tag via `helm_values` when the cluster runs an
  older minor).

## Essential Configuration Fields

### Required

- **`spec.namespace`**: installation namespace (`kube-system` is the
  upstream convention — it keeps the pod under the system-critical
  eviction umbrella) — literal or a KubernetesNamespace reference
- **`spec.cloud`**: exactly one provider arm (see the table below)

### Common

- **`spec.chart_version`**: pinned chart version (default `9.59.0`,
  which ships autoscaler 1.35.0 — chart and app versions move
  SEPARATELY; the chart pin governs)
- **`spec.scaling.expander`**: node-group selection strategy —
  comma-separated ordered expanders from `random` (upstream default),
  `most-pods`, `least-waste` (the common production choice), `price`,
  `priority` (reads the cluster-autoscaler-priority-expander ConfigMap),
  `grpc`
- **`spec.scaling.balance_similar_node_groups`**: keep identical groups
  balanced — the multi-AZ-ASG pattern on AWS, where one group per zone
  is the norm
- **`spec.scaling.scale_down`**: master switch (upstream default true),
  `utilization_threshold` (upstream default "0.5"), `unneeded_time`
  (10m), and the three cool-downs (`delay_after_add` 10m,
  `delay_after_delete` 0s, `delay_after_failure` 3m)
- **`spec.scaling.skip_nodes_with_local_storage` /
  `skip_nodes_with_system_pods`**: both upstream-default true — set
  false deliberately, knowing emptyDir/hostPath pods lose data on
  consolidation
- **`spec.extra_args`**: flag-name → value pairs for the autoscaler's
  long tail (rendered without the leading `--`; unknown flags fail at
  pod start)
- **`spec.deployment`**: replicas (chart default 1 — extras are
  leader-elected warm standbys), resources (upstream's example starting
  point is 100m/300Mi), `priority_class_name` (chart default
  `system-cluster-critical`), node selector and tolerations — the
  autoscaler should not run on a node it may delete
- **`spec.prometheus`**: `service_monitor` (requires the Prometheus
  operator CRDs — the release fails without them) and
  `service_monitor_selector_release` (chart default
  `prometheus-operator`)
- **`spec.helm_values`**: escape hatch for chart values beyond the
  typed fields — never the primary interface

## Environment Injection

Each provider arm carries its own credential posture:

| Environment | Arm | Identity surface | Mechanism |
|---|---|---|---|
| EKS / AWS | `aws` | `irsa_role_arn` (preferred) or `access_keys` (self-managed without IRSA; the secret key is stored as a managed secret) | IRSA rides the `eks.amazonaws.com/role-arn` service-account annotation; access keys land in the chart's credential Secret and reach the pod via secretKeyRef |
| AKS / Azure VMSS | `azure` | `identity`: exactly one of `use_workload_identity`, `use_managed_identity` (+ optional `user_assigned_identity_id`), or `service_principal` | Workload identity sets the chart's ARM workload-identity extension; managed identity uses the node's (or a named user-assigned) identity; the service principal is the declared-credential fallback |
| GCE (self-managed) | `gce` | `workload_identity_service_account_email` (empty = node default credentials) | the `iam.gke.io/gcp-service-account` service-account annotation; the GSA needs compute.instanceGroups permissions and a WI binding to the autoscaler's KSA |
| Cluster API | `cluster_api` | `kubeconfig_secret` (the NAME of an in-cluster Secret) for the kubeconfig-* modes | the chart mounts the Secret and derives the kubeconfig paths from `mode`; in-cluster modes need no credential at all |
| Civo | `civo` | `api_key` (stored as a managed secret) | lands in the chart's credential Secret; reaches the pod as CIVO_* env vars via secretKeyRef |
| Sandbox | `kwok` | none — nodes are FAKE, created by the KWOK controller | the provider ConfigMap (`kwok-provider-config` chart default) carries the simulated node templates |

The cloud-side half of each keyless contract (IRSA trust policy, GCP WI
binding, Entra federated credential) is written against the chart's
derived service-account name — which is why it is an output.

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Installation namespace |
| `release_name` | Helm release name (always `cluster-autoscaler`) |
| `service_account_name` | The chart-derived service account (e.g. `cluster-autoscaler-aws-cluster-autoscaler` for the aws arm) — the subject cloud-side keyless bindings are written against |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace, field path `spec.name`).
- **Cloud-side keyless identity** closes over the
  `service_account_name` output: the IRSA trust policy, GCP WI binding,
  or Entra federated credential names the derived service account in the
  installation namespace. Note the name embeds the provider arm — the
  chart's fullname is `<release>-<cloudProvider>-<chartName>`.
- **Node groups are cloud-side composition**: the ASGs/VMSS/MIGs the
  autoscaler drives are declared by the cluster infrastructure (with the
  discovery tags on AWS); this component only points at them.

## Examples

### EKS with tag-based auto-discovery and IRSA (recommended)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesClusterAutoscaler
metadata:
  name: cluster-autoscaler
spec:
  namespace:
    value: kube-system
  aws:
    region: us-west-2
    autoDiscovery:
      clusterName: my-eks-cluster
    irsaRoleArn: arn:aws:iam::111111111111:role/cluster-autoscaler
  scaling:
    expander: least-waste
    balanceSimilarNodeGroups: true
```

### Cluster API, namespace-scoped

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesClusterAutoscaler
metadata:
  name: cluster-autoscaler
spec:
  namespace:
    value: kube-system
  clusterApi:
    mode: incluster-incluster
    namespace: cluster-workloads
    namespaceScopedRbac: true
```

### KWOK sandbox (scaling-policy evaluation, no cloud account)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesClusterAutoscaler
metadata:
  name: cluster-autoscaler
spec:
  namespace:
    value: kube-system
  kwok: {} # nodes are fake — requires the KWOK controller on the cluster
  scaling:
    expander: least-waste
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

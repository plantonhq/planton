# Kubernetes Karpenter

## When NOT to Use This

**One installation per cluster.** Karpenter owns the cluster-wide
`karpenter.sh` label domain, its CRDs, and node lifecycle — a second
controller fleet would fight the first over every NodeClaim. The Helm
release names are therefore fixed (`karpenter-crd` and `karpenter`) and
never derive from `metadata.name`.

Also not the right kind when:

- **You want to declare WHAT gets provisioned** — the fleet declarations
  are separate resources: KubernetesKarpenterNodePool describes the shape
  and constraints of the fleet, and KubernetesKarpenterEc2NodeClass
  describes the cloud-level machine template (AMIs, subnets, security
  groups, IAM). This component installs and configures the ENGINE; an
  installation without at least one NodePool provisions nothing.
- **Your capacity is organized as pre-defined node groups** — EC2 Auto
  Scaling groups, VM scale sets, Cluster API MachineDeployments. Growing
  and shrinking EXISTING groups is KubernetesClusterAutoscaler territory;
  Karpenter's value is launching right-sized machines on demand with no
  groups at all.
- **The cluster is not on AWS** — Karpenter is per-cloud upstream (each
  cloud ships its own controller, chart, and NodeClass CRD), and this
  kind installs the AWS provider, the canonical, mature implementation.
  The cloud-specific settings live in a typed oneof so future providers
  land as additive arms.

## Overview

**KubernetesKarpenter** installs the Karpenter node-provisioning
controller from the official OCI-served Helm charts
(`oci://public.ecr.aws/karpenter/karpenter`, with the CRDs from the
companion `karpenter-crd` chart). Karpenter watches for unschedulable
pods and launches RIGHT-SIZED machines for them in seconds — no
pre-created node groups, no per-ASG scaling policies: it picks instance
types from the live catalog to fit the pending pods, consolidates
under-used nodes away, and handles spot interruptions.

The typed spec covers the chart's meaningful configuration surface, with
a `helm_values` escape hatch (merged last, Helm `-f` semantics, identical
on both engines) for anything beyond it.

**Key design points:**

- **Two Helm releases, CRDs first** — the CRDs (NodePool, NodeClaim,
  EC2NodeClass, ...) install as a dedicated `karpenter-crd` release,
  upstream's supported mechanism for keeping CRDs upgradable (Helm never
  upgrades CRDs bundled inside the main chart). The controller release
  always skips its bundled CRD copies.
- **CRDs are kept on uninstall by default** — the CRD chart serves CRDs
  as ordinary templates, so without the `helm.sh/resource-policy: keep`
  annotation (rendered through the chart's `additionalAnnotations`) a
  plain uninstall cascade-deletes every NodePool/EC2NodeClass/NodeClaim
  record in the cluster.
- **Extra replicas are warm standbys, not capacity** — the chart default
  is 2 (an active leader plus a standby, spread across zones by the
  chart's topology constraints).
- **The controller never disrupts its own machine** — the chart pins
  controller pods away from Karpenter-provisioned nodes (a node affinity
  requiring `karpenter.sh/nodepool` to NOT exist), so Karpenter must run
  on capacity it does not manage (a managed node group or Fargate).
- **The install waits for real readiness** — both releases install
  atomically (600s) and wait: a ServiceMonitor rendered without the
  Prometheus operator CRDs, or a bad IRSA trust policy, fails THIS deploy
  with a readiness timeout instead of surfacing later as pods that stay
  Pending forever because no nodes ever appear.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: installation namespace (`kube-system` is
  upstream's recommended home since v1 — it keeps the controller under
  the system-critical eviction umbrella) — literal or a
  KubernetesNamespace reference
- **`spec.cluster.name`**: the cluster Karpenter provisions for — the
  value nodes register under; the chart refuses to render without it

### Common

- **`spec.cluster.eks_control_plane`**: marks the control plane as EKS,
  letting the controller discover the endpoint and CA from the
  DescribeCluster API; `endpoint` is required in practice for non-EKS
  control planes
- **`spec.chart_version`**: pinned chart version (default `1.14.0` — the
  karpenter and karpenter-crd charts version together with the
  controller, so one version pins both releases)
- **`spec.crds`**: `install` (default true, via the companion CRD chart)
  and `keep_on_uninstall` (default true — the guard rail against
  cascade-deleting every fleet declaration in the cluster)
- **`spec.aws`**: the AWS provider arm — `irsa_role_arn` (or empty for
  EKS Pod Identity), `interruption_queue` (the SQS queue for spot
  interruptions and rebalance events; empty disables interruption
  handling), `isolated_vpc`, `reserved_enis` (VPC CNI custom networking),
  `enable_zonal_shift`, `vm_memory_overhead_percent` (chart default
  `0.075`)
- **`spec.controller`**: replicas (chart default 2 — leader-elected
  standbys) and container resources (upstream leaves the choice to the
  operator; 1 CPU / 1Gi is the commonly documented starting point for
  large clusters)
- **`spec.batching`**: how long pending pods wait to be considered
  together (`max_duration` chart default 10s, `idle_duration` 1s) —
  longer windows usually produce fewer, larger nodes
- **`spec.scheduling`**: `preference_policy` (`Respect` chart default —
  soft affinities shape the node; `Ignore` can pick cheaper shapes) and
  `min_values_policy` (`Strict` chart default or `BestEffort`)
- **`spec.feature_gates`**: alpha/beta gates mirroring the chart's
  featureGates block (`reserved_capacity` is BETA and chart-default true;
  the rest — node repair, node overlay, spot-to-spot consolidation,
  static capacity, capacity buffers — are ALPHA and off)
- **`spec.controller_scheduling`**: where the controller itself runs —
  `priority_class_name` (chart default `system-cluster-critical`), node
  selector (narrows the chart's `kubernetes.io/os=linux`), tolerations
  (chart default tolerates CriticalAddonsOnly), and `host_network` for
  clusters whose CNI cannot serve pod IPs before Karpenter runs
- **`spec.prometheus.service_monitor`**: scrape discovery for the
  controller's metrics — requires the Prometheus operator CRDs (the
  release fails without them)
- **`spec.helm_values`**: escape hatch for chart values beyond the typed
  fields — never the primary interface

## Environment Injection

How the controller reaches the EC2/EKS/SQS/Pricing APIs without stored
keys:

| Environment | Spec surface | Mechanism | Where it lands |
|---|---|---|---|
| EKS with IRSA | `aws.irsa_role_arn` (an AwsIamRole reference or a literal ARN) | IAM Roles for Service Accounts via the cluster's OIDC provider | `eks.amazonaws.com/role-arn` annotation on the chart's fixed `karpenter` service account |
| EKS Pod Identity | leave `aws.irsa_role_arn` empty | the association is configured on the AWS side (namespace + `karpenter` service account) | no annotation rendered — nothing needed in the values |

The cloud-side half of either contract (trust policy or Pod Identity
association) is written against the chart's fixed service-account name,
`karpenter` — which is why it is an output.

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Installation namespace (where the controller runs) |
| `release_name` | Controller Helm release name (always `karpenter`) |
| `crd_release_name` | CRD Helm release name (`karpenter-crd`; empty when `crds.install` is false and something else manages them) |
| `service_account_name` | Always `karpenter` — the subject IRSA trust policies and EKS Pod Identity associations are written against |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace, field path `spec.name`).
- **The fleet declarations reference each other, not this engine**:
  KubernetesKarpenterNodePool references its machine template through
  `node_class_ref` (a foreign key to KubernetesKarpenterEc2NodeClass's
  `status.outputs.node_class_name`), so the chain deploys in dependency
  order — engine, then NodeClass, then NodePool. The pools and classes
  are cluster-visible custom resources the controller watches; they need
  no reference to this component, but cannot be applied before its CRDs
  exist.
- **Cloud-side keyless identity** closes over the
  `service_account_name` output: the IRSA trust policy or EKS Pod
  Identity association names the `karpenter` service account in the
  installation namespace.

## Examples

### Minimal (EKS, IRSA, upstream defaults)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKarpenter
metadata:
  name: karpenter
spec:
  namespace:
    value: kube-system
  cluster:
    name: my-eks-cluster
    eksControlPlane: true
  aws:
    irsaRoleArn:
      value: arn:aws:iam::111111111111:role/karpenter-controller
```

### Production EKS (interruption handling, sizing, telemetry)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKarpenter
metadata:
  name: karpenter
spec:
  namespace:
    value: kube-system
  cluster:
    name: my-eks-cluster
    eksControlPlane: true
  aws:
    irsaRoleArn:
      value: arn:aws:iam::111111111111:role/karpenter-controller
    interruptionQueue: my-eks-cluster-karpenter
  controller:
    replicas: 2
    resources:
      requests:
        cpu: "1"
        memory: 1Gi
      limits:
        cpu: "1"
        memory: 1Gi
  prometheus:
    serviceMonitor: true # requires the Prometheus operator CRDs
```

### Larger batches for bursty batch workloads

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKarpenter
metadata:
  name: karpenter
spec:
  namespace:
    value: kube-system
  cluster:
    name: my-eks-cluster
    eksControlPlane: true
  aws:
    irsaRoleArn:
      value: arn:aws:iam::111111111111:role/karpenter-controller
  batching:
    maxDuration: 30s
    idleDuration: 2s
  scheduling:
    preferencePolicy: Ignore # only hard requirements shape the node
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

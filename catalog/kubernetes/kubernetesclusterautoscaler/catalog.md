# Cluster Autoscaler

Installs the Kubernetes Cluster Autoscaler from the official Helm chart. The autoscaler grows and shrinks EXISTING node groups: when pods are unschedulable it raises the desired size of a matching group (an EC2 Auto Scaling group, an Azure VMSS, a Cluster API MachineDeployment, ...), and it scales groups back down when nodes sit underutilized. Exactly one provider arm per installation — AWS, Azure, GCE, Cluster API, Civo, or the KWOK simulation sandbox. One installation per cluster.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Helm Release** (`cluster-autoscaler`) -- the autoscaler Deployment (chart default 1 replica; extras leader-elect as warm standbys), RBAC, the provider-specific credential Secret when declared credentials are used, and the chart-derived service account cloud-side keyless bindings are written against
- **Namespace** (optional) -- created with standard governance labels when `createNamespace` is true; a pre-existing `kube-system` is the upstream convention

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### Kubernetes Cluster

- Node capacity organized as pre-defined groups with size bounds — ASGs tagged for auto-discovery on AWS (`k8s.io/cluster-autoscaler/enabled` + `k8s.io/cluster-autoscaler/<cluster_name>`), VMSS on Azure, MIG name prefixes on GCE, annotated MachineDeployments on Cluster API.
- Cloud-side identity for the keyless postures: an IRSA role, GCP Workload Identity binding, or Azure workload/managed identity written against the chart's derived service account.
- **On GKE/AKS this is the exception, not the rule** — both platforms ship a managed autoscaler as a node-pool toggle; installing the self-managed autoscaler there is a deliberate decision.
- With the Prometheus ServiceMonitor: the Prometheus operator CRDs — the release fails to install without them.

## Deploy

### Console

Open the deployment store, find **Cluster Autoscaler**, and click **Deploy**. The creation wizard walks you through placement, the chart pin, the provider arm with its identity posture, scaling behavior, extra flags, availability, observability, and scheduling. Start from the **EKS Autodiscovery** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesClusterAutoscaler
metadata:
  name: cluster-autoscaler
  org: acme-corp
  env: prod
spec:
  namespace:
    value: kube-system
  aws:
    region: us-west-2
    autoDiscovery:
      clusterName: my-eks-cluster
    irsaRoleArn: arn:aws:iam::111111111111:role/cluster-autoscaler
```

```shell
planton apply -f cluster-autoscaler.yaml
```

This installs the autoscaler into `kube-system` with tag-based ASG auto-discovery and keyless IRSA identity — it then manages every ASG carrying the discovery tags for the cluster, and new node groups enroll by tagging alone. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, reference the namespace instead of hardcoding it:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: autoscaler-namespace
      fieldPath: spec.name
  aws:
    region: us-west-2
    autoDiscovery:
      clusterName: my-eks-cluster
    irsaRoleArn: arn:aws:iam::111111111111:role/cluster-autoscaler
```

The InfraPipeline creates the namespace first, then installs the autoscaler into it in dependency order.

## Key Configuration

These are the most important decisions when configuring the autoscaler. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Exactly one provider arm** -- the installation serves one cloud; switching arms is a redeploy decision, not a runtime toggle. The KWOK arm is a simulation sandbox — never production.

**Auto-discovery over explicit lists** -- tag-based discovery (AWS) enrolls new node groups with no autoscaler change; explicit group lists go stale as infrastructure evolves.

**Scale-down posture** -- the scaling block's thresholds control how aggressively idle capacity is reclaimed; the untouched dials keep upstream's defaults (an unset dial ships nothing). `least-waste` is the common production expander, and balancing similar node groups keeps one-group-per-zone layouts even.

**Version alignment** -- keep the autoscaler's minor version aligned with the cluster's Kubernetes minor per upstream guidance; the chart pin is that decision.

**The flag long tail** -- `extraArgs` carries the autoscaler's long tail of flags as key/value pairs (the chart's own contract) — the tier between the typed fields and the `helmValues` YAML hatch.

**Pick one fleet controller** -- the autoscaler and Karpenter must not manage the same capacity. For AWS clusters that would rather launch right-sized machines on demand than manage groups, Karpenter is the modern alternative.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Installation namespace | Debugging and composition |
| `release_name` | Helm release name (always `cluster-autoscaler`) | Debugging the release (`helm status`) |
| `service_account_name` | Chart-derived service account (embeds the provider arm, e.g. `cluster-autoscaler-aws-cluster-autoscaler`) | The subject cloud-side keyless bindings are written against |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**EKS autodiscovery** -- tag-discovered ASGs with IRSA identity — the standard AWS posture. Start from the **EKS Autodiscovery** preset.

**Cluster API** -- annotated MachineDeployments on management or workload clusters. Start from the **Cluster API** preset.

**Azure VMSS** -- VMSS-backed node pools with workload-identity or declared-credential postures. Start from the **Azure VMSS** preset.

## Works With

- [**Karpenter**](/infra-catalog/kubernetes-karpenter) -- the alternative fleet controller for just-in-time, right-sized nodes; never both on the same capacity.
- [**Kubernetes Deployment**](/infra-catalog/kubernetes-deployment) -- unschedulable pods from any workload (Deployments, StatefulSets, Jobs) trigger scale-up; no per-workload wiring needed.
- [**Kubernetes PodDisruptionBudget**](/infra-catalog/kubernetes-pod-disruption-budget) -- shapes what the autoscaler may evict during scale-down.
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- reference it when installing outside `kube-system` so Infra Charts create the namespace first.

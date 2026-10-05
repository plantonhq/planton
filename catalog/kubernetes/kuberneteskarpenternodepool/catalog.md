# Karpenter Node Pool

Declares a Karpenter NodePool — the fleet shape Karpenter provisions machines from. A NodePool answers "what nodes may exist": the instance-type/zone/capacity-type constraints, the taints and labels new nodes carry, how long nodes live, how aggressively under-used nodes are consolidated away, and the resource ceiling the pool may reach. Clusters routinely run several — a default on-demand pool, a spot pool, a GPU pool — and Karpenter picks among them by weight. The spec holds 100% fidelity with the upstream `karpenter.sh/v1` NodePool CRD, with the CRD's own validation rules mirrored so mistakes surface at validate time.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **NodePool** (cluster-scoped, named after `metadata.name`) -- the `karpenter.sh/v1` custom resource carrying the NodeClaim template (requirements, taints, labels, lifetime), the disruption policy (consolidation and budgets), pool-wide limits, and weight

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### Kubernetes Cluster

- A Karpenter installation (the **Karpenter** component) on the cluster — the NodePool CRD does not exist before it, and a NodePool without a live controller is accepted by the API but provisions nothing.
- A NodeClass carrying the machine template — on AWS, a **Karpenter EC2 Node Class** the pool references through its node class ref (one class is typically shared by several pools).

## Deploy

### Console

Open the deployment store, find **Karpenter Node Pool**, and click **Deploy**. The creation wizard walks you through the node class binding, scheduling requirements, taints and labels, node lifetime, the disruption policy with budgets, and pool limits. Start from the **General Purpose On-Demand** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKarpenterNodePool
metadata:
  name: general-spot
  org: acme-corp
  env: prod
spec:
  template:
    nodeClassRef:
      name:
        valueFrom:
          kind: KubernetesKarpenterEc2NodeClass
          name: default-al2023
          fieldPath: status.outputs.node_class_name
    requirements:
      - key: karpenter.sh/capacity-type
        operator: In
        values: [spot]
      - key: kubernetes.io/arch
        operator: In
        values: [amd64]
  limits:
    cpu: "1000"
```

```shell
planton apply -f node-pool.yaml
```

Karpenter then launches nodes from this pool whenever pending pods fit its constraints, consolidates under-used nodes per the disruption policy, and recycles nodes at expiry. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the node class binding from the EC2NodeClass itself so the pair cannot drift apart:

```yaml
spec:
  template:
    nodeClassRef:
      name:
        valueFrom:
          kind: KubernetesKarpenterEc2NodeClass
          name: default-al2023
          fieldPath: status.outputs.node_class_name
    requirements:
      - key: karpenter.sh/capacity-type
        operator: In
        values: [on-demand]
  limits:
    cpu: "500"
```

The InfraPipeline deploys the EC2NodeClass first, then declares the pool against it.

## Key Configuration

These are the most important decisions when configuring a NodePool. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The node class is the machine, the pool is the policy** -- AMIs, subnets, security groups, and IAM live on the EC2NodeClass; the pool decides which instance shapes may launch, what they are tainted and labeled with, and when they are disrupted.

**Layer pools by purpose** -- taint dedicated pools (GPU, batch) so only tolerating pods land there; rank overlapping pools with `weight` so Karpenter prefers the cheaper or safer pool first.

**Bound spot churn** -- `minValues` instance diversity keeps spot pools resilient, and disruption budgets bound how much of the fleet can be disrupted at once — a `"0"` budget on a cron schedule freezes voluntary disruption during business hours.

**Cap the pool** -- `limits` bound the total CPU/memory the pool may reach so a runaway workload cannot provision unbounded capacity.

**Expiry recycles nodes** -- node lifetime (720h upstream default) rolls the fleet through fresh AMIs and patched hosts; shorter lifetimes trade churn for freshness.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesKarpenterEc2NodeClass** | `template.nodeClassRef.name` | `status.outputs.node_class_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `node_pool_name` | Name of the cluster-scoped NodePool | The value of the `karpenter.sh/nodepool` label on every node the pool launches — node selectors and monitoring queries |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**The first fleet** -- a general-purpose on-demand pool with sane lifetime and consolidation. Start from the **General Purpose On-Demand** preset.

**Spot with diversity** -- capacity-type spot, wide instance diversity via `minValues`, budgets bounding churn. Start from the **Spot Diversified** preset.

**Dedicated GPU pool** -- tainted so only tolerating pods land, accelerated instance families only. Start from the **GPU Dedicated** preset.

## Works With

- [**Karpenter**](/infra-catalog/kubernetes-karpenter) -- the controller that watches this pool; install it first
- [**Karpenter EC2 Node Class**](/infra-catalog/kubernetes-karpenter-ec2-node-class) -- the machine template this pool references; one class typically serves several pools
- [**Kubernetes Deployment**](/infra-catalog/kubernetes-deployment) -- pending pods trigger provisioning against the pools whose requirements they fit (the same holds for StatefulSets and Jobs)

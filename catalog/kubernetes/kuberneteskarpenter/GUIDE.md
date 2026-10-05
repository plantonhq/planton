# KubernetesKarpenter Guide

The judgment this guide carries: Karpenter vs ClusterAutoscaler is a
decision about how a cluster's capacity is ORGANIZED — and an installed
Karpenter with no NodePool provisions nothing, silently.

## Node scaling: Karpenter vs ClusterAutoscaler (the comparison home)

- **Karpenter (this kind)** watches for unschedulable pods and launches
  RIGHT-SIZED machines in seconds, picking instance types from the live
  catalog — no pre-created node groups. The modern choice for AWS
  clusters that want capacity shaped to the pending pods, with
  consolidation reclaiming under-used nodes.
- **[KubernetesClusterAutoscaler](../kubernetesclusterautoscaler/GUIDE.md)**
  grows and shrinks EXISTING node groups (ASGs, VMSS, Cluster API). The
  right choice when capacity is already organized as groups, or on
  providers without an on-demand provisioner.
- **Neither, sometimes**: GKE and AKS ship a MANAGED autoscaler as a
  toggle on the node pool (on Planton, through the cluster kinds) — that
  guide states the trap. Don't deploy a standalone autoscaler where the
  cluster kind already owns it.

Karpenter is AWS-only today (per-cloud upstream; this kind installs the
AWS provider — [reference.md](v1alpha1/reference.md) carries the scope).

## Install-then-declare, or it provisions nothing

This kind installs the ENGINE. What to provision is separate:
declare at least one
[KubernetesKarpenterNodePool](../kuberneteskarpenternodepool/GUIDE.md)
(the fleet shape) referencing a KubernetesKarpenterEc2NodeClass (the
AWS machine template). A Karpenter with no NodePool is a controller
watching for work it can never satisfy — no error, just pods stuck
Pending. A complete proposal is all three.

## Once per cluster; CRDs are the prerequisite seam

Karpenter owns the cluster-wide `karpenter.sh` domain and node lifecycle
— one fleet per cluster (release names are fixed). Its CRDs install via
the companion chart and, like every keep-on-uninstall operator, are
never auto-upgraded once present; the NodePool/EC2NodeClass kinds depend
on them exactly as the
[operator-prerequisite pattern](../../_patterns/operator-prerequisite.md)
describes. Its own dedicated namespace is the
[namespace-ownership pattern](../../_patterns/namespace-ownership.md)'s
sole-tenant case.

## On the diagram

Karpenter renders in the shared-cluster layer; NodePools reference it and
each references its EC2NodeClass — the fleet topology is visible. Pods
never draw an edge to Karpenter (it reacts to unschedulability, not
references), so the reviewer confirms a NodePool exists whenever
Karpenter is the scaler.

## Pairs well with

- KubernetesKarpenterNodePool — the fleet shape (required to provision
  anything).
- KubernetesKarpenterEc2NodeClass — the AWS machine template each pool
  references.
- KubernetesClusterAutoscaler — the alternative node-scaling lane
  (comparison above).

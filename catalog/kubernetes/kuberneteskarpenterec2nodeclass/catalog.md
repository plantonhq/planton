# Karpenter EC2 Node Class

Declares a Karpenter EC2NodeClass — the AWS-level machine template NodePools launch instances from: which AMIs to boot, which subnets and security groups to attach, which IAM identity nodes assume, disk layout, kubelet configuration, and instance metadata posture. One NodeClass is typically shared by several NodePools — the pools differ in constraints and taints; the class is the common "how a node is built". The spec holds 100% fidelity with the upstream `karpenter.k8s.aws/v1` EC2NodeClass CRD, with the CRD's own validation rules mirrored so mistakes surface at validate time.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **EC2NodeClass** (cluster-scoped, named after `metadata.name`) -- the `karpenter.k8s.aws/v1` custom resource carrying the AMI, subnet, and security-group selector terms, the node IAM identity, block device mappings, kubelet configuration, IMDS options, tags, and extra user data

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS / Cluster

- A Karpenter installation (the **Karpenter** component) on an EKS or EKS-compatible cluster — the EC2NodeClass CRD does not exist before it.
- A node IAM role (recommended — Karpenter manages the instance profile, which needs `iam:PassRole` on the controller) or a pre-existing instance profile, registered in the cluster's access configuration so nodes can join.
- Subnets and security groups selectable by the declared terms — tagging them `karpenter.sh/discovery: <cluster>` is the EKS convention.

## Deploy

### Console

Open the deployment store, find **Karpenter EC2 Node Class**, and click **Deploy**. The creation wizard walks you through AMI selection, subnet and security-group discovery, the node IAM identity, storage, kubelet configuration, metadata posture, and tags. Start from the **AL2023 Standard** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKarpenterEc2NodeClass
metadata:
  name: default-al2023
  org: acme-corp
  env: prod
spec:
  amiSelectorTerms:
    - alias: al2023@v20240807
  subnetSelectorTerms:
    - tags:
        karpenter.sh/discovery: my-eks-cluster
  securityGroupSelectorTerms:
    - tags:
        karpenter.sh/discovery: my-eks-cluster
  role: KarpenterNodeRole-my-eks-cluster
```

```shell
planton apply -f node-class.yaml
```

NodePools then reference the class through their node class ref, and every node they launch is built from this template. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring an EC2NodeClass. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Pin AMI versions in production** -- an AMI alias at `latest` drifts the fleet the moment a new AMI ships; a pinned alias version rolls only when you advance it. Custom-AMI pipelines switch the term to an SSM parameter with an explicit AMI family and advance the parameter from the pipeline.

**Selector terms are OR-of-terms, AND-within-a-term** -- each AMI/subnet/security-group term matches by id OR by tags (per the CRD's own exclusivity rules); several terms union together. The `karpenter.sh/discovery: <cluster>` tag convention keeps discovery one-tag simple.

**Role vs instance profile** -- declaring the node role lets Karpenter manage the instance profile (recommended); a pre-existing instance profile bypasses that management. Exactly one of the two.

**Harden the template as it becomes load-bearing** -- an encrypted gp3 root volume via block device mappings, kubelet reserves and eviction thresholds, and the IMDS defaults (IMDSv2 required, hop limit 1) left in place.

**User data is additive** -- extra user data merges with what Karpenter generates for the AMI family (MIME multipart on AL2, TOML on Bottlerocket); it never replaces the bootstrap.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies — the AWS selectors (AMI ids, tags, role name) are plain values validated by the CRD's own rules.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `node_class_name` | Name of the cluster-scoped EC2NodeClass | Karpenter Node Pool's node class ref — the binding every pool declares |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**AL2023 standard** -- the current-generation Amazon Linux family with discovery-tag selectors and a managed node role. Start from the **AL2023 Standard** preset.

**Bottlerocket encrypted** -- the container-optimized OS with an encrypted root volume. Start from the **Bottlerocket Encrypted** preset.

**Custom AMI pipeline** -- an SSM-parameter AMI term an image pipeline advances, with an explicit AMI family. Start from the **Custom AMI Pipeline** preset.

## Works With

- [**Karpenter**](/infra-catalog/kubernetes-karpenter) -- the controller that realizes this template; install it first
- [**Karpenter Node Pool**](/infra-catalog/kubernetes-karpenter-node-pool) -- references this class; one class typically serves a default pool, a spot pool, and a GPU pool

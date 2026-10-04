# EKS Standard

This preset installs the Karpenter controller into `kube-system` on an EKS
cluster with the two integrations every production installation carries:
an IRSA role for keyless AWS API access and an SQS interruption queue so
spot interruptions and maintenance events are drained ahead of, not
reacted to. Karpenter is one installation per cluster — it owns the
cluster-wide `karpenter.sh` label domain and node lifecycle.

This kind installs the ENGINE only: an installation without at least
one `KubernetesKarpenterNodePool` (plus its
`KubernetesKarpenterEc2NodeClass`) provisions nothing.

## When to Use

- Any EKS cluster adopting Karpenter for right-sized, on-demand node
  provisioning instead of pre-defined node groups
- The 30-second choice: this is the standard first Karpenter installation

## Key Configuration Choices

- **`namespace: kube-system` + `createNamespace: false`** — upstream's
  recommended home since v1; the namespace already exists
- **`cluster.eksControlPlane: true`** — endpoint and CA are discovered at
  startup via DescribeCluster, so only the cluster name is declared
- **IRSA (`aws.irsaRoleArn`)** — the controller calls EC2/EKS/SQS/Pricing
  without stored keys; leave the field empty instead if you use EKS Pod
  Identity (that association is configured on the AWS side)
- **Interruption queue (`aws.interruptionQueue`)** — without it,
  provisioning still works but Karpenter cannot drain ahead of an
  interruption; the controller role needs the queue permissions from
  upstream's IAM guidance
- **`crds.install: true` + `keepOnUninstall: true`** (spec defaults, made
  explicit) — CRDs stay upgradable across chart upgrades, and uninstall
  does not purge every NodePool/EC2NodeClass/NodeClaim in the cluster
- **`chartVersion: "1.14.0"`** — the spec default, pinned deliberately;
  the karpenter and karpenter-crd charts version together

## Placeholders to Replace

| Placeholder                                            | Description                                    | Where to Find                              |
| ------------------------------------------------------ | ---------------------------------------------- | ------------------------------------------ |
| `<eks-cluster-name>`                                   | EKS cluster name (node registration/discovery) | EKS console or `AwsEksCluster` outputs     |
| `arn:aws:iam::123456789012:role/karpenter-controller`  | IRSA role ARN — replace account id and name    | IAM console (role per upstream IAM guide)  |
| `karpenter-interruptions`                              | SQS interruption queue name                    | SQS console (queue per upstream IAM guide) |

## Related Presets

- **02-eks-isolated-vpc** — clusters in VPCs without internet-reachable
  AWS endpoints, and VPC CNI custom networking
- **03-ha-tuned** — explicit HA sizing, batching tuned for fewer/larger
  nodes, spot-to-spot consolidation, Prometheus telemetry

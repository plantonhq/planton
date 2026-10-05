# Kubernetes Karpenter EC2 Node Class

## When NOT to Use This

**This is the machine template — not the fleet declaration and not the
engine.** An EC2NodeClass does nothing without a Karpenter installation
(KubernetesKarpenter) on an EKS or EKS-compatible cluster, and no node
launches from it until a KubernetesKarpenterNodePool references it.

Also not the right kind when:

- **You want to declare fleet constraints** — instance types, zones,
  capacity type, taints, lifetime, and consolidation policy live on the
  NodePool. One NodeClass is typically shared by several pools: the
  pools differ in constraints and taints; the class is the common "how a
  node is built".
- **The cluster is not on AWS** — EC2NodeClass is the AWS provider's
  NodeClass CRD (`karpenter.k8s.aws/v1`); other clouds ship their own.

## Overview

**KubernetesKarpenterEc2NodeClass** declares a Karpenter EC2NodeClass —
the AWS-level machine template NodePools launch instances from: which
AMIs to boot, which subnets and security groups to attach, which IAM
identity nodes assume, disk layout, kubelet configuration, and instance
metadata posture.

The rendered EC2NodeClass is CLUSTER-SCOPED (no namespace) and named
after `metadata.name`. The spec holds 100% fidelity with the upstream
`karpenter.k8s.aws/v1` EC2NodeClass CRD at the pinned release, and the
CRD's own CEL rules are mirrored into the spec so mistakes surface at
validate time, not at apply.

**Key design points:**

- **Selector terms are ORed; fields within a term are ANDed** — the
  pattern for AMIs, subnets, security groups, and capacity
  reservations alike. The simplest AMI arm is a single alias term
  (`al2023@v20240807`); tag selection (`karpenter.sh/discovery:
  <cluster>` is the EKS convention for subnets and security groups)
  serves discovery-tagged accounts; explicit ids pin exactly.
- **Pin AMI versions in production** — `latest` drifts nodes when a new
  AMI ships (and is the only option for Windows families). The CRD's
  exclusivity rules are mirrored: an alias term must be the ONLY AMI
  term, an `id` term combines with nothing else, and `ami_family` is
  required when no alias can infer it.
- **`role` XOR `instance_profile`** — exactly one. `role` is the
  recommended arm (Karpenter manages the instance profile; needs
  `iam:PassRole` on the controller); `instance_profile` brings your own
  for accounts where Karpenter must not manage profiles.
- **IMDS defaults enforce the EKS security best practice** — IMDSv2
  required, hop limit 1 (pods cannot reach the node's instance
  credentials), IPv6 endpoint disabled. These are the CRD's own
  defaults; the modules omit unset optionals entirely so CRD-side
  defaults are never overridden by rendered zero values.

## Essential Configuration Fields

### Required

- **`spec.ami_selector_terms`**: 1–30 terms — `alias`
  (`family@version`; families al2, al2023, bottlerocket, windows2019/
  2022/2025), explicit `id`, `name`+`owner`, `ssm_parameter` (the
  custom-AMI-pipeline arm), or `tags`
- **`spec.subnet_selector_terms`**: 1–30 terms — `tags` (the standard
  `karpenter.sh/discovery` pattern) or explicit `id`; Karpenter spreads
  across the selected subnets' zones
- **`spec.security_group_selector_terms`**: 1–30 terms — `tags`, group
  `name`, or group `id`
- **`spec.role` or `spec.instance_profile`**: the node IAM identity —
  exactly one

### Common

- **`spec.ami_family`**: the UserData/bootstrap family (AL2, AL2023,
  Bottlerocket, Windows2019/2022/2025, Custom) — inferred from an alias
  term, REQUIRED when selecting AMIs by id/name/tags/ssm-parameter
- **`spec.block_device_mappings`**: disk layout (at most one root
  volume); each EBS mapping needs `snapshot_id` or `volume_size`, with
  type/IOPS/throughput/encryption knobs
- **`spec.kubelet`**: the upstream-supported kubelet subset merged into
  bootstrap user data — cluster DNS, CFS quota, hard/soft eviction
  thresholds (soft signals must pair with grace periods), image GC
  thresholds (high must exceed low), kube/system reserved, max pods,
  pods per core
- **`spec.metadata_options`**: IMDS exposure — `http_tokens` (CRD
  default `required`), `http_put_response_hop_limit` (CRD default 1),
  `http_endpoint`, `http_protocol_ipv6`
- **`spec.tags`**: applied to EC2 resources Karpenter creates; the
  controller-owned keys (`kubernetes.io/cluster/*`,
  `karpenter.sh/nodepool`, `karpenter.sh/nodeclaim`,
  `karpenter.k8s.aws/ec2nodeclass`, `eks:eks-cluster-name`) are rejected
- **`spec.user_data`**: extra bootstrap merged into Karpenter's own
  required fields — MIME multi-part for AL2/AL2023, TOML for
  Bottlerocket, PowerShell for Windows
- **Advanced**: `capacity_reservation_selector_terms` (On-Demand
  Capacity Reservations; requires the reservedCapacity feature gate, on
  by default), `network_interfaces` (multi-card / EFA topologies; a
  primary `interface` at device 0 / card 0 is required),
  `placement_group_selector` (name XOR id), `instance_store_policy`
  (`RAID0` stripes ephemeral NVMe for pod storage), `ip_prefix_count`
  (prefix delegation for pod density), `associate_public_ip_address`,
  `connection_tracking`, `cpu_options`, `detailed_monitoring`

## Outputs

| Output | Purpose |
|---|---|
| `node_class_name` | Name of the cluster-scoped EC2NodeClass (equals `metadata.name`) — the value NodePools reference through their `node_class_ref` |

## Composing in Infra Charts

- **NodePools reference this kind**, not the other way around:
  KubernetesKarpenterNodePool's `node_class_ref.name` is a foreign key
  to this kind's `status.outputs.node_class_name` — wired with
  `valueFrom`, the pool deploys after its machine template.
- **The chain deploys in dependency order**: KubernetesKarpenter (the
  engine and CRDs) → this NodeClass → the NodePools sharing it. The
  NodeClass cannot be applied before the engine's CRDs exist.
- **The node IAM role is cloud-side composition**: `role` names an IAM
  role whose instances Karpenter launches — on EKS the role must also
  appear in the cluster's access configuration (aws-auth or access
  entries) for nodes to join.

## Examples

### Minimal (AL2023 alias, discovery tags, managed profile)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKarpenterEc2NodeClass
metadata:
  name: default-al2023
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

### Hardened production (encrypted gp3 root, kubelet reserves)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKarpenterEc2NodeClass
metadata:
  name: prod-al2023
spec:
  amiSelectorTerms:
    - alias: al2023@v20240807 # pinned — latest drifts nodes
  subnetSelectorTerms:
    - tags:
        karpenter.sh/discovery: my-eks-cluster
  securityGroupSelectorTerms:
    - tags:
        karpenter.sh/discovery: my-eks-cluster
  role: KarpenterNodeRole-my-eks-cluster
  blockDeviceMappings:
    - deviceName: /dev/xvda
      rootVolume: true
      ebs:
        volumeSize: 100Gi
        volumeType: gp3
        encrypted: true
        deleteOnTermination: true
  kubelet:
    maxPods: 110
    kubeReserved:
      cpu: 200m
      memory: 512Mi
    evictionHard:
      memory.available: 5%
  metadataOptions:
    httpTokens: required
    httpPutResponseHopLimit: 1
  tags:
    team: platform
```

### Custom-AMI pipeline (SSM parameter + explicit family)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKarpenterEc2NodeClass
metadata:
  name: golden-image
spec:
  amiSelectorTerms:
    - ssmParameter: /platform/golden-ami/latest-id
  amiFamily: AL2023 # required — no alias to infer it from
  subnetSelectorTerms:
    - tags:
        karpenter.sh/discovery: my-eks-cluster
  securityGroupSelectorTerms:
    - tags:
        karpenter.sh/discovery: my-eks-cluster
  role: KarpenterNodeRole-my-eks-cluster
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

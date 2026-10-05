# AWS EKS Fargate Profile

Declares which Kubernetes pods of an EKS cluster run on AWS Fargate — serverless, per-pod compute with no EC2 nodes to size, patch, or scale. Pods whose namespace (and optionally labels) match one of the profile's selectors are scheduled onto Fargate; everything else keeps running on the cluster's node groups. The profile composes onto its neighbors by reference: the cluster attaches through an AwsEksCluster's name output, the pod execution role through an AwsIamRole's role_arn output, and the subnets through AwsSubnet subnet_id outputs.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **EKS Fargate Profile** -- the profile on the target cluster, with its pod selectors, private subnets, and pod execution role
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied to the profile

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **EKS Cluster** -- the target cluster, ideally a Planton AwsEksCluster referenced by its `name` output so deploys order correctly.
- **Pod Execution Role** -- an AwsIamRole trusting `eks-fargate-pods.amazonaws.com` and carrying `AmazonEKSFargatePodExecutionRolePolicy`, referenced by its `role_arn` output.
- **Private Subnets** -- AwsSubnet resources with no direct internet-gateway route, referenced by their `subnet_id` outputs.

### AWS Account

- **EKS permissions** -- the credentials used by the Provider Connection must have `eks:CreateFargateProfile`, `eks:DescribeFargateProfile`, and `eks:DeleteFargateProfile`, plus `iam:PassRole` on the pod execution role.
- **NAT egress** -- Fargate pods have no public IPs; give them outbound internet through a NAT gateway on the private subnets' route tables.
- **CoreDNS on Fargate** -- a Fargate-only cluster needs the CoreDNS add-on's pods matched by a profile (the `kube-system` namespace) so DNS itself can schedule.

## Deploy

### Console

Open the deployment store, find **AWS EKS Fargate Profile**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Namespace on Fargate** preset in the [Presets](#presets) tab to run one namespace's pods on Fargate.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEksFargateProfile
metadata:
  name: batch-workloads
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  clusterName:
    valueFrom:
      kind: AwsEksCluster
      name: platform-cluster
      fieldPath: status.outputs.name
  podExecutionRoleArn:
    valueFrom:
      kind: AwsIamRole
      name: fargate-pod-execution
      fieldPath: status.outputs.role_arn
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: private-subnet-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: private-subnet-b
        fieldPath: status.outputs.subnet_id
  selectors:
    - namespace: batch
```

```shell
planton apply -f fargate-profile.yaml
```

This schedules every pod in the `batch` namespace onto Fargate — no nodes to provision, patch, or scale for those workloads. An Infra Job tracks the provisioning in real time.

### InfraChart

When the profile deploys alongside its cluster, role, and subnets in one chart, wire the references via ValueFromRef:

```yaml
spec:
  region: us-west-2
  clusterName:
    valueFrom:
      kind: AwsEksCluster
      name: platform-cluster
      fieldPath: status.outputs.name
  podExecutionRoleArn:
    valueFrom:
      kind: AwsIamRole
      name: fargate-pod-execution
      fieldPath: status.outputs.role_arn
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: private-subnet-a
        fieldPath: status.outputs.subnet_id
  selectors:
    - namespace: batch
```

The InfraPipeline resolves the dependency graph, deploys the cluster, role, and subnets first, then creates the profile.

## Key Configuration

These are the most important decisions when configuring a Fargate profile. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Everything is create-time immutable** -- name, cluster, role, subnets, and selectors are all fixed at creation. Changing anything replaces the profile, and pods keep running through the replacement window only if a second matching profile covers them. AWS also serializes profile operations: one profile per cluster creates or deletes at a time.

**Pod selectors** -- a pod runs on Fargate when it matches ANY selector: the selector's namespace (wildcards allowed — `*` any sequence, `?` any character), plus EVERY label when labels are given (AND semantics within a selector). At most 5 selectors per profile, 5 label pairs per selector. Selectors are the whole scheduling contract — nothing else moves pods onto Fargate.

**Private subnets only** -- AWS rejects subnets whose route table carries an internet-gateway route. Fargate pods get no public IPs; outbound internet flows through the subnets' NAT gateway.

**Pod execution role** -- the role Fargate uses to pull images and write logs for the matched pods. It must trust `eks-fargate-pods.amazonaws.com` and carry `AmazonEKSFargatePodExecutionRolePolicy` — attach both on the AwsIamRole itself; this component never modifies a role it merely references.

**Fargate vs node groups** -- Fargate removes node management for spiky, batch, or strongly-isolated workloads; node groups stay cheaper for steady high-utilization fleets and support DaemonSets, GPUs, and privileged pods (which Fargate does not).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsEksCluster** | `clusterName` | `status.outputs.name` |
| **AwsIamRole** | `podExecutionRoleArn` | `status.outputs.role_arn` |
| **AwsSubnet** | `subnetIds[]` | `status.outputs.subnet_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `fargate_profile_arn` | Amazon Resource Name of the profile | Auditing and support tooling |
| `fargate_profile_name` | The profile's name | Cross-referencing with `aws eks describe-fargate-profile` |
| `status` | The profile's state after provisioning (`ACTIVE` on success) | Deployment verification |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**One namespace on Fargate** -- Everything in a namespace (batch jobs, a team's sandbox) runs serverless. Start from the **Namespace on Fargate** preset.

**Label-selected workloads** -- Only pods carrying a label (e.g. `compute: fargate`) inside a namespace move to Fargate, letting the same namespace mix node-group and Fargate scheduling. Start from the **Label-Scoped Fargate** preset.

## Works With

- [**AWS EKS Cluster**](/infra-catalog/aws-eks-cluster) -- the cluster the profile attaches to, referenced by `clusterName`.
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- the pod execution role, referenced by `podExecutionRoleArn`.
- [**AWS Subnet**](/infra-catalog/aws-subnet) -- the private subnets the pods launch into, referenced by `subnetIds`.
- [**AWS EKS Addon**](/infra-catalog/aws-eks-addon) -- CoreDNS as a managed add-on pairs with a `kube-system` selector on Fargate-only clusters.

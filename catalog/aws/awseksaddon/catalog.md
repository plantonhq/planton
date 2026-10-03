# AWS EKS Addon

Installs an EKS managed add-on — cluster software like vpc-cni, CoreDNS, kube-proxy, or the EBS/EFS CSI drivers whose install, upgrades, configuration, and removal AWS manages through the EKS control plane instead of Helm charts you operate yourself. The add-on composes onto its neighbors by reference: the cluster attaches through an AwsEksCluster's name output, and IAM identity is wired to referenced AwsIamRole resources through EKS Pod Identity or IRSA.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **EKS Add-on** -- the managed add-on on the target cluster, at the pinned version or AWS's default for the cluster's Kubernetes version, with the chosen conflict-resolution posture and configuration values
- **Pod Identity Associations** -- one EKS Pod Identity association per entry in `podIdentityAssociations`, binding the add-on's Kubernetes service accounts to referenced IAM roles
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied to the add-on

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **EKS Cluster** -- the target cluster, ideally a Planton AwsEksCluster referenced by its `name` output so deploys order correctly.

### AWS Account

- **EKS permissions** -- the credentials used by the Provider Connection must have `eks:CreateAddon`, `eks:DescribeAddon`, `eks:UpdateAddon`, `eks:DeleteAddon`, and `eks:CreatePodIdentityAssociation` when Pod Identity is used.
- **Pod Identity prerequisites** -- the `eks-pod-identity-agent` add-on installed on the cluster, and each referenced role trusting `pods.eks.amazonaws.com`.
- **IRSA prerequisites** -- an IAM OIDC provider created on the cluster's `oidc_issuer_url` output; without it AWS rejects an install that sets `serviceAccountRoleArn`.

## Deploy

### Console

Open the deployment store, find **AWS EKS Addon**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Core Networking Add-on** preset in the [Presets](#presets) tab to adopt a bootstrapped cluster's networking add-ons.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEksAddon
metadata:
  name: platform-coredns
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  clusterName:
    valueFrom:
      kind: AwsEksCluster
      name: platform-cluster
      fieldPath: status.outputs.name
  addonName: coredns
  resolveConflictsOnCreate: OVERWRITE
```

```shell
planton apply -f eks-addon.yaml
```

This installs CoreDNS as a managed add-on at AWS's default version for the cluster's Kubernetes version, adopting the self-managed CoreDNS the cluster bootstrapped. An Infra Job tracks the provisioning in real time.

### InfraChart

When the add-on deploys alongside its cluster and IAM role in one chart, wire the references via ValueFromRef:

```yaml
spec:
  region: us-west-2
  clusterName:
    valueFrom:
      kind: AwsEksCluster
      name: platform-cluster
      fieldPath: status.outputs.name
  addonName: aws-ebs-csi-driver
  podIdentityAssociations:
    - serviceAccount: ebs-csi-controller-sa
      roleArn:
        valueFrom:
          kind: AwsIamRole
          name: ebs-csi-role
          fieldPath: status.outputs.role_arn
```

The InfraPipeline resolves the dependency graph, deploys the cluster and role first, then installs the add-on with the resolved values.

## Key Configuration

These are the most important decisions when configuring an EKS add-on. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Add-on identity** -- AWS keys the add-on on `(cluster, addonName)`: one add-on name per cluster, and both halves plus `region` are create-time immutable. AWS-built add-ons use short names (`vpc-cni`, `coredns`, `kube-proxy`, `aws-ebs-csi-driver`, `eks-pod-identity-agent`); marketplace add-ons use their vendor-prefixed catalog name. `aws eks describe-addon-versions` lists everything the cluster's Kubernetes version supports.

**Version** -- Empty `addonVersion` installs the AWS default for the cluster's Kubernetes version, the never-goes-stale choice. Pin a full version like `v1.18.1-eksbuild.3` for byte-identical clusters and controlled upgrades. Version changes update in place -- the add-on rolls its own pods.

**Conflict resolution** -- `resolveConflictsOnCreate: OVERWRITE` adopts the self-managed copies of vpc-cni/coredns/kube-proxy that bootstrapped clusters already run (the standard migration path); `NONE` fails the install on conflict. `resolveConflictsOnUpdate` decides what happens when an update finds out-of-band edits: `OVERWRITE` restores managed values, `PRESERVE` keeps hand-made changes, empty keeps the AWS default (fail loudly).

**IAM identity** -- With nothing configured, the add-on's pods use the node role's permissions. Add-ons that call AWS APIs (the EBS/EFS CSI drivers, CloudWatch observability) need their own role: prefer `podIdentityAssociations` (EKS Pod Identity -- per-service-account roles, no OIDC provider needed) on new clusters; `serviceAccountRoleArn` (IRSA) remains for clusters already wired for it. Both reference AwsIamRole resources -- this component never modifies a role it references.

**Configuration values** -- A single JSON document validated against the add-on's own published schema at install time (`aws eks describe-addon-configuration` shows what is configurable). Empty keeps every add-on default.

**Lifecycle** -- `preserve: true` keeps the add-on's Kubernetes resources running (self-managed) when this resource is deleted -- the no-outage handover. `namespaceConfig.namespace` installs into a custom namespace; it is create-only and supported by only some add-ons.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsEksCluster** | `clusterName` | `status.outputs.name` |
| **AwsIamRole** | `serviceAccountRoleArn`, `podIdentityAssociations[].roleArn` | `status.outputs.role_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `addon_arn` | Amazon Resource Name of the add-on | IAM policies, support tooling |
| `addon_name` | Resolved EKS catalog name | Auditing and reporting |
| `addon_version` | The version actually running | Answering "what is live?" on unpinned installs |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Adopt core networking** -- Install vpc-cni, coredns, and kube-proxy as managed add-ons with `OVERWRITE` on create, taking over the self-managed copies from cluster bootstrap. Start from the **Core Networking Add-on** preset.

**EBS CSI with Pod Identity** -- The EBS CSI driver bound to a dedicated IAM role through Pod Identity, unlocking dynamic EBS volume provisioning. Start from the **EBS CSI Driver with Pod Identity** preset.

**Pinned version fleet** -- Pin `addonVersion` across environments for byte-identical clusters and staged upgrades. Start from the **Pinned Version with Configuration** preset.

## Works With

- [**AWS EKS Cluster**](/infra-catalog/aws-eks-cluster) -- the parent cluster, referenced by `clusterName`.
- [**AWS EKS Node Group**](/infra-catalog/aws-eks-node-group) -- add-ons schedule onto the cluster's nodes; storage drivers serve the workloads those nodes run.
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- the IAM identities referenced by Pod Identity associations and IRSA.

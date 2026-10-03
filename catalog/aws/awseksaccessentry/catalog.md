# AWS EKS Access Entry

Grants one IAM principal (a role or user) access to an EKS cluster's Kubernetes API through EKS access entries — the modern access model that replaces hand-editing the aws-auth ConfigMap. Authorization comes from either side, or both: AWS-managed access policies attached with cluster or namespace scope, and Kubernetes group mappings your own RBAC bindings reference. The entry composes onto its neighbors by reference: the cluster attaches through an AwsEksCluster's name output and the principal through an AwsIamRole's role_arn output.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **EKS Access Entry** -- the (cluster, principal) grant, with the entry type, Kubernetes group mappings, and optional username
- **Access Policy Associations** -- one association per entry in `policyAssociations`, attaching an AWS-managed EKS access policy to the principal with cluster-wide or namespace scope
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied to the entry

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **EKS Cluster** -- the target cluster, ideally a Planton AwsEksCluster referenced by its `name` output so deploys order correctly.
- **IAM Principal** -- the role (or user) being granted access, ideally a Planton AwsIamRole referenced by its `role_arn` output.

### AWS Account

- **API authentication enabled on the cluster** -- the cluster's `accessConfig.authenticationMode` must be `API` or `API_AND_CONFIG_MAP`; a CONFIG_MAP-only cluster rejects access entries.
- **EKS permissions** -- the credentials used by the Provider Connection must have `eks:CreateAccessEntry`, `eks:DescribeAccessEntry`, `eks:UpdateAccessEntry`, `eks:DeleteAccessEntry`, and `eks:AssociateAccessPolicy` / `eks:DisassociateAccessPolicy` when policy associations are used.

## Deploy

### Console

Open the deployment store, find **AWS EKS Access Entry**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Cluster Viewer** preset in the [Presets](#presets) tab for the safe read-only default grant.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEksAccessEntry
metadata:
  name: platform-viewers
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  clusterName:
    valueFrom:
      kind: AwsEksCluster
      name: platform-cluster
      fieldPath: status.outputs.name
  principalArn:
    valueFrom:
      kind: AwsIamRole
      name: platform-readonly
      fieldPath: status.outputs.role_arn
  policyAssociations:
    - policyArn: arn:aws:eks::aws:cluster-access-policy/AmazonEKSViewPolicy
      accessScope:
        type: cluster
```

```shell
planton apply -f access-entry.yaml
```

This grants the referenced role read-only access across the whole cluster through the AWS-managed view policy — no in-cluster RBAC objects, no ConfigMap edits. An Infra Job tracks the provisioning in real time.

### InfraChart

When the entry deploys alongside its cluster and role in one chart, wire the references via ValueFromRef:

```yaml
spec:
  region: us-west-2
  clusterName:
    valueFrom:
      kind: AwsEksCluster
      name: platform-cluster
      fieldPath: status.outputs.name
  principalArn:
    valueFrom:
      kind: AwsIamRole
      name: platform-readonly
      fieldPath: status.outputs.role_arn
```

The InfraPipeline resolves the dependency graph, deploys the cluster and role first, then creates the access entry.

## Key Configuration

These are the most important decisions when configuring an EKS access entry. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Entry identity** -- AWS keys the entry on `(cluster, principal)`: one entry per principal per cluster, and both keys plus `region` are create-time immutable. Groups, username, and policy associations all update in place — the grant's *scope* can evolve without replacing the grant itself.

**Entry type** -- Empty or `STANDARD` is the human/workload entry this component exists for. The node types (`EC2`, `EC2_LINUX`, `EC2_WINDOWS`, `FARGATE_LINUX`, `HYBRID_LINUX`) exist for infrastructure roles — EKS creates them automatically for managed node groups and Fargate profiles, so set one only when registering self-managed or hybrid nodes. AWS forbids groups, username, and policy associations on non-STANDARD entries.

**Access policies vs RBAC groups** -- `policyAssociations` attach AWS-managed EKS access policies (`AmazonEKSViewPolicy`, `...EditPolicy`, `...AdminPolicy`, `...ClusterAdminPolicy`), scoped to the whole cluster or to namespaces — authorization without any in-cluster RBAC objects. `kubernetesGroups` maps the principal onto group names your OWN (Cluster)RoleBindings reference — you own the RBAC objects inside the cluster. Use policies for the standard grants, groups for fine-grained custom authorization, or both.

**Namespace scoping** -- a `namespace`-scoped association bounds the policy to listed namespaces (trailing `*` wildcards allowed, e.g. `team-*`) — give a team `AmazonEKSAdminPolicy` inside its own namespaces without cluster-wide admin. A `cluster`-scoped association applies everywhere and must not list namespaces.

**Username** -- Empty lets AWS default the Kubernetes username (the principal ARN for users; a session-templated name for roles — the right choice for audit trails, since it preserves the session name). Set it only when RBAC bindings need a fixed username.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsEksCluster** | `clusterName` | `status.outputs.name` |
| **AwsIamRole** | `principalArn` | `status.outputs.role_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `access_entry_arn` | Amazon Resource Name of the entry | Auditing and support tooling |
| `principal_arn` | The IAM principal as resolved at provisioning time | Verifying which identity the grant landed on |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Cluster viewer** -- Read-only across the whole cluster through `AmazonEKSViewPolicy` — the default grant for engineers who inspect but do not operate. Start from the **Cluster Viewer** preset.

**Namespace admin** -- Full admin inside a team's namespaces only, through a namespace-scoped `AmazonEKSAdminPolicy`. Start from the **Namespace Admin** preset.

**Bring-your-own RBAC** -- Map the principal onto Kubernetes groups your own RoleBindings reference, with no AWS-managed policies at all. Start from the **RBAC Group Mapping** preset.

## Works With

- [**AWS EKS Cluster**](/infra-catalog/aws-eks-cluster) -- the cluster whose Kubernetes API the entry grants access to, referenced by `clusterName`. The cluster's authentication mode must include API.
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- the principal being granted access, referenced by `principalArn`.
- [**AWS EKS Node Group**](/infra-catalog/aws-eks-node-group) and [**AWS EKS Fargate Profile**](/infra-catalog/aws-eks-fargate-profile) -- EKS auto-creates node-type entries for their infrastructure roles; model entries only for the principals you grant yourself.

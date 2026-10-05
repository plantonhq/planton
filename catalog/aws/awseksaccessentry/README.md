# Overview

The AwsEksAccessEntry API resource grants one IAM principal (a role or
user) access to an `AwsEksCluster`'s Kubernetes API through EKS access
entries -- the modern access model that replaces hand-editing the
aws-auth ConfigMap.

## Why We Created This API Resource

Cluster access is per-principal and changes on its own schedule --
teams onboard, CI roles rotate, break-glass roles come and go. Modeling
each grant as a first-class node:

- **Makes access auditable in the graph**: every principal with cluster
  access is a visible resource referencing the cluster and the
  `AwsIamRole` it empowers -- no hidden ConfigMap entries.
- **Composes with IAM**: the principal attaches by reference to the
  role's `role_arn` output; the same role node that carries policies
  shows where it can reach Kubernetes.
- **Keeps authorization honest**: AWS-managed access policies
  (cluster- or namespace-scoped) and/or your own RBAC group mappings,
  each modeled exactly as AWS defines them.

## Key Features

### Two Authorization Paths

- **Access policies** (`policyAssociations`): AWS-managed policies
  (AmazonEKSViewPolicy, ...EditPolicy, ...AdminPolicy,
  ...ClusterAdminPolicy) scoped to the cluster or to namespaces -- no
  in-cluster RBAC objects needed.
- **RBAC groups** (`kubernetesGroups`): map the principal onto group
  names your own (Cluster)RoleBindings reference; reserved `system:`
  groups are rejected at validation.

### Honest Constraints

- **One entry per principal per cluster** (the AWS key), both
  create-time immutable.
- **Node types modeled truthfully**: EC2/FARGATE/HYBRID entry types
  exist for self-managed node registration and forbid
  groups/username/associations -- enforced in CEL, mirroring AWS's
  runtime rules.
- **Scope shape enforced**: namespace-scoped associations must list
  namespaces; cluster-scoped ones must not.

### Folded Associations

Policy associations are AWS sub-resources of exactly this (cluster,
principal) pair, so they live in the spec -- while both engines still
materialize each one as its own provider resource keyed by policy name,
so adding, re-scoping, or removing one diffs in place.

## Benefits

- **Composability**: cluster and principal attach through `valueFrom`
  references.
- **Auditability**: the environment graph shows who can reach which
  cluster, with what scope.
- **Consistency**: identical behavior across Terraform and Pulumi.

## Outputs

- `access_entry_arn`: the entry's ARN
- `principal_arn`: the IAM principal granted access, as resolved at
  provisioning time

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

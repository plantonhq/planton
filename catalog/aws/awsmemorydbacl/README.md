# Overview

The AwsMemorydbAcl API resource creates a MemoryDB Access Control List —
the set of users a MemoryDB cluster authenticates against. An ACL is the
single attachment point between identities and clusters: users
(AwsMemorydbUser) are grouped into an ACL, and a cluster attaches exactly
one ACL.

## Why We Created This API Resource

Granting an application access to a MemoryDB cluster should be a membership
edit, not a cluster change:

- **The attachment unit**: clusters reference an ACL by name; the ACL
  references users. Access is granted or revoked in one place — in place,
  with no cluster or user replacement.
- **Composable membership**: `userNames` is a list of references to
  AwsMemorydbUser outputs, so the resource graph shows exactly which
  identities reach which clusters.
- **Environment-shaped**: one ACL is typically shared by every cluster in
  an environment that trusts the same set of application identities.

## Key Features

- **In-place membership**: adding or removing a user diffs the set on
  update — granting one application access never disturbs the others.
- **Empty ACLs are valid**: unlike ElastiCache user groups (which require a
  "default" member), MemoryDB has no mandatory member. A cluster attached
  to an empty ACL simply accepts no authenticated connections.
- **Engine-version awareness**: the ACL exports the minimum engine version
  its combined user set requires.

## Benefits

- **Least-privilege by construction**: pair per-application users (scoped
  access strings) with one environment ACL instead of sharing "open-access".
- **Honest modeling**: the built-in "open-access" ACL always exists in the
  account and is referenced by literal value on the cluster — it is never
  modeled as a resource.
- **Consistency**: identical behavior across Terraform and Pulumi.

## Outputs

- `acl_name`: what clusters attach via their `aclName` (same as
  `metadata.name`)
- `acl_arn`: the ACL's ARN for IAM policies
- `minimum_engine_version`: the minimum engine version the ACL's user set
  requires

## Deliberately Skipped (with reasons)

- **A standalone membership/association kind**: AWS has no such resource —
  membership is a property of the ACL — and per-edge glue kinds are never
  modeled; membership folds into `userNames` as references.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

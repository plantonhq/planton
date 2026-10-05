# AWS MemoryDB ACL

Deploys a MemoryDB Access Control List — the single attachment point between identities and clusters in MemoryDB's only authentication model. Users ([AwsMemorydbUser](/infra-catalog/aws-memorydb-user)) join the ACL, and a cluster ([AwsMemorydbCluster](/infra-catalog/aws-memorydb-cluster)) attaches exactly one ACL. Granting or revoking an application's database access is an in-place membership edit here — the cluster and the users themselves never change.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **MemoryDB ACL** -- one access control list whose AWS name is the resource name (create-time immutable, max 40 characters), carrying the user set it grants access to; membership edits apply in place
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Member users (optional at create)** -- deploy [AwsMemorydbUser](/infra-catalog/aws-memorydb-user) resources first to reference them in membership; an empty ACL is valid and can be populated later.

### AWS Account

- **A name within AWS's cap** -- the resource name IS the ACL name; AWS rejects names longer than 40 characters at create time.
- **Same-region members** -- an ACL can only contain users from its own region, and only a cluster in its region can attach it.

## Deploy

### Console

Open the deployment store, find **AWS MemoryDB ACL**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Environment ACL** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsMemorydbAcl
metadata:
  name: prod-services
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  userNames:
    - valueFrom:
        kind: AwsMemorydbUser
        name: orders-service
        fieldPath: status.outputs.user_name
```

```shell
planton apply -f memorydb-acl.yaml
```

This creates an ACL holding the orders-service user, ready for a cluster to attach. An Infra Job tracks the provisioning in real time.

### InfraChart

When the ACL deploys alongside its member users in one chart, wire the membership via ValueFromRef:

```yaml
spec:
  region: us-west-2
  userNames:
    - valueFrom:
        kind: AwsMemorydbUser
        name: orders-service
        fieldPath: status.outputs.user_name
    - valueFrom:
        kind: AwsMemorydbUser
        name: payments-service
        fieldPath: status.outputs.user_name
```

The InfraPipeline resolves the dependency graph, deploys the users first, then this ACL -- and the cluster that attaches it references the `acl_name` output afterwards.

## Key Configuration

These are the most important decisions when configuring a MemoryDB ACL. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One name, one attachment** -- the resource name IS the AWS ACL name (create-time immutable): what a cluster's ACL field references. A cluster attaches exactly one ACL, so the ACL is the whole access story for that cluster.

**Membership is the lever** -- every operational access change is a membership edit: onboarding a service adds its user, offboarding removes it. Edits apply in place — no user or cluster replacement, ever. Reference platform-managed users by their `user_name` output; users created outside the platform join by literal name. Deleting a user removes it from every ACL on the AWS side automatically, and the provider reconciles that — dropping an already-deleted member from the list is a clean no-op, never an error.

**Empty is legal, open-access is built in** -- unlike ElastiCache user groups (which demand a member named "default"), a MemoryDB ACL may be empty: it deploys but accepts no authenticated connections — the deploy-then-populate shape. For deliberately unauthenticated access, clusters reference AWS's built-in `open-access` ACL by name; it always exists and is never modeled as a resource.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsMemorydbUser** (optional) | `userNames` | `status.outputs.user_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `acl_name` | The AWS ACL name | AwsMemorydbCluster attachment (`aclName`) |
| `acl_arn` | Amazon Resource Name of the ACL | IAM policies scoping ACL management |
| `minimum_engine_version` | The engine floor the combined member set requires | Verifying the attaching cluster's engine version |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Environment-scoped ACL** -- one ACL per environment (e.g. prod-services) holding each application's user; new services onboard with a membership edit. Start from the **Environment ACL** preset.

**Deploy-then-populate** -- create the ACL empty alongside the cluster, then add users as applications come online — every step in place, no replacements.

## Works With

- [**AWS MemoryDB User**](/infra-catalog/aws-memorydb-user) -- the identities this ACL collects (references `user_name`)
- [**AWS MemoryDB Cluster**](/infra-catalog/aws-memorydb-cluster) -- the durable database that attaches this ACL via `acl_name`

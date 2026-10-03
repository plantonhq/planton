# AWS MemoryDB User

Deploys a MemoryDB user — one identity in MemoryDB's Access Control List (ACL) authentication system. MemoryDB has exactly one authentication model: every cluster attaches an ACL, and an ACL is a set of users. If an application should reach a MemoryDB cluster with credentials, a user is how that identity exists. Each user carries a Redis ACL access string scoping which commands and keys it may touch, so per-application least-privilege access is the natural shape: one user per application, grouped into ACLs ([AwsMemorydbAcl](/infra-catalog/aws-memorydb-acl)), with the ACL attached to the cluster. Password material lives in managed secrets referenced from the spec — never in the manifest.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **MemoryDB User** -- one ACL identity whose user name is the resource name (create-time immutable, unique per region, max 40 characters), carrying its Redis `ACL SETUSER` access string (scoping keys and command categories; tightening it later applies in place) and exactly one authentication mode: password (1–2 secrets, enabling zero-downtime rotation) or IAM-signed tokens
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Managed secrets for passwords** -- when using password authentication, create the password as an org secret first; the spec carries a `$secret/<slug>` reference and the runner resolves it just-in-time at deploy. Each password's value must be 16–128 characters.

### AWS Account

- **A name within AWS's cap** -- the resource name IS the user name; AWS rejects names longer than 40 characters at create time.
- **For IAM authentication** -- the attached cluster must run transit encryption (TLS — MemoryDB's default), and the client's IAM principal needs `memorydb:Connect` on both the user ARN and the cluster ARN.

## Deploy

### Console

Open the deployment store, find **AWS MemoryDB User**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Password-Authenticated Application User** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsMemorydbUser
metadata:
  name: orders-service
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  accessString: "on ~orders:* +@read +@write"
  authenticationMode:
    type: password
    passwords:
      - $secret/orders-db-password
```

```shell
planton apply -f memorydb-user.yaml
```

This creates a password-authenticated user scoped to the `orders:` key prefix with read/write command categories. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a MemoryDB user. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One name, one identity** -- unlike ElastiCache (where a user has a separate id and a reusable AUTH name), MemoryDB has a single flat user name that IS the identity: the resource name. Clients present it in the AUTH command, ACLs reference it in membership, it must be unique in the region, and it is create-time immutable — renaming means replacement.

**The access string** -- the same syntax as the Redis `ACL SETUSER` rule list: the on/off switch enables or disables the user, key patterns (`~orders:*`) grant matching keys, and command categories (`+@read`, `-@dangerous`) grant or subtract command groups. Updates apply in place on new connections. Scope production applications to their key prefix; reserve `~* +@all` for admin identities.

**Authentication** -- exactly one mechanism, and every user authenticates: MemoryDB has no credential-less user type (unauthenticated access exists only through the built-in `open-access` ACL, never through a user). `password` works everywhere (1–2 secret references; two enable rotation: add the new password, roll clients, remove the old). `iam` replaces long-lived secrets with short-lived tokens signed by the workload's AWS identity — it requires a TLS-enabled cluster. The whole authentication mode edits in place.

**Passwords are write-only at AWS** -- the API never returns them, so a password changed outside the manifest is undetectable drift, and a user adopted by import carries no password state. The manifest is the source of truth: re-assert the passwords there to manage them, and rotate through the two-entry overlap rather than by out-of-band edits.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies. It is a leaf the access-control graph builds on: in a chart, users deploy first, then the ACL that collects them (referencing this user's `user_name` output), then the cluster that attaches the ACL.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `user_name` | The identity clients present in AUTH | AwsMemorydbAcl membership (`userNames`) |
| `user_arn` | Amazon Resource Name of the user | IAM policies — `memorydb:Connect` for IAM-authenticated clients |
| `minimum_engine_version` | The engine floor this user's ACL feature set requires | Verifying the attaching cluster's engine version |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Password-authenticated application user** -- one user per service, scoped to its own key prefix, authenticating with a managed-secret password. Start from the **Password-Authenticated Application User** preset.

**IAM-authenticated user** -- workloads on ECS/EKS/Lambda with an IAM role skip password rotation entirely: short-lived signed tokens, no secret material anywhere. Start from the **IAM-Authenticated Application User** preset.

## Works With

- [**AWS MemoryDB ACL**](/infra-catalog/aws-memorydb-acl) -- collects users into the attachment unit clusters consume (references `user_name`)
- [**AWS MemoryDB Cluster**](/infra-catalog/aws-memorydb-cluster) -- the durable database the user ultimately authenticates against, via its attached ACL

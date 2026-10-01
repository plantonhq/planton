# DigitalOcean Database User

Creates an additional user on a DigitalOcean managed database cluster, with the MySQL authentication plugin choice and per-topic Kafka / per-index OpenSearch access-control lists. DigitalOcean generates the password (and Kafka mTLS certificate pair) server-side; they surface as secret stack outputs for application wiring and never appear in the manifest. One user per service is the shape that keeps credential rotation and revocation independent -- the built-in `doadmin` user working everywhere is exactly why production should not use it.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Database User** -- a named user on the referenced cluster, with a server-generated password
- **MySQL Auth Plugin** -- configured only when `mysqlAuthPlugin` is set; chooses between DigitalOcean's modern default (`caching_sha2_password`) and the legacy plugin for old clients
- **Kafka / OpenSearch ACLs** -- configured only when `settings` is set; grants per-topic or per-index permissions from closed permission lists. `settings` is set by engine: PostgreSQL users declare `settings: {}` (DigitalOcean stores a settings object for every PostgreSQL user), MySQL users leave it out (the API refuses settings updates on MySQL)

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **A DigitalOceanDatabaseCluster** -- the owning cluster, referenced by name (or an existing cluster's UUID as a literal).

### DigitalOcean Account

- Nothing beyond the cluster: users are a free feature of managed databases.

## Deploy

### Console

Open the deployment store, find **DigitalOcean Database User**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Per-Service Application User** preset in the [Presets](#presets) tab to give one service its own credential.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseUser
metadata:
  name: orders-service-user
  org: acme-corp
  env: prod
spec:
  cluster:
    valueFrom:
      kind: DigitalOceanDatabaseCluster
      name: orders-postgres
      fieldPath: status.outputs.cluster_id
  userName: orders-service
```

```shell
planton apply -f do-database-user.yaml
```

This creates a user named `orders-service` on the referenced cluster with a server-generated password exported as a secret output. A Stack Job tracks the provisioning in real time.

### InfraChart

When the user deploys alongside its cluster in one chart, wire the cluster reference via ValueFromRef:

```yaml
spec:
  cluster:
    valueFrom:
      kind: DigitalOceanDatabaseCluster
      name: orders-postgres
      fieldPath: status.outputs.cluster_id
  userName: orders-service
```

The InfraPipeline resolves the dependency graph, deploys the cluster first, then creates the user on it.

## Key Configuration

These are the most important decisions when configuring a database user. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One user per service, always** -- The cluster's built-in `doadmin` user works everywhere, which is exactly why production should not use it: one leaked credential exposes everything, and rotation breaks every consumer at once. Create one user per service; each gets its own server-generated password that rotates (by replacing the user) and revokes (by deleting it) independently.

**Passwords rotate by replacement** -- DigitalOcean generates passwords; there is no set-password surface here. To rotate a credential, create a replacement user (new name), move the service to it, then delete the old user -- deletion revokes access immediately. Changing `userName` or `cluster` replaces the user, and the replacement gets a NEW server-generated password.

**ACLs: the manifest is the source of truth** -- DigitalOcean returns Kafka and OpenSearch ACLs only in the create response; reads never include them, so the console will never show you what a user may do. Review permission changes in code review, not in the console. ACL edits apply in place (no replacement), topics and indexes accept wildcard patterns (`events-*`), and the API rejects Kafka ACLs on non-Kafka clusters (and OpenSearch ACLs on non-OpenSearch clusters) at request time.

**`mysqlAuthPlugin`: leave it unset** -- Unset means DigitalOcean's `caching_sha2_password`, the modern plugin. Set `mysql_native_password` only for clients too old to speak it, and treat that as a dated compatibility decision to revisit. It applies only to MySQL clusters (rejected on other engines at request time), and clearing it later resets the user to the modern default in place, preserving the password.

**Serialized operations on busy clusters** -- User creates and deletes serialize per cluster (an API constraint). A chart creating ten users on one cluster deploys them one at a time -- expect linear time, not a hang.

**MongoDB quirk** -- MongoDB clusters return the user's password only in the create response. The outputs still carry it (captured at create), but an imported MongoDB user has no recoverable password -- rotate by replacement instead.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **DigitalOceanDatabaseCluster** | `cluster` | `status.outputs.cluster_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `password` | The server-generated password (secret) -- it never appears in the manifest | Application authentication wired as a secret |
| `access_cert` / `access_key` | PEM certificate pair for mutual TLS (secrets); Kafka clusters only, empty elsewhere | Kafka client mTLS configuration |
| `role` | The role DigitalOcean assigned (normally `normal`; the cluster's built-in default user is `primary`) | Verifying least-privilege posture |

`cluster_id` and `user_name` are also echoed for addressing and verification -- DigitalOcean has no standalone user id; the (cluster, name) pair IS the identity.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Per-service application user** -- one dedicated user per service so credentials rotate and revoke independently, replacing shared `doadmin` use in application configs; compose it with a DigitalOceanDatabaseDb of the same service. Start from the **Per-Service Application User** preset.

**Least-privilege Kafka producer** -- a Kafka user scoped to producing on a topic-prefix wildcard with produce-and-consume rights on its dead-letter topic, authenticating with the issued mTLS pair instead of the cluster's admin default. Start from the **Kafka Producer with Topic ACLs** preset.

## Works With

- [**DigitalOcean Database Cluster**](/cloud-catalog/digital-ocean-database-cluster) -- the owning cluster, wired via the `cluster` reference
- [**DigitalOcean Logical Database**](/cloud-catalog/digital-ocean-database-db) -- the per-service database this user's credentials typically connect to
- [**DigitalOcean Database Connection Pool**](/cloud-catalog/digital-ocean-database-connection-pool) -- a dedicated pool authenticates as this user via its `user` field

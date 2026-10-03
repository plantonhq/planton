# GCP AlloyDB User

Deploys a database user (`google_alloydb_user`) on an AlloyDB cluster. Users are first-class nodes: one per application with ALLOYDB_BUILT_IN (password) or ALLOYDB_IAM_USER (passwordless IAM database authentication).

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **AlloyDB API enablement** (`alloydb.googleapis.com`) on the target project (never disabled on destroy)
- **AlloyDB User** — a database user on the cluster with the chosen authentication type and the declared PostgreSQL roles (e.g. `alloydbiamuser`) granted

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** — an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GcpAlloydbCluster** — the cluster this user lives on. Reference its `status.outputs.cluster_id`, or pass the full cluster resource path directly.
- **Password secret** — for ALLOYDB_BUILT_IN, store the password as an org secret reference (`$secret/slug`) or inline for dev only.

## Deploy

### Console

Open the deployment store, find **GCP AlloyDB User**, and click **Deploy**. Start from the **Application User (ALLOYDB_BUILT_IN)** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpAlloydbUser
metadata:
  name: orders-app
  org: acme-corp
  env: prod
spec:
  cluster:
    value: "projects/acme-prod-12345/locations/us-central1/clusters/orders-db"
  userId: orders-app
  password: $secret/orders-app-db-password
  databaseRoles:
    - alloydbiamuser
```

```shell
planton apply -f alloydb-user.yaml
```

This creates a password-authenticated application user on the cluster with the `alloydbiamuser` role, the password resolved from the org secret at deploy time. An Infra Job tracks the provisioning in real time.

### InfraChart

Wire the user to a cluster deployed in the same InfraPipeline:

```yaml
spec:
  cluster:
    valueFrom:
      kind: GcpAlloydbCluster
      name: orders-db
      fieldPath: status.outputs.cluster_id
  password: $secret/orders-app-db-password
```

## Key Configuration

These are the most important decisions when configuring an AlloyDB user. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Authentication type** — `ALLOYDB_BUILT_IN` (default) uses a password; updating the password rotates it in place. `ALLOYDB_IAM_USER` enables passwordless IAM database authentication via AlloyDB Auth Proxy or Language Connectors — the spec rejects a password on an IAM user, and the type is immutable: switching between the two means a new user.

**User ID** — Immutable after creation, and it is the PostgreSQL role name applications authenticate as. One user per application service is the recommended pattern — it makes credential rotation and revocation per-service instead of shared.

**Database roles** — Grant only the PostgreSQL roles the application needs. `alloydbiamuser` is the standard application role; `alloydbsuperuser` grants cluster-wide power and belongs on break-glass users only.

**Destroy behavior** — the default destroy drops the user from the cluster, but objects it owns inside PostgreSQL keep their ownership rows — reassign ownership first or the drop fails. `deletionPolicy: PREVENT` protects a credential applications still authenticate with; `ABANDON` leaves the user existing (and authenticating) on the cluster.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpAlloydbCluster** | `cluster` | `status.outputs.cluster_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | Fully qualified user resource name (`projects/{p}/locations/{l}/clusters/{c}/users/{u}`) | Management API calls, monitoring |
| `user_id` | The database role name as stored by AlloyDB | Application connection strings (the username apps authenticate as) |

`cluster_id` is also exported but only echoes the cluster the spec already names.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Application user** — password-authenticated built-in user with `alloydbiamuser` role. Start from **Application User**.

**IAM user** — passwordless IAM-authenticated user for services connecting via Auth Proxy. Start from **IAM User**.

## Works With

- [**GCP AlloyDB Cluster**](/infra-catalog/gcp-alloydb-cluster) — parent cluster this user lives on
- [**GCP AlloyDB Instance**](/infra-catalog/gcp-alloydb-instance) — read pools on the same cluster
- [**GCP Project**](/infra-catalog/gcp-project) — optional project override

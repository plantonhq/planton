# GCP AlloyDB User

Deploys a database user (`google_alloydb_user`) on an AlloyDB cluster. Users are first-class nodes: one per application with its own credential (`ALLOYDB_BUILT_IN`) or passwordless IAM authentication (`ALLOYDB_IAM_USER`).

## What Gets Created

- **AlloyDB API enablement** — `alloydb.googleapis.com` on the target project
- **AlloyDB user** — a `google_alloydb_user` on the referenced cluster

## Prerequisites

- **An existing AlloyDB cluster** — [GcpAlloydbCluster](/docs/catalog/gcp/gcpalloydbcluster) or literal cluster path
- **GCP credentials** — [`iac/permissions.yaml`](iac/permissions.yaml) lists the exact least-privilege permissions

## Quick Start

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpAlloydbUser
metadata:
  name: orders-app-user
spec:
  cluster:
    valueFrom:
      kind: GcpAlloydbCluster
      name: my-orders-cluster
      fieldPath: status.outputs.cluster_id
  userId: orders-app
  password: a-strong-generated-password
```

## Configuration Reference

| Field | Default | Description |
|-------|---------|-------------|
| `cluster` | — (required) | Cluster ref → `cluster_id`. Immutable. |
| `userId` | — (required) | Database role name. Immutable. |
| `userType` | `ALLOYDB_BUILT_IN` | `ALLOYDB_BUILT_IN` or `ALLOYDB_IAM_USER`. |
| `password` | — | BUILT_IN only; mutable (rotates in place). Secret. |
| `databaseRoles` | — | Roles granted to the user. |

## Outputs

| Output | Description |
|--------|-------------|
| `name` | Fully qualified user resource path |
| `user_id` | User ID as stored by AlloyDB |
| `cluster_id` | Cluster resource path |

## Related Kinds

- [GcpAlloydbCluster](/docs/catalog/gcp/gcpalloydbcluster)
- [GcpAlloydbInstance](/docs/catalog/gcp/gcpalloydbinstance)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

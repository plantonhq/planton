# Scaleway MongoDB Instance

Deploys a Scaleway Managed MongoDB instance with an admin user, optional additional users with role-based access control, Private Network attachment, and automated snapshot scheduling. Supports standalone (single-node) and replica set (three-node) topologies with block storage volumes and TLS certificates.

## What Gets Created

When you deploy a ScalewayMongodbInstance resource, OpenMCF provisions:

- **MongoDB Instance** — a `mongodb.Instance` resource providing a fully managed MongoDB engine with the specified node type, volume configuration, admin user, and TLS certificate
- **Private Network Endpoint** — created only when `privateNetworkId` is set, attaches the instance to a Private Network with IPAM-based IP assignment
- **Public Network Endpoint** — created when no Private Network is set, or when both `privateNetworkId` and `enablePublicNetwork` are set
- **Database Users** — one `mongodb.User` resource per entry in the `users` list, each with inline role assignments scoped to specific databases or all databases

## Prerequisites

- **Scaleway credentials** configured via environment variables or OpenMCF provider config
- **A supported MongoDB version** in semantic version format (e.g., `"7.0.12"`)
- **A Private Network** in the `fr-par` region if using private connectivity (can be created via a ScalewayPrivateNetwork resource)
- **Region availability** — Scaleway Managed MongoDB is currently only available in `fr-par` (Paris)

## Quick Start

Create a file `mongodb-instance.yaml`:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayMongodbInstance
metadata:
  name: my-mongo
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.ScalewayMongodbInstance.my-mongo
spec:
  region: fr-par
  version: "7.0.12"
  nodeType: MGDB-PLAY2-NANO
  nodeNumber: 1
  adminUser: admin
  adminPassword: change-me-strong-pw
```

Deploy:

```shell
openmcf apply -f mongodb-instance.yaml
```

This creates a single-node MongoDB 7.0 instance with default block storage (sbs_5k), no Private Network (public endpoint only), and no additional users beyond the admin account.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Scaleway region for the instance. Currently only `"fr-par"` is supported. Cannot be changed after creation. | Required |
| `version` | `string` | MongoDB engine version in semantic version format (e.g., `"7.0.12"`). Scaleway normalizes to major.minor internally. Cannot be changed after creation. | Required, pattern: `^[0-9]+\.[0-9]+\.[0-9]+$` |
| `nodeType` | `string` | Instance type determining CPU and RAM. Shared: `"MGDB-PLAY2-NANO"`, `"MGDB-PRO2-XXS"` through `"MGDB-PRO2-L"`. Dedicated: `"MGDB-POP2-2C-8G"` through `"MGDB-POP2-64C-256G"`. Can be changed after creation. | Required |
| `nodeNumber` | `uint32` | Number of nodes. `1` for standalone, `3` for replica set (automatic failover). No other values are valid. Changing between 1 and 3 may destroy and recreate the instance. | Required, must be 1 or 3 |
| `adminUser` | `string` | Username for the initial admin user created with the instance. Must differ from any user in the `users` list. | Required, max 63 characters |
| `adminPassword` | `string` | Password for the admin user. | Required, min 8 characters |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `privateNetworkId` | `StringValueOrRef` | — | Private Network UUID for private connectivity. Enables IPAM-based IP assignment. Can reference a ScalewayPrivateNetwork resource via `valueFrom`. |
| `enablePublicNetwork` | `bool` | `false` | When `true` and `privateNetworkId` is set, creates a public endpoint in addition to the private one. Has no effect when `privateNetworkId` is not set. No IP-based ACL is available for MongoDB. |
| `volumeType` | `string` | `"sbs_5k"` | Block storage volume type. Options: `"sbs_5k"` (5K IOPS) or `"sbs_15k"` (15K IOPS). Cannot be changed after creation. |
| `volumeSizeInGb` | `uint32` | `5` | Volume size in GB. Must be a multiple of 5, minimum 5. Can only be increased, never decreased. |
| `enableSnapshotSchedule` | `bool` | `false` | When `true`, enables automatic periodic snapshots. |
| `snapshotScheduleFrequencyHours` | `uint32` | — | Hours between automatic snapshots. Only used when `enableSnapshotSchedule` is `true`. |
| `snapshotScheduleRetentionDays` | `uint32` | — | Days to retain automatic snapshots. Only used when `enableSnapshotSchedule` is `true`. |
| `users` | `object[]` | `[]` | Additional database users to create on the instance. |
| `users[].name` | `string` | — | Username. Required per entry. Max 63 characters. |
| `users[].password` | `string` | — | User password. Required per entry. Min 8 characters. |
| `users[].roles` | `object[]` | `[]` | Role assignments for this user. If empty, the user exists but has no database access. |
| `users[].roles[].role` | `string` | — | Permission level. Options: `"read"`, `"read_write"`, `"db_admin"`. Required per role. |
| `users[].roles[].databaseName` | `string` | — | Specific database to scope this role to. Mutually exclusive with `anyDatabase`. |
| `users[].roles[].anyDatabase` | `bool` | `false` | When `true`, applies the role to all databases. Mutually exclusive with `databaseName`. |
| `settings` | `map<string, string>` | `{}` | MongoDB-specific engine configuration settings. Applied on creation and updates. |

## Examples

### Development Standalone

A minimal single-node MongoDB instance for development and testing:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayMongodbInstance
metadata:
  name: dev-mongo
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.ScalewayMongodbInstance.dev-mongo
spec:
  region: fr-par
  version: "7.0.12"
  nodeType: MGDB-PLAY2-NANO
  nodeNumber: 1
  adminUser: admin
  adminPassword: dev-admin-pw-2024
  users:
    - name: appuser
      password: app-user-pw-2024
      roles:
        - role: read_write
          databaseName: myapp
```

### Production Replica Set with Private Network

A three-node replica set with Private Network connectivity, high-IOPS storage, automated snapshots, and multiple users with scoped roles:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayMongodbInstance
metadata:
  name: prod-mongo
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.ScalewayMongodbInstance.prod-mongo
spec:
  region: fr-par
  version: "7.0.12"
  nodeType: MGDB-POP2-8C-32G
  nodeNumber: 3
  adminUser: mongoadmin
  adminPassword: strong-prod-password-2024
  volumeType: sbs_15k
  volumeSizeInGb: 100
  enableSnapshotSchedule: true
  snapshotScheduleFrequencyHours: 6
  snapshotScheduleRetentionDays: 30
  privateNetworkId: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  users:
    - name: webapp_svc
      password: webapp-svc-pw-2024
      roles:
        - role: read_write
          databaseName: webapp
    - name: analytics_ro
      password: analytics-ro-pw-2024
      roles:
        - role: read
          databaseName: webapp
        - role: read
          databaseName: analytics
    - name: dba_tools
      password: dba-tools-pw-2024
      roles:
        - role: db_admin
          anyDatabase: true
```

### Private Network Reference with Public Endpoint

A MongoDB instance referencing an OpenMCF-managed Private Network while also exposing a public endpoint for admin access:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayMongodbInstance
metadata:
  name: staging-mongo
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.ScalewayMongodbInstance.staging-mongo
spec:
  region: fr-par
  version: "7.0.12"
  nodeType: MGDB-PRO2-XXS
  nodeNumber: 1
  adminUser: admin
  adminPassword: staging-admin-pw-2024
  privateNetworkId:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: app-network
      fieldPath: status.outputs.private_network_id
  enablePublicNetwork: true
  volumeSizeInGb: 20
  users:
    - name: app_svc
      password: app-svc-pw-2024
      roles:
        - role: read_write
          databaseName: staging_db
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `instance_id` | `string` | Regional ID of the created MongoDB instance (e.g., `"fr-par/xxxxxxxx-..."`). Referenced by downstream resources (snapshots, monitoring, management automation). |
| `public_dns_record` | `string` | Public endpoint DNS hostname (e.g., `"{id}.mgdb.{region}.scw.cloud"`). Empty if the instance is private-only. |
| `public_port` | `uint32` | Public endpoint TCP port (typically 27017). Zero if the instance is private-only. |
| `private_dns_records` | `string[]` | Private Network endpoint DNS hostnames. Empty if no Private Network is attached. |
| `private_ips` | `string[]` | Private Network endpoint IPv4 addresses assigned via IPAM. Empty if no Private Network is attached. |
| `private_port` | `uint32` | Private Network endpoint TCP port (typically 27017). Zero if no Private Network is attached. |
| `tls_certificate` | `string` | TLS CA certificate in PEM format for verifying the database server. Use with the `tlsCAFile` MongoDB driver option or `--tlsCAFile` mongo shell flag. Always available. |

## Related Components

- [ScalewayPrivateNetwork](/docs/catalog/scaleway/scalewayprivatenetwork) — provides private connectivity between the database and application workloads
- [ScalewayKapsuleCluster](/docs/catalog/scaleway/scalewaykapsulecluster) — deploys Kubernetes clusters whose workloads connect to this database
- [ScalewayInstance](/docs/catalog/scaleway/scalewayinstance) — deploys compute instances that can connect to the database over a shared Private Network
- [ScalewayRdbInstance](/docs/catalog/scaleway/scalewayrdbinstance) — alternative managed database component for PostgreSQL and MySQL workloads

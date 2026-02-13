# Civo Database

Deploys a managed MySQL or PostgreSQL database instance on Civo Cloud with configurable replicas, private network attachment, and firewall integration. The component supports up to four read replicas and exposes connection details as stack outputs for wiring to application workloads.

## What Gets Created

When you deploy a CivoDatabase resource, OpenMCF provisions:

- **Managed Database Instance** — a `civo_database` resource running the specified engine (MySQL or PostgreSQL) at the chosen version and size, attached to the target private network
- **Replica Nodes** — created only when `replicas` is greater than 0, adding read replicas to the database cluster (total nodes = 1 primary + replica count)
- **Firewall Attachment** — created only when `firewallIds` is provided, attaches a firewall to control network access to the database instance

## Prerequisites

- **Civo credentials** configured via environment variables or OpenMCF provider config
- **An existing Civo network** in the target region for private connectivity (can be created with CivoVpc)
- **A Civo firewall** if restricting access to the database (can be created with CivoFirewall)

## Quick Start

Create a file `civo-db.yaml`:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoDatabase
metadata:
  name: my-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CivoDatabase.my-db
spec:
  dbInstanceName: my-db
  engine: postgres
  engineVersion: "14"
  region: nyc1
  sizeSlug: g3.db.small
  networkId:
    value: network-uuid-here
```

Deploy:

```shell
openmcf apply -f civo-db.yaml
```

This creates a single-node PostgreSQL 14 instance on a `g3.db.small` plan in New York, attached to the specified private network.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `dbInstanceName` | `string` | Human-readable name for the database instance. Must be unique within the region. | Required, max 64 characters |
| `engine` | `enum` | Database engine. Valid values: `mysql`, `postgres`. | Required |
| `engineVersion` | `string` | Engine version (e.g., `8.0` for MySQL, `14` for PostgreSQL). Major version only or major.minor. | Required, pattern: `^[0-9]+(\.[0-9]+)?$` |
| `region` | `enum` | Civo region where the database is created. Valid values: `lon1`, `lon2`, `fra1`, `nyc1`, `phx1`, `mum1`. | Required |
| `sizeSlug` | `string` | Plan identifier defining CPU, memory, and base storage (e.g., `g3.db.small`). | Required |
| `networkId` | `StringValueOrRef` | Private network ID for the database instance. Can reference a CivoVpc resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `replicas` | `uint32` | `0` | Number of read replicas. Use 0, 2, or 4 for a total of 1, 3, or 5 nodes. Maximum 4 replicas. |
| `firewallIds` | `StringValueOrRef[]` | `[]` | Firewall IDs to attach for access control. Can reference CivoFirewall resources via `valueFrom`. Currently only the first firewall ID is applied. |
| `storageGib` | `uint32` | — | Custom storage size in GiB, overriding the default provided by `sizeSlug`. |
| `tags` | `string[]` | `[]` | Tags for organizational purposes within Civo. |

## Examples

### Single-Node MySQL Instance

A minimal MySQL database for development:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoDatabase
metadata:
  name: dev-mysql
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CivoDatabase.dev-mysql
spec:
  dbInstanceName: dev-mysql
  engine: mysql
  engineVersion: "8.0"
  region: fra1
  sizeSlug: g3.db.small
  networkId:
    value: network-uuid-here
```

### PostgreSQL with Replicas and Firewall

A PostgreSQL cluster with two read replicas and firewall-controlled access:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoDatabase
metadata:
  name: prod-postgres
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CivoDatabase.prod-postgres
spec:
  dbInstanceName: prod-postgres
  engine: postgres
  engineVersion: "15"
  region: nyc1
  sizeSlug: g3.db.medium
  replicas: 2
  networkId:
    value: network-uuid-here
  firewallIds:
    - value: firewall-uuid-here
  tags:
    - environment:production
    - team:backend
```

### Using Foreign Key References

Reference OpenMCF-managed CivoVpc and CivoFirewall resources instead of hardcoding IDs:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoDatabase
metadata:
  name: ref-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CivoDatabase.ref-db
spec:
  dbInstanceName: ref-db
  engine: postgres
  engineVersion: "15"
  region: lon1
  sizeSlug: g3.db.large
  replicas: 4
  networkId:
    valueFrom:
      kind: CivoVpc
      name: my-network
      field: status.outputs.network_id
  firewallIds:
    - valueFrom:
        kind: CivoFirewall
        name: db-firewall
        field: status.outputs.firewall_id
  tags:
    - environment:production
    - managed-by:openmcf
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `database_id` | `string` | Unique identifier (UUID) of the created database instance |
| `host` | `string` | Hostname or IP address of the primary database endpoint |
| `port` | `uint32` | Network port on which the database is listening |
| `username` | `string` | Username for the default database user |
| `password_secret_ref` | `string` | Reference to the secret containing the default user's password |
| `replica_endpoints` | `string[]` | Host addresses of replica nodes; empty if no replicas were configured |

## Related Components

- [CivoVpc](/docs/catalog/civo/civovpc) — provides the private network for database connectivity
- [CivoFirewall](/docs/catalog/civo/civofirewall) — controls network access to the database instance
- [CivoKubernetesCluster](/docs/catalog/civo/civokubernetescluster) — application workloads that connect to the database
- [CivoDnsRecord](/docs/catalog/civo/civodnsrecord) — creates DNS records pointing to the database endpoint

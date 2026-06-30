# DigitalOcean Database Cluster

Deploys a managed database cluster on DigitalOcean supporting PostgreSQL, MySQL, Redis, and MongoDB engines. The component handles node sizing, version selection, optional VPC placement, and custom storage configuration, exposing connection details as stack outputs.

## What Gets Created

When you deploy a DigitalOceanDatabaseCluster resource, Planton provisions:

- **Database Cluster** — a `digitalocean_database_cluster` resource with the specified engine, version, region, node size, and node count
- **VPC Attachment** — created only when `vpc` is specified, places the cluster in a private network for secure access
- **Custom Storage** — configured only when `storageGib` is specified, overrides the default storage for the chosen node size

## Prerequisites

- **DigitalOcean credentials** configured via environment variables or Planton provider config
- **A DigitalOcean VPC** in the target region if using private networking (can reference a DigitalOceanVpc resource via `valueFrom`)
- **A supported engine version** for the chosen database engine (e.g., `16` for PostgreSQL 16, `8` for MySQL 8)

## Quick Start

Create a file `database.yaml`:

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanDatabaseCluster
metadata:
  name: my-database
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.DigitalOceanDatabaseCluster.my-database
spec:
  clusterName: my-database
  engine: pg
  engineVersion: "16"
  region: nyc3
  sizeSlug: db-s-1vcpu-2gb
  nodeCount: 1
```

Deploy:

```shell
planton apply -f database.yaml
```

This creates a single-node PostgreSQL 16 cluster in the NYC3 region with the `db-s-1vcpu-2gb` Droplet size.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `clusterName` | `string` | Name of the database cluster in DigitalOcean. | Required, max 64 characters |
| `engine` | `enum` | Database engine. Valid values: `pg` (PostgreSQL), `mysql`, `redis`, `mongodb`. | Required |
| `engineVersion` | `string` | Major version of the database engine (e.g., `16` for PostgreSQL, `8` for MySQL). | Required, pattern: `^[0-9]+(\.[0-9]+)?$` |
| `region` | `enum` | DigitalOcean region for the cluster. Valid values: `nyc3`, `sfo3`, `fra1`, `sgp1`, `lon1`, `tor1`, `blr1`, `ams3`. | Required |
| `sizeSlug` | `string` | Droplet size slug for cluster nodes (e.g., `db-s-2vcpu-4gb`). Determines CPU and memory per node. | Required |
| `nodeCount` | `uint32` | Number of nodes in the cluster. | Required, 1–3 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `vpc` | `StringValueOrRef` | — | VPC UUID for private network placement. Can reference a DigitalOceanVpc resource via `valueFrom`. |
| `storageGib` | `uint32` | — | Custom storage size in GiB. If not set, uses the default storage for the chosen `sizeSlug`. |
| `enablePublicConnectivity` | `bool` | `false` | When `true`, enables public network access to the cluster. When `false`, the cluster is accessible only via VPC or DigitalOcean internal network. |

## Examples

### MySQL Cluster in VPC

A single-node MySQL 8 cluster placed in a private network:

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanDatabaseCluster
metadata:
  name: mysql-app-db
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.DigitalOceanDatabaseCluster.mysql-app-db
spec:
  clusterName: mysql-app-db
  engine: mysql
  engineVersion: "8"
  region: fra1
  sizeSlug: db-s-2vcpu-4gb
  nodeCount: 1
  vpc:
    value: "vpc-fra1-uuid"
```

### HA PostgreSQL with Custom Storage

A three-node PostgreSQL cluster with increased storage for production workloads:

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanDatabaseCluster
metadata:
  name: prod-postgres
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.DigitalOceanDatabaseCluster.prod-postgres
spec:
  clusterName: prod-postgres
  engine: pg
  engineVersion: "16"
  region: nyc3
  sizeSlug: db-s-4vcpu-8gb
  nodeCount: 3
  storageGib: 100
  vpc:
    value: "vpc-prod-uuid"
```

### Redis Cache with VPC Reference

A Redis cluster using a foreign key reference to an Planton-managed VPC:

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanDatabaseCluster
metadata:
  name: cache-redis
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.DigitalOceanDatabaseCluster.cache-redis
spec:
  clusterName: cache-redis
  engine: redis
  engineVersion: "7"
  region: sgp1
  sizeSlug: db-s-2vcpu-4gb
  nodeCount: 2
  vpc:
    valueFrom:
      kind: DigitalOceanVpc
      name: prod-vpc
      field: status.outputs.vpc_id
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `cluster_id` | `string` | UUID of the created database cluster |
| `connection_uri` | `string` | Full connection URI including credentials and database name |
| `host` | `string` | Hostname or IP address of the database cluster |
| `port` | `uint32` | Network port the database cluster listens on |
| `database_user` | `string` | Username for the cluster's default database user |
| `database_password` | `string` | Password for the cluster's default database user |

## Related Components

- [DigitalOceanVpc](/docs/catalog/digitalocean/digitaloceanvpc) — provides the VPC for private network placement
- [DigitalOceanKubernetesCluster](/docs/catalog/digitalocean/digitaloceankubernetescluster) — commonly co-deployed to run applications that consume the database
- [DigitalOceanFirewall](/docs/catalog/digitalocean/digitaloceanfirewall) — controls network access to the database cluster

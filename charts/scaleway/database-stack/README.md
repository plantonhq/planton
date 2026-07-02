# Scaleway Database Stack

The **Scaleway Database Stack** InfraChart provisions managed database instances on
Scaleway Cloud -- PostgreSQL/MySQL (RDB), Redis, and MongoDB -- all connected to a
shared Private Network for secure, private-only access.

It supports **conditional resource generation via Jinja templates** -- you choose which
databases to deploy and whether to create a new VPC or reuse an existing Private Network
from another chart (e.g., Kapsule Environment).

Chart resources and configuration parameters are defined in the [`templates`](templates)
directory and documented in [`values.yaml`](values.yaml).

---

## Architecture

All database instances connect to the same Private Network, enabling private-only
communication. Applications on the same network (e.g., Kapsule pods, serverless
containers) can reach the databases without public internet exposure.

```
ScalewayVpc (conditional)
  └── ScalewayPrivateNetwork (conditional)
        ├── ScalewayRdbInstance (optional, PostgreSQL/MySQL)
        ├── ScalewayRedisCluster (optional, Redis cache)
        └── ScalewayMongodbInstance (optional, MongoDB)
```

### Network Sharing Pattern

When `create_network` is `true` (default), the chart creates its own VPC and Private
Network. When `false`, it reuses an existing Private Network via `private_network_id`.

**Common scenario:** Deploy Kapsule Environment first (creates VPC + Private Network),
then deploy Database Stack with `create_network: false` and pass the Kapsule
environment's `private_network_id`. This places databases on the same network as your
Kubernetes cluster, enabling pods to connect to databases over private IPs.

---

## Included Cloud Resources (conditional)

| Resource | Always created | Controlled by boolean flag |
|----------|----------------|----------------------------|
| **Scaleway VPC** | No | `create_network` |
| **Scaleway Private Network** | No | `create_network` |
| **Scaleway RDB Instance** (PostgreSQL/MySQL) | No | `rdb_enabled` |
| **Scaleway Redis Cluster** | No | `redis_enabled` |
| **Scaleway MongoDB Instance** | No | `mongodb_enabled` |

### How the `create_network` flag works

* `create_network: true` (default) -- Creates a dedicated VPC and Private Network for
  the database tier. All databases join this network. Best for isolated database
  environments or when no other infrastructure exists yet.
* `create_network: false` -- Uses the `private_network_id` parameter to connect databases
  to an existing Private Network. Best for adding databases to an existing infrastructure
  stack (e.g., alongside a Kapsule cluster).

---

## Chart Input Values

### Network

| Parameter | Description | Example / Options | Required / Default |
|-----------|-------------|-------------------|-------------------|
| **create_network** | Create new VPC and Private Network | `true` / `false` | Default `true` |
| **region** | Scaleway region | `fr-par`, `nl-ams`, `pl-waw` | Default `fr-par` |
| **stack_name** | Name prefix for resources | `db-stack` | Required |
| **private_network_id** | Existing PN ID | UUID | Required if `create_network=false` |

### RDB Instance (PostgreSQL/MySQL)

| Parameter | Description | Example / Options | Required / Default |
|-----------|-------------|-------------------|-------------------|
| **rdb_enabled** | Create an RDB instance | `true` / `false` | Default `true` |
| **rdb_engine** | Engine and version | `PostgreSQL-16`, `MySQL-8` | Default `PostgreSQL-16` |
| **rdb_node_type** | Node type | `DB-DEV-S`, `db-gp-xs`, `db-gp-s` | Default `DB-DEV-S` |
| **rdb_admin_user** | Admin username | `admin` | Required |
| **rdb_admin_password** | Admin password (min 8 chars) | -- | Required |
| **rdb_is_ha** | Enable high availability | `true` / `false` | Default `false` |
| **rdb_volume_type** | Volume type | `lssd`, `bssd`, `sbs_15k` | Default `lssd` |
| **rdb_volume_size_in_gb** | Volume size (0 = default) | `10`, `50`, `100` | Default `0` |
| **rdb_disable_backup** | Disable automated backups | `true` / `false` | Default `false` |

### Redis Cluster

| Parameter | Description | Example / Options | Required / Default |
|-----------|-------------|-------------------|-------------------|
| **redis_enabled** | Create a Redis cluster | `true` / `false` | Default `false` |
| **redis_zone** | Zone (Redis is zonal) | `fr-par-1`, `nl-ams-1` | Default `fr-par-1` |
| **redis_version** | Redis version | `7.2.5`, `6.2.7` | Default `7.2.5` |
| **redis_node_type** | Node type | `RED1-MICRO`, `RED1-S`, `RED1-M` | Default `RED1-MICRO` |
| **redis_cluster_size** | Nodes (1=standalone, 2=HA, 3+=cluster) | `1`, `2`, `3` | Default `1` |
| **redis_user_name** | Redis username | `default` | Required |
| **redis_password** | Redis password (min 8 chars) | -- | Required |
| **redis_tls_enabled** | Enable TLS encryption | `true` / `false` | Default `false` |

### MongoDB Instance

| Parameter | Description | Example / Options | Required / Default |
|-----------|-------------|-------------------|-------------------|
| **mongodb_enabled** | Create a MongoDB instance | `true` / `false` | Default `false` |
| **mongodb_version** | MongoDB version | `7.0.12` | Default `7.0.12` |
| **mongodb_node_type** | Node type | `MGDB-PLAY2-NANO`, `MGDB-PRO2-XXS` | Default `MGDB-PLAY2-NANO` |
| **mongodb_node_number** | Nodes (1=standalone, 3=replica set) | `1`, `3` | Default `1` |
| **mongodb_admin_user** | Admin username | `admin` | Required |
| **mongodb_admin_password** | Admin password (min 8 chars) | -- | Required |
| **mongodb_volume_type** | Volume type | `sbs_5k`, `sbs_15k` | Default `sbs_5k` |
| **mongodb_volume_size_in_gb** | Volume size (0 = default 5 GB) | `10`, `50` | Default `0` |

---

## Customization & Management

* Enable only the databases you need -- toggle `rdb_enabled`, `redis_enabled`, and
  `mongodb_enabled` independently.
* Use `create_network: false` with `private_network_id` to share the network with a
  Kapsule cluster, enabling pods to connect to databases over private IPs.
* Databases, users, privileges, and ACL rules are managed individually per resource after
  the stack is deployed -- the chart creates instances with admin credentials only.
* Resource references (`valueFrom` vs `value`) are automatically wired by the templates --
  no manual edits needed.

---

## Important Notes

* **Redis is zonal** -- Redis clusters are placed in a specific zone (e.g., `fr-par-1`),
  while RDB and MongoDB are regional. Ensure `redis_zone` is in the same region as the
  Private Network.
* **MongoDB region availability** -- As of writing, Scaleway Managed MongoDB is only
  available in `fr-par`. Check Scaleway documentation for updated region availability.
* **ACL and Private Network are mutually exclusive for Redis** -- When Redis is attached
  to a Private Network (the chart's default behavior), ACL rules cannot be configured.
  This is a Scaleway API constraint enforced at apply time.
* **MongoDB has no IP-based ACL** -- Unlike RDB, MongoDB has no network access control.
  When attached to a Private Network, private-only is the most secure configuration.
  Enable `enable_public_network` on the individual resource only when explicitly needed.
* **HA doubles cost** -- Enabling `rdb_is_ha` or using `redis_cluster_size: 2` provisions
  standby replicas with automatic failover. Use for production, skip for dev/test.
* **Passwords are not exported** -- Admin passwords are user-specified in the values file
  and should be managed through your organization's secrets workflow.
* **Change default passwords** -- The default `changeme123` passwords are placeholders.
  Always set strong passwords before deploying to any non-local environment.

---

© Planton. Licensed under [Apache-2.0](../../../LICENSE).

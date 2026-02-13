# ScalewayRdbInstance Examples

Copy-paste examples for common deployment patterns. Adjust `region`, `nodeType`, and credentials for your environment.

## Example 1: Development PostgreSQL (Minimal)

A single-node PostgreSQL instance for development with one database and one user.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayRdbInstance
metadata:
  name: dev-postgres
  env: development
spec:
  region: fr-par
  engine: PostgreSQL-16
  nodeType: DB-DEV-S
  disableBackup: true
  adminUser: admin
  adminPassword: dev-admin-password-123
  databases:
    - name: myapp
  users:
    - name: app_user
      password: app-user-password-123
      privileges:
        - databaseName: myapp
          permission: readwrite
```

## Example 2: Production PostgreSQL with HA

A high-availability PostgreSQL cluster with Private Network, encryption, ACL, and separate application users.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayRdbInstance
metadata:
  name: prod-postgres
  org: mycompany
  env: production
spec:
  region: fr-par
  engine: PostgreSQL-16
  nodeType: db-gp-xs
  isHaCluster: true
  volumeType: bssd
  volumeSizeInGb: 100
  encryptionAtRest: true
  backupScheduleFrequencyHours: 12
  backupScheduleRetentionDays: 30
  privateNetworkId:
    value: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  aclRules:
    - ip: "10.0.0.0/8"
      description: "Private network range"
    - ip: "203.0.113.10/32"
      description: "Office VPN egress"
  adminUser: pgadmin
  adminPassword: very-strong-random-password-here
  databases:
    - name: appdb
    - name: analytics
  users:
    - name: app_service
      password: app-service-password-random
      privileges:
        - databaseName: appdb
          permission: readwrite
    - name: analytics_reader
      password: analytics-reader-password
      privileges:
        - databaseName: analytics
          permission: all
        - databaseName: appdb
          permission: readonly
  settings:
    max_connections: "200"
    work_mem: "64MB"
    effective_cache_size: "4GB"
```

## Example 3: MySQL for Web Applications

A MySQL instance for traditional web application stacks.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayRdbInstance
metadata:
  name: web-mysql
  env: staging
spec:
  region: nl-ams
  engine: MySQL-8
  nodeType: DB-DEV-M
  adminUser: root_admin
  adminPassword: mysql-admin-password-123
  databases:
    - name: wordpress
    - name: sessions
  users:
    - name: wp_user
      password: wordpress-db-password
      privileges:
        - databaseName: wordpress
          permission: readwrite
        - databaseName: sessions
          permission: readwrite
  initSettings:
    lower_case_table_names: "1"
```

## Example 4: Multi-User with Fine-Grained Permissions

Multiple users with different permission levels on different databases.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayRdbInstance
metadata:
  name: multi-user-db
  org: dataplatform
  env: production
spec:
  region: fr-par
  engine: PostgreSQL-16
  nodeType: db-gp-s
  isHaCluster: true
  adminUser: db_admin
  adminPassword: admin-password-very-strong
  databases:
    - name: orders
    - name: inventory
    - name: reporting
  users:
    - name: order_service
      password: order-svc-password
      privileges:
        - databaseName: orders
          permission: readwrite
        - databaseName: inventory
          permission: readonly
    - name: inventory_service
      password: inventory-svc-password
      privileges:
        - databaseName: inventory
          permission: readwrite
    - name: reporting_etl
      password: etl-password
      privileges:
        - databaseName: orders
          permission: readonly
        - databaseName: inventory
          permission: readonly
        - databaseName: reporting
          permission: all
    - name: dashboard_viewer
      password: dashboard-readonly-password
      privileges:
        - databaseName: reporting
          permission: readonly
```

## Example 5: Locked-Down Instance (ACL Only, No Private Network)

A publicly accessible instance with strict ACL rules.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayRdbInstance
metadata:
  name: restricted-postgres
spec:
  region: pl-waw
  engine: PostgreSQL-15
  nodeType: DB-DEV-S
  aclRules:
    - ip: "198.51.100.0/24"
      description: "Corporate network"
    - ip: "203.0.113.5/32"
      description: "CI/CD runner"
  adminUser: admin
  adminPassword: restricted-admin-password
  databases:
    - name: cicd_db
```

## Example 6: Infra Chart Pattern (valueFrom Reference)

Using `valueFrom` to wire the Private Network from an upstream resource in an infra chart template.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayRdbInstance
metadata:
  name: "{{ values.env }}-postgres"
  org: "{{ values.org }}"
  env: "{{ values.env }}"
spec:
  region: "{{ values.region }}"
  engine: "PostgreSQL-16"
  nodeType: "{{ values.rdb_node_type | default('db-gp-xs') }}"
  isHaCluster: "{{ values.rdb_ha_enabled | default(false) }}"
  privateNetworkId:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: "{{ values.env }}-network"
      fieldPath: status.outputs.private_network_id
  aclRules:
    - ip: "10.0.0.0/8"
      description: "Private network range"
  adminUser: pgadmin
  adminPassword: "{{ values.rdb_admin_password }}"
  databases:
    - name: "{{ values.database_name | default('appdb') }}"
  users:
    - name: "{{ values.app_db_user | default('app') }}"
      password: "{{ values.app_db_password }}"
      privileges:
        - databaseName: "{{ values.database_name | default('appdb') }}"
          permission: readwrite
```

## Example 7: Bare Instance (No Databases or Users)

An instance with just the admin user -- databases and users managed externally (e.g., via application migrations).

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayRdbInstance
metadata:
  name: migration-managed-db
spec:
  region: fr-par
  engine: PostgreSQL-16
  nodeType: DB-DEV-S
  adminUser: admin
  adminPassword: admin-password-123
```

## Configuration Patterns Summary

| Pattern | HA | PN | ACL | Users | Use Case |
|---------|----|----|-----|-------|----------|
| Dev minimal | No | No | No | 1 | Local development |
| Staging | No | Yes | Yes | 1-2 | Pre-production testing |
| Production | Yes | Yes | Yes | 2+ | Live workloads |
| Multi-user | Yes | Yes | Yes | 3+ | Microservices with RBAC |
| Bare instance | No | No | No | 0 | Migration-managed DBs |

## Deployment Checklist

1. Choose engine and version (`PostgreSQL-16` or `MySQL-8`)
2. Select node type based on workload (dev vs production)
3. Enable HA for production (`isHaCluster: true`)
4. Attach Private Network for secure internal access
5. Set ACL rules to restrict public endpoint
6. Create application databases
7. Create application users with minimal permissions
8. Configure backup schedule appropriate for RPO
9. Enable encryption at rest for compliance
10. Tune engine settings if needed

## Next Steps

After deploying a `ScalewayRdbInstance`:

- **Connect applications** using `status.outputs.private_endpoint_ip` and `status.outputs.private_endpoint_port`
- **Verify TLS** using `status.outputs.certificate` as the CA cert
- **Monitor** via Scaleway console or integrate with your monitoring stack
- **Scale vertically** by changing `nodeType` (no data migration needed)
- **Scale storage** by increasing `volumeSizeInGb` (only increases)

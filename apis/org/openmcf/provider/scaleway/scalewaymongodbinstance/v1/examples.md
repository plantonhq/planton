# ScalewayMongodbInstance Examples

Copy-paste examples for common deployment patterns. Adjust `region`, `nodeType`, and credentials for your environment.

## Example 1: Development Instance (Minimal)

A single-node MongoDB instance for development with one application user.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayMongodbInstance
metadata:
  name: dev-mongodb
  env: development
spec:
  region: fr-par
  version: "7.0.12"
  nodeType: MGDB-PLAY2-NANO
  nodeNumber: 1
  adminUser: admin
  adminPassword: dev-admin-password-123
  users:
    - name: app_user
      password: app-user-password-123
      roles:
        - role: read_write
          databaseName: myapp
```

## Example 2: Production Replica Set with Private Network

A 3-node replica set with Private Network, dedicated vCPU, high-IOPS storage, and snapshot scheduling.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayMongodbInstance
metadata:
  name: prod-mongodb
  org: mycompany
  env: production
spec:
  region: fr-par
  version: "7.0.12"
  nodeType: MGDB-POP2-8C-32G
  nodeNumber: 3
  volumeType: sbs_15k
  volumeSizeInGb: 100
  enableSnapshotSchedule: true
  snapshotScheduleFrequencyHours: 12
  snapshotScheduleRetentionDays: 30
  privateNetworkId:
    value: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  adminUser: dbadmin
  adminPassword: very-strong-random-password-here
  users:
    - name: order_service
      password: order-svc-password-random
      roles:
        - role: read_write
          databaseName: orders
    - name: analytics_reader
      password: analytics-reader-password
      roles:
        - role: read
          anyDatabase: true
```

## Example 3: Multi-User with Fine-Grained Roles

Multiple users with different role assignments across different databases.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayMongodbInstance
metadata:
  name: multi-user-mongodb
  org: dataplatform
  env: production
spec:
  region: fr-par
  version: "7.0.12"
  nodeType: MGDB-PRO2-S
  nodeNumber: 3
  privateNetworkId:
    value: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  adminUser: db_admin
  adminPassword: admin-password-very-strong
  users:
    - name: order_service
      password: order-svc-password
      roles:
        - role: read_write
          databaseName: orders
        - role: read
          databaseName: inventory
    - name: inventory_service
      password: inventory-svc-password
      roles:
        - role: read_write
          databaseName: inventory
    - name: reporting_etl
      password: etl-password
      roles:
        - role: read
          anyDatabase: true
    - name: index_manager
      password: index-mgr-password
      roles:
        - role: db_admin
          databaseName: orders
        - role: db_admin
          databaseName: inventory
```

## Example 4: Private Network with Public Admin Access

A private MongoDB instance that also has a public endpoint for admin access (e.g., from a developer's laptop or CI/CD pipeline).

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayMongodbInstance
metadata:
  name: hybrid-mongodb
  env: staging
spec:
  region: fr-par
  version: "7.0.12"
  nodeType: MGDB-PRO2-XS
  nodeNumber: 1
  privateNetworkId:
    value: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  enablePublicNetwork: true
  adminUser: admin
  adminPassword: staging-admin-password
  users:
    - name: app_user
      password: app-user-password
      roles:
        - role: read_write
          databaseName: staging_app
```

**Note**: The public endpoint has no IP-based access control. Only enable it if you have other network-level controls or accept the security risk.

## Example 5: Bare Instance (No Additional Users)

An instance with just the admin user -- application users managed externally (e.g., via application code or scripts).

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayMongodbInstance
metadata:
  name: app-managed-mongodb
spec:
  region: fr-par
  version: "7.0.12"
  nodeType: MGDB-PLAY2-NANO
  nodeNumber: 1
  adminUser: admin
  adminPassword: admin-password-123
```

## Example 6: Infra Chart Pattern (valueFrom Reference)

Using `valueFrom` to wire the Private Network from an upstream resource in an infra chart template.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayMongodbInstance
metadata:
  name: "{{ values.env }}-mongodb"
  org: "{{ values.org }}"
  env: "{{ values.env }}"
spec:
  region: "{{ values.region }}"
  version: "{{ values.mongodb_version | default('7.0.12') }}"
  nodeType: "{{ values.mongodb_node_type | default('MGDB-PRO2-XS') }}"
  nodeNumber: "{{ values.mongodb_node_number | default(1) }}"
  volumeType: "{{ values.mongodb_volume_type | default('sbs_5k') }}"
  volumeSizeInGb: "{{ values.mongodb_volume_size_gb | default(10) }}"
  enableSnapshotSchedule: "{{ values.mongodb_snapshots_enabled | default(false) }}"
  privateNetworkId:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: "{{ values.env }}-network"
      fieldPath: status.outputs.private_network_id
  adminUser: "{{ values.mongodb_admin_user | default('dbadmin') }}"
  adminPassword: "{{ values.mongodb_admin_password }}"
  users:
    - name: "{{ values.app_db_user | default('app') }}"
      password: "{{ values.app_db_password }}"
      roles:
        - role: read_write
          databaseName: "{{ values.database_name | default('appdb') }}"
```

## Example 7: Large Volume with Custom Settings

An instance with large storage and MongoDB engine tuning.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayMongodbInstance
metadata:
  name: large-mongodb
  env: production
spec:
  region: fr-par
  version: "7.0.12"
  nodeType: MGDB-POP2-16C-64G
  nodeNumber: 3
  volumeType: sbs_15k
  volumeSizeInGb: 500
  enableSnapshotSchedule: true
  snapshotScheduleFrequencyHours: 6
  snapshotScheduleRetentionDays: 60
  privateNetworkId:
    value: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  adminUser: admin
  adminPassword: ultra-strong-password
  settings: {}
  users:
    - name: app_service
      password: app-service-password
      roles:
        - role: read_write
          databaseName: main_db
```

## Configuration Patterns Summary

| Pattern | Nodes | PN | Public | Users | Use Case |
|---------|-------|----|--------|-------|----------|
| Dev minimal | 1 | No | Yes (default) | 1 | Local development |
| Staging hybrid | 1 | Yes | Yes (opt-in) | 1-2 | Pre-production with admin access |
| Production private | 3 | Yes | No | 2+ | Live workloads |
| Multi-user | 3 | Yes | No | 3+ | Microservices with RBAC |
| Bare instance | 1 | No | Yes (default) | 0 | App-managed users |

## Deployment Checklist

1. Choose MongoDB version (currently 7.0.x)
2. Select node type based on workload (shared vs dedicated vCPU)
3. Set node count: 1 for dev/test, 3 for production
4. Attach Private Network for secure internal access
5. Do NOT enable public network unless needed (no ACL available)
6. Configure snapshot schedule for production
7. Create application users with minimal roles
8. Choose volume type and size based on IOPS needs and data growth

## Connecting to MongoDB

### Using the mongo shell

```bash
# Public endpoint
mongosh "mongodb://<user>:<password>@<public_dns_record>:<port>/mydb?tls=true&tlsCAFile=ca.pem"

# Private endpoint (from within the Private Network)
mongosh "mongodb://<user>:<password>@<private_ip>:<port>/mydb?tls=true&tlsCAFile=ca.pem"
```

### Using the connection string in applications

```
mongodb://<user>:<password>@<host>:<port>/mydb?tls=true&tlsCAFile=/path/to/ca.pem
```

The TLS CA certificate is available in `status.outputs.tls_certificate`.

## Next Steps

After deploying a `ScalewayMongodbInstance`:

- **Connect applications** using `status.outputs.private_ips` and `status.outputs.private_port`
- **Verify TLS** using `status.outputs.tls_certificate` as the CA cert
- **Monitor** via Scaleway console or integrate with your monitoring stack
- **Scale vertically** by changing `nodeType` (no data migration needed)
- **Scale storage** by increasing `volumeSizeInGb` (only increases)
- **Enable HA** by changing `nodeNumber` from 1 to 3 (may recreate instance)

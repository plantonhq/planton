# AliCloudPolardbCluster Examples

## Minimal: MySQL Development Cluster

A basic MySQL 8.0 PolarDB cluster for development and testing with a single node.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudPolardbCluster
metadata:
  name: dev-polardb
spec:
  region: cn-hangzhou
  dbType: MySQL
  dbVersion: "8.0"
  dbNodeClass: polar.mysql.x4.medium
  vswitchId:
    value: vsw-abc123
  dbNodeCount: 1
  creationCategory: Basic
  databases:
    - dbName: appdb
  accounts:
    - accountName: dev_user
      accountPassword: "${DEV_DB_PASSWORD}"
      privileges:
        - dbNames: [appdb]
          accountPrivilege: ReadWrite
```

## Production: MySQL HA with Encryption and Monitoring

A production-grade MySQL PolarDB cluster with 4 nodes (1 primary + 3 read replicas), TDE encryption, audit logging, and cross-AZ deployment.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudPolardbCluster
metadata:
  name: prod-mysql-polardb
  org: acme-corp
  env: production
spec:
  region: cn-shanghai
  dbType: MySQL
  dbVersion: "8.0"
  dbNodeClass: polar.mysql.x4.xlarge
  vswitchId:
    valueFrom:
      name: prod-db-vswitch
  dbNodeCount: 4
  zoneId: cn-shanghai-a
  creationCategory: Normal
  subCategory: Exclusive
  securityIps:
    - "10.0.0.0/8"
  securityGroupIds:
    - sg-db-access
  maintainTime: "02:00Z-03:00Z"
  resourceGroupId: rg-production
  tdeStatus: Enabled
  encryptionKey: kms-key-abc123
  deletionLock: 1
  collectorStatus: Enable
  backupRetentionPolicyOnClusterDeletion: LATEST
  tags:
    team: platform
    cost-center: database
  parameters:
    - name: loose_innodb_buffer_pool_size
      value: "4294967296"
    - name: max_connections
      value: "1000"
  databases:
    - dbName: api_service
      characterSetName: utf8mb4
      dbDescription: Primary API service database
    - dbName: analytics
      characterSetName: utf8mb4
      dbDescription: Analytics and reporting
  accounts:
    - accountName: api_svc
      accountPassword: "${API_DB_PASSWORD}"
      accountDescription: API service account
      privileges:
        - dbNames: [api_service]
          accountPrivilege: ReadWrite
    - accountName: analyst
      accountPassword: "${ANALYST_DB_PASSWORD}"
      accountDescription: Read-only analytics access
      privileges:
        - dbNames: [analytics]
          accountPrivilege: ReadWrite
        - dbNames: [api_service]
          accountPrivilege: ReadOnly
```

## PostgreSQL Cluster with Collation Settings

A PostgreSQL PolarDB cluster with database-level collation settings for internationalized applications.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudPolardbCluster
metadata:
  name: pg-polardb
  org: global-corp
  env: production
spec:
  region: cn-hangzhou
  dbType: PostgreSQL
  dbVersion: "14"
  dbNodeClass: polar.pg.x4.large
  vswitchId:
    valueFrom:
      name: db-vswitch
  dbNodeCount: 3
  securityIps:
    - "172.16.0.0/12"
  deletionLock: 1
  tags:
    compliance: gdpr
  databases:
    - dbName: user_service
      characterSetName: UTF8
      collate: en_US.UTF-8
      ctype: en_US.UTF-8
      dbDescription: User service database
    - dbName: audit_log
      characterSetName: UTF8
      collate: C
      ctype: C
  accounts:
    - accountName: app_svc
      accountPassword: "${APP_DB_PASSWORD}"
      privileges:
        - dbNames: [user_service]
          accountPrivilege: ReadWrite
        - dbNames: [audit_log]
          accountPrivilege: ReadOnly
    - accountName: dba_admin
      accountPassword: "${DBA_PASSWORD}"
      accountType: Super
```

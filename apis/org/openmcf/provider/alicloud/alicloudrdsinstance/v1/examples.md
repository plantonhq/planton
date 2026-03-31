# AliCloudRdsInstance Examples

## Minimal: MySQL Development Instance

A basic MySQL 8.0 instance for development and testing.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRdsInstance
metadata:
  name: dev-mysql
spec:
  region: cn-hangzhou
  engine: MySQL
  engineVersion: "8.0"
  instanceType: rds.mysql.t1.small
  instanceStorage: 20
  vswitchId:
    value: vsw-abc123
  category: Basic
  databases:
    - name: appdb
  accounts:
    - accountName: dev_user
      accountPassword: "${DEV_DB_PASSWORD}"
      privileges:
        - databaseNames: [appdb]
          privilege: ReadWrite
```

## Production: PostgreSQL HA with SSL and Monitoring

A production-grade PostgreSQL instance with high availability, SSL encryption, cross-AZ deployment, and fine-grained monitoring.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRdsInstance
metadata:
  name: prod-pg
  org: acme-corp
  env: production
spec:
  region: cn-shanghai
  engine: PostgreSQL
  engineVersion: "16.0"
  instanceType: rds.pg.s2.large
  instanceStorage: 200
  vswitchId:
    valueFrom:
      name: prod-db-vswitch-a
  category: HighAvailability
  dbInstanceStorageType: cloud_essd
  zoneId: cn-shanghai-a
  zoneIdSlaveA: cn-shanghai-b
  securityIps:
    - "10.0.0.0/8"
  securityGroupIds:
    - sg-db-access
  monitoringPeriod: 60
  maintainTime: "02:00Z-06:00Z"
  deletionProtection: true
  sslAction: Open
  resourceGroupId: rg-production
  tags:
    team: platform
    cost-center: database
  parameters:
    - name: shared_buffers
      value: "4096MB"
    - name: max_connections
      value: "500"
  databases:
    - name: api_service
      description: Primary API service database
    - name: analytics
      description: Analytics and reporting
  accounts:
    - accountName: api_svc
      accountPassword: "${API_DB_PASSWORD}"
      accountDescription: API service account
      privileges:
        - databaseNames: [api_service]
          privilege: ReadWrite
    - accountName: analyst
      accountPassword: "${ANALYST_DB_PASSWORD}"
      accountDescription: Read-only analytics access
      privileges:
        - databaseNames: [analytics]
          privilege: ReadWrite
        - databaseNames: [api_service]
          privilege: ReadOnly
```

## MySQL with Encryption and Prepaid Billing

A production MySQL instance with TDE encryption, KMS key, and subscription billing for cost optimization.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRdsInstance
metadata:
  name: encrypted-mysql
  org: fintech-corp
  env: production
spec:
  region: cn-hangzhou
  engine: MySQL
  engineVersion: "8.0"
  instanceType: rds.mysql.s2.xlarge
  instanceStorage: 500
  vswitchId:
    valueFrom:
      name: secure-db-vswitch
  category: Finance
  dbInstanceStorageType: cloud_essd
  instanceChargeType: Prepaid
  period: 12
  autoRenew: true
  autoRenewPeriod: 3
  zoneId: cn-hangzhou-a
  zoneIdSlaveA: cn-hangzhou-b
  securityIps:
    - "172.16.0.0/12"
  deletionProtection: true
  tdeStatus: Enabled
  encryptionKey: kms-key-abc123
  sslAction: Open
  tags:
    compliance: pci-dss
    data-class: confidential
  databases:
    - name: transactions
      characterSet: utf8mb4
    - name: audit_log
      characterSet: utf8mb4
  accounts:
    - accountName: app_svc
      accountPassword: "${APP_DB_PASSWORD}"
      privileges:
        - databaseNames: [transactions]
          privilege: ReadWrite
        - databaseNames: [audit_log]
          privilege: ReadOnly
    - accountName: dba_admin
      accountPassword: "${DBA_PASSWORD}"
      accountType: Super
```

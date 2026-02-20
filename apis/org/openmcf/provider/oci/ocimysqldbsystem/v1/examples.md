# OCI MySQL DB System — Examples

## Table of Contents

- [Minimal Development Instance](#minimal-development-instance)
- [High Availability with Backups and PITR](#high-availability-with-backups-and-pitr)
- [Production with BYOK Encryption and Read Endpoint](#production-with-byok-encryption-and-read-endpoint)
- [Foreign Key References](#foreign-key-references)
- [Database Console and REST API](#database-console-and-rest-api)
- [Common Operations](#common-operations)
- [Best Practices](#best-practices)

---

## Minimal Development Instance

A single-instance MySQL DB System with shape defaults. No HA, no backups, Oracle-managed encryption.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciMysqlDbSystem
metadata:
  name: dev-mysql
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciMysqlDbSystem.dev-mysql
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:PHX-AD-1"
  shapeName: "MySQL.VM.Standard.E4.1.8GB"
  subnetId:
    value: "ocid1.subnet.oc1.phx..example"
  adminUsername: "admin"
  adminPassword: "D3v!P4ssw0rd"
```

**What this creates:** A single MySQL DB System on the smallest E4 shape with 8 GB memory, Oracle-managed encryption, and the default MySQL configuration. The primary endpoint is accessible at the auto-assigned private IP on port 3306.

---

## High Availability with Backups and PITR

Three-instance HA with daily backups retained for 14 days, point-in-time recovery, auto-expanding storage, and a Sunday maintenance window.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciMysqlDbSystem
metadata:
  name: staging-mysql
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.OciMysqlDbSystem.staging-mysql
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:PHX-AD-1"
  shapeName: "MySQL.VM.Standard.E4.4.64GB"
  subnetId:
    value: "ocid1.subnet.oc1.phx..example"
  adminUsername: "admin"
  adminPassword: "Stag1ng$ecure!9"
  mysqlVersion: "8.0.36"
  isHighlyAvailable: true
  hostnameLabel: "staging-mysql"
  dataStorage:
    dataStorageSizeInGb: 200
    isAutoExpandStorageEnabled: true
    maxStorageSizeInGbs: 32768
  backupPolicy:
    isEnabled: true
    retentionInDays: 14
    windowStartTime: "03:00"
    pitrPolicy:
      isEnabled: true
  maintenance:
    windowStartTime: "sun 04:00"
    maintenanceScheduleType: regular
    versionPreference: second_newest
    versionTrackPreference: long_term_support
```

**What this creates:** A three-instance HA MySQL DB System with automatic failover across fault domains. Daily backups run at 03:00 UTC with 14-day retention and PITR enabled. Storage starts at 200 GB and auto-expands up to 32 TB. Maintenance patches apply on Sundays at 04:00 UTC on the regular schedule, preferring the second-newest LTS version.

---

## Production with BYOK Encryption and Read Endpoint

Full production setup: customer-managed encryption via OCI Vault, deletion protection, read endpoint for read scaling, NSG attachment, and customer contact notifications.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciMysqlDbSystem
metadata:
  name: prod-mysql
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciMysqlDbSystem.prod-mysql
  env: prod
  org: acme
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:PHX-AD-1"
  shapeName: "MySQL.VM.Standard.E4.8.128GB"
  subnetId:
    value: "ocid1.subnet.oc1.phx..example"
  adminUsername: "dbadmin"
  adminPassword: "Pr0d!Str0ng#42"
  mysqlVersion: "8.0.36"
  isHighlyAvailable: true
  hostnameLabel: "prod-mysql"
  faultDomain: "FAULT-DOMAIN-1"
  description: "Production MySQL for order processing"
  crashRecovery: "ENABLED"
  databaseManagement: "ENABLED"
  nsgIds:
    - value: "ocid1.networksecuritygroup.oc1.phx..example"
  dataStorage:
    dataStorageSizeInGb: 500
    isAutoExpandStorageEnabled: true
    maxStorageSizeInGbs: 65536
  backupPolicy:
    isEnabled: true
    retentionInDays: 30
    windowStartTime: "02:00"
    pitrPolicy:
      isEnabled: true
  maintenance:
    windowStartTime: "sun 05:00"
    maintenanceScheduleType: regular
    versionPreference: oldest
    versionTrackPreference: long_term_support
  deletionPolicy:
    automaticBackupRetention: "RETAIN"
    finalBackup: "REQUIRE_FINAL_BACKUP"
    isDeleteProtected: true
  encryptData:
    keyGenerationType: byok
    keyId:
      value: "ocid1.key.oc1.phx..example"
  secureConnections:
    certificateGenerationType: system_cert
  customerContacts:
    - email: "dba-team@example.com"
    - email: "oncall@example.com"
  readEndpoint:
    isEnabled: true
    readEndpointHostnameLabel: "prod-mysql-ro"
```

**What this creates:** A three-instance HA MySQL DB System on an 8-OCPU / 128 GB shape. Data is encrypted with a customer-managed Vault key. Deletion is blocked until `isDeleteProtected` is set to `false`, and a final backup is required before deletion. A read-only endpoint at `prod-mysql-ro` distributes read queries. The DBA team and on-call get email notifications for maintenance windows and critical alerts.

---

## Foreign Key References

Reference OpenMCF-managed compartments, subnets, and NSGs instead of hardcoding OCIDs. OpenMCF resolves the references at deployment time.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciMysqlDbSystem
metadata:
  name: ref-mysql
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciMysqlDbSystem.ref-mysql
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  availabilityDomain: "Uocm:PHX-AD-1"
  shapeName: "MySQL.VM.Standard.E4.4.64GB"
  subnetId:
    valueFrom:
      kind: OciSubnet
      name: db-subnet
      fieldPath: status.outputs.subnetId
  adminUsername: "admin"
  adminPassword: "R3fPass!word1"
  isHighlyAvailable: true
  nsgIds:
    - valueFrom:
        kind: OciSecurityGroup
        name: mysql-nsg
        fieldPath: status.outputs.networkSecurityGroupId
```

**What this creates:** The same DB System as the HA example, but all infrastructure OCIDs are resolved from other OpenMCF-managed resources. The compartment comes from `prod-compartment`, the subnet from `db-subnet`, and the NSG from `mysql-nsg`.

---

## Database Console and REST API

Enable the web-based MySQL management console and the MySQL Router REST API.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciMysqlDbSystem
metadata:
  name: console-mysql
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciMysqlDbSystem.console-mysql
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:PHX-AD-1"
  shapeName: "MySQL.VM.Standard.E4.1.8GB"
  subnetId:
    value: "ocid1.subnet.oc1.phx..example"
  adminUsername: "admin"
  adminPassword: "C0nsole!Pass1"
  databaseConsole:
    status: enabled
    port: 443
  rest:
    port: 8443
```

**What this creates:** A MySQL DB System with the database console available on port 443 and the REST API on port 8443. Both are accessible on the DB System's private IP.

---

## Common Operations

### Deploying

```shell
openmcf apply -f mysql-db.yaml
```

### Checking status

```shell
openmcf get OciMysqlDbSystem my-mysql -o yaml
```

The `status.outputs` section contains the DB System OCID, endpoint hostname, IP address, and port.

### Connecting to the database

After deployment, connect using the endpoint from stack outputs:

```shell
mysql -h <endpoint_ip_address> -P <endpoint_port> -u <adminUsername> -p
```

### Enabling HA on an existing instance

Update the manifest to set `isHighlyAvailable: true` and re-apply:

```shell
openmcf apply -f mysql-db.yaml
```

OCI will provision standby instances across fault domains. This is an in-place update — no recreation.

### Disabling deletion protection before teardown

Set `deletionPolicy.isDeleteProtected` to `false` in the manifest and apply before destroying:

```shell
openmcf apply -f mysql-db.yaml
openmcf destroy -f mysql-db.yaml
```

---

## Best Practices

1. **Always set `adminPassword` securely** — use a secret manager or environment variable injection rather than committing passwords to version control.
2. **Enable HA for any non-development workload** — `isHighlyAvailable: true` provisions three instances across fault domains. The cost increase is offset by automatic failover.
3. **Enable PITR alongside backups** — `pitrPolicy.isEnabled: true` within `backupPolicy` allows recovery to any point within the retention window, not just the daily snapshot.
4. **Use deletion protection in production** — set `deletionPolicy.isDeleteProtected: true` and `deletionPolicy.finalBackup: REQUIRE_FINAL_BACKUP` to prevent accidental data loss.
5. **Size storage with auto-expand** — set an initial `dataStorageSizeInGb` and enable `isAutoExpandStorageEnabled` with a `maxStorageSizeInGbs` ceiling to avoid manual intervention.
6. **Pin MySQL version** — specify `mysqlVersion` explicitly (e.g., `8.0.36`) to prevent unexpected major/minor version changes on recreation.
7. **Attach NSGs** — use `nsgIds` to restrict network access to the DB System VNIC instead of relying on subnet-level security lists alone.
8. **Use `valueFrom` references** — reference compartments, subnets, and NSGs via foreign keys to keep manifests portable across environments.

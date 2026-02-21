# OCI Autonomous Database — Examples

## Table of Contents

- [Always Free ATP](#always-free-atp)
- [Development ATP with ECPU](#development-atp-with-ecpu)
- [Autonomous Data Warehouse with Auto-Scaling](#autonomous-data-warehouse-with-auto-scaling)
- [Production ATP with Private Endpoint](#production-atp-with-private-endpoint)
- [Dedicated Exadata ATP with Data Guard](#dedicated-exadata-atp-with-data-guard)
- [Common Operations](#common-operations)
- [Best Practices](#best-practices)

---

## Always Free ATP

An Always Free Autonomous Transaction Processing database for experimentation. Limited to fixed compute and storage; reclaimed after extended inactivity.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciAutonomousDatabase
metadata:
  name: free-atp
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciAutonomousDatabase.free-atp
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  dbName: "freeatp"
  dbWorkload: oltp
  isFreeTier: true
  adminPassword: "FreePass#2026abc"
```

## Development ATP with ECPU

A serverless ATP database using the ECPU compute model with minimal resources. Suitable for application development and integration testing.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciAutonomousDatabase
metadata:
  name: dev-atp
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciAutonomousDatabase.dev-atp
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  dbName: "devatp"
  displayName: "Dev ATP"
  dbWorkload: oltp
  dbVersion: "23ai"
  computeModel: ecpu
  computeCount: 2
  dataStorageSizeInTbs: 1
  licenseModel: license_included
  isAutoScalingEnabled: false
  adminPassword: "DevPass#2026abc"
```

## Autonomous Data Warehouse with Auto-Scaling

An ADW database for analytic workloads with BYOL licensing. Compute auto-scaling handles burst queries without over-provisioning. Storage auto-scaling expands capacity on demand.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciAutonomousDatabase
metadata:
  name: analytics-adw
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: analytics
    pulumi.openmcf.org/stack.name: staging.OciAutonomousDatabase.analytics-adw
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  dbName: "analyticsadw"
  displayName: "Analytics ADW"
  dbWorkload: dw
  databaseEdition: enterprise_edition
  computeModel: ecpu
  computeCount: 4
  dataStorageSizeInTbs: 5
  licenseModel: bring_your_own_license
  isAutoScalingEnabled: true
  isAutoScalingForStorageEnabled: true
  adminPassword: "AnalyticsPass#2026"
  autonomousMaintenanceScheduleType: regular
  customerContacts:
    - email: "data-team@example.com"
```

## Production ATP with Private Endpoint

A production ATP database placed behind a private endpoint in a VCN subnet. Uses Vault-managed credentials, customer-managed KMS encryption, mTLS enforcement, Data Guard for HA, and 30-day backup retention. All infrastructure references use `valueFrom` for composability.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciAutonomousDatabase
metadata:
  name: prod-atp
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: prod-databases
    pulumi.openmcf.org/stack.name: prod.OciAutonomousDatabase.prod-atp
  env: prod
  org: acme
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  dbName: "prodatp"
  displayName: "Production ATP"
  dbWorkload: oltp
  dbVersion: "23ai"
  databaseEdition: enterprise_edition
  computeModel: ecpu
  computeCount: 8
  dataStorageSizeInTbs: 10
  licenseModel: bring_your_own_license
  isAutoScalingEnabled: true
  isAutoScalingForStorageEnabled: true
  secretId:
    value: "ocid1.vaultsecret.oc1..example"
  subnetId:
    valueFrom:
      kind: OciSubnet
      name: prod-db-subnet
      fieldPath: status.outputs.subnetId
  nsgIds:
    - valueFrom:
        kind: OciSecurityGroup
        name: prod-db-nsg
        fieldPath: status.outputs.networkSecurityGroupId
  privateEndpointLabel: "prodatp"
  isMtlsConnectionRequired: true
  kmsKeyId:
    value: "ocid1.key.oc1..example"
  vaultId:
    value: "ocid1.vault.oc1..example"
  isLocalDataGuardEnabled: true
  backupRetentionPeriodInDays: 30
  autonomousMaintenanceScheduleType: regular
  customerContacts:
    - email: "dba-team@acme.com"
    - email: "oncall@acme.com"
```

## Dedicated Exadata ATP with Data Guard

An ATP database on dedicated Exadata infrastructure with storage specified in gigabytes for finer granularity. Requires an existing autonomous container database.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciAutonomousDatabase
metadata:
  name: dedicated-atp
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: dedicated-infra
    pulumi.openmcf.org/stack.name: prod.OciAutonomousDatabase.dedicated-atp
  env: prod
  org: acme
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  dbName: "dedicatedatp"
  displayName: "Dedicated ATP"
  dbWorkload: oltp
  databaseEdition: enterprise_edition
  computeModel: ecpu
  computeCount: 16
  dataStorageSizeInGb: 512
  licenseModel: bring_your_own_license
  isDedicated: true
  autonomousContainerDatabaseId:
    value: "ocid1.autonomouscontainerdatabase.oc1..example"
  isLocalDataGuardEnabled: true
  backupRetentionPeriodInDays: 60
  secretId:
    value: "ocid1.vaultsecret.oc1..example"
  autonomousMaintenanceScheduleType: early
  customerContacts:
    - email: "dba-team@acme.com"
```

---

## Common Operations

### Scaling Compute

Change `computeCount` in the manifest and re-apply. With `isAutoScalingEnabled: true`, the database can temporarily use up to 3x the provisioned count, but the base allocation is what you pay for at rest.

### Scaling Storage

Update `dataStorageSizeInTbs` (or `dataStorageSizeInGb` for dedicated) and re-apply. With `isAutoScalingForStorageEnabled: true`, storage expands automatically — the field sets the allocated baseline.

### Rotating the Admin Password

Update `adminPassword` in the manifest and re-apply, or update the Vault secret referenced by `secretId` and bump `secretVersionNumber`.

### Switching Maintenance Schedule

Change `autonomousMaintenanceScheduleType` from `early` to `regular` (or vice versa) and re-apply. This controls when OCI applies quarterly patches.

### Adding IP Allowlist Entries

Add entries to `whitelistedIps` and re-apply. Entries can be individual IPs (`10.0.0.1`), CIDR blocks (`10.0.0.0/24`), or VCN OCIDs.

---

## Best Practices

1. **Use Vault secrets in production** — set `secretId` instead of `adminPassword` to avoid credentials in version control.
2. **Enable auto-scaling for variable workloads** — `isAutoScalingEnabled` handles compute bursts; `isAutoScalingForStorageEnabled` prevents storage exhaustion.
3. **Use private endpoints** — set `subnetId` and `nsgIds` to keep database traffic off the public internet.
4. **Enable Data Guard for production** — `isLocalDataGuardEnabled` provides automatic failover with no application changes required.
5. **Set `dbName` carefully** — it cannot be changed after creation. Keep it short, alphanumeric, and descriptive.
6. **Choose character sets at creation** — `characterSet` and `ncharacterSet` are immutable. `AL32UTF8` and `AL16UTF16` are appropriate for most workloads.
7. **Use `valueFrom` references** — reference OciCompartment, OciSubnet, and OciSecurityGroup resources instead of hardcoding OCIDs for composability and drift detection.
8. **Set backup retention explicitly** — `backupRetentionPeriodInDays` makes the retention policy visible in the manifest rather than relying on service defaults.

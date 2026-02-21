# OciDbSystem Examples

Complete YAML manifests for common OCI DB System deployment patterns.

## Table of Contents

- [Minimal Development DB](#minimal-development-db)
- [Enterprise Edition with Backups and PDB](#enterprise-edition-with-backups-and-pdb)
- [2-Node RAC Production Cluster](#2-node-rac-production-cluster)
- [KMS-Encrypted DB with Foreign Key References](#kms-encrypted-db-with-foreign-key-references)
- [Bare Metal with High-Performance Storage](#bare-metal-with-high-performance-storage)

---

## Minimal Development DB

A single-node VM DB System with Oracle 19c. No backups, no PDB, no NSGs — the smallest viable configuration for development or experimentation.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDbSystem
metadata:
  name: dev-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciDbSystem.dev-db
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:PHX-AD-1"
  shape: "VM.Standard2.2"
  subnetId:
    value: "ocid1.subnet.oc1.phx.example"
  sshPublicKeys:
    - "ssh-rsa AAAA...your-key"
  hostname: "devdb"
  dbHome:
    dbVersion: "19.0.0.0"
    database:
      adminPassword: "BEstr0ng#2024"
      dbName: "DEVDB"
```

## Enterprise Edition with Backups and PDB

A flex-shape VM with Enterprise Edition, BYOL licensing, explicit storage sizing, a pluggable database, and automatic daily backups with 30-day retention.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDbSystem
metadata:
  name: staging-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.OciDbSystem.staging-db
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:PHX-AD-1"
  shape: "VM.Standard.E4.Flex"
  cpuCoreCount: 4
  subnetId:
    value: "ocid1.subnet.oc1.phx.example"
  sshPublicKeys:
    - "ssh-rsa AAAA...your-key"
  hostname: "stagingdb"
  databaseEdition: enterprise_edition
  licenseModel: bring_your_own_license
  dataStorageSizeInGb: 1024
  timeZone: "UTC"
  dataCollectionOptions:
    isDiagnosticsEventsEnabled: true
    isHealthMonitoringEnabled: true
  dbSystemOptions:
    storageManagement: asm
  dbHome:
    dbVersion: "19.0.0.0"
    displayName: "staging-home"
    database:
      adminPassword: "BEstr0ng#2024"
      dbName: "STAGDB"
      characterSet: "AL32UTF8"
      ncharacterSet: "AL16UTF16"
      pdbName: "STAGPDB"
      dbBackupConfig:
        autoBackupEnabled: true
        autoBackupWindow: "SLOT_TWO"
        recoveryWindowInDays: 30
```

## 2-Node RAC Production Cluster

A 2-node RAC cluster with Extreme Performance edition, fault domain distribution, backup subnet, NSG isolation, and a quarterly rolling maintenance window. This configuration requires `enterprise_edition_extreme_performance`.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDbSystem
metadata:
  name: prod-rac
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciDbSystem.prod-rac
  env: prod
  org: acme
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:PHX-AD-1"
  shape: "VM.Standard2.8"
  cpuCoreCount: 8
  subnetId:
    value: "ocid1.subnet.oc1.phx.example"
  sshPublicKeys:
    - "ssh-rsa AAAA...key-one"
    - "ssh-rsa AAAA...key-two"
  hostname: "prodrac"
  databaseEdition: enterprise_edition_extreme_performance
  licenseModel: license_included
  dataStorageSizeInGb: 4096
  nodeCount: 2
  clusterName: "prodclstr"
  faultDomains:
    - "FAULT-DOMAIN-1"
    - "FAULT-DOMAIN-2"
  nsgIds:
    - value: "ocid1.networksecuritygroup.oc1.phx.example"
  backupSubnetId:
    value: "ocid1.subnet.oc1.phx.backup-example"
  backupNetworkNsgIds:
    - value: "ocid1.networksecuritygroup.oc1.phx.backup-nsg"
  timeZone: "UTC"
  storageVolumePerformanceMode: balanced
  dataCollectionOptions:
    isDiagnosticsEventsEnabled: true
    isHealthMonitoringEnabled: true
    isIncidentLogsEnabled: true
  maintenanceWindowDetails:
    preference: custom_preference
    patchingMode: rolling
    leadTimeInWeeks: 2
    months:
      - "JANUARY"
      - "APRIL"
      - "JULY"
      - "OCTOBER"
    weeksOfMonth:
      - 2
    daysOfWeek:
      - "SUNDAY"
    hoursOfDay:
      - 4
  dbHome:
    dbVersion: "19.0.0.0"
    displayName: "prod-rac-home"
    database:
      adminPassword: "BEstr0ng#Prod2024"
      dbName: "PRODDB"
      characterSet: "AL32UTF8"
      ncharacterSet: "AL16UTF16"
      pdbName: "PRODPDB"
      dbBackupConfig:
        autoBackupEnabled: true
        autoBackupWindow: "SLOT_FOUR"
        recoveryWindowInDays: 60
```

## KMS-Encrypted DB with Foreign Key References

Uses `valueFrom` references for compartment, subnet, and NSG instead of hardcoded OCIDs. Enables KMS encryption at both the DB System level and TDE at the database level.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDbSystem
metadata:
  name: encrypted-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciDbSystem.encrypted-db
  env: prod
  org: acme
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  availabilityDomain: "Uocm:PHX-AD-1"
  shape: "VM.Standard2.4"
  cpuCoreCount: 4
  subnetId:
    valueFrom:
      kind: OciSubnet
      name: db-subnet
      fieldPath: status.outputs.subnetId
  sshPublicKeys:
    - "ssh-rsa AAAA...your-key"
  hostname: "encrypteddb"
  databaseEdition: enterprise_edition_high_performance
  licenseModel: bring_your_own_license
  dataStorageSizeInGb: 2048
  kmsKeyId:
    value: "ocid1.key.oc1.phx.example"
  nsgIds:
    - valueFrom:
        kind: OciSecurityGroup
        name: db-nsg
        fieldPath: status.outputs.networkSecurityGroupId
  dbHome:
    dbVersion: "19.0.0.0"
    displayName: "encrypted-home"
    database:
      adminPassword: "BEstr0ng#Enc2024"
      dbName: "ENCDB"
      pdbName: "ENCPDB"
      kmsKeyId:
        value: "ocid1.key.oc1.phx.tde-example"
      kmsKeyVersionId: "ocid1.keyversion.oc1.phx.example"
      vaultId:
        value: "ocid1.vault.oc1.phx.example"
      dbBackupConfig:
        autoBackupEnabled: true
        recoveryWindowInDays: 45
```

## Bare Metal with High-Performance Storage

A Bare Metal DB System with high disk redundancy, 80% data allocation, high-performance storage volumes, and a sparse disk group. BM shapes provide dedicated hardware and full control over storage layout.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDbSystem
metadata:
  name: bm-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciDbSystem.bm-db
  env: prod
  org: acme
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:PHX-AD-1"
  shape: "BM.DenseIO2.52"
  subnetId:
    value: "ocid1.subnet.oc1.phx.example"
  sshPublicKeys:
    - "ssh-rsa AAAA...your-key"
  hostname: "bmdb"
  databaseEdition: enterprise_edition_extreme_performance
  licenseModel: license_included
  dataStoragePercentage: 80
  diskRedundancy: high
  sparseDiskgroup: true
  storageVolumePerformanceMode: high_performance
  timeZone: "US/Pacific"
  dbHome:
    dbVersion: "19.0.0.0"
    displayName: "bm-home"
    database:
      adminPassword: "BEstr0ng#BM2024"
      dbName: "BMDB"
      characterSet: "AL32UTF8"
      pdbName: "BMPDB"
      dbBackupConfig:
        autoBackupEnabled: true
        autoBackupWindow: "SLOT_SIX"
        recoveryWindowInDays: 60
```

---

## Common Operations

### Scaling CPU cores

Update `cpuCoreCount` in the manifest and re-apply. This is a non-destructive update for flex shapes:

```yaml
spec:
  cpuCoreCount: 8  # was 4
```

### Increasing storage

Update `dataStorageSizeInGb` to a larger value. Storage can only be scaled up, not down:

```yaml
spec:
  dataStorageSizeInGb: 2048  # was 1024
```

### Enabling backups on an existing DB

Add the `dbBackupConfig` block to the database section:

```yaml
spec:
  dbHome:
    database:
      dbBackupConfig:
        autoBackupEnabled: true
        recoveryWindowInDays: 30
```

---

## Best Practices

1. **Use foreign key references** — prefer `valueFrom` over hardcoded OCIDs to keep manifests portable across environments.
2. **Set explicit storage sizes for VM shapes** — `dataStorageSizeInGb` is required for VM DB Systems; omitting it causes provisioning to use defaults that may not match your needs.
3. **Use BYOL when you have existing licenses** — `bring_your_own_license` reduces costs when you already own Oracle Database licenses with active Software Update License & Support.
4. **Enable backups from day one** — adding backups later is possible but starting with them avoids a gap in the recovery timeline.
5. **Use rolling patching for RAC** — set `patchingMode: rolling` in maintenance window details to avoid downtime during patching on 2-node clusters.
6. **Separate NSGs for primary and backup VNICs** — use `nsgIds` for the primary VNIC and `backupNetworkNsgIds` for the backup VNIC to apply different security rules to each.
7. **Pin the DB version** — specify an explicit `dbVersion` (e.g., `19.0.0.0`) rather than relying on defaults to ensure consistent deployments.

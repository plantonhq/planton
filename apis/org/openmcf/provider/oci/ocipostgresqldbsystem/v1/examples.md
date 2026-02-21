# OciPostgresqlDbSystem Examples

Complete YAML examples for deploying OCI PostgreSQL Database Systems via OpenMCF.

## Table of Contents

- [Minimal Standalone Instance](#minimal-standalone-instance)
- [Production with Read Replicas](#production-with-read-replicas)
- [Single-AD Development Instance](#single-ad-development-instance)
- [Monthly Backups with Per-Instance Pinning](#monthly-backups-with-per-instance-pinning)
- [Foreign Key References](#foreign-key-references)

---

## Minimal Standalone Instance

A single-node PostgreSQL 16 instance with regionally durable storage. Uses a plain-text password and no explicit backup policy.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciPostgresqlDbSystem
metadata:
  name: dev-postgres
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciPostgresqlDbSystem.dev-postgres
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  dbVersion: "16"
  shape: VM.Standard.E4.Flex
  instanceOcpuCount: 1
  instanceMemorySizeInGbs: 8
  instanceCount: 1
  networkDetails:
    subnetId:
      value: "ocid1.subnet.oc1..example"
  storageDetails:
    isRegionallyDurable: true
  credentials:
    username: postgres
    passwordDetails:
      passwordType: plain_text
      password: "dev-password-change-me"
```

---

## Production with Read Replicas

Multi-node setup with a primary instance and one read replica. Uses Vault-managed credentials, NSG-secured networking, a reader endpoint for distributing read queries, and daily backups retained for 30 days.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciPostgresqlDbSystem
metadata:
  name: prod-postgres
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-org
    pulumi.openmcf.org/project: acme-data
    pulumi.openmcf.org/stack.name: prod.OciPostgresqlDbSystem.prod-postgres
  env: prod
  org: acme
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  displayName: "Production PostgreSQL"
  dbVersion: "16"
  shape: VM.Standard.E4.Flex
  instanceOcpuCount: 4
  instanceMemorySizeInGbs: 32
  instanceCount: 2
  networkDetails:
    subnetId:
      value: "ocid1.subnet.oc1..example"
    nsgIds:
      - value: "ocid1.networksecuritygroup.oc1..example"
    isReaderEndpointEnabled: true
  storageDetails:
    isRegionallyDurable: true
    iops: 75000
  credentials:
    username: postgres
    passwordDetails:
      passwordType: vault_secret
      secretId:
        value: "ocid1.vaultsecret.oc1..example"
  managementPolicy:
    backupPolicy:
      kind: daily
      backupStart: "03:00"
      retentionDays: 30
    maintenanceWindowStart: sun 04:00:00
  description: "Production PostgreSQL cluster for ACME data services"
```

---

## Single-AD Development Instance

A cost-optimized instance in a single availability domain with no regional durability. Uses plain-text password and weekly backups with 7-day retention.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciPostgresqlDbSystem
metadata:
  name: test-postgres
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: test.OciPostgresqlDbSystem.test-postgres
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  displayName: "Test PostgreSQL"
  dbVersion: "16"
  shape: VM.Standard.E4.Flex
  instanceOcpuCount: 1
  instanceMemorySizeInGbs: 8
  instanceCount: 1
  networkDetails:
    subnetId:
      value: "ocid1.subnet.oc1..example"
  storageDetails:
    isRegionallyDurable: false
    availabilityDomain: "Uocm:PHX-AD-1"
  credentials:
    username: postgres
    passwordDetails:
      passwordType: plain_text
      password: "test-password"
  managementPolicy:
    backupPolicy:
      kind: weekly
      backupStart: "02:00"
      retentionDays: 7
      daysOfTheWeek:
        - SUNDAY
```

---

## Monthly Backups with Per-Instance Pinning

A two-node setup with monthly backups on the 1st and 15th, per-instance display names and static private IPs, and a custom PostgreSQL configuration applied via `configId`.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciPostgresqlDbSystem
metadata:
  name: pinned-postgres
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.OciPostgresqlDbSystem.pinned-postgres
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  dbVersion: "15"
  shape: VM.Standard.E4.Flex
  instanceOcpuCount: 2
  instanceMemorySizeInGbs: 16
  instanceCount: 2
  networkDetails:
    subnetId:
      value: "ocid1.subnet.oc1..example"
    nsgIds:
      - value: "ocid1.networksecuritygroup.oc1..example"
    isReaderEndpointEnabled: true
    primaryDbEndpointPrivateIp: "10.0.1.10"
  storageDetails:
    isRegionallyDurable: true
    iops: 50000
  credentials:
    username: postgres
    passwordDetails:
      passwordType: vault_secret
      secretId:
        value: "ocid1.vaultsecret.oc1..example"
      secretVersion: "2"
  configId:
    value: "ocid1.postgresqlconfiguration.oc1..example"
  managementPolicy:
    backupPolicy:
      kind: monthly
      backupStart: "01:00"
      retentionDays: 90
      daysOfTheMonth:
        - 1
        - 15
    maintenanceWindowStart: sat 05:00:00
  instancesDetails:
    - displayName: "primary-node"
      description: "Primary read-write node"
      privateIp: "10.0.1.10"
    - displayName: "replica-node-1"
      description: "Read replica"
      privateIp: "10.0.1.11"
```

---

## Foreign Key References

Reference other OpenMCF-managed resources (compartment, subnet, NSG) instead of hardcoding OCIDs. The `valueFrom` block resolves the referenced resource's stack output at deploy time.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciPostgresqlDbSystem
metadata:
  name: ref-postgres
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciPostgresqlDbSystem.ref-postgres
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  dbVersion: "16"
  shape: VM.Standard.E4.Flex
  instanceOcpuCount: 2
  instanceMemorySizeInGbs: 16
  instanceCount: 1
  networkDetails:
    subnetId:
      valueFrom:
        kind: OciSubnet
        name: db-subnet
        fieldPath: status.outputs.subnetId
    nsgIds:
      - valueFrom:
          kind: OciSecurityGroup
          name: db-nsg
          fieldPath: status.outputs.networkSecurityGroupId
  storageDetails:
    isRegionallyDurable: true
  credentials:
    username: postgres
    passwordDetails:
      passwordType: vault_secret
      secretId:
        value: "ocid1.vaultsecret.oc1..example"
  managementPolicy:
    backupPolicy:
      kind: daily
      backupStart: "03:00"
      retentionDays: 14
```

---

## Common Operations

### Scaling Up Compute

`instanceOcpuCount` and `instanceMemorySizeInGbs` are updatable. Change the values and re-apply:

```yaml
spec:
  instanceOcpuCount: 8      # was 4
  instanceMemorySizeInGbs: 64  # was 32
```

### Adding a Read Replica

Increase `instanceCount` and enable the reader endpoint:

```yaml
spec:
  instanceCount: 3  # was 2
  networkDetails:
    isReaderEndpointEnabled: true
```

### Changing IOPS Tier

`storageDetails.iops` is updatable:

```yaml
spec:
  storageDetails:
    iops: 100000  # was 75000
```

### Disabling Backups

Set the backup policy kind to `none`:

```yaml
spec:
  managementPolicy:
    backupPolicy:
      kind: none
```

---

## Best Practices

1. **Use Vault secrets in production** — avoid plain-text passwords outside of development. The `vault_secret` password type keeps credentials out of state files.
2. **Enable regional durability for production** — `isRegionallyDurable: true` replicates data across multiple availability domains. Use AD-local storage only for cost-optimized development.
3. **Attach network security groups** — control traffic to the DB System at the VNIC level with `networkDetails.nsgIds`.
4. **Enable the reader endpoint for multi-node systems** — distributes read queries across replicas, reducing load on the primary.
5. **Set a maintenance window** — schedule `maintenanceWindowStart` during off-peak hours to minimize impact from OCI-managed patches.
6. **Use `valueFrom` for composability** — reference compartments, subnets, and NSGs from other OpenMCF resources instead of hardcoding OCIDs. This enables stack-level dependency tracking.
7. **Size `instancesDetails` to match `instanceCount`** — when using per-instance pinning, the list length must equal the instance count. Both are immutable after creation.

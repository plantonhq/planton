# OciBlockVolume Examples

## Minimal Development Volume

A 50 GB volume with default Balanced performance and Oracle-managed encryption:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciBlockVolume
metadata:
  name: dev-data
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciBlockVolume.dev-data
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:US-ASHBURN-AD-1"
  sizeInGbs: 50
```

## Higher Performance with KMS Encryption

A 500 GB volume at 20 VPUs/GB with customer-managed encryption for a production database:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciBlockVolume
metadata:
  name: db-data
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciBlockVolume.db-data
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:US-ASHBURN-AD-1"
  displayName: "db-data-prod"
  sizeInGbs: 500
  vpusPerGb: 20
  kmsKeyId:
    value: "ocid1.key.oc1..example"
```

## Autotune with Scheduled Backups

A volume that automatically reduces VPUs when detached and restores them on re-attach. A Gold backup policy provides daily incremental and weekly full backups:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciBlockVolume
metadata:
  name: app-storage
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciBlockVolume.app-storage
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:US-ASHBURN-AD-1"
  sizeInGbs: 200
  autotunePolicies:
    - autotuneType: detached_volume
  backupPolicyId:
    value: "ocid1.volumebackuppolicy.oc1..gold-example"
```

## Performance-Based Autotune with Cross-Region Replica

A 1 TB production volume with performance-based autotune (scales up to 60 VPUs/GB under load) and a cross-region replica for disaster recovery. Uses `valueFrom` to reference a compartment managed by OpenMCF:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciBlockVolume
metadata:
  name: critical-data
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciBlockVolume.critical-data
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  availabilityDomain: "Uocm:US-ASHBURN-AD-1"
  sizeInGbs: 1024
  vpusPerGb: 20
  kmsKeyId:
    value: "ocid1.key.oc1..example"
  autotunePolicies:
    - autotuneType: performance_based
      maxVpusPerGb: 60
  blockVolumeReplicas:
    - availabilityDomain: "Uocm:US-PHOENIX-AD-1"
      displayName: "critical-data-dr-phx"
      xrrKmsKeyId:
        value: "ocid1.key.oc1..phx-example"
```

## Ultra High Performance with Reservations

A 2 TB Ultra High Performance volume for Oracle RAC shared storage with SCSI reservations and cross-region backup encryption:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciBlockVolume
metadata:
  name: rac-shared
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciBlockVolume.rac-shared
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:US-ASHBURN-AD-1"
  sizeInGbs: 2048
  vpusPerGb: 120
  kmsKeyId:
    value: "ocid1.key.oc1..example"
  isReservationsEnabled: true
  backupPolicyId:
    value: "ocid1.volumebackuppolicy.oc1..gold-example"
  xrcKmsKeyId:
    value: "ocid1.key.oc1..xrc-example"
```

## Common Operations

### Change the performance tier

Update `vpusPerGb` to the desired value and re-apply. OCI adjusts performance online without detaching the volume. Transitioning from 0 (Lower Cost) to a higher tier may take several minutes.

### Add an autotune policy

Append a new entry to the `autotunePolicies` list and re-apply. Both `detached_volume` and `performance_based` policies can coexist on the same volume.

### Add a cross-region replica

Append a new entry to `blockVolumeReplicas` with the target availability domain and optional encryption key, then re-apply. Initial replication may take time proportional to volume size.

### Assign a backup policy

Set `backupPolicyId` to the OCID of an Oracle-defined or custom backup policy and re-apply. To find Oracle-defined policies, run `oci bv volume-backup-policy list`.

## Best Practices

1. **Set `sizeInGbs` explicitly** — prevents accidental creation at OCI's 1 TB default.
2. **Use autotune policies for cost control** — `detached_volume` avoids paying for high VPUs on volumes that sit idle.
3. **Match performance to workload** — databases benefit from 20+ VPUs/GB; log storage is fine at 0-10.
4. **Use cross-region replicas for DR** — faster failover than restoring from backups; encrypt replicas with dedicated KMS keys.
5. **Use `valueFrom` references** for `compartmentId` and `kmsKeyId` to avoid hardcoding OCIDs and maintain dependency ordering.
6. **Assign backup policies for production volumes** — automated backups prevent data loss from accidental deletion or corruption.

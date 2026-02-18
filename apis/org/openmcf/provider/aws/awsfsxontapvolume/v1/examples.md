# AwsFsxOntapVolume Examples

## 1. Minimal Volume

The simplest possible ONTAP volume — just the parent SVM, a name, and a size. The volume is created but unmounted (no junction path).

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapVolume
metadata:
  name: minimal-volume
  id: awsfxov-min001
  org: my-org
  env: dev
spec:
  region: us-east-1
  storage_virtual_machine_id:
    value: svm-0123456789abcdef0
  name: vol_minimal
  size_in_megabytes: 1024
```

## 2. NFS Data Volume with Tiering

A production NFS volume with AUTO tiering, storage efficiency, and a custom snapshot policy. Mounted at `/data` for client access.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapVolume
metadata:
  name: prod-nfs-data
  id: awsfxov-nfs001
  org: my-org
  env: prod
spec:
  region: us-east-1
  storage_virtual_machine_id:
    value: svm-0123456789abcdef0
  name: vol_prod_data
  size_in_megabytes: 512000
  junction_path: /data
  security_style: UNIX
  snapshot_policy: default
  storage_efficiency_enabled: true
  copy_tags_to_backups: true
  tiering_policy:
    name: AUTO
    cooling_period: 45
```

## 3. SMB Volume with NTFS Security

A Windows-compatible volume for SMB file shares. Requires the parent SVM to have Active Directory configured.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapVolume
metadata:
  name: windows-share
  id: awsfxov-smb001
  org: my-org
  env: prod
spec:
  region: us-east-1
  storage_virtual_machine_id:
    value: svm-0123456789abcdef0
  name: vol_windows_share
  size_in_megabytes: 204800
  junction_path: /shares/finance
  security_style: NTFS
  storage_efficiency_enabled: true
  tiering_policy:
    name: SNAPSHOT_ONLY
```

## 4. SnapLock Compliance Volume

A WORM-enabled volume for regulatory compliance. Files are automatically committed to immutable state after 24 hours of inactivity. Retention bounds enforce 1-10 year retention with a 5-year default.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapVolume
metadata:
  name: sec-compliance
  id: awsfxov-worm001
  org: my-org
  env: prod
spec:
  region: us-east-1
  storage_virtual_machine_id:
    value: svm-0123456789abcdef0
  name: vol_sec17a4
  size_in_megabytes: 1048576
  junction_path: /compliance/records
  security_style: UNIX
  storage_efficiency_enabled: true
  tiering_policy:
    name: SNAPSHOT_ONLY
  snaplock_configuration:
    snaplock_type: COMPLIANCE
    autocommit_period:
      type: DAYS
      value: 1
    retention_period:
      default_retention:
        type: YEARS
        value: 5
      minimum_retention:
        type: YEARS
        value: 1
      maximum_retention:
        type: YEARS
        value: 10
```

## 5. Cross-Resource Reference (valueFrom)

A volume that references its parent SVM via `valueFrom`, enabling dependency wiring in infra charts.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapVolume
metadata:
  name: referenced-volume
  id: awsfxov-ref001
  org: my-org
  env: prod
spec:
  region: us-east-1
  storage_virtual_machine_id:
    valueFrom:
      kind: AwsFsxOntapStorageVirtualMachine
      metadata:
        id: awsfxosvm-prod001
      fieldPath: status.outputs.svm_id
  name: vol_app_data
  size_in_megabytes: 102400
  junction_path: /app
  security_style: UNIX
  storage_efficiency_enabled: true
  tiering_policy:
    name: AUTO
    cooling_period: 31
```

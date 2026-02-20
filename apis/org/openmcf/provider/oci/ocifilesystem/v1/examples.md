# OciFileSystem Examples

## Minimal File System

A file system with one export and default NFS access:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciFileSystem
metadata:
  name: dev-storage
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciFileSystem.dev-storage
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:US-ASHBURN-AD-1"
  mountTarget:
    subnetId:
      value: "ocid1.subnet.oc1..example"
  exports:
    - path: "/data"
```

Mount on a compute instance:

```shell
mount -t nfs <mount_target_ip>:/data /mnt/data
```

## File System with Fixed IP and DNS

A file system with a predictable mount target address for static NFS client configurations:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciFileSystem
metadata:
  name: static-nfs
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.OciFileSystem.static-nfs
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:US-ASHBURN-AD-1"
  displayName: "Static NFS"
  mountTarget:
    subnetId:
      value: "ocid1.subnet.oc1..example"
    displayName: "static-nfs-mt"
    hostnameLabel: "staticnfs"
    ipAddress: "10.0.1.50"
  exports:
    - path: "/shared"
```

Mount using hostname (if VCN DNS is configured):

```shell
mount -t nfs staticnfs.privatesubnet.myvnc.oraclevcn.com:/shared /mnt/shared
```

## Encrypted File System with Snapshot Policy

A file system with customer-managed KMS encryption and automated snapshots:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciFileSystem
metadata:
  name: encrypted-fs
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciFileSystem.encrypted-fs
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:US-ASHBURN-AD-1"
  displayName: "Encrypted FS"
  kmsKeyId:
    value: "ocid1.key.oc1..example"
  filesystemSnapshotPolicyId:
    value: "ocid1.filesystemsnapshotpolicy.oc1..example"
  mountTarget:
    subnetId:
      value: "ocid1.subnet.oc1..example"
  exports:
    - path: "/secure-data"
      exportOptions:
        - source: "10.0.0.0/16"
          access: read_write
          identitySquash: root_squash
          requirePrivilegedSourcePort: true
          anonymousUid: 65534
          anonymousGid: 65534
```

## Multi-Export with Per-Source Access Control

A production file system with separate exports for different teams, each with distinct NFS permissions:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciFileSystem
metadata:
  name: prod-shared
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciFileSystem.prod-shared
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:US-ASHBURN-AD-1"
  displayName: "Production Shared"
  mountTarget:
    subnetId:
      value: "ocid1.subnet.oc1..example"
    displayName: "prod-shared-mt"
    nsgIds:
      - value: "ocid1.networksecuritygroup.oc1..example"
    requestedThroughput: 1024
    maxFsStatBytes: 1099511627776
    maxFsStatFiles: 10000000
  exports:
    - path: "/engineering"
      exportOptions:
        - source: "10.0.1.0/24"
          access: read_write
          identitySquash: root_squash
          requirePrivilegedSourcePort: true
          anonymousUid: 65534
          anonymousGid: 65534
        - source: "10.0.2.0/24"
          access: read_only
          identitySquash: all_squash
          anonymousUid: 65534
          anonymousGid: 65534
    - path: "/analytics"
      exportOptions:
        - source: "10.0.3.0/24"
          access: read_write
          identitySquash: no_squash
    - path: "/public-reports"
      exportOptions:
        - source: "0.0.0.0/0"
          access: read_only
          identitySquash: all_squash
          isAnonymousAccessAllowed: true
          anonymousUid: 65534
          anonymousGid: 65534
```

## Foreign Key References

Reference OpenMCF-managed compartment, subnet, and NSG resources:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciFileSystem
metadata:
  name: ref-fs
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciFileSystem.ref-fs
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  availabilityDomain: "Uocm:US-ASHBURN-AD-1"
  mountTarget:
    subnetId:
      valueFrom:
        kind: OciSubnet
        name: private-subnet
        fieldPath: status.outputs.subnetId
    nsgIds:
      - valueFrom:
          kind: OciSecurityGroup
          name: nfs-nsg
          fieldPath: status.outputs.networkSecurityGroupId
  exports:
    - path: "/shared"
```

## Common Operations

### Mount the file system from a compute instance

After deployment, use the `mount_target_ip_address` output:

```shell
sudo mkdir -p /mnt/shared
sudo mount -t nfs <mount_target_ip_address>:/shared /mnt/shared
```

For persistent mounts, add to `/etc/fstab`:

```
<mount_target_ip_address>:/shared /mnt/shared nfs defaults 0 0
```

### Add a new export to an existing file system

Append a new entry to the `exports` list and re-apply. Each export creates an independent `oci_file_storage_export` resource — existing exports are not affected.

### Change export access control

Modify the `exportOptions` for an existing export and re-apply. OCI updates the export options in place.

### Scale throughput

Set or update `mountTarget.requestedThroughput` to change the provisioned throughput tier.

## Best Practices

1. **Use `root_squash` for production exports** — prevents NFS clients from operating as root on the file system.
2. **Assign a fixed IP** (`mountTarget.ipAddress`) when clients use static NFS mount configurations.
3. **Use NSGs** to restrict NFS traffic to expected source CIDRs at the network level, in addition to export options.
4. **Set `maxFsStatBytes`** when you need NFS clients to see a specific capacity limit (e.g. quota enforcement via statfs).
5. **Use `valueFrom` references** for `compartmentId` and `subnetId` to avoid hardcoding OCIDs and maintain dependency ordering.
6. **One mount target per file system** — this component creates a dedicated mount target. If you need to share a mount target across file systems, manage the mount target separately.

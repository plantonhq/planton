# GcpFilestoreInstance Examples

## Minimal BASIC_SSD Instance

The simplest configuration — a 2.5 TiB SSD-backed NFS instance on the default network:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpFilestoreInstance
metadata:
  name: dev-nfs
spec:
  projectId: my-project
  instanceName: dev-nfs
  location: us-central1-a
  tier: BASIC_SSD
  fileShare:
    name: vol1
    capacityGb: 2560
  networkConfig:
    network: default
```

## Cost-Effective HDD Storage

A 1 TiB HDD-backed instance for infrequently accessed data:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpFilestoreInstance
metadata:
  name: archive-nfs
spec:
  projectId: my-project
  instanceName: archive-storage
  location: us-east1-b
  tier: STANDARD
  description: Archive storage for historical data
  fileShare:
    name: archive
    capacityGb: 1024
  networkConfig:
    network: my-vpc
```

## Enterprise HA with Deletion Protection

A production-grade instance with regional HA, private networking, and safety guards:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpFilestoreInstance
metadata:
  name: prod-nfs
spec:
  projectId: my-prod-project
  instanceName: prod-nfs-ha
  location: us-central1
  tier: ENTERPRISE
  description: Production shared NFS for application data
  deletionProtectionEnabled: true
  deletionProtectionReason: "critical production data"
  fileShare:
    name: data
    capacityGb: 2048
    nfsExportOptions:
      - ipRanges:
          - "10.0.0.0/8"
        accessMode: READ_WRITE
        squashMode: ROOT_SQUASH
  networkConfig:
    network: prod-vpc
    connectMode: PRIVATE_SERVICE_ACCESS
```

## CMEK-Encrypted Instance

An instance with customer-managed encryption keys:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpFilestoreInstance
metadata:
  name: encrypted-nfs
spec:
  projectId: my-project
  instanceName: encrypted-nfs
  location: us-central1-a
  tier: ZONAL
  kmsKeyName: projects/my-project/locations/us-central1/keyRings/my-ring/cryptoKeys/my-key
  fileShare:
    name: secure_data
    capacityGb: 1024
  networkConfig:
    network: my-vpc
    connectMode: PRIVATE_SERVICE_ACCESS
```

## High-Performance with Fixed IOPS

A zonal instance with guaranteed IOPS for demanding workloads:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpFilestoreInstance
metadata:
  name: perf-nfs
spec:
  projectId: my-project
  instanceName: render-nfs
  location: us-west1-a
  tier: ZONAL
  fileShare:
    name: renders
    capacityGb: 5120
  networkConfig:
    network: my-vpc
  performanceConfig:
    fixedIops:
      maxIops: 30000
```

## NFSv4.1 with Multiple Export Options

An instance using NFSv4.1 with different access levels per subnet:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpFilestoreInstance
metadata:
  name: v4-nfs
spec:
  projectId: my-project
  instanceName: nfsv4-server
  location: us-east1-c
  tier: ZONAL
  protocol: NFS_V4_1
  fileShare:
    name: shared
    capacityGb: 1024
    nfsExportOptions:
      - ipRanges:
          - "10.0.1.0/24"
        accessMode: READ_WRITE
        squashMode: ROOT_SQUASH
      - ipRanges:
          - "10.0.2.0/24"
        accessMode: READ_ONLY
        squashMode: NO_ROOT_SQUASH
  networkConfig:
    network: my-vpc
```

## Infra-Chart Composition with valueFrom

An instance wired to other OpenMCF resources via foreign key references:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpFilestoreInstance
metadata:
  name: composed-nfs
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
  instanceName: composed-nfs
  location: us-central1-a
  tier: BASIC_SSD
  kmsKeyName:
    valueFrom:
      kind: GcpKmsKey
      name: my-key
  fileShare:
    name: shared
    capacityGb: 2560
  networkConfig:
    network:
      valueFrom:
        kind: GcpVpc
        name: my-vpc
    connectMode: PRIVATE_SERVICE_ACCESS
    reservedIpRange: "10.10.0.0/29"
```

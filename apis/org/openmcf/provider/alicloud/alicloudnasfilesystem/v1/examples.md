# Examples

## Minimal NFS File System

Creates a standard NAS file system with NFS protocol and Performance storage. No custom access rules -- all VPC IPs get full read-write access via the default VPC access group.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudNasFileSystem
metadata:
  name: shared-data
spec:
  region: cn-hangzhou
  protocolType: NFS
  storageType: Performance
  vpcId: vpc-abc123
  vswitchId: vsw-abc123
```

## Production NFS with Encryption and Custom Access Rules

A production-grade file system with NAS-managed encryption and restrictive access rules. Only the application subnet gets read-write access, while the monitoring subnet gets read-only access with root squashing.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudNasFileSystem
metadata:
  name: prod-shared-storage
  org: my-org
  env: production
spec:
  region: cn-shanghai
  protocolType: NFS
  storageType: Performance
  description: Production shared storage for microservices
  encryption:
    encryptType: 1
  vpcId: vpc-prod-001
  vswitchId: vsw-prod-001
  accessRules:
    - sourceCidrIp: "10.0.1.0/24"
      rwAccessType: RDWR
      userAccessType: no_squash
      priority: 1
    - sourceCidrIp: "10.0.2.0/24"
      rwAccessType: RDONLY
      userAccessType: root_squash
      priority: 10
  resourceGroupId: rg-prod-123
  tags:
    team: platform
    costCenter: engineering
```

## Extreme NAS for High-Throughput Workloads

An extreme NAS file system with 500 GiB pre-allocated capacity, advanced storage tier, and KMS encryption. Suitable for ML training, media processing, or HPC workloads requiring dedicated throughput.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudNasFileSystem
metadata:
  name: hpc-scratch
  env: production
spec:
  region: cn-hangzhou
  fileSystemType: extreme
  protocolType: NFS
  storageType: advance
  description: High-throughput scratch storage for ML training
  capacity: 500
  zoneId: cn-hangzhou-a
  encryption:
    encryptType: 2
    kmsKeyId: "cmk-abc123def456"
  vpcId: vpc-hpc-001
  vswitchId: vsw-hpc-001
  accessRules:
    - sourceCidrIp: "10.0.0.0/8"
      rwAccessType: RDWR
      userAccessType: no_squash
  tags:
    workload: ml-training
    tier: compute
```

## Cost-Effective Capacity NAS for Archival Data

A Capacity-tier file system for warm/cold data that doesn't require high IOPS. Uses SMB protocol for Windows client compatibility.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudNasFileSystem
metadata:
  name: archive-share
spec:
  region: cn-beijing
  protocolType: SMB
  storageType: Capacity
  description: Archival file share for Windows workstations
  vpcId: vpc-office-001
  vswitchId: vsw-office-001
  accessRules:
    - sourceCidrIp: "192.168.1.0/24"
      rwAccessType: RDWR
      userAccessType: all_squash
```

# AWS OpenSearch Domain Pulumi Module Overview

This document provides a comprehensive overview of the Pulumi module architecture for deploying Amazon OpenSearch Service domains.

## Table of Contents

- [Module Architecture](#module-architecture)
- [Resource Flow](#resource-flow)
- [Implementation Details](#implementation-details)
- [Spec to Resource Mapping](#spec-to-resource-mapping)
- [Security and Networking](#security-and-networking)
- [Data Tiers](#data-tiers)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Module Architecture

The Pulumi module is structured into specialized files, each handling a specific aspect of OpenSearch domain deployment:

```
iac/pulumi/
├── main.go                     # Entry point for Pulumi program
├── Pulumi.yaml                 # Pulumi project configuration
├── Makefile                    # Deployment automation targets
├── debug.sh                    # Debug helper script
├── README.md                   # User-facing documentation
└── module/
    ├── main.go                 # Orchestration and resource coordination
    ├── locals.go               # Local variables and data transformations
    └── outputs.go              # Output constant definitions
```

### File Responsibilities

#### main.go (Entry Point)
- Initializes the Pulumi runtime
- Loads stack input from OpenMCF manifest
- Invokes the module's `Resources()` function
- Handles program-level errors

#### module/main.go (Orchestration)
- Coordinates creation of all OpenSearch-related resources
- Manages resource dependencies
- Configures AWS provider (explicit credentials or default)
- Exports stack outputs

#### module/locals.go
- Initializes local variables from stack input
- Computes AWS tags from metadata (organization, environment, resource kind, resource ID)
- Holds references to the target manifest and spec for easy access

#### module/outputs.go
- Defines output constant names: `domain_id`, `domain_name`, `domain_arn`, `endpoint`, `dashboard_endpoint`
- Ensures consistent output naming across the module

## Resource Flow

```
User Manifest (YAML)
       ↓
OpenMCF CLI
       ↓
Stack Input (Protobuf)
       ↓
main.go (Pulumi entry)
       ↓
module.Resources()
       ↓
    ┌──┴──┐
    │     │
    ↓     ↓
Provider  Locals
    │     │
    └──┬──┘
       ↓
 OpenSearch Domain
       ↓
   Outputs
       ↓
Stack Outputs (Protobuf)
       ↓
User (status.outputs)
```

### Resource Dependencies

1. **AWS Provider** (created first)
   - Uses explicit credentials from stack input, or default AWS credential chain

2. **OpenSearch Domain** (the primary resource)
   - Depends on: AWS Provider, Locals
   - Configures: cluster topology, EBS storage, encryption, VPC, FGAC, log publishing
   - Exports: domain ID, name, ARN, endpoint, dashboard endpoint

## Spec to Resource Mapping

The module translates the `AwsOpenSearchDomainSpec` protobuf message into a Pulumi `opensearch.Domain` resource. Here's how spec fields map to AWS resources:

### Core Domain

| Spec Field | AWS Resource Attribute | Notes |
|------------|----------------------|-------|
| `engineVersion` | `Domain.EngineVersion` | Format: `OpenSearch_2.11` |
| `encryptAtRestEnabled` | `Domain.EncryptAtRestOptions.Enabled` | ForceNew when disabling |
| `kmsKeyId` | `Domain.EncryptAtRestOptions.KmsKeyId` | ForceNew |
| `nodeToNodeEncryptionEnabled` | `Domain.NodeToNodeEncryptionOptions.Enabled` | — |
| `autoTuneEnabled` | `Domain.AutoTuneOptions.DesiredState` | Maps to `ENABLED_ROLLBACK`/`DISABLED` |
| `autoSoftwareUpdateEnabled` | `Domain.SoftwareUpdateOptions.AutoSoftwareUpdateEnabled` | — |
| `ipAddressType` | `Domain.IpAddressType` | `ipv4` or `dualstack` |
| `advancedOptions` | `Domain.AdvancedOptions` | Key-value map passed through |
| `accessPolicies` | `Domain.AccessPolicies` | JSON-serialized IAM policy |

### Cluster Configuration

| Spec Field | AWS Resource Attribute |
|------------|----------------------|
| `clusterConfig.instanceType` | `ClusterConfig.InstanceType` |
| `clusterConfig.instanceCount` | `ClusterConfig.InstanceCount` |
| `clusterConfig.dedicatedMasterEnabled` | `ClusterConfig.DedicatedMasterEnabled` |
| `clusterConfig.dedicatedMasterType` | `ClusterConfig.DedicatedMasterType` |
| `clusterConfig.dedicatedMasterCount` | `ClusterConfig.DedicatedMasterCount` |
| `clusterConfig.zoneAwarenessEnabled` | `ClusterConfig.ZoneAwarenessEnabled` |
| `clusterConfig.availabilityZoneCount` | `ClusterConfig.ZoneAwarenessConfig.AvailabilityZoneCount` |
| `clusterConfig.warmEnabled` | `ClusterConfig.WarmEnabled` |
| `clusterConfig.warmType` | `ClusterConfig.WarmType` |
| `clusterConfig.warmCount` | `ClusterConfig.WarmCount` |
| `clusterConfig.coldStorageEnabled` | `ClusterConfig.ColdStorageOptions.Enabled` |
| `clusterConfig.multiAzWithStandbyEnabled` | `ClusterConfig.MultiAzWithStandbyEnabled` |

### EBS Options

| Spec Field | AWS Resource Attribute |
|------------|----------------------|
| `ebsOptions.ebsEnabled` | `EbsOptions.EbsEnabled` |
| `ebsOptions.volumeType` | `EbsOptions.VolumeType` |
| `ebsOptions.volumeSize` | `EbsOptions.VolumeSize` |
| `ebsOptions.iops` | `EbsOptions.Iops` |
| `ebsOptions.throughput` | `EbsOptions.Throughput` |

### VPC Options

| Spec Field | AWS Resource Attribute |
|------------|----------------------|
| `vpcOptions.subnetIds` | `VpcOptions.SubnetIds` |
| `vpcOptions.securityGroupIds` | `VpcOptions.SecurityGroupIds` |

### Domain Endpoint Options

| Spec Field | AWS Resource Attribute |
|------------|----------------------|
| `domainEndpointOptions.enforceHttps` | `DomainEndpointOptions.EnforceHttps` |
| `domainEndpointOptions.tlsSecurityPolicy` | `DomainEndpointOptions.TlsSecurityPolicy` |
| `domainEndpointOptions.customEndpointEnabled` | `DomainEndpointOptions.CustomEndpointEnabled` |
| `domainEndpointOptions.customEndpoint` | `DomainEndpointOptions.CustomEndpoint` |
| `domainEndpointOptions.customEndpointCertificateArn` | `DomainEndpointOptions.CustomEndpointCertificateArn` |

### Advanced Security Options (FGAC)

| Spec Field | AWS Resource Attribute |
|------------|----------------------|
| `advancedSecurityOptions.enabled` | `AdvancedSecurityOptions.Enabled` |
| `advancedSecurityOptions.internalUserDatabaseEnabled` | `AdvancedSecurityOptions.InternalUserDatabaseEnabled` |
| `advancedSecurityOptions.masterUserArn` | `AdvancedSecurityOptions.MasterUserOptions.MasterUserArn` |
| `advancedSecurityOptions.masterUserName` | `AdvancedSecurityOptions.MasterUserOptions.MasterUserName` |
| `advancedSecurityOptions.masterUserPassword` | `AdvancedSecurityOptions.MasterUserOptions.MasterUserPassword` |

### Log Publishing Options

Each entry in `logPublishingOptions` maps to a `Domain.LogPublishingOptions` map entry keyed by `logType`:

| Spec Field | AWS Resource Attribute |
|------------|----------------------|
| `logPublishingOptions[].logType` | Map key (e.g., `INDEX_SLOW_LOGS`) |
| `logPublishingOptions[].cloudwatchLogGroupArn` | `LogPublishingOptions[type].CloudwatchLogGroupArn` |
| `logPublishingOptions[].enabled` | `LogPublishingOptions[type].Enabled` |

## Security and Networking

### Encryption Layers

The module configures three independent encryption layers:

1. **At-rest encryption** (`encryptAtRestEnabled`) — Encrypts EBS volumes, automated snapshots, UltraWarm/cold data
2. **Node-to-node encryption** (`nodeToNodeEncryptionEnabled`) — TLS for inter-node traffic
3. **In-transit encryption** (`domainEndpointOptions.enforceHttps`) — HTTPS for client-to-domain traffic

### VPC Deployment

When `vpcOptions` is set:
- OpenSearch creates ENIs in the specified subnets
- The domain endpoint resolves to a VPC-internal IP
- Security groups control which sources can reach port 443
- The domain is not accessible from the internet

### Fine-Grained Access Control

When `advancedSecurityOptions.enabled` is true:
- Internal user database provides username/password authentication
- IAM master user provides role-based authentication
- OpenSearch Dashboards supports role-based index and field-level permissions
- Audit logging can be enabled for compliance tracking

## Data Tiers

The module supports OpenSearch's three-tier storage model:

### Hot Tier (Data Nodes)
- Configured via `clusterConfig.instanceType`, `instanceCount`
- Storage via `ebsOptions`
- Active indexing and frequent queries

### Warm Tier (UltraWarm)
- Configured via `clusterConfig.warmEnabled`, `warmType`, `warmCount`
- S3-backed storage, read-only
- Infrequently accessed data

### Cold Tier
- Configured via `clusterConfig.coldStorageEnabled`
- Requires warm tier to be enabled
- S3-backed, detached from cluster compute
- Attach to warm tier on demand for querying

## Best Practices

### Production Deployment

```yaml
clusterConfig:
  instanceType: r6g.large.search
  instanceCount: 3
  dedicatedMasterEnabled: true
  dedicatedMasterType: r6g.large.search
  dedicatedMasterCount: 3
  zoneAwarenessEnabled: true
  availabilityZoneCount: 3
ebsOptions:
  ebsEnabled: true
  volumeType: gp3
  volumeSize: 100
encryptAtRestEnabled: true
nodeToNodeEncryptionEnabled: true
domainEndpointOptions:
  enforceHttps: true
  tlsSecurityPolicy: "Policy-Min-TLS-1-2-PFS-2023-10"
advancedSecurityOptions:
  enabled: true
autoTuneEnabled: true
```

### Cost Optimization

- Use `t3.small.search` or `t3.medium.search` for development
- Enable UltraWarm for data older than 14 days
- Enable cold storage for data older than 90 days
- Use gp3 (pay for what you need) instead of io1

### Monitoring

- Enable log publishing for `ES_APPLICATION_LOGS` at minimum
- Add `INDEX_SLOW_LOGS` and `SEARCH_SLOW_LOGS` for performance troubleshooting
- Enable `AUDIT_LOGS` when FGAC is active for compliance

## Troubleshooting

### Common Issues

#### 1. Domain Creation Timeout

**Symptom**: Pulumi times out waiting for domain to become active

**Causes**:
- OpenSearch domain creation takes 15-30 minutes (normal)
- VPC endpoint creation adds time
- Complex cluster configuration (dedicated masters, warm nodes)

**Solutions**:
- Increase Pulumi timeout settings
- Check VPC and subnet configurations
- Verify security group allows management traffic

#### 2. VPC Configuration Error

**Symptom**: Error about subnet or security group configuration

**Cause**: Subnet/security group IDs don't exist or are in wrong VPC

**Solution**:
- Verify subnets exist and are in different AZs
- Verify security groups belong to the same VPC as subnets
- Check IAM permissions for ENI creation in VPC

#### 3. FGAC Master User Error

**Symptom**: Error about master user configuration

**Cause**: Both `masterUserArn` and `masterUserName` specified, or neither

**Solution**:
```yaml
# Option A: Internal user database
advancedSecurityOptions:
  enabled: true
  internalUserDatabaseEnabled: true
  masterUserName: admin
  masterUserPassword:
    value: "MyStr0ng!Pass#"

# Option B: IAM master user
advancedSecurityOptions:
  enabled: true
  internalUserDatabaseEnabled: false
  masterUserArn:
    value: "arn:aws:iam::123456789012:role/opensearch-admin"
```

#### 4. Encryption Key Error

**Symptom**: Error about KMS key when `encryptAtRestEnabled` is true

**Cause**: `kmsKeyId` references a key the OpenSearch service can't access

**Solution**:
- Omit `kmsKeyId` to use the AWS-managed `aws/es` key
- Or ensure the KMS key policy allows the `es.amazonaws.com` service principal

### Debugging

**Enable Debug Output:**
```bash
export PULUMI_DEBUG_COMMANDS=true
pulumi up --logtostderr -v=9 2>&1 | tee pulumi-debug.log
```

**Check Domain Status:**
```bash
aws opensearch describe-domain \
  --domain-name my-search \
  --query 'DomainStatus.Processing'
```

**Check Domain Configuration:**
```bash
aws opensearch describe-domain-config \
  --domain-name my-search
```

## References

- [AWS OpenSearch Service Documentation](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/)
- [Pulumi AWS OpenSearch Domain](https://www.pulumi.com/registry/packages/aws/api-docs/opensearch/domain/)
- [OpenSearch Best Practices](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/bp.html)
- [OpenSearch Instance Types](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/supported-instance-types.html)

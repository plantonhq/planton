# AWS Serverless ElastiCache

Deploys an AWS ElastiCache Serverless cache with consumption-based pricing and automatic scaling of both compute (ECPU) and storage (GB). Supports Redis, Valkey, and Memcached engines with configurable scaling limits, VPC networking, encryption, snapshots, and Redis ACL authentication.

## What Gets Created

When you deploy an AwsServerlessElasticache resource, OpenMCF provisions:

- **Serverless Cache** — an `aws_elasticache_serverless_cache` resource using the specified engine (Redis, Valkey, or Memcached), with AWS managing all node scaling, replication, and patching automatically
- **Cache Usage Limits** — optional minimum and maximum bounds for data storage (GB) and compute (ECPU/s) that constrain the auto-scaling range
- **VPC Endpoints** — the cache creates endpoints in the specified subnets, with traffic controlled by the attached security groups
- **At-Rest Encryption** — uses the AWS-managed key by default, or a customer-managed KMS key when `kmsKeyId` is provided
- **Automatic Snapshots** — daily snapshots at the configured time with configurable retention (Redis/Valkey only)
- **AWS Resource Tags** — organization, environment, resource kind, and resource ID tags applied to the cache

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **A VPC with subnets** where the serverless cache endpoints will be placed
- **A security group** allowing inbound traffic on the cache port (default 6379 for Redis/Valkey, 11211 for Memcached)
- **A KMS key** if using customer-managed at-rest encryption
- **A Redis ACL user group** if using fine-grained access control (Redis/Valkey only)

## Quick Start

Create a file `serverless-cache.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsServerlessElasticache
metadata:
  name: my-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsServerlessElasticache.my-cache
spec:
  engine: redis
  subnetIds:
    - subnet-0a1b2c3d4e5f00001
    - subnet-0a1b2c3d4e5f00002
  securityGroupIds:
    - sg-0a1b2c3d4e5f00001
```

Deploy:

```shell
openmcf apply -f serverless-cache.yaml
```

This creates a Redis Serverless cache with AWS-managed scaling defaults, placed in two subnets.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `engine` | `string` | Cache engine to use. Values: `redis`, `valkey`, `memcached`. Switching between Redis and Valkey is in-place; switching to/from Memcached forces recreation. | Must be `redis`, `valkey`, or `memcached` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `majorEngineVersion` | `string` | Provider default | Major engine version. Examples: `7`, `8` for Redis/Valkey; `1.6` for Memcached. |
| `description` | `string` | — | Human-readable description of the serverless cache. |
| `dataStorageMaxGb` | `int` | AWS default | Maximum data storage in GB. AWS auto-scales up to this limit. Range: 1–5000. |
| `dataStorageMinGb` | `int` | AWS default | Minimum data storage in GB. AWS guarantees at least this capacity. Range: 1–5000. Must not exceed `dataStorageMaxGb`. |
| `ecpuMax` | `int` | AWS default | Maximum ElastiCache Processing Units per second. Range: 1000–15000000. |
| `ecpuMin` | `int` | AWS default | Minimum ElastiCache Processing Units per second. Range: 1000–15000000. Must not exceed `ecpuMax`. |
| `subnetIds` | `StringValueOrRef[]` | `[]` | Subnet IDs for the cache's VPC endpoints. **ForceNew** — changing this destroys and recreates the cache. Can reference `AwsVpc` via `valueFrom`. |
| `securityGroupIds` | `StringValueOrRef[]` | `[]` | Security group IDs to attach to the cache endpoint. Can reference `AwsSecurityGroup` via `valueFrom`. |
| `kmsKeyId` | `StringValueOrRef` | AWS-managed key | Customer-managed KMS key ARN for at-rest encryption. **ForceNew** — changing this destroys and recreates the cache. Can reference `AwsKmsKey` via `valueFrom`. |
| `dailySnapshotTime` | `string` | — | Daily automatic snapshot time in UTC, format `HH:mm` (e.g., `05:00`). Redis/Valkey only. |
| `snapshotRetentionLimit` | `int` | `0` | Number of days to retain automatic snapshots. Range: 0–35. 0 disables snapshots. Redis/Valkey only. |
| `userGroupId` | `string` | — | Redis ACL user group ID for fine-grained access control. Redis/Valkey only. |

## Examples

### Redis with Scaling Limits

A Redis serverless cache with explicit storage and compute boundaries:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsServerlessElasticache
metadata:
  name: session-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsServerlessElasticache.session-cache
spec:
  engine: redis
  majorEngineVersion: "7"
  description: Session store for web application
  dataStorageMinGb: 1
  dataStorageMaxGb: 10
  ecpuMin: 1000
  ecpuMax: 50000
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
  securityGroupIds:
    - sg-redis-cache
```

### Valkey with Snapshots and Encryption

A Valkey serverless cache with daily snapshots, customer-managed encryption, and Redis ACL authentication:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsServerlessElasticache
metadata:
  name: prod-kv-store
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsServerlessElasticache.prod-kv-store
spec:
  engine: valkey
  majorEngineVersion: "8"
  description: Production key-value store
  dataStorageMinGb: 5
  dataStorageMaxGb: 100
  ecpuMin: 5000
  ecpuMax: 500000
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
    - subnet-private-az3
  securityGroupIds:
    - sg-prod-cache
  kmsKeyId: arn:aws:kms:us-east-1:123456789012:key/abcd-1234-efgh-5678
  dailySnapshotTime: "05:00"
  snapshotRetentionLimit: 7
  userGroupId: my-redis-acl-group
```

### Memcached for Volatile Caching

A Memcached serverless cache for ephemeral data with no persistence or authentication:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsServerlessElasticache
metadata:
  name: html-fragment-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsServerlessElasticache.html-fragment-cache
spec:
  engine: memcached
  majorEngineVersion: "1.6"
  description: HTML fragment cache
  dataStorageMaxGb: 5
  ecpuMax: 10000
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
  securityGroupIds:
    - sg-memcached
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding IDs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsServerlessElasticache
metadata:
  name: ref-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsServerlessElasticache.ref-cache
spec:
  engine: redis
  majorEngineVersion: "7"
  dataStorageMinGb: 1
  dataStorageMaxGb: 50
  subnetIds:
    - valueFrom:
        kind: AwsVpc
        name: my-vpc
        field: status.outputs.private_subnets[0].id
    - valueFrom:
        kind: AwsVpc
        name: my-vpc
        field: status.outputs.private_subnets[1].id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: cache-sg
        field: status.outputs.security_group_id
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: cache-key
      field: status.outputs.key_arn
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `arn` | `string` | Amazon Resource Name of the serverless cache, used in IAM policies and cross-service permissions |
| `endpoint_address` | `string` | Primary connection endpoint DNS address for read-write operations |
| `endpoint_port` | `int` | Port of the primary connection endpoint |
| `reader_endpoint_address` | `string` | Reader endpoint DNS address for distributing read traffic (Redis/Valkey only; empty for Memcached) |
| `reader_endpoint_port` | `int` | Port of the reader endpoint |
| `full_engine_version` | `string` | Exact engine version deployed (e.g., `7.1.0`) |
| `name` | `string` | Name of the serverless cache, matches `metadata.id` |

## Related Components

- [AwsVpc](/docs/catalog/aws/awsvpc) — provides the subnets for cache endpoint placement
- [AwsSecurityGroup](/docs/catalog/aws/awssecuritygroup) — controls network-level access to the cache endpoint
- [AwsKmsKey](/docs/catalog/aws/awskmskey) — provides the customer-managed encryption key for at-rest encryption

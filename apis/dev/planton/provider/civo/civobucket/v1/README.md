# CivoBucket API

## Overview

The `CivoBucket` API provides a declarative way to provision and manage Civo Object Storage buckets using Planton. Civo Object Storage is S3-compatible, offering predictable pricing, collocated storage for low latency, and zero data transfer fees within the Civo platform.

This API abstracts the complexity of bucket provisioning, credential management, and S3 configuration into a simple Protobuf-based specification. You declare what you want (bucket name, region, versioning, tags), and Planton handles the infrastructure provisioning via Pulumi.

## API Structure

The `CivoBucket` resource follows the standard Planton API pattern:

```protobuf
message CivoBucket {
  string api_version = 1;                           // "civo.planton.dev/v1"
  string kind = 2;                                  // "CivoBucket"
  CloudResourceMetadata metadata = 3;               // Name, labels, description
  CivoBucketSpec spec = 4;                          // Bucket configuration
  CivoBucketStatus status = 5;                      // Runtime outputs
}
```

## Specification Fields

### `CivoBucketSpec`

Defines the desired state of your Civo Object Storage bucket:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `bucket_name` | `string` | Yes | DNS-compatible bucket name (3-63 chars, lowercase, alphanumeric + hyphens, no leading/trailing hyphens) |
| `region` | `CivoRegion` | Yes | Civo region (e.g., `LON1`, `NYC1`, `FRA1`) |
| `versioning_enabled` | `bool` | No | Enable versioning to protect against accidental deletion/overwrites (default: `false`) |
| `tags` | `repeated string` | No | Organizational tags (must be unique) |

#### Bucket Name Validation

The `bucket_name` field enforces DNS-compatible naming rules:

- **Length**: 3-63 characters
- **Characters**: Lowercase letters (a-z), numbers (0-9), hyphens (-)
- **Start/End**: Must start and end with alphanumeric characters
- **Pattern**: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`

**Valid Examples**:
- `my-app-storage`
- `prod-backups`
- `dev-assets-123`

**Invalid Examples**:
- `MyBucket` (uppercase)
- `my_bucket` (underscores)
- `-bucket` (starts with hyphen)
- `bucket-` (ends with hyphen)
- `ab` (too short)

#### Region Support

Civo Object Storage is available in multiple regions:

- `LON1` - London, United Kingdom
- `NYC1` - New York City, USA
- `FRA1` - Frankfurt, Germany
- `PHX1` - Phoenix, USA
- `SIN1` - Singapore

**Best Practice**: Choose the region closest to your compute resources to minimize latency and maximize throughput.

#### Versioning

When `versioning_enabled` is `true`, the bucket will protect against accidental deletions and overwrites by maintaining previous versions of objects.

**Note**: Civo Object Storage versioning is configured via the S3 API, not the Civo control plane. The Pulumi module logs a reminder to configure versioning post-deployment using the AWS CLI or SDK pointed at the Civo endpoint:

```bash
aws s3api put-bucket-versioning \
  --bucket my-bucket \
  --versioning-configuration Status=Enabled \
  --endpoint-url https://objectstore.civo.com
```

**Cost Consideration**: Versioning increases storage costs as each version is stored separately.

#### Tags

Tags provide organizational metadata for buckets. They're useful for:
- Environment tracking (`env:prod`, `env:staging`)
- Team ownership (`team:backend`, `team:data`)
- Compliance (`retention:90-days`, `criticality:high`)

**Note**: The Civo Pulumi provider doesn't currently support tags on ObjectStore resources. Tags in the spec are recorded in metadata but not applied to the Civo resource.

## Status and Outputs

### `CivoBucketStatus`

After provisioning, the `status` field contains runtime information:

```protobuf
message CivoBucketStatus {
  CivoBucketStackOutputs outputs = 1;
}
```

### `CivoBucketStackOutputs`

| Field | Type | Description |
|-------|------|-------------|
| `bucket_id` | `string` | Unique identifier (UUID) for the bucket |
| `endpoint_url` | `string` | S3-compatible endpoint URL (e.g., `https://objectstore.civo.com/my-bucket`) |
| `access_key_secret_ref` | `string` | Reference to the secret storing the access key ID |
| `secret_key_secret_ref` | `string` | Reference to the secret storing the secret access key |

These outputs are exported after successful provisioning and can be consumed by applications or other infrastructure resources.

## Quick Start

### Minimal Example

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoBucket
metadata:
  name: my-dev-bucket
spec:
  bucketName: my-dev-bucket
  region: FRA1
```

This creates a private bucket in Frankfurt with no versioning.

### Production Example

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoBucket
metadata:
  name: acme-prod-backups
  description: Production database backups for ACME Corp
spec:
  bucketName: acme-prod-backups
  region: LON1
  versioningEnabled: true
  tags:
    - env:prod
    - criticality:high
    - retention:180-days
```

This creates a versioned bucket in London with production-grade tagging.

## S3 Compatibility

Civo Object Storage is 100% S3-compatible for object operations. You can use:

- **AWS CLI**: Configure with `--endpoint-url https://objectstore.civo.com`
- **AWS SDKs**: Boto3 (Python), AWS SDK (JavaScript, Java, Go, .NET) with endpoint override
- **s3cmd, rclone, MinIO client**: Standard S3-compatible tools
- **Terraform S3 Backend**: Store state in Civo Object Storage

### Example: AWS CLI Usage

```bash
# Configure AWS CLI with Civo credentials
aws configure set aws_access_key_id <access-key>
aws configure set aws_secret_access_key <secret-key>

# List objects
aws s3 ls s3://my-bucket --endpoint-url https://objectstore.civo.com

# Upload file
aws s3 cp file.txt s3://my-bucket/ --endpoint-url https://objectstore.civo.com

# Enable versioning
aws s3api put-bucket-versioning \
  --bucket my-bucket \
  --versioning-configuration Status=Enabled \
  --endpoint-url https://objectstore.civo.com
```

**Important**: Use **path-style addressing** (`https://endpoint.civo.com/bucket/key`) rather than virtual-host style.

## Deployment Workflow

1. **Define the resource**: Create a YAML manifest with your bucket specification
2. **Apply via Planton CLI**: Use `planton apply` to provision
3. **Pulumi provisions resources**:
   - Creates Civo Object Store credential
   - Creates bucket in specified region
   - Exports endpoint and credentials
4. **Configure S3 settings** (if needed): Use AWS CLI/SDK for versioning, policies, lifecycle rules
5. **Consume outputs**: Applications retrieve credentials and endpoint from status

## Security Best Practices

1. **Never hardcode credentials**: Use environment variables, Kubernetes Secrets, or secret managers
2. **Separate credentials per service**: Limit blast radius if a key is compromised
3. **Default to private**: Only make buckets public when necessary
4. **Rotate keys periodically**: Create new credentials, migrate, delete old
5. **Encrypt sensitive data client-side**: Even though Civo encrypts at rest, you control the keys

## Cost Optimization

- **Right-size allocations**: Start with 500 GB, monitor usage, scale in increments
- **Use lifecycle rules**: Auto-delete old logs/backups after retention period
- **Leverage free in-platform transfer**: Run processing jobs on Civo instances in the same region
- **Compress data**: Text logs, JSON, CSVs compress 5-10x

## Monitoring and Operations

- Track bucket usage relative to allocated capacity
- Set alerts when usage exceeds 80% to avoid hitting limits
- Monitor application errors for failed writes (could indicate full bucket)
- Use Civo API to retrieve bucket metrics and capacity information

## Limitations and Known Issues

1. **Versioning Configuration**: Must be configured via S3 API after bucket creation (not through Civo control plane)
2. **Tags**: Civo ObjectStore provider doesn't currently support tags (recorded in metadata only)
3. **Lifecycle Policies**: Must be configured via S3 API (not in provider)
4. **CORS Configuration**: Must be configured via S3 API
5. **Bucket Policies**: Must be configured via S3 API

These limitations reflect the Civo provider's current capabilities. For advanced S3 configurations, use the AWS CLI or SDK with the Civo endpoint.

## Related Documentation

- **Examples**: See [examples.md](./examples.md) for real-world scenarios
- **Research**: See [docs/README.md](./docs/README.md) for deployment methods deep dive
- **IaC Implementation**: See [iac/pulumi/](./iac/pulumi/) for Pulumi module details
- **Civo Docs**: [Civo Object Storage Documentation](https://www.civo.com/docs/object-stores)

## Support and Troubleshooting

### Common Issues

**Issue**: Bucket name already taken
- **Solution**: Bucket names must be globally unique across all Civo customers. Choose a more specific name (e.g., prefix with company name).

**Issue**: Region mismatch between bucket and credentials
- **Solution**: Ensure credentials and bucket are in the same region.

**Issue**: Versioning not enabled after deployment
- **Solution**: Configure versioning via S3 API post-deployment (see Versioning section above).

**Issue**: Tags not visible in Civo dashboard
- **Solution**: Civo ObjectStore provider doesn't support tags. Tags are for organizational purposes in Planton only.

### Getting Help

- **Planton Issues**: [GitHub Issues](https://github.com/plantonhq/planton/issues)
- **Civo Support**: [Civo Support Portal](https://www.civo.com/support)
- **Community**: [Planton Discussions](https://github.com/plantonhq/planton/discussions)

## Version History

- **v1**: Initial release with bucket provisioning, region selection, versioning flag, and tags support


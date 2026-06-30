# DigitalOcean Spaces Bucket

## Overview

The **DigitalOcean Spaces Bucket** API resource provides a simplified, declarative interface for creating and managing object storage buckets on DigitalOcean Spaces. Following Planton's 80/20 principle, this API exposes essential configuration fields for the vast majority of object storage use cases while maintaining S3-compatible access patterns.

DigitalOcean Spaces is an S3-compatible object storage service with built-in CDN integration, offering predictable flat-rate pricing ($5/month per bucket with 250GB storage and 1TB outbound transfer included). Unlike AWS S3's complex pricing tiers and feature matrix, Spaces provides a streamlined experience optimized for common object storage needs: static asset hosting, file uploads, backups, and media storage.

This API resource integrates with Planton's unified infrastructure framework, using a Kubernetes-style manifest format (`apiVersion`, `kind`, `metadata`, `spec`) for consistency across cloud providers.

## Key Features

### Object Storage Essentials

- **S3-Compatible API**: Works with existing S3 tools (AWS CLI, s3cmd, rclone, SDKs)
- **Built-in CDN**: Integrated content delivery network for global edge caching
- **Predictable Pricing**: Flat-rate $5/month per bucket (250GB storage + 1TB transfer)
- **Multiple Regions**: Deploy buckets close to users (NYC, SFO, AMS, SGP, etc.)

### Access Control

- **Private Buckets**: Default access control restricts access to authenticated requests only
- **Public-Read Buckets**: Serve static assets publicly (websites, images, downloads)
- **Signed URLs**: Generate time-limited access URLs for controlled public access
- **Spaces Access Keys**: S3-compatible access keys for programmatic access

### Data Protection

- **Versioning**: Enable object versioning to protect against accidental overwrites and deletes
- **Lifecycle Rules**: Automatically expire old objects to control storage costs
- **CORS Configuration**: Configure cross-origin resource sharing for web applications

### Integration Features

- **CDN Integration**: Automatic CDN endpoint creation for public buckets
- **Custom Domains**: Map custom domains to Spaces buckets via CNAME records
- **Metadata and Tags**: Organize buckets with descriptive tags
- **Foreign Key References**: Reference buckets in other Planton resources

## Use Cases

### Static Website Hosting

Host static websites, SPAs, or JAMstack applications with automatic CDN distribution.

**Ideal for**: React/Vue/Angular frontends, Hugo/Jekyll sites, landing pages

### Media Storage

Store and serve user-uploaded images, videos, and documents with global CDN delivery.

**Ideal for**: User profile pictures, product images, video streaming

### Application Backups

Store database backups, log archives, and disaster recovery snapshots.

**Ideal for**: PostgreSQL dumps, application logs, VM snapshots

### Build Artifacts

Store CI/CD build artifacts, container images, or release packages.

**Ideal for**: npm packages, Docker layers, compiled binaries

## Basic Example

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanBucket
metadata:
  name: my-app-media
spec:
  bucket_name: my-app-media
  region: nyc3
  access_control: PRIVATE
  versioning_enabled: false
```

## Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `bucket_name` | string | Yes | Bucket name (DNS-compatible, 3-63 chars, lowercase alphanumeric and hyphens) |
| `region` | enum | Yes | DigitalOcean region (e.g., `nyc3`, `sfo3`, `ams3`, `sgp1`) |
| `access_control` | enum | No | Access control: `PRIVATE` (default) or `PUBLIC_READ` |
| `versioning_enabled` | bool | No | Enable object versioning (default: false) |
| `tags` | list | No | Tags for bucket organization |

### Access Control Options

| Value | Description | Use Case |
|-------|-------------|----------|
| `PRIVATE` | Bucket and objects are private (default) | Application data, backups, user uploads |
| `PUBLIC_READ` | Bucket and objects are publicly readable | Static websites, public downloads, CDN content |

**Note**: For controlled public access without making entire bucket public, use signed URLs.

### Versioning

When `versioning_enabled: true`:
- Every object update creates a new version
- Previous versions are retained
- Protects against accidental overwrites/deletes
- **Cannot be disabled once enabled** (only suspended)

**Cost impact**: Storage costs increase as versions accumulate. Use lifecycle policies to expire old versions.

## Stack Outputs

After deployment, the following outputs are available:

| Output | Description |
|--------|-------------|
| `bucket_id` | Unique identifier for the bucket (UUID format) |
| `endpoint` | Regional endpoint URL for the bucket (e.g., "https://<region>.digitaloceanspaces.com") |

## Available Regions

Common DigitalOcean Spaces regions:

- `nyc3` - New York (USA)
- `sfo3` - San Francisco (USA)
- `sfo2` - San Francisco (USA)
- `ams3` - Amsterdam (Netherlands)
- `sgp1` - Singapore
- `fra1` - Frankfurt (Germany)
- `blr1` - Bangalore (India)
- `syd1` - Sydney (Australia)

Choose the region closest to your users for lowest latency. Note: Each region requires a separate bucket (no cross-region replication).

## Pricing

**Flat Rate**: $5/month per bucket includes:
- 250GB storage
- 1TB outbound transfer
- Built-in CDN

**Overage**: $0.02/GB storage, $0.01/GB transfer beyond included amounts

**Example costs**:
- Small app (10GB media, 100GB/month transfer): $5/month
- Medium app (500GB storage, 2TB transfer): $15/month ($5 base + $5 storage + $5 transfer)

Compare to AWS S3: DigitalOcean's flat-rate is simpler and often more economical for typical workloads.

## S3 Compatibility

DigitalOcean Spaces implements the core S3 API, enabling use of:

- **AWS CLI**: `aws --endpoint-url https://nyc3.digitaloceanspaces.com s3 ls`
- **s3cmd**: Configure endpoint in `~/.s3cfg`
- **rclone**: Add Spaces as S3-compatible remote
- **SDKs**: AWS SDK for JavaScript/Python/Go/Ruby with custom endpoint

**Supported S3 features**:
- Object CRUD operations
- Multipart uploads
- Bucket/object ACLs
- Versioning
- CORS
- Lifecycle policies

**Not supported**:
- Storage classes (no Glacier equivalent)
- Server-side encryption with customer keys
- Cross-region replication
- Advanced IAM policies

## Examples

For comprehensive examples including:
- Private buckets for application data
- Public buckets for static websites
- Versioned buckets for backups
- Production configurations with tags
- Multi-region strategies

See [examples.md](./examples.md)

## Infrastructure as Code

### Pulumi
See [iac/pulumi/README.md](./iac/pulumi/README.md) for Pulumi-based deployment.

### Terraform
See [iac/tf/README.md](./iac/tf/README.md) for Terraform-based deployment.

## Best Practices

### Access Control
- **Default to private** - Only use `PUBLIC_READ` for truly public content
- **Use signed URLs** for temporary public access instead of making buckets public
- **Rotate access keys** regularly for security

### Versioning
- **Enable for critical data** - Backups, archives, important documents
- **Understand the cost** - Versioning increases storage usage
- **Plan lifecycle policies** - Expire old versions after retention period

### Naming
- **Use DNS-compatible names** - Lowercase, hyphens, no special characters
- **Include environment** - `prod-app-media`, `dev-uploads`
- **Be descriptive** - Name should indicate purpose

### Tags
- **Consistent tagging** - Use standard tags across all buckets
- **Include metadata** - environment, team, cost-center, purpose
- **Enable cost tracking** - Tags help analyze spending

### CDN Usage
- **Enable for public content** - Free CDN reduces latency and transfer costs
- **Set cache headers** - Control CDN caching behavior
- **Use custom domains** - Map `cdn.example.com` to Spaces CDN endpoint

## Troubleshooting

### Bucket Creation Fails with "Name Already Taken"
**Solution**: Bucket names must be globally unique across all DigitalOcean Spaces. Choose a different name.

### Cannot Access Objects
**Solution**: 
- Check bucket ACL (`PRIVATE` requires authentication)
- Verify Spaces access keys are configured correctly
- Confirm objects were uploaded successfully

### Versioning Cannot Be Disabled
**Note**: This is by design. Once versioning is enabled, it can only be suspended (not fully disabled). Plan accordingly before enabling.

### High Transfer Costs
**Solution**:
- Enable CDN to cache content at edge locations
- Review access patterns and optimize object sizes
- Consider lifecycle policies to delete unused objects

## Support

- **Planton**: [github.com/plantonhq/planton](https://github.com/plantonhq/planton)
- **DigitalOcean Spaces Docs**: [docs.digitalocean.com/products/spaces](https://docs.digitalocean.com/products/spaces/)
- **S3 API Reference**: [docs.aws.amazon.com/s3/](https://docs.aws.amazon.com/s3/) (compatible subset)

## References

- [DigitalOcean Spaces Overview](https://docs.digitalocean.com/products/spaces/)
- [S3 API Compatibility](https://docs.digitalocean.com/products/spaces/reference/s3-compatibility/)
- [Spaces CDN](https://docs.digitalocean.com/products/spaces/how-to/enable-cdn/)
- [Research Documentation](./docs/README.md) - Deep dive into deployment methods

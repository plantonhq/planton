

# Cloudflare R2 Bucket

A OpenMCF deployment component for deploying and managing Cloudflare R2 buckets - S3-compatible object storage with **zero egress fees**, strong consistency, and global performance powered by Cloudflare's network.

## Overview

Cloudflare R2 revolutionizes cloud object storage economics by eliminating the egress fees that plague traditional cloud providers. While AWS S3, Google Cloud Storage, and Azure Blob charge $0.08-0.12 per GB transferred out, R2 charges **zero** for egress. You pay only for storage ($0.015/GB-month) and API operations.

This fundamental shift makes R2 ideal for:
- **Content delivery and media hosting** (serve TBs without bandwidth charges)
- **Public datasets and software distribution** (no penalty for popularity)
- **Multi-cloud architectures** (store once, serve anywhere without vendor lock-in)
- **Backup repositories** (retrieve when needed without surprise bills)

R2 is S3-compatible for data operations (GET/PUT/DELETE, multipart uploads, presigned URLs), allowing most S3 tools and SDKs to work with minimal changes. It's simpler than S3 by design—optimized for the 80% use case of storing and serving content, without the complexity of advanced AWS features.

### Key Features

- **Zero Egress Fees**: Serve unlimited data globally at no bandwidth cost
- **S3 API Compatibility**: Works with existing S3 tools (AWS CLI, SDKs, rclone)
- **Strong Consistency**: Immediate read-after-write, no eventual consistency delays
- **Global Distribution**: Backed by Cloudflare's 330+ data center network
- **Geographic Location Hints**: Optimize latency with regional placement
- **Public Access Control**: Optional r2.dev subdomain for public buckets
- **Cloudflare Integration**: Seamless use with Workers, Pages, and CDN

### Use Cases

Use Cloudflare R2 when you need:

- **Escape egress fees**: Reduce cloud storage costs by 90%+ for egress-heavy workloads
- **Multi-cloud portability**: Store once, serve to AWS/GCP/Azure without exit penalties
- **Content delivery**: Serve media, assets, or downloads globally without bandwidth charges
- **Backup and archive**: Long-term storage with free retrieval
- **Hybrid cloud**: Bridge on-premises and cloud without egress lock-in

## Quick Start

### Basic R2 Bucket

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: media-bucket
spec:
  bucketName: my-media-assets
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"  # Your Cloudflare account ID (32 hex chars)
  location: weur  # Western Europe
  publicAccess: false
```

This creates an R2 bucket in Western Europe with private access (authentication required).

## Specification Fields

### Required Fields

- **`bucketName`** (string): DNS-compatible name (3-63 characters, lowercase alphanumeric + hyphens)
  - Must be globally unique within your Cloudflare account
  - Example: `my-app-assets`, `user-uploads-prod`

- **`accountId`** (string): Your Cloudflare account ID (exactly 32 hexadecimal characters)
  - Found in Cloudflare Dashboard → Account Home → Account ID
  - Example: `0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d`

### Optional Fields

- **`location`** (enum): Primary region for the bucket (location hint) - Default: `auto`
  - `auto`: No hint; Cloudflare selects the optimal region
  - `wnam`: Western North America (US West)
  - `enam`: Eastern North America (US East)
  - `weur`: Western Europe
  - `eeur`: Eastern Europe
  - `apac`: Asia-Pacific
  - `oc`: Oceania
  - The hint is only honored when the bucket is first created and is best-effort.

- **`publicAccess`** (bool): Expose the bucket via Cloudflare's managed r2.dev public URL - Default: `false`
  - `false`: Bucket requires authentication (API keys or presigned URLs)
  - **Note**: r2.dev URLs are rate-limited; use a custom domain for production public buckets

- **`customDomain`** (object): Custom domain configuration for the bucket
  - **`enabled`** (bool): Whether to enable custom domain access
  - **`zoneId`** (StringValueOrRef): Cloudflare Zone ID where the domain exists
    - Can be a literal value: `zoneId: { value: "..." }`
    - Or reference a CloudflareDnsZone: `zoneId: { valueFrom: { name: "my-dns-zone" } }`
  - **`domain`** (string): Full domain name (e.g., "media.example.com")

## Common Patterns

### Private Bucket for Application Data

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: app-data-bucket
spec:
  bucketName: myapp-user-uploads
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  location: enam  # Eastern North America
  publicAccess: false  # Requires authentication
```

**Access**: Use R2 API keys or presigned URLs in your application.

### Public Bucket for CDN/Media

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: cdn-assets
spec:
  bucketName: public-cdn-assets
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  location: weur  # Western Europe
  publicAccess: true  # Expose via the managed r2.dev public URL
```

**Access**: Once the managed public URL is enabled, objects are served from a Cloudflare-assigned `r2.dev` address.

**Production Best Practice**: Use a custom domain instead of r2.dev (see documentation).

### Multi-Region Strategy

For global applications, create regional buckets:

```yaml
---
# US bucket
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: assets-us
spec:
  bucketName: myapp-assets-us
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  location: enam

---
# EU bucket
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: assets-eu
spec:
  bucketName: myapp-assets-eu
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  location: weur
```

Route users to the nearest bucket for optimal latency.

## Location Hints

R2's location hints optimize data placement for performance:

| Location | Region | Best For |
|----------|--------|----------|
| **auto** | Cloudflare chooses | No strong regional preference |
| **wnam** | Western North America (US West) | Users in US West Coast, CA, Pacific Northwest |
| **enam** | Eastern North America (US East) | Users in US East Coast, NY, Midwest |
| **weur** | Western Europe | Users in UK, France, Spain, Western EU |
| **eeur** | Eastern Europe | Users in Germany, Poland, Eastern EU |
| **apac** | Asia-Pacific | Users in Australia, Japan, Singapore, India |
| **oc** | Oceania | Users in Australia, New Zealand |

**Important**: Location hints are optimization suggestions, not strict geo-fencing. R2 automatically replicates data globally for durability. The hint influences initial placement and caching.

## S3 API Compatibility

R2 is S3-compatible for data operations:

### Supported S3 Features

✅ **Standard Operations**: GET, PUT, DELETE, HEAD, LIST  
✅ **Multipart Uploads**: For files > 100MB  
✅ **Presigned URLs**: Temporary authenticated access  
✅ **Range Requests**: Partial object downloads  
✅ **Metadata**: Custom object metadata  
✅ **ETags**: Content integrity verification  
✅ **CORS**: Cross-origin resource sharing (configured via S3 API)  

### Not Supported

❌ **Object Versioning**: R2 does not support versioning  
❌ **Bucket Policies**: Use R2 API tokens for access control  
❌ **Static Website Hosting**: Use Cloudflare Pages instead  
❌ **Object Lock/Legal Hold**: Not available  
❌ **S3 Select**: Query-in-place not supported  

### Using AWS CLI with R2

```bash
# Configure AWS CLI with R2 endpoint
export AWS_ENDPOINT_URL="https://<account-id>.r2.cloudflarestorage.com"
export AWS_ACCESS_KEY_ID="<r2-access-key-id>"
export AWS_SECRET_ACCESS_KEY="<r2-secret-access-key>"

# List buckets
aws s3 ls

# Upload file
aws s3 cp file.txt s3://my-bucket/

# Download file
aws s3 cp s3://my-bucket/file.txt ./
```

## Public Access and Custom Domains

### r2.dev Public URLs (Development)

When `publicAccess: true`, objects are served from a Cloudflare-assigned `r2.dev` address for the bucket.

**Limitations**:
- Rate-limited (not suitable for high traffic)
- No CDN caching
- Cloudflare branding in URL

**Use for**: Development, testing, internal tools

### Custom Domains (Production)

For production, configure a custom domain directly in your manifest:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: media-bucket
spec:
  bucketName: media-assets
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  location: weur
  customDomain:
    enabled: true
    zoneId:
      value: "f1e2d3c4b5a6978899aabbccddeeff00"  # Your Cloudflare Zone ID
    domain: "media.example.com"
```

Or reference a `CloudflareDnsZone` resource:

```yaml
spec:
  customDomain:
    enabled: true
    zoneId:
      valueFrom:
        name: my-dns-zone  # References CloudflareDnsZone named "my-dns-zone"
    domain: "media.example.com"
```

**Benefits**:
- No rate limits
- Full CDN caching
- Your branding
- HTTPS with your certificate
- Automated via IaC (no manual dashboard configuration)

## Cost Model

R2's pricing is simple and predictable:

### Storage Costs

- **Standard Storage**: $0.015/GB-month
- **Infrequent Access**: $0.010/GB-month (30+ day retention minimum)

### Operation Costs

- **Class A Operations** (writes): $4.50 per million
  - PUT, POST, LIST, COPY
- **Class B Operations** (reads): $0.36 per million
  - GET, HEAD

### Free Tier

- **First 10 GB** of storage: Free
- **1 million Class A operations**: Free per month
- **10 million Class B operations**: Free per month

### Egress

- **All egress**: **$0.00** (completely free)

**Cost Example**: A 1TB media library serving 100TB/month:
- Storage: 1000 GB × $0.015 = **$15/month**
- Reads: ~300M requests × $0.36/M = **$108/month**
- Egress: 100TB × $0 = **$0/month**
- **Total: $123/month**

**AWS S3 Equivalent**: $15 (storage) + $27 (requests) + **$9,000 (egress)** = **$9,042/month**

**Savings: 98.6%**

## Migration from S3

R2's S3 compatibility makes migration straightforward:

### Option 1: Lazy Migration with Sippy

Cloudflare Sippy automatically migrates objects on-demand:
1. Enable Sippy on R2 bucket, pointing to S3
2. Update application to use R2 endpoint
3. On first access, R2 fetches from S3 and caches
4. Subsequent requests served from R2 (no egress fees)

**Pro**: Zero downtime, no upfront egress costs  
**Con**: First access per object is slower

### Option 2: Bulk Copy with rclone

```bash
# Configure rclone with R2 endpoint
rclone copy s3:my-s3-bucket r2:my-r2-bucket --progress

# Sync continuously
rclone sync s3:my-s3-bucket r2:my-r2-bucket --progress
```

**Pro**: Faster bulk migration  
**Con**: Incurs S3 egress fees during copy

## Best Practices

1. **Use location hints strategically**: Choose the region closest to most users
2. **Custom domains for production public buckets**: Don't rely on rate-limited r2.dev URLs
3. **Enable CORS via S3 API**: Configure cross-origin access for web apps
4. **Use presigned URLs for temporary access**: Better than embedding credentials
5. **Monitor costs in Cloudflare Dashboard**: Track storage and operation usage
6. **Plan for no versioning**: Implement application-level versioning if needed
7. **Leverage S3 API compatibility**: Reuse existing S3 tools and SDKs

## Anti-Patterns to Avoid

❌ **Using r2.dev URLs for production high-traffic sites** (rate-limited)  
❌ **Expecting S3 versioning to work** (R2 doesn't support it)  
❌ **Implementing bucket policies** (use R2 API tokens instead)  
❌ **Trying to use S3 Select or Glacier** (not supported)  
❌ **Forgetting to test S3 SDK compatibility** (most work, but verify)  

## Integration with Cloudflare Services

### Cloudflare Workers

Access R2 directly from Workers with zero latency:

```javascript
export default {
  async fetch(request, env) {
    const object = await env.MY_BUCKET.get("file.txt");
    return new Response(object.body);
  }
};
```

### Cloudflare Pages

Serve static assets from R2 via Pages Functions.

### Cloudflare CDN

Cache R2 content at Cloudflare's edge for global delivery.

## Deployment

OpenMCF handles creating the R2 bucket via either Pulumi or Terraform:

- **Pulumi**: See [iac/pulumi/README.md](./iac/pulumi/README.md)
- **Terraform**: See [iac/tf/README.md](./iac/tf/README.md)

## Examples

See the [Common Patterns](#common-patterns) above and the ready-to-use [presets](./presets/) for complete, working examples including:
- Private buckets for application data
- Public buckets for CDN/media delivery
- Multi-region deployment strategies
- Development vs production configurations

## Documentation

- **[Research Documentation](./docs/README.md)**: Deep dive into R2 architecture, deployment methods, and 80/20 scoping decisions
- **[Pulumi Module](./iac/pulumi/README.md)**: Pulumi deployment guide and architecture
- **[Terraform Module](./iac/tf/README.md)**: Terraform deployment guide

## Support

- Cloudflare R2: [Official Docs](https://developers.cloudflare.com/r2/)
- R2 Pricing: [Pricing Details](https://developers.cloudflare.com/r2/pricing/)
- S3 API Compatibility: [API Reference](https://developers.cloudflare.com/r2/api/s3/api/)
- Migration Guide: [Sippy Documentation](https://blog.cloudflare.com/sippy-incremental-migration-s3-r2/)

## What's Next?

- Explore the [presets](./presets/) for complete usage patterns
- Read [docs/README.md](./docs/README.md) for architectural deep dive
- Deploy using [iac/pulumi/](./iac/pulumi/) or [iac/tf/](./iac/tf/)
- Review [R2 API documentation](https://developers.cloudflare.com/r2/)

---

**Bottom Line**: Cloudflare R2 gives you S3-compatible object storage with zero egress fees, strong consistency, and production-grade durability at a fraction of the cost of AWS/GCP/Azure. Perfect for content delivery, multi-cloud architectures, and escaping vendor lock-in.


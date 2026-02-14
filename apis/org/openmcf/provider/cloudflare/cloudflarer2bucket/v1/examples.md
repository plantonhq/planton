# Cloudflare R2 Bucket Examples

Complete, working examples for common Cloudflare R2 Bucket patterns using OpenMCF.

## Table of Contents

- [Example 1: Basic Private Bucket](#example-1-basic-private-bucket)
- [Example 2: Public Bucket for CDN](#example-2-public-bucket-for-cdn)
- [Example 3: Multi-Region Deployment](#example-3-multi-region-deployment)
- [Example 4: Development Environment](#example-4-development-environment)
- [Example 5: Production Configuration](#example-5-production-configuration)
- [Example 6: Media Storage Bucket](#example-6-media-storage-bucket)
- [Example 7: User Uploads Bucket](#example-7-user-uploads-bucket)
- [Example 8: Backup and Archive](#example-8-backup-and-archive)
- [Example 9: Custom Domain with Literal Zone ID](#example-9-custom-domain-with-literal-zone-id)
- [Example 10: Custom Domain with Zone Reference](#example-10-custom-domain-with-zone-reference)

---

## Example 1: Basic Private Bucket

**Use Case**: Simple private bucket for application data with authentication required.

**Features**:
- Private access (authentication required)
- Regional placement for optimal latency
- Standard storage class

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: app-data-bucket
spec:
  bucketName: myapp-private-data
  accountId: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"  # Replace with your account ID
  location: ENAM  # Eastern North America
  publicAccess: false

```

**Expected Behavior**:
- Bucket created in US East region
- Objects require R2 API keys or presigned URLs for access
- No public r2.dev URL enabled

**Access Pattern**:
```bash
# Using AWS CLI with R2 endpoint
aws s3 --endpoint-url https://a1b2...p6.r2.cloudflarestorage.com \
  cp file.txt s3://myapp-private-data/
```

---

## Example 2: Public Bucket for CDN

**Use Case**: Publicly accessible bucket for serving static assets, images, or media files.

**Features**:
- Public access enabled (r2.dev subdomain)
- Geographic placement for content delivery
- No authentication required for reads

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: cdn-assets-bucket
spec:
  bucketName: public-cdn-assets
  accountId: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  location: WEUR  # Western Europe
  publicAccess: true  # Enable r2.dev public URL

```

**Expected Behavior**:
- Bucket accessible at `https://public-cdn-assets.<account-id>.r2.dev/`
- Anyone can read objects without authentication
- Writes still require API keys

**Access Pattern**:
```html
<!-- Direct access from HTML -->
<img src="https://public-cdn-assets.<account-id>.r2.dev/images/logo.png">

<!-- Or via JavaScript -->
<script src="https://public-cdn-assets.<account-id>.r2.dev/js/app.js"></script>
```

**Production Note**: For high-traffic production use, configure a custom domain instead of r2.dev (which is rate-limited).

---

## Example 3: Multi-Region Deployment

**Use Case**: Global application with regional buckets for optimal latency.

**Features**:
- Separate buckets per region
- Route users to nearest bucket
- Consistent naming convention

```yaml
---
# US East bucket
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: assets-us-east
spec:
  bucketName: myapp-assets-us-east
  accountId: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  location: ENAM
  publicAccess: false

---
# US West bucket
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: assets-us-west
spec:
  bucketName: myapp-assets-us-west
  accountId: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  location: WNAM
  publicAccess: false

---
# Europe bucket
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: assets-eu
spec:
  bucketName: myapp-assets-eu
  accountId: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  location: WEUR
  publicAccess: false

---
# Asia-Pacific bucket
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: assets-apac
spec:
  bucketName: myapp-assets-apac
  accountId: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  location: APAC
  publicAccess: false
```

**Expected Behavior**:
- Four regional buckets optimized for different geographies
- Application logic routes users based on location
- Lower latency for regional users

**Application Logic**:
```javascript
function getBucketForUser(userLocation) {
  const bucketMap = {
    'US-EAST': 'myapp-assets-us-east',
    'US-WEST': 'myapp-assets-us-west',
    'EU': 'myapp-assets-eu',
    'APAC': 'myapp-assets-apac'
  };
  return bucketMap[userLocation] || 'myapp-assets-us-east'; // Default
}
```

---

## Example 4: Development Environment

**Use Case**: Development/testing bucket with relaxed access for rapid iteration.

**Features**:
- Public access for easy testing
- Development naming convention
- Disposable (can be recreated easily)

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: dev-test-bucket
spec:
  bucketName: myapp-dev-testing
  accountId: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  location: ENAM
  publicAccess: true  # Easy testing without auth

```

**Expected Behavior**:
- Quick access for developers via r2.dev URL
- No production data or secrets
- Can be destroyed and recreated freely

**Development Workflow**:
```bash
# Quick upload for testing
aws s3 --endpoint-url https://a1b2...p6.r2.cloudflarestorage.com \
  cp test-file.json s3://myapp-dev-testing/

# Test access
curl https://myapp-dev-testing.<account-id>.r2.dev/test-file.json
```

---

## Example 5: Production Configuration

**Use Case**: Production-grade bucket with security and naming best practices.

**Features**:
- Private access with strict auth
- Production naming convention
- Regional placement for main audience

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: prod-app-data
  labels:
    environment: production
    team: platform
    cost-center: engineering
spec:
  bucketName: myapp-prod-data
  accountId: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  location: ENAM
  publicAccess: false  # Strict auth required

```

**Expected Behavior**:
- Secure, production-ready bucket
- All access requires authentication
- Clearly labeled for cost tracking and governance

**Access Control**:
```bash
# Use environment variables for credentials
export AWS_ACCESS_KEY_ID="<prod-r2-access-key>"
export AWS_SECRET_ACCESS_KEY="<prod-r2-secret-key>"
export AWS_ENDPOINT_URL="https://a1b2...p6.r2.cloudflarestorage.com"

# Operations require valid credentials
aws s3 cp sensitive-data.csv s3://myapp-prod-data/
```

---

## Example 6: Media Storage Bucket

**Use Case**: Store and serve video, images, and audio files for a media application.

**Features**:
- Public access for content delivery
- Location optimized for primary audience
- Large file handling

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: media-storage
spec:
  bucketName: myapp-media-assets
  accountId: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  location: WEUR  # Western Europe
  publicAccess: true

```

**Expected Behavior**:
- Optimized for large file uploads (multipart)
- Zero egress fees for video streaming
- Public URL access for embedded media

**Upload Large Files**:
```bash
# Multipart upload for large files (>100MB)
aws s3 cp large-video.mp4 \
  s3://myapp-media-assets/videos/ \
  --endpoint-url https://a1b2...p6.r2.cloudflarestorage.com
```

**Embed in Application**:
```html
<video controls>
  <source src="https://myapp-media-assets.<account-id>.r2.dev/videos/large-video.mp4" 
          type="video/mp4">
</video>
```

---

## Example 7: User Uploads Bucket

**Use Case**: Store user-generated content with controlled access.

**Features**:
- Private bucket for security
- Presigned URLs for temporary access
- User isolation via path prefixes

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: user-uploads
spec:
  bucketName: myapp-user-uploads
  accountId: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  location: ENAM
  publicAccess: false  # Users get presigned URLs

```

**Expected Behavior**:
- All uploads require authentication
- Application generates presigned URLs for users
- Users cannot list or access other users' files

**Generate Presigned URL (Python)**:
```python
import boto3
from botocore.client import Config

s3_client = boto3.client(
    's3',
    endpoint_url='https://a1b2...p6.r2.cloudflarestorage.com',
    aws_access_key_id='<access-key>',
    aws_secret_access_key='<secret-key>',
    config=Config(signature_version='s3v4')
)

# Generate presigned URL for upload (valid for 1 hour)
presigned_url = s3_client.generate_presigned_url(
    'put_object',
    Params={
        'Bucket': 'myapp-user-uploads',
        'Key': f'users/{user_id}/profile-photo.jpg'
    },
    ExpiresIn=3600
)

# User uploads directly to R2 using this URL
```

---

## Example 8: Backup and Archive

**Use Case**: Long-term backup storage with infrequent access.

**Features**:
- Private access for security
- Infrequent Access storage class (lower cost)
- Retention-optimized

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: backup-archive
spec:
  bucketName: myapp-backups-archive
  accountId: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  location: ENAM
  publicAccess: false

```

**Expected Behavior**:
- Secure storage for backups
- Zero egress fees for restore operations
- Cost-effective long-term retention

**Backup Workflow**:
```bash
# Daily backup upload
tar -czf backup-$(date +%Y-%m-%d).tar.gz /var/lib/myapp/
aws s3 cp backup-$(date +%Y-%m-%d).tar.gz \
  s3://myapp-backups-archive/daily/ \
  --endpoint-url https://a1b2...p6.r2.cloudflarestorage.com

# Restore when needed (no egress fees!)
aws s3 cp s3://myapp-backups-archive/daily/backup-2025-01-15.tar.gz ./ \
  --endpoint-url https://a1b2...p6.r2.cloudflarestorage.com
```

---

## Environment-Specific Configurations

### Development

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: dev-bucket
  labels:
    environment: development
spec:
  bucketName: myapp-dev
  accountId: "dev-account-id-32-hex-characters"
  location: ENAM
  publicAccess: true  # Easy testing
```

### Staging

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: staging-bucket
  labels:
    environment: staging
spec:
  bucketName: myapp-staging
  accountId: "staging-account-id-32-hex-chars"
  location: ENAM
  publicAccess: false  # Match prod security
```

### Production

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: prod-bucket
  labels:
    environment: production
    criticality: high
spec:
  bucketName: myapp-prod
  accountId: "prod-account-id-32-hex-characters"
  location: ENAM
  publicAccess: false  # Strict auth
```

---

## Testing Your Bucket

### Verify Bucket Creation

```bash
# List buckets
aws s3 ls --endpoint-url https://<account-id>.r2.cloudflarestorage.com

# Should show: myapp-bucket-name
```

### Upload Test File

```bash
echo "Hello R2" > test.txt
aws s3 cp test.txt s3://myapp-bucket-name/ \
  --endpoint-url https://<account-id>.r2.cloudflarestorage.com
```

### Download Test File

```bash
aws s3 cp s3://myapp-bucket-name/test.txt ./ \
  --endpoint-url https://<account-id>.r2.cloudflarestorage.com

cat test.txt  # Should print: Hello R2
```

### Test Public Access (if enabled)

```bash
curl https://myapp-bucket-name.<account-id>.r2.dev/test.txt
# Should return: Hello R2
```

---

## Common Issues and Solutions

### Issue: "AccessDenied" errors

**Cause**: Invalid or missing R2 API credentials

**Solution**:
1. Verify `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` are correct
2. Check that credentials have permission for the operation
3. Ensure using the correct endpoint URL

---

### Issue: "NoSuchBucket" errors

**Cause**: Bucket name doesn't exist or typo in name

**Solution**:
1. Verify bucket was created successfully
2. Check for typos in bucket name
3. Ensure using correct Cloudflare account

---

### Issue: Public URL returns 403

**Cause**: Public access not enabled or incorrect URL

**Solution**:
1. Verify `public_access: true` in spec
2. Check URL format: `https://<bucket>.<account-id>.r2.dev/<key>`
3. Wait a few minutes for propagation

---

### Issue: Rate limiting on r2.dev URL

**Cause**: r2.dev URLs have rate limits

**Solution**:
1. Use custom domain for production (no rate limits)
2. Configure custom domain in Cloudflare Dashboard
3. Enable Cloudflare CDN caching

---

## Example 9: Custom Domain with Literal Zone ID

**Use Case**: Production bucket accessible via a custom domain, with zone ID provided directly.

**Features**:
- Custom domain for production access (no rate limits)
- Full CDN caching via Cloudflare
- Professional branding with your domain

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: media-bucket
spec:
  bucketName: media-assets
  accountId: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  location: WEUR
  customDomain:
    enabled: true
    zoneId:
      value: "f1e2d3c4b5a6978899aabbccddeeff00"  # Your Cloudflare Zone ID
    domain: "media.example.com"
```

**Expected Behavior**:
- Bucket accessible at `https://media.example.com`
- Full Cloudflare CDN caching enabled
- No rate limits (unlike r2.dev URLs)

**Access Pattern**:
```html
<!-- Direct access from HTML -->
<img src="https://media.example.com/images/logo.png">

<!-- Or via JavaScript -->
<script src="https://media.example.com/js/app.js"></script>
```

---

## Example 10: Custom Domain with Zone Reference

**Use Case**: Production bucket with custom domain, referencing zone from a CloudflareDnsZone resource.

**Features**:
- Custom domain configuration via resource reference
- Zone ID automatically resolved from CloudflareDnsZone outputs
- Infrastructure dependencies managed automatically

```yaml
---
# First, define the DNS zone
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareDnsZone
metadata:
  name: example-zone
spec:
  zoneName: example.com
  accountId: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"

---
# Then, reference it in the R2 bucket
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: cdn-bucket
spec:
  bucketName: cdn-assets
  accountId: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  location: WEUR
  customDomain:
    enabled: true
    zoneId:
      valueFrom:
        name: example-zone  # References the CloudflareDnsZone above
    domain: "cdn.example.com"
```

**Expected Behavior**:
- Zone ID automatically resolved from `example-zone` resource outputs
- Bucket accessible at `https://cdn.example.com`
- Deployment order managed automatically (zone created before bucket)

**Benefits**:
- No hardcoded zone IDs
- Single source of truth for zone configuration
- Automatic dependency management

---

## Migration Examples

### Migrate from S3 to R2

```bash
# Using rclone for bulk migration
rclone sync s3:my-s3-bucket r2:my-r2-bucket \
  --progress \
  --transfers 10

# Using AWS CLI (slower but simple)
aws s3 sync s3://my-s3-bucket/ s3://my-r2-bucket/ \
  --endpoint-url https://<account-id>.r2.cloudflarestorage.com \
  --source-region us-east-1
```

---

## Next Steps

- Read the [README.md](./README.md) for detailed field documentation
- Review [research documentation](./docs/README.md) for architectural deep dive
- Deploy using [Pulumi](./iac/pulumi/README.md) or [Terraform](./iac/tf/README.md)
- Check [Cloudflare R2 documentation](https://developers.cloudflare.com/r2/)

---

**Questions or Issues?** Refer to the [README.md](./README.md) or Cloudflare's official R2 documentation.


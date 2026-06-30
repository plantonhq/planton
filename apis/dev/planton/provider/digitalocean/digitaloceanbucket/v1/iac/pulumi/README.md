# DigitalOcean Spaces Bucket - Pulumi Module

## Overview

This Pulumi module deploys DigitalOcean Spaces buckets (S3-compatible object storage) using Planton's declarative API. The module transforms the `DigitalOceanBucketSpec` protobuf definition into DigitalOcean Spaces resources with built-in CDN integration.

## Features

- **S3-Compatible Storage**: Full S3 API compatibility for existing tools and SDKs
- **Built-in CDN**: Automatic CDN endpoint for public buckets
- **Access Control**: Private or public-read bucket configurations
- **Versioning**: Object versioning for data protection
- **Tagging**: Bucket organization and cost allocation
- **State Management**: Pulumi tracks resources and detects configuration drift

## Prerequisites

- **Pulumi CLI**: Version 3.x or higher
- **Go**: Version 1.21+ (for module development)
- **DigitalOcean Account**: With Spaces enabled
- **DigitalOcean Credentials**:
  - API Token (for bucket management)
  - Spaces Access Key ID (for object access)
  - Spaces Secret Access Key (for object access)

## Installation

This module is part of the Planton monorepo:

```bash
git clone https://github.com/plantonhq/planton.git
cd planton/apis/dev/planton/provider/digitalocean/digitaloceanbucket/v1/iac/pulumi
```

## Usage

### Setting Up Credentials

```bash
# DigitalOcean API token (for infrastructure management)
export DIGITALOCEAN_TOKEN="dop_v1_..."

# Spaces access keys (for S3-compatible object access)
export SPACES_ACCESS_KEY_ID="your-spaces-key"
export SPACES_SECRET_ACCESS_KEY="your-spaces-secret"
```

### Initialize Pulumi Stack

```bash
pulumi stack init dev
pulumi config set digitalocean:token $DIGITALOCEAN_TOKEN --secret
```

### Basic Deployment

Create a Pulumi program or use the manifest-based deployment:

```go
package main

import (
    digitaloceanbucketv1 "github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean/digitaloceanbucket/v1"
    "github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean/digitaloceanbucket/v1/iac/pulumi/module"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        stackInput := &digitaloceanbucketv1.DigitalOceanBucketStackInput{
            Target: &digitaloceanbucketv1.DigitalOceanBucket{
                Metadata: &shared.CloudResourceMetadata{
                    Name: "my-bucket",
                },
                Spec: &digitaloceanbucketv1.DigitalOceanBucketSpec{
                    BucketName:        "my-app-data",
                    Region:            digitaloceanbucketv1.DigitalOceanRegion_nyc3,
                    AccessControl:     digitaloceanbucketv1.DigitalOceanBucketAccessControl_PRIVATE,
                    VersioningEnabled: false,
                },
            },
            ProviderConfig: &digitaloceanv1.DigitalOceanProviderConfig{
                Token: os.Getenv("DIGITALOCEAN_TOKEN"),
            },
        }

        return module.Resources(ctx, stackInput)
    })
}
```

### Deploy

```bash
# Preview changes
pulumi preview

# Deploy
pulumi up

# View outputs
pulumi stack output bucket_id
pulumi stack output endpoint
```

## Module Architecture

### Input Structure

The module accepts `DigitalOceanBucketStackInput` containing:
- **target**: The `DigitalOceanBucket` resource specification
- **provider_config**: DigitalOcean API credentials

### Processing Flow

1. **Initialize locals** - Extract metadata and configuration
2. **Create provider** - Set up DigitalOcean provider with credentials
3. **Map access control** - Convert protobuf enum to S3 ACL string
4. **Create bucket** - Deploy Spaces bucket with configuration
5. **Configure versioning** - Enable if specified
6. **Export outputs** - Return bucket ID and endpoint

### Outputs

The module exports:
- `bucket_id`: Unique bucket identifier (`region:bucket-name`)
- `endpoint`: Bucket endpoint URL

## Access Control Configuration

The module maps protobuf enums to S3 ACL strings:

| Protobuf Enum | ACL String | Public Access |
|---------------|------------|---------------|
| `PRIVATE` (0) | `private` | No |
| `PUBLIC_READ` (1) | `public-read` | Yes |

## Versioning Behavior

When `versioning_enabled: true`:
```go
bucketArgs.Versioning = &digitalocean.SpacesBucketVersioningArgs{
    Enabled: pulumi.Bool(true),
}
```

**Important**: Versioning cannot be disabled once enabled, only suspended.

## Debugging

### Enable Debug Mode

```bash
export PULUMI_DEBUG_COMMANDS=true
pulumi up --debug
```

Or use the provided debug script:

```bash
./debug.sh
```

### Common Issues

**Issue**: "Failed to setup digitalocean provider"
- **Solution**: Verify `DIGITALOCEAN_TOKEN` is set and valid

**Issue**: "Bucket name already in use"
- **Solution**: Bucket names must be globally unique across all DigitalOcean Spaces

**Issue**: "Access denied when uploading objects"
- **Solution**: Configure Spaces access keys (separate from API token)

## Best Practices

### Credential Management

- **Never commit credentials** to version control
- Use Pulumi secrets for sensitive values: `pulumi config set --secret`
- Rotate API tokens and Spaces keys regularly

### State Management

Use remote state for production:

```bash
# Pulumi Service (default)
pulumi login

# Or S3-compatible backend
pulumi login s3://my-state-bucket?endpoint=nyc3.digitaloceanspaces.com

# Or local filesystem (dev only)
pulumi login file://~/.pulumi
```

### Multi-Environment Pattern

Use separate stacks for environments:

```bash
# Development
pulumi stack init dev
pulumi config set digitalocean:token $DEV_TOKEN --secret
pulumi up

# Production
pulumi stack init prod
pulumi config set digitalocean:token $PROD_TOKEN --secret
pulumi up
```

### Tagging Strategy

Use consistent tags across environments:
```yaml
tags:
  - environment:production
  - team:platform
  - purpose:media-storage
  - cost-center:engineering
```

## Accessing Created Buckets

### Using AWS CLI

```bash
# List buckets
aws --endpoint-url https://nyc3.digitaloceanspaces.com s3 ls

# List objects in bucket
aws --endpoint-url https://nyc3.digitaloceanspaces.com s3 ls s3://my-bucket/

# Upload file
aws --endpoint-url https://nyc3.digitaloceanspaces.com s3 cp myfile.txt s3://my-bucket/

# Download file
aws --endpoint-url https://nyc3.digitaloceanspaces.com s3 cp s3://my-bucket/myfile.txt ./
```

### Using SDKs

**JavaScript/Node.js:**
```javascript
const AWS = require('aws-sdk');

const spacesEndpoint = new AWS.Endpoint('nyc3.digitaloceanspaces.com');
const s3 = new AWS.S3({
  endpoint: spacesEndpoint,
  accessKeyId: process.env.SPACES_KEY,
  secretAccessKey: process.env.SPACES_SECRET
});

// Upload
await s3.putObject({
  Bucket: 'my-bucket',
  Key: 'file.txt',
  Body: 'Hello World'
}).promise();
```

**Python (boto3):**
```python
import boto3

session = boto3.session.Session()
client = session.client('s3',
    region_name='nyc3',
    endpoint_url='https://nyc3.digitaloceanspaces.com',
    aws_access_key_id=os.getenv('SPACES_KEY'),
    aws_secret_access_key=os.getenv('SPACES_SECRET'))

# Upload
client.upload_file('file.txt', 'my-bucket', 'file.txt')
```

## Examples

For comprehensive examples including multi-region setups, versioning, and production configurations, see:
- [API-level examples](../../examples.md)
- [Pulumi-specific examples](./examples.md) (if available)

## References

- [Pulumi DigitalOcean Provider](https://www.pulumi.com/registry/packages/digitalocean/)
- [DigitalOcean Spaces API](https://docs.digitalocean.com/reference/api/spaces-api/)
- [Module Overview](./overview.md) - Architecture details

## Support

- **Planton**: [github.com/plantonhq/planton](https://github.com/plantonhq/planton)
- **Pulumi Docs**: [pulumi.com/docs](https://www.pulumi.com/docs/)
- **DigitalOcean Support**: [digitalocean.com/support](https://www.digitalocean.com/support/)

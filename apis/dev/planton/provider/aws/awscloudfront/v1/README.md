# AwsCloudFront

AWS CloudFront is a global content delivery network (CDN) that accelerates the delivery of static and dynamic content to users worldwide. This resource provisions a CloudFront distribution with a minimal 80/20 configuration: origins, default origin selection, optional custom domain aliases with ACM certificate, price class, and default root object.

The **AwsCloudFront** API provides an opinionated, simplified interface that covers the most common production use cases while eliminating configuration complexity and common pitfalls.

## Overview

CloudFront sits at the edge of your application architecture, serving as:
- **Performance Accelerator**: Caching content at 400+ global edge locations for sub-100ms latency
- **Security Gateway**: Integrates with AWS WAF and Shield for DDoS protection and threat filtering
- **Cost Optimizer**: Free data transfer from AWS origins (S3, ALB, EC2) to CloudFront edges
- **Global HTTP/HTTPS Frontend**: HTTPS enforcement, custom domain support, and HTTP/2 by default

This resource is ideal for:
- Static websites hosted on S3 (React, Vue, Next.js builds)
- API acceleration via ALB or API Gateway origins
- Multi-origin CDN architectures (static assets from S3, dynamic content from ALB)

## Spec Fields

### `enabled` (bool)
Whether the CloudFront distribution is enabled and actively serving traffic.
- **Default**: Typically set to `true` for production
- **Use Case**: Set to `false` to temporarily disable the distribution without deleting it (useful for maintenance windows)

### `aliases` ([]string, optional)
Custom domain names (CNAMEs) for the distribution, such as `cdn.example.com` or `www.example.com`.
- **Format**: Must be valid fully-qualified domain names (FQDNs)
- **Validation**: Each alias must be unique within the array
- **Requirement**: When aliases are set, `certificate_arn` **must** be provided
- **DNS Setup**: After deploying, create a Route53 CNAME or Alias record pointing your custom domain to the CloudFront `domain_name` output

**Example**:
```yaml
aliases:
  - cdn.example.com
  - assets.example.com
```

### `certificate_arn` (string, optional but required with aliases)
The ARN of an AWS Certificate Manager (ACM) certificate **in the us-east-1 region**.
- **Critical Requirement**: CloudFront requires certificates to be in `us-east-1`, regardless of your origin's region
- **Format**: `arn:aws:acm:us-east-1:123456789012:certificate/<uuid>`
- **Validation**: The ARN pattern is enforced at the API level to catch the us-east-1 requirement before deployment
- **How to Create**: Request a public certificate in ACM us-east-1 for your domain, validate it via DNS or email

**Example**:
```yaml
certificateArn: arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012
```

### `price_class` (enum)
Determines which edge locations CloudFront uses, impacting both performance and cost.

**Options**:
- **`PRICE_CLASS_100`** (recommended default): North America, Europe, Israel
  - Lowest cost
  - Ideal for applications serving primarily NA/EU users
  - ~70% of global internet users covered
- **`PRICE_CLASS_200`**: All locations in 100, plus Asia (excluding mainland China), Middle East, Africa
  - Moderate cost increase (~30% more than 100)
  - Better latency for Asian and African users
- **`PRICE_CLASS_ALL`**: All global edge locations including South America, Australia, New Zealand
  - Highest cost
  - Best global performance

**Cost Impact**: Price classes can reduce CloudFront costs by 40-60% for regionally-focused applications.

**Example**:
```yaml
priceClass: PRICE_CLASS_100
```

### `origins` ([]Origin, required)
List of origins (backend servers or S3 buckets) that CloudFront fetches content from.

Each origin includes:
- **`domain_name`** (string, required): DNS name of the origin
  - For S3: `bucket-name.s3.amazonaws.com` or `bucket-name.s3-website.region.amazonaws.com`
  - For ALB: `my-alb-1234567890.us-west-2.elb.amazonaws.com`
  - For custom: any valid domain or IP
- **`origin_path`** (string, optional): Prefix path appended to all requests to this origin
  - Example: `/assets` would prepend `/assets` to every request (viewer requests `/logo.png`, CloudFront fetches `/assets/logo.png` from origin)
- **`is_default`** (bool, required): Marks this origin as the default target for all requests
  - **Validation**: Exactly one origin must be marked as default

**Minimal Example (Single S3 Origin)**:
```yaml
origins:
  - domainName: my-bucket.s3.amazonaws.com
    isDefault: true
```

**Multi-Origin Example**:
```yaml
origins:
  - domainName: my-static-bucket.s3.amazonaws.com
    originPath: /assets
    isDefault: true
  - domainName: api.example.com
    isDefault: false
```

### `default_root_object` (string, optional)
The object CloudFront returns when a user requests the root URL (e.g., `https://example.com/`).
- **Common Value**: `index.html` (for static websites)
- **Behavior**: If omitted, CloudFront forwards requests to `/` directly to the origin
- **Format**: Must be a valid filename (alphanumeric, dashes, underscores, dots)

**Example**:
```yaml
defaultRootObject: index.html
```

## Validation Rules

The API enforces several critical validations to prevent common misconfigurations:

1. **Aliases Require Certificate**
   - If `aliases` is non-empty, `certificate_arn` must be provided
   - Prevents the "can't use custom domain without HTTPS certificate" error

2. **Certificate Must Be in us-east-1**
   - The `certificate_arn` pattern validation enforces the `us-east-1` region
   - Catches this gotcha at validation time, not deploy-time

3. **Exactly One Default Origin**
   - The `origins` array must contain exactly one origin with `is_default: true`
   - Prevents ambiguity about which origin to use for requests

4. **Unique Aliases**
   - All aliases in the `aliases` array must be unique
   - Prevents configuration errors and AWS API rejections

5. **Enum Enforcement**
   - `price_class` must be one of the defined enum values
   - Invalid values are rejected at validation time

## Stack Outputs

After deployment, the following outputs are available for use in DNS configuration, automation, or dependent resources:

### `distribution_id` (string)
The CloudFront distribution ID (e.g., `E2QWRUHAPOMQZL`).
- **Use Case**: Triggering cache invalidations via `aws cloudfront create-invalidation --distribution-id <id>`
- **Use Case**: Referencing the distribution in AWS Console or other infrastructure

### `domain_name` (string)
The CloudFront-assigned domain name (e.g., `d123abc.cloudfront.net`).
- **Use Case**: Creating DNS CNAME or Alias records pointing to this domain
- **Use Case**: Testing the distribution before associating custom domains

### `hosted_zone_id` (string)
The Route53 hosted zone ID for CloudFront distributions (always `Z2FDTNDATAQYW2`).
- **Use Case**: Creating Route53 Alias records for the apex domain (e.g., `example.com` instead of `www.example.com`)
- **Why It's Constant**: All CloudFront distributions use this same hosted zone ID—it's an AWS global constant

**Example Route53 Alias Record**:
```hcl
resource "aws_route53_record" "cdn" {
  zone_id = var.my_hosted_zone_id
  name    = "cdn.example.com"
  type    = "A"
  
  alias {
    name                   = aws_cloudfront_distribution.main.domain_name
    zone_id                = aws_cloudfront_distribution.main.hosted_zone_id  # Z2FDTNDATAQYW2
    evaluate_target_health = false
  }
}
```

## How It Works

This module can be provisioned with either Pulumi or Terraform via the Planton CLI:

1. **Define Your Manifest**: Create a YAML file conforming to the `AwsCloudFront` API
2. **Validate**: `planton validate --manifest manifest.yaml`
3. **Deploy**: 
   - Pulumi: `planton pulumi up --manifest manifest.yaml --stack org/project/stack`
   - Terraform: `planton terraform apply --manifest manifest.yaml --stack org/project/stack`

The CLI serializes your manifest into a protobuf `AwsCloudFrontStackInput`, injects provider credentials from the stack configuration, and invokes the chosen IaC backend. The IaC module (Pulumi or Terraform) provisions the CloudFront distribution and returns structured outputs.

## Common Patterns

### Static Website from S3
```yaml
apiVersion: aws.planton.dev/v1
kind: AwsCloudFront
metadata:
  name: my-website-cdn
spec:
  enabled: true
  priceClass: PRICE_CLASS_100
  origins:
    - domainName: my-website-bucket.s3.amazonaws.com
      isDefault: true
  defaultRootObject: index.html
```

### Static Website with Custom Domain
```yaml
apiVersion: aws.planton.dev/v1
kind: AwsCloudFront
metadata:
  name: my-website-cdn
spec:
  enabled: true
  aliases:
    - www.example.com
  certificateArn: arn:aws:acm:us-east-1:123456789012:certificate/abcd-1234-efgh-5678
  priceClass: PRICE_CLASS_100
  origins:
    - domainName: my-website-bucket.s3.amazonaws.com
      isDefault: true
  defaultRootObject: index.html
```

### Multi-Origin (Static Assets + API)
```yaml
apiVersion: aws.planton.dev/v1
kind: AwsCloudFront
metadata:
  name: multi-origin-cdn
spec:
  enabled: true
  aliases:
    - app.example.com
  certificateArn: arn:aws:acm:us-east-1:123456789012:certificate/abcd-1234-efgh-5678
  priceClass: PRICE_CLASS_100
  origins:
    - domainName: static-assets.s3.amazonaws.com
      originPath: /assets
      isDefault: true
    - domainName: api-alb-1234567890.us-west-2.elb.amazonaws.com
      isDefault: false
  defaultRootObject: index.html
```

## Common Troubleshooting

### "Invalid viewer certificate" error
**Cause**: The ACM certificate is not in `us-east-1`.
**Solution**: Create or import your certificate in the `us-east-1` region, regardless of where your origin is located.

### "403 Forbidden" from S3 origin
**Cause**: The S3 bucket policy doesn't grant CloudFront read access.
**Solution**: Add a bucket policy allowing the CloudFront distribution's service principal or OAI to read objects. (Note: This implementation uses custom origin config; for OAC, see the Pulumi or Terraform modules directly.)

### Cache not updating after deployment
**Cause**: Cached content hasn't expired yet.
**Best Practice**: Use file versioning (content-hashed filenames like `app.a1b2c3d4.js`) instead of cache invalidation. Only `index.html` should have a short TTL.
**Quick Fix**: Create a cache invalidation: `aws cloudfront create-invalidation --distribution-id <id> --paths "/*"`

### High costs
**Cause**: Using `PRICE_CLASS_ALL` when serving primarily NA/EU users.
**Solution**: Switch to `PRICE_CLASS_100` to reduce costs by 40-60%.

## Security Best Practices

1. **Always Use HTTPS**: This module enforces `redirect-to-https` by default
2. **Enable AWS WAF**: Integrate WAF for protection against SQL injection, XSS, and DDoS attacks (not exposed in this 80/20 API; configure separately)
3. **Private S3 Buckets**: Never make S3 buckets public—use Origin Access Control (OAC) or bucket policies
4. **Certificate Validation**: Use DNS validation for ACM certificates for automated renewal
5. **Geo-Restrictions**: If required, configure geo-blocking via CloudFront settings (advanced use case, not in 80/20 API)

## References

- **CloudFront Distributions**: https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/distribution-working-with.html
- **Price Classes**: https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/PriceClass.html
- **ACM Certificates for CloudFront**: https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/cnames-and-https-requirements.html
- **Origin Access Control (OAC)**: https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-restricting-access-to-s3.html
- **Cache Invalidation Best Practices**: https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/Invalidation.html

## Design Philosophy

This resource implements the **80/20 principle**: it exposes the 20% of CloudFront configuration that covers 80% of production use cases. Advanced features like Lambda@Edge, origin failover groups, custom error responses, and path-based cache behaviors are intentionally omitted to reduce complexity.

For advanced use cases, you can:
- Use the Pulumi or Terraform modules directly (in `iac/pulumi/` or `iac/tf/`)
- Extend the protobuf spec in a future API version
- Compose multiple `AwsCloudFront` resources for different distributions

See `docs/README.md` for the comprehensive research behind these design decisions.

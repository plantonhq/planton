# AWS Static Website

Provisions a static website backed by S3 for content storage and CloudFront for global CDN delivery. Optionally adds a custom domain with Route53 DNS and an ACM SSL certificate for HTTPS.

Upload your HTML, CSS, JavaScript, and other assets to the S3 bucket after deployment. CloudFront caches content at edge locations worldwide, delivering sub-100ms load times to visitors regardless of geography.

## Architecture

```
                              Visitors
                                 │
                                 ▼
{% if dnsEnabled %}
                    ┌────────────────────────┐
                    │   AwsRoute53DnsRecord   │
                    │   (A alias → CloudFront)│
                    └───────────┬────────────┘
                                │
{% endif %}
                    ┌───────────▼────────────┐
                    │    AwsCloudFront        │
                    │    (global CDN)         │
                    │                        │
{% if dnsEnabled %}
                    │  custom domain + SSL   │
{% endif %}
                    └───────────┬────────────┘
                                │ origin
                                ▼
                    ┌────────────────────────┐
                    │     AwsS3Bucket        │
                    │  (website content)     │
                    └────────────────────────┘

{% if dnsEnabled %}
  ┌──────────────────┐      ┌──────────────────┐
  │  AwsRoute53Zone  │◀────▶│ AwsCertManagerCert│
  │  (hosted zone)   │      │ (DNS-validated)  │
  └──────────────────┘      └──────────────────┘
{% endif %}
```

## Dependency Graph

```
Layer 0 (parallel):  AwsS3Bucket, AwsRoute53Zone
Layer 1 (dep R53):   AwsCertManagerCert
Layer 2 (dep S3+Cert): AwsCloudFront
Layer 3 (dep CF+R53):  AwsRoute53DnsRecord
```

## Included Cloud Resources

| Resource | Kind | Group | Condition | Purpose |
|----------|------|-------|-----------|---------|
| S3 Bucket | `AwsS3Bucket` | storage | Always | Static website content storage |
| CloudFront Distribution | `AwsCloudFront` | compute | Always | Global CDN edge caching |
| Route53 Hosted Zone | `AwsRoute53Zone` | network | `dnsEnabled` | DNS management for custom domain |
| ACM Certificate | `AwsCertManagerCert` | security | `dnsEnabled` | SSL/TLS for HTTPS |
| DNS Alias Record | `AwsRoute53DnsRecord` | network | `dnsEnabled` | Points custom domain to CloudFront |

## Parameters

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `aws_region` | AWS region for S3 bucket (CloudFront is global) | `us-east-1` | Yes |
| `s3_bucket_name` | Globally unique S3 bucket name | `my-website-content` | Yes |
| `default_root_object` | File served at root URL | `index.html` | Yes |
| **DNS & SSL** | | | |
| `dnsEnabled` | Create Route53 zone + ACM cert for custom domain | `true` | No |
| `domain_name` | Root domain for hosted zone (e.g., example.com) | `example.com` | When `dnsEnabled` |
| `website_domain_name` | Full website domain (e.g., www.example.com) | `www.example.com` | When `dnsEnabled` |

## Common Configurations

### Quick CDN (no custom domain)

Serve content via the CloudFront domain (e.g., `d1234abcd.cloudfront.net`):

```yaml
dnsEnabled: false
s3_bucket_name: my-app-assets
```

### Custom Domain with HTTPS

Serve content at `www.example.com` with SSL:

```yaml
dnsEnabled: true
domain_name: example.com
website_domain_name: www.example.com
s3_bucket_name: example-com-website
```

### Apex Domain

Serve content at `example.com` (no www):

```yaml
dnsEnabled: true
domain_name: example.com
website_domain_name: example.com
s3_bucket_name: example-com-website
```

## Post-Deployment Steps

1. **Upload content** to the S3 bucket:
   ```bash
   aws s3 sync ./dist s3://my-website-content
   ```

2. **Verify CloudFront** is serving your content at the distribution domain shown in deployment outputs.

3. **If using a custom domain**: Update your domain registrar's nameservers to the Route53 nameservers from the deployment outputs. DNS propagation may take up to 48 hours.

4. **Invalidate cache** after content updates:
   ```bash
   aws cloudfront create-invalidation --distribution-id EXXXXX --paths "/*"
   ```

## Important Notes

- The S3 bucket is configured with **public access** so CloudFront can serve objects. For production, consider adding a bucket policy to restrict direct S3 access and force traffic through CloudFront.
- ACM certificates require **DNS validation**. The certificate will not issue until the Route53 zone is authoritative for the domain (nameservers must be pointed).
- CloudFront is a **global** service. The distribution is deployed to all edge locations regardless of the S3 bucket's region.
- The `s3_bucket_name` must be **globally unique** across all AWS accounts.
- When `dnsEnabled` is `false`, the site is accessible only via the CloudFront domain (`d1234abcd.cloudfront.net`).

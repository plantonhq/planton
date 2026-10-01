# Custom Domain CDN with ACM Certificate

This preset serves a private S3 bucket through CloudFront on your own domain: an alias (CNAME) backed by an ACM certificate, with an Origin Access Control keeping the bucket private.

## When to Use

- Production websites and CDNs on a branded domain (`cdn.example.com`, `www.example.com`)
- Any distribution that must present your own TLS certificate instead of `*.cloudfront.net`

## Key Configuration Choices

- **Alias + ACM certificate** -- the certificate is referenced from an `AwsCertManagerCert` resource and **must live in us-east-1** (CloudFront's global requirement) and cover every alias; SNI-only serving and the `TLSv1.2_2021` floor are applied automatically
- **DNS not included** -- point the domain at the distribution with an `AwsRoute53DnsRecord` alias record targeting the `domain_name` and `hosted_zone_id` outputs
- **Origin Access Control** (`s3Origin.createOriginAccessControl: true`) -- the bucket stays private; allow the distribution's ARN in the bucket policy
- **IPv6 enabled** -- free dual-stack serving; create AAAA alias records alongside the A records

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `cdn.replaceme.example.com` | The custom domain the distribution answers for (a real DNS shape — the field's pattern rejects placeholders) | Your DNS plan |
| `<certificate-resource-name>` | The `AwsCertManagerCert` resource covering the domain (in us-east-1) | Your certificate manifest |
| `replace-me-bucket.s3.us-east-1.amazonaws.com` | The S3 bucket holding the content (replace the bucket and region segments) | `AwsS3Bucket` outputs |

## Related Presets

- **01-s3-static-website** -- Use when the default `*.cloudfront.net` domain is enough (no certificate needed)

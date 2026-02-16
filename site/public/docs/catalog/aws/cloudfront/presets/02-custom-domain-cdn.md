---
title: "Custom Domain CDN"
description: "This preset creates a CloudFront distribution with a custom domain name and HTTPS via an ACM certificate. It uses Price Class 200 for broader geographic coverage (US, Canada, Europe, Asia, Middle..."
type: "preset"
rank: "02"
presetSlug: "02-custom-domain-cdn"
componentSlug: "cloudfront"
componentTitle: "CloudFront"
provider: "aws"
icon: "package"
order: 2
---

# Custom Domain CDN

This preset creates a CloudFront distribution with a custom domain name and HTTPS via an ACM certificate. It uses Price Class 200 for broader geographic coverage (US, Canada, Europe, Asia, Middle East, Africa) and supports any origin (S3 bucket, ALB, API Gateway, or custom server).

## When to Use

- Production websites or APIs that need a custom domain (e.g., `cdn.example.com`) with HTTPS
- Global applications requiring edge caching across multiple continents
- Any CloudFront distribution that needs a branded domain instead of `*.cloudfront.net`

## Key Configuration Choices

- **Custom domain** (`aliases`) -- Associates a CNAME with the distribution; requires a matching DNS record pointing to the CloudFront domain
- **ACM certificate** (`certificateArn`) -- Must be in **us-east-1** regardless of where your origin is located (CloudFront requirement)
- **Price Class 200** (`priceClass: PRICE_CLASS_200`) -- Edge locations in US, Canada, Europe, Asia, Middle East, and Africa; good balance of coverage and cost
- **Single origin** -- One default origin; add more for multi-origin architectures (API + static assets)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-cdn-domain.com>` | Custom domain for the distribution (e.g., `cdn.example.com`) | Your domain registrar |
| `<acm-certificate-arn-us-east-1>` | ACM certificate ARN in us-east-1 covering the custom domain | AWS ACM console (us-east-1) or `AwsCertManagerCert` status outputs |
| `<origin-domain-name>` | Origin server DNS name (e.g., `my-bucket.s3.amazonaws.com` or `my-alb.us-east-1.elb.amazonaws.com`) | Your origin resource |

## Related Presets

- **01-s3-static-website** -- Use instead for quick S3 website hosting without a custom domain

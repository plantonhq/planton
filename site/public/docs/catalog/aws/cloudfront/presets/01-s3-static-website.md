---
title: "S3 Static Website"
description: "This preset creates a CloudFront distribution fronting an S3 bucket for static website hosting. It uses the most cost-effective price class (US, Canada, and Europe edge locations) and sets..."
type: "preset"
rank: "01"
presetSlug: "01-s3-static-website"
componentSlug: "cloudfront"
componentTitle: "CloudFront"
provider: "aws"
icon: "package"
order: 1
---

# S3 Static Website

This preset creates a CloudFront distribution fronting an S3 bucket for static website hosting. It uses the most cost-effective price class (US, Canada, and Europe edge locations) and sets `index.html` as the default root object. No custom domain or SSL certificate is configured -- the site is accessible via the CloudFront-assigned domain (e.g., `d1234abcd.cloudfront.net`).

## When to Use

- Static websites (HTML, CSS, JavaScript) hosted in S3
- Single-page applications (React, Vue, Angular) served from S3
- Quick deployments where a custom domain is not yet needed

## Key Configuration Choices

- **Price Class 100** (`priceClass: PRICE_CLASS_100`) -- Edge locations in US, Canada, and Europe only; lowest cost tier
- **S3 origin** -- Single origin pointing to an S3 bucket; CloudFront caches and serves content from edge locations
- **Default root object** (`defaultRootObject: index.html`) -- Requests to the root URL (`/`) serve `index.html`
- **No custom domain** -- Accessible via CloudFront's auto-assigned `*.cloudfront.net` domain

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<s3-bucket-domain-name>` | S3 bucket website endpoint (e.g., `my-website-bucket.s3.amazonaws.com`) | AWS S3 console or `AwsS3Bucket` status outputs |

## Related Presets

- **02-custom-domain-cdn** -- Use instead when you need a custom domain with HTTPS (requires ACM certificate in us-east-1)

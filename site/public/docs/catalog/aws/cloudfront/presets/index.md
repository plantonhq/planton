---
title: "Presets"
description: "Ready-to-deploy configuration presets for CloudFront"
type: "preset-list"
componentSlug: "cloudfront"
componentTitle: "CloudFront"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-s3-static-website"
    rank: "01"
    title: "S3 Static Website"
    excerpt: "This preset creates a CloudFront distribution fronting an S3 bucket for static website hosting. It uses the most cost-effective price class (US, Canada, and Europe edge locations) and sets..."
  - slug: "02-custom-domain-cdn"
    rank: "02"
    title: "Custom Domain CDN"
    excerpt: "This preset creates a CloudFront distribution with a custom domain name and HTTPS via an ACM certificate. It uses Price Class 200 for broader geographic coverage (US, Canada, Europe, Asia, Middle..."
---

# CloudFront Presets

Ready-to-deploy configuration presets for CloudFront. Each preset is a complete manifest you can copy, customize, and deploy.

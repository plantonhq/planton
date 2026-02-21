---
title: "OSS Static Assets CDN"
description: "This preset creates a CDN-accelerated domain with an Alibaba Cloud OSS bucket as the origin. This is the standard pattern for serving static website assets (images, CSS, JavaScript, fonts) or hosting..."
type: "preset"
rank: "02"
presetSlug: "02-oss-static-assets"
componentSlug: "cdndomain"
componentTitle: "CdnDomain"
provider: "alicloud"
icon: "package"
order: 2
---

# OSS Static Assets CDN

This preset creates a CDN-accelerated domain with an Alibaba Cloud OSS bucket as the origin. This is the standard pattern for serving static website assets (images, CSS, JavaScript, fonts) or hosting a static site behind CDN edge caching. No HTTPS is configured by default; add a `certificateConfig` section after completing DNS CNAME verification.

## When to Use

- Static websites hosted entirely in OSS
- Frontend SPA (single-page application) asset delivery
- Serving images, downloads, or media files stored in an OSS bucket
- Development or staging CDN setups where HTTPS can be added after DNS verification

## Key Configuration Choices

- **OSS origin** (`type: oss`) -- The CDN pulls content directly from an OSS bucket using Alibaba Cloud's internal network, which is faster and cheaper than pulling over the public internet. The content field must be the full OSS bucket domain.
- **Web CDN type** (`cdnType: web`) -- Suitable for static assets. The CDN optimizes for small file caching and connection reuse.
- **Health check URL** (`checkUrl`) -- Points to a small file in the bucket (e.g., `health.txt`) that Alibaba Cloud uses during domain creation to verify the origin is reachable. Create this file in the bucket root before deploying.
- **No HTTPS** -- HTTPS requires DNS CNAME verification first. After the CDN domain is created and you have pointed your DNS CNAME to the CDN's `cname` output, add a `certificateConfig` with either a free DV certificate (`certType: free`) or a CAS certificate (`certType: cas`).
- **Default scope** (omitted, defaults to `domestic`) -- Mainland China edge nodes. Set `scope: global` if serving international users.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<alibaba-cloud-region>` | Region for provider API calls | Any region, typically `cn-hangzhou` |
| `<your-cdn-domain>` | Domain name to accelerate (e.g., `assets.example.com`) | Your domain registrar |
| `<your-bucket-name>` | OSS bucket name | `AliCloudStorageBucket` stack outputs |
| `<bucket-region>` | Region where the OSS bucket is located (e.g., `cn-hangzhou`) | `AliCloudStorageBucket` configuration |

## Related Presets

- **01-web-https** -- Use when the origin is a domain or IP address and HTTPS is required from day one

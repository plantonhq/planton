---
title: "Presets"
description: "Ready-to-deploy configuration presets for CdnDomain"
type: "preset-list"
componentSlug: "cdndomain"
componentTitle: "CdnDomain"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-web-https"
    rank: "01"
    title: "Web Acceleration with HTTPS"
    excerpt: "This preset creates a CDN-accelerated domain for web content (images, small files, web pages) with HTTPS enabled via an Alibaba Cloud Certificate Management Service (CAS) certificate. A single..."
  - slug: "02-oss-static-assets"
    rank: "02"
    title: "OSS Static Assets CDN"
    excerpt: "This preset creates a CDN-accelerated domain with an Alibaba Cloud OSS bucket as the origin. This is the standard pattern for serving static website assets (images, CSS, JavaScript, fonts) or hosting..."
---

# CdnDomain Presets

Ready-to-deploy configuration presets for CdnDomain. Each preset is a complete manifest you can copy, customize, and deploy.

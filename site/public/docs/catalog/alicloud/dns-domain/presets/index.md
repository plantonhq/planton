---
title: "Presets"
description: "Ready-to-deploy configuration presets for DNS Domain"
type: "preset-list"
componentSlug: "dns-domain"
componentTitle: "DNS Domain"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard Domain Registration"
    excerpt: "This preset registers a domain in Alibaba Cloud DNS (Alidns) with only the required fields. After deployment, point your domain registrar's NS records to the DNS servers returned in the stack outputs."
  - slug: "02-organizational"
    rank: "02"
    title: "Organizational Domain Registration"
    excerpt: "This preset registers a domain in Alibaba Cloud DNS (Alidns) with resource group placement, remarks, and organizational tags. Suitable for production environments where governance, access control,..."
---

# DNS Domain Presets

Ready-to-deploy configuration presets for DNS Domain. Each preset is a complete manifest you can copy, customize, and deploy.

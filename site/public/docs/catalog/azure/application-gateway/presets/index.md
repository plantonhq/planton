---
title: "Presets"
description: "Ready-to-deploy configuration presets for Application Gateway"
type: "preset-list"
componentSlug: "application-gateway"
componentTitle: "Application Gateway"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-http-basic"
    rank: "01"
    title: "Basic HTTP Application Gateway"
    excerpt: "This preset creates an Azure Application Gateway v2 with Standard_v2 SKU, a single HTTP listener on port 80, one backend pool with FQDN-based targets, and a custom health probe. This is the simplest..."
  - slug: "02-https-waf"
    rank: "02"
    title: "HTTPS Application Gateway with WAF"
    excerpt: "This preset creates an Azure Application Gateway v2 with WAF_v2 SKU, HTTPS termination using a Key Vault certificate, Web Application Firewall in Prevention mode, and an HTTP listener for..."
---

# Application Gateway Presets

Ready-to-deploy configuration presets for Application Gateway. Each preset is a complete manifest you can copy, customize, and deploy.

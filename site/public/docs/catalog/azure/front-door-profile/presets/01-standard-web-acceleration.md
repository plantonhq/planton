---
title: "Standard Web Acceleration"
description: "A Standard-tier Front Door profile optimized for accelerating a web application with edge caching and compression."
type: "preset"
rank: "01"
presetSlug: "01-standard-web-acceleration"
componentSlug: "front-door-profile"
componentTitle: "Front Door Profile"
provider: "azure"
icon: "package"
order: 1
---

# Standard Web Acceleration

A Standard-tier Front Door profile optimized for accelerating a web application with edge caching and compression.

## What This Preset Provides

- **Standard SKU**: Global CDN with SSL offloading, caching, and compression at 99.99% SLA
- **Single endpoint**: One public entry point with a generated `*.azurefd.net` hostname
- **Single origin**: Configured for an Azure App Service backend with proper Host header
- **HTTPS health probes**: Periodic health checks every 30 seconds to detect origin failures
- **Edge caching**: Responses cached at edge locations with query string ignored for cache keys
- **Compression**: Gzip/Brotli compression enabled for common web content types (HTML, CSS, JS, JSON, SVG)
- **HTTPS redirect**: HTTP requests automatically redirected to HTTPS

## When to Use

- Single web application that needs global CDN acceleration
- Static + dynamic content mix served from a single backend
- App Service, Container App, or similar PaaS backend
- No need for multi-origin failover or path-based routing

## What to Customize

| Field | What to Set |
|-------|-------------|
| `resourceGroup.value` | Your Azure resource group name |
| `name` | Globally unique profile name |
| `origins[].hostName` | Your backend's hostname |
| `origins[].originHostHeader` | Usually same as `hostName` for PaaS backends |
| `contentTypesToCompress` | Add or remove MIME types based on your content |

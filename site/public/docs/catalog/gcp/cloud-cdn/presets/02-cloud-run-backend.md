---
title: "Cloud Run Backend CDN"
description: "This preset creates a Cloud CDN backed by a Cloud Run service, with caching controlled by the origin's Cache-Control headers. It includes a custom domain with HTTPS redirect, making it suitable for..."
type: "preset"
rank: "02"
presetSlug: "02-cloud-run-backend"
componentSlug: "cloud-cdn"
componentTitle: "Cloud CDN"
provider: "gcp"
icon: "package"
order: 2
---

# Cloud Run Backend CDN

This preset creates a Cloud CDN backed by a Cloud Run service, with caching controlled by the origin's Cache-Control headers. It includes a custom domain with HTTPS redirect, making it suitable for APIs or dynamic web applications that benefit from edge caching.

## When to Use

- Cloud Run services where some responses are cacheable (API with Cache-Control headers)
- Web applications that serve a mix of dynamic and cacheable content
- Services needing a custom domain with Google-managed HTTPS and global load balancing

## Key Configuration Choices

- **Cloud Run backend** (`cloudRunService`) -- creates a Serverless NEG pointing to the Cloud Run service
- **Origin headers mode** (`cacheMode: USE_ORIGIN_HEADERS`) -- only caches responses with explicit Cache-Control headers; gives the application full control
- **Negative caching** (`enableNegativeCaching: true`) -- caches error responses to protect the backend during failures
- **Custom domain with HTTPS redirect** -- routes traffic to the service via a custom domain with automatic HTTP-to-HTTPS redirect
- **Google-managed SSL** -- auto-provisioned and auto-renewed certificate for the custom domain

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |
| `<cloud-run-service-name>` | Name of the Cloud Run service | `GcpCloudRun` metadata name |
| `<gcp-region>` | Region where the Cloud Run service is deployed | `GcpCloudRun` spec region |
| `<your-domain.com>` | Custom domain for this CDN endpoint | Your DNS registrar |

## Related Presets

- **01-gcs-static-site** -- Use for static content hosted in a GCS bucket

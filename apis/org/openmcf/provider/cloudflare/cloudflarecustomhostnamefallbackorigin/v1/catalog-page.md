# Cloudflare Custom Hostname Fallback Origin

Set the default origin for a Cloudflare for SaaS zone — the backend that all of the
zone's custom hostnames route to by default. Required once per SaaS zone before
custom hostnames can serve traffic.

## What Gets Created

- A `cloudflare_custom_hostname_fallback_origin` for the zone.

## Prerequisites

- A SaaS zone (`CloudflareDnsZone`).
- A Cloudflare API token with `SSL and Certificates` permission.
- The fallback origin hostname must be a record within the SaaS zone.

## Configuration Reference

**Required**

- `zoneId` — the SaaS zone.
- `origin` — the fallback origin hostname.

## Stack Outputs

| Output | Description |
|---|---|
| `status` | Deployment status |
| `created_at` | Creation timestamp |
| `updated_at` | Last-update timestamp |
| `errors` | Any deployment errors |

## Related Components

- `CloudflareCustomHostname`, `CloudflareDnsZone`

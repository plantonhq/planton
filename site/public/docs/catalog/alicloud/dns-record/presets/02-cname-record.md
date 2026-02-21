---
title: "CNAME Record"
description: "This preset creates a DNS CNAME record that aliases a subdomain to another domain name. CNAME records are commonly used for CDN integration, service abstraction, and multi-environment routing."
type: "preset"
rank: "02"
presetSlug: "02-cname-record"
componentSlug: "dns-record"
componentTitle: "DNS Record"
provider: "alicloud"
icon: "package"
order: 2
---

# CNAME Record

This preset creates a DNS CNAME record that aliases a subdomain to another domain name. CNAME records are commonly used for CDN integration, service abstraction, and multi-environment routing.

## When to Use

- Pointing a subdomain at a CDN endpoint (e.g., `cdn.example.com` -> `example.com.cdn-provider.com`)
- Aliasing a service name to a load balancer DNS name (e.g., `api.example.com` -> `my-alb.cn-hangzhou.alb.aliyuncs.com`)
- Creating environment-specific aliases (e.g., `staging.example.com` -> `staging-app.internal.com`)
- Any scenario where you want a subdomain to resolve to whatever IP the target domain resolves to

## Key Configuration Choices

- **CNAME restrictions** -- CNAME records cannot coexist with other record types for the same `rr`. A CNAME at the apex (`@`) is not recommended as it conflicts with NS and SOA records.
- **No trailing dot** -- the Alibaba Cloud API does not expect a trailing dot on CNAME targets. The provider normalizes trailing dots during diff suppression.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`) | Your deployment region strategy |
| `<your-domain-name>` | The parent domain (e.g., `example.com`). Must already exist in Alidns. | Your AliCloudDnsZone resource or Alidns console |
| `<host-record>` | Subdomain label (e.g., `cdn`, `api`, `app`) | Your DNS design |
| `<target-domain>` | The domain to alias to (e.g., `example.com.cdn-provider.com`) | Your CDN or service provider |

## Post-Deployment Steps

1. Deploy the manifest to create the CNAME record
2. Verify with `dig CNAME <host-record>.<your-domain-name>`
3. Ensure the target domain resolves correctly

## Related Presets

- **01-a-record** -- use instead to point directly at an IP address
- **03-mx-record** -- use for email routing with priority

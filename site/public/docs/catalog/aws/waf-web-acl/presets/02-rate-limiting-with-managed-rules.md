---
title: "Rate Limiting with Managed Rules"
description: "This preset creates a Web ACL that combines IP-based rate limiting with three AWS Managed Rule Groups. The rate limit is evaluated first (lowest priority number) to block volumetric attacks before..."
type: "preset"
rank: "02"
presetSlug: "02-rate-limiting-with-managed-rules"
componentSlug: "waf-web-acl"
componentTitle: "WAF Web ACL"
provider: "aws"
icon: "package"
order: 2
---

# Rate Limiting with Managed Rules

This preset creates a Web ACL that combines IP-based rate limiting with three AWS Managed Rule Groups. The rate limit is evaluated first (lowest priority number) to block volumetric attacks before the more computationally expensive managed rules run.

## When to Use

- Public-facing APIs and web applications
- Services vulnerable to brute-force or credential stuffing
- Applications where both volumetric and content-based attacks are concerns
- Cost-conscious setups where rate limiting reduces WCU consumption by blocking floods early

## Key Configuration Choices

- **Rate limit at priority 1** -- evaluated first to block floods before managed rules
- **2000 requests per 5 minutes** -- approximately 6-7 requests/second per IP; tune based on your traffic pattern
- **Block action on rate limit** -- immediately blocks requests exceeding the limit
- **Three managed rule groups** -- Common Rules + SQL injection + Known Bad Inputs

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<acl-name>` | Unique name for the Web ACL (lowercase, alphanumeric, hyphens) |

## Tuning

- Adjust `limit` based on your application's normal traffic. API endpoints may need higher limits (5000-10000), while login endpoints may need lower limits (100-500).
- Set `evaluationWindowSec: 60` for faster burst detection.
- Add `aggregateKeyType: FORWARDED_IP` with `forwardedIpConfig` if behind a CDN or proxy.
- Add a `scopeDownStatement` to the rate rule to rate-limit only specific paths (e.g., /login, /api).

## Related Presets

- **01-managed-rules-basic** -- managed rules without rate limiting
- **03-production-web-app** -- full production config with geo blocking and logging

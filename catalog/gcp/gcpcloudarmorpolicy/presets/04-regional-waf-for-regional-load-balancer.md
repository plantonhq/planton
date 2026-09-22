# Regional WAF for a Regional Load Balancer

## Use Case

Put the same WAF a global load balancer gets in front of a regional external or internal Application Load Balancer. A regional backend service can attach only a regional Cloud Armor policy, so this preset builds one: a break-glass allow at priority 0, a corporate allowlist, a geo block in preview, and a per-IP throttle, with a default allow so the throttle -- not the allowlist -- is the guard.

## When to Use

- A service that lives in one region behind a regional external ALB (the regional `GcpBackendService`, `GcpUrlMap`, `GcpTargetHttpProxy`, and `GcpGlobalForwardingRule` chain)
- An internal ALB whose callers should still be rate-limited and geo-fenced
- Any policy a regional backend service's `securityPolicy` field must point at

## What This Creates

- A REGIONAL Cloud Armor policy of type `CLOUD_ARMOR` in `us-central1`
- Priority 0: allow the break-glass range (`198.51.100.0/24`) -- 0 is Google's highest priority and is evaluated first
- Priority 1000: allow the corporate ranges
- Priority 2000: deny(403) two region codes, in preview (logged, not enforced) until the logs prove it right
- Priority 3000: throttle every source to 100 requests per minute per IP, exceeding to deny(429)
- Priority 2147483647: the mandatory default rule, allow
- JSON body parsing and verbose logging (the two advanced options the regional collection carries)

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `region` | `us-central1` | The region of the backend service that attaches this policy; scopes must match. Immutable. |
| `srcIpRanges` (priorities 0 and 1000) | break-glass + RFC 1918 | Your on-call egress, VPN, and office ranges. Max 10 per rule. |
| `expression` (priority 2000) | CN, RU | The region codes your service does not serve; flip `preview` off once the logs agree. |
| `rateLimitThreshold` | 100 / 60s | Measure real traffic first; tighten from data, not intuition. |
| default rule `action` | `allow` | `deny(403)` turns the allowlist into the guard (deny-by-default). |

What this preset cannot carry, because the regional collection lacks it: labels, Adaptive Protection, reCAPTCHA redirects, header injection, and `requestBodyInspectionSize` -- each is rejected before deploy when `region` is set.

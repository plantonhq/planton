# Production Web Application

This preset creates a comprehensive Web ACL suitable for production web applications. It combines rate limiting, geographic blocking, five AWS Managed Rule Groups, custom response bodies, and logging with sensitive field redaction.

## When to Use

- Production web applications and APIs with public internet exposure
- Applications requiring defense-in-depth security
- Compliance-sensitive environments that need request logging
- Services that need to block traffic from specific geographic regions

## Key Configuration Choices

- **Rate limiting at priority 1** -- first line of defense against volumetric attacks
- **Geographic blocking at priority 2** -- blocks traffic from high-risk countries before expensive rule evaluation
- **IP reputation (priority 10-11)** -- blocks known malicious IPs and anonymous proxy traffic
- **Common Rules + SQLi + Known Bad Inputs (priority 20-40)** -- comprehensive OWASP protection
- **Custom response body** -- JSON error response for rate-limited requests
- **Logging enabled** -- sends request logs to CloudWatch Logs with Authorization, Cookie, and query string redacted

## Rule Evaluation Order

Rules are evaluated by priority (lowest first):
1. Rate limit -- blocks IPs exceeding 2000 requests per 5 minutes
2. Geo block -- blocks traffic from specified countries
3. IP reputation -- blocks known malicious and anonymous IPs
4. Common Rules -- blocks XSS, file inclusion, and other OWASP threats
5. SQLi Rules -- blocks SQL injection attempts
6. Known Bad Inputs -- blocks Log4j, Java deserialization exploits

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<acl-name>` | Unique name for the Web ACL |
| `<country-code-1>`, `<country-code-2>` | ISO 3166-1 alpha-2 country codes to block (e.g., RU, CN, KP) |
| `<log-group-arn>` | ARN of a CloudWatch Logs log group (name must start with `aws-waf-logs-`) |

## Tuning Recommendations

- Start with `overrideAction: count` on new managed rule groups, monitor for false positives, then switch to `none`
- Use `ruleActionOverrides` to set specific noisy rules to `count` within a group
- Adjust the rate limit based on your application's traffic patterns
- Add an IP set allowlist rule at priority 0 for trusted partners or monitoring services
- Consider adding `AWSManagedRulesBotControlRuleSet` for bot management (additional WCU cost)

## Related Presets

- **01-managed-rules-basic** -- simpler setup without rate limiting or geo blocking
- **02-rate-limiting-with-managed-rules** -- intermediate setup without geo blocking or logging

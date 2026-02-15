# AwsWafWebAcl — Architecture and Design

## Overview

AWS WAFv2 (Web Application Firewall) inspects incoming HTTP/HTTPS requests at the edge of your application and applies rules to allow, block, rate-limit, or challenge traffic. A Web ACL is the top-level container for these rules.

## Rule Evaluation Model

```
Request arrives
    │
    ▼
Priority 1 rule → match? → take action (STOP)
    │ no match
    ▼
Priority 2 rule → match? → take action (STOP)
    │ no match
    ▼
  ... (continue in priority order)
    │ no match
    ▼
Default action (allow or block)
```

Rules are evaluated in ascending priority order (lowest number first). When a rule matches, its action is taken immediately and no further rules are evaluated for that request. If no rule matches, the Web ACL's default action applies.

## Statement Types

### First-Class (Typed Proto Messages)

**Managed Rule Groups** — Pre-configured rule collections maintained by AWS or marketplace vendors. Each group contains multiple internal rules. Use `overrideAction` to control behavior:
- `none` — use the group's own actions (block/count as configured by AWS)
- `count` — override all internal actions to count (useful for testing)

Within a group, `ruleActionOverrides` lets you override specific internal rules (e.g., set a noisy rule to `count` while the rest enforce).

**Rate-Based** — Tracks request rates from individual keys (typically IPs) and triggers when the rate exceeds a threshold within an evaluation window. Effective against:
- DDoS and HTTP floods
- Brute-force login attempts
- API abuse and scraping

**Geo Match** — Matches based on the client's country of origin (determined from source IP or a forwarded header). Used for geographic restrictions and compliance.

**IP Set Reference** — Matches against a pre-defined WAFv2 IP Set (a separate resource containing IP addresses or CIDR ranges). Used for allowlists and blocklists.

### Escape Hatch (google.protobuf.Struct)

For all other statement types, use `customStatement` with the raw AWS WAFv2 JSON format:

- **SqliMatchStatement** — SQL injection detection
- **XssMatchStatement** — Cross-site scripting detection
- **ByteMatchStatement** — String/pattern matching in request fields
- **RegexMatchStatement** — Regex pattern matching
- **SizeConstraintStatement** — Request size limits
- **LabelMatchStatement** — Match labels from previous rules
- **AndStatement / OrStatement / NotStatement** — Compound logic
- **RuleGroupReferenceStatement** — Custom WAFv2 Rule Groups

## Capacity Model (WCUs)

Each statement type consumes Web ACL Capacity Units (WCUs). The total must not exceed 5,000 WCUs per Web ACL.

| Resource | Approximate WCUs |
|----------|-----------------|
| AWSManagedRulesCommonRuleSet | 700 |
| AWSManagedRulesSQLiRuleSet | 200 |
| AWSManagedRulesKnownBadInputsRuleSet | 200 |
| AWSManagedRulesAmazonIpReputationList | 25 |
| AWSManagedRulesAnonymousIpList | 50 |
| AWSManagedRulesBotControlRuleSet | 50 (common) / 100 (targeted) |
| Rate-based rule | 2 |
| Geo match rule | 1 |
| IP set reference | 1 |
| SQLi match | 20 |
| XSS match | 40 |

Monitor the `capacity` stack output to plan rule additions.

## Scope and Region

- **REGIONAL** — created in the same region as the protected resource. Protects ALB, API Gateway, AppSync, Cognito, App Runner.
- **CLOUDFRONT** — must be created in `us-east-1`. The CloudFront distribution references the Web ACL ARN directly.

## Association Model

This component does NOT bundle associations. The Web ACL ARN is exported as `web_acl_arn` and consumed by:

- **AwsAlb** — via `web_acl_arn` field or AWS WAF association
- **AwsHttpApiGateway** — via stage-level WAF association
- **AWS CloudFront** — via distribution configuration

This separation follows the principle that the Web ACL is a security policy with an independent lifecycle from the resources it protects.

## Logging Architecture

WAF logging is bundled as an optional inline configuration. When enabled, every inspected request generates a log entry containing:

- Timestamp and request ID
- Matched rule name and action taken
- Client IP and country
- HTTP method, URI, and headers
- Terminating rule details

**Destinations** (name must start with `aws-waf-logs-`):
- CloudWatch Logs — best for real-time monitoring and CloudWatch Insights queries
- S3 — best for long-term storage and compliance archival
- Kinesis Firehose — best for streaming to analytics platforms

**Field redaction** prevents sensitive data from appearing in logs:
- `redactedHeaderNames` — redacts specific headers (Authorization, Cookie, etc.)
- `redactUriPath` — redacts the URI path
- `redactQueryString` — redacts query string parameters

## IaC Implementation

Both Pulumi and Terraform modules use the `rule_json` approach, constructing rules as JSON in the AWS WAFv2 API format (PascalCase keys). This provides:

1. Uniform handling of typed and custom statements
2. Full WAFv2 API coverage via the escape hatch
3. Consistent behavior between Pulumi and Terraform

The visibility config applies smart defaults when omitted:
- `cloudwatch_metrics_enabled` = true
- `sampled_requests_enabled` = true
- `metric_name` = resource name (Web ACL level) or rule name (rule level)

## Security Best Practices

1. **Start permissive, tighten gradually** — deploy managed rules with `overrideAction: count` first, monitor for false positives, then switch to `none`.
2. **Rate limit early** — place rate-based rules at low priority numbers to block floods before expensive rule evaluation.
3. **Enable logging** — essential for tuning, forensics, and compliance.
4. **Redact sensitive headers** — always redact Authorization, Cookie, and API key headers from logs.
5. **Pin managed rule versions** — use the `version` field to prevent unexpected behavior changes.
6. **Monitor WCU usage** — the `capacity` output helps plan rule additions within the 5,000 limit.

## Cost Model

WAF pricing (as of 2024):
- $5/month per Web ACL
- $1/month per rule
- $0.60 per 1 million requests inspected
- Bot Control and Fraud Prevention have additional per-request charges

Managed rule group subscriptions are included at no extra cost for AWS Managed Rules.
